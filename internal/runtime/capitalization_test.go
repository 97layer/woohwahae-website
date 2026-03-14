package runtime

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCapitalization(t *testing.T) {
	t.Run("writes_entries_to_corpus_root", func(t *testing.T) {
		service, err := NewService(t.TempDir())
		if err != nil {
			t.Fatalf("new service: %v", err)
		}

		goal := "Preserve session intelligence"
		result, err := service.FinishSession(SessionFinishInput{
			CurrentFocus: "Stabilize daemon path",
			CurrentGoal:  &goal,
			NextSteps:    []string{"verify api"},
			OpenRisks:    []string{"drift"},
		})
		if err != nil {
			t.Fatalf("finish session: %v", err)
		}

		entries := service.ListCapitalizationEntries()
		if len(entries) != 1 {
			t.Fatalf("expected 1 capitalization entry, got %+v", entries)
		}
		if entries[0].SourceEventID != result.Event.EventID || entries[0].Situation.Summary != "Stabilize daemon path" {
			t.Fatalf("unexpected capitalization entry: %+v", entries[0])
		}

		rootPaths, err := filepath.Glob(filepath.Join(service.disk.capitalizationEntriesDir(), capitalizationEntryFilePattern))
		if err != nil {
			t.Fatalf("glob root entries: %v", err)
		}
		if len(rootPaths) != 1 {
			t.Fatalf("expected 1 root capitalization file, got %v", rootPaths)
		}
		if filepath.Base(rootPaths[0]) != entries[0].EntryID+".json" {
			t.Fatalf("expected root file for %s, got %s", entries[0].EntryID, filepath.Base(rootPaths[0]))
		}
		if _, err := os.Stat(service.disk.legacyCapitalizationEntriesDir()); !os.IsNotExist(err) {
			t.Fatalf("expected legacy capitalization directory to be absent, got err=%v", err)
		}
	})

	t.Run("reads_legacy_entries_as_fallback", func(t *testing.T) {
		service, err := NewService(t.TempDir())
		if err != nil {
			t.Fatalf("new service: %v", err)
		}

		legacyEntry := testCapitalizationEntry(
			"cap_legacy_only",
			"legacy summary",
			time.Date(2026, time.March, 8, 1, 0, 0, 0, time.UTC),
		)
		writeCapitalizationFixture(t, service.disk.legacyCapitalizationEntriesDir(), legacyEntry)

		entries := service.ListCapitalizationEntries()
		if len(entries) != 1 {
			t.Fatalf("expected 1 capitalization entry, got %+v", entries)
		}
		if entries[0].EntryID != legacyEntry.EntryID || entries[0].Situation.Summary != legacyEntry.Situation.Summary {
			t.Fatalf("expected legacy fallback entry %+v, got %+v", legacyEntry, entries[0])
		}
	})

	t.Run("dedupes_root_and_legacy_entries_by_entry_id", func(t *testing.T) {
		service, err := NewService(t.TempDir())
		if err != nil {
			t.Fatalf("new service: %v", err)
		}

		rootEntry := testCapitalizationEntry(
			"cap_duplicate",
			"root summary wins",
			time.Date(2026, time.March, 8, 2, 0, 0, 0, time.UTC),
		)
		legacyEntry := rootEntry
		legacyEntry.Situation.Summary = "legacy summary loses"

		writeCapitalizationFixture(t, service.disk.capitalizationEntriesDir(), rootEntry)
		writeCapitalizationFixture(t, service.disk.legacyCapitalizationEntriesDir(), legacyEntry)

		entries := service.ListCapitalizationEntries()
		if len(entries) != 1 {
			t.Fatalf("expected deduped capitalization entry, got %+v", entries)
		}
		if entries[0].EntryID != rootEntry.EntryID {
			t.Fatalf("expected entry id %s, got %+v", rootEntry.EntryID, entries[0])
		}
		if entries[0].Situation.Summary != rootEntry.Situation.Summary {
			t.Fatalf("expected root entry to win dedupe, got %+v", entries[0])
		}
	})
}

func testCapitalizationEntry(entryID, summary string, createdAt time.Time) CapitalizationEntry {
	return CapitalizationEntry{
		EntryID:       entryID,
		CreatedAt:     createdAt,
		Actor:         "system",
		SourceEventID: entryID + ":event",
		SourceKind:    "session.finished",
		Situation: CapitalizationFacet{
			Summary: summary,
		},
		Decision: CapitalizationFacet{
			Summary: "decision",
		},
		Cost: CapitalizationFacet{
			Summary: "cost",
		},
		Result: CapitalizationFacet{
			Summary: "result",
		},
	}
}

func writeCapitalizationFixture(t *testing.T, dir string, entry CapitalizationEntry) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	path := filepath.Join(dir, entry.EntryID+".json")
	if err := writeJSONAtomic(path, entry); err != nil {
		t.Fatalf("write capitalization fixture %s: %v", path, err)
	}
}
