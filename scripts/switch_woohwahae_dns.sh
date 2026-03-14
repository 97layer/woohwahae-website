#!/usr/bin/env bash
set -euo pipefail

ZONE_ID="${CLOUDFLARE_ZONE_ID:-ee861dda0e43bccc805a9a8fa5cf445f}"
ZONE_NAME="${CLOUDFLARE_ZONE_NAME:-woohwahae.kr}"
TARGET_IP="${TARGET_IP:-136.109.201.201}"
ADMIN_HOST="${ADMIN_HOST:-admin.woohwahae.kr}"
CHECK_ONLY=0
DELETE_EDGECHECK=0

usage() {
  cat >&2 <<'EOF'
usage: switch_woohwahae_dns.sh [--ip <ipv4>] [--check] [--delete-edgecheck]

Switch woohwahae.kr from the old Pages CNAME to the Layer OS VM A record while
keeping www.woohwahae.kr proxied to the apex and admin.woohwahae.kr pointed at
the same VM. Requires /tmp/CLOUDFLARE_API_TOKEN.
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

cf_api() {
  local token
  token="$(load_token)"
  curl -fsS \
    -H "Authorization: Bearer ${token}" \
    -H "Content-Type: application/json" \
    "$@"
}

json_filter() {
  local script="$1"
  python3 -c "import json,sys; data=json.load(sys.stdin); ${script}"
}

json_payload_a() {
  python3 - "$1" "$2" <<'PY'
import json, sys
name = sys.argv[1]
ip = sys.argv[2]
print(json.dumps({
    "type": "A",
    "name": name,
    "content": ip,
    "ttl": 1,
    "proxied": True,
}))
PY
}

json_payload_cname() {
  python3 - "$1" "$2" <<'PY'
import json, sys
name = sys.argv[1]
target = sys.argv[2]
print(json.dumps({
    "type": "CNAME",
    "name": name,
    "content": target,
    "ttl": 1,
    "proxied": True,
}))
PY
}

json_record_id() {
  local record_name="$1"
  local record_type="$2"
  json_filter "
for record in data.get('result', []):
    if record.get('type') == '${record_type}' and record.get('name') == '${record_name}':
        print(record.get('id', ''))
        break
"
}

find_record() {
  local name="$1"
  cf_api "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records?name=${name}&per_page=200"
}

delete_record_if_present() {
  local record_id="$1"
  if [[ -n "${record_id}" ]]; then
    cf_api -X DELETE "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records/${record_id}" >/dev/null
  fi
}

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --ip)
      [[ "$#" -ge 2 ]] || die "--ip requires a value"
      TARGET_IP="$2"
      shift 2
      ;;
    --check)
      CHECK_ONLY=1
      shift
      ;;
    --delete-edgecheck)
      DELETE_EDGECHECK=1
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

require_cmd curl
require_cmd python3

if [[ "${CHECK_ONLY}" -eq 1 ]]; then
  for name in woohwahae.kr www.woohwahae.kr "${ADMIN_HOST}" edgecheck.woohwahae.kr; do
    printf '== %s ==\n' "${name}"
    find_record "${name}" | json_filter '
for record in data.get("result", []):
    print("{}\t{}\t{}\t{}\tproxied={}".format(
        record.get("id", ""),
        record.get("type", ""),
        record.get("name", ""),
        record.get("content", ""),
        record.get("proxied", False),
    ))
'
  done
  exit 0
fi

apex_json="$(find_record 'woohwahae.kr')"
apex_cname_id="$(printf '%s' "${apex_json}" | json_record_id 'woohwahae.kr' 'CNAME')"
apex_a_id="$(printf '%s' "${apex_json}" | json_record_id 'woohwahae.kr' 'A')"

if [[ -n "${apex_cname_id}" ]]; then
  delete_record_if_present "${apex_cname_id}"
fi

if [[ -n "${apex_a_id}" ]]; then
  cf_api -X PATCH \
    "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records/${apex_a_id}" \
    --data "$(json_payload_a 'woohwahae.kr' "${TARGET_IP}")" >/dev/null
else
  cf_api -X POST \
    "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records" \
    --data "$(json_payload_a 'woohwahae.kr' "${TARGET_IP}")" >/dev/null
fi

www_json="$(find_record 'www.woohwahae.kr')"
www_cname_id="$(printf '%s' "${www_json}" | json_record_id 'www.woohwahae.kr' 'CNAME')"

if [[ -n "${www_cname_id}" ]]; then
  cf_api -X PATCH \
    "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records/${www_cname_id}" \
    --data "$(json_payload_cname 'www.woohwahae.kr' 'woohwahae.kr')" >/dev/null
else
  cf_api -X POST \
    "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records" \
    --data "$(json_payload_cname 'www.woohwahae.kr' 'woohwahae.kr')" >/dev/null
fi

admin_json="$(find_record "${ADMIN_HOST}")"
admin_a_id="$(printf '%s' "${admin_json}" | json_record_id "${ADMIN_HOST}" 'A')"
admin_cname_id="$(printf '%s' "${admin_json}" | json_record_id "${ADMIN_HOST}" 'CNAME')"

if [[ -n "${admin_cname_id}" ]]; then
  delete_record_if_present "${admin_cname_id}"
fi

if [[ -n "${admin_a_id}" ]]; then
  cf_api -X PATCH \
    "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records/${admin_a_id}" \
    --data "$(json_payload_a "${ADMIN_HOST}" "${TARGET_IP}")" >/dev/null
else
  cf_api -X POST \
    "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records" \
    --data "$(json_payload_a "${ADMIN_HOST}" "${TARGET_IP}")" >/dev/null
fi

if [[ "${DELETE_EDGECHECK}" -eq 1 ]]; then
  edgecheck_json="$(find_record 'edgecheck.woohwahae.kr')"
  edgecheck_id="$(printf '%s' "${edgecheck_json}" | json_filter '
for record in data.get("result", []):
    if record.get("name") == "edgecheck.woohwahae.kr":
        print(record.get("id", ""))
        break
')"
  delete_record_if_present "${edgecheck_id}"
fi

printf 'dns switched to %s\n' "${TARGET_IP}"
