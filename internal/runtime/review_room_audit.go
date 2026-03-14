package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (s *Service) currentReviewRoomLocked() ReviewRoom {
	if !auditPromotionEnabled() {
		return s.reviewRoom.current()
	}
	room, changed := syncAuditReviewRoom(s.reviewRoom.current(), s.repoRoot, s.daemonStartedAt)
	if changed {
		s.reviewRoom = newReviewRoomStore(room)
	}
	return room
}

func syncAuditReviewRoom(room ReviewRoom, root string, daemonStartedAt time.Time) (ReviewRoom, bool) {
	if !auditPromotionEnabled() {
		return copyReviewRoom(room), false
	}
	current := copyReviewRoom(room)
	synced := copyReviewRoom(room)
	synced.Accepted = filterManagedAuditItems(synced.Accepted)
	synced.Open = filterManagedAuditItems(synced.Open)
	synced.Deferred = filterManagedAuditItems(synced.Deferred)
	for _, item := range auditReviewRoomItems(root, daemonStartedAt) {
		updated, _, err := ensureReviewRoomItem(synced, "open", item)
		if err != nil {
			continue
		}
		synced = updated
	}
	synced = normalizeReviewRoom(synced)
	if !reviewRoomsEqual(current, synced) {
		now := zeroSafeNow()
		synced.UpdatedAt = &now
		return synced, true
	}
	return synced, false
}

func auditReviewRoomItems(root string, daemonStartedAt time.Time) []ReviewRoomItem {
	items := []ReviewRoomItem{}
	for _, issue := range AuditStructure(root).Issues {
		items = append(items, newStructuredReviewRoomItem("Structure drift: "+strings.TrimSpace(issue), "structure_drift", "high", "audit.structure", nil, &ReviewRoomRationale{Trigger: "audit.structure", Reason: "structure audit found workspace layout drift that needs cleanup or correction", Rule: "review_room.audit.structure"}, []string{strings.TrimSpace(issue)}))
	}
	for _, issue := range AuditContracts(root).Issues {
		items = append(items, newStructuredReviewRoomItem("Contract drift: "+strings.TrimSpace(issue), "contract_drift", "high", "audit.contracts", nil, &ReviewRoomRationale{Trigger: "audit.contracts", Reason: "contract audit found missing or unexpected schema entries", Rule: "review_room.audit.contracts"}, []string{strings.TrimSpace(issue)}))
	}
	geminiAudit := AuditGemini(root)
	if geminiAudit.Status != "ok" {
		evidence := []string{}
		if strings.TrimSpace(geminiAudit.ProjectPolicyPath) != "" {
			evidence = append(evidence, "policy:"+strings.TrimSpace(geminiAudit.ProjectPolicyPath))
		}
		evidence = append(evidence, geminiAudit.PolicyIssues...)
		evidence = append(evidence, geminiAudit.ArtifactMatches...)
		severity := "medium"
		if len(geminiAudit.PolicyIssues) > 0 {
			severity = "high"
		}
		items = append(items, newStructuredReviewRoomItem(
			fmt.Sprintf("Gemini containment drift: %d policy issue(s) and %d stray artifact(s) require cleanup or policy correction.", len(geminiAudit.PolicyIssues), len(geminiAudit.ArtifactMatches)),
			"gemini_drift",
			severity,
			"audit.gemini",
			nil,
			&ReviewRoomRationale{Trigger: "audit.gemini", Reason: "Gemini containment audit found policy gaps or stray scratchpads that should be absorbed into canonical reporting flows", Rule: "review_room.audit.gemini"},
			evidence,
		))
	}
	corpusAudit := AuditCorpus(root, filepath.Join(root, ".layer-os"))
	if corpusAudit.Status != "ok" {
		items = append(items, newStructuredReviewRoomItem(
			fmt.Sprintf("Corpus contamination drift: %d markdown artifact(s) require canonical recovery.", len(corpusAudit.ArtifactMatches)),
			"corpus_contamination",
			"high",
			"audit.corpus",
			nil,
			&ReviewRoomRationale{Trigger: "audit.corpus", Reason: "Corpus audit found markdown artifacts outside the canonical JSON ledger path", Rule: "review_room.audit.corpus"},
			append([]string{}, corpusAudit.ArtifactMatches...),
		))
	}
	for _, match := range AuditResidue(root).Matches {
		if isGeminiArtifactResidueMatch(match) {
			continue
		}
		items = append(items, newStructuredReviewRoomItem("Residue drift: "+strings.TrimSpace(match), "residue_drift", "medium", "audit.residue", nil, &ReviewRoomRationale{Trigger: "audit.residue", Reason: "residue audit found legacy or foreign terms that should be removed", Rule: "review_room.audit.residue"}, []string{strings.TrimSpace(match)}))
	}
	for _, issue := range AuditSurface(root).Issues {
		items = append(items, newStructuredReviewRoomItem("Surface drift: "+strings.TrimSpace(issue), "surface_drift", "high", "audit.surface", nil, &ReviewRoomRationale{Trigger: "audit.surface", Reason: "surface audit found CLI or API parity drift", Rule: "review_room.audit.surface"}, []string{strings.TrimSpace(issue)}))
	}
	daemonAudit := AuditDaemonFreshness(root, daemonStartedAt)
	if daemonAudit.Status != "ok" {
		evidence := append([]string{}, daemonAudit.Issues...)
		if daemonAudit.StartedAt != nil {
			evidence = append(evidence, "started_at:"+daemonAudit.StartedAt.UTC().Format(time.RFC3339))
		}
		if daemonAudit.LatestSourceAt != nil {
			evidence = append(evidence, "latest_source_at:"+daemonAudit.LatestSourceAt.UTC().Format(time.RFC3339))
		}
		if strings.TrimSpace(daemonAudit.LatestSourcePath) != "" {
			evidence = append(evidence, "latest_source_path:"+strings.TrimSpace(daemonAudit.LatestSourcePath))
		}
		items = append(items, newStructuredReviewRoomItem(
			"Daemon freshness drift: source changed after the current layer-osd start; restart the daemon before trusting new routes or contracts.",
			"daemon_freshness_drift",
			"high",
			"audit.daemon",
			nil,
			&ReviewRoomRationale{Trigger: "audit.daemon", Reason: "repo source is newer than the currently bound daemon start time", Rule: "review_room.audit.daemon"},
			evidence,
		))
	}
	return items
}

