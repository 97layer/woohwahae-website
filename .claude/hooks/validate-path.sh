#!/bin/bash
# validate-path.sh — Write 후크: 금지 경로 검증
# PostToolUse(Write) 시 실행됨
#
# exit 2 = 위반 시 차단
# exit 0 = 통과

FILE_PATH=$(echo "$CLAUDE_TOOL_INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('file_path',''))" 2>/dev/null)

if [ -z "$FILE_PATH" ]; then
  exit 0
fi

PROJECT_ROOT="/Users/97layer/97layerOS"
REL_PATH="${FILE_PATH#$PROJECT_ROOT/}"

# 프로젝트 외부 파일은 무시
if [ "$REL_PATH" = "$FILE_PATH" ]; then
  exit 0
fi

BASENAME=$(basename "$REL_PATH")
DIRNAME=$(dirname "$REL_PATH")

# 1. 루트에 .md/.json/.txt 파일 (CLAUDE.md, README.md 제외) — 차단
if [ "$DIRNAME" = "." ]; then
  case "$BASENAME" in
    *.md|*.json|*.txt)
      if [ "$BASENAME" != "CLAUDE.md" ] && [ "$BASENAME" != "README.md" ]; then
        echo "[ValidatePath] 🚫 BLOCKED: 루트에 파일 생성 금지 — $BASENAME (허용: CLAUDE.md, README.md)"
        exit 2
      fi
      ;;
  esac
fi

# 2. 금지 파일명 패턴 — 차단
case "$BASENAME" in
  SESSION_SUMMARY_*|WAKEUP_REPORT*|DEEP_WORK_PROGRESS*|DEPLOY_*|NEXT_STEPS*|audit_report_*|*_report_*.json)
    echo "[ValidatePath] 🚫 BLOCKED: 금지 파일명 패턴: $BASENAME"
    exit 2
    ;;
esac

# 3. 임시 파일명 — 차단
case "$BASENAME" in
  temp_*|untitled_*|무제*)
    echo "[ValidatePath] 🚫 BLOCKED: 임시 파일명 감지: $BASENAME"
    exit 2
    ;;
esac

exit 0
