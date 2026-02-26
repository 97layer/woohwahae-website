#!/bin/bash
# LAYER OS 배포 스크립트
# 웹: git push → Cloudflare Pages 자동 배포
# 코드: VM git pull + 서비스 재시작
#
# Usage:
#   ./deploy.sh              코드 pull만 (재시작 없음)
#   ./deploy.sh all          코드 pull + 전 서비스 재시작
#   ./deploy.sh web          Cloudflare Pages 안내
#   ./deploy.sh <서비스명>   특정 서비스 재시작
#   ./deploy.sh --status     서비스 상태 확인

set -e

VM_HOST="97layer-vm"
VM_PATH="/home/skyto5339_gmail_com/97layerOS"

SERVICES_ALL="97layer-telegram 97layer-ecosystem 97layer-gardener woohwahae-backend cortex-admin"

case "${1:-pull}" in
  web)
    echo "웹은 git push만 하면 Cloudflare Pages가 자동 배포합니다."
    echo "  git push origin main → 30초 내 woohwahae.kr 반영"
    ;;

  --status)
    echo "=== VM 서비스 상태 ==="
    ssh ${VM_HOST} "for s in ${SERVICES_ALL}; do printf '%-25s %s\n' \$s \$(systemctl is-active \$s 2>/dev/null || echo 'not-found'); done"
    ;;

  all)
    echo "[1/2] git pull..."
    ssh ${VM_HOST} "cd ${VM_PATH} && git pull origin main"
    echo "[2/2] 전 서비스 재시작..."
    ssh ${VM_HOST} "sudo systemctl restart ${SERVICES_ALL}"
    sleep 3
    ssh ${VM_HOST} "for s in ${SERVICES_ALL}; do printf '%-25s %s\n' \$s \$(systemctl is-active \$s); done"
    ;;

  pull)
    echo "git pull..."
    ssh ${VM_HOST} "cd ${VM_PATH} && git pull origin main"
    ;;

  *)
    echo "[1/2] git pull..."
    ssh ${VM_HOST} "cd ${VM_PATH} && git pull origin main"
    echo "[2/2] ${1} 재시작..."
    ssh ${VM_HOST} "sudo systemctl restart ${1}"
    sleep 2
    ssh ${VM_HOST} "systemctl is-active ${1}"
    ;;
esac
