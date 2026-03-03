#!/bin/bash
# PostToolUse: Edit/Write 후 Gemini가 변경된 .py 파일 자동 리뷰
# .py 파일이 아니면 즉시 종료

INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    # Edit: file_path / Write: file_path
    print(d.get('tool_input', {}).get('file_path', ''))
except Exception:
    print('')
" 2>/dev/null)

# .py 파일이 아니면 스킵
[[ "$FILE_PATH" != *.py ]] && exit 0

# 파일 존재 확인
[[ ! -f "$FILE_PATH" ]] && exit 0

# 환경변수 로드
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
set -a
[[ -f "$PROJECT_ROOT/.env" ]] && source "$PROJECT_ROOT/.env"
set +a

[[ -z "$GOOGLE_API_KEY" ]] && exit 0

# 파일 내용 (최대 200줄)
FILE_CONTENT=$(head -200 "$FILE_PATH")
REL_PATH="${FILE_PATH#$PROJECT_ROOT/}"

# Gemini Flash에 리뷰 요청
REVIEW=$(python3 - <<PYEOF
import os, sys, json
try:
    from google import genai
    client = genai.Client(api_key=os.environ['GOOGLE_API_KEY'])
    prompt = """You are a code reviewer. Review the following Python file and respond in Korean with ONLY this format (no other text):

score: [0-100]
verdict: [APPROVE|REVISE]
issues: [comma-separated list of up to 3 issues, or "없음"]

File: ${REL_PATH}
\`\`\`python
${FILE_CONTENT}
\`\`\`"""
    resp = client.models.generate_content(model='gemini-2.5-flash', contents=prompt)
    print(resp.text.strip())
except Exception as e:
    print(f"skip: {e}")
PYEOF
)

# score 추출
SCORE=$(echo "$REVIEW" | grep -oP 'score:\s*\K[0-9]+' | head -1)
VERDICT=$(echo "$REVIEW" | grep -oP 'verdict:\s*\K\S+' | head -1)
ISSUES=$(echo "$REVIEW" | grep -oP 'issues:\s*\K.+' | head -1)

[[ -z "$SCORE" ]] && exit 0

# 색상
if [[ "$VERDICT" == "APPROVE" || "$SCORE" -ge 85 ]]; then
    COLOR="\033[92m"  # green
elif [[ "$SCORE" -ge 65 ]]; then
    COLOR="\033[93m"  # yellow
else
    COLOR="\033[91m"  # red
fi
RESET="\033[0m"
BOLD="\033[1m"
DIM="\033[2m"

echo -e "\n${DIM}────────────────────────────────────────${RESET}"
echo -e "🤖 ${BOLD}Gemini Critic${RESET} · ${REL_PATH}"
echo -e "   score: ${COLOR}${BOLD}${SCORE}/100${RESET}  verdict: ${COLOR}${VERDICT}${RESET}"
[[ -n "$ISSUES" && "$ISSUES" != "없음" ]] && echo -e "   ${DIM}↳ ${ISSUES}${RESET}"
echo -e "${DIM}────────────────────────────────────────${RESET}"
