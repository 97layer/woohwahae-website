package main

import (
	"fmt"

	"layer-os/internal/runtime"
)

func auditStructure(root string) runtime.StructureAudit {
	return runtime.AuditStructure(root)
}

func auditContracts(root string) runtime.ContractAudit {
	return runtime.AuditContracts(root)
}

func auditResidue(root string) runtime.ResidueAudit {
	return runtime.AuditResidue(root)
}

func auditGemini(root string) runtime.GeminiAudit {
	return runtime.AuditGemini(root)
}

func auditCorpus(root string, dataDir string) runtime.CorpusAudit {
	return runtime.AuditCorpus(root, dataDir)
}

func auditAuthority(root string) runtime.AuthorityAudit {
	return runtime.AuditAuthority(root)
}

func auditSurface(root string) runtime.SurfaceAudit {
	return runtime.AuditSurface(root)
}

func auditSecurity(root string, dataDir string) runtime.SecurityAudit {
	return runtime.AuditSecurity(root, dataDir)
}

func auditReviewRoom(dataDir string) runtime.ReviewRoomIntegrityAudit {
	return runtime.AuditReviewRoomIntegrity(dataDir)
}

func auditDocs(root string) runtime.DocumentationAudit {
	return runtime.AuditDocumentation(root)
}

func selectAudit(kind string, root string, dataDir string) (any, bool, error) {
	switch kind {
	case "structure":
		audit := auditStructure(root)
		return audit, len(audit.Issues) > 0, nil
	case "contracts":
		audit := auditContracts(root)
		return audit, len(audit.Issues) > 0, nil
	case "residue":
		audit := auditResidue(root)
		return audit, len(audit.Matches) > 0, nil
	case "gemini":
		audit := auditGemini(root)
		return audit, audit.Status != "ok", nil
	case "corpus":
		audit := auditCorpus(root, dataDir)
		return audit, audit.Status != "ok", nil
	case "authority":
		audit := auditAuthority(root)
		return audit, audit.Status != "ok", nil
	case "surface":
		audit := auditSurface(root)
		return audit, len(audit.Issues) > 0, nil
	case "security":
		audit := auditSecurity(root, dataDir)
		return audit, audit.Status != "ok", nil
	case "review-room":
		audit := auditReviewRoom(dataDir)
		return audit, len(audit.Issues) > 0, nil
	case "runtime-data":
		audit := auditRuntimeData(root, dataDir)
		return audit, audit.Status != "ok", nil
	case "mcp":
		audit := auditMCP(root)
		return audit, audit.Status != "ok", nil
	case "docs":
		audit := auditDocs(root)
		return audit, audit.Status != "ok", nil
	default:
		return nil, false, fmt.Errorf("usage: layer-osctl audit <structure|contracts|residue|gemini|corpus|authority|surface|security|review-room|runtime-data|mcp|docs>")
	}
}

func auditMCP(root string) runtime.MCPAudit {
	return runtime.AuditMCP(root)
}

func auditRuntimeData(root string, dataDir string) runtime.RuntimeDataAudit {
	return runtime.AuditRuntimeData(root, dataDir)
}
