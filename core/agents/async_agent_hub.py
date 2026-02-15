#!/usr/bin/env python3
"""
Async Agent Hub - 비동기 에이전트 통신 허브
병렬 에이전트 실행 및 실시간 협업 지원

Features:
- asyncio 기반 병렬 메시지 라우팅
- SA + AD 동시 실행 지원
- 타임아웃 및 에러 핸들링
- Synapse Bridge 실시간 업데이트
"""

import asyncio
import json
import logging
from datetime import datetime
from typing import Dict, List, Optional, Any, Callable, Coroutine
from pathlib import Path
from enum import Enum
from collections import defaultdict, deque

logger = logging.getLogger(__name__)


class MessageType(Enum):
    """메시지 유형"""
    REQUEST = "request"
    RESPONSE = "response"
    BROADCAST = "broadcast"
    QUERY = "query"
    VOTE = "vote"
    DECISION = "decision"
    STATUS = "status"
    PARALLEL = "parallel"  # 병렬 실행 요청


class TaskPriority(Enum):
    """작업 우선순위"""
    CRITICAL = 0  # CD - Creative Director
    HIGH = 1      # SA - Strategy Analyst
    MEDIUM = 2    # CE - Chief Editor
    LOW = 3       # AD - Art Director


class AsyncAgentHub:
    """비동기 에이전트 통신 허브"""

    def __init__(self, project_root: str):
        self.project_root = Path(project_root)
        self.hub_dir = self.project_root / "knowledge" / "agent_hub"
        self.hub_dir.mkdir(parents=True, exist_ok=True)

        # Synapse Bridge 파일
        self.synapse_file = self.hub_dir / "synapse_bridge.json"

        # 에이전트 레지스트리
        self.agents: Dict[str, Dict[str, Any]] = {
            "CD": {
                "name": "Creative Director",
                "active": False,
                "handler": None,
                "priority": TaskPriority.CRITICAL
            },
            "TD": {
                "name": "Technical Director",
                "active": False,
                "handler": None,
                "priority": TaskPriority.HIGH
            },
            "SA": {
                "name": "Strategy Analyst",
                "active": False,
                "handler": None,
                "priority": TaskPriority.HIGH
            },
            "AD": {
                "name": "Art Director",
                "active": False,
                "handler": None,
                "priority": TaskPriority.LOW
            },
            "CE": {
                "name": "Chief Editor",
                "active": False,
                "handler": None,
                "priority": TaskPriority.MEDIUM
            }
        }

        # 비동기 메시지 큐
        self.message_queues: Dict[str, asyncio.Queue] = {
            agent: asyncio.Queue() for agent in self.agents
        }

        # 병렬 작업 추적
        self.parallel_tasks: Dict[str, List[asyncio.Task]] = {}

        # 작업 결과 캐시
        self.result_cache: Dict[str, Any] = {}

        # 통계
        self.stats = {
            "messages_routed": 0,
            "parallel_requests": 0,
            "cache_hits": 0,
            "start_time": datetime.now().isoformat()
        }

        # Synapse Bridge 초기화
        self._init_synapse_bridge()

    def _init_synapse_bridge(self):
        """Synapse Bridge 초기화"""
        synapse_data = {
            "active_agents": {},
            "collaboration_mode": "Parallel",
            "synapse_status": "Initialized",
            "last_update": datetime.now().isoformat(),
            "parallel_tasks": {},
            "performance": {
                "avg_response_time": 0,
                "throughput": 0,
                "efficiency": 0
            }
        }

        with open(self.synapse_file, 'w', encoding='utf-8') as f:
            json.dump(synapse_data, f, indent=2, ensure_ascii=False)

    def register_agent(self, agent_key: str, handler: Callable[[Dict], Coroutine]):
        """
        비동기 에이전트 등록

        Args:
            agent_key: 에이전트 키 (CD, TD, etc.)
            handler: 비동기 메시지 처리 핸들러
        """
        if agent_key in self.agents:
            self.agents[agent_key]["handler"] = handler
            self.agents[agent_key]["active"] = True
            logger.info(f"Agent {agent_key} registered to AsyncHub")

            # Synapse Bridge 업데이트
            self._update_synapse({
                "event": "agent_registered",
                "agent": agent_key,
                "timestamp": datetime.now().isoformat()
            })

    async def send_message_async(self, from_agent: str, to_agent: str,
                                 message_type: MessageType, data: Dict[str, Any],
                                 timeout: float = 30.0) -> Optional[Any]:
        """
        비동기 메시지 전송 및 응답 대기

        Args:
            from_agent: 발신 에이전트
            to_agent: 수신 에이전트
            message_type: 메시지 유형
            data: 메시지 데이터
            timeout: 타임아웃 (초)

        Returns:
            응답 결과
        """
        message_id = f"{datetime.now().timestamp()}-{from_agent}-{to_agent}"

        message = {
            "id": message_id,
            "timestamp": datetime.now().isoformat(),
            "from": from_agent,
            "to": to_agent,
            "type": message_type.value,
            "data": data
        }

        # 메시지 큐에 추가
        await self.message_queues[to_agent].put(message)

        # 즉시 처리
        if self.agents[to_agent]["active"] and self.agents[to_agent]["handler"]:
            try:
                handler = self.agents[to_agent]["handler"]
                result = await asyncio.wait_for(handler(message), timeout=timeout)
                self.stats["messages_routed"] += 1
                return result
            except asyncio.TimeoutError:
                logger.error(f"Message {message_id} timed out after {timeout}s")
                return None
            except Exception as e:
                logger.error(f"Error processing message {message_id}: {e}")
                return None

        return None

    async def parallel_request(self, from_agent: str,
                              targets: List[Dict[str, Any]],
                              timeout: float = 30.0) -> Dict[str, Any]:
        """
        병렬 에이전트 요청 (SA + AD 동시 실행)

        Args:
            from_agent: 발신 에이전트
            targets: 타겟 에이전트 정보 리스트
                     [{"agent": "SA", "data": {...}}, {"agent": "AD", "data": {...}}]
            timeout: 타임아웃 (초)

        Returns:
            {agent_key: result} 딕셔너리
        """
        parallel_id = f"parallel-{datetime.now().timestamp()}"
        logger.info(f"Starting parallel request {parallel_id} with {len(targets)} agents")

        # 병렬 작업 생성
        tasks = []
        agent_keys = []

        for target in targets:
            agent_key = target["agent"]
            data = target.get("data", {})
            message_type = MessageType[target.get("type", "REQUEST")]

            # 캐시 확인
            cache_key = f"{agent_key}-{json.dumps(data, sort_keys=True)}"
            if cache_key in self.result_cache:
                logger.info(f"Cache hit for {agent_key}")
                self.stats["cache_hits"] += 1
                continue

            agent_keys.append(agent_key)
            task = asyncio.create_task(
                self.send_message_async(from_agent, agent_key, message_type, data, timeout)
            )
            tasks.append(task)

        # Synapse Bridge 업데이트 - 병렬 작업 시작
        self._update_synapse({
            "event": "parallel_start",
            "parallel_id": parallel_id,
            "agents": agent_keys,
            "timestamp": datetime.now().isoformat()
        })

        # 병렬 실행
        start_time = datetime.now()
        results = await asyncio.gather(*tasks, return_exceptions=True)
        elapsed = (datetime.now() - start_time).total_seconds()

        # 결과 매핑
        result_dict = {}
        for i, agent_key in enumerate(agent_keys):
            result = results[i]
            if isinstance(result, Exception):
                logger.error(f"Agent {agent_key} failed: {result}")
                result_dict[agent_key] = {"error": str(result)}
            else:
                result_dict[agent_key] = result

                # 캐시 저장
                cache_key = f"{agent_key}-{json.dumps(targets[i]['data'], sort_keys=True)}"
                self.result_cache[cache_key] = result

        # Synapse Bridge 업데이트 - 병렬 작업 완료
        self._update_synapse({
            "event": "parallel_complete",
            "parallel_id": parallel_id,
            "elapsed_time": elapsed,
            "results": {k: "success" if "error" not in v else "failed" for k, v in result_dict.items()},
            "timestamp": datetime.now().isoformat()
        })

        self.stats["parallel_requests"] += 1
        logger.info(f"Parallel request {parallel_id} completed in {elapsed:.2f}s")

        return result_dict

    async def broadcast_async(self, from_agent: str, message_type: MessageType,
                             data: Dict[str, Any], exclude: List[str] = None):
        """
        비동기 브로드캐스트

        Args:
            from_agent: 발신 에이전트
            message_type: 메시지 유형
            data: 메시지 데이터
            exclude: 제외할 에이전트 리스트
        """
        exclude = exclude or []
        tasks = []

        for agent_key in self.agents:
            if agent_key != from_agent and agent_key not in exclude:
                if self.agents[agent_key]["active"]:
                    task = self.send_message_async(from_agent, agent_key, message_type, data)
                    tasks.append(task)

        await asyncio.gather(*tasks, return_exceptions=True)

    async def request_collaboration_async(self, initiator: str, participants: List[str],
                                         topic: str, context: Dict[str, Any],
                                         timeout: float = 60.0) -> Dict[str, Any]:
        """
        비동기 협업 요청

        Args:
            initiator: 협업 시작 에이전트
            participants: 참여 에이전트 리스트
            topic: 협업 주제
            context: 협업 컨텍스트
            timeout: 타임아웃 (초)

        Returns:
            협업 결과
        """
        collab_id = f"collab-{datetime.now().timestamp()}"
        logger.info(f"Starting collaboration {collab_id}: {topic}")

        # 협업 요청 병렬 전송
        targets = [
            {
                "agent": participant,
                "type": "REQUEST",
                "data": {
                    "collaboration_id": collab_id,
                    "topic": topic,
                    "context": context,
                    "action": "collaborate"
                }
            }
            for participant in participants if participant != initiator
        ]

        responses = await self.parallel_request(initiator, targets, timeout)

        # 결과 종합
        result = {
            "collaboration_id": collab_id,
            "topic": topic,
            "participants": participants,
            "responses": responses,
            "consensus": self._find_consensus(responses),
            "timestamp": datetime.now().isoformat()
        }

        logger.info(f"Collaboration {collab_id} completed")
        return result

    def _find_consensus(self, responses: Dict[str, Any]) -> str:
        """응답에서 합의점 찾기"""
        # 성공적인 응답만 추출
        valid_responses = [
            r for r in responses.values()
            if isinstance(r, dict) and "error" not in r
        ]

        if not valid_responses:
            return "합의 실패: 유효한 응답 없음"

        # 점수 기반 합의 (간단한 구현)
        if all("score" in r for r in valid_responses):
            avg_score = sum(r["score"] for r in valid_responses) / len(valid_responses)
            return f"평균 점수: {avg_score:.1f}/100"

        return f"참여자 {len(valid_responses)}명 응답 완료"

    def _update_synapse(self, event: Dict[str, Any]):
        """Synapse Bridge 실시간 업데이트"""
        try:
            # 기존 데이터 로드
            if self.synapse_file.exists():
                with open(self.synapse_file, 'r', encoding='utf-8') as f:
                    synapse_data = json.load(f)
            else:
                synapse_data = {"active_agents": {}, "events": []}

            # 이벤트 추가
            if "events" not in synapse_data:
                synapse_data["events"] = deque(maxlen=100)

            synapse_data["events"] = list(synapse_data.get("events", []))
            synapse_data["events"].append(event)
            synapse_data["events"] = synapse_data["events"][-100:]  # 최근 100개만 유지

            # 활성 에이전트 업데이트
            synapse_data["active_agents"] = {
                key: {
                    "name": info["name"],
                    "active": info["active"],
                    "priority": info["priority"].name,
                    "last_update": datetime.now().isoformat()
                }
                for key, info in self.agents.items()
            }

            synapse_data["last_update"] = datetime.now().isoformat()
            synapse_data["stats"] = self.stats

            # 저장
            with open(self.synapse_file, 'w', encoding='utf-8') as f:
                json.dump(synapse_data, f, indent=2, ensure_ascii=False)

        except Exception as e:
            logger.error(f"Synapse update error: {e}")

    def get_hub_status(self) -> Dict[str, Any]:
        """허브 상태 반환"""
        return {
            "timestamp": datetime.now().isoformat(),
            "active_agents": [
                key for key, info in self.agents.items() if info["active"]
            ],
            "pending_messages": {
                agent: self.message_queues[agent].qsize()
                for agent in self.agents
            },
            "parallel_tasks": len(self.parallel_tasks),
            "cache_size": len(self.result_cache),
            "stats": self.stats
        }

    async def shutdown(self):
        """허브 종료"""
        logger.info("Shutting down AsyncAgentHub")

        # 모든 진행 중인 작업 취소
        for tasks in self.parallel_tasks.values():
            for task in tasks:
                if not task.done():
                    task.cancel()

        # 캐시 클리어
        self.result_cache.clear()

        logger.info("AsyncAgentHub shutdown complete")


# 싱글톤 인스턴스
_async_hub_instance: Optional[AsyncAgentHub] = None


def get_async_hub(project_root: str = None) -> AsyncAgentHub:
    """싱글톤 AsyncHub 인스턴스 반환"""
    global _async_hub_instance

    if _async_hub_instance is None:
        if project_root is None:
            from pathlib import Path
            project_root = str(Path(__file__).resolve().parent.parent)
        _async_hub_instance = AsyncAgentHub(project_root)

    return _async_hub_instance
