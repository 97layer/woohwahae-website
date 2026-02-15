#!/usr/bin/env python3
"""
Ralph Loop - STAP Validation Engine
97layerOS 품질 보증 시스템

철학:
- "완벽주의 마비"를 극복하되, 품질은 강제한다
- 72시간 규칙: 불완전한 완료 > 완벽한 지연
- STAP 4단계 검증으로 최소 품질 보장

STAP Protocol:
  S (Stop)    - 결과물을 멈추고 평가 시작
  T (Task)    - 원래 목표/의도와 대조
  A (Assess)  - 품질 점수화 (0-100)
  P (Process) - Pass/Revise/Archive 결정

Author: 97layerOS Technical Director
Created: 2026-02-16
"""

import sys
from pathlib import Path
from datetime import datetime
from typing import Dict, Any, Optional, List
import json

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))


class RalphLoop:
    """
    Ralph Loop STAP Validation Engine

    완벽주의 마비와 낮은 품질 사이의 균형을 유지하는 품질 보증 시스템
    """

    def __init__(self):
        self.validation_log_path = PROJECT_ROOT / 'knowledge' / 'system' / 'ralph_validations.jsonl'
        self.validation_log_path.parent.mkdir(parents=True, exist_ok=True)

        # 품질 임계값
        self.THRESHOLDS = {
            'pass': 70,          # 70점 이상: 즉시 통과
            'revise': 50,        # 50-69점: 1회 재작업 권장
            'archive': 50        # 50점 미만: 아카이브 (향후 참고용)
        }

    def validate(
        self,
        asset_path: str,
        original_task: str,
        content: str,
        asset_type: str = "content",
        metadata: Optional[Dict] = None
    ) -> Dict[str, Any]:
        """
        STAP 검증 실행

        Args:
            asset_path: 자산 파일 경로
            original_task: 원래 작업 의도/목표
            content: 검증할 콘텐츠
            asset_type: 자산 유형 (content, code, visual, insight)
            metadata: 추가 메타데이터

        Returns:
            {
                'decision': 'pass' | 'revise' | 'archive',
                'quality_score': 0-100,
                'stap_report': {...},
                'recommendations': [...]
            }
        """
        print(f"\n{'='*70}")
        print(f"🔍 Ralph Loop - STAP Validation")
        print(f"{'='*70}")
        print(f"📄 Asset: {asset_path}")
        print(f"🎯 Task: {original_task[:60]}...")

        # S - Stop: 결과물 준비
        stop_result = self._step_stop(asset_path, content)

        # T - Task: 목표 대조
        task_result = self._step_task(original_task, content, asset_type)

        # A - Assess: 품질 점수화
        assess_result = self._step_assess(content, asset_type, task_result)

        # P - Process: 최종 결정
        process_result = self._step_process(assess_result['quality_score'])

        # 통합 결과
        stap_report = {
            'timestamp': datetime.now().isoformat(),
            'asset_path': asset_path,
            'original_task': original_task,
            'asset_type': asset_type,
            'stop': stop_result,
            'task': task_result,
            'assess': assess_result,
            'process': process_result,
            'metadata': metadata or {}
        }

        # 로그 저장
        self._log_validation(stap_report)

        # 결과 반환
        result = {
            'decision': process_result['decision'],
            'quality_score': assess_result['quality_score'],
            'stap_report': stap_report,
            'recommendations': process_result['recommendations']
        }

        # 출력
        self._print_result(result)

        return result

    def _step_stop(self, asset_path: str, content: str) -> Dict[str, Any]:
        """S - Stop: 결과물 멈추고 기본 정보 수집"""
        word_count = len(content.split())
        char_count = len(content)
        has_structure = any(marker in content for marker in ['#', '##', '```', '---'])

        return {
            'asset_path': asset_path,
            'word_count': word_count,
            'char_count': char_count,
            'has_structure': has_structure,
            'timestamp': datetime.now().isoformat()
        }

    def _step_task(self, original_task: str, content: str, asset_type: str) -> Dict[str, Any]:
        """T - Task: 원래 목표와 대조"""
        # 키워드 추출 (간단한 휴리스틱)
        task_keywords = set(original_task.lower().split())
        content_keywords = set(content.lower().split())

        # 키워드 일치율
        common_keywords = task_keywords & content_keywords
        if len(task_keywords) > 0:
            keyword_match_rate = len(common_keywords) / len(task_keywords)
        else:
            keyword_match_rate = 0.0

        # 작업 유형별 체크리스트
        checklist = self._get_type_checklist(asset_type, content)

        return {
            'task_keywords': list(task_keywords)[:10],  # 상위 10개만
            'keyword_match_rate': round(keyword_match_rate * 100, 2),
            'checklist': checklist,
            'alignment': 'high' if keyword_match_rate > 0.5 else 'medium' if keyword_match_rate > 0.3 else 'low'
        }

    def _get_type_checklist(self, asset_type: str, content: str) -> Dict[str, bool]:
        """자산 유형별 체크리스트"""
        if asset_type == "content":
            return {
                'has_title': content.startswith('#'),
                'has_sections': '##' in content,
                'has_conclusion': any(word in content.lower() for word in ['결론', '요약', 'conclusion', 'summary']),
                'sufficient_length': len(content.split()) >= 100
            }
        elif asset_type == "code":
            return {
                'has_docstring': '"""' in content or "'''" in content,
                'has_comments': '#' in content and not content.strip().startswith('#'),
                'has_functions': 'def ' in content,
                'has_main': '__main__' in content
            }
        elif asset_type == "visual":
            return {
                'has_description': len(content) > 50,
                'has_context': '**' in content or '__' in content,
                'has_metadata': 'From:' in content or 'Type:' in content
            }
        else:  # insight
            return {
                'has_context': len(content) > 100,
                'has_structure': '##' in content or '-' in content,
                'has_actionable': any(word in content for word in ['행동', 'action', '다음', 'next'])
            }

    def _step_assess(self, content: str, asset_type: str, task_result: Dict) -> Dict[str, Any]:
        """A - Assess: 품질 점수화 (0-100)"""
        scores = []

        # 1. 작업 정렬도 (Task Alignment) - 30점
        keyword_match = task_result['keyword_match_rate']
        scores.append(('task_alignment', keyword_match * 0.3))

        # 2. 구조 완성도 (Structure) - 30점
        checklist = task_result['checklist']
        structure_score = (sum(checklist.values()) / len(checklist)) * 30
        scores.append(('structure', structure_score))

        # 3. 내용 충실도 (Content Quality) - 25점
        word_count = len(content.split())
        if word_count >= 300:
            content_score = 25
        elif word_count >= 150:
            content_score = 20
        elif word_count >= 50:
            content_score = 15
        else:
            content_score = 10
        scores.append(('content_quality', content_score))

        # 4. 가독성 (Readability) - 15점
        has_paragraphs = '\n\n' in content
        has_formatting = any(marker in content for marker in ['**', '__', '`', '```'])
        readability_score = (has_paragraphs * 8) + (has_formatting * 7)
        scores.append(('readability', readability_score))

        # 총점 계산
        total_score = sum(score for _, score in scores)

        return {
            'quality_score': round(total_score, 2),
            'breakdown': dict(scores),
            'grade': self._get_grade(total_score)
        }

    def _get_grade(self, score: float) -> str:
        """점수를 등급으로 변환"""
        if score >= 90:
            return 'A+'
        elif score >= 80:
            return 'A'
        elif score >= 70:
            return 'B+'
        elif score >= 60:
            return 'B'
        elif score >= 50:
            return 'C'
        else:
            return 'D'

    def _step_process(self, quality_score: float) -> Dict[str, Any]:
        """P - Process: Pass/Revise/Archive 결정"""
        if quality_score >= self.THRESHOLDS['pass']:
            decision = 'pass'
            action = '✅ 통과 - 즉시 발행 가능'
            recommendations = [
                '품질 기준 충족',
                '자산으로 등록 및 발행'
            ]
        elif quality_score >= self.THRESHOLDS['revise']:
            decision = 'revise'
            action = '🔄 재작업 권장 - 1회 개선 후 재검증'
            recommendations = [
                '구조 강화 (섹션, 소제목 추가)',
                '내용 보충 (최소 300단어 목표)',
                '가독성 개선 (단락, 포맷팅)',
                '작업 의도와 정렬 확인'
            ]
        else:
            decision = 'archive'
            action = '📦 아카이브 - 현재는 보류, 향후 참고용'
            recommendations = [
                '현재 품질 기준 미달',
                'knowledge/archive/low_quality/로 이동',
                '향후 개선 또는 폐기 검토'
            ]

        return {
            'decision': decision,
            'action': action,
            'recommendations': recommendations,
            'threshold_used': self.THRESHOLDS
        }

    def _log_validation(self, stap_report: Dict):
        """검증 결과 로그 저장 (JSONL)"""
        with open(self.validation_log_path, 'a', encoding='utf-8') as f:
            f.write(json.dumps(stap_report, ensure_ascii=False) + '\n')

    def _print_result(self, result: Dict):
        """결과 출력"""
        print(f"\n{'='*70}")
        print(f"📊 Ralph Loop 검증 결과")
        print(f"{'='*70}")
        print(f"⭐ 품질 점수: {result['quality_score']}/100")
        print(f"   등급: {result['stap_report']['assess']['grade']}")
        print(f"\n🎯 최종 결정: {result['decision'].upper()}")
        print(f"   {result['stap_report']['process']['action']}")
        print(f"\n💡 권장 사항:")
        for i, rec in enumerate(result['recommendations'], 1):
            print(f"   {i}. {rec}")
        print(f"{'='*70}\n")

    def get_validation_history(self, limit: int = 10) -> List[Dict]:
        """최근 검증 이력 조회"""
        if not self.validation_log_path.exists():
            return []

        validations = []
        with open(self.validation_log_path, 'r', encoding='utf-8') as f:
            for line in f:
                validations.append(json.loads(line))

        return validations[-limit:]

    def get_statistics(self) -> Dict[str, Any]:
        """검증 통계"""
        history = self.get_validation_history(limit=1000)

        if not history:
            return {
                'total': 0,
                'pass_rate': 0,
                'avg_score': 0,
                'by_decision': {}
            }

        total = len(history)
        decisions = {}
        scores = []

        for validation in history:
            decision = validation['process']['decision']
            decisions[decision] = decisions.get(decision, 0) + 1
            scores.append(validation['assess']['quality_score'])

        return {
            'total': total,
            'pass_rate': round((decisions.get('pass', 0) / total) * 100, 2),
            'avg_score': round(sum(scores) / len(scores), 2) if scores else 0,
            'by_decision': decisions
        }


