package runtime

func canonicalPolicyMode(value string) string {
	switch value {
	case "dual":
		return "single"
	default:
		return value
	}
}

func normalizePolicyDecision(item PolicyDecision) PolicyDecision {
	originalMode := item.Mode
	item.Mode = canonicalPolicyMode(item.Mode)
	if originalMode == "dual" {
		item.Reasons = appendUniqueString(item.Reasons, "legacy dual mode normalized to single review lane")
	}
	return item
}

func normalizeExecuteRun(item ExecuteRun) ExecuteRun {
	item.Mode = canonicalPolicyMode(item.Mode)
	return item
}

func normalizePolicyDecisions(items []PolicyDecision) []PolicyDecision {
	if items == nil {
		return []PolicyDecision{}
	}
	out := make([]PolicyDecision, len(items))
	for i, item := range items {
		out[i] = normalizePolicyDecision(item)
	}
	return out
}

func normalizeExecuteRuns(items []ExecuteRun) []ExecuteRun {
	if items == nil {
		return []ExecuteRun{}
	}
	out := make([]ExecuteRun, len(items))
	for i, item := range items {
		out[i] = normalizeExecuteRun(item)
	}
	return out
}

func normalizeSnapshot(packet SnapshotPacket) SnapshotPacket {
	if packet.Branches == nil {
		packet.Branches = []Branch{}
	}
	if packet.Proposals == nil {
		packet.Proposals = []ProposalItem{}
	}
	if packet.AgentJobs == nil {
		packet.AgentJobs = []AgentJob{}
	}
	if packet.Observations == nil {
		packet.Observations = []ObservationRecord{}
	}
	if packet.ContinuityRecords == nil {
		packet.ContinuityRecords = []ContinuityRecord{}
	}
	packet.ContinuityRecords = normalizeContinuityRecordsAt(packet.ContinuityRecords, zeroSafeNow())
	if packet.SessionCheckpoint != nil {
		item := normalizeSessionCheckpoint(*packet.SessionCheckpoint)
		packet.SessionCheckpoint = &item
	}
	packet.ReviewRoom = normalizeReviewRoom(packet.ReviewRoom)
	packet.Policies = normalizePolicyDecisions(packet.Policies)
	packet.Executes = normalizeExecuteRuns(packet.Executes)
	return packet
}

func appendUniqueString(items []string, value string) []string {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}

func containsExactString(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
