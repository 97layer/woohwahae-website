#!/bin/bash
# GCP 브라우저 SSH에 붙여넣기용 명령어
# 97LAYER Systemd Services Installation

cd ~/97layerOS

echo "🔧 97LAYER Systemd 서비스 설치 시작..."

# 1. 기존 프로세스 중지
echo "1️⃣ 기존 프로세스 중지 중..."
pkill -f "technical_daemon.py" || true
pkill -f "telegram_daemon.py" || true
sleep 2

# 2. 서비스 파일 복사
echo "2️⃣ 서비스 파일 설치 중..."
sudo cp ~/97layerOS/97layer_technical.service /etc/systemd/system/
sudo cp ~/97layerOS/97layer_telegram.service /etc/systemd/system/
sudo chmod 644 /etc/systemd/system/97layer_technical.service
sudo chmod 644 /etc/systemd/system/97layer_telegram.service

# 3. Systemd 재로드
echo "3️⃣ Systemd 데몬 재로드 중..."
sudo systemctl daemon-reload

# 4. 부팅 시 자동 시작 활성화
echo "4️⃣ 자동 시작 활성화 중..."
sudo systemctl enable 97layer_technical.service
sudo systemctl enable 97layer_telegram.service

# 5. 서비스 시작
echo "5️⃣ 서비스 시작 중..."
sudo systemctl start 97layer_technical.service
sudo systemctl start 97layer_telegram.service

# 6. 상태 확인
sleep 3
echo ""
echo "✅ 설치 완료!"
echo ""
echo "📊 Technical Daemon 상태:"
sudo systemctl status 97layer_technical.service --no-pager -l | head -20
echo ""
echo "📊 Telegram Daemon 상태:"
sudo systemctl status 97layer_telegram.service --no-pager -l | head -20
echo ""
echo "🔍 프로세스 확인:"
ps aux | grep -E "technical_daemon|telegram_daemon" | grep -v grep
echo ""
echo "✨ Systemd 서비스 설치 완료! 이제 서버가 재부팅되어도 자동으로 실행됩니다."
