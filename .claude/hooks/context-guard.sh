#!/bin/bash
# context-guard.sh — PreToolUse hook
# 코드 수정 전 필수 맥락 확인 강제

# stdin에서 JSON 파싱
INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    print(data.get('tool', {}).get('name', ''))
except:
    pass
" 2>/dev/null)

# Edit/Write 도구 사용 시에만 체크
if [[ "$TOOL_NAME" != "Edit" && "$TOOL_NAME" != "Write" ]]; then
  exit 0
fi

cat << 'EOF'
━━━ CONTEXT GUARD ━━━
코드 수정 전 필수 확인:

✅ the_origin.md 읽었나?
✅ 관련 practice.md 읽었나?
✅ 기존 코드 Read 했나?
✅ Dependency Graph 확인했나?

읽지 않았으면 → 먼저 읽고 수정.
"아마도", "추측", "아마" → 금지.
━━━━━━━━━━━━━━━━━━━━━━━━
EOF

exit 0
