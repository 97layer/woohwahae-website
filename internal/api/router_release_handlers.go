package api

import (
	"encoding/json"
	"net/http"
	"time"

	"layer-os/internal/runtime"
)

func handleReleasesRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"items": service.ListReleases()})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var item runtime.ReleasePacket
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid release payload")
			return
		}
		if item.ReleasedAt == nil || item.ReleasedAt.IsZero() {
			releasedAt := time.Now().UTC()
			item.ReleasedAt = &releasedAt
		}
		if item.Artifacts == nil {
			item.Artifacts = []string{}
		}
		if item.Metrics == nil {
			item.Metrics = map[string]any{}
		}
		if item.ApprovalRefs == nil {
			item.ApprovalRefs = []string{}
		}
		if err := service.CreateRelease(item); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		methodNotAllowed(w)
	}
}

func handleDeploysRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"items": service.ListDeploys()})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var item runtime.DeployRun
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid deploy payload")
			return
		}
		if item.StartedAt.IsZero() {
			item.StartedAt = time.Now().UTC()
		}
		if item.FinishedAt == nil || item.FinishedAt.IsZero() {
			finishedAt := time.Now().UTC()
			item.FinishedAt = &finishedAt
		}
		if item.Notes == nil {
			item.Notes = []string{}
		}
		if err := service.CreateDeploy(item); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		methodNotAllowed(w)
	}
}

func handleDeployExecuteRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		DeployID  string   `json:"deploy_id"`
		ReleaseID string   `json:"release_id"`
		Notes     []string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid deploy execute payload")
		return
	}
	item, err := service.ExecuteDeploy(payload.DeployID, payload.ReleaseID, payload.Notes)
	if err != nil {
		if item.DeployID != "" {
			writeJSON(w, http.StatusBadGateway, map[string]any{
				"deploy": item,
				"error":  err.Error(),
			})
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func handleRollbacksRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"items": service.ListRollbacks()})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var item runtime.RollbackRun
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid rollback payload")
			return
		}
		if item.Status == "" {
			item.Status = "recorded"
		}
		if item.Notes == nil {
			item.Notes = []string{}
		}
		if item.StartedAt.IsZero() {
			item.StartedAt = time.Now().UTC()
		}
		if err := service.CreateRollback(item); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		methodNotAllowed(w)
	}
}

func handleRollbackExecuteRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		RollbackID string   `json:"rollback_id"`
		ReleaseID  string   `json:"release_id"`
		DeployID   string   `json:"deploy_id"`
		Notes      []string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid rollback execute payload")
		return
	}
	item, err := service.ExecuteRollback(payload.RollbackID, payload.ReleaseID, payload.DeployID, payload.Notes)
	if err != nil {
		if item.RollbackID != "" {
			writeJSON(w, http.StatusBadGateway, map[string]any{
				"rollback": item,
				"error":    err.Error(),
			})
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func handleDeployTargetsRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"items": service.ListTargets()})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var item runtime.DeployTarget
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid deploy target payload")
			return
		}
		if err := sanitizeDeployTargetPaths(&item, repoRoot()); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := service.PutTarget(item); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		methodNotAllowed(w)
	}
}
