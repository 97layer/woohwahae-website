#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Filename: execution/ops/sync_to_gdrive.py
Author: 97LAYER
Date: 2026-02-14
Description: 맥북 → 구글 드라이브 자동 동기화 (5분 주기)
"""

import os
import shutil
import sys
from pathlib import Path
from datetime import datetime

# Base directory
BASE_DIR = Path(__file__).resolve().parent.parent.parent
GDRIVE_BASE = Path.home() / "내 드라이브(skyto5339@gmail.com)" / "97layerOS"

# 동기화 대상
SYNC_DIRS = [
    "knowledge",
    "directives",
    "execution",
    "libs",
]

SYNC_FILES = [
    "task_status.json",
    ".env",
    "CLAUDE.md",
    "AGENTS.md",
    "GEMINI.md",
]

def ensure_gdrive_structure():
    """구글 드라이브 폴더 구조 생성"""
    try:
        GDRIVE_BASE.mkdir(parents=True, exist_ok=True)

        # Snapshots 폴더도 생성
        snapshots_dir = Path.home() / "내 드라이브(skyto5339@gmail.com)" / "97layerOS_Snapshots"
        snapshots_dir.mkdir(parents=True, exist_ok=True)

        print(f"[{datetime.now()}] ✅ 구글 드라이브 폴더 구조 생성 완료")
        return True
    except Exception as e:
        print(f"[{datetime.now()}] ❌ 폴더 생성 실패: {e}")
        return False

def sync_to_gdrive():
    """맥북 → 구글 드라이브 동기화"""
    print(f"[{datetime.now()}] 🔄 구글 드라이브 동기화 시작...")

    if not GDRIVE_BASE.exists():
        if not ensure_gdrive_structure():
            return False

    synced_count = 0
    error_count = 0

    # 1. 디렉토리 동기화
    for dir_name in SYNC_DIRS:
        src_dir = BASE_DIR / dir_name
        dst_dir = GDRIVE_BASE / dir_name

        if not src_dir.exists():
            continue

        try:
            # rsync 방식 (변경된 파일만 복사)
            if dst_dir.exists():
                shutil.rmtree(dst_dir)
            shutil.copytree(src_dir, dst_dir,
                          ignore=shutil.ignore_patterns('*.pyc', '__pycache__', '.DS_Store', '*.log'))
            synced_count += 1
            print(f"  ✅ {dir_name}/")
        except Exception as e:
            print(f"  ❌ {dir_name}/ - {e}")
            error_count += 1

    # 2. 파일 동기화
    for file_name in SYNC_FILES:
        src_file = BASE_DIR / file_name
        dst_file = GDRIVE_BASE / file_name

        if not src_file.exists():
            continue

        try:
            shutil.copy2(src_file, dst_file)
            synced_count += 1
            print(f"  ✅ {file_name}")
        except Exception as e:
            print(f"  ❌ {file_name} - {e}")
            error_count += 1

    print(f"[{datetime.now()}] 동기화 완료: {synced_count}개 성공, {error_count}개 실패")
    return error_count == 0

def main():
    """메인 실행"""
    try:
        success = sync_to_gdrive()
        sys.exit(0 if success else 1)
    except Exception as e:
        print(f"[{datetime.now()}] 동기화 오류: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
