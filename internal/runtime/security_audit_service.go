package runtime

func (s *Service) SecurityAudit() SecurityAudit {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.securityAuditLocked()
}

func (s *Service) securityAuditLocked() SecurityAudit {
	return evaluateSecurityAuditWithFindings(
		s.auth.status(),
		s.reviewRoom.current(),
		s.preflight.list(),
		s.release.list(),
		s.target.list(),
		s.event.list(),
		scanRuntimeSecretFiles(s.disk.baseDir),
		zeroSafeNow(),
	)
}
