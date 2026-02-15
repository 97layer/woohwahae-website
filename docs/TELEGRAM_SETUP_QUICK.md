# 텔레그램 봇 클라우드 배포 - 빠른 시작 가이드

## 🎯 목표
맥북 없이도 24/7 양방향 텔레그램 통신 가능하게 만들기

## 🚀 빠른 배포 (3단계)

### 1단계: 환경변수 설정

```bash
cd /Users/97layer/97layerOS

# .env 파일의 값들을 export
export TELEGRAM_BOT_TOKEN="8501568801:AAE-3fBl-p6uZcmrdsWSRQuz_eg8yDADwjI"
export GEMINI_API_KEY="AIzaSyBHpQRFjdZRzzkYGR6eqBezyPteaHX_uMQ"
export ANTHROPIC_API_KEY="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"
```

### 2단계: Google Cloud Run 배포

```bash
# 자동 배포 스크립트 실행
./deploy_google_cloud.sh
```

배포 완료 후 서비스 URL이 출력됩니다 (예: `https://telegram-bot-xxxxx-xx.a.run.app`)

### 3단계: Webhook으로 전환

```bash
# 출력된 서비스 URL을 사용
python execution/switch_to_webhook.py https://telegram-bot-xxxxx-xx.a.run.app
```

## ✅ 확인

텔레그램에서 봇에게 `/start` 보내기:

**예상 응답:**
```
97LAYER OS Online (Webhook Mode).

명령어:
/cd /td /ad /ce /sa - 에이전트 전환
/auto - 자동 라우팅
/status - 상태 확인
/evolve - 시스템 진화
/council [주제] - 위원회 소집
```

## 📊 모니터링

```bash
# 실시간 로그 확인
gcloud run logs tail telegram-bot --region asia-northeast3

# Health check
curl [서비스_URL]/health
```

## 🔧 문제 해결

### 409 Conflict 여전히 발생

```bash
# 기존 프로세스 확인 및 종료
ps aux | grep telegram_daemon
kill -9 [PID]

# Webhook 재설정
python execution/switch_to_webhook.py [서비스_URL]
```

### 응답 없음

```bash
# 로그 확인
gcloud run logs tail telegram-bot --region asia-northeast3

# Webhook 상태 확인
curl "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getWebhookInfo"
```

## 📁 생성된 파일

- `execution/telegram_webhook.py` - Webhook 서버 (Flask)
- `Dockerfile` - 컨테이너 이미지 정의
- `.dockerignore` - 컨테이너에서 제외할 파일
- `deploy_google_cloud.sh` - 자동 배포 스크립트
- `execution/switch_to_webhook.py` - 모드 전환 스크립트
- `docs/TELEGRAM_CLOUD_DEPLOYMENT.md` - 상세 가이드

## 🎉 완료!

이제 맥북을 끄더라도 텔레그램 봇이 Google Cloud에서 계속 작동합니다.

---

**자세한 내용**: [docs/TELEGRAM_CLOUD_DEPLOYMENT.md](docs/TELEGRAM_CLOUD_DEPLOYMENT.md)
