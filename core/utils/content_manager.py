#!/usr/bin/env python3
"""
WOOHWAHAE Content Manager
Manual content input and management system

Provides fallback when Instagram API is unavailable
Allows direct content submission for brand validation
"""

import json
import os
from pathlib import Path
from datetime import datetime
from typing import Dict, Any, List, Optional
from enum import Enum


class ContentSource(Enum):
    """Content source types"""
    INSTAGRAM = "instagram"
    MANUAL = "manual"
    WEBSITE = "website"
    ARCHIVE = "archive"


class ContentManager:
    """
    Content Management System for WOOHWAHAE

    Handles:
    - Manual content input
    - Content storage and retrieval
    - Content validation and preparation
    - Integration with Brand Consultant
    """

    def __init__(self, storage_path: Optional[Path] = None):
        """Initialize Content Manager"""
        self.storage_path = storage_path or Path.home() / '.woohwahae' / 'content'
        self.storage_path.mkdir(parents=True, exist_ok=True)

        # Content database file
        self.db_file = self.storage_path / 'content_db.json'
        self.content_db = self._load_database()

        print(f"[Content Manager] Initialized with {len(self.content_db)} items")

    def _load_database(self) -> List[Dict[str, Any]]:
        """Load content database"""
        if self.db_file.exists():
            try:
                with open(self.db_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
            except Exception as e:
                print(f"[Content Manager] Error loading database: {e}")
                return []
        return []

    def _save_database(self):
        """Save content database"""
        try:
            with open(self.db_file, 'w', encoding='utf-8') as f:
                json.dump(self.content_db, f, ensure_ascii=False, indent=2)
            print(f"[Content Manager] Database saved: {len(self.content_db)} items")
        except Exception as e:
            print(f"[Content Manager] Error saving database: {e}")

    def add_manual_content(
        self,
        title: str,
        content: str,
        content_type: str = "text",
        hashtags: Optional[List[str]] = None,
        metadata: Optional[Dict[str, Any]] = None
    ) -> str:
        """
        Add manual content to the system

        Args:
            title: Content title or caption
            content: Main content body
            content_type: Type of content (text, service, philosophy, etc.)
            hashtags: Optional hashtags
            metadata: Additional metadata

        Returns:
            Content ID
        """
        # Use microseconds for unique IDs
        import time
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        unique_suffix = str(int(time.time() * 1000000) % 1000000)
        content_id = f"manual_{timestamp}_{unique_suffix}"

        content_item = {
            'id': content_id,
            'source': ContentSource.MANUAL.value,
            'title': title,
            'content': content,
            'content_type': content_type,
            'hashtags': hashtags or [],
            'metadata': metadata or {},
            'created_at': datetime.now().isoformat(),
            'status': 'pending'
        }

        self.content_db.append(content_item)
        self._save_database()

        print(f"[Content Manager] Added manual content: {content_id}")
        return content_id

    def add_instagram_content(
        self,
        caption: str,
        post_url: str,
        hashtags: List[str],
        media_type: str = "image",
        metadata: Optional[Dict[str, Any]] = None
    ) -> str:
        """
        Add Instagram content manually (when API unavailable)

        Args:
            caption: Post caption
            post_url: Instagram post URL
            hashtags: Post hashtags
            media_type: Type of media (image, video, carousel)
            metadata: Additional metadata

        Returns:
            Content ID
        """
        content_id = f"ig_{datetime.now().strftime('%Y%m%d_%H%M%S')}"

        content_item = {
            'id': content_id,
            'source': ContentSource.INSTAGRAM.value,
            'caption': caption,
            'url': post_url,
            'hashtags': hashtags,
            'media_type': media_type,
            'metadata': metadata or {},
            'created_at': datetime.now().isoformat(),
            'status': 'pending'
        }

        self.content_db.append(content_item)
        self._save_database()

        print(f"[Content Manager] Added Instagram content: {content_id}")
        return content_id

    def get_pending_content(self, limit: int = 10) -> List[Dict[str, Any]]:
        """Get pending content for processing"""
        pending = [
            item for item in self.content_db
            if item.get('status') == 'pending'
        ]
        return pending[:limit]

    def get_content_by_id(self, content_id: str) -> Optional[Dict[str, Any]]:
        """Get specific content by ID"""
        for item in self.content_db:
            if item['id'] == content_id:
                return item
        return None

    def update_content_status(
        self,
        content_id: str,
        status: str,
        result: Optional[Dict[str, Any]] = None
    ):
        """Update content processing status"""
        for item in self.content_db:
            if item['id'] == content_id:
                item['status'] = status
                item['updated_at'] = datetime.now().isoformat()
                if result:
                    item['processing_result'] = result
                self._save_database()
                print(f"[Content Manager] Updated {content_id} status to: {status}")
                return
        print(f"[Content Manager] Content {content_id} not found")

    def prepare_for_brand_consultant(self, content_id: str) -> Dict[str, Any]:
        """
        Prepare content for Brand Consultant processing

        Args:
            content_id: Content ID to prepare

        Returns:
            Formatted content for Brand Consultant
        """
        content = self.get_content_by_id(content_id)
        if not content:
            return {}

        # Format based on source
        if content['source'] == ContentSource.INSTAGRAM.value:
            return {
                'type': 'instagram_post',
                'source': '@woosunhokr',
                'data': {
                    'caption': content.get('caption', ''),
                    'url': content.get('url', ''),
                    'hashtags': content.get('hashtags', []),
                    'media_type': content.get('media_type', 'image')
                },
                'metadata': content.get('metadata', {})
            }

        elif content['source'] == ContentSource.MANUAL.value:
            return {
                'type': content.get('content_type', 'text'),
                'source': 'manual_input',
                'data': {
                    'title': content.get('title', ''),
                    'content': content.get('content', ''),
                    'hashtags': content.get('hashtags', [])
                },
                'metadata': content.get('metadata', {})
            }

        return content

    def get_statistics(self) -> Dict[str, Any]:
        """Get content statistics"""
        stats = {
            'total': len(self.content_db),
            'by_source': {},
            'by_status': {},
            'recent': []
        }

        # Count by source
        for item in self.content_db:
            source = item.get('source', 'unknown')
            stats['by_source'][source] = stats['by_source'].get(source, 0) + 1

            status = item.get('status', 'unknown')
            stats['by_status'][status] = stats['by_status'].get(status, 0) + 1

        # Get 5 most recent
        sorted_content = sorted(
            self.content_db,
            key=lambda x: x.get('created_at', ''),
            reverse=True
        )
        stats['recent'] = [
            {
                'id': item['id'],
                'title': item.get('title') or item.get('caption', '')[:50],
                'source': item.get('source'),
                'status': item.get('status'),
                'created_at': item.get('created_at')
            }
            for item in sorted_content[:5]
        ]

        return stats

    def load_sample_content(self):
        """Load sample WOOHWAHAE content for testing"""
        sample_content = [
            {
                'title': '#읽는미장',
                'content': """레이어 BOB 스타일에 텍스처 펌을 가미한 형태입니다.
젠더리스한 느낌이 있으면서도 자유로운 분위기가 특징으로,
이솝 헤어 폴리쉬를 사용하여 전체적으로 역동적인 느낌을 연출했습니다.

스타일링이나 관리도 비교적 간단하여 편리하게 유지할 수 있습니다.
컬의 탄력은 적은 편이지만, 대신 본연의 볼륨을 잡아주어 분위기를 살려냅니다.""",
                'content_type': 'service',
                'hashtags': ['읽는미장', 'woohwahae', 'slowlife', '헤어디자인', '젠더리스']
            },
            {
                'title': '미용과 정신 건강의 접점',
                'content': """미용은 단순히 외적인 변화를 넘어, 정신 건강에 깊은 영향을 미친다고 믿어왔습니다.
내가 나를 어떻게 바라보는가는 자기 존중감, 자기 표현, 그리고 치유와 자아 인식을 촉진하는 중요한 역할을 합니다.

우리가 거울 앞에서 보내는 시간은 단순한 미용 루틴이 아니라, 자신과 연결되는 순간입니다.
그 과정 자체가 치유가 될 수 있다고 생각하는 관점에서, 우리 아틀리에는 그런 공간이 되길 바랍니다.""",
                'content_type': 'philosophy',
                'hashtags': ['woohwahae', 'wellness', '슬로우라이프', '치유']
            },
            {
                'title': '작업 중 듣는 음악',
                'content': """조용한 오후, 아틀리에에서 흘러나오는 음악.

이번 달의 플레이리스트:
- 이랑 '신의 놀이'
- Bill Evans 'Waltz for Debby'
- Ryuichi Sakamoto 'Merry Christmas Mr. Lawrence'

음악은 공간의 분위기를 만들고, 시간의 흐름을 느리게 합니다.""",
                'content_type': 'playlist',
                'hashtags': ['woohwahae', 'playlist', '슬로우라이프음악']
            }
        ]

        for content in sample_content:
            self.add_manual_content(
                title=content['title'],
                content=content['content'],
                content_type=content['content_type'],
                hashtags=content['hashtags']
            )

        print(f"[Content Manager] Loaded {len(sample_content)} sample items")


# Singleton instance
_manager_instance: Optional[ContentManager] = None


def get_content_manager() -> ContentManager:
    """Get singleton Content Manager instance"""
    global _manager_instance

    if _manager_instance is None:
        _manager_instance = ContentManager()

    return _manager_instance


# ================== Standalone Execution ==================

if __name__ == '__main__':
    # Test Content Manager
    manager = get_content_manager()

    print("\n" + "="*60)
    print("WOOHWAHAE Content Manager Test")
    print("="*60)

    # Load sample content
    print("\n1. Loading sample content...")
    manager.load_sample_content()

    # Get statistics
    print("\n2. Content Statistics:")
    stats = manager.get_statistics()
    print(f"   Total items: {stats['total']}")
    print(f"   By source: {stats['by_source']}")
    print(f"   By status: {stats['by_status']}")

    # Get pending content
    print("\n3. Pending Content:")
    pending = manager.get_pending_content(5)
    for item in pending:
        print(f"   - {item['id']}: {item.get('title', 'Untitled')[:50]}")

    print("\n" + "="*60)
    print("Test Complete!")
    print("="*60)