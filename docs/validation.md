# Validation

## Intent

Layer OS keeps documentation and runtime claims honest by pairing source-of-truth docs with repo-native audits and smoke tests.
This document makes the current verification surface explicit: what is automated now, what still needs a network-enabled environment, and where runtime visibility is expected to show up.

## Related Docs

- `docs/operator.md` defines the operator wrapper surface that runs these checks.
- `docs/verifier.md` defines the read-only audit lane and reporting posture.
- `docs/architecture.md` explains why review-room, handoff, and knowledge are the runtime evidence surfaces.
- `docs/agent-integration.md` defines the external-agent report path that feeds the same review-room and event loop.

## Automated Now

- `go test ./...` is the default pre/post change gate.
- `layer-osctl audit docs --strict` checks local markdown doc refs, `Related Docs` block shape, and forbidden legacy-label leaks in active docs.
- `scripts/doc_audit.sh` wraps that docs audit with sandbox-friendly Go caches for local operator use.
- Runtime smoke coverage now includes proposal promotion, failed agent-job escalation, and visibility in `KnowledgePacket`, `HandoffPacket`, and `ReviewRoomSummary`.

## Evidence Surfaces

- `ReviewRoomSummary` shows whether failures or unresolved tension are still open.
- `KnowledgePacket` shows whether the active lane is now blocked by review pressure.
- `HandoffPacket` shows counts and founder-priority interpretation for the same state.
- Existing execution-loop tests continue to cover proposal -> verification -> approval -> release/deploy flows and approval rejection behavior.

## Still Manual or Network-Dependent

- External URLs are recorded by the docs audit as deferred, not verified.
- Full external link validation still needs a network-enabled environment and a tool such as `markdown-link-check`.
- Any check that depends on remote packages, provider endpoints, or shared runtime infrastructure should be reported as `unverified` when the environment cannot support it.

## Recommended Command Sequence

```bash
./scripts/doc_audit.sh
```

```bash
go test ./...
```

```bash
markdown-link-check docs/**/*.md constitution/**/*.md
```

Use the third command only in a network-enabled environment and archive the output as evidence rather than treating a skipped run as a pass.
