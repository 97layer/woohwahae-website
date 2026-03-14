# Codex Job Worker Example

Use this wrapper when you want `layer-osctl job work` to keep watching for
packet-ready jobs and hand each one to Codex CLI.

If you want fewer manual terminal steps, use the native
`go run ./cmd/layer-osctl quickwork` entrypoint. It starts or reuses the daemon
plus workers, then submits the next job for you. `./scripts/work_now.sh` and
`scripts/work_now.command` are thin wrappers around that native command.

## 1. Start or reuse the daemon

```bash
./scripts/start.sh --check
./scripts/start.sh
```

If the daemon is already running, skip the second command.

## 2. Create and dispatch a backend job

```bash
go run ./cmd/layer-osctl job create \
  --kind implement \
  --role implementer \
  --summary "Stabilize the backend worker lane" \
  --source founder.manual \
  --stage compose
```

```bash
go run ./cmd/layer-osctl job list --status queued --limit 1
```

Copy the `job_id`, then dispatch it:

```bash
go run ./cmd/layer-osctl job dispatch --id <job-id>
```

## 3. Run the long-lived Codex worker

```bash
LAYER_OS_WRITE_TOKEN="$LAYER_OS_WRITE_TOKEN" \
go run ./cmd/layer-osctl job work \
  --roles implementer,verifier \
  --command './scripts/job_worker_codex.sh' \
  --poll 30s
```

The worker loop will:

- dispatch queued jobs when needed
- fetch the official `AgentRunPacket`
- save `packet.json`, `prompt.md`, stdout/stderr logs, and `result.json`
- invoke `codex exec` non-interactively
- submit the final `job report`

`job work` exports the full worker env contract to the wrapper, including
`LAYER_OS_REPO_ROOT`, `LAYER_OS_JOB_WORK_DIR`, `LAYER_OS_JOB_KIND`,
`LAYER_OS_JOB_STAGE`, `LAYER_OS_JOB_SOURCE`, `LAYER_OS_ALLOWED_PATHS`,
`LAYER_OS_REPORT_COMMAND`, `LAYER_OS_REPORT_PATH`, `LAYER_OS_REPORT_TOKEN_ENV`,
`LAYER_OS_PROMPT_PATH`, `LAYER_OS_STDOUT_PATH`, and `LAYER_OS_STDERR_PATH`.
The Codex subprocess does not submit the report itself; it only writes one JSON
object to `LAYER_OS_RESULT_PATH`, and `job work` closes the lane.

## Optional knobs

- Set `LAYER_OS_CODEX_MODEL` to force a specific Codex model.
- `./scripts/job_worker_codex.sh` defaults to `workspace-write` for implementer
  lanes and `read-only` for planner/verifier lanes; override with
  `LAYER_OS_CODEX_SANDBOX` only when the packet explicitly allows broader work.
- Set `LAYER_OS_CODEX_MAX_ATTEMPTS` and `LAYER_OS_CODEX_RETRY_BASE_DELAY_SECONDS`
  to soften transient Codex `429 Too Many Requests` / remote compact retries.
- Use `--packet-file /path/to/packet.json --once --dispatch-queued=false` to
  rehearse the wrapper offline without dispatching or reporting to the daemon.
- Add `--once` to `job work` for a single-pass worker.
- Add `--stop-on-failure` while hardening a new wrapper.
- Change `--work-root` if you want artifacts outside `/tmp/layer-os-job-work`.

## Result contract reminder

Your worker must emit one JSON object with these keys:

- `summary`
- `artifacts`
- `verification`
- `open_risks`
- `follow_on`
- `touched_paths`
- `blocked_paths`

`./scripts/job_worker_codex.sh` enforces that contract with a JSON schema before
writing the final result file back to Layer OS.
