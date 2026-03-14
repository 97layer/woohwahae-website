package runtime

import "time"

func (s *Service) BindDaemonRuntime(startedAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.daemonStartedAt = startedAt.UTC()
}

func (s *Service) DaemonFreshnessAudit() DaemonFreshnessAudit {
	s.mu.Lock()
	defer s.mu.Unlock()
	return AuditDaemonFreshness(s.repoRoot, s.daemonStartedAt)
}
