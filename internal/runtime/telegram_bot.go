package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// TelegramBotHandler handles incoming Telegram messages with Gemini intelligence.
type TelegramBotHandler struct {
	service    *Service
	geminiKey  string
	httpClient *http.Client
}

type telegramFounderConversationAction struct {
	Kind         string
	ApprovalID   string
	ApprovalVerb string
	JobID        string
	ReviewAction string
	ReviewTarget string
	ReviewReason string
	ReviewText   string
}

type telegramFounderIntentResult struct {
	Mode       string `json:"mode"`
	Action     string `json:"action"`
	Target     string `json:"target,omitempty"`
	Text       string `json:"text,omitempty"`
	Reason     string `json:"reason,omitempty"`
	Confidence string `json:"confidence"`
}

func NewTelegramBotHandler(service *Service) *TelegramBotHandler {
	geminiKey, _ := ProviderCredentialValue("gemini")
	return &TelegramBotHandler{
		service:    service,
		geminiKey:  geminiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (h *TelegramBotHandler) Enabled() bool {
	return h.geminiKey != ""
}

// HandleMessage routes an incoming message and returns a reply string.
func (h *TelegramBotHandler) HandleMessage(text string) string {
	return h.HandleMessageWithContext(TelegramInboundContext{RouteID: TelegramRouteFounder}, text)
}

// HandleMessageWithContext routes an incoming Telegram message with chat/route context.
func (h *TelegramBotHandler) HandleMessageWithContext(ctx TelegramInboundContext, text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	ctx = normalizeTelegramInboundContext(ctx)
	founderConversation := telegramFounderConversationAllowed(ctx)

	// Command routing
	lower := strings.ToLower(text)
	switch {
	case lower == "/start":
		return telegramRouteWelcome(ctx, founderConversation)
	case lower == "/whoami" || lower == "/chat" || lower == "/chatid" || lower == "내방":
		return telegramChatIdentityReply(ctx)
	case strings.HasPrefix(lower, "/drop ") || strings.HasPrefix(text, "드롭 "):
		if ctx.RouteID != TelegramRouteFounder {
			return telegramFounderDecisionRequiredReply()
		}
		dropText := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(text, "/drop "), "드롭 "))
		return h.handleSourceIntakeDrop(ctx, dropText)
	case lower == "/intake" || lower == "인박스" || lower == "/source":
		if ctx.RouteID != TelegramRouteFounder {
			return telegramFounderDecisionRequiredReply()
		}
		return h.handleSourceIntakeList()
	case lower == "/drafts" || lower == "/draft" || lower == "초안":
		if ctx.RouteID != TelegramRouteFounder {
			return telegramFounderDecisionRequiredReply()
		}
		return h.handleSourceDraftSeedList()
	case strings.HasPrefix(lower, "/redraft ") || strings.HasPrefix(text, "다듬기 "):
		if ctx.RouteID != TelegramRouteFounder {
			return telegramFounderDecisionRequiredReply()
		}
		rewriteText := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(text, "/redraft "), "다듬기 "))
		return h.handleSourceDraftSeedRedraft(rewriteText)
	case strings.HasPrefix(lower, "/prep "):
		if ctx.RouteID != TelegramRouteFounder {
			return telegramFounderDecisionRequiredReply()
		}
		prepText := strings.TrimSpace(strings.TrimPrefix(text, "/prep "))
		return h.handleSourceDraftSeedPrep(prepText)
	case strings.HasPrefix(lower, "/route ") || strings.HasPrefix(text, "라우트 "):
		if ctx.RouteID != TelegramRouteFounder {
			return telegramFounderDecisionRequiredReply()
		}
		decisionText := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(text, "/route "), "라우트 "))
		return h.handleSourceRouteDecision(ctx, decisionText)
	case lower == "/status" || lower == "상태":
		if !telegramRouteAllowsReadCommands(ctx.RouteID) {
			return telegramRouteNeedsMappingReply()
		}
		return h.handleStatus()
	case lower == "/review" || lower == "리뷰룸":
		if !telegramRouteAllowsReadCommands(ctx.RouteID) {
			return telegramRouteNeedsMappingReply()
		}
		return h.handleReview()
	case lower == "/approvals" || lower == "/approval":
		if ctx.RouteID != TelegramRouteFounder {
			return telegramFounderDecisionRequiredReply()
		}
		return h.handleApprovals()
	case strings.HasPrefix(lower, "/approve "):
		if ctx.RouteID != TelegramRouteFounder {
			return telegramFounderDecisionRequiredReply()
		}
		approvalID := strings.TrimSpace(strings.TrimPrefix(text, "/approve "))
		return h.handleApprovalDecision(approvalID, "approved")
	case strings.HasPrefix(lower, "/reject "):
		if ctx.RouteID != TelegramRouteFounder {
			return telegramFounderDecisionRequiredReply()
		}
		approvalID := strings.TrimSpace(strings.TrimPrefix(text, "/reject "))
		return h.handleApprovalDecision(approvalID, "rejected")
	case lower == "/jobs":
		if !telegramRouteAllowsReadCommands(ctx.RouteID) {
			return telegramRouteNeedsMappingReply()
		}
		return h.handleJobs()
	case strings.HasPrefix(lower, "/dispatch "):
		if ctx.RouteID != TelegramRouteFounder {
			return telegramFounderDecisionRequiredReply()
		}
		jobID := strings.TrimSpace(strings.TrimPrefix(text, "/dispatch "))
		return h.handleDispatch(jobID)
	case strings.HasPrefix(lower, "/accept "):
		if ctx.RouteID != TelegramRouteFounder {
			return telegramFounderDecisionRequiredReply()
		}
		return h.handleReviewAction(ctx, "accept", strings.TrimSpace(strings.TrimPrefix(text, "/accept ")))
	case strings.HasPrefix(lower, "/defer "):
		if ctx.RouteID != TelegramRouteFounder {
			return telegramFounderDecisionRequiredReply()
		}
		return h.handleReviewAction(ctx, "defer", strings.TrimSpace(strings.TrimPrefix(text, "/defer ")))
	case strings.HasPrefix(lower, "/resolve "):
		if ctx.RouteID != TelegramRouteFounder {
			return telegramFounderDecisionRequiredReply()
		}
		return h.handleReviewAction(ctx, "resolve", strings.TrimSpace(strings.TrimPrefix(text, "/resolve ")))
	case lower == "/handoff" || lower == "핸드오프":
		if !founderConversation {
			return telegramFounderConversationRequiredReply(ctx.RouteID)
		}
		return h.handleHandoff()
	case strings.HasPrefix(lower, "/note ") || strings.HasPrefix(text, "메모 "):
		if !founderConversation {
			return telegramFounderConversationRequiredReply(ctx.RouteID)
		}
		note := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(text, "/note "), "메모 "))
		return h.handleNote(ctx, note)
	}

	// Founder room source-like plain text can become intake automatically.
	if ctx.RouteID == TelegramRouteFounder && !founderConversation && telegramShouldAutoIntake(text) {
		return h.handleSourceIntakeDrop(ctx, text)
	}

	// Natural language → Gemini
	if !founderConversation {
		return telegramRouteNaturalLanguageReply(ctx.RouteID)
	}
	return h.handleFounderConversation(ctx, text)
}

func (h *TelegramBotHandler) handleStatus() string {
	knowledge := h.service.Knowledge()
	mem := h.service.Memory()
	review := h.service.ReviewRoomSummary()
	conversation := h.service.ConversationAutomationStatus()

	lines := []string{"<b>지금 상태</b>"}
	if mem.CurrentFocus != "" {
		lines = append(lines, "포커스: "+mem.CurrentFocus)
	}
	if knowledge.PrimaryAction != "" {
		lines = append(lines, "액션: "+telegramActionLabel(knowledge.PrimaryAction))
	}
	if knowledge.PrimaryRef != "" {
		lines = append(lines, "기준: "+knowledge.PrimaryRef)
	}
	lines = append(lines, fmt.Sprintf("리뷰룸 오픈: %d건", review.OpenCount))
	for _, risk := range limitStrings(mem.OpenRisks, 2) {
		lines = append(lines, "리스크: "+risk)
	}
	for _, candidate := range limitParallelCandidates(knowledge.ParallelCandidates, 2) {
		lines = append(lines, "병렬: "+candidate.Text)
	}
	if conversation.RecentFailureCount > 0 {
		lines = append(lines, fmt.Sprintf("대화 자동화 실패: %d건", conversation.RecentFailureCount))
	}
	return strings.Join(lines, "\n")
}

func (h *TelegramBotHandler) handleReview() string {
	review := h.service.ReviewRoomSummary()
	if review.OpenCount == 0 {
		return "지금 열린 리뷰 안건은 없어."
	}
	lines := []string{fmt.Sprintf("<b>지금 열린 리뷰 %d건</b>", review.OpenCount)}
	for _, item := range review.TopOpen {
		lines = append(lines, "• "+html.EscapeString(item.Text))
		if meta := telegramReviewItemMeta(item); meta != "" {
			lines = append(lines, "  "+html.EscapeString(meta))
		}
	}
	if review.OpenCount > len(review.TopOpen) {
		lines = append(lines, fmt.Sprintf("나머지 %d건은 cockpit/admin에서 이어서 보면 돼.", review.OpenCount-len(review.TopOpen)))
	}
	if len(review.TopOpen) > 0 {
		lines = append(lines, "처리는 /accept <ref|text>, /defer <ref|text> :: 이유, /resolve <ref|text> :: 이유 로 하면 돼.")
	}
	return strings.Join(lines, "\n")
}

func (h *TelegramBotHandler) handleApprovals() string {
	pending := telegramPendingApprovals(h.service.ListApprovals())
	if len(pending) == 0 {
		return "지금 대기 중인 승인은 없어."
	}
	lines := []string{fmt.Sprintf("<b>지금 대기 중인 승인 %d건</b>", len(pending))}
	for _, item := range pending {
		lines = append(lines, fmt.Sprintf("• %s · %s", item.ApprovalID, html.EscapeString(limitText(strings.TrimSpace(item.Summary), 88))))
		if len(item.Risks) > 0 {
			lines = append(lines, "  리스크: "+html.EscapeString(strings.Join(limitStrings(item.Risks, 2), ", ")))
		}
	}
	lines = append(lines, "결정은 /approve <approval_id> 또는 /reject <approval_id> 로 남기면 돼.")
	return strings.Join(lines, "\n")
}

