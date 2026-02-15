#!/usr/bin/env python3
"""
97layerOS Parallel Multi-Agent Orchestrator
멀티에이전트 병렬 실행 + Handoff 통합 + Container-First

Author: 97layerOS Technical Director
"""

import asyncio
import json
from pathlib import Path
from datetime import datetime
from typing import List, Dict, Any, Optional
import sys

PROJECT_ROOT = Path(__file__).parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from system.libs.engines.ai_engine import AIEngine
from core.agents.async_agent_hub import AsyncAgentHub
from core.system.handoff import HandoffEngine
from core.system.ralph_loop import RalphLoop
from core.agents.asset_manager import AssetManager


class ParallelOrchestrator:
    """
    멀티에이전트 병렬 오케스트레이터

    Features:
    - 5-Agent 병렬 실행 (SA + AD 동시, CE 순차, CD 최종 승인)
    - Handoff 통합 (작업 잠금, 자산 등록)
    - Ralph Loop 자동 검증
    - Container-First 원칙 준수
    """

    def __init__(self):
        self.ai = AIEngine()
        self.agent_hub = AsyncAgentHub(str(PROJECT_ROOT))
        self.handoff = HandoffEngine()
        self.ralph_loop = RalphLoop()
        self.asset_manager = AssetManager()

        # Agent 핸들러 등록
        self._register_agent_handlers()

    def _register_agent_handlers(self):
        """에이전트 핸들러 등록"""
        # SA (Strategy Analyst)
        async def sa_handler(prompt: str, image_data: Optional[bytes] = None) -> str:
            return await self.ai.generate_with_role("SA", prompt, image_data)

        # AD (Art Director)
        async def ad_handler(prompt: str, image_data: Optional[bytes] = None) -> str:
            return await self.ai.generate_with_role("AD", prompt, image_data)

        # CE (Chief Editor)
        async def ce_handler(prompt: str) -> str:
            return await self.ai.generate_with_role("CE", prompt)

        # CD (Creative Director - Claude Opus)
        async def cd_handler(prompt: str) -> str:
            return await self.ai.generate_with_role("CD", prompt)

        # TD (Technical Director)
        async def td_handler(prompt: str) -> str:
            return await self.ai.generate_with_role("TD", prompt)

        # 등록
        self.agent_hub.agents["SA"]["handler"] = sa_handler
        self.agent_hub.agents["SA"]["active"] = True

        self.agent_hub.agents["AD"]["handler"] = ad_handler
        self.agent_hub.agents["AD"]["active"] = True

        self.agent_hub.agents["CE"]["handler"] = ce_handler
        self.agent_hub.agents["CE"]["active"] = True

        self.agent_hub.agents["CD"]["handler"] = cd_handler
        self.agent_hub.agents["CD"]["active"] = True

        self.agent_hub.agents["TD"]["handler"] = td_handler
        self.agent_hub.agents["TD"]["active"] = True

    async def process_signal(self, signal_path: Path, image_path: Optional[Path] = None) -> Dict[str, Any]:
        """
        신호 처리 (멀티에이전트 병렬 실행)

        Workflow:
        1. Work Lock 획득
        2. [병렬] SA 분석 + AD 시각 분석 (if image)
        3. [순차] CE 콘텐츠 정제
        4. [순차] CD 최종 승인
        5. Ralph Loop 검증
        6. Asset 등록
        7. Work Lock 해제

        Args:
            signal_path: 신호 파일 경로
            image_path: 이미지 파일 경로 (옵션)

        Returns:
            처리 결과 딕셔너리
        """
        print("\n" + "="*70)
        print("🚀 Parallel Multi-Agent Processing")
        print("="*70)
        print(f"📄 Signal: {signal_path.name}")
        if image_path:
            print(f"🖼️  Image: {image_path.name}")
        print("="*70 + "\n")

        # 1. Work Lock 획득
        task_id = f"process-{signal_path.stem}-{datetime.now().strftime('%H%M%S')}"
        if not self.handoff.acquire_work_lock(
            agent_id="Parallel_Orchestrator",
            task=f"Processing {signal_path.name}",
            resources=[str(signal_path), str(image_path) if image_path else None]
        ):
            return {"error": "Work lock could not be acquired"}

        try:
            # 신호 내용 읽기
            signal_content = signal_path.read_text(encoding='utf-8')

            # 이미지 데이터 로드
            image_data = None
            if image_path and image_path.exists():
                image_data = image_path.read_bytes()

            # 2. [병렬] SA + AD 동시 실행
            print("🔄 Phase 1: Parallel Analysis (SA + AD)")
            print("-" * 70)

            sa_prompt = f"""
아래 신호를 분석하여 브랜드 전략 인사이트를 도출하십시오.

[신호]
{signal_content}

[분석 요구사항]
- 슬로우 라이프 철학과의 연관성
- 자기 긍정 요소 추출
- 본질 vs 소음 구분
- 아카이브 가치 평가
"""

            ad_prompt = f"""
아래 신호 및 이미지(있는 경우)의 시각적 요소를 분석하십시오.

[신호]
{signal_content}

[분석 요구사항]
- 여백의 미학 적용 가능성
- Monochrome 톤 적합성
- Aesop 벤치마크 대비 점수
- 시각적 개선 제안
"""

            # 병렬 실행
            sa_task = asyncio.create_task(self.agent_hub.agents["SA"]["handler"](sa_prompt, image_data))
            ad_task = asyncio.create_task(self.agent_hub.agents["AD"]["handler"](ad_prompt, image_data))

            # 결과 대기
            sa_analysis, ad_visual_guide = await asyncio.gather(sa_task, ad_task)

            print(f"✅ SA Analysis: {len(sa_analysis)} chars")
            print(f"✅ AD Visual Guide: {len(ad_visual_guide)} chars\n")

            # 3. [순차] CE 콘텐츠 정제
            print("🔄 Phase 2: Content Refinement (CE)")
            print("-" * 70)

            ce_prompt = f"""
SA의 분석과 AD의 시각 가이드를 바탕으로 브랜드 콘텐츠를 작성하십시오.

[SA 분석]
{sa_analysis}

[AD 시각 가이드]
{ad_visual_guide}

[작성 요구사항]
- 톤앤매너: 사려 깊고, 진정성 있으며, 정밀함
- 구조: Opening → Bridge → Core → Closing
- 스타일: 절제된 표현, 여백의 미학
- 길이: 500-1000자
"""

            ce_content = await self.agent_hub.agents["CE"]["handler"](ce_prompt)
            print(f"✅ CE Content: {len(ce_content)} chars\n")

            # 4. [순차] CD 최종 승인
            print("🔄 Phase 3: Final Approval (CD - Claude Opus)")
            print("-" * 70)

            cd_prompt = f"""
아래 콘텐츠가 97layerOS의 브랜드 철학과 일치하는지 최종 검토하십시오.

[콘텐츠]
{ce_content}

[검토 기준]
- 슬로우 라이프 철학 부합 여부
- 자기 긍정 요소 포함 여부
- 72시간 규칙 준수 (완성도 vs 속도)
- 알고리즘 저항 (본질 vs 자극)

[출력 형식]
승인 여부: ✅ 승인 / ❌ 거부
점수: X/100
의견: [한 줄 코멘트]
"""

            cd_decision = await self.agent_hub.agents["CD"]["handler"](cd_prompt)
            print(f"✅ CD Decision: {cd_decision}\n")

            # 5. Ralph Loop 검증
            print("🔄 Phase 4: Ralph Loop STAP Validation")
            print("-" * 70)

            ralph_result = self.ralph_loop.validate(
                asset_path=f"knowledge/content/{signal_path.stem}_final.md",
                original_task=f"신호 분석 및 브랜드 콘텐츠 생성: {signal_path.name}",
                content=ce_content,
                asset_type="content",
                metadata={
                    "signal_source": str(signal_path),
                    "has_image": image_path is not None,
                    "agents_involved": ["SA", "AD", "CE", "CD"]
                }
            )

            print(f"⭐ Quality Score: {ralph_result['quality_score']}/100")
            print(f"🎯 Decision: {ralph_result['decision'].upper()}\n")

            # 6. Asset 등록
            output_path = PROJECT_ROOT / "knowledge" / "content" / f"content_{datetime.now().strftime('%Y%m%d_%H%M%S')}.md"
            output_path.parent.mkdir(parents=True, exist_ok=True)

            # 콘텐츠 저장
            final_content = f"""---
title: {signal_path.stem}
created: {datetime.now().isoformat()}
source: {signal_path.name}
---

# {signal_path.stem}

## SA Analysis
{sa_analysis}

## AD Visual Guide
{ad_visual_guide}

## CE Content
{ce_content}

## CD Decision
{cd_decision}

---
Generated by 97layerOS Parallel Orchestrator
"""

            output_path.write_text(final_content, encoding='utf-8')

            # Asset 등록 with Ralph Loop results
            asset_id = self.asset_manager.register_asset(
                path=str(output_path),
                asset_type="content",
                source="parallel_orchestrator",
                metadata={
                    "signal": str(signal_path),
                    "image": str(image_path) if image_path else None,
                    "agents": ["SA", "AD", "CE", "CD"],
                    "cd_decision": cd_decision,
                    "quality_score": ralph_result['quality_score'],
                    "ralph_decision": ralph_result['decision']
                }
            )

            # Ralph Loop 결정에 따라 자산 상태 업데이트
            if ralph_result['decision'] == 'pass':
                self.asset_manager.update_asset_status(
                    asset_id=asset_id,
                    new_status="approved",
                    updated_by="ParallelOrchestrator",
                    quality_score=ralph_result['quality_score']
                )
            elif ralph_result['decision'] == 'revise':
                self.asset_manager.update_asset_status(
                    asset_id=asset_id,
                    new_status="refined",
                    updated_by="ParallelOrchestrator",
                    quality_score=ralph_result['quality_score']
                )
            else:  # archive
                self.asset_manager.update_asset_status(
                    asset_id=asset_id,
                    new_status="archived",
                    updated_by="ParallelOrchestrator",
                    quality_score=ralph_result['quality_score']
                )

            print(f"📦 Asset registered: {asset_id}")
            print(f"📁 Saved to: {output_path}\n")

            # 7. Work Lock 해제
            self.handoff.release_work_lock("Parallel_Orchestrator")

            print("="*70)
            print("✅ Processing Complete")
            print("="*70 + "\n")

            return {
                "status": "success",
                "asset_id": asset_id,
                "output_path": str(output_path),
                "agents_used": ["SA", "AD", "CE", "CD"],
                "cd_decision": cd_decision
            }

        except Exception as e:
            # 에러 시 Lock 해제
            self.handoff.release_work_lock("Parallel_Orchestrator")
            print(f"❌ Error: {e}")
            return {"status": "error", "error": str(e)}

    async def batch_process_signals(self, signals_dir: Path) -> List[Dict]:
        """
        신호 디렉토리 내 모든 신호를 병렬 처리

        Args:
            signals_dir: 신호 디렉토리 경로

        Returns:
            처리 결과 리스트
        """
        results = []

        for signal_file in signals_dir.glob("*.md"):
            # 연관된 이미지 찾기
            image_file = None
            for ext in ['.jpg', '.jpeg', '.png']:
                img_path = signal_file.with_suffix(ext)
                if img_path.exists():
                    image_file = img_path
                    break

            # 처리
            result = await self.process_signal(signal_file, image_file)
            results.append(result)

        return results


