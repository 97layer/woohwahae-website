#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF' >&2
usage: trim_local_operator_noise.sh [--check|--apply] [--keep-tmux <session>]...

Local operator seat cleanup for Layer OS.

Modes:
  --check   show counts, tmux sessions, listening ports, and noise candidates
  --apply   terminate stale tmux sessions and duplicated MCP/tooling noise

Examples:
  ./scripts/trim_local_operator_noise.sh --check
  ./scripts/trim_local_operator_noise.sh --apply
  ./scripts/trim_local_operator_noise.sh --apply --keep-tmux layeros
EOF
}

note() {
  printf '%s\n' "$1"
}

warn() {
  printf 'warning: %s\n' "$1" >&2
}

MODE="check"
KEEP_TMUX=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check)
      MODE="check"
      shift
      ;;
    --apply)
      MODE="apply"
      shift
      ;;
    --keep-tmux)
      [[ $# -ge 2 ]] || { usage; exit 2; }
      KEEP_TMUX+=("$2")
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

NOISE_REGEX='oh-my-codex/dist/mcp/|playwright-mcp|server-sequential-thinking|notebooklm-mcp|context7-mcp|chroma-mcp'
LISTEN_REGEX=':(8081|17808|3100|9700|9712)\b|node|go'

is_kept_tmux() {
  local session_name="$1"
  local kept
  for kept in "${KEEP_TMUX[@]}"; do
    if [[ "${kept}" == "${session_name}" ]]; then
      return 0
    fi
  done
  return 1
}

print_counts() {
  note "== process counts =="
  ps -axo pid=,comm= | awk '$2 ~ /^(node|go|codex|tmux)$/ {count[$2]++} END {for (k in count) print k, count[k]}' | sort || true
}

print_tmux_sessions() {
  note "== tmux sessions =="
  tmux ls 2>/dev/null || true
}

print_noise_processes() {
  note "== noise candidates =="
  ps -axo pid=,etime=,command= | awk -v re="${NOISE_REGEX}" '$0 ~ re && $0 !~ /awk|trim_local_operator_noise\.sh/ {print}' || true
}

print_listeners() {
  note "== listening ports =="
  lsof -nP -iTCP -sTCP:LISTEN | rg "${LISTEN_REGEX}" || true
}

kill_tmux_sessions() {
  local session_name
  while IFS= read -r session_name; do
    [[ -n "${session_name}" ]] || continue
    if is_kept_tmux "${session_name}"; then
      note "keeping tmux session: ${session_name}"
      continue
    fi
    tmux kill-session -t "${session_name}" 2>/dev/null || true
    note "killed tmux session: ${session_name}"
  done < <(tmux ls -F '#S' 2>/dev/null || true)
}

collect_noise_pids() {
  ps -axo pid=,etime=,command= | awk -v re="${NOISE_REGEX}" '$0 ~ re && $0 !~ /awk|trim_local_operator_noise\.sh/ {print $1}' || true
}

terminate_noise_processes() {
  local pids=()
  local pid
  while IFS= read -r pid; do
    [[ -n "${pid}" ]] || continue
    pids+=("${pid}")
  done < <(collect_noise_pids)

  if [[ "${#pids[@]}" -eq 0 ]]; then
    note "no noise processes matched"
    return 0
  fi

  note "terminating ${#pids[@]} noise processes"
  for pid in "${pids[@]}"; do
    kill "${pid}" 2>/dev/null || true
  done

  sleep 1

  for pid in "${pids[@]}"; do
    if kill -0 "${pid}" 2>/dev/null; then
      kill -9 "${pid}" 2>/dev/null || true
    fi
  done
}

note "Layer OS local operator sweet spot"
note "  1. control tower: layer-osd"
note "  2. cockpit: 8081/admin only when needed"
note "  3. builder lane: one active agent session"
note "  4. optional manual shell: verification or git only"

print_counts
print_tmux_sessions
print_noise_processes
print_listeners

if [[ "${MODE}" != "apply" ]]; then
  note "mode=check"
  exit 0
fi

kill_tmux_sessions
terminate_noise_processes

note "== post-cleanup =="
print_counts
print_tmux_sessions
print_noise_processes
print_listeners
