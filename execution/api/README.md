# 97layerOS PWA Backend API

FastAPI backend for real-time agent orchestration and hybrid system monitoring.

## Installation

```bash
cd execution/api
pip install -r requirements.txt
```

## Running the Server

```bash
# Development mode (auto-reload)
python main.py

# Or with uvicorn directly
uvicorn main:app --reload --host 0.0.0.0 --port 8080
```

## API Endpoints

### REST API

- `GET /` - Health check
- `GET /api/health` - System health status
- `GET /api/status` - Full hybrid system status
- `GET /api/agents` - Active agent list

### WebSocket

- `WS /ws` - Real-time bidirectional connection
  - Receives: `sync_state_update`, `agent_update`, `agent_thought`
  - Sends: `ping`, `get_status`, `chat` (Phase 2)

## Testing

### Test REST API
```bash
curl http://localhost:8080/api/health
curl http://localhost:8080/api/status
```

### Test WebSocket
```bash
# Install websocat: brew install websocat
websocat ws://localhost:8080/ws
# Then send: {"type": "get_status"}
```

### Test File Watcher
```bash
# In another terminal, modify sync_state.json:
echo '{"active_node": "macbook", "health": {"macbook": "online", "gcp_vm": "unknown"}}' > ../../knowledge/system/sync_state.json
# Watch the WebSocket broadcast in the server logs
```

## Architecture

```
FastAPI Server (main.py)
├── WebSocket Manager (websocket_manager.py)
│   └── Manages client connections + broadcasts
├── State Watcher (state_watcher.py)
│   └── Watches sync_state.json for changes
└── REST API Endpoints
    └── /api/status, /api/health, /api/agents
```

## Phase Roadmap

- ✅ **Phase 1**: WebSocket + Hybrid Health Monitor
- 🔜 **Phase 2**: Agent orchestration chat
- 🔜 **Phase 3**: Asset gallery + file browser
- 🔜 **Phase 4**: OAuth2 + GCP Cloud Run deployment
