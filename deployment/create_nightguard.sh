#!/bin/bash
# 97LAYER Night Guard (정찰기) 배치 스크립트
# GCP VM 생성 (무료 플랜 전용: us-west1-b)

set -e

echo "🛰️ 97LAYER Night Guard 배치 시작..."
echo "=============================================="

# 설정
INSTANCE_NAME="layer97-nightguard"  # 숫자로 시작 불가
ZONE="us-west1-b"  # 오리건 (무료 리전)
MACHINE_TYPE="e2-micro"  # 무료 플랜
BOOT_DISK_SIZE="30GB"
IMAGE_FAMILY="ubuntu-minimal-2204-lts"  # 경량 Ubuntu
IMAGE_PROJECT="ubuntu-os-cloud"
PROJECT_ID=${GCP_PROJECT_ID:-"layer97os"}

echo "📋 설정 확인:"
echo "   프로젝트: $PROJECT_ID"
echo "   인스턴스명: $INSTANCE_NAME"
echo "   리전/존: $ZONE"
echo "   머신 타입: $MACHINE_TYPE (무료)"
echo "   디스크: $BOOT_DISK_SIZE"
echo ""

# 1. 프로젝트 설정
echo "✓ Google Cloud 프로젝트 설정 중..."
gcloud config set project $PROJECT_ID

# 2. Compute Engine API 활성화
echo "✓ Compute Engine API 확인 중..."
gcloud services enable compute.googleapis.com

# 3. 인스턴스 생성
echo "✓ VM 인스턴스 생성 중..."
gcloud compute instances create $INSTANCE_NAME \
  --zone=$ZONE \
  --machine-type=$MACHINE_TYPE \
  --boot-disk-size=$BOOT_DISK_SIZE \
  --boot-disk-type=pd-standard \
  --image-family=$IMAGE_FAMILY \
  --image-project=$IMAGE_PROJECT \
  --tags=layer97-nightguard,http-server \
  --metadata=enable-oslogin=true \
  --scopes=cloud-platform

echo ""
echo "✅ Night Guard 배치 완료!"
echo "=============================================="
echo ""
echo "다음 단계:"
echo "1. SSH 접속:"
echo "   gcloud compute ssh $INSTANCE_NAME --zone=$ZONE"
echo ""
echo "2. 초기화 스크립트 실행:"
echo "   cd 97layerOS/deployment"
echo "   chmod +x init_nightguard.sh"
echo "   ./init_nightguard.sh"
echo ""
echo "3. VM 상태 확인:"
echo "   gcloud compute instances list"
echo ""
echo "4. 외부 IP 확인:"
gcloud compute instances describe $INSTANCE_NAME --zone=$ZONE --format="get(networkInterfaces[0].accessConfigs[0].natIP)"
echo ""
