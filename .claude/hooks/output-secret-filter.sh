#!/bin/bash
# output-secret-filter.sh — PostToolUse(Bash) 비밀 마스킹
# Bash 출력에서 API 키/토큰 패턴 감지 → 경고
#
# PostToolUse 훅은 stdin으로 JSON을 받는다 (CLAUDE_TOOL_OUTPUT 아님)
# exit 0 = 통과
# exit 2 = 차단 (stderr 메시지)

# stdin에서 JSON 읽기
INPUT=$(cat)

# tool_response 추출 (jq 없으면 python3 fallback)
if command -v jq &>/dev/null; then
  OUTPUT=$(echo "$INPUT" | jq -r '.tool_response // empty' 2>/dev/null)
else
  OUTPUT=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('tool_response',''))" 2>/dev/null)
fi

# 출력이 없으면 통과
if [ -z "$OUTPUT" ]; then
  exit 0
fi

# 비밀 패턴 정의
PATTERNS=(
  # Anthropic API keys
  'sk-ant-api[a-zA-Z0-9_-]{20,}'
  # Google API keys
  'AIza[a-zA-Z0-9_-]{35}'
  # GitHub tokens
  'ghp_[a-zA-Z0-9]{36}'
  'gho_[a-zA-Z0-9]{36}'
  'ghs_[a-zA-Z0-9]{36}'
  # Telegram bot tokens
  '[0-9]{8,10}:[a-zA-Z0-9_-]{35}'
  # OpenAI keys
  'sk-[a-zA-Z0-9]{20,}'
  # Generic KEY=value (환경변수 노출)
  '(API_KEY|SECRET_KEY|ACCESS_TOKEN|AUTH_TOKEN|PRIVATE_KEY|PASSWORD)=[^ ]{8,}'
  # AWS keys
  'AKIA[A-Z0-9]{16}'
)

for PATTERN in "${PATTERNS[@]}"; do
  if echo "$OUTPUT" | grep -qE "$PATTERN"; then
    echo "⚠️ SECRET DETECTED: Bash 출력에 비밀 정보 패턴 감지. 출력이 차단됩니다." >&2
    exit 2
  fi
done

exit 0
