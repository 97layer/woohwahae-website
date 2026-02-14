#!/usr/bin/env python3
"""
최종 작동 봇 - 실시간 대화 저장 및 응답
"""

import json
import time
import urllib.request
import urllib.error
from datetime import datetime
from pathlib import Path

# 설정
TOKEN = "8271602365:AAGQwvDfmLv11_CShkeTMSQvnAkDYbDiTxA"
BASE_URL = f"https://api.telegram.org/bot{TOKEN}"
MEMORY_FILE = Path.home() / "97layerOS" / "knowledge" / "chat_memory" / "realtime_chat.json"

def clear_all_updates():
    """모든 대기 중인 업데이트 클리어"""
    print("초기화 중...")
    try:
        # 타임아웃 0으로 모든 업데이트 가져오기
        url = f"{BASE_URL}/getUpdates?timeout=1"
        with urllib.request.urlopen(url, timeout=5) as response:
            result = json.loads(response.read())
            updates = result.get("result", [])

            if updates:
                last_id = updates[-1]["update_id"] + 1
                # 모두 확인 처리
                confirm_url = f"{BASE_URL}/getUpdates?offset={last_id}"
                urllib.request.urlopen(confirm_url, timeout=5)
                print(f"✅ {len(updates)}개 기존 업데이트 클리어")
                return last_id
    except:
        pass
    return None

def save_and_log(chat_id, text, role="user"):
    """메시지 저장 및 로깅"""
    MEMORY_FILE.parent.mkdir(parents=True, exist_ok=True)

    # 기존 메시지 로드
    messages = []
    if MEMORY_FILE.exists():
        try:
            with open(MEMORY_FILE, 'r', encoding='utf-8') as f:
                messages = json.load(f)
        except:
            messages = []

    # 새 메시지 추가
    entry = {
        "timestamp": datetime.now().isoformat(),
        "chat_id": str(chat_id),
        "role": role,
        "content": text
    }
    messages.append(entry)

    # 저장
    with open(MEMORY_FILE, 'w', encoding='utf-8') as f:
        json.dump(messages, f, ensure_ascii=False, indent=2)

    print(f"💾 [{datetime.now().strftime('%H:%M:%S')}] {role}: {text[:50]}")

def send_reply(chat_id, text):
    """응답 전송"""
    try:
        url = f"{BASE_URL}/sendMessage"
        data = json.dumps({
            "chat_id": chat_id,
            "text": text,
            "parse_mode": "Markdown"
        }).encode('utf-8')

        req = urllib.request.Request(url, data=data)
        req.add_header('Content-Type', 'application/json')

        with urllib.request.urlopen(req) as response:
            print(f"📤 응답 전송")
            return True
    except Exception as e:
        print(f"전송 실패: {e}")
        return False

def process_message(chat_id, text):
    """메시지 처리 및 응답 생성"""
    # 메시지 저장
    save_and_log(chat_id, text, "user")

    # 응답 생성
    now = datetime.now().strftime('%Y-%m-%d %H:%M:%S')

    if text.startswith("/"):
        if text == "/status":
            response = f"✅ *시스템 상태*\n\n시간: {now}\n봇: 정상 작동\n메모리: 실시간 저장 중"
        elif text == "/start":
            response = "🤖 *97LAYER OS*\n\n실시간 대화 시스템이 활성화되었습니다.\n모든 대화가 저장됩니다."
        elif text == "/help":
            response = "*사용 가능한 명령*\n\n/status - 상태 확인\n/help - 도움말\n/report - 보고서"
        elif text == "/report":
            # 최근 대화 요약
            if MEMORY_FILE.exists():
                with open(MEMORY_FILE, 'r') as f:
                    all_msgs = json.load(f)
                    recent = all_msgs[-5:] if len(all_msgs) > 5 else all_msgs
                    report = "*최근 대화*\n\n"
                    for msg in recent:
                        t = msg['timestamp'][:19].replace('T', ' ')
                        r = "👤" if msg['role'] == 'user' else "🤖"
                        report += f"{r} {t}\n{msg['content'][:50]}...\n\n"
                    response = report
            else:
                response = "대화 기록이 없습니다."
        else:
            response = f"명령어: {text}"
    else:
        # 일반 대화 응답
        response = f"메시지를 받았습니다.\n\n*내용*: {text}\n*시간*: {now}\n\n처리 중입니다..."

    # 응답 전송
    if send_reply(chat_id, response):
        # 응답도 저장
        save_and_log(chat_id, response, "assistant")

def main():
    print("=" * 60)
    print("🤖 FINAL BOT - 실시간 대화 시스템")
    print("=" * 60)

    # 초기 클리어
    offset = clear_all_updates()

    print(f"\n✅ 봇 시작! 텔레그램에서 메시지를 보내세요.")
    print(f"💾 저장 위치: {MEMORY_FILE}")
    print("-" * 60)

    error_count = 0

    while True:
        try:
            # 업데이트 가져오기
            url = f"{BASE_URL}/getUpdates?timeout=10"
            if offset:
                url += f"&offset={offset}"

            with urllib.request.urlopen(url, timeout=15) as response:
                result = json.loads(response.read())

                if result.get("ok"):
                    error_count = 0  # 에러 카운터 리셋

                    for update in result.get("result", []):
                        offset = update["update_id"] + 1

                        if "message" in update:
                            msg = update["message"]
                            chat_id = msg["chat"]["id"]
                            text = msg.get("text", "")

                            if text:
                                print(f"\n📩 받음: {text}")
                                process_message(chat_id, text)

        except urllib.error.HTTPError as e:
            if "409" in str(e):
                error_count += 1
                if error_count > 3:
                    print("\n⚠️ 409 Conflict 지속 - GCP 봇을 중지해주세요!")
                    print("해결 방법:")
                    print("1. GCP Console에서 VM 중지")
                    print("2. 또는 새 봇 토큰 생성 (@BotFather)")
                    time.sleep(30)
                else:
                    print(".", end="", flush=True)
                    time.sleep(10)
            else:
                print(f"\nHTTP 에러: {e}")
                time.sleep(5)

        except KeyboardInterrupt:
            print("\n\n봇 종료")
            break

        except Exception as e:
            print(f"\n오류: {e}")
            time.sleep(5)

if __name__ == "__main__":
    main()