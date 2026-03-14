#!/usr/bin/env bash
set -euo pipefail

REMOTE_HOST="${LAYER_OS_VM_HOST:-97layer-vm}"
REMOTE_USER="${LAYER_OS_VM_USER:-}"
REMOTE_PORT="${LAYER_OS_VM_PORT:-22}"
REMOTE_SSH_KEY="${LAYER_OS_VM_SSH_KEY:-}"
CHECK_ONLY=0

usage() {
  cat >&2 <<'EOF'
usage: vm_trim_legacy_runtime.sh [--host <ssh-host>] [--user <remote-user>] [--port <port>] [--ssh-key <path>] [--check]

Mask the known legacy runtime units on a VM and stop the orphaned cloudflared
tunnel that still points at the retired localhost:5001 stack.
EOF
}

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --host)
      [[ "$#" -ge 2 ]] || {
        usage
        exit 2
      }
      REMOTE_HOST="$2"
      shift 2
      ;;
    --user)
      [[ "$#" -ge 2 ]] || {
        usage
        exit 2
      }
      REMOTE_USER="$2"
      shift 2
      ;;
    --port)
      [[ "$#" -ge 2 ]] || {
        usage
        exit 2
      }
      REMOTE_PORT="$2"
      shift 2
      ;;
    --ssh-key)
      [[ "$#" -ge 2 ]] || {
        usage
        exit 2
      }
      REMOTE_SSH_KEY="$2"
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

SSH_ARGS=(-p "${REMOTE_PORT}" -o StrictHostKeyChecking=accept-new)
if [[ -n "${REMOTE_SSH_KEY}" ]]; then
  SSH_ARGS+=(-i "${REMOTE_SSH_KEY}")
fi
REMOTE_TARGET="${REMOTE_HOST}"
if [[ -n "${REMOTE_USER}" ]]; then
  REMOTE_TARGET="${REMOTE_USER}@${REMOTE_HOST}"
fi

ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "CHECK_ONLY=${CHECK_ONLY} bash -s" <<'EOF'
set -euo pipefail

BACKUP_DIR="${HOME}/layeros-legacy-unit-backup-$(date -u +%Y%m%d_%H%M%S)"
legacy_units=(
  97layer-code-agent.service
  97layer-ecosystem.service
  97layer-gardener.service
  97layer-nightguard.service
  97layeros-nightguard.service
  97layer-pip-audit.service
  97layer-pip-audit.timer
  97layer-telegram.service
  cortex-admin.service
  cortex-dashboard.service
  cortex-engine.service
  woohwahae-backend.service
  woohwahae-gateway.service
  woohwahae-ops-alert.service
  woohwahae-ops-alert.timer
)

cloudflared_pattern='^cloudflared tunnel --url http://localhost:5001$'

echo "== legacy units =="
for unit in "${legacy_units[@]}"; do
  enabled="$(systemctl is-enabled "$unit" 2>/dev/null || true)"
  active="$(systemctl is-active "$unit" 2>/dev/null || true)"
  printf '%s enabled=%s active=%s\n' "$unit" "${enabled:-unknown}" "${active:-unknown}"
done
echo

echo "== orphan tunnel =="
pgrep -af "$cloudflared_pattern" || echo "none"
echo

if [[ "${CHECK_ONLY}" == "1" ]]; then
  echo "backup_dir=${BACKUP_DIR}"
  echo "check only: no changes applied"
  exit 0
fi

mkdir -p "${BACKUP_DIR}"
for unit in "${legacy_units[@]}"; do
  sudo systemctl stop "$unit" 2>/dev/null || true
  sudo systemctl disable "$unit" 2>/dev/null || true
  sudo systemctl mask "$unit" 2>/dev/null || true
  if [[ -e "/etc/systemd/system/${unit}" ]]; then
    sudo cp -a "/etc/systemd/system/${unit}" "${BACKUP_DIR}/${unit}"
    sudo rm -f "/etc/systemd/system/${unit}"
  fi
done

sudo systemctl daemon-reload || true
sudo systemctl reset-failed || true

if pgrep -af "$cloudflared_pattern" >/dev/null 2>&1; then
  pkill -f "$cloudflared_pattern" || true
fi

echo "== post-cleanup legacy units =="
for unit in "${legacy_units[@]}"; do
  enabled="$(systemctl is-enabled "$unit" 2>/dev/null || true)"
  active="$(systemctl is-active "$unit" 2>/dev/null || true)"
  printf '%s enabled=%s active=%s\n' "$unit" "${enabled:-unknown}" "${active:-unknown}"
done
echo

echo "== post-cleanup orphan tunnel =="
pgrep -af "$cloudflared_pattern" || echo "none"
echo "backup_dir=${BACKUP_DIR}"
EOF
