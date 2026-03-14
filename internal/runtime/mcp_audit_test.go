package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditMCPOkWithRequiredAndOptionalServers(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config.toml")
	entrypoint := filepath.Join(root, "sequential.js")
	mustWriteMCPAuditFile(t, entrypoint, "console.log('ok')")
	mustWriteMCPAuditFile(t, configPath, `
[mcp_servers.sequential-thinking]
command = "node"
args = ["/tmp/sequential.js"]

[mcp_servers.playwright]
command = "npx"
args = ["@playwright/mcp@latest"]

[mcp_servers.notebooklm]
command = "npx"
args = ["@google-labs/notebooklm-mcp@latest"]

[mcp_servers.omx_code_intel]
command = "node"
args = ["/tmp/intel.js"]

[mcp_servers.omx_memory]
command = "node"
args = ["/tmp/memory.js"]

[mcp_servers.omx_state]
command = "node"
args = ["/tmp/state.js"]

[mcp_servers.omx_team_run]
command = "node"
args = ["/tmp/team-run.js"]

[mcp_servers.omx_trace]
command = "node"
args = ["/tmp/trace.js"]
`)
	t.Setenv("CODEX_CONFIG", configPath)
	runner := fakeMCPRunner{
		outputs: map[string]string{
			"sequential-thinking": "sequential-thinking\n  enabled: true\n  command: node\n  args: " + entrypoint + "\n",
			"playwright":          "playwright\n  enabled: true\n  command: npx\n  args: @playwright/mcp@latest\n",
			"notebooklm":          "notebooklm\n  enabled: true\n  command: npx\n  args: @google-labs/notebooklm-mcp@latest\n",
			"omx_code_intel":      "omx_code_intel\n  enabled: true\n  command: node\n  args: " + entrypoint + "\n",
			"omx_memory":          "omx_memory\n  enabled: true\n  command: node\n  args: " + entrypoint + "\n",
			"omx_state":           "omx_state\n  enabled: true\n  command: node\n  args: " + entrypoint + "\n",
			"omx_team_run":        "omx_team_run\n  enabled: true\n  command: node\n  args: " + entrypoint + "\n",
			"omx_trace":           "omx_trace\n  enabled: true\n  command: node\n  args: " + entrypoint + "\n",
		},
	}

	audit := auditMCPWithRunner(runner)
	if audit.Status != "ok" {
		t.Fatalf("expected ok, got %+v", audit)
	}
	if len(audit.Issues) != 0 {
		t.Fatalf("expected no issues, got %v", audit.Issues)
	}
}

func TestAuditMCPDegradedWhenRequiredServerMissing(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config.toml")
	mustWriteMCPAuditFile(t, configPath, "")
	t.Setenv("CODEX_CONFIG", configPath)
	runner := fakeMCPRunner{
		outputs: map[string]string{},
		errs:    map[string]error{},
	}

	audit := auditMCPWithRunner(runner)
	if audit.Status != "ok" {
		t.Fatalf("expected optional-only mcp audit to stay ok, got %+v", audit)
	}
	if len(audit.Issues) != 0 {
		t.Fatalf("expected no required issues, got %v", audit.Issues)
	}
	if len(audit.Servers) == 0 {
		t.Fatalf("expected optional server details, got %+v", audit)
	}
	joined := strings.Join(audit.Servers[0].Issues, "\n")
	if len(joined) == 0 {
		t.Fatalf("expected per-server optional issues, got %+v", audit.Servers[0])
	}
}

type fakeMCPRunner struct {
	outputs map[string]string
	errs    map[string]error
}

func (f fakeMCPRunner) run(_ string, args ...string) ([]byte, error) {
	name := args[len(args)-1]
	if err, ok := f.errs[name]; ok {
		return nil, err
	}
	return []byte(f.outputs[name]), nil
}

func mustWriteMCPAuditFile(t *testing.T, path string, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
