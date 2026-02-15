# 🚀 토큰 재발급 빠른 시작 (5분)

**이 문서는 토큰 재발급을 5분 안에 완료하는 초간단 가이드입니다.**

---

## 📋 준비물 확인

- [ ] Telegram 계정 (BotFather 접근)
- [ ] Google 계정 (AI Studio 접근)
- [ ] Anthropic 계정 (Console 접근)
- [ ] 터미널 열기

---

## ⚡ 3단계 재발급

### 1️⃣ 토큰 재발급 (3분)

**Telegram (1분)**:
```
1. Telegram 앱에서 @BotFather 검색
2. /mybots → 97LayerOSwoohwahae 선택
3. API Token → Revoke current token
4. 새 토큰 복사
```

**Gemini (1분)**:
```
1. https://aistudio.google.com/app/apikey 접속
2. 기존 키 삭제
3. Create API Key 클릭
4. 새 키 복사
```

**Anthropic (1분)**:
```
1. https://console.anthropic.com/settings/keys 접속
2. 기존 키 Revoke
3. Create Key 클릭
4. 새 키 복사
```

---

### 2️⃣ 자동 업데이트 실행 (2분)

터미널에서 아래 명령어 실행:

```bash
cd /Users/97layer/97layerOS
./execution/system/update_tokens.sh
```

**프롬프트 나오면**:
1. Telegram 토큰 붙여넣기 (Enter)
2. Gemini 키 붙여넣기 (Enter)
3. Anthropic 키 붙여넣기 (Enter)

스크립트가 자동으로:
- ✅ .env 업데이트
- ✅ Cloud Run 업데이트
- ✅ VM 업데이트
- ✅ Webhook 재등록
- ✅ 검증 완료

---

### 3️⃣ 검증 (30초)

Telegram 앱에서:
```
1. @97LayerOSwoohwahae 검색
2. /start 입력
3. "안녕" 입력
4. AI 응답 확인 ✅
```

---

## 🎉 완료!

**소요 시간**: 5분
**다운타임**: 2분 (재시작 동안)

모든 토큰이 새로 발급되어 안전합니다.

---

## 🔍 문제 해결

### Q: 스크립트 실행 안됨
```bash
chmod +x /Users/97layer/97layerOS/execution/system/update_tokens.sh
```

### Q: Cloud Run 업데이트 실패
```bash
# 수동 배포
cd /Users/97layer/97layerOS/deployment
./deploy_google_cloud.sh
```

### Q: 봇 무응답
```bash
# Webhook 수동 재등록
source /Users/97layer/97layerOS/.env
curl -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/setWebhook?url=https://telegram-bot-514569077225.asia-northeast3.run.app/webhook"
```

---

**상세 가이드**: [TOKEN_REVOCATION_GUIDE.md](TOKEN_REVOCATION_GUIDE.md)
