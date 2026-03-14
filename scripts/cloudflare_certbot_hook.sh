#!/usr/bin/env bash
set -euo pipefail

MODE="${1:-}"
ZONE_ID="${CLOUDFLARE_ZONE_ID:-ee861dda0e43bccc805a9a8fa5cf445f}"
ZONE_NAME="${CLOUDFLARE_ZONE_NAME:-woohwahae.kr}"
PROPAGATION_SECONDS="${CLOUDFLARE_DNS_PROPAGATION_SECONDS:-30}"

usage() {
  cat >&2 <<'EOF'
usage: cloudflare_certbot_hook.sh <auth|cleanup>

Certbot manual DNS-01 hook for woohwahae.kr using a scoped Cloudflare API token.
The hook reads the token from CLOUDFLARE_API_TOKEN, CLOUDFLARE_API_TOKEN_FILE,
or /tmp/CLOUDFLARE_API_TOKEN.
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

load_token() {
  if [[ -n "${CLOUDFLARE_API_TOKEN:-}" ]]; then
    printf '%s' "${CLOUDFLARE_API_TOKEN}"
    return
  fi
  if [[ -n "${CLOUDFLARE_API_TOKEN_FILE:-}" && -f "${CLOUDFLARE_API_TOKEN_FILE}" ]]; then
    cat "${CLOUDFLARE_API_TOKEN_FILE}"
    return
  fi
  if [[ -f /tmp/CLOUDFLARE_API_TOKEN ]]; then
    cat /tmp/CLOUDFLARE_API_TOKEN
    return
  fi
  die "Cloudflare API token not found"
}

json_get() {
  local expr="$1"
  python3 -c "import json,sys; data=json.load(sys.stdin); value=${expr}; print(value if value is not None else '')"
}

cf_api() {
  local token
  token="$(load_token)"
  curl -fsS \
    -H "Authorization: Bearer ${token}" \
    -H "Content-Type: application/json" \
    "$@"
}

record_name() {
  printf '_acme-challenge.%s' "${CERTBOT_DOMAIN}"
}

create_record() {
  local name payload response record_id
  name="$(record_name)"
  payload="$(python3 - "${name}" "${CERTBOT_VALIDATION}" <<'PY'
import json, sys
name = sys.argv[1]
value = sys.argv[2]
print(json.dumps({
    "type": "TXT",
    "name": name,
    "content": value,
    "ttl": 120,
}))
PY
)"
  response="$(cf_api -X POST "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records" --data "${payload}")"
  record_id="$(printf '%s' "${response}" | json_get 'data["result"]["id"]')"
  [[ -n "${record_id}" ]] || die "failed to create TXT record for ${name}"
  sleep "${PROPAGATION_SECONDS}"
  printf '%s\n' "${record_id}"
}

find_record_id() {
  local name response
  name="$(record_name)"
  response="$(cf_api "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records?type=TXT&name=${name}")"
  printf '%s' "${response}" | python3 - "${CERTBOT_VALIDATION}" <<'PY'
import json
import sys

target = sys.argv[1]
data = json.load(sys.stdin)
for record in data.get("result", []):
    if record.get("content") == target:
        print(record.get("id", ""))
        break
PY
}

delete_record() {
  local record_id
  record_id="${CERTBOT_AUTH_OUTPUT:-}"
  if [[ -z "${record_id}" ]]; then
    record_id="$(find_record_id || true)"
  fi
  if [[ -z "${record_id}" ]]; then
    exit 0
  fi
  cf_api -X DELETE "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records/${record_id}" >/dev/null
}

require_cmd curl
require_cmd python3

[[ -n "${CERTBOT_DOMAIN:-}" ]] || die "CERTBOT_DOMAIN is required"
[[ -n "${CERTBOT_VALIDATION:-}" ]] || die "CERTBOT_VALIDATION is required"

case "${MODE}" in
  auth)
    create_record
    ;;
  cleanup)
    delete_record
    ;;
  -h|--help|"")
    usage
    exit 2
    ;;
  *)
    die "unknown mode: ${MODE}"
    ;;
esac
