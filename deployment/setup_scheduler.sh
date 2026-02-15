#!/bin/bash
# Cloud Scheduler 설정 스크립트
# 무료 플랜: 3 job까지 무료

set -e

echo "⏰ Cloud Scheduler 설정 시작..."
echo "=============================================="

# 설정
PROJECT_ID=${GCP_PROJECT_ID:-"layer97os"}
REGION="us-central1"  # Cloud Scheduler 리전
CLOUD_RUN_URL="https://telegram-bot-514569077225.asia-northeast3.run.app"
VM_EXTERNAL_IP=""  # VM 생성 후 입력 필요

echo "📋 설정 확인:"
echo "   프로젝트: $PROJECT_ID"
echo "   Cloud Run URL: $CLOUD_RUN_URL"
echo "   VM IP: $VM_EXTERNAL_IP (미설정 시 스킵)"
echo ""

# 1. 프로젝트 설정
echo "✓ Google Cloud 프로젝트 설정 중..."
gcloud config set project $PROJECT_ID

# 2. Cloud Scheduler API 활성화
echo "✓ Cloud Scheduler API 활성화 중..."
gcloud services enable cloudscheduler.googleapis.com

# 3. Job 1: 매일 09:00 컨텐츠 아이디어 생성 (Cloud Run)
echo "✓ Job 1: daily-content (매일 09:00) 생성 중..."
gcloud scheduler jobs create http daily-content \
  --location=$REGION \
  --schedule="0 9 * * *" \
  --time-zone="Asia/Seoul" \
  --uri="$CLOUD_RUN_URL/scheduled/content" \
  --http-method=POST \
  --attempt-deadline=300s \
  --description="매일 오전 9시 컨텐츠 아이디어 생성 (Cloud Run)" \
  || echo "   ℹ️ Job 이미 존재하거나 생성 실패"

# 4. Job 2: 매일 06:00 트렌드 분석 (Cloud Run)
echo "✓ Job 2: daily-trends (매일 06:00) 생성 중..."
gcloud scheduler jobs create http daily-trends \
  --location=$REGION \
  --schedule="0 6 * * *" \
  --time-zone="Asia/Seoul" \
  --uri="$CLOUD_RUN_URL/scheduled/trends" \
  --http-method=POST \
  --attempt-deadline=300s \
  --description="매일 오전 6시 트렌드 분석 리포트 (Cloud Run)" \
  || echo "   ℹ️ Job 이미 존재하거나 생성 실패"

# 5. Job 3: 매주 일요일 00:00 Gardener 진화 (VM)
if [ -n "$VM_EXTERNAL_IP" ]; then
    echo "✓ Job 3: weekly-evolution (매주 일요일 00:00) 생성 중..."
    gcloud scheduler jobs create http weekly-evolution \
      --location=$REGION \
      --schedule="0 0 * * 0" \
      --time-zone="Asia/Seoul" \
      --uri="http://$VM_EXTERNAL_IP:8080/scheduled/evolution" \
      --http-method=POST \
      --attempt-deadline=600s \
      --description="매주 일요일 자정 Gardener 진화 사이클 (VM)" \
      || echo "   ℹ️ Job 이미 존재하거나 생성 실패"
else
    echo "   ⚠️ VM_EXTERNAL_IP 미설정, Job 3 스킵"
    echo "   VM 생성 후 수동으로 추가하세요:"
    echo "   export VM_EXTERNAL_IP=<VM_IP>"
    echo "   ./setup_scheduler.sh"
fi

echo ""
echo "✅ Cloud Scheduler 설정 완료!"
echo "=============================================="
echo ""
echo "생성된 Job 목록:"
gcloud scheduler jobs list --location=$REGION

echo ""
echo "Job 수동 실행 (테스트):"
echo "   gcloud scheduler jobs run daily-content --location=$REGION"
echo ""
echo "Job 삭제:"
echo "   gcloud scheduler jobs delete daily-content --location=$REGION"
echo ""
echo "비용 확인:"
echo "   무료 플랜: 3 job 무료"
echo "   현재 사용: 2-3 job"
echo "   예상 비용: \$0/월"
echo ""
