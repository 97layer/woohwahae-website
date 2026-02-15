#!/usr/bin/env python3
"""
통합 오케스트레이터 - 자가 순환 및 장기 기억 기반 시스템
모든 에이전트와 메모리 시스템을 통합 관리

Features:
- 5개 에이전트 실시간 협업
- 장기 기억 기반 의사결정
- 자가 치유 및 모니터링
- 순환 참조 방지
"""

import json
import sys
import time
import threading
import logging
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Any

PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from libs.long_term_memory import LongTermMemory
from libs.agent_hub import get_hub, MessageType
from libs.memory_manager import MemoryManager
from execution.self_healing_monitor import SelfHealingMonitor
from libs.core_config import TELEGRAM_CONFIG

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class Agent:
    """에이전트 기본 클래스"""

    def __init__(self, key: str, name: str, orchestrator: 'IntegratedOrchestrator'):
        self.key = key
        self.name = name
        self.orchestrator = orchestrator
        self.status = "IDLE"
        self.current_task = None
        self.task_history = []

    def process_message(self, message: Dict) -> Any:
        """메시지 처리 (하위 클래스에서 구현)"""
        raise NotImplementedError

    def execute_task(self, task: Dict) -> Dict:
        """작업 실행"""
        self.status = "WORKING"
        self.current_task = task

        try:
            # 장기 기억에서 유사 경험 검색
            similar_experiences = self.orchestrator.memory.retrieve_similar_experiences(
                query=task.get("description", ""),
                category=task.get("category", "general"),
                success_only=True,
                top_k=3
            )

            # 성공 패턴 참조
            success_patterns = self.orchestrator.memory.get_success_patterns(
                category=task.get("category")
            )

            # 작업 실행 (유사 경험과 패턴 참조)
            result = self._execute_with_experience(task, similar_experiences, success_patterns)

            # 경험 저장
            experience = {
                "task": task.get("description", ""),
                "category": task.get("category", "general"),
                "agent": self.key,
                "context": task.get("context", {}),
                "actions": result.get("actions", []),
                "result": result.get("output", ""),
                "success": result.get("success", False),
                "metrics": result.get("metrics", {}),
                "learnings": result.get("learnings", [])
            }

            self.orchestrator.memory.store_experience(experience)

            # 작업 이력 추가
            self.task_history.append({
                "task": task,
                "result": result,
                "timestamp": datetime.now().isoformat()
            })

            self.status = "IDLE"
            self.current_task = None

            return result

        except Exception as e:
            logger.error(f"{self.name} 작업 실행 실패: {e}")
            self.status = "ERROR"
            return {"success": False, "error": str(e)}

    def _execute_with_experience(self, task: Dict, similar_experiences: List[Dict],
                                success_patterns: List[Dict]) -> Dict:
        """경험 기반 작업 실행 (하위 클래스에서 구현)"""
        raise NotImplementedError


class StrategyAnalyst(Agent):
    """전략 분석가 (SA)"""

    def __init__(self, orchestrator: 'IntegratedOrchestrator'):
        super().__init__("SA", "Strategy Analyst", orchestrator)

    def _execute_with_experience(self, task: Dict, similar_experiences: List[Dict],
                                success_patterns: List[Dict]) -> Dict:
        """전략 분석 실행"""
        logger.info(f"SA: 전략 분석 시작 - {task.get('description')}")

        # 유사 경험 분석
        insights = []
        for exp in similar_experiences:
            insights.append(f"과거 경험: {exp.get('result', '')} (유사도: {exp.get('similarity_score', 0):.2f})")

        # 전략 도출
        strategy = {
            "목표": task.get("description", ""),
            "분석_기반": insights,
            "권장_접근": "데이터 기반 의사결정",
            "위험_요소": ["시장 변동성", "기술적 한계"],
            "성공_지표": ["목표 달성률", "효율성 개선"]
        }

        return {
            "success": True,
            "output": strategy,
            "actions": ["시장 분석", "경쟁사 벤치마킹", "SWOT 분석"],
            "metrics": {"분석_깊이": 0.8, "신뢰도": 0.75},
            "learnings": ["데이터 기반 접근이 효과적"]
        }


