#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Filename: execution/auto_publisher.py
Author: 97LAYER Mercenary
Date: 2026-02-14
Description: Imperfect Publish Protocol - 72시간 규칙 + 자동 발행 시스템
"""

import os
import json
import time
from pathlib import Path
from datetime import datetime, timedelta
from typing import Optional, Dict, List

# Base directory
BASE_DIR = Path(__file__).resolve().parent.parent

# Paths
DRAFT_DIR = BASE_DIR / "knowledge" / "assets" / "draft"
READY_DIR = BASE_DIR / "knowledge" / "assets" / "ready_to_publish"
PUBLISHED_DIR = BASE_DIR / "knowledge" / "assets" / "published"
DISCARDED_DIR = BASE_DIR / "knowledge" / "assets" / "discarded"

# Create directories
for d in [DRAFT_DIR, READY_DIR, PUBLISHED_DIR, DISCARDED_DIR]:
    d.mkdir(parents=True, exist_ok=True)


class AutoPublisher:
    """
    Imperfect Publish Protocol 자동 실행
    - 72시간 규칙 체크
    - CD 승인 콘텐츠 자동 예약
    - Instagram API 연동 (준비)
    """

    def __init__(self):
        self.now = datetime.now()

    def check_72h_rule(self) -> List[Dict]:
        """
        Draft 폴더 스캔하여 72시간 경과 파일 체크

        Returns:
            List of dicts with file info and elapsed time
        """
        violations = []

        if not DRAFT_DIR.exists():
            return violations

        for draft_file in DRAFT_DIR.glob("*.md"):
            # 파일 수정 시간 사용 (macOS에서 st_ctime은 metadata change time)
            created_time = datetime.fromtimestamp(draft_file.stat().st_mtime)
            elapsed = self.now - created_time

            # 메타데이터에서 생성 시간 확인 (우선순위: created > date_created)
            metadata = self._read_metadata(draft_file)
            if metadata and "created" in metadata:
                try:
                    created_time = datetime.fromisoformat(metadata["created"])
                    elapsed = self.now - created_time
                except (ValueError, TypeError):
                    pass  # Use file mtime
            elif metadata and "date_created" in metadata:
                try:
                    created_time = datetime.fromisoformat(metadata["date_created"])
                    elapsed = self.now - created_time
                except (ValueError, TypeError):
                    pass

            hours_elapsed = elapsed.total_seconds() / 3600

            # 72시간 (3일) 체크
            if hours_elapsed > 72:
                violations.append({
                    "file": draft_file.name,
                    "path": str(draft_file),
                    "created": created_time.isoformat(),
                    "elapsed_hours": round(hours_elapsed, 1),
                    "status": "violation" if hours_elapsed > 76 else "warning"
                })

        return violations

    def auto_discard(self, file_path: str) -> bool:
        """
        76시간(72h + 4h 유예) 경과 시 자동 폐기

        Args:
            file_path: Path to draft file

        Returns:
            True if discarded successfully
        """
        try:
            source = Path(file_path)
            if not source.exists():
                return False

            # Discard 폴더로 이동
            dest = DISCARDED_DIR / f"{source.stem}_{int(time.time())}{source.suffix}"
            source.rename(dest)

            # 로그
            log_msg = f"[{self.now}] Auto-discarded: {source.name} (76h+ elapsed)"
            self._log(log_msg)

            return True
        except Exception as e:
            print(f"Auto-discard failed: {e}")
            return False

    def notify_cd(self, violations: List[Dict]) -> str:
        """
        CD에게 72시간 경과 알림 생성

        Args:
            violations: List of violation dicts

        Returns:
            Formatted notification message
        """
        if not violations:
            return ""

        msg = "⏰ [TD → CD] 72시간 규칙 위반 감지\n\n"

        for v in violations:
            status_icon = "🚨" if v["status"] == "violation" else "⚠️"
            msg += f"{status_icon} {v['file']}\n"
            msg += f"   생성: {v['created'][:10]}\n"
            msg += f"   경과: {v['elapsed_hours']}h\n"

            if v["status"] == "violation":
                msg += f"   → 자동 폐기 예정 (4시간 유예 초과)\n"
            else:
                msg += f"   → CD 즉시 결정 필요 (4시간 유예 중)\n"
            msg += "\n"

        msg += "[Imperfect Publish Protocol]\n"
        msg += "MBQ 3가지 충족 시 즉시 승인.\n"
        msg += "의심스러우면 발행.\n"

        return msg

    def schedule_publish(self, content_file: str, image_file: Optional[str] = None,
                        schedule_time: Optional[str] = None) -> Dict:
        """
        CD 승인 후 Instagram 발행 예약 (준비)

        Args:
            content_file: Path to approved content (markdown)
            image_file: Path to image (optional)
            schedule_time: ISO format datetime (optional, default: next Monday 10:00)

        Returns:
            Dict with publish info
        """
        try:
            # 기본 발행 시간: 다음 월요일 오전 10시
            if not schedule_time:
                schedule_time = self._next_monday_10am()

            # 콘텐츠 읽기
            content_path = Path(content_file)
            with open(content_path, "r", encoding="utf-8") as f:
                content = f.read()

            # 메타데이터 추출
            metadata = self._read_metadata(content_path)

            # Ready 폴더로 이동
            ready_file = READY_DIR / content_path.name
            content_path.rename(ready_file)

            # 발행 정보 저장
            publish_info = {
                "id": f"woohwahae_{datetime.now().strftime('%Y%m%d_%H%M%S')}",
                "content_file": str(ready_file),
                "image_file": image_file,
                "scheduled_time": schedule_time,
                "status": "scheduled",
                "metadata": metadata,
                "locked": True  # 취소 불가
            }

            # 발행 큐에 추가
            self._add_to_publish_queue(publish_info)

            return publish_info

        except Exception as e:
            return {"error": str(e)}

    def publish_to_instagram(self, publish_info: Dict) -> Dict:
        """
        Instagram API로 실제 발행

        Meta Graph API를 사용하여 Instagram에 게시물을 발행합니다.
        단계:
        1. 이미지를 Instagram 서버에 업로드 (container 생성)
        2. Container를 publish하여 실제 게시

        Args:
            publish_info: Publish info dict

        Returns:
            Result dict
        """
        try:
            import requests
            import sys
            import os

            # Load config
            sys.path.insert(0, str(BASE_DIR))
            from libs.core_config import INSTAGRAM_CONFIG

            access_token = INSTAGRAM_CONFIG["ACCESS_TOKEN"]
            business_account_id = INSTAGRAM_CONFIG["BUSINESS_ACCOUNT_ID"]
            api_version = INSTAGRAM_CONFIG["API_VERSION"]

            if not access_token or not business_account_id:
                return {
                    "success": False,
                    "error": "Instagram credentials not configured. Set INSTAGRAM_ACCESS_TOKEN and INSTAGRAM_BUSINESS_ACCOUNT_ID in .env"
                }

            # Step 0: Read content
            content_file = Path(publish_info["content_file"])
            with open(content_file, "r", encoding="utf-8") as f:
                caption = f.read()

            # Truncate caption if needed
            max_length = INSTAGRAM_CONFIG["MAX_CAPTION_LENGTH"]
            if len(caption) > max_length:
                caption = caption[:max_length-3] + "..."

            image_url = publish_info.get("image_file")

            # Step 1: Create Media Container
            base_url = f"https://graph.facebook.com/{api_version}/{business_account_id}/media"

            container_params = {
                "access_token": access_token,
                "caption": caption
            }

            if image_url:
                container_params["image_url"] = image_url
            else:
                # Fallback: text-only not supported by Instagram, use placeholder
                return {
                    "success": False,
                    "error": "Instagram requires at least one image. Please provide image_file."
                }

            container_response = requests.post(base_url, params=container_params)
            container_data = container_response.json()

            if "error" in container_data:
                return {
                    "success": False,
                    "error": f"Instagram API Error (Container): {container_data['error']['message']}"
                }

            container_id = container_data.get("id")

            # Step 2: Publish the container
            publish_url = f"https://graph.facebook.com/{api_version}/{business_account_id}/media_publish"
            publish_params = {
                "access_token": access_token,
                "creation_id": container_id
            }

            publish_response = requests.post(publish_url, params=publish_params)
            publish_data = publish_response.json()

            if "error" in publish_data:
                return {
                    "success": False,
                    "error": f"Instagram API Error (Publish): {publish_data['error']['message']}"
                }

            post_id = publish_data.get("id")

            # Step 3: Move to published folder and update metadata
            published_file = PUBLISHED_DIR / content_file.name
            content_file.rename(published_file)

            metadata = publish_info.get("metadata", {})
            metadata["date_published"] = self.now.isoformat()
            metadata["status"] = "published"
            metadata["channel"] = "instagram"
            metadata["post_id"] = post_id
            metadata["container_id"] = container_id

            self._save_metadata(published_file, metadata)

            return {
                "success": True,
                "published_file": str(published_file),
                "published_time": self.now.isoformat(),
                "post_id": post_id,
                "post_url": f"https://www.instagram.com/p/{post_id}"
            }

        except requests.exceptions.RequestException as req_e:
            return {"success": False, "error": f"Network Error: {req_e}"}
        except Exception as e:
            return {"success": False, "error": str(e)}

    def _read_metadata(self, file_path: Path) -> Optional[Dict]:
        """마크다운 Front Matter에서 메타데이터 추출"""
        try:
            with open(file_path, "r", encoding="utf-8") as f:
                content = f.read()

            # Front Matter 파싱
            if content.startswith("---"):
                parts = content.split("---", 2)
                if len(parts) >= 3:
                    import yaml
                    metadata = yaml.safe_load(parts[1])
                    return metadata
        except:
            pass
        return None

    def _save_metadata(self, file_path: Path, metadata: Dict):
        """메타데이터를 별도 JSON 파일로 저장"""
        try:
            metadata_file = PUBLISHED_DIR / "metadata.json"

            # 기존 메타데이터 로드
            all_metadata = {}
            if metadata_file.exists():
                with open(metadata_file, "r", encoding="utf-8") as f:
                    all_metadata = json.load(f)

            # 추가
            all_metadata[file_path.stem] = metadata

            # 저장
            with open(metadata_file, "w", encoding="utf-8") as f:
                json.dump(all_metadata, f, indent=2, ensure_ascii=False)

        except Exception as e:
            print(f"Metadata save failed: {e}")

    def _next_monday_10am(self) -> str:
        """다음 월요일 오전 10시 ISO format 반환"""
        # 현재 요일 (0=월요일, 6=일요일)
        current_weekday = self.now.weekday()

        # 다음 월요일까지 일수
        days_until_monday = (7 - current_weekday) % 7
        if days_until_monday == 0:
            days_until_monday = 7  # 오늘이 월요일이면 다음주 월요일

        next_monday = self.now + timedelta(days=days_until_monday)
        next_monday = next_monday.replace(hour=10, minute=0, second=0, microsecond=0)

        return next_monday.isoformat()

    def _add_to_publish_queue(self, publish_info: Dict):
        """발행 큐에 추가 (JSON 파일)"""
        try:
            queue_file = READY_DIR / "publish_queue.json"

            queue = []
            if queue_file.exists():
                with open(queue_file, "r", encoding="utf-8") as f:
                    queue = json.load(f)

            queue.append(publish_info)

            with open(queue_file, "w", encoding="utf-8") as f:
                json.dump(queue, f, indent=2, ensure_ascii=False)

        except Exception as e:
            print(f"Queue add failed: {e}")

    def _log(self, message: str):
        """로그 기록"""
        log_file = BASE_DIR / "logs" / "auto_publisher.log"
        log_file.parent.mkdir(parents=True, exist_ok=True)

        with open(log_file, "a", encoding="utf-8") as f:
            f.write(f"{message}\n")


def main():
    """메인 실행 함수 (technical_daemon에서 호출)"""
    publisher = AutoPublisher()

    print(f"[{datetime.now()}] Auto Publisher: 72시간 규칙 체크...")

    # 1. 72시간 규칙 체크
    violations = publisher.check_72h_rule()

    if not violations:
        print("✅ 72시간 규칙 위반 없음")
        return

    # 2. 위반 처리
    for v in violations:
        if v["status"] == "violation":
            # 76시간 초과 → 자동 폐기
            print(f"🚨 자동 폐기: {v['file']} ({v['elapsed_hours']}h)")
            publisher.auto_discard(v["path"])
        else:
            # 72-76시간 → CD 알림
            print(f"⚠️ CD 결정 필요: {v['file']} ({v['elapsed_hours']}h)")

    # 3. CD에게 알림 생성
    notification = publisher.notify_cd(violations)
    if notification:
        print("\n" + notification)

        # TODO: 텔레그램 전송
        # from libs.notifier import Notifier
        # Notifier().broadcast(notification)


if __name__ == "__main__":
    main()
