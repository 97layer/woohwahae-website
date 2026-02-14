#!/usr/bin/env python3
"""
Agent Pusher - 에이전트 → 텔레그램 역방향 푸시 시스템
에이전트가 자율적으로 메시지를 전송할 수 있게 지원

Features:
- 에이전트 자율 메시지 발송
- 진행상황 실시간 보고
- 중요 이벤트 알림
- 메시지 큐잉 및 스로틀링
"""

import json
import logging
import asyncio
import aiohttp
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
from pathlib import Path
from collections import defaultdict
from enum import Enum

logger = logging.getLogger(__name__)


class MessagePriority(Enum):
    """메시지 우선순위"""
    CRITICAL = 1    # 즉시 전송
    HIGH = 2        # 높은 우선순위
    NORMAL = 3      # 일반
    LOW = 4         # 낮은 우선순위
    DEFERRED = 5    # 지연 가능


class AgentPusher:
    """에이전트 푸시 메시징 시스템"""

    def __init__(self, bot_token: str, project_root: str):
        self.bot_token = bot_token
        self.base_url = f"https://api.telegram.org/bot{bot_token}"
        self.project_root = Path(project_root)

        # 메시지 큐
        self.message_queue = asyncio.PriorityQueue()

        # 채널 관리
        self.registered_chats: List[int] = []
        self.agent_channels: Dict[str, List[int]] = defaultdict(list)

        # 레이트 리밋 관리
        self.rate_limiter = {
            "messages_per_second": 3,
            "last_sent": datetime.now(),
            "sent_count": 0
        }

        # 세션
        self.session: Optional[aiohttp.ClientSession] = None

        # 통계
        self.stats = {
            "messages_sent": 0,
            "messages_failed": 0,
            "agents_active": set()
        }

        # 설정 로드
        self._load_config()

    def _load_config(self):
        """설정 파일 로드"""
        config_file = self.project_root / "knowledge" / "telegram_chats.json"

        if config_file.exists():
            try:
                with open(config_file) as f:
                    config = json.load(f)
                    self.registered_chats = config.get("registered_chats", [])
                    logger.info(f"Loaded {len(self.registered_chats)} registered chats")
            except Exception as e:
                logger.error(f"Config load error: {e}")

    def register_chat(self, chat_id: int, agents: List[str] = None):
        """
        채팅 등록

        Args:
            chat_id: 텔레그램 채팅 ID
            agents: 이 채팅에 메시지를 보낼 수 있는 에이전트 리스트
        """
        if chat_id not in self.registered_chats:
            self.registered_chats.append(chat_id)

        if agents:
            for agent in agents:
                if chat_id not in self.agent_channels[agent]:
                    self.agent_channels[agent].append(chat_id)

        self._save_config()
        logger.info(f"Chat {chat_id} registered for agents: {agents}")

    def _save_config(self):
        """설정 저장"""
        config_file = self.project_root / "knowledge" / "telegram_chats.json"
        config_file.parent.mkdir(parents=True, exist_ok=True)

        config = {
            "registered_chats": self.registered_chats,
            "agent_channels": {k: list(v) for k, v in self.agent_channels.items()},
            "updated_at": datetime.now().isoformat()
        }

        with open(config_file, 'w') as f:
            json.dump(config, f, indent=2, ensure_ascii=False)

    async def push_message(self, agent_key: str, message: str,
                           chat_ids: List[int] = None,
                           priority: MessagePriority = MessagePriority.NORMAL,
                           metadata: Dict[str, Any] = None):
        """
        에이전트 메시지 푸시

        Args:
            agent_key: 에이전트 식별자
            message: 전송할 메시지
            chat_ids: 대상 채팅 ID 리스트 (None이면 등록된 모든 채팅)
            priority: 메시지 우선순위
            metadata: 추가 메타데이터
        """
        # 대상 채팅 결정
        if chat_ids is None:
            # 에이전트에게 할당된 채널 또는 전체
            chat_ids = self.agent_channels.get(agent_key, self.registered_chats)

        if not chat_ids:
            logger.warning(f"No registered chats for agent {agent_key}")
            return

        # 메시지 포맷팅
        formatted_message = self._format_message(agent_key, message, metadata)

        # 큐에 추가
        for chat_id in chat_ids:
            await self.message_queue.put((
                priority.value,
                datetime.now(),
                {
                    "chat_id": chat_id,
                    "text": formatted_message,
                    "agent": agent_key,
                    "metadata": metadata
                }
            ))

        self.stats["agents_active"].add(agent_key)
        logger.debug(f"Queued message from {agent_key} to {len(chat_ids)} chats")

    async def push_progress(self, agent_key: str, task: str,
                           progress: int, total: int,
                           status: str = None):
        """
        작업 진행상황 푸시

        Args:
            agent_key: 에이전트
            task: 작업 설명
            progress: 현재 진행
            total: 전체
            status: 상태 메시지
        """
        percentage = (progress / max(1, total)) * 100

        # 진행바 생성
        bar_length = 20
        filled = int(bar_length * progress / max(1, total))
        bar = "█" * filled + "░" * (bar_length - filled)

        message = f"📊 **작업 진행 상황**\n\n"
        message += f"작업: {task}\n"
        message += f"진행: [{bar}] {percentage:.1f}%\n"
        message += f"({progress}/{total})\n"

        if status:
            message += f"\n상태: {status}"

        await self.push_message(
            agent_key,
            message,
            priority=MessagePriority.LOW,
            metadata={"type": "progress", "task": task}
        )

    async def push_alert(self, agent_key: str, alert_type: str,
                        title: str, message: str,
                        actions: List[Dict[str, str]] = None):
        """
        중요 알림 푸시

        Args:
            agent_key: 에이전트
            alert_type: 알림 유형 (error, warning, info, success)
            title: 알림 제목
            message: 알림 내용
            actions: 액션 버튼 리스트
        """
        emoji_map = {
            "error": "🔴",
            "warning": "🟡",
            "info": "🔵",
            "success": "🟢"
        }

        emoji = emoji_map.get(alert_type, "⚪")

        alert_message = f"{emoji} **{title}**\n\n{message}"

        # 인라인 키보드 생성
        reply_markup = None
        if actions:
            keyboard = []
            for action in actions:
                keyboard.append([{
                    "text": action.get("text", "Action"),
                    "callback_data": action.get("data", "action")
                }])
            reply_markup = {"inline_keyboard": keyboard}

        await self.push_message(
            agent_key,
            alert_message,
            priority=MessagePriority.HIGH,
            metadata={
                "type": "alert",
                "alert_type": alert_type,
                "reply_markup": reply_markup
            }
        )

    async def push_report(self, agent_key: str, report_title: str,
                         sections: Dict[str, Any]):
        """
        보고서 푸시

        Args:
            agent_key: 에이전트
            report_title: 보고서 제목
            sections: 섹션별 내용
        """
        report = f"📋 **{report_title}**\n"
        report += f"_Generated by {agent_key} at {datetime.now().strftime('%Y-%m-%d %H:%M')}_\n\n"

        for section_title, content in sections.items():
            report += f"**{section_title}**\n"

            if isinstance(content, list):
                for item in content:
                    report += f"• {item}\n"
            elif isinstance(content, dict):
                for key, value in content.items():
                    report += f"  {key}: {value}\n"
            else:
                report += f"{content}\n"

            report += "\n"

        await self.push_message(
            agent_key,
            report,
            priority=MessagePriority.NORMAL,
            metadata={"type": "report", "title": report_title}
        )

    async def start_processor(self):
        """메시지 프로세서 시작"""
        if not self.session:
            self.session = aiohttp.ClientSession()

        logger.info("Starting Agent Pusher processor...")

        while True:
            try:
                # 레이트 리밋 확인
                await self._check_rate_limit()

                # 큐에서 메시지 가져오기
                if not self.message_queue.empty():
                    priority, timestamp, msg_data = await self.message_queue.get()

                    # 5분 이상 된 메시지는 스킵
                    if datetime.now() - timestamp > timedelta(minutes=5):
                        logger.debug(f"Skipping expired message")
                        continue

                    # 전송
                    await self._send_telegram_message(msg_data)

                await asyncio.sleep(0.1)

            except Exception as e:
                logger.error(f"Processor error: {e}")
                await asyncio.sleep(1)

    async def _send_telegram_message(self, msg_data: Dict[str, Any]):
        """텔레그램 메시지 전송"""
        try:
            url = f"{self.base_url}/sendMessage"

            data = {
                "chat_id": msg_data["chat_id"],
                "text": msg_data["text"],
                "parse_mode": "Markdown"
            }

            # 인라인 키보드 추가
            if msg_data.get("metadata", {}).get("reply_markup"):
                data["reply_markup"] = json.dumps(
                    msg_data["metadata"]["reply_markup"]
                )

            async with self.session.post(url, json=data, timeout=10) as resp:
                if resp.status == 200:
                    self.stats["messages_sent"] += 1
                    logger.debug(f"Message sent to {msg_data['chat_id']}")
                else:
                    error = await resp.text()
                    logger.error(f"Send error: {error}")
                    self.stats["messages_failed"] += 1

        except Exception as e:
            logger.error(f"Telegram send error: {e}")
            self.stats["messages_failed"] += 1

    async def _check_rate_limit(self):
        """레이트 리밋 확인 및 대기"""
        now = datetime.now()
        time_diff = (now - self.rate_limiter["last_sent"]).total_seconds()

        if time_diff >= 1:
            # 1초 경과, 카운터 리셋
            self.rate_limiter["sent_count"] = 0
            self.rate_limiter["last_sent"] = now
        elif self.rate_limiter["sent_count"] >= self.rate_limiter["messages_per_second"]:
            # 리밋 도달, 대기
            wait_time = 1 - time_diff
            await asyncio.sleep(wait_time)
            self.rate_limiter["sent_count"] = 0
            self.rate_limiter["last_sent"] = datetime.now()

        self.rate_limiter["sent_count"] += 1

    def _format_message(self, agent_key: str, message: str,
                       metadata: Optional[Dict] = None) -> str:
        """메시지 포맷팅"""
        # 에이전트 서명 추가
        agent_names = {
            "CD": "Creative Director",
            "TD": "Technical Director",
            "AD": "Art Director",
            "CE": "Chief Editor",
            "SA": "Strategy Analyst"
        }

        agent_name = agent_names.get(agent_key, agent_key)

        # 메시지 타입별 포맷
        if metadata and metadata.get("type") == "progress":
            # 진행상황은 서명 없이
            return message
        elif metadata and metadata.get("type") == "alert":
            # 알림도 서명 없이
            return message
        else:
            # 일반 메시지는 서명 포함
            return f"{message}\n\n— _{agent_name}_"

    async def close(self):
        """리소스 정리"""
        if self.session:
            await self.session.close()

    def get_stats(self) -> Dict[str, Any]:
        """통계 반환"""
        return {
            "messages_sent": self.stats["messages_sent"],
            "messages_failed": self.stats["messages_failed"],
            "active_agents": list(self.stats["agents_active"]),
            "registered_chats": len(self.registered_chats),
            "queue_size": self.message_queue.qsize()
        }


# 전역 인스턴스
_pusher_instance: Optional[AgentPusher] = None


def get_pusher(bot_token: str = None, project_root: str = None) -> AgentPusher:
    """싱글톤 Pusher 인스턴스 반환"""
    global _pusher_instance

    if _pusher_instance is None:
        if bot_token is None:
            from libs.core_config import TELEGRAM_CONFIG
            bot_token = TELEGRAM_CONFIG["BOT_TOKEN"]

        if project_root is None:
            project_root = str(Path(__file__).resolve().parent.parent)

        _pusher_instance = AgentPusher(bot_token, project_root)

    return _pusher_instance