# ─────────────────────────────────────────────────────────────────
# CLI Interface
# ─────────────────────────────────────────────────────────────────

def main():
    import argparse

    parser = argparse.ArgumentParser(description="97layerOS Parallel Multi-Agent Orchestrator")
    parser.add_argument('--signal', type=str, help='단일 신호 파일 처리')
    parser.add_argument('--image', type=str, help='연관 이미지 파일')
    parser.add_argument('--batch', type=str, help='신호 디렉토리 일괄 처리')
    parser.add_argument('--test', action='store_true', help='테스트 모드')

    args = parser.parse_args()

    orchestrator = ParallelOrchestrator()

    if args.test:
        print("🧪 Parallel Orchestrator Test Mode")
        print("   Creating test signal...")

        # 테스트 신호 생성
        test_signal = PROJECT_ROOT / "knowledge" / "signals" / "test_signal.md"
        test_signal.parent.mkdir(parents=True, exist_ok=True)
        test_signal.write_text("""# 테스트 신호

오늘 반지하 원룸에서 바라본 하늘이 유난히 맑았다.
속도에 쫓기지 않고, 나만의 속도로 살아가는 것의 가치를 다시 생각했다.

슬로우 라이프는 게으름이 아니라, 본질에 집중하는 것이다.
""")

        # 비동기 실행
        asyncio.run(orchestrator.process_signal(test_signal))

        print("\n✅ Test completed!")

    elif args.signal:
        signal_path = Path(args.signal)
        image_path = Path(args.image) if args.image else None

        asyncio.run(orchestrator.process_signal(signal_path, image_path))

    elif args.batch:
        signals_dir = Path(args.batch)
        results = asyncio.run(orchestrator.batch_process_signals(signals_dir))

        print(f"\n📊 Batch Processing Complete: {len(results)} signals processed")

    else:
        parser.print_help()


if __name__ == "__main__":
    main()
