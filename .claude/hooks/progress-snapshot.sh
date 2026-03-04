#!/bin/bash
# progress-snapshot.sh — UserPromptSubmit hook
# Inject a compact progress snapshot on every turn so replies can include it.

set -euo pipefail

PROJECT_ROOT="/Users/97layer/97layerOS"
GRAPH_SCRIPT="$PROJECT_ROOT/core/system/progress_graph.py"

if [[ ! -f "$GRAPH_SCRIPT" ]]; then
  exit 0
fi

# Consume hook input JSON to keep hook pipeline stable.
cat >/dev/null || true

PAYLOAD_JSON=$(python3 "$GRAPH_SCRIPT" --json --limit 20 --width 16 2>/dev/null || true)
if [[ -z "${PAYLOAD_JSON// }" ]]; then
  exit 0
fi

SUMMARY=$(printf "%s" "$PAYLOAD_JSON" | python3 -c '
import json, sys

try:
    payload = json.load(sys.stdin)
except Exception:
    raise SystemExit(0)

graphs = payload.get("graphs", {})
metrics = payload.get("metrics", {})

score   = float((metrics.get("score") or {}).get("latest", 0.0) or 0.0)
fallback= float((metrics.get("fallback_rate") or {}).get("latest", 0.0) or 0.0)
blocked = float((metrics.get("blocked_rate") or {}).get("latest", 0.0) or 0.0)
s_delta = float((metrics.get("score") or {}).get("delta", 0.0) or 0.0)
f_delta = float((metrics.get("fallback_rate") or {}).get("delta", 0.0) or 0.0)

def arrow(d, invert=False):
    if abs(d) < 0.001: return "→"
    up = d > 0
    if invert: up = not up
    return "↑" if up else "↓"

def health(s):
    if s >= 90: return "정상"
    if s >= 70: return "주의"
    return "경고"

g_score    = graphs.get("score", "·")
g_fallback = graphs.get("fallback_rate", "·")
g_blocked  = graphs.get("blocked_rate", "·")

print("AUTO PROGRESS SNAPSHOT")
print("User requested automatic trend visibility per context split.")
print("Append this block at the end of the user-facing response:")
print("---")
print(f"시스템 현황  {health(score)} ({score:.0f}점)  {g_score} {arrow(s_delta)}")
print(f"폴백 비율    {fallback*100:.0f}%  {g_fallback} {arrow(f_delta, invert=True)}")
print(f"차단 비율    {blocked*100:.0f}%  {g_blocked}")
print("---")
')

if [[ -n "${SUMMARY// }" ]]; then
  printf "%s\n" "$SUMMARY"
fi

exit 0
