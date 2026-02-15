#!/bin/bash
# Google Cloud Run 배포 스크립트

set -e

echo "🚀 97LAYER Telegram Bot - Google Cloud Run 배포"
echo "================================================"

# 프로젝트 설정
PROJECT_ID=${GCP_PROJECT_ID:-"97layer-os"}
REGION=${GCP_REGION:-"asia-northeast3"}  # 서울 리전
SERVICE_NAME="telegram-bot"

# 1. 환경변수 확인
echo "✓ 환경변수 확인 중..."
if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    echo "❌ TELEGRAM_BOT_TOKEN이 설정되지 않았습니다."
    echo "   export TELEGRAM_BOT_TOKEN=your_token"
    exit 1
fi

if [ -z "$GEMINI_API_KEY" ]; then
    echo "❌ GEMINI_API_KEY가 설정되지 않았습니다."
    echo "   export GEMINI_API_KEY=your_key"
    exit 1
fi

# 2. Google Cloud 프로젝트 설정
echo "✓ Google Cloud 프로젝트 설정 중..."
gcloud config set project $PROJECT_ID

# 3. 필요한 API 활성화
echo "✓ 필요한 API 활성화 중..."
gcloud services enable run.googleapis.com
gcloud services enable containerregistry.googleapis.com
gcloud services enable cloudbuild.googleapis.com

# 4. Cloud Run에 배포
echo "✓ Cloud Run에 배포 중..."
gcloud run deploy $SERVICE_NAME \
    --source . \
    --platform managed \
    --region $REGION \
    --allow-unauthenticated \
    --set-env-vars "TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN,GEMINI_API_KEY=$GEMINI_API_KEY,ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY" \
    --memory 1Gi \
    --timeout 300 \
    --min-instances 1 \
    --max-instances 10

# 5. 배포된 URL 가져오기
echo "✓ 배포된 URL 확인 중..."
SERVICE_URL=$(gcloud run services describe $SERVICE_NAME --region $REGION --format 'value(status.url)')

echo ""
echo "✅ 배포 완료!"
echo "================================================"
echo "서비스 URL: $SERVICE_URL"
echo ""
echo "다음 단계:"
echo "1. Webhook URL 설정:"
echo "   curl -X POST \"https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/setWebhook?url=$SERVICE_URL/webhook\""
echo ""
echo "2. Webhook 상태 확인:"
echo "   curl \"https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getWebhookInfo\""
echo ""
echo "3. Health Check:"
echo "   curl $SERVICE_URL/health"
echo ""
