#!/usr/bin/env python3
"""
텔레그램 봇을 Polling → Webhook 모드로 안전하게 전환

실행 방법:
python execution/switch_to_webhook.py [WEBHOOK_URL]

예시:
python execution/switch_to_webhook.py https://telegram-bot-xxxxx-xx.a.run.app
"""

import os
import sys
import json
import signal
import subprocess
import urllib.request
from pathlib import Path
from datetime import datetime

PROJECT_ROOT = Path(__file__).resolve().parent.parent
TASK_STATUS_FILE = PROJECT_ROOT / "task_status.json"

# .env에서 토큰 로드
try:
    from libs.core_config import TELEGRAM_CONFIG
    TOKEN = TELEGRAM_CONFIG["BOT_TOKEN"]
except:
    print("❌ TELEGRAM_CONFIG를 로드할 수 없습니다.")
    sys.exit(1)

def find_and_kill_telegram_daemon():
    """실행 중인 telegram_daemon 프로세스 찾아서 종료"""
    print("\n🔍 실행 중인 telegram_daemon 프로세스 검색...")

    try:
        result = subprocess.run(
            ["ps", "aux"],
            capture_output=True,
            text=True
        )

        killed_count = 0
        for line in result.stdout.split('\n'):
            if 'telegram_daemon.py' in line and 'grep' not in line:
                parts = line.split()
                if len(parts) > 1:
                    pid = int(parts[1])
                    print(f"   Found PID {pid}: {line[:80]}...")
                    try:
                        os.kill(pid, signal.SIGTERM)
                        print(f"   ✓ Terminated PID {pid}")
                        killed_count += 1
                    except ProcessLookupError:
                        print(f"   ⚠️ PID {pid} already terminated")
                    except PermissionError:
                        print(f"   ❌ Permission denied for PID {pid}")

        if killed_count == 0:
            print("   ℹ️ 실행 중인 telegram_daemon 없음")
        else:
            print(f"\n✓ {killed_count}개 프로세스 종료됨")

        return killed_count

    except Exception as e:
        print(f"❌ 프로세스 검색 중 오류: {e}")
        return 0

def delete_old_webhook():
    """기존 webhook 제거"""
    print("\n🧹 기존 Webhook 제거...")

    url = f"https://api.telegram.org/bot{TOKEN}/deleteWebhook"

    try:
        with urllib.request.urlopen(url) as response:
            result = json.loads(response.read().decode())
            if result.get("ok"):
                print("   ✓ 기존 webhook 제거 완료")
                return True
            else:
                print(f"   ⚠️ Webhook 제거 실패: {result}")
                return False
    except Exception as e:
        print(f"   ❌ 오류: {e}")
        return False

def set_new_webhook(webhook_url: str):
    """새 webhook 설정"""
    print(f"\n🔗 새 Webhook 설정 중...")
    print(f"   URL: {webhook_url}/webhook")

    url = f"https://api.telegram.org/bot{TOKEN}/setWebhook?url={webhook_url}/webhook"

    try:
        with urllib.request.urlopen(url) as response:
            result = json.loads(response.read().decode())
            if result.get("ok"):
                print("   ✓ Webhook 설정 완료")
                return True
            else:
                print(f"   ❌ Webhook 설정 실패: {result}")
                return False
    except Exception as e:
        print(f"   ❌ 오류: {e}")
        return False

def verify_webhook():
    """Webhook 상태 확인"""
    print("\n🔍 Webhook 상태 확인...")

    url = f"https://api.telegram.org/bot{TOKEN}/getWebhookInfo"

    try:
        with urllib.request.urlopen(url) as response:
            result = json.loads(response.read().decode())
            if result.get("ok"):
                info = result.get("result", {})
                webhook_url = info.get("url", "")
                pending_count = info.get("pending_update_count", 0)
                last_error = info.get("last_error_message", "")

                print(f"   URL: {webhook_url or '(없음)'}")
                print(f"   Pending Updates: {pending_count}")
                if last_error:
                    print(f"   ⚠️ Last Error: {last_error}")
                else:
                    print(f"   ✓ 오류 없음")

                return bool(webhook_url)
            else:
                print(f"   ❌ 상태 확인 실패: {result}")
                return False
    except Exception as e:
        print(f"   ❌ 오류: {e}")
        return False

def update_task_status(webhook_url: str):
    """task_status.json에 webhook 모드 기록"""
    print("\n📝 task_status.json 업데이트...")

    try:
        if TASK_STATUS_FILE.exists():
            status = json.loads(TASK_STATUS_FILE.read_text())
        else:
            status = {}

        status["telegram_mode"] = "webhook"
        status["telegram_webhook_url"] = webhook_url
        status["telegram_switched_at"] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        status["last_active"] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")

        TASK_STATUS_FILE.write_text(json.dumps(status, indent=2, ensure_ascii=False))
        print("   ✓ 업데이트 완료")
        return True

    except Exception as e:
        print(f"   ❌ 오류: {e}")
        return False

def main():
    print("=" * 60)
    print("🔄 텔레그램 봇 모드 전환: Polling → Webhook")
    print("=" * 60)

    # Webhook URL 입력
    if len(sys.argv) > 1:
        webhook_url = sys.argv[1].rstrip('/')
    else:
        webhook_url = input("\n🌐 배포된 Cloud Run URL을 입력하세요: ").strip().rstrip('/')

    if not webhook_url:
        print("❌ URL이 입력되지 않았습니다.")
        sys.exit(1)

    # Step 1: 기존 polling daemon 중지
    find_and_kill_telegram_daemon()

    # Step 2: 기존 webhook 제거
    delete_old_webhook()

    # Step 3: 새 webhook 설정
    if not set_new_webhook(webhook_url):
        print("\n❌ Webhook 설정 실패. 수동으로 설정해주세요:")
        print(f"   curl -X POST 'https://api.telegram.org/bot{TOKEN}/setWebhook?url={webhook_url}/webhook'")
        sys.exit(1)

    # Step 4: 설정 확인
    if not verify_webhook():
        print("\n⚠️ Webhook이 제대로 설정되지 않았을 수 있습니다.")

    # Step 5: task_status 업데이트
    update_task_status(webhook_url)

    # 완료
    print("\n" + "=" * 60)
    print("✅ 전환 완료!")
    print("=" * 60)
    print("\n다음 단계:")
    print("1. 텔레그램 봇에 메시지를 보내서 테스트")
    print("2. Cloud Run 로그 모니터링:")
    print("   gcloud run logs tail telegram-bot --region asia-northeast3")
    print(f"\n3. Health Check:")
    print(f"   curl {webhook_url}/health")
    print("\n")

if __name__ == "__main__":
    main()
