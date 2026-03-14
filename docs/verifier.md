# Verifier Audit

## Intent

Verifier support stays read-first and evidence-first.
This lane observes drift and classifies risk without becoming a hidden writer.

See `docs/legacy_inventory.md` for the baseline keep/port/drop inventory.
See `docs/validation.md` for the current automated-vs-manual verification map.

## Keep

- Thin bootstrap for current state.
- Structured audits for daemon visibility, structure, contracts, residue, Gemini containment, corpus contamination, surface, and review-room integrity.
- Structured docs audit coverage for local markdown references and `Related Docs` drift.
- Daemon status, review-room, handoff, and knowledge as read surfaces for evidence gathering.

## Port Carefully

- `layer-osctl verify run` exists, but it records a `VerificationRecord`; treat it as a writer command, not the default verifier read flow.
- When audits should become agenda, let the daemon or operator promote them through canonical review-room transitions.
- Use `session bootstrap --full` only when the thin packet is insufficient.

## Drop

- Direct edits to `.layer-os/review_room.json` or other runtime files.
- Freeform markdown audit ledgers as authority.
- Wrapper-local drift status that disagrees with daemon or API results.

## Daemon Visibility

Use the daemon visibility surface before any fallback decision.

- `layer-osctl daemon status` returns `DaemonStatus`.
- `GET /healthz` returns the same `DaemonStatus` shape directly.
- `GET /api/layer-os/daemon` returns `{ "daemon": DaemonStatus }`.

`DaemonStatus` exposes `service`, `status`, `address`, `started_at`, `uptime_seconds`, `memory_health`, `deploy_health`, `degraded_reasons`, and optional `architect` loop evidence as read-only evidence.

## Read-Only Check Procedure

1. Run `go run ./cmd/layer-osctl daemon status` and record `service`, `status`, `address`, `started_at`, `uptime_seconds`, `memory_health`, `deploy_health`, `degraded_reasons`, and when present the `architect` loop fields.
2. Check `curl -sS "${LAYER_OS_BASE_URL:-http://127.0.0.1:17808}/healthz"` to confirm the daemon visibility surface is reachable without wrapper-local state.
3. If the daemon is unreachable after a wifi or network-path change, run `go run ./cmd/layer-osctl session bootstrap --allow-local-fallback` only to recover read context.
4. Record whether the latest `security_review` preflight includes `write_auth_enabled`, `secret_plaintext_surface_minimized`, `edge_tls_required`, and `edge_access_control_required` before calling posture healthy.
5. Do not rewrite `.layer-os/review_room.json` or any other runtime file during verification.

Escalate to `--full` only when review-room summary, handoff counts, or capabilities are needed in the same read packet.

## Read-Only Audit Commands

| Command | Purpose | Mutates runtime |
|---|---|---|
| `go run ./cmd/layer-osctl daemon status` | daemon visibility and degraded reasons | no |
| `curl -sS "${LAYER_OS_BASE_URL:-http://127.0.0.1:17808}/healthz"` | raw daemon visibility payload | no |
| `go run ./cmd/layer-osctl status` | current company state | no |
| `go run ./cmd/layer-osctl knowledge` | compact working lane | no |
| `go run ./cmd/layer-osctl handoff` | baton and counts | no |
| `go run ./cmd/layer-osctl review-room` | agenda and summary | no |
| `go run ./cmd/layer-osctl writer` | lease visibility | no |
| `go run ./cmd/layer-osctl audit structure` | folder budget drift | no |
| `go run ./cmd/layer-osctl audit contracts` | contract set drift | no |
| `go run ./cmd/layer-osctl audit residue [--strict]` | forbidden residue scan | no |
| `go run ./cmd/layer-osctl audit gemini [--strict]` | Gemini containment policy + stray artifact scan | no |
| `go run ./cmd/layer-osctl audit corpus [--strict]` | markdown contamination in repo/runtime corpus paths | no |
| `go run ./cmd/layer-osctl audit authority [--strict]` | workspace authority boundary and legacy-path drift | no |
| `go run ./cmd/layer-osctl audit surface` | CLI and API surface drift | no |
| `go run ./cmd/layer-osctl audit security [--strict]` | write-auth, runtime secret plaintext, external-exposure gate, and security-review drift | no |
| `go run ./cmd/layer-osctl audit review-room` | seal and integrity drift | no |
| `go run ./cmd/layer-osctl audit docs [--strict]` | local markdown doc refs, `Related Docs` blocks, and active-doc legacy-label drift | no |
| `go run ./cmd/layer-osctl verify list` | recorded verification history | no |

External URLs are not validated by the offline docs audit. Run a network-enabled link checker separately and report it as additional evidence, not as part of the offline pass.

## Writer-Bound Commands

Keep these out of read-only verifier wrappers unless the lane is explicitly promoted to authenticated writer:

- `go run ./cmd/layer-osctl verify run`
- `go run ./cmd/layer-osctl review-room add`
- `go run ./cmd/layer-osctl review-room accept`
- `go run ./cmd/layer-osctl review-room defer`
- `go run ./cmd/layer-osctl review-room resolve`
- `go run ./cmd/layer-osctl ingest gemini --cleanup`
- `go run ./cmd/layer-osctl ingest corpus --cleanup`

## Reporting Pattern

1. Name the drift class: `structure`, `contracts`, `residue`, `gemini`, `surface`, or `review-room`.
2. Cite the exact command or API surface used.
3. Separate confirmed drift from suggested remediation.
4. Hand proposed fixes back through proposal, work, or review-room lanes instead of mutating runtime state directly.
