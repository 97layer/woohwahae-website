package main

import (
	"fmt"
	"strings"

	"layer-os/internal/runtime"
)

type nextActionKind string

const (
	nextActionAgenda    nextActionKind = "agenda"
	nextActionContext   nextActionKind = "context_promote"
	nextActionNone      nextActionKind = "none"
	nextActionRoute     nextActionKind = "route"
	nextActionPrimary   nextActionKind = "primary_action"
	nextActionProposal  nextActionKind = "proposal_candidate"
	nextActionOpenJob   nextActionKind = "open_job"
	nextActionNoPending string         = "no pending work"
)

const nextActionContextPromoteCommand = "layer-osctl job promote --limit 1 --dispatch"

type nextActionCandidate struct {
	Kind       nextActionKind
	Display    string
	Summary    string
	TargetLane string
	Source     string
	Job        *runtime.AgentJob
}

func nextJobTitle(item runtime.AgentJob) string {
	if item.Payload != nil {
		if title, ok := item.Payload["title"].(string); ok && strings.TrimSpace(title) != "" {
			return strings.TrimSpace(title)
		}
	}
	return strings.TrimSpace(item.Summary)
}

func nextActionCandidates(packet runtime.SessionBootstrapPacket, items []runtime.AgentJob) []nextActionCandidate {
	candidates := []nextActionCandidate{}
	for _, route := range packet.Knowledge.ActionRoutes {
		display := strings.TrimSpace(pointerString(route.Command))
		summary := strings.TrimSpace(route.Summary)
		if display == "" {
			display = summary
		}
		if summary == "" {
			summary = display
		}
		if display == "" {
			continue
		}
		candidates = append(candidates, nextActionCandidate{
			Kind:       nextActionRoute,
			Display:    display,
			Summary:    summary,
			TargetLane: strings.TrimSpace(route.TargetLane),
			Source:     strings.TrimSpace(route.Source),
		})
	}
	if agenda, ok := reviewAgendaCandidate(packet); ok {
		candidates = append(candidates, agenda)
	}
	if action := strings.TrimSpace(packet.Knowledge.PrimaryAction); action != "" {
		candidates = append(candidates, nextActionCandidate{
			Kind:    nextActionPrimary,
			Display: action,
			Summary: action,
			Source:  "knowledge.primary_action",
		})
	}
	for _, candidate := range packet.Knowledge.ProposalCandidates {
		display := strings.TrimSpace(candidate.CreateCommand)
		summary := strings.TrimSpace(candidate.Summary)
		if summary == "" {
			summary = strings.TrimSpace(candidate.Title)
		}
		if display == "" {
			display = summary
		}
		if summary == "" {
			summary = display
		}
		if display == "" {
			continue
		}
		candidates = append(candidates, nextActionCandidate{
			Kind:       nextActionProposal,
			Display:    display,
			Summary:    summary,
			TargetLane: "planner",
			Source:     strings.TrimSpace(candidate.Source),
		})
	}
	for _, item := range items {
		status := strings.TrimSpace(item.Status)
		if status != "queued" && status != "running" && status != "failed" {
			continue
		}
		job := item
		candidates = append(candidates, nextActionCandidate{
			Kind:       nextActionOpenJob,
			Display:    fmt.Sprintf("[%s] %s (%s)", job.JobID, nextJobTitle(job), job.Status),
			Summary:    strings.TrimSpace(job.Summary),
			TargetLane: strings.TrimSpace(job.Role),
			Source:     strings.TrimSpace(job.Source),
			Job:        &job,
		})
	}
	if context, ok := promotableContextCandidate(packet); ok {
		candidates = append(candidates, context)
	}
	return candidates
}

func reviewAgendaCandidate(packet runtime.SessionBootstrapPacket) (nextActionCandidate, bool) {
	count, text, source := reviewAgendaSummary(packet)
	if count == 0 || strings.TrimSpace(text) == "" {
		return nextActionCandidate{}, false
	}
	summary := "Review agenda: " + strings.TrimSpace(text)
	return nextActionCandidate{
		Kind:       nextActionAgenda,
		Display:    summary,
		Summary:    summary,
		TargetLane: "planner",
		Source:     source,
	}, true
}

