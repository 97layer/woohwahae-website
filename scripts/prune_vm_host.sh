#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

REMOTE_HOST="${LAYER_OS_VM_HOST:-97layer-vm}"
REMOTE_USER="${LAYER_OS_VM_USER:-}"
REMOTE_PORT="${LAYER_OS_VM_PORT:-22}"
REMOTE_SSH_KEY="${LAYER_OS_VM_SSH_KEY:-}"
KEEP_APP_RELEASES="${LAYER_OS_VM_KEEP_APP_RELEASES:-4}"
KEEP_WEB_RELEASES="${LAYER_OS_VM_KEEP_WEB_RELEASES:-4}"
JOURNAL_VACUUM_SIZE="${LAYER_OS_VM_JOURNAL_VACUUM_SIZE:-200M}"
CHECK_ONLY=0
SKIP_LEGACY_TRIM=0

usage() {
  cat >&2 <<'EOF'
usage: prune_vm_host.sh [--host <ssh-host>] [--user <remote-user>] [--port <port>] [--ssh-key <path>] [--keep-app <count>] [--keep-web <count>] [--journal-size <size>] [--skip-legacy-trim] [--check]

Prune a Layer OS VM while preserving the active always-on stack:
- keep current Layer OS services and runtimes
- optionally mask known legacy services/timers
- keep only the newest app/web release directories (plus current symlink target)
- clean apt cache, npm cache, and old npm logs
- vacuum systemd journal to a bounded size
EOF
}

die() {
  printf 'error: %s\n' "$1" >&2
  exit 1
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    die "required command not found: $1"
  fi
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
    --keep-app)
      [[ "$#" -ge 2 ]] || die "--keep-app requires a value"
      KEEP_APP_RELEASES="$2"
      shift 2
      ;;
    --keep-web)
      [[ "$#" -ge 2 ]] || die "--keep-web requires a value"
      KEEP_WEB_RELEASES="$2"
      shift 2
      ;;
    --journal-size)
      [[ "$#" -ge 2 ]] || die "--journal-size requires a value"
      JOURNAL_VACUUM_SIZE="$2"
      shift 2
      ;;
    --skip-legacy-trim)
      SKIP_LEGACY_TRIM=1
      shift
      ;;
    --check)
      CHECK_ONLY=1
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

require_cmd ssh

SSH_ARGS=(-p "${REMOTE_PORT}" -o StrictHostKeyChecking=accept-new)
if [[ -n "${REMOTE_SSH_KEY}" ]]; then
  SSH_ARGS+=(-i "${REMOTE_SSH_KEY}")
fi
REMOTE_TARGET="${REMOTE_HOST}"
if [[ -n "${REMOTE_USER}" ]]; then
  REMOTE_TARGET="${REMOTE_USER}@${REMOTE_HOST}"
fi

if [[ "${SKIP_LEGACY_TRIM}" -ne 1 ]]; then
  trim_cmd=("${SCRIPT_DIR}/vm_trim_legacy_runtime.sh" --host "${REMOTE_HOST}" --port "${REMOTE_PORT}")
  if [[ -n "${REMOTE_USER}" ]]; then
    trim_cmd+=(--user "${REMOTE_USER}")
  fi
  if [[ -n "${REMOTE_SSH_KEY}" ]]; then
    trim_cmd+=(--ssh-key "${REMOTE_SSH_KEY}")
  fi
  if [[ "${CHECK_ONLY}" -eq 1 ]]; then
    trim_cmd+=(--check)
  fi
  "${trim_cmd[@]}"
fi

ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "
  set -euo pipefail
  export CHECK_ONLY='${CHECK_ONLY}'
  export KEEP_APP_RELEASES='${KEEP_APP_RELEASES}'
  export KEEP_WEB_RELEASES='${KEEP_WEB_RELEASES}'
  export JOURNAL_VACUUM_SIZE='${JOURNAL_VACUUM_SIZE}'
  sudo CHECK_ONLY=\"${CHECK_ONLY}\" \
    KEEP_APP_RELEASES=\"${KEEP_APP_RELEASES}\" \
    KEEP_WEB_RELEASES=\"${KEEP_WEB_RELEASES}\" \
    JOURNAL_VACUUM_SIZE=\"${JOURNAL_VACUUM_SIZE}\" \
    python3 - <<'PY'
import os
import shutil
from pathlib import Path

check_only = os.environ['CHECK_ONLY'] == '1'
keep_app = int(os.environ['KEEP_APP_RELEASES'])
keep_web = int(os.environ['KEEP_WEB_RELEASES'])

def plan_release_prune(root_str: str, current_link_str: str, keep_count: int) -> None:
    root = Path(root_str)
    current_link = Path(current_link_str)
    current_target = None
    if current_link.exists() or current_link.is_symlink():
        try:
            current_target = current_link.resolve(strict=True)
        except FileNotFoundError:
            current_target = None
    dirs = sorted([path for path in root.iterdir() if path.is_dir()], key=lambda p: p.name) if root.exists() else []
    keep = set(dirs[-keep_count:]) if keep_count > 0 else set()
    if current_target is not None and current_target.parent == root:
        keep.add(current_target)
    delete = [path for path in dirs if path not in keep]

    print(f'== release prune {root} ==')
    print(f'current_target={current_target or \"none\"}')
    print(f'keep_count={keep_count}')
    print(f'total_dirs={len(dirs)}')
    for path in delete:
        print(f'prune={path}')
    if not delete:
        print('prune=none')
    if not check_only:
        for path in delete:
            shutil.rmtree(path)
    print()

plan_release_prune('/srv/layer-os/releases', '/srv/layer-os/current', keep_app)
plan_release_prune('/srv/layer-os/web/releases', '/srv/layer-os/web/current', keep_web)
PY

  echo '== cache cleanup =='
  if [[ \"${CHECK_ONLY}\" == '1' ]]; then
    du -sh \
      \"\$HOME/.npm\" \
      \"\$HOME/.npm/_cacache\" \
      \"\$HOME/.cache\" \
      \"\$HOME/.cache/pip\" \
      \"\$HOME/.cache/node-gyp\" \
      /var/cache/apt \
      /var/log/journal 2>/dev/null || true
  else
    if [[ -x \"\$HOME/.local/bin/npm\" ]]; then
      \"\$HOME/.local/bin/npm\" cache clean --force >/dev/null 2>&1 || true
    fi
    rm -rf \"\$HOME/.npm/_cacache\" 2>/dev/null || true
    rm -rf \"\$HOME/.npm/_logs\" 2>/dev/null || true
    rm -rf \"\$HOME/.cache/pip\" 2>/dev/null || true
    rm -rf \"\$HOME/.cache/node-gyp\" 2>/dev/null || true
    sudo apt-get clean >/dev/null 2>&1 || true
    sudo journalctl --vacuum-size=\"${JOURNAL_VACUUM_SIZE}\" >/dev/null 2>&1 || true
    du -sh \
      \"\$HOME/.npm\" \
      \"\$HOME/.cache\" \
      /var/cache/apt \
      /var/log/journal 2>/dev/null || true
  fi
  echo

  echo '== disk after =='
  df -h /
  echo
  sudo du -sh /srv/layer-os /srv/layer-os/web/releases /srv/layer-os/releases /srv/layer-os-dev 2>/dev/null || true
"
