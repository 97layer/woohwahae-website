"""
State Watcher - sync_state.json 파일 변경 감시
파일이 변경되면 모든 WebSocket 클라이언트에게 브로드캐스트
"""

import json
import asyncio
from pathlib import Path
from typing import Optional, Dict, Any
from watchfiles import awatch

from execution.api.websocket_manager import WebSocketManager


class StateWatcher:
    """sync_state.json 파일 감시자"""

    def __init__(self, sync_state_path: Path, ws_manager: WebSocketManager):
        self.sync_state_path = sync_state_path
        self.ws_manager = ws_manager
        self.is_running = False
        self.watch_task: Optional[asyncio.Task] = None
        self.last_state: Optional[Dict[str, Any]] = None

    async def start(self):
        """감시 시작"""
        if self.is_running:
            print("⚠️ State watcher already running")
            return

        self.is_running = True

        # 초기 상태 로드
        if self.sync_state_path.exists():
            with open(self.sync_state_path, 'r') as f:
                self.last_state = json.load(f)
                print(f"📖 Initial state loaded: {self.last_state.get('active_node')}")

        # 비동기 감시 태스크 시작
        self.watch_task = asyncio.create_task(self._watch_file())
        print(f"👁️  State watcher started: {self.sync_state_path}")

    async def stop(self):
        """감시 중지"""
        self.is_running = False
        if self.watch_task:
            self.watch_task.cancel()
            try:
                await self.watch_task
            except asyncio.CancelledError:
                pass
        print("🛑 State watcher stopped")

    async def _watch_file(self):
        """파일 변경 감시 루프"""
        watch_dir = self.sync_state_path.parent

        try:
            async for changes in awatch(watch_dir):
                if not self.is_running:
                    break

                # sync_state.json 파일 변경 확인
                for change_type, changed_path in changes:
                    if Path(changed_path) == self.sync_state_path:
                        await self._on_state_changed()

        except asyncio.CancelledError:
            print("State watcher cancelled")
        except Exception as e:
            print(f"❌ State watcher error: {e}")

    async def _on_state_changed(self):
        """상태 파일 변경 시 처리"""
        try:
            # 파일 읽기
            with open(self.sync_state_path, 'r') as f:
                new_state = json.load(f)

            # 변경 사항 확인
            if new_state != self.last_state:
                print(f"🔄 State changed: {new_state.get('active_node')} @ {new_state.get('last_sync')}")

                # 모든 WebSocket 클라이언트에게 브로드캐스트
                await self.ws_manager.broadcast_sync_state(new_state)

                self.last_state = new_state

        except json.JSONDecodeError as e:
            print(f"❌ Invalid JSON in sync_state.json: {e}")
        except Exception as e:
            print(f"❌ Error processing state change: {e}")

    async def force_broadcast(self):
        """현재 상태 강제 브로드캐스트 (테스트용)"""
        if self.last_state:
            await self.ws_manager.broadcast_sync_state(self.last_state)
            print("📡 Force broadcasted current state")
