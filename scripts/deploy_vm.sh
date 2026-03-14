#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

REMOTE_HOST="${LAYER_OS_VM_HOST:-97layer-vm}"
REMOTE_USER="${LAYER_OS_VM_USER:-}"
REMOTE_PORT="${LAYER_OS_VM_PORT:-22}"
REMOTE_SSH_KEY="${LAYER_OS_VM_SSH_KEY:-}"
REMOTE_BASE_DIR="${LAYER_OS_VM_BASE_DIR:-/srv/layer-os}"
REMOTE_RELEASES_DIR="${REMOTE_BASE_DIR}/releases"
REMOTE_CURRENT_LINK="${REMOTE_BASE_DIR}/current"
REMOTE_DATA_DIR="${LAYER_OS_VM_DATA_DIR:-/var/lib/layer-os}"
REMOTE_LOG_DIR="${LAYER_OS_VM_LOG_DIR:-/var/log/layer-os}"
REMOTE_ENV_DIR="${LAYER_OS_VM_ENV_DIR:-/etc/layer-os}"
SERVICE_NAME="${LAYER_OS_VM_SERVICE:-layer-osd}"
RELEASE_STAMP="$(date -u +%Y%m%d_%H%M%S)"
RESTART_SERVICE=1
CHECK_ONLY=0

usage() {
  cat >&2 <<'EOF'
usage: deploy_vm.sh [--host <ssh-host>] [--user <remote-user>] [--port <port>] [--ssh-key <path>] [--release-stamp <stamp>] [--check] [--no-restart]

Build a linux/amd64 Layer OS daemon bundle locally, sync it to the remote VM,
install the systemd unit if needed, and restart the remote layer-osd service.
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
    --release-stamp)
      [[ "$#" -ge 2 ]] || die "--release-stamp requires a value"
      RELEASE_STAMP="$2"
      shift 2
      ;;
    --check)
      CHECK_ONLY=1
      shift
      ;;
    --no-restart)
      RESTART_SERVICE=0
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

require_cmd go
require_cmd rsync
require_cmd ssh

TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/layer-os-vm-deploy.XXXXXX")"
cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

LOCAL_BIN_DIR="${TMP_DIR}/bin"
REMOTE_STAGE_DIR="/tmp/layer-os-release-${RELEASE_STAMP}"
REMOTE_RELEASE_DIR="${REMOTE_RELEASES_DIR}/${RELEASE_STAMP}"
SSH_ARGS=(-p "${REMOTE_PORT}" -o StrictHostKeyChecking=accept-new)
if [[ -n "${REMOTE_SSH_KEY}" ]]; then
  SSH_ARGS+=(-i "${REMOTE_SSH_KEY}")
fi
REMOTE_TARGET="${REMOTE_HOST}"
if [[ -n "${REMOTE_USER}" ]]; then
  REMOTE_TARGET="${REMOTE_USER}@${REMOTE_HOST}"
fi
RSYNC_SSH=(ssh -p "${REMOTE_PORT}" -o StrictHostKeyChecking=accept-new)
if [[ -n "${REMOTE_SSH_KEY}" ]]; then
  RSYNC_SSH+=(-i "${REMOTE_SSH_KEY}")
fi
printf -v RSYNC_SSH_COMMAND '%q ' "${RSYNC_SSH[@]}"
RSYNC_SSH_COMMAND="${RSYNC_SSH_COMMAND% }"

mkdir -p "${LOCAL_BIN_DIR}"

build_binary() {
  note "building linux/amd64 binaries"
  (
    cd "${ROOT_DIR}"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "${LOCAL_BIN_DIR}/layer-osd" ./cmd/layer-osd
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "${LOCAL_BIN_DIR}/layer-osctl" ./cmd/layer-osctl
  )
}

write_release_manifest() {
  local revision="unknown"
  if command -v git >/dev/null 2>&1; then
    revision="$(git -C "${ROOT_DIR}" rev-parse --short HEAD 2>/dev/null || printf 'unknown')"
  fi
  mkdir -p "${TMP_DIR}/.layer-os"
  cat > "${TMP_DIR}/.layer-os/RELEASE_MANIFEST.txt" <<EOF
release_stamp=${RELEASE_STAMP}
revision=${revision}
built_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)
source_root=${ROOT_DIR}
EOF
}

