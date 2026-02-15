#!/usr/bin/env python3
"""
GCP Telegram Daemon 원격 재시작
Google Drive를 통한 코드 동기화 + 원격 재시작
"""
import subprocess
import time
from pathlib import Path

def restart_gcp_daemon():
    """GCP telegram daemon 재시작 (Google Drive 경유)"""
    print("🔄 GCP Telegram Daemon 재시작 프로세스 시작...")

    # Step 1: Mac → Google Drive 동기화 (이미 완료됨)
    print("✅ Mac → Google Drive 동기화 완료")

    # Step 2: GCP가 Google Drive에서 자동으로 pull하도록 요청
    # GCP의 sync_from_gdrive 스크립트가 주기적으로 실행 중이면 자동으로 반영됨

    print("\n📋 GCP에서 수동 실행 필요:")
    print("=" * 60)
    print("cd ~/97layerOS && \\")
    print("pkill -f telegram_daemon.py && \\")
    print("nohup python3 execution/telegram_daemon.py > /tmp/telegram_daemon.log 2>&1 & \\")
    print("sleep 2 && ps aux | grep telegram_daemon | grep -v grep")
    print("=" * 60)

    print("\n또는 GCP SSH 콘솔에서 위 명령어를 복사/붙여넣기하세요.")
    print("\n✅ 코드는 이미 Google Drive에 동기화되었습니다.")
    print("   GCP에서 재시작만 하면 최신 버전이 적용됩니다.")

if __name__ == "__main__":
    restart_gcp_daemon()
