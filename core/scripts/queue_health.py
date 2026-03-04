#!/usr/bin/env python3
"""
Queue Health — pending/processing 상태 요약
"""

import json
from collections import Counter, defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent.parent
TASK_QUEUE = ROOT / ".infra/queue/tasks"
COUNCIL_QUEUE = ROOT / ".infra/queue/council"


def load_files(base: Path, subdir: str):
    path = base / subdir
    if not path.exists():
        return []
    files = []
    for f in path.glob("*.json"):
        try:
            data = json.loads(f.read_text())
            files.append((f.name, data))
        except Exception:
            files.append((f.name, {"error": "parse_error"}))
    return files


def summarize(name: str, items):
    print(f"[{name}] {len(items)}")
    if not items:
        return
    by_agent = Counter(d.get("agent_type") for _, d in items if d.get("agent_type"))
    by_task = Counter(d.get("task_type") for _, d in items if d.get("task_type"))
    if by_agent:
        print("  agents:", dict(by_agent))
    if by_task:
        print("  tasks :", dict(by_task))
    for fname, data in items[:5]:
        print("  ", fname, data.get("agent_type"), data.get("task_type"), data.get("status"))


def main():
    # SA/CE/AD 등 일반 태스크
    pending = load_files(TASK_QUEUE, "pending")
    processing = load_files(TASK_QUEUE, "processing")
    summarize("tasks/pending", pending)
    summarize("tasks/processing", processing)

    # Plan Council 큐
    council_pending = load_files(COUNCIL_QUEUE, "pending")
    council_processing = load_files(COUNCIL_QUEUE, "processing")
    summarize("council/pending", council_pending)
    summarize("council/processing", council_processing)


if __name__ == "__main__":
    main()
