#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

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

echo "============================================================"
echo "🛡️ LAYER OS: CROSS-CHECK ENVIRONMENT (Gemini Control Tower)"
echo "============================================================"

echo "[1] 📊 System Audit Check"
echo "-> Documentation Audit:"
"${ROOT_DIR}/scripts/doc_audit.sh" || echo "❌ Documentation Audit Failed"
echo ""
echo "-> Structure Audit (${CTL_COMMAND_DESC}):"
run_ctl audit structure || echo "❌ Structure Audit Failed"
echo ""
echo "-> Contracts Audit (${CTL_COMMAND_DESC}):"
run_ctl audit contracts || echo "❌ Contracts Audit Failed"
echo ""
echo "-> Security Audit (${CTL_COMMAND_DESC}):"
run_ctl audit security || echo "❌ Security Audit Failed"
echo "✅ Audit Complete."
echo ""

echo "[2] 🧪 Test Suite Verification"
(
  cd "${ROOT_DIR}"
  go test ./... -short
) || echo "❌ Tests Failed"
echo "✅ Tests Complete."
echo ""

echo "[3] 📝 Git Status & Diff (Changes made by sub-agents)"
git -C "${ROOT_DIR}" status -s
echo "--- Unstaged Diffs ---"
git -C "${ROOT_DIR}" diff --stat
echo "----------------------"
echo ""

echo "[4] 🧭 Bootstrap Context"
render_ctl_json '{source, read_only, degraded, tooling, knowledge: {current_focus, next_steps, open_risks, review_top_open}, founder_summary: .handoff.founder_summary}' session bootstrap --full --allow-local-fallback || echo "❌ Bootstrap Check Failed"
echo ""

echo "[5] 🕵️‍♂️ Active Jobs & Pending Reviews"
echo "-> Top 5 Recent Jobs:"
render_ctl_json 'sort_by(.updated_at // .created_at) | reverse | .[:5] | map({job_id, role, status, summary, updated_at})' job list || echo "No jobs found."
echo ""

echo "-> Review Room Summary:"
render_ctl_json '{summary, top_open: .summary.top_open[:5]}' review-room || echo "No reviews found."
echo ""

echo "============================================================"
echo "🎯 CROSS-CHECK READY. Awaiting Control Tower Verification."
echo "============================================================"
