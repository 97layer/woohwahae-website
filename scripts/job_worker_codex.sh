#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'USAGE'
usage: job_worker_codex.sh

Consumes the env vars exported by `layer-osctl job work` and runs Codex
non-interactively against the current job packet.

Required env:
  LAYER_OS_REPO_ROOT
  LAYER_OS_JOB_WORK_DIR
  LAYER_OS_PROMPT_PATH
  LAYER_OS_PACKET_PATH
  LAYER_OS_RESULT_PATH
USAGE
}

require_env() {
  local key="$1"
  if [[ -z "${!key:-}" ]]; then
    printf 'error: missing required env %s\n' "$key" >&2
    exit 1
  fi
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

require_env LAYER_OS_REPO_ROOT
require_env LAYER_OS_JOB_WORK_DIR
require_env LAYER_OS_PROMPT_PATH
require_env LAYER_OS_PACKET_PATH
require_env LAYER_OS_RESULT_PATH

if ! command -v codex >/dev/null 2>&1; then
  printf 'error: codex CLI is required for job_worker_codex.sh\n' >&2
  exit 1
fi

WORK_DIR="${LAYER_OS_JOB_WORK_DIR}"
PROMPT_FILE="${WORK_DIR}/codex-prompt.md"
SCHEMA_FILE="${WORK_DIR}/codex-result.schema.json"
RAW_OUTPUT_FILE="${WORK_DIR}/codex-last-message.json"
MODEL_FLAG=()
MAX_ATTEMPTS="${LAYER_OS_CODEX_MAX_ATTEMPTS:-4}"
RETRY_BASE_DELAY="${LAYER_OS_CODEX_RETRY_BASE_DELAY_SECONDS:-15}"
SANDBOX_MODE="${LAYER_OS_CODEX_SANDBOX:-}"

case "${MAX_ATTEMPTS}" in
  ''|*[!0-9]*)
    MAX_ATTEMPTS=4
    ;;
esac
case "${RETRY_BASE_DELAY}" in
  ''|*[!0-9]*)
    RETRY_BASE_DELAY=15
    ;;
esac

if [[ -n "${LAYER_OS_CODEX_MODEL:-}" ]]; then
  MODEL_FLAG=(--model "${LAYER_OS_CODEX_MODEL}")
fi

if [[ -z "${SANDBOX_MODE}" ]]; then
  case "${LAYER_OS_JOB_ROLE:-}" in
    verifier|planner)
      SANDBOX_MODE="read-only"
      ;;
    *)
      SANDBOX_MODE="workspace-write"
      ;;
  esac
fi

cat >"${PROMPT_FILE}" <<PROMPT
$(cat "${LAYER_OS_PROMPT_PATH}")

Codex worker addendum:
- Read the packet from ${LAYER_OS_PACKET_PATH}.
- Work inside ${LAYER_OS_REPO_ROOT}.
- Respect allowed paths from LAYER_OS_ALLOWED_PATHS=${LAYER_OS_ALLOWED_PATHS:-not-set}.
- Job role is ${LAYER_OS_JOB_ROLE:-unknown}; planner/verifier lanes should stay read-only unless the packet explicitly asks for a repair.
- Treat the packet as the thin-first context boundary; do not widen into broad docs or repo scans unless the packet is provably insufficient.
- Before the first edit, prefer one narrow `rg` search and 3-6 directly related files over broad exploration.
- Pull wider handoff or extra documentation only when blocked on missing facts, not by default.
- Before making any multi-step decision or implementation plan, reason step by step in the terminal, prefer packet/contracts/scripts over extra tools, and keep tool calls minimal.
- Favor CLI verification and local file inspection before reaching for optional MCP tooling or extra context servers.
- Keep the first pass narrow, not passive: stay in one lane, but continue through that lane until the objective is complete or an escalation trigger fires.
- Do not stop after the first bounded step when the packet, local evidence, and allowed paths are sufficient to finish the current lane safely.
- Write your final answer as one JSON object only.
- The JSON must follow the canonical Layer OS report contract.
- Keep summary/open_risks/follow_on founder-readable first, then include technical detail.
- Do not invent side-channel completion files; the final answer JSON is the completion contract.
PROMPT

cat >"${SCHEMA_FILE}" <<'SCHEMA'
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "additionalProperties": true,
  "required": [
    "summary",
    "artifacts",
    "verification",
    "open_risks",
    "follow_on",
    "touched_paths",
    "blocked_paths"
  ],
  "properties": {
    "status": {
      "type": "string",
      "enum": ["succeeded", "failed", "canceled"]
    },
    "notes": {
      "type": "array",
      "items": {"type": "string"}
    },
    "summary": {"type": "string", "minLength": 1},
    "artifacts": {
      "type": "array",
      "items": {"type": "string"}
    },
    "verification": {
      "anyOf": [
        {"type": "array", "items": {"type": "string"}},
        {"type": "object"},
        {"type": "string"}
      ]
    },
    "open_risks": {
      "type": "array",
      "items": {"type": "string"}
    },
    "follow_on": {
      "type": "array",
      "items": {"type": "string"}
    },
    "touched_paths": {
      "type": "array",
      "items": {"type": "string"}
    },
    "blocked_paths": {
      "type": "array",
      "items": {"type": "string"}
    }
  }
}
SCHEMA
attempt=1
while true; do
  ATTEMPT_STDOUT="${WORK_DIR}/codex-attempt-${attempt}.stdout.log"
  ATTEMPT_STDERR="${WORK_DIR}/codex-attempt-${attempt}.stderr.log"
  set +e
  codex exec \
    --full-auto \
    --sandbox "${SANDBOX_MODE}" \
    --cd "${LAYER_OS_REPO_ROOT}" \
    --add-dir "${WORK_DIR}" \
    --output-schema "${SCHEMA_FILE}" \
    -o "${RAW_OUTPUT_FILE}" \
    "${MODEL_FLAG[@]}" \
    - <"${PROMPT_FILE}" >"${ATTEMPT_STDOUT}" 2>"${ATTEMPT_STDERR}"
  exit_code=$?
  set -e
  if [[ ${exit_code} -eq 0 ]]; then
    break
  fi

  if [[ -s "${ATTEMPT_STDERR}" ]]; then
    cat "${ATTEMPT_STDERR}" >&2
  fi
  if [[ -s "${ATTEMPT_STDOUT}" ]]; then
    cat "${ATTEMPT_STDOUT}" >&2
  fi

  if (( attempt >= MAX_ATTEMPTS )) || ! grep -qiE '429|too many requests|exceeded retry limit|remote compact task' "${ATTEMPT_STDERR}" "${ATTEMPT_STDOUT}" 2>/dev/null; then
    exit "${exit_code}"
  fi

  sleep_seconds=$(( RETRY_BASE_DELAY * attempt ))
  printf 'codex worker retrying after rate limit (attempt %d/%d) in %ss\n' "$((attempt + 1))" "${MAX_ATTEMPTS}" "${sleep_seconds}" >&2
  sleep "${sleep_seconds}"
  attempt=$((attempt + 1))
done

python3 - "${RAW_OUTPUT_FILE}" "${LAYER_OS_RESULT_PATH}" <<'PY'
import json, pathlib, sys
raw_path = pathlib.Path(sys.argv[1])
out_path = pathlib.Path(sys.argv[2])
data = json.loads(raw_path.read_text())
out_path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n")
PY

cat "${LAYER_OS_RESULT_PATH}"