sync_bundle() {
  note "syncing release bundle to ${REMOTE_HOST}:${REMOTE_STAGE_DIR}"
  ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "rm -rf '${REMOTE_STAGE_DIR}' && mkdir -p '${REMOTE_STAGE_DIR}'"
  rsync -az --delete -e "${RSYNC_SSH_COMMAND}" \
    --exclude '.git/' \
    --exclude '.cache/' \
    --exclude '.layer-os/' \
    --exclude '.layer-os-dev/' \
    --exclude '.omx/' \
    --exclude '.tmp/' \
    --exclude '/bin/' \
    --exclude '/knowledge/' \
    --exclude '/임시.md' \
    --exclude 'docs/brand-home/node_modules/' \
    --exclude 'docs/brand-home/.next/' \
    --exclude 'output/' \
    --exclude '.DS_Store' \
    "${ROOT_DIR}/" "${REMOTE_TARGET}:${REMOTE_STAGE_DIR}/"
  ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "mkdir -p '${REMOTE_STAGE_DIR}/.layer-os'"
  rsync -az -e "${RSYNC_SSH_COMMAND}" "${LOCAL_BIN_DIR}/" "${REMOTE_TARGET}:${REMOTE_STAGE_DIR}/bin/"
  rsync -az -e "${RSYNC_SSH_COMMAND}" "${TMP_DIR}/.layer-os/RELEASE_MANIFEST.txt" "${REMOTE_TARGET}:${REMOTE_STAGE_DIR}/.layer-os/RELEASE_MANIFEST.txt"
}

install_remote_release() {
  note "installing remote release ${REMOTE_RELEASE_DIR}"
  ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "
    set -euo pipefail
    sudo install -d -o layeros -g layeros '${REMOTE_RELEASES_DIR}' '${REMOTE_DATA_DIR}' '${REMOTE_LOG_DIR}' '${REMOTE_ENV_DIR}'
    sudo rm -rf '${REMOTE_RELEASE_DIR}'
    sudo mv '${REMOTE_STAGE_DIR}' '${REMOTE_RELEASE_DIR}'
    sudo chown -R layeros:layeros '${REMOTE_RELEASE_DIR}'
    sudo ln -sfn '${REMOTE_RELEASE_DIR}' '${REMOTE_CURRENT_LINK}'
    if [ ! -f '${REMOTE_ENV_DIR}/layer-osd.env' ]; then
      sudo install -o layeros -g layeros -m 0640 '${REMOTE_CURRENT_LINK}/scripts/systemd/layer-osd.env.example' '${REMOTE_ENV_DIR}/layer-osd.env'
    fi
    if [ ! -f '${REMOTE_ENV_DIR}/providers.env' ]; then
      sudo install -o layeros -g layeros -m 0640 '${REMOTE_CURRENT_LINK}/scripts/systemd/providers.env.example' '${REMOTE_ENV_DIR}/providers.env'
    fi
    sudo install -o root -g root -m 0644 '${REMOTE_CURRENT_LINK}/scripts/systemd/layer-osd.service' '/etc/systemd/system/${SERVICE_NAME}.service'
    sudo systemctl daemon-reload
    sudo systemctl enable '${SERVICE_NAME}.service'
  "
}

restart_remote_service() {
  if [[ "${RESTART_SERVICE}" -ne 1 ]]; then
    note "skipping remote service restart (--no-restart)"
    return 0
  fi
  note "restarting ${SERVICE_NAME} on ${REMOTE_HOST}"
  ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "sudo systemctl restart '${SERVICE_NAME}.service'"
}

verify_remote_health() {
  note "verifying remote daemon health"
  ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "
    set -euo pipefail
    for _ in \$(seq 1 30); do
      if command -v curl >/dev/null 2>&1; then
        if curl -fsS 'http://127.0.0.1:17808/healthz'; then
          exit 0
        fi
      elif command -v wget >/dev/null 2>&1; then
        if wget -qO- 'http://127.0.0.1:17808/healthz'; then
          exit 0
        fi
      else
        sudo systemctl is-active '${SERVICE_NAME}.service' >/dev/null
        exit 0
      fi
      sleep 2
    done
    sudo systemctl status --no-pager '${SERVICE_NAME}.service' || true
    sudo journalctl -u '${SERVICE_NAME}.service' -n 40 --no-pager || true
    if command -v curl >/dev/null 2>&1; then
      curl -fsS 'http://127.0.0.1:17808/healthz'
    elif command -v wget >/dev/null 2>&1; then
      wget -qO- 'http://127.0.0.1:17808/healthz'
    else
      sudo systemctl is-active '${SERVICE_NAME}.service'
    fi
  "
}

build_binary
write_release_manifest

note "release_stamp=${RELEASE_STAMP}"
note "remote_host=${REMOTE_HOST}"
note "remote_release_dir=${REMOTE_RELEASE_DIR}"

if [[ "${CHECK_ONLY}" -eq 1 ]]; then
  note "deploy check complete"
  exit 0
fi

sync_bundle
install_remote_release
restart_remote_service
verify_remote_health

note "remote current -> ${REMOTE_RELEASE_DIR}"
note "service=${SERVICE_NAME}"
