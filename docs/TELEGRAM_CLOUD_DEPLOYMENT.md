# 텔레그램 봇 Google Cloud 배포 가이드

맥북 없이도 24/7 양방향 통신 가능한 텔레그램 봇 시스템

## 🎯 목표

- **409 Conflict 완전 해결**: Webhook 방식으로 중복 폴링 제거
- **맥북 독립 운영**: Google Cloud Run에서 24/7 자동 실행
- **안정적 양방향 통신**: 메시지 수신/발신 모두 클라우드에서 처리

## 📋 사전 준비

### 1. Google Cloud 프로젝트 생성

```bash
# gcloud CLI 설치 확인
gcloud --version

# 로그인
gcloud auth login

# 프로젝트 생성 (선택사항 - 기존 프로젝트 사용 가능)
gcloud projects create 97layer-os --name="97LAYER OS"

# 프로젝트 설정
gcloud config set project 97layer-os
```

### 2. 환경 변수 설정

```bash
# .env 파일에서 값 가져오기
export TELEGRAM_BOT_TOKEN="8501568801:AAE-3fBl-p6uZcmrdsWSRQuz_eg8yDADwjI"
export GEMINI_API_KEY="AIzaSyBHpQRFjdZRzzkYGR6eqBezyPteaHX_uMQ"
export ANTHROPIC_API_KEY="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"
```

## 🚀 배포 방법

### 방법 1: 자동 배포 스크립트 (권장)

```bash
cd /Users/97layer/97layerOS

# 환경변수 설정 후 실행
./deploy_google_cloud.sh
```

### 방법 2: 수동 배포

#### Step 1: API 활성화

```bash
gcloud services enable run.googleapis.com
gcloud services enable containerregistry.googleapis.com
gcloud services enable cloudbuild.googleapis.com
```

#### Step 2: Cloud Run 배포

```bash
gcloud run deploy telegram-bot \
    --source . \
    --platform managed \
    --region asia-northeast3 \
    --allow-unauthenticated \
    --set-env-vars "TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN,GEMINI_API_KEY=$GEMINI_API_KEY,ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY" \
    --memory 1Gi \
    --timeout 300 \
    --min-instances 1 \
    --max-instances 10
```

#### Step 3: 배포된 URL 확인

```bash
SERVICE_URL=$(gcloud run services describe telegram-bot --region asia-northeast3 --format 'value(status.url)')
echo $SERVICE_URL
```

#### Step 4: Webhook 설정

```bash
# Webhook 등록
curl -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/setWebhook?url=$SERVICE_URL/webhook"

# Webhook 상태 확인
curl "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getWebhookInfo"
```

## ✅ 배포 확인

### 1. Health Check

```bash
curl $SERVICE_URL/health
```

**예상 응답:**
```json
{
  "status": "healthy",
  "timestamp": "2026-02-15T12:00:00",
  "service": "97LAYER Telegram Webhook"
}
```

### 2. Telegram 봇 테스트

1. 텔레그램에서 봇과 대화 시작
2. `/start` 명령어 입력
3. 정상 응답 확인:
   ```
   97LAYER OS Online (Webhook Mode).

   명령어:
   /cd /td /ad /ce /sa - 에이전트 전환
   /auto - 자동 라우팅
   /status - 상태 확인
   /evolve - 시스템 진화
   /council [주제] - 위원회 소집
   ```

### 3. 로그 확인

```bash
# 실시간 로그 모니터링
gcloud run logs tail telegram-bot --region asia-northeast3

# 최근 로그 확인
gcloud run logs read telegram-bot --region asia-northeast3 --limit 50
```

## 🔧 문제 해결

### 409 Conflict 여전히 발생하는 경우

**원인**: 기존 polling 방식의 telegram_daemon.py가 여전히 실행 중

**해결**:
```bash
# 로컬에서 실행 중인 telegram_daemon 중지
ps aux | grep telegram_daemon
kill -9 [PID]

# 기존 webhook 제거 후 재설정
curl -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/deleteWebhook"
curl -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/setWebhook?url=$SERVICE_URL/webhook"
```

### 메시지 응답이 없는 경우

1. **로그 확인**:
   ```bash
   gcloud run logs tail telegram-bot --region asia-northeast3
   ```

2. **환경변수 확인**:
   ```bash
   gcloud run services describe telegram-bot --region asia-northeast3 --format="get(spec.template.spec.containers[0].env)"
   ```

3. **Webhook 상태 확인**:
   ```bash
   curl "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getWebhookInfo"
   ```

### 비용 최적화

기본 설정은 항상 1개 인스턴스가 실행됩니다 (`--min-instances 1`). 비용을 절약하려면:

```bash
# 최소 인스턴스 0으로 설정 (사용하지 않을 때 자동 종료)
gcloud run deploy telegram-bot \
    --region asia-northeast3 \
    --min-instances 0 \
    --max-instances 3
```

**주의**: `min-instances 0`으로 설정하면 첫 메시지 응답이 느릴 수 있습니다 (Cold Start).

## 🔄 업데이트 및 재배포

코드 수정 후 재배포:

```bash
# 간단히 배포 스크립트 재실행
./deploy_google_cloud.sh

# 또는 수동으로
gcloud run deploy telegram-bot --source . --region asia-northeast3
```

## 📊 모니터링

### Cloud Console에서 확인

1. [Google Cloud Console](https://console.cloud.google.com) 접속
2. Cloud Run > telegram-bot 서비스 선택
3. **지표** 탭에서 다음 확인:
   - 요청 수
   - 응답 시간
   - 오류율
   - 메모리 사용량

### 알림 설정 (선택사항)

```bash
# 오류율이 5% 이상일 때 알림
gcloud alpha monitoring policies create \
    --notification-channels=[CHANNEL_ID] \
    --display-name="Telegram Bot Error Rate" \
    --condition-display-name="Error rate > 5%" \
    --condition-threshold-value=5 \
    --condition-threshold-duration=300s
```

## 🛡️ 보안 고려사항

1. **환경변수 보호**: API 키를 코드에 직접 넣지 말고 환경변수 사용
2. **Secret Manager 사용 (고급)**:
   ```bash
   # Secret 생성
   echo -n "$TELEGRAM_BOT_TOKEN" | gcloud secrets create telegram-bot-token --data-file=-

   # Cloud Run에 Secret 마운트
   gcloud run deploy telegram-bot \
       --update-secrets=TELEGRAM_BOT_TOKEN=telegram-bot-token:latest
   ```

## 📝 다음 단계

배포 완료 후:

1. ✅ 로컬 `telegram_daemon.py` 중지
2. ✅ `task_status.json`에 webhook 모드 기록
3. ✅ 텔레그램 명령어로 테스트
4. ✅ 시스템 자동화 작업 모니터링

## 🆘 지원

문제가 발생하면:
1. 로그 확인: `gcloud run logs tail telegram-bot --region asia-northeast3`
2. Webhook 상태: `curl "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getWebhookInfo"`
3. Health Check: `curl [SERVICE_URL]/health`
