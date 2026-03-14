#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

REMOTE_HOST="${LAYER_OS_WEB_VM_HOST:-97layer-vm}"
LOCAL_HOOK="${ROOT_DIR}/scripts/cloudflare_certbot_hook.sh"
LOCAL_TLS_TEMPLATE="${ROOT_DIR}/scripts/nginx/woohwahae-public.tls.conf.example"
REMOTE_HOOK="/usr/local/lib/layer-os/cloudflare_certbot_hook.sh"
REMOTE_TOKEN="/etc/layer-os/cloudflare-dns.token"
REMOTE_SITE_CONF="/etc/nginx/conf.d/woohwahae-public.conf"
ZONE_ID="${CLOUDFLARE_ZONE_ID:-ee861dda0e43bccc805a9a8fa5cf445f}"
ZONE_NAME="${CLOUDFLARE_ZONE_NAME:-woohwahae.kr}"
PROPAGATION_SECONDS="${CLOUDFLARE_DNS_PROPAGATION_SECONDS:-30}"
LETSENCRYPT_EMAIL="${LAYER_OS_LETSENCRYPT_EMAIL:-}"
PUBLIC_EDGE_DOMAINS="${LAYER_OS_PUBLIC_EDGE_DOMAINS:-woohwahae.kr,www.woohwahae.kr,admin.woohwahae.kr}"
CHECK_ONLY=0

usage() {
  cat >&2 <<'EOF'
usage: issue_public_tls_vm.sh [--host <ssh-host>] [--email <email>] [--check]

Issue or expand the Let's Encrypt certificate used by the public/admin edge on
the Layer OS VM using Cloudflare DNS-01 hooks, then install the HTTPS nginx
edge. Requires a scoped token in /tmp/CLOUDFLARE_API_TOKEN on the local machine.
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
    --email)
      [[ "$#" -ge 2 ]] || die "--email requires a value"
      LETSENCRYPT_EMAIL="$2"
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

require_cmd rsync
require_cmd ssh
require_cmd stat

[[ -f "${LOCAL_HOOK}" ]] || die "missing hook script: ${LOCAL_HOOK}"
[[ -f "${LOCAL_TLS_TEMPLATE}" ]] || die "missing tls template: ${LOCAL_TLS_TEMPLATE}"
[[ -f /tmp/CLOUDFLARE_API_TOKEN ]] || die "missing /tmp/CLOUDFLARE_API_TOKEN"

TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/layer-os-public-tls.XXXXXX")"
cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

LOCAL_TOKEN_COPY="${TMP_DIR}/cloudflare-dns.token"
cp /tmp/CLOUDFLARE_API_TOKEN "${LOCAL_TOKEN_COPY}"
chmod 600 "${LOCAL_TOKEN_COPY}"

if [[ "${CHECK_ONLY}" -eq 1 ]]; then
  ssh "${REMOTE_HOST}" "bash -s" <<'EOF'
set -euo pipefail
printf 'certbot='
command -v certbot
printf 'nginx_active='
systemctl is-active nginx.service 2>/dev/null || true
printf '\n'
EOF
  exit 0
fi

rsync -az "${LOCAL_HOOK}" "${REMOTE_HOST}:/tmp/cloudflare_certbot_hook.sh"
rsync -az "${LOCAL_TLS_TEMPLATE}" "${REMOTE_HOST}:/tmp/woohwahae-public.tls.conf"
rsync -az "${LOCAL_TOKEN_COPY}" "${REMOTE_HOST}:/tmp/cloudflare-dns.token"

ssh "${REMOTE_HOST}" \
  "ZONE_ID='${ZONE_ID}' ZONE_NAME='${ZONE_NAME}' PROPAGATION_SECONDS='${PROPAGATION_SECONDS}' LETSENCRYPT_EMAIL='${LETSENCRYPT_EMAIL}' PUBLIC_EDGE_DOMAINS='${PUBLIC_EDGE_DOMAINS}' REMOTE_HOOK='${REMOTE_HOOK}' REMOTE_TOKEN='${REMOTE_TOKEN}' REMOTE_SITE_CONF='${REMOTE_SITE_CONF}' bash -s" <<'EOF'
set -euo pipefail

