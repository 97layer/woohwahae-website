package runtime

import (
	"strconv"
	"strings"
)

func ensureReviewRoomItem(room ReviewRoom, section string, item ReviewRoomItem) (ReviewRoom, bool, error) {
	section = normalizeReviewSection(section)
	if section == "" {
		return ReviewRoom{}, false, ErrInvalidReviewRoomSection
	}
	item = normalizeReviewRoomItem(item)
	if item.Text == "" {
		return ReviewRoom{}, false, ErrMissingReviewRoomItem
	}

	room = copyReviewRoom(room)
	if reviewRoomHasItem(room, item.Text) {
		return room, false, nil
	}
	item = annotateReviewRoomContradictions(room, section, item)
	items := reviewRoomSection(&room, section)
	if items == nil {
		return ReviewRoom{}, false, ErrInvalidReviewRoomSection
	}
	*items = append(*items, item)
	now := zeroSafeNow()
	room.UpdatedAt = &now
	room = normalizeReviewRoom(room)
	return room, true, nil
}

func reviewRoomHasItem(room ReviewRoom, item string) bool {
	item = strings.TrimSpace(item)
	if item == "" {
		return false
	}
	for _, existing := range room.Accepted {
		if existing.Text == item {
			return true
		}
	}
	for _, existing := range room.Open {
		if existing.Text == item {
			return true
		}
	}
	for _, existing := range room.Deferred {
		if existing.Text == item {
			return true
		}
	}
	return false
}

func reviewRoomSuggestionForObservation(item ObservationRecord) (ReviewRoomItem, bool) {
	if strings.TrimSpace(strings.ToLower(item.Topic)) != SourceIntakeTopic {
		return ReviewRoomItem{}, false
	}
	record := ParseSourceIntakeRawExcerpt(item.RawExcerpt)
	if record.PolicyColor == "red" {
		ref := item.ObservationID
		subject := sourceIntakeReviewSubject(record, item)
		routes := sourceIntakeRouteLabels(record.SuggestedRoutes)
		if strings.TrimSpace(routes) == "" {
			routes = "97layer"
		}
		evidence := append([]string{"observation:" + item.ObservationID, "policy:red"}, item.Refs...)
		if channel := strings.TrimSpace(item.SourceChannel); channel != "" {
			evidence = append(evidence, "channel:"+channel)
		}
		if note := strings.TrimSpace(record.FounderNote); note != "" {
			evidence = append(evidence, "founder_note:"+sourceIntakeLimitText(note, 80))
		}
		return newSignalReviewRoomItem(
			"소스 인입 `"+subject+"`이 `"+routes+"` 경로에서 red 정책에 걸렸어. 초안이나 발행 준비로 넘기기 전에 founder 검토가 필요해.",
			"source_intake.red_policy",
			&ref,
			"red-policy source intake requires founder review before draft seed or publish prep proceeds",
			"review_room.auto.source_intake_red_policy",
			evidence,
		), true
	}
	if !sourceIntakeNeedsRouteReview(record, item) {
		return ReviewRoomItem{}, false
	}
	ref := item.ObservationID
	subject := sourceIntakeReviewSubject(record, item)
	hints := explicitSourceIntakeRouteHints(item.Refs)
	evidence := append([]string{"observation:" + item.ObservationID, "policy:" + record.PolicyColor, "sensor_route_review:true"}, item.Refs...)
	if channel := strings.TrimSpace(item.SourceChannel); channel != "" {
		evidence = append(evidence, "channel:"+channel)
	}
	if note := strings.TrimSpace(record.FounderNote); note != "" {
		evidence = append(evidence, "founder_note:"+sourceIntakeLimitText(note, 80))
	}
	evidence = append(evidence, "route_hint_count:"+strconv.Itoa(len(hints)))
	if len(hints) > 0 {
		evidence = append(evidence, "route_hints:"+strings.Join(hints, ","))
	}
	text := "센서가 주운 소스 `" + subject + "`의 route를 founder가 먼저 정해야 해."
	detail := "sensor-collected source requires founder route review before draft seed or publish prep proceeds"
	if len(hints) == 0 {
		text = "센서가 주운 소스 `" + subject + "`은 route cue가 비어 있어 founder route 결정이 필요해."
		detail = "sensor-collected source has no explicit route hint, so founder route review is required before draft seed or publish prep proceeds"
	} else if len(hints) > 1 {
		text = "센서가 주운 소스 `" + subject + "`은 route cue가 여러 개라 founder route 결정이 필요해."
		detail = "sensor-collected source has conflicting route hints, so founder route review is required before draft seed or publish prep proceeds"
	}
	return newSignalReviewRoomItem(
		text,
		"source_intake.sensor_route",
		&ref,
		detail,
		"review_room.auto.source_intake_sensor_route",
		evidence,
	), true
}

