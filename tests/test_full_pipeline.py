#!/usr/bin/env python3
"""
WOOHWAHAE Full Pipeline Test
Complete system integration test with manual content
"""

import sys
import json
import asyncio
from pathlib import Path

# Add project root to path
PROJECT_ROOT = Path(__file__).parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from core.agents.brand_consultant import BrandConsultant, WOOHWAHAESection
from core.utils.content_manager import ContentManager


async def test_full_pipeline():
    """Full pipeline test with manual content"""

    print("="*60)
    print("WOOHWAHAE Full Pipeline Test")
    print("Archive for Slow Life - 완전한 시스템 테스트")
    print("="*60)

    # 1. Initialize components
    print("\n[1] 시스템 초기화")
    print("-"*40)

    # Content Manager
    content_manager = ContentManager()
    print("✅ Content Manager 준비")

    # Brand Consultant (Mock Mode)
    brand_consultant = BrandConsultant(mock_mode=True)
    print("✅ Brand Consultant 준비 (Mock Mode)")

    # 2. Load sample content
    print("\n[2] 샘플 콘텐츠 로드")
    print("-"*40)

    # Clear existing and load fresh
    content_manager.content_db = []
    content_manager.load_sample_content()

    stats = content_manager.get_statistics()
    print(f"✅ {stats['total']}개 콘텐츠 로드 완료")
    print(f"   - Manual: {stats['by_source'].get('manual', 0)}")
    print(f"   - Pending: {stats['by_status'].get('pending', 0)}")

    # 3. Process pending content
    print("\n[3] 콘텐츠 처리 파이프라인")
    print("-"*40)

    pending_content = content_manager.get_pending_content()

    for content_item in pending_content:
        content_id = content_item['id']
        title = content_item.get('title', 'Untitled')

        print(f"\n📝 Processing: {title[:50]}...")

        # Prepare for Brand Consultant
        formatted_content = content_manager.prepare_for_brand_consultant(content_id)

        # Brand Audit
        audit_result = await brand_consultant.audit_content(formatted_content)
        philosophy_score = audit_result.get('philosophy_score', 0)

        print(f"   철학 점수: {philosophy_score}/100")

        # 5 Pillars
        pillars = audit_result.get('pillars', {})
        pillar_scores = [f"{k[:4]}:{v}" for k, v in pillars.items()]
        print(f"   5 Pillars: {' | '.join(pillar_scores)}")

        # Section Classification
        sections = await brand_consultant.classify_for_sections(formatted_content)

        if sections:
            print(f"   섹션 분류:")
            for section, score in sections[:3]:
                print(f"     - {section.value}: {score}%")

            # Primary section
            primary_section = sections[0][0]
            print(f"   → 최종 섹션: {primary_section.value.upper()}")

        # Update status
        content_manager.update_content_status(
            content_id,
            'processed',
            {
                'philosophy_score': philosophy_score,
                'pillars': pillars,
                'sections': [(s.value, score) for s, score in sections],
                'primary_section': sections[0][0].value if sections else None
            }
        )

    # 4. Generate Summary Report
    print("\n[4] 최종 리포트")
    print("-"*40)

    final_stats = content_manager.get_statistics()

    print(f"\n✅ 처리 완료:")
    print(f"   - Total: {final_stats['total']} items")
    print(f"   - Processed: {final_stats['by_status'].get('processed', 0)}")
    print(f"   - Pending: {final_stats['by_status'].get('pending', 0)}")

    # Section distribution
    print(f"\n✅ 섹션 분포:")
    section_counts = {}
    for item in content_manager.content_db:
        if item.get('processing_result'):
            primary = item['processing_result'].get('primary_section')
            if primary:
                section_counts[primary] = section_counts.get(primary, 0) + 1

    for section, count in section_counts.items():
        print(f"   - {section}: {count} items")

    # Philosophy score average
    scores = []
    for item in content_manager.content_db:
        if item.get('processing_result'):
            score = item['processing_result'].get('philosophy_score', 0)
            scores.append(score)

    if scores:
        avg_score = sum(scores) / len(scores)
        print(f"\n✅ 평균 철학 점수: {avg_score:.1f}/100")

    # 5. Save results
    print("\n[5] 결과 저장")
    print("-"*40)

    output_file = PROJECT_ROOT / 'tests' / 'pipeline_result.json'
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump({
            'timestamp': content_manager.content_db[0].get('created_at') if content_manager.content_db else None,
            'statistics': final_stats,
            'section_distribution': section_counts,
            'average_philosophy_score': avg_score if scores else 0,
            'processed_items': len([i for i in content_manager.content_db if i.get('status') == 'processed'])
        }, f, ensure_ascii=False, indent=2)

    print(f"✅ Results saved to: {output_file}")

    print("\n" + "="*60)
    print("WOOHWAHAE Pipeline Test Complete!")
    print("Archive for Slow Life - 시스템 검증 완료")
    print("="*60)


if __name__ == '__main__':
    # Run test
    asyncio.run(test_full_pipeline())