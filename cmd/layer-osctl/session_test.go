package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

type sessionServiceStub struct {
	auth           runtime.AuthStatus
	bootstrap      runtime.SessionBootstrapPacket
	bootstrapFn    func(full bool) (runtime.SessionBootstrapPacket, error)
	bootstrapCalls int
	bootstrapFulls []bool
	checkpointed   runtime.SessionCheckpointInput
	checkpoint     runtime.SessionCheckpoint
	noted          runtime.SessionNoteInput
	noteResult     runtime.SessionNoteResult
	finished       runtime.SessionFinishInput
	result         runtime.SessionFinishResult
	err            error
}

func (s *sessionServiceStub) AuthStatus() runtime.AuthStatus { return s.auth }
func (s *sessionServiceStub) SessionBootstrap(full bool) (runtime.SessionBootstrapPacket, error) {
	s.bootstrapCalls++
	s.bootstrapFulls = append(s.bootstrapFulls, full)
	if s.bootstrapFn != nil {
		return s.bootstrapFn(full)
	}
	return s.bootstrap, s.err
}
func (s *sessionServiceStub) CheckpointSession(input runtime.SessionCheckpointInput) (runtime.SessionCheckpoint, error) {
	s.checkpointed = input
	if s.checkpoint.UpdatedAt.IsZero() {
		s.checkpoint.UpdatedAt = time.Now().UTC()
	}
	return s.checkpoint, s.err
}
func (s *sessionServiceStub) SessionNote(input runtime.SessionNoteInput) (runtime.SessionNoteResult, error) {
	s.noted = input
	now := time.Now().UTC()
	if s.noteResult.Record.CreatedAt.IsZero() {
		s.noteResult.Record.CreatedAt = now
	}
	if s.noteResult.Record.UpdatedAt.IsZero() {
		s.noteResult.Record.UpdatedAt = now
	}
	if s.noteResult.Note.CreatedAt.IsZero() {
		s.noteResult.Note.CreatedAt = now
	}
	if s.noteResult.Observation.CapturedAt.IsZero() {
		s.noteResult.Observation.CapturedAt = now
	}
	return s.noteResult, s.err
}
func (s *sessionServiceStub) SessionFinish(input runtime.SessionFinishInput) (runtime.SessionFinishResult, error) {
	s.finished = input
	if s.result.Memory.UpdatedAt.IsZero() {
		s.result.Memory.UpdatedAt = time.Now().UTC()
	}
	if s.result.Event.Timestamp.IsZero() {
		s.result.Event.Timestamp = time.Now().UTC()
	}
	return s.result, s.err
}

func TestRunSessionCheckpointParsesFields(t *testing.T) {
	service := &sessionServiceStub{
		checkpoint: runtime.SessionCheckpoint{
			CheckpointID: "session_checkpoint:001",
			Source:       "terminal",
			Actor:        "system",
			CurrentFocus: "Lock queue",
			NextSteps:    []string{"one"},
			OpenRisks:    []string{"drift"},
			Refs:         []string{"thread:terminal-1"},
		},
	}
	raw := captureStdout(t, func() {
		runSession(service, []string{"checkpoint", "--source", "terminal", "--focus", "Lock queue", "--goal", "Bus hardening", "--steps", "one,two", "--risks", "drift", "--handoff", "continue", "--note", "working", "--refs", "thread:terminal-1,proposal:queue"})
	})
	if service.checkpointed.CurrentFocus != "Lock queue" {
		t.Fatalf("unexpected focus: %+v", service.checkpointed)
	}
	if service.checkpointed.CurrentGoal == nil || *service.checkpointed.CurrentGoal != "Bus hardening" {
		t.Fatalf("unexpected goal: %+v", service.checkpointed.CurrentGoal)
	}
	if len(service.checkpointed.NextSteps) != 2 || service.checkpointed.NextSteps[0] != "one" {
		t.Fatalf("unexpected steps: %+v", service.checkpointed.NextSteps)
	}
	if len(service.checkpointed.Refs) != 2 || service.checkpointed.Refs[1] != "proposal:queue" {
		t.Fatalf("unexpected refs: %+v", service.checkpointed.Refs)
	}
	if !strings.Contains(raw, "session_checkpoint:001") {
		t.Fatalf("expected checkpoint output, got %s", raw)
	}
}

