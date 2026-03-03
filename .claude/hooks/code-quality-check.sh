#!/bin/bash
# code-quality-check.sh — PostToolUse(Edit|Write) Python 품질 체크
# core/**/*.py 대상으로 공통 위반 패턴 감지
#
# exit 2 = 위반 시 차단 (Claude가 즉시 수정해야 함)
# exit 0 = 통과

# stdin에서 JSON 읽기 (Claude Code가 stdin으로 전달)
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('file_path',''))" 2>/dev/null)

if [ -z "$FILE_PATH" ]; then
  exit 0
fi

# ─── CSS / HTML 시각 검증 강제 ──────────────────────────────
case "$FILE_PATH" in
  */website/assets/css/style.css)
    echo "━━━ VISUAL SCAN REQUIRED ━━━"
    echo "style.css 수정 감지. 커밋 전 의무 체크리스트:"
    echo "  1. getBoundingClientRect() 로 문제 요소 픽셀 측정"
    echo "  2. 모바일(390px) + 데스크탑(1280px) 양방향 스크린샷"
    echo "  3. 수정된 패턴이 사용되는 모든 페이지 타입 확인"
    echo "     (about / archive / essay / practice / home)"
    echo "  4. 수치가 증명한 후에만 커밋"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
    ;;
  */website/*.html|*/website/**/*.html)
    BASENAME=$(basename "$FILE_PATH")
    echo "━━━ HTML EDIT: ${BASENAME} ━━━"
    echo "HTML 수정 감지. 확인 사항:"
    echo "  1. 연관 컴포넌트(nav/footer) 변경 시 → build.py --components 실행"
    echo "  2. 수정 페이지 모바일 + 데스크탑 스크린샷 확인"
    echo "  3. 동일 패턴 사용 다른 페이지 파급 체크"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
    ;;
esac

# core/**/*.py 또는 scripts/**/*.py만 검사
case "$FILE_PATH" in
  */core/*.py|*/scripts/*.py)
    ;;
  *)
    exit 0
    ;;
esac

if [ ! -f "$FILE_PATH" ]; then
  exit 0
fi

VIOLATIONS=""

# 1. f-string 로깅 — 차단
FSTRING_HITS=$(grep -nE 'logger\.(info|debug|warning|error|critical)\(f["\x27]' "$FILE_PATH" 2>/dev/null)
if [ -n "$FSTRING_HITS" ]; then
  VIOLATIONS="${VIOLATIONS}\n🚫 f-string 로깅 위반 — lazy formatting (%) 사용 필수:\n${FSTRING_HITS}"
fi

# 2. 빈 except — 차단
BARE_EXCEPT=$(grep -nE '^\s*except:\s*$' "$FILE_PATH" 2>/dev/null)
if [ -n "$BARE_EXCEPT" ]; then
  VIOLATIONS="${VIOLATIONS}\n🚫 bare except 위반 — 구체적 예외 타입 지정 필수:\n${BARE_EXCEPT}"
fi

# 3. 하드코딩 비밀 — 차단
SECRET_HITS=$(grep -nE '(api_key|secret|token|password)\s*=\s*["\x27][a-zA-Z0-9_-]{16,}["\x27]' "$FILE_PATH" 2>/dev/null)
if [ -n "$SECRET_HITS" ]; then
  VIOLATIONS="${VIOLATIONS}\n🚫 하드코딩 비밀 위반 — os.getenv() 사용 필수:\n${SECRET_HITS}"
fi

# 4. import * — 경고만 (차단 아님)
if grep -qE '^\s*from\s+\S+\s+import\s+\*' "$FILE_PATH" 2>/dev/null; then
  echo "[CodeQuality] ⚠️  wildcard import 감지: $(basename "$FILE_PATH") — 명시적 import 권장"
fi

if [ -n "$VIOLATIONS" ]; then
  echo -e "[CodeQuality] 🚫 BLOCKED: $(basename "$FILE_PATH")${VIOLATIONS}"
  exit 2
fi

exit 0
