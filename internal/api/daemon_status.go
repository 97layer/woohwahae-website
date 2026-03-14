package api

import (
	"strings"
	"time"

	"layer-os/internal/runtime"
)

type DaemonRuntimeInfo struct {
	Address         string
	StartedAt       time.Time
	ArchitectStatus func() *runtime.DaemonArchitectStatus
}

func normalizeDaemonRuntimeInfo(info DaemonRuntimeInfo) DaemonRuntimeInfo {
	if strings.TrimSpace(info.Address) == "" {
		info.Address = runtime.DefaultDaemonAddr
	}
	if info.StartedAt.IsZero() {
		info.StartedAt = time.Now().UTC()
	}
	return info
}

func daemonStatus(service *runtime.Service, info DaemonRuntimeInfo) runtime.DaemonStatus {
	state := service.Status()
	reasons := []string{}
	var architect *runtime.DaemonArchitectStatus
	if strings.TrimSpace(state.MemoryHealth) == "degraded" {
		reasons = append(reasons, "memory_health=degraded")
	}
	if strings.TrimSpace(state.DeployHealth) == "degraded" {
		reasons = append(reasons, "deploy_health=degraded")
	}
	security := service.SecurityAudit()
	if security.Status != "ok" {
		reasons = append(reasons, "security_posture=degraded")
	}
	freshness := service.DaemonFreshnessAudit()
	if freshness.Status != "ok" {
		reasons = append(reasons, "daemon_source_freshness=degraded")
	}
	if info.ArchitectStatus != nil {
		architect = info.ArchitectStatus()
		if architect != nil && architect.Enabled && architect.LastError != nil {
			reasons = append(reasons, "architect_loop=degraded")
		}
	}
	status := "ok"
	if len(reasons) > 0 {
		status = "degraded"
	}
	uptimeSeconds := int64(time.Since(info.StartedAt).Seconds())
	if uptimeSeconds < 0 {
		uptimeSeconds = 0
	}
	item := runtime.DaemonStatus{
		Service:         "layer-osd",
		Status:          status,
		Address:         info.Address,
		StartedAt:       info.StartedAt,
		UptimeSeconds:   uptimeSeconds,
		MemoryHealth:    state.MemoryHealth,
		DeployHealth:    state.DeployHealth,
		DegradedReasons: reasons,
		Architect:       architect,
	}
	return item
}