func TestRunSessionFinishParsesFields(t *testing.T) {
	service := &sessionServiceStub{result: runtime.SessionFinishResult{Memory: runtime.SystemMemory{CurrentFocus: "Lock queue", NextSteps: []string{}, OpenRisks: []string{}}, Event: runtime.EventEnvelope{EventID: "event_001", Kind: "session.finished", Actor: "system", Surface: runtime.SurfaceAPI, WorkItemID: "system", Stage: runtime.StageDiscover, Data: map[string]any{}}}}
	raw := captureStdout(t, func() {
		runSession(service, []string{"finish", "--focus", "Lock queue", "--goal", "Bus hardening", "--steps", "one,two", "--risks", "drift", "--handoff", "continue", "--note", "done"})
	})
	if service.finished.CurrentFocus != "Lock queue" {
		t.Fatalf("unexpected focus: %+v", service.finished)
	}
	if service.finished.CurrentGoal == nil || *service.finished.CurrentGoal != "Bus hardening" {
		t.Fatalf("unexpected goal: %+v", service.finished.CurrentGoal)
	}
	if len(service.finished.NextSteps) != 2 || service.finished.NextSteps[0] != "one" {
		t.Fatalf("unexpected steps: %+v", service.finished.NextSteps)
	}
	if !strings.Contains(raw, "session.finished") {
		t.Fatalf("expected finish output, got %s", raw)
	}
}

type sessionWriteRequestStub struct {
	sessionServiceStub
	method string
	path   string
	body   any
}

func (s *sessionWriteRequestStub) request(method string, path string, payload any, out any) error {
	s.method = method
	s.path = path
	s.body = payload
	if target, ok := out.(*runtime.SessionCheckpoint); ok {
		*target = s.checkpoint
	}
	if target, ok := out.(*runtime.SessionFinishResult); ok {
		*target = s.result
	}
	if target, ok := out.(*runtime.SessionNoteResult); ok {
		*target = s.noteResult
	}
	return nil
}

func decodeSessionCheckpointPayload(payload any) (runtime.SessionCheckpointInput, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return runtime.SessionCheckpointInput{}, err
	}
	var input runtime.SessionCheckpointInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return runtime.SessionCheckpointInput{}, err
	}
	return input, nil
}

func decodeSessionFinishPayload(payload any) (runtime.SessionFinishInput, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return runtime.SessionFinishInput{}, err
	}
	var input runtime.SessionFinishInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return runtime.SessionFinishInput{}, err
	}
	return input, nil
}

func decodeSessionNotePayload(payload any) (runtime.SessionNoteInput, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return runtime.SessionNoteInput{}, err
	}
	var input runtime.SessionNoteInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return runtime.SessionNoteInput{}, err
	}
	return input, nil
}

