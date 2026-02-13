#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Filename: test_nightly_consolidation.py
Author: 97LAYER Test
Date: 2026-02-14
Description: Nightly Consolidation 테스트 - SA Connect 단계 검증
"""

import sys
from pathlib import Path
from datetime import datetime

BASE_DIR = Path(__file__).resolve().parent
sys.path.insert(0, str(BASE_DIR))

from libs.ai_engine import AIEngine
from libs.agent_router import AgentRouter
from libs.core_config import AI_MODEL_CONFIG

def test_nightly_consolidation():
    """
    Nightly Consolidation 테스트: SA Connect 단계 실행
    """
    print("=" * 60)
    print("Nightly Consolidation Test - Phase 5 Verification")
    print("=" * 60)

    # AI 엔진 + Router 초기화
    ai = AIEngine(AI_MODEL_CONFIG)
    router = AgentRouter(ai)

    # SA 시스템 프롬프트 로드 (Core Directives 포함)
    sa_system_prompt = router.build_system_prompt("SA")

    # 최근 3개 Raw Signal 수집
    raw_signals_dir = BASE_DIR / "knowledge" / "raw_signals"
    recent_signals = sorted(
        raw_signals_dir.glob("*.md"),
        key=lambda f: f.stat().st_mtime,
        reverse=True
    )[:5]

    print(f"\n[Input] 최근 Raw Signals {len(recent_signals)}건 발견\n")

    # Aggregate content
    signal_summaries = []
    for i, signal_file in enumerate(recent_signals, 1):
        with open(signal_file, "r", encoding="utf-8") as f:
            content = f.read()[:1000]  # Truncate to 1000 chars
            signal_summaries.append(f"[Signal {i}] {signal_file.name}\n{content}\n")
            print(f"  - {signal_file.name}")

    aggregated_signals = "\n---\n".join(signal_summaries)

    # Consolidation Instruction (from RITUALS_CONFIG)
    instruction = """
    [Junction Protocol - Connect 단계]

    오늘 수집된 모든 Raw Signals를 분석하여 연결 그래프를 생성하십시오.

    [작업 목표]
    1. 과거 기록과 유사성 탐색
    2. 반복 테마 식별
    3. 5가지 철학(Slow, 실용적 미학, 무언의 교감, 자기 긍정, 아카이브)과 연결
    4. 콘텐츠 후보 우선순위 제안

    97layer_identity.md와 연결 강도를 명시하십시오.

    [Output Format]
    JSON 형식의 연결 그래프 (각 신호마다):
    {
      "node_id": "rs-XXX",
      "connections": [
        {
          "target": "97layer_identity.md",
          "section": "시간 아키비스트",
          "strength": 0.9
        }
      ],
      "philosophy": ["Archive", "Slow"],
      "content_potential": "high" | "medium" | "low",
      "priority": 1-10
    }

    콘텐츠 가능성이 높은 신호 3개를 우선순위로 정리하십시오.
    """

    full_prompt = f"""
    {instruction}

    [Collected Raw Signals]
    {aggregated_signals}
    """

    print("\n" + "=" * 60)
    print("Running SA Consolidation (Connect Stage)...")
    print("=" * 60)

    # AI 실행
    result = ai.generate_response(
        prompt=full_prompt,
        system_instruction=sa_system_prompt
    )

    print("\n" + "=" * 60)
    print("Consolidation Result")
    print("=" * 60)
    print(result)
    print("\n" + "=" * 60)

    # Save to patterns folder
    today_str = datetime.now().strftime("%Y-%m-%d")
    pattern_file = BASE_DIR / "knowledge" / "patterns" / f"test_consolidation_{today_str}.md"
    pattern_file.parent.mkdir(parents=True, exist_ok=True)

    with open(pattern_file, "w", encoding="utf-8") as pf:
        pf.write(f"# Test Consolidation ({today_str})\n\n")
        pf.write(f"**Type**: Nightly Consolidation (SA Connect Stage)\n\n")
        pf.write(f"**Signals Analyzed**: {len(recent_signals)}\n\n")
        pf.write("---\n\n")
        pf.write(result)

    print(f"\n✅ Consolidation saved to: {pattern_file}")

    # Verify output quality
    if "node_id" in result and "philosophy" in result:
        print("\n✅ Test Status: SUCCESS")
        print("연결 그래프가 정상적으로 생성되었습니다.")
    else:
        print("\n⚠️ Test Status: WARNING")
        print("연결 그래프 형식이 예상과 다릅니다. 수동 검토가 필요합니다.")

if __name__ == "__main__":
    test_nightly_consolidation()
