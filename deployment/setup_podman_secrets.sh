#!/bin/bash
# Podman Secrets 초기화 스크립트
# API 키를 안전하게 Podman Secrets에 등록

set -e

echo "🔐 Podman Secrets 초기화 시작..."
echo "=============================================="

# 환경변수 확인
if [ -z "$TELEGRAM_BOT_TOKEN" ] || [ -z "$GEMINI_API_KEY" ]; then
    echo "⚠️ 환경변수가 설정되지 않았습니다."
    echo ""
    echo "다음 명령어로 환경변수를 설정하세요:"
    echo ""
    echo "  export TELEGRAM_BOT_TOKEN='your_token'"
    echo "  export GEMINI_API_KEY='your_key'"
    echo "  export ANTHROPIC_API_KEY='your_key'"
    echo ""
    echo "또는 .env 파일에서 읽어오려면:"
    echo ""
    echo "  source ~/97layerOS/.env"
    echo "  ./setup_podman_secrets.sh"
    echo ""
    exit 1
fi

# 기존 Secrets 삭제 (재설정 시)
echo "✓ 기존 Secrets 정리 중..."
podman secret rm telegram_bot_token 2>/dev/null || true
podman secret rm gemini_api_key 2>/dev/null || true
podman secret rm anthropic_api_key 2>/dev/null || true

# Telegram Bot Token 등록
echo "✓ Telegram Bot Token 등록 중..."
echo -n "$TELEGRAM_BOT_TOKEN" | podman secret create telegram_bot_token -

# Gemini API Key 등록
echo "✓ Gemini API Key 등록 중..."
echo -n "$GEMINI_API_KEY" | podman secret create gemini_api_key -

# Anthropic API Key 등록 (선택)
if [ -n "$ANTHROPIC_API_KEY" ]; then
    echo "✓ Anthropic API Key 등록 중..."
    echo -n "$ANTHROPIC_API_KEY" | podman secret create anthropic_api_key -
else
    echo "ℹ️ ANTHROPIC_API_KEY 미설정 (선택 사항)"
    # 빈 Secret 생성 (Compose 호환성)
    echo -n "" | podman secret create anthropic_api_key -
fi

echo ""
echo "✅ Podman Secrets 등록 완료!"
echo "=============================================="
echo ""

# Secrets 목록 확인
echo "등록된 Secrets:"
podman secret ls

echo ""
echo "📋 다음 단계:"
echo "  1. Podman Compose 실행:"
echo "     podman-compose -f deployment/podman-compose.nightguard.yml up -d"
echo ""
echo "  2. 로그 확인:"
echo "     podman logs -f 97layer-nightguard"
echo ""
echo "  3. Healthcheck 상태:"
echo "     podman inspect 97layer-nightguard | grep -A10 Health"
echo ""
