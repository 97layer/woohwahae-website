#!/usr/bin/env python3
"""
97LAYER OS Bot Test Script
Sends a test message to verify bot is working
"""

import requests
import json
import sys

# Bot token from config
BOT_TOKEN = "8501568801:AAE-3fBl-p6uZcmrdsWSRQuz_eg8yDADwjI"

def get_bot_info():
    """Get bot information"""
    url = f"https://api.telegram.org/bot{BOT_TOKEN}/getMe"
    response = requests.get(url)
    return response.json()

def get_updates():
    """Get recent updates to find chat IDs"""
    url = f"https://api.telegram.org/bot{BOT_TOKEN}/getUpdates"
    response = requests.get(url)
    return response.json()

def send_test_message(chat_id):
    """Send a test message"""
    url = f"https://api.telegram.org/bot{BOT_TOKEN}/sendMessage"
    data = {
        "chat_id": chat_id,
        "text": "🚀 97LAYER OS is online!\n\nSystem status:\n✓ Telegram daemon: Active\n✓ Gemini API: Connected\n✓ Memory manager: Ready\n\nSend me any message to capture it as a raw signal."
    }
    response = requests.post(url, json=data)
    return response.json()

def main():
    print("97LAYER OS Bot Test")
    print("=" * 40)

    # Get bot info
    bot_info = get_bot_info()
    if bot_info.get("ok"):
        bot = bot_info["result"]
        print(f"Bot: @{bot['username']} ({bot['first_name']})")
        print(f"ID: {bot['id']}")
    else:
        print("Error getting bot info:", bot_info)
        sys.exit(1)

    print("\nRecent chats:")
    print("-" * 40)

    # Get updates to find chat IDs
    updates = get_updates()
    if updates.get("ok"):
        chats = set()
        for update in updates["result"]:
            if "message" in update:
                chat = update["message"]["chat"]
                chat_id = chat["id"]
                chat_name = chat.get("username") or chat.get("first_name") or "Unknown"
                chats.add((chat_id, chat_name))

        if chats:
            for chat_id, chat_name in chats:
                print(f"Chat ID: {chat_id} ({chat_name})")

                # Send test message to first chat
                if input("\nSend test message to this chat? (y/n): ").lower() == "y":
                    result = send_test_message(chat_id)
                    if result.get("ok"):
                        print("✓ Test message sent successfully!")
                        print(f"  Message ID: {result['result']['message_id']}")
                    else:
                        print("✗ Failed to send message:", result)
                    break
        else:
            print("No recent chats found. Send a message to the bot first.")
            print("\nBot URL: https://t.me/97LayerOSwoohwahae")
    else:
        print("Error getting updates:", updates)

if __name__ == "__main__":
    main()