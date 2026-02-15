"""
WebSocket Connection Manager
실시간 클라이언트 연결 관리 및 브로드캐스트
"""

import json
from typing import List, Dict, Any
from fastapi import WebSocket
import asyncio


class WebSocketManager:
    """WebSocket 연결 관리자"""

    def __init__(self):
        self.active_connections: List[WebSocket] = []

    async def connect(self, websocket: WebSocket):
        """새 클라이언트 연결"""
        await websocket.accept()
        self.active_connections.append(websocket)
        print(f"✅ WebSocket connected. Total clients: {len(self.active_connections)}")

    def disconnect(self, websocket: WebSocket):
        """클라이언트 연결 해제"""
        if websocket in self.active_connections:
            self.active_connections.remove(websocket)
            print(f"❌ WebSocket disconnected. Total clients: {len(self.active_connections)}")

    async def send_personal_message(self, message: Dict[str, Any], websocket: WebSocket):
        """특정 클라이언트에게 메시지 전송"""
        try:
            await websocket.send_json(message)
        except Exception as e:
            print(f"Error sending personal message: {e}")
            self.disconnect(websocket)

    async def broadcast(self, message: Dict[str, Any]):
        """모든 연결된 클라이언트에게 브로드캐스트"""
        disconnected = []

        for connection in self.active_connections:
            try:
                await connection.send_json(message)
            except Exception as e:
                print(f"Error broadcasting to client: {e}")
                disconnected.append(connection)

        # 실패한 연결 제거
        for conn in disconnected:
            self.disconnect(conn)

    async def broadcast_sync_state(self, state_data: Dict[str, Any]):
        """sync_state 업데이트 브로드캐스트"""
        message = {
            "type": "sync_state_update",
            "data": state_data,
            "timestamp": state_data.get("last_sync")
        }
        await self.broadcast(message)

    async def broadcast_agent_update(self, agent_name: str, agent_data: Dict[str, Any]):
        """에이전트 상태 업데이트 브로드캐스트"""
        message = {
            "type": "agent_update",
            "agent": agent_name,
            "data": agent_data
        }
        await self.broadcast(message)

    async def broadcast_thought_stream(self, agent_name: str, thought: str):
        """에이전트 사고 과정 스트리밍 (Phase 2)"""
        message = {
            "type": "agent_thought",
            "agent": agent_name,
            "thought": thought
        }
        await self.broadcast(message)

    def get_connection_count(self) -> int:
        """현재 연결된 클라이언트 수"""
        return len(self.active_connections)
