#!/usr/bin/env bash
set -euo pipefail

CONFIG_PATH="${CODEX_CONFIG:-$HOME/.codex/config.toml}"

optional_servers=(
  "sequential-thinking"
  "playwright"
  "notebooklm"
  "omx_code_intel"
  "omx_memory"
  "omx_state"
  "omx_team_run"
  "omx_trace"
)

warnings=0

note() {
  printf '%s\n' "$1"
}

pass() {
  printf 'PASS %s\n' "$1"
}

warn() {
  printf 'WARN %s\n' "$1"
}

warn_issue() {
  printf 'WARN %s\n' "$1"
  warnings=$((warnings + 1))
}

first_path_arg() {
  local raw="${1:-}"
  raw="$(printf '%s' "${raw}" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
  if [[ -z "${raw}" ]]; then
    printf '%s' ""
    return
  fi
  if [[ -e "${raw}" ]]; then
    printf '%s' "${raw}"
    return
  fi
  printf '%s' "${raw%% *}"
}

if ! command -v codex >/dev/null 2>&1; then
  warn_issue "codex command not found in PATH"
  note "Result: MCP diagnostics skipped"
  exit 0
fi

if [[ ! -f "${CONFIG_PATH}" ]]; then
  warn_issue "Codex config not found at ${CONFIG_PATH}"
  note "Result: MCP diagnostics skipped"
  exit 0
fi

note "Checking optional Codex MCP diagnostics (advisory only)"
note "Config: ${CONFIG_PATH}"
note "Core Layer OS work stays Go/CLI/daemon-first even when MCP entries are missing."

for server in "${optional_servers[@]}"; do
  if ! details="$(codex mcp get "${server}" 2>/dev/null)"; then
    warn_issue "${server}: not registered"
    continue
  fi

  pass "${server}: registered"

  if [[ "${details}" != *"enabled: true"* ]]; then
    warn_issue "${server}: not enabled"
  else
    pass "${server}: enabled"
  fi

  command_name="$(printf '%s\n' "${details}" | sed -n 's/^  command: //p' | head -n 1)"
  args_value="$(printf '%s\n' "${details}" | sed -n 's/^  args: //p' | head -n 1)"

  case "${command_name}" in
    node)
      first_arg="$(first_path_arg "${args_value}")"
      if [[ -z "${first_arg}" || ! -e "${first_arg}" ]]; then
        warn_issue "${server}: node entrypoint missing (${first_arg:-empty})"
      else
        pass "${server}: node entrypoint exists"
      fi
      ;;
    npx)
      if [[ -z "${args_value}" ]]; then
        warn_issue "${server}: npx package missing"
      else
        pass "${server}: npx package declared (${args_value})"
      fi
      ;;
    *)
      warn "${server}: unrecognized command (${command_name}); manual review recommended"
      ;;
  esac
done

if [[ "${warnings}" -gt 0 ]]; then
  note "Result: ${warnings} optional MCP warning(s) found"
  note "This script is advisory only; Layer OS core flows stay Go/CLI/daemon-first."
  exit 0
fi

note "Result: optional MCP checks look healthy"
note "Note: a new Codex session may still be needed before newly added MCP tools show up in chat."
