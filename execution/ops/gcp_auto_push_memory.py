#!/usr/bin/env python3
"""
GCP에서 5분마다 실행되는 자동 push 스크립트
chat_memory를 Mac 서버로 HTTP POST 전송
"""
import json
import sys
from pathlib import Path
from datetime import datetime

try:
    import requests
except ImportError:
    print("❌ requests 모듈 없음. pip install requests 실행 필요")
    sys.exit(1)

# 설정
MAC_SERVER = "http://192.168.0.8:9876"  # Mac 로컬 IP
CHAT_MEMORY_FILE = Path.home() / "97layerOS" / "knowledge" / "chat_memory" / "7565534667.json"

def push_to_mac():
    """chat_memory를 Mac으로 전송"""
    try:
        # chat_memory 읽기
        if not CHAT_MEMORY_FILE.exists():
            print(f"[{datetime.now()}] ❌ chat_memory 파일 없음: {CHAT_MEMORY_FILE}")
            return False

        with open(CHAT_MEMORY_FILE, 'r', encoding='utf-8') as f:
            memory_data = json.load(f)

        print(f"[{datetime.now()}] 📤 Mac 서버로 전송 중... ({len(memory_data)}개 메시지)")

        # HTTP POST
        response = requests.post(
            f"{MAC_SERVER}/sync_memory",
            json=memory_data,
            timeout=10
        )

        if response.status_code == 200:
            result = response.json()
            print(f"[{datetime.now()}] ✅ 전송 성공: {result}")
            return True
        else:
            print(f"[{datetime.now()}] ❌ 전송 실패: {response.status_code}")
            return False

    except requests.exceptions.ConnectionError:
        print(f"[{datetime.now()}] ⚠️  Mac 서버 연결 실패 (꺼져있거나 네트워크 문제)")
        return False
    except Exception as e:
        print(f"[{datetime.now()}] ❌ 오류: {e}")
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    push_to_mac()
