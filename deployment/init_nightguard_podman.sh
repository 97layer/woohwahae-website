#!/bin/bash
# Night Guard 초기화 스크립트 (Podman 버전)
# GCP VM에서 Podman 컨테이너 환경 구축

set -e

echo "🛰️ Night Guard 초기화 시작 (Podman 최적화)..."
echo "=============================================="

# 1. Swap 2GB 생성 (RAM 1GB 극복)
echo "✓ Swap Memory 2GB 생성 중..."
if [ ! -f /swapfile ]; then
    sudo fallocate -l 2G /swapfile
    sudo chmod 600 /swapfile
    sudo mkswap /swapfile
    sudo swapon /swapfile
    echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
    echo "   ✅ Swap 활성화 완료"
else
    echo "   ℹ️ Swap 이미 존재"
fi

# Swap 확인
echo ""
echo "Swap 상태:"
free -h
echo ""

# 2. 패키지 업데이트
echo "✓ 패키지 업데이트 중..."
sudo apt update -qq

# 3. 필수 패키지 설치
echo "✓ 필수 패키지 설치 중 (Podman, Git, Python)..."
sudo apt install -y \
    podman \
    podman-compose \
    python3 \
    python3-pip \
    git \
    curl \
    jq

# Podman 버전 확인
echo ""
echo "Podman 버전:"
podman version | head -3
echo ""

# 4. 97layerOS 클론 또는 동기화
echo "✓ 97layerOS 동기화 중..."
if [ ! -d ~/97layerOS ]; then
    echo "   ⚠️ 97layerOS 디렉토리가 없습니다."
    echo "   수동으로 클론하거나 맥북에서 복사하세요:"
    echo ""
    echo "   # 맥북에서:"
    echo "   gcloud compute scp --recurse /Users/97layer/97layerOS 97layer-nightguard:~/ --zone=us-west1-b"
    echo ""
    echo "   또는 Git clone:"
    echo "   git clone git@github.com:your-org/97layerOS.git ~/97layerOS"
    echo ""
    read -p "97layerOS가 준비되었으면 Enter를 누르세요..."
else
    echo "   ℹ️ 97layerOS 이미 존재"
    cd ~/97layerOS

    # Git 업데이트 시도 (선택)
    if [ -d .git ]; then
        git pull || echo "   ⚠️ Git pull 실패 (수동 업데이트 필요)"
    fi
fi

# 5. Python 의존성 설치 (선택 - 컨테이너 빌드 시 포함됨)
echo "✓ Python 의존성 확인 중..."
cd ~/97layerOS
if [ -f requirements.txt ]; then
    echo "   ℹ️ requirements.txt 발견 (Docker 빌드 시 설치됨)"
else
    echo "   ⚠️ requirements.txt 없음"
fi

# 6. Podman Secrets 설정
echo "✓ Podman Secrets 설정 중..."
echo ""
echo "환경변수를 설정해주세요:"
echo ""

# .env 파일이 있으면 읽기
if [ -f ~/97layerOS/.env ]; then
    echo "   ℹ️ .env 파일 발견"
    source ~/97layerOS/.env
    echo "   ✅ 환경변수 로드 완료"
else
    echo "   ⚠️ .env 파일 없음"
    echo "   다음 명령어로 환경변수를 설정하세요:"
    echo ""
    echo "   export TELEGRAM_BOT_TOKEN='your_token'"
    echo "   export GEMINI_API_KEY='your_key'"
    echo "   export ANTHROPIC_API_KEY='your_key'"
    echo ""
    read -p "환경변수를 설정한 후 Enter를 누르세요..."
fi

# Secrets 초기화 스크립트 실행
cd ~/97layerOS/deployment
if [ -f setup_podman_secrets.sh ]; then
    ./setup_podman_secrets.sh
else
    echo "   ⚠️ setup_podman_secrets.sh 없음"
    echo "   Secrets를 수동으로 생성하세요:"
    echo ""
    echo "   echo -n '\$TELEGRAM_BOT_TOKEN' | podman secret create telegram_bot_token -"
    echo "   echo -n '\$GEMINI_API_KEY' | podman secret create gemini_api_key -"
    echo "   echo -n '\$ANTHROPIC_API_KEY' | podman secret create anthropic_api_key -"
fi

# 7. Podman 이미지 빌드
echo ""
echo "✓ Night Guard 컨테이너 이미지 빌드 중..."
cd ~/97layerOS
if [ -f deployment/Dockerfile.nightguard ]; then
    podman build -t 97layer-nightguard:latest -f deployment/Dockerfile.nightguard .
    echo "   ✅ 이미지 빌드 완료"
else
    echo "   ⚠️ Dockerfile.nightguard 없음"
    echo "   컨테이너 실행 시 자동으로 빌드됩니다."
fi

# 8. Podman Compose로 컨테이너 실행
echo ""
echo "✓ Night Guard 컨테이너 실행 중..."
cd ~/97layerOS
podman-compose -f deployment/podman-compose.nightguard.yml up -d

# 실행 확인
echo ""
echo "✅ Night Guard 가동 완료!"
echo "=============================================="
echo ""

# 9. 상태 확인
echo "컨테이너 상태:"
podman ps -a | grep nightguard
echo ""

echo "로그 확인 (최근 20줄):"
podman logs --tail 20 97layer-nightguard
echo ""

echo "Healthcheck 상태:"
podman inspect 97layer-nightguard --format '{{.State.Health.Status}}' || echo "Healthcheck 대기 중..."
echo ""

# 10. systemd 서비스 등록 (자동 시작)
echo "✓ systemd 서비스 등록 중..."
podman generate systemd --new --name 97layer-nightguard | sudo tee /etc/systemd/system/97layer-nightguard.service > /dev/null
sudo systemctl daemon-reload
sudo systemctl enable 97layer-nightguard
echo "   ✅ 자동 시작 활성화"

echo ""
echo "=============================================="
echo "🎉 Night Guard Podman 환경 구축 완료!"
echo "=============================================="
echo ""
echo "📋 관리 명령어:"
echo ""
echo "  # 로그 실시간 보기"
echo "  podman logs -f 97layer-nightguard"
echo ""
echo "  # 컨테이너 상태 확인"
echo "  podman ps"
echo ""
echo "  # 컨테이너 재시작"
echo "  podman-compose -f deployment/podman-compose.nightguard.yml restart"
echo ""
echo "  # 컨테이너 중지"
echo "  podman-compose -f deployment/podman-compose.nightguard.yml down"
echo ""
echo "  # Healthcheck 상태"
echo "  podman inspect 97layer-nightguard | grep -A10 Health"
echo ""
echo "  # 리소스 사용량"
echo "  podman stats 97layer-nightguard"
echo ""
echo "  # Secrets 확인"
echo "  podman secret ls"
echo ""
echo "  # systemd 서비스 상태"
echo "  sudo systemctl status 97layer-nightguard"
echo ""
