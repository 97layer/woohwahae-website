#!/bin/bash
# 97layerOS Cortex Global Deployment (Native / Systemd)
# Docker 빌드 실패 시 사용하는 경량화 배포 스크립트
set -e

VM_IP="136.109.201.201"
VM_USER="skyto5339_gmail_com"
VM_KEY="${HOME}/.ssh/google_compute_engine"
VM_HOST="${VM_USER}@${VM_IP}"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STAGING_DIR="/tmp/cortex-staging-native"

echo "🚀 Cortex Native Deployment 시작..."

# 1. Staging
echo "[1/4] Staging 영역 생성..."
rm -rf "${STAGING_DIR}"
mkdir -p "${STAGING_DIR}"

# 핵심 파일 복사
rsync -av --exclude='.DS_Store' --exclude='__pycache__' --exclude='*.pyc' --exclude='.git' \
    --exclude='website/assets/uploads' \
    --exclude='website/assets/img' \
    --exclude='website/assets/css/_archive' \
    core directives knowledge website requirements.txt .env \
    "${STAGING_DIR}/" || true

# 2. 패키징
echo "[2/4] 소스 코드 패키징..."
cd "${STAGING_DIR}"
find . -name "._*" -delete
COPYFILE_DISABLE=1 tar -czf /tmp/cortex-native.tar.gz .
cd "${PROJECT_ROOT}"
rm -rf "${STAGING_DIR}"

# 3. 이관
echo "[3/4] VM으로 전송..."
ssh -i ${VM_KEY} ${VM_HOST} "mkdir -p ~/97layerOS"
scp -i ${VM_KEY} /tmp/cortex-native.tar.gz ${VM_HOST}:~/
if [ -f .env ]; then
    scp -i ${VM_KEY} .env ${VM_HOST}:~/97layerOS/.env
fi
rm /tmp/cortex-native.tar.gz

# 4. 배포 실행 (Native Systemd)
echo "[4/4] Native 환경 설정 및 기동..."
ssh -i ${VM_KEY} ${VM_HOST} bash << 'ENDSSH'
set -e
mkdir -p ~/97layerOS
cd ~/97layerOS

# 압축 해제
tar -xzf ~/cortex-native.tar.gz
rm ~/cortex-native.tar.gz

# Docker 정리 (충돌 방지 및 리소스 확보)
echo "  - Stopping heavy processes..."
sudo systemctl stop docker 2>/dev/null || true
pkill -f "dockerd" || true
pkill -f "docker-proxy" || true

# Python venv 설정
echo "  - Setting up Python environment..."
if [ ! -d ".venv" ]; then
    python3 -m venv .venv
fi
.venv/bin/pip install -U pip setuptools wheel
.venv/bin/pip install -r requirements.txt

# Systemd 서비스 등록 (Cortex Ecosystem)
echo "  - Registering Systemd services..."

# 1. Cortex Admin (Web)
cat <<EOF | sudo tee /etc/systemd/system/cortex-admin.service
[Unit]
Description=97layerOS Cortex Admin
After=network.target

[Service]
User=$(whoami)
WorkingDirectory=/home/$(whoami)/97layerOS
ExecStart=/home/$(whoami)/97layerOS/.venv/bin/python core/admin/app.py
Restart=always
EnvironmentFile=/home/$(whoami)/97layerOS/.env

[Install]
WantedBy=multi-user.target
EOF

# 2. Cortex Dashboard
cat <<EOF | sudo tee /etc/systemd/system/cortex-dashboard.service
[Unit]
Description=97layerOS Cortex Dashboard
After=network.target

[Service]
User=$(whoami)
WorkingDirectory=/home/$(whoami)/97layerOS
ExecStart=/home/$(whoami)/97layerOS/.venv/bin/python core/daemons/dashboard_server.py
Restart=always
EnvironmentFile=/home/$(whoami)/97layerOS/.env

[Install]
WantedBy=multi-user.target
EOF

# 3. Cortex Engine (Signal Processor + Telegram)
cat <<EOF | sudo tee /etc/systemd/system/cortex-engine.service
[Unit]
Description=97layerOS Cortex Engine
After=network.target

[Service]
User=$(whoami)
WorkingDirectory=/home/$(whoami)/97layerOS
ExecStart=/bin/bash -c "/home/$(whoami)/97layerOS/.venv/bin/python core/system/signal_processor.py & /home/$(whoami)/97layerOS/.venv/bin/python core/daemons/telegram_secretary.py & wait"
Restart=always
EnvironmentFile=/home/$(whoami)/97layerOS/.env

[Install]
WantedBy=multi-user.target
EOF

# 서비스 리로드 및 재시작
sudo systemctl daemon-reload
sudo systemctl enable cortex-admin cortex-dashboard cortex-engine
sudo systemctl restart cortex-admin cortex-dashboard cortex-engine

# Cloudflare Tunnel 실행 (Native)
echo "  - Starting Cloudflare Tunnel..."
pkill -f "cloudflared tunnel" || true
# 로그 파일로 출력, 백그라운드 실행
nohup cloudflared tunnel --url http://localhost:5001 > ~/cloudflared.log 2>&1 &

echo "🎉 Native 배포 완료!"
sleep 5
sudo systemctl status cortex-admin --no-pager
echo ""
echo "🌍 Cloudflare Tunnel URL:"
grep -o 'https://.*\.trycloudflare.com' ~/cloudflared.log | head -1
ENDSSH
