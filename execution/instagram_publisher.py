#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Filename: execution/instagram_publisher.py
Author: 97LAYER Mercenary
Date: 2026-02-14
Description: Instagram 발행 큐 실행 스크립트 - CD 승인 후 예약된 콘텐츠 발행
"""

import os
import json
import sys
from pathlib import Path
from datetime import datetime
from typing import Dict, List

# Base directory
BASE_DIR = Path(__file__).resolve().parent.parent

# Add to path
if str(BASE_DIR) not in sys.path:
    sys.path.insert(0, str(BASE_DIR))

from execution.auto_publisher import AutoPublisher
from libs.core_config import INSTAGRAM_CONFIG


def check_publish_queue() -> List[Dict]:
    """
    발행 큐 확인 및 발행 대상 추출

    Returns:
        List of publish items ready to be published
    """
    queue_file = Path(INSTAGRAM_CONFIG["PUBLISH_QUEUE_PATH"])

    if not queue_file.exists():
        return []

    with open(queue_file, "r", encoding="utf-8") as f:
        queue = json.load(f)

    # 현재 시간 이전에 예약된 항목만 필터링
    now = datetime.now()
    ready_items = []

    for item in queue:
        scheduled_time = datetime.fromisoformat(item["scheduled_time"])
        if scheduled_time <= now and item["status"] == "scheduled":
            ready_items.append(item)

    return ready_items


def publish_scheduled_items():
    """
    예약된 항목 발행 및 큐 업데이트
    """
    publisher = AutoPublisher()
    ready_items = check_publish_queue()

    if not ready_items:
        print(f"[{datetime.now()}] 발행 대기 중인 콘텐츠 없음.")
        return

    print(f"[{datetime.now()}] 발행 대상 {len(ready_items)}건 발견")

    # 발행 실행
    results = []
    for item in ready_items:
        print(f"[{datetime.now()}] 발행 시작: {item['id']}")
        result = publisher.publish_to_instagram(item)

        if result.get("success"):
            print(f"[{datetime.now()}] ✅ 발행 완료: {result.get('post_url')}")
            item["status"] = "published"
            item["published_time"] = result.get("published_time")
            item["post_url"] = result.get("post_url")
        else:
            print(f"[{datetime.now()}] ❌ 발행 실패: {result.get('error')}")
            item["status"] = "failed"
            item["error"] = result.get("error")

        results.append(item)

    # 큐 업데이트 (발행된 항목 제거, 실패한 항목은 유지)
    queue_file = Path(INSTAGRAM_CONFIG["PUBLISH_QUEUE_PATH"])
    with open(queue_file, "r", encoding="utf-8") as f:
        full_queue = json.load(f)

    # 성공한 항목만 제거
    published_ids = [item["id"] for item in results if item["status"] == "published"]
    updated_queue = [item for item in full_queue if item["id"] not in published_ids]

    with open(queue_file, "w", encoding="utf-8") as f:
        json.dump(updated_queue, f, indent=2, ensure_ascii=False)

    # 발행 이력 저장
    history_dir = BASE_DIR / "knowledge" / "assets" / "published"
    history_file = history_dir / "publish_history.json"

    history = []
    if history_file.exists():
        with open(history_file, "r", encoding="utf-8") as f:
            history = json.load(f)

    history.extend(results)

    with open(history_file, "w", encoding="utf-8") as f:
        json.dump(history, f, indent=2, ensure_ascii=False)

    print(f"[{datetime.now()}] 발행 프로세스 완료. 성공: {len(published_ids)}, 실패: {len(results) - len(published_ids)}")


def main():
    """메인 실행"""
    print(f"[{datetime.now()}] Instagram Publisher 시작...")

    try:
        publish_scheduled_items()
    except Exception as e:
        print(f"[{datetime.now()}] Error: {e}")
        import traceback
        traceback.print_exc()


if __name__ == "__main__":
    main()
