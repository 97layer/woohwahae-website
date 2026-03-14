package runtime

import (
	"strings"
	"testing"
)

func TestSourceIntakeRedPolicyAutoOpensReviewRoomItem(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	observation, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_source_red_001",
		SourceChannel:     "telegram",
		Actor:             "founder",
		Topic:             SourceIntakeTopic,
		Confidence:        "high",
		Refs:              []string{"route:woohwahae", "domain:brand"},
		RawExcerpt:        BuildSourceIntakeRawExcerpt(SourceIntakeRecord{PolicyColor: "red", Title: "sensitive brand source", Excerpt: "raw source", FounderNote: "founder review 먼저", SuggestedRoutes: []string{"woohwahae"}}),
		NormalizedSummary: "Source intake · sensitive brand source -> 우화해",
	})
	if err != nil {
		t.Fatalf("create source observation: %v", err)
	}
	room := service.ReviewRoom()
	if got := len(room.Open); got != 1 {
		t.Fatalf("expected one open review-room item, got %+v", room.Open)
	}
	item := room.Open[0]
	if item.Source != "source_intake.red_policy" {
		t.Fatalf("unexpected review-room source: %+v", item)
	}
	if item.Ref == nil || *item.Ref != observation.ObservationID {
		t.Fatalf("expected review-room ref to match observation id, got %+v", item)
	}
	if !strings.Contains(item.Text, "founder 검토가 필요해") || !strings.Contains(item.Text, "sensitive brand source") {
		t.Fatalf("unexpected review-room text: %q", item.Text)
	}
	for _, want := range []string{"observation:" + observation.ObservationID, "policy:red", "route:woohwahae"} {
		if !containsStringLocal(item.Evidence, want) {
			t.Fatalf("expected review-room evidence %q in %+v", want, item.Evidence)
		}
	}
}

func TestSourceIntakeGreenPolicyDoesNotOpenReviewRoomItem(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_source_green_001",
		SourceChannel:     "telegram",
		Actor:             "founder",
		Topic:             SourceIntakeTopic,
		Confidence:        "high",
		Refs:              []string{"route:97layer"},
		RawExcerpt:        BuildSourceIntakeRawExcerpt(SourceIntakeRecord{PolicyColor: "green", Title: "normal source", Excerpt: "raw source", SuggestedRoutes: []string{"97layer"}}),
		NormalizedSummary: "Source intake · normal source -> 97layer",
	}); err != nil {
		t.Fatalf("create source observation: %v", err)
	}
	if room := service.ReviewRoom(); len(room.Open) != 0 {
		t.Fatalf("expected no open review-room items for green source intake, got %+v", room.Open)
	}
}

func TestSourceIntakeSensorRouteAmbiguityAutoOpensReviewRoomItem(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	observation, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_source_sensor_route_001",
		SourceChannel:     "crawler",
		Actor:             "rss_sensor",
		Topic:             SourceIntakeTopic,
		Confidence:        "high",
		Refs:              []string{"content_doc:post_sensor_route_001"},
		RawExcerpt:        BuildSourceIntakeRawExcerpt(SourceIntakeRecord{IntakeClass: "public_collector", PolicyColor: "yellow", Title: "ambiguous collector source", Excerpt: "raw source", FounderNote: "route cue가 비어 있어 founder 분류가 필요합니다."}),
		NormalizedSummary: "Source intake · ambiguous collector source -> 97layer",
	})
	if err != nil {
		t.Fatalf("create source observation: %v", err)
	}
	room := service.ReviewRoom()
	if got := len(room.Open); got != 1 {
		t.Fatalf("expected one open review-room item, got %+v", room.Open)
	}
	item := room.Open[0]
	if item.Source != "source_intake.sensor_route" {
		t.Fatalf("unexpected review-room source: %+v", item)
	}
	if item.Ref == nil || *item.Ref != observation.ObservationID {
		t.Fatalf("expected review-room ref to match observation id, got %+v", item)
	}
	if !strings.Contains(item.Text, "route cue가 비어") {
		t.Fatalf("unexpected review-room text: %q", item.Text)
	}
	for _, want := range []string{"observation:" + observation.ObservationID, "policy:yellow", "sensor_route_review:true", "route_hint_count:0"} {
		if !containsStringLocal(item.Evidence, want) {
			t.Fatalf("expected review-room evidence %q in %+v", want, item.Evidence)
		}
	}
}

