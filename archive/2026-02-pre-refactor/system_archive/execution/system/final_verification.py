#!/usr/bin/env python3
"""
최종 시스템 안정성 검증
물리적 증거 기반 완전 검증
"""

import subprocess
import json
import time
import os
from datetime import datetime

class FinalSystemVerification:
    def __init__(self):
        self.evidence = []
        self.critical_failures = []

    def execute_cmd(self, cmd):
        """명령 실행 및 증거 수집"""
        env = os.environ.copy()
        env['TMPDIR'] = '/tmp'

        result = subprocess.run(
            cmd,
            shell=True,
            capture_output=True,
            text=True,
            env=env
        )

        return {
            "command": cmd,
            "exit_code": result.returncode,
            "stdout_bytes": len(result.stdout),
            "stderr_bytes": len(result.stderr),
            "timestamp": datetime.now().isoformat(),
            "output": result.stdout[:500] if result.stdout else None
        }

    def verify_podman_infrastructure(self):
        """Podman 인프라 완전 검증"""
        print("🔍 Podman Infrastructure Verification")

        # 1. 컨테이너 실행 상태
        result = self.execute_cmd("export TMPDIR=/tmp && podman ps --format json")
        try:
            containers = json.loads(result.get("output", "[]") if result["exit_code"] == 0 else "[]")
        except json.JSONDecodeError:
            # JSON 파싱 실패 시 대체 방법 사용
            result = self.execute_cmd("export TMPDIR=/tmp && podman ps --format 'table {{.Names}}\t{{.State}}'")
            containers = []
            if result["exit_code"] == 0 and result.get("output"):
                lines = result["output"].strip().split('\n')[1:]  # 헤더 제거
                for line in lines:
                    if line.strip():
                        parts = line.split()
                        if len(parts) >= 2:
                            containers.append({
                                "Names": [parts[0]],
                                "State": parts[1]
                            })

        running_containers = []
        for container in containers:
            running_containers.append({
                "name": container.get("Names", ["unknown"])[0],
                "state": container.get("State", "unknown"),
                "created": container.get("Created", "unknown")
            })

        self.evidence.append({
            "type": "podman_containers",
            "running_count": len(running_containers),
            "containers": running_containers,
            "verified": len(running_containers) >= 3
        })

        # 2. 리소스 사용량 검증
        stats = self.execute_cmd("export TMPDIR=/tmp && podman stats --no-stream --format json")
        if stats["exit_code"] == 0:
            try:
                stats_data = json.loads(stats.get("output", "[]"))
                self.evidence.append({
                    "type": "resource_usage",
                    "containers_monitored": len(stats_data),
                    "verified": True
                })
            except json.JSONDecodeError:
                # JSON 파싱 실패 시 기본 검증
                stats = self.execute_cmd("export TMPDIR=/tmp && podman stats --no-stream | wc -l")
                if stats["exit_code"] == 0:
                    line_count = int(stats.get("output", "0").strip())
                    self.evidence.append({
                        "type": "resource_usage",
                        "containers_monitored": max(0, line_count - 1),  # 헤더 제외
                        "verified": line_count > 1
                    })

        # 3. 네트워크 포트 검증
        ports = [8081, 9876]
        active_ports = []
        for port in ports:
            result = self.execute_cmd(f"lsof -i :{port} | grep LISTEN | wc -l")
            if result["exit_code"] == 0 and int(result.get("output", "0").strip()) > 0:
                active_ports.append(port)

        self.evidence.append({
            "type": "network_ports",
            "expected": ports,
            "active": active_ports,
            "verified": len(active_ports) == len(ports)
        })

        return all(e["verified"] for e in self.evidence if "verified" in e)

    def verify_filesystem_integrity(self):
        """파일시스템 무결성 검증"""
        print("📁 Filesystem Integrity Verification")

        critical_paths = [
            "/Users/97layer/97layerOS/execution",
            "/Users/97layer/97layerOS/knowledge",
            "/Users/97layer/97layerOS/core",
            "/Users/97layer/97layerOS/libs"
        ]

        verified_paths = []
        for path in critical_paths:
            result = self.execute_cmd(f"test -d {path} && echo EXISTS")
            if "EXISTS" in result.get("output", ""):
                verified_paths.append(path)

        self.evidence.append({
            "type": "filesystem",
            "critical_paths": len(critical_paths),
            "verified_paths": len(verified_paths),
            "verified": len(verified_paths) == len(critical_paths)
        })

        return len(verified_paths) == len(critical_paths)

    def verify_recent_activity(self):
        """최근 활동 검증"""
        print("⚡ Recent Activity Verification")

        # 최근 5분 내 로그 확인
        containers = ["97layer-snapshot", "97layer-gcp-mgmt", "97layer-receiver"]
        active_containers = []

        for container in containers:
            result = self.execute_cmd(
                f"export TMPDIR=/tmp && podman logs {container} --since 5m 2>&1 | wc -l"
            )
            if result["exit_code"] == 0:
                log_lines = int(result.get("output", "0").strip())
                if log_lines > 0:
                    active_containers.append(container)

        self.evidence.append({
            "type": "recent_activity",
            "total_containers": len(containers),
            "active_containers": len(active_containers),
            "verified": len(active_containers) >= 2  # 최소 2개 이상 활성
        })

        return len(active_containers) >= 2

    def generate_final_report(self):
        """최종 보고서 생성"""
        print("\n" + "="*50)
        print("     97LAYEROS FINAL VERIFICATION REPORT")
        print("="*50)

        # 인프라 검증
        infra_ok = self.verify_podman_infrastructure()
        print(f"✓ Podman Infrastructure: {'STABLE' if infra_ok else 'UNSTABLE'}")

        # 파일시스템 검증
        fs_ok = self.verify_filesystem_integrity()
        print(f"✓ Filesystem Integrity: {'VERIFIED' if fs_ok else 'FAILED'}")

        # 활동 검증
        activity_ok = self.verify_recent_activity()
        print(f"✓ Recent Activity: {'ACTIVE' if activity_ok else 'INACTIVE'}")

        print("\n📊 Evidence Summary:")
        print("-" * 40)

        for evidence in self.evidence:
            print(f"• {evidence['type']}: {evidence.get('verified', 'N/A')}")
            if evidence['type'] == 'podman_containers':
                for c in evidence.get('containers', []):
                    print(f"  - {c['name']}: {c['state']}")

        # 최종 판정
        all_verified = all([infra_ok, fs_ok, activity_ok])

        print("\n" + "="*50)
        if all_verified:
            print("✅ All systems operational")
            print("✅ Physical verification complete")
            print("✅ Exit code: 0")
            print("\n" + "🟢"*20)
            print("        [SYSTEM_STABLE]")
            print("🟢"*20)
            return 0
        else:
            print("❌ System instability detected")
            print("❌ Manual intervention required")
            print("\n" + "🔴"*20)
            print("      [SYSTEM_UNSTABLE]")
            print("🔴"*20)
            return 1

if __name__ == "__main__":
    verifier = FinalSystemVerification()
    exit_code = verifier.generate_final_report()

    # 물리적 증거 파일 생성
    with open("/Users/97layer/97layerOS/execution/system/verification_evidence.json", "w") as f:
        json.dump({
            "timestamp": datetime.now().isoformat(),
            "evidence": verifier.evidence,
            "exit_code": exit_code,
            "token": "[SYSTEM_STABLE]" if exit_code == 0 else "[SYSTEM_UNSTABLE]"
        }, f, indent=2)

    exit(exit_code)