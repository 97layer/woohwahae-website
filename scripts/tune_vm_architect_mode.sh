#!/usr/bin/env bash
set -euo pipefail

REMOTE_HOST="${LAYER_OS_VM_HOST:-97layer-vm}"
REMOTE_ENV_PATH="${LAYER_OS_VM_ENV_PATH:-/etc/layer-os/layer-osd.env}"
REMOTE_SERVICE_NAME="${LAYER_OS_VM_SERVICE:-layer-osd}"
MODE="observe"
CHECK_ONLY=0

usage() {
  cat >&2 <<'EOF'
usage: tune_vm_architect_mode.sh [--host <ssh-host>] [--mode <observe|active>] [--check]

Tune the always-on VM architect loop posture.

Modes:
  observe  -> LAYER_OS_ARCHITECT_AUTODISPATCH=false, LAYER_OS_ARCHITECT_AUTOVERIFY=false
  active   -> LAYER_OS_ARCHITECT_AUTODISPATCH=true,  LAYER_OS_ARCHITECT_AUTOVERIFY=true
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
    --mode)
      [[ "$#" -ge 2 ]] || die "--mode requires a value"
      MODE="$2"
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

case "${MODE}" in
  observe)
    AUTODISPATCH="false"
    AUTOVERIFY="false"
    ;;
  active)
    AUTODISPATCH="true"
    AUTOVERIFY="true"
    ;;
  *)
    die "unsupported mode: ${MODE}"
    ;;
esac

printf 'remote_host=%s\n' "${REMOTE_HOST}"
printf 'mode=%s\n' "${MODE}"
printf 'architect_autodispatch=%s\n' "${AUTODISPATCH}"
printf 'architect_autoverify=%s\n' "${AUTOVERIFY}"

if [[ "${CHECK_ONLY}" -eq 1 ]]; then
  ssh "${REMOTE_HOST}" "
    set -euo pipefail
    sudo sh -c '
      if [ -f \"${REMOTE_ENV_PATH}\" ]; then
        grep -E \"^(LAYER_OS_ARCHITECT_AUTODISPATCH|LAYER_OS_ARCHITECT_AUTOVERIFY)=\" \"${REMOTE_ENV_PATH}\" || true
      else
        echo \"env_missing\"
      fi
    '
  "
  exit 0
fi

ssh "${REMOTE_HOST}" "
  set -euo pipefail
  tmp=\$(mktemp)
  trap 'rm -f \"\$tmp\"' EXIT
  sudo install -d -o layeros -g layeros \"\$(dirname '${REMOTE_ENV_PATH}')\"
  if sudo test -f '${REMOTE_ENV_PATH}'; then
    sudo cat '${REMOTE_ENV_PATH}' > \"\$tmp\"
  fi
  sed -i.bak '/^LAYER_OS_ARCHITECT_AUTODISPATCH=/d;/^LAYER_OS_ARCHITECT_AUTOVERIFY=/d' \"\$tmp\"
  rm -f \"\$tmp.bak\"
  {
    cat \"\$tmp\"
    printf 'LAYER_OS_ARCHITECT_AUTODISPATCH=%s\n' '${AUTODISPATCH}'
    printf 'LAYER_OS_ARCHITECT_AUTOVERIFY=%s\n' '${AUTOVERIFY}'
  } > \"\$tmp.next\"
  sudo install -o layeros -g layeros -m 0640 \"\$tmp.next\" '${REMOTE_ENV_PATH}'
  rm -f \"\$tmp.next\"
  sudo systemctl restart '${REMOTE_SERVICE_NAME}.service'
  for _ in \$(seq 1 30); do
    if curl -fsS 'http://127.0.0.1:17808/healthz' >/dev/null; then
      break
    fi
    sleep 2
  done
  curl -fsS 'http://127.0.0.1:17808/api/layer-os/daemon'
"
