#!/usr/bin/env python3
"""
97layerOS - Duplicate Daemon Process Killer
Resolves HTTP 409 Conflict by terminating duplicate daemon instances
"""

import subprocess
import sys
import os
from datetime import datetime

DAEMON_NAMES = ['technical_daemon.py', 'telegram_daemon.py', 'snapshot_daemon.py']

def log(message):
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    print(f"[{timestamp}] {message}")

def find_daemon_processes(daemon_name):
    """Find all PIDs for a given daemon name"""
    try:
        result = subprocess.run(
            ['ps', 'aux'],
            capture_output=True,
            text=True,
            check=True
        )

        processes = []
        for line in result.stdout.split('\n'):
            if daemon_name in line and 'grep' not in line:
                parts = line.split()
                if len(parts) > 1:
                    pid = parts[1]
                    processes.append({
                        'pid': pid,
                        'cmd': ' '.join(parts[10:])
                    })

        return processes
    except subprocess.CalledProcessError as e:
        log(f"Error finding processes: {e}")
        return []

def kill_process(pid):
    """Kill a process by PID"""
    try:
        subprocess.run(['kill', '-9', pid], check=True)
        return True
    except subprocess.CalledProcessError:
        return False

def main():
    log("Starting duplicate daemon cleanup...")

    total_killed = 0

    for daemon_name in DAEMON_NAMES:
        log(f"\nChecking {daemon_name}...")
        processes = find_daemon_processes(daemon_name)

        if not processes:
            log(f"  No running instances found")
            continue

        log(f"  Found {len(processes)} instance(s)")

        # Kill all instances (will be restarted fresh)
        for proc in processes:
            log(f"  Killing PID {proc['pid']}: {proc['cmd'][:80]}...")
            if kill_process(proc['pid']):
                log(f"    ✓ Killed successfully")
                total_killed += 1
            else:
                log(f"    ✗ Failed to kill")

    log(f"\n{'='*60}")
    log(f"Cleanup complete: {total_killed} process(es) terminated")
    log(f"{'='*60}")

    return 0

if __name__ == "__main__":
    sys.exit(main())
