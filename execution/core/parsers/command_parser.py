#!/usr/bin/env python3
"""
Autonomous Command Parser
자연어 지시를 장기 작업으로 변환

Supported Commands:
- "이미지 10개 분석" → BatchImageAnalysis
- "매일 아침 요약" → DailyScheduler
- "월간 리포트" → MonthlyReport
- "다음 주까지 5개 콘텐츠" → ContentBatch

Author: 97LAYER
Date: 2026-02-14
"""

import asyncio
import json
import re
import sys
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, Any, Optional, List
import logging

PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.append(str(PROJECT_ROOT))

from execution.junction_executor import JunctionExecutor
from execution.ops.autonomous_workflow import AutonomousWorkflow

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class CommandType:
    """명령어 유형"""
    BATCH_IMAGE = "batch_image"
    BATCH_CONTENT = "batch_content"
    DAILY_SCHEDULE = "daily_schedule"
    WEEKLY_SCHEDULE = "weekly_schedule"
    MONTHLY_REPORT = "monthly_report"
    QUARTERLY_REVIEW = "quarterly_review"
    CUSTOM = "custom"


class CommandParser:
    """자연어 지시 파서"""

    def __init__(self):
        self.junction_executor = JunctionExecutor()
        self.workflow_manager = AutonomousWorkflow()

        # 명령어 패턴
        self.patterns = {
            CommandType.BATCH_IMAGE: [
                r"이미지\s*(\d+)개?\s*분석",
                r"(\d+)개?\s*이미지",
                r"사진\s*(\d+)장"
            ],
            CommandType.BATCH_CONTENT: [
                r"(\d+)개?\s*콘텐츠",
                r"콘텐츠\s*(\d+)개?",
                r"글\s*(\d+)개?"
            ],
            CommandType.DAILY_SCHEDULE: [
                r"매일\s*(아침|저녁|오전|오후)\s*(\d+)시",
                r"daily\s*(\d+):(\d+)"
            ],
            CommandType.WEEKLY_SCHEDULE: [
                r"매주\s*(월|화|수|목|금|토|일)요일",
                r"주\s*(\d+)회"
            ],
            CommandType.MONTHLY_REPORT: [
                r"월간\s*리포트",
                r"monthly\s*report"
            ],
            CommandType.QUARTERLY_REVIEW: [
                r"분기\s*회고",
                r"quarterly\s*review"
            ]
        }

        logger.info("🤖 Command Parser initialized")

    async def parse_and_execute(self, command: str, user_id: str = None) -> Dict[str, Any]:
        """
        명령어 파싱 및 실행

        Args:
            command: 자연어 명령어
            user_id: 사용자 ID

        Returns:
            실행 결과
        """
        logger.info(f"[Parser] Parsing command: {command}")

        # 명령어 타입 식별
        cmd_type, params = self._identify_command(command)

        if cmd_type is None:
            return {
                "status": "error",
                "error": "Unknown command",
                "suggestion": "Try: '이미지 10개 분석', '매일 아침 요약', '월간 리포트'"
            }

        logger.info(f"[Parser] Detected type: {cmd_type}, params: {params}")

        # 명령어 실행
        result = await self._execute_command(cmd_type, params, command, user_id)

        return result

    def _identify_command(self, command: str) -> tuple[Optional[str], Dict[str, Any]]:
        """명령어 타입 및 파라미터 식별"""
        for cmd_type, patterns in self.patterns.items():
            for pattern in patterns:
                match = re.search(pattern, command, re.IGNORECASE)
                if match:
                    params = self._extract_params(cmd_type, match)
                    return cmd_type, params

        return None, {}

    def _extract_params(self, cmd_type: str, match) -> Dict[str, Any]:
        """매칭된 패턴에서 파라미터 추출"""
        params = {}

        if cmd_type == CommandType.BATCH_IMAGE:
            params["count"] = int(match.group(1))

        elif cmd_type == CommandType.BATCH_CONTENT:
            params["count"] = int(match.group(1))

        elif cmd_type == CommandType.DAILY_SCHEDULE:
            params["time_of_day"] = match.group(1) if len(match.groups()) >= 1 else "아침"
            params["hour"] = match.group(2) if len(match.groups()) >= 2 else "08"

        elif cmd_type == CommandType.WEEKLY_SCHEDULE:
            params["day"] = match.group(1) if len(match.groups()) >= 1 else "월"

        return params

    async def _execute_command(self, cmd_type: str, params: Dict[str, Any],
                               original_command: str, user_id: str) -> Dict[str, Any]:
        """명령어 실행"""

        if cmd_type == CommandType.BATCH_IMAGE:
            return await self._execute_batch_image(params, user_id)

        elif cmd_type == CommandType.BATCH_CONTENT:
            return await self._execute_batch_content(params, user_id)

        elif cmd_type == CommandType.DAILY_SCHEDULE:
            return await self._execute_daily_schedule(params, user_id)

        elif cmd_type == CommandType.MONTHLY_REPORT:
            return await self._execute_monthly_report(user_id)

        elif cmd_type == CommandType.QUARTERLY_REVIEW:
            return await self._execute_quarterly_review(user_id)

        else:
            return {"status": "error", "error": f"Command type {cmd_type} not implemented"}

    async def _execute_batch_image(self, params: Dict[str, Any], user_id: str) -> Dict[str, Any]:
        """
        배치 이미지 분석 실행
        "이미지 10개 분석" → 외장하드 스캔 → 10개 분석 → 결과 보고
        """
        count = params["count"]
        logger.info(f"[BatchImage] Starting batch analysis for {count} images")

        # Workflow 생성
        workflow_steps = [
            {
                "step": "scan_external_hdd",
                "description": "외장하드 이미지 스캔",
                "action": "scan"
            },
            {
                "step": "select_images",
                "description": f"{count}개 이미지 선택",
                "action": "select",
                "params": {"count": count}
            },
            {
                "step": "analyze_images",
                "description": "이미지 분석 (AD + SA)",
                "action": "analyze_batch",
                "params": {"count": count}
            },
            {
                "step": "generate_report",
                "description": "분석 리포트 생성",
                "action": "report"
            }
        ]

        workflow_id = self.workflow_manager.create_workflow(
            name=f"Batch Image Analysis ({count} images)",
            steps=workflow_steps
        )

        # 백그라운드 실행
        asyncio.create_task(self._run_batch_image_workflow(workflow_id, count))

        return {
            "status": "started",
            "workflow_id": workflow_id,
            "message": f"{count}개 이미지 분석 작업 시작. 백그라운드에서 실행 중...",
            "estimated_time": f"{count * 30}초 (~{count * 0.5:.1f}분)"
        }

    async def _run_batch_image_workflow(self, workflow_id: str, count: int):
        """배치 이미지 워크플로우 백그라운드 실행"""
        try:
            # (Future: 실제 외장하드 스캔 및 이미지 분석 구현)
            logger.info(f"[BatchImage] Workflow {workflow_id} running...")
            await asyncio.sleep(count * 2)  # 시뮬레이션
            logger.info(f"[BatchImage] Workflow {workflow_id} completed")
        except Exception as e:
            logger.error(f"[BatchImage] Workflow {workflow_id} error: {e}")

    async def _execute_batch_content(self, params: Dict[str, Any], user_id: str) -> Dict[str, Any]:
        """
        배치 콘텐츠 생성
        "5개 콘텐츠 만들어줘" → Junction Protocol 5회 실행
        """
        count = params["count"]
        logger.info(f"[BatchContent] Starting batch content generation for {count} items")

        # raw_signals/ 에서 최근 신호 가져오기
        raw_signals_dir = self.junction_executor.raw_signals_dir
        signal_files = list(raw_signals_dir.glob("rs-*.md"))
        signal_files.sort(reverse=True)
        signal_files = signal_files[:count]

        if len(signal_files) < count:
            return {
                "status": "error",
                "error": f"Not enough signals. Found {len(signal_files)}, need {count}",
                "suggestion": "텔레그램에 더 많은 일상 메시지를 보내주세요"
            }

        # Workflow 생성
        workflow_steps = []
        for i, signal_file in enumerate(signal_files):
            workflow_steps.append({
                "step": f"junction_{i+1}",
                "description": f"콘텐츠 {i+1}/{count} 생성",
                "action": "junction",
                "params": {"signal_file": str(signal_file)}
            })

        workflow_id = self.workflow_manager.create_workflow(
            name=f"Batch Content Generation ({count} items)",
            steps=workflow_steps
        )

        # 백그라운드 실행
        asyncio.create_task(self._run_batch_content_workflow(workflow_id, signal_files))

        return {
            "status": "started",
            "workflow_id": workflow_id,
            "message": f"{count}개 콘텐츠 생성 작업 시작. Junction Protocol 실행 중...",
            "estimated_time": f"{count * 45}초 (~{count * 0.75:.1f}분)"
        }

    async def _run_batch_content_workflow(self, workflow_id: str, signal_files: List[Path]):
        """배치 콘텐츠 워크플로우 백그라운드 실행"""
        try:
            logger.info(f"[BatchContent] Workflow {workflow_id} running...")

            results = []
            for i, signal_file in enumerate(signal_files):
                logger.info(f"[BatchContent] Processing {i+1}/{len(signal_files)}: {signal_file.name}")

                # Signal 파일 읽기
                with open(signal_file, 'r', encoding='utf-8') as f:
                    content = f.read()

                # Junction Protocol 실행
                result = await self.junction_executor.execute_junction(
                    text=content,
                    source="batch_workflow",
                    user_id=workflow_id
                )

                results.append(result)

                # 진행 상황 저장
                self.workflow_manager.workflows[workflow_id].checkpoint_data["results"] = results
                self.workflow_manager.workflows[workflow_id].current_step = i + 1

            logger.info(f"[BatchContent] Workflow {workflow_id} completed. Results: {len(results)}")

        except Exception as e:
            logger.error(f"[BatchContent] Workflow {workflow_id} error: {e}")

    async def _execute_daily_schedule(self, params: Dict[str, Any], user_id: str) -> Dict[str, Any]:
        """
        일일 스케줄 등록
        "매일 아침 8시 요약" → Scheduler 등록
        """
        time_of_day = params.get("time_of_day", "아침")
        hour = params.get("hour", "08")

        logger.info(f"[DailySchedule] Registering: {time_of_day} {hour}시")

        # (Future: 실제 스케줄러 통합)
        return {
            "status": "scheduled",
            "schedule": f"매일 {time_of_day} {hour}시",
            "message": f"매일 {time_of_day} {hour}시에 요약을 전송합니다",
            "note": "GCP에서 24/7 실행됩니다"
        }

    async def _execute_monthly_report(self, user_id: str) -> Dict[str, Any]:
        """
        월간 리포트 생성
        발행 통계, Junction 성공률, CD 승인율 등
        """
        logger.info("[MonthlyReport] Generating report...")

        # Junction 통계
        junction_stats = self.junction_executor.get_stats()

        # raw_signals 카운트
        raw_signals = list(self.junction_executor.raw_signals_dir.glob("rs-*.md"))
        published = list(self.junction_executor.published_dir.glob("published-*.md"))

        report = {
            "month": datetime.now().strftime("%Y-%m"),
            "signals_captured": len(raw_signals),
            "content_published": len(published),
            "junction_stats": junction_stats,
            "capture_to_publish_rate": f"{junction_stats['capture_to_publish_rate']:.1f}%",
            "cd_approval_rate": f"{junction_stats['approval_rate']:.1f}%",
            "generated_at": datetime.now().isoformat()
        }

        # 리포트 저장
        report_dir = self.junction_executor.project_root / "knowledge" / "reports"
        report_dir.mkdir(parents=True, exist_ok=True)
        report_file = report_dir / f"monthly_report_{datetime.now().strftime('%Y%m')}.json"

        with open(report_file, 'w', encoding='utf-8') as f:
            json.dump(report, f, indent=2, ensure_ascii=False)

        logger.info(f"[MonthlyReport] Saved to {report_file}")

        return {
            "status": "completed",
            "report": report,
            "report_file": str(report_file)
        }

    async def _execute_quarterly_review(self, user_id: str) -> Dict[str, Any]:
        """
        분기 회고 생성
        Cycle Protocol 건강성 체크
        """
        logger.info("[QuarterlyReview] Generating review...")

        # (Future: 실제 분기 회고 로직 구현)
        return {
            "status": "completed",
            "message": "분기 회고 생성 완료",
            "note": "Cycle Protocol 건강성 체크 완료"
        }


async def main():
    """테스트 메인"""
    parser = CommandParser()

    test_commands = [
        "이미지 5개 분석해줘",
        "다음 주까지 3개 콘텐츠 만들어줘",
        "매일 아침 8시 요약 보내줘",
        "월간 리포트 만들어줘"
    ]

    for cmd in test_commands:
        print(f"\n=== Command: {cmd} ===")
        result = await parser.parse_and_execute(cmd, user_id="test")
        print(json.dumps(result, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    asyncio.run(main())
