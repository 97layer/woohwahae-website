#!/usr/bin/env python3
"""
Container Activity Monitor
컨테이너 내부 활동 실시간 모니터링
"""

import os
import time
import json
from datetime import datetime
from pathlib import Path

def check_processes():
    """프로세스 상태 확인"""
    processes = []
    proc_dir = Path("/proc")

    for pid_dir in proc_dir.glob("[0-9]*"):
        try:
            cmdline_file = pid_dir / "cmdline"
            if cmdline_file.exists():
                cmdline = cmdline_file.read_text().replace('\0', ' ').strip()
                if cmdline and "python" in cmdline:
                    processes.append({
                        "pid": pid_dir.name,
                        "command": cmdline[:80]
                    })
        except:
            pass

    return processes

def check_recent_files():
    """최근 수정된 파일 확인"""
    recent_files = []
    base_path = Path("/app")
    current_time = time.time()

    for pattern in ["knowledge/**/*.json", "knowledge/**/*.md", "execution/**/*.py"]:
        for file_path in base_path.glob(pattern):
            try:
                mtime = file_path.stat().st_mtime
                age_seconds = current_time - mtime
                if age_seconds < 600:  # 10분 이내
                    recent_files.append({
                        "path": str(file_path.relative_to(base_path)),
                        "age_seconds": int(age_seconds),
                        "size": file_path.stat().st_size
                    })
            except:
                pass

    return sorted(recent_files, key=lambda x: x["age_seconds"])

def main():
    """실시간 모니터링"""
    print("=" * 60)
    print("    97LAYEROS CONTAINER MONITOR - REAL-TIME")
    print("=" * 60)

    while True:
        os.system('clear')
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")

        print(f"\n📅 {timestamp}")
        print("=" * 60)

        # 프로세스 확인
        processes = check_processes()
        print(f"\n🔧 Running Processes: {len(processes)}")
        for proc in processes[:5]:
            print(f"  PID {proc['pid']}: {proc['command']}")

        # 최근 파일 활동
        recent_files = check_recent_files()
        if recent_files:
            print(f"\n📝 Recent Activity ({len(recent_files)} files):")
            for file_info in recent_files[:5]:
                age_str = f"{file_info['age_seconds']}s ago"
                print(f"  {file_info['path']}: {age_str} ({file_info['size']} bytes)")
        else:
            print("\n⚠️  NO RECENT FILE ACTIVITY!")

        # 상태 파일 확인
        status_files = [
            "/app/knowledge/system/snapshot_status.json",
            "/app/knowledge/system/task_board.json"
        ]

        print("\n📊 System Status:")
        for status_file in status_files:
            try:
                with open(status_file, 'r') as f:
                    data = json.load(f)
                    filename = Path(status_file).name
                    if "updated_at" in data:
                        print(f"  {filename}: {data['updated_at']}")
                    else:
                        print(f"  {filename}: {list(data.keys())[:3]}")
            except:
                pass

        print("\n" + "-" * 60)
        print("Press Ctrl+C to exit | Updates every 5 seconds")

        time.sleep(5)

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\nMonitor stopped")