#!/usr/bin/env python3
"""
Cycle Manager
Cycle Protocol 순환 구조 자동화

Schedules:
- 주간 Council Meeting (월요일 오전 10시)
- 콘텐츠 후보 제안 (매주 목요일)
- 분기 회고 (3개월마다)

Author: 97LAYER
Date: 2026-02-14
"""

import asyncio
import json
import sys
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, Any, List
import logging
import schedule
import time

PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.append(str(PROJECT_ROOT))

from execution.junction_executor import JunctionExecutor

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class CycleManager:
    """
    Cycle Protocol 관리자
    순환 구조 자동 유지
    """

    def __init__(self):
        self.project_root = PROJECT_ROOT
        self.junction_executor = JunctionExecutor()
        self.council_dir = self.project_root / "knowledge" / "council_log"
        self.reports_dir = self.project_root / "knowledge" / "reports"

        self.council_dir.mkdir(parents=True, exist_ok=True)
        self.reports_dir.mkdir(parents=True, exist_ok=True)

        self.running = False

        logger.info("🔄 Cycle Manager initialized")

    def start(self):
        """스케줄 시작"""
        logger.info("🚀 Starting Cycle Manager schedules...")

        # 주간 Council Meeting (월요일 오전 10시)
        schedule.every().monday.at("10:00").do(self._run_async, self.council_meeting)

        # 콘텐츠 후보 제안 (목요일 오후 3시)
        schedule.every().thursday.at("15:00").do(self._run_async, self.suggest_content_candidates)

        # 분기 회고 (매달 1일)
        schedule.every().day.at("09:00").do(self._check_quarterly_review)

        self.running = True
        logger.info("✅ Schedules registered")

        # 메인 루프
        while self.running:
            schedule.run_pending()
            time.sleep(60)  # 1분마다 체크

    def _run_async(self, coro):
        """비동기 함수 실행 헬퍼"""
        asyncio.run(coro())

    async def council_meeting(self):
        """
        주간 Council Meeting
        - 지난주 발행 회고
        - 이번주 콘텐츠 후보 제안
        - 사이클 병목 지점 체크
        """
        logger.info("🏛️ Council Meeting started")

        meeting_date = datetime.now().strftime("%Y%m%d")
        meeting_file = self.council_dir / f"council_{meeting_date}.md"

        # 지난주 통계
        stats = self.junction_executor.get_stats()
        published_files = list(self.junction_executor.published_dir.glob("published-*.md"))

        # 지난 7일간 발행된 콘텐츠
        week_ago = datetime.now() - timedelta(days=7)
        recent_published = [
            f for f in published_files
            if datetime.fromtimestamp(f.stat().st_mtime) >= week_ago
        ]

        # Meeting 내용 생성
        meeting_content = f"""# Council Meeting - {datetime.now().strftime('%Y-%m-%d')}

## 📊 지난주 회고

**발행 통계**:
- 발행 콘텐츠: {len(recent_published)}개
- CD 승인율: {stats['approval_rate']:.1f}%
- Capture → Publish 비율: {stats['capture_to_publish_rate']:.1f}%

**총 통계**:
- 총 Capture: {stats['stats']['captured']}
- 총 발행: {stats['stats']['published']}
- CD 승인: {stats['stats']['cd_approved']}
- CD 거부: {stats['stats']['cd_rejected']}

## 📋 이번주 계획

**콘텐츠 후보**:
(목요일에 자동 제안됩니다)

**사이클 점검**:
- Capture 활성화: {'✅' if stats['stats']['captured'] > 0 else '⚠️'}
- Connect 작동: {'✅' if stats['stats']['connected'] > 0 else '⚠️'}
- Meaning 생성: {'✅' if stats['stats']['meaning_generated'] > 0 else '⚠️'}
- Manifest 완료: {'✅' if stats['stats']['cd_approved'] + stats['stats']['cd_rejected'] > 0 else '⚠️'}
- Cycle 순환: {'✅' if stats['stats']['published'] > 0 else '⚠️'}

## 💡 제안

- {"정상 운영 중" if len(recent_published) > 0 else "⚠️ 지난주 발행 없음 - Capture 활성화 필요"}
- {"" if stats['approval_rate'] >= 50 else "⚠️ CD 승인율 낮음 - 초고 품질 개선 필요"}

---
**Generated**: {datetime.now().isoformat()}
**Next Meeting**: {(datetime.now() + timedelta(days=7)).strftime('%Y-%m-%d')}
"""

        # 저장
        with open(meeting_file, 'w', encoding='utf-8') as f:
            f.write(meeting_content)

        logger.info(f"🏛️ Council Meeting saved to {meeting_file}")

        # (Future: 텔레그램 알림)

        return {
            "meeting_date": meeting_date,
            "meeting_file": str(meeting_file),
            "stats": stats
        }

    async def suggest_content_candidates(self):
        """
        콘텐츠 후보 제안
        raw_signals/ 분석 → 높은 점수 5개 제안
        """
        logger.info("💡 Suggesting content candidates...")

        # raw_signals/ 최근 20개
        signal_files = list(self.junction_executor.raw_signals_dir.glob("rs-*.md"))
        signal_files.sort(reverse=True)
        signal_files = signal_files[:20]

        candidates = []

        for signal_file in signal_files:
            # SA 분석
            with open(signal_file, 'r', encoding='utf-8') as f:
                content = f.read()

            # Signal ID 추출
            signal_id = signal_file.stem

            # connections.json에서 점수 가져오기
            if self.junction_executor.connections_file.exists():
                with open(self.junction_executor.connections_file, 'r', encoding='utf-8') as f:
                    connections = json.load(f)

                if signal_id in connections:
                    sa_score = connections[signal_id].get("sa_score", 0)
                    philosophy = connections[signal_id].get("philosophy", "unknown")

                    if sa_score >= 60:  # 60점 이상만
                        candidates.append({
                            "signal_id": signal_id,
                            "score": sa_score,
                            "philosophy": philosophy,
                            "preview": content[:100]
                        })

        # 점수순 정렬, 상위 5개
        candidates.sort(key=lambda x: x["score"], reverse=True)
        candidates = candidates[:5]

        # 제안 파일 저장
        suggestion_date = datetime.now().strftime("%Y%m%d")
        suggestion_file = self.reports_dir / f"content_candidates_{suggestion_date}.json"

        with open(suggestion_file, 'w', encoding='utf-8') as f:
            json.dump({
                "date": datetime.now().isoformat(),
                "candidates": candidates,
                "note": "상위 5개 콘텐츠 후보"
            }, f, indent=2, ensure_ascii=False)

        logger.info(f"💡 Suggested {len(candidates)} candidates, saved to {suggestion_file}")

        # (Future: 텔레그램 알림)

        return {
            "candidates": candidates,
            "suggestion_file": str(suggestion_file)
        }

    def _check_quarterly_review(self):
        """
        분기 회고 체크
        매달 1일에만 실행
        """
        today = datetime.now()
        if today.day == 1 and today.month % 3 == 1:  # 1, 4, 7, 10월 1일
            asyncio.run(self.quarterly_review())

    async def quarterly_review(self):
        """
        분기 회고
        Cycle Protocol 건강성 체크
        """
        logger.info("📈 Quarterly Review started")

        quarter = (datetime.now().month - 1) // 3 + 1
        year = datetime.now().year

        # 통계 수집
        stats = self.junction_executor.get_stats()

        # 분기 리포트
        review = {
            "year": year,
            "quarter": quarter,
            "period": f"{year}Q{quarter}",
            "stats": stats,
            "analysis": {
                "capture_활성화": stats['stats']['captured'] > 0,
                "junction_성공률": stats['capture_to_publish_rate'],
                "cd_승인율": stats['approval_rate'],
                "분기_발행_목표": "12-24개",
                "실제_발행": stats['stats']['published']
            },
            "generated_at": datetime.now().isoformat()
        }

        # 저장
        review_file = self.reports_dir / f"quarterly_review_{year}Q{quarter}.json"
        with open(review_file, 'w', encoding='utf-8') as f:
            json.dump(review, f, indent=2, ensure_ascii=False)

        logger.info(f"📈 Quarterly Review saved to {review_file}")

        # (Future: 텔레그램 알림)

        return {
            "review": review,
            "review_file": str(review_file)
        }

    def stop(self):
        """스케줄 중지"""
        self.running = False
        logger.info("🛑 Cycle Manager stopped")


def main():
    """메인 실행"""
    manager = CycleManager()

    try:
        logger.info("🔄 Cycle Manager running... (Press Ctrl+C to stop)")
        manager.start()
    except KeyboardInterrupt:
        manager.stop()
        logger.info("✅ Cycle Manager stopped gracefully")


if __name__ == "__main__":
    main()
