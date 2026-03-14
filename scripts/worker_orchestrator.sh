#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
STATE_DIR="${LAYER_OS_WORKER_STATE_DIR:-/tmp/layer-os-worker-orchestrator}"
mkdir -p "${STATE_DIR}"

export GOCACHE="${GOCACHE:-/tmp/gocache}"
export GOMODCACHE="${GOMODCACHE:-/tmp/gomodcache}"
mkdir -p "${GOCACHE}" "${GOMODCACHE}"

# Load write token from Keychain if not already set
if [[ -z "${LAYER_OS_WRITE_TOKEN:-}" ]] && command -v security >/dev/null 2>&1; then
  _tok="$(security find-generic-password -s "LAYER_OS_WRITE_TOKEN" -w 2>/dev/null || true)"
  if [[ -n "${_tok}" ]]; then
    export LAYER_OS_WRITE_TOKEN="${_tok}"
  fi
  unset _tok
fi

ROLE_ARGS_DEFAULT="implementer,verifier"
POLL_DEFAULT="${LAYER_OS_WORKER_POLL:-30s}"
IDLE_DEFAULT="${LAYER_OS_WORKER_IDLE_EXIT_AFTER:-2h}"

usage() {
  cat >&2 <<'USAGE'
usage:
  worker_orchestrator.sh up [--roles implementer,verifier] [--poll 30s] [--idle-exit-after 2h]
  worker_orchestrator.sh status
  worker_orchestrator.sh down
  worker_orchestrator.sh submit --summary <text> [--role implementer|verifier|planner|designer] [--kind <kind>] [--stage <stage>] [--source <source>] [--allowed-paths a,b] [--payload-json <json>]

Intent:
  - `up` starts or reuses the daemon, then launches long-lived worker loops in the background.
  - `submit` creates and dispatches one job with optional allowed_paths so you do not have to type create+dispatch manually.
USAGE
}

note() {
  printf '%s\n' "$1"
}

warn() {
  printf 'warning: %s\n' "$1" >&2
}

die() {
  printf 'error: %s\n' "$1" >&2
  exit 1
}

run_ctl() {
  (
    cd "${ROOT_DIR}"
    go run ./cmd/layer-osctl "$@"
  )
}

daemon_write_auth_enabled() {
  local raw
  raw="$(run_ctl auth status 2>/dev/null)" || return 1
  python3 -c 'import json,sys; print("true" if json.load(sys.stdin).get("write_auth_enabled") else "false")' <<<"${raw}"
}

daemon_reachable() {
  run_ctl daemon status >/dev/null 2>&1
}

pid_is_running() {
  local pid="$1"
  [[ -n "${pid}" ]] || return 1
  kill -0 "${pid}" >/dev/null 2>&1
}

read_pid_file() {
  local path="$1"
  [[ -f "${path}" ]] || return 1
  tr -d '[:space:]' < "${path}"
}

start_daemon_if_needed() {
  if daemon_reachable; then
    note "daemon: already reachable"
    return 0
  fi

  local pid_file="${STATE_DIR}/daemon.pid"
  local log_file="${STATE_DIR}/daemon.log"
  local pid=""

  if pid="$(read_pid_file "${pid_file}" 2>/dev/null || true)" && pid_is_running "${pid}"; then
    note "daemon: process ${pid} already running; waiting for readiness"
  else
    note "daemon: starting in background"
    (
      cd "${ROOT_DIR}"
      nohup "${SCRIPT_DIR}/start.sh" >"${log_file}" 2>&1 &
      echo $! >"${pid_file}"
    )
  fi

  local attempts=0
  until daemon_reachable; do
    attempts=$((attempts + 1))
    if (( attempts >= 30 )); then
      warn "daemon log tail:"
      tail -n 20 "${log_file}" 2>/dev/null || true
      die "daemon did not become reachable"
    fi
    sleep 1
  done
  note "daemon: ready"
}

worker_pid_file() {
  printf '%s/%s.pid' "${STATE_DIR}" "$1"
}

worker_log_file() {
  printf '%s/%s.log' "${STATE_DIR}" "$1"
}

