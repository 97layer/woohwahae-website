package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"layer-os/internal/runtime"
)

func TestObservationsRouteCreateAndQuery(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)
	raw := []byte(`{"source_channel":"terminal","topic":"queue drift","refs":["job_001"],"normalized_summary":"Terminal queue drift still needs triage."}`)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/observations", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var created runtime.ObservationRecord
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created observation: %v", err)
	}
	if created.ObservationID == "" || created.RawExcerpt == "" {
		t.Fatalf("expected service defaults in created observation, got %+v", created)
	}
	listReq := httptest.NewRequest(http.MethodGet, "/api/layer-os/observations?source_channel=terminal&text=triage", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", listRec.Code, listRec.Body.String())
	}
	var response struct {
		Items []runtime.ObservationRecord `json:"items"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode observations response: %v", err)
	}
	if len(response.Items) != 1 || response.Items[0].ObservationID != created.ObservationID {
		t.Fatalf("unexpected observations response: %+v", response.Items)
	}
}

func TestObservationsRouteRejectsInvalidLimit(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)
	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/observations?limit=-1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestObservationsRouteUsesHeaderActorOverPayloadActor(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)
	raw := []byte(`{"source_channel":"terminal","topic":"queue drift","actor":"founder","refs":["job_001"],"normalized_summary":"Terminal queue drift still needs triage."}`)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/observations", bytes.NewReader(raw))
	req.Header.Set("X-Layer-Actor", "gemini")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var created runtime.ObservationRecord
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created observation: %v", err)
	}
	if created.Actor != "gemini" {
		t.Fatalf("expected header actor gemini to win, got %+v", created)
	}
}

func TestConversationRouteStatusAndNote(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	statusReq := httptest.NewRequest(http.MethodGet, "/api/layer-os/conversation", nil)
	statusRec := httptest.NewRecorder()
	router.ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", statusRec.Code, statusRec.Body.String())
	}

	raw := []byte(`{"text":"Need founder review on this thread","source_channel":"telegram","tags":["todo"],"refs":["thread:001"],"confidence":"high"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/conversation", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var created runtime.ConversationNoteResult
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode conversation result: %v", err)
	}
	if created.ConversationID == "" || created.Observation.SourceChannel != "conversation:telegram" {
		t.Fatalf("unexpected conversation result: %+v", created)
	}
	if created.Proposal == nil || created.Job == nil {
		t.Fatalf("expected auto plan artifacts, got %+v", created)
	}
}

func TestConversationRouteRejectsInvalidPayload(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)
	raw := []byte(`{"source_channel":"telegram"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/conversation", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestConversationRouteRequiresBearerWhenEnabled(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.SetWriteToken("secret-token"); err != nil {
		t.Fatalf("set write token: %v", err)
	}
	router := NewRouter(service)
	raw := []byte(`{"text":"Need founder review on this thread","source_channel":"telegram","tags":["todo"],"refs":["thread:001"],"confidence":"high"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/conversation", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/layer-os/conversation", bytes.NewReader(raw))
	req.Header.Set("Authorization", "Bearer secret-token")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 with valid token, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestConversationRouteUsesHeaderActorOverPayloadActor(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)
	raw := []byte(`{"text":"Need founder review on this thread","source_channel":"telegram","actor":"founder","tags":["todo"],"refs":["thread:001"],"confidence":"high"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/conversation", bytes.NewReader(raw))
	req.Header.Set("X-Layer-Actor", "gemini")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var created runtime.ConversationNoteResult
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode conversation result: %v", err)
	}
	if created.Profile.Actor != "gemini" || created.Observation.Actor != "gemini" {
		t.Fatalf("expected header actor gemini to win, got %+v", created)
	}
}

func TestSourceDraftPrepRouteOpensCanonicalThreadsLane(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	source, err := service.CreateObservation(runtime.ObservationRecord{
		ObservationID:     "observation_source_400",
		SourceChannel:     "telegram",
		Actor:             "founder",
		Topic:             runtime.SourceIntakeTopic,
		Confidence:        "high",
		Refs:              []string{"route:97layer"},
		RawExcerpt:        runtime.BuildSourceIntakeRawExcerpt(runtime.SourceIntakeRecord{Title: "operating surface rebuild", Excerpt: "홈페이지와 운영면을 나누면서 기능보다 구조가 먼저 남았다.", FounderNote: "브랜드 구축 과정으로", DomainTags: []string{"system", "brand"}, WorldviewTags: []string{"identity"}, SuggestedRoutes: []string{"97layer"}}),
		NormalizedSummary: "Source intake · operating surface rebuild -> 97layer",
	})
	if err != nil {
		t.Fatalf("create source observation: %v", err)
	}
	draft, _, err := runtime.EnsureSourceDraftSeedObservation(service, source, "observation_route_400", "97layer")
	if err != nil {
		t.Fatalf("seed draft observation: %v", err)
	}

	router := NewRouter(service)
	raw := []byte(`{"draft_observation_id":"` + draft.ObservationID + `","channel":"threads"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/source-intake/drafts/prep", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var result runtime.SourceDraftPublishPrepResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode prep result: %v", err)
	}
	if result.Lane.Channel != "threads" || result.Approval.ApprovalID == "" || result.Observation.Topic != "brand_publish_prep" {
		t.Fatalf("unexpected prep route result: %+v", result)
	}
}

