package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

type ingestService interface {
	authStatusProvider
	ListObservations(query runtime.ObservationQuery) []runtime.ObservationRecord
	CreateConversationNote(input runtime.ConversationNoteInput) (runtime.ConversationNoteResult, error)
	CreateObservation(item runtime.ObservationRecord) (runtime.ObservationRecord, error)
	ListAgentJobs() []runtime.AgentJob
	ReportAgentJob(jobID string, status string, notes []string, result map[string]any) (runtime.AgentJobReportResult, error)
	RecoverCorpusMarkdown(limit int, cleanup bool, dryRun bool) (runtime.CorpusMarkdownRecoverResult, error)
}

type antigravityIngestOutput struct {
	Root                   string                        `json:"root"`
	DryRun                 bool                          `json:"dry_run"`
	Runs                   []runtime.AntigravityRun      `json:"runs"`
	CreatedObservations    []runtime.ObservationRecord   `json:"created_observations"`
	SkippedObservationRefs []string                      `json:"skipped_observation_refs"`
	SelectedRun            *runtime.AntigravityRun       `json:"selected_run,omitempty"`
	JobResultPreview       map[string]any                `json:"job_result_preview,omitempty"`
	JobReport              *runtime.AgentJobReportResult `json:"job_report,omitempty"`
	SkippedJobReport       bool                          `json:"skipped_job_report"`
	SkippedJobReason       string                        `json:"skipped_job_reason,omitempty"`
}

type observationIngestOutput struct {
	Adapter         string                    `json:"adapter"`
	DryRun          bool                      `json:"dry_run"`
	Mode            string                    `json:"mode,omitempty"`
	ConversationID  string                    `json:"conversation_id,omitempty"`
	Observation     runtime.ObservationRecord `json:"observation"`
	Proposal        *runtime.ProposalItem     `json:"proposal,omitempty"`
	Job             *runtime.AgentJob         `json:"job,omitempty"`
	ReviewItem      *runtime.ReviewRoomItem   `json:"review_item,omitempty"`
	Warnings        []string                  `json:"warnings,omitempty"`
	Created         bool                      `json:"created"`
	Duplicate       bool                      `json:"duplicate"`
	DedupeRef       *string                   `json:"dedupe_ref,omitempty"`
	DuplicateReason string                    `json:"duplicate_reason,omitempty"`
}

type geminiIngestOutput struct {
	Root                   string                        `json:"root"`
	DryRun                 bool                          `json:"dry_run"`
	Cleanup                bool                          `json:"cleanup"`
	Artifacts              []runtime.GeminiArtifact      `json:"artifacts"`
	CreatedObservations    []runtime.ObservationRecord   `json:"created_observations"`
	SkippedObservationRefs []string                      `json:"skipped_observation_refs"`
	SelectedArtifact       *runtime.GeminiArtifact       `json:"selected_artifact,omitempty"`
	JobResultPreview       map[string]any                `json:"job_result_preview,omitempty"`
	JobReport              *runtime.AgentJobReportResult `json:"job_report,omitempty"`
	SkippedJobReport       bool                          `json:"skipped_job_report"`
	SkippedJobReason       string                        `json:"skipped_job_reason,omitempty"`
	CleanedPaths           []string                      `json:"cleaned_paths,omitempty"`
	CleanupErrors          []string                      `json:"cleanup_errors,omitempty"`
}

const ingestUsage = "usage: layer-osctl ingest <antigravity|telegram|content|rss|gemini|corpus> ..."

func runIngest(service ingestService, args []string) {
	if len(args) == 0 {
		log.Fatal(ingestUsage)
	}
	switch args[0] {
	case "antigravity":
		runAntigravityIngest(service, args[1:])
	case "telegram":
		runTelegramIngest(service, args[1:])
	case "content":
		runContentIngest(service, args[1:])
	case "rss":
		runRSSIngest(service, args[1:])
	case "gemini":
		runGeminiIngest(service, args[1:])
	case "corpus":
		runCorpusIngest(service, args[1:])
	default:
		log.Fatal(ingestUsage)
	}
}

