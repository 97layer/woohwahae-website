#!/usr/bin/env python3
"""
Queue Health — pending/processing 상태 요약
"""

import json
from collections import Counter, defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent.parent
QUEUE_ROOT = ROOT / ".infra/queue/tasks"


def load_files(subdir: str):
    path = QUEUE_ROOT / subdir
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
    by_agent = Counter(d.get("agent_type") for _, d in items)
    by_task = Counter(d.get("task_type") for _, d in items)
    if by_agent:
        print("  agents:", dict(by_agent))
    if by_task:
        print("  tasks :", dict(by_task))
    for fname, data in items[:5]:
        print("  ", fname, data.get("agent_type"), data.get("task_type"), data.get("status"))


def main():
    pending = load_files("pending")
    processing = load_files("processing")
    summarize("pending", pending)
    summarize("processing", processing)


if __name__ == "__main__":
    main()
