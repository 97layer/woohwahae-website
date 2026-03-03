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

# 파일 경로 추출
FILE_PATH=$(echo "$INPUT" | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    ti = data.get('tool_input', {})
    print(ti.get('file_path', ''))
except:
    pass
" 2>/dev/null)

# CSS/HTML 시각 검증 강제
case "$FILE_PATH" in
  */website/assets/css/style.css)
    cat << 'CSSEOF'
━━━ CSS EDIT: 시각 검증 의무 ━━━
style.css 수정 전 확인:
  · 원인 Grep 선행 (추측 수정 금지)
  · 수정 후 의무: getBoundingClientRect() 픽셀 측정
  · 모바일(390px) + 데스크탑(1280px) 양방향 스크린샷
  · 동일 CSS 패턴 쓰는 다른 페이지 타입 파급 체크
  · 수치 증명 후에만 커밋 (눈대중 커밋 금지)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
CSSEOF
    exit 0
    ;;
  */website/*.html|*/website/**/*.html|*/website/_components/*.html)
    BASENAME=$(basename "$FILE_PATH" 2>/dev/null)
    echo "━━━ HTML EDIT: ${BASENAME} ━━━"
    echo "HTML 수정 전 확인:"
    echo "  · 컴포넌트 변경(nav/footer) 시 → 빌드 후 44개 파일 파급 확인"
    echo "  · 수정 페이지 모바일 + 데스크탑 스크린샷"
    echo "  · 동일 패턴 다른 페이지 자동 체크 (물어보기 전에)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
    ;;
esac

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