func runAntigravityIngest(service ingestService, args []string) {
	cmd := flag.NewFlagSet("ingest antigravity", flag.ExitOnError)
	root := cmd.String("root", "", "antigravity brain root")
	runID := cmd.String("run", "", "specific antigravity run id")
	limit := cmd.Int("limit", 0, "max runs to scan")
	jobID := cmd.String("job", "", "optional job id to report the selected run into")
	status := cmd.String("status", "succeeded", "terminal job status when --job is set")
	dryRun := cmd.Bool("dry-run", false, "scan and preview without writing observations or job reports")
	parseArgs(cmd, args)

	if *limit < 0 {
		log.Fatal("limit must not be negative")
	}
	if !*dryRun {
		requireCLIWriteAuth(service)
	}

	scan, err := runtime.ScanAntigravityRuns(strings.TrimSpace(*root), strings.TrimSpace(*runID), *limit)
	if err != nil {
		log.Fatal(err)
	}
	output := antigravityIngestOutput{
		Root:                   scan.Root,
		DryRun:                 *dryRun,
		Runs:                   scan.Runs,
		CreatedObservations:    []runtime.ObservationRecord{},
		SkippedObservationRefs: []string{},
	}

	if *dryRun {
		selected := selectedAntigravityRun(scan.Runs)
		if selected != nil {
			output.SelectedRun = selected
			output.JobResultPreview = antigravityJobResult(*selected)
		}
		writeJSON(output)
		return
	}

	for _, run := range scan.Runs {
		for _, artifact := range run.Artifacts {
			obs := runtime.AntigravityObservation(artifact)
			syncRef := antigravitySyncRef(artifact.SyncKey)
			existing := service.ListObservations(runtime.ObservationQuery{SourceChannel: "antigravity", Ref: syncRef, Limit: 1})
			if len(existing) > 0 {
				output.SkippedObservationRefs = append(output.SkippedObservationRefs, syncRef)
				continue
			}
			created, err := service.CreateObservation(obs)
			if err != nil {
				log.Fatal(err)
			}
			output.CreatedObservations = append(output.CreatedObservations, created)
		}
	}

	if job := strings.TrimSpace(*jobID); job != "" {
		selected := selectedAntigravityRun(scan.Runs)
		if selected == nil {
			log.Fatal("no antigravity run available for job report")
		}
		output.SelectedRun = selected
		preview := antigravityJobResult(*selected)
		output.JobResultPreview = preview
		current, ok := findAgentJob(service.ListAgentJobs(), job)
		if !ok {
			log.Fatal("job id not found")
		}
		if current.Result != nil {
			if syncKey, ok := current.Result["antigravity_sync_key"].(string); ok && strings.TrimSpace(syncKey) == selected.SyncKey {
				output.SkippedJobReport = true
				output.SkippedJobReason = "matching antigravity_sync_key already recorded"
				writeJSON(output)
				return
			}
		}
		primary, ok := runtime.AntigravityRunPrimaryArtifact(*selected)
		if !ok {
			log.Fatal("selected run has no primary antigravity artifact")
		}
		notes := []string{
			"ingest:antigravity",
			"antigravity_run:" + selected.RunID,
			"antigravity_artifact:" + primary.Stem,
		}
		result, err := service.ReportAgentJob(job, strings.TrimSpace(*status), notes, preview)
		if err != nil {
			log.Fatal(err)
		}
		output.JobReport = &result
	}

	writeJSON(output)
}

