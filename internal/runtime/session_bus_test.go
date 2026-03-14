package runtime

import (
	"strconv"
	"testing"
	"time"
)

func TestCreateEventStampsActorAndPersists(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.WithActor("gemini", func(s *Service) error {
		_, err := s.CreateEvent(EventCreateInput{
			Kind:       "session.finished",
			Surface:    SurfaceAPI,
			WorkItemID: "system",
			Stage:      StageDiscover,
			Data:       map[string]any{"note": "resume"},
		})
		return err
	}); err != nil {
		t.Fatalf("create event: %v", err)
	}
	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	events := reloaded.ListEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Actor != "gemini" || events[0].Kind != "session.finished" {
		t.Fatalf("unexpected event: %+v", events[0])
	}
}

func TestSessionBootstrapFullIncludesReadOnlyProjections(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.ReplaceMemory(SystemMemory{CurrentFocus: "Recover queue", NextSteps: []string{"resume"}, OpenRisks: []string{"daemon_down"}, UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}
	packet := service.SessionBootstrap("daemon", false, true)
	if err := packet.Validate(); err != nil {
		t.Fatalf("validate packet: %v", err)
	}
	if packet.Source != "daemon" || packet.Degraded {
		t.Fatalf("unexpected packet metadata: %+v", packet)
	}
	if packet.Tooling == nil || packet.Tooling.Status == "" {
		t.Fatalf("expected tooling summary, got %+v", packet.Tooling)
	}
	if packet.Knowledge.CurrentFocus != "Recover queue" {
		t.Fatalf("unexpected knowledge: %+v", packet.Knowledge)
	}
	if packet.Handoff == nil || packet.ReviewRoom == nil || packet.Capabilities == nil {
		t.Fatalf("expected full packet, got %+v", packet)
	}
}

func TestCheckpointSessionPersistsAndBootstrapReturnsResume(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	goal := "hold terminal context"
	handoff := "resume from checkpoint"
	note := "mid-flight memory"
	var checkpoint SessionCheckpoint
	if err := service.WithActor("gemini", func(s *Service) error {
		item, err := s.CheckpointSession(SessionCheckpointInput{
			Source:       "terminal",
			CurrentFocus: "Lock queue",
			CurrentGoal:  &goal,
			NextSteps:    []string{"sync cockpit", "trim backlog"},
			OpenRisks:    []string{"daemon drift"},
			HandoffNote:  &handoff,
			Note:         &note,
			Refs:         []string{"thread:terminal-1", "proposal:queue"},
		})
		checkpoint = item
		return err
	}); err != nil {
		t.Fatalf("checkpoint session: %v", err)
	}
	if checkpoint.Actor != "gemini" || checkpoint.Source != "terminal" {
		t.Fatalf("unexpected checkpoint: %+v", checkpoint)
	}
	packet := service.SessionBootstrap("daemon", false, false)
	if packet.Resume == nil || packet.Resume.CurrentFocus != "Lock queue" {
		t.Fatalf("expected resume checkpoint, got %+v", packet)
	}
	if packet.Continuity == nil || packet.Continuity.Record == nil || packet.Continuity.Record.Actor != "gemini" {
		t.Fatalf("expected continuity record in bootstrap, got %+v", packet.Continuity)
	}
	if packet.Tooling == nil {
		t.Fatalf("expected tooling summary in thin bootstrap")
	}
	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	reloadedPacket := reloaded.SessionBootstrap("daemon", false, false)
	if reloadedPacket.Resume == nil || reloadedPacket.Resume.Actor != "gemini" {
		t.Fatalf("expected persisted resume checkpoint, got %+v", reloadedPacket)
	}
	if reloadedPacket.Continuity == nil || reloadedPacket.Continuity.Record == nil || reloadedPacket.Continuity.Record.CurrentFocus != "Lock queue" {
		t.Fatalf("expected persisted continuity record, got %+v", reloadedPacket.Continuity)
	}
}

func TestFinishSessionUpdatesMemoryAndRecordsEvent(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	goal := "close bus loop"
	handoff := "continue tomorrow"
	note := "session closed"
	result, err := service.FinishSession(SessionFinishInput{
		CurrentFocus: "Lock queue",
		CurrentGoal:  &goal,
		NextSteps:    []string{"sync cockpit", "trim backlog"},
		OpenRisks:    []string{"drift"},
		HandoffNote:  &handoff,
		Note:         &note,
	})
	if err != nil {
		t.Fatalf("finish session: %v", err)
	}
	memory := service.Memory()
	if memory.CurrentFocus != "Lock queue" || memory.CurrentGoal == nil || *memory.CurrentGoal != "close bus loop" {
		t.Fatalf("unexpected memory: %+v", memory)
	}
	if result.Event.Kind != "session.finished" {
		t.Fatalf("unexpected session event: %+v", result.Event)
	}
	if result.Event.Data["note"] != "session closed" {
		t.Fatalf("expected note in event data, got %+v", result.Event.Data)
	}
}

func TestFinishSessionClearsActiveCheckpoint(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.CheckpointSession(SessionCheckpointInput{
		Source:       "terminal",
		CurrentFocus: "Recover queue",
		NextSteps:    []string{"resume"},
		OpenRisks:    []string{"lost context"},
		Refs:         []string{"thread:terminal-2"},
	}); err != nil {
		t.Fatalf("checkpoint session: %v", err)
	}
	if _, err := service.FinishSession(SessionFinishInput{
		CurrentFocus: "Recover queue",
		NextSteps:    []string{"resume tomorrow"},
		OpenRisks:    []string{},
	}); err != nil {
		t.Fatalf("finish session: %v", err)
	}
	packet := service.SessionBootstrap("daemon", false, false)
	if packet.Resume != nil {
		t.Fatalf("expected cleared resume checkpoint, got %+v", packet.Resume)
	}
	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	reloadedPacket := reloaded.SessionBootstrap("daemon", false, false)
	if reloadedPacket.Resume != nil {
		t.Fatalf("expected persisted checkpoint to be cleared, got %+v", reloadedPacket.Resume)
	}
}

func TestSnapshotRoundTripPreservesSessionCheckpoint(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.CheckpointSession(SessionCheckpointInput{
		Source:       "terminal",
		CurrentFocus: "Keep context warm",
		NextSteps:    []string{"resume"},
		OpenRisks:    []string{"session loss"},
		Refs:         []string{"thread:terminal-3"},
	}); err != nil {
		t.Fatalf("checkpoint session: %v", err)
	}
	packet := service.Snapshot()
	if packet.SessionCheckpoint == nil || packet.SessionCheckpoint.CurrentFocus != "Keep context warm" {
		t.Fatalf("expected snapshot checkpoint, got %+v", packet.SessionCheckpoint)
	}
	if len(packet.ContinuityRecords) != 1 || packet.ContinuityRecords[0].CurrentFocus != "Keep context warm" {
		t.Fatalf("expected continuity records in snapshot, got %+v", packet.ContinuityRecords)
	}
	target, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new target service: %v", err)
	}
	if err := target.ImportSnapshot(packet); err != nil {
		t.Fatalf("import snapshot: %v", err)
	}
	resume := target.SessionBootstrap("daemon", false, false).Resume
	if resume == nil || resume.CurrentFocus != "Keep context warm" {
		t.Fatalf("expected imported checkpoint resume, got %+v", resume)
	}
}

