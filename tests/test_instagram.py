#!/usr/bin/env python3
"""
Instagram Crawler Test
Tests the Instagram content crawling capability
"""

import sys
import json
import asyncio
from pathlib import Path
from datetime import datetime, timedelta

# Add project root to path
PROJECT_ROOT = Path(__file__).parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from core.utils.instagram_crawler import InstagramCrawler


async def test_crawler():
    """Test Instagram crawler"""

    print("="*60)
    print("Instagram Crawler Test")
    print("="*60)

    # Initialize crawler
    crawler = InstagramCrawler()

    # Test 1: Check if login is possible
    print("\n1. Checking Instagram connection...")
    try:
        # Try to get profile info
        crawler.loader.load_session_from_file("@woosunhokr")
        print("   ✅ Session loaded (may need login)")
    except Exception as e:
        print(f"   ⚠️  Session not found, will try to login: {e}")

    # Test 2: Crawl recent posts
    print("\n2. Crawling recent posts from @woosunhokr...")
    posts = await crawler.crawl_recent_posts(
        limit=5,
        days_back=30
    )

    if posts:
        print(f"   ✅ Found {len(posts)} posts")

        # Show sample post
        if posts:
            post = posts[0]
            print(f"\n   Sample post:")
            print(f"   - Date: {post.get('date')}")
            print(f"   - Caption: {post.get('caption', '')[:100]}...")
            print(f"   - Hashtags: {post.get('hashtags', [])}")
            print(f"   - URL: {post.get('url')}")

            # Test classification
            sections = post.get('sections', {})
            if sections:
                print(f"   - Classified sections:")
                for section, score in list(sections.items())[:3]:
                    print(f"     - {section}: {score}%")
    else:
        print("   ❌ No posts found (may need authentication)")
        print("\n   To authenticate:")
        print("   1. Create a session file manually")
        print("   2. Or use InstagramCrawler with login credentials")
        print("   3. Or use manual content input instead")

    # Test 3: Test hashtag analysis
    print("\n3. Testing hashtag analysis...")
    test_hashtags = ["woohwahae", "slowlife", "읽는미장", "헤어디자인"]

    classification = crawler._classify_post({
        'caption': '#읽는미장 테스트 포스트',
        'hashtags': test_hashtags
    })

    print(f"   Hashtag-based classification:")
    for section, score in list(classification.items())[:3]:
        if score > 0:
            print(f"   - {section}: {score}%")

    print("\n" + "="*60)
    print("Test Complete!")
    print("="*60)

    # Return results
    return {
        'posts_found': len(posts) if posts else 0,
        'authentication_needed': len(posts) == 0,
        'sample_post': posts[0] if posts else None
    }


if __name__ == '__main__':
    # Run test
    result = asyncio.run(test_crawler())

    # Save result to file for inspection
    output_file = PROJECT_ROOT / 'tests' / 'instagram_test_result.json'
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(result, f, ensure_ascii=False, indent=2)

    print(f"\nResults saved to: {output_file}")