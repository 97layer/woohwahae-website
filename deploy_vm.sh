#!/bin/bash
# 97layerOS → GCP VM 배포 스크립트
# 사용법: ./deploy_vm.sh [파일경로1] [파일경로2] ...
# 전체 배포: ./deploy_vm.sh --all
# 재시작만: ./deploy_vm.sh --restart

GCLOUD="/opt/homebrew/bin/gcloud"
VM="layer97-nightguard"
ZONE="us-west1-b"
PROJECT="layer97os"
REMOTE_BASE="~/97layerOS"

SCP() {
  $GCLOUD compute scp "$@" \
    --zone=$ZONE --project=$PROJECT --tunnel-through-iap 2>/dev/null
}

SSH() {
  $GCLOUD compute ssh $VM \
    --zone=$ZONE --project=$PROJECT --tunnel-through-iap \
    --ssh-flag="-T" < /dev/stdin 2>/dev/null
}

case "$1" in
  --restart)
    echo "=== 서비스 재시작 ==="
    echo "sudo systemctl restart 97layer-telegram && sudo systemctl status 97layer-telegram --no-pager | head -5" | SSH
    ;;

  --all)
    echo "=== 전체 core/ 배포 ==="
    # core/system
    SCP /Users/97layer/97layerOS/core/system/*.py $VM:$REMOTE_BASE/core/system/
    # core/agents
    SCP /Users/97layer/97layerOS/core/agents/*.py $VM:$REMOTE_BASE/core/agents/
    # core/bridges
    SCP /Users/97layer/97layerOS/core/bridges/*.py $VM:$REMOTE_BASE/core/bridges/
    echo "=== 서비스 재시작 ==="
    echo "sudo systemctl restart 97layer-telegram && echo '✅ 재시작 완료'" | SSH
    ;;

  --status)
    echo "sudo systemctl status 97layer-telegram --no-pager | head -10 && sudo journalctl -u 97layer-telegram -n 15 --no-pager" | SSH
    ;;

  --log)
    echo "sudo journalctl -u 97layer-telegram -n 30 --no-pager" | SSH
    ;;

  *)
    if [ $# -eq 0 ]; then
      echo "사용법:"
      echo "  ./deploy_vm.sh --all        전체 core/ 배포 + 재시작"
      echo "  ./deploy_vm.sh --restart    서비스 재시작만"
      echo "  ./deploy_vm.sh --status     서비스 상태 확인"
      echo "  ./deploy_vm.sh --log        로그 확인"
      echo "  ./deploy_vm.sh [파일]       특정 파일 배포"
      exit 0
    fi
    # 특정 파일 배포
    for f in "$@"; do
      # 경로에서 디렉토리 구조 유지
      rel="${f#/Users/97layer/97layerOS/}"
      remote_dir="$REMOTE_BASE/$(dirname $rel)"
      echo "📤 $rel → VM:$remote_dir/"
      SCP "$f" $VM:$remote_dir/
    done
    echo "=== 서비스 재시작 ==="
    echo "sudo systemctl restart 97layer-telegram && echo '✅ 완료'" | SSH
    ;;
esac
