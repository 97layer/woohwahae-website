#!/usr/bin/env python3
"""
Instagram Crawler for WOOHWAHAE
@woosunhokr 계정 콘텐츠 자동 수집 및 분류

Features:
- Instaloader를 사용한 공개 프로필 크롤링
- 포스트, 캡션, 해시태그, 이미지 자동 수집
- AI 기반 콘텐츠 분류 (7개 섹션)
- 로컬 캐싱으로 중복 수집 방지

Author: WOOHWAHAE System
Created: 2026-02-17
"""

import os
import json
import time
import hashlib
from pathlib import Path
from typing import Dict, Any, List, Optional, Tuple
from datetime import datetime, timedelta
import logging

# Instaloader for Instagram crawling
try:
    import instaloader
    INSTALOADER_AVAILABLE = True
except ImportError:
    INSTALOADER_AVAILABLE = False
    print("⚠️  instaloader not installed. Run: pip install instaloader")

# Image analysis
try:
    from PIL import Image
    PIL_AVAILABLE = True
except ImportError:
    PIL_AVAILABLE = False

logger = logging.getLogger(__name__)


class InstagramCrawler:
    """
    Instagram Crawler for @woosunhokr

    자동으로 포스트를 수집하고 WOOHWAHAE 7개 섹션에 분류
    """

    def __init__(self, username: str = "woosunhokr", cache_dir: Optional[str] = None):
        """
        Initialize Instagram Crawler

        Args:
            username: Instagram username to crawl
            cache_dir: Directory for caching (default: project_root/data/instagram_cache)
        """
        if not INSTALOADER_AVAILABLE:
            raise ImportError("instaloader required: pip install instaloader")

        self.username = username
        self.loader = instaloader.Instaloader(
            download_pictures=True,
            download_videos=False,
            download_video_thumbnails=False,
            download_geotags=False,
            download_comments=False,
            save_metadata=True,
            compress_json=False,
            post_metadata_txt_pattern="",  # JSON만 저장
            max_connection_attempts=3
        )

        # 캐시 디렉토리 설정
        if cache_dir:
            self.cache_dir = Path(cache_dir)
        else:
            project_root = Path(__file__).parent.parent.parent
            self.cache_dir = project_root / "data" / "instagram_cache" / username

        self.cache_dir.mkdir(parents=True, exist_ok=True)

        # 수집 기록
        self.history_file = self.cache_dir / "crawl_history.json"
        self.history = self._load_history()

        logger.info("Instagram Crawler initialized for @%s", username)

    def _load_history(self) -> Dict[str, Any]:
        """크롤링 기록 로드"""
        if self.history_file.exists():
            try:
                with open(self.history_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
            except (OSError, ValueError):
                pass
        return {
            'last_crawl': None,
            'posts_collected': {},
            'total_posts': 0
        }

    def _save_history(self):
        """크롤링 기록 저장"""
        self.history['last_crawl'] = datetime.now().isoformat()
        with open(self.history_file, 'w', encoding='utf-8') as f:
            json.dump(self.history, f, ensure_ascii=False, indent=2)

    def crawl_recent_posts(self, limit: int = 10, days_back: int = 7) -> List[Dict[str, Any]]:
        """
        최근 포스트 크롤링

        Args:
            limit: 최대 수집 개수
            days_back: 며칠 전까지 수집할지

        Returns:
            수집된 포스트 리스트
        """
        print(f"[Crawler] @{self.username} 최근 {days_back}일 포스트 수집 시작...")

        try:
            # 프로필 로드
            profile = instaloader.Profile.from_username(self.loader.context, self.username)
            posts_collected = []
            cutoff_date = datetime.now() - timedelta(days=days_back)

            # 포스트 순회
            for post in profile.get_posts():
                # 날짜 체크
                if post.date < cutoff_date:
                    break

                # 중복 체크
                post_id = str(post.shortcode)
                if post_id in self.history['posts_collected']:
                    logger.debug("Skip duplicate: %s", post_id)
                    continue

                # 포스트 데이터 추출
                post_data = self._extract_post_data(post)

                # 이미지 다운로드
                if post.is_video:
                    post_data['media_type'] = 'video'
                else:
                    post_data['media_type'] = 'image'
                    # 이미지 저장
                    image_path = self._download_image(post)
                    if image_path:
                        post_data['local_image_path'] = str(image_path)

                # 섹션 분류 힌트 추가
                post_data['section_hints'] = self._classify_post(post_data)

                # 캐시 저장
                self._save_post_cache(post_id, post_data)

                # 기록 업데이트
                self.history['posts_collected'][post_id] = {
                    'collected_at': datetime.now().isoformat(),
                    'caption_preview': post_data['caption'][:100]
                }

                posts_collected.append(post_data)

                print(f"[Crawler] 수집: {post_id} - {post_data['caption'][:50]}...")

                if len(posts_collected) >= limit:
                    break

                # Rate limiting
                time.sleep(2)

            self.history['total_posts'] = len(self.history['posts_collected'])
            self._save_history()

            print(f"[Crawler] 수집 완료: {len(posts_collected)}개 포스트")
            return posts_collected

        except Exception as e:
            logger.error("Crawling failed: %s", e)
            return []

    def _extract_post_data(self, post) -> Dict[str, Any]:
        """포스트 데이터 추출"""
        return {
            'post_id': post.shortcode,
            'url': f"https://www.instagram.com/p/{post.shortcode}/",
            'caption': post.caption if post.caption else "",
            'hashtags': list(post.caption_hashtags) if post.caption_hashtags else [],
            'mentions': list(post.caption_mentions) if post.caption_mentions else [],
            'date': post.date.isoformat(),
            'likes': post.likes,
            'comments': post.comments,
            'is_video': post.is_video,
            'location': post.location.name if post.location else None,
            'accessibility_caption': post.accessibility_caption,
            'crawled_at': datetime.now().isoformat()
        }

    def _download_image(self, post) -> Optional[Path]:
        """이미지 다운로드 및 저장"""
        try:
            image_dir = self.cache_dir / "images"
            image_dir.mkdir(exist_ok=True)

            # 이미지 파일명
            image_filename = f"{post.shortcode}.jpg"
            image_path = image_dir / image_filename

            if image_path.exists():
                return image_path

            # Instaloader로 다운로드
            self.loader.download_pic(
                filename=str(image_path.with_suffix('')),
                url=post.url,
                mtime=post.date
            )

            return image_path

        except Exception as e:
            logger.error("Image download failed: %s", e)
            return None

    def _classify_post(self, post_data: Dict[str, Any]) -> Dict[str, float]:
        """
        포스트를 7개 섹션으로 분류 (휴리스틱)

        Returns:
            섹션별 확률 점수
        """
        caption = post_data['caption'].lower()
        hashtags = [tag.lower() for tag in post_data.get('hashtags', [])]
        all_text = caption + ' ' + ' '.join(hashtags)

        scores = {
            'about': 0.0,
            'archive': 0.0,
            'shop': 0.0,
            'service': 0.0,
            'playlist': 0.0,
            'project': 0.0,
            'photography': 0.0
        }

        # Service (헤어 관련)
        hair_keywords = ['헤어', 'hair', '펌', 'perm', '컷', 'cut', '스타일', 'style', '미용']
        for keyword in hair_keywords:
            if keyword in all_text:
                scores['service'] += 20.0

        # Archive (매거진, 에세이)
        archive_keywords = ['archive', '아카이브', 'magazine', '매거진', '글', '에세이', '생각']
        for keyword in archive_keywords:
            if keyword in all_text:
                scores['archive'] += 20.0

        # Shop (제품 언급)
        shop_keywords = ['이솝', 'aesop', '제품', 'product', '밀본', 'milbon']
        for keyword in shop_keywords:
            if keyword in all_text:
                scores['shop'] += 15.0

        # Project (협업)
        project_keywords = ['협업', 'collaboration', '프로젝트', 'project', 'with']
        for keyword in project_keywords:
            if keyword in all_text:
                scores['project'] += 15.0

        # Photography (비주얼 중심)
        if post_data.get('media_type') == 'image' and len(caption) < 100:
            scores['photography'] += 25.0

        # Playlist (음악)
        music_keywords = ['음악', 'music', 'playlist', '플레이리스트', 'bgm']
        for keyword in music_keywords:
            if keyword in all_text:
                scores['playlist'] += 20.0

        # About (철학적)
        philosophy_keywords = ['slowlife', '슬로우라이프', '생각', '철학', 'philosophy', 'woohwahae']
        for keyword in philosophy_keywords:
            if keyword in all_text:
                scores['about'] += 15.0

        # 정규화 (합이 100이 되도록)
        total = sum(scores.values())
        if total > 0:
            for key in scores:
                scores[key] = round((scores[key] / total) * 100, 2)

        return scores

    def _save_post_cache(self, post_id: str, post_data: Dict[str, Any]):
        """포스트 데이터 캐시 저장"""
        post_file = self.cache_dir / f"{post_id}.json"
        with open(post_file, 'w', encoding='utf-8') as f:
            json.dump(post_data, f, ensure_ascii=False, indent=2)

    def get_cached_posts(self) -> List[Dict[str, Any]]:
        """캐시된 모든 포스트 반환"""
        posts = []
        for json_file in self.cache_dir.glob("*.json"):
            if json_file.name != "crawl_history.json":
                try:
                    with open(json_file, 'r', encoding='utf-8') as f:
                        posts.append(json.load(f))
                except (OSError, ValueError):
                    continue
        return posts

    def analyze_hashtag_trends(self) -> Dict[str, int]:
        """해시태그 트렌드 분석"""
        hashtag_counts = {}
        posts = self.get_cached_posts()

        for post in posts:
            for tag in post.get('hashtags', []):
                tag_lower = tag.lower()
                hashtag_counts[tag_lower] = hashtag_counts.get(tag_lower, 0) + 1

        # 빈도순 정렬
        return dict(sorted(hashtag_counts.items(), key=lambda x: x[1], reverse=True))

    def get_section_distribution(self) -> Dict[str, float]:
        """섹션별 콘텐츠 분포 분석"""
        section_totals = {
            'about': 0.0,
            'archive': 0.0,
            'shop': 0.0,
            'service': 0.0,
            'playlist': 0.0,
            'project': 0.0,
            'photography': 0.0
        }

        posts = self.get_cached_posts()
        for post in posts:
            hints = post.get('section_hints', {})
            for section, score in hints.items():
                section_totals[section] += score

        # 평균 계산
        if posts:
            for section in section_totals:
                section_totals[section] = round(section_totals[section] / len(posts), 2)

        return section_totals


# ================== Standalone Execution ==================

if __name__ == '__main__':
    import argparse

    parser = argparse.ArgumentParser(description='Instagram Crawler for WOOHWAHAE')
    parser.add_argument('--username', default='woosunhokr', help='Instagram username')
    parser.add_argument('--limit', type=int, default=10, help='Max posts to crawl')
    parser.add_argument('--days', type=int, default=7, help='Days back to crawl')
    parser.add_argument('--analyze', action='store_true', help='Analyze cached posts')

    args = parser.parse_args()

    # 크롤러 초기화
    crawler = InstagramCrawler(username=args.username)

    if args.analyze:
        # 캐시된 포스트 분석
        print("\n" + "="*50)
        print("Instagram Content Analysis")
        print("="*50 + "\n")

        # 섹션 분포
        print("📊 Section Distribution:")
        distribution = crawler.get_section_distribution()
        for section, avg_score in sorted(distribution.items(), key=lambda x: x[1], reverse=True):
            bar = "█" * int(avg_score / 5)
            print(f"  {section:12} {bar} {avg_score}%")

        # 해시태그 트렌드
        print("\n🏷️  Top Hashtags:")
        trends = crawler.analyze_hashtag_trends()
        for tag, count in list(trends.items())[:10]:
            print(f"  #{tag:20} {count}회")

        # 캐시 상태
        cached_posts = crawler.get_cached_posts()
        print(f"\n📦 Cached Posts: {len(cached_posts)}개")

    else:
        # 새 포스트 크롤링
        print(f"\n🔍 Crawling @{args.username}...")
        print(f"   최대 {args.limit}개, 최근 {args.days}일")
        print("-" * 50)

        posts = crawler.crawl_recent_posts(limit=args.limit, days_back=args.days)

        print(f"\n✅ 수집 완료: {len(posts)}개 포스트")

        if posts:
            print("\n최근 포스트:")
            for post in posts[:3]:
                print(f"\n  📝 {post['post_id']}")
                print(f"     {post['caption'][:100]}...")
                print(f"     섹션 힌트: {post.get('section_hints', {})}")