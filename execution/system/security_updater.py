#!/usr/bin/env python3
"""
보안 업데이트 스케줄러 (Security Update Scheduler)
주 1회 종속성 라이브러리 보안 업데이트 및 무결성 테스트

Features:
- pip 패키지 보안 업데이트 (pip-audit)
- requirements.txt 버전 갱신
- 시스템 무결성 테스트
- 업데이트 이력 로깅
"""

import os
import sys
import json
import subprocess
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Tuple
import logging

# Project Root
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

# Logging Setup
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - [SecurityUpdate] - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# 파일 경로
REQUIREMENTS_FILE = PROJECT_ROOT / "requirements.txt"
SECURITY_LOG = PROJECT_ROOT / "knowledge" / "system" / "security_update_log.json"
LAST_UPDATE_FILE = PROJECT_ROOT / ".tmp" / "last_security_update"


class SecurityUpdater:
    """보안 업데이트 관리자"""

    def __init__(self, auto_update: bool = True):
        self.auto_update = auto_update
        self.requirements_file = REQUIREMENTS_FILE
        self.security_log = SECURITY_LOG
        self.last_update_file = LAST_UPDATE_FILE

        # 초기화
        self.security_log.parent.mkdir(parents=True, exist_ok=True)
        self.last_update_file.parent.mkdir(parents=True, exist_ok=True)

        # 로그 로드
        self.update_history: List[Dict] = self._load_update_history()

    def _load_update_history(self) -> List[Dict]:
        """업데이트 이력 로드"""
        if self.security_log.exists():
            try:
                with open(self.security_log, 'r', encoding='utf-8') as f:
                    return json.load(f)
            except Exception as e:
                logger.error(f"Failed to load security log: {e}")
                return []
        return []

    def _save_update_history(self):
        """업데이트 이력 저장"""
        try:
            with open(self.security_log, 'w', encoding='utf-8') as f:
                json.dump(self.update_history, f, indent=2, ensure_ascii=False)
        except Exception as e:
            logger.error(f"Failed to save security log: {e}")

    def _get_last_update_time(self) -> Optional[datetime]:
        """마지막 업데이트 시간 조회"""
        if self.last_update_file.exists():
            try:
                with open(self.last_update_file, 'r') as f:
                    timestamp_str = f.read().strip()
                    return datetime.fromisoformat(timestamp_str)
            except Exception as e:
                logger.error(f"Failed to read last update time: {e}")
                return None
        return None

    def _set_last_update_time(self, timestamp: datetime):
        """마지막 업데이트 시간 기록"""
        try:
            with open(self.last_update_file, 'w') as f:
                f.write(timestamp.isoformat())
        except Exception as e:
            logger.error(f"Failed to write last update time: {e}")

    def should_update(self) -> Tuple[bool, Optional[str]]:
        """
        업데이트 필요 여부 확인 (주 1회 기준)

        Returns:
            (업데이트 필요 여부, 사유)
        """
        last_update = self._get_last_update_time()

        if last_update is None:
            return True, "First time security update"

        days_since_update = (datetime.now() - last_update).days

        if days_since_update >= 7:
            return True, f"Last update was {days_since_update} days ago"

        return False, f"Last update was {days_since_update} days ago (< 7 days)"

    def check_outdated_packages(self) -> List[Dict]:
        """구버전 패키지 확인"""
        logger.info("Checking for outdated packages...")

        try:
            result = subprocess.run(
                [sys.executable, "-m", "pip", "list", "--outdated", "--format=json"],
                capture_output=True,
                text=True,
                check=True
            )

            outdated = json.loads(result.stdout)
            logger.info(f"Found {len(outdated)} outdated packages")

            return outdated

        except subprocess.CalledProcessError as e:
            logger.error(f"Failed to check outdated packages: {e}")
            return []
        except json.JSONDecodeError as e:
            logger.error(f"Failed to parse pip output: {e}")
            return []

    def check_vulnerabilities(self) -> List[Dict]:
        """보안 취약점 검사 (pip-audit 사용)"""
        logger.info("Checking for vulnerabilities...")

        try:
            # pip-audit 설치 확인
            subprocess.run(
                [sys.executable, "-m", "pip", "show", "pip-audit"],
                capture_output=True,
                check=True
            )
        except subprocess.CalledProcessError:
            logger.warning("pip-audit not installed, installing...")
            try:
                subprocess.run(
                    [sys.executable, "-m", "pip", "install", "--quiet", "pip-audit"],
                    check=True
                )
                logger.info("✓ pip-audit installed")
            except subprocess.CalledProcessError as e:
                logger.error(f"Failed to install pip-audit: {e}")
                return []

        # 취약점 스캔
        try:
            result = subprocess.run(
                [sys.executable, "-m", "pip_audit", "--format=json"],
                capture_output=True,
                text=True,
                timeout=300  # 5분 타임아웃
            )

            # pip-audit은 취약점 발견 시 exit code 1 반환
            if result.returncode == 0:
                logger.info("✓ No vulnerabilities found")
                return []

            vulnerabilities = json.loads(result.stdout)
            logger.warning(f"Found {len(vulnerabilities.get('dependencies', []))} vulnerable packages")

            return vulnerabilities.get("dependencies", [])

        except subprocess.TimeoutExpired:
            logger.error("Vulnerability scan timed out")
            return []
        except subprocess.CalledProcessError as e:
            logger.error(f"Vulnerability scan failed: {e}")
            return []
        except json.JSONDecodeError as e:
            logger.error(f"Failed to parse pip-audit output: {e}")
            return []

    def update_packages(self, packages: List[str]) -> Dict[str, bool]:
        """패키지 업데이트 실행"""
        logger.info(f"Updating {len(packages)} packages...")

        results = {}

        for package in packages:
            try:
                subprocess.run(
                    [sys.executable, "-m", "pip", "install", "--upgrade", "--quiet", package],
                    capture_output=True,
                    check=True,
                    timeout=120  # 2분 타임아웃
                )
                logger.info(f"✓ Updated: {package}")
                results[package] = True

            except subprocess.CalledProcessError as e:
                logger.error(f"Failed to update {package}: {e}")
                results[package] = False
            except subprocess.TimeoutExpired:
                logger.error(f"Update timed out for {package}")
                results[package] = False

        success_count = sum(1 for v in results.values() if v)
        logger.info(f"✓ Updated {success_count}/{len(packages)} packages")

        return results

    def update_requirements_file(self):
        """requirements.txt 파일 갱신 (현재 설치된 버전으로)"""
        logger.info("Updating requirements.txt...")

        try:
            result = subprocess.run(
                [sys.executable, "-m", "pip", "freeze"],
                capture_output=True,
                text=True,
                check=True
            )

            # 백업
            if self.requirements_file.exists():
                backup_path = self.requirements_file.with_suffix(".txt.bak")
                import shutil
                shutil.copy2(self.requirements_file, backup_path)
                logger.info(f"✓ Backed up to {backup_path.name}")

            # 새 requirements.txt 작성
            with open(self.requirements_file, 'w', encoding='utf-8') as f:
                f.write(result.stdout)

            logger.info("✓ requirements.txt updated")

        except Exception as e:
            logger.error(f"Failed to update requirements.txt: {e}")

    def run_integrity_tests(self) -> Dict:
        """시스템 무결성 테스트"""
        logger.info("Running integrity tests...")

        tests = {
            "import_test": self._test_imports(),
            "environment_check": self._test_environment(),
            "critical_files": self._test_critical_files()
        }

        all_passed = all(tests.values())
        status = "PASS" if all_passed else "FAIL"

        logger.info(f"Integrity tests: {status}")

        return {
            "status": status,
            "tests": tests,
            "timestamp": datetime.now().isoformat()
        }

    def _test_imports(self) -> bool:
        """핵심 모듈 임포트 테스트"""
        try:
            # 핵심 라이브러리 임포트 시도
            import google.generativeai
            import anthropic
            import requests
            import psutil

            logger.info("✓ Import test passed")
            return True

        except ImportError as e:
            logger.error(f"Import test failed: {e}")
            return False

    def _test_environment(self) -> bool:
        """환경 설정 테스트"""
        try:
            from libs.core_config import ENVIRONMENT, AGENT_CREW

            if not AGENT_CREW:
                logger.error("AGENT_CREW not configured")
                return False

            logger.info("✓ Environment test passed")
            return True

        except Exception as e:
            logger.error(f"Environment test failed: {e}")
            return False

    def _test_critical_files(self) -> bool:
        """핵심 파일 존재 확인"""
        critical_files = [
            PROJECT_ROOT / "libs" / "core_config.py",
            PROJECT_ROOT / "CLAUDE.md",
            PROJECT_ROOT / "execution" / "system" / "hybrid_sync.py"
        ]

        all_exist = all(f.exists() for f in critical_files)

        if all_exist:
            logger.info("✓ Critical files test passed")
        else:
            missing = [f.name for f in critical_files if not f.exists()]
            logger.error(f"Missing critical files: {missing}")

        return all_exist

    def run_security_update(self) -> Dict:
        """보안 업데이트 실행"""
        logger.info("=" * 60)
        logger.info("Security Update Cycle Started")
        logger.info("=" * 60)

        update_record = {
            "timestamp": datetime.now().isoformat(),
            "status": "STARTED"
        }

        # 1. 구버전 패키지 확인
        outdated = self.check_outdated_packages()
        update_record["outdated_packages"] = len(outdated)

        # 2. 취약점 검사
        vulnerabilities = self.check_vulnerabilities()
        update_record["vulnerabilities"] = len(vulnerabilities)

        # 3. 업데이트 대상 결정
        packages_to_update = []

        # 취약점이 있는 패키지 우선
        for vuln in vulnerabilities:
            package_name = vuln.get("name")
            if package_name:
                packages_to_update.append(package_name)

        # 구버전 패키지 추가 (선택적)
        if self.auto_update:
            for pkg in outdated:
                if pkg["name"] not in packages_to_update:
                    packages_to_update.append(pkg["name"])

        update_record["packages_to_update"] = len(packages_to_update)

        # 4. 업데이트 실행
        if packages_to_update:
            logger.info(f"Updating {len(packages_to_update)} packages...")
            update_results = self.update_packages(packages_to_update)
            update_record["update_results"] = update_results

            # requirements.txt 갱신
            self.update_requirements_file()
        else:
            logger.info("No packages to update")
            update_record["update_results"] = {}

        # 5. 무결성 테스트
        integrity_result = self.run_integrity_tests()
        update_record["integrity_test"] = integrity_result

        # 6. 최종 상태
        if integrity_result["status"] == "PASS":
            update_record["status"] = "SUCCESS"
            logger.info("✅ Security update completed successfully")
        else:
            update_record["status"] = "FAILED"
            logger.error("⚠️ Security update completed with errors")

        # 7. 이력 저장
        self.update_history.append(update_record)
        self._save_update_history()

        # 8. 업데이트 시간 기록
        self._set_last_update_time(datetime.now())

        logger.info("=" * 60)

        return update_record


