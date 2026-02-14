#!/usr/bin/env python3
"""
ONE BOT - 단 하나만 실행되는 텔레그램 봇
GCP 충돌 없이 확실하게 작동
"""

import json
import time
import os
import sys
from datetime import datetime
from pathlib import Path
import urllib.request

# 봇 토큰
TOKEN = "8271602365:AAGQwvDfmLv11_CShkeTMSQvnAkDYbDiTxA"
BASE_URL = f"https://api.telegram.org/bot{TOKEN}"

# 파일 경로
PROJECT_ROOT = Path.home() / "97layerOS"
CHAT_FILE = PROJECT_ROOT / "knowledge" / "telegram_chat.json"
CHAT_FILE.parent.mkdir(parents=True, exist_ok=True)

print("=" * 60)
print("🤖 ONE BOT - 단일 텔레그램 봇")
print("=" * 60)

# 초기 업데이트 클리어
print("초기화 중...")
try:
    url = f"{BASE_URL}/getUpdates?offset=-1"
    urllib.request.urlopen(url)
    print("✅ 초기화 완료")
except:
    pass

offset = None

print("\n📱 텔레그램에서 메시지를 보내세요!")
print("💾 저장: " + str(CHAT_FILE))
print("-" * 60 + "\n")

def save_chat(text, role="user"):
    """채팅 저장"""
    chats = []
    if CHAT_FILE.exists():
        with open(CHAT_FILE) as f:
            chats = json.load(f)

    chats.append({
        "time": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
        "role": role,
        "text": text
    })

    with open(CHAT_FILE, 'w') as f:
        json.dump(chats, f, ensure_ascii=False, indent=2)

def send_msg(chat_id, text):
    """메시지 전송"""
    url = f"{BASE_URL}/sendMessage"
    data = json.dumps({"chat_id": chat_id, "text": text}).encode()
    req = urllib.request.Request(url, data=data, headers={'Content-Type': 'application/json'})
    urllib.request.urlopen(req)

# 메인 루프
while True:
    try:
        # 업데이트 받기
        url = f"{BASE_URL}/getUpdates?timeout=5"
        if offset:
            url += f"&offset={offset}"

        response = urllib.request.urlopen(url)
        data = json.loads(response.read())

        for update in data.get("result", []):
            offset = update["update_id"] + 1

            if "message" in update:
                msg = update["message"]
                chat_id = msg["chat"]["id"]
                text = msg.get("text", "")

                if text:
                    # 로그 출력
                    print(f"[{datetime.now().strftime('%H:%M:%S')}] 📩 {text}")

                    # 저장
                    save_chat(text, "user")

                    # 응답
                    if text == "/status":
                        reply = f"✅ 봇 정상 작동\n시간: {datetime.now().strftime('%H:%M:%S')}"
                    else:
                        reply = f"받음: {text}"

                    send_msg(chat_id, reply)
                    save_chat(reply, "bot")

                    print(f"[{datetime.now().strftime('%H:%M:%S')}] 📤 응답 전송")

    except Exception as e:
        if "409" in str(e):
            print("⚠️ 409 에러 - GCP 봇 실행 중...")
            time.sleep(10)
        else:
            print(f"오류: {e}")
            time.sleep(5)