func runTelegramIngest(service ingestService, args []string) {
	cmd := flag.NewFlagSet("ingest telegram", flag.ExitOnError)
	topic := cmd.String("topic", "", "telegram observation topic")
	interaction := cmd.String("interaction", "", "canonical interaction profile (inbound_conversation|outbound_publication|feedback_reply|youtube_link_injection|bookmark_share)")
	summary := cmd.String("summary", "", "normalized summary")
	excerpt := cmd.String("excerpt", "", "raw excerpt")
	excerptFile := cmd.String("excerpt-file", "", "path to excerpt text file")
	title := cmd.String("title", "", "optional title or headline")
	sourceURL := cmd.String("url", "", "optional source url")
	direction := cmd.String("direction", "", "interaction direction (inbound|outbound|feedback)")
	kind := cmd.String("kind", "message", "telegram content kind (message|link|publication|feedback)")
	messageID := cmd.String("message-id", "", "telegram message id for dedupe")
	chatID := cmd.String("chat", "", "telegram chat identifier")
	username := cmd.String("username", "", "telegram username")
	actor := cmd.String("actor", "", "actor name")
	confidence := cmd.String("confidence", "", "confidence override")
	refs := cmd.String("refs", "", "comma-separated semantic refs")
	tags := cmd.String("tags", "", "comma-separated tags")
	capturedAt := cmd.String("captured-at", "", "captured-at RFC3339 timestamp")
	dryRun := cmd.Bool("dry-run", false, "preview without writing observation")
	parseArgs(cmd, args)

	if !*dryRun {
		requireCLIWriteAuth(service)
	}
	text := loadIngestText(*excerpt, *excerptFile)
	captured := parseIngestCapturedAt(*capturedAt)
	input := runtime.TelegramObservationInput{
		Topic:      strings.TrimSpace(*topic),
		Summary:    strings.TrimSpace(*summary),
		Excerpt:    text,
		Title:      strings.TrimSpace(*title),
		SourceURL:  strings.TrimSpace(*sourceURL),
		Direction:  strings.TrimSpace(*direction),
		Kind:       strings.TrimSpace(*kind),
		MessageID:  strings.TrimSpace(*messageID),
		ChatID:     strings.TrimSpace(*chatID),
		Username:   strings.TrimSpace(*username),
		Actor:      strings.TrimSpace(*actor),
		Confidence: strings.TrimSpace(*confidence),
		Tags:       splitCSV(*tags),
		Refs:       splitCSV(*refs),
		CapturedAt: captured,
	}
	if value := strings.TrimSpace(*interaction); value != "" {
		// Let the profile own kind/topic/direction defaults even if the flag
		// provided a non-empty default value.
		input.Kind = ""
		input.Topic = ""
		if profile, ok := runtime.FindTelegramProfile(value); ok {
			input = runtime.ApplyTelegramProfile(profile, input)
		} else {
			log.Fatalf("unknown telegram interaction profile: %s", value)
		}
	}
	obs := runtime.BuildTelegramObservation(input)
	persistConversationIngest(service, "telegram", obs, input.Tags, runtime.TelegramDedupeRef(input.MessageID), *dryRun)
}

func runContentIngest(service ingestService, args []string) {
	cmd := flag.NewFlagSet("ingest content", flag.ExitOnError)
	channel := cmd.String("channel", "crawler", "source channel (crawler|personal_db|notebook_lm)")
	topic := cmd.String("topic", "", "content observation topic")
	profileName := cmd.String("profile", "", "canonical content profile (note|summary|qa|transcript|bookmark|outline|article|video|thread|post|content_capture)")
	title := cmd.String("title", "", "content title")
	summary := cmd.String("summary", "", "normalized summary")
	excerpt := cmd.String("excerpt", "", "raw excerpt")
	excerptFile := cmd.String("excerpt-file", "", "path to excerpt text file")
	sourceURL := cmd.String("url", "", "source url")
	author := cmd.String("author", "", "content author")
	kind := cmd.String("kind", "content_capture", "content kind (article|note|bookmark|video|content_capture)")
	docID := cmd.String("doc-id", "", "stable document id for dedupe")
	actor := cmd.String("actor", "", "actor name")
	confidence := cmd.String("confidence", "", "confidence override")
	refs := cmd.String("refs", "", "comma-separated semantic refs")
	tags := cmd.String("tags", "", "comma-separated tags")
	capturedAt := cmd.String("captured-at", "", "captured-at RFC3339 timestamp")
	dryRun := cmd.Bool("dry-run", false, "preview without writing observation")
	parseArgs(cmd, args)

	if !*dryRun {
		requireCLIWriteAuth(service)
	}
	text := loadIngestText(*excerpt, *excerptFile)
	captured := parseIngestCapturedAt(*capturedAt)
	input := runtime.ContentObservationInput{
		SourceChannel: strings.TrimSpace(*channel),
		Topic:         strings.TrimSpace(*topic),
		Title:         strings.TrimSpace(*title),
		Summary:       strings.TrimSpace(*summary),
		Excerpt:       text,
		SourceURL:     strings.TrimSpace(*sourceURL),
		Author:        strings.TrimSpace(*author),
		Kind:          strings.TrimSpace(*kind),
		DocID:         strings.TrimSpace(*docID),
		Actor:         strings.TrimSpace(*actor),
		Confidence:    strings.TrimSpace(*confidence),
		Refs:          splitCSV(*refs),
		Tags:          splitCSV(*tags),
		CapturedAt:    captured,
	}
	effectiveKind := strings.TrimSpace(*kind)
	if value := strings.TrimSpace(*profileName); value != "" {
		effectiveKind = value
		// Force profile application to set canonical kind/topic even if the flag
		// default provided a non-empty kind.
		input.Kind = ""
		input.Topic = ""
	}
	channelValue := strings.TrimSpace(*channel)
	switch runtime.ObservationChannelFamily(channelValue) {
	case "founder_archive":
		if profile, ok := runtime.FindPersonalContentProfile(channelValue, effectiveKind); ok {
			input = runtime.ApplyPersonalContentProfile(profile, input)
		} else if strings.TrimSpace(*profileName) != "" {
			log.Fatalf("unknown personal content profile: %s:%s", channelValue, effectiveKind)
		}
	default:
		if profile, ok := runtime.FindCrawlerContentProfile(effectiveKind); ok {
			input = runtime.ApplyCrawlerContentProfile(profile, input)
		} else if strings.TrimSpace(*profileName) != "" {
			log.Fatalf("unknown crawler content profile: %s", effectiveKind)
		}
	}
	obs := runtime.BuildContentObservation(input)
	persistConversationIngest(service, "content", obs, input.Tags, runtime.ContentDedupeRef(input.DocID, input.SourceURL), *dryRun)
}

