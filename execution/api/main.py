#!/usr/bin/env python3
"""
97layerOS PWA Backend - FastAPI Server
실시간 에이전트 오케스트레이션 API

Phase 1: 하이브리드 상태 모니터링 + WebSocket 실시간 스트리밍
"""

import json
import sys
from pathlib import Path
from datetime import datetime
from typing import Dict, Any, Optional
from contextlib import asynccontextmanager

from fastapi import FastAPI, WebSocket, WebSocketDisconnect, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

# 프로젝트 루트
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.append(str(PROJECT_ROOT))

# 내부 모듈
from execution.api.websocket_manager import WebSocketManager
from execution.api.state_watcher import StateWatcher
from execution.api.chat_handler import ChatHandler

# 경로 설정
STATUS_FILE = PROJECT_ROOT / "knowledge" / "system_state.json"
SYNC_STATE_FILE = PROJECT_ROOT / "knowledge" / "system" / "sync_state.json"
SYNAPSE_FILE = PROJECT_ROOT / "knowledge" / "agent_hub" / "synapse_bridge.json"
COUNCIL_DIR = PROJECT_ROOT / "knowledge" / "council_log"

# 글로벌 매니저
ws_manager = WebSocketManager()
state_watcher: Optional[StateWatcher] = None
chat_handler: Optional[ChatHandler] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """앱 생명주기 관리"""
    global state_watcher, chat_handler

    # 시작: State Watcher 초기화
    state_watcher = StateWatcher(
        sync_state_path=SYNC_STATE_FILE,
        ws_manager=ws_manager
    )
    await state_watcher.start()

    # 시작: Chat Handler 초기화
    chat_handler = ChatHandler(ws_manager=ws_manager)

    print("🚀 97layerOS PWA Backend started")
    print(f"📍 Watching: {SYNC_STATE_FILE}")
    print("💬 Chat handler ready")

    yield

    # 종료: 정리
    await state_watcher.stop()
    print("🛑 97layerOS PWA Backend stopped")


# FastAPI 앱 초기화
app = FastAPI(
    title="97layerOS PWA API",
    description="Real-time agent orchestration backend",
    version="1.0.0",
    lifespan=lifespan
)

# CORS 설정 (개발 단계: 모든 origin 허용)
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # TODO: Phase 4에서 특정 도메인으로 제한
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# ==========================================
# REST API 엔드포인트
# ==========================================

@app.get("/")
async def root():
    """헬스체크"""
    return {
        "service": "97layerOS PWA Backend",
        "status": "operational",
        "timestamp": datetime.now().isoformat()
    }


@app.get("/api/health")
async def health_check():
    """시스템 헬스체크"""
    try:
        # sync_state.json 읽기
        if not SYNC_STATE_FILE.exists():
            return JSONResponse(
                status_code=503,
                content={"status": "unhealthy", "reason": "sync_state.json not found"}
            )

        with open(SYNC_STATE_FILE, 'r') as f:
            sync_state = json.load(f)

        # 마지막 heartbeat 확인
        last_heartbeat_str = sync_state.get("last_heartbeat")
        if last_heartbeat_str:
            last_heartbeat = datetime.fromisoformat(last_heartbeat_str)
            delta = (datetime.now() - last_heartbeat).total_seconds()
            is_alive = delta < 300  # 5분 이내
        else:
            is_alive = False

        return {
            "status": "healthy" if is_alive else "stale",
            "active_node": sync_state.get("active_node"),
            "last_heartbeat": last_heartbeat_str,
            "health": sync_state.get("health", {}),
            "connected_clients": len(ws_manager.active_connections)
        }

    except Exception as e:
        return JSONResponse(
            status_code=500,
            content={"status": "error", "error": str(e)}
        )


