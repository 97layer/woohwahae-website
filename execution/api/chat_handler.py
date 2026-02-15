"""
Chat Handler - 에이전트 오케스트레이션 채팅 핸들러
사용자 메시지를 agent_router를 통해 처리하고 결과를 WebSocket으로 스트리밍
"""

import json
import sys
from pathlib import Path
from typing import Dict, Any, Optional
from datetime import datetime
import asyncio

# 프로젝트 루트
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.append(str(PROJECT_ROOT))

from libs.agent_router import AgentRouter, AGENT_REGISTRY
from libs.memory_manager import MemoryManager
from libs.ai_engine import AIEngine
from execution.api.websocket_manager import WebSocketManager


class ChatHandler:
    """채팅 메시지 처리 및 에이전트 라우팅"""

    def __init__(self, ws_manager: WebSocketManager):
        self.ws_manager = ws_manager

        # 코어 컴포넌트 초기화
        try:
            from libs.core_config import AI_MODEL_CONFIG
            self.ai = AIEngine(AI_MODEL_CONFIG)
        except Exception as e:
            print(f"⚠️ AIEngine initialization failed: {e}")
            self.ai = None

        self.memory = MemoryManager(str(PROJECT_ROOT))
        self.agent_router = AgentRouter(self.ai)

        print("✅ Chat handler initialized")

    async def process_message(
        self,
        user_id: str,
        message: str,
        images: Optional[list] = None
    ) -> Dict[str, Any]:
        """
        사용자 메시지 처리

        Args:
            user_id: 사용자 ID (텔레그램 chat_id 또는 PWA 세션 ID)
            message: 사용자 메시지
            images: 첨부 이미지 (Phase 3)

        Returns:
            응답 메시지 및 메타데이터
        """
        try:
            # 1. 메시지 저장
            self.memory.save_chat(user_id, message, role="user")

            # 2. WebSocket으로 "에이전트 사고 중" 알림
            await self.ws_manager.broadcast({
                "type": "agent_thinking",
                "message": "메시지를 분석하고 적절한 에이전트를 선택 중..."
            })

            # 3. 에이전트 라우팅
            agent_key = self.agent_router.route(message)
            agent_name = AGENT_REGISTRY.get(agent_key, {}).get("name", agent_key)

            # 4. 선택된 에이전트 알림
            await self.ws_manager.broadcast({
                "type": "agent_selected",
                "agent": agent_key,
                "agent_name": agent_name
            })

            # 5. 대화 기록 로드
            history = self.memory.load_chat(user_id, limit=10)

            # 6. 에이전트 응답 생성
            if self.ai:
                # 실제 AI 호출
                response = await self._generate_response(
                    agent_key=agent_key,
                    message=message,
                    history=history,
                    images=images
                )
            else:
                # AI 없을 경우 mock 응답
                response = f"[{agent_name}] 메시지를 수신했습니다: {message[:50]}..."

            # 7. 응답 저장
            self.memory.save_chat(user_id, response, role="assistant")

            # 8. WebSocket으로 최종 응답 전송
            await self.ws_manager.broadcast({
                "type": "agent_response",
                "agent": agent_key,
                "agent_name": agent_name,
                "message": response,
                "timestamp": datetime.now().isoformat()
            })

            return {
                "success": True,
                "agent": agent_key,
                "agent_name": agent_name,
                "response": response
            }

        except Exception as e:
            error_msg = f"Error processing message: {str(e)}"
            print(f"❌ {error_msg}")

            await self.ws_manager.broadcast({
                "type": "agent_error",
                "error": error_msg
            })

            return {
                "success": False,
                "error": error_msg
            }

    async def _generate_response(
        self,
        agent_key: str,
        message: str,
        history: list,
        images: Optional[list] = None
    ) -> str:
        """
        AI 엔진을 통해 에이전트 응답 생성
        """
        try:
            # 에이전트 페르소나 + 최근 대화 기록 결합
            persona = self.agent_router.personas.get(agent_key, "")

            # 프롬프트 구성
            context_prompt = ""

            # 최근 대화 기록 추가
            if history:
                context_prompt += "최근 대화 내용:\n"
                for msg in history[-3:]:  # 최근 3개만
                    role = "사용자" if msg["role"] == "user" else "에이전트"
                    context_prompt += f"{role}: {msg['content'][:300]}\n"
                context_prompt += "\n"

            # 현재 메시지
            context_prompt += f"사용자의 새로운 메시지: {message}\n\n"
            context_prompt += "위 대화 내용을 참고하여 적절한 응답을 생성하세요."

            # AI 호출
            response = await asyncio.to_thread(
                self.ai.generate,
                context_prompt,
                system_instruction=persona[:2000]  # 토큰 최적화
            )

            return response

        except Exception as e:
            print(f"❌ AI generation failed: {e}")
            return f"[시스템] 응답 생성 중 오류가 발생했습니다: {str(e)}"

    async def get_chat_history(self, user_id: str, limit: int = 50) -> list:
        """대화 기록 조회"""
        try:
            history = self.memory.load_chat(user_id, limit=limit)
            return history
        except Exception as e:
            print(f"❌ Error loading chat history: {e}")
            return []

    async def stream_council_log(self, log_file: Path) -> None:
        """
        council_log 파일을 실시간으로 스트리밍
        (Phase 2 후반부 구현 - 에이전트 내부 토론 과정)
        """
        try:
            if not log_file.exists():
                return

            with open(log_file, 'r', encoding='utf-8') as f:
                content = f.read()

            # 파싱 및 스트리밍 (간단한 구현)
            lines = content.split('\n')
            current_speaker = None
            current_text = ""

            for line in lines:
                if line.startswith("## 🗣️") or line.startswith("## 👑"):
                    if current_speaker and current_text:
                        await self.ws_manager.broadcast({
                            "type": "council_thought",
                            "speaker": current_speaker,
                            "text": current_text.strip()
                        })
                        await asyncio.sleep(0.5)  # 읽기 편하게 딜레이

                    current_speaker = line.replace("## 🗣️", "").replace("## 👑", "").strip()
                    current_text = ""
                else:
                    current_text += line + "\n"

            # 마지막 메시지
            if current_speaker and current_text:
                await self.ws_manager.broadcast({
                    "type": "council_thought",
                    "speaker": current_speaker,
                    "text": current_text.strip()
                })

        except Exception as e:
            print(f"❌ Error streaming council log: {e}")
