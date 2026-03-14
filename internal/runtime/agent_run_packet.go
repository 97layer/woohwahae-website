package runtime

import (
	"errors"
	"path/filepath"
	"strings"
)

func (s *Service) AgentRunPacket(jobID string) (AgentRunPacket, error) {
	return s.agentRunPacketForTransport(jobID, "job_packet")
}

func (s *Service) agentRunPacketForTransport(jobID string, transport string) (AgentRunPacket, error) {
	source, job, proposal, auth, err := s.agentRunPacketContext(jobID)
	if err != nil {
		return AgentRunPacket{}, err
	}
	return s.agentRunPacketWithTransport(source, job, proposal, auth, transport), nil
}

func (s *Service) agentRunPacketContext(jobID string) (string, AgentJob, *ProposalItem, AuthStatus, error) {
	trimmed := strings.TrimSpace(jobID)
	if trimmed == "" {
		return "", AgentJob{}, nil, AuthStatus{}, errors.New("job_id is required")
	}

	s.mu.Lock()
	job, ok := s.agentJob.get(trimmed)
	if !ok {
		s.mu.Unlock()
		return "", AgentJob{}, nil, AuthStatus{}, errors.New("job_id not found")
	}

	var proposal *ProposalItem
	if job.Ref != nil {
		if item, ok := s.proposal.get(strings.TrimSpace(*job.Ref)); ok {
			copy := item
			proposal = &copy
		}
	}

	source := filepath.ToSlash(s.disk.agentJobsPath())
	s.mu.Unlock()

	return source, job, proposal, s.AuthStatus(), nil
}

func (s *Service) agentRunPacketWithTransport(source string, job AgentJob, proposal *ProposalItem, auth AuthStatus, transport string) AgentRunPacket {
	knowledge := s.Knowledge()
	prompting := defaultJobPrompting(job, knowledge)
	handoff, summary := agentPacketHandoff(job.Role, s)
	return AgentRunPacket{
		GeneratedAt:    zeroSafeNow(),
		Source:         source,
		Job:            job,
		Runtime:        newAgentRuntimeContract(job.JobID, auth, transport),
		Knowledge:      knowledge,
		Prompting:      &prompting,
		Handoff:        handoff,
		HandoffSummary: summary,
		Proposal:       proposal,
	}
}

func newAgentRuntimeContract(jobID string, auth AuthStatus, transport string) AgentRuntimeContract {
	if strings.TrimSpace(transport) == "" {
		transport = "job_packet"
	}
	contract := AgentRuntimeContract{
		SourceOfTruth:     "daemon_api",
		DispatchTransport: transport,
		ReportPath:        "/api/layer-os/jobs/report",
		ReportCommand:     "layer-osctl job report --id " + jobID + " --status <succeeded|failed|canceled> [--notes a,b] [--result-file <path-to-json>]",
		TerminalStatuses:  []string{"succeeded", "failed", "canceled"},
		WriteAuthRequired: auth.WriteAuthEnabled,
	}
	if auth.WriteAuthEnabled {
		tokenEnv := "LAYER_OS_WRITE_TOKEN"
		contract.WriteTokenEnv = &tokenEnv
	}
	return contract
}
