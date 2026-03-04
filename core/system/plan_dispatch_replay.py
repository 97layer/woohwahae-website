#!/usr/bin/env python3
"""
plan_dispatch_replay.py

Replay recent plan_council task texts through plan_dispatch classifier.
"""

from __future__ import annotations

import argparse
import json
import sys
from collections import Counter
from datetime import datetime, timezone
from statistics import mean
from pathlib import Path
from typing import Any, Dict, List

PROJECT_ROOT = Path(__file__).resolve().parents[2]
if str(PROJECT_ROOT) not in sys.path:
    sys.path.insert(0, str(PROJECT_ROOT))

from core.system.plan_dispatch_classifier import classify_task

DEFAULT_REPORT_PATH = PROJECT_ROOT / "knowledge" / "system" / "plan_council_reports.jsonl"
DEFAULT_REPLAY_LOG_PATH = PROJECT_ROOT / "knowledge" / "system" / "plan_dispatch_replay_reports.jsonl"


def load_recent_tasks(report_path: Path, limit: int = 50) -> List[str]:
    if not report_path.exists():
        return []

    lines = report_path.read_text(encoding="utf-8").splitlines()
    tasks_rev: List[str] = []
    seen = set()

    for raw in reversed(lines):
        line = raw.strip()
        if not line:
            continue
        try:
            payload = json.loads(line)
        except Exception:
            continue
        if not isinstance(payload, dict):
            continue
        task = str(payload.get("task", "")).strip()
        if not task or task in seen:
            continue
        seen.add(task)
        tasks_rev.append(task)
        if len(tasks_rev) >= max(1, limit):
            break

    return list(reversed(tasks_rev))


def replay_tasks(tasks: List[str], min_complexity: str = "medium") -> List[Dict[str, Any]]:
    rows: List[Dict[str, Any]] = []
    for idx, task in enumerate(tasks, start=1):
        classified = classify_task(task, min_complexity=min_complexity)
        rows.append(
            {
                "index": idx,
                "task": task,
                "complexity": classified.get("complexity"),
                "score": classified.get("score"),
                "allowed": bool(classified.get("allowed")),
                "signals": classified.get("signals", []),
            }
        )
    return rows


def build_summary(rows: List[Dict[str, Any]], min_complexity: str, source: Path) -> Dict[str, Any]:
    complexity_counts = Counter()
    allowed_count = 0
    signal_counts = Counter()

    for row in rows:
        complexity_counts[str(row.get("complexity", "unknown"))] += 1
        if bool(row.get("allowed")):
            allowed_count += 1
        for sig in row.get("signals", []) or []:
            signal_counts[str(sig)] += 1

    total = len(rows)
    allowed_rate = round(allowed_count / total, 3) if total else 0.0
    top_signals = [sig for sig, _ in signal_counts.most_common(8)]

    return {
        "source": str(source),
        "min_complexity": min_complexity,
        "total": total,
        "allowed_count": allowed_count,
        "allowed_rate": allowed_rate,
        "complexity_counts": dict(complexity_counts),
        "top_signals": top_signals,
        "rows": rows,
    }


def append_replay_summary(report_log: Path, summary: Dict[str, Any]) -> Dict[str, Any]:
    payload = {
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "source": summary.get("source", ""),
        "min_complexity": summary.get("min_complexity", "medium"),
        "total": int(summary.get("total", 0)),
        "allowed_count": int(summary.get("allowed_count", 0)),
        "allowed_rate": float(summary.get("allowed_rate", 0.0)),
        "complexity_counts": summary.get("complexity_counts", {}),
        "top_signals": summary.get("top_signals", []),
    }
    report_log.parent.mkdir(parents=True, exist_ok=True)
    with report_log.open("a", encoding="utf-8") as fp:
        fp.write(json.dumps(payload, ensure_ascii=False) + "\n")
    return payload