func sourceIntakeNeedsRouteReview(record SourceIntakeRecord, item ObservationRecord) bool {
	if strings.TrimSpace(record.IntakeClass) != "public_collector" {
		return false
	}
	switch normalizeObservationChannel(item.SourceChannel) {
	case "crawler", "web", "feed":
	default:
		return false
	}
	return len(explicitSourceIntakeRouteHints(item.Refs)) != 1
}

func sourceIntakeReviewSubject(record SourceIntakeRecord, item ObservationRecord) string {
	if value := strings.TrimSpace(record.Title); value != "" {
		return sourceIntakeLimitText(value, 96)
	}
	if value := strings.TrimSpace(record.URL); value != "" {
		return sourceIntakeLimitText(value, 96)
	}
	if value := strings.TrimSpace(item.NormalizedSummary); value != "" {
		return sourceIntakeLimitText(value, 96)
	}
	return sourceIntakeLimitText(strings.Join(strings.Fields(strings.TrimSpace(record.Excerpt)), " "), 96)
}

func reviewRoomSuggestionForExecute(item ExecuteRun) (ReviewRoomItem, bool) {
	if item.Status != "failed" {
		return ReviewRoomItem{}, false
	}
	ref := item.ExecuteID
	if canonicalPolicyMode(item.Mode) == "blocked" || item.Mode == "blocked" {
		return newSignalReviewRoomItem("실행 `"+item.ExecuteID+"`이 작업 `"+item.WorkItemID+"`에서 정책 `"+item.PolicyDecisionID+"` 때문에 막혔어. 재시도 전에 founder 검토나 범위 축소가 필요해.", "execute.blocked", &ref, "blocked execute requires founder review or scope reduction before retry", "review_room.auto.execute_blocked", []string{"execute:" + item.ExecuteID, "work:" + item.WorkItemID, "policy:" + item.PolicyDecisionID}), true
	}
	return newSignalReviewRoomItem("실행 `"+item.ExecuteID+"`이 작업 `"+item.WorkItemID+"`에서 실패했어. 계속 가기 전에 실행 메모를 먼저 확인해야 해.", "execute.failed", &ref, "failed execute requires note inspection before another attempt", "review_room.auto.execute_failed", []string{"execute:" + item.ExecuteID, "work:" + item.WorkItemID, "policy:" + item.PolicyDecisionID}), true
}

func reviewRoomSuggestionForVerification(item VerificationRecord) (ReviewRoomItem, bool) {
	if item.Status != "failed" {
		return ReviewRoomItem{}, false
	}
	ref := item.RecordID
	return newSignalReviewRoomItem("검증 `"+item.RecordID+"`이 범위 `"+item.Scope+"`에서 실패했어. 릴리스 전에 명령 증거를 먼저 확인해야 해.", "verification.failed", &ref, "failed verification requires command evidence review before release", "review_room.auto.verification_failed", []string{"verification:" + item.RecordID, "scope:" + item.Scope}), true
}

func reviewRoomSuggestionForDeploy(item DeployRun) (ReviewRoomItem, bool) {
	if item.Status != "failed" {
		return ReviewRoomItem{}, false
	}
	ref := item.DeployID
	return newSignalReviewRoomItem("배포 `"+item.DeployID+"`이 릴리스 `"+item.ReleaseID+"`의 대상 `"+item.Target+"`에서 실패했어. 롤백 준비 상태를 먼저 검토해야 해.", "deploy.failed", &ref, "failed deploy requires rollback readiness review", "review_room.auto.deploy_failed", []string{"deploy:" + item.DeployID, "release:" + item.ReleaseID, "target:" + item.Target}), true
}

