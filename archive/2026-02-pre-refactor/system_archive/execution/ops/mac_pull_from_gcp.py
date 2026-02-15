#!/usr/bin/env python3
"""
Mac에서 5분마다 실행: GCP에서 chat_memory를 가져옴
SSH 대신 GCP에서 HTTP 서버를 띄우고 Mac이 pull
"""
import json
import subprocess
from pathlib import Path
from datetime import datetime

GCP_HOST = "35.184.30.182"
GCP_USER = "skyto5339"
SSH_KEY = Path.home() / ".ssh" / "id_ed25519_gcp"
MEMORY_FILE = Path.home() / "97layerOS" / "knowledge" / "chat_memory" / "7565534667.json"

def pull_from_gcp():
    """GCP에서 chat_memory를 가져옴"""
    try:
        # SSH로 GCP의 chat_memory 내용 가져오기
        cmd = [
            "ssh",
            "-i", str(SSH_KEY),
            "-o", "ConnectTimeout=10",
            "-o", "StrictHostKeyChecking=no",
            f"{GCP_USER}@{GCP_HOST}",
            f"cat ~/97layerOS/knowledge/chat_memory/7565534667.json"
        ]

        result = subprocess.run(cmd, capture_output=True, text=True, timeout=15)

        if result.returncode != 0:
            print(f"[{datetime.now()}] ⚠️ SSH 실패: {result.stderr}")
            return False

        # JSON 파싱
        gcp_data = json.loads(result.stdout)

        # 로컬과 비교
        local_data = []
        if MEMORY_FILE.exists():
            with open(MEMORY_FILE) as f:
                local_data = json.load(f)

        if len(gcp_data) > len(local_data):
            # 백업
            if MEMORY_FILE.exists():
                backup = MEMORY_FILE.with_suffix('.json.backup')
                MEMORY_FILE.rename(backup)

            # 저장
            with open(MEMORY_FILE, 'w', encoding='utf-8') as f:
                json.dump(gcp_data, f, ensure_ascii=False, indent=4)

            print(f"[{datetime.now()}] ✅ GCP에서 업데이트: {len(gcp_data)}개 메시지 (로컬: {len(local_data)})")
            return True
        else:
            print(f"[{datetime.now()}] ℹ️ 변경사항 없음: {len(gcp_data)}개")
            return True

    except subprocess.TimeoutExpired:
        print(f"[{datetime.now()}] ⚠️ SSH 타임아웃")
        return False
    except json.JSONDecodeError as e:
        print(f"[{datetime.now()}] ❌ JSON 파싱 실패: {e}")
        return False
    except Exception as e:
        print(f"[{datetime.now()}] ❌ 오류: {e}")
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    pull_from_gcp()
