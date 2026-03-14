# Legacy Inventory — `97layerOS` to `Layer OS`

This document is a boundary note, not an archaeology dump.
The old `97layerOS` tree is no longer part of current runtime operation.
Current runtime authority stays inside this Go workspace.

## Live Boundary

No external legacy path is required for current operation.
What remains is lineage, not a live dependency.

- previous website donor has been absorbed into `docs/brand-home/content/` and `docs/brand-home/public/assets/media/brand/`
- `the_origin` and `sage_architect` survive only as distilled constitutional source, documented through `constitution/lineage.md`

## What Was Ported

These meanings survive, but only through current contracts and runtime code.

| Legacy meaning | Current Layer OS owner |
|---|---|
| session bootstrap / finish | `layer-osctl session bootstrap`, `layer-osctl session finish`, `SessionBootstrapPacket` |
| explicit runtime baton | `AgentRunPacket.runtime`, `job packet -> job report` |
| raw signal vs derived learning split | `ObservationRecord` vs `.layer-os/knowledge/corpus/` |
| repeated evidence before promotion | `observation_links`, `proposal_candidates`, `open_threads`, `review_room` |
| single-writer caution on hot seams | current single-writer policy, `WriteLease`, review-room governance |

## What Is Intentionally Dead

Do not reintroduce these as runtime authority:

| Retired residue | Why it stays dead |
|---|---|
| legacy role labels (`SA`, `CE`, `AD`, `CD`, `SAGE_ARCHITECT`) | old theatre, not current law |
| `.infra/queue/` and file-ledger dispatch | splits truth across side channels |
| legacy markdown/state ledgers as SSOT | not atomic, not enforceable |
| Python/Telegram secretary runtime | replaced by Go runtime + canonical observations |
| `paperclip/` local control plane donor | UI donor only, not a second admin product |
| required OMX/MCP posture | optional tooling only, never kernel authority |

## Remaining Reference Paths

### Legacy repo

- no external legacy repo path is required now
- old website meaning lives in the absorbed brand snapshot under `docs/brand-home/content/`
- `the_origin` and `sage_architect` now matter only through the constitution port map

### Current repo

- canonical admin is `docs/brand-home` on `8081/admin`
- canonical kernel is `cmd/`, `internal/`, `contracts/`
- repo-local social style snapshot is `docs/brand-home/content/social-style-source.json`

## Current Cutover Rules

1. If a legacy idea is still needed, translate it into `contract -> runtime -> test`.
2. If it is only an implementation habit, leave it buried.
3. Do not bulk-port legacy folders.
4. Do not let donor surfaces become runtime authority.

## Supporting Ports

- Session semantics: `docs/legacy_session.md`
- External wrapper/operator boundary: `docs/operator.md`
- Verifier read-only audit lane: `docs/verifier.md`
- `docs/examples/job_worker_codex.md`
- `docs/examples/telegram_ingest_examples.md`

### Compatibility / Alias

- `docs/agent_integration.md` is an alias only; keep it thin and non-authoritative.
- `.gemini/GEMINI.md` is agent-local compatibility guidance, not repo authority.

### Archived Historical Notes

These are already outside the default read path and live under `docs/archive/`.

- `docs/archive/agent-quickstart-web-design.md`
- `docs/archive/content_flywheel_roadmap.md`
- `docs/archive/expansion.md`
- `docs/archive/founder-aiops-note.md`
- `docs/archive/home-shell-migration.md`
- `docs/archive/parallel_dispatch_5_agents.md`
- `docs/archive/parallel_milestone.md`
- `docs/archive/social_account_operating_model.md`

### Cleanup Rule

When markdown count grows again:

1. Exclude vendor/runtime/generated docs before judging sprawl.
2. Keep only authority, operator, integration, and example docs in the active read path.
3. Move historical notes into `docs/archive/` before deciding whether to delete them.
4. Do not let batch notes or personal strategy notes silently become runtime authority.