func TestEnsureSourceIntakeRouteDecisionObservationResolvesSensorRouteReview(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	source, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_source_sensor_route_010",
		SourceChannel:     "crawler",
		Actor:             "rss_sensor",
		Topic:             SourceIntakeTopic,
		Confidence:        "high",
		Refs:              []string{"content_doc:post_sensor_route_010"},
		RawExcerpt:        BuildSourceIntakeRawExcerpt(SourceIntakeRecord{IntakeClass: "public_collector", PolicyColor: "yellow", Title: "ambiguous collector source", Excerpt: "raw source", FounderNote: "route cue가 비어 있어 founder 분류가 필요합니다."}),
		NormalizedSummary: "Source intake · ambiguous collector source -> 97layer",
	})
	if err != nil {
		t.Fatalf("create source observation: %v", err)
	}
	if got := len(service.ReviewRoom().Open); got != 1 {
		t.Fatalf("expected sensor-route review item before route decision, got %+v", service.ReviewRoom().Open)
	}
	if _, _, err := EnsureSourceIntakeRouteDecisionObservation(service, source, "97layer", "rss_sensor", "feed_sensor"); err != nil {
		t.Fatalf("ensure route decision observation: %v", err)
	}
	if room := service.ReviewRoom(); len(room.Open) != 0 {
		t.Fatalf("expected sensor-route item to resolve after route decision, got %+v", room.Open)
	}
}

func TestEnsureSourceDraftSeedObservationResolvesRedPolicyReviewRoomItem(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	source, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_source_red_010",
		SourceChannel:     "telegram",
		Actor:             "founder",
		Topic:             SourceIntakeTopic,
		Confidence:        "high",
		Refs:              []string{"route:97layer"},
		RawExcerpt:        BuildSourceIntakeRawExcerpt(SourceIntakeRecord{PolicyColor: "red", Title: "route me carefully", Excerpt: "raw source", SuggestedRoutes: []string{"97layer"}}),
		NormalizedSummary: "Source intake · route me carefully -> 97layer",
	})
	if err != nil {
		t.Fatalf("create source observation: %v", err)
	}
	if got := len(service.ReviewRoom().Open); got != 1 {
		t.Fatalf("expected review-room item before route decision, got %+v", service.ReviewRoom().Open)
	}
	if _, _, err := EnsureSourceDraftSeedObservation(service, source, "observation_route_010", "97layer"); err != nil {
		t.Fatalf("ensure draft seed observation: %v", err)
	}
	if room := service.ReviewRoom(); len(room.Open) != 0 {
		t.Fatalf("expected red-policy item to resolve after route decision, got %+v", room.Open)
	}
}

