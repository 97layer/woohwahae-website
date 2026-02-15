#!/usr/bin/env python3
"""
Mac에서 5분마다 실행: GCP HTTP 서버에서 chat_memory 가져오기
"""
import json
import requests
from pathlib import Path
from datetime import datetime

GCP_URL = "http://35.184.30.182:8888/memory"
MEMORY_FILE = Path.home() / "97layerOS" / "knowledge" / "chat_memory" / "7565534667.json"

def pull_via_http():
    """HTTP로 GCP에서 chat_memory 가져오기"""
    try:
        response = requests.get(GCP_URL, timeout=10)

        if response.status_code != 200:
            print(f"[{datetime.now()}] ⚠️ HTTP {response.status_code}")
            return False

        gcp_data = response.json()

        # 로컬과 비교
        local_data = []
        if MEMORY_FILE.exists():
            with open(MEMORY_FILE) as f:
                local_data = json.load(f)

        if len(gcp_data) != len(local_data):
            # 백업
            if MEMORY_FILE.exists():
                backup = MEMORY_FILE.with_suffix('.json.backup')
                MEMORY_FILE.rename(backup)

            # 저장
            with open(MEMORY_FILE, 'w', encoding='utf-8') as f:
                json.dump(gcp_data, f, ensure_ascii=False, indent=4)

            print(f"[{datetime.now()}] ✅ 업데이트: {len(gcp_data)}개 (이전: {len(local_data)})")

            # Google Drive 동기화
            sync_script = MEMORY_FILE.parent.parent.parent / "execution" / "ops" / "sync_to_gdrive.py"
            if sync_script.exists():
                import subprocess
                subprocess.run(["/usr/bin/python3", str(sync_script)], check=False)

            return True
        else:
            print(f"[{datetime.now()}] ℹ️ 변경 없음: {len(gcp_data)}개")
            return True

    except requests.exceptions.RequestException as e:
        print(f"[{datetime.now()}] ⚠️ 연결 실패: {e}")
        return False
    except Exception as e:
        print(f"[{datetime.now()}] ❌ 오류: {e}")
        return False

if __name__ == "__main__":
    pull_via_http()
