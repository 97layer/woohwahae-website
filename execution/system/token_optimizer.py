#!/usr/bin/env python3
"""
Token Optimization Toolkit
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Filename: execution/system/token_optimizer.py
Purpose: Reduce AI token consumption through intelligent caching,
         compression, and snippet extraction
Author: 97LAYER System
Date: 2026-02-15
"""

import json
import hashlib
import logging
from pathlib import Path
from typing import Dict, List, Optional, Any
from datetime import datetime, timedelta

# Setup logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class TokenOptimizer:
    """
    시스템 전체의 토큰 사용을 최적화하는 중앙 관리자
    """

    def __init__(self, project_root: str = None):
        self.root = Path(project_root) if project_root else Path(__file__).parent.parent.parent
        self.cache_dir = self.root / ".tmp" / "token_cache"
        self.cache_dir.mkdir(parents=True, exist_ok=True)
        self.stats_file = self.cache_dir / "optimization_stats.json"

    def hash_content(self, content: str) -> str:
        """컨텐츠의 해시값 생성 (캐시 키로 사용)"""
        return hashlib.sha256(content.encode()).hexdigest()[:16]

    def get_cached_response(self, prompt: str, max_age_hours: int = 24) -> Optional[str]:
        """
        캐시된 응답 반환

        Args:
            prompt: 원본 프롬프트
            max_age_hours: 캐시 유효 시간 (기본 24시간)

        Returns:
            캐시된 응답 또는 None
        """
        cache_key = self.hash_content(prompt)
        cache_file = self.cache_dir / f"{cache_key}.json"

        if not cache_file.exists():
            return None

        try:
            with open(cache_file, 'r', encoding='utf-8') as f:
                cached = json.load(f)

            # 캐시 만료 확인
            cache_time = datetime.fromisoformat(cached['timestamp'])
            if datetime.now() - cache_time > timedelta(hours=max_age_hours):
                logger.info(f"Cache expired for key: {cache_key}")
                return None

            logger.info(f"✓ Cache HIT: {cache_key}")
            self._update_stats('cache_hits')
            return cached['response']

        except Exception as e:
            logger.error(f"Cache read error: {e}")
            return None

    def cache_response(self, prompt: str, response: str, metadata: Dict = None):
        """
        응답을 캐시에 저장

        Args:
            prompt: 원본 프롬프트
            response: AI 응답
            metadata: 추가 메타데이터
        """
        cache_key = self.hash_content(prompt)
        cache_file = self.cache_dir / f"{cache_key}.json"

        cache_data = {
            'timestamp': datetime.now().isoformat(),
            'prompt_hash': cache_key,
            'response': response,
            'metadata': metadata or {}
        }

        try:
            with open(cache_file, 'w', encoding='utf-8') as f:
                json.dump(cache_data, f, ensure_ascii=False, indent=2)
            logger.info(f"✓ Cached response: {cache_key}")
            self._update_stats('cache_writes')
        except Exception as e:
            logger.error(f"Cache write error: {e}")

    def extract_relevant_snippets(self,
                                  content: str,
                                  keywords: List[str],
                                  context_lines: int = 3) -> str:
        """
        키워드 기반으로 관련 있는 코드/텍스트 조각만 추출

        Args:
            content: 전체 컨텐츠
            keywords: 검색할 키워드 리스트
            context_lines: 키워드 주변 컨텍스트 라인 수

        Returns:
            필터링된 스니펫
        """
        lines = content.split('\n')
        relevant_indices = set()

        # 키워드가 포함된 라인 찾기
        for i, line in enumerate(lines):
            if any(keyword.lower() in line.lower() for keyword in keywords):
                # 컨텍스트 라인 포함
                start = max(0, i - context_lines)
                end = min(len(lines), i + context_lines + 1)
                relevant_indices.update(range(start, end))

        if not relevant_indices:
            # 키워드 없으면 처음 몇 줄만 반환
            logger.warning("No keywords found, returning first 20 lines")
            return '\n'.join(lines[:20]) + "\n\n[... truncated ...]"

        # 연속된 구간 추출
        sorted_indices = sorted(relevant_indices)
        snippets = []
        current_snippet = []
        last_idx = -2

        for idx in sorted_indices:
            if idx == last_idx + 1:
                current_snippet.append(lines[idx])
            else:
                if current_snippet:
                    snippets.append('\n'.join(current_snippet))
                current_snippet = [f"[Line {idx+1}]", lines[idx]]
            last_idx = idx

        if current_snippet:
            snippets.append('\n'.join(current_snippet))

        result = '\n\n...\n\n'.join(snippets)

        original_tokens = len(content) // 4  # 대략적인 토큰 수
        snippet_tokens = len(result) // 4
        saved = original_tokens - snippet_tokens

        logger.info(f"✓ Snippet extraction: {snippet_tokens}/{original_tokens} tokens (~{saved} saved)")
        self._update_stats('tokens_saved', saved)

        return result

    def compress_conversation_history(self,
                                     messages: List[Dict],
                                     max_messages: int = 10) -> List[Dict]:
        """
        대화 이력을 압축 (최근 N개 + 요약)

        Args:
            messages: 전체 메시지 리스트
            max_messages: 유지할 최대 메시지 수

        Returns:
            압축된 메시지 리스트
        """
        if len(messages) <= max_messages:
            return messages

        # 시스템 메시지는 보존
        system_msgs = [m for m in messages if m.get('role') == 'system']
        other_msgs = [m for m in messages if m.get('role') != 'system']

        # 최근 메시지 유지
        recent_msgs = other_msgs[-max_messages:]
        old_msgs = other_msgs[:-max_messages]

        # 오래된 메시지는 요약으로 대체
        if old_msgs:
            summary = {
                'role': 'system',
                'content': f"[Earlier conversation summary: {len(old_msgs)} messages compressed]"
            }
            compressed = system_msgs + [summary] + recent_msgs
        else:
            compressed = system_msgs + recent_msgs

        original_tokens = sum(len(str(m)) for m in messages) // 4
        compressed_tokens = sum(len(str(m)) for m in compressed) // 4
        saved = original_tokens - compressed_tokens

        logger.info(f"✓ History compression: {compressed_tokens}/{original_tokens} tokens (~{saved} saved)")
        self._update_stats('tokens_saved', saved)

        return compressed

    def estimate_tokens(self, text: str) -> int:
        """
        텍스트의 대략적인 토큰 수 추정
        (1 token ≈ 4 characters for English, ≈ 2 for Korean)
        """
        # 간단한 휴리스틱: 한글이 많으면 2로, 영어가 많으면 4로 나눔
        korean_chars = sum(1 for c in text if '\uac00' <= c <= '\ud7a3')
        total_chars = len(text)

        if korean_chars / max(total_chars, 1) > 0.3:
            # 한글 비중 높음
            return total_chars // 2
        else:
            # 영어 비중 높음
            return total_chars // 4

    def should_use_snippet(self, content: str, threshold_tokens: int = 1000) -> bool:
        """
        스니펫 추출을 사용할지 결정

        Args:
            content: 검사할 컨텐츠
            threshold_tokens: 스니펫 추출 임계값

        Returns:
            True if content exceeds threshold
        """
        estimated = self.estimate_tokens(content)
        return estimated > threshold_tokens

    def _update_stats(self, metric: str, value: int = 1):
        """내부: 최적화 통계 업데이트"""
        try:
            if self.stats_file.exists():
                with open(self.stats_file, 'r', encoding='utf-8') as f:
                    stats = json.load(f)
            else:
                stats = {
                    'cache_hits': 0,
                    'cache_writes': 0,
                    'tokens_saved': 0,
                    'last_updated': None
                }

            stats[metric] = stats.get(metric, 0) + value
            stats['last_updated'] = datetime.now().isoformat()

            with open(self.stats_file, 'w', encoding='utf-8') as f:
                json.dump(stats, f, ensure_ascii=False, indent=2)

        except Exception as e:
            logger.error(f"Stats update error: {e}")

    def get_optimization_report(self) -> Dict[str, Any]:
        """최적화 통계 리포트 반환"""
        if not self.stats_file.exists():
            return {
                'cache_hits': 0,
                'cache_writes': 0,
                'tokens_saved': 0,
                'cache_hit_rate': 0.0
            }

        try:
            with open(self.stats_file, 'r', encoding='utf-8') as f:
                stats = json.load(f)

            total_requests = stats['cache_hits'] + stats['cache_writes']
            hit_rate = (stats['cache_hits'] / max(total_requests, 1)) * 100

            return {
                **stats,
                'cache_hit_rate': round(hit_rate, 2),
                'total_requests': total_requests
            }
        except Exception as e:
            logger.error(f"Report generation error: {e}")
            return {}

    def clear_cache(self, older_than_hours: int = None):
        """
        캐시 클리어

        Args:
            older_than_hours: 특정 시간보다 오래된 캐시만 삭제 (None이면 전체 삭제)
        """
        deleted = 0

        for cache_file in self.cache_dir.glob("*.json"):
            if cache_file.name == "optimization_stats.json":
                continue

            try:
                if older_than_hours:
                    with open(cache_file, 'r', encoding='utf-8') as f:
                        cached = json.load(f)
                    cache_time = datetime.fromisoformat(cached['timestamp'])
                    if datetime.now() - cache_time <= timedelta(hours=older_than_hours):
                        continue

                cache_file.unlink()
                deleted += 1

            except Exception as e:
                logger.error(f"Error deleting cache file {cache_file}: {e}")

        logger.info(f"✓ Cleared {deleted} cache files")
        return deleted