func reviewRoomSuggestionForRollback(item RollbackRun) (ReviewRoomItem, bool) {
	if item.Status != "failed" {
		return ReviewRoomItem{}, false
	}
	ref := item.RollbackID
	return newSignalReviewRoomItem("롤백 `"+item.RollbackID+"`이 릴리스 `"+item.ReleaseID+"`의 대상 `"+item.Target+"`에서 실패했어. 수동 복구 검토가 필요해.", "rollback.failed", &ref, "failed rollback requires manual recovery review", "review_room.auto.rollback_failed", []string{"rollback:" + item.RollbackID, "release:" + item.ReleaseID, "target:" + item.Target}), true
}

func reviewRoomSuggestionForApproval(item ApprovalItem) (ReviewRoomItem, bool) {
	if item.Status != "rejected" {
		return ReviewRoomItem{}, false
	}
	ref := item.ApprovalID
	return newSignalReviewRoomItem("승인 `"+item.ApprovalID+"`이 작업 `"+item.WorkItemID+"`에서 거절됐어. 계속 진행하기 전에 founder 검토나 제안 재정리가 필요해.", "approval.rejected", &ref, "rejected approval requires founder review before execution continues", "review_room.auto.approval_rejected", []string{"approval:" + item.ApprovalID, "work:" + item.WorkItemID}), true
}

func reviewRoomSuggestionForAgentJob(item AgentJob) (ReviewRoomItem, bool) {
	if item.Status != "failed" {
		return ReviewRoomItem{}, false
	}
	ref := item.JobID
	evidence := []string{"job:" + item.JobID, "role:" + item.Role, "kind:" + item.Kind}
	if item.Ref != nil && strings.TrimSpace(*item.Ref) != "" {
		evidence = append(evidence, "ref:"+strings.TrimSpace(*item.Ref))
	}
	return newSignalReviewRoomItem("에이전트 작업 `"+item.JobID+"`이 역할 `"+item.Role+"`에서 실패했어. 레인을 이어가기 전에 founder 검토나 재시도 정책 정리가 필요해.", "agent_job.failed", &ref, "failed agent job requires founder review or retry policy before execution continues", "review_room.auto.agent_job_failed", evidence), true
}

func reviewRoomSuggestionForSessionFinish(item SessionFinishInput, actor string) (ReviewRoomItem, bool) {
	if len(item.OpenRisks) == 0 {
		return ReviewRoomItem{}, false
	}
	ref := strings.TrimSpace(actor)
	if ref == "" {
		ref = "system"
	}
	evidence := []string{"actor:" + ref, "focus:" + strings.TrimSpace(item.CurrentFocus)}
	for _, risk := range item.OpenRisks {
		value := strings.TrimSpace(risk)
		if value == "" {
			continue
		}
		evidence = append(evidence, "risk:"+value)
	}
	return newSignalReviewRoomItem("세션이 포커스 `"+strings.TrimSpace(item.CurrentFocus)+"`에서 열린 리스크를 남긴 채 종료됐어. 다음 실행 레인 전에 이월 검토가 필요해.", "session.finished", nil, "session finished with open risks requires carry-over review before the next execution lane", "review_room.auto.session_open_risks", evidence), true
}

func autoResolveReviewRoomItemLocked(store *reviewRoomStore, source string, ref *string, resolution *ReviewRoomResolution) (ReviewRoom, *ReviewRoomItem, bool, error) {
	room := store.current()
	target, ok := findOpenReviewRoomItem(room, source, ref)
	if !ok {
		return room, nil, false, nil
	}
	return transitionReviewRoomItem(room, "resolve", target, resolution)
}

