#!/usr/bin/env python3
"""
Agent Notifier System
실시간 에이전트 알림 및 메시지 큐 관리

Features:
- 각 에이전트별 독립 메시지 큐
- 우선순위 기반 메시지 처리
- 브로드캐스트 및 타겟팅 지원
- 메시지 만료 처리
"""

import json
import logging
import threading
import queue
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any, Callable
from pathlib import Path
from collections import defaultdict

logger = logging.getLogger(__name__)

class AgentNotifier:
    """실시간 에이전트 통신 시스템"""

    def __init__(self, project_root: str):
        self.project_root = Path(project_root)
        self.notification_dir = self.project_root / "knowledge" / "notifications"
        self.notification_dir.mkdir(parents=True, exist_ok=True)

        # 에이전트별 메시지 큐
        self.message_queues: Dict[str, queue.PriorityQueue] = defaultdict(lambda: queue.PriorityQueue())

        # 에이전트 구독자 (콜백 함수)
        self.subscribers: Dict[str, List[Callable]] = defaultdict(list)

        # 글로벌 이벤트 로그
        self.event_log = []
        self.max_log_size = 1000

        # 스레드 세이프 락
        self.lock = threading.RLock()

        # 에이전트 활성 상태
        self.active_agents = {
            "CD": False,  # Creative Director
            "TD": False,  # Technical Director
            "AD": False,  # Art Director
            "CE": False,  # Chief Editor
            "SA": False   # Strategy Analyst
        }

    def notify_agent(self, agent_key: str, message: Dict[str, Any], priority: int = 5):
        """
        특정 에이전트에게 메시지 전송

        Args:
            agent_key: 에이전트 키 (CD, TD, AD, CE, SA)
            message: 전송할 메시지
            priority: 우선순위 (1=긴급, 10=낮음)
        """
        with self.lock:
            # 메시지 메타데이터 추가
            enriched_message = {
                "id": f"{datetime.now().timestamp()}",
                "timestamp": datetime.now().isoformat(),
                "agent": agent_key,
                "priority": priority,
                "data": message
            }

            # 큐에 추가 (우선순위, 타임스탬프, 메시지)
            self.message_queues[agent_key].put((
                priority,
                datetime.now(),
                enriched_message
            ))

            # 이벤트 로그 기록
            self._log_event("notify", agent_key, enriched_message)

            # 구독자에게 즉시 알림
            self._trigger_subscribers(agent_key, enriched_message)

            # 영구 저장 (중요 메시지만)
            if priority <= 3:
                self._persist_notification(agent_key, enriched_message)

            logger.info(f"Notified {agent_key}: {message.get('type', 'message')}")

    def broadcast(self, message: Dict[str, Any], priority: int = 5, exclude: List[str] = None):
        """
        모든 활성 에이전트에게 브로드캐스트

        Args:
            message: 브로드캐스트할 메시지
            priority: 우선순위
            exclude: 제외할 에이전트 리스트
        """
        exclude = exclude or []

        with self.lock:
            broadcast_agents = [
                agent for agent, active in self.active_agents.items()
                if active and agent not in exclude
            ]

            for agent in broadcast_agents:
                self.notify_agent(agent, message, priority)

            logger.info(f"Broadcasted to {len(broadcast_agents)} agents")

    def subscribe(self, agent_key: str, callback: Callable):
        """
        에이전트 메시지 구독

        Args:
            agent_key: 구독할 에이전트
            callback: 메시지 수신 시 호출할 함수
        """
        with self.lock:
            self.subscribers[agent_key].append(callback)
            self.active_agents[agent_key] = True
            logger.info(f"Agent {agent_key} subscribed")

    def unsubscribe(self, agent_key: str, callback: Callable = None):
        """구독 해제"""
        with self.lock:
            if callback:
                self.subscribers[agent_key].remove(callback)
            else:
                self.subscribers[agent_key].clear()
                self.active_agents[agent_key] = False

    def get_messages(self, agent_key: str, limit: int = 10) -> List[Dict[str, Any]]:
        """
        에이전트 메시지 큐에서 메시지 가져오기

        Returns:
            최대 limit개의 메시지 리스트
        """
        messages = []

        with self.lock:
            agent_queue = self.message_queues[agent_key]

            # 만료된 메시지 제거하며 가져오기
            while not agent_queue.empty() and len(messages) < limit:
                try:
                    priority, timestamp, msg = agent_queue.get_nowait()

                    # 5분 이상 된 메시지는 만료 처리
                    if datetime.now() - timestamp < timedelta(minutes=5):
                        messages.append(msg)
                    else:
                        logger.debug(f"Expired message for {agent_key}: {msg['id']}")

                except queue.Empty:
                    break

        return messages

    def get_pending_count(self, agent_key: str) -> int:
        """대기 중인 메시지 수 반환"""
        with self.lock:
            return self.message_queues[agent_key].qsize()

    def clear_queue(self, agent_key: str):
        """에이전트 큐 초기화"""
        with self.lock:
            self.message_queues[agent_key] = queue.PriorityQueue()
            logger.info(f"Cleared queue for {agent_key}")

    def _trigger_subscribers(self, agent_key: str, message: Dict[str, Any]):
        """구독자 콜백 트리거"""
        for callback in self.subscribers.get(agent_key, []):
            try:
                # 별도 스레드에서 콜백 실행 (논블로킹)
                threading.Thread(
                    target=callback,
                    args=(message,),
                    daemon=True
                ).start()
            except Exception as e:
                logger.error(f"Subscriber callback error: {e}")

    def _persist_notification(self, agent_key: str, message: Dict[str, Any]):
        """중요 알림 파일 시스템에 저장"""
        try:
            date_str = datetime.now().strftime("%Y%m%d")
            log_file = self.notification_dir / f"{agent_key}_{date_str}.jsonl"

            with open(log_file, "a", encoding="utf-8") as f:
                f.write(json.dumps(message, ensure_ascii=False) + "\n")

        except Exception as e:
            logger.error(f"Failed to persist notification: {e}")

    def _log_event(self, event_type: str, agent_key: str, data: Any):
        """이벤트 로깅"""
        with self.lock:
            event = {
                "type": event_type,
                "agent": agent_key,
                "timestamp": datetime.now().isoformat(),
                "data": data
            }

            self.event_log.append(event)

            # 로그 크기 제한
            if len(self.event_log) > self.max_log_size:
                self.event_log = self.event_log[-self.max_log_size:]

    def get_agent_status(self) -> Dict[str, Any]:
        """전체 에이전트 상태 반환"""
        with self.lock:
            status = {
                "timestamp": datetime.now().isoformat(),
                "active_agents": self.active_agents,
                "pending_messages": {
                    agent: self.get_pending_count(agent)
                    for agent in self.active_agents
                },
                "recent_events": self.event_log[-10:]
            }

        return status

    def notify_telegram_received(self, chat_id: str, text: str, detected_agent: str = None):
        """
        텔레그램 메시지 수신 이벤트 전파

        Args:
            chat_id: 텔레그램 채팅 ID
            text: 수신된 메시지
            detected_agent: 감지된 타겟 에이전트
        """
        message = {
            "type": "telegram_received",
            "chat_id": chat_id,
            "text": text[:500],  # 토큰 절약
            "detected_agent": detected_agent,
            "source": "telegram_daemon"
        }

        if detected_agent:
            # 특정 에이전트에게 우선 알림
            self.notify_agent(detected_agent, message, priority=2)
            # 다른 에이전트들에게도 낮은 우선순위로 알림
            self.broadcast(message, priority=7, exclude=[detected_agent])
        else:
            # 모든 에이전트에게 브로드캐스트
            self.broadcast(message, priority=5)

    def notify_task_update(self, task_type: str, task_data: Dict[str, Any]):
        """
        작업 업데이트 알림

        Args:
            task_type: 작업 유형 (created, completed, failed)
            task_data: 작업 데이터
        """
        message = {
            "type": "task_update",
            "task_type": task_type,
            "data": task_data,
            "source": "task_system"
        }

        # TD에게 우선 알림, 다른 에이전트들도 알아야 함
        self.notify_agent("TD", message, priority=3)
        self.broadcast(message, priority=6, exclude=["TD"])

    def notify_memory_sync(self, sync_data: Dict[str, Any]):
        """
        메모리 동기화 이벤트 알림

        Args:
            sync_data: 동기화 정보
        """
        message = {
            "type": "memory_sync",
            "data": sync_data,
            "source": "memory_manager"
        }

        # 모든 에이전트에게 중간 우선순위로 알림
        self.broadcast(message, priority=5)


# 싱글톤 인스턴스
_notifier_instance: Optional[AgentNotifier] = None

def get_notifier(project_root: str = None) -> AgentNotifier:
    """싱글톤 Notifier 인스턴스 반환"""
    global _notifier_instance

    if _notifier_instance is None:
        if project_root is None:
            from pathlib import Path
            project_root = str(Path(__file__).resolve().parent.parent)
        _notifier_instance = AgentNotifier(project_root)

    return _notifier_instance