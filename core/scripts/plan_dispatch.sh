#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PLAN_COUNCIL_SCRIPT="$PROJECT_ROOT/core/system/plan_council.py"
PLAN_CLASSIFIER_SCRIPT="${PLAN_DISPATCH_CLASSIFIER_SCRIPT:-$PROJECT_ROOT/core/system/plan_dispatch_classifier.py}"
PLAN_METRICS_SCRIPT="$PROJECT_ROOT/core/system/plan_dispatch_metrics.py"
ENV_FILE="$PROJECT_ROOT/.env"
SAFE_ENV_EXPORT_SCRIPT="$PROJECT_ROOT/core/scripts/safe_env_export.py"

# Load .env safely (KEY=VALUE lines only) to avoid command execution side effects.
if [[ -f "$ENV_FILE" ]]; then
  if [[ -f "$SAFE_ENV_EXPORT_SCRIPT" ]]; then
    # shellcheck disable=SC2046
    eval "$(python3 "$SAFE_ENV_EXPORT_SCRIPT" --file "$ENV_FILE")"
  else
    while IFS= read -r raw || [[ -n "$raw" ]]; do
      line="${raw#"${raw%%[![:space:]]*}"}"
      [[ -z "$line" ]] && continue
      [[ "$line" == \#* ]] && continue
      if [[ "$line" =~ ^([A-Za-z_][A-Za-z0-9_]*)=(.*)$ ]]; then
        key="${BASH_REMATCH[1]}"
        value="${BASH_REMATCH[2]}"
        if [[ "$value" =~ ^\"(.*)\"$ ]]; then
          value="${BASH_REMATCH[1]}"
        elif [[ "$value" =~ ^\'(.*)\'$ ]]; then
          value="${BASH_REMATCH[1]}"
        fi
        export "${key}=${value}"
      fi
    done < "$ENV_FILE"
  fi
fi

if [[ $# -lt 1 ]]; then
  echo "Usage: bash core/scripts/plan_dispatch.sh \"<task>\" [--auto|--manual] [--smoke]" >&2
  exit 2
fi

TASK="$1"
shift || true

MODE="auto"
SMOKE="0"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --auto)
      MODE="auto"
      ;;
    --manual)
      MODE="manual"
      ;;
    --smoke)
      SMOKE="1"
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 2
      ;;
  esac
  shift
done

AUTO_ENABLED="${PLAN_COUNCIL_AUTO:-1}"
MIN_COMPLEXITY="${PLAN_COUNCIL_MIN_COMPLEXITY:-medium}"
ALLOW_DEGRADED="${PLAN_DISPATCH_ALLOW_DEGRADED:-0}"
STRICT_RUNTIME="${PLAN_DISPATCH_STRICT_RUNTIME:-1}"
MIN_RELIABILITY="${PLAN_DISPATCH_MIN_RELIABILITY:-0.65}"
LOG_METRICS="${PLAN_DISPATCH_LOG_METRICS:-1}"
CLASSIFY_COMPLEXITY="unknown"
CLASSIFY_SCORE="0"
CLASSIFIER_FALLBACK="0"

log_dispatch_metric() {
  local phase="$1"
  local reason="$2"
  local executed="$3"
  if [[ "$LOG_METRICS" != "1" ]]; then
    return 0
  fi
  if [[ ! -f "$PLAN_METRICS_SCRIPT" ]]; then
    return 0
  fi
  python3 "$PLAN_METRICS_SCRIPT" --append \
    --task "$TASK" \
    --mode "$MODE" \
    --phase "$phase" \
    --reason "$reason" \
    --executed "$executed" \
    --complexity "$CLASSIFY_COMPLEXITY" \
    --score "$CLASSIFY_SCORE" \
    --fallback "$CLASSIFIER_FALLBACK" >/dev/null 2>&1 || true
}

if [[ ! -f "$PLAN_COUNCIL_SCRIPT" ]]; then
  log_dispatch_metric "error" "plan_council_missing" "false"
  python3 - "$TASK" "$MODE" <<'PY'
import json
import sys

task = sys.argv[1]
mode = sys.argv[2]
payload = {
    "dispatcher": {
        "mode": mode,
        "executed": False,
        "reason": "plan_council_missing",
        "complexity": "unknown",
        "score": 0,
    },
    "consensus": {
        "status": "degraded",
        "models_used": [],
        "planner_primary": "claude",
        "verifier_secondary": "gemini",
        "intent": task[:120],
        "approach": "plan_council.py missing",
        "steps": [
            "Break the request into concrete work units",
            "Inspect related files and dependencies",
            "Implement and run validation checks",
        ],
        "risks": ["plan_council.py missing"],
        "checks": [
            "Restore core/system/plan_council.py",
            "Re-run session bootstrap until READY",
        ],
        "tools": [],
        "decision": "go",
        "decision_conflict": False,
    },
}
print(json.dumps(payload, ensure_ascii=False))
PY
  exit 0
fi

if [[ -f "$PLAN_CLASSIFIER_SCRIPT" ]]; then
  CLASSIFY_JSON="$(python3 "$PLAN_CLASSIFIER_SCRIPT" \
    --task "$TASK" \
    --min-complexity "$MIN_COMPLEXITY" \
    --json 2>/dev/null || true)"
fi

if [[ -n "${CLASSIFY_JSON:-}" ]]; then
  if ! python3 - "$CLASSIFY_JSON" >/dev/null 2>&1 <<'PY'
import json
import sys

payload = json.loads(sys.argv[1])
if not isinstance(payload, dict):
    raise ValueError("classifier output must be object")
required = {"complexity", "score", "level", "threshold", "allowed"}
missing = [key for key in required if key not in payload]
if missing:
    raise ValueError(f"classifier output missing keys: {missing}")
PY
  then
    CLASSIFY_JSON=""
  fi
fi

if [[ -z "${CLASSIFY_JSON:-}" ]]; then
  CLASSIFIER_FALLBACK="1"
  CLASSIFY_JSON="$(python3 - "$MIN_COMPLEXITY" <<'PY'
import json
import sys

min_complexity = (sys.argv[1] or "medium").strip().lower()
threshold_map = {"simple": 0, "medium": 1, "high": 2}
threshold = threshold_map.get(min_complexity, 1)
print(json.dumps({
    "complexity": "simple",
    "score": 0,
    "level": 0,
    "threshold": threshold,
    "allowed": False,
    "signals": ["classifier_fallback_or_invalid"],
}, ensure_ascii=False))
PY
)"
fi

CLASSIFY_META="$(python3 - "$CLASSIFY_JSON" <<'PY'
import json
import sys
payload = json.loads(sys.argv[1])
complexity = str(payload.get("complexity", "unknown")).strip().lower() or "unknown"
score = payload.get("score", 0)
try:
    score = int(score)
except Exception:
    score = 0
print(f"{complexity}\t{score}")
PY
)"
CLASSIFY_COMPLEXITY="${CLASSIFY_META%%$'\t'*}"
CLASSIFY_SCORE="${CLASSIFY_META#*$'\t'}"

should_execute="1"
skip_reason=""

if [[ "$MODE" == "auto" ]]; then
  if [[ "$AUTO_ENABLED" != "1" ]]; then
    should_execute="0"
    skip_reason="auto_disabled"
  else
    allowed="$(python3 - "$CLASSIFY_JSON" <<'PY'
import json
import sys
obj = json.loads(sys.argv[1])
print("1" if obj.get("allowed") else "0")
PY
)"
    if [[ "$allowed" != "1" ]]; then
      should_execute="0"
      skip_reason="simple_task"
    fi
  fi
fi

if [[ "$should_execute" != "1" ]]; then
  log_dispatch_metric "skip" "${skip_reason:-simple_task}" "false"
  python3 - "$TASK" "$MODE" "$CLASSIFY_JSON" "$skip_reason" <<'PY'
import json
import sys

task = sys.argv[1]
mode = sys.argv[2]
classifier = json.loads(sys.argv[3])
reason = sys.argv[4]

payload = {
    "dispatcher": {
        "mode": mode,
        "executed": False,
        "reason": reason,
        "complexity": classifier.get("complexity", "simple"),
        "score": classifier.get("score", 0),
    },
    "consensus": {
        "status": "skipped",
        "models_used": [],
        "planner_primary": "claude",
        "verifier_secondary": "gemini",
        "intent": task[:120],
        "approach": "Plan Council skipped by dispatcher policy",
        "steps": [],
        "risks": [],
        "checks": [],
        "tools": [],
        "decision": "go",
        "decision_conflict": False,
    },
}
print(json.dumps(payload, ensure_ascii=False))
PY
  exit 0
fi

if [[ "$SMOKE" == "1" ]]; then
  if [[ "$should_execute" == "1" ]]; then
    log_dispatch_metric "smoke" "smoke_mode" "true"
  else
    log_dispatch_metric "smoke" "${skip_reason:-smoke_mode_skipped}" "false"
  fi
  python3 - "$TASK" "$MODE" "$CLASSIFY_JSON" "$should_execute" "$skip_reason" <<'PY'
import json
import sys
from datetime import datetime, timedelta, timezone

task = sys.argv[1]
mode = sys.argv[2]
classifier = json.loads(sys.argv[3])
should_execute = (sys.argv[4] == "1")
skip_reason = (sys.argv[5] or "").strip()
now = datetime.now(timezone.utc)
expires = now + timedelta(minutes=10)

payload = {
    "timestamp": now.isoformat(),
    "mode": "preflight",
    "task": task,
    "claude": {"ok": mode == "manual", "error": "smoke_mode", "plan": None},
    "gemini": {"ok": mode == "manual", "error": "smoke_mode", "plan": None},
    "consensus": {
        "status": "smoke",
        "models_used": ["smoke"],
        "planner_primary": "claude",
        "verifier_secondary": "gemini",
        "intent": task[:120],
        "approach": "Smoke mode: dispatcher/contract integrity check only (no live model call).",
        "steps": ["Validate dispatcher payload schema", "Validate manual execution branch"],
        "risks": ["No live model connectivity verification in smoke mode"],
        "checks": ["Run without --smoke for live council check before real implementation"],
        "tools": ["plan_dispatch"],
        "decision": "go",
        "decision_conflict": False,
        "runtime": {
            "gate_recommendation": "go",
            "reliability_score": 1.0,
            "reliability_tier": "high",
            "generated_at_utc": now.isoformat(),
            "expires_at_utc": expires.isoformat(),
            "ttl_seconds": 600,
            "stability_window": 0,
            "unstable": False,
        },
    },
}
payload["dispatcher"] = {
    "mode": mode,
    "executed": should_execute,
    "reason": "smoke_mode" if should_execute else (skip_reason or "smoke_mode_skipped"),
    "complexity": classifier.get("complexity", "unknown"),
    "score": classifier.get("score", 0),
}
print(json.dumps(payload, ensure_ascii=False))
PY
  exit 0
fi

# exit code 계약: 0=go, 1=hard_stop, 2=needs_clarification, 3=degraded_or_caution
COUNCIL_JSON=""
COUNCIL_EXIT=0
COUNCIL_JSON="$(python3 "$PLAN_COUNCIL_SCRIPT" --task "$TASK" --mode preflight --json 2>/dev/null)" || COUNCIL_EXIT=$?

if [[ "$COUNCIL_EXIT" -eq 1 ]]; then
  log_dispatch_metric "blocked" "hard_stop_model_unavailable" "false"
  echo "━━━ PLAN COUNCIL: HARD STOP ━━━" >&2
  echo "두 모델 모두 호출 실패 (네트워크/키 오류)." >&2
  echo "구현 금지. 원인 확인 후 재시도하세요." >&2
  exit 1
fi

if [[ "$COUNCIL_EXIT" -eq 2 ]]; then
  log_dispatch_metric "blocked" "needs_clarification_model" "false"
  echo "━━━ PLAN COUNCIL: NEEDS CLARIFICATION ━━━" >&2
  echo "Claude/Gemini 모두 범위 불명확으로 판정." >&2
  echo "사용자에게 요청 범위 확인 후 재실행하세요." >&2
  exit 2
fi

if [[ "$COUNCIL_EXIT" -eq 3 ]]; then
  echo "━━━ PLAN COUNCIL: DEGRADED (한 모델만 응답) ━━━" >&2
  echo "단일 모델 기준으로 진행합니다. 리스크 주의." >&2
  if [[ "$ALLOW_DEGRADED" != "1" ]]; then
    log_dispatch_metric "blocked" "degraded_not_allowed" "false"
    echo "명시적 승인/허용 없이 진행 금지. PLAN_DISPATCH_ALLOW_DEGRADED=1 설정 후 재실행하세요." >&2
    exit 3
  fi
fi

RUNTIME_GATE="$(python3 - "$COUNCIL_JSON" "$MIN_RELIABILITY" <<'PY'
import json
import sys

raw = (sys.argv[1] or "").strip()
min_reliability = float(sys.argv[2])

try:
    payload = json.loads(raw)
except Exception:
    print("hard_stop\tplan_council_response_parse_failed")
    raise SystemExit(0)

consensus = payload.get("consensus") or {}
runtime = consensus.get("runtime") or payload.get("runtime") or {}
action = str(runtime.get("gate_recommendation", "go")).strip().lower()
score = runtime.get("reliability_score")
unstable = bool(runtime.get("unstable", False))

try:
    score_val = float(score)
except Exception:
    score_val = 0.0

if score_val < min_reliability and action == "go":
    action = "caution"

reason = f"score={score_val:.3f}, unstable={str(unstable).lower()}"
print(f"{action}\t{reason}")
PY
)"

RUNTIME_ACTION="${RUNTIME_GATE%%$'\t'*}"
RUNTIME_REASON="${RUNTIME_GATE#*$'\t'}"

if [[ "$RUNTIME_ACTION" == "hard_stop" ]]; then
  log_dispatch_metric "blocked" "runtime_hard_stop" "false"
  echo "━━━ PLAN COUNCIL: HARD STOP (runtime gate) ━━━" >&2
  echo "$RUNTIME_REASON" >&2
  exit 1
fi

if [[ "$RUNTIME_ACTION" == "needs_clarification" ]]; then
  log_dispatch_metric "blocked" "runtime_needs_clarification" "false"
  echo "━━━ PLAN COUNCIL: NEEDS CLARIFICATION (runtime gate) ━━━" >&2
  echo "$RUNTIME_REASON" >&2
  exit 2
fi

if [[ "$RUNTIME_ACTION" == "caution" && "$STRICT_RUNTIME" == "1" && "$ALLOW_DEGRADED" != "1" ]]; then
  log_dispatch_metric "blocked" "runtime_caution_not_allowed" "false"
  echo "━━━ PLAN COUNCIL: CAUTION (runtime gate) ━━━" >&2
  echo "$RUNTIME_REASON" >&2
  echo "신뢰도 부족 상태입니다. 승인/완화 설정 없이 진행 금지." >&2
  exit 3
fi

log_dispatch_metric "execute" "executed" "true"
python3 - "$TASK" "$MODE" "$CLASSIFY_JSON" "$COUNCIL_JSON" <<'PY'
import json
import sys

task = sys.argv[1]
mode = sys.argv[2]
classifier = json.loads(sys.argv[3])
raw_council = (sys.argv[4] or "").strip()

payload = None
if raw_council:
    try:
        parsed = json.loads(raw_council)
        if isinstance(parsed, dict):
            payload = parsed
    except Exception:
        payload = None

if payload is None:
    payload = {
        "timestamp": "",
        "mode": "preflight",
        "task": task,
        "claude": {"ok": False, "error": "parse_failed", "plan": None},
        "gemini": {"ok": False, "error": "parse_failed", "plan": None},
        "consensus": {
            "status": "degraded",
            "models_used": [],
            "planner_primary": "claude",
            "verifier_secondary": "gemini",
            "intent": task[:120],
            "approach": "Plan Council call failed or returned invalid JSON",
            "steps": [
                "Break the request into concrete work units",
                "Inspect related files and dependencies",
                "Implement and run validation checks",
            ],
            "risks": ["Plan Council response parse failed"],
            "checks": [
                "core/system/plan_council.py --self-check --require-both",
                "Validate knowledge/system/plan_council_reports.jsonl integrity",
            ],
            "tools": [],
            "decision": "go",
            "decision_conflict": False,
        },
    }

payload["dispatcher"] = {
    "mode": mode,
    "executed": True,
    "reason": "executed",
    "complexity": classifier.get("complexity", "unknown"),
    "score": classifier.get("score", 0),
}

print(json.dumps(payload, ensure_ascii=False))
PY