func runCorpusIngest(service ingestService, args []string) {
	cmd := flag.NewFlagSet("ingest corpus", flag.ExitOnError)
	limit := cmd.Int("limit", 0, "max markdown artifacts to scan")
	cleanup := cmd.Bool("cleanup", false, "delete absorbed markdown artifacts after canonical recovery")
	dryRun := cmd.Bool("dry-run", false, "scan and preview without writing corpus entries or deleting files")
	parseArgs(cmd, args)
	if *limit < 0 {
		log.Fatal("limit must not be negative")
	}
	if !*dryRun {
		requireCLIWriteAuth(service)
	}
	result, err := service.RecoverCorpusMarkdown(*limit, *cleanup, *dryRun)
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(result)
}

func runGeminiIngest(service ingestService, args []string) {
	cmd := flag.NewFlagSet("ingest gemini", flag.ExitOnError)
	root := cmd.String("root", "", "workspace root to scan for stray Gemini artifacts")
	limit := cmd.Int("limit", 0, "max artifacts to scan")
	jobID := cmd.String("job", "", "optional job id to report the selected artifact into")
	status := cmd.String("status", "succeeded", "terminal job status when --job is set")
	cleanup := cmd.Bool("cleanup", false, "delete absorbed stray artifact files after ingest succeeds or is already deduplicated")
	dryRun := cmd.Bool("dry-run", false, "scan and preview without writing observations, job reports, or cleanup")
	parseArgs(cmd, args)

	if *limit < 0 {
		log.Fatal("limit must not be negative")
	}
	if !*dryRun {
		requireCLIWriteAuth(service)
	}
	rootValue := strings.TrimSpace(*root)
	if rootValue == "" {
		rootValue = repoRootForLocal()
	}
	scan, err := runtime.ScanGeminiArtifacts(rootValue, *limit)
	if err != nil {
		log.Fatal(err)
	}
	output := geminiIngestOutput{
		Root:                   scan.Root,
		DryRun:                 *dryRun,
		Cleanup:                *cleanup,
		Artifacts:              scan.Artifacts,
		CreatedObservations:    []runtime.ObservationRecord{},
		SkippedObservationRefs: []string{},
		CleanedPaths:           []string{},
		CleanupErrors:          []string{},
	}

	selected := selectedGeminiArtifact(scan.Artifacts)
	if selected != nil {
		output.SelectedArtifact = selected
		output.JobResultPreview = geminiJobResult(*selected)
	}
	if *dryRun {
		writeJSON(output)
		return
	}

	cleanupCandidates := make([]runtime.GeminiArtifact, 0, len(scan.Artifacts))
	for _, artifact := range scan.Artifacts {
		obs := runtime.GeminiObservation(artifact)
		syncRef := geminiSyncRef(artifact.SyncKey)
		existing := service.ListObservations(runtime.ObservationQuery{SourceChannel: obs.SourceChannel, Ref: syncRef, Limit: 1})
		if len(existing) > 0 {
			output.SkippedObservationRefs = append(output.SkippedObservationRefs, syncRef)
			cleanupCandidates = append(cleanupCandidates, artifact)
			continue
		}
		created, err := service.CreateObservation(obs)
		if err != nil {
			log.Fatal(err)
		}
		output.CreatedObservations = append(output.CreatedObservations, created)
		cleanupCandidates = append(cleanupCandidates, artifact)
	}

	if job := strings.TrimSpace(*jobID); job != "" {
		if selected == nil {
			log.Fatal("no gemini artifact available for job report")
		}
		current, ok := findAgentJob(service.ListAgentJobs(), job)
		if !ok {
			log.Fatal("job id not found")
		}
		if current.Result != nil {
			if syncKey, ok := current.Result["gemini_sync_key"].(string); ok && strings.TrimSpace(syncKey) == selected.SyncKey {
				output.SkippedJobReport = true
				output.SkippedJobReason = "matching gemini_sync_key already recorded"
			} else {
				notes := []string{"ingest:gemini", "gemini_artifact:" + selected.Stem}
				result, err := service.ReportAgentJob(job, strings.TrimSpace(*status), notes, geminiJobResult(*selected))
				if err != nil {
					log.Fatal(err)
				}
				output.JobReport = &result
			}
		} else {
			notes := []string{"ingest:gemini", "gemini_artifact:" + selected.Stem}
			result, err := service.ReportAgentJob(job, strings.TrimSpace(*status), notes, geminiJobResult(*selected))
			if err != nil {
				log.Fatal(err)
			}
			output.JobReport = &result
		}
	}

	if *cleanup {
		cleaned, cleanupErrors := cleanupGeminiArtifacts(scan.Root, cleanupCandidates)
		output.CleanedPaths = cleaned
		output.CleanupErrors = cleanupErrors
	}

	writeJSON(output)
}

