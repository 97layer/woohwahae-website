package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

type ingestServiceStub struct {
	auth                 runtime.AuthStatus
	observationsByRef    map[string][]runtime.ObservationRecord
	createdObservations  []runtime.ObservationRecord
	createdConversations []runtime.ConversationNoteInput
	jobs                 []runtime.AgentJob
	reportedJobID        string
	reportedStatus       string
	reportedNotes        []string
	reportedResult       map[string]any
	reportResult         runtime.AgentJobReportResult
	corpusRecover        runtime.CorpusMarkdownRecoverResult
	corpusRecoverErr     error
	corpusLimit          int
	corpusCleanup        bool
	corpusDryRun         bool
}

func (s *ingestServiceStub) AuthStatus() runtime.AuthStatus { return s.auth }
func (s *ingestServiceStub) ListObservations(query runtime.ObservationQuery) []runtime.ObservationRecord {
	if s.observationsByRef == nil {
		return nil
	}
	ref := strings.TrimSpace(query.Ref)
	if ref == "" {
		items := []runtime.ObservationRecord{}
		seen := map[string]bool{}
		for _, bucket := range s.observationsByRef {
			for _, item := range bucket {
				if seen[item.ObservationID] {
					continue
				}
				seen[item.ObservationID] = true
				items = append(items, item)
			}
		}
		return append([]runtime.ObservationRecord{}, items...)
	}
	return append([]runtime.ObservationRecord{}, s.observationsByRef[ref]...)
}
func (s *ingestServiceStub) CreateObservation(item runtime.ObservationRecord) (runtime.ObservationRecord, error) {
	s.appendObservation(item)
	return item, nil
}
func (s *ingestServiceStub) CreateConversationNote(input runtime.ConversationNoteInput) (runtime.ConversationNoteResult, error) {
	s.createdConversations = append(s.createdConversations, input)
	capturedAt := time.Now().UTC()
	if input.CapturedAt != nil && !input.CapturedAt.IsZero() {
		capturedAt = input.CapturedAt.UTC()
	}
	summary := strings.TrimSpace(input.Summary)
	if summary == "" {
		summary = strings.TrimSpace(input.Text)
	}
	sourceChannel := strings.TrimSpace(strings.TrimPrefix(input.SourceChannel, "conversation:"))
	observation := runtime.ObservationRecord{
		ObservationID:     "observation_001",
		SourceChannel:     "conversation:" + sourceChannel,
		CapturedAt:        capturedAt,
		Actor:             strings.TrimSpace(input.Actor),
		Topic:             "conversation.note",
		Refs:              append([]string{}, input.Refs...),
		Confidence:        strings.TrimSpace(input.Confidence),
		RawExcerpt:        strings.TrimSpace(input.Text),
		NormalizedSummary: summary,
	}
	s.appendObservation(observation)
	return runtime.ConversationNoteResult{
		ConversationID: "conversation_001",
		Observation:    observation,
		Warnings:       []string{},
	}, nil
}
func (s *ingestServiceStub) ListAgentJobs() []runtime.AgentJob {
	return append([]runtime.AgentJob{}, s.jobs...)
}
func (s *ingestServiceStub) ReportAgentJob(jobID string, status string, notes []string, result map[string]any) (runtime.AgentJobReportResult, error) {
	s.reportedJobID = jobID
	s.reportedStatus = status
	s.reportedNotes = append([]string{}, notes...)
	s.reportedResult = result
	if s.reportResult.Job.JobID == "" {
		s.reportResult.Job = runtime.AgentJob{JobID: jobID, Status: status}
		s.reportResult.Event = runtime.EventEnvelope{EventID: "event_001", Kind: "agent_job." + status, Timestamp: time.Now().UTC()}
	}
	return s.reportResult, nil
}

func (s *ingestServiceStub) RecoverCorpusMarkdown(limit int, cleanup bool, dryRun bool) (runtime.CorpusMarkdownRecoverResult, error) {
	s.corpusLimit = limit
	s.corpusCleanup = cleanup
	s.corpusDryRun = dryRun
	return s.corpusRecover, s.corpusRecoverErr
}

func (s *ingestServiceStub) appendObservation(item runtime.ObservationRecord) {
	s.createdObservations = append(s.createdObservations, item)
	if s.observationsByRef == nil {
		s.observationsByRef = map[string][]runtime.ObservationRecord{}
	}
	for _, ref := range item.Refs {
		s.observationsByRef[ref] = append(s.observationsByRef[ref], item)
	}
}