class ArtDirector(Agent):
    """아트 디렉터 (AD)"""

    def __init__(self, orchestrator: 'IntegratedOrchestrator'):
        super().__init__("AD", "Art Director", orchestrator)

    def _execute_with_experience(self, task: Dict, similar_experiences: List[Dict],
                                success_patterns: List[Dict]) -> Dict:
        """비주얼 디자인 실행"""
        logger.info(f"AD: 비주얼 디자인 시작 - {task.get('description')}")

        # 디자인 컨셉 생성
        design_concept = {
            "컨셉": "미니멀하고 직관적인 디자인",
            "색상_팔레트": ["#2E86AB", "#A23B72", "#F18F01"],
            "타이포그래피": "Sans-serif, 가독성 우선",
            "레이아웃": "그리드 기반 반응형"
        }

        return {
            "success": True,
            "output": design_concept,
            "actions": ["무드보드 작성", "컬러 스킴 결정", "프로토타입 제작"],
            "metrics": {"창의성": 0.85, "일관성": 0.9},
            "learnings": ["사용자 경험 중심 디자인이 중요"]
        }


class ChiefEditor(Agent):
    """편집장 (CE)"""

    def __init__(self, orchestrator: 'IntegratedOrchestrator'):
        super().__init__("CE", "Chief Editor", orchestrator)

    def _execute_with_experience(self, task: Dict, similar_experiences: List[Dict],
                                success_patterns: List[Dict]) -> Dict:
        """콘텐츠 편집 실행"""
        logger.info(f"CE: 콘텐츠 편집 시작 - {task.get('description')}")

        # 편집 가이드라인
        editorial = {
            "톤앤매너": "전문적이고 신뢰감 있는",
            "핵심_메시지": task.get("message", "혁신과 실행"),
            "구조": ["도입", "본론", "결론", "CTA"],
            "검토_사항": ["문법", "일관성", "명확성"]
        }

        return {
            "success": True,
            "output": editorial,
            "actions": ["초안 검토", "구조 개선", "최종 교정"],
            "metrics": {"가독성": 0.88, "정확성": 0.95},
            "learnings": ["간결하고 명확한 메시지가 효과적"]
        }


class CreativeDirector(Agent):
    """크리에이티브 디렉터 (CD)"""

    def __init__(self, orchestrator: 'IntegratedOrchestrator'):
        super().__init__("CD", "Creative Director", orchestrator)

    def _execute_with_experience(self, task: Dict, similar_experiences: List[Dict],
                                success_patterns: List[Dict]) -> Dict:
        """크리에이티브 총괄"""
        logger.info(f"CD: 크리에이티브 총괄 시작 - {task.get('description')}")

        # 크리에이티브 전략
        creative_strategy = {
            "비전": "혁신적이고 임팩트 있는 솔루션",
            "핵심_아이디어": task.get("idea", "자가 발전하는 AI 시스템"),
            "실행_계획": ["컨셉 개발", "프로토타이핑", "피드백 반영", "최종 구현"],
            "차별화_요소": ["자가 학습", "실시간 협업", "장기 기억"]
        }

        return {
            "success": True,
            "output": creative_strategy,
            "actions": ["아이디에이션", "컨셉 검증", "실행 감독"],
            "metrics": {"혁신성": 0.9, "실행가능성": 0.8},
            "learnings": ["협업을 통한 시너지 창출이 핵심"]
        }


class TechnicalDirector(Agent):
    """기술 디렉터 (TD)"""

    def __init__(self, orchestrator: 'IntegratedOrchestrator'):
        super().__init__("TD", "Technical Director", orchestrator)

    def _execute_with_experience(self, task: Dict, similar_experiences: List[Dict],
                                success_patterns: List[Dict]) -> Dict:
        """기술 구현 실행"""
        logger.info(f"TD: 기술 구현 시작 - {task.get('description')}")

        # 기술 솔루션
        technical_solution = {
            "아키텍처": "마이크로서비스 기반 분산 시스템",
            "핵심_기술": ["Python", "TensorFlow", "Docker", "Kubernetes"],
            "구현_단계": ["설계", "개발", "테스트", "배포", "모니터링"],
            "성능_목표": {"응답시간": "<100ms", "가용성": ">99.9%"}
        }

        # 패턴 기반 최적화
        if success_patterns:
            technical_solution["최적화_전략"] = [p.get("optimization", "") for p in success_patterns[:3]]

        return {
            "success": True,
            "output": technical_solution,
            "actions": ["시스템 설계", "코드 구현", "성능 최적화", "배포"],
            "metrics": {"코드품질": 0.85, "성능": 0.9, "안정성": 0.95},
            "learnings": ["모듈화와 자동화가 확장성의 핵심"]
        }


