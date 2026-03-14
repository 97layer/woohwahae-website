package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

func handleKnowledgeSearchRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeError(w, http.StatusBadRequest, "knowledge search query is required")
		return
	}
	writeJSON(w, http.StatusOK, service.SearchKnowledge(query))
}

func handleConversationRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{
			"conversation": service.ConversationAutomationStatus(),
		})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var input runtime.ConversationNoteInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid conversation payload")
			return
		}
		if actor := requestActor(r); actor != "" {
			input.Actor = actor
		}
		result, err := service.CreateConversationNote(input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, result)
	default:
		methodNotAllowed(w)
	}
}

func handleObservationsRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		query := runtime.ObservationQuery{
			SourceChannel: r.URL.Query().Get("source_channel"),
			Topic:         r.URL.Query().Get("topic"),
			Actor:         r.URL.Query().Get("actor"),
			Ref:           r.URL.Query().Get("ref"),
			Text:          r.URL.Query().Get("text"),
		}
		if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
			limit, err := strconv.Atoi(raw)
			if err != nil || limit < 0 {
				writeError(w, http.StatusBadRequest, "invalid observations limit")
				return
			}
			query.Limit = limit
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"items": service.ListObservations(query),
		})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var item runtime.ObservationRecord
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid observation payload")
			return
		}
		if actor := requestActor(r); actor != "" {
			item.Actor = actor
		}
		created, err := service.CreateObservation(item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		methodNotAllowed(w)
	}
}

func handleSourceDraftPrepRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !requireWriteAuth(service, w, r) {
		return
	}
	var payload struct {
		DraftObservationID string `json:"draft_observation_id"`
		Channel            string `json:"channel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid source draft prep payload")
		return
	}
	result, err := service.OpenSourceDraftSeedPublishPrep(payload.DraftObservationID, payload.Channel)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func handleSourceIntakeRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, buildSourceIntakeView(service))
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var payload struct {
			SourceChannel   string    `json:"source_channel"`
			Actor           string    `json:"actor"`
			Confidence      string    `json:"confidence"`
			CapturedAt      time.Time `json:"captured_at"`
			IntakeClass     string    `json:"intake_class"`
			PolicyColor     string    `json:"policy_color"`
			Title           string    `json:"title"`
			URL             string    `json:"url"`
			Excerpt         string    `json:"excerpt"`
			FounderNote     string    `json:"founder_note"`
			DomainTags      []string  `json:"domain_tags"`
			WorldviewTags   []string  `json:"worldview_tags"`
			SuggestedRoutes []string  `json:"suggested_routes"`
			Refs            []string  `json:"refs"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid source intake payload")
			return
		}
		observation, _, err := runtime.BuildSourceIntakeObservation(runtime.SourceIntakeObservationInput{
			SourceChannel: payload.SourceChannel,
			Actor:         nonEmptyString(requestActor(r), payload.Actor),
			Confidence:    payload.Confidence,
			Refs:          payload.Refs,
			CapturedAt:    payload.CapturedAt,
			Record: runtime.SourceIntakeRecord{
				IntakeClass:     payload.IntakeClass,
				PolicyColor:     payload.PolicyColor,
				Title:           payload.Title,
				URL:             payload.URL,
				Excerpt:         payload.Excerpt,
				FounderNote:     payload.FounderNote,
				DomainTags:      payload.DomainTags,
				WorldviewTags:   payload.WorldviewTags,
				SuggestedRoutes: payload.SuggestedRoutes,
			},
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		created, err := service.CreateObservation(observation)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, buildSourceIntakeCreateResult(service, created))
	default:
		methodNotAllowed(w)
	}
}

func handleOpenThreadsRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	limit := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			writeError(w, http.StatusBadRequest, "invalid open threads limit")
			return
		}
		limit = value
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": service.OpenThreads(limit),
	})
}

func handleThreadsRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		limit := 0
		if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil || value < 0 {
				writeError(w, http.StatusBadRequest, "invalid threads limit")
				return
			}
			limit = value
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": service.Threads(limit)})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var item runtime.OpenThread
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid thread payload")
			return
		}
		saved, err := service.SaveOpenThread(item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, saved)
	default:
		methodNotAllowed(w)
	}
}

func handleReviewRoomRoute(service *runtime.Service, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{
			"review_room": service.ReviewRoom(),
			"summary":     service.ReviewRoomSummary(),
		})
	case http.MethodPost:
		if !requireWriteAuth(service, w, r) {
			return
		}
		var payload struct {
			Action         string   `json:"action"`
			Section        string   `json:"section"`
			Item           string   `json:"item"`
			Kind           string   `json:"kind"`
			Severity       string   `json:"severity"`
			Source         string   `json:"source"`
			Ref            string   `json:"ref"`
			Why            string   `json:"why"`
			WhyUnresolved  string   `json:"why_unresolved"`
			Contradiction  string   `json:"contradiction"`
			Contradictions []string `json:"contradictions"`
			PatternRefs    []string `json:"pattern_refs"`
			Reason         string   `json:"reason"`
			Rule           string   `json:"rule"`
			Evidence       []string `json:"evidence"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid review room payload")
			return
		}
		var (
			room   runtime.ReviewRoom
			err    error
			status = http.StatusCreated
		)
		if strings.TrimSpace(payload.Action) != "" {
			room, err = service.TransitionStructuredReviewRoomItem(payload.Action, payload.Item, &runtime.ReviewRoomResolution{
				Action:   payload.Action,
				Reason:   payload.Reason,
				Rule:     payload.Rule,
				Evidence: payload.Evidence,
			})
			status = http.StatusOK
		} else {
			var ref *string
			if strings.TrimSpace(payload.Ref) != "" {
				value := strings.TrimSpace(payload.Ref)
				ref = &value
			}
			room, err = service.AddStructuredReviewRoomItem(payload.Section, runtime.ReviewRoomItem{
				Text:           payload.Item,
				Kind:           payload.Kind,
				Severity:       payload.Severity,
				Source:         payload.Source,
				Ref:            ref,
				Why:            strings.TrimSpace(payload.Why),
				WhyUnresolved:  stringRef(payload.WhyUnresolved),
				Contradiction:  stringRef(payload.Contradiction),
				Contradictions: append(append([]string{}, payload.Contradictions...), strings.TrimSpace(payload.Contradiction)),
				PatternRefs:    append([]string{}, payload.PatternRefs...),
			})
		}
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, status, map[string]any{
			"review_room": room,
			"summary":     runtime.SummarizeReviewRoom(room),
		})
	default:
		methodNotAllowed(w)
	}
}
