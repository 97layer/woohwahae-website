#!/bin/bash
# 97layerOS GCP VM 배포 스크립트
# Usage: ./deploy.sh
#
# 로컬 코드 → GCP VM 97layerOS 업로드 → 서비스 재시작

set -e

VM_IP="136.109.201.201"
VM_USER="skyto5339_gmail_com"
VM_KEY="${HOME}/.ssh/google_compute_engine"
VM_HOST="${VM_USER}@${VM_IP}"
VM_PATH="/home/${VM_USER}/97layerOS"
SSH="ssh -i ${VM_KEY} -o ConnectTimeout=15 -o StrictHostKeyChecking=no"
SCP="scp -i ${VM_KEY} -o StrictHostKeyChecking=no"

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🚀 97layerOS 배포 시작 → ${VM_IP}"
echo ""

# [1] SSH 연결 확인
echo "[1/5] SSH 연결 확인..."
${SSH} ${VM_HOST} "echo 'OK'" > /dev/null 2>&1 && echo "✅ SSH OK" || { echo "❌ SSH 실패"; exit 1; }

# [2] 패키지 생성
echo "[2/5] 배포 패키지 생성..."
cd "${PROJECT_ROOT}"
tar \
    --exclude='*.pyc' \
    --exclude='*/__pycache__' \
    --exclude='.git' \
    --exclude='.env' \
    --no-xattrs \
    -czf /tmp/97layer-deploy.tar.gz \
    core/ \
    directives/ \
    knowledge/docs/ \
    knowledge/agent_hub/ \
    knowledge/long_term_memory.json \
    requirements.txt \
    website/
SIZE=$(du -h /tmp/97layer-deploy.tar.gz | cut -f1)
echo "✅ 패키지 생성 완료 (${SIZE})"

# [3] VM 업로드
echo "[3/5] VM 업로드..."
${SCP} /tmp/97layer-deploy.tar.gz ${VM_HOST}:~/
rm /tmp/97layer-deploy.tar.gz
echo "✅ 업로드 완료"

# [4] VM에서 추출 및 의존성 설치
echo "[4/5] VM 코드 반영..."
${SSH} ${VM_HOST} bash << 'ENDSSH'
set -e
cd ~/97layerOS
tar -xzf ~/97layer-deploy.tar.gz 2>/dev/null
rm ~/97layer-deploy.tar.gz
mkdir -p .infra/logs .infra/queue/tasks/{pending,processing,completed,failed} .infra/cache .infra/tmp
.venv/bin/pip install -q -r requirements.txt 2>&1 | tail -3
echo "✅ 코드 반영 완료"
ENDSSH

# [5] 서비스 재시작
echo "[5/5] 서비스 재시작..."
${SSH} ${VM_HOST} bash << 'ENDSSH'
sudo systemctl restart 97layer-telegram
sleep 4
STATUS=$(systemctl is-active 97layer-telegram)
if [ "$STATUS" = "active" ]; then
    echo "✅ 97layer-telegram: active"
    sudo systemctl status 97layer-telegram --no-pager | grep -E 'Active|Memory|PID'
else
    echo "❌ 서비스 상태: $STATUS"
    sudo journalctl -u 97layer-telegram -n 20 --no-pager
    exit 1
fi
ENDSSH

echo ""
echo "🎉 배포 완료 | GCP VM ${VM_IP}"
echo "   로그: ssh -i ~/.ssh/google_compute_engine ${VM_HOST} 'tail -f ~/97layerOS/.infra/logs/telegram.log'"
