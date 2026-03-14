#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Antigravity and similar sandboxes behave best when Go caches stay under /tmp.
export GOCACHE="${GOCACHE:-/tmp/gocache}"
export GOMODCACHE="${GOMODCACHE:-/tmp/gomodcache}"
mkdir -p "${GOCACHE}" "${GOMODCACHE}"

usage() {
  printf 'usage: %s [--check]\n' "${0##*/}" >&2
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

local_operator_noise_snapshot() {
  local node_count="unknown"
  local tmux_count="unknown"
  local noise_count="unknown"
  local current_counts=""
  local current_tmux=""
  local current_noise=""

  if command -v ps >/dev/null 2>&1; then
    current_counts="$(ps -axo pid=,comm= 2>/dev/null | awk '$2 ~ /^(node|go|codex|tmux)$/ {count[$2]++} END {for (k in count) printf "%s=%d\n", k, count[k]}' | sort | tr '\n' ' ' || true)"
    current_noise="$(ps -axo pid=,etime=,command= 2>/dev/null | awk 'match($0, /oh-my-codex\/dist\/mcp\/|playwright-mcp|server-sequential-thinking|notebooklm-mcp|context7-mcp|chroma-mcp/) && $0 !~ /awk|trim_local_operator_noise\.sh/ {count++} END {print count+0}' || true)"
    if [[ "${current_counts}" =~ node=([0-9]+) ]]; then
      node_count="${BASH_REMATCH[1]}"
    else
      node_count="0"
    fi
    if [[ -n "${current_noise}" ]]; then
      noise_count="${current_noise}"
    fi
  fi

  if command -v tmux >/dev/null 2>&1; then
    current_tmux="$(tmux ls 2>/dev/null | wc -l | tr -d ' ' || true)"
    if [[ -n "${current_tmux}" ]]; then
      tmux_count="${current_tmux}"
    fi
  fi

  note "local_seat_target=layer-osd + optional 8081/admin + one builder lane (+ optional manual shell)"
  note "local_seat_noise_check=./scripts/trim_local_operator_noise.sh --check"
  note "local_seat_cleanup=./scripts/trim_local_operator_noise.sh --apply"
  note "local_node_processes=${node_count}"
  note "local_tmux_sessions=${tmux_count}"
  note "local_noise_processes=${noise_count}"

  if [[ "${node_count}" != "unknown" && "${node_count}" -gt 12 ]]; then
    warn "local operator seat looks noisy (node=${node_count}); consider ./scripts/trim_local_operator_noise.sh --apply"
  fi
  if [[ "${tmux_count}" != "unknown" && "${tmux_count}" -gt 4 ]]; then
    warn "local tmux session count is above the sweet spot (tmux=${tmux_count})"
  fi
  if [[ "${noise_count}" != "unknown" && "${noise_count}" -gt 0 ]]; then
    warn "detected duplicated MCP/tooling noise processes (${noise_count})"
  fi
}

trim_line() {
  local value="$1"
  value="${value%$'\r'}"
  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"
  printf '%s' "$value"
}

resolve_path() {
  local value="$1"
  if [[ "${value}" = /* ]]; then
    printf '%s' "${value}"
    return
  fi
  printf '%s/%s' "${ROOT_DIR}" "${value#./}"
}

source_tree_newer_than_binary() {
  local binary_path="$1"
  local probe_path
  local newer_match=""
  local probe_roots=(
    "${ROOT_DIR}/cmd"
    "${ROOT_DIR}/internal"
    "${ROOT_DIR}/contracts"
    "${ROOT_DIR}/constitution"
    "${ROOT_DIR}/docs"
    "${ROOT_DIR}/scripts"
    "${ROOT_DIR}/go.mod"
    "${ROOT_DIR}/go.sum"
  )

  if [[ ! -e "${binary_path}" ]]; then
    return 0
  fi

  for probe_path in "${probe_roots[@]}"; do
    if [[ ! -e "${probe_path}" ]]; then
      continue
    fi
    newer_match="$(find "${probe_path}" -type f -newer "${binary_path}" -print -quit 2>/dev/null || true)"
    if [[ -n "${newer_match}" ]]; then
      return 0
    fi
  done

  return 1
}

CHECK_ONLY=0
if [[ "${1:-}" == "--check" ]]; then
  CHECK_ONLY=1
  shift
fi
if [[ "$#" -ne 0 ]]; then
  usage
  exit 2
fi

export LAYER_OS_REPO_ROOT="${LAYER_OS_REPO_ROOT:-${ROOT_DIR}}"
export LAYER_OS_ADDR="${LAYER_OS_ADDR:-127.0.0.1:17808}"
export LAYER_OS_DATA_DIR="${LAYER_OS_DATA_DIR:-.layer-os}"
DATA_DIR_PATH="$(resolve_path "${LAYER_OS_DATA_DIR}")"
export LAYER_OS_DATA_DIR="${DATA_DIR_PATH}"

mkdir -p "${DATA_DIR_PATH}"

ENV_FILE="${LAYER_OS_PROVIDER_ENV_FILE:-${DATA_DIR_PATH}/providers.env}"
ENV_FILE="$(resolve_path "${ENV_FILE}")"
export LAYER_OS_PROVIDER_ENV_FILE="${ENV_FILE}"

ENV_FILE_STATUS="missing"
if [[ -f "${ENV_FILE}" ]]; then
  ENV_FILE_STATUS="loaded"
  while IFS= read -r raw_line || [[ -n "${raw_line}" ]]; do
    line="$(trim_line "${raw_line}")"

    if [[ -z "${line}" || "${line}" == \#* ]]; then
      continue
    fi

    if [[ "${line}" =~ ^[A-Za-z_][A-Za-z0-9_]*=.*$ ]]; then
      export "${line}"
    else
      warn "skipping invalid provider entry: ${raw_line}"
    fi
  done < "${ENV_FILE}"
else
  warn "${ENV_FILE} not found; starting daemon without provider env"
fi

# Load secrets from macOS Keychain (overrides any value in providers.env).
KEYCHAIN_STATUS="unavailable"
if command -v security >/dev/null 2>&1; then
  KEYCHAIN_STATUS="available"
  _load_keychain_secret() {
    local key="$1"
    local val
    val="$(security find-generic-password -a layer-os -s "${key}" -w 2>/dev/null || true)"
    if [[ -n "${val}" ]]; then
      export "${key}=${val}"
    fi
  }
  _load_keychain_secret GEMINI_API_KEY
  _load_keychain_secret GOOGLE_API_KEY
  _load_keychain_secret OPENAI_API_KEY
  _load_keychain_secret ANTHROPIC_API_KEY
  _load_keychain_secret TELEGRAM_BOT_TOKEN
  _load_keychain_secret TELEGRAM_FOUNDER_CHAT_ID
  _load_keychain_secret TELEGRAM_FOUNDER_DM_CHAT_ID
  _load_keychain_secret TELEGRAM_OPS_CHAT_ID
  _load_keychain_secret TELEGRAM_BRAND_CHAT_ID
  _load_keychain_secret TELEGRAM_CHAT_ID
  _load_keychain_secret THREADS_ACCESS_TOKEN
  _load_keychain_secret LAYER_OS_WRITE_TOKEN
fi

if [[ -z "${GEMINI_API_KEY:-}" && -n "${GOOGLE_API_KEY:-}" ]]; then
  export GEMINI_API_KEY="${GOOGLE_API_KEY}"
fi
if [[ -z "${GOOGLE_API_KEY:-}" && -n "${GEMINI_API_KEY:-}" ]]; then
  export GOOGLE_API_KEY="${GEMINI_API_KEY}"
fi

validate_runtime_data() {
  local file
  local invalid=()
  local unreadable=()

  if ! command -v jq >/dev/null 2>&1; then
    warn "jq not found; skipping runtime JSON validation"
    return 0
  fi

  if [[ ! -r "${DATA_DIR_PATH}" || ! -x "${DATA_DIR_PATH}" ]]; then
    if [[ "${CHECK_ONLY}" -eq 1 ]]; then
      warn "runtime data directory is not readable from this shell; skipping JSON validation and preferring live daemon surfaces"
      return 0
    fi
    die "runtime data directory is not readable from this shell; prefer live daemon surfaces or use a less restricted shell"
  fi

  for file in "${DATA_DIR_PATH}"/*.json; do
    if [[ ! -e "${file}" ]]; then
      continue
    fi
    if [[ "$(basename "${file}")" == "events_archive.json" ]]; then
      continue
    fi
    if [[ ! -r "${file}" ]]; then
      unreadable+=("${file}")
      continue
    fi
    if ! jq empty "${file}" >/dev/null 2>&1; then
      invalid+=("${file}")
    fi
  done

  if [[ "${#unreadable[@]}" -gt 0 ]]; then
    if [[ "${CHECK_ONLY}" -eq 1 ]]; then
      warn "runtime JSON is unreadable from this shell; skipping JSON validation and preferring live daemon surfaces"
      for file in "${unreadable[@]}"; do
        warn "unreadable JSON: ${file}"
      done
      return 0
    fi
    for file in "${unreadable[@]}"; do
      warn "unreadable JSON: ${file}"
    done
    die "runtime JSON is not readable from this shell; prefer live daemon surfaces or use a less restricted shell"
  fi

  if [[ "${#invalid[@]}" -eq 0 ]]; then
    return 0
  fi

  warn "runtime data validation failed for ${DATA_DIR_PATH}"
  for file in "${invalid[@]}"; do
    warn "invalid JSON: ${file}"
  done
  die "repair the runtime snapshot or start an isolated daemon with LAYER_OS_DATA_DIR=.layer-os-dev"
}

export LAYER_OS_GATEWAY_ADAPTER="${LAYER_OS_GATEWAY_ADAPTER:-gemini}"
export LAYER_OS_ARCHITECT_AUTODISPATCH="${LAYER_OS_ARCHITECT_AUTODISPATCH:-true}"
export LAYER_OS_ARCHITECT_AUTOVERIFY="${LAYER_OS_ARCHITECT_AUTOVERIFY:-true}"
export LAYER_OS_ARCHITECT_GEMINI_RECOVERY="${LAYER_OS_ARCHITECT_GEMINI_RECOVERY:-true}"
export LAYER_OS_ARCHITECT_GEMINI_CLEANUP="${LAYER_OS_ARCHITECT_GEMINI_CLEANUP:-true}"
export LAYER_OS_ARCHITECT_CORPUS_RECOVERY="${LAYER_OS_ARCHITECT_CORPUS_RECOVERY:-true}"
export LAYER_OS_ARCHITECT_CORPUS_CLEANUP="${LAYER_OS_ARCHITECT_CORPUS_CLEANUP:-true}"
export LAYER_OS_AGENT_ROLE_PROVIDERS="${LAYER_OS_AGENT_ROLE_PROVIDERS:-}"
export LAYER_OS_AGENT_ROLE_MODELS="${LAYER_OS_AGENT_ROLE_MODELS:-}"

DAEMON_COMMAND_DESC="go run ./cmd/layer-osd"
DAEMON_COMMAND=(go run ./cmd/layer-osd)
if [[ -x "${ROOT_DIR}/bin/layer-osd" ]]; then
  if source_tree_newer_than_binary "${ROOT_DIR}/bin/layer-osd"; then
    warn "bin/layer-osd is older than the source tree; using go run ./cmd/layer-osd"
  else
    DAEMON_COMMAND_DESC="${ROOT_DIR}/bin/layer-osd"
    DAEMON_COMMAND=("${ROOT_DIR}/bin/layer-osd")
  fi
fi

validate_runtime_data

if [[ "${CHECK_ONLY}" -eq 1 ]]; then
  note "layer-os bootstrap check: ok"
  note "repo_root=${ROOT_DIR}"
  note "data_dir=${DATA_DIR_PATH}"
  note "provider_env_file=${ENV_FILE}"
  note "provider_env_status=${ENV_FILE_STATUS}"
  note "keychain=${KEYCHAIN_STATUS}"
  note "daemon_command=${DAEMON_COMMAND_DESC}"
  note "gateway_adapter=${LAYER_OS_GATEWAY_ADAPTER}"
  note "agent_role_providers=${LAYER_OS_AGENT_ROLE_PROVIDERS}"
  note "agent_role_models=${LAYER_OS_AGENT_ROLE_MODELS}"
  note "architect_autodispatch=${LAYER_OS_ARCHITECT_AUTODISPATCH}"
  note "architect_autoverify=${LAYER_OS_ARCHITECT_AUTOVERIFY}"
  note "architect_gemini_recovery=${LAYER_OS_ARCHITECT_GEMINI_RECOVERY}"
  note "architect_gemini_cleanup=${LAYER_OS_ARCHITECT_GEMINI_CLEANUP}"
  note "architect_corpus_recovery=${LAYER_OS_ARCHITECT_CORPUS_RECOVERY}"
  note "architect_corpus_cleanup=${LAYER_OS_ARCHITECT_CORPUS_CLEANUP}"
  local_operator_noise_snapshot
  exit 0
fi

cd "${ROOT_DIR}"
exec "${DAEMON_COMMAND[@]}"
