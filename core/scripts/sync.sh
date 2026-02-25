#!/bin/bash
# 97layerOS GCP VM → 로컬 knowledge 동기화
# Usage: ./sync.sh
#
# GCP VM에 쌓인 signals, long_term_memory를 로컬로 pull

VM_IP="136.109.201.201"
VM_USER="skyto5339_gmail_com"
VM_KEY="${HOME}/.ssh/google_compute_engine"
VM_HOST="${VM_USER}@${VM_IP}"
VM_PATH="/home/${VM_USER}/97layerOS/knowledge"
LOCAL_PATH="/Users/97layer/97layerOS/knowledge"

echo "🔄 GCP VM → 로컬 knowledge 동기화..."
echo "   출처: ${VM_IP}:${VM_PATH}"
echo "   대상: ${LOCAL_PATH}"
echo ""

# signals 전체 동기화 (GCP → 로컬, 삭제 없이)
rsync -avz --progress \
    -e "ssh -i ${VM_KEY} -o StrictHostKeyChecking=no" \
    ${VM_HOST}:${VM_PATH}/signals/ \
    ${LOCAL_PATH}/signals/ \
    2>/dev/null

# long_term_memory.json 동기화 (더 최신 것으로)
rsync -avz \
    -e "ssh -i ${VM_KEY} -o StrictHostKeyChecking=no" \
    ${VM_HOST}:${VM_PATH}/long_term_memory.json \
    ${LOCAL_PATH}/long_term_memory.json \
    2>/dev/null

# corpus 동기화 (Gardener 로컬 실행을 위해 필수)
rsync -avz --delete \
    -e "ssh -i ${VM_KEY} -o StrictHostKeyChecking=no" \
    ${VM_HOST}:${VM_PATH}/corpus/ \
    ${LOCAL_PATH}/corpus/ \
    2>/dev/null

echo ""
echo "✅ 동기화 완료"
echo ""
echo "=== 로컬 signals 현황 ==="
echo "  텍스트: $(ls ${LOCAL_PATH}/signals/*.json 2>/dev/null | wc -l | tr -d ' ')개"
echo "  이미지: $(ls ${LOCAL_PATH}/signals/images/*.json 2>/dev/null | wc -l | tr -d ' ')개"
echo "  유튜브: $(ls ${LOCAL_PATH}/signals/youtube*.json 2>/dev/null | wc -l | tr -d ' ')개"
total=$(find ${LOCAL_PATH}/signals -name "*.json" | wc -l | tr -d ' ')
echo "  전체: ${total}개"
