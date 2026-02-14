#!/usr/bin/env python3
"""
Telegram 봇 409 Conflict 에러 해결 스크립트
모든 중복 인스턴스를 정리하고 새로 시작
"""

import os
import sys
import json
import time
import subprocess
import urllib.request
from pathlib import Path

PROJECT_ROOT = Path(__file__).parent

# 설정 로드
sys.path.append(str(PROJECT_ROOT))
from libs.core_config import TELEGRAM_CONFIG

TOKEN = TELEGRAM_CONFIG["BOT_TOKEN"]
BASE_URL = f"https://api.telegram.org/bot{TOKEN}"

def clear_webhook():
    """웹훅 제거 (폴링 모드 사용)"""
    print("1. 웹훅 제거 중...")
    try:
        url = f"{BASE_URL}/deleteWebhook"
        with urllib.request.urlopen(url) as response:
            result = json.loads(response.read())
            print(f"   웹훅 제거: {result}")
    except Exception as e:
        print(f"   웹훅 제거 실패: {e}")

def get_updates_and_clear():
    """마지막 업데이트 ID 가져오고 클리어"""
    print("2. 업데이트 큐 정리 중...")
    try:
        # 대기 중인 업데이트 가져오기
        url = f"{BASE_URL}/getUpdates?timeout=1"
        with urllib.request.urlopen(url, timeout=5) as response:
            result = json.loads(response.read())
            updates = result.get("result", [])

            if updates:
                # 마지막 업데이트 ID 저장
                last_update_id = updates[-1]["update_id"] + 1

                # task_status.json에 저장
                status_file = PROJECT_ROOT / "task_status.json"
                status_data = {}
                if status_file.exists():
                    with open(status_file) as f:
                        status_data = json.load(f)

                status_data["last_telegram_update_id"] = last_update_id

                with open(status_file, 'w') as f:
                    json.dump(status_data, f, indent=2)

                print(f"   마지막 update_id 저장: {last_update_id}")

                # 모든 업데이트 확인 (클리어)
                url = f"{BASE_URL}/getUpdates?offset={last_update_id}"
                urllib.request.urlopen(url, timeout=5)
                print(f"   {len(updates)}개 업데이트 클리어")
            else:
                print("   대기 중인 업데이트 없음")

    except Exception as e:
        print(f"   업데이트 정리 실패: {e}")

def kill_all_instances():
    """모든 텔레그램 봇 인스턴스 종료"""
    print("3. 기존 프로세스 종료 중...")

    # Mac에서 실행 중인 프로세스 종료
    subprocess.run(["pkill", "-f", "telegram_daemon"], stderr=subprocess.DEVNULL)
    subprocess.run(["pkill", "-f", "async_telegram"], stderr=subprocess.DEVNULL)

    # GCP 인스턴스 확인 및 종료 시도
    try:
        ssh_key = Path.home() / ".ssh" / "id_ed25519_gcp"
        if ssh_key.exists():
            print("   GCP 인스턴스 확인 중...")
            # GCP에서 프로세스 종료
            subprocess.run([
                "ssh", "-i", str(ssh_key),
                "-o", "ConnectTimeout=5",
                "skyto5339@35.184.30.182",
                "pkill -f telegram_daemon"
            ], stderr=subprocess.DEVNULL, timeout=10)
            print("   GCP 프로세스 종료 시도")
    except:
        pass

    time.sleep(3)
    print("   모든 프로세스 종료 완료")

def start_fresh_instance():
    """새로운 인스턴스 시작"""
    print("4. 새 인스턴스 시작 중...")

    # 로그 디렉토리 확인
    log_dir = PROJECT_ROOT / "logs"
    log_dir.mkdir(exist_ok=True)

    # 새 프로세스 시작
    log_file = log_dir / f"telegram_daemon_{time.strftime('%Y%m%d_%H%M%S')}.log"

    with open(log_file, 'w') as log:
        process = subprocess.Popen(
            [sys.executable, "execution/telegram_daemon.py"],
            cwd=PROJECT_ROOT,
            stdout=log,
            stderr=subprocess.STDOUT
        )

    print(f"   PID: {process.pid}")
    print(f"   로그: {log_file}")

    # 시작 확인
    time.sleep(5)

    if process.poll() is None:
        print("   ✅ 텔레그램 봇 정상 시작!")

        # 테스트 메시지
        print("\n5. 봇 테스트:")
        print("   텔레그램에서 봇에게 /start 또는 /status 명령을 보내보세요")

        # 로그 모니터링 시작
        print(f"\n로그 모니터링: tail -f {log_file}")

        return True
    else:
        print("   ❌ 시작 실패")
        return False

def main():
    print("=" * 60)
    print("Telegram Bot 409 Conflict 해결 스크립트")
    print("=" * 60)

    # 1. 웹훅 제거
    clear_webhook()

    # 2. 업데이트 큐 정리
    get_updates_and_clear()

    # 3. 모든 인스턴스 종료
    kill_all_instances()

    # 4. 새로 시작
    success = start_fresh_instance()

    print("=" * 60)

    if success:
        print("✅ 완료! 봇이 정상적으로 시작되었습니다.")
        print("\n텔레그램 봇 테스트:")
        print("1. 텔레그램 앱 열기")
        print("2. 봇 검색 및 채팅 시작")
        print("3. /start 또는 /status 입력")
    else:
        print("❌ 실패. 로그를 확인하세요.")

if __name__ == "__main__":
    main()