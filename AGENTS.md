# Layer OS — Agent Entry Point

This is the thin bootstrap document for the new Layer OS workspace.
It is an entrypoint, not a rule encyclopedia.

## Read Order

1. `constitution/charter.md`
2. `constitution/orchestra.md`
3. `docs/grammar.md`
4. `docs/architecture.md`
5. `docs/agent-integration.md`

## Path Map

- Current Layer OS constitution lives in `constitution/` under this workspace root: `/Users/97layer/layer OS/constitution`.
- Legacy `97layerOS` reference philosophy does **not** live under `constitution/`; read `/Users/97layer/97layerOS/directives/` there instead.
- When shelling absolute paths, quote `/Users/97layer/layer OS/...` because the workspace root contains a space.

## Legacy Boundary

- Legacy material under `/Users/97layer/97layerOS` is reference-only in this workspace.
- Do not use legacy docs to justify file paths, write targets, schemas, or runtime authority for `/Users/97layer/layer OS`.
- For file creation, edits, contracts, and validation in this workspace, authority must come from this workspace's `AGENTS.md`, `constitution/`, `docs/`, and `contracts/`.
- If legacy guidance conflicts with current Layer OS docs, current workspace rules win.

## Validate

Use this as the default pre/post change validation command:

```bash
go test ./...
```

## Runtime Roles

- Canonical runtime roles are `planner`, `implementer`, `verifier`, and `designer`.
- `Codex`, `Claude`, and `Gemini` are provider/interface names, not fixed job ownership.
- Route work by packet `prompting`, dispatch profiles, and runtime state rather than vendor name.

## Hard Rules

- Do not reintroduce legacy labels: `SA`, `CE`, `AD`, `RALPH`, `CD`.
- Do not add a new root folder before proving the change cannot fit inside `constitution`, `contracts`, `docs`, `skills`, `cmd`, or `internal`.
- If you change `contracts/*.schema.json`, update `internal/runtime` types and validation in the same change.
- Keep `internal/` first-level packages limited to `api` and `runtime`.
- When write auth is enabled, API `POST` routes require a bearer token and CLI write commands require `LAYER_OS_WRITE_TOKEN`.
- External agents (Python or otherwise) receive `AgentRunPacket` and report back via `/api/layer-os/jobs/report`. Follow `docs/agent-integration.md`; do not invent side-channel report paths.
- `review_room.json` is runtime state, not an agent scratchpad. Do not rewrite `.layer-os/review_room.json` directly; all meeting-state changes must go through `layer-osd` review-room transitions. The verifier lane is read-only for meeting-state unless explicitly authorized.

## Allowed Write Paths

Agents may only create new files under:
- `internal/`, `cmd/`, `contracts/`, `constitution/`, `docs/`, `scripts/`, `skills/` — source changes

Do NOT create or hand-author repository `knowledge/` files. Canonical corpus entries are runtime-owned JSON under `.layer-os/knowledge/corpus/` and must be emitted through official daemon/service flows.
Do NOT place any file directly at the repository root.
If content does not fit the above paths, raise a review-room item instead of creating a new path.

## Session Protocol

- Default session entry is `layer-osctl session bootstrap`; use `--full` only when the thin packet is insufficient.
- If `layer-osd` is unreachable and only read context is needed, use `layer-osctl session bootstrap --allow-local-fallback`.
- Default session close is `layer-osctl session finish --focus <text> ...` so memory and a `session.finished` event stay aligned.
- Prefer `knowledge` for low-token startup and pull full `handoff` only when the task proves it needs wider state.
