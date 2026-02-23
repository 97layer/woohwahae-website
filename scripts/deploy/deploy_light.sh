#!/bin/bash
# 97layerOS Cortex Global Deployment (Ultra Light)
set -e

VM_IP="136.109.201.201"
VM_USER="skyto5339_gmail_com"
VM_KEY="${HOME}/.ssh/google_compute_engine"
VM_HOST="${VM_USER}@${VM_IP}"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STAGING_DIR="/tmp/cortex-staging-light"

echo "🚀 Cortex Light Deployment 시작..."

# 1. Staging (최소한의 소스만 포함)
echo "[1/4] Staging 영역 생성..."
rm -rf "${STAGING_DIR}"
mkdir -p "${STAGING_DIR}"

# 핵심 파일만 복사 (이미지 에셋 제외)
# .env 파일은 별도 복사
rsync -av --exclude='.DS_Store' --exclude='__pycache__' --exclude='*.pyc' --exclude='.git' \
    --exclude='website/assets/uploads' \
    --exclude='website/assets/img' \
    --exclude='website/assets/css/_archive' \
    core directives knowledge website requirements.txt Dockerfile docker-compose.yml \
    "${STAGING_DIR}/" || true

# 2. 패키징
echo "[2/4] 소스 코드 패키징..."
cd "${STAGING_DIR}"
find . -name "._*" -delete
COPYFILE_DISABLE=1 tar -czf /tmp/cortex-light.tar.gz .
cd "${PROJECT_ROOT}"
rm -rf "${STAGING_DIR}"

SIZE=$(du -h /tmp/cortex-light.tar.gz | cut -f1)
echo "✅ 패키지 크기: ${SIZE}"

# 3. 이관
echo "[3/4] VM으로 전송..."
ssh -i ${VM_KEY} ${VM_HOST} "mkdir -p ~/97layerOS"
scp -i ${VM_KEY} /tmp/cortex-light.tar.gz ${VM_HOST}:~/
if [ -f .env ]; then
    scp -i ${VM_KEY} .env ${VM_HOST}:~/97layerOS/.env
fi
rm /tmp/cortex-light.tar.gz

# 4. 배포 실행 (순차 빌드 및 메모리 제한)
echo "[4/4] 컨테이너 순차 빌드 및 기동..."
ssh -i ${VM_KEY} ${VM_HOST} bash << 'ENDSSH'
set -e
mkdir -p ~/97layerOS
cd ~/97layerOS

# 압축 해제
tar -xzf ~/cortex-light.tar.gz
rm ~/cortex-light.tar.gz

# 기존 정리 (메모리 확보)
echo "  - Cleaning up..."
docker system prune -f >/dev/null 2>&1 || true

# 순차 빌드 (OOM 방지)
echo "  - Building services sequentially..."
export DOCKER_BUILDKIT=1
docker compose build --no-cache cortex-admin
docker compose build --no-cache cortex-dashboard
docker compose build --no-cache cortex-engine

# 실행
echo "  - Starting services..."
docker compose up -d --no-build

# 상태 확인
sleep 5
docker compose ps
echo ""
echo "🎉 배포 완료!"
echo "터널링 로그: docker compose logs cortex-tunnel | grep trycloudflare"
ENDSSH
