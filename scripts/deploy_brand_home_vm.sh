#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
BRAND_HOME_DIR="${ROOT_DIR}/docs/brand-home"
LOCAL_WEB_SERVICE_FILE="${ROOT_DIR}/scripts/systemd/layer-os-web.service"
LOCAL_WEB_ENV_EXAMPLE="${ROOT_DIR}/scripts/systemd/layer-os-web.env.example"
LOCAL_WEB_NGINX_EXAMPLE="${ROOT_DIR}/scripts/nginx/layer-os-web.local.conf.example"
LOCAL_NEXT_DIR="${BRAND_HOME_DIR}/.next"
LOCAL_STANDALONE_DIR=""
LOCAL_STATIC_DIR=""

REMOTE_HOST="${LAYER_OS_WEB_VM_HOST:-97layer-vm}"
REMOTE_USER="${LAYER_OS_WEB_VM_USER:-}"
REMOTE_PORT="${LAYER_OS_WEB_VM_PORT:-22}"
REMOTE_SSH_KEY="${LAYER_OS_WEB_VM_SSH_KEY:-}"
REMOTE_BASE_DIR="${LAYER_OS_WEB_VM_BASE_DIR:-/srv/layer-os/web}"
REMOTE_RELEASES_DIR="${REMOTE_BASE_DIR}/releases"
REMOTE_CURRENT_LINK="${REMOTE_BASE_DIR}/current"
REMOTE_ENV_DIR="${LAYER_OS_WEB_VM_ENV_DIR:-/etc/layer-os}"
REMOTE_NODE_DIR="${LAYER_OS_WEB_VM_NODE_DIR:-/srv/layer-os/node}"
REMOTE_SERVICE_NAME="${LAYER_OS_WEB_VM_SERVICE:-layer-os-web}"
REMOTE_LOG_DIR="${LAYER_OS_WEB_VM_LOG_DIR:-/var/log/layer-os}"
REMOTE_WEB_PORT="${LAYER_OS_WEB_PORT:-3081}"
NODE_VERSION="${LAYER_OS_WEB_NODE_VERSION:-$(node -p 'process.versions.node')}"
RELEASE_STAMP="$(date -u +%Y%m%d_%H%M%S)"
CHECK_ONLY=0
RESTART_SERVICE=1
NEEDS_ENV_SEED=1
REMOTE_NODE_INSTALLED=0

usage() {
  cat >&2 <<'EOF'
usage: deploy_brand_home_vm.sh [--host <ssh-host>] [--user <remote-user>] [--port <port>] [--ssh-key <path>] [--node-version <version>] [--release-stamp <stamp>] [--check] [--no-restart]

Build the founder/admin web as a Next standalone bundle, sync it to the VM,
install a matching Linux Node runtime, and restart the remote web service.
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

sha256_hex() {
  local value="$1"
  if command -v shasum >/dev/null 2>&1; then
    printf '%s' "${value}" | shasum -a 256 | awk '{print $1}'
    return
  fi
  if command -v openssl >/dev/null 2>&1; then
    printf '%s' "${value}" | openssl dgst -sha256 -binary | xxd -p -c 256
    return
  fi
  die "need shasum or openssl to hash the admin password seed"
}

rand_hex() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 32
    return
  fi
  if command -v python3 >/dev/null 2>&1; then
    python3 - <<'PY'
import secrets
print(secrets.token_hex(32))
PY
    return
  fi
  die "need openssl or python3 to generate a session secret"
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
    --node-version)
      [[ "$#" -ge 2 ]] || die "--node-version requires a value"
      NODE_VERSION="$2"
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

require_cmd node
require_cmd npm
require_cmd curl
require_cmd rsync
require_cmd ssh
require_cmd tar

TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/layer-os-web-deploy.XXXXXX")"
cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

REMOTE_STAGE_DIR="/tmp/layer-os-web-release-${RELEASE_STAMP}"
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
NODE_DIST="node-v${NODE_VERSION}-linux-x64"
NODE_TARBALL="${NODE_DIST}.tar.xz"
NODE_URL="https://nodejs.org/dist/v${NODE_VERSION}/${NODE_TARBALL}"
LOCAL_NODE_TARBALL="${TMP_DIR}/${NODE_TARBALL}"
SESSION_SECRET="${SESSION_HMAC_SECRET:-}"
WRITE_TOKEN="${LAYER_OS_WRITE_TOKEN:-}"
ADMIN_PASSWORD_SHA256="${LAYER_OS_ADMIN_PASSWORD_SHA256:-}"