func (h *TelegramBotHandler) handleApprovalDecision(approvalID string, status string) string {
	approvalID = telegramFirstField(approvalID)
	if approvalID == "" {
		switch status {
		case "approved":
			return "형식은 /approve <approval_id> 이야."
		default:
			return "형식은 /reject <approval_id> 이야."
		}
	}
	item, err := h.service.ResolveApproval(approvalID, status)
	if err != nil {
		return "승인 처리에 실패했어. " + err.Error()
	}
	switch status {
	case "approved":
		return fmt.Sprintf("승인 처리했어: %s\n%s", item.ApprovalID, html.EscapeString(limitText(strings.TrimSpace(item.Summary), 160)))
	default:
		return fmt.Sprintf("반려 처리했어: %s\n%s\n리뷰룸에도 rejection 안건을 반영했어.", item.ApprovalID, html.EscapeString(limitText(strings.TrimSpace(item.Summary), 160)))
	}
}

func (h *TelegramBotHandler) handleJobs() string {
	jobs := telegramDispatchableJobs(h.service.ListAgentJobs())
	if len(jobs) == 0 {
		return "지금 바로 dispatch할 잡은 없어."
	}
	lines := []string{fmt.Sprintf("<b>지금 dispatch 가능한 잡 %d건</b>", len(jobs))}
	for _, job := range jobs {
		lines = append(lines, fmt.Sprintf("• %s · %s / %s", job.JobID, job.Role, job.Status))
		lines = append(lines, "  "+html.EscapeString(limitText(strings.TrimSpace(job.Summary), 88)))
		if job.Ref != nil && strings.TrimSpace(*job.Ref) != "" {
			lines = append(lines, "  ref: "+html.EscapeString(strings.TrimSpace(*job.Ref)))
		}
	}
	lines = append(lines, "실행은 /dispatch <job_id> 로 걸면 돼.")
	return strings.Join(lines, "\n")
}

func (h *TelegramBotHandler) handleDispatch(jobID string) string {
	jobID = telegramFirstField(jobID)
	if jobID == "" {
		return "형식은 /dispatch <job_id> 이야."
	}
	result, err := h.service.DispatchAgentJob(jobID)
	if err != nil {
		return "dispatch에 실패했어. " + err.Error()
	}
	lines := []string{fmt.Sprintf("dispatch 걸었어: %s", result.Job.JobID)}
	if state := telegramStringResultValue(result.Job.Result, "dispatch_state"); state != "" {
		lines = append(lines, "상태: "+state)
	}
	lines = append(lines, "gateway: "+result.Gateway.Status)
	if state := telegramStringResultValue(result.Job.Result, "dispatch_state"); state == "packet_ready" {
		lines = append(lines, "지금은 packet-ready 상태라 packet-capable worker가 이어받아야 해.")
		if command := telegramStringResultValue(result.Job.Result, "job_packet_command"); command != "" {
			lines = append(lines, "패킷: "+command)
		}
	} else if preview := telegramStringResultValue(result.Job.Result, "response_preview"); preview != "" {
		lines = append(lines, "응답 미리보기: "+html.EscapeString(limitText(preview, 140)))
	}
	return strings.Join(lines, "\n")
}

func (h *TelegramBotHandler) handleReviewAction(ctx TelegramInboundContext, action string, raw string) string {
	target, reason := telegramParseReviewAction(raw)
	if target == "" {
		switch action {
		case "accept":
			return "형식은 /accept <ref|text> 이야."
		case "defer":
			return "형식은 /defer <ref|text> :: 이유 이야."
		default:
			return "형식은 /resolve <ref|text> :: 이유 이야."
		}
	}
	item, ok, problem := telegramFindReviewRoomItem(h.service.ReviewRoom(), target)
	if !ok {
		if problem == "" {
			problem = "그 리뷰 안건은 찾지 못했어. /review로 다시 확인해줘."
		}
		return problem
	}
	resolution := &ReviewRoomResolution{
		Action: action,
		Rule:   "telegram.review." + action,
		Evidence: []string{
			fmt.Sprintf("telegram_chat:%d", ctx.ChatID),
			"telegram_route:" + ctx.RouteID,
		},
	}
	if reason != "" {
		resolution.Reason = reason
	}
	room, err := h.service.TransitionStructuredReviewRoomItem(action, item.Text, resolution)
	if err != nil {
		return "리뷰 안건 처리에 실패했어. " + err.Error()
	}
	label := telegramReviewItemTarget(item)
	lines := []string{fmt.Sprintf("%s 처리했어: %s", telegramReviewActionLabel(action), html.EscapeString(label))}
	if reason != "" {
		lines = append(lines, "이유: "+html.EscapeString(reason))
	}
	lines = append(lines, fmt.Sprintf("현황: open %d / accepted %d / deferred %d", len(room.Open), len(room.Accepted), len(room.Deferred)))
	return strings.Join(lines, "\n")
}

func (h *TelegramBotHandler) handleHandoff() string {
	mem := h.service.Memory()
	lines := []string{"<b>지금 이어받을 내용</b>"}
	if mem.CurrentFocus != "" {
		lines = append(lines, "포커스: "+mem.CurrentFocus)
	}
	for _, step := range limitStrings(mem.NextSteps, 3) {
		lines = append(lines, "다음: "+step)
	}
	if mem.HandoffNote != nil && *mem.HandoffNote != "" {
		lines = append(lines, "노트: "+*mem.HandoffNote)
	}
	return strings.Join(lines, "\n")
}

func (h *TelegramBotHandler) handleSourceIntakeDrop(ctx TelegramInboundContext, text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "형식은 /drop <링크나 텍스트> 이야. 먼저 raw 소스를 넣고, 그다음 /intake 와 /route 로 이어가자."
	}

	record := NormalizeSourceIntakeRecord(SourceIntakeRecord{
		IntakeClass:     "manual_drop",
		PolicyColor:     "green",
		URL:             ExtractFirstSourceURL(text),
		Excerpt:         text,
		SuggestedRoutes: []string{"97layer"},
	})
	if record.URL == "" {
		record.Title = sourceIntakeLimitText(strings.Join(strings.Fields(text), " "), 72)
	}

	refs := append(SourceIntakeRefs(record),
		fmt.Sprintf("telegram_chat:%d", ctx.ChatID),
		"telegram_route:"+ctx.RouteID,
	)
	if value := strings.TrimSpace(ctx.ChatType); value != "" {
		refs = append(refs, "telegram_chat_type:"+value)
	}
	created, err := h.service.CreateObservation(ObservationRecord{
		Topic:             SourceIntakeTopic,
		SourceChannel:     "telegram",
		Actor:             telegramRouteActor(ctx.RouteID),
		Refs:              refs,
		Confidence:        "high",
		RawExcerpt:        BuildSourceIntakeRawExcerpt(record),
		NormalizedSummary: SourceIntakeSummary(record),
		CapturedAt:        zeroSafeNow(),
	})
	if err != nil {
		return "source intake 저장이 안 됐어. " + err.Error()
	}

	subject := telegramSourceIntakeSubject(created)
	return fmt.Sprintf(
		"source intake로 넣었어: %s\nid: %s\n다음은 /route %s 97layer 처럼 보내면 돼.",
		html.EscapeString(subject),
		created.ObservationID,
		created.ObservationID,
	)
}

func (h *TelegramBotHandler) handleSourceIntakeList() string {
	items := telegramPendingSourceIntakeItems(
		h.service.ListObservations(ObservationQuery{Topic: SourceIntakeTopic, Limit: 12}),
		h.service.ListObservations(ObservationQuery{Topic: IntakeRouteDecisionTopic, Limit: 48}),
	)
	if len(items) == 0 {
		return "지금 route 결정 대기 중인 source intake는 없어. founder 방에서 /drop <링크나 텍스트> 로 먼저 넣을 수 있어."
	}

	lines := []string{"<b>최근 source intake</b>"}
	for _, item := range items {
		subject := html.EscapeString(telegramSourceIntakeSubject(item.Observation))
		routes := telegramSourceIntakeRouteLabels(item.SuggestedRoutes)
		lines = append(lines, fmt.Sprintf("• %s", subject))
		lines = append(lines, fmt.Sprintf("  %s", item.Observation.ObservationID))
		if routes != "" {
			lines = append(lines, fmt.Sprintf("  제안 route: %s", routes))
		}
	}
	lines = append(lines, "결정은 /route <observation_id> <97layer|woosunhokr|woohwahae|hold> 로 남기면 돼.")
	return strings.Join(lines, "\n")
}

func (h *TelegramBotHandler) handleSourceDraftSeedList() string {
	items := telegramRecentSourceDraftSeedItems(h.service.ListObservations(ObservationQuery{Topic: SourceDraftSeedTopic, Limit: 12}))
	if len(items) == 0 {
		return "지금 열려 있는 route 초안은 없어. founder 방에서 source를 넣고 /route로 먼저 방향을 정하면 여기부터 쌓여."
	}

	lines := []string{"<b>최근 route 초안</b>"}
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("• %s · %s", html.EscapeString(item.TargetLabel), html.EscapeString(item.Title)))
		lines = append(lines, fmt.Sprintf("  %s", item.ObservationID))
		if item.Preview != "" {
			lines = append(lines, fmt.Sprintf("  %s", html.EscapeString(item.Preview)))
		}
	}
	lines = append(lines, "Threads prep은 /prep <draft_observation_id> 로 열면 돼.")
	return strings.Join(lines, "\n")
}

