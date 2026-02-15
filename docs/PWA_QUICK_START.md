# 97layerOS PWA - Quick Start

**1분 만에 시작하기**

---

## 🚀 시작

### 1. Backend 실행 (Terminal 1)
```bash
cd /Users/97layer/97layerOS/execution/api
python3 main.py
```
✅ Server running on `http://localhost:8080`

### 2. Frontend 실행 (Terminal 2)
```bash
cd /Users/97layer/97layerOS/frontend
npm run dev
```
✅ PWA running on `http://localhost:3000`

### 3. 브라우저에서 열기
```
http://localhost:3000
```

---

## ✅ 정상 작동 확인

PWA 화면에서 확인할 항목:
- 🟢 **Green dot**: "Real-time Connected"
- 🖥️ **Active Node**: MacBook
- 📊 **Health**: MacBook (online), GCP VM (unknown)

---

## 🧪 실시간 업데이트 테스트

```bash
# Terminal 3
cd /Users/97layer/97layerOS/knowledge/system
echo '{
  "last_sync": "'$(date -u +"%Y-%m-%dT%H:%M:%S.%6N")'",
  "location": "LOCAL_MAC",
  "pending_changes": [],
  "active_node": "macbook",
  "last_heartbeat": "'$(date -u +"%Y-%m-%dT%H:%M:%S.%6N")'",
  "pending_handover": false,
  "node_history": [],
  "health": {"macbook": "online", "gcp_vm": "unknown"}
}' > sync_state.json
```

**결과**: PWA가 새로고침 없이 즉시 업데이트됨

---

## 📱 휴대폰에서 접속

### ngrok 사용
```bash
brew install ngrok
ngrok http 3000
# 생성된 URL을 휴대폰에서 접속
```

---

## 🛑 종료

Backend/Frontend 터미널에서 `Ctrl+C`

---

## 📚 자세한 문서

- **설치 가이드**: [PWA_LAUNCH_GUIDE.md](PWA_LAUNCH_GUIDE.md)
- **완성 보고서**: [PWA_PHASE1_COMPLETE.md](PWA_PHASE1_COMPLETE.md)
- **API 문서**: [../execution/api/README.md](../execution/api/README.md)

---

**현재 상태**: Phase 1 완성 ✅
**다음 단계**: Phase 2 - Agent Orchestration Chat
