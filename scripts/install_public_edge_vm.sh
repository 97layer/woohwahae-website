#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

REMOTE_HOST="${LAYER_OS_WEB_VM_HOST:-97layer-vm}"
SITE_NAME="woohwahae-public"
LOCAL_TEMPLATE="${ROOT_DIR}/scripts/nginx/woohwahae-public.conf.example"
LOCAL_NGINX_CONF="${ROOT_DIR}/scripts/nginx/nginx.layer-os.conf.example"
CHECK_ONLY=0

usage() {
  cat >&2 <<'EOF'
usage: install_public_edge_vm.sh [--host <ssh-host>] [--check]

Install the public nginx edge for woohwahae.kr on the Layer OS VM and proxy it
to the local founder/admin web service on 127.0.0.1:3081.
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

TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/layer-os-nginx-edge.XXXXXX")"
cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

LOCAL_RENDERED="${TMP_DIR}/${SITE_NAME}.conf"
cp "${LOCAL_TEMPLATE}" "${LOCAL_RENDERED}"
LOCAL_RENDERED_NGINX="${TMP_DIR}/nginx.conf"
cp "${LOCAL_NGINX_CONF}" "${LOCAL_RENDERED_NGINX}"

if [[ "${CHECK_ONLY}" -eq 1 ]]; then
  ssh "${REMOTE_HOST}" "SITE_NAME=${SITE_NAME} bash -s" <<'EOF'
set -euo pipefail

site_name="${SITE_NAME}"
site_conf="/etc/nginx/conf.d/${site_name}.conf"

echo "== current nginx state =="
systemctl is-enabled nginx.service 2>/dev/null || true
systemctl is-active nginx.service 2>/dev/null || true
echo

if [[ -f "${site_conf}" ]]; then
  echo "existing_site_conf=${site_conf}"
fi
echo "check only: no remote nginx changes applied"
EOF
  exit 0
fi

rsync -az "${LOCAL_RENDERED}" "${REMOTE_HOST}:/tmp/${SITE_NAME}.conf"
rsync -az "${LOCAL_RENDERED_NGINX}" "${REMOTE_HOST}:/tmp/nginx.conf"

ssh "${REMOTE_HOST}" "CHECK_ONLY=0 SITE_NAME=${SITE_NAME} bash -s" <<'EOF'
set -euo pipefail

backup_root="/etc/nginx/layer-os-backup"
stamp="$(date -u +%Y%m%d_%H%M%S)"
site_name="${SITE_NAME}"
site_conf="/etc/nginx/conf.d/${site_name}.conf"
incoming_conf="/tmp/${site_name}.conf"
incoming_nginx_conf="/tmp/nginx.conf"
legacy_files=(
  /etc/nginx/sites-enabled/default.bak
  /etc/nginx/sites-enabled/default.disabled
)

sudo install -d -m 0755 "${backup_root}/${stamp}" /etc/nginx/conf.d
if [[ -f /etc/nginx/nginx.conf ]]; then
  sudo cp /etc/nginx/nginx.conf "${backup_root}/${stamp}/nginx.conf"
fi
for legacy in "${legacy_files[@]}"; do
  if [[ -f "${legacy}" ]]; then
    sudo mv "${legacy}" "${backup_root}/${stamp}/"
  fi
done

sudo mv "${incoming_nginx_conf}" /etc/nginx/nginx.conf
sudo chown root:root /etc/nginx/nginx.conf
sudo chmod 0644 /etc/nginx/nginx.conf
sudo mv "${incoming_conf}" "${site_conf}"
sudo chown root:root "${site_conf}"
sudo chmod 0644 "${site_conf}"

sudo nginx -t
sudo systemctl unmask nginx.service
sudo systemctl enable nginx.service
sudo systemctl restart nginx.service

curl -fsS -H 'Host: woohwahae.kr' -H 'X-Forwarded-Proto: https' http://127.0.0.1/ >/dev/null
curl -fsS -H 'Host: woohwahae.kr' -H 'X-Forwarded-Proto: https' http://127.0.0.1/admin/login >/dev/null
sudo systemctl is-active nginx.service
EOF
