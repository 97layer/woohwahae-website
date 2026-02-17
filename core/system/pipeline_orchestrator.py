#!/usr/bin/env python3
"""
97layerOS Pipeline Orchestrator
SA → AD → CE → CD → ContentPublisher 자동 파이프라인

역할:
- SA 완료 태스크 감지 → AD 태스크 생성
- AD 완료 태스크 감지 → CE 태스크 생성
- CE 완료 태스크 감지 → CD(review) 태스크 생성
- CD 승인 → ContentPublisher 호출
- CD 거절 → CE 재작업 (max 2회)
- Ralph 품질 게이트 통합 (70+ 통과)

중복 방지: .infra/queue/orchestrated.json

Author: 97layerOS
Created: 2026-02-17
"""

import json
import logging
import asyncio
from pathlib import Path
from datetime import datetime
from typing import Dict, Optional, Any

logger = logging.getLogger(__name__)


class PipelineOrchestrator:
    """
    파이프라인 자동 흐름 관리자.
    각 에이전트는 독립적으로 동작, Orchestrator만 다음 단계 태스크를 생성.
    """

    PIPELINE_STAGES = {
        "SA": {
            "task_types": ["analyze_signal", "analyze"],
            "next_agent": "AD",
            "next_task_type": "create_visual_concept"
        },
        "AD": {
            "task_types": ["create_visual_concept"],
            "next_agent": "CE",
            "next_task_type": "write_content"
        },
        "CE": {
            "task_types": ["write_content"],
            "next_agent": "CD",
            "next_task_type": "review_content"
        }
    }

    MAX_CE_RETRIES = 2  # CD 거절 후 CE 재작업 최대 횟수
    RALPH_PASS_SCORE = 70  # Ralph 품질 게이트

    def __init__(self, base_path: Optional[Path] = None):
        self.base_path = base_path or Path("/app")
        self.queue_path = self.base_path / ".infra" / "queue"
        self.orchestrated_file = self.queue_path / "orchestrated.json"
        self.completed_path = self.queue_path / "tasks" / "completed"
        self.pending_path = self.queue_path / "tasks" / "pending"

        # 처리된 태스크 추적 (중복 방지)
        self._orchestrated: Dict[str, Any] = self._load_orchestrated()
        self._running = False

    def _load_orchestrated(self) -> Dict:
        """이미 처리한 task_id 목록 로드"""
        if self.orchestrated_file.exists():
            try:
                return json.loads(self.orchestrated_file.read_text())
            except Exception:
                return {}
        return {}

    def _save_orchestrated(self):
        """처리된 task_id 목록 저장"""
        self.orchestrated_file.parent.mkdir(parents=True, exist_ok=True)
        self.orchestrated_file.write_text(json.dumps(self._orchestrated, indent=2, ensure_ascii=False))

    def _is_orchestrated(self, task_id: str) -> bool:
        return task_id in self._orchestrated

    def _mark_orchestrated(self, task_id: str, next_task_id: str):
        self._orchestrated[task_id] = {
            "orchestrated_at": datetime.now().isoformat(),
            "next_task_id": next_task_id
        }
        self._save_orchestrated()

    def _create_task(self, agent_type: str, task_type: str, payload: Dict) -> str:
        """큐에 새 태스크 파일 생성"""
        import time
        task_id = f"{int(time.time() * 1000)}_{agent_type}_{task_type}"
        task = {
            "task_id": task_id,
            "agent_type": agent_type,
            "task_type": task_type,
            "payload": payload,
            "status": "pending",
            "created_at": datetime.now().isoformat(),
            "claimed_at": None,
            "completed_at": None,
            "claimed_by": None,
            "result": None,
            "error": None
        }
        self.pending_path.mkdir(parents=True, exist_ok=True)
        task_file = self.pending_path / f"{task_id}.json"
        task_file.write_text(json.dumps(task, indent=2, ensure_ascii=False))
        logger.info(f"[Orchestrator] 태스크 생성: {task_id}")
        return task_id

    def _load_completed_tasks(self):
        """완료된 태스크 파일 목록 로드"""
        if not self.completed_path.exists():
            return []
        tasks = []
        for f in sorted(self.completed_path.glob("*.json")):
            try:
                data = json.loads(f.read_text())
                tasks.append(data)
            except Exception as e:
                logger.warning(f"태스크 파일 읽기 실패: {f.name} - {e}")
        return tasks

    def _get_ce_retry_count(self, signal_id: str) -> int:
        """특정 signal_id에 대한 CE 재작업 횟수"""
        count = 0
        for v in self._orchestrated.values():
            if isinstance(v, dict) and v.get("signal_id") == signal_id and v.get("retry_for") == "CE":
                count += 1
        return count

    def _ralph_score(self, ce_result: Dict) -> int:
        """CE 결과물 Ralph 점수 계산 (간단 휴리스틱)"""
        score = 50  # 기본
        content = ce_result.get("result", {})

        # 인스타 캡션 존재 여부
        if content.get("instagram_caption"):
            caption = content["instagram_caption"]
            if len(caption) > 50:
                score += 10
            if len(caption) <= 300:
                score += 5

        # 해시태그 존재
        if content.get("hashtags"):
            score += 5

        # 아카이브 에세이 존재
        if content.get("archive_essay"):
            essay = content["archive_essay"]
            if 200 <= len(essay) <= 1500:
                score += 15
            elif len(essay) > 100:
                score += 8

        # SA 전략 점수 반영
        sa_score = content.get("sa_strategic_score", 0)
        if sa_score >= 80:
            score += 10
        elif sa_score >= 60:
            score += 5

        return min(score, 100)

    def process_completed_tasks(self):
        """완료된 태스크 스캔 → 다음 파이프라인 단계 생성"""
        tasks = self._load_completed_tasks()
        processed_count = 0

        for task in tasks:
            task_id = task.get("task_id", "")
            agent_type = task.get("agent_type", "")
            task_type = task.get("task_type", "")
            status = task.get("status", "")
            result = task.get("result", {}) or {}

            # 이미 처리된 태스크 스킵
            if self._is_orchestrated(task_id):
                continue

            # 실패한 태스크 스킵
            if status != "completed":
                continue

            # SA 완료 → AD 태스크 생성
            if agent_type == "SA" and task_type in ["analyze_signal", "analyze"]:
                next_task_id = self._handle_sa_completed(task, result)
                if next_task_id:
                    self._mark_orchestrated(task_id, next_task_id)
                    processed_count += 1
                else:
                    # 건너뜀 (낮은 점수 등) - 중복 방지용으로 마킹
                    self._mark_orchestrated(task_id, "skipped")
                    processed_count += 1

            # AD 완료 → CE 태스크 생성
            elif agent_type == "AD" and task_type == "create_visual_concept":
                next_task_id = self._handle_ad_completed(task, result)
                self._mark_orchestrated(task_id, next_task_id)
                processed_count += 1

            # CE 완료 → Ralph 검증 → CD 태스크 생성
            elif agent_type == "CE" and task_type == "write_content":
                next_task_id = self._handle_ce_completed(task, result)
                self._mark_orchestrated(task_id, next_task_id)
                processed_count += 1

            # CD 완료 → ContentPublisher 호출 or CE 재작업
            elif agent_type == "CD" and task_type == "review_content":
                next_task_id = self._handle_cd_completed(task, result)
                self._mark_orchestrated(task_id, next_task_id)
                processed_count += 1

        if processed_count > 0:
            logger.info(f"[Orchestrator] {processed_count}개 태스크 처리 완료")

        return processed_count

    def _handle_sa_completed(self, task: Dict, result: Dict) -> Optional[str]:
        """SA 완료 → AD 태스크 생성"""
        sa_result = result.get("result", {})
        strategic_score = sa_result.get("strategic_score", 0)

        # 전략 점수 50 미만이면 파이프라인 스킵
        if strategic_score < 50:
            logger.info(f"[Orchestrator] SA 점수 {strategic_score} < 50, 파이프라인 스킵: {task['task_id']}")
            return None

        signal_id = task.get("payload", {}).get("signal_id", "unknown")

        ad_payload = {
            "signal_id": signal_id,
            "sa_result": sa_result,
            "themes": sa_result.get("themes", []),
            "key_insights": sa_result.get("key_insights", []),
            "strategic_score": strategic_score,
            "source_task_id": task["task_id"]
        }

        logger.info(f"[Orchestrator] SA→AD: signal={signal_id}, score={strategic_score}")
        return self._create_task("AD", "create_visual_concept", ad_payload)

    def _handle_ad_completed(self, task: Dict, result: Dict) -> str:
        """AD 완료 → CE 태스크 생성"""
        ad_result = result.get("result", {})
        sa_result = task.get("payload", {}).get("sa_result", {})
        signal_id = task.get("payload", {}).get("signal_id", "unknown")

        ce_payload = {
            "signal_id": signal_id,
            "sa_result": sa_result,
            "ad_result": ad_result,
            "visual_concept": ad_result.get("visual_concept", {}),
            "themes": task.get("payload", {}).get("themes", []),
            "source_task_id": task["task_id"]
        }

        logger.info(f"[Orchestrator] AD→CE: signal={signal_id}")
        return self._create_task("CE", "write_content", ce_payload)

    def _handle_ce_completed(self, task: Dict, result: Dict) -> str:
        """CE 완료 → Ralph 점수 검증 → CD 태스크 생성"""
        ralph_score = self._ralph_score(result)
        signal_id = task.get("payload", {}).get("signal_id", "unknown")
        sa_result = task.get("payload", {}).get("sa_result", {})
        ad_result = task.get("payload", {}).get("ad_result", {})
        ce_result = result.get("result", {})

        logger.info(f"[Orchestrator] CE 완료, Ralph 점수: {ralph_score}/100 (signal={signal_id})")

        # 점수 낮으면 재작업 (최대 2회)
        retry_count = self._get_ce_retry_count(signal_id)
        if ralph_score < 50 and retry_count < self.MAX_CE_RETRIES:
            logger.info(f"[Orchestrator] Ralph {ralph_score}<50 → CE 재작업 ({retry_count+1}/{self.MAX_CE_RETRIES})")
            retry_payload = {
                **task.get("payload", {}),
                "retry_count": retry_count + 1,
                "ralph_score": ralph_score,
                "previous_output": ce_result,
                "feedback": f"품질 점수 {ralph_score}/100. 인스타 캡션(300자 이내)과 아카이브 에세이(500-800자)를 더 충실히 작성하세요."
            }
            new_task_id = self._create_task("CE", "write_content", retry_payload)
            # 재작업 추적
            self._orchestrated[f"retry_{signal_id}_{retry_count}"] = {
                "orchestrated_at": datetime.now().isoformat(),
                "next_task_id": new_task_id,
                "signal_id": signal_id,
                "retry_for": "CE"
            }
            self._save_orchestrated()
            return new_task_id

        # CD 리뷰 태스크 생성
        cd_payload = {
            "signal_id": signal_id,
            "sa_result": sa_result,
            "ad_result": ad_result,
            "ce_result": ce_result,
            "ralph_score": ralph_score,
            "instagram_caption": ce_result.get("instagram_caption", ""),
            "hashtags": ce_result.get("hashtags", ""),
            "archive_essay": ce_result.get("archive_essay", ""),
            "source_task_id": task["task_id"]
        }

        logger.info(f"[Orchestrator] CE→CD: signal={signal_id}, ralph={ralph_score}")
        return self._create_task("CD", "review_content", cd_payload)

    def _handle_cd_completed(self, task: Dict, result: Dict) -> str:
        """CD 완료 → 승인 시 ContentPublisher, 거절 시 CE 재작업"""
        cd_result = result.get("result", {})
        approved = cd_result.get("approved", False)
        feedback = cd_result.get("feedback", "")
        signal_id = task.get("payload", {}).get("signal_id", "unknown")
        ce_result = task.get("payload", {}).get("ce_result", {})

        if approved:
            logger.info(f"[Orchestrator] CD 승인! ContentPublisher 호출: signal={signal_id}")
            publisher_payload = {
                "signal_id": signal_id,
                "sa_result": task.get("payload", {}).get("sa_result", {}),
                "ad_result": task.get("payload", {}).get("ad_result", {}),
                "ce_result": ce_result,
                "cd_result": cd_result,
                "instagram_caption": task.get("payload", {}).get("instagram_caption", ""),
                "hashtags": task.get("payload", {}).get("hashtags", ""),
                "archive_essay": task.get("payload", {}).get("archive_essay", ""),
                "source_task_id": task["task_id"]
            }
            # ContentPublisher는 별도 태스크가 아닌 직접 호출
            self._trigger_content_publisher(publisher_payload)
            return f"published_{signal_id}"
        else:
            # CE 재작업 (피드백 포함)
            retry_count = self._get_ce_retry_count(signal_id)
            if retry_count >= self.MAX_CE_RETRIES:
                logger.warning(f"[Orchestrator] CE 재작업 최대 횟수 초과, 파이프라인 종료: signal={signal_id}")
                return f"max_retry_{signal_id}"

            logger.info(f"[Orchestrator] CD 거절 → CE 재작업: {feedback[:100]}")
            retry_payload = {
                **task.get("payload", {}),
                "retry_count": retry_count + 1,
                "cd_feedback": feedback,
                "previous_output": ce_result,
                "feedback": f"CD 검토 결과: {feedback}"
            }
            new_task_id = self._create_task("CE", "write_content", retry_payload)
            self._orchestrated[f"cd_retry_{signal_id}_{retry_count}"] = {
                "orchestrated_at": datetime.now().isoformat(),
                "next_task_id": new_task_id,
                "signal_id": signal_id,
                "retry_for": "CE"
            }
            self._save_orchestrated()
            return new_task_id

    def _trigger_content_publisher(self, payload: Dict):
        """ContentPublisher 직접 트리거 (비동기 태스크로 실행)"""
        try:
            from core.system.content_publisher import ContentPublisher
            publisher = ContentPublisher(base_path=self.base_path)
            # 비동기 실행 (블로킹 방지)
            import threading
            t = threading.Thread(target=publisher.publish, args=(payload,), daemon=True)
            t.start()
            logger.info(f"[Orchestrator] ContentPublisher 스레드 시작: {payload.get('signal_id')}")
        except ImportError:
            logger.warning("[Orchestrator] ContentPublisher 미구현 - 스킵")
        except Exception as e:
            logger.error(f"[Orchestrator] ContentPublisher 오류: {e}")

    async def run_forever(self, interval_seconds: int = 30):
        """무한 폴링 루프 (AgentWatcher 패턴)"""
        self._running = True
        logger.info(f"[Orchestrator] 시작. 폴링 간격: {interval_seconds}초")

        while self._running:
            try:
                count = self.process_completed_tasks()
                if count > 0:
                    logger.info(f"[Orchestrator] 이번 사이클: {count}개 처리")
            except Exception as e:
                logger.error(f"[Orchestrator] 오류: {e}", exc_info=True)

            await asyncio.sleep(interval_seconds)

    def stop(self):
        self._running = False
        logger.info("[Orchestrator] 정지")


def main():
    """독립 실행 진입점"""
    import sys
    # 로그 경로: 현재 작업 디렉토리 기준 (GCP VM: /home/.../97layerOS)
    log_dir = Path(".infra/logs")
    log_dir.mkdir(parents=True, exist_ok=True)

    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s [%(name)s] %(levelname)s: %(message)s',
        handlers=[
            logging.StreamHandler(sys.stdout),
            logging.FileHandler(log_dir / "orchestrator.log", mode='a')
        ]
    )

    # base_path: GCP VM은 실행 디렉토리, 컨테이너는 /app
    base = Path("/app") if Path("/app").exists() else Path(".")
    orchestrator = PipelineOrchestrator(base_path=base)
    asyncio.run(orchestrator.run_forever(interval_seconds=30))


if __name__ == "__main__":
    main()
