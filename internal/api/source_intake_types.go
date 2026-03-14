package api

import "time"

type sourceIntakeOption struct {
	Value string `json:"value,omitempty"`
	Label string `json:"label"`
	Cue   string `json:"cue,omitempty"`
}

type sourceIntakeRouteOption struct {
	RouteID string `json:"routeId"`
	Label   string `json:"label"`
	Cue     string `json:"cue"`
}

type sourceIntakeDraftSeedView struct {
	ObservationID            string    `json:"observationId"`
	CapturedAt               time.Time `json:"capturedAt"`
	Summary                  string    `json:"summary"`
	TargetAccount            string    `json:"targetAccount"`
	TargetAccountLabel       string    `json:"targetAccountLabel"`
	TargetTone               string    `json:"targetTone"`
	Title                    string    `json:"title"`
	SourceObservationID      string    `json:"sourceObservationId"`
	RouteDecisionID          string    `json:"routeDecisionId"`
	ParentDraftObservationID string    `json:"parentDraftObservationId"`
	RevisionNote             string    `json:"revisionNote"`
	SourceTitle              string    `json:"sourceTitle"`
	SourceURL                string    `json:"sourceURL"`
	FounderNote              string    `json:"founderNote"`
	DomainTags               []string  `json:"domainTags"`
	WorldviewTags            []string  `json:"worldviewTags"`
	Draft                    string    `json:"draft"`
	Preview                  string    `json:"preview"`
	Refs                     []string  `json:"refs"`
}

type sourceIntakeRouteDecisionView struct {
	ObservationID       string    `json:"observationId"`
	CapturedAt          time.Time `json:"capturedAt"`
	Summary             string    `json:"summary"`
	SourceObservationID string    `json:"sourceObservationId"`
	Decision            string    `json:"decision"`
	DecisionLabel       string    `json:"decisionLabel"`
	Title               string    `json:"title"`
	RouteSource         string    `json:"routeSource"`
	RouteSourceLabel    string    `json:"routeSourceLabel"`
	Refs                []string  `json:"refs"`
}

type sourceIntakePrepLaneView struct {
	ObservationID       string    `json:"observationId"`
	CapturedAt          time.Time `json:"capturedAt"`
	Summary             string    `json:"summary"`
	Channel             string    `json:"channel"`
	ChannelLabel        string    `json:"channelLabel"`
	TargetAccount       string    `json:"targetAccount"`
	TargetAccountLabel  string    `json:"targetAccountLabel"`
	Title               string    `json:"title"`
	BodyPreview         string    `json:"bodyPreview"`
	SourceObservationID string    `json:"sourceObservationId"`
	DraftObservationID  string    `json:"draftObservationId"`
	ApprovalID          string    `json:"approvalId"`
	FlowID              string    `json:"flowId"`
	Refs                []string  `json:"refs"`
}

type sourceIntakeItemView struct {
	ObservationID    string                          `json:"observationId"`
	CapturedAt       time.Time                       `json:"capturedAt"`
	Summary          string                          `json:"summary"`
	IntakeClass      string                          `json:"intakeClass"`
	IntakeClassLabel string                          `json:"intakeClassLabel"`
	PolicyColor      string                          `json:"policyColor"`
	PolicyColorLabel string                          `json:"policyColorLabel"`
	Title            string                          `json:"title"`
	URL              string                          `json:"url"`
	Excerpt          string                          `json:"excerpt"`
	FounderNote      string                          `json:"founderNote"`
	PriorityScore    int                             `json:"priorityScore"`
	Disposition      string                          `json:"disposition"`
	DispositionLabel string                          `json:"dispositionLabel"`
	DispositionNote  string                          `json:"dispositionNote"`
	DomainTags       []string                        `json:"domainTags"`
	WorldviewTags    []string                        `json:"worldviewTags"`
	SuggestedRoutes  []string                        `json:"suggestedRoutes"`
	Refs             []string                        `json:"refs"`
	FeedSource       string                          `json:"feedSource"`
	FeedKind         string                          `json:"feedKind"`
	OriginLabel      string                          `json:"originLabel"`
	FeedSourceLabel  string                          `json:"feedSourceLabel"`
	DraftSeed        *sourceIntakeDraftSeedView      `json:"draftSeed,omitempty"`
	DraftSeeds       []sourceIntakeDraftSeedView     `json:"draftSeeds,omitempty"`
	RouteDecision    *sourceIntakeRouteDecisionView  `json:"routeDecision,omitempty"`
	RouteDecisions   []sourceIntakeRouteDecisionView `json:"routeDecisions,omitempty"`
	PrepLane         *sourceIntakePrepLaneView       `json:"prepLane,omitempty"`
	PrepLanes        []sourceIntakePrepLaneView      `json:"prepLanes,omitempty"`
}

