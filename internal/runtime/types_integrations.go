package runtime

import (
	"errors"
	"strings"
	"time"
)

type AdapterSummary struct {
	Gateway                string `json:"gateway"`
	GatewaySemantics       string `json:"gateway_semantics"`
	GatewayDispatchEnabled bool   `json:"gateway_dispatch_enabled"`
	GatewayRequiredMode    string `json:"gateway_required_mode"`
	Verify                 string `json:"verify"`
	Deploy                 string `json:"deploy"`
	Rollback               string `json:"rollback"`
}

type ProviderSummary struct {
	Provider        string   `json:"provider"`
	Endpoint        *string  `json:"endpoint,omitempty"`
	Declared        bool     `json:"declared"`
	DispatchEnabled bool     `json:"dispatch_enabled"`
	AuthReady       bool     `json:"auth_ready"`
	AuthSource      *string  `json:"auth_source,omitempty"`
	AuthEnvKeys     []string `json:"auth_env_keys"`
	GatewayAdapter  string   `json:"gateway_adapter"`
	Semantics       string   `json:"semantics"`
	Notes           []string `json:"notes"`
}

type TelegramStatus struct {
	Adapter           string                `json:"adapter"`
	SendAdapter       string                `json:"send_adapter"`
	SendConfigured    bool                  `json:"send_configured"`
	PollingConfigured bool                  `json:"polling_configured"`
	ChatConfigured    bool                  `json:"chat_configured"`
	GeminiConfigured  bool                  `json:"gemini_configured"`
	InboundMode       string                `json:"inbound_mode"`
	FounderDelivery   string                `json:"founder_delivery"`
	Routes            []TelegramRouteStatus `json:"routes"`
	Notes             []string              `json:"notes"`
}

type TelegramRouteStatus struct {
	RouteID        string   `json:"route_id"`
	Label          string   `json:"label"`
	ChatConfigured bool     `json:"chat_configured"`
	Delivery       string   `json:"delivery"`
	Notes          []string `json:"notes"`
}

type ThreadsStatus struct {
	Adapter           string   `json:"adapter"`
	PublishConfigured bool     `json:"publish_configured"`
	Notes             []string `json:"notes"`
}

type ThreadsPublishReceipt struct {
	ApprovalID    string    `json:"approval_id"`
	ProposalID    string    `json:"proposal_id,omitempty"`
	WorkItemID    string    `json:"work_item_id,omitempty"`
	FlowID        string    `json:"flow_id,omitempty"`
	ObservationID string    `json:"observation_id"`
	CreationID    string    `json:"creation_id"`
	ThreadID      string    `json:"thread_id"`
	TargetAccount string    `json:"target_account,omitempty"`
	Title         string    `json:"title"`
	TopicTag      string    `json:"topic_tag,omitempty"`
	SourceIDs     []string  `json:"source_ids"`
	PublishedAt   time.Time `json:"published_at"`
}

type TelegramPacket struct {
	GeneratedAt     time.Time `json:"generated_at"`
	Headline        string    `json:"headline"`
	BodyLines       []string  `json:"body_lines"`
	PrimaryAction   string    `json:"primary_action"`
	PrimaryRef      string    `json:"primary_ref"`
	CurrentFocus    string    `json:"current_focus"`
	CurrentGoal     *string   `json:"current_goal,omitempty"`
	NextSteps       []string  `json:"next_steps"`
	OpenRisks       []string  `json:"open_risks"`
	ReviewOpenCount int       `json:"review_open_count"`
	ReviewTopOpen   []string  `json:"review_top_open"`
	FounderNotice   string    `json:"founder_notice"`
	RecommendedMode string    `json:"recommended_mode"`
}

type ToolingHealth struct {
	Status           string   `json:"status"`
	RequiredMCPReady bool     `json:"required_mcp_ready"`
	SessionNote      *string  `json:"session_note,omitempty"`
	MissingRequired  []string `json:"missing_required"`
}

func (t ToolingHealth) Validate() error {
	if strings.TrimSpace(t.Status) == "" {
		return errors.New("tooling health status is required")
	}
	if t.MissingRequired == nil {
		return errors.New("tooling health missing_required is required")
	}
	return nil
}
