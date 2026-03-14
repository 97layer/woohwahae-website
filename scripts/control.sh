#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Antigravity and similar sandboxes behave best when Go caches stay under /tmp.
export GOCACHE="${GOCACHE:-/tmp/gocache}"
export GOMODCACHE="${GOMODCACHE:-/tmp/gomodcache}"
mkdir -p "${GOCACHE}" "${GOMODCACHE}"

usage() {
  cat >&2 <<'EOF'
usage: control.sh [report|watch] [--interval <seconds>]

report  Print central runtime status plus current worktree dispatch lanes.
watch   Re-render the same control-tower report on a fixed interval.
EOF
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

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    die "required command not found: $1"
  fi
}

MODE="report"
INTERVAL=5

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    report|watch)
      MODE="$1"
      shift
      ;;
    --interval)
      [[ "$#" -ge 2 ]] || die "--interval requires a value"
      INTERVAL="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

if [[ ! "${INTERVAL}" =~ ^[0-9]+$ ]] || [[ "${INTERVAL}" -lt 1 ]]; then
  die "--interval must be a positive integer"
fi

require_cmd jq

DAEMON_CTL_COMMAND_DESC="go run ./cmd/layer-osctl"
DAEMON_CTL_COMMAND=(go run ./cmd/layer-osctl)
if [[ -x "${ROOT_DIR}/bin/layer-osctl" ]]; then
  DAEMON_CTL_COMMAND_DESC="${ROOT_DIR}/bin/layer-osctl"
  DAEMON_CTL_COMMAND=("${ROOT_DIR}/bin/layer-osctl")
fi

LOCAL_CTL_COMMAND_DESC="go run ./cmd/layer-osctl"
LOCAL_CTL_COMMAND=(go run ./cmd/layer-osctl)

run_daemon_ctl() {
  (
    cd "${ROOT_DIR}"
    "${DAEMON_CTL_COMMAND[@]}" "$@"
  )
}

run_local_ctl() {
  (
    cd "${ROOT_DIR}"
    "${LOCAL_CTL_COMMAND[@]}" "$@"
  )
}

capture_daemon_ctl() {
  local output
  if ! output="$(run_daemon_ctl "$@" 2>&1)"; then
    printf '%s\n' "${output}" >&2
    return 1
  fi
  printf '%s' "${output}"
}

capture_local_ctl() {
  local output
  if ! output="$(run_local_ctl "$@" 2>&1)"; then
    printf '%s\n' "${output}" >&2
    return 1
  fi
  printf '%s' "${output}"
}

print_lines() {
  local title="$1"
  local lines="$2"
  if [[ -z "${lines}" ]]; then
    note "${title}: -"
    return
  fi
  note "${title}:"
  while IFS= read -r line; do
    [[ -n "${line}" ]] || continue
    printf '  - %s\n' "${line}"
  done <<<"${lines}"
}

path_state_label() {
  local status="$1"
  local index_state worktree_state

  if [[ "${status}" == "??" ]]; then
    printf 'untracked\n'
    return 0
  fi

  index_state="${status:0:1}"
  worktree_state="${status:1:1}"
  if [[ "${index_state}" != " " ]] && [[ "${worktree_state}" != " " ]]; then
    printf 'staged+unstaged\n'
    return 0
  fi
  if [[ "${index_state}" != " " ]]; then
    printf 'staged\n'
    return 0
  fi
  if [[ "${worktree_state}" != " " ]]; then
    printf 'unstaged\n'
    return 0
  fi
  printf 'tracked\n'
}

