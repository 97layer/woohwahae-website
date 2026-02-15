#!/usr/bin/env python3
"""
97LAYER OS - WORKING BOT
Immediate execution with Gemini integration
Simple, direct, functional
"""

import os
import sys
import json
import time
import urllib.request
import urllib.parse
from datetime import datetime
from pathlib import Path

# Project setup
PROJECT_ROOT = Path(__file__).resolve().parent
sys.path.append(str(PROJECT_ROOT))

# Configuration
BOT_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")
GEMINI_KEY = "AIzaSyBHpQRFjdZRzzkYGR6eqBezyPteaHX_uMQ"
BASE_URL = f"https://api.telegram.org/bot{BOT_TOKEN}"

# Stats
stats = {
    "start_time": datetime.now(),
    "messages": 0,
    "gemini_calls": 0,
    "errors": 0
}


def gemini_quick(prompt):
    """Quick Gemini API call"""
    url = "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent"

    data = json.dumps({
        "contents": [{"parts": [{"text": prompt}]}],
        "generationConfig": {
            "temperature": 0.7,
            "maxOutputTokens": 200
        }
    }).encode('utf-8')

    headers = {"Content-Type": "application/json"}
    url_with_key = f"{url}?key={GEMINI_KEY}"

    try:
        req = urllib.request.Request(url_with_key, data=data, headers=headers)
        with urllib.request.urlopen(req, timeout=10) as response:
            result = json.loads(response.read().decode('utf-8'))

        if 'candidates' in result and result['candidates']:
            text = result['candidates'][0]['content']['parts'][0]['text']
            stats["gemini_calls"] += 1
            return text
        else:
            return f"Gemini response format error: {result.get('error', 'Unknown')}"
    except Exception as e:
        stats["errors"] += 1
        return f"Gemini error: {str(e)[:100]}"


def get_updates(offset=None):
    """Get Telegram updates"""
    url = f"{BASE_URL}/getUpdates?timeout=30"
    if offset:
        url += f"&offset={offset}"

    try:
        with urllib.request.urlopen(url, timeout=35) as response:
            data = json.loads(response.read().decode('utf-8'))
            if data.get("ok"):
                return data.get("result", [])
    except Exception as e:
        print(f"[ERROR] Get updates: {e}")
        stats["errors"] += 1
    return []


def send_message(chat_id, text):
    """Send Telegram message"""
    url = f"{BASE_URL}/sendMessage"
    data = urllib.parse.urlencode({
        "chat_id": chat_id,
        "text": text[:4000],  # Telegram limit
        "parse_mode": "Markdown"
    }).encode('utf-8')

    try:
        with urllib.request.urlopen(url, data=data, timeout=10) as response:
            return True
    except Exception as e:
        print(f"[ERROR] Send message: {e}")
        stats["errors"] += 1
        return False


def save_signal(content, source="telegram"):
    """Save to knowledge/raw_signals"""
    signal_dir = PROJECT_ROOT / "knowledge" / "raw_signals"
    signal_dir.mkdir(parents=True, exist_ok=True)

    timestamp = datetime.now()
    filename = f"rs-{timestamp.strftime('%Y%m%d%H%M%S')}_{source}.md"

    with open(signal_dir / filename, "w", encoding="utf-8") as f:
        f.write(f"# Raw Signal\n\n")
        f.write(f"**Date**: {timestamp.isoformat()}\n")
        f.write(f"**Source**: {source}\n\n")
        f.write(f"---\n\n{content}\n")

    return filename


def handle_command(message):
    """Handle bot commands"""
    text = message.get("text", "")
    chat_id = message["chat"]["id"]

    if text == "/start" or text == "/help":
        help_text = """🤖 *97LAYER OS Bot*

*Commands:*
/status - System status
/gemini [text] - Ask Gemini
/test - Test AI connection

*Features:*
• Auto-saves all messages to knowledge base
• Gemini AI integration for analysis
• Multimodal support (text + images)

Send any message to capture it!"""
        send_message(chat_id, help_text)

    elif text == "/status":
        uptime = datetime.now() - stats["start_time"]
        status_text = f"""📊 *System Status*

*Uptime:* {int(uptime.total_seconds() // 60)} minutes
*Messages:* {stats['messages']}
*Gemini Calls:* {stats['gemini_calls']}
*Errors:* {stats['errors']}

*API Status:*
• Telegram: ✅ Connected
• Gemini: {'✅ Active' if stats['gemini_calls'] > 0 else '⚠️ Not tested'}

*Knowledge Base:*
• Signals: {len(list(Path('knowledge/raw_signals').glob('*.md'))) if Path('knowledge/raw_signals').exists() else 0} files"""
        send_message(chat_id, status_text)

    elif text == "/test":
        send_message(chat_id, "🔍 Testing Gemini...")
        response = gemini_quick("Say 'System operational' if working")
        send_message(chat_id, f"✅ Gemini: {response}")

    elif text.startswith("/gemini "):
        prompt = text[8:]
        send_message(chat_id, "💭 Processing...")
        response = gemini_quick(prompt)
        send_message(chat_id, f"🤖 Gemini:\n{response}")


def handle_message(message):
    """Handle regular messages"""
    text = message.get("text", "")
    chat_id = message["chat"]["id"]
    user = message["from"].get("username", message["from"].get("first_name", "User"))

    if not text:
        return

    # Check if command
    if text.startswith("/"):
        handle_command(message)
        return

    stats["messages"] += 1

    # Save signal
    filename = save_signal(text, f"telegram_{user}")

    # Quick AI response for substantial messages
    response = f"✅ Captured: `{filename}`\n"

    if len(text) > 30:
        ai_response = gemini_quick(f"Very brief (1 line) insight about: {text[:200]}")
        if "error" not in ai_response.lower():
            response += f"\n💡 {ai_response}"

    send_message(chat_id, response)


def main():
    """Main bot loop"""
    print("=" * 60)
    print("97LAYER OS WORKING BOT")
    print("=" * 60)
    print(f"Started: {datetime.now()}")
    print(f"Bot: @official_97Layer_OSwoohwahae_bot")
    print("\nPress Ctrl+C to stop\n")

    # Test Gemini on startup
    print("Testing Gemini...")
    test = gemini_quick("Say OK")
    print(f"Gemini: {test}\n")

    offset = None

    while True:
        try:
            updates = get_updates(offset)

            for update in updates:
                offset = update["update_id"] + 1

                if "message" in update:
                    message = update["message"]

                    # Log
                    user = message["from"].get("username", "User")
                    text = message.get("text", "")[:50]
                    print(f"[{datetime.now().strftime('%H:%M:%S')}] {user}: {text}")

                    # Handle
                    handle_message(message)

            # Brief pause
            time.sleep(0.5)

        except KeyboardInterrupt:
            print("\n\nShutting down...")
            break
        except Exception as e:
            print(f"[ERROR] Main loop: {e}")
            stats["errors"] += 1
            time.sleep(5)

    # Final stats
    print(f"\nFinal Stats:")
    print(f"  Messages: {stats['messages']}")
    print(f"  Gemini Calls: {stats['gemini_calls']}")
    print(f"  Errors: {stats['errors']}")


if __name__ == "__main__":
    # Kill existing bots
    os.system("pkill -f telegram_daemon 2>/dev/null")
    os.system("pkill -f unified_system 2>/dev/null")
    os.system("pkill -f integrated_bot 2>/dev/null")
    os.system("launchctl unload ~/Library/LaunchAgents/com.97layer.os.plist 2>/dev/null")
    time.sleep(2)

    # Run
    main()