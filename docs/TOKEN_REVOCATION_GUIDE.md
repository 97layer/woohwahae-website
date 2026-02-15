# API 토큰 즉시 재발급 가이드

**날짜**: 2026-02-15
**상태**: 🔴 **URGENT - 즉시 실행 필요**
**예상 소요 시간**: 10분
**다운타임**: 5분 (재시작 동안)

---

## 🚨 왜 재발급해야 하나요?

1. **GitHub Public Repo**: 모든 토큰이 인터넷에 노출됨
2. **Git 히스토리**: 3개 커밋에 토큰 포함
3. **악용 가능성**: 누구나 97layerOS 봇 사용 가능

---

## 📋 재발급 순서 (단계별)

### 1단계: Telegram Bot Token 재발급 (2분)

#### 방법 A: 기존 봇 토큰 갱신 (권장)
```
1. Telegram에서 @BotFather 검색
2. /mybots 입력
3. "97LayerOSwoohwahae" 선택
4. "API Token" 선택
5. "Revoke current token" 클릭
6. 새 토큰 복사 (형식: 1234567890:ABCdef...)
```

#### 방법 B: 새 봇 생성
```
1. @BotFather에서 /newbot
2. 봇 이름: 97LayerOS v2
3. 사용자명: 97layeros_v2_bot
4. 토큰 복사
```

**복사한 토큰 임시 저장**:
```
NEW_TELEGRAM_TOKEN=여기에_붙여넣기
```

---

### 2단계: Gemini API Key 재발급 (3분)

```
1. https://aistudio.google.com/app/apikey 접속
2. 기존 키 "AIzaSyCGgHVPjEEI3OI3tSNW3SSHNbZuYpHrH-g" 찾기
3. "Delete" 클릭
4. "Create API Key" 클릭
5. 프로젝트 선택: Default Gemini Project (또는 신규 생성)
6. 키 복사
```

**복사한 키 임시 저장**:
```
NEW_GEMINI_KEY=여기에_붙여넣기
```

---

### 3단계: Anthropic API Key 재발급 (3분)

```
1. https://console.anthropic.com/settings/keys 접속
2. 기존 키 "sk-ant-api03-PKAkuoznR_..." 찾기
3. "Revoke" 클릭
4. "Create Key" 클릭
5. 이름: 97layerOS Production
6. 키 복사
```

**복사한 키 임시 저장**:
```
NEW_ANTHROPIC_KEY=여기에_붙여넣기
```

---

### 4단계: .env 파일 업데이트 (1분)

**터미널에서 실행**:
```bash
cd /Users/97layer/97layerOS

# .env 백업
cp .env .env.backup

# 새 키로 교체 (아래 명령어를 복사하고, 키를 실제 값으로 교체)
cat > .env << 'EOF'
TELEGRAM_BOT_TOKEN=여기에_1단계_토큰_붙여넣기
GEMINI_API_KEY=여기에_2단계_키_붙여넣기
ANTHROPIC_API_KEY=여기에_3단계_키_붙여넣기
EOF

# 확인 (키가 제대로 들어갔는지)
cat .env
```

**⚠️ 주의**: 실제 키 값으로 교체하세요!

---

### 5단계: Cloud Run 환경변수 업데이트 (2분)

```bash
cd /Users/97layer/97layerOS

# 환경변수 로드
source .env

# Cloud Run 업데이트
gcloud run services update telegram-bot \
  --region=asia-northeast3 \
  --set-env-vars "TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN,GEMINI_API_KEY=$GEMINI_API_KEY,ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY"
```

---

### 6단계: VM 환경변수 업데이트 (3분)

```bash
# VM에 .env 동기화
gcloud compute scp .env layer97-nightguard:~/97layerOS/.env --zone=us-west1-b

# VM에서 Night Guard 재시작
gcloud compute ssh layer97-nightguard --zone=us-west1-b --command="sudo systemctl restart 97layeros-nightguard"

# 상태 확인
gcloud compute ssh layer97-nightguard --zone=us-west1-b --command="sudo systemctl status 97layeros-nightguard | head -15"
```

---

### 7단계: Telegram Webhook 재등록 (1분)

```bash
# 새 토큰으로 webhook 재등록
source .env

curl -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/setWebhook?url=https://telegram-bot-514569077225.asia-northeast3.run.app/webhook"

# 확인
curl "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getWebhookInfo" | jq
```

**기대 결과**:
```json
{
  "ok": true,
  "result": {
    "url": "https://telegram-bot-514569077225.asia-northeast3.run.app/webhook",
    "has_custom_certificate": false,
    "pending_update_count": 0
  }
}
```

---

## ✅ 검증 체크리스트

### 1. 기존 토큰 무효화 확인
```bash
# 기존 토큰으로 API 호출 (실패해야 정상)
curl "https://api.telegram.org/bot8501568801:AAE-3fBl-p6uZcmrdsWSRQuz_eg8yDADwjI/getMe"

# 기대 결과: {"ok":false,"error_code":401,"description":"Unauthorized"}
```

### 2. 새 토큰 작동 확인
```bash
# 새 토큰으로 API 호출 (성공해야 정상)
source .env
curl "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getMe"

# 기대 결과: {"ok":true,"result":{"id":...,"username":"97LayerOSwoohwahae"}}
```

### 3. Cloud Run 확인
```bash
curl https://telegram-bot-514569077225.asia-northeast3.run.app/health
```