func filterManagedAuditItems(items []ReviewRoomItem) []ReviewRoomItem {
	filtered := make([]ReviewRoomItem, 0, len(items))
	for _, item := range items {
		if isManagedAuditSource(item.Source) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func isManagedAuditSource(source string) bool {
	source = strings.TrimSpace(strings.ToLower(source))
	return strings.HasPrefix(source, "audit.")
}

func reviewRoomsEqual(left ReviewRoom, right ReviewRoom) bool {
	left = normalizeReviewRoom(left)
	right = normalizeReviewRoom(right)
	return reviewRoomItemsEqual(left.Accepted, right.Accepted) &&
		reviewRoomItemsEqual(left.Open, right.Open) &&
		reviewRoomItemsEqual(left.Deferred, right.Deferred) &&
		stringSlicesEqual(left.Issues, right.Issues) &&
		strings.TrimSpace(left.Source) == strings.TrimSpace(right.Source)
}

func reviewRoomItemsEqual(left []ReviewRoomItem, right []ReviewRoomItem) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Text != right[index].Text ||
			left[index].Kind != right[index].Kind ||
			left[index].Severity != right[index].Severity ||
			left[index].Source != right[index].Source ||
			refString(left[index].Ref) != refString(right[index].Ref) {
			return false
		}
	}
	return true
}

func stringSlicesEqual(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func refString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func auditPromotionEnabled() bool {
	return strings.TrimSpace(os.Getenv("LAYER_OS_REPO_ROOT")) != ""
}
