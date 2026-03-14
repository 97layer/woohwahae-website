package api

import (
	"encoding/json"
	"net/http"
	"time"

	"layer-os/internal/runtime"
)

func handleFlowsRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"items": service.ListFlows()})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var item runtime.FlowRun
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid flow payload")
			return
		}
		if item.Status == "" {
			item.Status = "active"
		}
		if item.Notes == nil {
			item.Notes = []string{}
		}
		if item.CreatedAt.IsZero() {
			item.CreatedAt = time.Now().UTC()
		}
		if item.UpdatedAt.IsZero() {
			item.UpdatedAt = item.CreatedAt
		}
		if err := service.CreateFlow(item); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		methodNotAllowed(w)
	}
}

func handleFlowSyncRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		FlowID           string   `json:"flow_id"`
		WorkItemID       string   `json:"work_item_id"`
		ApprovalID       string   `json:"approval_id"`
		PolicyDecisionID string   `json:"policy_decision_id"`
		ExecuteID        string   `json:"execute_id"`
		VerificationID   string   `json:"verification_id"`
		ReleaseID        string   `json:"release_id"`
		DeployID         string   `json:"deploy_id"`
		RollbackID       string   `json:"rollback_id"`
		Notes            []string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid flow sync payload")
		return
	}
	item, err := service.SyncFlow(
		payload.FlowID,
		payload.WorkItemID,
		stringRef(payload.ApprovalID),
		stringRef(payload.PolicyDecisionID),
		stringRef(payload.ExecuteID),
		stringRef(payload.VerificationID),
		stringRef(payload.ReleaseID),
		stringRef(payload.DeployID),
		stringRef(payload.RollbackID),
		payload.Notes,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func handleAuthRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{
			"auth": service.AuthStatus(),
		})
	case http.MethodPost:
		if !requireAuthBootstrap(service, w, r) {
			return
		}
		var payload struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid auth payload")
			return
		}
		status, err := service.SetWriteToken(payload.Token)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"auth": status})
	case http.MethodDelete:
		if !requireAuthBootstrap(service, w, r) {
			return
		}
		status, err := service.ClearWriteToken()
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"auth": status})
	default:
		methodNotAllowed(w)
	}
}

func handlePreflightsRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"items": service.ListPreflights()})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var item runtime.PreflightRecord
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid preflight payload")
			return
		}
		if item.GeneratedAt.IsZero() {
			item.GeneratedAt = time.Now().UTC()
		}
		if item.ModelsUsed == nil {
			item.ModelsUsed = []string{}
		}
		item.ModelsUsed = mergeModels(item.ModelsUsed, requestModels(r))
		if item.Steps == nil {
			item.Steps = []string{}
		}
		if item.Risks == nil {
			item.Risks = []string{}
		}
		if item.Checks == nil {
			item.Checks = []string{}
		}
		if err := service.CreatePreflight(item); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		methodNotAllowed(w)
	}
}

func handlePoliciesRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": service.ListPolicies()})
}

func handlePolicyEvaluateRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		DecisionID       string `json:"decision_id"`
		Intent           string `json:"intent"`
		Scope            string `json:"scope"`
		Risk             string `json:"risk"`
		Novelty          string `json:"novelty"`
		TokenClass       string `json:"token_class"`
		RequiresApproval bool   `json:"requires_approval"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid policy payload")
		return
	}
	item, err := service.EvaluatePolicy(payload.DecisionID, payload.Intent, payload.Scope, payload.Risk, payload.Novelty, payload.TokenClass, payload.RequiresApproval)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func handleGatewayCallsRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"items": service.ListGatewayCalls()})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var item runtime.GatewayCall
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid gateway payload")
			return
		}
		if item.Status == "" {
			item.Status = "recorded"
		}
		if item.Notes == nil {
			item.Notes = []string{}
		}
		if item.CreatedAt.IsZero() {
			item.CreatedAt = time.Now().UTC()
		}
		if err := service.CreateGatewayCall(item); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		methodNotAllowed(w)
	}
}

func handleExecuteRunsRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": service.ListExecutes()})
}

func handleExecuteRunRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		ExecuteID        string   `json:"execute_id"`
		WorkItemID       string   `json:"work_item_id"`
		PolicyDecisionID string   `json:"policy_decision_id"`
		Notes            []string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid execute payload")
		return
	}
	item, err := service.RunExecute(payload.ExecuteID, payload.WorkItemID, payload.PolicyDecisionID, payload.Notes)
	if err != nil {
		if item.ExecuteID != "" {
			writeJSON(w, http.StatusBadGateway, map[string]any{
				"execute": item,
				"error":   err.Error(),
			})
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func handleMemoryRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"memory": service.Memory()})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var memory runtime.SystemMemory
		if err := json.NewDecoder(r.Body).Decode(&memory); err != nil {
			writeError(w, http.StatusBadRequest, "invalid system memory payload")
			return
		}
		if memory.UpdatedAt.IsZero() {
			memory.UpdatedAt = time.Now().UTC()
		}
		if memory.NextSteps == nil {
			memory.NextSteps = []string{}
		}
		if memory.OpenRisks == nil {
			memory.OpenRisks = []string{}
		}
		if err := service.ReplaceMemory(memory); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, memory)
	default:
		methodNotAllowed(w)
	}
}
