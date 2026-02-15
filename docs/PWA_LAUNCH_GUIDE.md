# 97layerOS PWA Launch Guide

**Phase 1 Complete**: 메신저 뼈대 + 하이브리드 상태 모니터

---

## 🎯 Phase 1 완성 항목

✅ **Backend (FastAPI)**
- Real-time WebSocket server
- State watcher for sync_state.json
- REST API endpoints: `/api/health`, `/api/status`, `/api/agents`

✅ **Frontend (Next.js PWA)**
- Hybrid Health Monitor UI
- Real-time WebSocket client
- 우화해 brand design system (minimal & high-end)

---

## 🚀 Quick Start

### 1. Start Backend (Terminal 1)

```bash
cd /Users/97layer/97layerOS/execution/api
python3 main.py
```

Expected output:
```
🚀 97layerOS PWA Backend Server
📍 Monitoring: sync_state.json
👁️  State watcher started
```

Backend runs on: **http://localhost:8080**

### 2. Start Frontend (Terminal 2)

```bash
cd /Users/97layer/97layerOS/frontend
npm run dev
```

Frontend runs on: **http://localhost:3000**

### 3. Open in Browser

Navigate to: **http://localhost:3000**

You should see:
- ✅ **Green dot**: "Real-time Connected" (WebSocket active)
- 🖥️ **Active Node**: MacBook
- 📊 **Health Status**: MacBook (online), GCP VM (unknown)
- 🕐 **Last Sync/Heartbeat**: timestamps

---

## 🧪 Testing Real-Time Updates

### Test 1: File Watcher

Modify `sync_state.json` and watch the PWA update instantly:

```bash
# In Terminal 3
cd /Users/97layer/97layerOS/knowledge/system
echo '{
  "last_sync": "'$(date -u +"%Y-%m-%dT%H:%M:%S.%6N")'",
  "location": "LOCAL_MAC",
  "pending_changes": [],
  "active_node": "macbook",
  "last_heartbeat": "'$(date -u +"%Y-%m-%dT%H:%M:%S.%6N")'",
  "pending_handover": false,
  "node_history": [],
  "health": {
    "macbook": "online",
    "gcp_vm": "unknown"
  }
}' > sync_state.json
```

**Expected**: PWA updates "Last Sync" timestamp in real-time (no refresh needed)

### Test 2: Simulate GCP Takeover

```bash
# Simulate MacBook going offline, GCP taking over
echo '{
  "last_sync": "'$(date -u +"%Y-%m-%dT%H:%M:%S.%6N")'",
  "location": "GCP_VM",
  "pending_changes": [],
  "active_node": "gcp_vm",
  "last_heartbeat": "'$(date -u +"%Y-%m-%dT%H:%M:%S.%6N")'",
  "pending_handover": false,
  "node_history": [{"from": "macbook", "to": "gcp_vm", "timestamp": "'$(date -u +"%Y-%m-%dT%H:%M:%S.%6N")'", "reason": "timeout"}],
  "health": {
    "macbook": "offline",
    "gcp_vm": "online"
  }
}' > sync_state.json
```

**Expected**:
- Active Node changes to "☁️ GCP VM"
- MacBook status becomes red (offline)
- GCP VM status becomes green (online)
- Gold "● ACTIVE" label moves to GCP VM

### Test 3: REST API Health Check

```bash
curl http://localhost:8080/api/health | jq
```

Expected response:
```json
{
  "status": "healthy",
  "active_node": "macbook",
  "last_heartbeat": "2026-02-15T16:52:50.000000",
  "health": {
    "macbook": "online",
    "gcp_vm": "unknown"
  },
  "connected_clients": 1
}
```

### Test 4: WebSocket Connection Test

```bash
# Install websocat: brew install websocat
websocat ws://localhost:8080/ws
```

After connecting, send:
```json
{"type": "get_status"}
```

You should receive a `sync_state_update` message with current state.

---

## 📱 Mobile Testing (Phase 1+)

### Option A: ngrok (Recommended for testing)

```bash
# Install ngrok: brew install ngrok
ngrok http 3000
```

You'll get a URL like `https://abc123.ngrok.io` - open this on your phone!

### Option B: Local Network (Same WiFi)

1. Find your MacBook's IP:
```bash
ipconfig getifaddr en0  # Usually something like 192.168.1.xxx
```

2. Update frontend WebSocket URL:
```typescript
// frontend/lib/websocket.ts
constructor(url: string = 'ws://192.168.1.xxx:8080/ws') {
```

3. Open on phone: `http://192.168.1.xxx:3000`