func TestRunIngestAntigravityDryRunScansWithoutWrites(t *testing.T) {
	root := seedAntigravityRun(t)
	service := &ingestServiceStub{}
	raw := captureStdout(t, func() {
		runIngest(service, []string{"antigravity", "--root", root, "--dry-run"})
	})
	if len(service.createdObservations) != 0 || service.reportedJobID != "" {
		t.Fatalf("expected dry-run to avoid writes, got created=%d reported=%q", len(service.createdObservations), service.reportedJobID)
	}
	if !strings.Contains(raw, "run-001") || !strings.Contains(raw, "Task summary from metadata") {
		t.Fatalf("unexpected dry-run output: %s", raw)
	}
}

func TestRunIngestAntigravityCreatesObservationAndSkipsDuplicate(t *testing.T) {
	root := seedAntigravityRun(t)
	service := &ingestServiceStub{}
	runIngest(service, []string{"antigravity", "--root", root})
	if len(service.createdObservations) != 1 {
		t.Fatalf("expected first ingest to create one observation, got %d", len(service.createdObservations))
	}
	raw := captureStdout(t, func() {
		runIngest(service, []string{"antigravity", "--root", root})
	})
	if len(service.createdObservations) != 1 {
		t.Fatalf("expected duplicate ingest to be skipped, got %d created observations", len(service.createdObservations))
	}
	if !strings.Contains(raw, "skipped_observation_refs") || !strings.Contains(raw, "antigravity_sync:") {
		t.Fatalf("expected skipped observation refs in output: %s", raw)
	}
}

func TestRunIngestAntigravityReportsJobWhenRequested(t *testing.T) {
	root := seedAntigravityRun(t)
	service := &ingestServiceStub{jobs: []runtime.AgentJob{{JobID: "job_001", Status: "queued", Result: map[string]any{}}}}
	raw := captureStdout(t, func() {
		runIngest(service, []string{"antigravity", "--root", root, "--job", "job_001", "--status", "succeeded"})
	})
	if service.reportedJobID != "job_001" || service.reportedStatus != "succeeded" {
		t.Fatalf("unexpected job report call: id=%q status=%q", service.reportedJobID, service.reportedStatus)
	}
	if service.reportedResult["report_source"] != "antigravity" || service.reportedResult["source_run_id"] != "run-001" {
		t.Fatalf("unexpected reported result: %+v", service.reportedResult)
	}
	artifacts, ok := service.reportedResult["artifacts"].([]string)
	if !ok || len(artifacts) == 0 {
		t.Fatalf("expected artifacts in job report result, got %+v", service.reportedResult)
	}
	if !strings.Contains(raw, "agent_job.succeeded") {
		t.Fatalf("expected job report output, got %s", raw)
	}
}

func TestRunIngestTelegramCreatesObservationAndSkipsDuplicate(t *testing.T) {
	service := &ingestServiceStub{}
	runIngest(service, []string{"telegram", "--title", "Founder shared a YouTube link", "--excerpt", "Need to study this recommender angle.", "--url", "https://youtube.com/watch?v=abc123", "--kind", "link", "--direction", "inbound", "--message-id", "msg-001", "--chat", "room-a", "--username", "founder", "--tags", "recommendation,media"})
	if len(service.createdObservations) != 1 {
		t.Fatalf("expected telegram ingest to create one observation, got %d", len(service.createdObservations))
	}
	created := service.createdObservations[0]
	if created.SourceChannel != "conversation:telegram" || !containsString(created.Refs, "telegram_message:msg_001") || !containsString(created.Refs, "content_host:youtube_com") {
		t.Fatalf("unexpected telegram observation: %+v", created)
	}
	raw := captureStdout(t, func() {
		runIngest(service, []string{"telegram", "--title", "Founder shared a YouTube link", "--excerpt", "Need to study this recommender angle.", "--message-id", "msg-001"})
	})
	if len(service.createdObservations) != 1 {
		t.Fatalf("expected duplicate telegram ingest to be skipped, got %d", len(service.createdObservations))
	}
	if !strings.Contains(raw, "matching dedupe ref already recorded") || !strings.Contains(raw, "telegram_message:msg_001") {
		t.Fatalf("expected telegram duplicate output, got %s", raw)
	}
}

