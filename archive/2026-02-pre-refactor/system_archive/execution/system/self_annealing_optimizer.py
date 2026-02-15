#!/usr/bin/env python3
"""
Self-Annealing Token Optimizer
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Filename: execution/system/self_annealing_optimizer.py
Purpose: Automatically detect inefficient token usage patterns
         and update directives with learnings
Author: 97LAYER System
Date: 2026-02-15
"""

import json
import logging
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Any
from collections import defaultdict

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class SelfAnnealingOptimizer:
    """
    자가 개선형 토큰 최적화 시스템
    - 비효율적 패턴 자동 감지
    - Directive 자동 업데이트
    - 최적화 전략 학습
    """

    def __init__(self, project_root: str = None):
        self.root = Path(project_root) if project_root else Path(__file__).parent.parent.parent
        self.cache_dir = self.root / ".tmp" / "token_cache"
        self.learning_file = self.root / ".tmp" / "optimization_learnings.json"
        self.learning_file.parent.mkdir(parents=True, exist_ok=True)
        self.directive_file = self.root / "directives" / "token_optimization_protocol.md"

    def analyze_inefficiencies(self) -> Dict[str, List[Dict]]:
        """비효율적인 패턴 분석"""
        logger.info("Analyzing token usage patterns for inefficiencies...")

        inefficiencies = {
            'large_prompts': [],
            'repeated_queries': [],
            'uncached_queries': [],
            'missed_opportunities': []
        }

        if not self.cache_dir.exists():
            logger.warning("No cache directory found")
            return inefficiencies

        cache_files = list(self.cache_dir.glob("*.json"))
        if not cache_files:
            logger.warning("No cache files found")
            return inefficiencies

        prompt_hashes = defaultdict(int)
        large_prompts = []

        for cache_file in cache_files:
            if cache_file.name == "optimization_stats.json":
                continue

            try:
                with open(cache_file, 'r', encoding='utf-8') as f:
                    cached = json.load(f)

                prompt_hash = cached.get('prompt_hash')
                response = cached.get('response', '')
                metadata = cached.get('metadata', {})

                # 반복 쿼리 감지
                prompt_hashes[prompt_hash] += 1

                # 큰 프롬프트 감지 (추정 토큰 > 2000)
                estimated_tokens = len(response) // 4
                if estimated_tokens > 2000:
                    large_prompts.append({
                        'hash': prompt_hash,
                        'estimated_tokens': estimated_tokens,
                        'context': metadata.get('context', 'unknown'),
                        'timestamp': cached.get('timestamp')
                    })

            except Exception as e:
                logger.error(f"Error analyzing cache file {cache_file}: {e}")

        # 반복 쿼리 (2회 이상)
        for prompt_hash, count in prompt_hashes.items():
            if count > 1:
                inefficiencies['repeated_queries'].append({
                    'hash': prompt_hash,
                    'count': count,
                    'recommendation': 'This query was repeated. Good that it was cached!'
                })

        # 큰 프롬프트 (상위 10개)
        inefficiencies['large_prompts'] = sorted(
            large_prompts,
            key=lambda x: x['estimated_tokens'],
            reverse=True
        )[:10]

        # 놓친 기회 계산
        total_queries = len(cache_files) - 1  # optimization_stats.json 제외
        repeated_count = sum(1 for c in prompt_hashes.values() if c > 1)
        cache_potential = (repeated_count / max(total_queries, 1)) * 100

        if cache_potential < 30:
            inefficiencies['missed_opportunities'].append({
                'type': 'low_cache_reuse',
                'metric': f'{cache_potential:.1f}%',
                'recommendation': 'Consider increasing cache duration or identifying more reusable queries'
            })

        logger.info(f"✓ Found {len(inefficiencies['large_prompts'])} large prompts, "
                   f"{len(inefficiencies['repeated_queries'])} repeated queries")

        return inefficiencies

    def generate_learnings(self, inefficiencies: Dict) -> List[Dict]:
        """비효율성을 학습 항목으로 변환"""
        learnings = []

        # 큰 프롬프트 학습
        large_prompts = inefficiencies.get('large_prompts', [])
        if large_prompts:
            avg_tokens = sum(p['estimated_tokens'] for p in large_prompts) / len(large_prompts)

            if avg_tokens > 3000:
                learnings.append({
                    'category': 'prompt_size',
                    'severity': 'high',
                    'observation': f'Average large prompt size: {avg_tokens:.0f} tokens',
                    'learning': 'Prompts are consistently large. Implement snippet extraction before querying.',
                    'action': 'Update code to use token_optimizer.extract_relevant_snippets()',
                    'expected_savings': '60-80% token reduction',
                    'timestamp': datetime.now().isoformat()
                })

        # 반복 쿼리 학습
        repeated = inefficiencies.get('repeated_queries', [])
        if len(repeated) > 5:
            learnings.append({
                'category': 'query_patterns',
                'severity': 'medium',
                'observation': f'{len(repeated)} queries were repeated',
                'learning': 'Good caching behavior detected. Cache hit rate is improving.',
                'action': 'Continue current caching strategy',
                'expected_savings': 'Maintained',
                'timestamp': datetime.now().isoformat()
            })

        # 놓친 기회 학습
        missed = inefficiencies.get('missed_opportunities', [])
        for opportunity in missed:
            learnings.append({
                'category': 'missed_optimization',
                'severity': 'high',
                'observation': opportunity['recommendation'],
                'learning': 'Cache reuse is low. Query patterns may be too diverse.',
                'action': 'Review token_optimization_protocol.md section on cache strategies',
                'expected_savings': '20-40% additional reduction',
                'timestamp': datetime.now().isoformat()
            })

        return learnings

    def save_learnings(self, learnings: List[Dict]):
        """학습 항목 저장"""
        existing_learnings = []

        if self.learning_file.exists():
            try:
                with open(self.learning_file, 'r', encoding='utf-8') as f:
                    existing_learnings = json.load(f)
            except Exception as e:
                logger.error(f"Error loading existing learnings: {e}")

        # 새 학습 추가
        all_learnings = existing_learnings + learnings

        # 최근 50개만 유지
        all_learnings = all_learnings[-50:]

        try:
            with open(self.learning_file, 'w', encoding='utf-8') as f:
                json.dump(all_learnings, f, ensure_ascii=False, indent=2)
            logger.info(f"✓ Saved {len(learnings)} new learnings")
        except Exception as e:
            logger.error(f"Error saving learnings: {e}")

    def update_directive_if_needed(self, learnings: List[Dict]) -> bool:
        """필요 시 directive 자동 업데이트"""
        if not learnings:
            logger.info("No significant learnings to update directive")
            return False

        # High severity 학습만 directive에 반영
        high_severity = [l for l in learnings if l.get('severity') == 'high']

        if not high_severity:
            logger.info("No high-severity learnings found")
            return False

        if not self.directive_file.exists():
            logger.warning(f"Directive file not found: {self.directive_file}")
            return False

        try:
            with open(self.directive_file, 'r', encoding='utf-8') as f:
                directive_content = f.read()

            # Learning section 찾기 또는 생성
            learning_section = "\n\n---\n\n## 🔄 Recent Learnings (Auto-Generated)\n\n"
            learning_section += f"_Last updated: {datetime.now().strftime('%Y-%m-%d %H:%M')}_\n\n"

            for learning in high_severity[-3:]:  # 최근 3개만
                learning_section += f"### {learning['category'].replace('_', ' ').title()}\n\n"
                learning_section += f"**Observation**: {learning['observation']}\n\n"
                learning_section += f"**Learning**: {learning['learning']}\n\n"
                learning_section += f"**Action**: {learning['action']}\n\n"
                learning_section += f"**Expected Savings**: {learning['expected_savings']}\n\n"

            # 기존 learning section 제거
            if "## 🔄 Recent Learnings" in directive_content:
                # 기존 섹션 찾아서 교체
                parts = directive_content.split("## 🔄 Recent Learnings")
                if len(parts) == 2:
                    # 다음 섹션까지 또는 파일 끝까지
                    after_section = parts[1]
                    next_section_pos = after_section.find("\n## ")
                    if next_section_pos != -1:
                        directive_content = parts[0] + learning_section + after_section[next_section_pos:]
                    else:
                        directive_content = parts[0] + learning_section
            else:
                # 파일 끝에 추가
                directive_content += learning_section

            # 파일 저장
            with open(self.directive_file, 'w', encoding='utf-8') as f:
                f.write(directive_content)

            logger.info(f"✓ Updated directive with {len(high_severity)} learnings")
            return True

        except Exception as e:
            logger.error(f"Error updating directive: {e}")
            return False

    def run_self_annealing_cycle(self):
        """자가 개선 사이클 실행"""
        logger.info("Starting self-annealing optimization cycle...")

        # 1. 비효율성 분석
        inefficiencies = self.analyze_inefficiencies()

        # 2. 학습 생성
        learnings = self.generate_learnings(inefficiencies)

        # 3. 학습 저장
        if learnings:
            self.save_learnings(learnings)

        # 4. Directive 업데이트 (필요시)
        updated = self.update_directive_if_needed(learnings)

        # 5. 결과 리포트
        result = {
            'timestamp': datetime.now().isoformat(),
            'inefficiencies_found': {
                'large_prompts': len(inefficiencies.get('large_prompts', [])),
                'repeated_queries': len(inefficiencies.get('repeated_queries', [])),
                'missed_opportunities': len(inefficiencies.get('missed_opportunities', []))
            },
            'learnings_generated': len(learnings),
            'directive_updated': updated
        }

        logger.info("✓ Self-annealing cycle completed")
        logger.info(f"  Large prompts: {result['inefficiencies_found']['large_prompts']}")
        logger.info(f"  Repeated queries: {result['inefficiencies_found']['repeated_queries']}")
        logger.info(f"  Learnings: {result['learnings_generated']}")
        logger.info(f"  Directive updated: {result['directive_updated']}")

        return result

    def get_learning_summary(self) -> List[Dict]:
        """학습 요약 반환"""
        if not self.learning_file.exists():
            return []

        try:
            with open(self.learning_file, 'r', encoding='utf-8') as f:
                learnings = json.load(f)

            # 카테고리별 그룹화
            by_category = defaultdict(list)
            for learning in learnings:
                by_category[learning['category']].append(learning)

            summary = []
            for category, items in by_category.items():
                summary.append({
                    'category': category,
                    'count': len(items),
                    'latest': items[-1] if items else None
                })

            return summary

        except Exception as e:
            logger.error(f"Error reading learning summary: {e}")
            return []


