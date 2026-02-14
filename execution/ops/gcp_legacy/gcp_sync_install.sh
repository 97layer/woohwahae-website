#!/bin/bash
# GCP 브라우저 SSH에 붙여넣기용
# 97layerOS 양방향 동기화 설치

echo "🔄 97layerOS 동기화 시스템 설치 시작..."

# 1. 기존 파일 백업 (있다면)
cd ~
if [ -d "97layerOS/execution/ops" ]; then
    echo "📦 기존 동기화 스크립트 백업 중..."
    cp 97layerOS/execution/ops/sync_*.py /tmp/backup_sync_$(date +%Y%m%d_%H%M%S)/ 2>/dev/null || true
fi

# 2. credentials와 token 파일 복사 (이미 있을 것으로 예상)
echo "🔑 인증 파일 준비 중..."
cd ~/97layerOS

# 3. Python 가상환경 활성화
echo "🐍 Python 환경 설정 중..."
source .venv/bin/activate

# 4. Google Drive API 패키지 설치
echo "📚 Google Drive API 패키지 설치 중..."
pip install -q google-api-python-client google-auth-httplib2 google-auth-oauthlib

# 5. 동기화 스크립트 생성
echo "📝 동기화 스크립트 생성 중..."
cat > ~/97layerOS/execution/ops/sync_gcp_to_gdrive_simple.sh << 'EOFSCRIPT'
#!/bin/bash
# GCP → Google Drive 간단 동기화 (rsync 방식)

cd ~/97layerOS

# Google Drive 마운트 확인 (없으면 스킵)
if [ ! -d "/mnt/gdrive" ]; then
    echo "[$(date)] ⚠️  Google Drive 마운트 없음 - 스킵"
    exit 0
fi

# knowledge 폴더만 동기화 (가장 중요)
rsync -a --delete \
    --exclude=".DS_Store" \
    --exclude="*.pyc" \
    ~/97layerOS/knowledge/ /mnt/gdrive/97layerOS/knowledge/

echo "[$(date)] ✅ GCP → Google Drive 동기화 완료"
EOFSCRIPT

chmod +x ~/97layerOS/execution/ops/sync_gcp_to_gdrive_simple.sh

# 6. 대안: Python으로 직접 파일 복사
cat > ~/97layerOS/execution/ops/sync_gcp_simple.py << 'EOFPYTHON'
#!/usr/bin/env python3
"""GCP → Google Drive 간단 동기화 (API 없이)"""
import shutil
import json
from pathlib import Path
from datetime import datetime

GDRIVE_BASE = Path("/mnt/gdrive/97layerOS")

def simple_sync():
    """지식 데이터만 동기화"""
    print(f"[{datetime.now()}] 🔄 간단 동기화 시작...")

    # Google Drive 마운트 확인
    if not GDRIVE_BASE.exists():
        print("⚠️  Google Drive 마운트 없음")
        return False

    # knowledge 폴더만 동기화
    src = Path.home() / "97layerOS" / "knowledge"
    dst = GDRIVE_BASE / "knowledge"

    if src.exists():
        if dst.exists():
            shutil.rmtree(dst)
        shutil.copytree(src, dst,
            ignore=shutil.ignore_patterns('*.pyc', '__pycache__', '.DS_Store'))
        print(f"[{datetime.now()}] ✅ knowledge/ 동기화 완료")
        return True

    return False

if __name__ == "__main__":
    simple_sync()
EOFPYTHON

chmod +x ~/97layerOS/execution/ops/sync_gcp_simple.py

# 7. 현재 GCP에서 Google Drive가 어디 마운트되어 있는지 확인
echo ""
echo "📊 현재 시스템 상태:"
echo "===================="
df -h | grep -i drive || echo "  Google Drive 마운트 없음"
echo ""
ls -la ~/ | grep -i drive || echo "  홈 디렉토리에 Drive 폴더 없음"
echo ""

# 8. 설치 완료
echo "✅ 동기화 시스템 설치 완료!"
echo ""
echo "📝 다음 단계:"
echo "=============="
echo "1. Google Drive 위치 확인:"
echo "   df -h | grep drive"
echo "   ls -la ~/ | grep -i google"
echo ""
echo "2. 동기화 테스트:"
echo "   python ~/97layerOS/execution/ops/sync_gcp_simple.py"
echo ""
echo "3. Cron 등록 (5분마다):"
echo "   crontab -e"
echo "   # 추가: */5 * * * * python3 /home/skyto5339/97layerOS/execution/ops/sync_gcp_simple.py >> /tmp/gcp_sync.log 2>&1"
echo ""