sudo install -d -m 0755 /usr/local/lib/layer-os /etc/layer-os /etc/nginx/conf.d
sudo install -o root -g root -m 0700 /tmp/cloudflare_certbot_hook.sh "${REMOTE_HOOK}"
sudo install -o root -g root -m 0600 /tmp/cloudflare-dns.token "${REMOTE_TOKEN}"
rm -f /tmp/cloudflare_certbot_hook.sh /tmp/cloudflare-dns.token

cert_path="/etc/letsencrypt/live/woohwahae.kr/fullchain.pem"
key_path="/etc/letsencrypt/live/woohwahae.kr/privkey.pem"

normalize_domains() {
  python3 - "$1" <<'PY'
import sys

seen = set()
for part in sys.argv[1].split(','):
    value = part.strip().lower()
    if value and value not in seen:
        seen.add(value)
        print(value)
PY
}

extract_cert_domains() {
  local path="$1"
  python3 - "$path" <<'PY'
import re
import subprocess
import sys

path = sys.argv[1]
output = subprocess.check_output(
    ['openssl', 'x509', '-in', path, '-noout', '-ext', 'subjectAltName'],
    text=True,
    stderr=subprocess.DEVNULL,
)
seen = set()
for match in re.findall(r'DNS:([^,\s]+)', output):
    name = match.strip().lower()
    if name and name not in seen:
        seen.add(name)
        print(name)
PY
}

contains_domain() {
  local needle="$1"
  shift
  local item
  for item in "$@"; do
    if [[ "${item}" == "${needle}" ]]; then
      return 0
    fi
  done
  return 1
}

mapfile -t requested_domains < <(normalize_domains "${PUBLIC_EDGE_DOMAINS}")
current_domains=()
if [[ -f "${cert_path}" ]]; then
  mapfile -t current_domains < <(extract_cert_domains "${cert_path}")
fi

issue_domains=("${current_domains[@]}")
needs_issue=0
if [[ ! -f "${cert_path}" || ! -f "${key_path}" ]]; then
  needs_issue=1
  issue_domains=()
fi

for domain in "${requested_domains[@]}"; do
  if ! contains_domain "${domain}" "${current_domains[@]}"; then
    needs_issue=1
  fi
  if ! contains_domain "${domain}" "${issue_domains[@]}"; then
    issue_domains+=("${domain}")
  fi
done

email_args=(--register-unsafely-without-email)
if [[ -n "${LETSENCRYPT_EMAIL}" ]]; then
  email_args=(--email "${LETSENCRYPT_EMAIL}")
fi

if [[ "${needs_issue}" -eq 1 ]]; then
  domain_args=()
  for domain in "${issue_domains[@]}"; do
    domain_args+=(-d "${domain}")
  done
  sudo env \
    CLOUDFLARE_API_TOKEN_FILE="${REMOTE_TOKEN}" \
    CLOUDFLARE_ZONE_ID="${ZONE_ID}" \
    CLOUDFLARE_ZONE_NAME="${ZONE_NAME}" \
    CLOUDFLARE_DNS_PROPAGATION_SECONDS="${PROPAGATION_SECONDS}" \
    certbot certonly \
      --manual \
      --preferred-challenges dns \
      --manual-auth-hook "${REMOTE_HOOK} auth" \
      --manual-cleanup-hook "${REMOTE_HOOK} cleanup" \
      --non-interactive \
      --agree-tos \
      --expand \
      --cert-name woohwahae.kr \
      "${email_args[@]}" \
      "${domain_args[@]}"
fi

sudo install -o root -g root -m 0644 /tmp/woohwahae-public.tls.conf "${REMOTE_SITE_CONF}"
rm -f /tmp/woohwahae-public.tls.conf

sudo nginx -t
sudo systemctl restart nginx.service

curl -kfsS --resolve woohwahae.kr:443:127.0.0.1 https://woohwahae.kr/ >/dev/null
curl -kfsS --resolve woohwahae.kr:443:127.0.0.1 https://woohwahae.kr/admin/login >/dev/null
curl -kfsS --resolve admin.woohwahae.kr:443:127.0.0.1 https://admin.woohwahae.kr/admin/login >/dev/null
sudo systemctl is-active nginx.service >/dev/null
EOF