def main():
    """CLI 테스트"""
    import argparse

    parser = argparse.ArgumentParser(description="Ralph Loop STAP Validation")
    parser.add_argument('--test', action='store_true', help='Run test validation')
    parser.add_argument('--stats', action='store_true', help='Show validation statistics')
    args = parser.parse_args()

    loop = RalphLoop()

    if args.stats:
        stats = loop.get_statistics()
        print(f"\n📊 Ralph Loop 통계")
        print(f"{'='*70}")
        print(f"총 검증 횟수: {stats['total']}")
        print(f"통과율: {stats['pass_rate']}%")
        print(f"평균 점수: {stats['avg_score']}/100")
        print(f"\n결정 분포:")
        for decision, count in stats['by_decision'].items():
            print(f"  {decision}: {count}회")
        print(f"{'='*70}\n")
        return

    if args.test:
        # 테스트 케이스
        test_content = """# 슬로우 라이프의 실천

## 핵심 철학

속도보다 방향, 효율보다 본질을 선택하는 삶의 방식입니다.

### 실천 방법

1. 완벽보다 완료를 우선
2. 72시간 규칙 적용
3. 과정의 흔적 보존

## 결론

불완전함을 수용하며 나다운 속도로 나아갑니다.
"""

        result = loop.validate(
            asset_path="tests/test_slow_life.md",
            original_task="슬로우 라이프 철학을 설명하는 콘텐츠 작성",
            content=test_content,
            asset_type="content"
        )

        print(f"\n✅ 테스트 완료")
        print(f"결정: {result['decision']}")
        print(f"점수: {result['quality_score']}/100")


if __name__ == "__main__":
    main()
