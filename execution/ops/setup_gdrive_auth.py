#!/usr/bin/env python3
"""
Filename: setup_gdrive_auth.py
Purpose: Google Drive API 인증 자동화
Usage: python3 setup_gdrive_auth.py
"""

import os
import json
from pathlib import Path

# Google Drive API는 이미 있는 credentials.json을 사용
CREDENTIALS_FILE = Path.home() / "97layerOS" / "credentials.json"
TOKEN_FILE = Path.home() / "97layerOS" / "token.json"

def main():
    print("🔍 기존 Google Drive 인증 확인 중...")

    # credentials.json 확인
    if CREDENTIALS_FILE.exists():
        print(f"✅ credentials.json 발견: {CREDENTIALS_FILE}")
        creds_data = json.loads(CREDENTIALS_FILE.read_text())
        print(f"   Client ID: {creds_data.get('installed', {}).get('client_id', 'N/A')[:40]}...")
    else:
        print(f"❌ credentials.json 없음: {CREDENTIALS_FILE}")
        print("\n대안: Google Drive Desktop 앱 사용")
        print("   - Google Drive Desktop이 이미 설치되어 있으면")
        print("   - ~/Google Drive/내 드라이브/ 경로를 직접 사용")
        return False

    # token.json 확인
    if TOKEN_FILE.exists():
        print(f"✅ token.json 발견: {TOKEN_FILE}")
        print("   이미 인증된 상태입니다!")
        return True
    else:
        print(f"⚠️  token.json 없음: {TOKEN_FILE}")
        print("   OAuth 인증이 필요합니다.")

    # Google Drive Desktop 경로 확인
    gdrive_path = Path.home() / "Google Drive" / "내 드라이브"
    if gdrive_path.exists():
        print(f"\n✅ Google Drive Desktop 감지: {gdrive_path}")
        print("   rclone 없이 직접 파일 시스템으로 동기화 가능!")
        return True
    else:
        print(f"\n❌ Google Drive Desktop 없음: {gdrive_path}")

    return False

if __name__ == "__main__":
    success = main()
    exit(0 if success else 1)