func TestEnsureSourceDraftSeedProposalCreatesProposal(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	source, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_source_100",
		SourceChannel:     "telegram",
		Actor:             "founder",
		Topic:             SourceIntakeTopic,
		Confidence:        "high",
		Refs:              []string{"route:97layer"},
		RawExcerpt:        BuildSourceIntakeRawExcerpt(SourceIntakeRecord{Title: "worldview note", Excerpt: "raw source", SuggestedRoutes: []string{"97layer"}}),
		NormalizedSummary: "Source intake · worldview note -> 97layer",
	})
	if err != nil {
		t.Fatalf("create source observation: %v", err)
	}

	proposal, created, err := EnsureSourceDraftSeedProposal(service, source, "observation_route_001", "woohwahae")
	if err != nil {
		t.Fatalf("ensure source draft seed proposal: %v", err)
	}
	if !created {
		t.Fatalf("expected proposal to be created")
	}
	if proposal.Surface != SurfaceTelegram || proposal.Status != "proposed" {
		t.Fatalf("unexpected proposal: %+v", proposal)
	}
	for _, needle := range []string{SourceDraftSeedLaneNote, "account:woohwahae", "source_observation:observation_source_100", "route_decision:observation_route_001"} {
		if !strings.Contains(strings.Join(proposal.Notes, " "), needle) {
			t.Fatalf("expected note %q in %+v", needle, proposal.Notes)
		}
	}
	drafts := service.ListObservations(ObservationQuery{Topic: SourceDraftSeedTopic, Limit: 4})
	if len(drafts) != 1 {
		t.Fatalf("expected one source draft seed observation, got %+v", drafts)
	}
	draft := ParseSourceDraftSeedRawExcerpt(drafts[0].RawExcerpt)
	if draft.TargetAccount != "woohwahae" || draft.TargetToneLevel != "polished" {
		t.Fatalf("unexpected draft seed record: %+v", draft)
	}
	if draft.Draft == "" {
		t.Fatalf("expected draft body to be generated")
	}
}

func TestEnsureSourceDraftSeedProposalDedupesBySourceAndAccount(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	source, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_source_101",
		SourceChannel:     "telegram",
		Actor:             "founder",
		Topic:             SourceIntakeTopic,
		Confidence:        "high",
		Refs:              []string{"route:97layer"},
		RawExcerpt:        BuildSourceIntakeRawExcerpt(SourceIntakeRecord{Title: "maker note", Excerpt: "raw source", SuggestedRoutes: []string{"97layer"}}),
		NormalizedSummary: "Source intake · maker note -> 97layer",
	})
	if err != nil {
		t.Fatalf("create source observation: %v", err)
	}

	first, created, err := EnsureSourceDraftSeedProposal(service, source, "observation_route_002", "97layer")
	if err != nil || !created {
		t.Fatalf("first ensure failed: proposal=%+v created=%v err=%v", first, created, err)
	}
	second, createdAgain, err := EnsureSourceDraftSeedProposal(service, source, "observation_route_003", "97layer")
	if err != nil {
		t.Fatalf("second ensure failed: %v", err)
	}
	if createdAgain {
		t.Fatalf("expected second ensure to reuse proposal")
	}
	if first.ProposalID != second.ProposalID {
		t.Fatalf("expected same proposal id, got %s vs %s", first.ProposalID, second.ProposalID)
	}
	if got := len(service.ListProposals()); got != 1 {
		t.Fatalf("expected one proposal, got %d", got)
	}
	if got := len(service.ListObservations(ObservationQuery{Topic: SourceDraftSeedTopic, Limit: 4})); got != 1 {
		t.Fatalf("expected one source draft observation, got %d", got)
	}
}

