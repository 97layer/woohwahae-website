#!/usr/bin/env python3
"""
Weekly Token Optimization Monitor
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Filename: execution/system/weekly_optimization_monitor.py
Purpose: Generate weekly reports on token usage and optimization effectiveness
Author: 97LAYER System
Date: 2026-02-15
"""

import json
import logging
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, List, Any
from collections import defaultdict

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class WeeklyOptimizationMonitor:
    """주간 토큰 최적화 모니터링 및 리포트 생성"""

    def __init__(self, project_root: str = None):
        self.root = Path(project_root) if project_root else Path(__file__).parent.parent.parent
        self.cache_dir = self.root / ".tmp" / "token_cache"
        self.reports_dir = self.root / ".tmp" / "optimization_reports"
        self.reports_dir.mkdir(parents=True, exist_ok=True)
        self.stats_file = self.cache_dir / "optimization_stats.json"
        self.history_file = self.reports_dir / "weekly_history.json"

    def generate_weekly_report(self) -> Dict[str, Any]:
        """주간 최적화 리포트 생성"""
        logger.info("Generating weekly optimization report...")

        # 현재 통계 로드
        current_stats = self._load_stats()

        # 지난주 통계 로드
        last_week_stats = self._load_last_week_stats()

        # 캐시 파일 분석
        cache_analysis = self._analyze_cache_files()

        # 비교 분석
        comparison = self._compare_weeks(current_stats, last_week_stats)

        # 권장사항 생성
        recommendations = self._generate_recommendations(current_stats, cache_analysis)

        report = {
            'report_date': datetime.now().isoformat(),
            'period': {
                'start': (datetime.now() - timedelta(days=7)).strftime('%Y-%m-%d'),
                'end': datetime.now().strftime('%Y-%m-%d')
            },
            'current_stats': current_stats,
            'last_week_stats': last_week_stats,
            'comparison': comparison,
            'cache_analysis': cache_analysis,
            'recommendations': recommendations,
            'overall_grade': self._calculate_grade(current_stats)
        }

        # 리포트 저장
        self._save_report(report)

        # 히스토리 업데이트
        self._update_history(current_stats)

        return report

    def _load_stats(self) -> Dict:
        """현재 최적화 통계 로드"""
        if not self.stats_file.exists():
            return {
                'cache_hits': 0,
                'cache_writes': 0,
                'tokens_saved': 0,
                'last_updated': None
            }

        try:
            with open(self.stats_file, 'r', encoding='utf-8') as f:
                return json.load(f)
        except Exception as e:
            logger.error(f"Error loading stats: {e}")
            return {}

    def _load_last_week_stats(self) -> Dict:
        """지난주 통계 로드"""
        if not self.history_file.exists():
            return {}

        try:
            with open(self.history_file, 'r', encoding='utf-8') as f:
                history = json.load(f)

            # 가장 최근 기록 반환
            if history:
                return history[-1] if isinstance(history, list) else history

            return {}
        except Exception as e:
            logger.error(f"Error loading history: {e}")
            return {}

    def _analyze_cache_files(self) -> Dict:
        """캐시 파일 분석"""
        if not self.cache_dir.exists():
            return {}

        cache_files = list(self.cache_dir.glob("*.json"))
        cache_files = [f for f in cache_files if f.name != "optimization_stats.json"]

        analysis = {
            'total_cache_files': len(cache_files),
            'cache_size_mb': 0,
            'oldest_cache': None,
            'newest_cache': None,
            'expired_count': 0,
            'popular_queries': []
        }

        if not cache_files:
            return analysis

        timestamps = []
        query_counts = defaultdict(int)
        total_size = 0

        for cache_file in cache_files:
            try:
                total_size += cache_file.stat().st_size

                with open(cache_file, 'r', encoding='utf-8') as f:
                    cached = json.load(f)

                cache_time = datetime.fromisoformat(cached['timestamp'])
                timestamps.append(cache_time)

                # 만료 확인 (24시간 기준)
                if datetime.now() - cache_time > timedelta(hours=24):
                    analysis['expired_count'] += 1

                # 쿼리 타입 집계
                metadata = cached.get('metadata', {})
                query_type = metadata.get('context', 'unknown')
                query_counts[query_type] += 1

            except Exception as e:
                logger.error(f"Error analyzing cache file {cache_file}: {e}")

        if timestamps:
            analysis['oldest_cache'] = min(timestamps).isoformat()
            analysis['newest_cache'] = max(timestamps).isoformat()

        analysis['cache_size_mb'] = round(total_size / (1024 * 1024), 2)
        analysis['popular_queries'] = [
            {'type': k, 'count': v}
            for k, v in sorted(query_counts.items(), key=lambda x: x[1], reverse=True)[:5]
        ]

        return analysis

    def _compare_weeks(self, current: Dict, last_week: Dict) -> Dict:
        """주간 비교 분석"""
        if not last_week:
            return {'status': 'No historical data available'}

        comparison = {}

        metrics = ['cache_hits', 'cache_writes', 'tokens_saved']

        for metric in metrics:
            current_val = current.get(metric, 0)
            last_val = last_week.get(metric, 0)

            if last_val > 0:
                change_pct = ((current_val - last_val) / last_val) * 100
            else:
                change_pct = 100 if current_val > 0 else 0

            comparison[metric] = {
                'current': current_val,
                'last_week': last_val,
                'change': current_val - last_val,
                'change_pct': round(change_pct, 2)
            }

        # 캐시 히트율 비교
        current_total = current.get('cache_hits', 0) + current.get('cache_writes', 0)
        last_total = last_week.get('cache_hits', 0) + last_week.get('cache_writes', 0)

        current_hit_rate = (current.get('cache_hits', 0) / max(current_total, 1)) * 100
        last_hit_rate = (last_week.get('cache_hits', 0) / max(last_total, 1)) * 100

        comparison['cache_hit_rate'] = {
            'current': round(current_hit_rate, 2),
            'last_week': round(last_hit_rate, 2),
            'change': round(current_hit_rate - last_hit_rate, 2)
        }

        return comparison

    def _generate_recommendations(self, stats: Dict, cache_analysis: Dict) -> List[str]:
        """권장사항 생성"""
        recommendations = []

        # 캐시 히트율 분석
        total_requests = stats.get('cache_hits', 0) + stats.get('cache_writes', 0)
        hit_rate = (stats.get('cache_hits', 0) / max(total_requests, 1)) * 100

        if hit_rate < 30:
            recommendations.append({
                'priority': 'HIGH',
                'area': 'Cache Efficiency',
                'issue': f'Low cache hit rate ({hit_rate:.1f}%)',
                'action': 'Review query patterns and increase cache duration for stable queries'
            })
        elif hit_rate < 50:
            recommendations.append({
                'priority': 'MEDIUM',
                'area': 'Cache Efficiency',
                'issue': f'Moderate cache hit rate ({hit_rate:.1f}%)',
                'action': 'Identify frequently repeated queries and optimize caching strategy'
            })

        # 만료된 캐시
        expired = cache_analysis.get('expired_count', 0)
        if expired > 10:
            recommendations.append({
                'priority': 'LOW',
                'area': 'Cache Maintenance',
                'issue': f'{expired} expired cache files',
                'action': 'Run: python execution/system/token_optimizer.py clear 24'
            })

        # 캐시 크기
        cache_size = cache_analysis.get('cache_size_mb', 0)
        if cache_size > 50:
            recommendations.append({
                'priority': 'MEDIUM',
                'area': 'Storage',
                'issue': f'Cache size is {cache_size} MB',
                'action': 'Consider clearing old caches or implementing size-based eviction'
            })

        # 토큰 절약 부족
        tokens_saved = stats.get('tokens_saved', 0)
        if tokens_saved < 10000 and total_requests > 50:
            recommendations.append({
                'priority': 'HIGH',
                'area': 'Token Optimization',
                'issue': f'Only {tokens_saved} tokens saved from {total_requests} requests',
                'action': 'Review token_optimization_protocol.md and apply snippet extraction'
            })

        # 권장사항 없으면 긍정적 피드백
        if not recommendations:
            recommendations.append({
                'priority': 'INFO',
                'area': 'Overall',
                'issue': 'Optimization system performing well',
                'action': 'Continue current practices and monitor weekly'
            })

        return recommendations

    def _calculate_grade(self, stats: Dict) -> str:
        """전체 성적 계산"""
        total_requests = stats.get('cache_hits', 0) + stats.get('cache_writes', 0)

        if total_requests == 0:
            return 'N/A'

        hit_rate = (stats.get('cache_hits', 0) / max(total_requests, 1)) * 100
        tokens_saved = stats.get('tokens_saved', 0)

        # 점수 계산
        score = 0

        # 캐시 히트율 (50점)
        if hit_rate >= 50:
            score += 50
        elif hit_rate >= 40:
            score += 40
        elif hit_rate >= 30:
            score += 30
        else:
            score += hit_rate * 0.5

        # 토큰 절약량 (50점)
        if tokens_saved >= 100000:
            score += 50
        elif tokens_saved >= 50000:
            score += 40
        elif tokens_saved >= 20000:
            score += 30
        elif tokens_saved >= 10000:
            score += 20
        else:
            score += (tokens_saved / 10000) * 20

        # 등급 부여
        if score >= 90:
            return 'A+ (Excellent)'
        elif score >= 80:
            return 'A (Very Good)'
        elif score >= 70:
            return 'B+ (Good)'
        elif score >= 60:
            return 'B (Fair)'
        elif score >= 50:
            return 'C (Needs Improvement)'
        else:
            return 'D (Poor)'

    def _save_report(self, report: Dict):
        """리포트 저장"""
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        report_file = self.reports_dir / f"weekly_report_{timestamp}.json"

        try:
            with open(report_file, 'w', encoding='utf-8') as f:
                json.dump(report, f, ensure_ascii=False, indent=2)
            logger.info(f"✓ Report saved to {report_file}")
        except Exception as e:
            logger.error(f"Error saving report: {e}")

    def _update_history(self, current_stats: Dict):
        """히스토리 업데이트"""
        history = []

        if self.history_file.exists():
            try:
                with open(self.history_file, 'r', encoding='utf-8') as f:
                    history = json.load(f)
            except Exception as e:
                logger.error(f"Error loading history: {e}")

        # 현재 통계 추가
        history.append({
            **current_stats,
            'recorded_at': datetime.now().isoformat()
        })

        # 최근 12주만 유지
        history = history[-12:]

        try:
            with open(self.history_file, 'w', encoding='utf-8') as f:
                json.dump(history, f, ensure_ascii=False, indent=2)
            logger.info(f"✓ History updated")
        except Exception as e:
            logger.error(f"Error saving history: {e}")

    def print_report(self, report: Dict):
        """리포트 출력 (CLI용)"""
        print("\n" + "="*70)
        print("📊 WEEKLY TOKEN OPTIMIZATION REPORT")
        print("="*70)
        print(f"Period: {report['period']['start']} to {report['period']['end']}")
        print(f"Generated: {report['report_date'][:10]}")
        print(f"Overall Grade: {report['overall_grade']}")
        print("="*70)

        # 현재 통계
        stats = report['current_stats']
        print("\n📈 CURRENT STATISTICS")
        print("-" * 70)
        print(f"  Cache Hits:        {stats.get('cache_hits', 0):>10,}")
        print(f"  Cache Writes:      {stats.get('cache_writes', 0):>10,}")
        print(f"  Tokens Saved:      {stats.get('tokens_saved', 0):>10,}")

        total = stats.get('cache_hits', 0) + stats.get('cache_writes', 0)
        hit_rate = (stats.get('cache_hits', 0) / max(total, 1)) * 100
        print(f"  Cache Hit Rate:    {hit_rate:>10.2f}%")

        # 주간 비교
        comparison = report.get('comparison', {})
        if comparison and comparison.get('status') != 'No historical data available':
            print("\n📊 WEEK-OVER-WEEK COMPARISON")
            print("-" * 70)

            for metric, data in comparison.items():
                if isinstance(data, dict) and 'change_pct' in data:
                    change_symbol = "↑" if data['change'] > 0 else "↓" if data['change'] < 0 else "→"
                    print(f"  {metric.replace('_', ' ').title():20s}: "
                          f"{change_symbol} {abs(data['change_pct']):>6.1f}% "
                          f"({data['current']:,} vs {data['last_week']:,})")

        # 캐시 분석
        cache = report.get('cache_analysis', {})
        if cache:
            print("\n💾 CACHE ANALYSIS")
            print("-" * 70)
            print(f"  Total Cache Files: {cache.get('total_cache_files', 0):>10,}")
            print(f"  Cache Size:        {cache.get('cache_size_mb', 0):>10.2f} MB")
            print(f"  Expired Caches:    {cache.get('expired_count', 0):>10,}")

        # 권장사항
        recommendations = report.get('recommendations', [])
        if recommendations:
            print("\n💡 RECOMMENDATIONS")
            print("-" * 70)
            for rec in recommendations:
                priority_emoji = {
                    'HIGH': '🔴',
                    'MEDIUM': '🟡',
                    'LOW': '🟢',
                    'INFO': 'ℹ️'
                }.get(rec['priority'], '•')

                print(f"\n  {priority_emoji} [{rec['priority']}] {rec['area']}")
                print(f"     Issue:  {rec['issue']}")
                print(f"     Action: {rec['action']}")

        print("\n" + "="*70 + "\n")


def main():
    """CLI 인터페이스"""
    import sys

    monitor = WeeklyOptimizationMonitor()

    if len(sys.argv) > 1 and sys.argv[1] == "show-last":
        # 가장 최근 리포트 표시
        reports = sorted(monitor.reports_dir.glob("weekly_report_*.json"))
        if reports:
            with open(reports[-1], 'r', encoding='utf-8') as f:
                report = json.load(f)
            monitor.print_report(report)
        else:
            print("No reports found. Run without arguments to generate first report.")
        return

    # 새 리포트 생성
    report = monitor.generate_weekly_report()
    monitor.print_report(report)


if __name__ == "__main__":
    main()