func TestRunSessionCheckpointPostsToDaemonHook(t *testing.T) {
	goal := "Bus hardening"
	handoff := "continue"
	note := "working"
	service := &sessionWriteRequestStub{
		sessionServiceStub: sessionServiceStub{
			checkpoint: runtime.SessionCheckpoint{
				CheckpointID: "session_checkpoint:daemon",
				Source:       "terminal",
				Actor:        "system",
				CurrentFocus: "Lock queue",
				NextSteps:    []string{"one", "two"},
				OpenRisks:    []string{"drift"},
				Refs:         []string{"thread:terminal-1"},
			},
		},
	}
	service.sessionServiceStub.err = errors.New("session checkpoint fallback should not run")

	raw := captureStdout(t, func() {
		runSession(service, []string{"checkpoint", "--source", "terminal", "--focus", "Lock queue", "--goal", goal, "--steps", "one,two", "--risks", "drift", "--handoff", handoff, "--note", note, "--refs", "thread:terminal-1"})
	})

	if service.method != http.MethodPost {
		t.Fatalf("expected POST, got %q", service.method)
	}
	if service.path != "/api/layer-os/session/checkpoint" {
		t.Fatalf("unexpected checkpoint path: %q", service.path)
	}
	body, err := decodeSessionCheckpointPayload(service.body)
	if err != nil {
		t.Fatalf("decode checkpoint body: %v", err)
	}
	if body.CurrentFocus != "Lock queue" {
		t.Fatalf("unexpected checkpoint body: %+v", body)
	}
	if body.CurrentGoal == nil || *body.CurrentGoal != goal {
		t.Fatalf("unexpected goal: %+v", body.CurrentGoal)
	}
	if body.HandoffNote == nil || *body.HandoffNote != handoff {
		t.Fatalf("unexpected handoff: %+v", body.HandoffNote)
	}
	if body.Note == nil || *body.Note != note {
		t.Fatalf("unexpected note: %+v", body.Note)
	}
	if len(body.Refs) != 1 || body.Refs[0] != "thread:terminal-1" {
		t.Fatalf("unexpected refs: %+v", body.Refs)
	}
	if !strings.Contains(raw, "session_checkpoint:daemon") {
		t.Fatalf("expected checkpoint output, got %s", raw)
	}
	if service.sessionServiceStub.checkpointed.CurrentFocus != "" {
		t.Fatalf("expected daemon hook instead of service fallback, got %+v", service.sessionServiceStub.checkpointed)
	}
}

func TestRunSessionNoteParsesFields(t *testing.T) {
	service := &sessionServiceStub{
		noteResult: runtime.SessionNoteResult{
			Record: runtime.ContinuityRecord{
				RecordID:     "continuity_record_001",
				Source:       "terminal",
				Actor:        "system",
				Status:       "active",
				CurrentFocus: "Lock queue",
				NextSteps:    []string{"one"},
				OpenRisks:    []string{"drift"},
				Refs:         []string{"thread:terminal-1"},
				Notes:        []runtime.ContinuityNote{},
			},
			Note: runtime.ContinuityNote{
				NoteID: "continuity_note_001",
				Kind:   "todo",
				Text:   "check queue",
				Refs:   []string{"thread:terminal-1"},
			},
			Observation: runtime.ObservationRecord{
				ObservationID:     "observation_001",
				SourceChannel:     "terminal",
				Actor:             "system",
				Topic:             "continuity.note",
				Refs:              []string{"thread:terminal-1"},
				Confidence:        "high",
				RawExcerpt:        "kind=todo | text=check queue",
				NormalizedSummary: "Continuity todo: check queue",
			},
		},
	}
	raw := captureStdout(t, func() {
		runSession(service, []string{"note", "--source", "terminal", "--kind", "todo", "--text", "check queue", "--refs", "thread:terminal-1,proposal:queue"})
	})
	if service.noted.Source != "terminal" || service.noted.Kind != "todo" || service.noted.Text != "check queue" {
		t.Fatalf("unexpected note input: %+v", service.noted)
	}
	if len(service.noted.Refs) != 2 || service.noted.Refs[1] != "proposal:queue" {
		t.Fatalf("unexpected note refs: %+v", service.noted.Refs)
	}
	if !strings.Contains(raw, "continuity_note_001") {
		t.Fatalf("expected note output, got %s", raw)
	}
}

