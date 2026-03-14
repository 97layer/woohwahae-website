# Operator Wrappers

## Intent

External wrappers may improve ergonomics, packaging, or automation.
They still delegate reads and writes to `layer-osctl` or `/api/layer-os/*`.
A wrapper is an outer operator surface, not a second runtime kernel.

## Related Docs

- `docs/agent-quickstart.md` and `docs/agent-role-seeds.md` explain how agents start, self-assign lanes safely, and choose roles before hitting these wrappers.
- `docs/examples/content_ingest_examples.md` and `docs/examples/telegram_ingest_examples.md` show runnable commands surfaced in this guide.
- `docs/validation.md` maps which checks are automated locally versus deferred to network-enabled verification.

See `docs/legacy_inventory.md` for the baseline keep/port/drop inventory.

### Self-start quick prompt (KR)

이 문서는 wrapper surface만 다룬다.
실제 self-start prompt 문구는 `docs/agent-quickstart.md`와 `docs/agent-role-seeds.md`만 canonical로 유지한다.

## Keep

- One daemon/API canonical write path.
- Explicit auth, actor, and model metadata on requests.
- Thin read surfaces for daemon visibility, status, bootstrap, handoff, capability, and audit access.
- Terminal-first execution as the default operator path; optional MCP tooling may help locally but must not become the authority path or a prerequisite for routine work.
- `layer-osctl audit mcp` and `./scripts/check_codex_mcp.sh` are advisory environment diagnostics only, not readiness gates.

## Token Sweet Spot Defaults

For routine work, keep the operator path thin by default:

1. `layer-osctl session bootstrap` before any broad read
2. `layer-osctl next` or `layer-osctl knowledge` before `handoff`
3. `layer-osctl job packet --id <job>` before repo-wide browsing
4. full `handoff` only when the thin packet cannot explain the lane safely
5. one bounded verification loop before escalating into deeper context

The goal is not maximum compression.
It is the smallest context that still keeps decisions safe and legible.

## Local Seat Sweet Spot

For local founder operation, optimize for a calm seat rather than maximum concurrency:

1. `layer-osd` as the control-tower daemon and log source
2. `8081/admin` only when the cockpit is actively needed
3. one active builder lane for Codex or another agent session
4. one optional manual shell for verification or git

If the machine starts accumulating stale tmux seats, MCP daemons, or repeated
Playwright/NotebookLM helper processes, trim them before debugging product
issues:

```bash
./scripts/trim_local_operator_noise.sh --check
./scripts/trim_local_operator_noise.sh --apply
```

This cleanup should keep `layer-osd` alive while removing duplicated local tool
noise that causes port conflicts, Telegram collision, and cognitive overload.

## Port Carefully

- `LAYER_OS_WRITER_ID` is identity metadata for logging and lease context, not side-channel authority.
- `LAYER_OS_REPO_ROOT` scopes local audit and verification work only; it does not authorize direct state rewrites.
- `session bootstrap --allow-local-fallback` is for read-only fallback, not offline mutation.
- Repeated localhost-blocked read sessions can opt into read-only fallback by setting `LAYER_OS_ALLOW_LOCAL_FALLBACK=1` for commands such as `status`, `handoff`, `review-room`, and `founder summary`.

## Drop

- Wrapper-local state files as SSOT.
- Wrapper-managed lock files or queue folders.
- Direct writes to `.layer-os/*.json`.
- Repackaging POST flows that omit daemon auth.

## Central Control Rules

1. Start local runtime through `./scripts/start.sh`, not ad-hoc env shells.
2. Read central state through `./scripts/control.sh` or official `layer-osctl` / `/api/layer-os/*` surfaces only.
3. Treat `.layer-os/*.json` as runtime-owned evidence; mutate state only through authenticated daemon or CLI flows.

## Wrapper Classes

| Wrapper lane | Reads | Writes | Notes |
|---|---|---|---|
| observer | `daemon status`, `status`, `knowledge`, `handoff`, `session bootstrap`, `review-room`, `writer`, `capabilities`, `audit ...` | none | safe default |
| operator | same reads | `session finish`, `review-room` transitions, control-plane and workflow commands | requires canonical write path |
| daemon shell | `layer-osd` process start, stop, and health checks | daemon lifecycle only | does not own runtime semantics |

## Environment Contract