func TestOpenSourceIntakeDraftLaneCreatesRouteDecisionDraftAndProposal(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	source, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_source_auto_001",
		SourceChannel:     "crawler",
		Actor:             "rss_sensor",
		Topic:             SourceIntakeTopic,
		Confidence:        "high",
		Refs:              []string{"route:woosunhokr", "content_doc:post_1"},
		RawExcerpt:        BuildSourceIntakeRawExcerpt(SourceIntakeRecord{Title: "beauty practice note", URL: "https://example.com/post-1", Excerpt: "실무 감각을 더 조용하게 번역한 메모.", DomainTags: []string{"beauty"}, WorldviewTags: []string{"subtraction"}, SuggestedRoutes: []string{"woosunhokr"}}),
		NormalizedSummary: "Source intake · beauty practice note -> 우순호",
	})
	if err != nil {
		t.Fatalf("create source observation: %v", err)
	}

	result, err := OpenSourceIntakeDraftLane(service, source, "woosunhokr", "rss_sensor", "feed_sensor")
	if err != nil {
		t.Fatalf("open source intake draft lane: %v", err)
	}
	if result.Reused {
		t.Fatalf("expected first lane open to create fresh route/draft/proposal")
	}
	if result.RouteDecision.Topic != IntakeRouteDecisionTopic {
		t.Fatalf("unexpected route decision: %+v", result.RouteDecision)
	}
	if !containsStringLocal(result.RouteDecision.Refs, "decision_route:woosunhokr") || !containsStringLocal(result.RouteDecision.Refs, "route_source:feed_sensor") {
		t.Fatalf("unexpected route decision refs: %+v", result.RouteDecision.Refs)
	}
	draft := ParseSourceDraftSeedRawExcerpt(result.DraftSeed.RawExcerpt)
	if draft.TargetAccount != "woosunhokr" {
		t.Fatalf("unexpected draft seed record: %+v", draft)
	}
	if !containsStringLocal(result.Proposal.Notes, "account:woosunhokr") {
		t.Fatalf("unexpected proposal notes: %+v", result.Proposal.Notes)
	}
}

func TestReviseSourceDraftSeedObservationCreatesLatestRevision(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	source, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_source_102",
		SourceChannel:     "telegram",
		Actor:             "founder",
		Topic:             SourceIntakeTopic,
		Confidence:        "high",
		Refs:              []string{"route:97layer"},
		RawExcerpt:        BuildSourceIntakeRawExcerpt(SourceIntakeRecord{Title: "maker note", Excerpt: "raw source", FounderNote: "덜 추상적으로", SuggestedRoutes: []string{"97layer"}}),
		NormalizedSummary: "Source intake · maker note -> 97layer",
	})
	if err != nil {
		t.Fatalf("create source observation: %v", err)
	}
	seed, _, err := EnsureSourceDraftSeedObservation(service, source, "observation_route_010", "97layer")
	if err != nil {
		t.Fatalf("seed draft observation: %v", err)
	}

	revised, err := ReviseSourceDraftSeedObservation(service, seed, "조금 더 구체적으로")
	if err != nil {
		t.Fatalf("revise source draft seed: %v", err)
	}
	if revised.ObservationID == seed.ObservationID {
		t.Fatalf("expected new observation id for revision")
	}
	draft := ParseSourceDraftSeedRawExcerpt(revised.RawExcerpt)
	if draft.ParentDraftObservationID != seed.ObservationID || draft.RevisionNote != "조금 더 구체적으로" {
		t.Fatalf("unexpected revised draft seed record: %+v", draft)
	}
	if !strings.Contains(draft.Draft, "조금 더 구체적으로") {
		t.Fatalf("expected revised draft body to include instruction, got %q", draft.Draft)
	}
}

func TestBuildSourceDraftSeedBodySeparatesAccountVoices(t *testing.T) {
	record := SourceIntakeRecord{
		Title:         "브랜드 구축 메모",
		Excerpt:       "홈페이지와 운영면을 같이 정리하다 보니 결국 기능보다 구조를 먼저 봐야 한다는 쪽으로 생각이 모였다.",
		FounderNote:   "미용 실무와 공적 브랜드 문장으로도 번역 가능해야 함",
		DomainTags:    []string{"system", "beauty"},
		WorldviewTags: []string{"subtraction", "identity"},
	}

	raw := buildSourceDraftSeedBody("97layer", "브랜드 구축 메모", record)
	refined := buildSourceDraftSeedBody("woosunhokr", "브랜드 구축 메모", record)
	polished := buildSourceDraftSeedBody("woohwahae", "브랜드 구축 메모", record)

	if raw == refined || refined == polished || raw == polished {
		t.Fatalf("expected account voices to diverge, got raw=%q refined=%q polished=%q", raw, refined, polished)
	}
	if !strings.Contains(raw, "97layer에서는") || !strings.Contains(raw, "기록에 가까운 메모") {
		t.Fatalf("expected raw 97layer voice markers, got %q", raw)
	}
	if !strings.Contains(refined, "우순호에서는") || !strings.Contains(refined, "미용사의 단상") {
		t.Fatalf("expected refined woosunhokr voice markers, got %q", refined)
	}
	if !strings.Contains(polished, "우화해에서는") || !strings.Contains(polished, "슬로우 라이프") || !strings.Contains(polished, "공적 문장") {
		t.Fatalf("expected polished woohwahae voice markers, got %q", polished)
	}
}