func TestRunIngestContentReadsExcerptFileAndClassifiesNotebook(t *testing.T) {
	service := &ingestServiceStub{}
	excerptFile := filepath.Join(t.TempDir(), "note.txt")
	mustWrite(t, excerptFile, "This NotebookLM note tracks second-brain clustering and retrieval patterns.")
	raw := captureStdout(t, func() {
		runIngest(service, []string{"content", "--channel", "notebook-lm", "--title", "Second brain note", "--excerpt-file", excerptFile, "--url", "https://example.com/research/clustering", "--author", "Founder", "--kind", "note", "--doc-id", "note-001", "--tags", "research,second-brain"})
	})
	if len(service.createdObservations) != 1 {
		t.Fatalf("expected content ingest to create one observation, got %d", len(service.createdObservations))
	}
	created := service.createdObservations[0]
	if created.SourceChannel != "conversation:notebook_lm" || !containsString(created.Refs, "content_doc:note_001") || !containsString(created.Refs, "source_family:founder_archive") {
		t.Fatalf("unexpected notebook observation: %+v", created)
	}
	if !strings.Contains(raw, "\"mode\": \"conversation\"") || !strings.Contains(raw, "content_doc:note_001") {
		t.Fatalf("expected notebook ingest output, got %s", raw)
	}
}

