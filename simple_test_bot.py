#!/usr/bin/env python3
"""
Simple standalone bot for testing
Runs in foreground with clear output
"""

import time
import os
import json
import urllib.request
import urllib.error
from datetime import datetime

BOT_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")
BASE_URL = f"https://api.telegram.org/bot{BOT_TOKEN}"

def get_updates(offset=None):
    """Get new messages"""
    url = f"{BASE_URL}/getUpdates"
    if offset:
        url += f"?offset={offset}"

    try:
        with urllib.request.urlopen(url, timeout=30) as response:
            data = json.loads(response.read().decode('utf-8'))
            if data.get("ok"):
                return data.get("result", [])
    except Exception as e:
        print(f"Error getting updates: {e}")
    return []

def send_message(chat_id, text):
    """Send a message"""
    url = f"{BASE_URL}/sendMessage"
    data = json.dumps({
        "chat_id": chat_id,
        "text": text
    }).encode('utf-8')

    req = urllib.request.Request(url, data=data, headers={'Content-Type': 'application/json'})

    try:
        with urllib.request.urlopen(req) as response:
            result = json.loads(response.read().decode('utf-8'))
            return result.get("ok", False)
    except Exception as e:
        print(f"Error sending message: {e}")
        return False

def main():
    print("=" * 50)
    print("97LAYER OS Simple Test Bot")
    print("=" * 50)
    print(f"Started at {datetime.now()}")
    print("Bot: @official_97Layer_OSwoohwahae_bot")
    print("\nSend a message to the bot to test it!")
    print("Press Ctrl+C to stop\n")

    offset = None
    message_count = 0

    while True:
        try:
            updates = get_updates(offset)

            for update in updates:
                offset = update["update_id"] + 1

                if "message" in update:
                    message = update["message"]
                    chat_id = message["chat"]["id"]
                    chat_name = message["chat"].get("username") or message["chat"].get("first_name") or "User"
                    text = message.get("text", "")

                    message_count += 1
                    print(f"[{datetime.now().strftime('%H:%M:%S')}] Message #{message_count}")
                    print(f"  From: {chat_name} (ID: {chat_id})")
                    print(f"  Text: {text[:100]}")

                    # Send response
                    response_text = f"✓ Received your message #{message_count}\n\n"
                    response_text += f"Text: {text}\n"
                    response_text += f"Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n"
                    response_text += "97LAYER OS is processing your signal..."

                    if send_message(chat_id, response_text):
                        print(f"  Response: Sent successfully")
                    else:
                        print(f"  Response: Failed to send")
                    print()

            time.sleep(1)

        except KeyboardInterrupt:
            print("\n\nStopping bot...")
            break
        except Exception as e:
            print(f"Error: {e}")
            time.sleep(5)

    print(f"\nProcessed {message_count} messages")
    print("Bot stopped")

if __name__ == "__main__":
    main()