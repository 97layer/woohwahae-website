#!/bin/bash
# session-stop.sh — Stop hook
# 세션 종료 시 QUANTA 갱신 상태 체크 + 핸드오프 리마인더
#
# exit 0 = 항상 통과

PROJECT_ROOT="/Users/97layer/97layerOS"
QUANTA="$PROJECT_ROOT/knowledge/agent_hub/INTELLIGENCE_QUANTA.md"

echo "━━━ LAYER OS Session Stop ━━━"

# ─── QUANTA 갱신 시각 체크 ────────────────────────────────

if [ -f "$QUANTA" ]; then
  LAST_MOD=$(stat -f "%m" "$QUANTA" 2>/dev/null || stat -c "%Y" "$QUANTA" 2>/dev/null)
  NOW=$(date +%s)

  if [ -n "$LAST_MOD" ]; then
    DIFF=$(( (NOW - LAST_MOD) / 60 ))

    if [ "$DIFF" -gt 120 ]; then
      echo "⚠️  QUANTA 미갱신: ${DIFF}분 전 마지막 수정"
      echo "→ /handoff 실행을 권장합니다."
    elif [ "$DIFF" -gt 60 ]; then
      echo "QUANTA: ${DIFF}분 전 갱신 (갱신 권장)"
    else
      echo "QUANTA: ${DIFF}분 전 갱신 (정상)"
    fi
  fi
else
  echo "WARNING: QUANTA 파일 없음"
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

exit 0