def main():
    """CLI Entry Point"""
    import argparse

    parser = argparse.ArgumentParser(description="97layerOS Security Updater")
    parser.add_argument(
        "--force",
        action="store_true",
        help="Force update regardless of last update time"
    )
    parser.add_argument(
        "--no-auto-update",
        action="store_true",
        help="Only update vulnerable packages, not all outdated"
    )
    parser.add_argument(
        "--check-only",
        action="store_true",
        help="Check only, do not update"
    )

    args = parser.parse_args()

    updater = SecurityUpdater(auto_update=not args.no_auto_update)

    # 업데이트 필요 확인
    should_update, reason = updater.should_update()

    if not should_update and not args.force:
        logger.info(f"No update needed: {reason}")
        sys.exit(0)

    logger.info(f"Running security update: {reason}")

    if args.check_only:
        # 체크만
        outdated = updater.check_outdated_packages()
        vulnerabilities = updater.check_vulnerabilities()

        print(f"\nOutdated packages: {len(outdated)}")
        print(f"Vulnerabilities: {len(vulnerabilities)}")
        sys.exit(0)

    # 업데이트 실행
    result = updater.run_security_update()

    print(json.dumps(result, indent=2, ensure_ascii=False))

    sys.exit(0 if result["status"] == "SUCCESS" else 1)


if __name__ == "__main__":
    main()
