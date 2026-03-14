#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

REMOTE_HOST="${LAYER_OS_VM_HOST:-97layer-vm}"
REMOTE_USER="${LAYER_OS_VM_DEV_USER:-skyto5339_gmail_com}"
REMOTE_DEV_ROOT="${LAYER_OS_VM_DEV_ROOT:-/srv/layer-os-dev}"
TMUX_SERVICE="layer-os-tmux@${REMOTE_USER}.service"
CHECK_ONLY=0
FORCE_SYNC=0

usage() {
  cat >&2 <<'EOF'
usage: install_vm_dev_workspace.sh [--host <ssh-host>] [--user <remote-user>] [--dev-root <path>] [--force-sync] [--check]

Bootstrap a real development checkout on the VM at /srv/layer-os-dev and refresh
the canonical tmux cockpit so mobile and remote edits can happen off the live release path.
EOF
}

note() {
  printf '%s\n' "$1"
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
    --dev-root)
      [[ "$#" -ge 2 ]] || die "--dev-root requires a value"
      REMOTE_DEV_ROOT="$2"
      shift 2
      ;;
    --force-sync)
      FORCE_SYNC=1
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
require_cmd rsync

note "remote_host=${REMOTE_HOST}"
note "remote_user=${REMOTE_USER}"
note "remote_dev_root=${REMOTE_DEV_ROOT}"

ssh "${REMOTE_HOST}" "
  set -euo pipefail
  sudo install -d -o '${REMOTE_USER}' -g '${REMOTE_USER}' '${REMOTE_DEV_ROOT}'
"

if [[ "${CHECK_ONLY}" -eq 1 ]]; then
  note "dev workspace check complete"
  exit 0
fi

if [[ "${FORCE_SYNC}" -ne 1 ]]; then
  remote_state="$(ssh "${REMOTE_HOST}" "set -euo pipefail; if [ -d '${REMOTE_DEV_ROOT}/.git' ] || [ -n \"\$(find '${REMOTE_DEV_ROOT}' -mindepth 1 -maxdepth 1 -print -quit 2>/dev/null)\" ]; then echo populated; else echo empty; fi")"
  if [[ "${remote_state}" != "empty" ]]; then
    die "remote dev root already has content; rerun with --force-sync once you are sure you want to refresh it"
  fi
fi

note "syncing local repo to ${REMOTE_HOST}:${REMOTE_DEV_ROOT}"
rsync -az \
  --delete \
  --exclude '.layer-os/' \
  --exclude '.layer-os-dev/' \
  --exclude '.omx/' \
  --exclude '.tmp/' \
  --exclude '/bin/' \
  --exclude 'docs/brand-home/.next/' \
  --exclude '.DS_Store' \
  "${ROOT_DIR}/" "${REMOTE_HOST}:${REMOTE_DEV_ROOT}/"

ssh "${REMOTE_HOST}" "
  set -euo pipefail
  sudo chown -R '${REMOTE_USER}:${REMOTE_USER}' '${REMOTE_DEV_ROOT}'
  sudo systemctl start '${TMUX_SERVICE}' >/dev/null 2>&1 || true
  sudo -u '${REMOTE_USER}' /usr/local/bin/layer-os-tmux-bootstrap >/dev/null
  sudo -u '${REMOTE_USER}' tmux list-windows -t layeros -F 'window=#{window_index}:#{window_name}'
  sudo -u '${REMOTE_USER}' git -C '${REMOTE_DEV_ROOT}' status --short --branch --untracked-files=no | sed -n '1,80p' || true
"

note "dev workspace ready on ${REMOTE_HOST}"
note "attach with: ./scripts/vm_tmux_attach.sh --host ${REMOTE_HOST}"