func TestOpenSourceDraftSeedPublishPrepCreatesCanonicalThreadsLane(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	source, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_source_200",
		SourceChannel:     "telegram",
		Actor:             "founder",
		Topic:             SourceIntakeTopic,
		Confidence:        "high",
		Refs:              []string{"route:97layer"},
		RawExcerpt:        BuildSourceIntakeRawExcerpt(SourceIntakeRecord{Title: "operating surface rebuild", Excerpt: "홈페이지를 다시 세우고 운영 페이지를 분리하면서 결국 기능보다 구조와 순서를 먼저 보게 됐다.", FounderNote: "브랜드 구축 과정에 더 가깝게", DomainTags: []string{"system", "brand"}, WorldviewTags: []string{"identity"}, SuggestedRoutes: []string{"97layer"}}),
		NormalizedSummary: "Source intake · operating surface rebuild -> 97layer",
	})
	if err != nil {
		t.Fatalf("create source observation: %v", err)
	}
	draftObservation, _, err := EnsureSourceDraftSeedObservation(service, source, "observation_route_200", "97layer")
	if err != nil {
		t.Fatalf("ensure draft seed observation: %v", err)
	}

	result, err := service.OpenSourceDraftSeedPublishPrep(draftObservation.ObservationID, "threads")
	if err != nil {
		t.Fatalf("open source draft seed publish prep: %v", err)
	}
	if result.Reused {
		t.Fatalf("expected first prep lane open to create new records")
	}
	if result.Lane.Channel != "threads" || result.Lane.TargetAccount != "97layer" {
		t.Fatalf("unexpected lane: %+v", result.Lane)
	}
	if !strings.Contains(strings.Join(result.Proposal.Notes, " "), brandPublishLaneNote) || !strings.Contains(strings.Join(result.Proposal.Notes, " "), "source_draft_seed:"+draftObservation.ObservationID) {
		t.Fatalf("expected canonical brand lane notes, got %+v", result.Proposal.Notes)
	}
	if result.WorkItem.ID == "" || result.Approval.ApprovalID == "" || result.Flow.FlowID == "" {
		t.Fatalf("expected promoted work/approval/flow, got %+v", result)
	}
	if result.Observation.Topic != threadsBrandPrepTopic {
		t.Fatalf("expected prep observation, got %+v", result.Observation)
	}
	if !strings.Contains(result.Observation.RawExcerpt, "notes="+brandPublishLaneNote) && !strings.Contains(result.Observation.RawExcerpt, brandPublishLaneNote) {
		t.Fatalf("expected prep observation to capture canonical notes, got %q", result.Observation.RawExcerpt)
	}
	if got := len(service.ListApprovals()); got != 1 {
		t.Fatalf("expected one approval, got %d", got)
	}
	if got := len(service.ListFlows()); got != 1 {
		t.Fatalf("expected one flow, got %d", got)
	}
}

