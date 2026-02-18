#!/bin/bash
# VM 인프라 초기화 스크립트 (Cleanup & Robust Install)
set -e

echo "🧹 기존 Docker 관련 설정 클린업..."
sudo rm -f /etc/apt/sources.list.d/docker.list || true
sudo rm -f /etc/apt/keyrings/docker.gpg || true

echo "📦 VM 패키지 업데이트..."
sudo apt-get update || true # Ignore errors initially

# 1. Docker 설치 (공식 가이드 준수)
echo "🐳 Docker 엔진 설치 (수동 키 등록)..."
sudo apt-get install -y ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# 저장소 추가
echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# 2. 권한 설정
sudo usermod -aG docker $(whoami)

# 3. Cloudflared 설치
echo "☁️ Cloudflared 설치 중..."
curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared.deb
rm cloudflared.deb

echo "✅ 인프라 설치 완료."
docker --version
cloudflared --version
