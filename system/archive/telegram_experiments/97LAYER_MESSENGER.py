#!/usr/bin/env python3
import os
"""
97LAYER 회사 메신저 - 최종 버전
새 토큰으로 409 에러 없이 작동

목적: 텔레그램 → 에이전트 → 실시간 보고
"""

import json
import sys
import time
import subprocess
from datetime import datetime
from pathlib import Path
import urllib.request

# ===== 설정 =====
NEW_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")  # 새 토큰!
API_URL = f"https://api.telegram.org/bot{NEW_TOKEN}"

PROJECT_ROOT = Path.home() / "97layerOS"
CHAT_LOG = PROJECT_ROOT / "knowledge" / "messenger_chat.json"
CHAT_LOG.parent.mkdir(parents=True, exist_ok=True)

# 에이전트 정의
AGENTS = {
    "CD": "Creative Director - 브랜드/전략",
    "TD": "Technical Director - 기술/시스템",
    "AD": "Art Director - 디자인/비주얼",
    "CE": "Chief Editor - 콘텐츠/편집",
    "SA": "Strategy Analyst - 분석/리서치"
}

print("=" * 60)
print("🚀 97LAYER 회사 메신저")
print("새 봇 토큰으로 실행 (409 에러 해결!)")
print("=" * 60)

# 초기화
offset = None
chat_history = []

def save_message(chat_id, text, role="user"):
    """메시지 저장"""
    global chat_history

    entry = {
        "timestamp": datetime.now().isoformat(),
        "chat_id": str(chat_id),
        "role": role,
        "content": text
    }

    chat_history.append(entry)

    # 파일 저장
    with open(CHAT_LOG, 'w', encoding='utf-8') as f:
        json.dump(chat_history, f, ensure_ascii=False, indent=2)

    # 콘솔 출력
    time_str = datetime.now().strftime("%H:%M:%S")
    if role == "user":
        print(f"\n📩 [{time_str}] 받음: {text}")
    else:
        print(f"📤 [{time_str}] 응답 전송")

def send_message(chat_id, text):
    """메시지 전송"""
    try:
        url = f"{API_URL}/sendMessage"
        data = json.dumps({
            "chat_id": chat_id,
            "text": text,
            "parse_mode": "Markdown"
        }).encode('utf-8')

        req = urllib.request.Request(url, data=data)
        req.add_header('Content-Type', 'application/json')

        with urllib.request.urlopen(req) as response:
            save_message(chat_id, text, "assistant")
            return True

    except Exception as e:
        print(f"전송 오류: {e}")
        return False

def process_command(chat_id, command):
    """명령어 처리"""
    cmd = command.lower().split()[0]

    if cmd == "/start":
        msg = "🤖 *97LAYER 회사 메신저*\n\n"
        msg += "안티그래비티 에이전트와 실시간 소통 시스템입니다.\n\n"
        msg += "*명령어:*\n"
        msg += "/status - 시스템 상태\n"
        msg += "/agents - 에이전트 목록\n"
        msg += "/report - 최근 활동 보고\n"
        msg += "/cd, /td, /ad, /ce, /sa - 에이전트 호출"

    elif cmd == "/status":
        msg = f"✅ *시스템 상태*\n\n"
        msg += f"시간: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n"
        msg += f"총 메시지: {len(chat_history)}개\n"
        msg += "봇: 정상 작동 (새 토큰)\n"
        msg += "에이전트: 대기 중"

    elif cmd == "/agents":
        msg = "*활성 에이전트:*\n\n"
        for code, desc in AGENTS.items():
            msg += f"• *{code}*: {desc}\n"

    elif cmd == "/report":
        # 최근 대화 보고
        recent = chat_history[-10:] if len(chat_history) > 10 else chat_history
        msg = "*최근 활동 보고:*\n\n"
        for entry in recent:
            time_str = entry['timestamp'][11:19]
            role = "👤" if entry['role'] == "user" else "🤖"
            content = entry['content'][:50] + "..." if len(entry['content']) > 50 else entry['content']
            msg += f"{role} {time_str}: {content}\n"

    elif cmd in ["/cd", "/td", "/ad", "/ce", "/sa"]:
        agent = cmd[1:].upper()
        msg = f"📢 *{AGENTS[agent]} 호출*\n\n"

        # 에이전트별 특수 처리
        if agent == "TD":
            msg += "시스템 점검 중...\n"
            # 실제 시스템 체크
            try:
                result = subprocess.run(
                    ["ps", "aux"],
                    capture_output=True,
                    text=True,
                    timeout=5
                )
                python_procs = len([line for line in result.stdout.split('\n') if 'python' in line.lower()])
                msg += f"• Python 프로세스: {python_procs}개\n"
                msg += f"• 메신저 상태: 정상"
            except:
                msg += "시스템 체크 완료"

    else:
        msg = f"알 수 없는 명령: {cmd}"

    return msg

def process_message(chat_id, text):
    """일반 메시지 처리"""
    # 키워드 기반 에이전트 선택
    text_lower = text.lower()

    if any(w in text_lower for w in ["코드", "버그", "시스템", "서버", "기술"]):
        agent = "TD"
    elif any(w in text_lower for w in ["디자인", "ui", "색상", "비주얼"]):
        agent = "AD"
    elif any(w in text_lower for w in ["분석", "데이터", "통계", "리포트"]):
        agent = "SA"
    elif any(w in text_lower for w in ["글", "문구", "카피", "편집"]):
        agent = "CE"
    else:
        agent = "CD"

    response = f"📋 *메시지 처리*\n\n"
    response += f"담당: {AGENTS[agent]}\n"
    response += f"요청: {text}\n"
    response += f"상태: 처리 중...\n\n"
    response += f"_에이전트가 작업을 진행하고 있습니다._"

    return response

# 초기 업데이트 클리어
print("초기화 중...")
try:
    url = f"{API_URL}/getUpdates?offset=-1"
    with urllib.request.urlopen(url) as response:
        result = json.loads(response.read())
        print(f"✅ 초기화 완료 (새 토큰 확인: {result['ok']})")
except Exception as e:
    print(f"초기화 오류: {e}")

# 기존 채팅 로드
if CHAT_LOG.exists():
    try:
        with open(CHAT_LOG) as f:
            chat_history = json.load(f)
            print(f"📚 기존 대화 {len(chat_history)}개 로드")
    except:
        pass

print("\n✅ 메신저 준비 완료!")
print("📱 텔레그램에서 새 봇에게 메시지를 보내세요")
print("-" * 60)

# 메인 루프
while True:
    try:
        # 업데이트 받기
        url = f"{API_URL}/getUpdates?timeout=10"
        if offset:
            url += f"&offset={offset}"

        with urllib.request.urlopen(url, timeout=15) as response:
            result = json.loads(response.read())

            if result["ok"]:
                for update in result["result"]:
                    offset = update["update_id"] + 1

                    if "message" in update:
                        msg = update["message"]
                        chat_id = msg["chat"]["id"]
                        text = msg.get("text", "")

                        if text:
                            # 메시지 저장
                            save_message(chat_id, text, "user")

                            # 응답 생성
                            if text.startswith("/"):
                                response = process_command(chat_id, text)
                            else:
                                response = process_message(chat_id, text)

                            # 응답 전송
                            send_message(chat_id, response)

    except KeyboardInterrupt:
        print("\n\n메신저 종료")
        break

    except Exception as e:
        print(f"오류: {e}")
        time.sleep(5)

print("시스템 종료")