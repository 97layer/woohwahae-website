package api

import (
	"net/http"
	"path/filepath"

	"layer-os/internal/runtime"
)

type auditRouteSpec struct {
	path string
	run  func(root string) any
}

var auditRouteSpecs = []auditRouteSpec{
	{path: "/api/layer-os/audit/structure", run: func(root string) any { return runtime.AuditStructure(root) }},
	{path: "/api/layer-os/audit/contracts", run: func(root string) any { return runtime.AuditContracts(root) }},
	{path: "/api/layer-os/audit/residue", run: func(root string) any { return runtime.AuditResidue(root) }},
	{path: "/api/layer-os/audit/gemini", run: func(root string) any { return runtime.AuditGemini(root) }},
	{path: "/api/layer-os/audit/authority", run: func(root string) any { return runtime.AuditAuthority(root) }},
	{path: "/api/layer-os/audit/surface", run: func(root string) any { return runtime.AuditSurface(root) }},
	{path: "/api/layer-os/audit/review-room", run: func(root string) any { return runtime.AuditReviewRoomIntegrity(filepath.Join(root, ".layer-os")) }},
	{path: "/api/layer-os/audit/runtime-data", run: func(root string) any { return runtime.AuditRuntimeData(root, filepath.Join(root, ".layer-os")) }},
	{path: "/api/layer-os/audit/mcp", run: func(root string) any { return runtime.AuditMCP(root) }},
	{path: "/api/layer-os/audit/docs", run: func(root string) any { return runtime.AuditDocumentation(root) }},
}

func registerAuditRoutes(mux *http.ServeMux) {
	for _, spec := range auditRouteSpecs {
		spec := spec
		mux.HandleFunc(spec.path, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				methodNotAllowed(w)
				return
			}
			writeJSON(w, http.StatusOK, spec.run(repoRoot()))
		})
	}
}