| Variable | Use |
|---|---|
| `LAYER_OS_BASE_URL` | explicit CLI base URL for `layer-osd` |
| `LAYER_OS_ADDR` | daemon listen address, or CLI shorthand when `LAYER_OS_BASE_URL` is unset |
| `LAYER_OS_DATA_DIR` | daemon runtime data directory |
| `LAYER_OS_REPO_ROOT` | repo root for local audits and default verification workdir |
| `LAYER_OS_ACTOR` | actor stamped into runtime events and request headers |
| `LAYER_OS_WRITER_ID` | writer identity fallback when actor is unset |
| `LAYER_OS_AGENT_ROLE_PROVIDERS` | role-to-provider routing; also seeds the visible provider roster when `LAYER_OS_PROVIDERS` is unset |
| `LAYER_OS_MODELS` / `LAYER_OS_MODEL` | model ids attached to requests and planning evidence |
| `LAYER_OS_WRITE_TOKEN` | required by mutating CLI commands when write auth is enabled |
| `LAYER_OS_AUTH_FILE` | optional override for the write-auth hash file; default storage is a per-runtime file outside `.layer-os/` |

Provider auth is also normalized at the runtime boundary. `claude` reads
`ANTHROPIC_API_KEY`, `openai` reads `OPENAI_API_KEY`, and `gemini` accepts
either `GEMINI_API_KEY` or `GOOGLE_API_KEY`. `/api/layer-os/providers` and the
cockpit snapshot expose redacted provider summaries; the daemon and CLI keep
using the internal auth-ready/readiness truth for council routing without
publishing env-key metadata on read-only HTTP surfaces.

## HTTP and Auth Rules

- Read routes stay plain `GET`.
- When write auth is enabled, mutating API routes send `Authorization: Bearer <token>`.
- Before write auth is enabled, `/api/layer-os/auth` bootstrap is loopback-only; do not rely on remote bootstrap through an exposed edge.
- `layer-osctl` automatically forwards `LAYER_OS_WRITE_TOKEN`, `LAYER_OS_ACTOR`, and `LAYER_OS_MODELS` / `LAYER_OS_MODEL` into request headers.
- Wrappers should forward or set those inputs, not create alternative auth channels.

## Daemon Visibility Surface

The daemon visibility surface is read-only and exposes process health without creating a second authority path.

- `GET /healthz` returns the raw `DaemonStatus` payload.
- `GET /api/layer-os/daemon` returns `{ "daemon": DaemonStatus }`.
- `GET /api/layer-os/cockpit` is the unified dashboard snapshot for admin surfaces: daemon/status, provider readiness, capability registry, Telegram/Threads status, conversation automation, founder summaries, open jobs, and recent events in one read.
- The default cockpit response is summary-first so `8081/admin` can stay responsive. Use `GET /api/layer-os/cockpit?full=1` only when you need the expanded debug dump.
- `layer-osctl daemon status` returns `DaemonStatus`.

| Field | Meaning |
|---|---|
| `service` | fixed daemon service identity; currently `layer-osd` |
| `status` | daemon visibility summary; currently `ok` or `degraded` |
| `address` | daemon bind or advertised address |
| `started_at` | daemon process start time |
| `uptime_seconds` | non-negative process uptime in seconds |
| `memory_health` | memory lane health copied from company state |
| `deploy_health` | deploy lane health copied from company state |
| `degraded_reasons` | read-only explanation list for degraded state; examples include `memory_health=degraded`, `deploy_health=degraded`, and `security_posture=degraded` |
| `architect` | optional architect-loop visibility block with enablement, interval, promote limit, last run, last error, and last promotion counters |

## Operator Interpretation

- `status=ok` means the daemon read surface is responding normally.
- If `degraded_reasons` is non-empty, interpret those reasons before assuming total daemon failure.
- `architect.last_error` means the background context-to-job promotion loop hit a loop-local failure; it does not by itself mean snapshot state is corrupted.
- A wifi transition, proxy change, or daemon-unreachable path can justify `layer-osctl session bootstrap --allow-local-fallback`, but only to recover read context.
- Fallback does not restore write authority; writes remain daemon/API canonical.

## Antigravity Handoff Note