detect_remote_env() {
  if [[ "${CHECK_ONLY}" -eq 1 ]]; then
    NEEDS_ENV_SEED=0
    return 0
  fi
  if ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "test -f '${REMOTE_ENV_DIR}/layer-os-web.env'"; then
    NEEDS_ENV_SEED=0
    note "remote web env already exists; keeping existing write token and admin password seed"
    return 0
  fi
  NEEDS_ENV_SEED=1
}

detect_remote_node_runtime() {
  if [[ "${CHECK_ONLY}" -eq 1 ]]; then
    return 0
  fi
  if ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "test -x '${REMOTE_NODE_DIR}/${NODE_DIST}/bin/node'"; then
    REMOTE_NODE_INSTALLED=1
    note "remote node runtime v${NODE_VERSION} already exists; reusing it"
  fi
}

resolve_env_seed() {
  if [[ "${NEEDS_ENV_SEED}" -ne 1 ]]; then
    return 0
  fi
  if [[ -z "${SESSION_SECRET}" ]]; then
    SESSION_SECRET="$(rand_hex)"
  fi
  if [[ -z "${WRITE_TOKEN}" ]]; then
    WRITE_TOKEN="$(security find-generic-password -a layer-os -s LAYER_OS_WRITE_TOKEN -w 2>/dev/null || true)"
  fi
  if [[ -z "${WRITE_TOKEN}" ]]; then
    die "LAYER_OS_WRITE_TOKEN is required (or must exist in macOS Keychain) on first web boot so the admin surface can write through Layer OS"
  fi
  if [[ -z "${ADMIN_PASSWORD_SHA256}" ]]; then
    ADMIN_PASSWORD_SHA256="$(sha256_hex "${WRITE_TOKEN}")"
  fi
}

build_web() {
  note "building standalone founder/admin web"
  (
    cd "${BRAND_HOME_DIR}"
    chmod -R u+w .next 2>/dev/null || true
    find .next -name '.DS_Store' -delete 2>/dev/null || true
    rm -rf .next 2>/dev/null || true
    if [[ -e .next ]]; then
      node -e "require('node:fs').rmSync('.next', { recursive: true, force: true })"
    fi
    NEXT_TELEMETRY_DISABLED=1 npm run build
  )
  resolve_local_build_artifacts
}

download_node_runtime() {
  if [[ "${CHECK_ONLY}" -eq 1 ]]; then
    return 0
  fi
  if [[ "${REMOTE_NODE_INSTALLED}" -eq 1 ]]; then
    return 0
  fi
  note "downloading linux node runtime v${NODE_VERSION}"
  curl -fsSLo "${LOCAL_NODE_TARBALL}" "${NODE_URL}"
}

resolve_local_build_artifacts() {
  LOCAL_STANDALONE_DIR=""
  LOCAL_STATIC_DIR=""

  if [[ -f "${LOCAL_NEXT_DIR}/standalone/server.js" ]]; then
    LOCAL_STANDALONE_DIR="${LOCAL_NEXT_DIR}/standalone"
  fi

  if [[ -d "${LOCAL_NEXT_DIR}/static" ]]; then
    LOCAL_STATIC_DIR="${LOCAL_NEXT_DIR}/static"
  elif [[ -n "${LOCAL_STANDALONE_DIR}" && -d "${LOCAL_STANDALONE_DIR}/.next/static" ]]; then
    LOCAL_STATIC_DIR="${LOCAL_STANDALONE_DIR}/.next/static"
  fi

  if [[ -z "${LOCAL_STANDALONE_DIR}" ]]; then
    die "standalone output missing at ${LOCAL_NEXT_DIR}/standalone; confirm docs/brand-home next.config.mjs uses output=standalone"
  fi
  if [[ -z "${LOCAL_STATIC_DIR}" ]]; then
    die "static output missing under ${LOCAL_NEXT_DIR}; cannot stage founder/admin assets"
  fi
}

