#!/bin/bash
# LAYER OS 배포 스크립트
# SSOT: knowledge/system/vm_services.json
#
# Usage:
#   ./deploy.sh              코드 pull만 (재시작 없음)
#   ./deploy.sh all          코드 pull + active 서비스 전체 재시작
#   ./deploy.sh web          Cloudflare Pages 안내
#   ./deploy.sh <서비스명>   특정 서비스 재시작 (active만 허용)
#   ./deploy.sh --status     서비스 상태 확인
#   ./deploy.sh --list       배포 가능 서비스 목록

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
VM_SERVICES_JSON="$PROJECT_ROOT/knowledge/system/vm_services.json"

# vm_services.json에서 설정 읽기
if [ ! -f "$VM_SERVICES_JSON" ]; then
  echo "ERROR: $VM_SERVICES_JSON 없음. 배포 중단." >&2
  exit 1
fi

VM_HOST=$(python3 -c "import json; d=json.load(open('$VM_SERVICES_JSON')); print(d['vm']['host_alias'])")
VM_PATH=$(python3 -c "import json; d=json.load(open('$VM_SERVICES_JSON')); print(d['vm']['app_path'])")
SERVICES_ACTIVE=$(python3 -c "import json; d=json.load(open('$VM_SERVICES_JSON')); print(' '.join(d['active'].keys()))")

# 서비스가 active 목록에 있는지 확인
is_active_service() {
  python3 -c "
import json, sys
d = json.load(open('$VM_SERVICES_JSON'))
sys.exit(0 if '$1' in d['active'] else 1)
" 2>/dev/null
}

case "${1:-pull}" in
  web)
    echo "웹은 git push만 하면 Cloudflare Pages가 자동 배포합니다."
    echo "  git push origin main → 30초 내 woohwahae.kr 반영"
    ;;

  --status)
    echo "=== VM 서비스 상태 (active 목록) ==="
    ssh ${VM_HOST} "for s in ${SERVICES_ACTIVE}; do printf '%-25s %s\n' \$s \$(systemctl is-active \$s 2>/dev/null || echo 'not-found'); done"
    ;;

  --list)
    echo "=== 배포 가능 서비스 (active) ==="
    python3 -c "
import json
d = json.load(open('$VM_SERVICES_JSON'))
for name, info in d['active'].items():
    port = info.get('port') or '-'
    print(f'  {name:<25} port={port:<6} {info[\"role\"]}')
print()
print('=== 배포 불가 (inactive) ===')
for name, info in d['inactive'].items():
    print(f'  {name:<25} {info[\"reason\"]}')
"
    ;;

  all)
    echo "[1/3] git pull..."
    ssh ${VM_HOST} "cd ${VM_PATH} && git fetch origin main && git reset --hard origin/main"
    echo "[2/3] corpus 군집 마이그레이션..."
    ssh ${VM_HOST} "cd ${VM_PATH} && python3 core/scripts/migrate_corpus_clusters.py"
    echo "[3/3] active 서비스 전체 재시작 (${SERVICES_ACTIVE})..."
    ssh ${VM_HOST} "sudo systemctl restart ${SERVICES_ACTIVE}"
    sleep 3
    ssh ${VM_HOST} "for s in ${SERVICES_ACTIVE}; do printf '%-25s %s\n' \$s \$(systemctl is-active \$s); done"
    ;;

  pull)
    echo "git pull..."
    ssh ${VM_HOST} "cd ${VM_PATH} && git fetch origin main && git reset --hard origin/main"
    ;;

  *)
    # 서비스명 검증
    if ! is_active_service "$1"; then
      echo "ERROR: '$1'은(는) active 서비스가 아닙니다." >&2
      echo "배포 가능 목록: ${SERVICES_ACTIVE}" >&2
      echo "전체 목록: ./deploy.sh --list" >&2
      exit 1
    fi

    echo "[1/2] git pull..."
    ssh ${VM_HOST} "cd ${VM_PATH} && git fetch origin main && git reset --hard origin/main"
    echo "[2/2] ${1} 재시작..."
    ssh ${VM_HOST} "sudo systemctl restart ${1}"
    sleep 2
    ssh ${VM_HOST} "systemctl is-active ${1}"
    ;;
esac
