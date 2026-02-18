#!/usr/bin/env python3
"""
WOOHWAHAE Content Pipeline
Instagram → Brand Consultant → 7 Sections → Website

자동화 파이프라인:
1. Instagram @woosunhokr 크롤링
2. Brand Consultant 철학 검증
3. 7개 섹션 분류 및 배포
4. 웹사이트 자동 업데이트

Author: WOOHWAHAE System
Created: 2026-02-17
"""

import os
import sys
import json
import asyncio
import logging
from pathlib import Path
from typing import Dict, Any, List, Optional
from datetime import datetime, timedelta
import schedule
import time

# Project setup
PROJECT_ROOT = Path(__file__).parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from core.utils.instagram_crawler import InstagramCrawler
from core.agents.brand_consultant import BrandConsultant, WOOHWAHAESection
from core.system.queue_manager import QueueManager, Task

logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')


class WOOHWAHAEPipeline:
    """
    WOOHWAHAE 콘텐츠 파이프라인

    Instagram 포스트를 자동으로 수집하여
    7개 섹션에 맞게 분류하고 웹사이트에 배포
    """

    def __init__(self, instagram_username: str = "woosunhokr"):
        """
        Initialize Pipeline

        Args:
            instagram_username: Instagram account to crawl
        """
        self.instagram_username = instagram_username
        self.project_root = PROJECT_ROOT

        # Components
        self.crawler = InstagramCrawler(username=instagram_username)
        self.consultant = BrandConsultant()
        self.queue_manager = QueueManager(str(PROJECT_ROOT))

        # Pipeline state
        self.pipeline_dir = PROJECT_ROOT / "data" / "pipeline"
        self.pipeline_dir.mkdir(parents=True, exist_ok=True)

        self.state_file = self.pipeline_dir / "pipeline_state.json"
        self.state = self._load_state()

        # Section content directories
        self.content_dirs = {
            WOOHWAHAESection.ABOUT: PROJECT_ROOT / "website" / "content" / "about",
            WOOHWAHAESection.ARCHIVE: PROJECT_ROOT / "website" / "content" / "archive",
            WOOHWAHAESection.SHOP: PROJECT_ROOT / "website" / "content" / "shop",
            WOOHWAHAESection.SERVICE: PROJECT_ROOT / "website" / "content" / "service",
            WOOHWAHAESection.PLAYLIST: PROJECT_ROOT / "website" / "content" / "playlist",
            WOOHWAHAESection.PROJECT: PROJECT_ROOT / "website" / "content" / "project",
            WOOHWAHAESection.PHOTOGRAPHY: PROJECT_ROOT / "website" / "content" / "photography"
        }

        # Create directories
        for dir_path in self.content_dirs.values():
            dir_path.mkdir(parents=True, exist_ok=True)

        logger.info(f"WOOHWAHAE Pipeline initialized for @{instagram_username}")

    def _load_state(self) -> Dict[str, Any]:
        """파이프라인 상태 로드"""
        if self.state_file.exists():
            try:
                with open(self.state_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
            except:
                pass

        return {
            'last_run': None,
            'posts_processed': 0,
            'section_counts': {},
            'errors': []
        }

    def _save_state(self):
        """파이프라인 상태 저장"""
        self.state['last_run'] = datetime.now().isoformat()
        with open(self.state_file, 'w', encoding='utf-8') as f:
            json.dump(self.state, f, ensure_ascii=False, indent=2)

    async def run_pipeline(self, days_back: int = 7, limit: int = 10):
        """
        파이프라인 실행

        Args:
            days_back: 며칠 전까지 크롤링할지
            limit: 최대 처리 포스트 수
        """
        logger.info("="*50)
        logger.info("WOOHWAHAE Pipeline Starting")
        logger.info("="*50)

        try:
            # Step 1: Instagram 크롤링
            logger.info(f"Step 1: Crawling @{self.instagram_username}...")
            posts = self.crawler.crawl_recent_posts(limit=limit, days_back=days_back)
            logger.info(f"  → {len(posts)} posts collected")

            if not posts:
                logger.info("No new posts to process")
                return

            # Step 2: 각 포스트 처리
            for post in posts:
                await self._process_post(post)

            # Step 3: 상태 저장
            self._save_state()

            # Step 4: 리포트 생성
            self._generate_report()

            logger.info("="*50)
            logger.info("Pipeline completed successfully")
            logger.info("="*50)

        except Exception as e:
            logger.error(f"Pipeline failed: {e}")
            self.state['errors'].append({
                'timestamp': datetime.now().isoformat(),
                'error': str(e)
            })
            self._save_state()

    async def _process_post(self, post_data: Dict[str, Any]):
        """
        개별 포스트 처리

        1. Brand Consultant 검증
        2. 섹션 분류
        3. 콘텐츠 생성
        4. 배포
        """
        post_id = post_data.get('post_id', 'unknown')
        logger.info(f"\nProcessing post: {post_id}")

        try:
            # Step 1: 브랜드 철학 검증
            logger.info("  → Brand audit...")
            audit_result = await self.consultant.audit_content({
                'type': 'instagram_post',
                'source': f'@{self.instagram_username}',
                'data': post_data
            })

            philosophy_score = audit_result.get('philosophy_score', 0)
            logger.info(f"    Philosophy score: {philosophy_score}/100")

            if philosophy_score < 50:
                logger.info("    ❌ Rejected: Low philosophy alignment")
                return

            # Step 2: 섹션 분류
            logger.info("  → Section classification...")
            sections = await self.consultant.classify_for_sections({
                'type': 'instagram_post',
                'data': post_data
            })

            if not sections:
                logger.info("    ❌ No suitable section found")
                return

            primary_section = sections[0][0]
            section_score = sections[0][1]
            logger.info(f"    Primary section: {primary_section.value} ({section_score}%)")

            # Step 3: 콘텐츠 생성 및 저장
            content = self._create_content(post_data, primary_section, audit_result)
            self._save_content(content, primary_section)

            # Step 4: 통계 업데이트
            self.state['posts_processed'] += 1
            section_name = primary_section.value
            self.state['section_counts'][section_name] = \
                self.state['section_counts'].get(section_name, 0) + 1

            logger.info(f"    ✅ Successfully processed to {primary_section.value}")

        except Exception as e:
            logger.error(f"  ❌ Failed to process post {post_id}: {e}")

    def _create_content(self, post_data: Dict[str, Any],
                       section: WOOHWAHAESection,
                       audit_result: Dict[str, Any]) -> Dict[str, Any]:
        """
        섹션별 콘텐츠 생성
        """
        content = {
            'post_id': post_data.get('post_id'),
            'url': post_data.get('url'),
            'date': post_data.get('date'),
            'section': section.value,
            'philosophy_score': audit_result.get('philosophy_score'),
            'created_at': datetime.now().isoformat()
        }

        # 섹션별 특화 콘텐츠
        if section == WOOHWAHAESection.SERVICE:
            # 헤어 서비스 관련
            content.update({
                'type': 'hair_work',
                'title': self._extract_title(post_data.get('caption', '')),
                'description': post_data.get('caption', ''),
                'image': post_data.get('local_image_path'),
                'hashtags': post_data.get('hashtags', [])
            })

        elif section == WOOHWAHAESection.ARCHIVE:
            # 매거진 아카이브
            content.update({
                'type': 'magazine_entry',
                'title': self._extract_title(post_data.get('caption', '')),
                'essay': post_data.get('caption', ''),
                'themes': audit_result.get('themes', []),
                'image': post_data.get('local_image_path')
            })

        elif section == WOOHWAHAESection.PHOTOGRAPHY:
            # 사진 갤러리
            content.update({
                'type': 'photo',
                'caption': post_data.get('caption', ''),
                'image': post_data.get('local_image_path'),
                'alt_text': post_data.get('accessibility_caption', '')
            })

        # 다른 섹션들도 추가...

        return content

    def _extract_title(self, caption: str) -> str:
        """캡션에서 제목 추출"""
        lines = caption.split('\n')
        for line in lines:
            line = line.strip()
            if line and not line.startswith('#'):
                return line[:50]
        return "Untitled"

    def _save_content(self, content: Dict[str, Any], section: WOOHWAHAESection):
        """콘텐츠를 섹션별 디렉토리에 저장"""
        content_dir = self.content_dirs[section]

        # 파일명 생성 (날짜-포스트ID)
        post_id = content.get('post_id', 'unknown')
        date_str = datetime.now().strftime('%Y%m%d')
        filename = f"{date_str}-{post_id}.json"

        filepath = content_dir / filename

        with open(filepath, 'w', encoding='utf-8') as f:
            json.dump(content, f, ensure_ascii=False, indent=2)

        logger.debug(f"Content saved to {filepath}")

    def _generate_report(self):
        """파이프라인 실행 리포트 생성"""
        report = {
            'timestamp': datetime.now().isoformat(),
            'posts_processed': self.state['posts_processed'],
            'section_distribution': self.state['section_counts'],
            'last_run': self.state['last_run']
        }

        report_file = self.pipeline_dir / f"report_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
        with open(report_file, 'w', encoding='utf-8') as f:
            json.dump(report, f, ensure_ascii=False, indent=2)

        # 콘솔 출력
        print("\n" + "="*50)
        print("Pipeline Report")
        print("="*50)
        print(f"Posts processed: {report['posts_processed']}")
        print("\nSection distribution:")
        for section, count in report['section_distribution'].items():
            print(f"  {section:15} : {count} posts")
        print("="*50)

    def schedule_daily(self, hour: int = 0, minute: int = 0):
        """
        매일 정해진 시간에 파이프라인 실행

        Args:
            hour: 실행 시간 (0-23)
            minute: 실행 분 (0-59)
        """
        schedule_time = f"{hour:02d}:{minute:02d}"
        schedule.every().day.at(schedule_time).do(
            lambda: asyncio.run(self.run_pipeline())
        )

        logger.info(f"Pipeline scheduled daily at {schedule_time}")

        while True:
            schedule.run_pending()
            time.sleep(60)  # Check every minute


# ================== Standalone Execution ==================

if __name__ == '__main__':
    import argparse

    parser = argparse.ArgumentParser(description='WOOHWAHAE Content Pipeline')
    parser.add_argument('--username', default='woosunhokr', help='Instagram username')
    parser.add_argument('--days', type=int, default=7, help='Days back to crawl')
    parser.add_argument('--limit', type=int, default=10, help='Max posts to process')
    parser.add_argument('--schedule', action='store_true', help='Run daily schedule')

    args = parser.parse_args()

    # Initialize pipeline
    pipeline = WOOHWAHAEPipeline(instagram_username=args.username)

    if args.schedule:
        # 매일 자정 실행
        print("Starting daily schedule (00:00)...")
        pipeline.schedule_daily(hour=0, minute=0)
    else:
        # 즉시 실행
        print(f"Running pipeline for @{args.username}...")
        asyncio.run(pipeline.run_pipeline(
            days_back=args.days,
            limit=args.limit
        ))