hot_seam_kind() {
  case "$1" in
    .layer-os/*)
      printf 'runtime_state\n'
      ;;
    internal/runtime/types.go)
      printf 'runtime_types\n'
      ;;
    internal/runtime/knowledge.go)
      printf 'runtime_knowledge\n'
      ;;
    internal/runtime/continuity.go)
      printf 'runtime_continuity\n'
      ;;
    internal/runtime/service_session_bus.go)
      printf 'session_bus\n'
      ;;
    internal/api/router.go)
      printf 'api_router\n'
      ;;
    contracts/*.schema.json)
      printf 'contract_schema\n'
      ;;
    *)
      return 1
      ;;
  esac
}

classify_path() {
  local path="$1"
  local hot=""

  if hot="$(hot_seam_kind "${path}" 2>/dev/null)"; then
    printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
      "hot_seam" \
      "hold" \
      "control_tower" \
      "${path}" \
      "go test ./..." \
      "0" \
      "keep this lane gated at the control tower; hot seams must not be parallelized without explicit review"
    return 0
  fi

  case "${path}" in
    docs/brand-home/*)
      printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
        "brand_frontend" \
        "frontend" \
        "frontend implementer" \
        "docs/brand-home/**" \
        "targeted brand-home preview smoke" \
        "10" \
        "finish the brand-home surface only and keep hands off internal/, contracts/, and runtime state"
      ;;
    internal/runtime/*profile*.go|internal/runtime/*profile*_test.go)
      printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
        "backend_profiles" \
        "backend_writer" \
        "backend writer" \
        "internal/runtime/*profile*" \
        "go test ./..." \
        "20" \
        "complete the runtime profile lane only; stay inside the profile files and paired tests"
      ;;
    internal/runtime/runtime_data_audit.go|cmd/layer-osctl/work_flow_approval_test.go|cmd/layer-osctl/*approval*)
      printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
        "backend_audit" \
        "verifier" \
        "verifier" \
        "cmd/layer-osctl/** internal/runtime/runtime_data_audit.go" \
        "go test ./..." \
        "30" \
        "wire the audit and approval verification lane only; do not rewrite runtime profile logic unless blocked"
      ;;
    docs/agent-quickstart.md|docs/agent-role-seeds.md|docs/agent-integration.md|docs/prompting.md)
      printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
        "docs_agent_ops" \
        "support_writer" \
        "docs implementer" \
        "docs/agent-quickstart.md docs/agent-role-seeds.md docs/agent-integration.md docs/prompting.md" \
        "consistency review" \
        "40" \
        "align agent prompts and integration docs only; do not widen into runtime implementation"
      ;;
    internal/api/*)
      printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
        "backend_api" \
        "backend_writer" \
        "backend writer" \
        "internal/api/**" \
        "go test ./..." \
        "45" \
        "keep this API lane self-contained and avoid touching runtime hot seams unless approved"
      ;;
    internal/runtime/*)
      printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
        "backend_runtime" \
        "backend_writer" \
        "backend writer" \
        "internal/runtime/**" \
        "go test ./..." \
        "50" \
        "keep this runtime lane narrow and avoid shared seam files already reserved by control tower"
      ;;
    cmd/*)
      printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
        "backend_cli" \
        "backend_writer" \
        "backend writer" \
        "cmd/**" \
        "go test ./..." \
        "55" \
        "complete this CLI lane without widening into unrelated runtime or docs work"
      ;;
    scripts/*)
      printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
        "operator_scripts" \
        "support_writer" \
        "operator implementer" \
        "scripts/**" \
        "bash -n <touched script>" \
        "60" \
        "tighten operator tooling only and keep runtime semantics unchanged"
      ;;
    docs/*)
      printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
        "docs_support" \
        "support_writer" \
        "docs implementer" \
        "docs/**" \
        "consistency review" \
        "70" \
        "finish docs cleanup only and avoid turning a docs lane into a product or runtime lane"
      ;;
    constitution/*)
      printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
        "governance" \
        "support_writer" \
        "governance editor" \
        "constitution/**" \
        "consistency review" \
        "80" \
        "keep governance edits isolated and do not mix them with implementation work"
      ;;
    skills/*)
      printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
        "skills" \
        "support_writer" \
        "skill maintainer" \
        "skills/**" \
        "targeted skill smoke" \
        "85" \
        "keep this skill lane isolated and do not leak it into runtime changes"
      ;;
    *)
      printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
        "unclassified" \
        "hold" \
        "control_tower" \
        "${path}" \
        "manual review" \
        "90" \
        "review this lane centrally before dispatching it to a worker"
      ;;
  esac
}

lane_progress_label() {
  local staged="$1"
  local unstaged="$2"
  local untracked="$3"

  if [[ "${staged}" -gt 0 ]] && ([[ "${unstaged}" -gt 0 ]] || [[ "${untracked}" -gt 0 ]]); then
    printf 'mixed_inflight\n'
    return 0
  fi
  if [[ "${staged}" -gt 0 ]] && [[ "${unstaged}" -eq 0 ]] && [[ "${untracked}" -eq 0 ]]; then
    printf 'ready_for_verify\n'
    return 0
  fi
  if [[ "${staged}" -eq 0 ]] && ([[ "${unstaged}" -gt 0 ]] || [[ "${untracked}" -gt 0 ]]); then
    printf 'draft_inflight\n'
    return 0
  fi
  printf 'queued\n'
}

assign_slot() {
  local bucket="$1"
  case "${bucket}" in
    frontend)
      FRONTEND_SLOT_COUNT=$((FRONTEND_SLOT_COUNT + 1))
      ASSIGNED_SLOT="brand_page_${FRONTEND_SLOT_COUNT}"
      ;;
    backend_writer)
      BACKEND_SLOT_COUNT=$((BACKEND_SLOT_COUNT + 1))
      ASSIGNED_SLOT="mini_backend_${BACKEND_SLOT_COUNT}"
      ;;
    verifier)
      VERIFIER_SLOT_COUNT=$((VERIFIER_SLOT_COUNT + 1))
      ASSIGNED_SLOT="mini_verifier_${VERIFIER_SLOT_COUNT}"
      ;;
    support_writer)
      SUPPORT_SLOT_COUNT=$((SUPPORT_SLOT_COUNT + 1))
      ASSIGNED_SLOT="mini_support_${SUPPORT_SLOT_COUNT}"
      ;;
    hold|*)
      HOLD_SLOT_COUNT=$((HOLD_SLOT_COUNT + 1))
      ASSIGNED_SLOT="control_tower_review_${HOLD_SLOT_COUNT}"
      ;;
  esac
}

render_header() {
  local source="$1"
  local control_path="$2"
  note "Layer OS Control"
  note "time: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  note "source: ${source}"
  note "control_path: ${control_path}"
}

render_daemon_report() {
  local daemon_json bootstrap_json jobs_json next_line

  daemon_json="$(run_daemon_ctl daemon status 2>/dev/null)" || return 1
  bootstrap_json="$(capture_daemon_ctl session bootstrap --full)"
  jobs_json="$(run_daemon_ctl job list 2>/dev/null || true)"
  next_line="$(run_daemon_ctl next 2>/dev/null || true)"

  render_header "daemon" "${DAEMON_CTL_COMMAND_DESC}"
  note "daemon: $(jq -r '[.status, "@", .address, "uptime=" + (.uptime_seconds|tostring) + "s", "memory=" + .memory_health, "deploy=" + .deploy_health] | join(" ")' <<<"${daemon_json}")"
  note "bootstrap: read_only=$(jq -r '.read_only' <<<"${bootstrap_json}") degraded=$(jq -r '.degraded' <<<"${bootstrap_json}") source=$(jq -r '.source' <<<"${bootstrap_json}") route=session_bootstrap_full"
  note "progress: $(jq -r 'if .handoff.company_state.progress then (.handoff.company_state.progress.overall_percent|tostring) + "% " + .handoff.company_state.progress.overall_status else "-" end' <<<"${bootstrap_json}")"
  note "lane_counts: work_items_active=$(jq -r '.handoff.company_state.work_items_active' <<<"${bootstrap_json}") approvals_pending=$(jq -r '.handoff.company_state.approvals_pending' <<<"${bootstrap_json}")"
  note "focus: $(jq -r '.knowledge.current_focus // "-"' <<<"${bootstrap_json}")"
  note "goal: $(jq -r '.knowledge.current_goal // "-"' <<<"${bootstrap_json}")"
  note "founder_action: $(jq -r '[.knowledge.primary_action, .knowledge.primary_ref] | map(select(. != null and . != "")) | join(" @ ")' <<<"${bootstrap_json}")"
  note "founder_counts: now=$(jq -r '.handoff.founder_summary.now_count' <<<"${bootstrap_json}") waiting=$(jq -r '.handoff.founder_summary.waiting_count' <<<"${bootstrap_json}") risk=$(jq -r '.handoff.founder_summary.risk_count' <<<"${bootstrap_json}") review_open=$(jq -r '.handoff.founder_summary.review_open_count' <<<"${bootstrap_json}")"

  local rationale
  rationale="$(jq -r 'if .handoff.founder_summary.priority_rationale then [.handoff.founder_summary.priority_rationale.lane, .handoff.founder_summary.priority_rationale.source, .handoff.founder_summary.priority_rationale.reason] | join(" | ") else empty end' <<<"${bootstrap_json}")"
  if [[ -n "${rationale}" ]]; then
    note "priority_rationale: ${rationale}"
  fi

  note "environment: $(jq -r '[.handoff.founder_summary.environment_advisory.host_class, .handoff.founder_summary.environment_advisory.power_mode, .handoff.founder_summary.environment_advisory.continuity_role, .handoff.founder_summary.environment_advisory.agent_mode] | join(" / ")' <<<"${bootstrap_json}")"
  note "founder_notice: $(jq -r '.handoff.founder_summary.environment_advisory.founder_notice // "-"' <<<"${bootstrap_json}")"

  print_lines "next_steps" "$(jq -r '.knowledge.next_steps[:3] | .[]?' <<<"${bootstrap_json}")"
  print_lines "open_risks" "$(jq -r '.knowledge.open_risks[:3] | .[]?' <<<"${bootstrap_json}")"
  print_lines "review_top_open" "$(jq -r '.review_room.top_open[:3] | map("[" + .severity + "/" + .kind + "] " + .text) | .[]?' <<<"${bootstrap_json}")"
  print_lines "open_threads" "$(jq -r '.knowledge.open_threads[:3] | map(.question) | .[]?' <<<"${bootstrap_json}")"

  if [[ -n "${jobs_json}" ]] && jq empty >/dev/null 2>&1 <<<"${jobs_json}"; then
    local total queued running succeeded failed canceled
    total="$(jq 'length' <<<"${jobs_json}")"
    queued="$(jq '[.[] | select(.status == "queued")] | length' <<<"${jobs_json}")"
    running="$(jq '[.[] | select(.status == "running")] | length' <<<"${jobs_json}")"
    succeeded="$(jq '[.[] | select(.status == "succeeded")] | length' <<<"${jobs_json}")"
    failed="$(jq '[.[] | select(.status == "failed")] | length' <<<"${jobs_json}")"
    canceled="$(jq '[.[] | select(.status == "canceled")] | length' <<<"${jobs_json}")"
    note "jobs: total=${total} queued=${queued} running=${running} succeeded=${succeeded} failed=${failed} canceled=${canceled}"
  else
    note "jobs: unavailable (keeping live daemon bootstrap as the canonical read path)"
  fi
  if [[ -n "${next_line}" ]]; then
    note "next_job: ${next_line}"
  else
    note "next_job: -"
  fi
}

render_fallback_report() {
  local bootstrap_json start_check_output

  if ! bootstrap_json="$(run_local_ctl session bootstrap --allow-local-fallback --full 2>/dev/null)"; then
    warn "daemon unavailable and local fallback bootstrap failed"
    start_check_output="$("${ROOT_DIR}/scripts/start.sh" --check 2>&1 || true)"
    if [[ -n "${start_check_output}" ]]; then
      printf '%s\n' "${start_check_output}" >&2
    fi
    return 1
  fi

  render_header "local_fallback" "${LOCAL_CTL_COMMAND_DESC}"
  note "bootstrap: read_only=$(jq -r '.read_only' <<<"${bootstrap_json}") degraded=$(jq -r '.degraded' <<<"${bootstrap_json}") source=$(jq -r '.source' <<<"${bootstrap_json}")"
  note "focus: $(jq -r '.knowledge.current_focus // "-"' <<<"${bootstrap_json}")"
  note "goal: $(jq -r '.knowledge.current_goal // "-"' <<<"${bootstrap_json}")"
  note "founder_action: $(jq -r '[.knowledge.primary_action, .knowledge.primary_ref] | map(select(. != null and . != "")) | join(" @ ")' <<<"${bootstrap_json}")"
  note "review_open: $(jq -r '.knowledge.review_open_count' <<<"${bootstrap_json}")"
  note "environment: $(jq -r '[.knowledge.environment_advisory.host_class, .knowledge.environment_advisory.power_mode, .knowledge.environment_advisory.continuity_role, .knowledge.environment_advisory.agent_mode] | join(" / ")' <<<"${bootstrap_json}")"

  local tooling
  tooling="$(jq -r 'if .tooling then [.tooling.status, "required_mcp_ready=" + (.tooling.required_mcp_ready|tostring)] | join(" ") else empty end' <<<"${bootstrap_json}")"
  if [[ -n "${tooling}" ]]; then
    note "tooling: ${tooling}"
  fi

  print_lines "next_steps" "$(jq -r '.knowledge.next_steps[:3] | .[]?' <<<"${bootstrap_json}")"
  print_lines "open_risks" "$(jq -r '.knowledge.open_risks[:3] | .[]?' <<<"${bootstrap_json}")"
  print_lines "review_top_open" "$(jq -r '.knowledge.review_top_open[:3] | .[]?' <<<"${bootstrap_json}")"
  print_lines "open_threads" "$(jq -r '.knowledge.open_threads[:3] | map(.question) | .[]?' <<<"${bootstrap_json}")"

  local resume
  resume="$(jq -r 'if .resume then [.resume.current_focus, .resume.updated_at] | map(select(. != null and . != "")) | join(" @ ") else empty end' <<<"${bootstrap_json}")"
  if [[ -n "${resume}" ]]; then
    note "resume: ${resume}"
  fi
}

render_runtime_failure() {
  render_header "runtime_unavailable" "${ROOT_DIR}/scripts/start.sh --check"
  note "runtime: daemon unavailable and local fallback bootstrap failed"
  note "state_gate: fix runtime JSON or switch to an isolated LAYER_OS_DATA_DIR for full control-tower visibility"
}

render_worktree_report() {
  local status_lines line status path state staged_count unstaged_count untracked_count
  local lane bucket role owned verify priority dispatch hot record records_json summary_json
  local lane_json slot progress lane_lines dispatch_lines changed_lines hot_lines path_preview
  local files_total staged_total unstaged_total untracked_total hot_total lane_total

  note ""
  note "Worktree Control"

  if ! command -v git >/dev/null 2>&1; then
    note "worktree: unavailable (git not found)"
    return 0
  fi

  if ! git -C "${ROOT_DIR}" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    note "worktree: unavailable (not a git worktree)"
    return 0
  fi

  status_lines="$(git -C "${ROOT_DIR}" status --porcelain=v1)"
  if [[ -z "${status_lines}" ]]; then
    note "worktree: clean"
    note "agent_mix: frontend=0 backend=0 verifier=0 support=0 hold=0"
    note "dispatch_plan: no active lanes"
    return 0
  fi

  records_json=""
  while IFS= read -r line; do
    [[ -n "${line}" ]] || continue
    status="${line:0:2}"
    [[ "${status}" == "!!" ]] && continue

    path="${line:3}"
    if [[ "${path}" == *" -> "* ]]; then
      path="${path##* -> }"
    fi
    path="${path%\"}"
    path="${path#\"}"

    state="$(path_state_label "${status}")"
    staged_count=0
    unstaged_count=0
    untracked_count=0
    if [[ "${status}" == "??" ]]; then
      untracked_count=1
    else
      if [[ "${status:0:1}" != " " ]]; then
        staged_count=1
      fi
      if [[ "${status:1:1}" != " " ]]; then
        unstaged_count=1
      fi
    fi

    IFS=$'\t' read -r lane bucket role owned verify priority dispatch <<<"$(classify_path "${path}")"
    hot="$(hot_seam_kind "${path}" 2>/dev/null || true)"

    record="$(jq -nc \
      --arg path "${path}" \
      --arg state "${state}" \
      --arg status "${status}" \
      --arg lane "${lane}" \
      --arg bucket "${bucket}" \
      --arg role "${role}" \
      --arg owned "${owned}" \
      --arg verify "${verify}" \
      --arg dispatch "${dispatch}" \
      --arg hot "${hot}" \
      --argjson priority "${priority}" \
      --argjson staged "${staged_count}" \
      --argjson unstaged "${unstaged_count}" \
      --argjson untracked "${untracked_count}" \
      '{
        path: $path,
        state: $state,
        status: $status,
        lane: $lane,
        bucket: $bucket,
        role: $role,
        owned: $owned,
        verify: $verify,
        dispatch: $dispatch,
        hot: $hot,
        priority: $priority,
        staged: $staged,
        unstaged: $unstaged,
        untracked: $untracked
      }')"

    if [[ -z "${records_json}" ]]; then
      records_json="${record}"
    else
      records_json="${records_json}"$'\n'"${record}"
    fi
  done <<<"${status_lines}"

  if [[ -z "${records_json}" ]]; then
    note "worktree: clean"
    note "agent_mix: frontend=0 backend=0 verifier=0 support=0 hold=0"
    note "dispatch_plan: no active lanes"
    return 0
  fi

  summary_json="$(jq -s '
    {
      totals: {
        files: length,
        staged: (map(.staged) | add),
        unstaged: (map(.unstaged) | add),
        untracked: (map(.untracked) | add),
        hot: (map(select(.hot != "")) | length)
      },
      files: (sort_by(.priority, .lane, .path)),
      lanes: (
        group_by(.lane)
        | map({
            lane: .[0].lane,
            bucket: .[0].bucket,
            role: .[0].role,
            owned: .[0].owned,
            verify: .[0].verify,
            priority: .[0].priority,
            dispatch: .[0].dispatch,
            files: length,
            staged: (map(.staged) | add),
            unstaged: (map(.unstaged) | add),
            untracked: (map(.untracked) | add),
            hot: (map(select(.hot != "")) | length),
            paths: (map(.path) | unique),
            hot_refs: (map(select(.hot != "") | (.path + " [" + .hot + "]")) | unique)
          })
        | sort_by(.priority, .lane)
      ),
      hot: (map(select(.hot != "") | (.path + " [" + .hot + "]")) | unique)
    }
  ' <<<"${records_json}")"

  files_total="$(jq -r '.totals.files' <<<"${summary_json}")"
  staged_total="$(jq -r '.totals.staged' <<<"${summary_json}")"
  unstaged_total="$(jq -r '.totals.unstaged' <<<"${summary_json}")"
  untracked_total="$(jq -r '.totals.untracked' <<<"${summary_json}")"
  hot_total="$(jq -r '.totals.hot' <<<"${summary_json}")"
  lane_total="$(jq -r '.lanes | length' <<<"${summary_json}")"
  note "worktree: files=${files_total} staged=${staged_total} unstaged=${unstaged_total} untracked=${untracked_total} lanes=${lane_total} hot_seams=${hot_total}"

  changed_lines="$(jq -r '.files[:12] | .[] | "[" + .state + "] " + .path + " -> " + .lane' <<<"${summary_json}")"
  print_lines "changed_files" "${changed_lines}"

  FRONTEND_SLOT_COUNT=0
  BACKEND_SLOT_COUNT=0
  VERIFIER_SLOT_COUNT=0
  SUPPORT_SLOT_COUNT=0
  HOLD_SLOT_COUNT=0
  lane_lines=""
  dispatch_lines=""
  while IFS= read -r lane_json; do
    [[ -n "${lane_json}" ]] || continue
    bucket="$(jq -r '.bucket' <<<"${lane_json}")"
    assign_slot "${bucket}"
    slot="${ASSIGNED_SLOT}"
    progress="$(lane_progress_label \
      "$(jq -r '.staged' <<<"${lane_json}")" \
      "$(jq -r '.unstaged' <<<"${lane_json}")" \
      "$(jq -r '.untracked' <<<"${lane_json}")")"
    path_preview="$(jq -r '.paths[:3] | join(", ")' <<<"${lane_json}")"
    lane_lines="${lane_lines}${slot} | lane=$(jq -r '.lane' <<<"${lane_json}") | role=$(jq -r '.role' <<<"${lane_json}") | files=$(jq -r '.files' <<<"${lane_json}") | progress=${progress} | owned=$(jq -r '.owned' <<<"${lane_json}") | sample=${path_preview}"$'\n'
    dispatch_lines="${dispatch_lines}${slot} => $(jq -r '.dispatch' <<<"${lane_json}") | verify=$(jq -r '.verify' <<<"${lane_json}")"$'\n'
  done < <(jq -c '.lanes[]' <<<"${summary_json}")

  note "agent_mix: frontend=${FRONTEND_SLOT_COUNT} backend=${BACKEND_SLOT_COUNT} verifier=${VERIFIER_SLOT_COUNT} support=${SUPPORT_SLOT_COUNT} hold=${HOLD_SLOT_COUNT}"
  print_lines "lane_summary" "${lane_lines}"
  print_lines "dispatch_plan" "${dispatch_lines}"

  hot_lines="$(jq -r '.hot[]?' <<<"${summary_json}")"
  print_lines "hot_seams" "${hot_lines}"
}

report_once() {
  local status=0
  if render_daemon_report; then
    :
  elif render_fallback_report; then
    :
  else
    status=1
    render_runtime_failure
  fi
  render_worktree_report
  return "${status}"
}

clear_frame() {
  if [[ -t 1 ]]; then
    printf '\033[2J\033[H'
  fi
}

watch_report() {
  while true; do
    clear_frame
    if ! report_once; then
      warn "control report failed; retrying in ${INTERVAL}s"
    fi
    sleep "${INTERVAL}"
  done
}

case "${MODE}" in
  report)
    report_once
    ;;
  watch)
    watch_report
    ;;
  *)
    usage
    exit 2
    ;;
esac
