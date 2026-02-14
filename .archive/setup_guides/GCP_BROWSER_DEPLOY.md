# GCP 브라우저 SSH로 배포하기

**날짜**: 2026-02-14
**파일 위치**: `/tmp/97layerOS_deploy.tar.gz` (이미 GCP 서버에 업로드됨)

---

## 1단계: GCP 브라우저 SSH 열기

1. https://console.cloud.google.com/compute/instances 접속
2. 35.184.30.182 인스턴스 옆의 **"SSH"** 버튼 클릭
3. 브라우저 SSH 창이 열림

---

## 2단계: 아래 명령어 전체 복사해서 붙여넣기

```bash
cd ~
echo "🔄 배포 시작..."

# 1. 기존 Daemon 종료
pkill -f "technical_daemon.py" || true
pkill -f "telegram_daemon.py" || true
sleep 2

# 2. 압축 해제
rm -rf 97layerOS
tar xzf /tmp/97layerOS_deploy.tar.gz
cd 97layerOS

# 3. .env 설정
if [ -f ".env.txt" ]; then
  cat .env.txt > .env
fi

# Telegram Bot Token 추가 (중요!)
echo "TELEGRAM_BOT_TOKEN=8271602365:AAGQwvDfmLv11_CShkeTMSQvnAkDYbDiTxA" >> .env

# 4. Python 환경 설정
python3 -m venv .venv
source .venv/bin/activate
pip install -q google-generativeai python-dotenv requests

# 5. Daemon 재시작
nohup python execution/technical_daemon.py > /tmp/technical_daemon.log 2>&1 &
echo "✅ Technical Daemon (PID: $!)"

nohup python execution/telegram_daemon.py > /tmp/telegram_daemon.log 2>&1 &
echo "✅ Telegram Daemon (PID: $!)"

sleep 3

# 6. 상태 확인
ps aux | grep -E "technical_daemon|telegram_daemon" | grep -v grep

echo ""
echo "=== Technical Daemon 로그 ==="
tail -10 /tmp/technical_daemon.log

echo ""
echo "=== Telegram Daemon 로그 ==="
tail -10 /tmp/telegram_daemon.log

echo ""
echo "🎉 배포 완료!"
```

---

## 3단계: 결과 확인

정상적으로 실행되면 다음과 같이 표시됩니다:

```
✅ Technical Daemon (PID: 12345)
✅ Telegram Daemon (PID: 12346)

skyto5339  12345  0.0  1.7  28372 17300 ?  Ss  08:45  0:00 python execution/technical_daemon.py
skyto5339  12346  0.0  1.8  28316 18164 ?  Ss  08:45  0:00 python execution/telegram_daemon.py

🎉 배포 완료!
```

---

## 4단계: 텔레그램 테스트

배포 완료 후 텔레그램에서 테스트:

```
/status
```

GCP 서버가 응답하면 성공!

---

## 문제 해결

### Daemon이 실행되지 않으면:

```bash
# 로그 확인
tail -50 /tmp/technical_daemon.log
tail -50 /tmp/telegram_daemon.log

# 수동 실행으로 에러 확인
cd ~/97layerOS
source .venv/bin/activate
python execution/telegram_daemon.py
```

### .env 파일 확인:

```bash
cat ~/97layerOS/.env
```

다음 2줄이 있어야 함:
```
GEMINI_API_KEY=AIzaSyBHpQRFjdZRzzkYGR6eqBezyPteaHX_uMQ
TELEGRAM_BOT_TOKEN=8271602365:AAGQwvDfmLv11_CShkeTMSQvnAkDYbDiTxA
```

---

이 파일을 보고 GCP 브라우저 SSH에서 직접 배포해주세요!
