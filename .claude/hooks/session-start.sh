#!/bin/bash
# session-start.sh — SessionStart hook
# QUANTA 전체 출력 + work_lock 체크
# (QUANTA가 ~65줄로 경량화되어 선택 로드 불필요)

PROJECT_ROOT="/Users/97layer/97layerOS"
QUANTA="$PROJECT_ROOT/knowledge/agent_hub/INTELLIGENCE_QUANTA.md"
WORK_LOCK="$PROJECT_ROOT/knowledge/system/work_lock.json"

echo "━━━ LAYER OS Session Start ━━━"

# ─── QUANTA 전체 출력 ─────────────────────────────────────

if [ -f "$QUANTA" ]; then
  cat "$QUANTA"
else
  echo "WARNING: QUANTA 파일 없음 — $QUANTA"
fi

echo ""

# ─── work_lock 상태 ───────────────────────────────────────

if [ -f "$WORK_LOCK" ]; then
  python3 - "$WORK_LOCK" <<'PYEOF'
import json, sys
try:
    d = json.load(open(sys.argv[1]))
    if d.get('locked'):
        print(f"⚠️  WORK LOCK ACTIVE — Agent: {d.get('agent_id','unknown')}, Task: {d.get('task','unknown')}")
        print("다른 에이전트가 작업 중입니다. 충돌 주의.")
    else:
        print("work_lock: unlocked")
except Exception as e:
    print(f"work_lock: 파싱 오류 ({e})")
PYEOF
else
  echo "work_lock: 파일 없음 (정상)"
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

exit 0
