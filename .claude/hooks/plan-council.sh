#!/bin/bash
# plan-council.sh — UserPromptSubmit hook
# 복잡 태스크 감지 시 Claude+Gemini 계획 협의를 선행하고 요약을 주입한다.

set -euo pipefail

INPUT=$(cat)
PROJECT_ROOT="/Users/97layer/97layerOS"
PLAN_SCRIPT="$PROJECT_ROOT/core/system/plan_council.py"

if [[ ! -f "$PLAN_SCRIPT" ]]; then
  exit 0
fi

MESSAGE=$(echo "$INPUT" | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    msg = data.get('message', '')
    if isinstance(msg, list):
        chunks = []
        for block in msg:
            if isinstance(block, dict) and block.get('type') == 'text':
                chunks.append(str(block.get('text', '')))
        print('\\n'.join(chunks)[:1200])
    else:
        print(str(msg)[:1200])
except Exception:
    print('')
" 2>/dev/null)

if [[ -z "${MESSAGE// }" ]]; then
  exit 0
fi

NORMALIZED=$(echo "$MESSAGE" | tr '[:upper:]' '[:lower:]' | xargs)
SKIP_PATTERNS="^(오케이|ok|ㅇㅋ|응|네|아니|ㄴ|ㅇ|맞아|좋아|고마워|감사|됐어|확인)$"
if echo "$NORMALIZED" | grep -qiE "$SKIP_PATTERNS"; then
  exit 0
fi

IS_COMPLEX=$(printf "%s" "$MESSAGE" | python3 -c "
import re, sys
msg = sys.stdin.read().strip()
if not msg:
    print('0')
    raise SystemExit(0)

keywords = r'(구축|연결|설계|리팩토링|아키텍처|통합|연동|개선|수정|파이프라인|하네스|배포|마이그레이션|테스트|자동화|계획|플랜)'
length = len(msg)
complex_flag = bool(re.search(keywords, msg)) or length >= 90
print('1' if complex_flag else '0')
" 2>/dev/null)

if [[ "$IS_COMPLEX" != "1" ]]; then
  exit 0
fi

OUTPUT=$(python3 "$PLAN_SCRIPT" --task "$MESSAGE" --mode hook --max-items 3 2>/dev/null || true)
if [[ -n "${OUTPUT// }" ]]; then
  echo "$OUTPUT"
fi

exit 0
