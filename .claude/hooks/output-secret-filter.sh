#!/bin/bash
# output-secret-filter.sh — PostToolUse(Bash) 비밀 마스킹
# Bash 출력에서 API 키/토큰 패턴 감지 → 마스킹
#
# exit 0 = 통과 (비밀 없음)
# exit 2 = 출력 대체 (마스킹된 출력을 stdout으로)

OUTPUT="$CLAUDE_TOOL_OUTPUT"

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
  '(API_KEY|SECRET_KEY|ACCESS_TOKEN|AUTH_TOKEN|PRIVATE_KEY)=[^ ]{8,}'
  # AWS keys
  'AKIA[A-Z0-9]{16}'
  # Generic long hex tokens (40+ chars)
  '[a-f0-9]{40,}'
)

MASKED="$OUTPUT"
FOUND=false

for PATTERN in "${PATTERNS[@]}"; do
  if echo "$MASKED" | grep -qE "$PATTERN"; then
    MASKED=$(echo "$MASKED" | sed -E "s/$PATTERN/****MASKED****/g")
    FOUND=true
  fi
done

if [ "$FOUND" = true ]; then
  echo "$MASKED"
  echo ""
  echo "[SecretFilter] 비밀 정보가 마스킹되었습니다."
  exit 2
fi

exit 0
