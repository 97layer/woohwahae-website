package api

import (
	"encoding/json"
	"net/http"
	"time"

	"layer-os/internal/runtime"
)

func handleProposalsRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"items": service.ListProposals()})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var item runtime.ProposalItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid proposal payload")
			return
		}
		if item.Summary == "" {
			item.Summary = item.Title
		}
		if item.Status == "" {
			item.Status = "proposed"
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
		if err := service.CreateProposal(item); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		methodNotAllowed(w)
	}
}

func handleProposalPromoteRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		ProposalID string `json:"proposal_id"`
		WorkItemID string `json:"work_item_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid proposal promote payload")
		return
	}
	proposal, work, err := service.PromoteProposal(payload.ProposalID, payload.WorkItemID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"proposal": proposal, "work_item": work})
}

func handleWorkItemsRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"items": service.ListWorkItems()})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var item runtime.WorkItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid work item payload")
			return
		}
		if item.CreatedAt.IsZero() {
			item.CreatedAt = time.Now().UTC()
		}
		if item.Payload == nil {
			item.Payload = map[string]any{}
		}
		if err := service.CreateWorkItem(item); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		methodNotAllowed(w)
	}
}

func handleApprovalInboxRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"items": service.ListApprovals()})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var item runtime.ApprovalItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid approval payload")
			return
		}
		if item.RequestedAt.IsZero() {
			item.RequestedAt = time.Now().UTC()
		}
		if err := service.CreateApproval(item); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		methodNotAllowed(w)
	}
}

func handleApprovalResolveRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		ApprovalID string `json:"approval_id"`
		Status     string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid approval resolve payload")
		return
	}
	item, err := service.ResolveApproval(payload.ApprovalID, payload.Status)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}