def _load_replay_history(report_log: Path, window: int) -> List[Dict[str, Any]]:
    if not report_log.exists():
        return []
    rows: List[Dict[str, Any]] = []
    for raw in report_log.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line:
            continue
        try:
            payload = json.loads(line)
        except Exception:
            continue
        if isinstance(payload, dict):
            rows.append(payload)
    if window > 0:
        rows = rows[-window:]
    return rows


def detect_drift(
    summary: Dict[str, Any],
    report_log: Path,
    *,
    window: int = 14,
    min_samples: int = 5,
    allowed_rate_threshold: float = 0.20,
    complexity_rate_threshold: float = 0.25,
) -> Dict[str, Any]:
    history = _load_replay_history(report_log, window=window)
    baseline_samples = len(history)

    if baseline_samples < min_samples:
        return {
            "detected": False,
            "reason": "insufficient_history",
            "baseline_samples": baseline_samples,
        }

    allowed_rates = []
    complexity_rates: Dict[str, List[float]] = {"simple": [], "medium": [], "high": []}

    for entry in history:
        total = int(entry.get("total", 0))
        if total <= 0:
            continue
        allowed_rates.append(float(entry.get("allowed_rate", 0.0)))
        raw_counts = entry.get("complexity_counts", {})
        if not isinstance(raw_counts, dict):
            raw_counts = {}
        for key in ("simple", "medium", "high"):
            value = raw_counts.get(key, 0)
            try:
                rate = float(value) / float(total)
            except Exception:
                rate = 0.0
            complexity_rates[key].append(rate)

    if not allowed_rates:
        return {
            "detected": False,
            "reason": "insufficient_valid_history",
            "baseline_samples": baseline_samples,
        }

    baseline_allowed_rate = float(mean(allowed_rates))
    current_allowed_rate = float(summary.get("allowed_rate", 0.0))
    allowed_delta = abs(current_allowed_rate - baseline_allowed_rate)

    current_total = int(summary.get("total", 0))
    current_counts = summary.get("complexity_counts", {})
    if not isinstance(current_counts, dict):
        current_counts = {}

    complexity_delta = {}
    max_complexity_delta = 0.0
    for key in ("simple", "medium", "high"):
        baseline_rate = float(mean(complexity_rates[key])) if complexity_rates[key] else 0.0
        current_count = current_counts.get(key, 0)
        try:
            current_rate = float(current_count) / float(current_total) if current_total > 0 else 0.0
        except Exception:
            current_rate = 0.0
        delta = abs(current_rate - baseline_rate)
        complexity_delta[key] = round(delta, 3)
        if delta > max_complexity_delta:
            max_complexity_delta = delta

    detected = bool(
        allowed_delta >= float(allowed_rate_threshold)
        or max_complexity_delta >= float(complexity_rate_threshold)
    )

    return {
        "detected": detected,
        "reason": "threshold_exceeded" if detected else "within_threshold",
        "baseline_samples": baseline_samples,
        "baseline_allowed_rate": round(baseline_allowed_rate, 3),
        "current_allowed_rate": round(current_allowed_rate, 3),
        "allowed_delta": round(allowed_delta, 3),
        "allowed_rate_threshold": float(allowed_rate_threshold),
        "max_complexity_delta": round(max_complexity_delta, 3),
        "complexity_rate_threshold": float(complexity_rate_threshold),
        "complexity_delta": complexity_delta,
    }


def backfill_replay_history(
    *,
    source: Path,
    report_log: Path,
    min_complexity: str = "medium",
    limit: int = 120,
    snapshots: int = 12,
) -> Dict[str, Any]:
    tasks = load_recent_tasks(source, limit=max(1, limit))
    if not tasks:
        return {"appended": 0, "reason": "no_tasks"}

    tasks_len = len(tasks)
    snapshots = max(1, min(int(snapshots), tasks_len))
    min_window = min(5, tasks_len)
    if snapshots == 1:
        indices = [tasks_len]
    else:
        span = max(1, tasks_len - min_window)
        indices = []
        for i in range(snapshots):
            ratio = i / (snapshots - 1)
            idx = min_window + int(round(span * ratio))
            idx = max(min_window, min(tasks_len, idx))
            if not indices or idx != indices[-1]:
                indices.append(idx)
        if indices[-1] != tasks_len:
            indices.append(tasks_len)

    appended = 0
    for idx in indices:
        subset = tasks[:idx]
        rows = replay_tasks(subset, min_complexity=min_complexity)
        summary = build_summary(rows, min_complexity=min_complexity, source=source)
        append_replay_summary(report_log, summary)
        appended += 1

    return {
        "appended": appended,
        "reason": "ok",
        "tasks_used": tasks_len,
        "indices": indices,
    }


