#!/usr/bin/env python3
"""
System Monitor - 실시간 시스템 모니터링 대시보드
터미널에서 실행하는 TUI 기반 모니터링

Features:
- 실시간 프로세스 상태
- 리소스 사용량
- 에이전트 활동
- 로그 스트리밍
"""

import os
import sys
import json
import time
import psutil
import threading
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Optional, Any
from collections import deque

# Rich TUI 지원 (설치 필요 시 자동 설치)
try:
    from rich.console import Console
    from rich.table import Table
    from rich.panel import Panel
    from rich.layout import Layout
    from rich.live import Live
    from rich.text import Text
    from rich.progress import Progress, BarColumn, TextColumn
except ImportError:
    print("Installing rich for better display...")
    os.system(f"{sys.executable} -m pip install rich")
    from rich.console import Console
    from rich.table import Table
    from rich.panel import Panel
    from rich.layout import Layout
    from rich.live import Live
    from rich.text import Text
    from rich.progress import Progress, BarColumn, TextColumn

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.append(str(PROJECT_ROOT))

console = Console()


class SystemMonitor:
    """시스템 모니터"""

    def __init__(self):
        self.project_root = PROJECT_ROOT
        self.running = True

        # 모니터링 데이터
        self.process_data = {}
        self.resource_data = {}
        self.agent_data = {}
        self.log_buffer = deque(maxlen=20)
        self.alert_buffer = deque(maxlen=10)

        # 서비스 목록
        self.services = {
            "telegram_daemon": "Telegram Bot",
            "async_telegram": "Async Telegram",
            "mac_realtime_receiver": "Sync Receiver",
            "gcp_management": "GCP Manager"
        }

    def collect_data(self):
        """데이터 수집"""
        while self.running:
            try:
                # 프로세스 상태
                self._collect_process_status()

                # 리소스 사용량
                self._collect_resource_usage()

                # 에이전트 상태
                self._collect_agent_status()

                # 최신 로그
                self._collect_logs()

                time.sleep(1)

            except Exception as e:
                self.log_buffer.append(f"[ERROR] Monitor: {e}")

    def _collect_process_status(self):
        """프로세스 상태 수집"""
        for service_id, service_name in self.services.items():
            # 프로세스 찾기
            found = False
            for proc in psutil.process_iter(['pid', 'name', 'cmdline', 'memory_info', 'cpu_percent']):
                try:
                    cmdline = ' '.join(proc.info['cmdline'] or [])
                    if f"{service_id}.py" in cmdline:
                        self.process_data[service_id] = {
                            "name": service_name,
                            "pid": proc.info['pid'],
                            "status": "Running",
                            "memory_mb": proc.info['memory_info'].rss / 1024 / 1024,
                            "cpu_percent": proc.info['cpu_percent']
                        }
                        found = True
                        break
                except (psutil.NoSuchProcess, psutil.AccessDenied):
                    continue

            if not found:
                self.process_data[service_id] = {
                    "name": service_name,
                    "pid": None,
                    "status": "Stopped",
                    "memory_mb": 0,
                    "cpu_percent": 0
                }

    def _collect_resource_usage(self):
        """리소스 사용량 수집"""
        self.resource_data = {
            "cpu_percent": psutil.cpu_percent(interval=0.1),
            "memory": psutil.virtual_memory(),
            "disk": psutil.disk_usage('/'),
            "network": psutil.net_io_counters()
        }

    def _collect_agent_status(self):
        """에이전트 상태 수집"""
        try:
            # system_state.json 읽기
            state_file = self.project_root / "knowledge" / "system_state.json"
            if state_file.exists():
                with open(state_file) as f:
                    state = json.load(f)
                    self.agent_data = state.get("agents", {})

            # 알림 확인
            notif_dir = self.project_root / "knowledge" / "notifications"
            if notif_dir.exists():
                latest_files = sorted(notif_dir.glob("*.jsonl"), key=lambda x: x.stat().st_mtime)
                if latest_files:
                    # 최신 알림 읽기
                    with open(latest_files[-1]) as f:
                        lines = f.readlines()
                        if lines:
                            latest = json.loads(lines[-1])
                            self.alert_buffer.append(
                                f"[{latest.get('timestamp', '')}] {latest.get('data', {}).get('type', 'notification')}"
                            )

        except Exception as e:
            pass

    def _collect_logs(self):
        """최신 로그 수집"""
        try:
            log_dir = self.project_root / "logs"
            if log_dir.exists():
                # 오늘 날짜의 로그 파일들
                today = datetime.now().strftime("%Y%m%d")
                log_files = list(log_dir.glob(f"*{today}*.log"))

                for log_file in log_files[-3:]:  # 최근 3개 파일
                    try:
                        # 마지막 줄 읽기
                        with open(log_file, 'r') as f:
                            lines = f.readlines()
                            if lines:
                                last_line = lines[-1].strip()
                                if last_line:
                                    service = log_file.stem.split('_')[0]
                                    self.log_buffer.append(f"[{service}] {last_line[:100]}")
                    except:
                        pass

        except Exception as e:
            pass

    def create_dashboard(self) -> Layout:
        """대시보드 레이아웃 생성"""
        layout = Layout()

        # 메인 레이아웃 분할
        layout.split_column(
            Layout(name="header", size=3),
            Layout(name="body"),
            Layout(name="footer", size=3)
        )

        # 헤더
        layout["header"].update(self._create_header())

        # 바디를 2개 컬럼으로
        layout["body"].split_row(
            Layout(name="left"),
            Layout(name="right")
        )

        # 왼쪽: 프로세스 + 리소스
        layout["body"]["left"].split_column(
            Layout(self._create_process_table(), name="processes"),
            Layout(self._create_resource_panel(), name="resources", size=10)
        )

        # 오른쪽: 에이전트 + 로그
        layout["body"]["right"].split_column(
            Layout(self._create_agent_table(), name="agents", size=12),
            Layout(self._create_log_panel(), name="logs")
        )

        # 푸터
        layout["footer"].update(self._create_footer())

        return layout

    def _create_header(self) -> Panel:
        """헤더 생성"""
        text = Text("97LAYER OS SYSTEM MONITOR", style="bold cyan", justify="center")
        return Panel(text, style="bold blue")

    def _create_process_table(self) -> Panel:
        """프로세스 테이블"""
        table = Table(title="Service Status", show_header=True, header_style="bold magenta")
        table.add_column("Service", style="cyan", width=20)
        table.add_column("Status", width=10)
        table.add_column("PID", width=8)
        table.add_column("Memory", width=10)
        table.add_column("CPU", width=8)

        for service_id, data in self.process_data.items():
            status = data["status"]
            status_style = "green" if status == "Running" else "red"

            table.add_row(
                data["name"],
                Text(status, style=status_style),
                str(data["pid"] or "-"),
                f"{data['memory_mb']:.1f} MB",
                f"{data['cpu_percent']:.1f}%"
            )

        return Panel(table, title="🔧 Services", border_style="cyan")

    def _create_resource_panel(self) -> Panel:
        """리소스 패널"""
        if not self.resource_data:
            return Panel("Loading...", title="📊 Resources")

        # CPU 바
        cpu = self.resource_data.get("cpu_percent", 0)
        cpu_bar = self._create_bar(cpu, 100, 40)

        # 메모리 바
        mem = self.resource_data.get("memory", psutil.virtual_memory())
        mem_bar = self._create_bar(mem.percent, 100, 40)

        # 디스크 바
        disk = self.resource_data.get("disk", psutil.disk_usage('/'))
        disk_bar = self._create_bar(disk.percent, 100, 40)

        text = f"""CPU:  {cpu_bar} {cpu:.1f}%
MEM:  {mem_bar} {mem.percent:.1f}% ({mem.used//1024//1024//1024}GB/{mem.total//1024//1024//1024}GB)
DISK: {disk_bar} {disk.percent:.1f}% ({disk.used//1024//1024//1024}GB/{disk.total//1024//1024//1024}GB)"""

        return Panel(text, title="📊 System Resources", border_style="yellow")

    def _create_agent_table(self) -> Panel:
        """에이전트 테이블"""
        table = Table(show_header=True, header_style="bold magenta")
        table.add_column("Agent", style="cyan", width=18)
        table.add_column("Status", width=10)
        table.add_column("Task", width=30)

        for agent_key, agent_data in self.agent_data.items():
            status = agent_data.get("status", "UNKNOWN")
            status_style = "green" if status == "ONLINE" else "yellow"

            table.add_row(
                agent_key.replace("_", " "),
                Text(status, style=status_style),
                agent_data.get("current_task", "-")[:30]
            )

        # 최근 알림 추가
        if self.alert_buffer:
            table.add_row("", "", "")
            table.add_row(
                Text("Recent Alerts:", style="bold yellow"),
                "",
                ""
            )
            for alert in list(self.alert_buffer)[-3:]:
                table.add_row("", "", alert[:50])

        return Panel(table, title="🤖 Agents", border_style="magenta")

    def _create_log_panel(self) -> Panel:
        """로그 패널"""
        log_text = "\n".join(list(self.log_buffer)[-10:]) or "No recent logs"
        return Panel(log_text, title="📜 Recent Logs", border_style="green")

    def _create_footer(self) -> Panel:
        """푸터"""
        commands = "[q] Quit  [r] Restart Service  [s] Stop All  [h] Help"
        return Panel(commands, style="dim")

    def _create_bar(self, value: float, max_value: float, width: int) -> str:
        """프로그레스 바 생성"""
        filled = int((value / max_value) * width)
        bar = "█" * filled + "░" * (width - filled)

        # 색상 결정
        if value < 50:
            color = "green"
        elif value < 80:
            color = "yellow"
        else:
            color = "red"

        return f"[{color}]{bar}[/{color}]"

    def run(self):
        """모니터 실행"""
        # 데이터 수집 스레드
        collector = threading.Thread(target=self.collect_data, daemon=True)
        collector.start()

        # 초기 데이터 수집 대기
        time.sleep(1)

        # Live 디스플레이
        with Live(self.create_dashboard(), refresh_per_second=1, screen=True) as live:
            try:
                while self.running:
                    live.update(self.create_dashboard())
                    time.sleep(1)

                    # 키 입력 체크 (간단한 구현)
                    # 실제로는 keyboard 라이브러리 사용 권장

            except KeyboardInterrupt:
                self.running = False

        console.print("\n[bold red]Monitor stopped.[/bold red]")


