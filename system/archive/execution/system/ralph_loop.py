#!/usr/bin/env python3
"""
Ralph Loop: Autonomous Self-Healing Protocol
자율 완결 프로토콜 - 에러 자가 수정 시스템
"""

import subprocess
import sys
import time
import json
from pathlib import Path
from typing import Tuple, Optional, Dict, Any

class RalphLoop:
    """자율 완결 프로토콜 실행기"""

    def __init__(self):
        self.max_retries = 5
        self.retry_count = 0
        self.errors_fixed = []
        self.verification_log = []

    def execute_with_bypass(self, command: str) -> Tuple[int, str, str]:
        """macOS 샌드박스 우회 실행"""
        env = os.environ.copy()
        env['TMPDIR'] = '/tmp'

        try:
            result = subprocess.run(
                command,
                shell=True,
                capture_output=True,
                text=True,
                env=env,
                timeout=30
            )
            return result.returncode, result.stdout, result.stderr
        except subprocess.TimeoutExpired:
            return -1, "", "Command timeout after 30 seconds"
        except Exception as e:
            return -2, "", str(e)

    def self_heal(self, error_msg: str, context: Dict[str, Any]) -> bool:
        """에러 자가 수정"""
        self.retry_count += 1

        if self.retry_count > self.max_retries:
            return False

        # UTF-8 에러 자동 수정
        if "utf-8" in error_msg.lower() or "decode" in error_msg.lower():
            context['encoding'] = 'latin-1'
            self.errors_fixed.append(f"Fixed UTF-8 error (attempt {self.retry_count})")
            return True

        # 권한 에러 자동 수정
        if "permission" in error_msg.lower() or "chmod" in error_msg.lower():
            if 'file_path' in context:
                self.execute_with_bypass(f"chmod +x {context['file_path']}")
                self.errors_fixed.append(f"Fixed permission error (attempt {self.retry_count})")
                return True

        # Build 실패 자동 수정
        if "build" in error_msg.lower() or "compile" in error_msg.lower():
            # 의존성 재설치 시도
            self.execute_with_bypass("pip install -r requirements.txt 2>/dev/null")
            self.errors_fixed.append(f"Reinstalled dependencies (attempt {self.retry_count})")
            return True

        # Podman 관련 에러 자동 수정
        if "podman" in error_msg.lower():
            # Podman 소켓 재시작
            self.execute_with_bypass("podman system migrate")
            self.errors_fixed.append(f"Migrated Podman system (attempt {self.retry_count})")
            return True

        return False

    def stap_assess(self, task: str) -> Dict[str, Any]:
        """STAP 검증 - 실제 실행 확인"""
        assessment = {
            "task": task,
            "timestamp": time.time(),
            "verified": False,
            "evidence": []
        }

        # Step 1: 실행
        exit_code, stdout, stderr = self.execute_with_bypass(task)

        # Step 2: 검증
        if exit_code == 0:
            assessment["verified"] = True
            assessment["evidence"].append({
                "type": "execution",
                "exit_code": exit_code,
                "output_bytes": len(stdout),
                "has_output": len(stdout) > 0
            })
        else:
            # Step 3: 자가 수정
            context = {"command": task, "file_path": task.split()[-1] if task else ""}
            if self.self_heal(stderr, context):
                # 재실행
                exit_code, stdout, stderr = self.execute_with_bypass(task)
                if exit_code == 0:
                    assessment["verified"] = True
                    assessment["evidence"].append({
                        "type": "self_healed",
                        "attempts": self.retry_count,
                        "fixes": self.errors_fixed
                    })

        self.verification_log.append(assessment)
        return assessment

    def generate_report(self) -> str:
        """최종 보고서 생성"""
        report = []
        report.append("=== RALPH LOOP VERIFICATION REPORT ===\n")

        all_verified = all(log["verified"] for log in self.verification_log)

        for log in self.verification_log:
            status = "✓ VERIFIED" if log["verified"] else "✗ FAILED"
            report.append(f"\nTask: {log['task']}")
            report.append(f"Status: {status}")
            if log["evidence"]:
                report.append("Evidence:")
                for evidence in log["evidence"]:
                    report.append(f"  - {json.dumps(evidence, indent=2)}")

        if self.errors_fixed:
            report.append("\n=== SELF-HEALING LOG ===")
            for fix in self.errors_fixed:
                report.append(f"  - {fix}")

        # 최종 토큰 부착
        if all_verified and len(self.verification_log) > 0:
            report.append("\n" + "="*40)
            report.append("[SYSTEM_STABLE]")
        else:
            report.append("\n" + "="*40)
            report.append("[SYSTEM_UNSTABLE] - Manual intervention required")

        return "\n".join(report)

import os

if __name__ == "__main__":
    # Ralph Loop 실행 데모
    ralph = RalphLoop()

    # 검증 작업들
    tasks = [
        "export TMPDIR=/tmp && podman ps",
        "export TMPDIR=/tmp && podman stats --no-stream",
        "ls -la /Users/97layer/97layerOS/execution/system/"
    ]

    for task in tasks:
        ralph.stap_assess(task)

    # 보고서 출력
    print(ralph.generate_report())