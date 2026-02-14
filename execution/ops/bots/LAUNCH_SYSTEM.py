#!/usr/bin/env python3
"""
97LAYER OS - FULL SYSTEM LAUNCHER
완전 자동화된 시스템 런처 - 모든 프로세스를 관리

이 스크립트 하나로 전체 시스템이 자동으로 구동됩니다.
"""

import os
import sys
import json
import time
import subprocess
import signal
from pathlib import Path
from datetime import datetime

# 색상 코드
class Colors:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'

def print_banner():
    """배너 출력"""
    banner = f"""{Colors.OKCYAN}{Colors.BOLD}
╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║        ███████╗███████╗██╗      █████╗ ██╗   ██╗███████╗██████╗ ║
║        ██╔══██║╚════██║██║     ██╔══██╗╚██╗ ██╔╝██╔════╝██╔══██╗║
║        ╚██████║   ██╔═╝██║     ███████║ ╚████╔╝ █████╗  ██████╔╝║
║         ╚═══██║   ██║  ██║     ██╔══██║  ╚██╔╝  ██╔══╝  ██╔══██╗║
║        ███████║   ██║  ███████╗██║  ██║   ██║   ███████╗██║  ██║║
║        ╚══════╝   ╚═╝  ╚══════╝╚═╝  ╚═╝   ╚═╝   ╚══════╝╚═╝  ╚═╝║
║                                                                  ║
║                   A N T I - G R A V I T Y   O S                 ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
{Colors.ENDC}"""
    print(banner)


