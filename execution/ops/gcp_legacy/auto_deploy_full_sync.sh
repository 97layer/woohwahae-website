#!/bin/bash
# 완전 자동화된 GCP 동기화 시스템 배포
set -e

echo "🚀 97layerOS 양방향 동기화 완전 자동 배포"
echo "========================================"

# 1. 전체 시스템 패키지 생성
echo "📦 1. 배포 패키지 생성 중..."
cd /Users/97layer
tar czf /tmp/97layerOS_full_deploy.tar.gz \
    --exclude='97layerOS/.venv*' \
    --exclude='97layerOS/.DS_Store' \
    --exclude='97layerOS/__pycache__' \
    --exclude='97layerOS/*.pyc' \
    --exclude='97layerOS/.git' \
    --exclude='97layerOS/node_modules' \
    --exclude='97layerOS/.tmp' \
    97layerOS/

PACKAGE_SIZE=$(ls -lh /tmp/97layerOS_full_deploy.tar.gz | awk '{print $5}')
echo "   ✅ 패키지 생성 완료: $PACKAGE_SIZE"

# 2. GCP 배포 명령어 생성
echo ""
echo "📝 2. GCP 배포 스크립트 생성 중..."

cat > /tmp/deploy_on_gcp.sh << 'EOFGCP'
#!/bin/bash
# GCP에서 실행될 배포 스크립트
set -e

echo "🔄 97layerOS 배포 시작..."

# 기존 프로세스 중지
echo "1️⃣ 기존 데몬 중지..."
pkill -f "technical_daemon.py" || true
pkill -f "telegram_daemon.py" || true
sleep 2

# 백업
echo "2️⃣ 기존 설정 백업..."
cd ~
if [ -d "97layerOS" ]; then
    cp 97layerOS/.env /tmp/backup_env 2>/dev/null || true
    cp 97layerOS/credentials.json /tmp/backup_creds.json 2>/dev/null || true
    cp 97layerOS/token.json /tmp/backup_token.json 2>/dev/null || true
fi

# 압축 해제
echo "3️⃣ 새 버전 배포..."
rm -rf 97layerOS_old
mv 97layerOS 97layerOS_old 2>/dev/null || true
tar xzf /tmp/97layerOS_full_deploy.tar.gz

# 설정 복원
echo "4️⃣ 설정 복원..."
cd 97layerOS
cp /tmp/backup_env .env 2>/dev/null || true
cp /tmp/backup_creds.json credentials.json 2>/dev/null || true
cp /tmp/backup_token.json token.json 2>/dev/null || true

# Python 환경 설정
echo "5️⃣ Python 환경 설정..."
python3 -m venv .venv
source .venv/bin/activate
pip install -q google-generativeai python-dotenv requests

# 동기화 스크립트 실행 권한
chmod +x execution/ops/*.py 2>/dev/null || true
chmod +x execution/ops/*.sh 2>/dev/null || true

# 데몬 재시작
echo "6️⃣ 데몬 시작..."
nohup python execution/technical_daemon.py > /tmp/technical_daemon.log 2>&1 &
TECH_PID=$!
echo "   Technical Daemon: $TECH_PID"

nohup python execution/telegram_daemon.py > /tmp/telegram_daemon.log 2>&1 &
TELE_PID=$!
echo "   Telegram Daemon: $TELE_PID"

sleep 3

# 확인
echo "7️⃣ 프로세스 확인..."
ps aux | grep -E "technical_daemon|telegram_daemon" | grep -v grep

echo ""
echo "✅ 배포 완료!"
echo ""
echo "📊 로그 확인:"
echo "   tail -f /tmp/technical_daemon.log"
echo "   tail -f /tmp/telegram_daemon.log"

EOFGCP

chmod +x /tmp/deploy_on_gcp.sh

# 3. GCP 브라우저 SSH용 한 줄 명령어 생성
cat > /tmp/gcp_oneliner.sh << 'EOFONELINE'
cd ~ && curl -o /tmp/deploy.sh https://pastebin.com/raw/PLACEHOLDER && bash /tmp/deploy.sh
EOFONELINE

echo "   ✅ GCP 스크립트 생성 완료"

# 4. 사용자 안내
echo ""
echo "=========================================="
echo "📋 GCP 배포 방법 (선택)"
echo "=========================================="
echo ""
echo "방법 1: 브라우저 SSH로 파일 업로드 (권장)"
echo "----------------------------------------"
echo "1. GCP Console → Compute Engine → SSH"
echo "2. 톱니바퀴 → Upload file"
echo "3. 업로드: /tmp/97layerOS_full_deploy.tar.gz ($PACKAGE_SIZE)"
echo "4. 업로드: /tmp/deploy_on_gcp.sh"
echo "5. 실행: bash /tmp/deploy_on_gcp.sh"
echo ""
echo "방법 2: wget으로 직접 다운로드 (실험적)"
echo "----------------------------------------"
echo "Mac에서 간이 웹서버 실행:"
echo "  cd /tmp && python3 -m http.server 8000"
echo ""
echo "GCP에서 다운로드:"
echo "  wget http://[MAC_IP]:8000/97layerOS_full_deploy.tar.gz -O /tmp/97layerOS_full_deploy.tar.gz"
echo "  wget http://[MAC_IP]:8000/deploy_on_gcp.sh -O /tmp/deploy_on_gcp.sh"
echo "  bash /tmp/deploy_on_gcp.sh"
echo ""
echo "=========================================="
echo "✅ 준비 완료!"
echo "=========================================="
echo ""
echo "다음: GCP Console 브라우저 SSH로 접속하여 파일 업로드"