func (h *TelegramBotHandler) handleSourceDraftSeedRedraft(raw string) string {
	parts := strings.Fields(strings.TrimSpace(raw))
	if len(parts) < 2 {
		return "형식은 /redraft <draft_observation_id> <메모> 이야. 예: /redraft observation_123 조금 더 구체적으로"
	}
	draftObservationID := strings.TrimSpace(parts[0])
	instruction := strings.TrimSpace(strings.Join(parts[1:], " "))
	draftObservation, ok := telegramFindSourceDraftSeedObservation(h.service, draftObservationID)
	if !ok {
		return "그 draft seed는 최근 목록에서 못 찾았어. /drafts로 다시 확인해줘."
	}
	revised, err := ReviseSourceDraftSeedObservation(h.service, draftObservation, instruction)
	if err != nil {
		return "초안을 다시 여는 데 실패했어. " + err.Error()
	}
	revisedRecord := ParseSourceDraftSeedRawExcerpt(revised.RawExcerpt)
	return fmt.Sprintf(
		"초안 다시 열었어: %s\nid: %s\n초안 미리보기: %s",
		html.EscapeString(sourceDraftSeedTargetLabel(revisedRecord.TargetAccount)),
		revised.ObservationID,
		html.EscapeString(sourceDraftSeedPreview(revisedRecord.Draft)),
	)
}

func (h *TelegramBotHandler) handleSourceDraftSeedPrep(raw string) string {
	parts := strings.Fields(strings.TrimSpace(raw))
	if len(parts) == 0 {
		return "형식은 /prep <draft_observation_id> 이야. 예: /prep observation_123"
	}
	draftObservationID := strings.TrimSpace(parts[0])
	channel := "threads"
	if len(parts) > 1 {
		channel = strings.TrimSpace(parts[1])
	}
	result, err := h.service.OpenSourceDraftSeedPublishPrep(draftObservationID, channel)
	if err != nil {
		return "prep 작업을 여는 데 실패했어. " + err.Error()
	}
	status := "Threads prep 열었어"
	if result.Reused {
		status = "기존 Threads prep 이어서 잡았어"
	}
	return fmt.Sprintf(
		"%s: %s\napproval: %s\nflow: %s\n본문 미리보기: %s",
		status,
		html.EscapeString(result.Lane.TargetAccountLabel),
		result.Approval.ApprovalID,
		result.Flow.FlowID,
		html.EscapeString(result.Lane.BodyPreview),
	)
}

func (h *TelegramBotHandler) handleSourceRouteDecision(ctx TelegramInboundContext, decisionText string) string {
	parts := strings.Fields(strings.TrimSpace(decisionText))
	if len(parts) < 2 {
		return "형식은 /route <observation_id> <97layer|woosunhokr|woohwahae|hold> 이야."
	}
	observationID := strings.TrimSpace(parts[0])
	decision := normalizeSourceIntakeRouteChoice(parts[1])
	if decision == "" {
		return "route 값은 97layer, woosunhokr, woohwahae, hold 중 하나로 보내줘."
	}

	item, ok := telegramFindSourceIntakeObservation(h.service, observationID)
	if !ok {
		return "그 source intake는 최근 목록에서 못 찾았어. /intake로 다시 확인해줘."
	}

	subject := telegramSourceIntakeSubject(item.Observation)
	rawExcerpt := strings.Join([]string{
		"source_observation_id=" + item.Observation.ObservationID,
		"decision=" + decision,
		"title=" + item.Title,
		"summary=" + item.Observation.NormalizedSummary,
		"telegram_chat_id=" + strconv.FormatInt(ctx.ChatID, 10),
		"telegram_route=" + ctx.RouteID,
	}, "\n")
	created, err := h.service.CreateObservation(ObservationRecord{
		Topic:         IntakeRouteDecisionTopic,
		SourceChannel: "telegram",
		Actor:         telegramRouteActor(ctx.RouteID),
		Refs: []string{
			item.Observation.ObservationID,
			"decision_route:" + decision,
			fmt.Sprintf("telegram_chat:%d", ctx.ChatID),
			"telegram_route:" + ctx.RouteID,
		},
		Confidence:        "high",
		RawExcerpt:        rawExcerpt,
		NormalizedSummary: fmt.Sprintf("Intake route decided · %s -> %s", subject, SourceIntakeRouteChoiceLabel(decision)),
		CapturedAt:        zeroSafeNow(),
	})
	if err != nil {
		return "route 결정 저장이 안 됐어. " + err.Error()
	}

	if decision == "hold" {
		return fmt.Sprintf("hold로 남겼어: %s\n기록: %s", html.EscapeString(subject), created.ObservationID)
	}
	draftSeed, draftOpened, err := EnsureSourceDraftSeedObservation(h.service, item.Observation, created.ObservationID, decision)
	if err != nil {
		return "route 결정은 남겼지만 초안 본문을 여는 데 실패했어. " + err.Error()
	}
	draftPreview := sourceDraftSeedPreview(ParseSourceDraftSeedRawExcerpt(draftSeed.RawExcerpt).Draft)
	proposal, opened, err := EnsureSourceDraftSeedProposal(h.service, item.Observation, created.ObservationID, decision)
	if err != nil {
		return "route 결정은 남겼지만 draft seed를 여는 데 실패했어. " + err.Error()
	}
	if !opened {
		prefix := "기존"
		if draftOpened {
			prefix = "새 초안 + 기존"
		}
		return fmt.Sprintf(
			"route 결정 남겼어: %s -> %s\n%s draft seed: %s\n초안 미리보기: %s\nprep: /prep %s",
			html.EscapeString(subject),
			html.EscapeString(SourceIntakeRouteChoiceLabel(decision)),
			prefix,
			proposal.ProposalID,
			html.EscapeString(draftPreview),
			draftSeed.ObservationID,
		)
	}
	return fmt.Sprintf(
		"route 결정 남겼어: %s -> %s\nproposal: %s\n초안 미리보기: %s\nprep: /prep %s",
		html.EscapeString(subject),
		html.EscapeString(SourceIntakeRouteChoiceLabel(decision)),
		proposal.ProposalID,
		html.EscapeString(draftPreview),
		draftSeed.ObservationID,
	)
}

func (h *TelegramBotHandler) handleNote(ctx TelegramInboundContext, note string) string {
	if note == "" {
		return "메모 내용을 입력해줘. 예: /note 오늘 결정사항"
	}
	refs := telegramConversationRefs(ctx)
	capturedAt := zeroSafeNow()
	if _, err := h.service.CreateConversationNote(ConversationNoteInput{
		Text:          note,
		Actor:         telegramRouteActor(ctx.RouteID),
		SourceChannel: "telegram",
		Summary:       note,
		Refs:          refs,
		Confidence:    "high",
		CapturedAt:    &capturedAt,
	}); err != nil {
		return "메모 저장이 안 됐어. " + err.Error()
	}
	return "메모로 남겼어: " + note
}

func (h *TelegramBotHandler) handleFounderConversation(ctx TelegramInboundContext, text string) string {
	if reply, ok := h.handleFounderConversationMetaQuery(ctx, text); ok {
		return reply
	}
	if reply, ok := h.handleFounderConversationModelIntent(ctx, text); ok {
		return reply
	}
	if action, ok := telegramParseFounderConversationAction(text); ok {
		reply := h.handleFounderConversationAction(ctx, action)
		return h.appendFounderControlLog(ctx, text, reply)
	}
	if reply, ok := h.handleFounderConversationReadQuery(text); ok {
		return reply
	}

	result, err := h.createFounderConversationNote(ctx, text)
	if err != nil {
		if !h.Enabled() {
			return "대화 저장이 안 됐어. " + err.Error()
		}
		reply := h.chatWithGemini(userMessageWithTrimmedSpace(text), nil)
		if strings.TrimSpace(reply) == "" {
			return "대화 저장이 안 됐어. " + err.Error()
		}
		return reply + "\n\n주의: 이번 founder DM은 runtime conversation에 남기지 못했어. " + err.Error()
	}

	assistant := ""
	if h.Enabled() {
		assistant = h.chatWithGemini(userMessageWithTrimmedSpace(text), &result)
	}
	summary := telegramFounderConversationSummary(result, !h.Enabled())
	if strings.TrimSpace(assistant) == "" {
		return summary
	}
	if strings.TrimSpace(summary) == "" {
		return assistant
	}
	return assistant + "\n\n" + summary
}

func (h *TelegramBotHandler) handleFounderConversationMetaQuery(ctx TelegramInboundContext, text string) (string, bool) {
	if !telegramAsksWhereConversationWasSaved(text) {
		return "", false
	}
	observation, ok := telegramLatestFounderConversationObservation(h.service, ctx, nil)
	if !ok {
		return "방금 founder DM 저장 기록은 아직 못 찾았어. 이번 턴부터는 observation id까지 같이 답하게 맞추고 있어.", true
	}
	lines := []string{"최근 founder DM은 conversation spine observation으로 남아 있어: " + observation.ObservationID}
	if summary := strings.TrimSpace(observation.NormalizedSummary); summary != "" {
		lines = append(lines, "요약: "+html.EscapeString(limitText(summary, 160)))
	}
	if refs := telegramConversationRefSummary(observation.Refs); refs != "" {
		lines = append(lines, refs)
	}
	return strings.Join(lines, "\n"), true
}