start_worker_role() {
  local role="$1"
  local poll="$2"
  local idle_exit_after="$3"
  local pid_file log_file pid
  pid_file="$(worker_pid_file "${role}")"
  log_file="$(worker_log_file "${role}")"

  pid="$(read_pid_file "${pid_file}" 2>/dev/null || true)"
  if pid_is_running "${pid}"; then
    note "worker:${role}: already running as pid ${pid}"
    return 0
  fi

  note "worker:${role}: starting"
  (
    cd "${ROOT_DIR}"
    nohup go run ./cmd/layer-osctl job work \
      --roles "${role}" \
      --command './scripts/job_worker_codex.sh' \
      --poll "${poll}" \
      --idle-exit-after "${idle_exit_after}" \
      >"${log_file}" 2>&1 &
    echo $! >"${pid_file}"
  )
}

ensure_worker_roles_running() {
  local roles="$1"
  local poll="$2"
  local idle_exit_after="$3"
  local role
  IFS=',' read -r -a role_items <<<"${roles}"
  for role in "${role_items[@]}"; do
    role="${role// /}"
    [[ -n "${role}" ]] || continue
    start_worker_role "${role}" "${poll}" "${idle_exit_after}"
  done
}

warn_if_write_token_missing() {
  if [[ -n "${LAYER_OS_WRITE_TOKEN:-}" ]]; then
    return 0
  fi
  case "$(daemon_write_auth_enabled 2>/dev/null || printf 'unknown')" in
    true)
      warn "write auth is enabled on the daemon; set LAYER_OS_WRITE_TOKEN before expecting workers to report successfully"
      ;;
    false)
      ;;
    *)
      warn "could not confirm daemon write-auth state; if reports fail, set LAYER_OS_WRITE_TOKEN and retry"
      ;;
  esac
}

command_up() {
  local roles="${ROLE_ARGS_DEFAULT}"
  local poll="${POLL_DEFAULT}"
  local idle_exit_after="${IDLE_DEFAULT}"
  while [[ "$#" -gt 0 ]]; do
    case "$1" in
      --roles)
        [[ "$#" -ge 2 ]] || die "--roles requires a value"
        roles="$2"
        shift 2
        ;;
      --poll)
        [[ "$#" -ge 2 ]] || die "--poll requires a value"
        poll="$2"
        shift 2
        ;;
      --idle-exit-after)
        [[ "$#" -ge 2 ]] || die "--idle-exit-after requires a value"
        idle_exit_after="$2"
        shift 2
        ;;
      *)
        usage
        exit 2
        ;;
    esac
  done

  start_daemon_if_needed
  warn_if_write_token_missing
  ensure_worker_roles_running "${roles}" "${poll}" "${idle_exit_after}"

  command_status
}

command_status() {
  note "state_dir=${STATE_DIR}"
  if daemon_reachable; then
    note "daemon=reachable"
  else
    note "daemon=unreachable"
  fi

  local role pid
  for role in implementer verifier planner; do
    pid="$(read_pid_file "$(worker_pid_file "${role}")" 2>/dev/null || true)"
    if pid_is_running "${pid}"; then
      note "worker_${role}=running pid=${pid} log=$(worker_log_file "${role}")"
    else
      note "worker_${role}=stopped"
    fi
  done
}

command_down() {
  local role pid stopped=0
  for role in implementer verifier planner; do
    pid="$(read_pid_file "$(worker_pid_file "${role}")" 2>/dev/null || true)"
    if pid_is_running "${pid}"; then
      kill "${pid}" >/dev/null 2>&1 || true
      note "worker:${role}: stopped pid ${pid}"
      stopped=1
    fi
    rm -f "$(worker_pid_file "${role}")"
  done
  if [[ -f "${STATE_DIR}/daemon.pid" ]]; then
    pid="$(read_pid_file "${STATE_DIR}/daemon.pid" 2>/dev/null || true)"
    if pid_is_running "${pid}"; then
      note "daemon: left running as pid ${pid}"
    fi
  fi
  if [[ "${stopped}" -eq 0 ]]; then
    note "workers: nothing to stop"
  fi
}

