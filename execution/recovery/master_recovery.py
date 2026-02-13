#!/usr/bin/env python3
"""
97layerOS - Master Recovery Workflow
Orchestrates complete system recovery in proper sequence
"""

import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path

SCRIPT_DIR = Path(__file__).parent
SCRIPTS = [
    ('kill_duplicate_daemons.py', 'Terminating duplicate processes'),
    ('refresh_oauth_token.py', 'Refreshing OAuth tokens'),
    ('system_recovery.py', 'System health check and daemon restart')
]

def log(message, level="INFO"):
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    symbols = {
        "INFO": "ℹ️",
        "SUCCESS": "✓",
        "ERROR": "✗",
        "STEP": "▶"
    }
    print(f"[{timestamp}] {symbols.get(level, '•')} {message}")

def run_script(script_name, description):
    """Run a recovery script"""
    script_path = SCRIPT_DIR / script_name

    if not script_path.exists():
        log(f"Script not found: {script_name}", "ERROR")
        return False

    log(f"{description}...", "STEP")

    try:
        result = subprocess.run(
            [sys.executable, str(script_path)],
            capture_output=True,
            text=True,
            timeout=60
        )

        # Print script output
        if result.stdout:
            for line in result.stdout.split('\n'):
                if line.strip():
                    print(f"  {line}")

        if result.returncode == 0:
            log(f"Completed: {script_name}", "SUCCESS")
            return True
        else:
            log(f"Failed: {script_name} (exit code: {result.returncode})", "ERROR")
            if result.stderr:
                print(f"Error output:\n{result.stderr}")
            return False

    except subprocess.TimeoutExpired:
        log(f"Timeout: {script_name}", "ERROR")
        return False
    except Exception as e:
        log(f"Exception running {script_name}: {e}", "ERROR")
        return False

def main():
    log("="*70)
    log("97layerOS - MASTER RECOVERY WORKFLOW")
    log("="*70)
    log("This will perform complete system recovery in 3 steps")
    log("")

    start_time = time.time()
    success_count = 0

    for i, (script, description) in enumerate(SCRIPTS, 1):
        log(f"\nStep {i}/{len(SCRIPTS)}: {description}")
        log("-" * 70)

        if run_script(script, description):
            success_count += 1
            time.sleep(2)  # Brief pause between steps
        else:
            log(f"\nRecovery workflow stopped at step {i}", "ERROR")
            break

    elapsed = time.time() - start_time

    log("\n" + "="*70)
    log(f"Recovery workflow completed: {success_count}/{len(SCRIPTS)} steps successful")
    log(f"Time elapsed: {elapsed:.1f} seconds")
    log("="*70)

    if success_count == len(SCRIPTS):
        log("\n✓ SYSTEM RECOVERY COMPLETE", "SUCCESS")
        log("All daemons should now be running properly.")
        log("Run 'ps aux | grep daemon' to verify.")
        return 0
    else:
        log("\n✗ RECOVERY INCOMPLETE", "ERROR")
        log("Please review errors above and retry.")
        return 1

if __name__ == "__main__":
    sys.exit(main())
