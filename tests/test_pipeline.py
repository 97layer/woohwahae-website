#!/usr/bin/env python3
"""
WOOHWAHAE Pipeline Test
전체 시스템 통합 테스트
"""

import sys
import json
import asyncio
from pathlib import Path

# Add project root to path
PROJECT_ROOT = Path(__file__).parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from core.agents.brand_consultant import BrandConsultant, WOOHWAHAESection


async def test_pipeline():
    """파이프라인 테스트"""

    print("="*60)
    print("WOOHWAHAE Pipeline Test")
    print("="*60)

    # 1. 테스트 데이터 로드
    test_file = PROJECT_ROOT / "tests" / "test_content.json"
    with open(test_file, 'r', encoding='utf-8') as f:
        test_data = json.load(f)

    posts = test_data['test_posts']
    print(f"\n✅ {len(posts)}개 테스트 포스트 로드")

    # 2. Brand Consultant 초기화 (Mock 모드)
    try:
        consultant = BrandConsultant(mock_mode=True)
        print("✅ Brand Consultant 초기화 성공 (Mock Mode)")
    except Exception as e:
        print(f"❌ Brand Consultant 초기화 실패: {e}")
        return

    # 3. 각 포스트 처리
    print("\n" + "-"*60)
    print("포스트 처리 시작")
    print("-"*60)

    for post in posts:
        print(f"\n📝 포스트: {post['post_id']}")
        print(f"   제목: {post['caption'][:50]}...")

        # 브랜드 감사
        content = {
            'type': 'instagram_post',
            'source': '@woosunhokr',
            'data': post
        }

        try:
            # 철학 검증
            audit_result = await consultant.audit_content(content)
            philosophy_score = audit_result.get('philosophy_score', 0)
            print(f"   철학 점수: {philosophy_score}/100")

            # 5 Pillars 점수
            pillars = audit_result.get('pillars', {})
            if pillars:
                print(f"   5 Pillars:")
                for pillar, score in pillars.items():
                    print(f"      - {pillar}: {score}/20")

            # 섹션 분류
            sections = await consultant.classify_for_sections(content)
            if sections:
                print(f"   추천 섹션:")
                for section, score in sections[:3]:
                    print(f"      - {section.value}: {score}%")

        except Exception as e:
            print(f"   ❌ 처리 실패: {e}")

    print("\n" + "="*60)
    print("테스트 완료")
    print("="*60)


if __name__ == '__main__':
    # Run test
    asyncio.run(test_pipeline())