write_release_manifest() {
  mkdir -p "${TMP_DIR}/.layer-os"
  cat > "${TMP_DIR}/.layer-os/WEB_RELEASE_MANIFEST.txt" <<EOF
release_stamp=${RELEASE_STAMP}
node_version=${NODE_VERSION}
remote_host=${REMOTE_HOST}
remote_release_dir=${REMOTE_RELEASE_DIR}
web_port=${REMOTE_WEB_PORT}
built_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)
EOF
}

write_env_seed() {
  if [[ "${NEEDS_ENV_SEED}" -ne 1 ]]; then
    return 0
  fi
  cat > "${TMP_DIR}/layer-os-web.env.seed" <<EOF
NODE_ENV=production
HOSTNAME=127.0.0.1
PORT=${REMOTE_WEB_PORT}
LAYER_OS_BASE_URL=http://127.0.0.1:17808
LAYER_OS_LOCAL_ADMIN_BOOTSTRAP=false
SESSION_HMAC_SECRET=${SESSION_SECRET}
LAYER_OS_WRITE_TOKEN=${WRITE_TOKEN}
LAYER_OS_ADMIN_PASSWORD_SHA256=${ADMIN_PASSWORD_SHA256}
EOF
}

sync_bundle() {
  note "syncing founder/admin web bundle to ${REMOTE_HOST}:${REMOTE_STAGE_DIR}"
  [[ -n "${LOCAL_STANDALONE_DIR}" ]] || die "local standalone path is unresolved"
  [[ -n "${LOCAL_STATIC_DIR}" ]] || die "local static path is unresolved"
  ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "rm -rf '${REMOTE_STAGE_DIR}' && mkdir -p '${REMOTE_STAGE_DIR}/.next' '${REMOTE_STAGE_DIR}/public'"
  rsync -az --delete -e "${RSYNC_SSH_COMMAND}" "${LOCAL_STANDALONE_DIR}/" "${REMOTE_TARGET}:${REMOTE_STAGE_DIR}/"
  rsync -az --delete -e "${RSYNC_SSH_COMMAND}" "${LOCAL_STATIC_DIR}/" "${REMOTE_TARGET}:${REMOTE_STAGE_DIR}/.next/static/"
  rsync -az --delete -e "${RSYNC_SSH_COMMAND}" "${BRAND_HOME_DIR}/public/" "${REMOTE_TARGET}:${REMOTE_STAGE_DIR}/public/"
  ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "mkdir -p '${REMOTE_STAGE_DIR}/.layer-os' '${REMOTE_STAGE_DIR}/deploy-assets'"
  if [[ -f "${LOCAL_NODE_TARBALL}" ]]; then
    rsync -az -e "${RSYNC_SSH_COMMAND}" "${LOCAL_NODE_TARBALL}" "${REMOTE_TARGET}:${REMOTE_STAGE_DIR}/${NODE_TARBALL}"
  fi
  rsync -az -e "${RSYNC_SSH_COMMAND}" "${TMP_DIR}/.layer-os/WEB_RELEASE_MANIFEST.txt" "${REMOTE_TARGET}:${REMOTE_STAGE_DIR}/.layer-os/WEB_RELEASE_MANIFEST.txt"
  if [[ -f "${TMP_DIR}/layer-os-web.env.seed" ]]; then
    rsync -az -e "${RSYNC_SSH_COMMAND}" "${TMP_DIR}/layer-os-web.env.seed" "${REMOTE_TARGET}:${REMOTE_STAGE_DIR}/layer-os-web.env.seed"
  fi
  rsync -az -e "${RSYNC_SSH_COMMAND}" "${LOCAL_WEB_SERVICE_FILE}" "${REMOTE_TARGET}:${REMOTE_STAGE_DIR}/deploy-assets/layer-os-web.service"
  rsync -az -e "${RSYNC_SSH_COMMAND}" "${LOCAL_WEB_ENV_EXAMPLE}" "${REMOTE_TARGET}:${REMOTE_STAGE_DIR}/deploy-assets/layer-os-web.env.example"
  rsync -az -e "${RSYNC_SSH_COMMAND}" "${LOCAL_WEB_NGINX_EXAMPLE}" "${REMOTE_TARGET}:${REMOTE_STAGE_DIR}/deploy-assets/layer-os-web.local.conf.example"
}

