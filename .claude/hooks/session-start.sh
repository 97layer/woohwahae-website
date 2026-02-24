#!/bin/bash
# session-start.sh — SessionStart hook
# QUANTA 상태 요약 + work_lock 체크를 자동 출력
# CLAUDE.md의 "cat QUANTA" 지시를 hook으로 자동화
#
# exit 0 = 항상 통과 (정보 제공만)

PROJECT_ROOT="/Users/97layer/97layerOS"
QUANTA="$PROJECT_ROOT/knowledge/agent_hub/INTELLIGENCE_QUANTA.md"
WORK_LOCK="$PROJECT_ROOT/knowledge/system/work_lock.json"

echo "━━━ LAYER OS Session Start ━━━"

# ─── QUANTA 상태 ──────────────────────────────────────────

if [ -f "$QUANTA" ]; then
  # 최종 갱신 시각 추출
  LAST_UPDATE=$(grep -m1 "Last Updated" "$QUANTA" | sed 's/.*: //')
  if [ -n "$LAST_UPDATE" ]; then
    echo "QUANTA: $LAST_UPDATE"
  fi

  # 요약 30줄 출력
  head -30 "$QUANTA"
else
  echo "WARNING: QUANTA 파일 없음 — $QUANTA"
fi

echo ""

# ─── work_lock 상태 ───────────────────────────────────────

if [ -f "$WORK_LOCK" ]; then
  LOCKED=$(python3 -c "import json; d=json.load(open('$WORK_LOCK')); print(d.get('locked', False))" 2>/dev/null)
  if [ "$LOCKED" = "True" ]; then
    AGENT=$(python3 -c "import json; d=json.load(open('$WORK_LOCK')); print(d.get('agent_id', 'unknown'))" 2>/dev/null)
    TASK=$(python3 -c "import json; d=json.load(open('$WORK_LOCK')); print(d.get('task', 'unknown'))" 2>/dev/null)
    echo "⚠️  WORK LOCK ACTIVE — Agent: $AGENT, Task: $TASK"
    echo "다른 에이전트가 작업 중입니다. 충돌 주의."
  else
    echo "work_lock: unlocked"
  fi
else
  echo "work_lock: 파일 없음 (정상)"
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

exit 0
