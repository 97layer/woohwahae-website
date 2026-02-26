#!/bin/bash
# command-guard.sh — PreToolUse(Bash) 파괴적 명령 + VM 배포 차단
# 위험 명령 패턴 감지 → 차단
#
# exit 0 = 허용
# exit 2 = 차단 (stderr 메시지 출력)

# stdin에서 JSON 읽기 (Claude Code가 stdin으로 전달)
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('command',''))" 2>/dev/null)

if [ -z "$COMMAND" ]; then
  exit 0
fi

# ─── VM 직접 접근 차단 (/deploy 스킬 강제) ────────────────────
# SSH: alias, IP, user@host 전부 차단
# SCP: 동일
# 허용: deploy.sh 경유만 (deploy.sh 내부에서 ssh 호출은 사용자가 직접 실행)

VM_PATTERNS='97layer-vm|136\.109\.201\.201|skyto5339_gmail_com@'

if echo "$COMMAND" | grep -qE "(ssh|scp)\s+.*($VM_PATTERNS)"; then
  echo "BLOCKED: VM 직접 SSH/SCP 금지." >&2
  echo "  → /deploy 스킬 또는 deploy.sh 사용하세요." >&2
  echo "  → 배포 가능 서비스: knowledge/system/vm_services.json 참조" >&2
  exit 2
fi

# systemctl 원격 실행 차단 (ssh 경유)
if echo "$COMMAND" | grep -qE "ssh.*systemctl"; then
  echo "BLOCKED: VM systemctl 직접 실행 금지. /deploy 스킬 사용하세요." >&2
  exit 2
fi

# ─── 파괴적 명령 패턴 ───────────────────────────────────────

# 시스템 파괴
if echo "$COMMAND" | grep -qE 'rm\s+(-[a-zA-Z]*f[a-zA-Z]*\s+)?/[^U]'; then
  echo "BLOCKED: 루트 경로 삭제 시도 감지" >&2
  exit 2
fi

# 환경변수 노출 (비밀 유출 위험)
if echo "$COMMAND" | grep -qE '^\s*(env|printenv)\s*$'; then
  echo "BLOCKED: 환경변수 전체 출력 금지 (비밀 유출 위험)" >&2
  exit 2
fi

# .env 파일 읽기 (로컬 + SSH 원격 포함)
if echo "$COMMAND" | grep -qE '(cat|less|more|head|tail|vi|vim|nano)\s+.*\.env'; then
  echo "BLOCKED: .env 파일 직접 읽기 금지 (비밀 유출 위험)" >&2
  exit 2
fi

# 파괴적 git 명령
if echo "$COMMAND" | grep -qE 'git\s+reset\s+--hard'; then
  echo "BLOCKED: git reset --hard 금지 (작업 손실 위험)" >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'git\s+clean\s+(-[a-zA-Z]*f|-fd)'; then
  echo "BLOCKED: git clean -f/-fd 금지 (추적되지 않은 파일 삭제)" >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'git\s+push\s+.*--force'; then
  echo "BLOCKED: force push 금지" >&2
  exit 2
fi

# Path traversal 공격
if echo "$COMMAND" | grep -qE '\.\./\.\./\.\./'; then
  echo "BLOCKED: 과도한 path traversal 감지 (3단계+)" >&2
  exit 2
fi

# 시스템 권한 변경
if echo "$COMMAND" | grep -qE 'chmod\s+(-[a-zA-Z]*\s+)?(777|666|o\+w)'; then
  echo "BLOCKED: 위험한 권한 변경 감지" >&2
  exit 2
fi

# 시스템 디렉토리 수정
if echo "$COMMAND" | grep -qE '(rm|mv|cp)\s+.*(/etc/|/usr/|/var/|/System/)'; then
  echo "BLOCKED: 시스템 디렉토리 수정 시도" >&2
  exit 2
fi

exit 0
