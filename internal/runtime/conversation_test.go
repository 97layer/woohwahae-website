package runtime

import (
	"strings"
	"testing"
	"time"
)

func TestCreateConversationNoteCreatesMaskedAutomationArtifacts(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	result, err := service.CreateConversationNote(ConversationNoteInput{
		Text:          "Need to wire deploy retry today. founder@example.com token: abc123",
		SourceChannel: "api",
		Tags:          []string{"todo"},
		Refs:          []string{"thread:founder-001"},
		Confidence:    "high",
	})
	if err != nil {
		t.Fatalf("create conversation note: %v", err)
	}

	if result.Event.Kind != conversationEventKind {
		t.Fatalf("expected %q event, got %+v", conversationEventKind, result.Event)
	}
	if strings.Contains(result.Observation.RawExcerpt, "founder@example.com") || strings.Contains(result.Observation.RawExcerpt, "abc123") {
		t.Fatalf("expected masked excerpt, got %q", result.Observation.RawExcerpt)
	}
	if !strings.Contains(result.Observation.RawExcerpt, "[redacted-email]") || !strings.Contains(result.Observation.RawExcerpt, "[redacted-secret]") {
		t.Fatalf("expected redaction markers, got %q", result.Observation.RawExcerpt)
	}
	if result.Proposal == nil || result.Job == nil {
		t.Fatalf("expected proposal and planner job, got %+v", result)
	}
	if result.Job.Result != nil {
		t.Fatalf("expected queued planner job before dispatch, got %+v", result.Job.Result)
	}
	roleProfile, ok := result.Job.Payload["role_profile"].(map[string]any)
	if !ok {
		t.Fatalf("expected JSON-safe role_profile map, got %#v", result.Job.Payload["role_profile"])
	}
	if got := strings.TrimSpace(roleProfile["role"].(string)); got != "planner" {
		t.Fatalf("expected planner role profile, got %+v", roleProfile)
	}
	if got := strings.TrimSpace(result.Job.Payload["instructions"].(string)); got == "" || !strings.Contains(got, "Conversation signal:") {
		t.Fatalf("expected planner instructions, got %#v", result.Job.Payload["instructions"])
	}
	if len(service.ListObservations(ObservationQuery{SourceChannel: "conversation:api"})) != 1 {
		t.Fatalf("expected one persisted conversation observation")
	}
	if len(service.ListProposals()) != 1 || len(service.ListAgentJobs()) != 1 {
		t.Fatalf("expected one proposal and one job, got proposals=%d jobs=%d", len(service.ListProposals()), len(service.ListAgentJobs()))
	}
}

func TestCreateConversationNoteDedupesRiskItemsWithoutPlannerSpam(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	input := ConversationNoteInput{
		Text:          "Critical security risk: same login failure keeps repeating.",
		SourceChannel: "api",
		Tags:          []string{"risk", "critique"},
		Refs:          []string{"incident:001"},
		Confidence:    "high",
		CapturedAt:    ptrTime(time.Date(2026, 3, 9, 12, 0, 0, 0, time.UTC)),
	}
	first, err := service.CreateConversationNote(input)
	if err != nil {
		t.Fatalf("first conversation note: %v", err)
	}
	second, err := service.CreateConversationNote(input)
	if err != nil {
		t.Fatalf("second conversation note: %v", err)
	}

	if first.ReviewItem == nil || second.ReviewItem == nil {
		t.Fatalf("expected review-room risk items, got first=%+v second=%+v", first.ReviewItem, second.ReviewItem)
	}
	if got := len(service.ListProposals()); got != 0 {
		t.Fatalf("expected zero proposals for passive risk capture, got %d", got)
	}
	if got := len(service.ListAgentJobs()); got != 0 {
		t.Fatalf("expected zero planner jobs for passive risk capture, got %d", got)
	}
	room := service.ReviewRoom()
	if len(room.Open) != 1 {
		t.Fatalf("expected one open review item, got %+v", room.Open)
	}
	if room.Open[0].Severity != "critical" {
		t.Fatalf("expected repeated risk to escalate to critical, got %+v", room.Open[0])
	}
	dedupeKey := conversationDedupeKey(
		first.Profile.Actor,
		input.SourceChannel,
		first.Observation.NormalizedSummary,
		conversationRefs(input.Refs, input.SourceChannel, classifyConversationTags(input.Tags, input.Text, first.Observation.NormalizedSummary)),
	)
	if !reviewRoomItemHasEvidence(room.Open[0], conversationDedupeNotePrefix+dedupeKey) {
		t.Fatalf("expected dedupe evidence on review item, got %+v", room.Open[0].Evidence)
	}
}

func TestCreateConversationNoteDoesNotAutoPlanFromRefsAlone(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	result, err := service.CreateConversationNote(ConversationNoteInput{
		Text:          "Notebook capture about retrieval architecture.",
		SourceChannel: "notebook_lm",
		Refs:          []string{"content_doc:note_001", "content_host:example_com"},
		Confidence:    "high",
	})
	if err != nil {
		t.Fatalf("create conversation note: %v", err)
	}

	if result.Proposal != nil || result.Job != nil {
		t.Fatalf("expected passive capture to avoid auto-plan, got %+v", result)
	}
	if got := len(service.ListProposals()); got != 0 {
		t.Fatalf("expected zero proposals, got %d", got)
	}
	if got := len(service.ListAgentJobs()); got != 0 {
		t.Fatalf("expected zero jobs, got %d", got)
	}
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
