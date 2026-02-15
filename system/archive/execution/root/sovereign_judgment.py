#!/usr/bin/env python3
# Sovereign Judgment Execution Tool
# MBQ (Manifest Brand Quality) 승인 도구
# Author: 97LAYER
# Date: 2026-02-14

import os
import sys
import json
import logging
from pathlib import Path
from typing import Optional, Dict, Any, List
from datetime import datetime

# Path setup
PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.append(str(PROJECT_ROOT))

from libs.claude_engine import ClaudeEngine, DualAIEngine
from libs.memory_manager import MemoryManager

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)


class SovereignJudgment:
    """
    Sovereign (Creative Director)의 최종 판단을 실행하는 도구.
    Cycle Protocol의 Stage 4: Manifest 단계에서 사용.
    """

    def __init__(self):
        self.dual_ai = DualAIEngine()
        self.memory = MemoryManager()

        # MBQ 기준 로드
        self.mbq_criteria = self.load_mbq_criteria()

        # 판단 히스토리 디렉토리
        self.judgment_dir = PROJECT_ROOT / "knowledge" / "council_log" / "sovereign_judgments"
        self.judgment_dir.mkdir(parents=True, exist_ok=True)

    def load_mbq_criteria(self) -> Dict[str, str]:
        """Load MBQ criteria from brand constitution"""
        criteria_file = PROJECT_ROOT / "directives" / "brand_constitution.md"

        # 기본 MBQ 기준
        default_criteria = {
            "철학적_일치성": "97layer의 5대 철학 축(고독, 불완전, 시간, 선례, 반알고리즘)과 일치하는가?",
            "톤_일관성": "Aesop처럼 절제되고 지적이며, 과장이 없는가?",
            "구조_완성도": "Hook → Manuscript → Afterglow 구조를 따르는가?",
            "반알고리즘성": "자극과 휘발성을 거부하고 깊이를 추구하는가?",
            "시각적_일치": "모노크롬, 60% 여백, 미니멀리즘을 구현하는가?"
        }

        # 실제 파일에서 추가 기준 로드 시도
        if criteria_file.exists():
            try:
                with open(criteria_file, "r", encoding="utf-8") as f:
                    content = f.read()
                    # MBQ 섹션 찾기
                    if "## MBQ" in content or "## Manifest Brand Quality" in content:
                        # 추가 로직으로 파일에서 기준 추출 가능
                        pass
            except Exception as e:
                logger.error(f"Error loading MBQ criteria: {e}")

        return default_criteria

    def judge_content(self,
                     content_path: str,
                     content_type: str = "essay",
                     metadata: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        콘텐츠에 대한 Sovereign 판단 실행

        Args:
            content_path: 판단할 콘텐츠 파일 경로
            content_type: essay, visual, strategy 중 하나
            metadata: 추가 메타데이터

        Returns:
            판단 결과
        """

        # 콘텐츠 로드
        content_file = Path(content_path)
        if not content_file.exists():
            return {
                "approved": False,
                "reason": f"Content file not found: {content_path}",
                "timestamp": datetime.now().isoformat()
            }

        with open(content_file, "r", encoding="utf-8") as f:
            content = f.read()

        # 콘텐츠 타입별 기준 조정
        criteria = self.mbq_criteria.copy()
        if content_type == "visual":
            criteria["시각적_완성도"] = "이미지가 브랜드의 시각적 정체성을 구현하는가?"
        elif content_type == "strategy":
            criteria["전략적_타당성"] = "장기적 브랜드 방향과 일치하는가?"

        logger.info(f"Sovereign judging {content_type}: {content_path}")

        # Claude로 판단 (월 20회 제한)
        result = self.dual_ai.process(
            prompt=f"Sovereign judgment required for {content_type}",
            context=content,
            force_claude=True  # Sovereign은 항상 Claude 사용
        )

        # 판단 결과 처리
        if result["engine"] == "claude":
            judgment = result["response"]
        else:
            # Fallback to Gemini if Claude unavailable
            logger.warning("Claude unavailable, using Gemini for judgment")
            judgment = self._gemini_fallback_judgment(content, criteria)

        # 판단 기록 저장
        self.save_judgment_record(content_path, content_type, judgment)

        # 승인된 경우 다음 단계로 이동
        if isinstance(judgment, dict) and judgment.get("approved"):
            self.move_to_publish(content_path, content_type)

        return judgment

    def _gemini_fallback_judgment(self,
                                 content: str,
                                 criteria: Dict[str, str]) -> Dict[str, Any]:
        """Gemini를 사용한 대체 판단 (Claude 한도 초과 시)"""

        # Gemini는 이미 dual_ai에 통합되어 있음
        prompt = f"""WOOHWAHAE Creative Director로서 콘텐츠를 판단하세요.

콘텐츠:
{content[:2000]}  # 토큰 절약

기준:
{json.dumps(criteria, indent=2, ensure_ascii=False)}

JSON 형식으로 응답:
- approved: true/false
- score: 0-100
- reason: 판단 근거
- suggestions: 개선 제안 리스트"""

        response = self.dual_ai.process(
            prompt=prompt,
            force_gemini=True
        )

        # Parse response
        try:
            if isinstance(response["response"], str):
                # JSON 추출 시도
                import re
                json_match = re.search(r'\{.*\}', response["response"], re.DOTALL)
                if json_match:
                    return json.loads(json_match.group())
        except:
            pass

        # 파싱 실패 시 기본 응답
        return {
            "approved": False,
            "score": 50,
            "reason": "Gemini fallback judgment - manual review recommended",
            "engine": "gemini_fallback"
        }

    def save_judgment_record(self,
                            content_path: str,
                            content_type: str,
                            judgment: Dict[str, Any]):
        """판단 기록을 저장"""

        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        record_file = self.judgment_dir / f"judgment_{timestamp}.json"

        record = {
            "timestamp": datetime.now().isoformat(),
            "content_path": str(content_path),
            "content_type": content_type,
            "judgment": judgment,
            "engine": judgment.get("engine", "unknown")
        }

        with open(record_file, "w", encoding="utf-8") as f:
            json.dump(record, f, indent=2, ensure_ascii=False)

        logger.info(f"Judgment record saved: {record_file}")

    def move_to_publish(self, content_path: str, content_type: str):
        """승인된 콘텐츠를 발행 준비 폴더로 이동"""

        source = Path(content_path)
        publish_dir = PROJECT_ROOT / "knowledge" / "assets" / "ready_to_publish"
        publish_dir.mkdir(parents=True, exist_ok=True)

        # 타임스탬프 추가
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        dest_name = f"{content_type}_{timestamp}_{source.name}"
        dest = publish_dir / dest_name

        # 파일 복사 (원본 유지)
        import shutil
        shutil.copy2(source, dest)

        logger.info(f"Content moved to publish queue: {dest}")

        # Memory Manager에 기록
        self.memory.save({
            "action": "content_approved",
            "source": str(content_path),
            "destination": str(dest),
            "timestamp": datetime.now().isoformat()
        }, category="publish_queue")

    def batch_judge(self, content_dir: str, content_type: str = "essay") -> Dict[str, Any]:
        """
        디렉토리의 모든 콘텐츠를 일괄 판단

        Args:
            content_dir: 판단할 콘텐츠들이 있는 디렉토리
            content_type: 콘텐츠 타입

        Returns:
            일괄 판단 결과
        """

        content_path = Path(content_dir)
        if not content_path.exists():
            return {"error": f"Directory not found: {content_dir}"}

        results = {
            "total": 0,
            "approved": 0,
            "rejected": 0,
            "details": []
        }

        # 모든 마크다운 파일 판단
        for file_path in content_path.glob("*.md"):
            result = self.judge_content(
                content_path=str(file_path),
                content_type=content_type
            )

            results["total"] += 1
            if isinstance(result, dict) and result.get("approved"):
                results["approved"] += 1
            else:
                results["rejected"] += 1

            results["details"].append({
                "file": file_path.name,
                "result": result
            })

            # Claude 한도 체크
            if self.dual_ai.claude.usage["calls"] >= 19:
                logger.warning("Approaching Claude monthly limit, stopping batch")
                break

        return results

    def get_judgment_stats(self, days: int = 30) -> Dict[str, Any]:
        """최근 판단 통계 조회"""

        stats = {
            "total_judgments": 0,
            "approved": 0,
            "rejected": 0,
            "by_type": {},
            "by_engine": {},
            "recent_judgments": []
        }

        # 최근 판단 기록 로드
        from datetime import timedelta
        cutoff_date = datetime.now() - timedelta(days=days)

        for record_file in self.judgment_dir.glob("judgment_*.json"):
            try:
                with open(record_file, "r", encoding="utf-8") as f:
                    record = json.load(f)

                # 날짜 필터링
                record_date = datetime.fromisoformat(record["timestamp"])
                if record_date < cutoff_date:
                    continue

                stats["total_judgments"] += 1

                # 승인/거부 카운트
                if record["judgment"].get("approved"):
                    stats["approved"] += 1
                else:
                    stats["rejected"] += 1

                # 타입별 통계
                content_type = record.get("content_type", "unknown")
                stats["by_type"][content_type] = stats["by_type"].get(content_type, 0) + 1

                # 엔진별 통계
                engine = record.get("engine", "unknown")
                stats["by_engine"][engine] = stats["by_engine"].get(engine, 0) + 1

                # 최근 5개 기록
                if len(stats["recent_judgments"]) < 5:
                    stats["recent_judgments"].append({
                        "timestamp": record["timestamp"],
                        "content": Path(record["content_path"]).name,
                        "approved": record["judgment"].get("approved", False),
                        "score": record["judgment"].get("score", 0)
                    })

            except Exception as e:
                logger.error(f"Error loading judgment record {record_file}: {e}")

        # 승인률 계산
        if stats["total_judgments"] > 0:
            stats["approval_rate"] = round(stats["approved"] / stats["total_judgments"] * 100, 1)
        else:
            stats["approval_rate"] = 0

        return stats


def main():
    """CLI 인터페이스"""
    import argparse

    parser = argparse.ArgumentParser(description="Sovereign Judgment Tool")
    parser.add_argument("action", choices=["judge", "batch", "stats"],
                       help="Action to perform")
    parser.add_argument("--content", help="Content file path for judgment")
    parser.add_argument("--type", default="essay",
                       choices=["essay", "visual", "strategy"],
                       help="Content type")
    parser.add_argument("--dir", help="Directory for batch judgment")
    parser.add_argument("--days", type=int, default=30,
                       help="Days to include in stats (default: 30)")

    args = parser.parse_args()

    sovereign = SovereignJudgment()

    if args.action == "judge":
        if not args.content:
            print("Error: --content required for judge action")
            sys.exit(1)

        result = sovereign.judge_content(
            content_path=args.content,
            content_type=args.type
        )
        print(json.dumps(result, indent=2, ensure_ascii=False))

    elif args.action == "batch":
        if not args.dir:
            print("Error: --dir required for batch action")
            sys.exit(1)

        results = sovereign.batch_judge(
            content_dir=args.dir,
            content_type=args.type
        )
        print(f"\nBatch Judgment Results:")
        print(f"  Total: {results['total']}")
        print(f"  Approved: {results['approved']}")
        print(f"  Rejected: {results['rejected']}")

    elif args.action == "stats":
        stats = sovereign.get_judgment_stats(days=args.days)
        print(f"\nSovereign Judgment Stats (Last {args.days} days):")
        print(f"  Total Judgments: {stats['total_judgments']}")
        print(f"  Approved: {stats['approved']}")
        print(f"  Rejected: {stats['rejected']}")
        print(f"  Approval Rate: {stats['approval_rate']}%")
        print(f"\nBy Type:")
        for ctype, count in stats['by_type'].items():
            print(f"    {ctype}: {count}")
        print(f"\nBy Engine:")
        for engine, count in stats['by_engine'].items():
            print(f"    {engine}: {count}")

    # Claude 사용량 표시
    print(f"\n{sovereign.dual_ai.claude.get_usage_summary()}")


if __name__ == "__main__":
    main()