type sourceIntakeFormDefaults struct {
	IntakeClass     string   `json:"intakeClass"`
	PolicyColor     string   `json:"policyColor"`
	Title           string   `json:"title"`
	URL             string   `json:"url"`
	Excerpt         string   `json:"excerpt"`
	FounderNote     string   `json:"founderNote"`
	PriorityScore   int      `json:"priorityScore"`
	Disposition     string   `json:"disposition"`
	DispositionNote string   `json:"dispositionNote"`
	DomainTags      []string `json:"domainTags"`
	WorldviewTags   []string `json:"worldviewTags"`
	SuggestedRoutes []string `json:"suggestedRoutes"`
}

type sourceIntakeAttentionView struct {
	Mode                string `json:"mode"`
	Summary             string `json:"summary"`
	Detail              string `json:"detail"`
	Ref                 string `json:"ref,omitempty"`
	SourceObservationID string `json:"sourceObservationId,omitempty"`
	DraftObservationID  string `json:"draftObservationId,omitempty"`
	ActionLabel         string `json:"actionLabel,omitempty"`
}

type sourceIntakeView struct {
	GeneratedAt      time.Time                       `json:"generatedAt"`
	RuntimeAvailable bool                            `json:"runtimeAvailable"`
	DegradedReason   string                          `json:"degradedReason,omitempty"`
	SummaryNote      string                          `json:"summaryNote"`
	SummaryMeta      string                          `json:"summaryMeta"`
	QuietNote        string                          `json:"quietNote"`
	Defaults         sourceIntakeFormDefaults        `json:"defaults"`
	RouteOptions     []sourceIntakeRouteOption       `json:"routeOptions"`
	IntakeClasses    []sourceIntakeOption            `json:"intakeClasses"`
	PolicyColors     []sourceIntakeOption            `json:"policyColors"`
	Items            []sourceIntakeItemView          `json:"items"`
	ActionCount      int                             `json:"actionCount"`
	QuietCount       int                             `json:"quietCount"`
	QuietItems       []sourceIntakeItemView          `json:"quietItems"`
	Drafts           []sourceIntakeDraftSeedView     `json:"drafts"`
	RouteDecisions   []sourceIntakeRouteDecisionView `json:"routeDecisions"`
	PrepLanes        []sourceIntakePrepLaneView      `json:"prepLanes"`
	Attention        sourceIntakeAttentionView       `json:"attention"`
	RecentCount      int                             `json:"recentCount"`
	FeedCount        int                             `json:"feedCount"`
	DraftSeedCount   int                             `json:"draftSeedCount"`
	PrepCount        int                             `json:"prepCount"`
	ReviewCount      int                             `json:"reviewCount"`
	PrepReadyCount   int                             `json:"prepReadyCount"`
	AutoRoutedCount  int                             `json:"autoRoutedCount"`
}

type sourceIntakeCreateResult struct {
	Created    sourceIntakeItemView      `json:"created"`
	NextAction sourceIntakeAttentionView `json:"next_action"`
}