### 4. VM Night Guard 확인
```bash
gcloud compute ssh layer97-nightguard --zone=us-west1-b --command="sudo journalctl -u 97layeros-nightguard -n 20"
```

### 5. Telegram 봇 테스트
```
1. Telegram 앱에서 @97LayerOSwoohwahae 검색
2. /start 입력
3. "97LAYER OS Online" 응답 확인
4. "안녕" 입력
5. AI 응답 확인
```

---

## 🔒 보안 강화 (추가 조치)

### 옵션 A: Git 히스토리 정리

**⚠️ 주의**: 협업자가 있다면 협의 필요

```bash
cd /Users/97layer/97layerOS

# 백업
git branch backup-before-cleanup

# Git filter-repo 설치
pip install git-filter-repo

# 토큰 제거
git filter-repo --invert-paths \
  --path test_bot.py \
  --path simple_test_bot.py \
  --path execution/five_agent_hub_integrated.py \
  --force

# 강제 푸시
git push --force-with-lease
```

### 옵션 B: Repository 재생성 (권장)

```bash
# 1. GitHub에서 새 private repo 생성: 97layerOS-secure

# 2. 현재 코드만 깨끗하게 커밋
cd /Users/97layer/97layerOS
rm -rf .git
git init
git add .
git commit -m "Initial commit - Clean security"

# 3. 새 repo에 푸시
git remote add origin https://github.com/97layer/97layerOS-secure.git
git push -u origin main

# 4. 기존 public repo 삭제 (GitHub 웹에서)
```

---

## 📞 문제 해결

### Q: Cloud Run 업데이트 실패
```bash
# 권한 확인
gcloud projects get-iam-policy layer97os

# 재배포
cd /Users/97layer/97layerOS/deployment
./deploy_google_cloud.sh
```

### Q: VM Night Guard 재시작 실패
```bash
# 로그 확인
gcloud compute ssh layer97-nightguard --zone=us-west1-b --command="sudo journalctl -u 97layeros-nightguard -n 50"

# .env 확인
gcloud compute ssh layer97-nightguard --zone=us-west1-b --command="cat ~/97layerOS/.env"

# 수동 재시작
gcloud compute ssh layer97-nightguard --zone=us-west1-b --command="sudo systemctl stop 97layeros-nightguard && sleep 2 && sudo systemctl start 97layeros-nightguard"
```

### Q: Telegram 봇 무응답
```bash
# Webhook 상태 확인
source .env
curl "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getWebhookInfo" | jq

# Webhook 재설정
curl -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/setWebhook?url=https://telegram-bot-514569077225.asia-northeast3.run.app/webhook"

# Cloud Run 로그 확인
gcloud run logs read telegram-bot --region=asia-northeast3 --limit=50
```

---

## 🎯 완료 확인

모든 단계 완료 후:

```bash
# 최종 검증 스크립트
cd /Users/97layer/97layerOS

cat > /tmp/verify_security.sh << 'EOFVERIFY'
#!/bin/bash
echo "🔒 보안 검증 시작..."
echo ""

# 1. 하드코딩 확인
echo "1️⃣ 하드코딩 토큰 확인..."
HARDCODED=$(grep -r "8501568801" . --exclude-dir=.git --exclude="*.md" 2>/dev/null | wc -l)
if [ "$HARDCODED" -eq 0 ]; then
    echo "   ✅ 하드코딩 토큰 없음"
else
    echo "   ❌ 하드코딩 토큰 발견: $HARDCODED개"
fi

# 2. .env 퍼미션
echo ""
echo "2️⃣ .env 파일 퍼미션..."
PERM=$(stat -f "%Lp" .env)
if [ "$PERM" = "600" ]; then
    echo "   ✅ 퍼미션 안전 (600)"
else
    echo "   ⚠️  퍼미션: $PERM (600 권장)"
    chmod 600 .env
    echo "   ✅ 퍼미션 수정 완료"
fi

# 3. .gitignore 확인
echo ""
echo "3️⃣ .gitignore 보호..."
if grep -q "^\.env$" .gitignore && grep -q "^config\.json$" .gitignore; then
    echo "   ✅ .env, config.json 보호됨"
else
    echo "   ❌ .gitignore 미흡"
fi

# 4. Telegram API 테스트
echo ""
echo "4️⃣ Telegram API 테스트..."
source .env
RESPONSE=$(curl -s "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getMe")
if echo "$RESPONSE" | grep -q "\"ok\":true"; then
    echo "   ✅ Telegram API 정상"
else
    echo "   ❌ Telegram API 실패"
fi

# 5. Cloud Run 헬스체크
echo ""
echo "5️⃣ Cloud Run 헬스체크..."
HEALTH=$(curl -s https://telegram-bot-514569077225.asia-northeast3.run.app/health)
if echo "$HEALTH" | grep -q "healthy"; then
    echo "   ✅ Cloud Run 정상"
else
    echo "   ❌ Cloud Run 오류"
fi

echo ""
echo "✅ 보안 검증 완료!"
EOFVERIFY

bash /tmp/verify_security.sh
```

---

## 📊 재발급 완료 보고

모든 단계 완료 후 아래 정보를 기록하세요:

```
✅ 재발급 완료 시각: ___________
✅ 새 Telegram 봇: @___________
✅ Cloud Run 업데이트: ✅
✅ VM 업데이트: ✅
✅ Webhook 재등록: ✅
✅ 검증 완료: ✅
```

---

**다음 액션**: 이 가이드를 따라 토큰을 즉시 재발급하세요!
