#!/usr/bin/env python3
"""
97layerOS Auto Reporter
매일 아침/저녁 자동 보고

Features:
- 아침 9시: 일일 브리핑
- 저녁 9시: 데일리 요약
- 주간 리포트

Author: 97layerOS Technical Director
Created: 2026-02-16
"""

import os
import sys
import json
import logging
from pathlib import Path
from typing import Dict, List
from datetime import datetime, timedelta
from apscheduler.schedulers.blocking import BlockingScheduler
from apscheduler.triggers.cron import CronTrigger

# Add project root to path
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

# Load .env
try:
    from dotenv import load_dotenv
    load_dotenv(PROJECT_ROOT / '.env')
except ImportError:
    pass

logger = logging.getLogger(__name__)


class AutoReporter:
    """
    자동 보고 시스템
    """

    def __init__(self):
        """Initialize Auto Reporter"""
        self.signals_dir = PROJECT_ROOT / 'knowledge' / 'signals'
        self.reports_dir = PROJECT_ROOT / 'knowledge' / 'reports'
        self.reports_dir.mkdir(parents=True, exist_ok=True)

        logger.info("✅ Auto Reporter initialized")

    def generate_morning_briefing(self):
        """아침 브리핑 생성 (09:00)"""
        logger.info("☀️ Generating morning briefing...")

        today = datetime.now().strftime('%Y-%m-%d')

        # 어제부터 오늘까지의 신호 수집
        yesterday = (datetime.now() - timedelta(days=1)).strftime('%Y%m%d')
        today_str = datetime.now().strftime('%Y%m%d')

        signals = self._collect_signals_by_date([yesterday, today_str])

        report = self._create_briefing_report(signals, "morning")

        # 보고서 저장
        report_file = self.reports_dir / f"morning_briefing_{today}.md"
        with open(report_file, 'w', encoding='utf-8') as f:
            f.write(report)

        logger.info("✅ Morning briefing saved: %s", report_file)

        # TODO: Telegram으로 전송
        return report

    def generate_evening_summary(self):
        """저녁 요약 생성 (21:00)"""
        logger.info("🌙 Generating evening summary...")

        today = datetime.now().strftime('%Y-%m-%d')
        today_str = datetime.now().strftime('%Y%m%d')

        signals = self._collect_signals_by_date([today_str])

        report = self._create_briefing_report(signals, "evening")

        # 보고서 저장
        report_file = self.reports_dir / f"evening_summary_{today}.md"
        with open(report_file, 'w', encoding='utf-8') as f:
            f.write(report)

        logger.info("✅ Evening summary saved: %s", report_file)

        # TODO: Telegram으로 전송
        return report

    def _collect_signals_by_date(self, date_strings: List[str]) -> Dict:
        """날짜별 신호 수집"""
        signals = {
            'youtube': [],
            'images': [],
            'texts': []
        }

        for date_str in date_strings:
            # YouTube 신호
            youtube_files = list(self.signals_dir.glob(f'youtube_*_{date_str}_*.json'))
            for file in youtube_files:
                try:
                    with open(file, 'r', encoding='utf-8') as f:
                        data = json.load(f)
                        signals['youtube'].append(data)
                except Exception as e:
                    logger.error("Error reading %s: %s", file, e)

            # 이미지 신호
            images_dir = self.signals_dir / 'images'
            if images_dir.exists():
                image_files = list(images_dir.glob(f'image_*_{date_str}_*.json'))
                for file in image_files:
                    try:
                        with open(file, 'r', encoding='utf-8') as f:
                            data = json.load(f)
                            signals['images'].append(data)
                    except Exception as e:
                        logger.error("Error reading %s: %s", file, e)

            # 텍스트 신호
            text_files = list(self.signals_dir.glob(f'text_{date_str}_*.json'))
            for file in text_files:
                try:
                    with open(file, 'r', encoding='utf-8') as f:
                        data = json.load(f)
                        signals['texts'].append(data)
                except Exception as e:
                    logger.error("Error reading %s: %s", file, e)

        return signals

    def _create_briefing_report(self, signals: Dict, report_type: str) -> str:
        """브리핑 보고서 생성"""
        now = datetime.now()
        date_str = now.strftime('%Y-%m-%d')
        time_str = now.strftime('%H:%M')

        if report_type == "morning":
            title = "☀️ 아침 브리핑"
            period = "어제부터 오늘까지"
        else:
            title = "🌙 저녁 요약"
            period = "오늘 하루"

        youtube_count = len(signals['youtube'])
        images_count = len(signals['images'])
        texts_count = len(signals['texts'])
        total_count = youtube_count + images_count + texts_count

        report = f"""# {title}

**날짜**: {date_str} {time_str}
**기간**: {period}

---

## 📊 활동 요약

- 🎥 YouTube 분석: {youtube_count}개
- 📷 이미지 분석: {images_count}개
- 💬 텍스트 인사이트: {texts_count}개
- **총 신호**: {total_count}개

---

## 🎥 YouTube 분석

"""

        if signals['youtube']:
            for idx, signal in enumerate(signals['youtube'], 1):
                video_id = signal.get('video_id', 'unknown')
                source = signal.get('source', '')
                status = signal.get('status', 'unknown')
                transcript_length = signal.get('full_transcript_length', 0)

                report += f"""### {idx}. Video: {video_id}

- **링크**: {source}
- **자막 길이**: {transcript_length} 글자
- **상태**: {status}
- **처리 시간**: {signal.get('captured_at', 'N/A')}

"""
        else:
            report += "_오늘은 YouTube 분석이 없습니다._\n\n"

        report += """---

## 📷 이미지 분석

"""

        if signals['images']:
            for idx, signal in enumerate(signals['images'], 1):
                description = signal.get('analysis', {}).get('description', '')[:100]
                status = signal.get('status', 'unknown')

                report += f"""### {idx}. 이미지

- **설명**: {description}...
- **상태**: {status}
- **처리 시간**: {signal.get('captured_at', 'N/A')}

"""
        else:
            report += "_오늘은 이미지 분석이 없습니다._\n\n"

        report += """---

## 💬 텍스트 인사이트

"""

        if signals['texts']:
            for idx, signal in enumerate(signals['texts'], 1):
                content = signal.get('content', '')[:100]

                report += f"""### {idx}. {content}...

"""
        else:
            report += "_오늘은 텍스트 인사이트가 없습니다._\n\n"

        report += """---

## 🎯 다음 액션

- Multi-Agent 분석 대기 중: {pending} 건
- 자동 처리 예정

---

_Generated by 97layer AI Secretary_
_Report Type: {type}_
""".format(pending=total_count, type=report_type)

        return report

    def start_scheduled_reports(self):
        """스케줄된 보고 시작"""
        logger.info("📅 Starting scheduled reports...")

        scheduler = BlockingScheduler()

        # 아침 9시 브리핑
        scheduler.add_job(
            self.generate_morning_briefing,
            CronTrigger(hour=9, minute=0),
            id='morning_briefing',
            name='Morning Briefing (09:00)',
            replace_existing=True
        )

        # 저녁 9시 요약
        scheduler.add_job(
            self.generate_evening_summary,
            CronTrigger(hour=21, minute=0),
            id='evening_summary',
            name='Evening Summary (21:00)',
            replace_existing=True
        )

        logger.info("✅ Scheduled:")
        logger.info("  - Morning Briefing: 09:00")
        logger.info("  - Evening Summary: 21:00")

        try:
            scheduler.start()
        except (KeyboardInterrupt, SystemExit):
            logger.info("🛑 Scheduler stopped")


def main():
    """Main entry point"""
    # Configure logging
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )

    logger.info("🚀 Starting Auto Reporter...")

    reporter = AutoReporter()

    # 테스트: 즉시 보고서 생성
    print("\n" + "="*60)
    print("Testing Morning Briefing:")
    print("="*60)
    report = reporter.generate_morning_briefing()
    print(report)

    print("\n" + "="*60)
    print("Testing Evening Summary:")
    print("="*60)
    report = reporter.generate_evening_summary()
    print(report)

    # 스케줄 시작 (프로덕션)
    # reporter.start_scheduled_reports()


if __name__ == "__main__":
    main()