func TestRunSessionNotePostsToDaemonHook(t *testing.T) {
	service := &sessionWriteRequestStub{
		sessionServiceStub: sessionServiceStub{
			noteResult: runtime.SessionNoteResult{
				Record: runtime.ContinuityRecord{
					RecordID:     "continuity_record_daemon",
					Source:       "terminal",
					Actor:        "system",
					Status:       "active",
					CurrentFocus: "Lock queue",
					NextSteps:    []string{"one"},
					OpenRisks:    []string{"drift"},
					Refs:         []string{"thread:terminal-1"},
					Notes:        []runtime.ContinuityNote{},
				},
				Note: runtime.ContinuityNote{
					NoteID: "continuity_note_daemon",
					Kind:   "risk",
					Text:   "queue drift",
					Refs:   []string{"thread:terminal-1"},
				},
				Observation: runtime.ObservationRecord{
					ObservationID:     "observation_daemon",
					SourceChannel:     "terminal",
					Actor:             "system",
					Topic:             "continuity.note",
					Refs:              []string{"thread:terminal-1"},
					Confidence:        "high",
					RawExcerpt:        "kind=risk | text=queue drift",
					NormalizedSummary: "Continuity risk: queue drift",
				},
			},
		},
	}
	service.sessionServiceStub.err = errors.New("session note fallback should not run")

	raw := captureStdout(t, func() {
		runSession(service, []string{"note", "--source", "terminal", "--kind", "risk", "--text", "queue drift", "--refs", "thread:terminal-1"})
	})

	if service.method != http.MethodPost {
		t.Fatalf("expected POST, got %q", service.method)
	}
	if service.path != "/api/layer-os/session/note" {
		t.Fatalf("unexpected note path: %q", service.path)
	}
	body, err := decodeSessionNotePayload(service.body)
	if err != nil {
		t.Fatalf("decode note body: %v", err)
	}
	if body.Source != "terminal" || body.Kind != "risk" || body.Text != "queue drift" {
		t.Fatalf("unexpected note body: %+v", body)
	}
	if len(body.Refs) != 1 || body.Refs[0] != "thread:terminal-1" {
		t.Fatalf("unexpected note refs: %+v", body.Refs)
	}
	if !strings.Contains(raw, "continuity_note_daemon") {
		t.Fatalf("expected note output, got %s", raw)
	}
	if service.sessionServiceStub.noted.Text != "" {
		t.Fatalf("expected daemon hook instead of service fallback, got %+v", service.sessionServiceStub.noted)
	}
}

func TestRunSessionFinishPostsToDaemonHook(t *testing.T) {
	goal := "Bus hardening"
	handoff := "continue"
	note := "done"
	service := &sessionWriteRequestStub{
		sessionServiceStub: sessionServiceStub{
			result: runtime.SessionFinishResult{
				Memory: runtime.SystemMemory{CurrentFocus: "Lock queue", NextSteps: []string{"one", "two"}, OpenRisks: []string{"drift"}},
				Event:  runtime.EventEnvelope{EventID: "event_001", Kind: "session.finished", Actor: "system", Surface: runtime.SurfaceAPI, WorkItemID: "system", Stage: runtime.StageDiscover, Data: map[string]any{}},
			},
		},
	}
	service.sessionServiceStub.err = errors.New("session finish fallback should not run")

	raw := captureStdout(t, func() {
		runSession(service, []string{"finish", "--focus", "Lock queue", "--goal", goal, "--steps", "one,two", "--risks", "drift", "--handoff", handoff, "--note", note})
	})

	if service.method != http.MethodPost {
		t.Fatalf("expected POST, got %q", service.method)
	}
	if service.path != "/api/layer-os/session/finish" {
		t.Fatalf("unexpected finish path: %q", service.path)
	}
	body, err := decodeSessionFinishPayload(service.body)
	if err != nil {
		t.Fatalf("decode finish body: %v", err)
	}
	if body.CurrentFocus != "Lock queue" {
		t.Fatalf("unexpected finish body: %+v", body)
	}
	if body.CurrentGoal == nil || *body.CurrentGoal != goal {
		t.Fatalf("unexpected goal: %+v", body.CurrentGoal)
	}
	if body.HandoffNote == nil || *body.HandoffNote != handoff {
		t.Fatalf("unexpected handoff: %+v", body.HandoffNote)
	}
	if body.Note == nil || *body.Note != note {
		t.Fatalf("unexpected note: %+v", body.Note)
	}
	if !strings.Contains(raw, "session.finished") {
		t.Fatalf("expected finish output, got %s", raw)
	}
	if service.sessionServiceStub.finished.CurrentFocus != "" {
		t.Fatalf("expected daemon hook instead of service fallback, got %+v", service.sessionServiceStub.finished)
	}
}