@app.get("/api/status")
async def get_status():
    """하이브리드 시스템 상태 조회"""
    try:
        status_data = {"status": "INITIALIZING", "agents": {}}

        # 1. System State 읽기
        if STATUS_FILE.exists():
            with open(STATUS_FILE, 'r', encoding='utf-8') as f:
                status_data = json.load(f)

        # 2. Sync State 병합 (하이브리드 정보)
        if SYNC_STATE_FILE.exists():
            with open(SYNC_STATE_FILE, 'r', encoding='utf-8') as f:
                sync_state = json.load(f)
                status_data["hybrid"] = {
                    "active_node": sync_state.get("active_node"),
                    "location": sync_state.get("location"),
                    "last_sync": sync_state.get("last_sync"),
                    "health": sync_state.get("health", {})
                }

        # 3. Synapse Bridge 병합 (에이전트 협업 정보)
        if SYNAPSE_FILE.exists():
            with open(SYNAPSE_FILE, 'r', encoding='utf-8') as f:
                bridge_data = json.load(f)

                if "active_agents" in bridge_data:
                    for agent, info in bridge_data["active_agents"].items():
                        status_data["agents"][agent] = info

                status_data["parallel_mode"] = bridge_data.get("collaboration_mode") == "Parallel"
                status_data["performance"] = bridge_data.get("performance", {})

        # 4. Heartbeat 체크
        last_update_str = status_data.get("last_update")
        if last_update_str:
            try:
                last_ts = datetime.fromisoformat(last_update_str)
                delta = (datetime.now() - last_ts).total_seconds()
                status_data["is_alive"] = delta < 300
            except:
                status_data["is_alive"] = False

        return status_data

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/api/agents")
async def get_agents():
    """활성 에이전트 목록"""
    try:
        agents = []

        if SYNAPSE_FILE.exists():
            with open(SYNAPSE_FILE, 'r', encoding='utf-8') as f:
                bridge_data = json.load(f)

                for agent_name, agent_info in bridge_data.get("active_agents", {}).items():
                    agents.append({
                        "name": agent_name,
                        "status": agent_info.get("status"),
                        "task": agent_info.get("current_task"),
                        "last_update": agent_info.get("last_update")
                    })

        return {"agents": agents, "count": len(agents)}

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


# ==========================================
# Chat API 엔드포인트 (Phase 2)
# ==========================================

@app.post("/api/chat")
async def send_chat_message(request: Dict[str, Any]):
    """
    채팅 메시지 전송

    Body:
        {
            "user_id": "string",  # PWA 세션 ID 또는 user ID
            "message": "string",  # 사용자 메시지
            "images": []          # Optional: 이미지 URL 리스트 (Phase 3)
        }
    """
    try:
        user_id = request.get("user_id", "pwa_user")
        message = request.get("message")
        images = request.get("images")

        if not message:
            raise HTTPException(status_code=400, detail="Message is required")

        if not chat_handler:
            raise HTTPException(status_code=503, detail="Chat handler not initialized")

        # 비동기 메시지 처리
        result = await chat_handler.process_message(
            user_id=user_id,
            message=message,
            images=images
        )

        return result

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/api/chat/history/{user_id}")
async def get_chat_history(user_id: str, limit: int = 50):
    """
    대화 기록 조회

    Args:
        user_id: 사용자 ID
        limit: 최대 메시지 수 (default: 50)
    """
    try:
        if not chat_handler:
            raise HTTPException(status_code=503, detail="Chat handler not initialized")

        history = await chat_handler.get_chat_history(user_id, limit)

        return {
            "user_id": user_id,
            "messages": history,
            "count": len(history)
        }

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


# ==========================================
# WebSocket 엔드포인트
# ==========================================

@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket):
    """실시간 WebSocket 연결"""
    await ws_manager.connect(websocket)

    try:
        # 연결 직후 현재 상태 전송
        if SYNC_STATE_FILE.exists():
            with open(SYNC_STATE_FILE, 'r') as f:
                sync_state = json.load(f)
                await websocket.send_json({
                    "type": "sync_state_update",
                    "data": sync_state
                })

        # 클라이언트 메시지 수신 루프
        while True:
            data = await websocket.receive_text()
            message = json.loads(data)

            # 메시지 타입에 따라 처리
            msg_type = message.get("type")

            if msg_type == "ping":
                await websocket.send_json({"type": "pong"})

            elif msg_type == "get_status":
                # 현재 상태 요청
                if SYNC_STATE_FILE.exists():
                    with open(SYNC_STATE_FILE, 'r') as f:
                        sync_state = json.load(f)
                        await websocket.send_json({
                            "type": "sync_state_update",
                            "data": sync_state
                        })

            elif msg_type == "chat":
                # Phase 2: 채팅 메시지 처리
                user_id = message.get("user_id", "pwa_user")
                user_message = message.get("message")
                images = message.get("images")

                if user_message and chat_handler:
                    # 비동기 처리 (결과는 WebSocket broadcast로 전송됨)
                    asyncio.create_task(
                        chat_handler.process_message(user_id, user_message, images)
                    )

    except WebSocketDisconnect:
        ws_manager.disconnect(websocket)
    except Exception as e:
        print(f"WebSocket error: {e}")
        ws_manager.disconnect(websocket)


if __name__ == "__main__":
    import uvicorn

    print("=" * 60)
    print("🚀 97layerOS PWA Backend Server")
    print("=" * 60)
    print(f"📍 Project Root: {PROJECT_ROOT}")
    print(f"🔍 Monitoring: {SYNC_STATE_FILE}")
    print("=" * 60)

    uvicorn.run(
        app,
        host="0.0.0.0",
        port=8080,
        log_level="info"
    )