- Observed on Antigravity `1.19.6` / build `1.107.0`: the effective agent sandbox state may live in `~/Library/Application Support/Antigravity/User/globalStorage/state.vscdb`, not only in `settings.json`.
- In that environment, `127.0.0.1` daemon reads may still work while direct reads of `.layer-os/*` files fail under the terminal sandbox. Prefer daemon-backed reads (`status`, `handoff`, `founder summary`, `review-room`, `session bootstrap`) before assuming localhost itself is blocked.
- Use `/tmp`-scoped caches for Go commands inside Antigravity when possible: `GOCACHE=/tmp/gocache` and `GOMODCACHE=/tmp/gomodcache`. The bundled `./scripts/start.sh` and `./scripts/control.sh` now default those caches automatically.
- If Antigravity starts showing generic failures and the local logs mention `can't send message because channel is full`, reload or restart Antigravity before treating the issue as Layer OS state drift.
- Canonical write handoff still goes through `layer-osctl session finish` or `session note`, while agent baton close still goes through `layer-osctl job report --result-file ...` or `runtime.report_path`. Those write surfaces still require `LAYER_OS_WRITE_TOKEN` whenever write auth is enabled.


## Security Posture Baseline

- Treat `layer-osctl audit security --strict` as the canonical read-only posture check.
- Record a `security_review` preflight at least weekly, before release, and before external exposure.
- Keep write auth enabled before any shared or externally reachable daemon exposure.
- If deploy targets exist while write auth is disabled or the latest release is newer than the latest security review, treat the posture as degraded until the gate is re-run.

## External Exposure Gate

- Before any external exposure, require a fresh `security_review` preflight with checks `write_auth_enabled`, `secret_plaintext_surface_minimized`, `edge_tls_required`, and `edge_access_control_required`.
- Treat TLS termination at the edge and an explicit access-control boundary (private network, VPN, identity proxy, or equivalent) as mandatory, not optional hardening.
- If write auth is disabled, or if the latest security review is `hold` / stale / incomplete, keep the daemon posture degraded and do not treat the surface as externally ready.

## Minimal Patterns

### Read-only daemon status wrapper

```bash
LAYER_OS_BASE_URL="${LAYER_OS_BASE_URL:-http://127.0.0.1:17808}" \
go run ./cmd/layer-osctl daemon status
```

### Read-only bootstrap wrapper

```bash
LAYER_OS_BASE_URL="${LAYER_OS_BASE_URL:-http://127.0.0.1:17808}" \
go run ./cmd/layer-osctl session bootstrap --allow-local-fallback
```

### Read-only handoff wrapper

```bash
LAYER_OS_BASE_URL="${LAYER_OS_BASE_URL:-http://127.0.0.1:17808}" \
go run ./cmd/layer-osctl handoff --allow-local-fallback
```

### Authenticated finish wrapper

```bash
LAYER_OS_BASE_URL="${LAYER_OS_BASE_URL:-http://127.0.0.1:17808}" \
LAYER_OS_WRITE_TOKEN="$LAYER_OS_WRITE_TOKEN" \
go run ./cmd/layer-osctl session finish \
  --focus "..." \
  --steps step_1,step_2 \
  --risks risk_1
```

### Agent packet wrapper

```bash
LAYER_OS_BASE_URL="${LAYER_OS_BASE_URL:-http://127.0.0.1:17808}" \
go run ./cmd/layer-osctl job packet --id job_123
```

### Agent job list wrapper

```bash
LAYER_OS_BASE_URL="${LAYER_OS_BASE_URL:-http://127.0.0.1:17808}" \
go run ./cmd/layer-osctl job list --status open --limit 5
```

### Agent report wrapper

```bash
LAYER_OS_BASE_URL="${LAYER_OS_BASE_URL:-http://127.0.0.1:17808}" \
LAYER_OS_WRITE_TOKEN="$LAYER_OS_WRITE_TOKEN" \
go run ./cmd/layer-osctl job report \
  --id job_123 \
  --status succeeded \
  --result-file /tmp/job-report.json
```

### CLI worker loop

Use the built-in worker loop when you want one terminal command to keep watching
for packet-ready jobs, run a local CLI workflow, and close the lane with
`job report`.

The worker command receives these env vars:

