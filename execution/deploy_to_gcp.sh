#!/bin/bash
# GCP 서버에 최신 97layerOS 배포 스크립트
# Usage: ./execution/deploy_to_gcp.sh

set -e

GCP_HOST="skyto5339@35.184.30.182"
SSH_KEY="$HOME/.ssh/id_ed25519_gcp"
LOCAL_DIR="$HOME/97layerOS"
REMOTE_DIR="~/97layerOS"

echo "🚀 GCP 서버에 97layerOS 배포 시작..."

# 1. 최신 코드 전송 (rsync)
echo "📦 Step 1: 코드 전송 중..."
rsync -avz --delete \
  --exclude='.venv' \
  --exclude='node_modules' \
  --exclude='.git' \
  --exclude='*.pyc' \
  --exclude='__pycache__' \
  --exclude='.DS_Store' \
  --exclude='*.log' \
  --exclude='.local_node' \
  --exclude='.mcp-source' \
  -e "ssh -i $SSH_KEY" \
  "$LOCAL_DIR/" "$GCP_HOST:$REMOTE_DIR/"

echo "✅ 코드 전송 완료"

# 2. .env 파일 생성
echo "🔐 Step 2: .env 파일 생성 중..."
ssh -i $SSH_KEY $GCP_HOST << 'EOF'
cd ~/97layerOS
if [ -f ".env.txt" ]; then
  cp .env.txt .env
elif [ ! -f ".env" ]; then
  echo "❌ .env 파일이 없습니다. 수동으로 생성해주세요."
  exit 1
fi
echo "✅ .env 파일 확인 완료"
EOF

# 3. Python 패키지 설치
echo "📚 Step 3: Python 패키지 설치 중..."
ssh -i $SSH_KEY $GCP_HOST << 'EOF'
cd ~/97layerOS
python3 -m venv .venv
source .venv/bin/activate
pip install -q google-generativeai python-dotenv requests
echo "✅ 패키지 설치 완료"
EOF

# 4. 기존 Daemon 종료
echo "🛑 Step 4: 기존 Daemon 종료 중..."
ssh -i $SSH_KEY $GCP_HOST << 'EOF'
pkill -f "technical_daemon.py" || true
pkill -f "telegram_daemon.py" || true
sleep 2
echo "✅ 기존 Daemon 종료 완료"
EOF

# 5. Daemon 재시작
echo "🔄 Step 5: Daemon 재시작 중..."
ssh -i $SSH_KEY $GCP_HOST << 'EOF'
cd ~/97layerOS
source .venv/bin/activate

# Technical Daemon
nohup python execution/technical_daemon.py > /tmp/technical_daemon.log 2>&1 &
echo "✅ Technical Daemon 시작 (PID: $!)"

# Telegram Daemon
nohup python execution/telegram_daemon.py > /tmp/telegram_daemon.log 2>&1 &
echo "✅ Telegram Daemon 시작 (PID: $!)"

sleep 3
ps aux | grep -E "technical_daemon|telegram_daemon" | grep -v grep
EOF

echo ""
echo "🎉 배포 완료!"
echo ""
echo "다음 명령어로 로그 확인:"
echo "  ssh -i $SSH_KEY $GCP_HOST 'tail -f /tmp/technical_daemon.log'"
echo "  ssh -i $SSH_KEY $GCP_HOST 'tail -f /tmp/telegram_daemon.log'"
