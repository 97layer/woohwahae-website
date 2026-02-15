#!/usr/bin/env python3
"""
Master Controller - 97LAYER OS 중앙 제어 시스템
모든 프로세스를 자동으로 관리하고 모니터링

Features:
- 전체 서비스 자동 시작/중지
- 프로세스 상태 모니터링
- 자동 복구 시스템
- 로그 수집 및 분석
- 성능 최적화
"""

import os
import sys
import json
import time
import subprocess
import psutil
import signal
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Optional, Any
import threading
import logging

# 프로젝트 루트
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.append(str(PROJECT_ROOT))

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class MasterController:
    """마스터 컨트롤러"""

    def __init__(self):
        self.project_root = PROJECT_ROOT
        self.processes: Dict[str, subprocess.Popen] = {}
        self.service_config = self._load_service_config()
        self.monitoring_active = True
        self.stats = {
            "start_time": datetime.now(),
            "restarts": {},
            "errors": {},
            "uptime": {}
        }

    def _load_service_config(self) -> Dict[str, Any]:
        """서비스 설정 로드"""
        return {
            "telegram_daemon": {
                "name": "Telegram Daemon",
                "command": [sys.executable, "execution/telegram_daemon.py"],
                "workdir": self.project_root,
                "auto_restart": True,
                "restart_delay": 5,
                "critical": True,
                "health_check": self._check_telegram_health
            },
            "async_telegram": {
                "name": "Async Telegram Multimodal Bot",
                "command": [sys.executable, "execution/async_telegram_daemon.py"],
                "workdir": self.project_root,
                "auto_restart": True,
                "restart_delay": 5,
                "critical": True,  # Multimodal system is now critical
                "health_check": self._check_async_telegram_health
            },
            "mac_sync_receiver": {
                "name": "Mac Sync Receiver",
                "command": [sys.executable, "execution/ops/mac_realtime_receiver.py"],
                "workdir": self.project_root,
                "auto_restart": True,
                "restart_delay": 3,
                "critical": True,
                "health_check": self._check_sync_health
            },
            "gcp_management": {
                "name": "GCP Management Server",
                "command": [sys.executable, "execution/ops/gcp_management_server.py"],
                "workdir": self.project_root,
                "auto_restart": True,
                "restart_delay": 5,
                "critical": False,
                "health_check": None
            }
        }

    def start_all(self):
        """모든 서비스 시작"""
        logger.info("🚀 Starting all services...")

        # 환경 체크
        self._check_environment()

        # 서비스 시작 순서 (의존성 고려)
        start_order = [
            "mac_sync_receiver",    # 동기화 수신 먼저
            "async_telegram",       # 비동기 멀티모달 버전 (우선)
            "gcp_management"        # 관리 서버
        ]

        for service_id in start_order:
            if service_id in self.service_config:
                self.start_service(service_id)
                time.sleep(2)  # 서비스 간 시작 간격

        # 모니터링 스레드 시작
        monitor_thread = threading.Thread(target=self._monitor_loop, daemon=True)
        monitor_thread.start()

        logger.info("✅ All services started successfully")
        self._show_status()

    def start_service(self, service_id: str) -> bool:
        """개별 서비스 시작"""
        if service_id in self.processes and self.processes[service_id].poll() is None:
            logger.warning(f"{service_id} is already running")
            return True

        config = self.service_config.get(service_id)
        if not config:
            logger.error(f"Unknown service: {service_id}")
            return False

        try:
            logger.info(f"Starting {config['name']}...")

            # 로그 파일 준비
            log_dir = self.project_root / "logs"
            log_dir.mkdir(exist_ok=True)
            log_file = log_dir / f"{service_id}_{datetime.now().strftime('%Y%m%d')}.log"

            # 프로세스 시작
            with open(log_file, 'a') as log:
                process = subprocess.Popen(
                    config['command'],
                    cwd=config['workdir'],
                    stdout=log,
                    stderr=subprocess.STDOUT,
                    preexec_fn=os.setsid if os.name != 'nt' else None
                )

            self.processes[service_id] = process
            self.stats["uptime"][service_id] = datetime.now()

            # 시작 확인 (3초 대기)
            time.sleep(3)
            if process.poll() is None:
                logger.info(f"✅ {config['name']} started (PID: {process.pid})")
                return True
            else:
                logger.error(f"❌ {config['name']} failed to start")
                return False

        except Exception as e:
            logger.error(f"Failed to start {service_id}: {e}")
            self._record_error(service_id, str(e))
            return False

    def stop_service(self, service_id: str):
        """서비스 중지"""
        if service_id not in self.processes:
            logger.warning(f"{service_id} is not running")
            return

        process = self.processes[service_id]
        if process.poll() is not None:
            logger.warning(f"{service_id} is already stopped")
            return

        config = self.service_config[service_id]
        logger.info(f"Stopping {config['name']}...")

        try:
            # SIGTERM 전송
            if os.name != 'nt':
                os.killpg(os.getpgid(process.pid), signal.SIGTERM)
            else:
                process.terminate()

            # 5초 대기
            process.wait(timeout=5)
            logger.info(f"✅ {config['name']} stopped gracefully")

        except subprocess.TimeoutExpired:
            # 강제 종료
            logger.warning(f"Force killing {config['name']}...")
            if os.name != 'nt':
                os.killpg(os.getpgid(process.pid), signal.SIGKILL)
            else:
                process.kill()
            process.wait()
            logger.info(f"⚠️ {config['name']} force killed")

        finally:
            del self.processes[service_id]

    def stop_all(self):
        """모든 서비스 중지"""
        logger.info("🛑 Stopping all services...")
        self.monitoring_active = False

        for service_id in list(self.processes.keys()):
            self.stop_service(service_id)

        logger.info("✅ All services stopped")

    def restart_service(self, service_id: str):
        """서비스 재시작"""
        config = self.service_config.get(service_id)
        if not config:
            return

        logger.info(f"♻️ Restarting {config['name']}...")

        # 재시작 카운트
        if service_id not in self.stats["restarts"]:
            self.stats["restarts"][service_id] = 0
        self.stats["restarts"][service_id] += 1

        # 중지
        self.stop_service(service_id)

        # 대기
        time.sleep(config.get("restart_delay", 3))

        # 시작
        self.start_service(service_id)

    def _monitor_loop(self):
        """모니터링 루프"""
        logger.info("🔍 Monitoring started")

        while self.monitoring_active:
            try:
                for service_id, process in list(self.processes.items()):
                    config = self.service_config[service_id]

                    # 프로세스 상태 확인
                    if process.poll() is not None:
                        # 프로세스 종료됨
                        logger.warning(f"⚠️ {config['name']} has stopped")

                        if config.get("auto_restart", True):
                            # 자동 재시작
                            restart_count = self.stats["restarts"].get(service_id, 0)

                            if restart_count < 5:  # 최대 5회
                                logger.info(f"Auto-restarting {config['name']}...")
                                self.restart_service(service_id)
                            else:
                                logger.error(f"❌ {config['name']} restart limit reached")
                                if config.get("critical", False):
                                    self._handle_critical_failure(service_id)

                    else:
                        # 헬스 체크
                        health_check = config.get("health_check")
                        if health_check and not health_check():
                            logger.warning(f"⚠️ {config['name']} health check failed")
                            self.restart_service(service_id)

                # CPU/메모리 체크
                self._check_system_resources()

            except Exception as e:
                logger.error(f"Monitor loop error: {e}")

            time.sleep(30)  # 30초마다 체크

    def _check_telegram_health(self) -> bool:
        """텔레그램 헬스 체크"""
        try:
            # 프로세스 메모리 사용량 체크
            if "telegram_daemon" in self.processes:
                process = self.processes["telegram_daemon"]
                if process.poll() is None:
                    proc = psutil.Process(process.pid)
                    memory_mb = proc.memory_info().rss / 1024 / 1024

                    if memory_mb > 500:  # 500MB 초과
                        logger.warning(f"Telegram daemon using {memory_mb:.1f}MB")
                        return False

            return True

        except Exception as e:
            logger.error(f"Health check error: {e}")
            return False

    def _check_async_telegram_health(self) -> bool:
        """Async Telegram 헬스 체크"""
        try:
            # 프로세스 메모리 사용량 체크
            if "async_telegram" in self.processes:
                process = self.processes["async_telegram"]
                if process.poll() is None:
                    proc = psutil.Process(process.pid)
                    memory_mb = proc.memory_info().rss / 1024 / 1024

                    if memory_mb > 800:  # 800MB 초과 (멀티모달이라 더 높음)
                        logger.warning(f"Async Telegram using {memory_mb:.1f}MB")
                        return False

            return True

        except Exception as e:
            logger.error(f"Async Telegram health check error: {e}")
            return False

    def _check_sync_health(self) -> bool:
        """동기화 서버 헬스 체크"""
        try:
            import requests
            response = requests.get("http://localhost:9876/status", timeout=5)
            return response.status_code == 200
        except:
            return False

    def _check_system_resources(self):
        """시스템 리소스 체크"""
        try:
            # CPU 사용률
            cpu_percent = psutil.cpu_percent(interval=1)
            if cpu_percent > 80:
                logger.warning(f"⚠️ High CPU usage: {cpu_percent}%")

            # 메모리 사용률
            memory = psutil.virtual_memory()
            if memory.percent > 85:
                logger.warning(f"⚠️ High memory usage: {memory.percent}%")

            # 디스크 사용률
            disk = psutil.disk_usage('/')
            if disk.percent > 90:
                logger.warning(f"⚠️ Low disk space: {disk.percent}% used")

        except Exception as e:
            logger.error(f"Resource check error: {e}")

    def _check_environment(self):
        """환경 체크"""
        logger.info("Checking environment...")

        # 필수 디렉토리 생성
        required_dirs = [
            self.project_root / "logs",
            self.project_root / "knowledge" / "notifications",
            self.project_root / "knowledge" / "agent_hub",
            self.project_root / ".tmp" / "ai_cache"
        ]

        for dir_path in required_dirs:
            dir_path.mkdir(parents=True, exist_ok=True)

        # Python 패키지 체크
        required_packages = ["aiohttp", "psutil", "requests"]
        missing = []

        for package in required_packages:
            try:
                __import__(package)
            except ImportError:
                missing.append(package)

        if missing:
            logger.warning(f"Missing packages: {missing}")
            logger.info("Installing missing packages...")
            subprocess.run([sys.executable, "-m", "pip", "install"] + missing)

        logger.info("✅ Environment check passed")

    def _handle_critical_failure(self, service_id: str):
        """크리티컬 실패 처리"""
        logger.error(f"🚨 CRITICAL FAILURE: {service_id}")

        # 텔레그램 알림 (가능한 경우)
        try:
            from libs.core_config import TELEGRAM_CONFIG
            import requests

            token = TELEGRAM_CONFIG.get("BOT_TOKEN")
            chat_id = TELEGRAM_CONFIG.get("ADMIN_CHAT_ID")

            if token and chat_id:
                url = f"https://api.telegram.org/bot{token}/sendMessage"
                message = f"🚨 CRITICAL FAILURE\n\nService: {service_id}\nTime: {datetime.now()}\nAction: Manual intervention required"

                requests.post(url, json={
                    "chat_id": chat_id,
                    "text": message
                })
        except:
            pass

        # 로그 파일에 기록
        error_log = self.project_root / "logs" / "critical_errors.log"
        with open(error_log, 'a') as f:
            f.write(f"{datetime.now()} - CRITICAL: {service_id}\n")

    def _record_error(self, service_id: str, error: str):
        """에러 기록"""
        if service_id not in self.stats["errors"]:
            self.stats["errors"][service_id] = []

        self.stats["errors"][service_id].append({
            "time": datetime.now().isoformat(),
            "error": error
        })

        # 최근 10개만 유지
        self.stats["errors"][service_id] = self.stats["errors"][service_id][-10:]

    def _show_status(self):
        """상태 출력"""
        print("\n" + "=" * 60)
        print("97LAYER OS - Service Status")
        print("=" * 60)

        for service_id, process in self.processes.items():
            config = self.service_config[service_id]
            status = "🟢 Running" if process.poll() is None else "🔴 Stopped"

            print(f"{config['name']:<30} {status}")

            if service_id in self.stats["uptime"]:
                uptime = datetime.now() - self.stats["uptime"][service_id]
                print(f"  Uptime: {uptime}")

            if service_id in self.stats["restarts"]:
                print(f"  Restarts: {self.stats['restarts'][service_id]}")

        print("=" * 60)

    def get_status(self) -> Dict[str, Any]:
        """상태 정보 반환"""
        status = {
            "timestamp": datetime.now().isoformat(),
            "services": {},
            "system": {
                "cpu": psutil.cpu_percent(),
                "memory": psutil.virtual_memory().percent,
                "disk": psutil.disk_usage('/').percent
            },
            "stats": self.stats
        }

        for service_id, process in self.processes.items():
            config = self.service_config[service_id]
            status["services"][service_id] = {
                "name": config["name"],
                "running": process.poll() is None,
                "pid": process.pid if process.poll() is None else None
            }

        return status


def main():
    """메인 함수"""
    controller = MasterController()

    # 시그널 핸들러
    def signal_handler(sig, frame):
        print("\n🛑 Shutting down...")
        controller.stop_all()
        sys.exit(0)

    signal.signal(signal.SIGINT, signal_handler)

    # 인자 처리
    if len(sys.argv) > 1:
        command = sys.argv[1]

        if command == "start":
            controller.start_all()
            # 계속 실행
            while True:
                time.sleep(1)

        elif command == "stop":
            controller.stop_all()

        elif command == "restart":
            if len(sys.argv) > 2:
                controller.restart_service(sys.argv[2])
            else:
                controller.stop_all()
                time.sleep(2)
                controller.start_all()

        elif command == "status":
            status = controller.get_status()
            print(json.dumps(status, indent=2))

        else:
            print("Usage: master_controller.py [start|stop|restart|status]")

    else:
        # 기본: 시작
        controller.start_all()
        while True:
            time.sleep(1)


if __name__ == "__main__":
    main()