package runtime

import (
	"errors"
	"strings"
	"time"
)

type CorpusLesson struct {
	EntryID    string    `json:"entry_id"`
	SourceKind string    `json:"source_kind"`
	Summary    string    `json:"summary"`
	CreatedAt  time.Time `json:"created_at"`
}

type ObservationQuery struct {
	SourceChannel string `json:"source_channel,omitempty"`
	Topic         string `json:"topic,omitempty"`
	Actor         string `json:"actor,omitempty"`
	Ref           string `json:"ref,omitempty"`
	Text          string `json:"text,omitempty"`
	Limit         int    `json:"limit,omitempty"`
}

type ObservationRecord struct {
	ObservationID     string    `json:"observation_id"`
	SourceChannel     string    `json:"source_channel"`
	CapturedAt        time.Time `json:"captured_at"`
	Actor             string    `json:"actor"`
	Topic             string    `json:"topic"`
	Refs              []string  `json:"refs"`
	Confidence        string    `json:"confidence"`
	RawExcerpt        string    `json:"raw_excerpt"`
	NormalizedSummary string    `json:"normalized_summary"`
}

type ObservationLink struct {
	LinkID   string    `json:"link_id"`
	Kind     string    `json:"kind"`
	Summary  string    `json:"summary"`
	Channels []string  `json:"channels"`
	Refs     []string  `json:"refs"`
	Count    int       `json:"count"`
	LatestAt time.Time `json:"latest_at"`
}

type ProposalCandidate struct {
	CandidateID    string   `json:"candidate_id"`
	ProposalID     string   `json:"proposal_id"`
	WorkItemID     string   `json:"work_item_id"`
	Title          string   `json:"title"`
	Intent         string   `json:"intent"`
	Summary        string   `json:"summary"`
	Priority       string   `json:"priority"`
	Risk           string   `json:"risk"`
	Surface        Surface  `json:"surface"`
	Source         string   `json:"source"`
	Notes          []string `json:"notes"`
	LinkIDs        []string `json:"link_ids"`
	ThreadIDs      []string `json:"thread_ids"`
	Refs           []string `json:"refs"`
	CreatePath     string   `json:"create_path"`
	CreateCommand  string   `json:"create_command"`
	PromotePath    string   `json:"promote_path"`
	PromoteCommand string   `json:"promote_command"`
}

type OpenThread struct {
	ThreadID    string   `json:"thread_id"`
	Question    string   `json:"question"`
	Status      string   `json:"status"`
	PatternRefs []string `json:"pattern_refs"`
	Evidence    []string `json:"evidence"`
	Source      string   `json:"source"`
}

type KnowledgeSearchResult struct {
	Entry       CapitalizationEntry `json:"entry"`
	MatchFields []string            `json:"match_fields"`
	MatchCount  int                 `json:"match_count"`
}

type KnowledgeSearchResponse struct {
	Query       string                  `json:"query"`
	Results     []KnowledgeSearchResult `json:"results"`
	AutoThreads []OpenThread            `json:"auto_threads"`
}

type ActionRoute struct {
	RouteID    string  `json:"route_id"`
	Kind       string  `json:"kind"`
	Summary    string  `json:"summary"`
	TargetLane string  `json:"target_lane"`
	TargetRef  *string `json:"target_ref,omitempty"`
	Command    *string `json:"command,omitempty"`
	Source     string  `json:"source"`
}

