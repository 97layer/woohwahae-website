package runtime

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestTelegramBotCommandRepliesStayNatural(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	if got := handler.HandleMessage("/start"); got != "올라와 있어. 편하게 말해줘. 이 방 번호가 필요하면 `/whoami` 보내면 돼." {
		t.Fatalf("unexpected /start reply: %q", got)
	}
	if got := handler.HandleMessage("/review"); got != "지금 열린 리뷰 안건은 없어." {
		t.Fatalf("unexpected /review reply: %q", got)
	}
}

func TestTelegramBotWhoAmIReplyShowsChatIdentity(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   -10012345,
		RouteID:  TelegramRouteOps,
		ChatType: "supergroup",
	}, "/whoami")
	for _, needle := range []string{
		"chat_id: -10012345",
		"chat_type: supergroup",
		"route: ops",
	} {
		if !strings.Contains(reply, needle) {
			t.Fatalf("expected %q in %q", needle, reply)
		}
	}
}

func TestTelegramBotScopesNonFounderRoutes(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	if got := handler.HandleMessageWithContext(TelegramInboundContext{RouteID: TelegramRouteOps}, "지금 뭐부터 해야 해?"); !strings.Contains(got, "ops route") {
		t.Fatalf("expected ops route guardrail reply, got %q", got)
	}
	if got := handler.HandleMessageWithContext(TelegramInboundContext{RouteID: TelegramRouteBrand}, "/handoff"); got != "이 기능은 founder 1:1 DM에서만 열어둘게." {
		t.Fatalf("expected founder-only handoff reply, got %q", got)
	}
	if got := handler.HandleMessageWithContext(TelegramInboundContext{}, "/status"); !strings.Contains(got, "/whoami") {
		t.Fatalf("expected unmapped route reply, got %q", got)
	}
}

func TestTelegramBotFounderRoomBlocksConversationUntilDMIsConfigured(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	t.Setenv("TELEGRAM_FOUNDER_CHAT_ID", "-5060692298")

	handler := NewTelegramBotHandler(service)
	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   -5060692298,
		RouteID:  TelegramRouteFounder,
		ChatType: "supergroup",
	}, "지금 뭐부터 해야 해?")
	if !strings.Contains(reply, "founder room") {
		t.Fatalf("expected founder room guardrail reply, got %q", reply)
	}

	reply = handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   -5060692298,
		RouteID:  TelegramRouteFounder,
		ChatType: "supergroup",
	}, "/handoff")
	if !strings.Contains(reply, "1:1 DM") {
		t.Fatalf("expected founder DM requirement reply, got %q", reply)
	}
}

func TestTelegramBotFounderDMAllowsConversationWhenSplitFromRoom(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	t.Setenv("TELEGRAM_FOUNDER_CHAT_ID", "-5060692298")
	t.Setenv("TELEGRAM_FOUNDER_DM_CHAT_ID", "123456789")

	handler := NewTelegramBotHandler(service)
	if got := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   123456789,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "/handoff"); !strings.Contains(got, "지금 이어받을 내용") {
		t.Fatalf("expected founder DM handoff reply, got %q", got)
	}
}

func TestTelegramBotFounderNoteKeepsRouteProvenance(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	reply := handler.HandleMessageWithContext(TelegramInboundContext{ChatID: 1234, RouteID: TelegramRouteFounder, ChatType: "private"}, "/note 오늘 founder 정리")
	if reply != "메모로 남겼어: 오늘 founder 정리" {
		t.Fatalf("unexpected note reply: %q", reply)
	}

	items := service.ListObservations(ObservationQuery{SourceChannel: "conversation:telegram", Limit: 1})
	if len(items) != 1 {
		t.Fatalf("expected one observation, got %+v", items)
	}
	if items[0].Actor != "founder" {
		t.Fatalf("expected founder actor, got %+v", items[0])
	}
	refs := strings.Join(items[0].Refs, " ")
	for _, needle := range []string{"telegram_chat:1234", "telegram_route:founder", "telegram_chat_type:private"} {
		if !strings.Contains(refs, needle) {
			t.Fatalf("expected ref %q in %+v", needle, items[0].Refs)
		}
	}
	if got := len(service.ListProposals()); got != 0 {
		t.Fatalf("expected plain note to avoid auto proposal, got %d", got)
	}
}