def main():
    """CLI 인터페이스"""
    import sys

    optimizer = SelfAnnealingOptimizer()

    if len(sys.argv) > 1:
        command = sys.argv[1]

        if command == "run":
            # 자가 개선 사이클 실행
            result = optimizer.run_self_annealing_cycle()
            print("\n" + "="*60)
            print("🔄 SELF-ANNEALING OPTIMIZATION CYCLE")
            print("="*60)
            print(f"Large prompts found:     {result['inefficiencies_found']['large_prompts']}")
            print(f"Repeated queries:        {result['inefficiencies_found']['repeated_queries']}")
            print(f"Learnings generated:     {result['learnings_generated']}")
            print(f"Directive updated:       {'Yes' if result['directive_updated'] else 'No'}")
            print("="*60 + "\n")

        elif command == "summary":
            # 학습 요약
            summary = optimizer.get_learning_summary()
            print("\n" + "="*60)
            print("📚 LEARNING SUMMARY")
            print("="*60)
            for item in summary:
                print(f"\n{item['category'].replace('_', ' ').title()}: {item['count']} learnings")
                if item['latest']:
                    print(f"  Latest: {item['latest']['observation']}")
            print("\n" + "="*60 + "\n")

        elif command == "analyze":
            # 분석만 실행
            inefficiencies = optimizer.analyze_inefficiencies()
            print("\n" + "="*60)
            print("🔍 INEFFICIENCY ANALYSIS")
            print("="*60)
            print(f"Large prompts:           {len(inefficiencies['large_prompts'])}")
            print(f"Repeated queries:        {len(inefficiencies['repeated_queries'])}")
            print(f"Missed opportunities:    {len(inefficiencies['missed_opportunities'])}")
            print("="*60 + "\n")

    else:
        print("Usage:")
        print("  python self_annealing_optimizer.py run        # Run full cycle")
        print("  python self_annealing_optimizer.py summary    # Show learning summary")
        print("  python self_annealing_optimizer.py analyze    # Analyze only")


if __name__ == "__main__":
    main()
