#!/bin/bash
# session-start.sh — SessionStart hook
# 관련 QUANTA 섹션만 선택 출력 (토큰 절약)
#
# 출력 순서: 핵심 정보 → 다음 작업 → 현재 상태
# 스킵: 완료된 작업 누적 (가장 긴 섹션), 콘텐츠 전략, 실행 명령

PROJECT_ROOT="/Users/97layer/97layerOS"
QUANTA="$PROJECT_ROOT/knowledge/agent_hub/INTELLIGENCE_QUANTA.md"
WORK_LOCK="$PROJECT_ROOT/knowledge/system/work_lock.json"

echo "━━━ LAYER OS Session Start ━━━"

# ─── QUANTA 선택 로드 ─────────────────────────────────────

if [ -f "$QUANTA" ]; then
  python3 - "$QUANTA" <<'PYEOF'
import sys
import re

path = sys.argv[1]
content = open(path, encoding="utf-8").read()

# 섹션 추출 함수
def extract_section(text, header):
    """헤더부터 다음 ## 헤더 직전까지 추출."""
    idx = text.find(header)
    if idx == -1:
        return None
    rest = text[idx:]
    # 다음 ## 헤더 찾기 (현재 헤더 제외)
    next_h = re.search(r'\n## ', rest[3:])
    if next_h:
        return rest[:next_h.start() + 3].strip()
    return rest.strip()

# 1. 파일 상단 메타 (5줄)
lines = content.splitlines()
meta_lines = [l for l in lines[:6] if l.strip()]
print("\n".join(meta_lines[:4]))
print()

# 2. 사람 소개 (압축 — 첫 6줄만)
person_section = extract_section(content, "## 👤 이 사람에 대해")
if person_section:
    person_lines = person_section.splitlines()[:8]
    print("\n".join(person_lines))
    print()

# 3. 다음 작업 (전체)
next_section = extract_section(content, "## 🎯 다음 작업")
if next_section:
    print(next_section)
    print()

# 4. 현재 상태 (전체)
current_section = extract_section(content, "## 📍 현재 상태 (CURRENT STATE)")
if current_section:
    print(current_section)
    print()

# 스킵: ✅ 완료된 작업 누적 / 📐 콘텐츠 전략 / 🚀 실행 명령 / 🏗️ 시스템 아키텍처
PYEOF
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
