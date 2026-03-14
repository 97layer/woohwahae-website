#!/usr/bin/env bash
set -euo pipefail

REMOTE_HOST="${LAYER_OS_VM_HOST:-97layer-vm}"
REMOTE_USER="${LAYER_OS_VM_TMUX_USER:-skyto5339_gmail_com}"
TMUX_SESSION="${LAYER_OS_VM_TMUX_SESSION:-layeros}"
MODE="attach"

usage() {
  cat >&2 <<'EOF'
usage: vm_tmux_attach.sh [--host <ssh-host>] [--session <name>] [--list|--ensure]

Attach to the canonical Layer OS tmux cockpit on the VM, or just ensure/list it.
EOF
}

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --host)
      REMOTE_HOST="$2"
      shift 2
      ;;
    --session)
      TMUX_SESSION="$2"
      shift 2
      ;;
    --list)
      MODE="list"
      shift
      ;;
    --ensure)
      MODE="ensure"
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

case "${MODE}" in
  attach)
    exec ssh -t "${REMOTE_HOST}" "sudo systemctl start 'layer-os-tmux@${REMOTE_USER}.service' >/dev/null 2>&1 || true; tmux attach -t '${TMUX_SESSION}'"
    ;;
  list)
    exec ssh "${REMOTE_HOST}" "sudo systemctl start 'layer-os-tmux@${REMOTE_USER}.service' >/dev/null 2>&1 || true; tmux list-windows -t '${TMUX_SESSION}' -F 'window=#{window_index}:#{window_name}'"
    ;;
  ensure)
    exec ssh "${REMOTE_HOST}" "sudo systemctl start 'layer-os-tmux@${REMOTE_USER}.service' >/dev/null 2>&1 || true; sudo -u '${REMOTE_USER}' /usr/local/bin/layer-os-tmux-bootstrap"
    ;;
esac
