#!/usr/bin/env python3
"""
97LAYER OS - Intelligent Telegram Bot
Google Gemini API integration
"""

import os
import json
import time
import urllib.request
from datetime import datetime
from pathlib import Path
from dotenv import load_dotenv
import google.genai as genai

# Load environment
load_dotenv('/app/.env')

# Configuration
TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")
GEMINI_API_KEY = os.getenv("GEMINI_API_KEY")

if not TOKEN:
    raise ValueError("❌ TELEGRAM_BOT_TOKEN not found")
if not GEMINI_API_KEY:
    raise ValueError("❌ GEMINI_API_KEY not found")

BASE_URL = f"https://api.telegram.org/bot{TOKEN}"
CHAT_MEMORY = Path("/app/knowledge/chat_memory/7565534667.json")

# Initialize Gemini with new API
client = genai.Client(api_key=GEMINI_API_KEY)

print("✅ Gemini AI initialized")

def load_conversation_history():
    """대화 히스토리 로드 (최근 20개)"""
    if CHAT_MEMORY.exists():
        with open(CHAT_MEMORY, 'r', encoding='utf-8') as f:
            messages = json.load(f)
            return messages[-20:] if len(messages) > 20 else messages
    return []

def save_message(chat_id, text, role="user"):
    """메시지 저장"""
    CHAT_MEMORY.parent.mkdir(parents=True, exist_ok=True)

    messages = []
    if CHAT_MEMORY.exists():
        with open(CHAT_MEMORY, 'r', encoding='utf-8') as f:
            messages = json.load(f)

    messages.append({
        "timestamp": datetime.now().isoformat(),
        "role": role,
        "content": text
    })

    with open(CHAT_MEMORY, 'w', encoding='utf-8') as f:
        json.dump(messages, f, ensure_ascii=False, indent=2)

    print(f"💾 [{role}] {text[:50]}...")

def send_message(chat_id, text):
    """메시지 전송"""
    url = f"{BASE_URL}/sendMessage"
    data = json.dumps({
        "chat_id": chat_id,
        "text": text
    }).encode('utf-8')

    req = urllib.request.Request(url, data=data)
    req.add_header('Content-Type', 'application/json')

    try:
        with urllib.request.urlopen(req) as response:
            print(f"📤 전송: {text[:50]}...")
            return True
    except Exception as e:
        print(f"전송 실패: {e}")
        return False

def generate_ai_response(user_message, context_messages):
    """Gemini AI 응답 생성"""
    try:
        # Build context from history
        context = "대화 기록:\n"
        for msg in context_messages[-10:]:  # Last 10 messages
            role = "사용자" if msg["role"] == "user" else "AI"
            context += f"{role}: {msg['content']}\n"

        # System instruction + context + new message
        prompt = f"""당신은 97LAYER OS의 AI 어시스턴트입니다.

특징:
- 한국어로 자연스럽게 대화
- 간결하고 실용적인 답변
- 불필요한 격식 없이 편하게
- 창의적이고 기술적인 작업 지원

{context}

사용자: {user_message}
AI:"""

        # Generate response with new API
        response = client.models.generate_content(
            model='models/gemini-2.5-flash',
            contents=prompt
        )
        return response.text

    except Exception as e:
        print(f"AI 응답 생성 실패: {e}")
        return f"죄송합니다. 응답 생성 중 오류가 발생했습니다: {str(e)}"

def process_command(chat_id, command):
    """명령어 처리"""
    cmd = command.lower().split()[0]

    if cmd == "/start":
        return """🤖 97LAYER OS Bot

Google Gemini AI와 함께 작동합니다.

명령어:
- /status - 시스템 상태
- /clear - 대화 기록 초기화
- 일반 메시지 - AI와 자연스러운 대화"""

    elif cmd == "/status":
        history = load_conversation_history()
        return f"""✅ 시스템 상태

시간: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
컨테이너: 97layer-workspace
AI 엔진: Google Gemini 2.5 Flash
대화 기록: {len(history)}개 메시지"""

    elif cmd == "/clear":
        if CHAT_MEMORY.exists():
            CHAT_MEMORY.unlink()
        return "🗑️ 대화 기록이 초기화되었습니다."

    else:
        return f"알 수 없는 명령어: {cmd}\n/start 를 입력해 도움말을 확인하세요."

def main():
    print("=" * 60)
    print("🤖 97LAYER OS Intelligent Bot (Gemini)")
    print("=" * 60)
    print(f"📍 Container: 97layer-workspace")
    print(f"📂 Memory: {CHAT_MEMORY}")
    print(f"🧠 AI: Google Gemini 1.5 Flash")
    print("=" * 60)

    offset = None

    while True:
        try:
            url = f"{BASE_URL}/getUpdates?timeout=30"
            if offset:
                url += f"&offset={offset}"

            with urllib.request.urlopen(url, timeout=35) as response:
                result = json.loads(response.read())

                for update in result.get("result", []):
                    offset = update["update_id"] + 1

                    if "message" in update:
                        msg = update["message"]
                        chat_id = msg["chat"]["id"]
                        text = msg.get("text", "")

                        if text:
                            print(f"\n📩 받음: {text}")
                            save_message(chat_id, text, "user")

                            # Process message
                            if text.startswith("/"):
                                response = process_command(chat_id, text)
                            else:
                                # AI conversation
                                context = load_conversation_history()
                                response = generate_ai_response(text, context)

                            # Send response
                            send_message(chat_id, response)
                            save_message(chat_id, response, "assistant")

        except Exception as e:
            if "409" in str(e):
                print("⚠️ 409 Conflict - 다른 봇이 실행 중입니다. 10초 대기...")
                time.sleep(10)
            elif "timeout" in str(e).lower():
                # Normal timeout, just continue
                pass
            else:
                print(f"오류: {e}")
                time.sleep(5)

if __name__ == "__main__":
    main()
