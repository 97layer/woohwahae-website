package runtime

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type GatewayAdapter interface {
	Name() string
	Semantics() string
	DispatchEnabled() bool
	RequiredMode() string
	Prepare(call GatewayCall, decision PolicyDecision) (GatewayCall, error)
	Dispatch(call GatewayCall, decision PolicyDecision) (GatewayCall, error)
}

type gatewayPayloadAdapter interface {
	GatewayAdapter
	DispatchWithPayload(call GatewayCall, decision PolicyDecision, payload map[string]any) (GatewayCall, error)
}

type VerifyAdapter interface {
	Name() string
	Run(command []string, workdir string) (string, error)
}

type DeployAdapter interface {
	Name() string
	Run(target DeployTarget) (string, error)
}

type recordGatewayAdapter struct{}

func prepareRecordedGatewayCall(call GatewayCall, decision PolicyDecision) GatewayCall {
	if call.Status == "" {
		call.Status = "recorded"
	}
	if call.Notes == nil {
		call.Notes = []string{}
	}
	call.AttemptCount = 0
	call.LastHTTPStatus = nil
	call.LastError = nil
	call.ResponsePreview = nil
	call.DispatchedAt = nil
	if call.TokenBudget == 0 {
		switch decision.TokenClass {
		case "large":
			call.TokenBudget = 64000
		case "medium":
			call.TokenBudget = 16000
		default:
			call.TokenBudget = 8000
		}
	}
	return call
}

func (recordGatewayAdapter) Name() string          { return "record" }
func (recordGatewayAdapter) Semantics() string     { return "record_only" }
func (recordGatewayAdapter) DispatchEnabled() bool { return false }
func (recordGatewayAdapter) RequiredMode() string  { return "single" }

func (recordGatewayAdapter) Prepare(call GatewayCall, decision PolicyDecision) (GatewayCall, error) {
	call = prepareRecordedGatewayCall(call, decision)
	call.Notes = appendUniqueString(call.Notes, "adapter:record")
	call.Notes = appendUniqueString(call.Notes, "dispatch:recorded-intent")
	return call, nil
}

func (recordGatewayAdapter) Dispatch(call GatewayCall, decision PolicyDecision) (GatewayCall, error) {
	return call, nil
}

type packetGatewayAdapter struct {
	base   GatewayAdapter
	reason string
}

func (a packetGatewayAdapter) Name() string {
	if a.base != nil {
		if value := strings.TrimSpace(a.base.Name()); value != "" {
			return value
		}
	}
	return "packet"
}

func (a packetGatewayAdapter) Semantics() string     { return "job_packet" }
func (a packetGatewayAdapter) DispatchEnabled() bool { return false }
func (a packetGatewayAdapter) RequiredMode() string {
	if a.base != nil {
		return a.base.RequiredMode()
	}
	return "single"
}

func (a packetGatewayAdapter) Prepare(call GatewayCall, decision PolicyDecision) (GatewayCall, error) {
	call = prepareRecordedGatewayCall(call, decision)
	call.Notes = appendUniqueString(call.Notes, "adapter:"+a.Name())
	call.Notes = appendUniqueString(call.Notes, "dispatch:job_packet")
	if reason := strings.TrimSpace(a.reason); reason != "" {
		call.Notes = appendUniqueString(call.Notes, "dispatch:reason="+reason)
	}
	return call, nil
}

func (a packetGatewayAdapter) Dispatch(call GatewayCall, decision PolicyDecision) (GatewayCall, error) {
	return call, nil
}

type apiGatewayAdapter struct {
	httpClient *http.Client
}

func (a apiGatewayAdapter) Name() string          { return "api" }
func (a apiGatewayAdapter) Semantics() string     { return "dispatch" }
func (a apiGatewayAdapter) DispatchEnabled() bool { return true }
func (a apiGatewayAdapter) RequiredMode() string  { return "single" }

func (a apiGatewayAdapter) Prepare(call GatewayCall, decision PolicyDecision) (GatewayCall, error) {
	call, err := recordGatewayAdapter{}.Prepare(call, decision)
	if err != nil {
		return call, err
	}
	call.Notes = appendUniqueString(call.Notes, "adapter:api")
	call.Notes = appendUniqueString(call.Notes, "dispatch:provider-api")
	return call, nil
}

func (a apiGatewayAdapter) Dispatch(call GatewayCall, decision PolicyDecision) (GatewayCall, error) {
	return a.DispatchWithPayload(call, decision, nil)
}