func TestSourceDraftPrepRouteRequiresBearerWhenEnabled(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.SetWriteToken("secret-token"); err != nil {
		t.Fatalf("set write token: %v", err)
	}
	router := NewRouter(service)
	raw := []byte(`{"draft_observation_id":"observation_123","channel":"threads"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/source-intake/drafts/prep", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSourceIntakeRouteCreatesCanonicalObservation(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)
	raw := []byte(`{"source_channel":"cockpit","actor":"founder","confidence":"high","intake_class":"public_collector","policy_color":"yellow","title":"ambiguous feed note","excerpt":"A public signal arrived without a clear publishing lane.","domain_tags":["beauty"],"worldview_tags":["subtraction"]}`)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/source-intake", bytes.NewReader(raw))
	req.Header.Set("X-Layer-Actor", "gemini")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Created struct {
			ObservationID string `json:"observationId"`
			PolicyColor   string `json:"policyColor"`
			PriorityScore int    `json:"priorityScore"`
			Disposition   string `json:"disposition"`
		} `json:"created"`
		NextAction struct {
			Mode                string `json:"mode"`
			Summary             string `json:"summary"`
			SourceObservationID string `json:"sourceObservationId"`
			Ref                 string `json:"ref"`
		} `json:"next_action"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode source intake create result: %v", err)
	}
	if payload.Created.ObservationID == "" || payload.NextAction.Summary == "" {
		t.Fatalf("expected canonical source intake create result, got %+v", payload)
	}
	if payload.Created.PriorityScore <= 0 {
		t.Fatalf("expected computed priority score, got %+v", payload)
	}
	if payload.Created.Disposition != "review" || payload.Created.PolicyColor != "yellow" {
		t.Fatalf("expected yellow public collector source to require review, got %+v", payload)
	}
	if payload.NextAction.SourceObservationID != payload.Created.ObservationID {
		t.Fatalf("expected next action to point at created observation, got %+v", payload)
	}
	if payload.NextAction.Ref != "observation:"+payload.Created.ObservationID {
		t.Fatalf("expected next action ref to point at created observation, got %+v", payload)
	}

	items := service.ListObservations(runtime.ObservationQuery{Topic: runtime.SourceIntakeTopic, Limit: 1})
	if len(items) != 1 || items[0].Actor != "gemini" {
		t.Fatalf("expected created observation actor gemini, got %+v", items)
	}
}

func TestSourceIntakeRouteRequiresBearerWhenEnabled(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.SetWriteToken("secret-token"); err != nil {
		t.Fatalf("set write token: %v", err)
	}
	router := NewRouter(service)
	raw := []byte(`{"source_channel":"cockpit","confidence":"high","intake_class":"manual_drop","policy_color":"green","title":"manual note","excerpt":"Founders keep dropping brand fragments here.","suggested_routes":["97layer"]}`)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/source-intake", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/layer-os/source-intake", bytes.NewReader(raw))
	req.Header.Set("Authorization", "Bearer secret-token")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 with valid token, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSourceIntakeRouteReturnsCanonicalView(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.CreateObservation(runtime.ObservationRecord{
		ObservationID:     "observation_source_view_001",
		SourceChannel:     "feed",
		Actor:             "sensor",
		Topic:             runtime.SourceIntakeTopic,
		Confidence:        "high",
		Refs:              []string{"feed_source:https://example.com/feed.xml", "route:97layer"},
		RawExcerpt:        runtime.BuildSourceIntakeRawExcerpt(runtime.FinalizeSourceIntakeRecord(runtime.SourceIntakeRecord{IntakeClass: "public_collector", PolicyColor: "green", Title: "sensor note", Excerpt: "Signal strong enough for prep.", SuggestedRoutes: []string{"97layer"}})),
		NormalizedSummary: "Source intake · sensor note -> 97layer",
	}); err != nil {
		t.Fatalf("create source observation: %v", err)
	}
	router := NewRouter(service)
	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/source-intake", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		RuntimeAvailable bool `json:"runtimeAvailable"`
		ActionCount      int  `json:"actionCount"`
		Items            []struct {
			ObservationID string `json:"observationId"`
			Disposition   string `json:"disposition"`
		} `json:"items"`
		Attention struct {
			Summary string `json:"summary"`
		} `json:"attention"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode source intake view: %v", err)
	}
	if !payload.RuntimeAvailable || payload.ActionCount == 0 || len(payload.Items) == 0 {
		t.Fatalf("expected populated source intake view, got %+v", payload)
	}
	if payload.Items[0].ObservationID == "" || payload.Attention.Summary == "" {
		t.Fatalf("expected canonical source intake item and attention, got %+v", payload)
	}
}