func persistObservationIngest(service ingestService, adapter string, obs runtime.ObservationRecord, dedupeRef string, dryRun bool) {
	output := observationIngestOutput{
		Adapter:     adapter,
		DryRun:      dryRun,
		Observation: obs,
	}
	if ref := strings.TrimSpace(dedupeRef); ref != "" {
		output.DedupeRef = &ref
		if !dryRun {
			existing := service.ListObservations(runtime.ObservationQuery{SourceChannel: obs.SourceChannel, Ref: ref, Limit: 1})
			if len(existing) > 0 {
				output.Duplicate = true
				output.DuplicateReason = "matching dedupe ref already recorded"
				writeJSON(output)
				return
			}
		}
	}
	if dryRun {
		writeJSON(output)
		return
	}
	created, err := service.CreateObservation(obs)
	if err != nil {
		log.Fatal(err)
	}
	output.Observation = created
	output.Created = true
	writeJSON(output)
}

func persistConversationIngest(service ingestService, adapter string, obs runtime.ObservationRecord, tags []string, dedupeRef string, dryRun bool) {
	output := observationIngestOutput{
		Adapter:     adapter,
		DryRun:      dryRun,
		Mode:        "conversation",
		Observation: obs,
	}
	if ref := strings.TrimSpace(dedupeRef); ref != "" {
		output.DedupeRef = &ref
		if !dryRun {
			existing := service.ListObservations(runtime.ObservationQuery{Ref: ref, Limit: 1})
			if len(existing) > 0 {
				output.Duplicate = true
				output.DuplicateReason = "matching dedupe ref already recorded"
				writeJSON(output)
				return
			}
		}
	}
	if dryRun {
		writeJSON(output)
		return
	}
	result, err := service.CreateConversationNote(runtime.ConversationNoteInputFromObservation(obs, tags))
	if err != nil {
		log.Fatal(err)
	}
	output.Observation = result.Observation
	output.ConversationID = result.ConversationID
	output.Proposal = result.Proposal
	output.Job = result.Job
	output.ReviewItem = result.ReviewItem
	output.Warnings = append([]string{}, result.Warnings...)
	output.Created = true
	writeJSON(output)
}

func loadIngestText(raw string, path string) string {
	if value := strings.TrimSpace(path); value != "" {
		content, err := os.ReadFile(value)
		if err != nil {
			log.Fatalf("read excerpt file: %v", err)
		}
		return strings.TrimSpace(string(content))
	}
	return strings.TrimSpace(raw)
}

func parseIngestCapturedAt(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		log.Fatalf("invalid captured-at: %v", err)
	}
	return parsed.UTC()
}

func antigravitySyncRef(syncKey string) string {
	return "antigravity_sync:" + strings.TrimSpace(syncKey)
}

func selectedAntigravityRun(runs []runtime.AntigravityRun) *runtime.AntigravityRun {
	if len(runs) == 0 {
		return nil
	}
	run := runs[0]
	return &run
}

func findAgentJob(items []runtime.AgentJob, jobID string) (runtime.AgentJob, bool) {
	for _, item := range items {
		if strings.TrimSpace(item.JobID) == strings.TrimSpace(jobID) {
			return item, true
		}
	}
	return runtime.AgentJob{}, false
}

