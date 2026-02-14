#!/usr/bin/env python3
"""
GCP에서 실행될 간단한 Google Drive 동기화
Google Drive Desktop이 아닌 직접 파일 복사 방식
"""
import os
import shutil
import tarfile
from pathlib import Path
from datetime import datetime

BASE_DIR = Path.home() / "97layerOS"

def create_sync_package():
    """동기화용 tar 패키지 생성"""
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    tar_path = Path(f"/tmp/97layerOS_sync_{timestamp}.tar.gz")

    print(f"[{datetime.now()}] 📦 동기화 패키지 생성 중...")

    with tarfile.open(tar_path, "w:gz") as tar:
        # knowledge 폴더 (가장 중요)
        tar.add(BASE_DIR / "knowledge", arcname="knowledge")
        # task_status.json
        if (BASE_DIR / "task_status.json").exists():
            tar.add(BASE_DIR / "task_status.json", arcname="task_status.json")

    print(f"[{datetime.now()}] ✅ 패키지 생성 완료: {tar_path}")
    print(f"   크기: {tar_path.stat().st_size / 1024:.1f} KB")

    return tar_path

def main():
    """메인 실행"""
    try:
        # GCP에서 실행 중인지 확인
        hostname = os.uname().nodename
        print(f"[{datetime.now()}] 🖥️  호스트: {hostname}")

        # 동기화 패키지 생성
        tar_path = create_sync_package()

        print(f"""
[{datetime.now()}] 📋 다음 단계:

1. Mac에서 이 파일을 가져오기:
   scp -i ~/.ssh/id_ed25519_gcp skyto5339@35.184.30.182:{tar_path} /tmp/

2. Mac에서 압축 해제:
   cd ~/내\\ 드라이브\\(skyto5339@gmail.com\\)/97layerOS
   tar xzf /tmp/{tar_path.name}

또는 GCP 브라우저 SSH에서 다운로드:
   Download file: {tar_path}
""")

        return True

    except Exception as e:
        print(f"[{datetime.now()}] ❌ 오류: {e}")
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    main()