func TestTelegramBotFounderIntakeListsPendingSourceItems(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_source_001",
		SourceChannel:     "cockpit",
		Actor:             "founder",
		Topic:             "source_intake",
		Confidence:        "high",
		Refs:              []string{"route:97layer"},
		RawExcerpt:        "intake_class=manual_drop\npolicy_color=green\ntitle=brand note from link\nsuggested_routes=97layer,woohwahae\nexcerpt:\nsource body",
		NormalizedSummary: "Source intake · brand note from link -> 97layer",
	}); err != nil {
		t.Fatalf("create source intake: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "/intake")

	for _, needle := range []string{
		"최근 source intake",
		"observation_source_001",
		"brand note from link",
		"97layer, 우화해",
	} {
		if !strings.Contains(reply, needle) {
			t.Fatalf("expected %q in %q", needle, reply)
		}
	}
}

func TestTelegramBotFounderDropCreatesSourceIntakeObservation(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "/drop https://example.com/article 이 글은 브랜드 세계관 쪽 참고")

	if !strings.Contains(reply, "source intake로 넣었어") || !strings.Contains(reply, "다음은 /route") {
		t.Fatalf("unexpected drop reply: %q", reply)
	}

	items := service.ListObservations(ObservationQuery{Topic: SourceIntakeTopic, Limit: 1})
	if len(items) != 1 {
		t.Fatalf("expected one source intake observation, got %+v", items)
	}
	if items[0].SourceChannel != "telegram" || items[0].Actor != "founder" {
		t.Fatalf("unexpected source intake observation: %+v", items[0])
	}
	raw := items[0].RawExcerpt
	for _, needle := range []string{
		"intake_class=manual_drop",
		"policy_color=green",
		"url=https://example.com/article",
		"suggested_routes=97layer",
		"excerpt:",
	} {
		if !strings.Contains(raw, needle) {
			t.Fatalf("expected %q in raw excerpt %q", needle, raw)
		}
	}
}

func TestTelegramBotFounderRoomPlainTextAutoCreatesSourceIntake(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   -5060692298,
		RouteID:  TelegramRouteFounder,
		ChatType: "supergroup",
	}, "https://example.com/story 이건 나중에 우화해 쪽으로도 볼만한 소재")

	if !strings.Contains(reply, "source intake로 넣었어") {
		t.Fatalf("expected auto intake reply, got %q", reply)
	}

	items := service.ListObservations(ObservationQuery{Topic: SourceIntakeTopic, Limit: 1})
	if len(items) != 1 {
		t.Fatalf("expected one auto-created source intake, got %+v", items)
	}
	if items[0].SourceChannel != "telegram" {
		t.Fatalf("expected telegram source channel, got %+v", items[0])
	}
	if !strings.Contains(items[0].RawExcerpt, "url=https://example.com/story") {
		t.Fatalf("expected url to be captured, got %q", items[0].RawExcerpt)
	}
}

