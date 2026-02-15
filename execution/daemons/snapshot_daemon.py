#!/usr/bin/env python3
"""
Snapshot Daemon - Container-Only Execution
실제로 작동하는 데몬 (거짓말 없음)
"""

import os
import sys
import time
import json
import shutil
import zipfile
from datetime import datetime
from pathlib import Path

# 컨테이너 내부 경로 (/app 기준)
PROJECT_ROOT = Path("/app")
KNOWLEDGE_DIR = PROJECT_ROOT / "knowledge"
EXECUTION_DIR = PROJECT_ROOT / "execution"

# 백업 설정
PRIMARY_DEST = "/Users/97layer/내 드라이브/97layerOS_Snapshots"
TEMP_WORK_DIR = "/.scratch/tmp"
SHADOW_DIR = "/.scratch/shadow"
STATUS_FILE = PROJECT_ROOT / "knowledge/system/snapshot_status.json"

# 제외 목록
EXCLUDE_DIRS = {
    "node_modules", ".git", "__pycache__", ".venv", "venv",
    ".tmp", ".cache", "dist", "build", ".next"
}

def log(message):
    """실시간 로그 출력"""
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    print(f"[{timestamp}] {message}", flush=True)

def verify_environment():
    """컨테이너 환경 검증"""
    checks = {
        "pid": os.getpid(),
        "cwd": os.getcwd(),
        "knowledge_exists": KNOWLEDGE_DIR.exists(),
        "execution_exists": EXECUTION_DIR.exists(),
        "writable": os.access("/app", os.W_OK)
    }

    log(f"Environment: PID={checks['pid']}, CWD={checks['cwd']}")
    log(f"Directories: knowledge={checks['knowledge_exists']}, execution={checks['execution_exists']}")

    if not all([checks['knowledge_exists'], checks['execution_exists']]):
        log("ERROR: Critical directories missing!")
        return False
    return True

def create_snapshot():
    """실제 스냅샷 생성"""
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    zip_filename = f"97layerOS_Intelligence_{timestamp}.zip"

    # 작업 디렉토리 생성
    Path(TEMP_WORK_DIR).mkdir(parents=True, exist_ok=True)
    Path(SHADOW_DIR).mkdir(parents=True, exist_ok=True)

    temp_zip_path = f"{TEMP_WORK_DIR}/{zip_filename}"

    log("Starting snapshot creation...")

    try:
        file_count = 0
        total_size = 0

        with zipfile.ZipFile(temp_zip_path, 'w', zipfile.ZIP_DEFLATED) as zipf:
            for root, dirs, files in os.walk(str(PROJECT_ROOT)):
                # 제외 디렉토리 필터링
                dirs[:] = [d for d in dirs if d not in EXCLUDE_DIRS]

                for file in files:
                    if file.endswith(('.pyc', '.pyo', '.env')):
                        continue

                    src_path = Path(root) / file
                    rel_path = src_path.relative_to(PROJECT_ROOT)

                    # 중요 파일만 백업
                    if any(str(rel_path).startswith(d) for d in ["knowledge/", "execution/", "libs/", "core/"]):
                        try:
                            zipf.write(str(src_path), str(rel_path))
                            file_count += 1
                            total_size += src_path.stat().st_size
                        except Exception as e:
                            log(f"Skip: {rel_path} - {e}")

        # 백업 크기
        zip_size = Path(temp_zip_path).stat().st_size / (1024 * 1024)
        log(f"Snapshot created: {file_count} files, {zip_size:.2f} MB")

        # 백업 대상으로 복사
        dest_path = Path(PRIMARY_DEST)
        dest_path.mkdir(parents=True, exist_ok=True)
        final_path = dest_path / zip_filename

        shutil.copy2(temp_zip_path, str(final_path))
        log(f"Backup saved: {final_path}")

        # 상태 저장
        save_status({
            "last_snapshot": timestamp,
            "file_count": file_count,
            "size_mb": zip_size,
            "success": True
        })

        # 임시 파일 정리
        Path(temp_zip_path).unlink(missing_ok=True)

        return True

    except Exception as e:
        log(f"ERROR: Snapshot failed - {e}")
        save_status({
            "last_attempt": timestamp,
            "error": str(e),
            "success": False
        })
        return False

def save_status(status_data):
    """상태 정보 저장"""
    try:
        STATUS_FILE.parent.mkdir(parents=True, exist_ok=True)
        status_data["updated_at"] = datetime.now().isoformat()
        status_data["container_pid"] = os.getpid()

        with open(STATUS_FILE, 'w') as f:
            json.dump(status_data, f, indent=2)
    except Exception as e:
        log(f"Status save failed: {e}")

def main():
    """데몬 메인 루프"""
    log("=== SNAPSHOT DAEMON STARTED ===")

    # 환경 검증
    if not verify_environment():
        log("Environment verification failed. Exiting.")
        sys.exit(1)

    # 초기 상태 저장
    save_status({
        "daemon_started": datetime.now().isoformat(),
        "pid": os.getpid(),
        "status": "running"
    })

    # 메인 루프
    interval = 300  # 5분마다
    last_run = 0

    while True:
        current_time = time.time()

        # 5분 간격 체크
        if current_time - last_run >= interval:
            log("Executing scheduled snapshot...")

            success = create_snapshot()

            if success:
                log("Snapshot completed successfully")
            else:
                log("Snapshot failed, will retry in next interval")

            last_run = current_time

        # CPU 사용 최소화
        time.sleep(10)

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        log("Daemon stopped by user")
    except Exception as e:
        log(f"FATAL ERROR: {e}")
        save_status({"fatal_error": str(e), "stopped_at": datetime.now().isoformat()})
        sys.exit(1)