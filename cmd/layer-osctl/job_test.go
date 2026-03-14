package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

type jobServiceStub struct {
	auth            runtime.AuthStatus
	job             runtime.AgentJob
	jobs            []runtime.AgentJob
	status          string
	notes           []string
	dispatchResult  runtime.AgentDispatchResult
	reportResult    runtime.AgentJobReportResult
	dispatchedJobID string
	reportedJobID   string
	reportedStatus  string
	reportedNotes   []string
	reportedResult  map[string]any
	profiles        []runtime.AgentDispatchProfile
	packet          runtime.AgentRunPacket
	packetJobID     string
	promoteLimit    int
	promoteDispatch bool
	promoteResult   runtime.AgentJobPromotionResult
}

func (s *jobServiceStub) AuthStatus() runtime.AuthStatus    { return s.auth }
func (s *jobServiceStub) ListAgentJobs() []runtime.AgentJob { return s.jobs }
func (s *jobServiceStub) CreateAgentJob(item runtime.AgentJob) error {
	s.job = item
	s.jobs = append(s.jobs, item)
	return nil
}
func (s *jobServiceStub) PromoteContextJobs(limit int, dispatch bool) (runtime.AgentJobPromotionResult, error) {
	s.promoteLimit = limit
	s.promoteDispatch = dispatch
	return s.promoteResult, nil
}
func (s *jobServiceStub) UpdateAgentJob(jobID string, status string, notes []string, result map[string]any) (runtime.AgentJob, error) {
	s.status = status
	s.notes = append([]string{}, notes...)
	s.job = runtime.AgentJob{JobID: jobID, Kind: "plan", Role: "planner", Summary: "Plan founder lane", Status: status, Source: "manual", Surface: runtime.SurfaceAPI, Stage: runtime.StageDiscover, Notes: append([]string{}, notes...), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	return s.job, nil
}
func (s *jobServiceStub) ReportAgentJob(jobID string, status string, notes []string, result map[string]any) (runtime.AgentJobReportResult, error) {
	s.reportedJobID = jobID
	s.reportedStatus = status
	s.reportedNotes = append([]string{}, notes...)
	s.reportedResult = result
	if s.reportResult.Job.JobID == "" {
		s.reportResult.Job.JobID = jobID
		s.reportResult.Job.Status = status
	}
	return s.reportResult, nil
}
func (s *jobServiceStub) DispatchAgentJob(jobID string) (runtime.AgentDispatchResult, error) {
	s.dispatchedJobID = jobID
	if s.dispatchResult.Job.JobID == "" {
		s.dispatchResult.Job.JobID = jobID
	}
	return s.dispatchResult, nil
}
func (s *jobServiceStub) AgentDispatchProfiles() []runtime.AgentDispatchProfile {
	return s.profiles
}
func (s *jobServiceStub) AgentRunPacket(jobID string) (runtime.AgentRunPacket, error) {
	s.packetJobID = jobID
	if s.packet.Job.JobID == "" {
		s.packet.Job.JobID = jobID
	}
	return s.packet, nil
}
func (s *jobServiceStub) SessionFinish(input runtime.SessionFinishInput) (runtime.SessionFinishResult, error) {
	return runtime.SessionFinishResult{}, nil
}

func TestRunJobCreateWritesAgentJob(t *testing.T) {
	service := &jobServiceStub{}
	raw := captureStdout(t, func() {
		runJob(service, []string{"create", "--id", "job_001", "--branch", "branch_001", "--kind", "plan", "--role", "planner", "--summary", "Plan founder lane", "--source", "founder.manual", "--surface", "api", "--stage", "discover", "--ref", "proposal_001", "--notes", "one,two", "--allowed-paths", "cmd/layer-osctl/,internal/runtime/", "--payload-json", `{"priority":"high"}`, "--council", "claude,openai", "--council-primary", "claude"})
	})
	if service.job.JobID != "job_001" || service.job.Role != "planner" || service.job.Ref == nil || *service.job.Ref != "proposal_001" {
		t.Fatalf("unexpected created job: %+v", service.job)
	}
	if service.job.BranchID == nil || *service.job.BranchID != "branch_001" {
		t.Fatalf("unexpected job branch: %+v", service.job.BranchID)
	}
	if len(service.job.Notes) != 2 || service.job.Notes[0] != "one" {
		t.Fatalf("unexpected job notes: %+v", service.job.Notes)
	}
	if service.job.Payload["priority"] != "high" {
		t.Fatalf("expected payload-json merge, got %+v", service.job.Payload)
	}
	allowed, ok := service.job.Payload["allowed_paths"].([]string)
	if !ok || len(allowed) != 2 || allowed[0] != "cmd/layer-osctl/" {
		t.Fatalf("unexpected allowed_paths payload: %#v", service.job.Payload["allowed_paths"])
	}
	council, ok := service.job.Payload["council"].(map[string]any)
	if !ok || council["primary_provider"] != "claude" {
		t.Fatalf("unexpected council payload: %#v", service.job.Payload["council"])
	}
	if !strings.Contains(raw, "job_001") {
		t.Fatalf("expected job output, got %s", raw)
	}
}

func TestRunJobCreateAutoGeneratesIDWhenMissing(t *testing.T) {
	service := &jobServiceStub{}
	raw := captureStdout(t, func() {
		runJob(service, []string{"create", "--kind", "plan", "--role", "planner", "--summary", "Plan founder lane"})
	})
	if !strings.HasPrefix(service.job.JobID, "job_") {
		t.Fatalf("expected auto-generated job id, got %q", service.job.JobID)
	}
	if strings.TrimPrefix(service.job.JobID, "job_") == "" {
		t.Fatalf("expected unix ms suffix in job id, got %q", service.job.JobID)
	}
	if !strings.Contains(raw, service.job.JobID) {
		t.Fatalf("expected generated job id in output, got %s", raw)
	}
}

func TestParseJobCreatePayloadInput(t *testing.T) {
	payload, err := parseJobCreatePayloadInput(`{"role_hint":"backend"}`, "cmd/layer-osctl/,internal/runtime/", "claude,openai", "claude")
	if err != nil {
		t.Fatalf("parse payload: %v", err)
	}
	if payload["role_hint"] != "backend" {
		t.Fatalf("unexpected payload-json merge: %+v", payload)
	}
	allowed, ok := payload["allowed_paths"].([]string)
	if !ok || len(allowed) != 2 {
		t.Fatalf("unexpected allowed paths: %#v", payload["allowed_paths"])
	}
	council, ok := payload["council"].(map[string]any)
	if !ok || council["primary_provider"] != "claude" {
		t.Fatalf("unexpected council payload: %#v", payload["council"])
	}
}

func TestParseJobCreatePayloadInputRejectsCouncilPrimaryWithoutProviders(t *testing.T) {
	if _, err := parseJobCreatePayloadInput("", "", "", "claude"); err == nil {
		t.Fatal("expected council primary without providers to fail")
	}
}

func TestParseJobCreatePayloadInputRejectsUnknownCouncilProvider(t *testing.T) {
	if _, err := parseJobCreatePayloadInput("", "", "claude,foo", "claude"); err == nil {
		t.Fatal("expected unknown council provider to fail")
	}
}

func TestRunJobListFiltersStatusAndLimit(t *testing.T) {
	now := time.Now().UTC()
	service := &jobServiceStub{jobs: []runtime.AgentJob{
		{JobID: "job_001", Status: "queued", Role: "planner", Kind: "plan", Summary: "Queued", Source: "manual", Surface: runtime.SurfaceAPI, Stage: runtime.StageDiscover, CreatedAt: now, UpdatedAt: now},
		{JobID: "job_002", Status: "running", Role: "implementer", Kind: "implement", Summary: "Running", Source: "manual", Surface: runtime.SurfaceAPI, Stage: runtime.StageCompose, CreatedAt: now, UpdatedAt: now},
		{JobID: "job_003", Status: "succeeded", Role: "verifier", Kind: "verify", Summary: "Done", Source: "manual", Surface: runtime.SurfaceAPI, Stage: runtime.StageVerify, CreatedAt: now, UpdatedAt: now},
	}}

	raw := captureStdout(t, func() {
		runJob(service, []string{"list", "--status", "open", "--limit", "1"})
	})

	if !strings.Contains(raw, "job_001") {
		t.Fatalf("expected first open job in output, got %s", raw)
	}
	if strings.Contains(raw, "job_002") || strings.Contains(raw, "job_003") {
		t.Fatalf("expected output to be filtered and limited, got %s", raw)
	}
}

func TestRunJobUpdateWritesStatus(t *testing.T) {
	service := &jobServiceStub{}
	raw := captureStdout(t, func() {
		runJob(service, []string{"update", "--id", "job_001", "--status", "failed", "--notes", "exit=1,retry"})
	})
	if service.status != "failed" || len(service.notes) != 2 || service.notes[0] != "exit=1" {
		t.Fatalf("unexpected update call: status=%q notes=%v", service.status, service.notes)
	}
	if !strings.Contains(raw, "failed") {
		t.Fatalf("expected updated job output, got %s", raw)
	}
}

func TestRunJobDispatchCallsService(t *testing.T) {
	service := &jobServiceStub{dispatchResult: runtime.AgentDispatchResult{Job: runtime.AgentJob{JobID: "job_001", Status: "running"}}}
	raw := captureStdout(t, func() {
		runJob(service, []string{"dispatch", "--id", "job_001"})
	})
	if service.dispatchedJobID != "job_001" {
		t.Fatalf("expected dispatched job id job_001, got %q", service.dispatchedJobID)
	}
	if !strings.Contains(raw, "job_001") {
		t.Fatalf("unexpected dispatch output: %s", raw)
	}
}

func TestRunJobProfilesWritesProfiles(t *testing.T) {
	service := &jobServiceStub{profiles: []runtime.AgentDispatchProfile{{Role: "planner", Provider: "openai", DispatchReady: false, Notes: []string{"provider dispatch not enabled"}}}}
	raw := captureStdout(t, func() {
		runJob(service, []string{"profiles"})
	})
	if !strings.Contains(raw, "planner") || !strings.Contains(raw, "openai") {
		t.Fatalf("unexpected profiles output: %s", raw)
	}
}

func TestRunJobReportCallsService(t *testing.T) {
	service := &jobServiceStub{reportResult: runtime.AgentJobReportResult{Job: runtime.AgentJob{JobID: "job_001", Status: "succeeded"}, Event: runtime.EventEnvelope{EventID: "event_001", Kind: "agent_job.succeeded"}}}
	raw := captureStdout(t, func() {
		runJob(service, []string{"report", "--id", "job_001", "--status", "succeeded", "--notes", "done,clean", "--result", "provider=openai,gateway_call_id=gateway_001"})
	})
	if service.reportedJobID != "job_001" || service.reportedStatus != "succeeded" {
		t.Fatalf("unexpected report call: id=%q status=%q", service.reportedJobID, service.reportedStatus)
	}
	if len(service.reportedNotes) != 2 || service.reportedNotes[0] != "done" {
		t.Fatalf("unexpected report notes: %+v", service.reportedNotes)
	}
	if service.reportedResult["provider"] != "openai" || service.reportedResult["gateway_call_id"] != "gateway_001" {
		t.Fatalf("unexpected report result: %+v", service.reportedResult)
	}
	if !strings.Contains(raw, "agent_job.succeeded") {
		t.Fatalf("unexpected report output: %s", raw)
	}
}

func TestRunJobReportReadsResultFile(t *testing.T) {
	reportPath := t.TempDir() + "/job-report.json"
	if err := os.WriteFile(reportPath, []byte(`{
  "summary": "Kept the live daemon handoff path canonical.",
  "artifacts": ["docs/operator.md"],
  "verification": {"status": "passed", "meaning": "doc and wrapper paths align"},
  "open_risks": [],
  "follow_on": ["Monitor Antigravity channel pressure during the next live session."],
  "touched_paths": ["docs/operator.md"],
  "blocked_paths": []
}`), 0o644); err != nil {
		t.Fatalf("write result file: %v", err)
	}

	service := &jobServiceStub{reportResult: runtime.AgentJobReportResult{Job: runtime.AgentJob{JobID: "job_002", Status: "succeeded"}, Event: runtime.EventEnvelope{EventID: "event_002", Kind: "agent_job.succeeded"}}}
	raw := captureStdout(t, func() {
		runJob(service, []string{"report", "--id", "job_002", "--status", "succeeded", "--result-file", reportPath})
	})

	if service.reportedResult["summary"] != "Kept the live daemon handoff path canonical." {
		t.Fatalf("unexpected summary: %+v", service.reportedResult["summary"])
	}
	artifacts, ok := service.reportedResult["artifacts"].([]any)
	if !ok || len(artifacts) != 1 || artifacts[0] != "docs/operator.md" {
		t.Fatalf("unexpected artifacts: %+v", service.reportedResult["artifacts"])
	}
	verification, ok := service.reportedResult["verification"].(map[string]any)
	if !ok || verification["status"] != "passed" {
		t.Fatalf("unexpected verification: %+v", service.reportedResult["verification"])
	}
	if !strings.Contains(raw, "event_002") {
		t.Fatalf("unexpected report output: %s", raw)
	}
}

func TestParseJobResultInputRejectsJSONAndFileTogether(t *testing.T) {
	if _, err := parseJobResultInput("", `{"summary":"ok"}`, "report.json"); err == nil {
		t.Fatal("expected mixed result source error")
	}
}

func TestRunJobPacketCallsService(t *testing.T) {
	service := &jobServiceStub{packet: runtime.AgentRunPacket{
		Job:       runtime.AgentJob{JobID: "job_001"},
		Runtime:   runtime.AgentRuntimeContract{ReportCommand: "layer-osctl job report --id job_001 --status <succeeded|failed|canceled> [--notes a,b] [--result-file <path-to-json>]"},
		Knowledge: runtime.KnowledgePacket{CurrentFocus: "Focus"},
		Handoff:   &runtime.HandoffPacket{GeneratedAt: time.Now().UTC()},
	}}
	raw := captureStdout(t, func() {
		runJob(service, []string{"packet", "--id", "job_001"})
	})
	if service.packetJobID != "job_001" {
		t.Fatalf("expected packet job id job_001, got %q", service.packetJobID)
	}
	if !strings.Contains(raw, "job_001") {
		t.Fatalf("unexpected packet output: %s", raw)
	}
	if !strings.Contains(raw, "report_command") {
		t.Fatalf("expected runtime contract in packet output: %s", raw)
	}
	if !strings.Contains(raw, "--result-file") {
		t.Fatalf("expected result-file report command, got %s", raw)
	}
}

func TestRunJobPromoteCallsService(t *testing.T) {
	service := &jobServiceStub{promoteResult: runtime.AgentJobPromotionResult{Created: 1, Items: []runtime.AgentJobPromotionItem{{SourceKind: "parallel_candidate", Status: "created", Summary: "Plan queue drift lane"}}}}
	raw := captureStdout(t, func() {
		runJob(service, []string{"promote", "--limit", "2", "--dispatch"})
	})
	if service.promoteLimit != 2 || !service.promoteDispatch {
		t.Fatalf("unexpected promote call: limit=%d dispatch=%v", service.promoteLimit, service.promoteDispatch)
	}
	if !strings.Contains(raw, "parallel_candidate") || !strings.Contains(raw, "created") {
		t.Fatalf("unexpected promote output: %s", raw)
	}
}

func TestRunJobWorkProcessesPacketReadyJob(t *testing.T) {
	repoRoot := t.TempDir()
	workRoot := filepath.Join(repoRoot, "worker")
	now := time.Now().UTC()
	job := runtime.AgentJob{
		JobID:     "job_work_001",
		Kind:      "implement",
		Role:      "implementer",
		Summary:   "Patch the backend lane",
		Status:    "running",
		Source:    "founder.manual",
		Surface:   runtime.SurfaceAPI,
		Stage:     runtime.StageCompose,
		Result:    map[string]any{"dispatch_state": "packet_ready"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	command := `cat >"$LAYER_OS_RESULT_PATH" <<'JSON'
{"summary":"Patched the backend lane.","artifacts":["cmd/layer-osctl/job_work.go"],"verification":["go test ./... not run in worker test"],"open_risks":[],"follow_on":[],"touched_paths":["cmd/layer-osctl/job_work.go"],"blocked_paths":[]}
JSON`
	service := &jobServiceStub{
		jobs: []runtime.AgentJob{job},
		packet: runtime.AgentRunPacket{
			Job:       job,
			Runtime:   runtime.AgentRuntimeContract{ReportCommand: "layer-osctl job report --id job_work_001 --status <succeeded|failed|canceled> [--notes a,b] [--result-file <path-to-json>]"},
			Knowledge: runtime.KnowledgePacket{CurrentFocus: "Patch backend lane"},
		},
		reportResult: runtime.AgentJobReportResult{
			Job:   runtime.AgentJob{JobID: "job_work_001", Status: "succeeded"},
			Event: runtime.EventEnvelope{EventID: "event_work_001", Kind: "agent_job.succeeded"},
		},
	}

	raw := captureStdout(t, func() {
		runJob(service, []string{"work", "--once", "--repo-root", repoRoot, "--work-root", workRoot, "--command", command})
	})

	if service.packetJobID != "job_work_001" {
		t.Fatalf("expected packet fetch for job_work_001, got %q", service.packetJobID)
	}
	if service.reportedJobID != "job_work_001" || service.reportedStatus != "succeeded" {
		t.Fatalf("unexpected report call: id=%q status=%q", service.reportedJobID, service.reportedStatus)
	}
	if service.reportedResult["summary"] != "Patched the backend lane." {
		t.Fatalf("unexpected worker summary: %+v", service.reportedResult)
	}
	if !strings.Contains(raw, "processed_job_ids") || !strings.Contains(raw, "job_work_001") {
		t.Fatalf("unexpected worker output: %s", raw)
	}
}

func TestRunJobWorkDispatchesQueuedJobBeforeProcessing(t *testing.T) {
	repoRoot := t.TempDir()
	workRoot := filepath.Join(repoRoot, "worker")
	now := time.Now().UTC()
	queued := runtime.AgentJob{
		JobID:     "job_work_queued_001",
		Kind:      "implement",
		Role:      "implementer",
		Summary:   "Patch queued backend lane",
		Status:    "queued",
		Source:    "founder.manual",
		Surface:   runtime.SurfaceAPI,
		Stage:     runtime.StageCompose,
		CreatedAt: now,
		UpdatedAt: now,
	}
	running := queued
	running.Status = "running"
	running.Result = map[string]any{"dispatch_state": "packet_ready"}
	command := `cat >"$LAYER_OS_RESULT_PATH" <<'JSON'
{"summary":"Queued worker completed.","artifacts":["cmd/layer-osctl/job.go"],"verification":[],"open_risks":[],"follow_on":[],"touched_paths":["cmd/layer-osctl/job.go"],"blocked_paths":[]}
JSON`
	service := &jobServiceStub{
		jobs:           []runtime.AgentJob{queued},
		dispatchResult: runtime.AgentDispatchResult{Job: running},
		packet: runtime.AgentRunPacket{
			Job:       running,
			Runtime:   runtime.AgentRuntimeContract{ReportCommand: "layer-osctl job report --id job_work_queued_001 --status <succeeded|failed|canceled> [--notes a,b] [--result-file <path-to-json>]"},
			Knowledge: runtime.KnowledgePacket{CurrentFocus: "Dispatch then patch"},
		},
		reportResult: runtime.AgentJobReportResult{
			Job:   runtime.AgentJob{JobID: "job_work_queued_001", Status: "succeeded"},
			Event: runtime.EventEnvelope{EventID: "event_work_queued_001", Kind: "agent_job.succeeded"},
		},
	}

	raw := captureStdout(t, func() {
		runJob(service, []string{"work", "--once", "--repo-root", repoRoot, "--work-root", workRoot, "--command", command})
	})

	if service.dispatchedJobID != "job_work_queued_001" {
		t.Fatalf("expected queued dispatch before work, got %q", service.dispatchedJobID)
	}
	if service.reportedJobID != "job_work_queued_001" || service.reportedStatus != "succeeded" {
		t.Fatalf("unexpected queued worker report: id=%q status=%q", service.reportedJobID, service.reportedStatus)
	}
	if !strings.Contains(raw, "\"dispatched\": 1") {
		t.Fatalf("expected worker summary to record dispatch, got %s", raw)
	}
}

func TestRunJobWorkPacketFileModeRunsOffline(t *testing.T) {
	repoRoot := t.TempDir()
	workRoot := filepath.Join(repoRoot, "worker")
	packetPath := filepath.Join(repoRoot, "packet.json")
	packet := runtime.AgentRunPacket{
		Job: runtime.AgentJob{
			JobID:   "job_packet_001",
			Kind:    "implement",
			Role:    "implementer",
			Summary: "Rehearse offline worker packet",
			Status:  "running",
			Source:  "founder.manual",
			Stage:   runtime.StageCompose,
			Payload: map[string]any{"allowed_paths": []string{"cmd/layer-osctl/"}},
		},
		Knowledge: runtime.KnowledgePacket{CurrentFocus: "Offline packet rehearsal"},
	}
	rawPacket, err := json.Marshal(packet)
	if err != nil {
		t.Fatalf("marshal packet: %v", err)
	}
	if err := os.WriteFile(packetPath, rawPacket, 0o644); err != nil {
		t.Fatalf("write packet file: %v", err)
	}

	command := `cat >"$LAYER_OS_RESULT_PATH" <<'JSON'
{"summary":"Offline packet worker completed.","artifacts":["cmd/layer-osctl/job_work.go"],"verification":[],"open_risks":[],"follow_on":[],"touched_paths":["cmd/layer-osctl/job_work.go"],"blocked_paths":[]}
JSON`
	service := &jobServiceStub{}

	raw := captureStdout(t, func() {
		runJob(service, []string{"work", "--once", "--dispatch-queued=false", "--packet-file", packetPath, "--repo-root", repoRoot, "--work-root", workRoot, "--command", command})
	})

	if service.packetJobID != "" || service.reportedJobID != "" || service.dispatchedJobID != "" {
		t.Fatalf("expected offline packet mode to avoid daemon mutations: packet=%q report=%q dispatch=%q", service.packetJobID, service.reportedJobID, service.dispatchedJobID)
	}
	if !strings.Contains(raw, "packet_file_mode") || !strings.Contains(raw, "job_packet_001") {
		t.Fatalf("unexpected offline worker output: %s", raw)
	}
}
