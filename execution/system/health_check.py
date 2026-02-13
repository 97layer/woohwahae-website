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

def check_daemon_status():
    print(f"\n--- 97LAYER System Health Check ---")
    print(f"Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    
    # 1. Check Technical Daemon Heartbeat
    if not TASK_FILE.exists():
        print(f"[{RED}CRITICAL{RESET}] task_status.json missing!")
        return
    
    try:
        with open(TASK_FILE, 'r', encoding='utf-8') as f:
            status = json.load(f)
            
        last_active_str = status.get("last_active")
        if not last_active_str:
             print(f"[{YELLOW}WARN{RESET}] No last_active timestamp found in task_status.json")
        else:
            last_active = datetime.strptime(last_active_str, "%Y-%m-%d %H:%M:%S")
            delta = datetime.now() - last_active
            
            if delta < timedelta(minutes=20):
                 print(f"[{GREEN}OK{RESET}] Technical Daemon Active (Last heartbeat: {last_active_str}, {delta.seconds // 60} min ago)")
            else:
                 print(f"[{RED}WARN{RESET}] Technical Daemon Inactive? (Last heartbeat: {last_active_str}, {delta.seconds // 60} min ago)")
                 
    except Exception as e:
        print(f"[{RED}ERROR{RESET}] Failed to read task_status.json: {e}")

    # 2. Check Script Existence
    check_file_exists(BASE_DIR / "execution" / "technical_daemon.py", "Technical Daemon Script")
    check_file_exists(BASE_DIR / "execution" / "telegram_daemon.py", "Telegram Daemon Script")
    check_file_exists(BASE_DIR / "execution" / "snapshot_daemon.py", "Snapshot Daemon Script")
    
    # 3. Check Test Scripts
    check_file_exists(BASE_DIR / "tests" / "verify_synapse.py", "Synapse Verification Test")

if __name__ == "__main__":
    check_daemon_status()
