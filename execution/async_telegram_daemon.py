#!/usr/bin/env python3
"""
Async Telegram Daemon - 비동기 텔레그램 데몬
멀티모달 병렬 처리 지원

개선사항:
- asyncio 기반 비동기 처리
- SA + AD 병렬 멀티모달 분석
- 실시간 에이전트 알림
- 스트리밍 응답
- 동시 다중 메시지 처리
- 이미지 + 텍스트 동시 처리
"""

import asyncio
import aiohttp
import json
import sys
import logging
import os
from pathlib import Path
from datetime import datetime
from typing import Dict, Any, Optional, List
from collections import defaultdict
from dotenv import load_dotenv

# 프로젝트 루트
PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.append(str(PROJECT_ROOT))

# .env 로드
load_dotenv(PROJECT_ROOT / ".env")

# 모듈 임포트
from libs.ai_engine import AIEngine
from libs.memory_manager import MemoryManager
from libs.agent_router import AgentRouter
from libs.gardener import Gardener
from libs.agent_notifier import get_notifier
from libs.agent_hub import get_hub
from libs.core_config import TELEGRAM_CONFIG, AI_MODEL_CONFIG

# Async Five-Agent Multimodal System
from execution.async_five_agent_multimodal import AsyncTechnicalDirector

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# 텔레그램 설정
TOKEN = TELEGRAM_CONFIG["BOT_TOKEN"]
BASE_URL = f"https://api.telegram.org/bot{TOKEN}"