func TestSessionBootstrapPacketRetriesDaemonUnavailable(t *testing.T) {
	oldAttempts := sessionBootstrapRetryAttempts
	oldSleep := sessionBootstrapSleep
	oldFallback := sessionBootstrapFallback
	t.Cleanup(func() {
		sessionBootstrapRetryAttempts = oldAttempts
		sessionBootstrapSleep = oldSleep
		sessionBootstrapFallback = oldFallback
	})

	sessionBootstrapRetryAttempts = 3
	sleepCalls := 0
	sessionBootstrapSleep = func(time.Duration) { sleepCalls++ }
	sessionBootstrapFallback = func(full bool) (runtime.SessionBootstrapPacket, error) {
		t.Fatal("unexpected local fallback")
		return runtime.SessionBootstrapPacket{}, nil
	}

	service := &sessionServiceStub{}
	service.bootstrapFn = func(full bool) (runtime.SessionBootstrapPacket, error) {
		if !full {
			t.Fatalf("expected full bootstrap retry")
		}
		if service.bootstrapCalls < 3 {
			return runtime.SessionBootstrapPacket{}, wrapDaemonError(&url.Error{Op: "Get", URL: daemonBaseURL(), Err: errors.New("dial tcp: i/o timeout")})
		}
		return runtime.SessionBootstrapPacket{Source: "daemon", Knowledge: runtime.KnowledgePacket{CurrentFocus: "Resume work"}}, nil
	}

	packet, err := sessionBootstrapPacket(service, true, false)
	if err != nil {
		t.Fatalf("session bootstrap packet: %v", err)
	}
	if packet.Source != "daemon" || packet.Knowledge.CurrentFocus != "Resume work" {
		t.Fatalf("unexpected packet: %+v", packet)
	}
	if service.bootstrapCalls != 3 {
		t.Fatalf("expected 3 bootstrap attempts, got %d", service.bootstrapCalls)
	}
	if sleepCalls != 2 {
		t.Fatalf("expected 2 retry sleeps, got %d", sleepCalls)
	}
}

func TestSessionBootstrapPacketFallsBackOnlyWhenDaemonUnavailable(t *testing.T) {
	oldAttempts := sessionBootstrapRetryAttempts
	oldSleep := sessionBootstrapSleep
	oldFallback := sessionBootstrapFallback
	t.Cleanup(func() {
		sessionBootstrapRetryAttempts = oldAttempts
		sessionBootstrapSleep = oldSleep
		sessionBootstrapFallback = oldFallback
	})

	sessionBootstrapRetryAttempts = 2
	sessionBootstrapSleep = func(time.Duration) {}
	fallbackCalls := 0
	sessionBootstrapFallback = func(full bool) (runtime.SessionBootstrapPacket, error) {
		fallbackCalls++
		if full {
			t.Fatalf("expected thin fallback packet")
		}
		return runtime.SessionBootstrapPacket{Source: "local_fallback", Degraded: true, Knowledge: runtime.KnowledgePacket{CurrentFocus: "Read only"}}, nil
	}

	service := &sessionServiceStub{}
	service.bootstrapFn = func(full bool) (runtime.SessionBootstrapPacket, error) {
		return runtime.SessionBootstrapPacket{}, wrapDaemonError(&url.Error{Op: "Get", URL: daemonBaseURL(), Err: errors.New("network changed")})
	}

	packet, err := sessionBootstrapPacket(service, false, true)
	if err != nil {
		t.Fatalf("session bootstrap fallback: %v", err)
	}
	if fallbackCalls != 1 {
		t.Fatalf("expected one fallback, got %d", fallbackCalls)
	}
	if service.bootstrapCalls != 2 {
		t.Fatalf("expected retry attempts before fallback, got %d", service.bootstrapCalls)
	}
	if packet.Source != "local_fallback" || !packet.Degraded {
		t.Fatalf("unexpected fallback packet: %+v", packet)
	}
}

