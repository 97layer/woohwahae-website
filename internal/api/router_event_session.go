package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

func handleEvents(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"items": service.ListEvents()})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var input runtime.EventCreateInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid event payload")
			return
		}
		if strings.TrimSpace(input.WorkItemID) == "" {
			input.WorkItemID = "system"
		}
		if input.Surface == "" {
			input.Surface = runtime.SurfaceAPI
		}
		if input.Stage == "" {
			input.Stage = runtime.StageDiscover
		}
		if input.Data == nil {
			input.Data = map[string]any{}
		}
		item, err := service.CreateEvent(input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		methodNotAllowed(w)
	}
}

func handleEventStream(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	lastID := strings.TrimSpace(r.Header.Get("Last-Event-ID"))
	sent := seedSeenEvents(service.ListEvents(), lastID)
	_, _ = fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	ticker := time.NewTicker(750 * time.Millisecond)
	defer ticker.Stop()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			for _, item := range service.ListEvents() {
				if sent[item.EventID] {
					continue
				}
				raw, err := json.Marshal(item)
				if err != nil {
					continue
				}
				_, _ = fmt.Fprintf(w, "id: %s\n", item.EventID)
				_, _ = fmt.Fprintf(w, "event: %s\n", item.Kind)
				_, _ = fmt.Fprintf(w, "data: %s\n\n", raw)
				sent[item.EventID] = true
				flusher.Flush()
			}
		case <-heartbeat.C:
			_, _ = fmt.Fprintf(w, ": ping %d\n\n", time.Now().UTC().Unix())
			flusher.Flush()
		}
	}
}

func seedSeenEvents(items []runtime.EventEnvelope, lastID string) map[string]bool {
	seen := map[string]bool{}
	if strings.TrimSpace(lastID) == "" {
		return seen
	}
	matched := false
	for _, item := range items {
		seen[item.EventID] = true
		if item.EventID == lastID {
			matched = true
			break
		}
	}
	if matched {
		return seen
	}
	return map[string]bool{}
}

func handleSessionBootstrap(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	full := false
	if value := strings.TrimSpace(r.URL.Query().Get("full")); value == "1" || strings.EqualFold(value, "true") {
		full = true
	}
	packet := service.SessionBootstrap("daemon", false, full)
	writeJSON(w, http.StatusOK, map[string]any{"session": packet})
}

func handleSessionCheckpoint(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var input runtime.SessionCheckpointInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid session checkpoint payload")
		return
	}
	if input.NextSteps == nil {
		input.NextSteps = []string{}
	}
	if input.OpenRisks == nil {
		input.OpenRisks = []string{}
	}
	if input.Refs == nil {
		input.Refs = []string{}
	}
	item, err := service.CheckpointSession(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func handleSessionNote(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var input runtime.SessionNoteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid session note payload")
		return
	}
	if input.Refs == nil {
		input.Refs = []string{}
	}
	result, err := service.SessionNote(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func handleSessionFinish(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var input runtime.SessionFinishInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid session finish payload")
		return
	}
	if input.NextSteps == nil {
		input.NextSteps = []string{}
	}
	if input.OpenRisks == nil {
		input.OpenRisks = []string{}
	}
	result, err := service.FinishSession(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, result)
}
