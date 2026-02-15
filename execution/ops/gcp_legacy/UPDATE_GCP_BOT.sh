#!/bin/bash
# GCP 봇 토큰 업데이트 스크립트

echo "========================================"
echo "🔧 GCP 봇 토큰 업데이트 스크립트"
echo "========================================"

# 새 봇 토큰
NEW_TOKEN="${TELEGRAM_BOT_TOKEN}"  # 환경변수에서 읽기
GCP_IP="35.184.30.182"
GCP_USER="skyto5339"
SSH_KEY="$HOME/.ssh/id_ed25519_gcp"

echo ""
echo "1️⃣ .env 파일 업데이트 준비..."
cat > /tmp/update_bot_token.sh << 'EOF'
#!/bin/bash

# 환경 변수 파일 업데이트
cd ~/97layerOS
echo "현재 토큰 확인:"
grep TELEGRAM_BOT_TOKEN .env

# 새 토큰으로 교체 (환경변수에서 읽음)
sed -i "s/TELEGRAM_BOT_TOKEN=.*/TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}/" .env

echo ""
echo "업데이트 후:"
grep TELEGRAM_BOT_TOKEN .env

# 기존 telegram_daemon 종료
echo ""
echo "기존 프로세스 종료..."
pkill -f telegram_daemon.py || true

# 5초 대기
sleep 5

# 새 토큰으로 재시작
echo "새 토큰으로 재시작..."
cd ~/97layerOS
source .venv/bin/activate
nohup python execution/telegram_daemon.py > logs/telegram.log 2>&1 &

echo "✅ 완료! 새 PID: $!"
EOF

echo ""
echo "2️⃣ GCP로 스크립트 전송..."
scp -i "$SSH_KEY" /tmp/update_bot_token.sh $GCP_USER@$GCP_IP:/tmp/

echo ""
echo "3️⃣ GCP에서 실행..."
ssh -i "$SSH_KEY" $GCP_USER@$GCP_IP "bash /tmp/update_bot_token.sh"

echo ""
echo "========================================"
echo "✅ GCP 봇 토큰 업데이트 완료!"
echo "새 토큰: @official_97Layer_OSwoohwahae_bot"
echo "========================================"