func (h *TelegramBotHandler) handleFounderConversationModelIntent(ctx TelegramInboundContext, text string) (string, bool) {
	intent, ok := h.inferFounderIntentWithGemini(text)
	if !ok {
		return "", false
	}
	switch strings.TrimSpace(intent.Action) {
	case "read.status":
		return h.handleStatus(), true
	case "read.review":
		return h.handleReview(), true
	case "read.approvals":
		return h.handleApprovals(), true
	case "read.jobs":
		return h.handleJobs(), true
	case "read.team":
		return h.handleHarnessTeamStatus(), true
	case "read.planner":
		return h.handleHarnessLaneStatus("planner"), true
	case "read.designer":
		return h.handleHarnessLaneStatus("designer"), true
	case "read.implementer":
		return h.handleHarnessLaneStatus("implementer"), true
	case "read.verifier":
		return h.handleHarnessLaneStatus("verifier"), true
	case "approval.approve":
		if strings.TrimSpace(intent.Target) == "" {
			return "", false
		}
		return h.appendFounderControlLog(ctx, text, h.handleApprovalDecision(intent.Target, "approved")), true
	case "approval.reject":
		if strings.TrimSpace(intent.Target) == "" {
			return "", false
		}
		return h.appendFounderControlLog(ctx, text, h.handleApprovalDecision(intent.Target, "rejected")), true
	case "job.dispatch":
		if strings.TrimSpace(intent.Target) == "" {
			return "", false
		}
		return h.appendFounderControlLog(ctx, text, h.handleDispatch(intent.Target)), true
	case "review.accept":
		if strings.TrimSpace(intent.Target) == "" {
			return "", false
		}
		return h.appendFounderControlLog(ctx, text, h.handleReviewAction(ctx, "accept", telegramJoinTargetAndReason(intent.Target, intent.Reason))), true
	case "review.defer":
		if strings.TrimSpace(intent.Target) == "" {
			return "", false
		}
		return h.appendFounderControlLog(ctx, text, h.handleReviewAction(ctx, "defer", telegramJoinTargetAndReason(intent.Target, intent.Reason))), true
	case "review.resolve":
		if strings.TrimSpace(intent.Target) == "" {
			return "", false
		}
		return h.appendFounderControlLog(ctx, text, h.handleReviewAction(ctx, "resolve", telegramJoinTargetAndReason(intent.Target, intent.Reason))), true
	case "review.add":
		return h.appendFounderControlLog(ctx, text, h.handleReviewAdd(ctx, intent.Text)), true
	default:
		return "", false
	}
}

func (h *TelegramBotHandler) handleFounderConversationReadQuery(text string) (string, bool) {
	lower := strings.ToLower(strings.TrimSpace(text))
	switch {
	case telegramAsksHarnessTeamStatus(lower):
		return h.handleHarnessTeamStatus(), true
	case telegramAsksHarnessLaneStatus(lower, "planner"):
		return h.handleHarnessLaneStatus("planner"), true
	case telegramAsksHarnessLaneStatus(lower, "designer"):
		return h.handleHarnessLaneStatus("designer"), true
	case telegramAsksHarnessLaneStatus(lower, "implementer"):
		return h.handleHarnessLaneStatus("implementer"), true
	case telegramAsksHarnessLaneStatus(lower, "verifier"):
		return h.handleHarnessLaneStatus("verifier"), true
	case telegramAsksApprovalsRead(lower):
		return h.handleApprovals(), true
	case telegramAsksJobsRead(lower):
		return h.handleJobs(), true
	case telegramAsksReviewRead(lower):
		return h.handleReview(), true
	case telegramAsksStatusRead(lower):
		return h.handleStatus(), true
	default:
		return "", false
	}
}

func (h *TelegramBotHandler) appendFounderControlLog(ctx TelegramInboundContext, text string, reply string) string {
	autoDisabled := false
	result, err := h.createFounderConversationNoteWithOverrides(ctx, text, &autoDisabled, &autoDisabled, &autoDisabled)
	if err != nil {
		return reply + "\n\n주의: 이번 founder DM은 runtime conversation에 남기지 못했어. " + err.Error()
	}
	note := telegramFounderControlNote(result)
	if strings.TrimSpace(note) == "" {
		return reply
	}
	return reply + "\n\n" + note
}

func userMessageWithTrimmedSpace(value string) string {
	return strings.TrimSpace(value)
}

func (h *TelegramBotHandler) createFounderConversationNote(ctx TelegramInboundContext, text string) (ConversationNoteResult, error) {
	return h.createFounderConversationNoteWithOverrides(ctx, text, nil, nil, nil)
}

func (h *TelegramBotHandler) createFounderConversationNoteWithOverrides(ctx TelegramInboundContext, text string, autoPlan *bool, autoRisk *bool, autoDispatch *bool) (ConversationNoteResult, error) {
	capturedAt := zeroSafeNow()
	return h.service.CreateConversationNote(ConversationNoteInput{
		Text:          text,
		Actor:         telegramRouteActor(ctx.RouteID),
		SourceChannel: "telegram",
		Refs:          telegramConversationRefs(ctx),
		Confidence:    "high",
		CapturedAt:    &capturedAt,
		AutoPlan:      autoPlan,
		AutoRisk:      autoRisk,
		AutoDispatch:  autoDispatch,
	})
}

func (h *TelegramBotHandler) handleFounderConversationAction(ctx TelegramInboundContext, action telegramFounderConversationAction) string {
	switch action.Kind {
	case "approval":
		return h.handleApprovalDecision(action.ApprovalID, action.ApprovalVerb)
	case "dispatch":
		return h.handleDispatch(action.JobID)
	case "review":
		raw := action.ReviewTarget
		if strings.TrimSpace(action.ReviewReason) != "" {
			raw += " :: " + action.ReviewReason
		}
		return h.handleReviewAction(ctx, action.ReviewAction, raw)
	case "review_add":
		return h.handleReviewAdd(ctx, action.ReviewText)
	default:
		return ""
	}
}

func (h *TelegramBotHandler) handleReviewAdd(ctx TelegramInboundContext, itemText string) string {
	text, ref := telegramResolveReviewAddText(h.service, ctx, itemText)
	if text == "" {
		return "리뷰룸에 올릴 안건 본문이 아직 비어 있어. `리뷰룸에 안건 올려 :: <내용>` 처럼 보내주면 바로 반영할게."
	}
	item := newManualReviewRoomItem(text, "agenda", "medium", "telegram.manual", ref)
	room, err := h.service.AddStructuredReviewRoomItem("open", item)
	if err != nil {
		return "리뷰룸 안건 추가에 실패했어. " + err.Error()
	}
	lines := []string{"리뷰룸에 올렸어: " + html.EscapeString(text)}
	if ref != nil && strings.TrimSpace(*ref) != "" {
		lines = append(lines, "근거 ref: "+html.EscapeString(strings.TrimSpace(*ref)))
	}
	lines = append(lines, fmt.Sprintf("현황: open %d / accepted %d / deferred %d", len(room.Open), len(room.Accepted), len(room.Deferred)))
	return strings.Join(lines, "\n")
}

func (h *TelegramBotHandler) handleHarnessTeamStatus() string {
	proposals := telegramOpenProposals(h.service.ListProposals())
	jobs := h.service.ListAgentJobs()
	review := h.service.ReviewRoomSummary()
	approvals := telegramPendingApprovals(h.service.ListApprovals())

	lines := []string{"<b>풀스택 하네스 팀</b>"}
	lines = append(lines, telegramHarnessLaneCountLine("기획", len(proposals), jobs, "planner"))
	lines = append(lines, telegramHarnessLaneCountLine("디자인", 0, jobs, "designer"))
	lines = append(lines, telegramHarnessLaneCountLine("구현", 0, jobs, "implementer"))
	lines = append(lines, telegramHarnessLaneCountLine("검증", 0, jobs, "verifier"))
	lines = append(lines, fmt.Sprintf("승인: pending %d", len(approvals)))
	lines = append(lines, fmt.Sprintf("리뷰: open %d", review.OpenCount))
	return strings.Join(lines, "\n")
}

func (h *TelegramBotHandler) handleHarnessLaneStatus(role string) string {
	proposals := telegramOpenProposals(h.service.ListProposals())
	jobs := telegramJobsForRole(h.service.ListAgentJobs(), role)
	label := telegramHarnessLaneLabel(role)
	lines := []string{fmt.Sprintf("<b>%s 레인</b>", label)}

	if role == "planner" {
		open := proposals
		if len(open) == 0 {
			lines = append(lines, "열린 제안은 없어.")
		} else {
			lines = append(lines, fmt.Sprintf("열린 제안 %d건", len(open)))
			for _, item := range limitProposalItems(open, 3) {
				lines = append(lines, fmt.Sprintf("• %s · %s", item.ProposalID, html.EscapeString(limitText(strings.TrimSpace(item.Summary), 88))))
			}
		}
	}

	if len(jobs) == 0 {
		lines = append(lines, "열린 잡은 없어.")
		return strings.Join(lines, "\n")
	}
	lines = append(lines, fmt.Sprintf("활성 잡 %d건", len(jobs)))
	for _, job := range jobs {
		lines = append(lines, fmt.Sprintf("• %s · %s", job.JobID, job.Status))
		lines = append(lines, "  "+html.EscapeString(limitText(strings.TrimSpace(job.Summary), 88)))
	}
	return strings.Join(lines, "\n")
}

func telegramConversationRefs(ctx TelegramInboundContext) []string {
	refs := []string{
		fmt.Sprintf("telegram_chat:%d", ctx.ChatID),
		"telegram_route:" + ctx.RouteID,
	}
	if value := strings.TrimSpace(ctx.ChatType); value != "" {
		refs = append(refs, "telegram_chat_type:"+value)
	}
	return refs
}

func telegramFounderConversationSummary(result ConversationNoteResult, includeEngineNote bool) string {
	lines := []string{"conversation spine에 남겼어: " + result.Observation.ObservationID}
	if strings.TrimSpace(result.ConversationID) != "" {
		lines = append(lines, "대화 ID: "+strings.TrimSpace(result.ConversationID))
	}
	if result.ReviewItem != nil {
		lines = append(lines, "리스크 안건도 열어뒀어: "+result.ReviewItem.Text)
	}
	switch {
	case result.Proposal != nil && result.Job != nil:
		lines = append(lines, fmt.Sprintf("planner 레인도 붙였어: %s / %s", result.Proposal.ProposalID, result.Job.JobID))
	case result.Job != nil:
		lines = append(lines, "작업 레인도 열어뒀어: "+result.Job.JobID)
	case result.Proposal != nil:
		lines = append(lines, "제안도 열어뒀어: "+result.Proposal.ProposalID)
	}
	if result.Dispatch != nil {
		lines = append(lines, "자동 dispatch까지 이어졌어.")
	}
	if len(result.Warnings) > 0 {
		lines = append(lines, "주의: "+limitText(strings.TrimSpace(result.Warnings[0]), 140))
	}
	if includeEngineNote {
		lines = append(lines, "지금은 답변 엔진이 꺼져 있어서 판단 코멘트는 붙이지 않았어.")
	}
	return strings.Join(lines, "\n")
}

