import time
import os
import sys
import subprocess
from datetime import datetime

# 프로젝트 루트를 sys.path에 추가
root_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if root_dir not in sys.path:
    sys.path.insert(0, root_dir)

try:
    from execution.create_snapshot import create_snapshot
except ImportError:
    sys.path.append("/Users/97layer/97layerOS")
    from execution.create_snapshot import create_snapshot

def continuous_sanitization():
    """Sentinel: 스냅샷 전 시스템을 정화하여 미니멀리즘 상태를 유지"""
    print(f"[{datetime.now()}] Sentinel: Continuous Sanitization sequence initiated...", flush=True)
    
    # 1. 제거 대상 정의
    targets = [
        ".venv_runtime", "venv_new", "venv_old_broken", 
        ".local_node", ".mcp-source", ".tmp"
    ]
    
    # 2. 강력한 소거 집행
    purged_count = 0
    for target in targets:
        path = os.path.join("/Users/97layer/97layerOS", target)
        if os.path.exists(path):
            try:
                if os.path.isdir(path):
                    subprocess.run(["rm", "-rf", path], check=True)
                else:
                    os.remove(path)
                purged_count += 1
                print(f"[SENTINEL] Purged: {target}", flush=True)
            except Exception as e:
                print(f"[SENTINEL] Failed to purge {target}: {e}", flush=True)
                
    # 3. 누수된 node_modules 검색 및 소거
    try:
        # find 명령어로 node_modules 폴더 탐색 후 삭제
        subprocess.run(
            "find /Users/97layer/97layerOS -name 'node_modules' -type d -exec rm -rf {} +",
            shell=True, capture_output=True
        )
    except Exception as e:
        print(f"[SENTINEL] node_modules purge failed: {e}", flush=True)

    print(f"[{datetime.now()}] Sentinel: Sanitization complete. (Purged {purged_count} core targets)", flush=True)

def run_daemon():
    print(f"[{datetime.now()}] 97LAYER Snapshot Sentinel Daemon Started.", flush=True)
    
    while True:
        try:
            # Step 1: Sentinel Sanitization
            continuous_sanitization()
            
            # Step 2: Intelligent Snapshot
            print(f"[{datetime.now()}] Starting scheduled snapshot...", flush=True)
            success = create_snapshot()
            
            if success:
                print(f"[{datetime.now()}] Snapshot successful. Intelligence preserved.", flush=True)
            else:
                print(f"[{datetime.now()}] Snapshot failed. Will retry.", flush=True)
                
        except Exception as e:
            print(f"[{datetime.now()}] Daemon Error: {e}", flush=True)
            
        # 1시간 대기
        time.sleep(3600)

if __name__ == "__main__":
    run_daemon()