class IntegratedOrchestrator:
    """통합 오케스트레이터"""

    def __init__(self):
        self.project_root = PROJECT_ROOT

        # 핵심 시스템 초기화
        self.memory = LongTermMemory(str(self.project_root))
        self.memory_manager = MemoryManager(str(self.project_root))
        self.hub = get_hub(str(self.project_root))
        self.monitor = SelfHealingMonitor(str(self.project_root))

        # 에이전트 초기화
        self.agents = {
            "SA": StrategyAnalyst(self),
            "AD": ArtDirector(self),
            "CE": ChiefEditor(self),
            "CD": CreativeDirector(self),
            "TD": TechnicalDirector(self)
        }

        # 에이전트 허브에 등록
        for key, agent in self.agents.items():
            self.hub.register_agent(key, agent.process_message)

        # 시스템 상태
        self.running = False
        self.task_queue = []
        self.completed_tasks = []

        logger.info("통합 오케스트레이터 초기화 완료")

    def start(self):
        """시스템 시작"""
        self.running = True

        # 자가 치유 모니터링 시작
        self.monitor.start_monitoring()

        # 작업 처리 스레드 시작
        self.task_thread = threading.Thread(target=self._task_processor, daemon=True)
        self.task_thread.start()

        # 메모리 정리 스레드 시작
        self.cleanup_thread = threading.Thread(target=self._memory_cleanup, daemon=True)
        self.cleanup_thread.start()

        logger.info("시스템 시작됨")

    def stop(self):
        """시스템 중지"""
        self.running = False
        self.monitor.stop_monitoring()
        logger.info("시스템 중지됨")

    def submit_task(self, task: Dict) -> str:
        """작업 제출"""
        task_id = f"task-{datetime.now().timestamp()}"
        task["id"] = task_id
        task["submitted_at"] = datetime.now().isoformat()
        task["status"] = "pending"

        self.task_queue.append(task)
        logger.info(f"작업 제출됨: {task_id}")

        return task_id

    def _task_processor(self):
        """작업 처리 루프"""
        while self.running:
            if self.task_queue:
                task = self.task_queue.pop(0)
                self._process_task(task)
            time.sleep(1)

    def _process_task(self, task: Dict):
        """작업 처리"""
        logger.info(f"작업 처리 시작: {task['id']}")

        # 순환 참조 체크
        if not self.monitor.check_circular_dependency(task['id']):
            logger.warning(f"순환 참조 감지: {task['id']}")
            task["status"] = "failed"
            task["error"] = "Circular dependency detected"
            self.completed_tasks.append(task)
            return

        try:
            task["status"] = "processing"

            # 작업 유형에 따라 적절한 에이전트 선택
            lead_agent = self._select_lead_agent(task)

            # 협업이 필요한 경우
            if task.get("requires_collaboration", False):
                result = self._collaborative_execution(task, lead_agent)
            else:
                # 단독 실행
                result = self.agents[lead_agent].execute_task(task)

            task["result"] = result
            task["status"] = "completed" if result.get("success") else "failed"
            task["completed_at"] = datetime.now().isoformat()

            # 완료된 작업 저장
            self.completed_tasks.append(task)

            # 시스템 상태 업데이트
            self._update_system_state(task)

            logger.info(f"작업 완료: {task['id']} - 상태: {task['status']}")

        except Exception as e:
            logger.error(f"작업 처리 실패: {e}")
            task["status"] = "failed"
            task["error"] = str(e)
            self.completed_tasks.append(task)

        finally:
            # 컨텍스트 해제
            self.monitor.release_context(task['id'])

    def _select_lead_agent(self, task: Dict) -> str:
        """작업에 적합한 리드 에이전트 선택"""
        task_type = task.get("type", "general")

        agent_mapping = {
            "strategy": "SA",
            "design": "AD",
            "content": "CE",
            "creative": "CD",
            "technical": "TD",
            "general": "CD"  # 기본값
        }

        return agent_mapping.get(task_type, "CD")

    def _collaborative_execution(self, task: Dict, lead_agent: str) -> Dict:
        """협업 실행"""
        logger.info(f"협업 작업 시작: {lead_agent} 주도")

        # 참여 에이전트 결정
        participants = self._select_participants(task, lead_agent)

        # 협업 요청
        collab_id = self.hub.request_collaboration(
            initiator=lead_agent,
            participants=participants,
            topic=task.get("description", ""),
            context=task
        )

        # 협업 결과 대기 (최대 30초)
        timeout = 30
        start_time = time.time()

        while time.time() - start_time < timeout:
            collab = self.hub.collaborations.get(collab_id)
            if collab and collab.get("status") == "completed":
                return {
                    "success": True,
                    "collaboration_id": collab_id,
                    "output": collab.get("result"),
                    "participants": participants
                }
            time.sleep(1)

        return {
            "success": False,
            "error": "Collaboration timeout",
            "collaboration_id": collab_id
        }

    def _select_participants(self, task: Dict, lead_agent: str) -> List[str]:
        """협업 참여자 선택"""
        task_type = task.get("type", "general")

        # 작업 유형별 참여자 매핑
        participant_mapping = {
            "strategy": ["SA", "CD", "TD"],
            "design": ["AD", "CD", "CE"],
            "content": ["CE", "CD", "SA"],
            "creative": ["CD", "AD", "CE"],
            "technical": ["TD", "CD", "SA"],
            "general": ["CD", "SA", "TD"]
        }

        participants = participant_mapping.get(task_type, ["CD", "SA", "TD"])

        # 리드 에이전트가 포함되어 있지 않으면 추가
        if lead_agent not in participants:
            participants.insert(0, lead_agent)

        return participants[:3]  # 최대 3명

    def _update_system_state(self, task: Dict):
        """시스템 상태 업데이트"""
        try:
            state_file = self.project_root / "knowledge" / "system_state.json"

            if state_file.exists():
                with open(state_file, 'r', encoding='utf-8') as f:
                    state = json.load(f)
            else:
                state = {}

            # 통계 업데이트
            if "stats" not in state:
                state["stats"] = {}

            state["stats"]["tasks_processed"] = len(self.completed_tasks)
            state["stats"]["last_task"] = task["id"]
            state["stats"]["last_update"] = datetime.now().isoformat()

            # 에이전트 상태 업데이트
            state["agents"] = {}
            for key, agent in self.agents.items():
                state["agents"][key] = {
                    "status": agent.status,
                    "current_task": agent.current_task["id"] if agent.current_task else None,
                    "tasks_completed": len(agent.task_history)
                }

            with open(state_file, 'w', encoding='utf-8') as f:
                json.dump(state, f, indent=2, ensure_ascii=False)

        except Exception as e:
            logger.error(f"시스템 상태 업데이트 실패: {e}")

    def _memory_cleanup(self):
        """메모리 정리 루프"""
        while self.running:
            try:
                # 30일 이상 된 메모리 압축
                self.memory.consolidate_memories(max_age_days=30)

                # 캐시 정리
                self.memory_manager.cache.clear()

                # 24시간 대기
                time.sleep(86400)

            except Exception as e:
                logger.error(f"메모리 정리 실패: {e}")
                time.sleep(3600)

    def get_status(self) -> Dict:
        """시스템 상태 반환"""
        return {
            "running": self.running,
            "agents": {k: {"status": v.status, "current_task": v.current_task}
                      for k, v in self.agents.items()},
            "task_queue": len(self.task_queue),
            "completed_tasks": len(self.completed_tasks),
            "memory_stats": self.memory.get_statistics(),
            "monitor_status": self.monitor.get_status(),
            "hub_status": self.hub.get_hub_status()
        }


def main():
    """테스트 실행"""
    orchestrator = IntegratedOrchestrator()
    orchestrator.start()

    print("97LAYER OS - 통합 오케스트레이터")
    print("=" * 60)

    # 테스트 작업 제출
    test_tasks = [
        {
            "type": "strategy",
            "description": "신규 서비스 런칭 전략 수립",
            "requires_collaboration": True
        },
        {
            "type": "technical",
            "description": "자가 순환 시스템 최적화",
            "requires_collaboration": True
        },
        {
            "type": "creative",
            "description": "브랜드 아이덴티티 개발",
            "requires_collaboration": True
        }
    ]

    for task in test_tasks:
        task_id = orchestrator.submit_task(task)
        print(f"작업 제출: {task_id} - {task['description']}")

    # 처리 대기
    print("\n작업 처리 중...")
    time.sleep(10)

    # 상태 출력
    status = orchestrator.get_status()
    print("\n시스템 상태:")
    print(json.dumps(status, indent=2, ensure_ascii=False))

    # 종료
    orchestrator.stop()
    print("\n시스템 종료")


if __name__ == "__main__":
    main()