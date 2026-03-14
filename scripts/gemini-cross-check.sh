#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPORT_FILE="${REPORT_FILE:-${TMPDIR:-/tmp}/layer-os-gemini-cross-check-report.md}"

export GOCACHE="${GOCACHE:-/tmp/gocache}"
export GOMODCACHE="${GOMODCACHE:-/tmp/gomodcache}"

CTL_COMMAND_DESC="go run ./cmd/layer-osctl"
CTL_COMMAND=(go run ./cmd/layer-osctl)
if [[ -x "${ROOT_DIR}/bin/layer-osctl" ]]; then
  CTL_COMMAND_DESC="${ROOT_DIR}/bin/layer-osctl"
  CTL_COMMAND=("${ROOT_DIR}/bin/layer-osctl")
fi

run_ctl() {
  (
    cd "${ROOT_DIR}"
    "${CTL_COMMAND[@]}" "$@"
  )
}

capture_ctl() {
  local output
  if ! output="$(run_ctl "$@" 2>&1)"; then
    printf '%s\n' "${output}"
    return 1
  fi
  printf '%s\n' "${output}"
}

render_ctl_json() {
  local filter="$1"
  shift

  local output
  if ! output="$(capture_ctl "$@")"; then
    printf '%s\n' "${output}"
    return 1
  fi
  if command -v jq >/dev/null 2>&1 && [[ -n "${filter}" ]]; then
    printf '%s\n' "${output}" | jq "${filter}"
    return 0
  fi
  printf '%s\n' "${output}"
}

echo "Running Gemini Cross-Check Pipeline..."

mkdir -p "$(dirname "${REPORT_FILE}")"

{
  echo "# 🛡️ Gemini Control Tower: Cross-Check Report"
  echo "**Generated At:** $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  echo "**Layer OS Surface:** ${CTL_COMMAND_DESC}"
  echo ""

  echo "## 1. 📚 Documentation Audit"
  echo '```json'
  "${ROOT_DIR}/scripts/doc_audit.sh" || echo "❌ Documentation Audit Failed"
  echo ""
  echo '```'

  echo "## 2. 📊 System Audit Status"
  echo '```json'
  run_ctl audit structure || echo "❌ Structure Audit Failed"
  echo ""
  run_ctl audit contracts || echo "❌ Contracts Audit Failed"
  echo ""
  run_ctl audit security || echo "❌ Security Audit Failed"
  echo '```'
  echo ""

  echo "## 3. 🧪 Test Suite Verification"
  echo '```text'
  (
    cd "${ROOT_DIR}"
    go test ./... -short
  ) || echo "❌ Tests Failed"
  echo '```'
  echo ""

  echo "## 4. 🧭 Bootstrap Context"
  echo '```json'
  render_ctl_json '{source, read_only, degraded, tooling, knowledge: {current_focus, next_steps, open_risks, review_top_open}, founder_summary: .handoff.founder_summary}' session bootstrap --full --allow-local-fallback || echo "❌ Bootstrap Check Failed"
  echo '```'
  echo ""

  echo "## 5. 🕵️‍♂️ Recent Jobs"
  echo '```json'
  render_ctl_json 'sort_by(.updated_at // .created_at) | reverse | .[:5] | map({job_id, role, status, summary, updated_at})' job list || echo "❌ Job List Failed"
  echo '```'
  echo ""

  echo "## 6. 🧾 Review Room Summary"
  echo '```json'
  render_ctl_json '{summary, top_open: .summary.top_open[:5]}' review-room || echo "❌ Review Room Read Failed"
  echo '```'
  echo ""

  echo "## 7. 📝 Active Sub-Agent Modifications (Git Status)"
  echo '```text'
  git -C "${ROOT_DIR}" status -s
  echo '```'
  echo ""

  echo "## 8. 🔍 Diff Summary"
  echo '```text'
  git -C "${ROOT_DIR}" diff --stat
  echo '```'
} > "${REPORT_FILE}"

echo "✅ Cross-Check Environment Ready. Report generated at: ${REPORT_FILE}"
