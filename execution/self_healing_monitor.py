#!/usr/bin/env python3
"""
자가 치유 모니터링 시스템
시스템 상태를 지속적으로 모니터링하고 문제를 자동으로 감지 및 수정

Features:
- 순환 참조 감지 및 차단
- 메모리 누수 방지
- 자동 오류 복구
- 시스템 건강도 추적
"""

import json
import os
import sys
import time
import threading
import logging
import traceback
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Optional, Any, Set
from collections import defaultdict
import psutil
import signal

# 프로젝트 루트 경로 설정
PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from libs.memory_manager import MemoryManager
from libs.agent_hub import get_hub

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class CircularDependencyTracker:
    """순환 참조 추적 및 방지"""

    def __init__(self):
        self.call_stack: List[str] = []
        self.dependency_graph: Dict[str, Set[str]] = defaultdict(set)
        self.circular_detections: List[Dict] = []

    def enter_context(self, context: str) -> bool:
        """
        컨텍스트 진입 시 순환 참조 체크

        Returns:
            True if safe, False if circular dependency detected
        """
        if context in self.call_stack:
            # 순환 참조 감지
            cycle_start = self.call_stack.index(context)
            cycle = self.call_stack[cycle_start:] + [context]

            self.circular_detections.append({
                "timestamp": datetime.now().isoformat(),
                "cycle": cycle,
                "full_stack": self.call_stack.copy()
            })

            logger.warning(f"순환 참조 감지: {' -> '.join(cycle)}")
            return False

        self.call_stack.append(context)

        # 의존성 그래프 업데이트
        if len(self.call_stack) > 1:
            caller = self.call_stack[-2]
            self.dependency_graph[caller].add(context)

        return True

    def exit_context(self, context: str):
        """컨텍스트 종료"""
        if self.call_stack and self.call_stack[-1] == context:
            self.call_stack.pop()

    def get_cycle_report(self) -> Dict:
        """순환 참조 보고서"""
        return {
            "total_cycles_detected": len(self.circular_detections),
            "recent_cycles": self.circular_detections[-5:],
            "dependency_graph_size": len(self.dependency_graph)
        }


class MemoryLeakDetector:
    """메모리 누수 감지기"""

    def __init__(self, threshold_mb: float = 500.0):
        self.threshold_mb = threshold_mb
        self.baseline_memory = None
        self.memory_samples: List[Dict] = []
        self.leak_warnings: List[Dict] = []

    def initialize_baseline(self):
        """기준 메모리 사용량 설정"""
        self.baseline_memory = self._get_memory_usage()

    def check_memory(self) -> Optional[Dict]:
        """메모리 상태 체크"""
        current_memory = self._get_memory_usage()

        if self.baseline_memory is None:
            self.initialize_baseline()
            return None

        # 메모리 증가량 계산
        increase = current_memory - self.baseline_memory

        # 샘플 기록
        self.memory_samples.append({
            "timestamp": datetime.now().isoformat(),
            "usage_mb": current_memory,
            "increase_mb": increase
        })

        # 최근 10개 샘플만 유지
        if len(self.memory_samples) > 10:
            self.memory_samples.pop(0)

        # 누수 감지
        if increase > self.threshold_mb:
            warning = {
                "timestamp": datetime.now().isoformat(),
                "baseline_mb": self.baseline_memory,
                "current_mb": current_memory,
                "increase_mb": increase,
                "threshold_mb": self.threshold_mb
            }

            self.leak_warnings.append(warning)
            logger.warning(f"메모리 누수 가능성 감지: {increase:.2f}MB 증가")

            return warning

        return None

    def _get_memory_usage(self) -> float:
        """현재 프로세스 메모리 사용량 (MB)"""
        process = psutil.Process(os.getpid())
        return process.memory_info().rss / 1024 / 1024


