#!/usr/bin/env bash
set -euo pipefail

REMOTE_HOST="${LAYER_OS_VM_HOST:-97layer-vm}"
REMOTE_USER="${LAYER_OS_VM_USER:-}"
REMOTE_PORT="${LAYER_OS_VM_PORT:-22}"
REMOTE_SSH_KEY="${LAYER_OS_VM_SSH_KEY:-}"

usage() {
  cat >&2 <<'EOF'
usage: vm_runtime_inventory.sh [--host <ssh-host>] [--user <remote-user>] [--port <port>] [--ssh-key <path>]

Show the minimal always-on Layer OS inventory for a remote VM:
- core systemd services
- key listeners
- provider readiness without printing secrets
- top process memory snapshot
- Layer OS disk paths
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

ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" 'bash -s' <<'EOF'
set -euo pipefail

services=(
  layer-osd.service
  layer-os-web.service
  nginx.service
)

echo "== core services =="
for service in "${services[@]}"; do
  if systemctl list-unit-files "${service}" >/dev/null 2>&1; then
    systemctl show \
      -p Id \
      -p LoadState \
      -p ActiveState \
      -p SubState \
      -p MainPID \
      -p ExecMainStartTimestamp \
      -p MemoryCurrent \
      "${service}"
    echo
  else
    echo "Id=${service}"
    echo "LoadState=missing"
    echo
  fi
done

echo "== key listeners =="
sudo ss -ltnp | grep -E ":(80|443|17808|3081)\\b" || echo "no matching listeners"
echo

echo "== provider readiness =="
sudo python3 - <<'PY'
from pathlib import Path

service_env_path = Path("/etc/layer-os/layer-osd.env")
repo_root = Path("/srv/layer-os/current")
env_path = Path("/etc/layer-os/providers.env")
service_values = {}

if service_env_path.exists():
    for raw in service_env_path.read_text().splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        service_values[key.strip()] = value.strip()

provider_override = service_values.get("LAYER_OS_PROVIDER_ENV_FILE", "").strip()
if provider_override:
    override_path = Path(provider_override)
    env_path = override_path if override_path.is_absolute() else repo_root / override_path

values = {}
if env_path.exists():
    for raw in env_path.read_text().splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        values[key.strip()] = value.strip()

def present(name: str) -> bool:
    return bool(values.get(name, ""))

def present_any(*names: str) -> bool:
    return any(present(name) for name in names)

token_ready = present("TELEGRAM_BOT_TOKEN")
founder_specific_ready = present("TELEGRAM_FOUNDER_CHAT_ID")
founder_dm_ready = present("TELEGRAM_FOUNDER_DM_CHAT_ID")
legacy_founder_ready = present("TELEGRAM_CHAT_ID")
founder_chat_ready = founder_specific_ready or legacy_founder_ready
founder_conversation_ready = founder_dm_ready or (founder_chat_ready and not str(values.get("TELEGRAM_FOUNDER_CHAT_ID", values.get("TELEGRAM_CHAT_ID", ""))).startswith("-"))
ops_chat_ready = present("TELEGRAM_OPS_CHAT_ID")
brand_chat_ready = present("TELEGRAM_BRAND_CHAT_ID")
gemini_ready = present_any("GEMINI_API_KEY", "GOOGLE_API_KEY")
threads_ready = present("THREADS_ACCESS_TOKEN")
legacy_founder_alias = legacy_founder_ready and not founder_specific_ready

if token_ready and gemini_ready and founder_conversation_ready:
    inbound_mode = "assistant"
elif token_ready:
    inbound_mode = "command_only"
else:
    inbound_mode = "off"

if token_ready and founder_chat_ready:
    founder_delivery = "ready"
elif token_ready:
    founder_delivery = "chat_missing"
elif founder_chat_ready:
    founder_delivery = "token_missing"
else:
    founder_delivery = "disabled"

if token_ready and ops_chat_ready:
    ops_delivery = "ready"
elif token_ready and founder_chat_ready:
    ops_delivery = "founder_fallback"
elif ops_chat_ready:
    ops_delivery = "token_missing"
else:
    ops_delivery = "disabled"

if token_ready and brand_chat_ready:
    brand_delivery = "ready"
elif brand_chat_ready:
    brand_delivery = "token_missing"
else:
    brand_delivery = "disabled"

for key in (
    "OPENAI_API_KEY",
    "ANTHROPIC_API_KEY",
    "GEMINI_API_KEY",
    "GOOGLE_API_KEY",
    "TELEGRAM_BOT_TOKEN",
    "TELEGRAM_FOUNDER_CHAT_ID",
    "TELEGRAM_FOUNDER_DM_CHAT_ID",
    "TELEGRAM_OPS_CHAT_ID",
    "TELEGRAM_BRAND_CHAT_ID",
    "TELEGRAM_CHAT_ID",
    "THREADS_ACCESS_TOKEN",
):
    print(f"{key}={'present' if present(key) else 'missing'}")
print(f"provider_env_path={env_path}")
print(f"inbound_mode={inbound_mode}")
print(f"founder_delivery={founder_delivery}")
print(f"ops_delivery={ops_delivery}")
print(f"brand_delivery={brand_delivery}")
print(f"threads_publish={'ready' if threads_ready else 'disabled'}")
print(f"legacy_founder_alias={'true' if legacy_founder_alias else 'false'}")
PY
echo

echo "== top memory processes =="
ps -eo pid,ppid,pmem,rss,etime,comm,args --sort=-rss | head -n 12
echo

echo "== layer os disk =="
df -h /srv/layer-os /var/lib/layer-os /var/log/layer-os 2>/dev/null || true
echo
du -sh /srv/layer-os /var/lib/layer-os /var/log/layer-os 2>/dev/null || true
EOF
