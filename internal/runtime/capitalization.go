package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const capitalizationEntryFilePattern = "*.json"

var (
	capitalizationCorpusPathSegments = []string{"knowledge", "corpus"}
	legacyCapitalizationPathSegments = []string{"knowledge", "corpus", "entries"}
)

func newCapitalizationEntry(input SessionFinishInput, event EventEnvelope, review ReviewRoomSummary, actor string) CapitalizationEntry {
	situationItems := []string{}
	if input.CurrentGoal != nil && *input.CurrentGoal != "" {
		situationItems = append(situationItems, *input.CurrentGoal)
	}
	situationItems = append(situationItems, limitStrings(input.OpenRisks, 3)...)
	decisionItems := append([]string{}, limitStrings(input.NextSteps, 3)...)
	if input.HandoffNote != nil && *input.HandoffNote != "" {
		decisionItems = append(decisionItems, *input.HandoffNote)
	}
	if input.Note != nil && *input.Note != "" {
		decisionItems = append(decisionItems, *input.Note)
	}
	costItems := append([]string{}, limitStrings(input.OpenRisks, 3)...)
	if review.OpenCount > 0 {
		costItems = append(costItems, fmt.Sprintf("review_open:%d", review.OpenCount))
	}
	resultItems := []string{"event:" + event.EventID, "actor:" + actor}
	if review.OpenCount > 0 {
		resultItems = append(resultItems, topReviewTexts(review.TopOpen, 1)...)
	}
	return CapitalizationEntry{
		EntryID:       "cap_" + event.EventID,
		CreatedAt:     event.Timestamp,
		Actor:         actor,
		SourceEventID: event.EventID,
		SourceKind:    event.Kind,
		Situation: CapitalizationFacet{
			Summary: input.CurrentFocus,
			Items:   situationItems,
		},
		Decision: CapitalizationFacet{
			Summary: "Session closed with explicit next steps",
			Items:   decisionItems,
		},
		Cost: CapitalizationFacet{
			Summary: "Carry-over cost and queue pressure",
			Items:   costItems,
		},
		Result: CapitalizationFacet{
			Summary: "Memory updated and session evidence emitted",
			Items:   resultItems,
		},
	}
}

func (s *Service) ListCapitalizationEntries() []CapitalizationEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	items, err := s.disk.loadCapitalizationEntries()
	if err != nil {
		return []CapitalizationEntry{}
	}
	return items
}

func (s *Service) DebugAppendCapitalizationEntry(item CapitalizationEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.disk.appendCapitalizationEntry(item)
}

func (d *diskStore) appendCapitalizationEntry(item CapitalizationEntry) error {
	dir := d.capitalizationEntriesDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	entryID := strings.TrimSpace(item.EntryID)
	if entryID == "" {
		return fmt.Errorf("capitalization entry_id is required")
	}
	if strings.Contains(entryID, "/") || strings.Contains(entryID, `\\`) || strings.Contains(entryID, "..") {
		return fmt.Errorf("capitalization entry_id must not contain path traversal segments")
	}
	return writeJSONAtomic(filepath.Join(dir, entryID+".json"), item)
}

func (d *diskStore) loadCapitalizationEntries() ([]CapitalizationEntry, error) {
	itemsByID := map[string]CapitalizationEntry{}
	for _, dir := range d.capitalizationEntryReadDirs() {
		paths, err := capitalizationEntryPaths(dir)
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			item, err := readJSON(path, CapitalizationEntry{})
			if err != nil {
				return nil, err
			}
			key := item.EntryID
			if key == "" {
				key = filepath.Base(path)
			}
			if _, exists := itemsByID[key]; exists {
				continue
			}
			itemsByID[key] = item
		}
	}
	if len(itemsByID) == 0 {
		return []CapitalizationEntry{}, nil
	}
	items := make([]CapitalizationEntry, 0, len(itemsByID))
	for _, item := range itemsByID {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items, nil
}

func (d *diskStore) capitalizationEntriesDir() string {
	return joinCapitalizationPath(d.baseDir, capitalizationCorpusPathSegments)
}

func (d *diskStore) capitalizationEntryReadDirs() []string {
	return []string{d.capitalizationEntriesDir(), d.legacyCapitalizationEntriesDir()}
}

