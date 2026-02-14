#!/usr/bin/env python3
"""
Junction Auto-Executor
Junction Protocol 5단계 완전 자동화

Flow:
1. Capture: 텔레그램 메시지 → knowledge/raw_signals/
2. Connect: SA가 과거 경험과 자동 연결
3. Meaning: CE가 초고 자동 생성
4. Manifest: CD 승인 대기 (30분 타이머)
5. Cycle: 발행 후 피드백 수집

Author: 97LAYER
Date: 2026-02-14
"""

import asyncio
import json
import sys
import logging
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, Any, Optional, List
import os
from dotenv import load_dotenv

# Project Root
PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.append(str(PROJECT_ROOT))

# Load .env
load_dotenv(PROJECT_ROOT / ".env")

# Import Async Five-Agent System
from execution.async_five_agent_multimodal import AsyncTechnicalDirector

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class JunctionExecutor:
    """
    Junction Protocol 자동 실행기
    97layer 경험 → WOOHWAHAE 콘텐츠 자동 변환
    """

    def __init__(self):
        self.project_root = PROJECT_ROOT
        self.raw_signals_dir = self.project_root / "knowledge" / "raw_signals"
        self.connections_file = self.project_root / "knowledge" / "connections.json"
        self.draft_dir = self.project_root / "knowledge" / "assets" / "draft"
        self.published_dir = self.project_root / "knowledge" / "assets" / "published"

        # 디렉토리 생성
        self.raw_signals_dir.mkdir(parents=True, exist_ok=True)
        self.draft_dir.mkdir(parents=True, exist_ok=True)
        self.published_dir.mkdir(parents=True, exist_ok=True)

        # Async Five-Agent System
        gemini_key = os.getenv("GEMINI_API_KEY")
        claude_key = os.getenv("ANTHROPIC_API_KEY")

        if not gemini_key or not claude_key:
            raise ValueError("API keys not found in .env")

        self.async_td = AsyncTechnicalDirector(gemini_key, claude_key)

        # Junction 실행 통계
        self.stats = {
            "captured": 0,
            "connected": 0,
            "meaning_generated": 0,
            "cd_approved": 0,
            "cd_rejected": 0,
            "published": 0
        }

        logger.info("🔗 Junction Executor initialized")

    async def execute_junction(self, text: str, image_bytes: Optional[bytes] = None,
                               source: str = "telegram", user_id: str = None) -> Dict[str, Any]:
        """
        Junction Protocol 전체 실행

        Args:
            text: 입력 텍스트 (97layer의 경험)
            image_bytes: 이미지 (옵션)
            source: 출처 (telegram, memo, etc.)
            user_id: 사용자 ID

        Returns:
            Junction 실행 결과
        """
        logger.info(f"[Junction] Starting execution for: {text[:50]}...")

        junction_id = f"junction-{datetime.now().timestamp()}"
        result = {
            "junction_id": junction_id,
            "status": "in_progress",
            "phases": {},
            "started_at": datetime.now().isoformat()
        }

        try:
            # Phase 1: Capture
            logger.info("[Junction] Phase 1: Capture")
            capture_result = await self._phase1_capture(text, image_bytes, source, user_id)
            result["phases"]["capture"] = capture_result
            self.stats["captured"] += 1

            # Phase 2: Connect (SA 자동 연결)
            logger.info("[Junction] Phase 2: Connect")
            connect_result = await self._phase2_connect(capture_result["signal_id"], text)
            result["phases"]["connect"] = connect_result
            self.stats["connected"] += 1

            # Phase 3: Meaning (CE 초고 생성)
            logger.info("[Junction] Phase 3: Meaning")
            meaning_result = await self._phase3_meaning(text, image_bytes, connect_result)
            result["phases"]["meaning"] = meaning_result
            self.stats["meaning_generated"] += 1

            # Phase 4: Manifest (CD 최종 판단)
            logger.info("[Junction] Phase 4: Manifest (CD Opus judgment)")
            manifest_result = await self._phase4_manifest(meaning_result)
            result["phases"]["manifest"] = manifest_result

            if manifest_result["approved"]:
                self.stats["cd_approved"] += 1

                # Phase 5: Cycle (자동 발행 및 피드백 수집)
                logger.info("[Junction] Phase 5: Cycle")
                cycle_result = await self._phase5_cycle(manifest_result)
                result["phases"]["cycle"] = cycle_result
                self.stats["published"] += 1

                result["status"] = "published"
            else:
                self.stats["cd_rejected"] += 1
                result["status"] = "rejected"
                result["reason"] = manifest_result.get("reason", "CD rejected")

            result["completed_at"] = datetime.now().isoformat()
            elapsed = (datetime.fromisoformat(result["completed_at"]) -
                      datetime.fromisoformat(result["started_at"])).total_seconds()
            result["elapsed_time"] = elapsed

            logger.info(f"[Junction] Completed in {elapsed:.2f}s - Status: {result['status']}")
            return result

        except Exception as e:
            logger.error(f"[Junction] Error: {e}")
            result["status"] = "error"
            result["error"] = str(e)
            return result

    async def _phase1_capture(self, text: str, image_bytes: Optional[bytes],
                              source: str, user_id: str) -> Dict[str, Any]:
        """
        Phase 1: Capture
        모든 경험을 일단 기록 (필터링 없음)
        """
        signal_id = f"rs-{int(datetime.now().timestamp())}_{source}_{datetime.now().strftime('%Y%m%d')}"

        metadata = {
            "id": signal_id,
            "date": datetime.now().isoformat(),
            "source": source,
            "user_id": user_id,
            "raw_text": text,
            "has_image": image_bytes is not None,
            "tags": [],
            "emotion": None
        }

        # raw_signals/ 저장
        signal_file = self.raw_signals_dir / f"{signal_id}.md"
        with open(signal_file, 'w', encoding='utf-8') as f:
            f.write(f"# Raw Signal: {signal_id}\n\n")
            f.write(f"**Date**: {metadata['date']}\n")
            f.write(f"**Source**: {source}\n")
            f.write(f"**Has Image**: {metadata['has_image']}\n\n")
            f.write(f"---\n\n{text}\n")

        logger.info(f"[Capture] Saved to {signal_file}")

        return {
            "signal_id": signal_id,
            "file_path": str(signal_file),
            "metadata": metadata,
            "status": "captured"
        }

    async def _phase2_connect(self, signal_id: str, text: str) -> Dict[str, Any]:
        """
        Phase 2: Connect
        SA가 과거 경험과 자동 연결
        """
        # SA 분석 (AsyncTD 사용)
        sa_analysis = await self.async_td.sa.analyze_signal_async(text)

        # 과거 신호 파일 로드 (최근 50개)
        past_signals = list(self.raw_signals_dir.glob("rs-*.md"))
        past_signals.sort(reverse=True)
        past_signals = past_signals[:50]

        # 간단한 연결 찾기 (키워드 기반)
        connections = []
        keywords = text.lower().split()[:10]  # 상위 10개 단어

        for past_file in past_signals:
            if past_file.stem == signal_id.replace(".md", ""):
                continue  # 자기 자신 제외

            with open(past_file, 'r', encoding='utf-8') as f:
                past_content = f.read().lower()

            # 유사도 계산 (단순 키워드 매칭)
            matches = sum(1 for kw in keywords if kw in past_content and len(kw) > 3)
            if matches > 0:
                similarity = min(matches / len(keywords), 1.0)
                connections.append({
                    "target": past_file.stem,
                    "similarity": round(similarity, 2),
                    "matches": matches
                })

        # 상위 5개만
        connections.sort(key=lambda x: x["similarity"], reverse=True)
        connections = connections[:5]

        # connections.json 업데이트
        if self.connections_file.exists():
            with open(self.connections_file, 'r', encoding='utf-8') as f:
                all_connections = json.load(f)
        else:
            all_connections = {}

        all_connections[signal_id] = {
            "signal_id": signal_id,
            "connections": connections,
            "philosophy": sa_analysis.get("category", "unknown"),
            "content_potential": "high" if sa_analysis.get("score", 0) >= 70 else "medium",
            "sa_score": sa_analysis.get("score", 0),
            "timestamp": datetime.now().isoformat()
        }

        with open(self.connections_file, 'w', encoding='utf-8') as f:
            json.dump(all_connections, f, indent=2, ensure_ascii=False)

        logger.info(f"[Connect] Found {len(connections)} connections, SA score: {sa_analysis.get('score')}")

        return {
            "signal_id": signal_id,
            "connections": connections,
            "sa_analysis": sa_analysis,
            "content_potential": all_connections[signal_id]["content_potential"],
            "status": "connected"
        }

    async def _phase3_meaning(self, text: str, image_bytes: Optional[bytes],
                              connect_result: Dict[str, Any]) -> Dict[str, Any]:
        """
        Phase 3: Meaning
        CE가 멀티모달 초고 자동 생성
        """
        sa_result = connect_result["sa_analysis"]

        # AsyncTD로 멀티모달 처리 (SA + AD + CE)
        multimodal_result = await self.async_td.process_multimodal_signal(
            text=text,
            image_bytes=image_bytes,
            signal_id=connect_result["signal_id"]
        )

        ce_result = multimodal_result["phases"].get("ce", {})

        if "error" in ce_result:
            return {"status": "error", "error": ce_result["error"]}

        # 초고 저장
        draft_id = f"draft-{datetime.now().timestamp()}"
        draft_file = self.draft_dir / f"{draft_id}.md"

        content = ce_result.get("content", "")
        with open(draft_file, 'w', encoding='utf-8') as f:
            f.write(f"# Draft: {draft_id}\n\n")
            f.write(f"**Signal ID**: {connect_result['signal_id']}\n")
            f.write(f"**SA Score**: {sa_result.get('score')}/100\n")
            f.write(f"**Created**: {datetime.now().isoformat()}\n\n")
            f.write(f"---\n\n{content}\n")

        logger.info(f"[Meaning] Draft saved to {draft_file}")

        return {
            "draft_id": draft_id,
            "draft_file": str(draft_file),
            "content": content,
            "ce_result": ce_result,
            "multimodal": image_bytes is not None,
            "status": "meaning_generated"
        }

    async def _phase4_manifest(self, meaning_result: Dict[str, Any]) -> Dict[str, Any]:
        """
        Phase 4: Manifest
        CD (Claude Opus) 최종 판단 (30분 타이머)
        """
        content = meaning_result["content"]

        # CD Opus 판단
        cd_result = await self.async_td.cd.sovereign_judgment_async(content)

        approved = cd_result.get("approved", False)
        score = cd_result.get("score", 0)

        logger.info(f"[Manifest] CD Judgment: {'APPROVED' if approved else 'REJECTED'} (Score: {score}/100)")

        return {
            "draft_id": meaning_result["draft_id"],
            "approved": approved,
            "score": score,
            "cd_result": cd_result,
            "status": "manifest_complete"
        }

    async def _phase5_cycle(self, manifest_result: Dict[str, Any]) -> Dict[str, Any]:
        """
        Phase 5: Cycle
        자동 발행 및 피드백 수집
        """
        draft_id = manifest_result["draft_id"]
        draft_file = self.draft_dir / f"{draft_id}.md"

        # published/ 디렉토리로 이동
        published_id = f"published-{datetime.now().timestamp()}"
        published_file = self.published_dir / f"{published_id}.md"

        # 파일 이동
        if draft_file.exists():
            with open(draft_file, 'r', encoding='utf-8') as f:
                draft_content = f.read()

            with open(published_file, 'w', encoding='utf-8') as f:
                f.write(draft_content)
                f.write(f"\n\n---\n**Published**: {datetime.now().isoformat()}\n")
                f.write(f"**CD Score**: {manifest_result['score']}/100\n")

            # 초고 삭제
            draft_file.unlink()

        logger.info(f"[Cycle] Published to {published_file}")

        # 메타데이터 기록
        metadata = {
            "published_id": published_id,
            "draft_id": draft_id,
            "published_at": datetime.now().isoformat(),
            "cd_score": manifest_result["score"],
            "channel": "instagram",  # Future: 실제 발행
            "engagement": {
                "likes": 0,
                "comments": 0,
                "saves": 0,
                "note": "조회수 추적 안 함"
            }
        }

        # 순환: 발행된 콘텐츠가 다시 경험이 됨
        # (Future: 피드백 수집 자동화)

        return {
            "published_id": published_id,
            "published_file": str(published_file),
            "metadata": metadata,
            "status": "published"
        }

    def get_stats(self) -> Dict[str, Any]:
        """Junction 실행 통계"""
        return {
            "stats": self.stats,
            "approval_rate": (self.stats["cd_approved"] /
                            max(self.stats["cd_approved"] + self.stats["cd_rejected"], 1) * 100),
            "capture_to_publish_rate": (self.stats["published"] /
                                       max(self.stats["captured"], 1) * 100)
        }


async def main():
    """테스트 메인"""
    executor = JunctionExecutor()

    # 테스트 신호
    test_text = """외장하드 정리했다. 2년 전 영상 발견.
결국 편집 안 했네. 시간만 흘렀다."""

    result = await executor.execute_junction(
        text=test_text,
        source="test",
        user_id="test_user"
    )

    print("\n=== Junction Execution Result ===")
    print(json.dumps(result, indent=2, ensure_ascii=False))

    print("\n=== Junction Stats ===")
    print(json.dumps(executor.get_stats(), indent=2))


if __name__ == "__main__":
    asyncio.run(main())