type KnowledgePacket struct {
	GeneratedAt         time.Time               `json:"generated_at"`
	EnvironmentAdvisory EnvironmentAdvisory     `json:"environment_advisory"`
	Authority           AuthorityBoundary       `json:"authority"`
	CurrentFocus        string                  `json:"current_focus"`
	CurrentGoal         *string                 `json:"current_goal,omitempty"`
	NextSteps           []string                `json:"next_steps"`
	OpenRisks           []string                `json:"open_risks"`
	PrimaryAction       string                  `json:"primary_action"`
	PrimaryRef          string                  `json:"primary_ref"`
	PriorityRationale   *PriorityRationale      `json:"priority_rationale,omitempty"`
	ReviewOpenCount     int                     `json:"review_open_count"`
	ReviewTopOpen       []string                `json:"review_top_open"`
	CorpusLessons       []CorpusLesson          `json:"corpus_lessons"`
	CorpusResults       []KnowledgePacketResult `json:"corpus_results,omitempty"`
	Surprising          []string                `json:"surprising"`
	ObservationLinks    []ObservationLink       `json:"observation_links"`
	ProposalCandidates  []ProposalCandidate     `json:"proposal_candidates"`
	OpenThreads         []OpenThread            `json:"open_threads"`
	ActionHints         []string                `json:"action_hints"`
	ActionRoutes        []ActionRoute           `json:"action_routes"`
	ActiveBranches      []Branch                `json:"active_branches"`
	ParallelCandidates  []ParallelCandidate     `json:"parallel_candidates"`
	DefaultActor        string                  `json:"default_actor"`
	Actors              []string                `json:"actors"`
	Providers           []string                `json:"providers"`
	GatewaySemantics    string                  `json:"gateway_semantics"`
}

func validObservationConfidence(value string) bool {
	switch strings.TrimSpace(value) {
	case "low", "medium", "high":
		return true
	default:
		return false
	}
}

func validObservationLinkKind(value string) bool {
	switch strings.TrimSpace(value) {
	case "ref", "semantic":
		return true
	default:
		return false
	}
}

func validActionRouteTargetLane(value string) bool {
	switch strings.TrimSpace(value) {
	case "proposal", "review_room", "session_memory", "job":
		return true
	default:
		return false
	}
}