class SystemLauncher:
    """시스템 런처"""

    def __init__(self):
        self.project_root = Path.home() / "97layerOS"
        self.processes = {}
        self.log_dir = self.project_root / "logs"
        self.log_dir.mkdir(exist_ok=True)
        self.start_time = datetime.now()

    def check_environment(self):
        """환경 체크"""
        print(f"\n{Colors.OKBLUE}[1/5] 환경 체크 중...{Colors.ENDC}")

        # Python 버전 체크
        python_version = sys.version_info
        print(f"  ✓ Python {python_version.major}.{python_version.minor}.{python_version.micro}")

        # 필수 디렉토리 생성
        dirs_to_create = [
            "logs",
            "knowledge/notifications",
            "knowledge/agent_hub",
            "knowledge/chat_memory",
            "knowledge/model_context",
            "knowledge/raw_signals",
            "knowledge/reports",
            "knowledge/inbox",
            ".tmp/ai_cache"
        ]

        for dir_path in dirs_to_create:
            full_path = self.project_root / dir_path
            full_path.mkdir(parents=True, exist_ok=True)

        print(f"  ✓ 필수 디렉토리 생성 완료")

        # .env 파일 체크
        env_file = self.project_root / ".env"
        if env_file.exists():
            print(f"  ✓ 환경 변수 파일 확인")
        else:
            print(f"  {Colors.WARNING}⚠ .env 파일이 없습니다. API 키 설정 필요{Colors.ENDC}")

        return True

    def install_dependencies(self):
        """의존성 설치"""
        print(f"\n{Colors.OKBLUE}[2/5] 의존성 체크 중...{Colors.ENDC}")

        required_packages = [
            "aiohttp",
            "psutil",
            "requests",
            "python-dotenv",
            "rich"
        ]

        # pip list로 설치된 패키지 확인
        result = subprocess.run(
            [sys.executable, "-m", "pip", "list", "--format=json"],
            capture_output=True,
            text=True
        )

        if result.returncode == 0:
            installed = {pkg["name"].lower() for pkg in json.loads(result.stdout)}
            missing = [pkg for pkg in required_packages if pkg not in installed]

            if missing:
                print(f"  설치 필요: {', '.join(missing)}")
                subprocess.run([sys.executable, "-m", "pip", "install"] + missing)
                print(f"  ✓ 패키지 설치 완료")
            else:
                print(f"  ✓ 모든 패키지 설치됨")
        else:
            print(f"  {Colors.WARNING}⚠ pip list 실패, 패키지 설치 시도{Colors.ENDC}")
            subprocess.run([sys.executable, "-m", "pip", "install"] + required_packages)

        return True

    def stop_existing_processes(self):
        """기존 프로세스 정리"""
        print(f"\n{Colors.OKBLUE}[3/5] 기존 프로세스 정리 중...{Colors.ENDC}")

        processes_to_kill = [
            "telegram_daemon.py",
            "async_telegram_daemon.py",
            "mac_realtime_receiver.py",
            "mac_sync_receiver.py",
            "gcp_management_server.py",
            "master_controller.py"
        ]

        for process_name in processes_to_kill:
            subprocess.run(["pkill", "-f", process_name], stderr=subprocess.DEVNULL)

        time.sleep(2)
        print(f"  ✓ 기존 프로세스 정리 완료")
        return True

    def start_core_services(self):
        """핵심 서비스 시작"""
        print(f"\n{Colors.OKBLUE}[4/5] 핵심 서비스 시작 중...{Colors.ENDC}")

        services = [
            {
                "name": "Sync Receiver",
                "script": "execution/ops/mac_realtime_receiver.py",
                "critical": True
            },
            {
                "name": "Telegram Daemon",
                "script": "execution/telegram_daemon.py",
                "critical": True
            },
            {
                "name": "Master Controller",
                "script": "execution/ops/master_controller.py",
                "args": ["start"],
                "critical": False
            }
        ]

        for service in services:
            script_path = self.project_root / service["script"]

            if not script_path.exists():
                print(f"  {Colors.WARNING}⚠ {service['name']}: 스크립트 없음{Colors.ENDC}")
                continue

            try:
                # 로그 파일
                log_file = self.log_dir / f"{script_path.stem}_{datetime.now().strftime('%Y%m%d')}.log"

                # 프로세스 시작
                with open(log_file, 'a') as log:
                    cmd = [sys.executable, str(script_path)]
                    if service.get("args"):
                        cmd.extend(service["args"])

                    process = subprocess.Popen(
                        cmd,
                        cwd=self.project_root,
                        stdout=log,
                        stderr=subprocess.STDOUT,
                        preexec_fn=os.setsid if os.name != 'nt' else None
                    )

                self.processes[service["name"]] = process

                # 시작 확인
                time.sleep(2)
                if process.poll() is None:
                    print(f"  ✓ {service['name']}: {Colors.OKGREEN}실행 중{Colors.ENDC} (PID: {process.pid})")
                else:
                    if service.get("critical"):
                        print(f"  {Colors.FAIL}✗ {service['name']}: 시작 실패 (중요){Colors.ENDC}")
                        return False
                    else:
                        print(f"  {Colors.WARNING}⚠ {service['name']}: 시작 실패{Colors.ENDC}")

            except Exception as e:
                print(f"  {Colors.FAIL}✗ {service['name']}: {e}{Colors.ENDC}")
                if service.get("critical"):
                    return False

        return True

    def verify_system(self):
        """시스템 검증"""
        print(f"\n{Colors.OKBLUE}[5/5] 시스템 검증 중...{Colors.ENDC}")

        # 프로세스 상태 확인
        running = 0
        for name, process in self.processes.items():
            if process.poll() is None:
                running += 1

        print(f"  ✓ 실행 중인 서비스: {running}/{len(self.processes)}")

        # 시스템 상태 파일 확인
        state_file = self.project_root / "knowledge" / "system_state.json"
        if state_file.exists():
            with open(state_file) as f:
                state = json.load(f)
                print(f"  ✓ 시스템 상태: {state.get('system_status', 'UNKNOWN')}")

        # 메모리/CPU 체크
        import psutil
        cpu = psutil.cpu_percent(interval=1)
        mem = psutil.virtual_memory()

        print(f"  ✓ CPU: {cpu:.1f}%")
        print(f"  ✓ 메모리: {mem.percent:.1f}% ({mem.used//1024//1024//1024}GB/{mem.total//1024//1024//1024}GB)")

        return True

    def show_final_status(self):
        """최종 상태 표시"""
        print(f"\n{Colors.BOLD}{'='*70}{Colors.ENDC}")
        print(f"{Colors.OKGREEN}{Colors.BOLD}✅ 97LAYER OS 시스템 구동 완료!{Colors.ENDC}")
        print(f"{Colors.BOLD}{'='*70}{Colors.ENDC}")

        print(f"""
{Colors.OKCYAN}📱 텔레그램 봇 명령어:{Colors.ENDC}
  /status - 시스템 상태 확인
  /cd, /td, /ad, /ce, /sa - 에이전트 전환
  /auto - 자동 라우팅 모드
  /council [주제] - 에이전트 위원회 소집
  /evolve - 시스템 진화

{Colors.OKCYAN}💻 터미널 명령어:{Colors.ENDC}
  python3 execution/ops/system_monitor.py - 실시간 모니터링
  python3 execution/ops/system_monitor.py quick - 빠른 상태 확인
  python3 execution/ops/master_controller.py status - 서비스 상태

{Colors.OKCYAN}🔧 관리 명령어:{Colors.ENDC}
  ./start_system.sh - 시스템 재시작
  python3 LAUNCH_SYSTEM.py - 전체 재구동

{Colors.WARNING}⚠️ 시스템 종료: Ctrl+C{Colors.ENDC}
""")

    def handle_shutdown(self, signum, frame):
        """종료 처리"""
        print(f"\n{Colors.WARNING}시스템 종료 중...{Colors.ENDC}")

        for name, process in self.processes.items():
            if process.poll() is None:
                process.terminate()
                print(f"  ✓ {name} 종료")

        time.sleep(2)
        print(f"{Colors.OKGREEN}시스템 종료 완료{Colors.ENDC}")
        sys.exit(0)

    def run(self):
        """실행"""
        print_banner()

        # 시그널 핸들러 등록
        signal.signal(signal.SIGINT, self.handle_shutdown)

        # 단계별 실행
        steps = [
            self.check_environment,
            self.install_dependencies,
            self.stop_existing_processes,
            self.start_core_services,
            self.verify_system
        ]

        for step in steps:
            if not step():
                print(f"\n{Colors.FAIL}❌ 시스템 구동 실패{Colors.ENDC}")
                sys.exit(1)

        # 최종 상태 표시
        self.show_final_status()

        # 백그라운드 실행
        print(f"\n{Colors.OKGREEN}시스템이 백그라운드에서 실행 중입니다.{Colors.ENDC}")
        print("모니터링을 시작하려면: python3 execution/ops/system_monitor.py")

        # 프로세스 유지 (선택적)
        if "--foreground" in sys.argv:
            print("\nForeground 모드 - Ctrl+C로 종료")
            try:
                while True:
                    time.sleep(1)
            except KeyboardInterrupt:
                self.handle_shutdown(None, None)


def main():
    """메인 함수"""
    launcher = SystemLauncher()
    launcher.run()


if __name__ == "__main__":
    main()