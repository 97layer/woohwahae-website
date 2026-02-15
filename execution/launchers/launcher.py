#!/usr/bin/env python3
"""
97LAYER OS Automation Launcher (Ultra-Stable Version)
통합 가동 제어 센터 - 모든 서브 데몬의 상시 가동 보장
"""

import subprocess
import time
import sys
import os
import signal
from pathlib import Path
from datetime import datetime

PROJECT_ROOT = Path(__file__).resolve().parent
LOG_DIR = PROJECT_ROOT / "logs"
if not LOG_DIR.exists():
    try:
        LOG_DIR.mkdir(parents=True, exist_ok=True)
    except:
        LOG_DIR = PROJECT_ROOT / "logs_fallback"
        LOG_DIR.mkdir(parents=True, exist_ok=True)

# 가상환경 파이썬 경로 (상대 경로 및 시스템 경로 우선 순위 설정)
def get_python_exe():
    venv_python = PROJECT_ROOT / ".venv" / "bin" / "python3"
    if venv_python.exists():
        return str(venv_python)
    return sys.executable # 대체: 현재 실행 중인 파이썬 사용

PYTHON_EXE = get_python_exe()

# 관리 대상 서비스들
SERVICES = [
    {
        "name": "97LAYER Integrated Engine",
        "command": [PYTHON_EXE, str(PROJECT_ROOT / "execution" / "five_agent_hub_integrated.py")],
        "log": LOG_DIR / "five_agent_hub.log"
    },
    {
        "name": "Dashboard Server",
        "command": [PYTHON_EXE, str(PROJECT_ROOT / "execution" / "dashboard_server.py")],
        "log": LOG_DIR / "dashboard.log"
    },
    {
        "name": "Technical Daemon",
        "command": [PYTHON_EXE, str(PROJECT_ROOT / "execution" / "technical_daemon.py")],
        "log": LOG_DIR / "technical.log"
    }
]

def cleanup_existing():
    """모든 관련 프로세스를 사전에 확실히 종료"""
    patterns = [
        "five_agent_multimodal.py", 
        "five_agent_async.py", 
        "dashboard_server.py", 
        "technical_daemon.py",
        "telegram_daemon.py",
        "five_agent_hub_integrated.py",
        "WORKING_BOT.py"
    ]
    print(f"[{datetime.now()}] Cleaning up existing processes...")
    for pattern in patterns:
        # pkill -f 를 사용하여 더 정확하게 프로세스 식별 및 강제 종료
        try:
            subprocess.run(["pkill", "-9", "-f", pattern], stderr=subprocess.DEVNULL)
        except:
            pass
    time.sleep(2)

def launch_services():
    processes = []
    
    print(f"[{datetime.now()}] 97LAYER OS - Launching Integrated Automation Universe...")
    cleanup_existing()

    for svc in SERVICES:
        print(f"Starting {svc['name']}...")
        # 로그 파일 핸들 오픈 (권한 오류 대비 예외 처리)
        log_file = None
        try:
            log_file = open(svc["log"], "a", encoding="utf-8")
        except Exception as e:
            print(f"  [WARNING] Could not open log file {svc['log']}: {e}")
            
        proc = subprocess.Popen(
            svc["command"],
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            cwd=str(PROJECT_ROOT),
            text=True,
            bufsize=1
        )
        processes.append({"proc": proc, "svc": svc, "log_file": log_file})

    print(f"\n[{datetime.now()}] All systems operational. Monitoring health...")
    
    try:
        while True:
            for p_info in processes:
                poll = p_info["proc"].poll()
                if poll is not None:
                    print(f"⚠️ [ALERT] {p_info['svc']['name']} terminated (code: {poll}).")
                    time.sleep(5)
                    print(f"Restarting {p_info['svc']['name']}...")
                    p_info["proc"] = subprocess.Popen(
                        p_info["svc"]["command"],
                        stdout=subprocess.PIPE,
                        stderr=subprocess.STDOUT,
                        cwd=str(PROJECT_ROOT),
                        text=True,
                        bufsize=1
                    )
                
                # 비차단 방식으로 로그 비우기 (메모리 누수 방지용, 실제 로그 파일링은 나중에 정교화)
                # 여기서는 간단히 한 줄씩 읽거나 건너뜀
                try:
                    import select
                    if select.select([p_info["proc"].stdout], [], [], 0.0)[0]:
                        line = p_info["proc"].stdout.readline()
                        if line:
                            clean_line = line.strip()
                            timestamp = datetime.now().strftime("%H:%M:%S")
                            print(f"[{timestamp}][{p_info['svc']['name']}] {clean_line}")
                            
                            # 파일에도 기록 (핸들이 유효한 경우만)
                            if p_info.get("log_file"):
                                p_info["log_file"].write(line)
                                p_info["log_file"].flush()
                except:
                    pass
            
            time.sleep(5)
    except KeyboardInterrupt:
        print("\n[{datetime.now()}] Shutting down all services...")
        for p_info in processes:
            p_info["proc"].terminate()
            p_info["log_file"].close()
        print("Done.")

if __name__ == "__main__":
    launch_services()