func TestRunIngestRSSCreatesObservationsAndSkipsDuplicate(t *testing.T) {
	oldFetch := fetchRSSFeedBody
	fetchRSSFeedBody = func(feedURL string) ([]byte, error) {
		return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Layer Feed</title>
    <link>https://example.com/feed</link>
    <item>
      <title>First article</title>
      <link>https://example.com/post-1</link>
      <guid>post-1</guid>
      <description>First excerpt.</description>
      <pubDate>Sat, 14 Mar 2026 00:30:00 +0900</pubDate>
      <category>strategy</category>
    </item>
    <item>
      <title>Second article</title>
      <link>https://example.com/post-2</link>
      <guid>post-2</guid>
      <description>Second excerpt.</description>
    </item>
  </channel>
</rss>`), nil
	}
	defer func() { fetchRSSFeedBody = oldFetch }()

	service := &ingestServiceStub{}
	runIngest(service, []string{"rss", "--feed", "https://example.com/feed.xml", "--limit", "1", "--tags", "signals"})
	if len(service.createdObservations) != 2 {
		t.Fatalf("expected rss ingest to create content + source intake observations, got %d", len(service.createdObservations))
	}
	contentCreated := service.createdObservations[0]
	if contentCreated.SourceChannel != "crawler" || !containsString(contentCreated.Refs, "content_doc:post_1") || !containsString(contentCreated.Refs, "feed_kind:rss") {
		t.Fatalf("unexpected rss content observation: %+v", contentCreated)
	}
	sourceIntakeCreated := service.createdObservations[1]
	if sourceIntakeCreated.Topic != runtime.SourceIntakeTopic || !containsString(sourceIntakeCreated.Refs, "source_observation:"+contentCreated.ObservationID) || !containsString(sourceIntakeCreated.Refs, "content_doc:post_1") {
		t.Fatalf("unexpected rss source intake observation: %+v", sourceIntakeCreated)
	}
	raw := captureStdout(t, func() {
		runIngest(service, []string{"rss", "--feed", "https://example.com/feed.xml", "--limit", "1"})
	})
	if len(service.createdObservations) != 2 {
		t.Fatalf("expected duplicate rss ingest to reuse both observations, got %d", len(service.createdObservations))
	}
	if !strings.Contains(raw, "skipped_observation_refs") || !strings.Contains(raw, "content_doc:post_1") || !strings.Contains(raw, "\"created_source_intakes\": []") {
		t.Fatalf("expected rss duplicate output, got %s", raw)
	}
}

func TestRunIngestRSSDryRunShowsFeedItems(t *testing.T) {
	oldFetch := fetchRSSFeedBody
	fetchRSSFeedBody = func(feedURL string) ([]byte, error) {
		return []byte(`<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Layer Atom</title>
  <entry>
    <id>tag:example.com,2026:1</id>
    <title>Atom note</title>
    <link href="https://example.com/atom-note" rel="alternate"></link>
    <summary>Atom excerpt.</summary>
    <updated>2026-03-14T00:45:00+09:00</updated>
  </entry>
</feed>`), nil
	}
	defer func() { fetchRSSFeedBody = oldFetch }()

	service := &ingestServiceStub{}
	raw := captureStdout(t, func() {
		runIngest(service, []string{"rss", "--feed", "https://example.com/atom.xml", "--dry-run"})
	})
	if len(service.createdObservations) != 0 {
		t.Fatalf("expected dry-run to avoid writes, got %d observations", len(service.createdObservations))
	}
	if !strings.Contains(raw, "\"kind\": \"atom\"") || !strings.Contains(raw, "Atom note") {
		t.Fatalf("unexpected rss dry-run output: %s", raw)
	}
}

func seedAntigravityRun(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runRoot := filepath.Join(root, "run-001")
	if err := os.MkdirAll(runRoot, 0o755); err != nil {
		t.Fatalf("mkdir run root: %v", err)
	}
	mustWrite(t, filepath.Join(runRoot, "task.md.resolved"), "# Primary task\n\nProposalItem proposal_001 remains open.")
	mustWrite(t, filepath.Join(runRoot, "task.md.metadata.json"), `{"artifactType":"ARTIFACT_TYPE_TASK","summary":"Task summary from metadata","updatedAt":"2026-03-08T10:29:56.121373Z","version":"2"}`)
	mustWrite(t, filepath.Join(runRoot, "media__1771917569587.png"), "png")
	return root
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func TestRunIngestGeminiCreatesObservationReportsJobAndCleanup(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "task.md.resolved"), `# Task

proposal_001 remains open.`)
	mustWrite(t, filepath.Join(root, "task.md.metadata.json"), `{"summary":"Recovered Gemini task","updatedAt":"2026-03-08T10:29:56Z","version":"3"}`)
	service := &ingestServiceStub{jobs: []runtime.AgentJob{{JobID: "job_001", Status: "queued", Result: map[string]any{}}}}
	raw := captureStdout(t, func() {
		runIngest(service, []string{"gemini", "--root", root, "--job", "job_001", "--status", "succeeded", "--cleanup"})
	})
	if len(service.createdObservations) != 1 {
		t.Fatalf("expected gemini ingest to create one observation, got %d", len(service.createdObservations))
	}
	created := service.createdObservations[0]
	if created.SourceChannel != "gemini" || !containsString(created.Refs, "gemini_sync:"+strings.TrimSpace(service.reportedResult["gemini_sync_key"].(string))) {
		t.Fatalf("unexpected gemini observation: %+v result=%+v", created, service.reportedResult)
	}
	if service.reportedJobID != "job_001" || service.reportedResult["report_source"] != "gemini" {
		t.Fatalf("unexpected gemini job report: id=%q result=%+v", service.reportedJobID, service.reportedResult)
	}
	if _, err := os.Stat(filepath.Join(root, "task.md.resolved")); !os.IsNotExist(err) {
		t.Fatalf("expected cleanup to remove task artifact, err=%v", err)
	}
	if !strings.Contains(raw, "cleaned_paths") || !strings.Contains(raw, "task.md.metadata.json") {
		t.Fatalf("expected cleanup paths in output, got %s", raw)
	}
}

func TestRunIngestGeminiSkipsDuplicateObservation(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "walkthrough.md.resolved"), `# Walkthrough

thread_001 is still open.`)
	service := &ingestServiceStub{}
	runIngest(service, []string{"gemini", "--root", root})
	if len(service.createdObservations) != 1 {
		t.Fatalf("expected first gemini ingest to create one observation, got %d", len(service.createdObservations))
	}
	raw := captureStdout(t, func() {
		runIngest(service, []string{"gemini", "--root", root})
	})
	if len(service.createdObservations) != 1 {
		t.Fatalf("expected duplicate gemini ingest to be skipped, got %d created observations", len(service.createdObservations))
	}
	if !strings.Contains(raw, "skipped_observation_refs") || !strings.Contains(raw, "gemini_sync:") {
		t.Fatalf("expected skipped gemini observation refs in output: %s", raw)
	}
}

func TestRunIngestCorpusForwardsRecoveryRequest(t *testing.T) {
	service := &ingestServiceStub{corpusRecover: runtime.CorpusMarkdownRecoverResult{
		Considered: 1,
		Created:    1,
		Cleaned:    1,
		Artifacts:  []runtime.CorpusMarkdownArtifact{{RelativePath: "knowledge/corpus/entries/analysis.md", Summary: "Recovered analysis"}},
	}}
	raw := captureStdout(t, func() {
		runIngest(service, []string{"corpus", "--cleanup"})
	})
	if service.corpusLimit != 0 || !service.corpusCleanup || service.corpusDryRun {
		t.Fatalf("unexpected corpus recovery call: limit=%d cleanup=%t dry=%t", service.corpusLimit, service.corpusCleanup, service.corpusDryRun)
	}
	if !strings.Contains(raw, "knowledge/corpus/entries/analysis.md") || !strings.Contains(raw, "created") {
		t.Fatalf("unexpected corpus ingest output: %s", raw)
	}
}