func TestTelegramBotFounderRouteDecisionCreatesObservation(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_source_002",
		SourceChannel:     "cockpit",
		Actor:             "founder",
		Topic:             "source_intake",
		Confidence:        "high",
		Refs:              []string{"route:97layer"},
		RawExcerpt:        "intake_class=manual_drop\npolicy_color=yellow\ntitle=beauty practice note\nsuggested_routes=woosunhokr\nexcerpt:\nsource body",
		NormalizedSummary: "Source intake · beauty practice note -> woosunhokr",
	}); err != nil {
		t.Fatalf("create source intake: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "/route observation_source_002 woosunhokr")

	if !strings.Contains(reply, "route 결정 남겼어") || !strings.Contains(reply, "우순호") || !strings.Contains(reply, "proposal:") || !strings.Contains(reply, "초안 미리보기:") || !strings.Contains(reply, "prep: /prep") {
		t.Fatalf("unexpected route reply: %q", reply)
	}

	items := service.ListObservations(ObservationQuery{Topic: "intake_route_decision", Limit: 1})
	if len(items) != 1 {
		t.Fatalf("expected one route decision observation, got %+v", items)
	}
	if items[0].Topic != "intake_route_decision" {
		t.Fatalf("unexpected topic: %+v", items[0])
	}
	refs := strings.Join(items[0].Refs, " ")
	for _, needle := range []string{"observation_source_002", "decision_route:woosunhokr", "telegram_chat:7565534667", "telegram_route:founder"} {
		if !strings.Contains(refs, needle) {
			t.Fatalf("expected ref %q in %+v", needle, items[0].Refs)
		}
	}
	proposals := service.ListProposals()
	if len(proposals) != 1 {
		t.Fatalf("expected one draft seed proposal, got %+v", proposals)
	}
	if !strings.Contains(strings.Join(proposals[0].Notes, " "), "account:woosunhokr") {
		t.Fatalf("expected target account note, got %+v", proposals[0])
	}
	drafts := service.ListObservations(ObservationQuery{Topic: SourceDraftSeedTopic, Limit: 4})
	if len(drafts) != 1 {
		t.Fatalf("expected one source draft seed observation, got %+v", drafts)
	}
	draft := ParseSourceDraftSeedRawExcerpt(drafts[0].RawExcerpt)
	if draft.TargetAccount != "woosunhokr" || !strings.Contains(draft.Draft, "미용사의 단상") {
		t.Fatalf("unexpected source draft seed: %+v", draft)
	}
}

func TestTelegramBotFounderDraftsListShowsRecentDraftSeeds(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_draft_seed_001",
		SourceChannel:     "telegram",
		Actor:             "founder",
		Topic:             SourceDraftSeedTopic,
		Confidence:        "medium",
		Refs:              []string{"observation_source_002", "account:97layer"},
		RawExcerpt:        BuildSourceDraftSeedRawExcerpt(SourceDraftSeedRecord{TargetAccount: "97layer", TargetToneLevel: "raw", Title: "97layer raw draft · worldview note", SourceObservationID: "observation_source_002", Draft: "요즘은 만드는 중이라기보다 정리하는 중에 가깝다."}),
		NormalizedSummary: "Draft seed opened · worldview note -> 97layer",
	}); err != nil {
		t.Fatalf("create source draft seed: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "/drafts")

	if !strings.Contains(reply, "최근 route 초안") || !strings.Contains(reply, "97layer") || !strings.Contains(reply, "정리하는 중에 가깝다") {
		t.Fatalf("unexpected drafts reply: %q", reply)
	}
}

func TestTelegramBotFounderRedraftCreatesLatestDraftRevision(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	source, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_source_103",
		SourceChannel:     "telegram",
		Actor:             "founder",
		Topic:             SourceIntakeTopic,
		Confidence:        "high",
		Refs:              []string{"route:97layer"},
		RawExcerpt:        BuildSourceIntakeRawExcerpt(SourceIntakeRecord{Title: "worldview note", Excerpt: "raw source", FounderNote: "덜 추상적으로", SuggestedRoutes: []string{"97layer"}}),
		NormalizedSummary: "Source intake · worldview note -> 97layer",
	})
	if err != nil {
		t.Fatalf("create source observation: %v", err)
	}
	seed, _, err := EnsureSourceDraftSeedObservation(service, source, "observation_route_011", "97layer")
	if err != nil {
		t.Fatalf("seed draft observation: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "/redraft "+seed.ObservationID+" 조금 더 구체적으로")

	if !strings.Contains(reply, "초안 다시 열었어") || !strings.Contains(reply, "초안 미리보기:") {
		t.Fatalf("unexpected redraft reply: %q", reply)
	}

	items := telegramRecentSourceDraftSeedItems(service.ListObservations(ObservationQuery{Topic: SourceDraftSeedTopic, Limit: 8}))
	if len(items) != 1 {
		t.Fatalf("expected one latest draft item, got %+v", items)
	}
	if items[0].ObservationID == seed.ObservationID {
		t.Fatalf("expected latest draft item to move past the original seed, got %+v", items[0])
	}
	if !strings.Contains(reply, items[0].ObservationID) {
		t.Fatalf("expected redraft reply to reference the latest draft item, got reply=%q item=%+v", reply, items[0])
	}

	drafts := service.ListObservations(ObservationQuery{Topic: SourceDraftSeedTopic, Limit: 8})
	if len(drafts) != 2 {
		t.Fatalf("expected original and revised draft observations, got %+v", drafts)
	}
	revised := ParseSourceDraftSeedRawExcerpt(drafts[0].RawExcerpt)
	if revised.ParentDraftObservationID != seed.ObservationID || revised.RevisionNote != "조금 더 구체적으로" {
		t.Fatalf("unexpected revised draft seed record: %+v", revised)
	}
}

func TestTelegramBotFounderPrepOpensThreadsLaneFromDraftSeed(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	source, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_source_301",
		SourceChannel:     "telegram",
		Actor:             "founder",
		Topic:             SourceIntakeTopic,
		Confidence:        "high",
		Refs:              []string{"route:97layer"},
		RawExcerpt:        BuildSourceIntakeRawExcerpt(SourceIntakeRecord{Title: "operating surface rebuild", Excerpt: "홈페이지와 운영 페이지를 나누면서 기능보다 구조가 먼저 남았다.", FounderNote: "브랜드 구축 과정으로", DomainTags: []string{"system", "brand"}, WorldviewTags: []string{"identity"}, SuggestedRoutes: []string{"97layer"}}),
		NormalizedSummary: "Source intake · operating surface rebuild -> 97layer",
	})
	if err != nil {
		t.Fatalf("create source observation: %v", err)
	}
	seed, _, err := EnsureSourceDraftSeedObservation(service, source, "observation_route_301", "97layer")
	if err != nil {
		t.Fatalf("seed draft observation: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "/prep "+seed.ObservationID)

	if !strings.Contains(reply, "Threads prep 열었어") || !strings.Contains(reply, "approval:") || !strings.Contains(reply, "본문 미리보기:") {
		t.Fatalf("unexpected prep reply: %q", reply)
	}
	if got := len(service.ListObservations(ObservationQuery{Topic: threadsBrandPrepTopic, Limit: 4})); got != 1 {
		t.Fatalf("expected one prep observation, got %d", got)
	}
	if got := len(service.ListApprovals()); got != 1 {
		t.Fatalf("expected one approval, got %d", got)
	}
}

func TestTelegramBotBlocksIntakeDecisionOutsideFounderRoute(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   -5060692298,
		RouteID:  TelegramRouteOps,
		ChatType: "supergroup",
	}, "/intake")
	if !strings.Contains(reply, "founder route") {
		t.Fatalf("expected founder route guardrail for /intake, got %q", reply)
	}

	reply = handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   -5060692298,
		RouteID:  TelegramRouteOps,
		ChatType: "supergroup",
	}, "/drop https://example.com")
	if !strings.Contains(reply, "founder route") {
		t.Fatalf("expected founder route guardrail for /drop, got %q", reply)
	}

	reply = handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   -5060692298,
		RouteID:  TelegramRouteBrand,
		ChatType: "supergroup",
	}, "/route observation_source_002 hold")
	if !strings.Contains(reply, "founder route") {
		t.Fatalf("expected founder route guardrail for /route, got %q", reply)
	}
}