func telegramFounderControlNote(result ConversationNoteResult) string {
	lines := []string{"기록도 남겼어: " + result.Observation.ObservationID}
	if strings.TrimSpace(result.ConversationID) != "" {
		lines = append(lines, "대화 ID: "+strings.TrimSpace(result.ConversationID))
	}
	if len(result.Warnings) > 0 {
		lines = append(lines, "주의: "+limitText(strings.TrimSpace(result.Warnings[0]), 140))
	}
	return strings.Join(lines, "\n")
}

func telegramJoinTargetAndReason(target string, reason string) string {
	target = strings.TrimSpace(target)
	reason = strings.TrimSpace(reason)
	if target == "" {
		return ""
	}
	if reason == "" {
		return target
	}
	return target + " :: " + reason
}

func (h *TelegramBotHandler) chatWithGemini(userMessage string, capture *ConversationNoteResult) string {
	if !h.Enabled() {
		return "지금은 답변 엔진이 꺼져 있어. `/status` 나 `/review` 는 바로 볼 수 있어."
	}

	systemPrompt := h.buildSystemPrompt()
	fullPrompt := systemPrompt
	if capture != nil {
		fullPrompt += "\n\n=== 이번 founder DM 처리 결과 ===\n" + telegramConversationAutomationContext(*capture)
	}
	fullPrompt += "\n\n창업자 메시지: " + userMessage
	text, err := h.geminiGenerateText(fullPrompt)
	if err != nil {
		return telegramAssistantError("지금 답변 엔진 연결이 불안정해. 잠시 뒤 다시 보내줘.", err.Error())
	}
	return normalizeTelegramAssistantReply(text)
}

func (h *TelegramBotHandler) inferFounderIntentWithGemini(text string) (telegramFounderIntentResult, bool) {
	if !h.Enabled() {
		return telegramFounderIntentResult{}, false
	}
	raw, err := h.geminiGenerateText(h.buildFounderIntentPrompt(text))
	if err != nil {
		return telegramFounderIntentResult{}, false
	}
	intent, err := parseTelegramFounderIntent(raw)
	if err != nil {
		return telegramFounderIntentResult{}, false
	}
	if !telegramFounderIntentUsable(intent) {
		return telegramFounderIntentResult{}, false
	}
	return intent, true
}

func (h *TelegramBotHandler) buildFounderIntentPrompt(text string) string {
	openProposals := telegramOpenProposals(h.service.ListProposals())
	pendingApprovals := telegramPendingApprovals(h.service.ListApprovals())
	activeJobs := telegramActiveHarnessJobs(h.service.ListAgentJobs())
	dispatchableJobs := telegramDispatchableJobs(h.service.ListAgentJobs())
	review := h.service.ReviewRoomSummary()

	var b strings.Builder
	b.WriteString("너는 Layer OS founder DM intent router다.\n")
	b.WriteString("자유문장을 canonical runtime read/write intent로 분류한다.\n")
	b.WriteString("target id/ref는 후보 목록에 있는 값만 써라. 모르면 invent하지 말고 none으로 돌려.\n")
	b.WriteString("확신이 낮으면 mode=chat, action=none 으로 반환해.\n")
	b.WriteString("반환은 JSON 한 개만 해라.\n")
	b.WriteString("allowed actions: read.status, read.review, read.approvals, read.jobs, read.team, read.planner, read.designer, read.implementer, read.verifier, approval.approve, approval.reject, job.dispatch, review.accept, review.defer, review.resolve, review.add, none\n")
	b.WriteString("schema: {\"mode\":\"read|write|chat\",\"action\":\"...\",\"target\":\"optional exact id/ref\",\"text\":\"optional\",\"reason\":\"optional\",\"confidence\":\"low|medium|high\"}\n")
	b.WriteString("\nopen proposals:\n")
	for _, item := range limitProposalItems(openProposals, 6) {
		b.WriteString("- " + item.ProposalID + " | " + strings.TrimSpace(item.Summary) + "\n")
	}
	b.WriteString("\nactive harness jobs:\n")
	for _, item := range activeJobs {
		b.WriteString("- " + item.JobID + " | " + item.Role + " | " + item.Status + " | " + strings.TrimSpace(item.Summary) + "\n")
	}
	b.WriteString("\npending approvals:\n")
	for _, item := range pendingApprovals {
		b.WriteString("- " + item.ApprovalID + " | " + strings.TrimSpace(item.Summary) + "\n")
	}
	b.WriteString("\ndispatchable jobs:\n")
	for _, item := range dispatchableJobs {
		b.WriteString("- " + item.JobID + " | " + item.Role + " | " + strings.TrimSpace(item.Summary) + "\n")
	}
	b.WriteString("\nopen review items:\n")
	for _, item := range review.TopOpen {
		target := item.Text
		if item.Ref != nil && strings.TrimSpace(*item.Ref) != "" {
			target = strings.TrimSpace(*item.Ref)
		}
		b.WriteString("- " + target + " | " + strings.TrimSpace(item.Text) + "\n")
	}
	b.WriteString("\nfounder message:\n")
	b.WriteString(text)
	return b.String()
}

func (h *TelegramBotHandler) geminiGenerateText(prompt string) (string, error) {
	type part struct {
		Text string `json:"text"`
	}
	type content struct {
		Parts []part `json:"parts"`
	}
	type reqBody struct {
		Contents []content `json:"contents"`
	}

	body, _ := json.Marshal(reqBody{
		Contents: []content{{Parts: []part{{Text: prompt}}}},
	})

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + h.geminiKey
	resp, err := h.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	if result.Error.Message != "" {
		return "", fmt.Errorf("%s", result.Error.Message)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty candidate")
	}
	return result.Candidates[0].Content.Parts[0].Text, nil
}

func (h *TelegramBotHandler) buildSystemPrompt() string {
	knowledge := h.service.Knowledge()
	mem := h.service.Memory()
	review := h.service.ReviewRoomSummary()

	var b strings.Builder
	b.WriteString("너는 Layer OS의 Telegram 비서가 아니라, 창업자 옆에서 바로 판단을 도와주는 실무형 참모다.\n")
	b.WriteString("말투는 자연스러운 한국어로 짧고 또렷하게 유지해.\n")
	b.WriteString("첫 문장은 바로 답이나 판단으로 시작하고, 필요할 때만 짧은 bullet 2~4개를 붙여.\n")
	b.WriteString("과장하거나 아는 척하지 말고, 지금 주어진 시스템 상태에 없는 사실은 추측하지 마.\n")
	b.WriteString("모르면 모른다고 말하고 다음 확인 행동을 한 줄로 제안해.\n")
	b.WriteString("실제 runtime 결과가 없으면 완료형으로 말하지 말고, 제안인지 실행 결과인지 분명히 구분해.\n")
	b.WriteString("시스템 용어, 헤더 남발, 번역투, 장황한 서론, 코드블록은 피하고 사람 대화처럼 답해.\n\n")
	b.WriteString("내부 액션 id(review_room, rollback_or_fix 같은 토큰)를 그대로 반복하지 말고 자연스러운 한국어 행동으로 풀어써.\n")
	b.WriteString("아래 시스템 상태는 배경 정보야. 창업자 메시지가 시스템 운영과 직접 관련 없는 주제(콘텐츠, 브랜드, 아이디어, 일상 대화 등)면 그 주제 자체로 자연스럽게 응답해. 시스템 상태를 모든 답변에 끼워 넣지 마.\n\n")
	b.WriteString("운영 원칙은 세 가지야.\n")
	b.WriteString("1. 모델 비교보다 세팅 품질, lane fit, tool budget이 더 큰 차이를 만든다.\n")
	b.WriteString("2. 무의미한 MCP나 긴 컨텍스트로 토큰과 rate limit를 낭비하지 말고 꼭 필요한 도구만 남긴다.\n")
	b.WriteString("3. 반복되는 요청과 실패 패턴은 personal harness처럼 축적해서 다음 판단 품질을 높인다.\n\n")
	b.WriteString("=== 현재 시스템 상태 ===\n")

	if mem.CurrentFocus != "" {
		b.WriteString("현재 포커스: " + mem.CurrentFocus + "\n")
	}
	if knowledge.PrimaryAction != "" {
		b.WriteString("우선 액션: " + telegramActionLabel(knowledge.PrimaryAction) + "\n")
	}
	b.WriteString(fmt.Sprintf("리뷰룸 오픈: %d건\n", review.OpenCount))
	for _, item := range topReviewTexts(review.TopOpen, 2) {
		b.WriteString("  안건: " + item + "\n")
	}
	for _, candidate := range limitParallelCandidates(knowledge.ParallelCandidates, 3) {
		b.WriteString("병렬 후보: " + candidate.Text + "\n")
	}
	for _, risk := range limitStrings(mem.OpenRisks, 3) {
		b.WriteString("리스크: " + risk + "\n")
	}
	for _, step := range limitStrings(mem.NextSteps, 3) {
		b.WriteString("다음 단계: " + step + "\n")
	}
	if mem.HandoffNote != nil && *mem.HandoffNote != "" {
		b.WriteString("핸드오프 노트: " + *mem.HandoffNote + "\n")
	}
	b.WriteString("\n답변은 기본적으로 2~5문장 또는 짧은 bullet 위주로, 길어도 450자 안쪽으로 유지해.")
	return b.String()
}