install_remote_release() {
  note "installing remote founder/admin web release ${REMOTE_RELEASE_DIR}"
  ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "
    set -euo pipefail
    sudo install -d -o layeros -g layeros '${REMOTE_RELEASES_DIR}' '${REMOTE_NODE_DIR}' '${REMOTE_ENV_DIR}' '${REMOTE_LOG_DIR}'
    sudo rm -rf '${REMOTE_RELEASE_DIR}'
    sudo mv '${REMOTE_STAGE_DIR}' '${REMOTE_RELEASE_DIR}'
    sudo chown -R layeros:layeros '${REMOTE_RELEASE_DIR}'
    if [ ! -d '${REMOTE_NODE_DIR}/${NODE_DIST}' ]; then
      [ -f '${REMOTE_RELEASE_DIR}/${NODE_TARBALL}' ] || {
        echo 'missing node runtime tarball for first web boot' >&2
        exit 1
      }
      sudo tar -xJf '${REMOTE_RELEASE_DIR}/${NODE_TARBALL}' -C '${REMOTE_NODE_DIR}'
      sudo chown -R layeros:layeros '${REMOTE_NODE_DIR}/${NODE_DIST}'
    fi
    sudo ln -sfn '${REMOTE_NODE_DIR}/${NODE_DIST}' '${REMOTE_NODE_DIR}/current'
    sudo ln -sfn '${REMOTE_RELEASE_DIR}' '${REMOTE_CURRENT_LINK}'
    if [ ! -f '${REMOTE_ENV_DIR}/layer-os-web.env' ]; then
      [ -f '${REMOTE_RELEASE_DIR}/layer-os-web.env.seed' ] || {
        echo 'missing layer-os-web.env.seed for first boot' >&2
        exit 1
      }
      sudo install -o layeros -g layeros -m 0640 '${REMOTE_RELEASE_DIR}/layer-os-web.env.seed' '${REMOTE_ENV_DIR}/layer-os-web.env'
    fi
    sudo install -o root -g root -m 0644 '${REMOTE_RELEASE_DIR}/deploy-assets/layer-os-web.service' '/etc/systemd/system/${REMOTE_SERVICE_NAME}.service'
    sudo systemctl daemon-reload
    sudo systemctl enable '${REMOTE_SERVICE_NAME}.service'
  "
}

restart_remote_service() {
  if [[ "${RESTART_SERVICE}" -ne 1 ]]; then
    note "skipping remote web service restart (--no-restart)"
    return 0
  fi
  note "restarting ${REMOTE_SERVICE_NAME} on ${REMOTE_HOST}"
  ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "sudo systemctl restart '${REMOTE_SERVICE_NAME}.service'"
}

verify_remote_health() {
  note "verifying remote founder/admin web health"
  ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "
    set -euo pipefail
    for _ in \$(seq 1 30); do
      if curl -fsS 'http://127.0.0.1:${REMOTE_WEB_PORT}/admin/login' >/dev/null && \
         curl -fsS 'http://127.0.0.1:${REMOTE_WEB_PORT}/api/public/proof' >/dev/null; then
        sudo systemctl is-active '${REMOTE_SERVICE_NAME}.service' >/dev/null
        exit 0
      fi
      sleep 2
    done
    sudo systemctl status --no-pager '${REMOTE_SERVICE_NAME}.service' || true
    sudo journalctl -u '${REMOTE_SERVICE_NAME}.service' -n 40 --no-pager || true
    curl -fsS 'http://127.0.0.1:${REMOTE_WEB_PORT}/admin/login' >/dev/null
    curl -fsS 'http://127.0.0.1:${REMOTE_WEB_PORT}/api/public/proof' >/dev/null
    sudo systemctl is-active '${REMOTE_SERVICE_NAME}.service'
  "
}

build_web
detect_remote_env
detect_remote_node_runtime
download_node_runtime
resolve_env_seed
write_release_manifest
write_env_seed

note "release_stamp=${RELEASE_STAMP}"
note "remote_host=${REMOTE_HOST}"
note "remote_release_dir=${REMOTE_RELEASE_DIR}"
note "node_version=${NODE_VERSION}"

if [[ "${CHECK_ONLY}" -eq 1 ]]; then
  note "web deploy check complete"
  exit 0
fi

sync_bundle
install_remote_release
restart_remote_service
verify_remote_health

note "remote current -> ${REMOTE_RELEASE_DIR}"
note "service=${REMOTE_SERVICE_NAME}"
