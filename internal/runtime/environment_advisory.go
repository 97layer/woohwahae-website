package runtime

import (
	"errors"
	"os"
	"strings"
)

type EnvironmentAdvisory struct {
	HostClass      string `json:"host_class"`
	PowerMode      string `json:"power_mode"`
	ContinuityRole string `json:"continuity_role"`
	AgentMode      string `json:"agent_mode"`
	FounderNotice  string `json:"founder_notice"`
	Rule           string `json:"rule"`
}

func (e EnvironmentAdvisory) Validate() error {
	if strings.TrimSpace(e.HostClass) == "" {
		return errors.New("environment advisory host_class is required")
	}
	if strings.TrimSpace(e.PowerMode) == "" {
		return errors.New("environment advisory power_mode is required")
	}
	if strings.TrimSpace(e.ContinuityRole) == "" {
		return errors.New("environment advisory continuity_role is required")
	}
	if strings.TrimSpace(e.AgentMode) == "" {
		return errors.New("environment advisory agent_mode is required")
	}
	if strings.TrimSpace(e.FounderNotice) == "" {
		return errors.New("environment advisory founder_notice is required")
	}
	if strings.TrimSpace(e.Rule) == "" {
		return errors.New("environment advisory rule is required")
	}
	return nil
}

func currentEnvironmentAdvisory() EnvironmentAdvisory {
	hostClass := normalizeHostClass(os.Getenv("LAYER_OS_HOST_CLASS"))
	powerMode := normalizePowerMode(os.Getenv("LAYER_OS_POWER_MODE"))
	continuityRole := continuityRoleForHost(hostClass)

	advisory := EnvironmentAdvisory{
		HostClass:      hostClass,
		PowerMode:      powerMode,
		ContinuityRole: continuityRole,
	}

	switch {
	case powerMode == "low":
		advisory.AgentMode = "conserve"
		advisory.FounderNotice = "Low-power host detected; throttle non-critical agent work and defer heavy verification, release, and long fan-out tasks."
		advisory.Rule = "environment_advisory.low_power_conserve"
	case continuityRole == "burst_workstation":
		advisory.AgentMode = "burst_only"
		advisory.FounderNotice = "Burst workstation host detected; keep always-on runtime continuity on the VM and use this node for local iteration or recovery only."
		advisory.Rule = "environment_advisory.burst_workstation"
	case continuityRole == "continuity_host":
		advisory.AgentMode = "full"
		advisory.FounderNotice = "Continuity host is in normal power mode; full agent throughput is allowed."
		advisory.Rule = "environment_advisory.continuity_full"
	default:
		advisory.AgentMode = "observe"
		advisory.FounderNotice = "Host class is unspecified; keep continuity-critical work on the VM until the environment is declared explicitly."
		advisory.Rule = "environment_advisory.unspecified_host"
	}

	return advisory
}

func normalizeHostClass(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "vm", "server":
		return "vm"
	case "laptop", "notebook", "macbook":
		return "laptop"
	case "desktop":
		return "desktop"
	case "container", "podman", "docker":
		return "container"
	default:
		return "unknown"
	}
}

func normalizePowerMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "low", "low_power", "conserve", "battery":
		return "low"
	case "normal", "full", "ac":
		return "normal"
	default:
		return "normal"
	}
}

func continuityRoleForHost(hostClass string) string {
	switch hostClass {
	case "vm":
		return "continuity_host"
	case "laptop", "desktop":
		return "burst_workstation"
	case "container":
		return "encapsulated_runtime"
	default:
		return "unspecified"
	}
}