func TestSessionNoteAppendsContinuityAndObservationWithoutAutomation(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.CheckpointSession(SessionCheckpointInput{
		Source:       "terminal",
		CurrentFocus: "Lock queue",
		NextSteps:    []string{},
		OpenRisks:    []string{},
		Refs:         []string{"thread:terminal-1"},
	}); err != nil {
		t.Fatalf("checkpoint session: %v", err)
	}
	result, err := service.SessionNote(SessionNoteInput{
		Source: "terminal",
		Kind:   "todo",
		Text:   "check queue health",
		Refs:   []string{"thread:terminal-1", "thread:terminal-1"},
	})
	if err != nil {
		t.Fatalf("session note: %v", err)
	}
	if result.Note.Kind != "todo" || result.Note.Text != "check queue health" {
		t.Fatalf("unexpected continuity note: %+v", result.Note)
	}
	if len(result.Record.NextSteps) != 1 || result.Record.NextSteps[0] != "check queue health" {
		t.Fatalf("expected todo note to project into next steps, got %+v", result.Record.NextSteps)
	}
	if result.Observation.Topic != continuityObservationTopic || result.Observation.SourceChannel != continuitySourceTerminal {
		t.Fatalf("unexpected continuity observation: %+v", result.Observation)
	}
	packet := service.SessionBootstrap("daemon", false, false)
	if packet.Continuity == nil || packet.Continuity.Record == nil || len(packet.Continuity.Record.Notes) != 1 {
		t.Fatalf("expected bootstrap continuity notes, got %+v", packet.Continuity)
	}
	observations := service.ListObservations(ObservationQuery{Topic: continuityObservationTopic, SourceChannel: continuitySourceTerminal})
	if len(observations) != 1 {
		t.Fatalf("expected one continuity observation, got %+v", observations)
	}
	if len(service.ListAgentJobs()) != 0 || len(service.ListProposals()) != 0 {
		t.Fatalf("session note must not auto-create automation, jobs=%d proposals=%d", len(service.ListAgentJobs()), len(service.ListProposals()))
	}
}