def _print_human(summary: Dict[str, Any]) -> None:
    print("=== plan_dispatch replay ===")
    print(f"source: {summary.get('source')}")
    print(
        f"total={summary.get('total')} "
        f"allowed={summary.get('allowed_count')} "
        f"allowed_rate={summary.get('allowed_rate')}"
    )
    print(f"complexity_counts: {summary.get('complexity_counts')}")
    print(f"top_signals: {summary.get('top_signals')}")


def main() -> int:
    parser = argparse.ArgumentParser(description="Replay classifier against recent plan council tasks")
    parser.add_argument(
        "--source",
        default=str(DEFAULT_REPORT_PATH),
        help="plan_council_reports jsonl path",
    )
    parser.add_argument("--limit", type=int, default=50, help="max unique tasks to replay")
    parser.add_argument("--min-complexity", default="medium", help="simple|medium|high")
    parser.add_argument("--json", action="store_true", help="print JSON")
    parser.add_argument("--include-rows", action="store_true", help="include replay rows in output")
    parser.add_argument("--append-report", action="store_true", help="append summary to replay report jsonl")
    parser.add_argument("--backfill-report", action="store_true", help="backfill replay report history")
    parser.add_argument(
        "--report-log",
        default=str(DEFAULT_REPLAY_LOG_PATH),
        help="replay summary jsonl path",
    )
    parser.add_argument("--drift-window", type=int, default=14, help="history window for drift detection")
    parser.add_argument("--drift-min-samples", type=int, default=5, help="minimum samples for drift detection")
    parser.add_argument(
        "--allowed-rate-drift-threshold",
        type=float,
        default=0.20,
        help="absolute allowed_rate drift threshold",
    )
    parser.add_argument(
        "--complexity-rate-drift-threshold",
        type=float,
        default=0.25,
        help="absolute complexity rate drift threshold",
    )
    parser.add_argument("--backfill-limit", type=int, default=120, help="max source tasks for backfill")
    parser.add_argument("--backfill-snapshots", type=int, default=12, help="number of synthetic history snapshots")
    args = parser.parse_args()

    source = Path(args.source)
    report_log = Path(args.report_log)
    backfill = None
    if args.backfill_report:
        backfill = backfill_replay_history(
            source=source,
            report_log=report_log,
            min_complexity=args.min_complexity,
            limit=max(1, args.backfill_limit),
            snapshots=max(1, args.backfill_snapshots),
        )

    tasks = load_recent_tasks(source, limit=max(1, args.limit))
    rows = replay_tasks(tasks, min_complexity=args.min_complexity)
    summary = build_summary(rows, args.min_complexity, source)
    drift = detect_drift(
        summary,
        report_log,
        window=max(1, args.drift_window),
        min_samples=max(1, args.drift_min_samples),
        allowed_rate_threshold=float(args.allowed_rate_drift_threshold),
        complexity_rate_threshold=float(args.complexity_rate_drift_threshold),
    )
    summary["drift"] = drift
    if args.append_report:
        appended = append_replay_summary(report_log, summary)
        summary["report_log"] = str(report_log)
        summary["appended"] = appended
    if backfill is not None:
        summary["backfill"] = backfill
    if not args.include_rows:
        summary = {k: v for k, v in summary.items() if k != "rows"}

    if args.json:
        print(json.dumps(summary, ensure_ascii=False))
    else:
        _print_human(summary)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