func (a apiGatewayAdapter) DispatchWithPayload(call GatewayCall, decision PolicyDecision, payload map[string]any) (GatewayCall, error) {
	endpoints := providerEndpointMapFromEnv()
	endpoint, ok := endpoints[strings.TrimSpace(call.Provider)]
	if !ok || strings.TrimSpace(endpoint) == "" {
		call.Status = "failed"
		call.AttemptCount = 0
		errText := "no_provider_endpoint"
		call.LastError = &errText
		call.Notes = appendUniqueString(call.Notes, "dispatch:error=no_provider_endpoint")
		return call, nil
	}
	if err := validateProviderEndpoint(endpoint); err != nil {
		call.Status = "failed"
		call.AttemptCount = 0
		errText := "provider_endpoint_rejected"
		call.LastError = &errText
		call.Notes = appendUniqueString(call.Notes, "dispatch:error=provider_endpoint_rejected")
		return call, nil
	}
	requestPayload := payload
	if requestPayload == nil {
		requestPayload = map[string]any{
			"call_id":      call.CallID,
			"decision_id":  call.DecisionID,
			"provider":     call.Provider,
			"model":        call.Model,
			"request_kind": call.RequestKind,
			"token_budget": call.TokenBudget,
			"notes":        call.Notes,
			"policy": map[string]any{
				"mode":        decision.Mode,
				"decision":    decision.Decision,
				"scope":       decision.Scope,
				"risk":        decision.Risk,
				"novelty":     decision.Novelty,
				"token_class": decision.TokenClass,
			},
		}
	}
	raw, err := json.Marshal(requestPayload)
	if err != nil {
		return call, err
	}
	client := a.httpClient
	if client == nil {
		client = &http.Client{Timeout: providerDispatchTimeout()}
	}
	retries := providerDispatchRetries()
	for attempt := 1; attempt <= retries+1; attempt++ {
		request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
		if err != nil {
			return call, err
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Layer-Call-ID", call.CallID)
		response, err := client.Do(request)
		call.AttemptCount = attempt
		now := zeroSafeNow()
		call.DispatchedAt = &now
		call.Notes = appendUniqueString(call.Notes, "dispatch:endpoint="+endpoint)
		call.Notes = appendUniqueString(call.Notes, "dispatch:attempt="+strconvItoa(attempt))
		if err != nil {
			errText := strings.TrimSpace(err.Error())
			if errText == "" {
				errText = "http_request_failed"
			}
			call.LastError = &errText
			call.Notes = appendUniqueString(call.Notes, "dispatch:error=http_request_failed")
			if attempt <= retries {
				call.Notes = appendUniqueString(call.Notes, "dispatch:retry=scheduled")
				continue
			}
			call.Status = "failed"
			return call, nil
		}
		body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
		_ = response.Body.Close()
		statusCode := response.StatusCode
		call.LastHTTPStatus = &statusCode
		preview := strings.TrimSpace(string(body))
		if preview != "" {
			if len(preview) > 280 {
				preview = preview[:280]
			}
			call.ResponsePreview = &preview
		}
		call.Notes = appendUniqueString(call.Notes, "dispatch:http_status="+strconvItoa(response.StatusCode))
		if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
			call.Status = "sent"
			call.LastError = nil
			return call, nil
		}
		statusErr := "http_status_" + strconvItoa(response.StatusCode)
		call.LastError = &statusErr
		if !shouldRetryProviderStatus(response.StatusCode) || attempt > retries {
			call.Status = "failed"
			return call, nil
		}
		call.Notes = appendUniqueString(call.Notes, "dispatch:retry=scheduled")
	}
	call.Status = "failed"
	return call, nil
}

type commandVerifyAdapter struct{}

func (commandVerifyAdapter) Name() string { return "command" }

func (commandVerifyAdapter) Run(command []string, workdir string) (string, error) {
	return executeVerificationCommand(command, workdir)
}

type commandDeployAdapter struct{}

func (commandDeployAdapter) Name() string { return "command" }

func (commandDeployAdapter) Run(target DeployTarget) (string, error) {
	return executeTargetCommand(target)
}

func strconvItoa(value int) string {
	return strconv.Itoa(value)
}

func providerDispatchRetries() int {
	raw := strings.TrimSpace(os.Getenv("LAYER_OS_PROVIDER_RETRIES"))
	if raw == "" {
		return 1
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 1
	}
	return value
}

func providerDispatchTimeout() time.Duration {
	raw := strings.TrimSpace(os.Getenv("LAYER_OS_PROVIDER_TIMEOUT_MS"))
	if raw == "" {
		return 10 * time.Second
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 10 * time.Second
	}
	return time.Duration(value) * time.Millisecond
}

func shouldRetryProviderStatus(status int) bool {
	if status == http.StatusTooManyRequests {
		return true
	}
	return status >= http.StatusInternalServerError
}
