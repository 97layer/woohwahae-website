#!/usr/bin/env bash
set -euo pipefail

REMOTE_HOST="${LAYER_OS_VM_HOST:-97layer-vm}"
REMOTE_USER="${LAYER_OS_VM_USER:-}"
REMOTE_PORT="${LAYER_OS_VM_PORT:-22}"
REMOTE_SSH_KEY="${LAYER_OS_VM_SSH_KEY:-}"
TAIL_LINES="${TAIL_LINES:-200}"

usage() {
  cat >&2 <<'EOF'
usage: discover_vm_telegram_chats.sh [--host <ssh-host>] [--user <remote-user>] [--port <port>] [--ssh-key <path>] [--tail <lines>]

Read recent layer-osd journal lines from the VM and print candidate Telegram
chat ids seen by the live polling loop.
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
    --tail)
      [[ "$#" -ge 2 ]] || {
        usage
        exit 2
      }
      TAIL_LINES="$2"
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

ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "sudo journalctl -u layer-osd.service -n '${TAIL_LINES}' --no-pager" | python3 - <<'PY'
import re
import sys

pattern = re.compile(
    r'telegram bot: inbound chat_id=(?P<chat>-?\d+) '
    r'(?:route=(?P<route>[a-z_]+) )?'
    r'(?:chat_type=(?P<type>[a-z_]+) chat_label=(?P<label>".*?") )?'
    r'message_id=(?P<message>\d+) text_len=(?P<length>\d+) update_id=(?P<update>\d+)'
)
seen = {}
for raw in sys.stdin:
    match = pattern.search(raw)
    if not match:
        continue
    chat_id = match.group("chat")
    seen[chat_id] = {
        "route": match.group("route") or "unknown",
        "chat_type": match.group("type") or "unknown",
        "chat_label": match.group("label") or '""',
        "message_id": match.group("message"),
        "text_len": match.group("length"),
        "update_id": match.group("update"),
    }

if not seen:
    print("no_recent_chats_found")
    raise SystemExit(0)

for chat_id, meta in sorted(seen.items()):
    print(
        f"chat_id={chat_id} "
        f"route={meta['route']} "
        f"chat_type={meta['chat_type']} "
        f"chat_label={meta['chat_label']} "
        f"latest_message_id={meta['message_id']} "
        f"text_len={meta['text_len']} "
        f"update_id={meta['update_id']}"
    )
PY