func (d *diskStore) legacyCapitalizationEntriesDir() string {
	return joinCapitalizationPath(d.baseDir, legacyCapitalizationPathSegments)
}

func capitalizationEntryPaths(dir string) ([]string, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return filepath.Glob(filepath.Join(dir, capitalizationEntryFilePattern))
}

func joinCapitalizationPath(baseDir string, segments []string) string {
	parts := make([]string, 0, len(segments)+1)
	parts = append(parts, baseDir)
	parts = append(parts, segments...)
	return filepath.Join(parts...)
}

func newAgentJobCapitalizationEntry(item AgentJob, event EventEnvelope, review ReviewRoomSummary, actor string) CapitalizationEntry {
	situationItems := []string{"role:" + item.Role, "kind:" + item.Kind, "stage:" + string(item.Stage), "source:" + item.Source}
	if item.Ref != nil && strings.TrimSpace(*item.Ref) != "" {
		situationItems = append(situationItems, "ref:"+strings.TrimSpace(*item.Ref))
	}
	// Extract LLM response for corpus surfacing.
	agentResponse := ""
	if item.Result != nil {
		if raw, ok := item.Result["response"].(string); ok {
			agentResponse = limitText(strings.TrimSpace(raw), 300)
		}
	}
	decisionItems := append([]string{}, limitStrings(item.Notes, 3)...)
	if agentResponse == "" {
		decisionItems = append(decisionItems, summarizeJSONObject(item.Result, 3)...)
	}
	costItems := []string{}
	if item.Result != nil {
		for _, key := range []string{"attempt_count", "last_http_status", "last_error"} {
			if value, ok := item.Result[key]; ok {
				costItems = append(costItems, key+":"+stringifyCapitalizationValue(value))
			}
		}
	}
	if review.OpenCount > 0 {
		costItems = append(costItems, fmt.Sprintf("review_open:%d", review.OpenCount))
	}
	resultItems := []string{"status:" + item.Status, "event:" + event.EventID, "actor:" + actor}
	if item.Result != nil {
		for _, key := range []string{"gateway_call_id", "provider", "model", "dispatch_state", "report_source", "source_channel", "source_family", "source_run_id", "antigravity_sync_key", "gemini_sync_key", "artifact_topic", "artifact_type", "artifact_stem", "artifact_relative_path", "severity"} {
			if value, ok := item.Result[key]; ok {
				resultItems = append(resultItems, key+":"+stringifyCapitalizationValue(value))
			}
		}
	}
	return CapitalizationEntry{
		EntryID:       "cap_" + event.EventID,
		CreatedAt:     event.Timestamp,
		Actor:         actor,
		SourceEventID: event.EventID,
		SourceKind:    event.Kind,
		Situation: CapitalizationFacet{
			Summary: item.Summary,
			Items:   limitStrings(situationItems, 5),
		},
		Decision: CapitalizationFacet{
			Summary: func() string {
				if agentResponse != "" {
					return agentResponse
				}
				return "Agent job terminal result recorded through official daemon path"
			}(),
			Items: limitStrings(decisionItems, 5),
		},
		Cost: CapitalizationFacet{
			Summary: "Dispatch/runtime cost markers",
			Items:   limitStrings(costItems, 5),
		},
		Result: CapitalizationFacet{
			Summary: "Agent result persisted into runtime evidence and corpus",
			Items:   limitStrings(resultItems, 12),
		},
	}
}

func summarizeJSONObject(items map[string]any, max int) []string {
	if len(items) == 0 || max <= 0 {
		return []string{}
	}
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, key+":"+stringifyCapitalizationValue(items[key]))
		if len(out) >= max {
			break
		}
	}
	return out
}

func stringifyCapitalizationValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case int:
		return fmt.Sprintf("%d", typed)
	case int64:
		return fmt.Sprintf("%d", typed)
	case float64:
		return fmt.Sprintf("%g", typed)
	case []string:
		return strings.Join(limitStrings(typed, 3), "|")
	case []any:
		parts := []string{}
		for _, item := range typed {
			parts = append(parts, stringifyCapitalizationValue(item))
			if len(parts) >= 3 {
				break
			}
		}
		return strings.Join(parts, "|")
	default:
		return fmt.Sprintf("%v", typed)
	}
}
