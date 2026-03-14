package runtime

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var requiredMCPServers = []string{}

var optionalMCPServers = []string{
	"sequential-thinking",
	"playwright",
	"notebooklm",
	"omx_code_intel",
	"omx_memory",
	"omx_state",
	"omx_team_run",
	"omx_trace",
}

type MCPServerAudit struct {
	Name       string   `json:"name"`
	Required   bool     `json:"required"`
	Registered bool     `json:"registered"`
	Enabled    bool     `json:"enabled"`
	Command    string   `json:"command,omitempty"`
	Args       string   `json:"args,omitempty"`
	Issues     []string `json:"issues"`
	Checks     []string `json:"checks"`
}

type MCPAudit struct {
	Status            string           `json:"status"`
	ConfigPath        string           `json:"config_path"`
	RequiredServers   []string         `json:"required_servers"`
	OptionalServers   []string         `json:"optional_servers"`
	Servers           []MCPServerAudit `json:"servers"`
	Issues            []string         `json:"issues"`
	SuggestedCommands []string         `json:"suggested_commands"`
}

func AuditMCP(_ string) MCPAudit {
	return auditMCPWithRunner(execCommandRunner{})
}

type commandRunner interface {
	run(name string, args ...string) ([]byte, error)
}

type execCommandRunner struct{}

func (execCommandRunner) run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

func auditMCPWithRunner(runner commandRunner) MCPAudit {
	configPath := codexConfigPath()
	audit := MCPAudit{
		Status:          "ok",
		ConfigPath:      configPath,
		RequiredServers: append([]string{}, requiredMCPServers...),
		OptionalServers: append([]string{}, optionalMCPServers...),
		Servers:         []MCPServerAudit{},
		Issues:          []string{},
		SuggestedCommands: []string{
			"layer-osctl audit mcp --strict",
			"codex mcp list",
			"./scripts/check_codex_mcp.sh",
		},
	}

	configRaw, err := os.ReadFile(configPath)
	if err != nil {
		if len(requiredMCPServers) > 0 {
			audit.Status = "degraded"
			audit.Issues = append(audit.Issues, "codex config unreadable: "+configPath)
		}
		return audit
	}
	configText := string(configRaw)

	serverNames := append([]string{}, requiredMCPServers...)
	serverNames = append(serverNames, optionalMCPServers...)
	for _, name := range serverNames {
		server := MCPServerAudit{
			Name:     name,
			Required: containsAuditString(requiredMCPServers, name),
			Issues:   []string{},
			Checks:   []string{},
		}
		if strings.Contains(configText, "[mcp_servers."+name+"]") {
			server.Checks = append(server.Checks, "config stanza present")
		} else {
			server.Issues = append(server.Issues, "config stanza missing")
		}
		raw, err := runner.run("codex", "mcp", "get", name)
		if err != nil {
			server.Issues = append(server.Issues, "not registered")
			audit.Servers = append(audit.Servers, server)
			if server.Required {
				audit.Issues = append(audit.Issues, name+": not registered")
			}
			continue
		}
		server.Registered = true
		server.Checks = append(server.Checks, "registered")
		server.Enabled = strings.Contains(string(raw), "enabled: true")
		if server.Enabled {
			server.Checks = append(server.Checks, "enabled")
		} else {
			server.Issues = append(server.Issues, "not enabled")
		}
		server.Command = parseMCPField(string(raw), "command")
		server.Args = parseMCPField(string(raw), "args")
		switch strings.TrimSpace(server.Command) {
		case "node":
			entrypoint := firstMCPArg(server.Args)
			if entrypoint == "" {
				server.Issues = append(server.Issues, "node entrypoint missing")
			} else if _, err := os.Stat(entrypoint); err != nil {
				server.Issues = append(server.Issues, "node entrypoint missing")
			} else {
				server.Checks = append(server.Checks, "node entrypoint exists")
			}
		case "npx":
			if strings.TrimSpace(server.Args) == "" {
				server.Issues = append(server.Issues, "npx package missing")
			} else {
				server.Checks = append(server.Checks, "npx package declared")
			}
		default:
			if strings.TrimSpace(server.Command) == "" {
				server.Issues = append(server.Issues, "command missing")
			}
		}
		if len(server.Issues) > 0 && server.Required {
			audit.Issues = append(audit.Issues, name+": "+strings.Join(server.Issues, ", "))
		}
		audit.Servers = append(audit.Servers, server)
	}

	sort.Strings(audit.RequiredServers)
	sort.Strings(audit.OptionalServers)
	sort.Slice(audit.Servers, func(i, j int) bool { return audit.Servers[i].Name < audit.Servers[j].Name })
	sort.Strings(audit.Issues)
	if len(audit.Issues) > 0 {
		audit.Status = "degraded"
	}
	return audit
}

func codexConfigPath() string {
	if value := strings.TrimSpace(os.Getenv("CODEX_CONFIG")); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".codex", "config.toml")
	}
	return filepath.Join(home, ".codex", "config.toml")
}

func parseMCPField(raw string, name string) string {
	prefix := "  " + strings.TrimSpace(name) + ": "
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func firstMCPArg(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	// Pragmatic fix: if the whole argument string is a valid file,
	// assume it's a path with spaces (common in macOS environments).
	if _, err := os.Stat(raw); err == nil {
		return raw
	}
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func containsAuditString(items []string, needle string) bool {
	for _, item := range items {
		if item == needle {
			return true
		}
	}
	return false
}
