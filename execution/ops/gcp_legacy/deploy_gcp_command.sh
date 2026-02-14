cd ~
echo "🔄 배포 시작..."
pkill -f "technical_daemon.py" || true
pkill -f "telegram_daemon.py" || true
sleep 2
rm -rf 97layerOS
tar xzf /tmp/97layerOS_deploy.tar.gz
cd 97layerOS
if [ -f ".env.txt" ]; then cat .env.txt > .env; fi
echo "TELEGRAM_BOT_TOKEN=8271602365:AAGQwvDfmLv11_CShkeTMSQvnAkDYbDiTxA" >> .env
python3 -m venv .venv
source .venv/bin/activate
pip install -q google-generativeai python-dotenv requests
nohup python execution/technical_daemon.py > /tmp/technical_daemon.log 2>&1 &
echo "✅ Technical Daemon (PID: $!)"
nohup python execution/telegram_daemon.py > /tmp/telegram_daemon.log 2>&1 &
echo "✅ Telegram Daemon (PID: $!)"
sleep 3
ps aux | grep -E "technical_daemon|telegram_daemon" | grep -v grep
echo ""
echo "=== Technical Daemon 로그 ==="
tail -10 /tmp/technical_daemon.log
echo ""
echo "=== Telegram Daemon 로그 ==="
tail -10 /tmp/telegram_daemon.log
echo ""
echo "🎉 배포 완료!"
