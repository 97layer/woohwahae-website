#!/bin/bash
# deepwork-tracker.sh — PostToolUse(Edit|Write) 딥워크 상태 추적
#
# 파일 수정 발생 시 deepwork.md에 자동 기록.
# 컨텍스트 압축 후 재개 시에도 "어디까지 했는지" 복원 가능.
#
# exit 0 = 항상 통과 (추적 실패해도 작업 중단 안 함)

PROJECT_ROOT="/Users/97layer/97layerOS"
DEEPWORK="$PROJECT_ROOT/knowledge/agent_hub/deepwork.md"

INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('tool_input', {}).get('file_path', ''))
except:
    print('')
" 2>/dev/null)

if [ -z "$FILE_PATH" ]; then
  exit 0
fi

# deepwork.md가 없거나 status: idle이면 추적 안 함
if [ ! -f "$DEEPWORK" ]; then
  exit 0
fi

STATUS=$(grep -m1 "^status:" "$DEEPWORK" 2>/dev/null | awk '{print $2}')
if [ "$STATUS" != "active" ]; then
  exit 0
fi

# 상대 경로로 변환
REL_PATH="${FILE_PATH#$PROJECT_ROOT/}"
TS=$(date "+%H:%M")

# Modified 섹션에 파일 기록 (중복 방지: 같은 파일이면 타임스탬프만 갱신)
if grep -q "^- $REL_PATH" "$DEEPWORK" 2>/dev/null; then
  # 기존 항목 타임스탬프 갱신
  sed -i.bak "s|^- $REL_PATH.*|— $REL_PATH ($TS)|" "$DEEPWORK" 2>/dev/null
  rm -f "$DEEPWORK.bak"
else
  # 새 항목 추가 (## Modified 섹션)
  if grep -q "^## Modified" "$DEEPWORK"; then
    sed -i.bak "/^## Modified/a\\
— $REL_PATH ($TS)" "$DEEPWORK" 2>/dev/null
    rm -f "$DEEPWORK.bak"
  fi
fi

# last_activity 갱신
sed -i.bak "s/^last_activity:.*/last_activity: $(date '+%H:%M:%S')/" "$DEEPWORK" 2>/dev/null
rm -f "$DEEPWORK.bak"

exit 0