func TestOpenSourceDraftSeedPublishPrepDedupesByDraftObservation(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	source, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_source_201",
		SourceChannel:     "telegram",
		Actor:             "founder",
		Topic:             SourceIntakeTopic,
		Confidence:        "high",
		Refs:              []string{"route:97layer"},
		RawExcerpt:        BuildSourceIntakeRawExcerpt(SourceIntakeRecord{Title: "quiet beauty note", Excerpt: "실무에서는 결국 손기술보다 기준이 먼저 남는다는 생각이 다시 들었다.", FounderNote: "미용사의 단상으로", DomainTags: []string{"beauty", "practice"}, WorldviewTags: []string{"subtraction"}, SuggestedRoutes: []string{"woosunhokr"}}),
		NormalizedSummary: "Source intake · quiet beauty note -> woosunhokr",
	})
	if err != nil {
		t.Fatalf("create source observation: %v", err)
	}
	draftObservation, _, err := EnsureSourceDraftSeedObservation(service, source, "observation_route_201", "woosunhokr")
	if err != nil {
		t.Fatalf("ensure draft seed observation: %v", err)
	}

	first, err := service.OpenSourceDraftSeedPublishPrep(draftObservation.ObservationID, "threads")
	if err != nil {
		t.Fatalf("first prep open: %v", err)
	}
	second, err := service.OpenSourceDraftSeedPublishPrep(draftObservation.ObservationID, "threads")
	if err != nil {
		t.Fatalf("second prep open: %v", err)
	}
	if !second.Reused {
		t.Fatalf("expected second prep open to reuse existing lane")
	}
	if first.Approval.ApprovalID != second.Approval.ApprovalID || first.Observation.ObservationID != second.Observation.ObservationID {
		t.Fatalf("expected second open to reuse the same lane ids, got first=%+v second=%+v", first, second)
	}
	if got := len(service.ListProposals()); got != 1 {
		t.Fatalf("expected one proposal, got %d", got)
	}
	if got := len(service.ListWorkItems()); got != 1 {
		t.Fatalf("expected one work item, got %d", got)
	}
	if got := len(service.ListApprovals()); got != 1 {
		t.Fatalf("expected one approval, got %d", got)
	}
	if got := len(service.ListFlows()); got != 1 {
		t.Fatalf("expected one flow, got %d", got)
	}
	if got := len(service.ListObservations(ObservationQuery{Topic: threadsBrandPrepTopic, Limit: 8})); got != 1 {
		t.Fatalf("expected one prep observation, got %d", got)
	}
}

func TestOpenSourceDraftSeedPublishPrepResolvesLingeringRedPolicyItem(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	source, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_source_red_200",
		SourceChannel:     "telegram",
		Actor:             "founder",
		Topic:             SourceIntakeTopic,
		Confidence:        "high",
		Refs:              []string{"route:woohwahae"},
		RawExcerpt:        BuildSourceIntakeRawExcerpt(SourceIntakeRecord{PolicyColor: "red", Title: "sensitive prep note", Excerpt: "raw source", SuggestedRoutes: []string{"woohwahae"}}),
		NormalizedSummary: "Source intake · sensitive prep note -> 우화해",
	})
	if err != nil {
		t.Fatalf("create source observation: %v", err)
	}
	draftObservation, _, err := EnsureSourceDraftSeedObservation(service, source, "observation_route_200", "woohwahae")
	if err != nil {
		t.Fatalf("ensure draft seed observation: %v", err)
	}
	ref := source.ObservationID
	manual := newSignalReviewRoomItem(
		"Lingering red policy reminder for prep path.",
		"source_intake.red_policy",
		&ref,
		"manual re-open for prep-path regression",
		"review_room.manual.source_intake_red_policy",
		[]string{"observation:" + source.ObservationID, "policy:red"},
	)
	if _, err := service.AddStructuredReviewRoomItem("open", manual); err != nil {
		t.Fatalf("re-open review-room item: %v", err)
	}
	if _, err := service.OpenSourceDraftSeedPublishPrep(draftObservation.ObservationID, "threads"); err != nil {
		t.Fatalf("open source draft publish prep: %v", err)
	}
	if room := service.ReviewRoom(); len(room.Open) != 0 {
		t.Fatalf("expected lingering red-policy item to resolve after prep open, got %+v", room.Open)
	}
}

func containsStringLocal(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