func telegramConversationAutomationContext(result ConversationNoteResult) string {
	lines := []string{"conversation_id: " + strings.TrimSpace(result.ConversationID)}
	if result.Proposal != nil {
		lines = append(lines, "proposal_id: "+strings.TrimSpace(result.Proposal.ProposalID))
	}
	if result.Job != nil {
		lines = append(lines, "job_id: "+strings.TrimSpace(result.Job.JobID))
		lines = append(lines, "job_status: "+strings.TrimSpace(result.Job.Status))
	}
	if result.Dispatch != nil {
		if result.Dispatch.Job.Result != nil {
			if state, ok := result.Dispatch.Job.Result["dispatch_state"].(string); ok && strings.TrimSpace(state) != "" {
				lines = append(lines, "dispatch_state: "+strings.TrimSpace(state))
			}
		}
	}
	if result.ReviewItem != nil {
		lines = append(lines, "review_item: "+strings.TrimSpace(result.ReviewItem.Text))
	}
	for _, warning := range limitStrings(result.Warnings, 2) {
		lines = append(lines, "warning: "+strings.TrimSpace(warning))
	}
	return strings.Join(lines, "\n")
}

func telegramAssistantError(summary string, detail string) string {
	lower := strings.ToLower(strings.TrimSpace(detail))
	switch {
	case strings.Contains(lower, "quota"), strings.Contains(lower, "rate"):
		return "지금 답변 엔진 사용 한도에 걸렸어. 조금 있다가 다시 보내줘."
	case strings.Contains(lower, "auth"), strings.Contains(lower, "permission"), strings.Contains(lower, "api key"):
		return "지금 답변 엔진 인증이 꼬여 있어. 설정부터 다시 볼게."
	default:
		return summary
	}
}

func normalizeTelegramAssistantReply(text string) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	cleaned := make([]string, 0, len(lines))
	lastBlank := false
	inCodeBlock := false

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}
		line = strings.TrimPrefix(line, "### ")
		line = strings.TrimPrefix(line, "## ")
		line = strings.TrimPrefix(line, "# ")
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			line = "• " + strings.TrimSpace(line[2:])
		}
		if line == "" {
			if lastBlank || len(cleaned) == 0 {
				continue
			}
			cleaned = append(cleaned, "")
			lastBlank = true
			continue
		}
		lastBlank = false
		cleaned = append(cleaned, html.EscapeString(line))
	}

	reply := strings.TrimSpace(strings.Join(cleaned, "\n"))
	if reply == "" {
		return "지금은 짧게 정리하기 어려워. 한 번만 더 물어봐줘."
	}

	runes := []rune(reply)
	if len(runes) > 450 {
		reply = strings.TrimSpace(string(runes[:450])) + "…"
	}
	return reply
}

func parseTelegramFounderIntent(raw string) (telegramFounderIntentResult, error) {
	cleaned := telegramExtractJSONObject(raw)
	if strings.TrimSpace(cleaned) == "" {
		return telegramFounderIntentResult{}, fmt.Errorf("intent json missing")
	}
	var intent telegramFounderIntentResult
	if err := json.Unmarshal([]byte(cleaned), &intent); err != nil {
		return telegramFounderIntentResult{}, err
	}
	intent.Mode = strings.TrimSpace(strings.ToLower(intent.Mode))
	intent.Action = strings.TrimSpace(strings.ToLower(intent.Action))
	intent.Target = strings.TrimSpace(intent.Target)
	intent.Text = strings.TrimSpace(intent.Text)
	intent.Reason = strings.TrimSpace(intent.Reason)
	intent.Confidence = strings.TrimSpace(strings.ToLower(intent.Confidence))
	return intent, nil
}

func telegramExtractJSONObject(raw string) string {
	text := strings.TrimSpace(raw)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start == -1 || end == -1 || end < start {
		return text
	}
	return text[start : end+1]
}

func telegramFounderIntentUsable(intent telegramFounderIntentResult) bool {
	if intent.Action == "" || intent.Action == "none" {
		return false
	}
	switch intent.Mode {
	case "read":
		return intent.Confidence == "high" || intent.Confidence == "medium"
	case "write":
		switch intent.Action {
		case "review.add":
			return intent.Confidence == "high" || intent.Confidence == "medium"
		default:
			return intent.Confidence == "high" && strings.TrimSpace(intent.Target) != ""
		}
	default:
		return false
	}
}

func telegramParseFounderConversationAction(text string) (telegramFounderConversationAction, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" || strings.HasPrefix(trimmed, "/") {
		return telegramFounderConversationAction{}, false
	}
	fields := strings.Fields(trimmed)
	compact := telegramCompactText(trimmed)

	if approvalID := telegramFirstTokenWithPrefix(fields, "approval_"); approvalID != "" {
		switch {
		case telegramIntentRejectsApproval(compact) && !telegramIntentAsksState(compact):
			return telegramFounderConversationAction{Kind: "approval", ApprovalID: approvalID, ApprovalVerb: "rejected"}, true
		case telegramIntentApprovesApproval(compact) && !telegramIntentAsksState(compact):
			return telegramFounderConversationAction{Kind: "approval", ApprovalID: approvalID, ApprovalVerb: "approved"}, true
		}
	}

	if jobID := telegramFirstTokenWithPrefix(fields, "job_"); jobID != "" && telegramIntentDispatchesJob(compact) && !telegramIntentAsksState(compact) {
		return telegramFounderConversationAction{Kind: "dispatch", JobID: jobID}, true
	}

	if reviewTarget := telegramReviewTargetToken(fields); reviewTarget != "" {
		if reviewAction := telegramDetectReviewAction(compact); reviewAction != "" && !telegramIntentAsksState(compact) {
			return telegramFounderConversationAction{
				Kind:         "review",
				ReviewAction: reviewAction,
				ReviewTarget: reviewTarget,
				ReviewReason: telegramExtractReviewReason(trimmed),
			}, true
		}
	}

	if telegramDetectReviewAddAction(compact) && !telegramIntentAsksState(compact) {
		return telegramFounderConversationAction{
			Kind:       "review_add",
			ReviewText: telegramExtractReviewAddText(trimmed),
		}, true
	}

	return telegramFounderConversationAction{}, false
}

func normalizeTelegramInboundContext(ctx TelegramInboundContext) TelegramInboundContext {
	ctx.RouteID = strings.TrimSpace(strings.ToLower(ctx.RouteID))
	if ctx.RouteID == "" && ctx.ChatID != 0 {
		ctx.RouteID = TelegramRouteForChatID(ctx.ChatID)
	}
	return ctx
}

func telegramRouteWelcome(ctx TelegramInboundContext, founderConversation bool) string {
	if ctx.RouteID == TelegramRouteFounder && !founderConversation {
		return "이 방은 founder room으로 붙어 있어. 알림과 결정은 여기서 받고, 대화형 비서는 1:1 founder DM에서 이어가자. DM 쪽 `/whoami` 값만 붙이면 바로 전환할 수 있어."
	}
	switch ctx.RouteID {
	case TelegramRouteFounder:
		return "올라와 있어. 편하게 말해줘. 이 방 번호가 필요하면 `/whoami` 보내면 돼."
	case TelegramRouteOps:
		return "이 방은 ops route로 붙어 있어. `/status` 와 `/review` 는 바로 볼 수 있고, 긴 대화형 답변은 founder route에서 이어갈게. 방 번호가 필요하면 `/whoami` 보내줘."
	case TelegramRouteBrand:
		return "이 방은 brand route로 붙어 있어. 여기선 리뷰와 승인 흐름 위주로 쓰고, 자유 대화는 founder route에서 이어가자. 방 번호가 필요하면 `/whoami` 보내줘."
	default:
		return telegramRouteNeedsMappingReply()
	}
}

func telegramRouteNeedsMappingReply() string {
	return "이 방은 아직 founder·ops·brand route로 연결되지 않았어. `/whoami` 보내면 chat id를 바로 보여줄게."
}

func telegramRouteAllowsReadCommands(routeID string) bool {
	switch routeID {
	case TelegramRouteFounder, TelegramRouteOps, TelegramRouteBrand:
		return true
	default:
		return false
	}
}

func telegramRouteNaturalLanguageReply(routeID string) string {
	switch routeID {
	case TelegramRouteFounder:
		return "이 방은 founder room이라서 알림과 결정 위주로 둘게. 개인비서 대화는 founder 1:1 DM에서 이어가자."
	case TelegramRouteOps:
		return "이 방은 ops route라서 긴 대화형 답변은 막아둘게. 여기서는 `/status` 와 `/review` 위주로 쓰고, 자유 대화는 founder route에서 이어가자."
	case TelegramRouteBrand:
		return "이 방은 brand route라서 대화형 비서는 founder route에서만 열어둘게. 여기서는 리뷰·승인·퍼블리시 흐름 위주로 쓰자."
	default:
		return telegramRouteNeedsMappingReply()
	}
}

func telegramRouteActor(routeID string) string {
	switch routeID {
	case TelegramRouteFounder:
		return "founder"
	case TelegramRouteOps:
		return "ops"
	case TelegramRouteBrand:
		return "brand"
	default:
		return "telegram"
	}
}

func telegramFounderConversationAllowed(ctx TelegramInboundContext) bool {
	if ctx.RouteID != TelegramRouteFounder {
		return false
	}
	if ctx.ChatID == 0 {
		return true
	}
	conversationChatID := strings.TrimSpace(telegramFounderConversationChatID())
	if conversationChatID == "" {
		return strings.EqualFold(strings.TrimSpace(ctx.ChatType), "private")
	}
	return strconv.FormatInt(ctx.ChatID, 10) == conversationChatID
}

func telegramFounderConversationRequiredReply(routeID string) string {
	if routeID == TelegramRouteFounder {
		return "개인비서 대화, 핸드오프, 메모 저장은 founder 1:1 DM에서만 받을게."
	}
	return "이 기능은 founder 1:1 DM에서만 열어둘게."
}

func telegramFounderDecisionRequiredReply() string {
	return "이 결정은 founder route에서만 받을게. founder 방이나 1:1 DM에서 /intake, /route로 이어가자."
}

func telegramChatIdentityReply(ctx TelegramInboundContext) string {
	lines := []string{"이 방 정보야."}
	if ctx.ChatID != 0 {
		lines = append(lines, fmt.Sprintf("chat_id: %d", ctx.ChatID))
	} else {
		lines = append(lines, "chat_id: 아직 못 읽었어")
	}
	if value := strings.TrimSpace(ctx.ChatType); value != "" {
		lines = append(lines, "chat_type: "+value)
	}
	switch ctx.RouteID {
	case TelegramRouteFounder, TelegramRouteOps, TelegramRouteBrand:
		lines = append(lines, "route: "+ctx.RouteID)
	default:
		lines = append(lines, "route: unmapped")
	}
	lines = append(lines, "이 숫자를 founder, ops, brand 중 맞는 route에 붙이면 돼.")
	return strings.Join(lines, "\n")
}

