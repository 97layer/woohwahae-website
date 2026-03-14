# Grammar

## Naming Family

Layer OS uses one naming family across folders, packages, files, routes, and contracts.

## Related Docs

- External wrapper/operator boundary: `docs/operator.md`
- Verifier read-only audit lane: `docs/verifier.md`

## Folder Rules

- Top-level folder names are lowercase nouns.
- Internal package names are lowercase singular nouns.
- No `shared`, `common`, `misc`, `helpers`, or `utils` dumping grounds.
- Root folders are budgeted; adding one requires a responsibility that cannot be expressed by `constitution`, `contracts`, `docs`, `skills`, `cmd`, or `internal`.
- Inside `internal/`, only `api/` and `runtime/` are first-level packages.
- Prefer a new file before a new folder.
- `.layer-os/` is reserved for runtime state only.

## File Rules

- Prefer short role names: `types.go`, `store.go`, `service.go`, `router.go`, `config.go`.
- If a file must specialize, use `<domain>_<role>.go`.
- Avoid duplicated prefixes inside a package.
- Skill names prefer short verbs or role nouns: `orchestrate`, `verify`, `browser`, `capture`.
- Command names prefer one daemon noun and one control noun: `layer-osd`, `layer-osctl`.
- Command verbs stay short and operator-facing: `status`, `handoff`, `list`, `create`, `resolve`, `run`, `get`, `set`, `evaluate`, `audit`.

Good:
- `internal/runtime/state.go`
- `internal/api/router.go`

Bad:
- `internal/runtime/runtime_state_store_service.go`
- `internal/api/layer_os_http_router_handlers.go`

## Route Rules

- Health routes stay thin.
- Business routes live under `/api/layer-os/*`.
- Route names use the same nouns as contracts.

## Contract Rules

- Contract names are singular PascalCase nouns.
- JSON field names are lowercase snake_case.
- Enums use small vocabulary families.

## Stage Vocabulary

- `discover`
- `compose`
- `experience`
- `verify`
- `release`

No legacy specialist labels appear in public contracts.

## Surface Vocabulary

- `cockpit`
- `telegram`
- `api`

## Event Rules

- Events use nouns and verbs that match contracts.
- One event envelope shape, many event kinds.

## Interpretation Rules

- A new agent should infer where a change belongs within seconds.
- The same idea must not appear under multiple folder names.
- If two folders could plausibly own the same change, the structure is wrong.
- Active skills are the first stop; imported skills are deeper references.