---

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────┐
│  PWA Frontend (Next.js)                         │
│  - HealthMonitor component                      │
│  - WebSocket client                             │
│  - Tailwind CSS (우화해 brand)                  │
└─────────────────┬───────────────────────────────┘
                  │ WebSocket (ws://localhost:8080/ws)
                  │ + REST API (/api/*)
┌─────────────────▼───────────────────────────────┐
│  FastAPI Backend                                │
│  ├─ WebSocket Manager (connection handling)    │
│  ├─ State Watcher (monitors sync_state.json)   │
│  └─ REST API (status, health, agents)          │
└─────────────────┬───────────────────────────────┘
                  │ File Watcher
┌─────────────────▼───────────────────────────────┐
│  knowledge/system/sync_state.json               │
│  - Updated by hybrid_sync.py                    │
│  - Tracks MacBook ↔ GCP state                   │
└─────────────────────────────────────────────────┘
```

---

## 🎨 Design System (우화해 Brand)

**Colors**:
- Primary: `#0A0A0A` (brand-black)
- Background: `#FAFAFA` (brand-white)
- Accent: `#D4AF37` (brand-gold)
- Status indicators: Green (online), Red (offline), Gray (unknown)

**Typography**: Inter font family (minimal, clean)

**Aesthetic**: High-end, minimal, professional - reflecting the 우화해 brand identity

---

## 🔮 Next Steps (Phase 2)

### Agent Orchestration Chat

**Backend additions**:
- `POST /api/chat` - Send messages to agent_router
- `GET /api/chat/history` - Retrieve conversation history
- WebSocket events: `agent:thinking`, `agent:response`, `agent:error`

**Frontend additions**:
- `AgentChat.tsx` - Telegram-style chat interface
- `AgentStatus.tsx` - Show which agent is currently processing
- `ThoughtExpander.tsx` - Expandable view of agent internal discussion

**Integration**:
- Connect to existing `agent_router` from `async_telegram_daemon.py`
- Use `libs/memory_manager.py` for chat history
- Stream agent thoughts from `council_log/*.md`

### Implementation Time: 4-5 hours

---

## 📊 Performance Notes

- **WebSocket reconnection**: Automatic with 3s interval
- **Heartbeat**: Ping every 30s to keep connection alive
- **File watcher**: Uses `watchfiles` (Rust-based, very fast)
- **State updates**: Broadcast to all clients instantly (<50ms)

---

## 🐛 Troubleshooting

### Issue: "WebSocket disconnected"

**Check**:
1. Backend is running: `curl http://localhost:8080/api/health`
2. Port 8080 not blocked by firewall
3. Browser console for error details

### Issue: "No updates when modifying sync_state.json"

**Check**:
1. File path is correct: `knowledge/system/sync_state.json`
2. JSON is valid (use `jq` to validate)
3. Backend logs show: "🔄 State changed"
4. At least one WebSocket client is connected

### Issue: Frontend won't start

**Check**:
1. Node.js version >=18: `node -v`
2. Dependencies installed: `npm install`
3. Port 3000 not in use: `lsof -i :3000`

---

## 🎉 Success Criteria

**Phase 1 is successful when**:
- ✅ PWA loads without errors
- ✅ Green "Real-time Connected" dot appears
- ✅ MacBook/GCP health status displays correctly
- ✅ Modifying sync_state.json triggers instant UI update
- ✅ REST API endpoints respond correctly
- ✅ Design matches 우화해 brand aesthetic (minimal & high-end)

---

## 📝 File Locations

**Backend**:
- `/Users/97layer/97layerOS/execution/api/main.py`
- `/Users/97layer/97layerOS/execution/api/websocket_manager.py`
- `/Users/97layer/97layerOS/execution/api/state_watcher.py`

**Frontend**:
- `/Users/97layer/97layerOS/frontend/components/HealthMonitor.tsx`
- `/Users/97layer/97layerOS/frontend/lib/websocket.ts`
- `/Users/97layer/97layerOS/frontend/app/page.tsx`

**State File**:
- `/Users/97layer/97layerOS/knowledge/system/sync_state.json`

---

## 🚢 Deployment (Phase 4)

### Local Podman (Current)
Already running! Use for development and testing.

### GCP Cloud Run (Future)
- Containerize FastAPI + Next.js build
- Deploy to Cloud Run with auto-scaling
- Update WebSocket URL to cloud endpoint
- Enable OAuth2 authentication (97layer@gmail.com only)

---

**Phase 1 완성을 축하합니다!** 🎊

이제 커트하는 도중 태블릿에서 "지능의 가시성"을 실시간으로 볼 수 있습니다.
