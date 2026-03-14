#!/usr/bin/env bash
set -euo pipefail

REMOTE_HOST="${LAYER_OS_VM_HOST:-97layer-vm}"
REMOTE_USER="${LAYER_OS_VM_USER:-}"
REMOTE_PORT="${LAYER_OS_VM_PORT:-22}"
REMOTE_SSH_KEY="${LAYER_OS_VM_SSH_KEY:-}"
REMOTE_ENV_PATH="${LAYER_OS_VM_PROVIDER_ENV_PATH:-/etc/layer-os/providers.env}"
REMOTE_SERVICE_NAME="${LAYER_OS_VM_SERVICE:-layer-osd}"
REMOTE_CHECK_PATH="${LAYER_OS_VM_PROVIDER_CHECK_PATH:-}"
APPLY=0
SELECTED_KEYS=()

KNOWN_KEYS=(
  OPENAI_API_KEY
  ANTHROPIC_API_KEY
  GOOGLE_API_KEY
  TELEGRAM_BOT_TOKEN
  TELEGRAM_FOUNDER_CHAT_ID
  TELEGRAM_FOUNDER_DM_CHAT_ID
  TELEGRAM_OPS_CHAT_ID
  TELEGRAM_BRAND_CHAT_ID
  TELEGRAM_CHAT_ID
  THREADS_ACCESS_TOKEN
)

usage() {
  cat >&2 <<'EOF'
usage: seed_vm_providers.sh [--host <ssh-host>] [--user <remote-user>] [--port <port>] [--ssh-key <path>] [--check-path <daemon-route>] [--keys <comma-separated-keys>] [--apply]

Check or seed provider secrets for the VM continuity host without printing the
secret values themselves.

Default mode prints local/remote presence only.
Use --apply to push any locally available secrets to the VM provider env file
and restart layer-osd. The post-apply probe defaults to the most relevant
runtime surface for the selected keys.
EOF
}

note() {
  printf '%s\n' "$1"
}

die() {
  printf 'error: %s\n' "$1" >&2
  exit 1
}

local_secret() {
  local key="$1"
  local value="${!key:-}"
  if [[ -n "${value}" ]]; then
    printf '%s' "${value}"
    return 0
  fi
  if command -v security >/dev/null 2>&1; then
    value="$(security find-generic-password -a layer-os -s "${key}" -w 2>/dev/null || true)"
    if [[ -n "${value}" ]]; then
      printf '%s' "${value}"
      return 0
    fi
  fi
  if [[ "${key}" == "TELEGRAM_FOUNDER_CHAT_ID" ]]; then
    value="${TELEGRAM_CHAT_ID:-}"
    if [[ -n "${value}" ]]; then
      printf '%s' "${value}"
      return 0
    fi
    if command -v security >/dev/null 2>&1; then
      value="$(security find-generic-password -a layer-os -s "TELEGRAM_CHAT_ID" -w 2>/dev/null || true)"
      if [[ -n "${value}" ]]; then
        printf '%s' "${value}"
        return 0
      fi
    fi
  fi
  return 1
}

selected_key() {
  local key="$1"
  local item
  if [[ "${#SELECTED_KEYS[@]}" -eq 0 ]]; then
    return 0
  fi
  for item in "${SELECTED_KEYS[@]}"; do
    if [[ "${item}" == "${key}" ]]; then
      return 0
    fi
  done
  return 1
}

selected_any_key() {
  local item
  for item in "$@"; do
    if selected_key "${item}"; then
      return 0
    fi
  done
  return 1
}

post_apply_check_path() {
  if [[ -n "${REMOTE_CHECK_PATH}" ]]; then
    printf '%s' "${REMOTE_CHECK_PATH}"
    return 0
  fi
  if selected_any_key THREADS_ACCESS_TOKEN; then
    printf '/api/layer-os/social/threads'
    return 0
  fi
  if selected_any_key \
    TELEGRAM_BOT_TOKEN \
    TELEGRAM_FOUNDER_CHAT_ID \
    TELEGRAM_FOUNDER_DM_CHAT_ID \
    TELEGRAM_OPS_CHAT_ID \
    TELEGRAM_BRAND_CHAT_ID \
    TELEGRAM_CHAT_ID \
    GOOGLE_API_KEY; then
    printf '/api/layer-os/telegram'
    return 0
  fi
  printf '/api/layer-os/status'
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
    --check-path)
      [[ "$#" -ge 2 ]] || die "--check-path requires a value"
      REMOTE_CHECK_PATH="$2"
      shift 2
      ;;
    --keys)
      [[ "$#" -ge 2 ]] || die "--keys requires a value"
      IFS=',' read -r -a SELECTED_KEYS <<< "$2"
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

TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/layer-os-vm-providers.XXXXXX")"
cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

LOCAL_ENV_FILE="${TMP_DIR}/providers.local.env"
REMOTE_ENV_FILE="${TMP_DIR}/providers.remote.env"
MERGED_ENV_FILE="${TMP_DIR}/providers.merged.env"
REMOTE_STAGE_FILE="/tmp/layer-os-providers.env"

touch "${LOCAL_ENV_FILE}" "${REMOTE_ENV_FILE}" "${MERGED_ENV_FILE}"
chmod 600 "${LOCAL_ENV_FILE}" "${REMOTE_ENV_FILE}" "${MERGED_ENV_FILE}"

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

local_present=0
note "== provider readiness =="
for key in "${KNOWN_KEYS[@]}"; do
  if ! selected_key "${key}"; then
    continue
  fi
  if value="$(local_secret "${key}")"; then
    printf '%s=%s\n' "${key}" "${value}" >> "${LOCAL_ENV_FILE}"
    printf '%s local=present\n' "${key}"
    local_present=$((local_present + 1))
  else
    printf '%s local=missing\n' "${key}"
  fi
done

ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "sudo test -f '${REMOTE_ENV_PATH}' && sudo cat '${REMOTE_ENV_PATH}' || true" > "${REMOTE_ENV_FILE}"

for key in "${KNOWN_KEYS[@]}"; do
  if ! selected_key "${key}"; then
    continue
  fi
  if grep -q "^${key}=" "${REMOTE_ENV_FILE}"; then
    printf '%s remote=present\n' "${key}"
  else
    printf '%s remote=missing\n' "${key}"
  fi
done

if [[ "${APPLY}" -ne 1 ]]; then
  note "mode=check"
  exit 0
fi

if [[ "${local_present}" -eq 0 ]]; then
  die "no local provider secrets found; set env vars or macOS Keychain items first"
fi

cp "${REMOTE_ENV_FILE}" "${MERGED_ENV_FILE}"
for key in "${KNOWN_KEYS[@]}"; do
  if ! selected_key "${key}"; then
    continue
  fi
  sed -i.bak "/^${key}=/d" "${MERGED_ENV_FILE}"
done
rm -f "${MERGED_ENV_FILE}.bak"
cat "${LOCAL_ENV_FILE}" >> "${MERGED_ENV_FILE}"

if grep -q '^TELEGRAM_FOUNDER_CHAT_ID=' "${MERGED_ENV_FILE}" && grep -q '^TELEGRAM_CHAT_ID=' "${MERGED_ENV_FILE}"; then
  founder_value="$(grep '^TELEGRAM_FOUNDER_CHAT_ID=' "${MERGED_ENV_FILE}" | tail -n 1 | cut -d= -f2-)"
  legacy_value="$(grep '^TELEGRAM_CHAT_ID=' "${MERGED_ENV_FILE}" | tail -n 1 | cut -d= -f2-)"
  if [[ -n "${founder_value}" && "${founder_value}" == "${legacy_value}" ]]; then
    sed -i.bak '/^TELEGRAM_CHAT_ID=/d' "${MERGED_ENV_FILE}"
    rm -f "${MERGED_ENV_FILE}.bak"
  fi
fi

rsync -az -e "${RSYNC_SSH_COMMAND}" "${MERGED_ENV_FILE}" "${REMOTE_TARGET}:${REMOTE_STAGE_FILE}"
CHECK_PATH="$(post_apply_check_path)"
note "post_apply_check=${CHECK_PATH}"
ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "
  set -euo pipefail
  sudo install -o layeros -g layeros -m 0640 '${REMOTE_STAGE_FILE}' '${REMOTE_ENV_PATH}'
  rm -f '${REMOTE_STAGE_FILE}'
  sudo systemctl restart '${REMOTE_SERVICE_NAME}.service'
  for _ in \$(seq 1 30); do
    if curl -fsS 'http://127.0.0.1:17808/healthz' >/dev/null; then
      break
    fi
    sleep 2
  done
  curl -fsS 'http://127.0.0.1:17808${CHECK_PATH}'
"
