#!/usr/bin/env python3
"""
File Watcher — 파일 변경 감지 → Cascade Manager 자동 실행

watchdog 기반 실시간 파일 감지
dependency_graph.json에 등록된 파일만 추적

Author: LAYER OS
Created: 2026-02-26
"""

import sys
import time
import json
from pathlib import Path
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler

# cascade_manager import
sys.path.insert(0, str(Path(__file__).parent.parent))
from system.cascade_manager import CascadeManager


class DependencyGraphHandler(FileSystemEventHandler):
    """Dependency Graph 파일 변경 감지"""

    def __init__(self, graph_path: Path, debounce_seconds: float = 1.0):
        self.graph_path = graph_path
        self.manager = CascadeManager(str(graph_path))
        self.debounce_seconds = debounce_seconds
        self.last_modified = {}

        # 추적 대상 파일 목록
        self.tracked_files = set(self.manager.graph['nodes'].keys())
        print(f"📡 Watching {len(self.tracked_files)} files:")
        for f in sorted(self.tracked_files):
            print(f"   - {f}")

    def on_modified(self, event):
        """파일 수정 감지"""
        if event.is_directory:
            return

        # 절대경로 → 상대경로
        filepath = Path(event.src_path)
        try:
            rel_path = str(filepath.relative_to(self.manager.project_root))
        except ValueError:
            return  # 프로젝트 외부 파일

        # 추적 대상 여부
        if rel_path not in self.tracked_files:
            return

        # Debounce (연속 수정 방지)
        now = time.time()
        if rel_path in self.last_modified:
            if now - self.last_modified[rel_path] < self.debounce_seconds:
                return

        self.last_modified[rel_path] = now

        # Cascade Manager 실행
        print(f"\n🔔 File changed: {rel_path}")
        try:
            report = self.manager.on_file_change(str(filepath))
            print(f"   Tier: {report.tier}")
            print(f"   Affected: {len(report.affected_nodes)} nodes")
            print(f"   Actions: {', '.join(report.cascade_actions)}")
        except Exception as e:
            print(f"   ❌ Error: {e}")


def main():
    """File Watcher 실행"""
    project_root = Path(__file__).parent.parent.parent
    graph_path = project_root / 'knowledge/system/dependency_graph.json'

    if not graph_path.exists():
        print(f"❌ Dependency graph not found: {graph_path}")
        sys.exit(1)

    print("🚀 File Watcher Starting...")
    print(f"   Project: {project_root}")
    print(f"   Graph: {graph_path}")
    print()

    # Handler 생성
    event_handler = DependencyGraphHandler(graph_path)

    # Observer 시작
    observer = Observer()
    observer.schedule(event_handler, str(project_root), recursive=True)
    observer.start()

    print("✅ File Watcher Running (Ctrl+C to stop)")
    print()

    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        print("\n🛑 Stopping File Watcher...")
        observer.stop()

    observer.join()
    print("✅ File Watcher Stopped")


if __name__ == "__main__":
    main()