func TestTelegramBotFounderDMPlainTextPersistsConversationAndOpensPlannerLane(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "텔레그램 founder DM을 canonical conversation spine에 붙이고 다음 작업도 구현해줘")
	if !strings.Contains(reply, "conversation spine에 남겼어:") || !strings.Contains(reply, "planner 레인도 붙였어:") {
		t.Fatalf("expected persisted founder DM reply, got %q", reply)
	}
	items := service.ListObservations(ObservationQuery{SourceChannel: "conversation:telegram", Limit: 1})
	if len(items) != 1 {
		t.Fatalf("expected one founder DM observation, got %+v", items)
	}
	refs := strings.Join(items[0].Refs, " ")
	for _, needle := range []string{"telegram_chat:7565534667", "telegram_route:founder", "telegram_chat_type:private"} {
		if !strings.Contains(refs, needle) {
			t.Fatalf("expected founder DM ref %q in %+v", needle, items[0].Refs)
		}
	}
	if got := len(service.ListProposals()); got != 1 {
		t.Fatalf("expected one proposal from founder DM follow-up, got %d", got)
	}
	if got := len(service.ListAgentJobs()); got != 1 {
		t.Fatalf("expected one planner job from founder DM follow-up, got %d", got)
	}
}

func TestTelegramBotFounderApprovalsListAndApprove(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateApproval(ApprovalItem{
		ApprovalID:      "approval_telegram_001",
		WorkItemID:      "work_telegram_001",
		Stage:           StageVerify,
		Summary:         "Ship founder Telegram control surface",
		Risks:           []string{"review drift"},
		RollbackPlan:    "hold release",
		DecisionSurface: SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     zeroSafeNow(),
	}); err != nil {
		t.Fatalf("create approval: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	listReply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "/approvals")
	for _, needle := range []string{"approval_telegram_001", "Ship founder Telegram control surface", "review drift"} {
		if !strings.Contains(listReply, needle) {
			t.Fatalf("expected %q in %q", needle, listReply)
		}
	}

	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "/approve approval_telegram_001")
	if !strings.Contains(reply, "승인 처리했어: approval_telegram_001") {
		t.Fatalf("unexpected approve reply: %q", reply)
	}

	items := service.ListApprovals()
	if len(items) != 1 || items[0].Status != "approved" {
		t.Fatalf("expected approval to resolve, got %+v", items)
	}
}

func TestTelegramBotFounderJobsListAndDispatch(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "api")
	t.Setenv("LAYER_OS_PROVIDERS", "openai")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")

	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateAgentJob(AgentJob{
		JobID:     "job_telegram_001",
		Kind:      "plan",
		Role:      "planner",
		Summary:   "Plan the Telegram control layer",
		Status:    "queued",
		Source:    "founder.manual",
		Surface:   SurfaceTelegram,
		Stage:     StageDiscover,
		Notes:     []string{},
		CreatedAt: zeroSafeNow(),
		UpdatedAt: zeroSafeNow(),
	}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	listReply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   -10012345,
		RouteID:  TelegramRouteOps,
		ChatType: "supergroup",
	}, "/jobs")
	for _, needle := range []string{"job_telegram_001", "planner / queued", "Plan the Telegram control layer"} {
		if !strings.Contains(listReply, needle) {
			t.Fatalf("expected %q in %q", needle, listReply)
		}
	}

	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "/dispatch job_telegram_001")
	for _, needle := range []string{"dispatch 걸었어: job_telegram_001", "상태: packet_ready", "packet-capable worker"} {
		if !strings.Contains(reply, needle) {
			t.Fatalf("expected %q in %q", needle, reply)
		}
	}

	jobs := service.ListAgentJobs()
	if len(jobs) != 1 || jobs[0].Status != "running" {
		t.Fatalf("expected job to move into running, got %+v", jobs)
	}
	if jobs[0].Result == nil || jobs[0].Result["dispatch_state"] != "packet_ready" {
		t.Fatalf("expected packet-ready result, got %+v", jobs[0].Result)
	}
}

