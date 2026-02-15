#!/bin/bash
# Night Guard 초기 설정 스크립트
# VM SSH 접속 후 실행

set -e

echo "🛰️ Night Guard 초기화 시작..."
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

# 2. 패키지 업데이트
echo "✓ 패키지 업데이트 중..."
sudo apt update -qq

# 3. Python 3.10+ 설치
echo "✓ Python 3.10+ 설치 중..."
sudo apt install -y python3 python3-pip git curl

# 4. Podman 설치 (경량 컨테이너, 선택 사항)
echo "✓ Podman 설치 중..."
sudo apt install -y podman

# 5. 97layerOS 클론 (이미 클론되지 않은 경우)
echo "✓ 97layerOS 클론 중..."
if [ ! -d ~/97layerOS ]; then
    # 임시: GitHub 인증 없이 public repo 가정
    # 실제로는 SSH 키 설정 필요
    echo "   ⚠️ GitHub 인증이 필요합니다."
    echo "   수동으로 클론하거나 SSH 키를 설정하세요:"
    echo "   git clone git@github.com:your-org/97layerOS.git"
    echo ""
    echo "   또는 deployment/로 파일 복사:"
    echo "   gcloud compute scp --recurse ../97layerOS 97layer-nightguard:~/ --zone=us-west1-b"
else
    echo "   ℹ️ 97layerOS 이미 존재"
    cd ~/97layerOS
    git pull
fi

# 6. 환경변수 설정
echo "✓ 환경변수 설정 중..."
cd ~/97layerOS
cat > .env << EOF
ENVIRONMENT=GCP_VM
PROCESSING_MODE=sequential
ENABLE_MULTIMODAL=false
TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
GEMINI_API_KEY=${GEMINI_API_KEY}
ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
EOF

# 7. Python 의존성 설치
echo "✓ Python 의존성 설치 중..."
if [ -f requirements.txt ]; then
    pip3 install -r requirements.txt --quiet
else
    echo "   ⚠️ requirements.txt 없음"
fi

# 8. systemd 서비스 등록
echo "✓ systemd 서비스 등록 중..."
sudo cp deployment/97layeros-nightguard.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable 97layeros-nightguard
sudo systemctl start 97layeros-nightguard

# 9. 상태 확인
echo ""
echo "✅ Night Guard 가동 완료!"
echo "=============================================="
echo ""
echo "상태 확인:"
sudo systemctl status 97layeros-nightguard --no-pager

echo ""
echo "로그 확인:"
echo "   sudo journalctl -u 97layeros-nightguard -f"
echo ""
echo "서비스 관리:"
echo "   sudo systemctl stop 97layeros-nightguard"
echo "   sudo systemctl restart 97layeros-nightguard"
echo ""
echo "Swap 확인:"
free -h
echo ""