- `LAYER_OS_REPO_ROOT`
- `LAYER_OS_JOB_WORK_DIR`
- `LAYER_OS_PACKET_PATH`
- `LAYER_OS_PROMPT_PATH`
- `LAYER_OS_RESULT_PATH`
- `LAYER_OS_STDOUT_PATH`
- `LAYER_OS_STDERR_PATH`
- `LAYER_OS_JOB_ID`
- `LAYER_OS_JOB_ROLE`
- `LAYER_OS_JOB_KIND`
- `LAYER_OS_JOB_SUMMARY`
- `LAYER_OS_JOB_STAGE`
- `LAYER_OS_JOB_SOURCE`
- `LAYER_OS_ALLOWED_PATHS`
- `LAYER_OS_REPORT_COMMAND`
- `LAYER_OS_REPORT_PATH`
- `LAYER_OS_REPORT_TOKEN_ENV`

The worker still writes exactly one JSON object using the canonical seven-key
result contract from `docs/agent-integration.md`. The wrapper may add transport
metadata later, but it should not invent a second human report shape.

Minimal example:

```bash
LAYER_OS_BASE_URL="${LAYER_OS_BASE_URL:-http://127.0.0.1:17808}" \
LAYER_OS_WRITE_TOKEN="$LAYER_OS_WRITE_TOKEN" \
go run ./cmd/layer-osctl job work \
  --roles implementer,verifier \
  --command 'cat >"$LAYER_OS_RESULT_PATH" <<'"'"'JSON'"'"'
{"summary":"Stub worker completed.","artifacts":[],"verification":[],"open_risks":[],"follow_on":[],"touched_paths":[],"blocked_paths":[]}
JSON' \
  --poll 30s
```

For real long-running agent work, replace the stub command with your Codex /
Claude Code / Antigravity CLI wrapper and make that wrapper write one JSON
object to `LAYER_OS_RESULT_PATH`. The worker loop handles the authenticated
`job report`; the subprocess only needs to honor the packet/prompt/result
contract.

For wrapper hardening without touching the daemon, save a packet JSON first and
use `--packet-file <path> --once --dispatch-queued=false`. That runs the same
prompt/result/log flow offline and skips both dispatch and `job report`.

`./scripts/job_worker_codex.sh` defaults Codex to `workspace-write` for
implementer lanes, and `read-only` for planner/verifier lanes. Override that
with `LAYER_OS_CODEX_SANDBOX` only when the packet explicitly justifies it.

### Native quickwork

Use `layer-osctl quickwork` as the primary native entrypoint when you want the
daemon plus long-lived workers without stitching subcommands together yourself.
It now prefers packet-first runtime hints before falling back to manual typing.

```bash
go run ./cmd/layer-osctl quickwork
```

If the runtime already exposes a safe next action through `session bootstrap`,
quickwork can turn that into one bounded job without making the founder restate
the lane by hand. If not, it still falls back to an interactive prompt.

Planner quickwork now also wakes the planner worker automatically instead of
stopping at the default implementer/verifier pair, so founder/admin planner
submissions do not look healthy while the planner lane is still asleep.

Designer quickwork uses the same pattern for experience lanes.

One-shot submit:

```bash
go run ./cmd/layer-osctl quickwork \
  --summary "Stabilize the backend worker lane" \
  --role implementer \
  --allowed-paths cmd/layer-osctl/,scripts/
```

Looped bounded execution:

```bash
go run ./cmd/layer-osctl quickwork \
  --summary "Tighten the current frontend lane" \
  --role designer \
  --loop \
  --exit-on-idle 10m
```

In loop mode quickwork submits one bounded job, waits for terminal state, then
checks `session bootstrap` for the next safe action. It stops when no fresh
action appears before the idle timeout.

If bootstrap has no direct route but knowledge still carries promotable
`open_threads` or `parallel_candidates`, quickwork now asks the daemon to
promote one bounded context job before falling back to manual founder input or
idle exit.

`layer-osctl next` mirrors the same packet-first posture. When no direct route,
agenda item, proposal command, or open job is available, it now falls back to
`layer-osctl job promote --limit 1 --dispatch` if bootstrap still exposes one
promotable context signal.

If one of those context threads is already linked to a proposal candidate,
context promotion leaves it alone. The founder should see one proposal helper
instead of a duplicate planner job for the same planning seam.

