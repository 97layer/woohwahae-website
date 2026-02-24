#!/bin/bash
# validate-path.sh — Write 후크: 금지 경로 검증
# PostToolUse(Write) 시 실행됨

# $CLAUDE_TOOL_INPUT에서 file_path 추출
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

# 금지 패턴 체크
BASENAME=$(basename "$REL_PATH")
DIRNAME=$(dirname "$REL_PATH")

# 1. 루트에 .md 파일 (CLAUDE.md, README.md 제외)
if [ "$DIRNAME" = "." ] && [[ "$BASENAME" == *.md ]] && [ "$BASENAME" != "CLAUDE.md" ] && [ "$BASENAME" != "README.md" ]; then
  echo "⚠️  WARN: 루트에 .md 파일 생성 감지: $BASENAME — FILESYSTEM_MANIFEST 위반 가능"
  exit 0
fi

# 2. 금지 파일명 패턴
case "$BASENAME" in
  SESSION_SUMMARY_*|WAKEUP_REPORT*|DEEP_WORK_PROGRESS*|DEPLOY_*|NEXT_STEPS*)
    echo "⚠️  WARN: 금지 파일명 패턴: $BASENAME"
    exit 0
    ;;
esac

# 3. 임시 파일명
case "$BASENAME" in
  temp_*|untitled_*|무제*)
    echo "⚠️  WARN: 임시 파일명 감지: $BASENAME"
    exit 0
    ;;
esac

# 4. QUANTA 갱신 리마인더 (4시간 이상 미갱신)
QUANTA_FILE="$PROJECT_ROOT/knowledge/agent_hub/INTELLIGENCE_QUANTA.md"
if [ -f "$QUANTA_FILE" ]; then
  LAST_MOD=$(stat -f "%m" "$QUANTA_FILE" 2>/dev/null || stat -c "%Y" "$QUANTA_FILE" 2>/dev/null)
  NOW=$(date +%s)
  if [ -n "$LAST_MOD" ]; then
    DIFF=$(( (NOW - LAST_MOD) / 60 ))
    if [ "$DIFF" -gt 240 ]; then
      echo "⚠️  QUANTA ${DIFF}분 미갱신 — 업데이트 권장"
    fi
  fi
fi

exit 0
