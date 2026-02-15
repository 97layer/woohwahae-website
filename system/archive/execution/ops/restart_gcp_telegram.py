#!/usr/bin/env python3
"""
GCP Telegram Daemon 완전 자동 재시작
HTTP POST /restart 엔드포인트로 원격 재시작
"""
import requests
import json
from datetime import datetime

GCP_URL = "http://35.184.30.182:8888"

def restart_telegram_daemon():
    """GCP에 재시작 요청"""
    print(f"[{datetime.now()}] 🔄 GCP Telegram Daemon 재시작 요청 중...")

    try:
        response = requests.post(f"{GCP_URL}/restart", timeout=15)

        if response.status_code == 200:
            result = response.json()
            print(f"[{datetime.now()}] ✅ {result['message']}")
            print(f"   상태: {result['status']}")
            return True
        else:
            print(f"[{datetime.now()}] ❌ 재시작 실패: {response.status_code}")
            print(f"   응답: {response.text}")
            return False

    except requests.exceptions.ConnectionError:
        print(f"[{datetime.now()}] ⚠️ GCP 연결 실패")
        print("   GCP에서 management server가 실행 중인지 확인하세요:")
        print("   nohup python3 execution/ops/gcp_management_server.py > /tmp/gcp_mgmt.log 2>&1 &")
        return False

    except Exception as e:
        print(f"[{datetime.now()}] ❌ 오류: {e}")
        return False

def check_gcp_status():
    """GCP 시스템 상태 확인"""
    try:
        response = requests.get(f"{GCP_URL}/status", timeout=10)
        if response.status_code == 200:
            status = response.json()
            print(f"\n📊 GCP 상태:")
            print(f"   Telegram Daemon: {status['telegram_daemon']}")
            print(f"   호스트: {status['hostname']}")
            print(f"   시간: {status['timestamp']}")
            return True
        return False
    except:
        return False

if __name__ == "__main__":
    print("=" * 60)
    print("GCP Telegram Daemon 자동 재시작")
    print("=" * 60)

    # 1. 현재 상태 확인
    if not check_gcp_status():
        print("\n⚠️ GCP management server에 연결할 수 없습니다.")
        print("\nGCP SSH에서 다음 명령어를 실행하세요:")
        print("-" * 60)
        print("cd ~/97layerOS && \\")
        print("nohup python3 execution/ops/gcp_management_server.py > /tmp/gcp_mgmt.log 2>&1 &")
        print("-" * 60)
        exit(1)

    # 2. 재시작 요청
    print()
    if restart_telegram_daemon():
        print("\n🎉 완료! Telegram Bot이 자연스러운 대화 모드로 업데이트되었습니다.")
        print("   텔레그램에서 테스트해보세요: '안녕' 또는 '현재 상태 알려줘'")
    else:
        print("\n❌ 재시작 실패. 수동으로 GCP SSH에서 실행하세요:")
        print("   cd ~/97layerOS && pkill -f telegram_daemon.py && \\")
        print("   nohup python3 execution/telegram_daemon.py > /tmp/telegram_daemon.log 2>&1 &")
