package api

import (
	"encoding/json"
	"net/http"

	"layer-os/internal/runtime"
)

func handleFounderStartRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		FlowID     string   `json:"flow_id"`
		WorkItemID string   `json:"work_item_id"`
		ApprovalID string   `json:"approval_id"`
		Title      string   `json:"title"`
		Intent     string   `json:"intent"`
		Notes      []string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid founder start payload")
		return
	}
	item, err := service.StartFounderFlow(payload.FlowID, payload.WorkItemID, payload.ApprovalID, payload.Title, payload.Intent, payload.Notes)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func handleFounderApproveRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		FlowID string   `json:"flow_id"`
		Notes  []string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid founder approve payload")
		return
	}
	item, err := service.ApproveFounderFlow(payload.FlowID, payload.Notes)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func handleFounderReleaseRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		FlowID    string   `json:"flow_id"`
		ReleaseID string   `json:"release_id"`
		DeployID  string   `json:"deploy_id"`
		Target    string   `json:"target"`
		Channel   string   `json:"channel"`
		Notes     []string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid founder release payload")
		return
	}
	item, err := service.ReleaseFounderFlow(payload.FlowID, payload.ReleaseID, payload.DeployID, payload.Target, payload.Channel, payload.Notes)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func handleFounderRollbackRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		FlowID     string   `json:"flow_id"`
		RollbackID string   `json:"rollback_id"`
		Notes      []string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid founder rollback payload")
		return
	}
	item, err := service.RollbackFounderFlow(payload.FlowID, payload.RollbackID, payload.Notes)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}
