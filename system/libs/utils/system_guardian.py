import os
import subprocess
import logging
from pathlib import Path
from datetime import datetime

logger = logging.getLogger(__name__)

class SystemGuardian:
    def __init__(self, project_root: str):
        self.project_root = Path(project_root)
        self.daemons = ["technical_daemon.py", "telegram_daemon.py", "snapshot_daemon.py"]
        self.log_dir = Path("/tmp") # Default log location for nohup in this project

    def check_daemon_health(self) -> dict:
        """Checks if configured daemons are running."""
        health_status = {}
        for daemon in self.daemons:
            # Use pgrep to find process by full name
            try:
                # Use ps aux for more reliable process detection on different Linux distros
                cmd = f"ps aux | grep {daemon} | grep -v grep"
                result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
                is_running = result.returncode == 0
                health_status[daemon] = {
                    "status": "Running" if is_running else "Down",
                    "pids": [line.split()[1] for line in result.stdout.strip().split("\n") if line] if is_running else []
                }
            except Exception as e:
                health_status[daemon] = {"status": "Error", "message": str(e)}
        return health_status

    def analyze_logs(self, limit_lines: int = 50) -> dict:
        """Scans recent logs for [ERROR] or Traceback patterns."""
        errors = {}
        for daemon in self.daemons:
            log_file = self.log_dir / f"{daemon.replace('.py', '')}.log"
            if not log_file.exists():
                # Try common names like telegram_daemon.log
                log_file = self.log_dir / f"{daemon.replace('.py', '')}.log"
            
            if log_file.exists():
                try:
                    # Read last N lines
                    with open(log_file, "r", encoding="utf-8", errors="ignore") as f:
                        lines = f.readlines()
                        recent_lines = lines[-limit_lines:]
                        
                        found_errors = []
                        for i, line in enumerate(recent_lines):
                            if "ERROR" in line.upper() or "TRACEBACK" in line.upper() or "EXCEPTION" in line.upper():
                                # Capture some context
                                context = recent_lines[max(0, i-2):min(len(recent_lines), i+3)]
                                found_errors.append("".join(context))
                        
                        if found_errors:
                            errors[daemon] = found_errors[-3:] # Last 3 unique errors
                except Exception as e:
                    errors[daemon] = [f"Log reading error: {e}"]
        return errors

    def get_system_report(self) -> str:
        """Generates a combined health and error report."""
        health = self.check_daemon_health()
        errors = self.analyze_logs()
        
        report = "◈ [System Guardian Report]\n\n"
        
        report += "[Daemon Status]\n"
        for d, status in health.items():
            icon = "✅" if status["status"] == "Running" else "❌"
            report += f"{icon} {d}: {status['status']}\n"
            
        if errors:
            report += "\n[Recent Errors Detected]\n"
            for d, errs in errors.items():
                report += f"📍 {d}:\n"
                for e in errs:
                    report += f"  - {e[:200]}...\n"
        else:
            report += "\n✅ No critical errors detected in logs."
            
        return report
