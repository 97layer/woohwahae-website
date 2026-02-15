#!/usr/bin/env python3
"""
자동 업데이트 프로토콜 (Auto Update Protocol)
Google Drive Core 폴더의 새 스크립트를 자동으로 시스템에 반영

Features:
- Google Drive 폴더 모니터링
- 신규/수정 파일 감지
- 기존 파일 Archive/ 백업
- 즉시 시스템 반영 (Apply)
- 변경 이력 로깅
"""

import os
import sys
import json
import shutil
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Tuple, Optional
import logging
import hashlib

# Project Root
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

# Logging Setup
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - [AutoUpdate] - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Google Drive 경로
DRIVE_ROOT = Path.home() / "내 드라이브" / "97layerOS"
DRIVE_CORE = DRIVE_ROOT / "Core"
DRIVE_ARCHIVE = DRIVE_ROOT / "Archive"

# 로컬 경로
LOCAL_ARCHIVE = PROJECT_ROOT / ".tmp" / "archive"
UPDATE_LOG = PROJECT_ROOT / "knowledge" / "system" / "update_log.json"


class AutoUpdater:
    """자동 업데이트 관리자"""

    def __init__(self, drive_path: Path = DRIVE_CORE, local_root: Path = PROJECT_ROOT):
        self.drive_path = drive_path
        self.local_root = local_root
        self.archive_path = LOCAL_ARCHIVE
        self.update_log_path = UPDATE_LOG

        # 초기화
        self.archive_path.mkdir(parents=True, exist_ok=True)
        self.update_log_path.parent.mkdir(parents=True, exist_ok=True)

        # 업데이트 이력
        self.update_history: List[Dict] = self._load_update_history()

    def _load_update_history(self) -> List[Dict]:
        """업데이트 이력 로드"""
        if self.update_log_path.exists():
            try:
                with open(self.update_log_path, 'r', encoding='utf-8') as f:
                    return json.load(f)
            except Exception as e:
                logger.error(f"Failed to load update history: {e}")
                return []
        return []

    def _save_update_history(self):
        """업데이트 이력 저장"""
        try:
            with open(self.update_log_path, 'w', encoding='utf-8') as f:
                json.dump(self.update_history, f, indent=2, ensure_ascii=False)
        except Exception as e:
            logger.error(f"Failed to save update history: {e}")

    def _compute_file_hash(self, file_path: Path) -> str:
        """파일 해시 계산 (변경 감지용)"""
        try:
            with open(file_path, 'rb') as f:
                return hashlib.sha256(f.read()).hexdigest()
        except Exception as e:
            logger.error(f"Failed to compute hash for {file_path}: {e}")
            return ""

    def _get_relative_path(self, file_path: Path, base_path: Path) -> Path:
        """상대 경로 계산"""
        try:
            return file_path.relative_to(base_path)
        except ValueError:
            return file_path

    def scan_drive_for_updates(self) -> List[Dict]:
        """Google Drive에서 업데이트 대상 파일 스캔"""
        if not self.drive_path.exists():
            logger.error(f"Google Drive path not found: {self.drive_path}")
            return []

        logger.info(f"Scanning for updates in {self.drive_path}...")

        updates = []

        # Python 파일만 스캔 (.py)
        for drive_file in self.drive_path.rglob("*.py"):
            relative_path = self._get_relative_path(drive_file, self.drive_path)
            local_file = self.local_root / relative_path

            # 해시 비교
            drive_hash = self._compute_file_hash(drive_file)

            if not local_file.exists():
                # 신규 파일
                updates.append({
                    "type": "NEW",
                    "drive_file": str(drive_file),
                    "local_file": str(local_file),
                    "relative_path": str(relative_path),
                    "drive_hash": drive_hash
                })
                logger.info(f"[NEW] {relative_path}")

            else:
                # 기존 파일 - 변경 확인
                local_hash = self._compute_file_hash(local_file)

                if drive_hash != local_hash:
                    updates.append({
                        "type": "MODIFIED",
                        "drive_file": str(drive_file),
                        "local_file": str(local_file),
                        "relative_path": str(relative_path),
                        "drive_hash": drive_hash,
                        "local_hash": local_hash
                    })
                    logger.info(f"[MODIFIED] {relative_path}")

        logger.info(f"Found {len(updates)} updates")
        return updates

    def backup_to_archive(self, local_file: Path) -> Optional[Path]:
        """기존 파일을 Archive로 백업"""
        if not local_file.exists():
            return None

        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        relative_path = self._get_relative_path(local_file, self.local_root)

        # Archive 경로 생성
        archive_file = self.archive_path / f"{relative_path.stem}_{timestamp}{relative_path.suffix}"
        archive_file.parent.mkdir(parents=True, exist_ok=True)

        try:
            shutil.copy2(local_file, archive_file)
            logger.info(f"✓ Backed up: {relative_path} → Archive/")
            return archive_file
        except Exception as e:
            logger.error(f"Failed to backup {local_file}: {e}")
            return None

    def apply_update(self, update: Dict) -> bool:
        """업데이트 적용"""
        drive_file = Path(update["drive_file"])
        local_file = Path(update["local_file"])
        update_type = update["type"]

        try:
            # 1. 기존 파일 백업 (MODIFIED인 경우)
            if update_type == "MODIFIED":
                archive_path = self.backup_to_archive(local_file)
                update["archive_path"] = str(archive_path) if archive_path else None

            # 2. 로컬 디렉토리 생성
            local_file.parent.mkdir(parents=True, exist_ok=True)

            # 3. Drive → Local 복사
            shutil.copy2(drive_file, local_file)
            logger.info(f"✓ Applied: {update['relative_path']} [{update_type}]")

            # 4. 업데이트 이력 기록
            update["timestamp"] = datetime.now().isoformat()
            update["status"] = "SUCCESS"
            self.update_history.append(update)

            return True

        except Exception as e:
            logger.error(f"Failed to apply update for {local_file}: {e}")
            update["timestamp"] = datetime.now().isoformat()
            update["status"] = "FAILED"
            update["error"] = str(e)
            self.update_history.append(update)
            return False

    def run_update_cycle(self) -> Dict:
        """업데이트 사이클 실행"""
        logger.info("=" * 60)
        logger.info("Auto Update Cycle Started")
        logger.info("=" * 60)

        # 1. Drive 스캔
        updates = self.scan_drive_for_updates()

        if not updates:
            logger.info("No updates found")
            return {
                "status": "NO_UPDATES",
                "updates_count": 0
            }

        # 2. 업데이트 적용
        success_count = 0
        failed_count = 0

        for update in updates:
            if self.apply_update(update):
                success_count += 1
            else:
                failed_count += 1

        # 3. 이력 저장
        self._save_update_history()

        logger.info("=" * 60)
        logger.info(f"✅ Update Complete: {success_count} success, {failed_count} failed")
        logger.info("=" * 60)

        return {
            "status": "COMPLETED",
            "updates_count": len(updates),
            "success_count": success_count,
            "failed_count": failed_count,
            "timestamp": datetime.now().isoformat()
        }

    def get_recent_updates(self, limit: int = 10) -> List[Dict]:
        """최근 업데이트 이력 조회"""
        return self.update_history[-limit:]


def main():
    """CLI Entry Point"""
    import argparse

    parser = argparse.ArgumentParser(description="97layerOS Auto Updater")
    parser.add_argument(
        "--drive-path",
        type=str,
        default=str(DRIVE_CORE),
        help="Google Drive Core folder path"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Scan only, do not apply updates"
    )
    parser.add_argument(
        "--history",
        action="store_true",
        help="Show recent update history"
    )

    args = parser.parse_args()

    updater = AutoUpdater(drive_path=Path(args.drive_path))

    if args.history:
        # 이력 출력
        history = updater.get_recent_updates(limit=20)
        print(json.dumps(history, indent=2, ensure_ascii=False))
        return

    if args.dry_run:
        # Dry-run: 스캔만
        updates = updater.scan_drive_for_updates()
        print(f"\nFound {len(updates)} updates (dry-run mode, not applied)")
        for update in updates:
            print(f"  [{update['type']}] {update['relative_path']}")
        return

    # 업데이트 실행
    result = updater.run_update_cycle()
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
