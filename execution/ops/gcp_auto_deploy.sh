#!/bin/bash
# GCP 완전 자동 배포 스크립트
# 맥북에서 이 스크립트 1번만 실행 → GCP 완전 자동 시작
#
# Usage:
#   ./gcp_auto_deploy.sh
#
# Author: 97LAYER
# Date: 2026-02-14

set -e  # 에러 발생 시 중단

# SSH Key 설정
SSH_KEY="$HOME/.ssh/id_ed25519_gcp"
SSH_OPTS="-i $SSH_KEY -o StrictHostKeyChecking=no"

echo "🚀 97layerOS GCP Auto-Deploy Starting..."
echo ""

# ====================
# 1. 환경 변수 체크
# ====================
echo "📋 Checking environment..."

if [ ! -f ".env" ]; then
    echo "❌ .env file not found!"
    exit 1
fi

# GCP 인스턴스 정보 (기존 설정 자동 감지)
GCP_IP="35.184.30.182"
GCP_USER="skyto5339"

echo "🔍 Auto-detected GCP instance from existing config"

echo "✅ GCP Instance: $GCP_USER@$GCP_IP"
echo ""

# ====================
# 2. 파일 업로드
# ====================
echo "📤 Uploading files to GCP..."

# 필수 디렉토리 생성
ssh $SSH_OPTS $GCP_USER@$GCP_IP "mkdir -p ~/97layerOS"

# 파일 업로드 (rsync 사용)
rsync -avz -e "ssh $SSH_OPTS" --exclude '.git' --exclude '__pycache__' --exclude '*.pyc' \
    --exclude '.tmp' --exclude 'logs' \
    ./ $GCP_USER@$GCP_IP:~/97layerOS/

echo "✅ Files uploaded"
echo ""

# ====================
# 3. systemd 서비스 생성
# ====================
echo "⚙️ Creating systemd services..."

# 3-1. Master Controller 서비스
cat << 'EOF' | ssh $SSH_OPTS $GCP_USER@$GCP_IP "sudo tee /etc/systemd/system/97layer-master.service"
[Unit]
Description=97layerOS Master Controller
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=/home/$USER/97layerOS
ExecStart=/usr/bin/python3 /home/$USER/97layerOS/execution/ops/master_controller.py start_all
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# 3-2. Cycle Manager 서비스
cat << 'EOF' | ssh $SSH_OPTS $GCP_USER@$GCP_IP "sudo tee /etc/systemd/system/97layer-cycle.service"
[Unit]
Description=97layerOS Cycle Manager
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=/home/$USER/97layerOS
ExecStart=/usr/bin/python3 /home/$USER/97layerOS/execution/cycle_manager.py
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

echo "✅ systemd services created"
echo ""

# ====================
# 4. Python 의존성 설치
# ====================
echo "📦 Installing Python dependencies..."

ssh $SSH_OPTS $GCP_USER@$GCP_IP << 'ENDSSH'
cd ~/97layerOS

# pip 업그레이드
python3 -m pip install --upgrade pip

# 의존성 설치
if [ -f "requirements.txt" ]; then
    pip3 install -r requirements.txt
else
    # 최소 의존성
    pip3 install asyncio aiohttp python-telegram-bot google-generativeai anthropic schedule psutil python-dotenv
fi

echo "✅ Dependencies installed"
ENDSSH

echo ""

# ====================
# 5. systemd 서비스 활성화
# ====================
echo "🔧 Enabling systemd services..."

ssh $SSH_OPTS $GCP_USER@$GCP_IP << 'ENDSSH'
# systemd 리로드
sudo systemctl daemon-reload

# 서비스 활성화
sudo systemctl enable 97layer-master.service
sudo systemctl enable 97layer-cycle.service

# 서비스 시작
sudo systemctl start 97layer-master.service
sudo systemctl start 97layer-cycle.service

echo "✅ Services enabled and started"
ENDSSH

echo ""

# ====================
# 6. 상태 확인
# ====================
echo "🔍 Checking status..."
sleep 5

ssh $SSH_OPTS $GCP_USER@$GCP_IP << 'ENDSSH'
echo ""
echo "=== Master Controller Status ==="
sudo systemctl status 97layer-master.service --no-pager | head -15

echo ""
echo "=== Cycle Manager Status ==="
sudo systemctl status 97layer-cycle.service --no-pager | head -15

echo ""
echo "=== Running Processes ==="
ps aux | grep -E "(telegram|junction|cycle)" | grep -v grep
ENDSSH

echo ""
echo "✅ GCP Auto-Deploy Complete!"
echo ""
echo "📝 Next Steps:"
echo "  1. GCP 인스턴스 재부팅 시에도 자동 시작됩니다"
echo "  2. 텔레그램으로 메시지 전송 → 자동 처리"
echo "  3. 로그 확인: ssh $GCP_USER@$GCP_IP 'sudo journalctl -u 97layer-master -f'"
echo ""
echo "🎉 시스템 완전 자율 실행 중!"