func TestTelegramBotFounderReviewReplyShowsRefAndAcceptsItem(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	ref := "verification:telegram_control_001"
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{
		Text:      "Triage Telegram control drift.",
		Kind:      "automation_failure",
		Severity:  "high",
		Source:    "verification.failed",
		Ref:       &ref,
		CreatedAt: zeroSafeNow(),
		UpdatedAt: zeroSafeNow(),
	}); err != nil {
		t.Fatalf("add review item: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	reviewReply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "/review")
	for _, needle := range []string{"Triage Telegram control drift.", "ref: verification:telegram_control_001", "/accept <ref|text>"} {
		if !strings.Contains(reviewReply, needle) {
			t.Fatalf("expected %q in %q", needle, reviewReply)
		}
	}

	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "/accept verification:telegram_control_001 :: founder control lane now")
	if !strings.Contains(reply, "채택 처리했어: verification:telegram_control_001") {
		t.Fatalf("unexpected review accept reply: %q", reply)
	}

	room := service.ReviewRoom()
	if len(room.Open) != 0 || len(room.Accepted) != 1 {
		t.Fatalf("expected accepted review item, got %+v", room)
	}
	if room.Accepted[0].Resolution == nil || room.Accepted[0].Resolution.Action != "accept" || room.Accepted[0].Resolution.Reason != "founder control lane now" {
		t.Fatalf("expected acceptance resolution, got %+v", room.Accepted[0].Resolution)
	}
}

