package main

import (
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

// Smoke test that walks a minimal session → job create → job report flow using existing stubs.
func TestSmokeSessionJobFlow(t *testing.T) {
	session := &sessionServiceStub{
		auth: runtime.AuthStatus{WriteAuthEnabled: false},
		result: runtime.SessionFinishResult{
			Memory: runtime.SystemMemory{
				CurrentFocus: "Stabilize queue",
				NextSteps:    []string{"cutover"},
				OpenRisks:    []string{"drift"},
				UpdatedAt:    time.Now().UTC(),
			},
			Event: runtime.EventEnvelope{EventID: "event_smoke", Kind: "session.finished", Actor: "system", Surface: runtime.SurfaceAPI, WorkItemID: "system", Stage: runtime.StageDiscover, Timestamp: time.Now().UTC(), Data: map[string]any{}},
		},
	}
	job := &jobServiceStub{
		auth: runtime.AuthStatus{WriteAuthEnabled: false},
		reportResult: runtime.AgentJobReportResult{
			Job:   runtime.AgentJob{JobID: "job_smoke", Status: "succeeded"},
			Event: runtime.EventEnvelope{EventID: "event_job_smoke", Kind: "agent_job.succeeded", Actor: "system", Surface: runtime.SurfaceAPI, WorkItemID: "system", Stage: runtime.StageDiscover, Timestamp: time.Now().UTC(), Data: map[string]any{}},
		},
	}

	finishOut := captureStdout(t, func() {
		runSession(session, []string{"finish", "--focus", "Stabilize queue", "--steps", "cutover", "--risks", "drift", "--note", "smoke-run"})
	})
	if !strings.Contains(finishOut, "session.finished") || session.finished.CurrentFocus != "Stabilize queue" {
		t.Fatalf("unexpected session finish output/state: %q %+v", finishOut, session.finished)
	}

	createOut := captureStdout(t, func() {
		runJob(job, []string{"create", "--id", "job_smoke", "--kind", "plan", "--role", "planner", "--summary", "Smoke pipeline"})
	})
	if job.job.JobID != "job_smoke" || job.job.Status != "queued" {
		t.Fatalf("unexpected job create: %+v", job.job)
	}
	if !strings.Contains(createOut, "job_smoke") {
		t.Fatalf("job create output missing id: %q", createOut)
	}

	reportOut := captureStdout(t, func() {
		runJob(job, []string{"report", "--id", "job_smoke", "--status", "succeeded", "--notes", "ok"})
	})
	if job.reportedJobID != "job_smoke" || job.reportedStatus != "succeeded" {
		t.Fatalf("unexpected job report call: id=%q status=%q", job.reportedJobID, job.reportedStatus)
	}
	if !strings.Contains(reportOut, "agent_job.succeeded") {
		t.Fatalf("job report output missing success event: %q", reportOut)
	}
}
