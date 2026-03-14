#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

REMOTE_HOST="${LAYER_OS_VM_HOST:-97layer-vm}"
REMOTE_USER="${LAYER_OS_VM_TMUX_USER:-skyto5339_gmail_com}"
TMUX_SESSION="${LAYER_OS_VM_TMUX_SESSION:-layeros}"
TMUX_ROOT="${LAYER_OS_VM_TMUX_ROOT:-/srv/layer-os/current}"
TMUX_SHELL="${LAYER_OS_VM_TMUX_SHELL:-/bin/bash}"
DAEMON_SERVICE="${LAYER_OS_VM_TMUX_DAEMON_SERVICE:-layer-osd}"
WEB_SERVICE="${LAYER_OS_VM_TMUX_WEB_SERVICE:-layer-os-web}"
CHECK_ONLY=0

usage() {
  cat >&2 <<'EOF'
usage: install_vm_tmux.sh [--host <ssh-host>] [--user <remote-user>] [--session <name>] [--check]

Install tmux on the Layer OS VM, seed the canonical tmux cockpit bootstrap,
enable a boot-persistent systemd unit, and create the first session.
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
    --session)
      [[ "$#" -ge 2 ]] || die "--session requires a value"
      TMUX_SESSION="$2"
      shift 2
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
require_cmd bash

TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/layer-os-tmux-install.XXXXXX")"
cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

cat > "${TMP_DIR}/tmux.env" <<EOF
LAYER_OS_TMUX_SESSION=${TMUX_SESSION}
LAYER_OS_TMUX_ROOT=${TMUX_ROOT}
LAYER_OS_TMUX_SHELL=${TMUX_SHELL}
LAYER_OS_TMUX_DAEMON_SERVICE=${DAEMON_SERVICE}
LAYER_OS_TMUX_WEB_SERVICE=${WEB_SERVICE}
EOF

REMOTE_STAGE_DIR="/tmp/layer-os-tmux-install-$(date -u +%Y%m%d_%H%M%S)"

note "remote_host=${REMOTE_HOST}"
note "remote_user=${REMOTE_USER}"
note "tmux_session=${TMUX_SESSION}"
note "tmux_root=${TMUX_ROOT}"

if [[ "${CHECK_ONLY}" -eq 1 ]]; then
  note "tmux install check complete"
  exit 0
fi

ssh "${REMOTE_HOST}" "rm -rf '${REMOTE_STAGE_DIR}' && mkdir -p '${REMOTE_STAGE_DIR}'"
rsync -az \
  "${ROOT_DIR}/scripts/layer-os-tmux-bootstrap.sh" \
  "${ROOT_DIR}/scripts/systemd/layer-os-tmux@.service" \
  "${TMP_DIR}/tmux.env" \
  "${REMOTE_HOST}:${REMOTE_STAGE_DIR}/"

ssh "${REMOTE_HOST}" "
  set -euo pipefail
  sudo apt-get update -y
  sudo apt-get install -y tmux
  sudo install -d -o root -g root /usr/local/bin /etc/layer-os /etc/systemd/system
  sudo install -o root -g root -m 0755 '${REMOTE_STAGE_DIR}/layer-os-tmux-bootstrap.sh' '/usr/local/bin/layer-os-tmux-bootstrap'
  sudo install -o root -g root -m 0644 '${REMOTE_STAGE_DIR}/layer-os-tmux@.service' '/etc/systemd/system/layer-os-tmux@.service'
  sudo install -o root -g root -m 0644 '${REMOTE_STAGE_DIR}/tmux.env' '/etc/layer-os/tmux.env'
  sudo systemctl daemon-reload
  sudo systemctl enable --now 'layer-os-tmux@${REMOTE_USER}.service'
  sudo systemctl reload 'layer-os-tmux@${REMOTE_USER}.service' || true
  sudo -u '${REMOTE_USER}' tmux ls
  sudo -u '${REMOTE_USER}' tmux list-windows -t '${TMUX_SESSION}' -F 'window=#{window_index}:#{window_name}'
  rm -rf '${REMOTE_STAGE_DIR}'
"

note "tmux cockpit ready on ${REMOTE_HOST}"
note "attach with: ssh -t ${REMOTE_HOST} 'tmux attach -t ${TMUX_SESSION}'"
