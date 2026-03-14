package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func schemaRequiredFields(t *testing.T, schemaPath string) []string {
	t.Helper()
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read schema %s: %v", schemaPath, err)
	}
	var schema struct {
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("parse schema %s: %v", schemaPath, err)
	}
	return schema.Required
}

func structJSONTagSet(v any) map[string]bool {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	tags := make(map[string]bool)
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if name != "" && name != "-" {
			tags[name] = true
		}
	}
	return tags
}

// TestSchemaRequiredFieldsCoveredByGoStructs verifies that every field listed
// as required in a contract schema has a matching json tag in the Go struct.
// This catches the gap the compiler cannot: schema drift after a struct rename.
func TestSchemaRequiredFieldsCoveredByGoStructs(t *testing.T) {
	contractsDir := filepath.Join("..", "..", "contracts")

	cases := map[string]any{
		"adapter_summary.schema.json":          AdapterSummary{},
		"agent_dispatch_profile.schema.json":   AgentDispatchProfile{},
		"agent_dispatch_result.schema.json":    AgentDispatchResult{},
		"agent_job.schema.json":                AgentJob{},
		"agent_job_report_result.schema.json":  AgentJobReportResult{},
		"agent_run_packet.schema.json":         AgentRunPacket{},
		"approval_item.schema.json":            ApprovalItem{},
		"auth_status.schema.json":              AuthStatus{},
		"authority_boundary.schema.json":       AuthorityBoundary{},
		"branch.schema.json":                   Branch{},
		"capability_registry.schema.json":      CapabilityRegistry{},
		"capitalization_entry.schema.json":     CapitalizationEntry{},
		"chain_rules.schema.json":              ChainRules{},
		"company_state.schema.json":            CompanyState{},
		"daemon_status.schema.json":            DaemonStatus{},
		"deploy_run.schema.json":               DeployRun{},
		"deploy_target.schema.json":            DeployTarget{},
		"event_envelope.schema.json":           EventEnvelope{},
		"execute_run.schema.json":              ExecuteRun{},
		"flow_run.schema.json":                 FlowRun{},
		"founder_summary.schema.json":          FounderSummary{},
		"founder_view.schema.json":             FounderView{},
		"gateway_call.schema.json":             GatewayCall{},
		"handoff_packet.schema.json":           HandoffPacket{},
		"knowledge_packet.schema.json":         KnowledgePacket{},
		"observation_record.schema.json":       ObservationRecord{},
		"open_thread.schema.json":              OpenThread{},
		"policy_decision.schema.json":          PolicyDecision{},
		"preflight_record.schema.json":         PreflightRecord{},
		"proposal_item.schema.json":            ProposalItem{},
		"provider_summary.schema.json":         ProviderSummary{},
		"release_packet.schema.json":           ReleasePacket{},
		"review_room.schema.json":              ReviewRoom{},
		"review_room_summary.schema.json":      ReviewRoomSummary{},
		"rollback_run.schema.json":             RollbackRun{},
		"session_bootstrap_packet.schema.json": SessionBootstrapPacket{},
		"snapshot_packet.schema.json":          SnapshotPacket{},
		"system_memory.schema.json":            SystemMemory{},
		"telegram_packet.schema.json":          TelegramPacket{},
		"verification_record.schema.json":      VerificationRecord{},
		"work_item.schema.json":                WorkItem{},
		"write_lease.schema.json":              WriteLease{},
	}

	expected := CanonicalContractFilenames()
	actual := sortedUniqueStrings(mapsKeys(cases))
	if missing := setDifference(expected, actual); len(missing) > 0 {
		t.Fatalf("schema fixture mapping missing contracts: %v", missing)
	}
	if unexpected := setDifference(actual, expected); len(unexpected) > 0 {
		t.Fatalf("schema fixture mapping has unexpected contracts: %v", unexpected)
	}

	for _, schema := range actual {
		schema := schema
		t.Run(schema, func(t *testing.T) {
			required := schemaRequiredFields(t, filepath.Join(contractsDir, schema))
			tags := structJSONTagSet(cases[schema])
			for _, field := range required {
				if !tags[field] {
					t.Errorf("%s: schema requires %q but Go struct has no matching json tag", schema, field)
				}
			}
		})
	}
}

func mapsKeys(items map[string]any) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	return keys
}