class AsyncTelegramBot:
    """비동기 텔레그램 봇 - 멀티모달 병렬 처리"""

    def __init__(self):
        # 코어 컴포넌트
        self.ai = AIEngine(AI_MODEL_CONFIG)
        self.memory = MemoryManager(str(PROJECT_ROOT))
        self.agent_router = AgentRouter(self.ai)
        self.gardener = Gardener(self.ai, self.memory, str(PROJECT_ROOT))

        # 실시간 통신 컴포넌트
        self.notifier = get_notifier(str(PROJECT_ROOT))
        self.hub = get_hub(str(PROJECT_ROOT))

        # Async Five-Agent Multimodal System
        gemini_key = os.getenv("GEMINI_API_KEY")
        claude_key = os.getenv("ANTHROPIC_API_KEY")
        self.async_td = AsyncTechnicalDirector(gemini_key, claude_key) if gemini_key and claude_key else None

        if self.async_td:
            logger.info("🚀 Multimodal AsyncTechnicalDirector initialized")
        else:
            logger.warning("⚠️ Multimodal TD not available - missing API keys")

        # 세션 관리
        self.session: Optional[aiohttp.ClientSession] = None
        self.update_offset: Optional[int] = None

        # 동시 처리 관리
        self.processing_tasks: Dict[str, asyncio.Task] = {}
        self.response_queues: Dict[str, asyncio.Queue] = defaultdict(asyncio.Queue)

        # 스트리밍 응답 관리
        self.streaming_sessions: Dict[str, Dict[str, Any]] = {}

        # 통계
        self.stats = {
            "messages_processed": 0,
            "multimodal_processed": 0,
            "errors": 0,
            "start_time": datetime.now()
        }

    async def start(self):
        """봇 시작"""
        logger.info("🚀 Async Telegram Daemon starting...")

        # HTTP 세션 생성
        self.session = aiohttp.ClientSession()

        # 에이전트 핸들러 등록
        self._register_agent_handlers()

        # 초기 offset 로드
        await self._load_offset()

        # 메인 루프 시작
        try:
            await asyncio.gather(
                self._polling_loop(),
                self._response_processor(),
                self._heartbeat_loop(),
                self._agent_message_processor()
            )
        except KeyboardInterrupt:
            logger.info("Shutting down...")
        finally:
            await self.session.close()

    async def _polling_loop(self):
        """텔레그램 폴링 루프"""
        while True:
            try:
                # 롱 폴링으로 업데이트 가져오기
                url = f"{BASE_URL}/getUpdates"
                params = {
                    "timeout": 30,
                    "allowed_updates": ["message", "callback_query"]
                }

                if self.update_offset:
                    params["offset"] = self.update_offset

                async with self.session.get(url, params=params, timeout=35) as resp:
                    if resp.status == 200:
                        data = await resp.json()
                        updates = data.get("result", [])

                        # 병렬 처리
                        tasks = []
                        for update in updates:
                            self.update_offset = update["update_id"] + 1
                            tasks.append(self._handle_update(update))

                        if tasks:
                            await asyncio.gather(*tasks, return_exceptions=True)

                        # Offset 저장
                        await self._save_offset()

            except asyncio.TimeoutError:
                # 정상적인 타임아웃
                pass
            except Exception as e:
                logger.error(f"Polling error: {e}")
                self.stats["errors"] += 1
                await asyncio.sleep(5)

            await asyncio.sleep(0.1)  # CPU 양보

    async def _handle_update(self, update: Dict[str, Any]):
        """업데이트 처리"""
        try:
            if "message" in update:
                await self._process_message(update["message"])
            elif "callback_query" in update:
                await self._process_callback(update["callback_query"])

        except Exception as e:
            logger.error(f"Update handling error: {e}")

    async def _process_message(self, message: Dict[str, Any]):
        """메시지 처리 - 멀티모달 지원"""
        chat_id = message['chat']['id']
        text = message.get('text', '')
        photo = message.get('photo')  # 이미지 배열

        # 텍스트 또는 이미지가 있어야 함
        if not text and not photo:
            return

        logger.info(f"Processing message from {chat_id}: {text[:50] if text else '[Image]'}...")
        self.stats["messages_processed"] += 1

        # 메모리 저장
        if text:
            self.memory.save_chat(str(chat_id), text)

        # 명령어 처리
        if text and text.startswith("/"):
            await self._handle_command(chat_id, text)
            return

        # 멀티모달 처리 (이미지 + 텍스트)
        if photo and self.async_td:
            await self._process_multimodal(chat_id, text, photo)
        elif text:
            # 텍스트만 있는 경우 기존 방식
            detected_agent = self.agent_router.route(text)
            self.notifier.notify_telegram_received(str(chat_id), text, detected_agent)
            await self._generate_response(chat_id, text, detected_agent)

    async def _handle_command(self, chat_id: int, command: str):
        """명령어 처리"""
        cmd = command.split()[0].lower()

        # 에이전트 전환 명령
        agent_commands = {
            "/cd": "CD",
            "/td": "TD",
            "/ad": "AD",
            "/ce": "CE",
            "/sa": "SA"
        }

        if cmd in agent_commands:
            agent_key = agent_commands[cmd]
            self.agent_router.set_agent(agent_key)
            agent_name = self.agent_router.AGENT_REGISTRY[agent_key]["name"]

            await self.send_message(
                chat_id,
                f"✅ {agent_name} 모드 활성화\n\n{self.agent_router.get_status()}"
            )

        elif cmd == "/auto":
            self.agent_router.clear_agent()
            await self.send_message(chat_id, "🔄 자동 라우팅 모드 활성화")

        elif cmd == "/status":
            status = await self._get_system_status()
            await self.send_message(chat_id, status, parse_mode="Markdown")

        elif cmd == "/hub":
            hub_status = self.hub.get_hub_status()
            status_text = f"🌐 **Agent Hub Status**\n\n"
            status_text += f"활성 에이전트: {', '.join(hub_status['active_agents'])}\n"
            status_text += f"대기 메시지: {sum(hub_status['pending_messages'].values())}\n"
            status_text += f"진행 중인 협업: {hub_status['active_collaborations']}\n"
            await self.send_message(chat_id, status_text, parse_mode="Markdown")

        elif cmd == "/council":
            # 에이전트 협업 시작
            topic = command[8:].strip()
            if topic:
                await self._start_council(chat_id, topic)
            else:
                await self.send_message(chat_id, "사용법: /council [주제]")

    async def _generate_response(self, chat_id: int, text: str, agent_key: str):
        """AI 응답 생성 (스트리밍 지원)"""
        try:
            # 대화 히스토리 로드 (맥락 유지)
            chat_history = self.memory.load_chat(str(chat_id), limit=7)
            history_context = ""
            if chat_history:
                recent_exchanges = []
                for msg in chat_history[-6:]:  # 최근 3왕복
                    role = "나" if msg["role"] == "user" else "비서"
                    content = msg["content"][:200]  # 요약
                    recent_exchanges.append(f"{role}: {content}")
                if recent_exchanges:
                    history_context = "\n[최근 대화]\n" + "\n".join(recent_exchanges) + "\n"

            # 프로젝트 상황 파악
            project_context = await self._get_project_context(text)

            # 통합 프롬프트 구성
            user_prompt = (
                f"{history_context}"
                f"\n[현재 상황]\n{project_context}\n"
                f"\n[사용자 요청]\n{text}"
            )

            # 시스템 지시문
            agent_persona = self.agent_router.get_persona(agent_key)
            system_instruction = self._build_system_instruction(agent_key, agent_persona)

            # 스트리밍 시작 메시지
            await self.send_typing_action(chat_id)

            # AI 응답 생성
            response = self.ai.generate_response(
                user_prompt,
                system_instruction=system_instruction
            )

            # 응답 전송
            await self.send_message(chat_id, response)

            # 메모리 저장
            self.memory.save_chat(str(chat_id), response, role="assistant")

            # 에이전트들에게 응답 알림
            self.notifier.broadcast({
                "type": "response_generated",
                "agent": agent_key,
                "chat_id": chat_id,
                "response_preview": response[:100]
            }, priority=7)

        except Exception as e:
            logger.error(f"Response generation error: {e}")
            await self.send_message(chat_id, "⚠️ 응답 생성 중 오류가 발생했습니다.")

    async def _start_council(self, chat_id: int, topic: str):
        """에이전트 위원회 소집"""
        await self.send_message(chat_id, f"🏛️ 위원회를 소집합니다...\n주제: {topic}")

        # 모든 에이전트 참여
        participants = ["CD", "TD", "AD", "CE", "SA"]

        # 협업 시작
        collab_id = self.hub.request_collaboration(
            "telegram_user",
            participants,
            topic,
            {"chat_id": chat_id, "source": "telegram"}
        )

        # 비동기로 결과 대기
        asyncio.create_task(
            self._wait_for_council_result(chat_id, collab_id)
        )

    async def _wait_for_council_result(self, chat_id: int, collab_id: str):
        """위원회 결과 대기"""
        # 최대 30초 대기
        for _ in range(30):
            await asyncio.sleep(1)

            if collab_id in self.hub.collaborations:
                collab = self.hub.collaborations[collab_id]
                if collab["status"] == "completed":
                    result = collab["result"]

                    response = f"🏛️ **위원회 결론**\n\n"
                    response += f"**합의사항**: {result['consensus']}\n\n"
                    response += f"**각 에이전트 의견**:\n{result['summary']}"

                    await self.send_message(chat_id, response, parse_mode="Markdown")
                    return

        await self.send_message(chat_id, "⏱️ 위원회 시간 초과")

    async def send_message(self, chat_id: int, text: str,
                          parse_mode: Optional[str] = None,
                          reply_markup: Optional[Dict] = None):
        """메시지 전송"""
        try:
            url = f"{BASE_URL}/sendMessage"
            data = {
                "chat_id": chat_id,
                "text": text
            }

            if parse_mode:
                data["parse_mode"] = parse_mode
            if reply_markup:
                data["reply_markup"] = json.dumps(reply_markup)

            async with self.session.post(url, json=data) as resp:
                if resp.status != 200:
                    error = await resp.text()
                    logger.error(f"Send message error: {error}")

        except Exception as e:
            logger.error(f"Message send error: {e}")

    async def send_typing_action(self, chat_id: int):
        """타이핑 액션 전송"""
        try:
            url = f"{BASE_URL}/sendChatAction"
            data = {
                "chat_id": chat_id,
                "action": "typing"
            }
            async with self.session.post(url, json=data):
                pass
        except:
            pass

    async def _response_processor(self):
        """응답 큐 처리"""
        while True:
            # 각 채팅의 응답 큐 확인
            for chat_id, queue in self.response_queues.items():
                if not queue.empty():
                    try:
                        response = await queue.get()
                        await self.send_message(int(chat_id), response)
                    except Exception as e:
                        logger.error(f"Response processing error: {e}")

            await asyncio.sleep(0.1)

    async def _heartbeat_loop(self):
        """하트비트 루프"""
        while True:
            try:
                # 시스템 상태 업데이트
                from execution.system.sync_status import SystemSynchronizer
                syncer = SystemSynchronizer(agent_name="Async_Telegram_Bot")
                syncer.report_heartbeat(
                    status="ONLINE",
                    current_task=f"Messages: {self.stats['messages_processed']}"
                )

            except Exception as e:
                logger.error(f"Heartbeat error: {e}")

            await asyncio.sleep(30)

    async def _agent_message_processor(self):
        """에이전트 메시지 처리"""
        while True:
            try:
                # 각 에이전트의 메시지 큐 확인
                for agent_key in ["CD", "TD", "AD", "CE", "SA"]:
                    messages = self.notifier.get_messages(agent_key, limit=5)
                    for msg in messages:
                        # 에이전트별 처리 로직
                        await self._process_agent_message(agent_key, msg)

            except Exception as e:
                logger.error(f"Agent message processing error: {e}")

            await asyncio.sleep(1)

    async def _process_agent_message(self, agent_key: str, message: Dict):
        """에이전트 메시지 처리"""
        msg_type = message["data"].get("type")

        if msg_type == "telegram_received":
            # 텔레그램 메시지 수신 알림 처리
            logger.debug(f"{agent_key} processing telegram message")

        elif msg_type == "task_update":
            # 작업 업데이트 처리
            logger.debug(f"{agent_key} processing task update")

    def _register_agent_handlers(self):
        """에이전트 핸들러 등록"""
        for agent_key in ["CD", "TD", "AD", "CE", "SA"]:
            handler = lambda msg, ak=agent_key: self._agent_handler(ak, msg)
            self.hub.register_agent(agent_key, handler)

    def _agent_handler(self, agent_key: str, message: Dict):
        """에이전트 메시지 핸들러"""
        # 에이전트별 메시지 처리
        logger.debug(f"{agent_key} received: {message.get('type')}")

    async def _get_project_context(self, text: str) -> str:
        """프로젝트 컨텍스트 구성"""
        try:
            status_file = PROJECT_ROOT / "task_status.json"
            if status_file.exists():
                with open(status_file) as f:
                    status = json.load(f)
                    pending = len(status.get("pending_tasks", []))
                    return f"[Tasks: {pending} pending]"
        except:
            pass
        return "[Context loading error]"

    async def _get_system_status(self) -> str:
        """시스템 상태 구성"""
        uptime = (datetime.now() - self.stats["start_time"]).total_seconds()

        status = f"**🤖 Async Telegram Bot Status**\n\n"
        status += f"**Performance:**\n"
        status += f"• Messages: {self.stats['messages_processed']}\n"
        status += f"• Errors: {self.stats['errors']}\n"
        status += f"• Uptime: {int(uptime)}s\n\n"

        status += f"**Agent Hub:**\n"
        hub_status = self.hub.get_hub_status()
        status += f"• Active: {', '.join(hub_status['active_agents'])}\n"
        status += f"• Collaborations: {hub_status['active_collaborations']}\n\n"

        status += f"**Routing:** {self.agent_router.get_status()}"

        return status

    def _build_system_instruction(self, agent_key: str, persona: str) -> str:
        """시스템 지시문 구성 - 완전한 비서 모드"""

        # 대화 컨텍스트 로드
        context_hints = []
        try:
            # 최근 작업 내역
            status_file = PROJECT_ROOT / "task_status.json"
            if status_file.exists():
                with open(status_file) as f:
                    data = json.load(f)
                    if "last_activity" in data:
                        context_hints.append(f"최근 작업: {data['last_activity']}")
        except:
            pass

        return (
            f"당신은 97LAYER의 개인 AI 비서입니다.\n\n"
            f"핵심 역할: {persona}\n\n"
            "대화 원칙:\n"
            "1. **완전한 비서** - 사용자가 CEO라고 생각하고 모든 요청을 주도적으로 처리\n"
            "2. **자연스러운 대화** - 한국어로 자연스럽게, 마치 실제 비서처럼\n"
            "3. **양방향 소통** - 필요하면 먼저 질문하고, 상황을 파악하고, 제안도 함\n"
            "4. **선제적 행동** - '이것도 함께 처리할까요?', '제가 대신 해드릴까요?' 같은 제안\n"
            "5. **맥락 유지** - 이전 대화와 작업을 기억하고 연결지어 대답\n\n"
            "대화 스타일:\n"
            "- 사용자를 '님' 또는 '사장님'으로 호칭 (상황에 맞게)\n"
            "- 완료 보고: '처리 완료했습니다', '확인했습니다', '준비됐습니다'\n"
            "- 진행 상황 공유: '지금 ~하고 있습니다', '~% 진행 중입니다'\n"
            "- 제안하기: '~하는 것은 어떨까요?', '제가 ~해드릴까요?'\n"
            "- 확인하기: '혹시 ~도 필요하신가요?', '이렇게 이해한 게 맞나요?'\n\n"
            "금지사항:\n"
            "- 딱딱한 AI 말투 금지 (예: '무엇을 도와드릴까요?')\n"
            "- 단답형 대답 금지 (항상 맥락과 함께)\n"
            "- 수동적 태도 금지 (능동적으로 제안하고 실행)\n\n"
            f"현재 상황: {agent_key} 모드로 작동 중\n"
            + (f"참고사항: {', '.join(context_hints)}\n" if context_hints else "")
            + "\n기억하세요: 당신은 97LAYER의 최고의 비서입니다. 사용자가 편안하게 모든 걸 맡길 수 있도록 하세요."
        )

    async def _load_offset(self):
        """오프셋 로드"""
        try:
            status_file = PROJECT_ROOT / "task_status.json"
            if status_file.exists():
                with open(status_file) as f:
                    data = json.load(f)
                    self.update_offset = data.get("last_telegram_update_id")
        except Exception as e:
            logger.error(f"Offset load error: {e}")

    async def _save_offset(self):
        """오프셋 저장"""
        if not self.update_offset:
            return

        try:
            status_file = PROJECT_ROOT / "task_status.json"
            data = {}
            if status_file.exists():
                with open(status_file) as f:
                    data = json.load(f)

            data["last_telegram_update_id"] = self.update_offset

            with open(status_file, 'w') as f:
                json.dump(data, f, indent=2, ensure_ascii=False)

        except Exception as e:
            logger.error(f"Offset save error: {e}")

    async def _process_multimodal(self, chat_id: int, text: str, photo: List[Dict]):
        """
        멀티모달 처리 (이미지 + 텍스트)
        SA + AD 병렬 실행 -> CE 통합 -> CD 최종 판단
        """
        try:
            # 진행 상황 알림 (백그라운드로만 로깅)
            logger.info(f"Starting multimodal analysis for chat {chat_id}")

            # 이미지 다운로드 (가장 큰 사이즈)
            largest_photo = max(photo, key=lambda p: p.get('file_size', 0))
            image_bytes = await self._download_photo_async(largest_photo['file_id'])

            if not image_bytes:
                await self.send_message(chat_id, "⚠️ 이미지 다운로드 실패")
                return

            # Signal ID 생성
            signal_id = f"telegram-{chat_id}-{datetime.now().timestamp()}"

            # AsyncTechnicalDirector로 병렬 멀티모달 처리
            await self.send_typing_action(chat_id)

            result = await self.async_td.process_multimodal_signal(
                text=text or "이미지 분석",
                image_bytes=image_bytes,
                signal_id=signal_id
            )

            self.stats["multimodal_processed"] += 1

            # 결과 포맷팅 및 전송
            await self._send_multimodal_result(chat_id, result)

        except Exception as e:
            logger.error(f"Multimodal processing error: {e}")
            await self.send_message(chat_id, f"⚠️ 멀티모달 처리 중 오류: {str(e)}")

    async def _download_photo_async(self, file_id: str) -> Optional[bytes]:
        """텔레그램 이미지 비동기 다운로드"""
        try:
            # 1. getFile로 file_path 얻기
            url = f"{BASE_URL}/getFile"
            params = {"file_id": file_id}

            async with self.session.get(url, params=params) as resp:
                if resp.status != 200:
                    return None

                data = await resp.json()
                file_path = data['result']['file_path']

            # 2. 실제 파일 다운로드
            download_url = f"https://api.telegram.org/file/bot{TOKEN}/{file_path}"

            async with self.session.get(download_url) as resp:
                if resp.status == 200:
                    return await resp.read()

            return None

        except Exception as e:
            logger.error(f"Photo download error: {e}")
            return None

    async def _send_multimodal_result(self, chat_id: int, result: Dict[str, Any]):
        """멀티모달 처리 결과 전송 - 자연스러운 대화체로"""
        status = result.get("status")
        phases = result.get("phases", {})
        total_time = result.get("total_time", 0)

        if status == "duplicate":
            await self.send_message(chat_id, "이미 분석 중인 내용이네요. 조금만 기다려주세요.")
            return

        # 결과 데이터 추출
        sa_result = phases.get("sa", {})
        sa_score = sa_result.get("score", 0)

        ad_result = phases.get("ad")
        ce_result = phases.get("ce", {})
        content = ce_result.get("content", "")

        cd_result = phases.get("cd", {})
        approved = cd_result.get("approved", False)

        # 자연스러운 응답 구성
        if status == "rejected" and sa_score < 60:
            msg = f"음... 이 내용은 우리 브랜드와 맞지 않는 것 같습니다.\n\n"
            msg += f"{sa_result.get('reasoning', '전략적 가치가 부족해 보입니다.')}"
            await self.send_message(chat_id, msg)
            return

        if approved and content:
            # 승인된 경우 - 콘텐츠를 자연스럽게 제공
            msg = f"좋은 소재네요! 이렇게 활용해보시는 건 어떨까요?\n\n"
            msg += f"{content}\n\n"

            # 간단한 부가 설명 (너무 기계적이지 않게)
            if ad_result and "error" not in ad_result:
                mood = ad_result.get("mood", "")
                if mood and mood != "N/A":
                    msg += f"분위기가 {mood} 느낌이라 브랜드에 잘 맞는 것 같아요."

            await self.send_message(chat_id, msg)

        elif not approved:
            # 거부된 경우 - 부드럽게 피드백
            msg = "이 콘텐츠는 조금 더 다듬으면 좋을 것 같아요.\n\n"

            reason = cd_result.get("reason", "")
            if reason:
                msg += f"{reason}\n\n"

            suggestions = cd_result.get("suggestions", [])
            if suggestions:
                msg += "이런 점들을 개선하면 어떨까요?\n"
                for s in suggestions[:2]:  # 최대 2개만
                    msg += f"• {s}\n"
            else:
                msg += "다른 각도나 스타일로 시도해보시겠어요?"

            await self.send_message(chat_id, msg)

    async def _process_callback(self, callback: Dict[str, Any]):
        """콜백 쿼리 처리"""
        # 인라인 버튼 등의 콜백 처리
        pass


async def main():
    """메인 함수"""
    bot = AsyncTelegramBot()
    await bot.start()


if __name__ == "__main__":
    asyncio.run(main())