func TestTelegramBotFounderDMNaturalApprovalIntentActsWithoutSlash(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateApproval(ApprovalItem{
		ApprovalID:      "approval_telegram_002",
		WorkItemID:      "work_telegram_002",
		Stage:           StageVerify,
		Summary:         "Approve the Telegram assistant lane",
		Risks:           []string{"shell-only drift"},
		RollbackPlan:    "hold release",
		DecisionSurface: SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     zeroSafeNow(),
	}); err != nil {
		t.Fatalf("create approval: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "approval_telegram_002 승인해")

	for _, needle := range []string{"승인 처리했어: approval_telegram_002", "기록도 남겼어:"} {
		if !strings.Contains(reply, needle) {
			t.Fatalf("expected %q in %q", needle, reply)
		}
	}
	items := service.ListApprovals()
	if len(items) != 1 || items[0].Status != "approved" {
		t.Fatalf("expected approval to resolve, got %+v", items)
	}
	if got := len(service.ListProposals()); got != 0 {
		t.Fatalf("expected control DM to avoid auto proposal noise, got %d", got)
	}
}

func TestTelegramBotFounderDMNaturalDispatchIntentActsWithoutSlash(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "api")
	t.Setenv("LAYER_OS_PROVIDERS", "openai")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")

	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateAgentJob(AgentJob{
		JobID:     "job_telegram_002",
		Kind:      "plan",
		Role:      "planner",
		Summary:   "Dispatch the Telegram assistant lane",
		Status:    "queued",
		Source:    "founder.manual",
		Surface:   SurfaceTelegram,
		Stage:     StageDiscover,
		Notes:     []string{},
		CreatedAt: zeroSafeNow(),
		UpdatedAt: zeroSafeNow(),
	}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "job_telegram_002 지금 디스패치해")

	for _, needle := range []string{"dispatch 걸었어: job_telegram_002", "상태: packet_ready", "기록도 남겼어:"} {
		if !strings.Contains(reply, needle) {
			t.Fatalf("expected %q in %q", needle, reply)
		}
	}
	jobs := service.ListAgentJobs()
	if len(jobs) != 1 || jobs[0].Status != "running" {
		t.Fatalf("expected job to move into running, got %+v", jobs)
	}
	if got := len(service.ListProposals()); got != 0 {
		t.Fatalf("expected control DM to avoid auto proposal noise, got %d", got)
	}
}

func TestTelegramBotFounderDMNaturalReviewIntentActsWithoutSlash(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	ref := "verification:telegram_control_002"
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{
		Text:      "Close the Telegram shell gap.",
		Kind:      "automation_failure",
		Severity:  "high",
		Source:    "verification.failed",
		Ref:       &ref,
		CreatedAt: zeroSafeNow(),
		UpdatedAt: zeroSafeNow(),
	}); err != nil {
		t.Fatalf("add review item: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "verification:telegram_control_002 채택해")

	for _, needle := range []string{"채택 처리했어: verification:telegram_control_002", "기록도 남겼어:"} {
		if !strings.Contains(reply, needle) {
			t.Fatalf("expected %q in %q", needle, reply)
		}
	}
	room := service.ReviewRoom()
	if len(room.Open) != 0 || len(room.Accepted) != 1 {
		t.Fatalf("expected review item to move into accepted, got %+v", room)
	}
	if got := len(service.ListProposals()); got != 0 {
		t.Fatalf("expected control DM to avoid auto proposal noise, got %d", got)
	}
}

func TestTelegramBotFounderDMReportsWhereItSavedThePreviousNote(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "텔레그램 founder DM 저장 위치를 명확히 보이게 해줘")

	items := service.ListObservations(ObservationQuery{SourceChannel: "conversation:telegram", Ref: "telegram_chat:7565534667", Limit: 1})
	if len(items) != 1 {
		t.Fatalf("expected one saved observation, got %+v", items)
	}

	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "어디로 남겼는데")
	for _, needle := range []string{"conversation spine observation으로 남아 있어", items[0].ObservationID, "telegram_chat:7565534667"} {
		if !strings.Contains(reply, needle) {
			t.Fatalf("expected %q in %q", needle, reply)
		}
	}
}

func TestTelegramBotFounderDMReviewAddUsesRecentContext(t *testing.T) {
	t.Setenv("LAYER_OS_CONVERSATION_AUTOPLAN", "0")
	t.Setenv("LAYER_OS_CONVERSATION_AUTORISK", "0")

	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "architect verify implementer가 계속 같은 실패패턴을 반복하고 있어")

	items := service.ListObservations(ObservationQuery{SourceChannel: "conversation:telegram", Ref: "telegram_chat:7565534667", Limit: 1})
	if len(items) != 1 {
		t.Fatalf("expected one saved observation before review add, got %+v", items)
	}

	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "리뷰룸에 안건올려")
	if !strings.Contains(reply, "리뷰룸에 올렸어:") {
		t.Fatalf("expected review add reply, got %q", reply)
	}

	room := service.ReviewRoom()
	if len(room.Open) != 1 {
		t.Fatalf("expected one review item, got %+v", room)
	}
	if room.Open[0].Ref == nil || *room.Open[0].Ref != items[0].ObservationID {
		t.Fatalf("expected review item ref to point at recent observation, got %+v", room.Open[0])
	}
	if !strings.Contains(room.Open[0].Text, "architect verify implementer가 계속 같은 실패패턴을 반복하고 있어") {
		t.Fatalf("expected review item to reuse recent context, got %+v", room.Open[0])
	}
}

