package main

import (
	"strings"
	"testing"

	"layer-os/internal/runtime"
)

type nextServiceStub struct {
	status    string
	limit     int
	items     []runtime.AgentJob
	bootstrap runtime.SessionBootstrapPacket
}

func (s *nextServiceStub) SessionBootstrap(full bool) (runtime.SessionBootstrapPacket, error) {
	return s.bootstrap, nil
}

func (s *nextServiceStub) ListAgentJobsByStatus(status string, limit int) []runtime.AgentJob {
	s.status = status
	s.limit = limit
	return append([]runtime.AgentJob{}, s.items...)
}

func TestRunNextPrefersActionRouteCommand(t *testing.T) {
	command := "layer-osctl job create --summary \"Route first\""
	service := &nextServiceStub{
		bootstrap: runtime.SessionBootstrapPacket{
			Source: "daemon",
			Knowledge: runtime.KnowledgePacket{
				ActionRoutes: []runtime.ActionRoute{{
					RouteID:    "route_001",
					Kind:       "job",
					Summary:    "Run routed work",
					TargetLane: "implementer",
					Command:    &command,
					Source:     "knowledge.route",
				}},
				PrimaryAction: "fallback action",
			},
		},
		items: []runtime.AgentJob{{JobID: "job_001", Summary: "Plan founder lane", Status: "queued"}},
	}
	raw := captureStdout(t, func() {
		runNext(service, nil)
	})
	if strings.TrimSpace(raw) != command {
		t.Fatalf("expected route command, got %q", raw)
	}
}

func TestRunNextFallsBackToPrimaryActionThenOpenJob(t *testing.T) {
	service := &nextServiceStub{
		bootstrap: runtime.SessionBootstrapPacket{
			Source: "daemon",
			Knowledge: runtime.KnowledgePacket{
				PrimaryAction: "Review the current verification lane",
			},
		},
		items: []runtime.AgentJob{{JobID: "job_001", Summary: "Plan founder lane", Status: "queued"}},
	}
	raw := captureStdout(t, func() {
		runNext(service, nil)
	})
	if service.status != "open" || service.limit != 1 {
		t.Fatalf("expected open limit=1 query, got status=%q limit=%d", service.status, service.limit)
	}
	if strings.TrimSpace(raw) != "Review the current verification lane" {
		t.Fatalf("expected primary action, got %q", raw)
	}
}

func TestRunNextPrefersReviewAgendaOverOpaqueReviewRoomPrimaryAction(t *testing.T) {
	service := &nextServiceStub{
		bootstrap: runtime.SessionBootstrapPacket{
			Source: "daemon",
			Knowledge: runtime.KnowledgePacket{
				PrimaryAction:   "review_room",
				ReviewOpenCount: 1,
				ReviewTopOpen:   []string{"Provider drift still needs triage before new execution lanes."},
			},
			ReviewRoom: &runtime.ReviewRoomSummary{
				Source:    "review_room.json",
				OpenCount: 1,
				TopOpen: []runtime.ReviewRoomItem{{
					Text:     "Provider drift still needs triage before new execution lanes.",
					Kind:     "agenda",
					Severity: "high",
					Source:   "review.auto",
				}},
				TopAccepted: []runtime.ReviewRoomItem{},
				TopDeferred: []runtime.ReviewRoomItem{},
				Issues:      []string{},
			},
		},
	}
	raw := captureStdout(t, func() {
		runNext(service, nil)
	})
	if strings.TrimSpace(raw) != "Review agenda: Provider drift still needs triage before new execution lanes." {
		t.Fatalf("expected review agenda guidance, got %q", raw)
	}
}

func TestResolveNextActionFallsBackToProposalCommandAndNoPendingWork(t *testing.T) {
	packet := runtime.SessionBootstrapPacket{
		Source: "daemon",
		Knowledge: runtime.KnowledgePacket{
			ProposalCandidates: []runtime.ProposalCandidate{{
				CreateCommand: "layer-osctl proposal create --title \"Proposal\"",
			}},
		},
	}
	if got := resolveNextAction(packet, nil); got != "layer-osctl proposal create --title \"Proposal\"" {
		t.Fatalf("expected proposal create command, got %q", got)
	}
	packet.Knowledge.ProposalCandidates = nil
	if got := resolveNextAction(packet, nil); got != "no pending work" {
		t.Fatalf("expected no pending work, got %q", got)
	}
}

func TestResolveNextActionFallsBackToContextPromotionCommandForParallelCandidate(t *testing.T) {
	packet := runtime.SessionBootstrapPacket{
		Source: "daemon",
		Knowledge: runtime.KnowledgePacket{
			ParallelCandidates: []runtime.ParallelCandidate{{
				Text:   "Review room carry-over still needs a bounded planner pass.",
				Kind:   "proposal",
				Source: "knowledge.parallel_candidate",
				Reason: "review-room carry-over",
			}},
		},
	}
	if got := resolveNextAction(packet, nil); got != nextActionContextPromoteCommand {
		t.Fatalf("expected context promotion command, got %q", got)
	}
}

func TestResolveNextActionFallsBackToContextPromotionCommandForOpenThread(t *testing.T) {
	packet := runtime.SessionBootstrapPacket{
		Source: "daemon",
		Knowledge: runtime.KnowledgePacket{
			OpenThreads: []runtime.OpenThread{{
				ThreadID: "pattern_001",
				Question: "Why does planner work keep returning to the same review seam?",
				Status:   "open",
				Source:   "corpus_signal",
			}},
		},
	}
	if got := resolveNextAction(packet, nil); got != nextActionContextPromoteCommand {
		t.Fatalf("expected context promotion command, got %q", got)
	}
}

func TestResolveNextActionPrefersOpenJobOverContextPromotion(t *testing.T) {
	packet := runtime.SessionBootstrapPacket{
		Source: "daemon",
		Knowledge: runtime.KnowledgePacket{
			OpenThreads: []runtime.OpenThread{{
				ThreadID: "pattern_001",
				Question: "Why does planner work keep returning to the same review seam?",
				Status:   "open",
				Source:   "corpus_signal",
			}},
		},
	}
	items := []runtime.AgentJob{{
		JobID:   "job_001",
		Summary: "Plan founder lane",
		Status:  "queued",
		Role:    "planner",
	}}
	if got := resolveNextAction(packet, items); got != "[job_001] Plan founder lane (queued)" {
		t.Fatalf("expected open job to win over context promotion, got %q", got)
	}
}

func TestNextJobTitlePrefersPayloadTitle(t *testing.T) {
	item := runtime.AgentJob{Summary: "Fallback summary", Payload: map[string]any{"title": "Primary title"}}
	if got := nextJobTitle(item); got != "Primary title" {
		t.Fatalf("expected payload title, got %q", got)
	}
}
