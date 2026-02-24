#!/bin/bash
# code-quality-check.sh — PostToolUse(Edit|Write) Python 품질 체크
# core/**/*.py 대상으로 공통 위반 패턴 감지
#
# exit 0 = 항상 통과 (경고만 출력)

# Write 도구: file_path 추출
FILE_PATH=$(echo "$CLAUDE_TOOL_INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('file_path',''))" 2>/dev/null)

# Edit 도구: file_path 추출
if [ -z "$FILE_PATH" ]; then
  FILE_PATH=$(echo "$CLAUDE_TOOL_INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('file_path',''))" 2>/dev/null)
fi

if [ -z "$FILE_PATH" ]; then
  exit 0
fi

# core/**/*.py 또는 scripts/**/*.py만 검사
case "$FILE_PATH" in
  */core/*.py|*/scripts/*.py)
    ;;
  *)
    exit 0
    ;;
esac

# 파일 존재 확인
if [ ! -f "$FILE_PATH" ]; then
  exit 0
fi

WARNINGS=""

# 1. f-string 로깅 감지
if grep -nE 'logger\.(info|debug|warning|error|critical)\(f["\x27]' "$FILE_PATH" 2>/dev/null; then
  WARNINGS="${WARNINGS}\n⚠️  f-string 로깅 감지 — lazy formatting (%) 사용 필수"
fi

# 2. 빈 except 감지
if grep -nE '^\s*except:\s*$' "$FILE_PATH" 2>/dev/null; then
  WARNINGS="${WARNINGS}\n⚠️  빈 except 감지 — 구체적 예외 타입 지정 필수"
fi

# 3. 하드코딩 비밀 감지
if grep -nE '(api_key|secret|token|password)\s*=\s*["\x27][a-zA-Z0-9_-]{16,}["\x27]' "$FILE_PATH" 2>/dev/null; then
  WARNINGS="${WARNINGS}\n⚠️  하드코딩 비밀 감지 — os.getenv() 사용 필수"
fi

# 4. import * 감지
if grep -nE '^\s*from\s+\S+\s+import\s+\*' "$FILE_PATH" 2>/dev/null; then
  WARNINGS="${WARNINGS}\n⚠️  wildcard import 감지 — 명시적 import 사용 권장"
fi

if [ -n "$WARNINGS" ]; then
  echo -e "[CodeQuality] $(basename "$FILE_PATH")$WARNINGS"
fi

exit 0
