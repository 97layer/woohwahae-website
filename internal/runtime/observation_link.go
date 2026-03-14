package runtime

import (
	"fmt"
	"hash/crc32"
	"sort"
	"strings"
	"time"
)

type observationLinkGroup struct {
	kind          string
	key           string
	observations  map[string]ObservationRecord
	channels      map[string]struct{}
	refs          map[string]struct{}
	latestSummary string
	latestAt      time.Time
}

type observationLinkCandidate struct {
	fingerprint string
	link        ObservationLink
}

func deriveObservationLinks(items []ObservationRecord) []ObservationLink {
	if len(items) == 0 {
		return []ObservationLink{}
	}
	groups := map[string]*observationLinkGroup{}
	for _, raw := range items {
		item := normalizeObservationRecord(raw, raw.Actor)
		for _, ref := range item.Refs {
			if !linkableObservationRef(ref) {
				continue
			}
			addObservationLinkGroup(groups, "ref", ref, item)
		}
		if key := observationSemanticKey(item); key != "" {
			addObservationLinkGroup(groups, "semantic", key, item)
		}
	}
	byFingerprint := map[string]observationLinkCandidate{}
	for _, group := range groups {
		if len(group.observations) < 2 {
			continue
		}
		candidate := observationLinkCandidateFromGroup(group)
		if existing, ok := byFingerprint[candidate.fingerprint]; ok {
			if preferObservationLinkCandidate(candidate, existing) {
				byFingerprint[candidate.fingerprint] = candidate
			}
			continue
		}
		byFingerprint[candidate.fingerprint] = candidate
	}
	links := make([]ObservationLink, 0, len(byFingerprint))
	for _, candidate := range byFingerprint {
		links = append(links, candidate.link)
	}
	sort.SliceStable(links, func(i, j int) bool {
		if links[i].Count == links[j].Count {
			if len(links[i].Channels) == len(links[j].Channels) {
				if links[i].LatestAt.Equal(links[j].LatestAt) {
					return links[i].LinkID < links[j].LinkID
				}
				return links[i].LatestAt.After(links[j].LatestAt)
			}
			return len(links[i].Channels) > len(links[j].Channels)
		}
		return links[i].Count > links[j].Count
	})
	return links
}

func addObservationLinkGroup(groups map[string]*observationLinkGroup, kind string, key string, item ObservationRecord) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	groupKey := kind + ":" + key
	group, ok := groups[groupKey]
	if !ok {
		group = &observationLinkGroup{
			kind:         kind,
			key:          key,
			observations: map[string]ObservationRecord{},
			channels:     map[string]struct{}{},
			refs:         map[string]struct{}{},
		}
		groups[groupKey] = group
	}
	group.observations[item.ObservationID] = item
	group.channels[normalizeObservationChannel(item.SourceChannel)] = struct{}{}
	for _, ref := range item.Refs {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		group.refs[ref] = struct{}{}
	}
	if item.CapturedAt.After(group.latestAt) || group.latestAt.IsZero() {
		group.latestAt = item.CapturedAt
		group.latestSummary = strings.TrimSpace(item.NormalizedSummary)
	}
}

func observationLinkCandidateFromGroup(group *observationLinkGroup) observationLinkCandidate {
	ids := make([]string, 0, len(group.observations))
	for id := range group.observations {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	channels := mapKeys(group.channels)
	refs := mapKeys(group.refs)
	link := ObservationLink{
		LinkID:   deriveObservationLinkID(group.kind, group.key, ids),
		Kind:     group.kind,
		Summary:  observationLinkSummary(group.kind, group.key, group.latestSummary, len(ids)),
		Channels: channels,
		Refs:     refs,
		Count:    len(ids),
		LatestAt: group.latestAt,
	}
	return observationLinkCandidate{fingerprint: strings.Join(ids, "|"), link: link}
}

func preferObservationLinkCandidate(next observationLinkCandidate, current observationLinkCandidate) bool {
	if next.link.Kind != current.link.Kind {
		return next.link.Kind == "ref"
	}
	if next.link.Count != current.link.Count {
		return next.link.Count > current.link.Count
	}
	if len(next.link.Channels) != len(current.link.Channels) {
		return len(next.link.Channels) > len(current.link.Channels)
	}
	return next.link.LatestAt.After(current.link.LatestAt)
}

func observationLinkSummary(kind string, key string, latestSummary string, count int) string {
	base := strings.TrimSpace(latestSummary)
	if base == "" {
		switch kind {
		case "ref":
			base = key
		default:
			base = strings.ReplaceAll(key, " ", " ")
		}
	}
	return limitText(fmt.Sprintf("%s [%d linked observations]", base, count), 140)
}

func deriveObservationLinkID(kind string, key string, ids []string) string {
	payload := strings.Join(append([]string{strings.TrimSpace(kind), strings.TrimSpace(key)}, ids...), "|")
	return fmt.Sprintf("oblink_%08x", crc32.ChecksumIEEE([]byte(payload)))
}

func linkableObservationRef(ref string) bool {
	return !isNonLinkableObservationRef(ref)
}

func observationSemanticKey(item ObservationRecord) string {
	if tokens := filteredObservationTokens(item.Topic); !genericObservationTopic(item.Topic) && len(tokens) > 0 {
		return strings.Join(limitStrings(tokens, 3), " ")
	}
	if tokens := filteredObservationTokens(item.NormalizedSummary); len(tokens) > 0 {
		return strings.Join(limitStrings(tokens, 3), " ")
	}
	return ""
}

func filteredObservationTokens(text string) []string {
	tokens := extractCorpusTokens(text)
	if len(tokens) == 0 {
		return []string{}
	}
	items := make([]string, 0, len(tokens))
	seen := map[string]struct{}{}
	for _, token := range tokens {
		if genericObservationToken(token) {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		items = append(items, token)
	}
	return items
}

func genericObservationTopic(topic string) bool {
	switch strings.ToLower(strings.TrimSpace(topic)) {
	case "", "session.finished", "agent_job.succeeded", "agent_job.failed", "agent_job.canceled", "agent_job.updated", "observation.created", "message", "telegram_message", "content_capture", "article", "note", "bookmark", "video", "publication", "feedback", "conversation", "link", "task", "walkthrough", "implementation_plan", "integrity_diagnosis", "session_summary", "wakeup_report", "deep_work_progress", "deploy_scratchpad", "deployment_checklist", "next_steps", "legacy_council", "legacy_state":
		return true
	default:
		return false
	}
}

func genericObservationToken(token string) bool {
	switch strings.ToLower(strings.TrimSpace(token)) {
	case "session", "finished", "agent", "job", "succeeded", "failed", "canceled", "updated", "observation", "created", "report", "reported", "lane", "still", "into", "with", "from", "that", "this", "then", "them", "they", "the", "and", "for", "while", "over", "under", "after", "before", "review", "message", "messages", "telegram", "gemini", "content", "capture", "article", "note", "bookmark", "video", "publication", "publish", "feedback", "conversation", "comment", "link", "links", "inbound", "outbound", "task", "walkthrough", "implementation", "plan", "diagnosis", "session_summary", "next_steps", "deployment", "checklist", "artifact", "artifacts", "recovered", "recovery", "was", "were":
		return true
	default:
		return false
	}
}

func mapKeys(items map[string]struct{}) []string {
	out := make([]string, 0, len(items))
	for item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}