class QuickStatus:
    """빠른 상태 확인"""

    @staticmethod
    def show():
        """간단한 상태 표시"""
        console = Console()

        # 프로세스 체크
        services = {
            "telegram_daemon": "Telegram Bot",
            "async_telegram": "Async Telegram",
            "mac_realtime_receiver": "Sync Receiver",
            "gcp_management": "GCP Manager"
        }

        table = Table(title="97LAYER OS Status", show_header=True)
        table.add_column("Service", style="cyan")
        table.add_column("Status")
        table.add_column("PID")

        for service_id, service_name in services.items():
            found = False
            for proc in psutil.process_iter(['pid', 'cmdline']):
                try:
                    cmdline = ' '.join(proc.info['cmdline'] or [])
                    if f"{service_id}.py" in cmdline:
                        table.add_row(
                            service_name,
                            Text("✅ Running", style="green"),
                            str(proc.info['pid'])
                        )
                        found = True
                        break
                except:
                    continue

            if not found:
                table.add_row(
                    service_name,
                    Text("❌ Stopped", style="red"),
                    "-"
                )

        # 시스템 리소스
        cpu = psutil.cpu_percent(interval=0.5)
        mem = psutil.virtual_memory()
        disk = psutil.disk_usage('/')

        console.print(table)
        console.print("\n[bold]System Resources:[/bold]")
        console.print(f"  CPU:  {cpu:.1f}%")
        console.print(f"  MEM:  {mem.percent:.1f}% ({mem.used//1024//1024//1024}GB/{mem.total//1024//1024//1024}GB)")
        console.print(f"  DISK: {disk.percent:.1f}% ({disk.used//1024//1024//1024}GB/{disk.total//1024//1024//1024}GB)")


def main():
    """메인 함수"""
    if len(sys.argv) > 1 and sys.argv[1] == "quick":
        # 빠른 상태 확인
        QuickStatus.show()
    else:
        # 전체 모니터링
        monitor = SystemMonitor()
        monitor.run()


if __name__ == "__main__":
    main()