def main():
    """CLI 테스트 인터페이스"""
    import sys

    optimizer = TokenOptimizer()

    if len(sys.argv) < 2:
        print("Usage:")
        print("  python token_optimizer.py report          # Show optimization stats")
        print("  python token_optimizer.py clear [hours]   # Clear cache")
        print("  python token_optimizer.py test            # Run test")
        return

    command = sys.argv[1]

    if command == "report":
        report = optimizer.get_optimization_report()
        print("\n" + "="*50)
        print("TOKEN OPTIMIZATION REPORT")
        print("="*50)
        for key, value in report.items():
            print(f"{key:20s}: {value}")
        print("="*50 + "\n")

    elif command == "clear":
        hours = int(sys.argv[2]) if len(sys.argv) > 2 else None
        optimizer.clear_cache(older_than_hours=hours)

    elif command == "test":
        # 테스트
        test_content = """
        def calculate_total(items):
            total = 0
            for item in items:
                total += item.price
            return total

        def calculate_tax(total):
            return total * 0.1

        def process_order(order):
            total = calculate_total(order.items)
            tax = calculate_tax(total)
            return total + tax
        """

        # 스니펫 추출 테스트
        snippets = optimizer.extract_relevant_snippets(
            test_content,
            keywords=["calculate_total", "tax"],
            context_lines=2
        )
        print("\n=== Snippet Extraction Test ===")
        print(snippets)

        # 캐시 테스트
        test_prompt = "What is the capital of France?"
        optimizer.cache_response(test_prompt, "Paris")
        cached = optimizer.get_cached_response(test_prompt)
        print(f"\n=== Cache Test ===")
        print(f"Cached response: {cached}")

        # 리포트
        report = optimizer.get_optimization_report()
        print(f"\n=== Report ===")
        print(json.dumps(report, indent=2))


if __name__ == "__main__":
    main()
