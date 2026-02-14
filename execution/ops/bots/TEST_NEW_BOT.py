#!/usr/bin/env python3
"""
새 봇 토큰 테스트 및 링크 생성
"""

import json
import urllib.request

# 새 봇 토큰
NEW_TOKEN = "8501568801:AAE-3fBl-p6uZcmrdsWSRQuz_eg8yDADwjI"
API_URL = f"https://api.telegram.org/bot{NEW_TOKEN}"

def get_bot_info():
    """봇 정보 가져오기"""
    try:
        url = f"{API_URL}/getMe"
        with urllib.request.urlopen(url) as response:
            result = json.loads(response.read())

            if result["ok"]:
                bot_info = result["result"]
                print("=" * 60)
                print("🤖 새로운 97LAYER 봇 정보")
                print("=" * 60)
                print(f"봇 이름: {bot_info.get('first_name', 'Unknown')}")
                print(f"유저네임: @{bot_info.get('username', 'Unknown')}")
                print(f"봇 ID: {bot_info.get('id', 'Unknown')}")
                print()
                print("✅ 봇이 정상적으로 작동합니다!")
                print()
                print("📱 텔레그램에서 봇 시작하기:")
                print(f"   https://t.me/{bot_info.get('username', 'Unknown')}")
                print()
                print("또는 텔레그램에서 검색:")
                print(f"   @{bot_info.get('username', 'Unknown')}")
                print("=" * 60)

                return bot_info
            else:
                print(f"❌ 봇 정보 가져오기 실패: {result}")
                return None

    except Exception as e:
        print(f"❌ 오류 발생: {e}")
        return None

def check_updates():
    """최근 업데이트 확인"""
    try:
        url = f"{API_URL}/getUpdates?limit=5"
        with urllib.request.urlopen(url) as response:
            result = json.loads(response.read())

            if result["ok"]:
                updates = result["result"]
                print("\n📨 최근 메시지:")
                print("-" * 40)

                if not updates:
                    print("아직 받은 메시지가 없습니다.")
                    print("위 링크로 봇에게 메시지를 보내보세요!")
                else:
                    for update in updates:
                        if "message" in update:
                            msg = update["message"]
                            text = msg.get("text", "")
                            from_user = msg.get("from", {}).get("username", "Unknown")
                            print(f"From @{from_user}: {text}")

                print("-" * 40)
                return True
            else:
                print(f"❌ 업데이트 확인 실패: {result}")
                return False

    except Exception as e:
        print(f"❌ 업데이트 확인 중 오류: {e}")
        return False

if __name__ == "__main__":
    # 봇 정보 가져오기
    bot_info = get_bot_info()

    # 최근 업데이트 확인
    if bot_info:
        check_updates()

        print("\n💡 사용 방법:")
        print("1. 위 링크로 텔레그램 봇 열기")
        print("2. /start 명령어로 시작")
        print("3. 메시지를 보내면 97LAYER_MESSENGER.py가 처리")
        print("\n현재 실행 중인 메신저:")
        print("  python3 97LAYER_MESSENGER.py")