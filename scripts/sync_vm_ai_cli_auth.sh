#!/usr/bin/env bash
set -euo pipefail

REMOTE_HOST="${LAYER_OS_VM_HOST:-97layer-vm}"
REMOTE_USER="${LAYER_OS_VM_USER:-skyto5339_gmail_com}"
REMOTE_PORT="${LAYER_OS_VM_PORT:-22}"
REMOTE_SSH_KEY="${LAYER_OS_VM_SSH_KEY:-}"
REMOTE_PROVIDER_ENV="${LAYER_OS_VM_PROVIDER_ENV_PATH:-/etc/layer-os/providers.env}"
REMOTE_AUTH_FILE="${LAYER_OS_VM_AI_CLI_AUTH_FILE:-}"
APPLY=0

usage() {
  cat >&2 <<'EOF'
usage: sync_vm_ai_cli_auth.sh [--host <ssh-host>] [--user <remote-user>] [--port <port>] [--ssh-key <path>] [--apply]

Mirror the VM provider env into a user-shell AI CLI auth bridge without
printing secret values. The bridge maps GOOGLE_API_KEY -> GEMINI_API_KEY and
passes through OPENAI_API_KEY / ANTHROPIC_API_KEY when present.
EOF
}

die() {
  printf 'error: %s\n' "$1" >&2
  exit 1
}

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --host)
      [[ "$#" -ge 2 ]] || die "--host requires a value"
      REMOTE_HOST="$2"
      shift 2
      ;;
    --user)
      [[ "$#" -ge 2 ]] || die "--user requires a value"
      REMOTE_USER="$2"
      shift 2
      ;;
    --port)
      [[ "$#" -ge 2 ]] || die "--port requires a value"
      REMOTE_PORT="$2"
      shift 2
      ;;
    --ssh-key)
      [[ "$#" -ge 2 ]] || die "--ssh-key requires a value"
      REMOTE_SSH_KEY="$2"
      shift 2
      ;;
    --apply)
      APPLY=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

SSH_ARGS=(-p "${REMOTE_PORT}" -o StrictHostKeyChecking=accept-new)
if [[ -n "${REMOTE_SSH_KEY}" ]]; then
  SSH_ARGS+=(-i "${REMOTE_SSH_KEY}")
fi
REMOTE_TARGET="${REMOTE_USER}@${REMOTE_HOST}"

MODE="check"
if [[ "${APPLY}" -eq 1 ]]; then
  MODE="apply"
fi

ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "
  set -euo pipefail
  MODE='${MODE}'
  REMOTE_PROVIDER_ENV='${REMOTE_PROVIDER_ENV}'
  REMOTE_AUTH_FILE='${REMOTE_AUTH_FILE}'
  REMOTE_USER='${REMOTE_USER}'
  if [[ -z \"\${REMOTE_AUTH_FILE}\" ]]; then
    REMOTE_AUTH_FILE=\"\$HOME/.config/layer-os/ai-cli-auth.sh\"
  fi

  sudo MODE=\"\${MODE}\" \
    REMOTE_PROVIDER_ENV=\"\${REMOTE_PROVIDER_ENV}\" \
    REMOTE_AUTH_FILE=\"\${REMOTE_AUTH_FILE}\" \
    REMOTE_USER=\"\${REMOTE_USER}\" \
    python3 - <<'PY'
import os
import pwd
import shlex
from pathlib import Path

mode = os.environ['MODE']
provider_path = Path(os.environ['REMOTE_PROVIDER_ENV'])
auth_file = Path(os.environ['REMOTE_AUTH_FILE'].replace('$HOME', str(Path.home())))
remote_user = os.environ['REMOTE_USER']

values = {}
if provider_path.exists():
    for raw in provider_path.read_text().splitlines():
        line = raw.strip()
        if not line or line.startswith('#') or '=' not in line:
            continue
        key, value = line.split('=', 1)
        values[key.strip()] = value.strip()

openai = values.get('OPENAI_API_KEY', '')
anthropic = values.get('ANTHROPIC_API_KEY', '')
google = values.get('GOOGLE_API_KEY', '')
openai_status = 'present' if openai else 'missing'
anthropic_status = 'present' if anthropic else 'missing'
google_status = 'present' if google else 'missing'

print(f'provider_env={provider_path}')
print(f'OPENAI_API_KEY={openai_status}')
print(f'ANTHROPIC_API_KEY={anthropic_status}')
print(f'GOOGLE_API_KEY={google_status}')
print(f'GEMINI_API_KEY={google_status}')

if mode != 'apply':
    raise SystemExit(0)

auth_file.parent.mkdir(parents=True, exist_ok=True)
lines = [
    '# Generated from /etc/layer-os/providers.env for interactive AI CLIs.',
    '# Re-run scripts/sync_vm_ai_cli_auth.sh after provider changes.',
]
if openai:
    lines.append(f'export OPENAI_API_KEY={shlex.quote(openai)}')
else:
    lines.append('unset OPENAI_API_KEY')
if anthropic:
    lines.append(f'export ANTHROPIC_API_KEY={shlex.quote(anthropic)}')
else:
    lines.append('unset ANTHROPIC_API_KEY')
if google:
    quoted = shlex.quote(google)
    lines.append(f'export GOOGLE_API_KEY={quoted}')
    lines.append(f'export GEMINI_API_KEY={quoted}')
else:
    lines.append('unset GOOGLE_API_KEY')
    lines.append('unset GEMINI_API_KEY')
content = '\n'.join(lines) + '\n'
auth_file.write_text(content)

user_info = pwd.getpwnam(remote_user)
os.chown(auth_file, user_info.pw_uid, user_info.pw_gid)
os.chmod(auth_file, 0o600)
print(f'auth_file={auth_file}')
print('mode=apply')
PY

  if [[ \"\${MODE}\" == 'apply' ]]; then
    mkdir -p \"\$HOME/.config/layer-os\"
    for shell_rc in \"\$HOME/.profile\" \"\$HOME/.bash_profile\" \"\$HOME/.bashrc\" \"\$HOME/.zprofile\" \"\$HOME/.zshrc\"; do
      touch \"\${shell_rc}\"
      if ! grep -Fq 'ai-cli-auth.sh' \"\${shell_rc}\"; then
        printf '\n[ -f \"\$HOME/.config/layer-os/ai-cli-auth.sh\" ] && . \"\$HOME/.config/layer-os/ai-cli-auth.sh\"\n' >> \"\${shell_rc}\"
      fi
    done
  fi
"
