#!/usr/bin/env python3
"""
가장 단순한 작동하는 텔레그램 봇
메시지 받고, 저장하고, 응답하기
"""

import json
import time
import urllib.request
from datetime import datetime
from pathlib import Path

# 설정
TOKEN = "8271602365:AAGQwvDfmLv11_CShkeTMSQvnAkDYbDiTxA"
BASE_URL = f"https://api.telegram.org/bot{TOKEN}"
CHAT_MEMORY = Path.home() / "97layerOS" / "knowledge" / "chat_memory" / "7565534667.json"

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
    print("🤖 단순 작동 텔레그램 봇")
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
                                    response = f"✅ 봇 작동 중\n시간: {datetime.now().strftime('%H:%M:%S')}"
                                elif text == "/start":
                                    response = "🤖 97LAYER OS 봇입니다.\n무엇을 도와드릴까요?"
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