#!/bin/bash
# API 토큰 업데이트 자동화 스크립트
# 사용법: ./update_tokens.sh

set -e

PROJECT_ROOT="/Users/97layer/97layerOS"
cd "$PROJECT_ROOT"

echo "🔒 97layerOS API 토큰 업데이트"
echo "================================"
echo ""

# Step 1: .env 백업
echo "1️⃣ .env 백업 중..."
if [ -f .env ]; then
    cp .env .env.backup.$(date +%Y%m%d_%H%M%S)
    echo "   ✅ 백업 완료: .env.backup.$(date +%Y%m%d_%H%M%S)"
else
    echo "   ⚠️  .env 파일 없음"
fi

echo ""
echo "2️⃣ 새 토큰 입력"
echo "   (재발급한 토큰을 붙여넣기)"
echo ""

# 입력 받기
read -p "Telegram Bot Token: " TELEGRAM_TOKEN
read -p "Gemini API Key: " GEMINI_KEY
read -p "Anthropic API Key: " ANTHROPIC_KEY

# 검증
echo ""
echo "3️⃣ 입력 검증 중..."

if [ -z "$TELEGRAM_TOKEN" ] || [ ${#TELEGRAM_TOKEN} -lt 40 ]; then
    echo "   ❌ Telegram 토큰 형식 오류"
    exit 1
fi

if [ -z "$GEMINI_KEY" ] || [ ${#GEMINI_KEY} -lt 30 ]; then
    echo "   ❌ Gemini 키 형식 오류"
    exit 1
fi

if [ -z "$ANTHROPIC_KEY" ] || [ ${#ANTHROPIC_KEY} -lt 50 ]; then
    echo "   ❌ Anthropic 키 형식 오류"
    exit 1
fi

echo "   ✅ 형식 검증 통과"

# .env 파일 생성
echo ""
echo "4️⃣ .env 파일 업데이트 중..."
cat > .env << EOF
TELEGRAM_BOT_TOKEN=$TELEGRAM_TOKEN
GEMINI_API_KEY=$GEMINI_KEY
ANTHROPIC_API_KEY=$ANTHROPIC_KEY
EOF

chmod 600 .env
echo "   ✅ .env 파일 생성 완료 (퍼미션: 600)"

# Telegram API 테스트
echo ""
echo "5️⃣ Telegram API 테스트 중..."
RESPONSE=$(curl -s "https://api.telegram.org/bot$TELEGRAM_TOKEN/getMe")
if echo "$RESPONSE" | grep -q '"ok":true'; then
    BOT_NAME=$(echo "$RESPONSE" | grep -o '"username":"[^"]*"' | cut -d'"' -f4)
    echo "   ✅ Telegram API 정상 (봇: @$BOT_NAME)"
else
    echo "   ❌ Telegram API 실패"
    echo "   응답: $RESPONSE"
    exit 1
fi

# Cloud Run 업데이트
echo ""
echo "6️⃣ Cloud Run 환경변수 업데이트 중..."
if gcloud run services update telegram-bot \
    --region=asia-northeast3 \
    --set-env-vars "TELEGRAM_BOT_TOKEN=$TELEGRAM_TOKEN,GEMINI_API_KEY=$GEMINI_KEY,ANTHROPIC_API_KEY=$ANTHROPIC_KEY" \
    --quiet 2>&1 | grep -q "Done"; then
    echo "   ✅ Cloud Run 업데이트 완료"
else
    echo "   ⚠️  Cloud Run 업데이트 실패 (수동 확인 필요)"
fi

# VM 업데이트
echo ""
echo "7️⃣ VM 환경변수 업데이트 중..."
if gcloud compute scp .env layer97-nightguard:~/97layerOS/.env --zone=us-west1-b --quiet 2>/dev/null; then
    echo "   ✅ VM .env 동기화 완료"

    # Night Guard 재시작
    if gcloud compute ssh layer97-nightguard --zone=us-west1-b \
        --command="sudo systemctl restart 97layeros-nightguard" --quiet 2>/dev/null; then
        echo "   ✅ Night Guard 재시작 완료"
    else
        echo "   ⚠️  Night Guard 재시작 실패"
    fi
else
    echo "   ⚠️  VM 동기화 실패 (수동 확인 필요)"
fi

# Webhook 재등록
echo ""
echo "8️⃣ Telegram Webhook 재등록 중..."
WEBHOOK_URL="https://telegram-bot-514569077225.asia-northeast3.run.app/webhook"
WEBHOOK_RESPONSE=$(curl -s -X POST "https://api.telegram.org/bot$TELEGRAM_TOKEN/setWebhook?url=$WEBHOOK_URL")

if echo "$WEBHOOK_RESPONSE" | grep -q '"ok":true'; then
    echo "   ✅ Webhook 등록 완료"
else
    echo "   ❌ Webhook 등록 실패"
    echo "   응답: $WEBHOOK_RESPONSE"
fi

# 최종 검증
echo ""
echo "9️⃣ 최종 검증 중..."
echo ""

# Cloud Run 헬스체크
echo "   [Cloud Run]"
HEALTH=$(curl -s https://telegram-bot-514569077225.asia-northeast3.run.app/health)
if echo "$HEALTH" | grep -q "healthy"; then
    echo "   ✅ 헬스체크 통과"
else
    echo "   ❌ 헬스체크 실패"
fi

# Webhook 상태
echo ""
echo "   [Telegram Webhook]"
WEBHOOK_INFO=$(curl -s "https://api.telegram.org/bot$TELEGRAM_TOKEN/getWebhookInfo")
WEBHOOK_STATUS=$(echo "$WEBHOOK_INFO" | grep -o '"url":"[^"]*"' | cut -d'"' -f4)
PENDING=$(echo "$WEBHOOK_INFO" | grep -o '"pending_update_count":[0-9]*' | cut -d':' -f2)
echo "   URL: $WEBHOOK_STATUS"
echo "   대기 중: $PENDING 개"

# VM 상태
echo ""
echo "   [VM Night Guard]"
VM_STATUS=$(gcloud compute ssh layer97-nightguard --zone=us-west1-b \
    --command="sudo systemctl is-active 97layeros-nightguard" --quiet 2>/dev/null || echo "unknown")
echo "   상태: $VM_STATUS"

# 완료
echo ""
echo "================================"
echo "✅ 토큰 업데이트 완료!"
echo ""
echo "📋 다음 단계:"
echo "1. Telegram에서 봇 테스트: /start"
echo "2. 메시지 보내서 AI 응답 확인"
echo "3. VM 로그 확인 (선택):"
echo "   gcloud compute ssh layer97-nightguard --zone=us-west1-b \\"
echo "     --command='sudo journalctl -u 97layeros-nightguard -n 20'"
echo ""