func TestSessionBootstrapPacketDoesNotFallbackOnGenericError(t *testing.T) {
	oldAttempts := sessionBootstrapRetryAttempts
	oldSleep := sessionBootstrapSleep
	oldFallback := sessionBootstrapFallback
	t.Cleanup(func() {
		sessionBootstrapRetryAttempts = oldAttempts
		sessionBootstrapSleep = oldSleep
		sessionBootstrapFallback = oldFallback
	})

	sessionBootstrapRetryAttempts = 3
	sessionBootstrapSleep = func(time.Duration) {}
	fallbackCalls := 0
	sessionBootstrapFallback = func(full bool) (runtime.SessionBootstrapPacket, error) {
		fallbackCalls++
		return runtime.SessionBootstrapPacket{}, nil
	}

	service := &sessionServiceStub{}
	genericErr := errors.New("decode daemon response: invalid character '}'")
	service.bootstrapFn = func(full bool) (runtime.SessionBootstrapPacket, error) {
		return runtime.SessionBootstrapPacket{}, genericErr
	}

	_, err := sessionBootstrapPacket(service, false, true)
	if !errors.Is(err, genericErr) {
		t.Fatalf("expected generic error to pass through, got %v", err)
	}
	if service.bootstrapCalls != 1 {
		t.Fatalf("expected no retry for generic error, got %d calls", service.bootstrapCalls)
	}
	if fallbackCalls != 0 {
		t.Fatalf("expected no fallback for generic error, got %d", fallbackCalls)
	}
}

func TestWrapDaemonErrorMarksDaemonUnavailable(t *testing.T) {
	err := wrapDaemonError(&url.Error{Op: "Get", URL: daemonBaseURL(), Err: errors.New("connection reset by peer")})
	if !errors.Is(err, errDaemonUnavailable) {
		t.Fatalf("expected daemon unavailable classification, got %v", err)
	}
	if !strings.Contains(err.Error(), "cannot reach layer-osd") {
		t.Fatalf("unexpected wrapped error: %v", err)
	}
}

func TestLocalSessionBootstrapReadsLocalProjection(t *testing.T) {
	dataDir := t.TempDir()
	service, err := runtime.NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.ReplaceMemory(runtime.SystemMemory{CurrentFocus: "Recover queue", NextSteps: []string{"resume"}, OpenRisks: []string{"daemon_down"}, UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}
	oldRoot := t.TempDir()
	if err := os.Setenv("LAYER_OS_REPO_ROOT", oldRoot); err != nil {
		t.Fatalf("set repo root: %v", err)
	}
	defer os.Unsetenv("LAYER_OS_REPO_ROOT")
	if err := os.MkdirAll(localRuntimeDataDir(), 0o755); err != nil {
		t.Fatalf("mkdir local runtime dir: %v", err)
	}
	copyDir(t, dataDir, localRuntimeDataDir())
	packet, err := localSessionBootstrap(true)
	if err != nil {
		t.Fatalf("local bootstrap: %v", err)
	}
	if packet.Source != "local_fallback" || !packet.Degraded {
		t.Fatalf("unexpected fallback packet: %+v", packet)
	}
	if packet.Knowledge.CurrentFocus != "Recover queue" {
		t.Fatalf("unexpected fallback knowledge: %+v", packet.Knowledge)
	}
	if packet.Handoff == nil || packet.ReviewRoom == nil || packet.Capabilities == nil {
		t.Fatalf("expected full fallback packet, got %+v", packet)
	}
}

func copyDir(t *testing.T, src string, dst string) {
	t.Helper()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(src, entry.Name()))
		if err != nil {
			t.Fatalf("read file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dst, entry.Name()), raw, 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}
}