func findOpenReviewRoomItem(room ReviewRoom, source string, ref *string) (string, bool) {
	targetRef := strings.TrimSpace(refString(ref))
	for _, item := range room.Open {
		if strings.TrimSpace(item.Source) != strings.TrimSpace(source) {
			continue
		}
		if targetRef != "" && strings.TrimSpace(refString(item.Ref)) != targetRef {
			continue
		}
		return item.Text, true
	}
	return "", false
}

func autoResolveReviewRoomItemsByEvidenceLocked(store *reviewRoomStore, source string, evidence string, resolution *ReviewRoomResolution) (ReviewRoom, int, error) {
	room := store.current()
	evidence = strings.TrimSpace(evidence)
	if evidence == "" {
		return room, 0, nil
	}
	resolved := 0
	for {
		target, ok := findOpenReviewRoomItemByEvidence(room, source, evidence)
		if !ok {
			return room, resolved, nil
		}
		var removed *ReviewRoomItem
		var changed bool
		var err error
		room, removed, changed, err = transitionReviewRoomItem(room, "resolve", target, resolution)
		if err != nil {
			return room, resolved, err
		}
		if !changed || removed == nil {
			return room, resolved, nil
		}
		resolved++
	}
}

func findOpenReviewRoomItemByEvidence(room ReviewRoom, source string, evidence string) (string, bool) {
	evidence = strings.TrimSpace(evidence)
	if evidence == "" {
		return "", false
	}
	for _, item := range room.Open {
		if strings.TrimSpace(item.Source) != strings.TrimSpace(source) {
			continue
		}
		if !reviewRoomItemHasEvidence(item, evidence) {
			continue
		}
		return item.Text, true
	}
	return "", false
}

func reviewRoomItemHasEvidence(item ReviewRoomItem, evidence string) bool {
	evidence = strings.TrimSpace(evidence)
	if evidence == "" {
		return false
	}
	for _, value := range item.Evidence {
		if strings.TrimSpace(value) == evidence {
			return true
		}
	}
	return false
}

func transitionReviewRoomItem(room ReviewRoom, action string, item string, resolution *ReviewRoomResolution) (ReviewRoom, *ReviewRoomItem, bool, error) {
	action = normalizeReviewAction(action)
	if action == "" {
		return ReviewRoom{}, nil, false, ErrInvalidReviewRoomAction
	}
	item = strings.TrimSpace(item)
	if item == "" {
		return ReviewRoom{}, nil, false, ErrMissingReviewRoomItem
	}

	room = copyReviewRoom(room)
	room.Issues = []string{}
	changed := false
	var removedItem *ReviewRoomItem
	room.Accepted, removedItem, changed = removeReviewRoomItem(room.Accepted, item)
	if !changed {
		room.Open, removedItem, changed = removeReviewRoomItem(room.Open, item)
	}
	if !changed {
		room.Deferred, removedItem, changed = removeReviewRoomItem(room.Deferred, item)
	}
	if !changed || removedItem == nil {
		return room, nil, false, nil
	}

	now := zeroSafeNow()
	removedItem.UpdatedAt = now
	removedItem.Resolution = defaultReviewRoomResolution(action, resolution, now)
	if action == "defer" {
		stamped := stampDeferredWhyUnresolved(*removedItem)
		removedItem = &stamped
	}
	switch action {
	case "accept":
		room.Accepted = append(room.Accepted, *removedItem)
	case "defer":
		room.Deferred = append(room.Deferred, *removedItem)
	case "resolve":
		// removal only
	}
	room.UpdatedAt = &now
	room = normalizeReviewRoom(room)
	return room, removedItem, true, nil
}

func removeReviewRoomItem(items []ReviewRoomItem, target string) ([]ReviewRoomItem, *ReviewRoomItem, bool) {
	if len(items) == 0 {
		return items, nil, false
	}
	filtered := make([]ReviewRoomItem, 0, len(items))
	var removed *ReviewRoomItem
	for _, item := range items {
		if item.Text == target && removed == nil {
			copyItem := item
			removed = &copyItem
			continue
		}
		filtered = append(filtered, item)
	}
	if removed == nil {
		return items, nil, false
	}
	return filtered, removed, true
}