func TestSessionBootstrapContinuitySuggestionsRankRiskTodoThenRoute(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.CheckpointSession(SessionCheckpointInput{
		Source:       "terminal",
		CurrentFocus: "Lock queue",
		NextSteps:    []string{},
		OpenRisks:    []string{"daemon drift"},
		Refs:         []string{"thread:terminal-4"},
	}); err != nil {
		t.Fatalf("checkpoint session: %v", err)
	}
	if _, err := service.SessionNote(SessionNoteInput{
		Source: "terminal",
		Kind:   "todo",
		Text:   "capture retry plan",
		Refs:   []string{"thread:terminal-4"},
	}); err != nil {
		t.Fatalf("session note: %v", err)
	}
	why := "Why is queue drift still unresolved?"
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{
		Text:     "Queue drift still unresolved.",
		Kind:     "agenda",
		Severity: "high",
		Source:   "review.auto",
		Why:      why,
	}); err != nil {
		t.Fatalf("add review room item: %v", err)
	}
	packet := service.SessionBootstrap("daemon", false, false)
	if packet.Continuity == nil || len(packet.Continuity.Suggestions) < 3 {
		t.Fatalf("expected ranked continuity suggestions, got %+v", packet.Continuity)
	}
	if packet.Continuity.Suggestions[0].Kind != continuitySuggestionKindRisk {
		t.Fatalf("expected first suggestion to be risk, got %+v", packet.Continuity.Suggestions)
	}
	if packet.Continuity.Suggestions[1].Kind != continuitySuggestionKindTodo {
		t.Fatalf("expected second suggestion to be todo, got %+v", packet.Continuity.Suggestions)
	}
	if packet.Continuity.Suggestions[2].Kind != continuitySuggestionKindRoute {
		t.Fatalf("expected third suggestion to be route, got %+v", packet.Continuity.Suggestions)
	}
}

func TestSessionBootstrapMarksStaleContinuityAfterOneDay(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	packet := service.Snapshot()
	oldUpdated := time.Now().UTC().Add(-25 * time.Hour)
	packet.ContinuityRecords = []ContinuityRecord{{
		RecordID:     "continuity_record_stale",
		Source:       continuitySourceTerminal,
		Actor:        "system",
		Status:       continuityStatusActive,
		CurrentFocus: "Resume old thread",
		NextSteps:    []string{"resume"},
		OpenRisks:    []string{"context loss"},
		Refs:         []string{"thread:terminal-stale"},
		Notes:        []ContinuityNote{},
		CreatedAt:    oldUpdated.Add(-time.Hour),
		UpdatedAt:    oldUpdated,
	}}
	if err := service.ImportSnapshot(packet); err != nil {
		t.Fatalf("import snapshot: %v", err)
	}
	bootstrap := service.SessionBootstrap("daemon", false, false)
	if bootstrap.Continuity == nil || bootstrap.Continuity.Record == nil || bootstrap.Continuity.Record.Status != continuityStatusStale {
		t.Fatalf("expected stale continuity record, got %+v", bootstrap.Continuity)
	}
	if bootstrap.Resume == nil || bootstrap.Resume.CurrentFocus != "Resume old thread" {
		t.Fatalf("expected stale record to remain resumable, got %+v", bootstrap.Resume)
	}
}

func TestSessionNoteCapsContinuityHistoryAtFiftyEntries(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	packet := service.Snapshot()
	now := time.Now().UTC()
	notes := make([]ContinuityNote, 0, continuityNoteHistoryLimit)
	for idx := 0; idx < continuityNoteHistoryLimit; idx++ {
		text := "note " + strconv.Itoa(idx)
		notes = append(notes, ContinuityNote{
			NoteID:    "continuity_note_" + strconv.Itoa(idx),
			Kind:      continuityNoteKindNote,
			Text:      text,
			Refs:      []string{},
			CreatedAt: now.Add(time.Duration(idx) * time.Second),
		})
	}
	packet.ContinuityRecords = []ContinuityRecord{{
		RecordID:     "continuity_record_cap",
		Source:       continuitySourceTerminal,
		Actor:        "system",
		Status:       continuityStatusActive,
		CurrentFocus: "Keep note cap",
		NextSteps:    []string{},
		OpenRisks:    []string{},
		Refs:         []string{},
		Notes:        notes,
		CreatedAt:    now,
		UpdatedAt:    now.Add(time.Duration(continuityNoteHistoryLimit) * time.Second),
	}}
	if err := service.ImportSnapshot(packet); err != nil {
		t.Fatalf("import snapshot: %v", err)
	}
	result, err := service.SessionNote(SessionNoteInput{
		Source: continuitySourceTerminal,
		Kind:   continuityNoteKindNote,
		Text:   "fresh note",
		Refs:   []string{},
	})
	if err != nil {
		t.Fatalf("session note: %v", err)
	}
	if len(result.Record.Notes) != continuityNoteHistoryLimit {
		t.Fatalf("expected note history cap of %d, got %d", continuityNoteHistoryLimit, len(result.Record.Notes))
	}
	if result.Record.Notes[0].Text != "note 1" || result.Record.Notes[len(result.Record.Notes)-1].Text != "fresh note" {
		t.Fatalf("expected oldest note to roll off, got %+v", result.Record.Notes)
	}
}