class SelfHealingMonitor:
    """자가 치유 모니터링 시스템"""

    def __init__(self, project_root: str = None):
        self.project_root = Path(project_root or PROJECT_ROOT)
        self.knowledge_dir = self.project_root / "knowledge"
        self.system_state_file = self.knowledge_dir / "system_state.json"
        self.health_log_file = self.knowledge_dir / "health_log.json"

        # 모니터링 컴포넌트
        self.circular_tracker = CircularDependencyTracker()
        self.memory_detector = MemoryLeakDetector()
        self.memory_manager = MemoryManager(str(self.project_root))
        self.agent_hub = get_hub(str(self.project_root))

        # 시스템 상태
        self.system_health = {
            "status": "HEALTHY",
            "last_check": None,
            "issues": [],
            "auto_fixes": [],
            "uptime_seconds": 0
        }

        # 모니터링 설정
        self.monitoring_enabled = True
        self.auto_heal_enabled = True
        self.check_interval = 30  # 초

        # 스레드
        self.monitor_thread = None
        self.start_time = datetime.now()

    def start_monitoring(self):
        """모니터링 시작"""
        self.monitoring_enabled = True
        self.monitor_thread = threading.Thread(
            target=self._monitor_loop,
            daemon=True
        )
        self.monitor_thread.start()
        logger.info("자가 치유 모니터링 시작")

    def stop_monitoring(self):
        """모니터링 중지"""
        self.monitoring_enabled = False
        if self.monitor_thread:
            self.monitor_thread.join(timeout=5)
        logger.info("자가 치유 모니터링 중지")

    def _monitor_loop(self):
        """모니터링 루프"""
        while self.monitoring_enabled:
            try:
                self._perform_health_check()
                time.sleep(self.check_interval)
            except Exception as e:
                logger.error(f"모니터링 오류: {e}")
                time.sleep(5)

    def _perform_health_check(self):
        """시스템 건강도 체크"""
        issues = []
        auto_fixes = []

        # 1. 순환 참조 체크
        cycle_report = self.circular_tracker.get_cycle_report()
        if cycle_report["total_cycles_detected"] > 0:
            issues.append({
                "type": "CIRCULAR_DEPENDENCY",
                "severity": "WARNING",
                "details": cycle_report
            })

        # 2. 메모리 누수 체크
        memory_warning = self.memory_detector.check_memory()
        if memory_warning:
            issues.append({
                "type": "MEMORY_LEAK",
                "severity": "WARNING",
                "details": memory_warning
            })

            # 자동 치유: 캐시 정리
            if self.auto_heal_enabled:
                self._clear_caches()
                auto_fixes.append({
                    "issue": "MEMORY_LEAK",
                    "action": "CACHE_CLEARED",
                    "timestamp": datetime.now().isoformat()
                })

        # 3. 에이전트 상태 체크
        agent_status = self._check_agent_health()
        if agent_status["inactive_agents"]:
            issues.append({
                "type": "INACTIVE_AGENTS",
                "severity": "INFO",
                "details": agent_status
            })

        # 4. 파일 시스템 체크
        fs_status = self._check_filesystem_health()
        if fs_status["issues"]:
            issues.extend(fs_status["issues"])

        # 5. 시스템 상태 업데이트
        self.system_health = {
            "status": "HEALTHY" if not issues else "DEGRADED",
            "last_check": datetime.now().isoformat(),
            "issues": issues,
            "auto_fixes": auto_fixes,
            "uptime_seconds": (datetime.now() - self.start_time).total_seconds()
        }

        # 상태 저장
        self._save_system_state()

        # 건강도 로그 기록
        self._log_health_status()

        # 중대한 문제 발견 시 알림
        if any(issue["severity"] == "CRITICAL" for issue in issues):
            self._send_critical_alert(issues)

    def _check_agent_health(self) -> Dict:
        """에이전트 건강도 체크"""
        hub_status = self.agent_hub.get_hub_status()
        inactive_agents = []

        for agent_key, agent_info in self.agent_hub.agents.items():
            if not agent_info["active"]:
                inactive_agents.append(agent_key)

        return {
            "total_agents": len(self.agent_hub.agents),
            "active_agents": hub_status["active_agents"],
            "inactive_agents": inactive_agents,
            "pending_messages": hub_status["pending_messages"]
        }

    def _check_filesystem_health(self) -> Dict:
        """파일 시스템 건강도 체크"""
        issues = []

        # 중요 디렉토리 존재 확인
        critical_dirs = [
            self.project_root / "execution",
            self.project_root / "libs",
            self.project_root / "knowledge",
            self.project_root / "directives"
        ]

        for dir_path in critical_dirs:
            if not dir_path.exists():
                issues.append({
                    "type": "MISSING_DIRECTORY",
                    "severity": "CRITICAL",
                    "path": str(dir_path)
                })

                # 자동 치유: 디렉토리 생성
                if self.auto_heal_enabled:
                    dir_path.mkdir(parents=True, exist_ok=True)
                    logger.info(f"자동 복구: {dir_path} 디렉토리 생성")

        # 디스크 공간 체크
        disk_usage = psutil.disk_usage(str(self.project_root))
        if disk_usage.percent > 90:
            issues.append({
                "type": "LOW_DISK_SPACE",
                "severity": "WARNING",
                "usage_percent": disk_usage.percent
            })

        return {"issues": issues}

    def _clear_caches(self):
        """캐시 정리"""
        try:
            # 메모리 매니저 캐시 정리
            self.memory_manager.cache.clear()

            # 임시 파일 정리
            tmp_dir = self.project_root / ".tmp"
            if tmp_dir.exists():
                for file in tmp_dir.glob("*.cache"):
                    if (datetime.now() - datetime.fromtimestamp(file.stat().st_mtime)) > timedelta(hours=24):
                        file.unlink()

            logger.info("캐시 정리 완료")

        except Exception as e:
            logger.error(f"캐시 정리 실패: {e}")

    def _save_system_state(self):
        """시스템 상태 저장"""
        try:
            # 기존 상태 파일 읽기
            if self.system_state_file.exists():
                with open(self.system_state_file, 'r', encoding='utf-8') as f:
                    state = json.load(f)
            else:
                state = {}

            # 자가 치유 상태 추가
            state["self_healing"] = {
                "health": self.system_health,
                "circular_dependencies": self.circular_tracker.get_cycle_report(),
                "memory_status": {
                    "samples": self.memory_detector.memory_samples[-5:],
                    "warnings": self.memory_detector.leak_warnings[-5:]
                }
            }

            # 파일 저장
            with open(self.system_state_file, 'w', encoding='utf-8') as f:
                json.dump(state, f, indent=2, ensure_ascii=False)

        except Exception as e:
            logger.error(f"시스템 상태 저장 실패: {e}")

    def _log_health_status(self):
        """건강도 로그 기록"""
        try:
            # 로그 파일 읽기 또는 생성
            if self.health_log_file.exists():
                with open(self.health_log_file, 'r', encoding='utf-8') as f:
                    logs = json.load(f)
            else:
                logs = []

            # 새 로그 추가
            log_entry = {
                "timestamp": datetime.now().isoformat(),
                "status": self.system_health["status"],
                "issues_count": len(self.system_health["issues"]),
                "auto_fixes_count": len(self.system_health["auto_fixes"]),
                "uptime_seconds": self.system_health["uptime_seconds"]
            }

            logs.append(log_entry)

            # 최근 100개만 유지
            if len(logs) > 100:
                logs = logs[-100:]

            # 파일 저장
            with open(self.health_log_file, 'w', encoding='utf-8') as f:
                json.dump(logs, f, indent=2, ensure_ascii=False)

        except Exception as e:
            logger.error(f"건강도 로그 기록 실패: {e}")

    def _send_critical_alert(self, issues: List[Dict]):
        """중대한 문제 알림"""
        critical_issues = [i for i in issues if i["severity"] == "CRITICAL"]

        alert_message = {
            "type": "CRITICAL_ALERT",
            "timestamp": datetime.now().isoformat(),
            "issues": critical_issues,
            "action_required": "즉시 확인 필요"
        }

        # 에이전트 허브를 통해 브로드캐스트
        self.agent_hub.broadcast_message(
            from_agent="MONITOR",
            message_type="ALERT",
            data=alert_message
        )

        logger.critical(f"중대한 시스템 문제 감지: {len(critical_issues)}개")

    def check_circular_dependency(self, context: str) -> bool:
        """순환 참조 체크 (외부 호출용)"""
        return self.circular_tracker.enter_context(context)

    def release_context(self, context: str):
        """컨텍스트 해제 (외부 호출용)"""
        self.circular_tracker.exit_context(context)

    def get_status(self) -> Dict:
        """현재 시스템 상태 반환"""
        return {
            "health": self.system_health,
            "monitoring_enabled": self.monitoring_enabled,
            "auto_heal_enabled": self.auto_heal_enabled,
            "uptime": str(datetime.now() - self.start_time)
        }


def main():
    """테스트 및 데몬 실행"""
    monitor = SelfHealingMonitor()

    print("자가 치유 모니터링 시스템 시작")
    print("=" * 60)

    # 시그널 핸들러
    def signal_handler(signum, frame):
        print("\n모니터링 종료 중...")
        monitor.stop_monitoring()
        sys.exit(0)

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    # 모니터링 시작
    monitor.start_monitoring()

    # 메인 루프
    try:
        while True:
            status = monitor.get_status()
            print(f"\r[{datetime.now().strftime('%H:%M:%S')}] "
                  f"상태: {status['health']['status']} | "
                  f"문제: {len(status['health']['issues'])} | "
                  f"가동시간: {status['uptime']}", end="")
            time.sleep(5)

    except KeyboardInterrupt:
        print("\n종료 중...")
        monitor.stop_monitoring()


if __name__ == "__main__":
    main()