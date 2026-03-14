package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

func handleVerificationsRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": service.ListVerifications()})
}

func handleVerificationRunRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		RecordID string   `json:"record_id"`
		Scope    string   `json:"scope"`
		Notes    []string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid verification payload")
		return
	}
	command, err := runtime.DefaultGoTestVerificationCommand()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	item, err := service.RunVerification(payload.RecordID, payload.Scope, repoRoot(), command, payload.Notes)
	if err != nil {
		if item.RecordID != "" {
			writeJSON(w, http.StatusBadGateway, map[string]any{
				"verification": item,
				"error":        err.Error(),
			})
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func handleBranchesRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"items": service.ListBranches()})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var item runtime.Branch
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid branch payload")
			return
		}
		if item.Summary == "" {
			item.Summary = item.Title
		}
		if item.Stage == "" {
			item.Stage = runtime.StageCompose
		}
		if item.Surface == "" {
			item.Surface = runtime.SurfaceAPI
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
		if err := service.CreateBranch(item); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		methodNotAllowed(w)
	}
}

func handleBranchMergeRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		BranchID       string   `json:"branch_id"`
		TargetBranchID string   `json:"target_branch_id"`
		Notes          []string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid branch merge payload")
		return
	}
	item, err := service.MergeBranch(payload.BranchID, payload.TargetBranchID, payload.Notes)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func handleJobProfilesRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": service.AgentDispatchProfiles()})
}

func handleJobPacketRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	packet, err := service.AgentRunPacket(r.URL.Query().Get("job_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, packet)
}

func handleJobsRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items := service.ListAgentJobs()
		status := strings.TrimSpace(r.URL.Query().Get("status"))
		if status != "" {
			filtered := make([]runtime.AgentJob, 0, len(items))
			for _, item := range items {
				if status == "open" {
					if item.Status != "queued" && item.Status != "running" {
						continue
					}
				} else if item.Status != status {
					continue
				}
				filtered = append(filtered, item)
			}
			items = filtered
		}
		if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
			limit, err := strconv.Atoi(rawLimit)
			if err != nil || limit < 1 {
				writeError(w, http.StatusBadRequest, "invalid jobs limit")
				return
			}
			if len(items) > limit {
				items = items[:limit]
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var item runtime.AgentJob
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid agent job payload")
			return
		}
		if err := service.CreateAgentJob(item); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		methodNotAllowed(w)
	}
}

func handleJobPromoteRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		Limit    int  `json:"limit"`
		Dispatch bool `json:"dispatch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid agent job promote payload")
		return
	}
	result, err := service.PromoteContextJobs(payload.Limit, payload.Dispatch)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func handleJobDispatchRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		JobID string `json:"job_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid agent job dispatch payload")
		return
	}
	result, err := service.DispatchAgentJob(payload.JobID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func handleJobUpdateRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		JobID  string         `json:"job_id"`
		Status string         `json:"status"`
		Notes  []string       `json:"notes"`
		Result map[string]any `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid agent job update payload")
		return
	}
	item, err := service.UpdateAgentJob(payload.JobID, payload.Status, payload.Notes, payload.Result)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func handleJobReportRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		JobID  string         `json:"job_id"`
		Status string         `json:"status"`
		Notes  []string       `json:"notes"`
		Result map[string]any `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid agent job report payload")
		return
	}
	item, err := service.ReportAgentJob(payload.JobID, payload.Status, payload.Notes, payload.Result)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}