When quickwork derives a planner/designer lane from bootstrap rather than a
manual founder summary, it may also auto-attach a bounded `council` payload if
multiple direct providers are live. The daemon mirrors the same posture at
dispatch time: non-manual planner/designer lanes may be auto-upgraded into the
same bounded council even when the original job payload stayed single-provider.
That keeps review/planning loops multi-model without forcing hand-written
`--payload-json`.

Native status / lifecycle helpers:

```bash
go run ./cmd/layer-osctl quickwork --status
go run ./cmd/layer-osctl quickwork --up
go run ./cmd/layer-osctl quickwork --down
```

The admin quickwork runtime surface now mirrors the same backend truth: daemon
reachability, worker states, write-auth readiness, dispatch profiles, and the
current open job slice all come from the canonical Layer OS status + job APIs.
Provider auth/readiness now follows that same rule internally: quickwork
council routing uses the same provider-credential contract instead of
re-deriving Gemini/Claude/OpenAI readiness separately, while public HTTP read
surfaces redact auth-source/env-key metadata.

### Orchestrator script

`./scripts/worker_orchestrator.sh` remains the backend helper that quickwork
delegates to. Use it only when you specifically want the lower-level split
commands.

```bash
./scripts/worker_orchestrator.sh up
```

Submit one backend lane with automatic create + dispatch:

```bash
./scripts/worker_orchestrator.sh submit \
  --summary "Stabilize the backend worker lane" \
  --role implementer \
  --allowed-paths cmd/layer-osctl/,scripts/
```

`submit` also ensures the matching worker lane is running first, so planner
submissions do not rely on a separate manual `up --roles ...` step.

Check or stop the background workers:

```bash
./scripts/worker_orchestrator.sh status
./scripts/worker_orchestrator.sh down
```

`./scripts/work_now.sh` is now a thin wrapper around `layer-osctl quickwork`.
On macOS you can also double-click `scripts/work_now.command`.

### Local daemon wrapper

```bash
./scripts/start.sh --check
```

```bash
./scripts/start.sh
```

Use an isolated runtime lane when you need a clean local environment for verification or cross-check without touching the current `.layer-os` state:

```bash
LAYER_OS_DATA_DIR=.layer-os-dev ./scripts/start.sh
```

`./scripts/start.sh` is the canonical local bootstrap. It loads `LAYER_OS_PROVIDER_ENV_FILE` (default: `${LAYER_OS_DATA_DIR}/providers.env`), imports macOS Keychain secrets when available, defaults Go caches to `/tmp` for sandbox-friendly runs, keeps write-token verification hashes in the default external auth path unless `LAYER_OS_AUTH_FILE` overrides it, fixes the default daemon env contract, validates runtime JSON except `events_archive.json`, soft-skips that validation during `--check` when the shell cannot read runtime files, and falls back to `go run ./cmd/layer-osd` when `bin/layer-osd` is absent.

### Central control wrapper

```bash
./scripts/control.sh
```

```bash
./scripts/control.sh watch --interval 5
```

`./scripts/control.sh` is the human operator wrapper for central read visibility. It does not read `.layer-os/*.json` directly; it defaults Go caches to `/tmp`, prefers `daemon status` plus live `session bootstrap --full`, keeps optional daemon reads best-effort, and falls back to `session bootstrap --allow-local-fallback --full` only when the daemon is unavailable.

The same wrapper also reads the current git worktree and groups staged, unstaged, and untracked source changes into conflict-aware lanes. Use that lane summary as the control-tower dispatch surface for uncommitted work: frontend lanes, backend writer lanes, verifier lanes, support-writer lanes for docs/scripts/governance work, and hold-for-review hot seams are separated so parallel agents do not collide.

The architect loop now recovers stray Gemini artifacts and stray corpus markdown before context-job promotion, marks explicit idle ticks in daemon status when no work is available, and runs `go test ./...` once per repo stamp. Set `LAYER_OS_ARCHITECT_AUTOVERIFY=false` to disable the automatic verification lane, set Gemini envs to `false` to disable Gemini recovery/cleanup, or set `LAYER_OS_ARCHITECT_CORPUS_RECOVERY=false` / `LAYER_OS_ARCHITECT_CORPUS_CLEANUP=false` to disable corpus recovery lanes without changing promotion behavior.

### Read-only drift audit wrapper