func reviewAgendaSummary(packet runtime.SessionBootstrapPacket) (int, string, string) {
	if packet.ReviewRoom != nil {
		if packet.ReviewRoom.OpenCount > 0 {
			text := ""
			source := "review_room.open"
			if len(packet.ReviewRoom.TopOpen) > 0 {
				text = strings.TrimSpace(packet.ReviewRoom.TopOpen[0].Text)
				if value := strings.TrimSpace(packet.ReviewRoom.TopOpen[0].Source); value != "" {
					source = value
				}
			}
			if text == "" {
				text = "review room requires attention before widening execution lanes"
			}
			return packet.ReviewRoom.OpenCount, text, source
		}
	}
	if packet.Knowledge.ReviewOpenCount > 0 {
		text := ""
		if len(packet.Knowledge.ReviewTopOpen) > 0 {
			text = strings.TrimSpace(packet.Knowledge.ReviewTopOpen[0])
		}
		if text == "" {
			text = "review room requires attention before widening execution lanes"
		}
		return packet.Knowledge.ReviewOpenCount, text, "knowledge.review_top_open"
	}
	return 0, "", ""
}

func promotableContextCandidate(packet runtime.SessionBootstrapPacket) (nextActionCandidate, bool) {
	if summary, source, ok := firstPromotableParallelCandidate(packet.Knowledge.ParallelCandidates); ok {
		return nextActionCandidate{
			Kind:       nextActionContext,
			Display:    nextActionContextPromoteCommand,
			Summary:    summary,
			TargetLane: "planner",
			Source:     source,
		}, true
	}
	if summary, source, ok := firstPromotableOpenThread(packet.Knowledge.OpenThreads); ok {
		return nextActionCandidate{
			Kind:       nextActionContext,
			Display:    nextActionContextPromoteCommand,
			Summary:    summary,
			TargetLane: "planner",
			Source:     source,
		}, true
	}
	return nextActionCandidate{}, false
}

func firstPromotableParallelCandidate(items []runtime.ParallelCandidate) (string, string, bool) {
	for _, item := range items {
		if !parallelCandidatePromotable(item) {
			continue
		}
		summary := strings.TrimSpace(item.Text)
		if summary == "" {
			continue
		}
		source := strings.TrimSpace(item.Source)
		if source == "" {
			source = "knowledge.parallel_candidate"
		}
		return summary, source, true
	}
	return "", "", false
}

func parallelCandidatePromotable(item runtime.ParallelCandidate) bool {
	summary := strings.TrimSpace(item.Text)
	if summary == "" {
		return false
	}
	kind := strings.TrimSpace(item.Kind)
	source := strings.TrimSpace(item.Source)
	reason := strings.ToLower(strings.TrimSpace(item.Reason))
	if kind == "agent_job" || kind == "approval" || kind == "release" || kind == "rollback" {
		return false
	}
	if source == "founder_view.now" {
		return false
	}
	switch {
	case strings.Contains(reason, "review-room"):
		return true
	case kind == "proposal":
		return true
	case kind == "flow" && source == "founder_view.risk":
		return true
	default:
		return false
	}
}

func firstPromotableOpenThread(items []runtime.OpenThread) (string, string, bool) {
	for _, item := range items {
		if !openThreadPromotable(item) {
			continue
		}
		summary := strings.TrimSpace(item.Question)
		if summary == "" {
			continue
		}
		source := strings.TrimSpace(item.Source)
		if source == "" {
			source = "knowledge.open_thread"
		}
		return summary, source, true
	}
	return "", "", false
}

func openThreadPromotable(item runtime.OpenThread) bool {
	status := strings.ToLower(strings.TrimSpace(item.Status))
	if status == "resolved" || status == "dismissed" {
		return false
	}
	return openThreadPromotionKind(item) != "stale"
}

func openThreadPromotionKind(item runtime.OpenThread) string {
	threadID := strings.TrimSpace(item.ThreadID)
	if prefix, _, ok := strings.Cut(threadID, "_"); ok {
		switch prefix {
		case "pattern", "contradiction", "stale", "curiosity":
			return prefix
		}
	}
	if strings.TrimSpace(strings.ToLower(item.Source)) == "corpus_signal" {
		return "pattern"
	}
	for _, ref := range item.PatternRefs {
		if strings.Contains(strings.ToLower(ref), "contradiction") {
			return "contradiction"
		}
	}
	return "curiosity"
}

func resolveNextAction(packet runtime.SessionBootstrapPacket, items []runtime.AgentJob) string {
	candidates := nextActionCandidates(packet, items)
	if len(candidates) == 0 {
		return nextActionNoPending
	}
	return candidates[0].Display
}

func pointerString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
