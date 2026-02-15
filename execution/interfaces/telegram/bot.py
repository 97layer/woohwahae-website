#!/usr/bin/env python3
"""
97LAYER OS Telegram Bot - Secure Version
- Uses environment variables for tokens
- Runs in Podman container
"""

import os
import json
import time
import urllib.request
from datetime import datetime
from pathlib import Path
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

# Configuration
TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")
if not TOKEN:
    raise ValueError("❌ TELEGRAM_BOT_TOKEN not found in environment variables")

BASE_URL = f"https://api.telegram.org/bot{TOKEN}"
CHAT_MEMORY = Path("/app/knowledge/chat_memory/7565534667.json")

def save_message(chat_id, text, role="user"):
    """메시지 저장"""
    CHAT_MEMORY.parent.mkdir(parents=True, exist_ok=True)

    # 기존 메시지 로드
    messages = []
    if CHAT_MEMORY.exists():
        with open(CHAT_MEMORY, 'r', encoding='utf-8') as f:
            messages = json.load(f)

    # 새 메시지 추가
    messages.append({
        "timestamp": datetime.now().isoformat(),
        "role": role,
        "content": text
    })

    # 저장
    with open(CHAT_MEMORY, 'w', encoding='utf-8') as f:
        json.dump(messages, f, ensure_ascii=False, indent=2)

    print(f"💾 저장됨: [{role}] {text[:50]}")

def send_message(chat_id, text):
    """메시지 전송"""
    url = f"{BASE_URL}/sendMessage"
    data = json.dumps({
        "chat_id": chat_id,
        "text": text
    }).encode('utf-8')

    req = urllib.request.Request(url, data=data)
    req.add_header('Content-Type', 'application/json')

    with urllib.request.urlopen(req) as response:
        print(f"📤 전송됨: {text[:50]}")

def main():
    print("=" * 60)
    print("🤖 97LAYER OS Telegram Bot (Secure)")
    print("=" * 60)
    print(f"📍 Container: 97layer-workspace")
    print(f"📂 Memory: {CHAT_MEMORY}")
    print("=" * 60)

    offset = None

    while True:
        try:
            # 업데이트 가져오기
            url = f"{BASE_URL}/getUpdates?timeout=5"
            if offset:
                url += f"&offset={offset}"

            with urllib.request.urlopen(url, timeout=10) as response:
                result = json.loads(response.read())

                for update in result.get("result", []):
                    offset = update["update_id"] + 1

                    if "message" in update:
                        msg = update["message"]
                        chat_id = msg["chat"]["id"]
                        text = msg.get("text", "")

                        if text:
                            print(f"\n📩 받음: {text}")

                            # 메시지 저장
                            save_message(chat_id, text, "user")

                            # 응답 생성
                            if text.startswith("/"):
                                if text == "/status":
                                    response = f"✅ 봇 작동 중 (Container Mode)\n시간: {datetime.now().strftime('%H:%M:%S')}"
                                elif text == "/start":
                                    response = "🤖 97LAYER OS Bot\n컨테이너 환경에서 실행 중입니다."
                                else:
                                    response = f"명령어: {text}"
                            else:
                                response = f"메시지 받았습니다: {text}\n\n처리 중..."

                            # 응답 전송
                            send_message(chat_id, response)

                            # 응답도 저장
                            save_message(chat_id, response, "assistant")

        except Exception as e:
            if "409" in str(e):
                print("⚠️ 409 에러 - 다른 봇 실행 중. 10초 대기...")
                time.sleep(10)
            else:
                print(f"오류: {e}")
                time.sleep(5)

if __name__ == "__main__":
    main()