func antigravityJobResult(run runtime.AntigravityRun) map[string]any {
	sourceChannel := "antigravity"
	sourceFamily := runtime.ObservationChannelFamily(sourceChannel)
	if primary, ok := runtime.AntigravityRunPrimaryArtifact(run); !ok {
		return map[string]any{
			"summary":              "Antigravity run has no primary artifact",
			"artifacts":            append([]string{}, run.MediaPaths...),
			"artifact_count":       0,
			"report_source":        sourceChannel,
			"source_channel":       sourceChannel,
			"source_family":        sourceFamily,
			"source_run_id":        run.RunID,
			"antigravity_sync_key": run.SyncKey,
			"severity":             "low",
		}
	} else {
		artifacts := []string{primary.PrimaryPath}
		for _, item := range run.Artifacts {
			if item.PrimaryPath == primary.PrimaryPath {
				continue
			}
			artifacts = append(artifacts, item.PrimaryPath)
		}
		artifacts = append(artifacts, run.MediaPaths...)
		sort.Strings(artifacts)
		refs := append([]string{}, primary.Refs...)
		sort.Strings(refs)
		result := map[string]any{
			"summary":              primary.Summary,
			"artifacts":            artifacts,
			"artifact_count":       len(run.Artifacts),
			"artifact_refs":        refs,
			"artifact_stem":        primary.Stem,
			"artifact_topic":       primary.Topic,
			"artifact_type":        primary.ArtifactType,
			"artifact_updated_at":  primary.CapturedAt.Format(time.RFC3339),
			"report_source":        sourceChannel,
			"source_channel":       sourceChannel,
			"source_family":        sourceFamily,
			"source_run_id":        run.RunID,
			"antigravity_sync_key": run.SyncKey,
			"severity":             primary.Severity,
		}
		if primary.Version != nil {
			result["artifact_version"] = *primary.Version
		}
		return result
	}
}

func geminiSyncRef(syncKey string) string {
	return "gemini_sync:" + strings.TrimSpace(syncKey)
}

func selectedGeminiArtifact(items []runtime.GeminiArtifact) *runtime.GeminiArtifact {
	if len(items) == 0 {
		return nil
	}
	item := items[0]
	return &item
}

func geminiJobResult(item runtime.GeminiArtifact) map[string]any {
	sourceChannel := "gemini"
	sourceFamily := runtime.ObservationChannelFamily(sourceChannel)
	artifacts := append([]string{}, item.RelatedPaths...)
	sort.Strings(artifacts)
	relativeArtifacts := append([]string{}, item.RelatedRelativePaths...)
	sort.Strings(relativeArtifacts)
	refs := append([]string{}, item.Refs...)
	sort.Strings(refs)
	result := map[string]any{
		"summary":                 item.Summary,
		"artifacts":               artifacts,
		"artifact_count":          len(item.RelatedPaths),
		"artifact_refs":           refs,
		"artifact_stem":           item.Stem,
		"artifact_topic":          item.Topic,
		"artifact_type":           item.ArtifactType,
		"artifact_relative_path":  item.RelativePath,
		"artifact_relative_paths": relativeArtifacts,
		"artifact_updated_at":     item.CapturedAt.Format(time.RFC3339),
		"report_source":           sourceChannel,
		"source_channel":          sourceChannel,
		"source_family":           sourceFamily,
		"gemini_sync_key":         item.SyncKey,
		"severity":                item.Severity,
	}
	if item.Version != nil {
		result["artifact_version"] = *item.Version
	}
	return result
}

func cleanupGeminiArtifacts(root string, artifacts []runtime.GeminiArtifact) ([]string, []string) {
	seen := map[string]bool{}
	cleaned := []string{}
	cleanupErrors := []string{}
	for _, artifact := range artifacts {
		for _, path := range artifact.RelatedPaths {
			value := strings.TrimSpace(path)
			if value == "" || seen[value] {
				continue
			}
			seen[value] = true
			if err := os.Remove(value); err != nil {
				if os.IsNotExist(err) {
					continue
				}
				rel, relErr := filepath.Rel(root, value)
				if relErr != nil {
					rel = value
				}
				cleanupErrors = append(cleanupErrors, filepath.ToSlash(rel)+": "+err.Error())
				continue
			}
			rel, relErr := filepath.Rel(root, value)
			if relErr != nil {
				rel = value
			}
			cleaned = append(cleaned, filepath.ToSlash(rel))
		}
	}
	sort.Strings(cleaned)
	sort.Strings(cleanupErrors)
	return cleaned, cleanupErrors
}