type telegramSourceIntakeItem struct {
	Observation     ObservationRecord
	Title           string
	SuggestedRoutes []string
}

type telegramSourceDraftSeedItem struct {
	ObservationID string
	TargetLabel   string
	Title         string
	Preview       string
}

func telegramPendingSourceIntakeItems(items []ObservationRecord, decisions []ObservationRecord) []telegramSourceIntakeItem {
	decided := map[string]bool{}
	for _, decision := range decisions {
		for _, ref := range decision.Refs {
			if strings.HasPrefix(strings.TrimSpace(ref), "observation_") {
				decided[strings.TrimSpace(ref)] = true
			}
		}
	}

	out := make([]telegramSourceIntakeItem, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Topic) != SourceIntakeTopic {
			continue
		}
		if decided[item.ObservationID] {
			continue
		}
		parsed := ParseSourceIntakeRawExcerpt(item.RawExcerpt)
		out = append(out, telegramSourceIntakeItem{
			Observation:     item,
			Title:           parsed.Title,
			SuggestedRoutes: parsed.SuggestedRoutes,
		})
	}
	return out
}

func telegramFindSourceIntakeObservation(service *Service, observationID string) (telegramSourceIntakeItem, bool) {
	observationID = strings.TrimSpace(observationID)
	if observationID == "" {
		return telegramSourceIntakeItem{}, false
	}
	items := service.ListObservations(ObservationQuery{Topic: SourceIntakeTopic, Limit: 64})
	for _, item := range items {
		if strings.TrimSpace(item.Topic) != SourceIntakeTopic {
			continue
		}
		if item.ObservationID != observationID {
			continue
		}
		parsed := ParseSourceIntakeRawExcerpt(item.RawExcerpt)
		return telegramSourceIntakeItem{
			Observation:     item,
			Title:           parsed.Title,
			SuggestedRoutes: parsed.SuggestedRoutes,
		}, true
	}
	return telegramSourceIntakeItem{}, false
}

func telegramRecentSourceDraftSeedItems(items []ObservationRecord) []telegramSourceDraftSeedItem {
	out := make([]telegramSourceDraftSeedItem, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		if strings.TrimSpace(item.Topic) != SourceDraftSeedTopic {
			continue
		}
		parsed := ParseSourceDraftSeedRawExcerpt(item.RawExcerpt)
		key := strings.TrimSpace(parsed.SourceObservationID) + ":" + strings.TrimSpace(parsed.TargetAccount)
		if key != ":" && seen[key] {
			continue
		}
		if key != ":" {
			seen[key] = true
		}
		title := strings.TrimSpace(parsed.Title)
		if title == "" {
			title = strings.TrimSpace(parsed.SourceTitle)
		}
		if title == "" {
			title = item.ObservationID
		}
		out = append(out, telegramSourceDraftSeedItem{
			ObservationID: item.ObservationID,
			TargetLabel:   sourceDraftSeedTargetLabel(parsed.TargetAccount),
			Title:         title,
			Preview:       sourceDraftSeedPreview(parsed.Draft),
		})
	}
	return out
}

func telegramFindSourceDraftSeedObservation(service *Service, observationID string) (ObservationRecord, bool) {
	observationID = strings.TrimSpace(observationID)
	if observationID == "" {
		return ObservationRecord{}, false
	}
	for _, item := range service.ListObservations(ObservationQuery{Topic: SourceDraftSeedTopic, Limit: 64}) {
		if item.ObservationID == observationID {
			return item, true
		}
	}
	return ObservationRecord{}, false
}

func telegramSourceIntakeSubject(item ObservationRecord) string {
	parsed := ParseSourceIntakeRawExcerpt(item.RawExcerpt)
	if parsed.Title != "" {
		return parsed.Title
	}
	if parsed.URL != "" {
		return parsed.URL
	}
	text := strings.TrimSpace(item.NormalizedSummary)
	if text == "" {
		text = strings.TrimSpace(item.RawExcerpt)
	}
	runes := []rune(text)
	if len(runes) > 96 {
		return strings.TrimSpace(string(runes[:96])) + "…"
	}
	if text == "" {
		return item.ObservationID
	}
	return text
}

func telegramSourceIntakeRouteLabels(routes []string) string {
	if len(routes) == 0 {
		return SourceIntakeRouteChoiceLabel("97layer")
	}
	return sourceIntakeRouteLabels(routes)
}

func telegramShouldAutoIntake(text string) bool {
	text = strings.TrimSpace(text)
	if text == "" || strings.HasPrefix(text, "/") {
		return false
	}
	if ExtractFirstSourceURL(text) != "" {
		return true
	}
	runes := []rune(text)
	if strings.Contains(text, "\n") && len(runes) >= 24 {
		return true
	}
	if len(runes) >= 120 {
		return true
	}
	return false
}

func telegramFirstField(raw string) string {
	fields := strings.Fields(strings.TrimSpace(raw))
	if len(fields) == 0 {
		return ""
	}
	return strings.TrimSpace(fields[0])
}

func telegramFirstTokenWithPrefix(fields []string, prefix string) string {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	for _, field := range fields {
		candidate := telegramTrimToken(field)
		if strings.HasPrefix(strings.ToLower(candidate), prefix) {
			return candidate
		}
	}
	return ""
}

func telegramTrimToken(value string) string {
	return strings.Trim(strings.TrimSpace(value), "`'\"()[]{}.,;!?")
}

func telegramTextHasAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(text, strings.ToLower(strings.TrimSpace(needle))) {
			return true
		}
	}
	return false
}

func telegramAsksWhereConversationWasSaved(text string) bool {
	compact := telegramCompactText(text)
	return telegramCompactHasAny(compact, "어디로남겼", "어디에남겼", "뭐로남겼", "어디에저장", "wheredidyousave", "wherewasthatsaved")
}

func telegramAsksHarnessTeamStatus(lower string) bool {
	compact := telegramCompactText(lower)
	return telegramMentionsHarnessTeam(compact) && telegramIntentAsksState(compact)
}

func telegramAsksHarnessLaneStatus(lower string, role string) bool {
	compact := telegramCompactText(lower)
	return telegramMentionsHarnessLane(compact, role) && telegramIntentAsksState(compact)
}

func telegramAsksApprovalsRead(lower string) bool {
	compact := telegramCompactText(lower)
	return telegramMentionsApprovalDomain(compact) && telegramIntentAsksState(compact) && !telegramIntentApprovesApproval(compact) && !telegramIntentRejectsApproval(compact)
}

func telegramAsksJobsRead(lower string) bool {
	compact := telegramCompactText(lower)
	return (telegramMentionsJobDomain(compact) || telegramMentionsDispatchDomain(compact)) && telegramIntentAsksState(compact) && !telegramIntentDispatchesJob(compact)
}

func telegramAsksReviewRead(lower string) bool {
	compact := telegramCompactText(lower)
	return telegramMentionsReviewDomain(compact) && telegramIntentAsksState(compact) && !telegramDetectReviewAddAction(compact)
}

func telegramAsksStatusRead(lower string) bool {
	compact := telegramCompactText(lower)
	return telegramIntentAsksState(compact) &&
		!telegramMentionsReviewDomain(compact) &&
		!telegramMentionsApprovalDomain(compact) &&
		!telegramMentionsJobDomain(compact) &&
		!telegramMentionsDispatchDomain(compact) &&
		!telegramMentionsHarnessTeam(compact) &&
		!telegramMentionsAnyHarnessLane(compact)
}

func telegramPendingApprovals(items []ApprovalItem) []ApprovalItem {
	out := make([]ApprovalItem, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Status) == "pending" {
			out = append(out, item)
		}
	}
	if len(out) > 5 {
		return append([]ApprovalItem{}, out[:5]...)
	}
	return out
}

func telegramDispatchableJobs(items []AgentJob) []AgentJob {
	out := make([]AgentJob, 0, len(items))
	for _, item := range items {
		switch strings.TrimSpace(item.Status) {
		case "queued", "failed":
			out = append(out, item)
		}
	}
	if len(out) > 6 {
		return append([]AgentJob{}, out[:6]...)
	}
	return out
}

