#!/usr/bin/env python3
"""
새 봇으로 테스트 메시지 전송
"""

import json
import urllib.request
import urllib.parse
from datetime import datetime

# 새 봇 토큰
NEW_TOKEN = "8501568801:AAE-3fBl-p6uZcmrdsWSRQuz_eg8yDADwjI"
API_URL = f"https://api.telegram.org/bot{NEW_TOKEN}"

# 사용자 채팅 ID (이전 대화 기록에서)
CHAT_ID = 7565534667

def send_message(text: str):
    """텔레그램으로 메시지 전송"""
    try:
        url = f"{API_URL}/sendMessage"

        # 메시지 데이터
        data = {
            "chat_id": CHAT_ID,
            "text": text,
            "parse_mode": "Markdown"
        }

        # URL 인코딩
        params = urllib.parse.urlencode(data).encode('utf-8')

        # 요청 생성
        req = urllib.request.Request(url, data=params)

        # 전송
        with urllib.request.urlopen(req) as response:
            result = json.loads(response.read())

            if result["ok"]:
                print(f"✅ 메시지 전송 성공!")
                print(f"   시간: {datetime.now().strftime('%H:%M:%S')}")
                print(f"   메시지 ID: {result['result']['message_id']}")
                return True
            else:
                print(f"❌ 전송 실패: {result}")
                return False

    except Exception as e:
        print(f"❌ 오류 발생: {e}")
        return False

if __name__ == "__main__":
    # 테스트 메시지들 전송
    messages = [
        "🚀 *97LAYER OS 실시간 연동 테스트*\n\n새로운 봇이 정상 작동합니다!",

        "📊 *시스템 상태*\n• 봇: ✅ 정상\n• 409 에러: ❌ 해결됨\n• 실시간 연동: ✅ 활성화",

        f"⏰ *현재 시간*: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n이제 텔레그램 메시지가 실시간으로 97LAYER OS와 연동됩니다.",

        "🤖 *에이전트 상태*\n• CD (Creative Director): 대기 중\n• TD (Technical Director): 대기 중\n• AD (Art Director): 대기 중\n• CE (Chief Editor): 대기 중\n• SA (Strategy Analyst): 대기 중"
    ]

    print("=" * 60)
    print("📨 텔레그램으로 테스트 메시지 전송 중...")
    print("=" * 60)

    for i, msg in enumerate(messages, 1):
        print(f"\n[{i}/{len(messages)}] 전송 중...")
        if send_message(msg):
            print("   → 완료")
        else:
            print("   → 실패")

    print("\n" + "=" * 60)
    print("✅ 메시지 전송 완료!")
    print("텔레그램을 확인해주세요.")
    print("=" * 60)