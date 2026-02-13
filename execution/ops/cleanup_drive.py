#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Filename: execution/ops/cleanup_drive.py
Author: 97LAYER Mercenary
Date: 2026-02-12
Description: Google Drive에서 .driveignore 패턴에 해당하는 파일/폴더 삭제
"""

import sys
import logging
from pathlib import Path
from typing import List, Set

# Path Setup
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
if str(PROJECT_ROOT) not in sys.path:
    sys.path.append(str(PROJECT_ROOT))

from libs.google_workspace import GoogleWorkspace

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# 삭제 대상 (driveignore 주요 항목)
CLEANUP_TARGETS = [
    "venv",
    "venv_old_broken",
    "__pycache__",
    ".git",
    ".tmp",
    ".tmp.driveupload",
    ".tmp.drivedownload",
    ".antigravity",
    ".claude",
    ".vscode",
    ".idea",
]

def find_97layeros_folder(gw: GoogleWorkspace) -> str:
    """97layerOS 폴더 찾기"""
    query = "name = '97layerOS' and mimeType = 'application/vnd.google-apps.folder' and trashed = false"
    folders = gw.search_files(query)

    if not folders:
        logger.error("97layerOS 폴더를 찾을 수 없습니다.")
        sys.exit(1)

    folder_id = folders[0]['id']
    logger.info("97layerOS 폴더 발견: %s", folder_id)
    return folder_id

def find_files_to_delete(gw: GoogleWorkspace, parent_id: str) -> List[dict]:
    """삭제 대상 파일/폴더 검색"""
    to_delete = []

    for target in CLEANUP_TARGETS:
        query = f"name = '{target}' and '{parent_id}' in parents and trashed = false"
        results = gw.search_files(query)

        if results:
            for item in results:
                to_delete.append(item)
                logger.info("삭제 대상 발견: %s (%s)", item['name'], item['id'])

    return to_delete

def delete_files(gw: GoogleWorkspace, items: List[dict], dry_run: bool = True) -> None:
    """파일/폴더 삭제"""
    if not items:
        logger.info("삭제할 항목이 없습니다.")
        return

    logger.info("총 %d개 항목 발견", len(items))

    if dry_run:
        logger.warning("DRY RUN 모드: 실제 삭제하지 않습니다.")
        logger.warning("실제 삭제하려면 스크립트에 --execute 플래그를 추가하세요.")
        for item in items:
            logger.info("  - %s (ID: %s)", item['name'], item['id'])
        return

    # 실제 삭제
    service = gw.get_service('drive', 'v3')
    for item in items:
        try:
            service.files().delete(fileId=item['id']).execute()
            logger.info("삭제 완료: %s", item['name'])
        except Exception as e:
            logger.error("삭제 실패 (%s): %s", item['name'], e)

def main() -> None:
    logger.info("Google Drive 정리 시작...")

    # Dry run 여부 확인
    dry_run = "--execute" not in sys.argv

    # Google Workspace 연결
    gw = GoogleWorkspace()

    # 97layerOS 폴더 찾기
    folder_id = find_97layeros_folder(gw)

    # 삭제 대상 검색
    items = find_files_to_delete(gw, folder_id)

    # 삭제 실행
    delete_files(gw, items, dry_run=dry_run)

    logger.info("작업 완료.")

if __name__ == "__main__":
    main()
