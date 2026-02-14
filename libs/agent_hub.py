#!/usr/bin/env python3
"""
Agent Hub - 중앙 에이전트 통신 허브
에이전트 간 실시간 협업 및 작업 조정

Features:
- 에이전트 간 직접 메시징
- 작업 분담 및 조정
- 합의 도출 메커니즘
- 실시간 상태 동기화
"""

import json
import logging
import threading
import queue
import time
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any, Callable, Tuple
from pathlib import Path
from enum import Enum
from collections import defaultdict

logger = logging.getLogger(__name__)


class MessageType(Enum):
    """메시지 유형"""
    REQUEST = "request"      # 작업 요청
    RESPONSE = "response"    # 응답
    BROADCAST = "broadcast"  # 브로드캐스트
    QUERY = "query"         # 질의
    VOTE = "vote"          # 투표
    DECISION = "decision"   # 결정
    STATUS = "status"       # 상태 업데이트


class AgentHub:
    """에이전트 통신 허브"""

    def __init__(self, project_root: str):
        self.project_root = Path(project_root)
        self.hub_dir = self.project_root / "knowledge" / "agent_hub"
        self.hub_dir.mkdir(parents=True, exist_ok=True)

        # 에이전트 레지스트리
        self.agents: Dict[str, Dict[str, Any]] = {
            "CD": {"name": "Creative Director", "active": False, "handler": None},
            "TD": {"name": "Technical Director", "active": False, "handler": None},
            "AD": {"name": "Art Director", "active": False, "handler": None},
            "CE": {"name": "Chief Editor", "active": False, "handler": None},
            "SA": {"name": "Strategy Analyst", "active": False, "handler": None}
        }

        # 메시지 라우팅 테이블
        self.routing_table: Dict[str, queue.Queue] = defaultdict(queue.Queue)

        # 작업 큐
        self.task_queue = queue.PriorityQueue()

        # 진행 중인 협업
        self.collaborations: Dict[str, Dict[str, Any]] = {}

        # 투표 세션
        self.voting_sessions: Dict[str, Dict[str, Any]] = {}

        # 스레드 풀
        self.worker_threads: List[threading.Thread] = []
        self.running = False

        # 통계
        self.stats = {
            "messages_routed": 0,
            "collaborations_completed": 0,
            "decisions_made": 0,
            "start_time": datetime.now()
        }

    def register_agent(self, agent_key: str, handler: Callable[[Dict], Any]):
        """
        에이전트 등록

        Args:
            agent_key: 에이전트 키 (CD, TD, etc.)
            handler: 메시지 처리 핸들러 함수
        """
        if agent_key in self.agents:
            self.agents[agent_key]["handler"] = handler
            self.agents[agent_key]["active"] = True
            logger.info(f"Agent {agent_key} registered to hub")

    def unregister_agent(self, agent_key: str):
        """에이전트 등록 해제"""
        if agent_key in self.agents:
            self.agents[agent_key]["active"] = False
            self.agents[agent_key]["handler"] = None
            logger.info(f"Agent {agent_key} unregistered from hub")

    def send_message(self, from_agent: str, to_agent: str,
                     message_type: MessageType, data: Dict[str, Any]) -> str:
        """
        에이전트 간 메시지 전송

        Args:
            from_agent: 발신 에이전트
            to_agent: 수신 에이전트
            message_type: 메시지 유형
            data: 메시지 데이터

        Returns:
            메시지 ID
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

        # 라우팅 테이블에 추가
        self.routing_table[to_agent].put(message)

        # 즉시 처리 (수신 에이전트가 활성 상태인 경우)
        if self.agents[to_agent]["active"] and self.agents[to_agent]["handler"]:
            threading.Thread(
                target=self._process_message,
                args=(to_agent, message),
                daemon=True
            ).start()

        self.stats["messages_routed"] += 1
        logger.debug(f"Message {message_id} routed from {from_agent} to {to_agent}")

        return message_id

    def broadcast_message(self, from_agent: str, message_type: MessageType,
                         data: Dict[str, Any], exclude: List[str] = None):
        """
        모든 에이전트에게 브로드캐스트

        Args:
            from_agent: 발신 에이전트
            message_type: 메시지 유형
            data: 메시지 데이터
            exclude: 제외할 에이전트 리스트
        """
        exclude = exclude or []

        for agent_key in self.agents:
            if agent_key != from_agent and agent_key not in exclude:
                if self.agents[agent_key]["active"]:
                    self.send_message(from_agent, agent_key, message_type, data)

    def request_collaboration(self, initiator: str, participants: List[str],
                             topic: str, context: Dict[str, Any]) -> str:
        """
        협업 요청

        Args:
            initiator: 협업 시작 에이전트
            participants: 참여 에이전트 리스트
            topic: 협업 주제
            context: 협업 컨텍스트

        Returns:
            협업 ID
        """
        collab_id = f"collab-{datetime.now().timestamp()}"

        collaboration = {
            "id": collab_id,
            "initiator": initiator,
            "participants": participants,
            "topic": topic,
            "context": context,
            "status": "active",
            "started_at": datetime.now().isoformat(),
            "responses": {},
            "result": None
        }

        self.collaborations[collab_id] = collaboration

        # 참여자들에게 협업 요청 전송
        for participant in participants:
            if participant != initiator:
                self.send_message(
                    initiator, participant,
                    MessageType.REQUEST,
                    {
                        "collaboration_id": collab_id,
                        "topic": topic,
                        "context": context,
                        "action": "collaborate"
                    }
                )

        logger.info(f"Collaboration {collab_id} initiated by {initiator}")

        return collab_id

    def submit_collaboration_response(self, collab_id: str, agent_key: str,
                                     response: Dict[str, Any]):
        """
        협업 응답 제출

        Args:
            collab_id: 협업 ID
            agent_key: 응답 에이전트
            response: 응답 내용
        """
        if collab_id in self.collaborations:
            self.collaborations[collab_id]["responses"][agent_key] = {
                "agent": agent_key,
                "response": response,
                "timestamp": datetime.now().isoformat()
            }

            # 모든 참여자가 응답했는지 확인
            collab = self.collaborations[collab_id]
            if len(collab["responses"]) == len(collab["participants"]) - 1:
                self._finalize_collaboration(collab_id)

    def _finalize_collaboration(self, collab_id: str):
        """협업 완료 처리"""
        collab = self.collaborations[collab_id]
        collab["status"] = "completed"
        collab["completed_at"] = datetime.now().isoformat()

        # 결과 종합 (간단한 합의 도출)
        responses = collab["responses"]

        # 응답 분석 및 결과 도출
        result = {
            "consensus": self._find_consensus(responses),
            "summary": self._summarize_responses(responses),
            "participants": list(responses.keys())
        }

        collab["result"] = result

        # 발의자에게 결과 전송
        self.send_message(
            "hub", collab["initiator"],
            MessageType.DECISION,
            {
                "collaboration_id": collab_id,
                "result": result
            }
        )

        self.stats["collaborations_completed"] += 1
        logger.info(f"Collaboration {collab_id} completed")

    def start_voting(self, topic: str, options: List[str],
                    voters: List[str], timeout: int = 60) -> str:
        """
        투표 시작

        Args:
            topic: 투표 주제
            options: 선택지
            voters: 투표자 리스트
            timeout: 제한 시간 (초)

        Returns:
            투표 ID
        """
        vote_id = f"vote-{datetime.now().timestamp()}"

        voting_session = {
            "id": vote_id,
            "topic": topic,
            "options": options,
            "voters": voters,
            "votes": {},
            "status": "active",
            "started_at": datetime.now(),
            "timeout": timeout
        }

        self.voting_sessions[vote_id] = voting_session

        # 투표 요청 전송
        for voter in voters:
            self.send_message(
                "hub", voter,
                MessageType.VOTE,
                {
                    "vote_id": vote_id,
                    "topic": topic,
                    "options": options,
                    "timeout": timeout
                }
            )

        # 타임아웃 스레드
        threading.Thread(
            target=self._voting_timeout,
            args=(vote_id,),
            daemon=True
        ).start()

        return vote_id

    def cast_vote(self, vote_id: str, voter: str, choice: str):
        """투표"""
        if vote_id in self.voting_sessions:
            session = self.voting_sessions[vote_id]
            if session["status"] == "active" and voter in session["voters"]:
                session["votes"][voter] = choice

                # 모든 투표 완료 확인
                if len(session["votes"]) == len(session["voters"]):
                    self._finalize_voting(vote_id)

    def _finalize_voting(self, vote_id: str):
        """투표 완료 처리"""
        session = self.voting_sessions[vote_id]
        session["status"] = "completed"

        # 결과 집계
        vote_counts = defaultdict(int)
        for choice in session["votes"].values():
            vote_counts[choice] += 1

        # 승자 결정
        winner = max(vote_counts, key=vote_counts.get)

        result = {
            "winner": winner,
            "counts": dict(vote_counts),
            "participation": f"{len(session['votes'])}/{len(session['voters'])}"
        }

        session["result"] = result

        # 모든 참여자에게 결과 전송
        for voter in session["voters"]:
            self.send_message(
                "hub", voter,
                MessageType.DECISION,
                {
                    "vote_id": vote_id,
                    "result": result
                }
            )

        self.stats["decisions_made"] += 1

    def _voting_timeout(self, vote_id: str):
        """투표 타임아웃 처리"""
        session = self.voting_sessions.get(vote_id)
        if not session:
            return

        time.sleep(session["timeout"])

        if session["status"] == "active":
            self._finalize_voting(vote_id)

    def _process_message(self, agent_key: str, message: Dict[str, Any]):
        """메시지 처리"""
        try:
            handler = self.agents[agent_key]["handler"]
            if handler:
                handler(message)
        except Exception as e:
            logger.error(f"Error processing message for {agent_key}: {e}")

    def _find_consensus(self, responses: Dict[str, Any]) -> Optional[str]:
        """응답에서 합의점 찾기"""
        # 간단한 구현: 가장 많이 언급된 키워드 찾기
        keywords = defaultdict(int)

        for agent_response in responses.values():
            response_text = str(response_response.get("response", ""))
            # 키워드 추출 (간단한 구현)
            words = response_text.lower().split()
            for word in words:
                if len(word) > 3:
                    keywords[word] += 1

        if keywords:
            consensus_word = max(keywords, key=keywords.get)
            return f"합의: {consensus_word} (언급 횟수: {keywords[consensus_word]})"

        return "명확한 합의점을 찾을 수 없음"

    def _summarize_responses(self, responses: Dict[str, Any]) -> str:
        """응답 요약"""
        summary_parts = []

        for agent_key, response_data in responses.items():
            agent_name = self.agents[agent_key]["name"]
            response = response_data.get("response", {})

            # 응답에서 핵심 추출
            if isinstance(response, dict):
                key_point = response.get("key_point", response.get("summary", "응답 없음"))
            else:
                key_point = str(response)[:100]

            summary_parts.append(f"- {agent_name}: {key_point}")

        return "\n".join(summary_parts)

    def get_hub_status(self) -> Dict[str, Any]:
        """허브 상태 반환"""
        return {
            "timestamp": datetime.now().isoformat(),
            "active_agents": [
                key for key, info in self.agents.items()
                if info["active"]
            ],
            "pending_messages": {
                agent: self.routing_table[agent].qsize()
                for agent in self.agents
            },
            "active_collaborations": len([
                c for c in self.collaborations.values()
                if c["status"] == "active"
            ]),
            "active_votes": len([
                v for v in self.voting_sessions.values()
                if v["status"] == "active"
            ]),
            "stats": self.stats
        }


# 싱글톤 인스턴스
_hub_instance: Optional[AgentHub] = None


def get_hub(project_root: str = None) -> AgentHub:
    """싱글톤 Hub 인스턴스 반환"""
    global _hub_instance

    if _hub_instance is None:
        if project_root is None:
            from pathlib import Path
            project_root = str(Path(__file__).resolve().parent.parent)
        _hub_instance = AgentHub(project_root)

    return _hub_instance