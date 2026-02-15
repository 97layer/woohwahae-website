#!/usr/bin/env python3
"""
환경 정합성 검사 (Environment Integrity Check)
모든 실행 전 필수 패키지 및 시스템 요구사항 확인

Features:
- Python 패키지 자동 설치 (Silent Install)
- 필수 디렉토리 생성
- 환경변수 검증
- 시스템 리소스 체크 (메모리, 디스크)
"""

import sys
import subprocess
import json
import os
from pathlib import Path
from typing import List, Dict, Tuple
import logging

# Logging Setup
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - [EnvCheck] - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Project Root
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
REQUIREMENTS_FILE = PROJECT_ROOT / "requirements.txt"
ENV_FILE = PROJECT_ROOT / ".env"

class EnvironmentChecker:
    """환경 정합성 검사 및 자동 복구"""

    def __init__(self, silent_install: bool = True):
        self.silent_install = silent_install
        self.issues: List[Dict] = []
        self.fixes_applied: List[str] = []

    def check_python_version(self) -> bool:
        """Python 버전 확인 (>= 3.9)"""
        version_info = sys.version_info
        required_major = 3
        required_minor = 9

        if version_info.major < required_major or (
            version_info.major == required_major and version_info.minor < required_minor
        ):
            self.issues.append({
                "type": "PYTHON_VERSION",
                "severity": "CRITICAL",
                "current": f"{version_info.major}.{version_info.minor}.{version_info.micro}",
                "required": f">= {required_major}.{required_minor}",
                "fixable": False
            })
            logger.error(f"Python {required_major}.{required_minor}+ required (current: {version_info.major}.{version_info.minor})")
            return False

        logger.info(f"✓ Python {version_info.major}.{version_info.minor}.{version_info.micro}")
        return True

    def check_required_packages(self) -> bool:
        """필수 패키지 확인 및 설치"""
        if not REQUIREMENTS_FILE.exists():
            logger.warning(f"requirements.txt not found at {REQUIREMENTS_FILE}")
            return True

        logger.info("Checking required packages...")

        # requirements.txt 파싱
        with open(REQUIREMENTS_FILE, 'r', encoding='utf-8') as f:
            requirements = [
                line.strip()
                for line in f
                if line.strip() and not line.startswith('#')
            ]

        missing_packages = []

        for requirement in requirements:
            # 패키지명 추출 (버전 정보 제거)
            package_name = requirement.split('==')[0].split('>=')[0].split('<=')[0].strip()

            try:
                __import__(package_name.replace('-', '_'))
            except ImportError:
                missing_packages.append(requirement)
                logger.warning(f"Missing package: {package_name}")

        # 자동 설치 (Silent Install)
        if missing_packages:
            if self.silent_install:
                logger.info(f"Installing {len(missing_packages)} missing packages...")
                try:
                    subprocess.check_call(
                        [sys.executable, "-m", "pip", "install", "--quiet", "--upgrade"] + missing_packages,
                        stdout=subprocess.DEVNULL,
                        stderr=subprocess.PIPE
                    )
                    self.fixes_applied.append(f"Installed {len(missing_packages)} packages")
                    logger.info(f"✓ {len(missing_packages)} packages installed")
                    return True
                except subprocess.CalledProcessError as e:
                    self.issues.append({
                        "type": "PACKAGE_INSTALL_FAILED",
                        "severity": "ERROR",
                        "packages": missing_packages,
                        "error": str(e)
                    })
                    logger.error(f"Failed to install packages: {e}")
                    return False
            else:
                self.issues.append({
                    "type": "MISSING_PACKAGES",
                    "severity": "WARNING",
                    "packages": missing_packages
                })
                logger.warning(f"{len(missing_packages)} packages missing (silent_install=False)")
                return False

        logger.info(f"✓ All {len(requirements)} packages installed")
        return True

    def check_required_directories(self) -> bool:
        """필수 디렉토리 존재 확인 및 생성"""
        required_dirs = [
            PROJECT_ROOT / "knowledge",
            PROJECT_ROOT / "knowledge" / "system",
            PROJECT_ROOT / "knowledge" / "inbox",
            PROJECT_ROOT / "knowledge" / "raw_signals",
            PROJECT_ROOT / "directives",
            PROJECT_ROOT / "directives" / "agents",
            PROJECT_ROOT / "execution",
            PROJECT_ROOT / "execution" / "system",
            PROJECT_ROOT / "libs",
            PROJECT_ROOT / "assets",
            PROJECT_ROOT / ".tmp",
            PROJECT_ROOT / ".tmp" / "nightguard"
        ]

        missing_dirs = []

        for dir_path in required_dirs:
            if not dir_path.exists():
                missing_dirs.append(dir_path)
                logger.warning(f"Missing directory: {dir_path}")

        # 자동 생성
        if missing_dirs:
            if self.silent_install:
                for dir_path in missing_dirs:
                    dir_path.mkdir(parents=True, exist_ok=True)
                    logger.info(f"✓ Created: {dir_path.relative_to(PROJECT_ROOT)}")

                self.fixes_applied.append(f"Created {len(missing_dirs)} directories")
                return True
            else:
                self.issues.append({
                    "type": "MISSING_DIRECTORIES",
                    "severity": "ERROR",
                    "directories": [str(d.relative_to(PROJECT_ROOT)) for d in missing_dirs]
                })
                return False

        logger.info(f"✓ All {len(required_dirs)} directories exist")
        return True

    def check_environment_variables(self) -> bool:
        """환경변수 검증"""
        required_vars = [
            "TELEGRAM_BOT_TOKEN",
            "GEMINI_API_KEY"
        ]

        optional_vars = [
            "ANTHROPIC_API_KEY",
            "INSTAGRAM_ACCESS_TOKEN"
        ]

        missing_vars = []

        for var in required_vars:
            if not os.getenv(var):
                missing_vars.append(var)
                logger.warning(f"Missing required env var: {var}")

        if missing_vars:
            self.issues.append({
                "type": "MISSING_ENV_VARS",
                "severity": "WARNING",
                "variables": missing_vars,
                "action": f"Add to {ENV_FILE}"
            })
            logger.warning(f"{len(missing_vars)} required env vars missing")
            return False

        logger.info(f"✓ {len(required_vars)} required env vars present")
        return True

    def check_system_resources(self) -> bool:
        """시스템 리소스 체크 (메모리, 디스크)"""
        try:
            import psutil

            # 메모리 체크
            memory = psutil.virtual_memory()
            memory_gb = memory.total / (1024 ** 3)

            # 디스크 체크
            disk = psutil.disk_usage(str(PROJECT_ROOT))
            disk_free_gb = disk.free / (1024 ** 3)

            logger.info(f"✓ Memory: {memory_gb:.1f}GB total, {memory.percent}% used")
            logger.info(f"✓ Disk: {disk_free_gb:.1f}GB free, {disk.percent}% used")

            # 경고 임계값
            if memory.percent > 90:
                self.issues.append({
                    "type": "HIGH_MEMORY_USAGE",
                    "severity": "WARNING",
                    "usage_percent": memory.percent
                })
                logger.warning(f"High memory usage: {memory.percent}%")

            if disk.percent > 90:
                self.issues.append({
                    "type": "LOW_DISK_SPACE",
                    "severity": "WARNING",
                    "usage_percent": disk.percent,
                    "free_gb": disk_free_gb
                })
                logger.warning(f"Low disk space: {disk_free_gb:.1f}GB free")

            return True

        except ImportError:
            logger.warning("psutil not installed, skipping resource check")
            return True

    def check_critical_files(self) -> bool:
        """핵심 파일 존재 확인"""
        critical_files = [
            PROJECT_ROOT / "libs" / "core_config.py",
            PROJECT_ROOT / "execution" / "system" / "hybrid_sync.py",
            PROJECT_ROOT / "CLAUDE.md"
        ]

        missing_files = []

        for file_path in critical_files:
            if not file_path.exists():
                missing_files.append(file_path)
                logger.error(f"CRITICAL: Missing file: {file_path}")

        if missing_files:
            self.issues.append({
                "type": "MISSING_CRITICAL_FILES",
                "severity": "CRITICAL",
                "files": [str(f.relative_to(PROJECT_ROOT)) for f in missing_files]
            })
            return False

        logger.info(f"✓ All {len(critical_files)} critical files exist")
        return True

    def run_full_check(self) -> Tuple[bool, Dict]:
        """전체 환경 검사 실행"""
        logger.info("=" * 60)
        logger.info("97layerOS Environment Integrity Check")
        logger.info("=" * 60)

        checks = [
            ("Python Version", self.check_python_version),
            ("Required Packages", self.check_required_packages),
            ("Required Directories", self.check_required_directories),
            ("Environment Variables", self.check_environment_variables),
            ("System Resources", self.check_system_resources),
            ("Critical Files", self.check_critical_files)
        ]

        results = {}
        all_passed = True

        for check_name, check_func in checks:
            logger.info(f"\n[Check] {check_name}...")
            try:
                passed = check_func()
                results[check_name] = "PASS" if passed else "FAIL"
                if not passed:
                    all_passed = False
            except Exception as e:
                logger.error(f"Error during {check_name}: {e}")
                results[check_name] = "ERROR"
                all_passed = False

        logger.info("\n" + "=" * 60)

        if all_passed:
            logger.info("✅ Environment check PASSED")
        else:
            logger.warning(f"⚠️ Environment check FAILED ({len(self.issues)} issues)")
            logger.info("\nIssues:")
            for issue in self.issues:
                logger.warning(f"  - [{issue['severity']}] {issue['type']}")

        if self.fixes_applied:
            logger.info(f"\n✓ Auto-fixes applied: {len(self.fixes_applied)}")
            for fix in self.fixes_applied:
                logger.info(f"  - {fix}")

        logger.info("=" * 60)

        return all_passed, {
            "passed": all_passed,
            "results": results,
            "issues": self.issues,
            "fixes_applied": self.fixes_applied
        }


def main():
    """CLI Entry Point"""
    import argparse

    parser = argparse.ArgumentParser(description="97layerOS Environment Checker")
    parser.add_argument(
        "--no-install",
        action="store_true",
        help="Do not auto-install missing packages"
    )
    parser.add_argument(
        "--json",
        action="store_true",
        help="Output results as JSON"
    )

    args = parser.parse_args()

    checker = EnvironmentChecker(silent_install=not args.no_install)
    passed, report = checker.run_full_check()

    if args.json:
        print(json.dumps(report, indent=2))

    sys.exit(0 if passed else 1)


if __name__ == "__main__":
    main()