default_kind_for_role() {
  case "$1" in
    verifier) printf 'verify' ;;
    planner) printf 'plan' ;;
    designer) printf 'design' ;;
    *) printf 'implement' ;;
  esac
}

default_stage_for_role() {
  case "$1" in
    verifier) printf 'verify' ;;
    planner) printf 'discover' ;;
    designer) printf 'experience' ;;
    *) printf 'compose' ;;
  esac
}

validate_submit_role() {
  case "$1" in
    implementer|verifier|planner|designer)
      ;;
    *)
      die "role must be implementer, verifier, planner, or designer"
      ;;
  esac
}

worker_roles_for_submit() {
  case "$1" in
    planner) printf 'implementer,verifier,planner' ;;
    designer) printf 'implementer,verifier,designer' ;;
    *) printf '%s' "${ROLE_ARGS_DEFAULT}" ;;
  esac
}

extract_job_id() {
  python3 -c 'import json,sys; print(json.load(sys.stdin)["job_id"])'
}

command_submit() {
  local summary=""
  local role="implementer"
  local kind=""
  local stage=""
  local source="founder.manual"
  local allowed_paths=""
  local payload_json=""

  while [[ "$#" -gt 0 ]]; do
    case "$1" in
      --summary)
        [[ "$#" -ge 2 ]] || die "--summary requires a value"
        summary="$2"
        shift 2
        ;;
      --role)
        [[ "$#" -ge 2 ]] || die "--role requires a value"
        role="$2"
        shift 2
        ;;
      --kind)
        [[ "$#" -ge 2 ]] || die "--kind requires a value"
        kind="$2"
        shift 2
        ;;
      --stage)
        [[ "$#" -ge 2 ]] || die "--stage requires a value"
        stage="$2"
        shift 2
        ;;
      --source)
        [[ "$#" -ge 2 ]] || die "--source requires a value"
        source="$2"
        shift 2
        ;;
      --allowed-paths)
        [[ "$#" -ge 2 ]] || die "--allowed-paths requires a value"
        allowed_paths="$2"
        shift 2
        ;;
      --payload-json)
        [[ "$#" -ge 2 ]] || die "--payload-json requires a value"
        payload_json="$2"
        shift 2
        ;;
      *)
        usage
        exit 2
        ;;
    esac
  done

  [[ -n "${summary}" ]] || die "submit requires --summary"
  validate_submit_role "${role}"
  start_daemon_if_needed
  warn_if_write_token_missing
  ensure_worker_roles_running "$(worker_roles_for_submit "${role}")" "${POLL_DEFAULT}" "${IDLE_DEFAULT}"

  if [[ -z "${kind}" ]]; then
    kind="$(default_kind_for_role "${role}")"
  fi
  if [[ -z "${stage}" ]]; then
    stage="$(default_stage_for_role "${role}")"
  fi

  local create_output job_id
  local -a create_args
  create_args=(
    job create
    --kind "${kind}"
    --role "${role}"
    --summary "${summary}"
    --source "${source}"
    --stage "${stage}"
  )
  if [[ -n "${allowed_paths}" ]]; then
    create_args+=(--allowed-paths "${allowed_paths}")
  fi
  if [[ -n "${payload_json}" ]]; then
    create_args+=(--payload-json "${payload_json}")
  fi
  create_output="$(run_ctl "${create_args[@]}")"
  job_id="$(printf '%s' "${create_output}" | extract_job_id)"
  run_ctl job dispatch --id "${job_id}" >/dev/null
  note "submitted_job_id=${job_id}"
  note "role=${role}"
  note "summary=${summary}"
}

main() {
  [[ "$#" -ge 1 ]] || {
    usage
    exit 2
  }

  local subcommand="$1"
  shift
  case "${subcommand}" in
    up) command_up "$@" ;;
    status) command_status ;;
    down) command_down ;;
    submit) command_submit "$@" ;;
    -h|--help|help)
      usage
      ;;
    *)
      usage
      exit 2
      ;;
  esac
}

main "$@"
