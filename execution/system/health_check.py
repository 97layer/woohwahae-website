import json
import os
import sys
from datetime import datetime, timedelta
from pathlib import Path

# Color codes
GREEN = "\033[92m"
YELLOW = "\033[93m"
RED = "\033[91m"
RESET = "\033[0m"

BASE_DIR = Path(__file__).resolve().parent.parent.parent
TASK_FILE = BASE_DIR / "task_status.json"

def check_file_exists(path: Path, description: str):
    if path.exists():
        print(f"[{GREEN}OK{RESET}] {description}: Found")
        return True
    else:
        print(f"[{RED}MISSING{RESET}] {description}: Not found at {path}")
        return False

def check_daemon_status() -> int:
    """
    시스템 건강도 체크

    Returns:
        0: 정상 (모든 체크 통과)
        1: 경고/에러 (하나 이상 실패)
    """
    print(f"\n--- 97LAYER System Health Check ---")
    print(f"Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")

    has_errors = False

    # 1. Check Technical Daemon Heartbeat
    if not TASK_FILE.exists():
        print(f"[{RED}CRITICAL{RESET}] task_status.json missing!")
        has_errors = True
    else:
        try:
            with open(TASK_FILE, 'r', encoding='utf-8') as f:
                status = json.load(f)

            last_active_str = status.get("last_active")
            if not last_active_str:
                 print(f"[{YELLOW}WARN{RESET}] No last_active timestamp found in task_status.json")
                 has_errors = True
            else:
                last_active = datetime.strptime(last_active_str, "%Y-%m-%d %H:%M:%S")
                delta = datetime.now() - last_active

                if delta < timedelta(minutes=20):
                     print(f"[{GREEN}OK{RESET}] Technical Daemon Active (Last heartbeat: {last_active_str}, {delta.seconds // 60} min ago)")
                else:
                     print(f"[{RED}WARN{RESET}] Technical Daemon Inactive? (Last heartbeat: {last_active_str}, {delta.seconds // 60} min ago)")
                     has_errors = True

        except Exception as e:
            print(f"[{RED}ERROR{RESET}] Failed to read task_status.json: {e}")
            has_errors = True

    # 2. Check Script Existence
    if not check_file_exists(BASE_DIR / "execution" / "technical_daemon.py", "Technical Daemon Script"):
        has_errors = True

    # Night Guard 환경에서는 technical_daemon 대신 nightguard_daemon 체크
    nightguard_script = BASE_DIR / "execution" / "system" / "nightguard_daemon.py"
    if nightguard_script.exists():
        check_file_exists(nightguard_script, "Night Guard Daemon")

    # 3. 핵심 파일 체크
    critical_files = [
        BASE_DIR / "libs" / "core_config.py",
        BASE_DIR / "execution" / "system" / "hybrid_sync.py",
        BASE_DIR / "CLAUDE.md"
    ]

    for file_path in critical_files:
        if not file_path.exists():
            print(f"[{RED}CRITICAL{RESET}] {file_path.name} missing!")
            has_errors = True

    # 결과 반환
    print("")
    if has_errors:
        print(f"[{RED}HEALTH CHECK FAILED{RESET}]")
        return 1
    else:
        print(f"[{GREEN}HEALTH CHECK PASSED{RESET}]")
        return 0

if __name__ == "__main__":
    exit_code = check_daemon_status()
    sys.exit(exit_code)
