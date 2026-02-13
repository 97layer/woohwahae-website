#!/usr/bin/env python3
"""
97layerOS - System Recovery & Health Check
Comprehensive system diagnosis and automated recovery
"""

import subprocess
import json
import os
import sys
from datetime import datetime
from pathlib import Path

PROJECT_ROOT = Path("/Users/97layer/97layerOS")
VENV_PYTHON = PROJECT_ROOT / ".venv/bin/python3"
TASK_STATUS = PROJECT_ROOT / "task_status.json"

DAEMONS = {
    'technical_daemon.py': PROJECT_ROOT / 'execution/technical_daemon.py',
    'telegram_daemon.py': PROJECT_ROOT / 'execution/telegram_daemon.py',
    'snapshot_daemon.py': PROJECT_ROOT / 'execution/snapshot_daemon.py'
}

def log(message, level="INFO"):
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    prefix = {
        "INFO": "ℹ️",
        "SUCCESS": "✓",
        "ERROR": "✗",
        "WARNING": "⚠️"
    }.get(level, "•")
    print(f"[{timestamp}] {prefix} {message}")

def check_daemon_status(daemon_name):
    """Check if a daemon is running"""
    try:
        result = subprocess.run(
            ['pgrep', '-f', daemon_name],
            capture_output=True,
            text=True
        )
        return bool(result.stdout.strip())
    except Exception:
        return False

def start_daemon(daemon_path):
    """Start a daemon process"""
    if not daemon_path.exists():
        log(f"Daemon not found: {daemon_path}", "ERROR")
        return False

    try:
        subprocess.Popen(
            [str(VENV_PYTHON), str(daemon_path)],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            start_new_session=True
        )
        return True
    except Exception as e:
        log(f"Failed to start {daemon_path.name}: {e}", "ERROR")
        return False

def update_task_status(updates):
    """Update task_status.json with new information"""
    if not TASK_STATUS.exists():
        log(f"Task status file not found: {TASK_STATUS}", "WARNING")
        return

    try:
        with open(TASK_STATUS, 'r') as f:
            data = json.load(f)

        data.update(updates)
        data['last_active'] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")

        with open(TASK_STATUS, 'w') as f:
            json.dump(data, f, indent=4, ensure_ascii=False)

        log("Task status updated", "SUCCESS")
    except Exception as e:
        log(f"Failed to update task status: {e}", "ERROR")

def check_environment():
    """Check critical environment components"""
    log("\nChecking environment...")

    checks = {
        "Python venv": VENV_PYTHON.exists(),
        "Project root": PROJECT_ROOT.exists(),
        "Task status": TASK_STATUS.exists(),
        "credentials.json": (PROJECT_ROOT / "credentials.json").exists(),
        "token.json": (PROJECT_ROOT / "token.json").exists()
    }

    all_ok = True
    for name, status in checks.items():
        if status:
            log(f"  {name}: OK", "SUCCESS")
        else:
            log(f"  {name}: MISSING", "ERROR")
            all_ok = False

    return all_ok

def main():
    log("="*70)
    log("97layerOS - SYSTEM RECOVERY & HEALTH CHECK")
    log("="*70)

    # Environment check
    if not check_environment():
        log("\nEnvironment check failed. Please verify installation.", "ERROR")
        return 1

    # Daemon status check
    log("\nChecking daemon status...")
    daemon_status = {}

    for daemon_name, daemon_path in DAEMONS.items():
        is_running = check_daemon_status(daemon_name)
        daemon_status[daemon_name] = is_running

        status_text = "RUNNING" if is_running else "DOWN"
        level = "SUCCESS" if is_running else "ERROR"
        log(f"  {daemon_name}: {status_text}", level)

    # Recovery actions
    daemons_down = [name for name, status in daemon_status.items() if not status]

    if daemons_down:
        log(f"\n{len(daemons_down)} daemon(s) need recovery", "WARNING")
        log("Attempting to start daemons...")

        started = 0
        for daemon_name in daemons_down:
            daemon_path = DAEMONS[daemon_name]
            log(f"  Starting {daemon_name}...")

            if start_daemon(daemon_path):
                log(f"    Started successfully", "SUCCESS")
                started += 1
            else:
                log(f"    Failed to start", "ERROR")

        log(f"\nRecovery summary: {started}/{len(daemons_down)} daemon(s) started")
    else:
        log("\nAll daemons are running", "SUCCESS")

    # Update system status
    update_task_status({
        "pending_issue": "None" if not daemons_down else f"{len(daemons_down)} daemon(s) down",
        "system_entropy": "Low" if len(daemons_down) <= 1 else "Medium"
    })

    log("="*70)
    log("Health check complete")
    log("="*70)

    return 0 if not daemons_down else 1

if __name__ == "__main__":
    sys.exit(main())
