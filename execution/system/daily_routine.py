#!/usr/bin/env python3
"""
97layerOS Daily Routine Automation
일일 자동화 루틴 - 아침 브리핑 + 저녁 리포트

철학:
- 슬로우 라이프: 하루의 시작과 끝을 의식적으로 정리
- 자기 긍정: 완료된 것에 집중, 미완성도 인정
- 기록: 매일의 흔적을 담담히 보존

Features:
- 아침 브리핑 (09:00): Pending assets 리뷰, 오늘의 우선순위
- 저녁 리포트 (21:00): Completed assets 요약, 품질 통계
- 주간 요약 (일요일 21:00): 7일 통합 리포트

Author: 97layerOS Technical Director
Created: 2026-02-16
"""

import sys
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, Any, List, Optional
import json

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from system.libs.agents.asset_manager import AssetManager
from execution.system.ralph_loop import RalphLoop


class DailyRoutine:
    """
    일일 자동화 루틴 관리자

    하루의 시작(아침 브리핑)과 끝(저녁 리포트)을 자동화하여
    슬로우 라이프 실천을 지원합니다.
    """

    def __init__(self):
        self.asset_manager = AssetManager()
        self.ralph_loop = RalphLoop()
        self.reports_dir = PROJECT_ROOT / 'knowledge' / 'reports' / 'daily'
        self.reports_dir.mkdir(parents=True, exist_ok=True)

    def morning_briefing(self) -> Dict[str, Any]:
        """
        아침 브리핑 (09:00)

        - Pending/Refined assets 리뷰
        - 오늘의 우선순위 제안
        - 어제 완료 항목 요약

        Returns:
            브리핑 데이터
        """
        print("\n" + "="*70)
        print("🌅 좋은 아침입니다 - Daily Morning Briefing")
        print(f"📅 {datetime.now().strftime('%Y년 %m월 %d일 %A')}")
        print("="*70 + "\n")

        registry = self.asset_manager._load_registry()
        assets = registry.get('assets', [])

        # 상태별 분류
        pending_assets = [a for a in assets if a['status'] in ['captured', 'analyzed']]
        refined_assets = [a for a in assets if a['status'] == 'refined']
        approved_yesterday = [
            a for a in assets
            if a['status'] == 'approved' and self._is_yesterday(a['created_at'])
        ]

        # 브리핑 생성
        briefing = {
            'date': datetime.now().isoformat(),
            'type': 'morning_briefing',
            'summary': {
                'pending_count': len(pending_assets),
                'refined_count': len(refined_assets),
                'completed_yesterday': len(approved_yesterday)
            },
            'pending_assets': pending_assets[:5],  # 상위 5개
            'refined_assets': refined_assets[:5],
            'yesterday_highlights': approved_yesterday
        }

        # 출력
        print(f"📊 현황:")
        print(f"   대기 중: {len(pending_assets)}개")
        print(f"   재작업 필요: {len(refined_assets)}개")
        print(f"   어제 완료: {len(approved_yesterday)}개\n")

        if pending_assets:
            print(f"🎯 오늘의 우선순위 (Pending Assets):")
            for i, asset in enumerate(pending_assets[:3], 1):
                age_days = self._get_age_days(asset['created_at'])
                print(f"   {i}. {Path(asset['path']).name} ({age_days}일 경과)")
            print()

        if refined_assets:
            print(f"🔄 재작업 권장 (Refined Assets):")
            for i, asset in enumerate(refined_assets[:3], 1):
                quality = asset.get('quality_score', 0)
                print(f"   {i}. {Path(asset['path']).name} (품질: {quality}/100)")
            print()

        if approved_yesterday:
            print(f"✅ 어제의 성과:")
            for asset in approved_yesterday[:3]:
                print(f"   • {Path(asset['path']).name}")
            print()

        print("💡 슬로우 라이프 리마인더:")
        print("   속도보다 방향, 효율보다 본질을 기억하세요.")
        print("   오늘도 나다운 속도로 나아갑니다.\n")

        # 저장
        briefing_path = self.reports_dir / f"morning_{datetime.now().strftime('%Y%m%d')}.json"
        with open(briefing_path, 'w', encoding='utf-8') as f:
            json.dump(briefing, f, indent=2, ensure_ascii=False)

        print(f"📄 브리핑 저장: {briefing_path}")
        print("="*70 + "\n")

        return briefing

    def evening_report(self) -> Dict[str, Any]:
        """
        저녁 리포트 (21:00)

        - 오늘 완료된 assets 요약
        - 품질 통계 (Ralph Loop)
        - 내일 권장 작업

        Returns:
            리포트 데이터
        """
        print("\n" + "="*70)
        print("🌙 하루를 마무리합니다 - Daily Evening Report")
        print(f"📅 {datetime.now().strftime('%Y년 %m월 %d일 %A')}")
        print("="*70 + "\n")

        registry = self.asset_manager._load_registry()
        assets = registry.get('assets', [])

        # 오늘 완료/수정된 assets
        today_approved = [
            a for a in assets
            if a['status'] == 'approved' and self._is_today(a['created_at'])
        ]
        today_archived = [
            a for a in assets
            if a['status'] == 'archived' and self._is_today(a['created_at'])
        ]

        # Ralph Loop 통계
        ralph_stats = self.ralph_loop.get_statistics()

        # 리포트 생성
        report = {
            'date': datetime.now().isoformat(),
            'type': 'evening_report',
            'summary': {
                'approved_today': len(today_approved),
                'archived_today': len(today_archived),
                'ralph_stats': ralph_stats
            },
            'completed_assets': today_approved,
            'archived_assets': today_archived,
            'quality_insights': self._generate_quality_insights(ralph_stats)
        }

        # 출력
        print(f"📊 오늘의 성과:")
        print(f"   완료: {len(today_approved)}개")
        print(f"   아카이브: {len(today_archived)}개\n")

        if today_approved:
            print(f"✅ 완료된 작업:")
            for asset in today_approved:
                quality = asset.get('quality_score', 0)
                print(f"   • {Path(asset['path']).name} (품질: {quality}/100)")
            print()

        # Ralph Loop 통계
        print(f"🔍 품질 관리 통계:")
        print(f"   총 검증: {ralph_stats['total']}회")
        print(f"   통과율: {ralph_stats['pass_rate']}%")
        print(f"   평균 점수: {ralph_stats['avg_score']}/100\n")

        # 내일 권장
        pending = [a for a in assets if a['status'] in ['captured', 'analyzed', 'refined']]
        if pending:
            print(f"🎯 내일 추천 작업:")
            sorted_pending = sorted(pending, key=lambda x: self._get_age_days(x['created_at']), reverse=True)
            for i, asset in enumerate(sorted_pending[:3], 1):
                age = self._get_age_days(asset['created_at'])
                print(f"   {i}. {Path(asset['path']).name} ({age}일 경과)")
            print()

        print("💭 하루 마무리:")
        print("   완벽하지 않아도 괜찮습니다.")
        print("   과정의 흔적을 남긴 것만으로도 충분합니다.")
        print("   내일도 슬로우 라이프로 나아갑니다.\n")

        # 저장
        report_path = self.reports_dir / f"evening_{datetime.now().strftime('%Y%m%d')}.json"
        with open(report_path, 'w', encoding='utf-8') as f:
            json.dump(report, f, indent=2, ensure_ascii=False)

        print(f"📄 리포트 저장: {report_path}")
        print("="*70 + "\n")

        return report

    def weekly_summary(self) -> Dict[str, Any]:
        """
        주간 요약 (일요일 21:00)

        - 지난 7일 통합 통계
        - 품질 트렌드 분석
        - 다음 주 목표 제안

        Returns:
            주간 요약 데이터
        """
        print("\n" + "="*70)
        print("📊 Weekly Summary - 지난 한 주를 돌아봅니다")
        print(f"📅 {datetime.now().strftime('%Y년 %m월 %d일')}")
        print("="*70 + "\n")

        registry = self.asset_manager._load_registry()
        assets = registry.get('assets', [])

        # 지난 7일 데이터
        seven_days_ago = datetime.now() - timedelta(days=7)
        week_assets = [
            a for a in assets
            if datetime.fromisoformat(a['created_at']) >= seven_days_ago
        ]

        # 상태별 분류
        by_status = {}
        for asset in week_assets:
            status = asset['status']
            by_status[status] = by_status.get(status, 0) + 1

        # 품질 통계
        quality_scores = [a.get('quality_score', 0) for a in week_assets if a.get('quality_score')]
        avg_quality = sum(quality_scores) / len(quality_scores) if quality_scores else 0

        # Ralph Loop 통계
        ralph_stats = self.ralph_loop.get_statistics()

        # 요약 생성
        summary = {
            'date': datetime.now().isoformat(),
            'type': 'weekly_summary',
            'period': {
                'start': seven_days_ago.isoformat(),
                'end': datetime.now().isoformat()
            },
            'summary': {
                'total_assets': len(week_assets),
                'by_status': by_status,
                'avg_quality': round(avg_quality, 2),
                'ralph_stats': ralph_stats
            }
        }

        # 출력
        print(f"📈 지난 주 성과:")
        print(f"   생성된 자산: {len(week_assets)}개")
        print(f"   평균 품질: {avg_quality:.1f}/100\n")

        print(f"📊 상태별 분포:")
        for status, count in sorted(by_status.items(), key=lambda x: x[1], reverse=True):
            print(f"   {status}: {count}개")
        print()

        print(f"🔍 품질 관리:")
        print(f"   총 검증: {ralph_stats['total']}회")
        print(f"   통과율: {ralph_stats['pass_rate']}%")
        print(f"   평균 점수: {ralph_stats['avg_score']}/100\n")

        # 다음 주 목표
        print(f"🎯 다음 주 제안:")
        if ralph_stats['pass_rate'] < 70:
            print(f"   • 품질 개선에 집중 (현재 통과율: {ralph_stats['pass_rate']}%)")
        if avg_quality < 70:
            print(f"   • 콘텐츠 충실도 향상 (현재: {avg_quality:.1f}/100)")
        pending_count = by_status.get('captured', 0) + by_status.get('analyzed', 0)
        if pending_count > 5:
            print(f"   • Pending 자산 처리 ({pending_count}개 대기 중)")
        print()

        print("💡 슬로우 라이프 회고:")
        print("   한 주 동안 나다운 속도로 잘 나아갔습니다.")
        print("   다음 주도 본질에 집중하며 천천히 나아갑니다.\n")

        # 저장
        week_num = datetime.now().isocalendar()[1]
        summary_path = self.reports_dir / f"weekly_{datetime.now().year}W{week_num:02d}.json"
        with open(summary_path, 'w', encoding='utf-8') as f:
            json.dump(summary, f, indent=2, ensure_ascii=False)

        print(f"📄 주간 요약 저장: {summary_path}")
        print("="*70 + "\n")

        return summary

    def _is_today(self, timestamp: str) -> bool:
        """오늘 생성/수정 여부"""
        try:
            dt = datetime.fromisoformat(timestamp)
            return dt.date() == datetime.now().date()
        except:
            return False

    def _is_yesterday(self, timestamp: str) -> bool:
        """어제 생성/수정 여부"""
        try:
            dt = datetime.fromisoformat(timestamp)
            yesterday = (datetime.now() - timedelta(days=1)).date()
            return dt.date() == yesterday
        except:
            return False

    def _get_age_days(self, timestamp: str) -> int:
        """생성 후 경과 일수"""
        try:
            dt = datetime.fromisoformat(timestamp)
            delta = datetime.now() - dt
            return delta.days
        except:
            return 0

    def _generate_quality_insights(self, ralph_stats: Dict) -> List[str]:
        """품질 통계 기반 인사이트 생성"""
        insights = []

        if ralph_stats['total'] == 0:
            insights.append("아직 품질 검증 데이터가 없습니다.")
            return insights

        # 통과율 분석
        pass_rate = ralph_stats['pass_rate']
        if pass_rate >= 80:
            insights.append(f"✅ 우수한 품질 유지 중 (통과율 {pass_rate}%)")
        elif pass_rate >= 60:
            insights.append(f"📊 양호한 품질 수준 (통과율 {pass_rate}%)")
        else:
            insights.append(f"⚠️ 품질 개선 필요 (통과율 {pass_rate}%)")

        # 평균 점수 분석
        avg_score = ralph_stats['avg_score']
        if avg_score >= 75:
            insights.append(f"⭐ 높은 평균 품질 ({avg_score}/100)")
        elif avg_score >= 60:
            insights.append(f"📈 중간 수준 품질 ({avg_score}/100)")
        else:
            insights.append(f"💡 품질 향상 기회 ({avg_score}/100)")

        return insights


def main():
    """CLI 인터페이스"""
    import argparse

    parser = argparse.ArgumentParser(description="97layerOS Daily Routine")
    parser.add_argument('--morning', action='store_true', help='Run morning briefing')
    parser.add_argument('--evening', action='store_true', help='Run evening report')
    parser.add_argument('--weekly', action='store_true', help='Run weekly summary')
    parser.add_argument('--all', action='store_true', help='Run all routines (test mode)')

    args = parser.parse_args()

    routine = DailyRoutine()

    if args.morning or args.all:
        routine.morning_briefing()

    if args.evening or args.all:
        routine.evening_report()

    if args.weekly or args.all:
        routine.weekly_summary()

    if not (args.morning or args.evening or args.weekly or args.all):
        parser.print_help()


if __name__ == "__main__":
    main()