func telegramStringResultValue(result map[string]any, key string) string {
	if len(result) == 0 {
		return ""
	}
	value, ok := result[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func telegramReviewItemMeta(item ReviewRoomItem) string {
	parts := make([]string, 0, 4)
	if item.Ref != nil && strings.TrimSpace(*item.Ref) != "" {
		parts = append(parts, "ref: "+strings.TrimSpace(*item.Ref))
	}
	if strings.TrimSpace(item.Kind) != "" {
		parts = append(parts, strings.TrimSpace(item.Kind))
	}
	if strings.TrimSpace(item.Severity) != "" {
		parts = append(parts, strings.TrimSpace(item.Severity))
	}
	if strings.TrimSpace(item.Source) != "" {
		parts = append(parts, strings.TrimSpace(item.Source))
	}
	return strings.Join(parts, " · ")
}

func telegramParseReviewAction(raw string) (string, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	if parts := strings.SplitN(raw, "::", 2); len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return raw, ""
}

func telegramReviewTargetToken(fields []string) string {
	for _, field := range fields {
		candidate := telegramTrimToken(field)
		switch {
		case candidate == "":
			continue
		case strings.Contains(candidate, "://"):
			continue
		case strings.Contains(candidate, ":"):
			return candidate
		case strings.HasPrefix(strings.ToLower(candidate), "review_"):
			return candidate
		}
	}
	return ""
}

func telegramDetectReviewAction(lower string) string {
	switch {
	case telegramTextHasAny(lower, "채택해", "accept"):
		return "accept"
	case telegramTextHasAny(lower, "보류해", "defer", "미뤄"):
		return "defer"
	case telegramTextHasAny(lower, "해결해", "resolve", "닫아", "close"):
		return "resolve"
	default:
		return ""
	}
}

func telegramExtractReviewReason(text string) string {
	_, reason := telegramParseReviewAction(text)
	return reason
}

func telegramDetectReviewAddAction(lower string) bool {
	return telegramMentionsReviewDomain(lower) &&
		telegramCompactHasAny(lower, "안건", "이슈", "agenda", "issue") &&
		telegramCompactHasAny(lower, "올려", "올리", "추가", "등록", "넣어", "넣자", "raise", "create")
}

func telegramExtractReviewAddText(text string) string {
	if target, reason := telegramParseReviewAction(text); reason != "" {
		if telegramDetectReviewAddAction(telegramCompactText(target)) {
			return reason
		}
	}
	if parts := strings.SplitN(text, ":", 2); len(parts) == 2 && telegramDetectReviewAddAction(strings.ToLower(parts[0])) {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

func telegramResolveReviewAddText(service *Service, ctx TelegramInboundContext, explicit string) (string, *string) {
	explicit = strings.TrimSpace(explicit)
	if explicit != "" {
		return explicit, nil
	}
	observation, ok := telegramLatestFounderConversationObservation(service, ctx, func(item ObservationRecord) bool {
		return !telegramAsksWhereConversationWasSaved(item.NormalizedSummary) && !telegramDetectReviewAddAction(strings.ToLower(item.NormalizedSummary))
	})
	if !ok {
		return "", nil
	}
	ref := observation.ObservationID
	text := strings.TrimSpace(observation.NormalizedSummary)
	if text == "" {
		text = strings.TrimSpace(observation.RawExcerpt)
	}
	return limitText(text, 180), &ref
}

func telegramLatestFounderConversationObservation(service *Service, ctx TelegramInboundContext, predicate func(ObservationRecord) bool) (ObservationRecord, bool) {
	query := ObservationQuery{
		SourceChannel: "conversation:telegram",
		Actor:         telegramRouteActor(ctx.RouteID),
		Limit:         16,
	}
	if ctx.ChatID != 0 {
		query.Ref = fmt.Sprintf("telegram_chat:%d", ctx.ChatID)
	}
	for _, item := range service.ListObservations(query) {
		if predicate != nil && !predicate(item) {
			continue
		}
		return item, true
	}
	return ObservationRecord{}, false
}

func telegramConversationRefSummary(refs []string) string {
	parts := make([]string, 0, 2)
	for _, ref := range refs {
		switch {
		case strings.HasPrefix(ref, "telegram_chat:"):
			parts = append(parts, ref)
		case strings.HasPrefix(ref, "telegram_route:"):
			parts = append(parts, ref)
		}
	}
	return strings.Join(parts, " · ")
}

func telegramOpenProposals(items []ProposalItem) []ProposalItem {
	out := make([]ProposalItem, 0, len(items))
	for _, item := range reverseProposals(items) {
		if strings.TrimSpace(item.Status) != "proposed" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func limitProposalItems(items []ProposalItem, max int) []ProposalItem {
	if len(items) <= max {
		return append([]ProposalItem{}, items...)
	}
	return append([]ProposalItem{}, items[:max]...)
}

func telegramJobsForRole(items []AgentJob, role string) []AgentJob {
	out := make([]AgentJob, 0, len(items))
	for _, item := range reverseAgentJobs(items) {
		if strings.TrimSpace(item.Role) != role {
			continue
		}
		switch strings.TrimSpace(item.Status) {
		case "queued", "running", "failed":
			out = append(out, item)
		}
	}
	if len(out) > 4 {
		return append([]AgentJob{}, out[:4]...)
	}
	return out
}

func telegramActiveHarnessJobs(items []AgentJob) []AgentJob {
	out := make([]AgentJob, 0, len(items))
	for _, item := range reverseAgentJobs(items) {
		switch strings.TrimSpace(item.Status) {
		case "queued", "running", "failed":
			out = append(out, item)
		}
	}
	if len(out) > 12 {
		return append([]AgentJob{}, out[:12]...)
	}
	return out
}

func telegramHarnessLaneCountLine(label string, proposalCount int, jobs []AgentJob, role string) string {
	queued, running, failed := telegramRoleJobCounts(jobs, role)
	if proposalCount > 0 {
		return fmt.Sprintf("%s: proposal %d / queued %d / running %d / failed %d", label, proposalCount, queued, running, failed)
	}
	return fmt.Sprintf("%s: queued %d / running %d / failed %d", label, queued, running, failed)
}

func telegramRoleJobCounts(items []AgentJob, role string) (int, int, int) {
	queued := 0
	running := 0
	failed := 0
	for _, item := range items {
		if strings.TrimSpace(item.Role) != role {
			continue
		}
		switch strings.TrimSpace(item.Status) {
		case "queued":
			queued++
		case "running":
			running++
		case "failed":
			failed++
		}
	}
	return queued, running, failed
}

func telegramHarnessLaneLabel(role string) string {
	switch role {
	case "planner":
		return "기획"
	case "designer":
		return "디자인"
	case "implementer":
		return "구현"
	case "verifier":
		return "검증"
	default:
		return role
	}
}

func telegramHarnessLaneKeywords(role string) []string {
	switch role {
	case "planner":
		return []string{"기획", "플랜", "planner"}
	case "designer":
		return []string{"디자인", "designer"}
	case "implementer":
		return []string{"구현", "개발", "implementer"}
	case "verifier":
		return []string{"검증", "verify", "verifier", "테스트"}
	default:
		return []string{role}
	}
}

func telegramCompactText(text string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(text)) {
		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r), r == '_', r == ':':
			b.WriteRune(r)
		}
	}
	return b.String()
}

func telegramCompactHasAny(compact string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(compact, strings.ToLower(strings.TrimSpace(needle))) {
			return true
		}
	}
	return false
}

func telegramIntentAsksState(compact string) bool {
	return telegramCompactHasAny(compact, "상태", "요약", "정리", "뭐", "무엇", "어디", "보여", "목록", "남아", "열려", "막혔", "어떻게", "how", "what")
}

func telegramIntentApprovesApproval(compact string) bool {
	return telegramCompactHasAny(compact, "승인", "approve")
}

func telegramIntentRejectsApproval(compact string) bool {
	return telegramCompactHasAny(compact, "반려", "거절", "reject")
}

func telegramIntentDispatchesJob(compact string) bool {
	return telegramCompactHasAny(compact, "디스패치", "dispatch", "실행", "보내", "runjob")
}

func telegramMentionsReviewDomain(compact string) bool {
	return telegramCompactHasAny(compact, "리뷰룸", "리뷰안건", "review")
}

func telegramMentionsApprovalDomain(compact string) bool {
	return telegramCompactHasAny(compact, "approval_", "승인")
}

func telegramMentionsJobDomain(compact string) bool {
	return telegramCompactHasAny(compact, "job_", "잡", "작업")
}

func telegramMentionsDispatchDomain(compact string) bool {
	return telegramCompactHasAny(compact, "디스패치", "dispatch")
}

func telegramMentionsHarnessTeam(compact string) bool {
	return telegramCompactHasAny(compact, "하네스", "풀스택", "풀스택팀", "팀상태", "팀구조")
}

func telegramMentionsHarnessLane(compact string, role string) bool {
	return telegramCompactHasAny(compact, telegramHarnessLaneKeywords(role)...)
}

func telegramMentionsAnyHarnessLane(compact string) bool {
	for _, role := range []string{"planner", "designer", "implementer", "verifier"} {
		if telegramMentionsHarnessLane(compact, role) {
			return true
		}
	}
	return false
}

func telegramFindReviewRoomItem(room ReviewRoom, target string) (ReviewRoomItem, bool, string) {
	target = strings.TrimSpace(target)
	if target == "" {
		return ReviewRoomItem{}, false, "리뷰 안건 대상이 비어 있어."
	}
	items := make([]ReviewRoomItem, 0, len(room.Open)+len(room.Accepted)+len(room.Deferred))
	items = append(items, room.Open...)
	items = append(items, room.Accepted...)
	items = append(items, room.Deferred...)
	if match, ok, problem := telegramMatchReviewRoomItem(items, target, true); ok || problem != "" {
		return match, ok, problem
	}
	return telegramMatchReviewRoomItem(items, target, false)
}

func telegramMatchReviewRoomItem(items []ReviewRoomItem, target string, exact bool) (ReviewRoomItem, bool, string) {
	target = strings.TrimSpace(target)
	lowerTarget := strings.ToLower(target)
	matches := make([]ReviewRoomItem, 0, 2)
	for _, item := range items {
		fields := []string{strings.TrimSpace(item.Text)}
		if item.Ref != nil && strings.TrimSpace(*item.Ref) != "" {
			fields = append(fields, strings.TrimSpace(*item.Ref))
		}
		for _, field := range fields {
			if field == "" {
				continue
			}
			lowerField := strings.ToLower(field)
			if exact {
				if lowerField != lowerTarget {
					continue
				}
			} else if !strings.Contains(lowerField, lowerTarget) {
				continue
			}
			matches = append(matches, item)
			break
		}
	}
	switch len(matches) {
	case 0:
		return ReviewRoomItem{}, false, ""
	case 1:
		return matches[0], true, ""
	default:
		return ReviewRoomItem{}, false, "리뷰 안건이 여러 개 걸렸어. /review에 나온 exact ref로 다시 찍어줘."
	}
}

func telegramReviewActionLabel(action string) string {
	switch strings.TrimSpace(action) {
	case "accept":
		return "채택"
	case "defer":
		return "보류"
	case "resolve":
		return "해결"
	default:
		return strings.TrimSpace(action)
	}
}

func telegramReviewItemTarget(item ReviewRoomItem) string {
	if item.Ref != nil && strings.TrimSpace(*item.Ref) != "" {
		return strings.TrimSpace(*item.Ref)
	}
	return item.Text
}

func reviewTopOpenTexts(items []ReviewRoomItem) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Text) != "" {
			out = append(out, item.Text)
		}
	}
	return out
}