func TestTelegramBotFounderDMNaturalHarnessLaneQueries(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateProposal(ProposalItem{
		ProposalID: "proposal_telegram_001",
		Title:      "Plan the founder harness lane",
		Intent:     "stabilize founder assistant",
		Summary:    "Founder harness planning lane",
		Surface:    SurfaceTelegram,
		Priority:   "high",
		Risk:       "medium",
		Status:     "proposed",
		Notes:      []string{},
		CreatedAt:  zeroSafeNow(),
		UpdatedAt:  zeroSafeNow(),
	}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	for _, job := range []AgentJob{
		{JobID: "job_planner_001", Kind: "plan", Role: "planner", Summary: "Planner lane summary", Status: "queued", Source: "founder.manual", Surface: SurfaceTelegram, Stage: StageDiscover, Notes: []string{}, CreatedAt: zeroSafeNow(), UpdatedAt: zeroSafeNow()},
		{JobID: "job_designer_001", Kind: "review", Role: "designer", Summary: "Designer lane summary", Status: "running", Source: "founder.manual", Surface: SurfaceTelegram, Stage: StageExperience, Notes: []string{}, CreatedAt: zeroSafeNow(), UpdatedAt: zeroSafeNow()},
		{JobID: "job_implementer_001", Kind: "implement", Role: "implementer", Summary: "Implementer lane summary", Status: "queued", Source: "founder.manual", Surface: SurfaceTelegram, Stage: StageCompose, Notes: []string{}, CreatedAt: zeroSafeNow(), UpdatedAt: zeroSafeNow()},
		{JobID: "job_verifier_001", Kind: "verify", Role: "verifier", Summary: "Verifier lane summary", Status: "failed", Source: "founder.manual", Surface: SurfaceTelegram, Stage: StageVerify, Notes: []string{}, CreatedAt: zeroSafeNow(), UpdatedAt: zeroSafeNow()},
	} {
		if err := service.CreateAgentJob(job); err != nil {
			t.Fatalf("create agent job %s: %v", job.JobID, err)
		}
	}
	if err := service.CreateApproval(ApprovalItem{
		ApprovalID:      "approval_telegram_003",
		WorkItemID:      "work_telegram_003",
		Stage:           StageVerify,
		Summary:         "Approve harness lane",
		Risks:           []string{"lane drift"},
		RollbackPlan:    "hold",
		DecisionSurface: SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     zeroSafeNow(),
	}); err != nil {
		t.Fatalf("create approval: %v", err)
	}
	if _, err := service.AddReviewRoomItem("open", "Verifier lane is blocked."); err != nil {
		t.Fatalf("seed review room: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	cases := []struct {
		message string
		needles []string
	}{
		{"기획 뭐 열렸어", []string{"<b>기획 레인</b>", "proposal_telegram_001", "job_planner_001"}},
		{"디자인 뭐 있어", []string{"<b>디자인 레인</b>", "job_designer_001", "running"}},
		{"구현 뭐 남았어", []string{"<b>구현 레인</b>", "job_implementer_001", "queued"}},
		{"검증 뭐 막혔어", []string{"<b>검증 레인</b>", "job_verifier_001", "failed"}},
		{"풀스택 하네스 팀 상태 보여줘", []string{"<b>풀스택 하네스 팀</b>", "기획: proposal 1 / queued 1 / running 0 / failed 0", "검증: queued 0 / running 0 / failed 1", "승인: pending 1", "리뷰: open 1"}},
	}
	for _, tc := range cases {
		reply := handler.HandleMessageWithContext(TelegramInboundContext{
			ChatID:   7565534667,
			RouteID:  TelegramRouteFounder,
			ChatType: "private",
		}, tc.message)
		for _, needle := range tc.needles {
			if !strings.Contains(reply, needle) {
				t.Fatalf("message %q expected %q in %q", tc.message, needle, reply)
			}
		}
	}

	if got := len(service.ListObservations(ObservationQuery{SourceChannel: "conversation:telegram", Ref: "telegram_chat:7565534667", Limit: 8})); got != 0 {
		t.Fatalf("expected read-only harness queries to avoid conversation noise, got %d observations", got)
	}
}

func TestTelegramBotFounderDMGeminiIntentApprovesPendingCandidate(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateApproval(ApprovalItem{
		ApprovalID:      "approval_telegram_010",
		WorkItemID:      "work_telegram_010",
		Stage:           StageVerify,
		Summary:         "Approve practical aide lane",
		Risks:           []string{"operator drift"},
		RollbackPlan:    "hold",
		DecisionSurface: SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     zeroSafeNow(),
	}); err != nil {
		t.Fatalf("create approval: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	handler.geminiKey = "test-key"
	handler.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if !strings.Contains(string(body), "approval_telegram_010") {
			t.Fatalf("expected intent prompt to include pending approval candidate, got %s", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"{\\\"mode\\\":\\\"write\\\",\\\"action\\\":\\\"approval.approve\\\",\\\"target\\\":\\\"approval_telegram_010\\\",\\\"confidence\\\":\\\"high\\\"}\"}]}}]}")),
			Header:     make(http.Header),
		}, nil
	})}

	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "이건 그냥 진행하자")
	for _, needle := range []string{"승인 처리했어: approval_telegram_010", "기록도 남겼어:"} {
		if !strings.Contains(reply, needle) {
			t.Fatalf("expected %q in %q", needle, reply)
		}
	}
	items := service.ListApprovals()
	if len(items) != 1 || items[0].Status != "approved" {
		t.Fatalf("expected model-routed approval, got %+v", items)
	}
}

