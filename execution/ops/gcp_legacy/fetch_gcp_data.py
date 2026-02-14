#!/usr/bin/env python3
"""
GCP Telegram Bot을 통해 chat_memory 가져오기
"""
import os
import sys
import json
import time
import requests
from pathlib import Path

# Telegram Bot 설정
BOT_TOKEN = "8271602365:AAGQwvDfmLv11_CShkeTMSQvnAkDYbDiTxA"
CHAT_ID = "7565534667"
BASE_URL = f"https://api.telegram.org/bot{BOT_TOKEN}"

def send_telegram_command(command):
    """Telegram으로 명령 전송"""
    url = f"{BASE_URL}/sendMessage"
    data = {
        "chat_id": CHAT_ID,
        "text": command
    }
    response = requests.post(url, json=data)
    return response.json()

def get_telegram_updates(offset=None):
    """Telegram 메시지 가져오기"""
    url = f"{BASE_URL}/getUpdates"
    params = {"timeout": 30, "offset": offset} if offset else {"timeout": 30}
    response = requests.get(url, params=params)
    return response.json()

def main():
    print("📱 Telegram Bot을 통해 GCP에 명령 전송 중...")

    # 1. GCP에게 knowledge 패키지 생성 명령
    print("\n1️⃣ GCP에게 데이터 패키지 생성 요청...")
    send_telegram_command("/exec cd ~/97layerOS && tar czf /tmp/knowledge.tar.gz knowledge/ && echo 'Package created'")

    time.sleep(5)

    # 2. 최근 메시지 확인
    print("\n2️⃣ 응답 확인 중...")
    updates = get_telegram_updates()

    if updates.get("ok"):
        messages = updates.get("result", [])
        if messages:
            last_message = messages[-1].get("message", {}).get("text", "")
            print(f"   응답: {last_message}")
        else:
            print("   응답 없음")

    # 3. 대안: chat_memory를 Telegram 메시지로 직접 요청
    print("\n3️⃣ Chat memory 직접 요청...")
    send_telegram_command("/dump_memory")

    time.sleep(3)

    updates = get_telegram_updates()
    if updates.get("ok"):
        messages = updates.get("result", [])
        if messages:
            for msg in messages[-5:]:
                text = msg.get("message", {}).get("text", "")
                if "WOOHWAHAE" in text or "72H" in text:
                    print(f"\n✅ 발견: {text[:200]}...")
                    return True

    print("\n⚠️  Telegram을 통한 자동 가져오기 실패")
    print("    GCP Telegram Bot이 /exec 또는 /dump_memory 명령을 지원하지 않을 수 있습니다.")

    return False

if __name__ == "__main__":
    main()