func (k KnowledgePacket) Validate() error {
	if k.GeneratedAt.IsZero() {
		return errors.New("knowledge packet generated_at is required")
	}
	if err := k.EnvironmentAdvisory.Validate(); err != nil {
		return err
	}
	if err := k.Authority.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(k.CurrentFocus) == "" {
		return errors.New("knowledge packet current_focus is required")
	}
	if k.NextSteps == nil {
		return errors.New("knowledge packet next_steps is required")
	}
	if k.OpenRisks == nil {
		return errors.New("knowledge packet open_risks is required")
	}
	if strings.TrimSpace(k.PrimaryAction) == "" {
		return errors.New("knowledge packet primary_action is required")
	}
	if k.ReviewTopOpen == nil {
		return errors.New("knowledge packet review_top_open is required")
	}
	if k.CorpusLessons == nil {
		return errors.New("knowledge packet corpus_lessons is required")
	}
	for _, item := range k.CorpusLessons {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	if k.Surprising == nil {
		return errors.New("knowledge packet surprising is required")
	}
	if k.ObservationLinks == nil {
		return errors.New("knowledge packet observation_links is required")
	}
	for _, item := range k.ObservationLinks {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	if k.ProposalCandidates == nil {
		return errors.New("knowledge packet proposal_candidates is required")
	}
	for _, item := range k.ProposalCandidates {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	if k.OpenThreads == nil {
		return errors.New("knowledge packet open_threads is required")
	}
	for _, item := range k.OpenThreads {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	if k.ActionHints == nil {
		return errors.New("knowledge packet action_hints is required")
	}
	for _, item := range k.ActionHints {
		if strings.TrimSpace(item) == "" {
			return errors.New("knowledge packet action_hints must not contain empty items")
		}
	}
	if k.ActionRoutes == nil {
		return errors.New("knowledge packet action_routes is required")
	}
	for _, item := range k.ActionRoutes {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	if k.ActiveBranches == nil {
		return errors.New("knowledge packet active_branches is required")
	}
	for _, item := range k.ActiveBranches {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	if k.ParallelCandidates == nil {
		return errors.New("knowledge packet parallel_candidates is required")
	}
	if strings.TrimSpace(k.DefaultActor) == "" {
		return errors.New("knowledge packet default_actor is required")
	}
	if k.Actors == nil {
		return errors.New("knowledge packet actors is required")
	}
	if k.Providers == nil {
		return errors.New("knowledge packet providers is required")
	}
	if strings.TrimSpace(k.GatewaySemantics) == "" {
		return errors.New("knowledge packet gateway_semantics is required")
	}
	return nil
}

func (l CorpusLesson) Validate() error {
	if strings.TrimSpace(l.EntryID) == "" {
		return errors.New("corpus lesson entry_id is required")
	}
	if strings.TrimSpace(l.SourceKind) == "" {
		return errors.New("corpus lesson source_kind is required")
	}
	if strings.TrimSpace(l.Summary) == "" {
		return errors.New("corpus lesson summary is required")
	}
	if l.CreatedAt.IsZero() {
		return errors.New("corpus lesson created_at is required")
	}
	return nil
}

func (l OpenThread) Validate() error {
	if strings.TrimSpace(l.ThreadID) == "" {
		return errors.New("open thread thread_id is required")
	}
	if strings.TrimSpace(l.Question) == "" {
		return errors.New("open thread question is required")
	}
	status := strings.TrimSpace(strings.ToLower(l.Status))
	if status != "open" && status != "resolved" && status != "dismissed" {
		return errors.New("open thread status is invalid")
	}
	if l.PatternRefs == nil {
		return errors.New("open thread pattern_refs is required")
	}
	for _, item := range l.PatternRefs {
		if strings.TrimSpace(item) == "" {
			return errors.New("open thread pattern_refs must not contain empty items")
		}
	}
	if l.Evidence == nil {
		return errors.New("open thread evidence is required")
	}
	for _, item := range l.Evidence {
		if strings.TrimSpace(item) == "" {
			return errors.New("open thread evidence must not contain empty items")
		}
	}
	if strings.TrimSpace(l.Source) == "" {
		return errors.New("open thread source is required")
	}
	source := strings.TrimSpace(strings.ToLower(l.Source))
	if source != "corpus_signal" && source != "review_room_drift" && source != "manual" {
		return errors.New("open thread source is invalid")
	}
	return nil
}

func (a ActionRoute) Validate() error {
	if strings.TrimSpace(a.RouteID) == "" {
		return errors.New("action route route_id is required")
	}
	if strings.TrimSpace(a.Kind) == "" {
		return errors.New("action route kind is required")
	}
	if strings.TrimSpace(a.Summary) == "" {
		return errors.New("action route summary is required")
	}
	if !validActionRouteTargetLane(a.TargetLane) {
		return errors.New("action route target_lane is invalid")
	}
	if a.TargetRef != nil && strings.TrimSpace(*a.TargetRef) == "" {
		return errors.New("action route target_ref must not be empty")
	}
	if a.Command != nil && strings.TrimSpace(*a.Command) == "" {
		return errors.New("action route command must not be empty")
	}
	if strings.TrimSpace(a.Source) == "" {
		return errors.New("action route source is required")
	}
	return nil
}

func (o ObservationRecord) Validate() error {
	if strings.TrimSpace(o.ObservationID) == "" {
		return errors.New("observation record observation_id is required")
	}
	if strings.TrimSpace(o.SourceChannel) == "" {
		return errors.New("observation record source_channel is required")
	}
	if o.CapturedAt.IsZero() {
		return errors.New("observation record captured_at is required")
	}
	if strings.TrimSpace(o.Actor) == "" {
		return errors.New("observation record actor is required")
	}
	if strings.TrimSpace(o.Topic) == "" {
		return errors.New("observation record topic is required")
	}
	if o.Refs == nil {
		return errors.New("observation record refs is required")
	}
	if !validObservationConfidence(o.Confidence) {
		return errors.New("observation record confidence is invalid")
	}
	if strings.TrimSpace(o.RawExcerpt) == "" {
		return errors.New("observation record raw_excerpt is required")
	}
	if strings.TrimSpace(o.NormalizedSummary) == "" {
		return errors.New("observation record normalized_summary is required")
	}
	return nil
}

func (o ObservationLink) Validate() error {
	if strings.TrimSpace(o.LinkID) == "" {
		return errors.New("observation link link_id is required")
	}
	if !validObservationLinkKind(o.Kind) {
		return errors.New("observation link kind is invalid")
	}
	if strings.TrimSpace(o.Summary) == "" {
		return errors.New("observation link summary is required")
	}
	if o.Channels == nil {
		return errors.New("observation link channels is required")
	}
	for _, item := range o.Channels {
		if strings.TrimSpace(item) == "" {
			return errors.New("observation link channels must not contain empty items")
		}
	}
	if o.Refs == nil {
		return errors.New("observation link refs is required")
	}
	if o.Count < 2 {
		return errors.New("observation link count must be at least 2")
	}
	if o.LatestAt.IsZero() {
		return errors.New("observation link latest_at is required")
	}
	return nil
}

func (p ProposalCandidate) Validate() error {
	if strings.TrimSpace(p.CandidateID) == "" {
		return errors.New("proposal candidate candidate_id is required")
	}
	if strings.TrimSpace(p.ProposalID) == "" {
		return errors.New("proposal candidate proposal_id is required")
	}
	if strings.TrimSpace(p.WorkItemID) == "" {
		return errors.New("proposal candidate work_item_id is required")
	}
	if strings.TrimSpace(p.Title) == "" {
		return errors.New("proposal candidate title is required")
	}
	if strings.TrimSpace(p.Intent) == "" {
		return errors.New("proposal candidate intent is required")
	}
	if strings.TrimSpace(p.Summary) == "" {
		return errors.New("proposal candidate summary is required")
	}
	if strings.TrimSpace(p.Priority) == "" {
		return errors.New("proposal candidate priority is required")
	}
	if strings.TrimSpace(p.Risk) == "" {
		return errors.New("proposal candidate risk is required")
	}
	if !validSurface(p.Surface) {
		return errors.New("proposal candidate surface is invalid")
	}
	if strings.TrimSpace(p.Source) == "" {
		return errors.New("proposal candidate source is required")
	}
	if p.Notes == nil {
		return errors.New("proposal candidate notes is required")
	}
	for _, item := range p.Notes {
		if strings.TrimSpace(item) == "" {
			return errors.New("proposal candidate notes must not contain empty items")
		}
	}
	if p.LinkIDs == nil {
		return errors.New("proposal candidate link_ids is required")
	}
	if len(p.LinkIDs) == 0 {
		return errors.New("proposal candidate link_ids must not be empty")
	}
	for _, item := range p.LinkIDs {
		if strings.TrimSpace(item) == "" {
			return errors.New("proposal candidate link_ids must not contain empty items")
		}
	}
	if p.ThreadIDs == nil {
		return errors.New("proposal candidate thread_ids is required")
	}
	for _, item := range p.ThreadIDs {
		if strings.TrimSpace(item) == "" {
			return errors.New("proposal candidate thread_ids must not contain empty items")
		}
	}
	if p.Refs == nil {
		return errors.New("proposal candidate refs is required")
	}
	for _, item := range p.Refs {
		if strings.TrimSpace(item) == "" {
			return errors.New("proposal candidate refs must not contain empty items")
		}
	}
	if strings.TrimSpace(p.CreatePath) == "" {
		return errors.New("proposal candidate create_path is required")
	}
	if strings.TrimSpace(p.CreateCommand) == "" {
		return errors.New("proposal candidate create_command is required")
	}
	if strings.TrimSpace(p.PromotePath) == "" {
		return errors.New("proposal candidate promote_path is required")
	}
	if strings.TrimSpace(p.PromoteCommand) == "" {
		return errors.New("proposal candidate promote_command is required")
	}
	return nil
}