func TestTelegramBotFounderDMGeminiIntentReadsPlannerLane(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateProposal(ProposalItem{
		ProposalID: "proposal_telegram_020",
		Title:      "Plan the practical aide lane",
		Intent:     "shape the planner harness",
		Summary:    "Planner harness proposal",
		Surface:    SurfaceTelegram,
		Priority:   "high",
		Risk:       "medium",
		Status:     "proposed",
		Notes:      []string{},
		CreatedAt:  zeroSafeNow(),
		UpdatedAt:  zeroSafeNow(),
	}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	if err := service.CreateAgentJob(AgentJob{
		JobID:     "job_planner_020",
		Kind:      "plan",
		Role:      "planner",
		Summary:   "Planner lane active",
		Status:    "running",
		Source:    "founder.manual",
		Surface:   SurfaceTelegram,
		Stage:     StageDiscover,
		Notes:     []string{},
		CreatedAt: zeroSafeNow(),
		UpdatedAt: zeroSafeNow(),
	}); err != nil {
		t.Fatalf("create planner job: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	handler.geminiKey = "test-key"
	handler.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if !strings.Contains(string(body), "proposal_telegram_020") || !strings.Contains(string(body), "job_planner_020") {
			t.Fatalf("expected planner candidates in intent prompt, got %s", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"{\\\"mode\\\":\\\"read\\\",\\\"action\\\":\\\"read.planner\\\",\\\"confidence\\\":\\\"high\\\"}\"}]}}]}")),
			Header:     make(http.Header),
		}, nil
	})}

	reply := handler.HandleMessageWithContext(TelegramInboundContext{
		ChatID:   7565534667,
		RouteID:  TelegramRouteFounder,
		ChatType: "private",
	}, "앞단 팀부터 보자")
	for _, needle := range []string{"<b>기획 레인</b>", "proposal_telegram_020", "job_planner_020"} {
		if !strings.Contains(reply, needle) {
			t.Fatalf("expected %q in %q", needle, reply)
		}
	}
	if got := len(service.ListObservations(ObservationQuery{SourceChannel: "conversation:telegram", Ref: "telegram_chat:7565534667", Limit: 4})); got != 0 {
		t.Fatalf("expected read intent to avoid conversation persistence noise, got %d observations", got)
	}
}

func TestTelegramBotBlocksControlCommandsOutsideFounderRoute(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	for _, input := range []string{"/approvals", "/approve approval_001", "/dispatch job_001", "/accept review_ref"} {
		reply := handler.HandleMessageWithContext(TelegramInboundContext{
			ChatID:   -10012345,
			RouteID:  TelegramRouteOps,
			ChatType: "supergroup",
		}, input)
		if !strings.Contains(reply, "founder route") {
			t.Fatalf("expected founder-route guard for %q, got %q", input, reply)
		}
	}
}

func TestTelegramBotBuildSystemPromptCarriesPracticalAideGuide(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.ReplaceMemory(SystemMemory{
		CurrentFocus: "Founder route stabilize",
		NextSteps:    []string{"capture founder chat", "separate ops room"},
		OpenRisks:    []string{"tone still awkward"},
		UpdatedAt:    zeroSafeNow(),
	}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}
	if _, err := service.AddReviewRoomItem("open", "Telegram tone patch still pending."); err != nil {
		t.Fatalf("seed review room: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	prompt := handler.buildSystemPrompt()
	for _, needle := range []string{
		"실무형 참모",
		"모르면 모른다고",
		"실제 runtime 결과가 없으면 완료형으로 말하지 말고",
		"번역투",
		"길어도 450자 안쪽",
		"Founder route stabilize",
		"tool budget",
		"personal harness",
	} {
		if !strings.Contains(prompt, needle) {
			t.Fatalf("expected prompt to contain %q, got %q", needle, prompt)
		}
	}
}

func TestTelegramBotChatWithGeminiNormalizesReply(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	handler := NewTelegramBotHandler(service)
	handler.geminiKey = "test-key"
	handler.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if !strings.Contains(string(body), "실무형 참모") {
			t.Fatalf("expected request prompt to include tone guide, got %s", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"### 지금 판단\\n\\n- founder chat부터 잡자\\n- 그다음 ops 방 분리\\n\\n```note\\nignore\\n```\"}]}}]}")),
			Header:     make(http.Header),
		}, nil
	})}

	reply := handler.chatWithGemini("지금 뭐부터 해야 해?", nil)
	expected := "지금 판단\n\n• founder chat부터 잡자\n• 그다음 ops 방 분리"
	if reply != expected {
		t.Fatalf("unexpected normalized reply:\nwant: %q\ngot:  %q", expected, reply)
	}
}
