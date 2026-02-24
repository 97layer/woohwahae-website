#!/bin/bash
# proactive-advisor.sh — UserPromptSubmit hook
# 매 사용자 메시지 전에 능동 사고 트리거를 주입한다.
# stdout → Claude의 system-reminder로 삽입됨.

# 입력 JSON에서 메시지 추출 (분기 판단용)
INPUT=$(cat)
MESSAGE=$(echo "$INPUT" | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    # message is in data['message'] or nested
    msg = data.get('message', '')
    if isinstance(msg, list):
        # array of content blocks
        for block in msg:
            if isinstance(block, dict) and block.get('type') == 'text':
                print(block.get('text', '')[:300])
                break
    else:
        print(str(msg)[:300])
except Exception:
    print('')
" 2>/dev/null)

# 단순 확인/단답 요청은 트리거 생략 (노이즈 방지)
SKIP_PATTERNS="^(오케이|ok|ㅇㅋ|응|네|아니|ㄴ|ㅇ|맞아|좋아|고마워|감사|됐어|확인)$"
NORMALIZED=$(echo "$MESSAGE" | tr '[:upper:]' '[:lower:]' | xargs)

if echo "$NORMALIZED" | grep -qiE "$SKIP_PATTERNS"; then
  exit 0
fi

cat << 'EOF'
━━━ PROACTIVE ADVISOR MODE ━━━
응답 전 반드시 아래 순서로 스캔하라:

① INTENT — 사용자가 실제로 원하는 게 뭔가? (표면 요청 ≠ 진짜 의도)
② SIDE EFFECTS — 이걸 실행하면 뭐가 함께 바뀌거나 깨지나?
③ BLIND SPOTS — 사용자가 모르고 있는데 알아야 할 것이 있나?
④ SIMPLER PATH — 더 짧거나 더 좋은 방법이 있나?

스캔 결과 중 하나라도 발견되면 → 실행 전에 먼저 말한다.
아무것도 없으면 → 즉시 실행.

금지: 빈 공감 / 실행 후 침묵 / 단점 생략
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
EOF

exit 0
