#!/usr/bin/env bash
# Plan Council Hook — 비동기 큐잉 전용
# 복잡 메시지를 감지하면 큐에 적재만 하고 즉시 종료한다 (지연/비용 없음).

set -euo pipefail

INPUT=$(cat)
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
TS="$(date -u +%Y%m%dT%H%M%SZ)"

# 메시지 추출 (기존 로직 재사용)
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

if [[ -z \"${MESSAGE// }\" ]]; then
  exit 0
fi

# 단순 응답은 스킵
NORMALIZED=$(echo \"$MESSAGE\" | tr '[:upper:]' '[:lower:]' | xargs)
SKIP_PATTERNS=\"^(오케이|ok|ㅇㅋ|응|네|아니|ㄴ|ㅇ|맞아|좋아|고마워|감사|됐어|확인)$\"
if echo \"$NORMALIZED\" | grep -qiE \"$SKIP_PATTERNS\"; then
  exit 0
fi

# 복잡 키워드/길이 필터
IS_COMPLEX=$(printf \"%s\" \"$MESSAGE\" | python3 -c \"\nimport re, sys\nmsg = sys.stdin.read().strip()\nif not msg:\n    print('0'); raise SystemExit(0)\nkeywords = r'(구축|연결|설계|리팩토링|아키텍처|통합|연동|개선|수정|파이프라인|하네스|배포|마이그레이션|테스트|자동화|계획|플랜)'\nlength = len(msg)\ncomplex_flag = bool(re.search(keywords, msg)) or length >= 90\nprint('1' if complex_flag else '0')\n\" 2>/dev/null)
if [[ \"$IS_COMPLEX\" != \"1\" ]]; then
  exit 0
fi

# 큐 적재
pending_dir=\"$ROOT/.infra/queue/council/pending\"
mkdir -p \"$pending_dir\"
cat > \"$pending_dir/${TS}_council.json\" <<EOF
{
  "intent": "${MESSAGE}",
  "proposal_id": "council_${TS}",
  "created_at": "${TS}"
}
EOF

echo \"[plan-council hook] queued intent: $MESSAGE\"
exit 0
