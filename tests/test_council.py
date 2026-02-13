#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Filename: test_council.py
Author: 97LAYER Test
Date: 2026-02-14
Description: Council Meeting 테스트 - 새로운 Agent Directives 검증
"""

import sys
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent
sys.path.insert(0, str(BASE_DIR))

from libs.ai_engine import AIEngine
from libs.synapse import Synapse
from libs.core_config import AI_MODEL_CONFIG

def test_council_meeting():
    """
    Council Meeting 테스트: 새로운 Agent Directives로 토론 실행
    """
    print("=" * 60)
    print("Council Meeting Test - Phase 5 Verification")
    print("=" * 60)

    # AI 엔진 초기화
    ai = AIEngine(AI_MODEL_CONFIG)
    synapse = Synapse(ai)

    # 테스트 안건
    topic = """
    [Phase 5 Verification Test]

    WOOHWAHAE 브랜드의 첫 Instagram 콘텐츠 후보가 제출되었습니다:

    주제: "외장하드 정리의 철학"
    Hook: "무엇을 남기고, 무엇을 지울 것인가"
    Manuscript: 디지털 데이터를 정리하는 행위는 단순한 저장 공간 확보가 아니라,
    과거와의 화해이자 미래를 위한 선택이다. 오래된 프로젝트 파일,
    잊혀진 사진들을 보며 우리는 당시의 자신과 마주한다.
    Afterglow: "당신의 외장하드에는 어떤 과거가 남아있습니까?"

    [질문]
    이 콘텐츠가 MBQ 기준을 통과하는가?
    - 철학 일치 (Archive, Slow 철학)
    - 톤 일관성 (Aesop 벤치마크)
    - 구조 완결성 (Hook/Manuscript/Afterglow)

    각 에이전트는 자신의 전문 영역에서 평가하고, CD가 최종 승인/거부를 결정하십시오.
    """

    print(f"\n[Topic]\n{topic}\n")
    print("=" * 60)
    print("Starting Council Meeting...")
    print("=" * 60)

    # Council Meeting 실행
    participants = ["Creative_Director", "Strategy_Analyst", "Chief_Editor"]

    result = synapse.council_meeting(topic, participants=participants)

    print("\n" + "=" * 60)
    print("Council Meeting Result")
    print("=" * 60)
    print(result)
    print("\n" + "=" * 60)

    # 결과 분석
    if "승인" in result or "통과" in result or "발행" in result:
        print("\n✅ Test Status: APPROVED")
        print("MBQ 기준을 통과한 것으로 판단됩니다.")
    elif "거부" in result or "재작업" in result:
        print("\n⚠️ Test Status: REJECTED")
        print("개선 필요 사항이 있습니다.")
    else:
        print("\n🔍 Test Status: REVIEW NEEDED")
        print("수동 검토가 필요합니다.")

    print("\nCouncil Log saved to: knowledge/council_log/")

if __name__ == "__main__":
    test_council_meeting()