```bash
LAYER_OS_REPO_ROOT="${LAYER_OS_REPO_ROOT:-$PWD}" \
go run ./cmd/layer-osctl audit surface --strict
```

### Documentation audit wrapper

```bash
./scripts/doc_audit.sh
```

This offline check validates local markdown doc refs, `Related Docs` blocks, and forbidden legacy-label leaks in active docs. External URLs remain deferred until a network-enabled link check is run.

### Authority drift audit wrapper

```bash
LAYER_OS_REPO_ROOT="${LAYER_OS_REPO_ROOT:-$PWD}" \
go run ./cmd/layer-osctl audit authority --strict
```

### Gemini containment audit wrapper

```bash
LAYER_OS_REPO_ROOT="${LAYER_OS_REPO_ROOT:-$PWD}" \
go run ./cmd/layer-osctl audit gemini --strict
```

### Gemini recovery wrapper

```bash
LAYER_OS_REPO_ROOT="${LAYER_OS_REPO_ROOT:-$PWD}" \
go run ./cmd/layer-osctl ingest gemini --cleanup
```

### Telegram/profiled ingest wrapper

Use canonical interaction profiles to stamp topics/kinds/tags without ad-hoc refs:

```bash
LAYER_OS_WRITE_TOKEN="$LAYER_OS_WRITE_TOKEN" \
go run ./cmd/layer-osctl ingest telegram \
  --interaction inbound_conversation \
  --message-id 12345 \
  --chat founder_chat \
  --excerpt "Founder pinged about deploy gate"
```

### Content/profiled ingest wrapper

Personal archive and crawler captures can pick a profile to set kinds/tags/dedupe hints:

```bash
LAYER_OS_WRITE_TOKEN="$LAYER_OS_WRITE_TOKEN" \
go run ./cmd/layer-osctl ingest content \
  --channel personal_db \
  --profile outline \
  --title "Q2 release outline" \
  --excerpt "steps and owners" \
  --doc-id outline_q2_release
```

```bash
LAYER_OS_WRITE_TOKEN="$LAYER_OS_WRITE_TOKEN" \
go run ./cmd/layer-osctl ingest content \
  --channel crawler \
  --profile article \
  --url "https://example.com/story?utm_source=twitter" \
  --title "OS adoption case study"
```

### RSS/Atom sensor ingest wrapper

Use this when an external feed should become canonical observations and open the same `source_intake` lane the founder uses manually:

```bash
LAYER_OS_WRITE_TOKEN="$LAYER_OS_WRITE_TOKEN" \
go run ./cmd/layer-osctl ingest rss \
  --feed "https://example.com/feed.xml" \
  --limit 3 \
  --tags signals,auto_ingest \
  --refs route:97layer,domain:brand
```

For 24/7 VM intake, set `LAYER_OS_SENSOR_RSS_FEEDS` on the daemon and optionally enable `LAYER_OS_SENSOR_AUTO_ROUTE=1` only when feed items carry an explicit single `route:` hint. If those items should open the canonical Threads prep lane too, add `LAYER_OS_SENSOR_AUTO_PREP=1`. The daemon now also stamps each `source_intake` item with a lightweight `priority_score` and `disposition` (`observe`, `review`, `prep`) so founder surfaces can separate quiet corpus fodder from route review or draft-ready material. The default VM posture stays conservative: collect first, auto-route only on clearly tagged lanes, auto-prep only when you explicitly opt in.

More runnable samples: see `docs/examples/telegram_ingest_examples.md` and `docs/examples/content_ingest_examples.md`.

### Corpus contamination audit wrapper

```bash
LAYER_OS_REPO_ROOT="${LAYER_OS_REPO_ROOT:-$PWD}" \
go run ./cmd/layer-osctl audit corpus --strict
```

### Corpus recovery wrapper

```bash
LAYER_OS_REPO_ROOT="${LAYER_OS_REPO_ROOT:-$PWD}" \
go run ./cmd/layer-osctl ingest corpus --cleanup
```

## Guardrails

1. Wrap the canonical CLI or API; do not bypass it.
2. Keep wrapper output disposable; runtime truth stays inside daemon-managed contracts.
3. Treat `writer` as a diagnostic surface, not as a second lock manager.
4. Use review-room transitions through CLI or API only.
5. If a wrapper needs richer state, request `--full` bootstrap instead of reading runtime files directly.
