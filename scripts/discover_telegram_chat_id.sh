#!/usr/bin/env bash
set -euo pipefail

OFFSET=""
LIMIT="25"

usage() {
  cat >&2 <<'EOF'
usage: discover_telegram_chat_id.sh [--offset <update-id>] [--limit <count>]

Inspect recent Telegram bot updates and print unique chat ids without revealing
the bot token. If no chats appear, send a message to the bot from the target
chat and run the script again.
EOF
}

die() {
  printf 'error: %s\n' "$1" >&2
  exit 1
}

bot_token() {
  local value="${TELEGRAM_BOT_TOKEN:-}"
  if [[ -n "${value}" ]]; then
    printf '%s' "${value}"
    return 0
  fi
  if command -v security >/dev/null 2>&1; then
    value="$(security find-generic-password -a layer-os -s TELEGRAM_BOT_TOKEN -w 2>/dev/null || true)"
    if [[ -n "${value}" ]]; then
      printf '%s' "${value}"
      return 0
    fi
  fi
  return 1
}

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --offset)
      [[ "$#" -ge 2 ]] || die "--offset requires a value"
      OFFSET="$2"
      shift 2
      ;;
    --limit)
      [[ "$#" -ge 2 ]] || die "--limit requires a value"
      LIMIT="$2"
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

command -v curl >/dev/null 2>&1 || die "curl is required"
command -v python3 >/dev/null 2>&1 || die "python3 is required"

TOKEN="$(bot_token || true)"
[[ -n "${TOKEN}" ]] || die "TELEGRAM_BOT_TOKEN is required in env or macOS Keychain"

URL="https://api.telegram.org/bot${TOKEN}/getUpdates?limit=${LIMIT}"
if [[ -n "${OFFSET}" ]]; then
  URL="${URL}&offset=${OFFSET}"
fi

TMP_JSON="$(mktemp "${TMPDIR:-/tmp}/telegram-updates.XXXXXX.json")"
cleanup() {
  rm -f "${TMP_JSON}"
}
trap cleanup EXIT

curl -fsSL "${URL}" > "${TMP_JSON}"

python3 - "${TMP_JSON}" <<'PY'
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text())
if not payload.get("ok"):
    raise SystemExit("telegram api did not return ok=true")

def pick_chat(update):
    for field in ("message", "edited_message", "channel_post", "edited_channel_post", "my_chat_member", "chat_member"):
        value = update.get(field)
        if isinstance(value, dict):
            chat = value.get("chat")
            if isinstance(chat, dict):
                return chat
    return None

chats = {}
max_update = None
for item in payload.get("result", []):
    update_id = item.get("update_id")
    if isinstance(update_id, int):
      max_update = update_id if max_update is None else max(max_update, update_id)
    chat = pick_chat(item)
    if not chat:
        continue
    chat_id = chat.get("id")
    if chat_id is None:
        continue
    chats[str(chat_id)] = {
        "type": chat.get("type", "unknown"),
        "title": chat.get("title") or chat.get("username") or " ".join(
            part for part in [chat.get("first_name"), chat.get("last_name")] if part
        ).strip() or "unknown",
        "username": chat.get("username") or "",
    }

if not chats:
    print("no_chats_found")
    if max_update is not None:
        print(f"next_offset={max_update + 1}")
    raise SystemExit(0)

for chat_id, meta in sorted(chats.items()):
    username = f" username=@{meta['username']}" if meta["username"] else ""
    print(f"chat_id={chat_id} type={meta['type']} title={meta['title']}{username}")

if max_update is not None:
    print(f"next_offset={max_update + 1}")
PY
