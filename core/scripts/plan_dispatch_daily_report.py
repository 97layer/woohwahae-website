#!/usr/bin/env python3
"""
plan_dispatch_daily_report.py

Generate a daily stability report for plan_dispatch classifier/dispatcher health.
"""

from __future__ import annotations

import argparse
import json
import os
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict

PROJECT_ROOT = Path(__file__).resolve().parents[2]
DEFAULT_OUTPUT_DIR = PROJECT_ROOT / "knowledge" / "system" / "plan_dispatch_daily"
if str(PROJECT_ROOT) not in sys.path:
    sys.path.insert(0, str(PROJECT_ROOT))

from core.system.plan_dispatch_metrics import (  # noqa: E402
    DEFAULT_LOG_PATH as DEFAULT_METRICS_LOG_PATH,
    summarize_metrics,
)
from core.system.plan_dispatch_replay import (  # noqa: E402
    DEFAULT_REPORT_PATH as DEFAULT_SOURCE_PATH,
    DEFAULT_REPLAY_LOG_PATH,
    append_replay_summary,
    backfill_replay_history,
    build_summary,
    detect_drift,
    load_recent_tasks,
    replay_tasks,
)


def _now_utc() -> datetime:
    return datetime.now(timezone.utc)


def _date_key_utc(ts: datetime) -> str:
    return ts.strftime("%Y%m%d")


def _resolve_output_path(date_key: str, output: str | None, output_dir: str) -> Path:
    if output:
        path = Path(output)
        if not path.is_absolute():
            path = PROJECT_ROOT / path
        return path
    path = Path(output_dir)
    if not path.is_absolute():
        path = PROJECT_ROOT / path
    return path / f"plan_dispatch_{date_key}.json"


def _evaluate_health(
    *,
    metrics: Dict[str, Any],
    replay: Dict[str, Any],
    fallback_threshold: float,
    fallback_min_samples: int,
) -> Dict[str, Any]:
    issues = []
    warnings = []

    metric_total = int(metrics.get("total", 0))
    fallback_rate = float(metrics.get("fallback_rate", 0.0))
    if metric_total < fallback_min_samples:
        warnings.append(
            f"insufficient metrics samples total={metric_total} (<{fallback_min_samples})"
        )
    elif fallback_rate >= fallback_threshold:
        issues.append(
            f"fallback_rate={fallback_rate:.3f} >= threshold={fallback_threshold:.3f}"
        )

    replay_total = int(replay.get("total", 0))
    if replay_total <= 0:
        warnings.append("no replay tasks available")

    drift = replay.get("drift", {})
    if isinstance(drift, dict) and drift.get("detected") is True:
        issues.append(
            "drift detected "
            f"allowed_delta={drift.get('allowed_delta')} "
            f"max_complexity_delta={drift.get('max_complexity_delta')}"
        )

    status = "pass"
    if issues:
        status = "fail"
    elif warnings:
        status = "warn"

    return {
        "status": status,
        "issues": issues,
        "warnings": warnings,
    }


def generate_report(
    *,
    source: Path,
    metrics_log: Path,
    replay_log: Path,
    metrics_window: int,
    replay_limit: int,
    min_complexity: str,
    drift_window: int,
    drift_min_samples: int,
    allowed_rate_drift_threshold: float,
    complexity_rate_drift_threshold: float,
    fallback_threshold: float,
    fallback_min_samples: int,
    backfill_report: bool,
    backfill_limit: int,
    backfill_snapshots: int,
    append_replay: bool,
) -> Dict[str, Any]:
    metrics = summarize_metrics(log_path=metrics_log, window=max(1, metrics_window))

    backfill = None
    if backfill_report:
        backfill = backfill_replay_history(
            source=source,
            report_log=replay_log,
            min_complexity=min_complexity,
            limit=max(1, backfill_limit),
            snapshots=max(1, backfill_snapshots),
        )

    tasks = load_recent_tasks(source, limit=max(1, replay_limit))
    rows = replay_tasks(tasks, min_complexity=min_complexity)
    replay = build_summary(rows, min_complexity=min_complexity, source=source)
    replay["drift"] = detect_drift(
        replay,
        replay_log,
        window=max(1, drift_window),
        min_samples=max(1, drift_min_samples),
        allowed_rate_threshold=float(allowed_rate_drift_threshold),
        complexity_rate_threshold=float(complexity_rate_drift_threshold),
    )
    if append_replay:
        replay["appended"] = append_replay_summary(replay_log, replay)
    if backfill is not None:
        replay["backfill"] = backfill
    replay.pop("rows", None)

    health = _evaluate_health(
        metrics=metrics,
        replay=replay,
        fallback_threshold=float(fallback_threshold),
        fallback_min_samples=max(1, fallback_min_samples),
    )

    now = _now_utc()
    return {
        "generated_at": now.isoformat(),
        "date_utc": _date_key_utc(now),
        "type": "plan_dispatch_daily",
        "sources": {
            "plan_council_reports": str(source),
            "metrics_log": str(metrics_log),
            "replay_log": str(replay_log),
        },
        "thresholds": {
            "fallback_threshold": float(fallback_threshold),
            "fallback_min_samples": int(fallback_min_samples),
            "drift_window": int(drift_window),
            "drift_min_samples": int(drift_min_samples),
            "allowed_rate_drift_threshold": float(allowed_rate_drift_threshold),
            "complexity_rate_drift_threshold": float(complexity_rate_drift_threshold),
        },
        "metrics": metrics,
        "replay": replay,
        "health": health,
    }


def main() -> int:
    parser = argparse.ArgumentParser(description="Generate plan_dispatch daily stability report")
    parser.add_argument("--source", default=str(DEFAULT_SOURCE_PATH), help="plan_council_reports jsonl")
    parser.add_argument(
        "--metrics-log",
        default=str(DEFAULT_METRICS_LOG_PATH),
        help="plan_dispatch_metrics jsonl",
    )
    parser.add_argument(
        "--replay-log",
        default=str(DEFAULT_REPLAY_LOG_PATH),
        help="plan_dispatch replay report jsonl",
    )
    parser.add_argument("--output", help="explicit output JSON path")
    parser.add_argument("--output-dir", default=str(DEFAULT_OUTPUT_DIR), help="default output directory")
    parser.add_argument("--write", action="store_true", help="write JSON report file")
    parser.add_argument("--json", action="store_true", help="print JSON payload")
    parser.add_argument(
        "--fail-on-health",
        action="store_true",
        help="return non-zero when health.status is fail",
    )

    parser.add_argument("--metrics-window", type=int, default=200, help="metrics summary window")
    parser.add_argument("--replay-limit", type=int, default=30, help="max tasks for replay")
    parser.add_argument("--min-complexity", default="medium", help="simple|medium|high")
    parser.add_argument("--drift-window", type=int, default=14, help="drift history window")
    parser.add_argument("--drift-min-samples", type=int, default=5, help="drift minimum history samples")
    parser.add_argument(
        "--allowed-rate-drift-threshold",
        type=float,
        default=0.20,
        help="absolute allowed-rate drift threshold",
    )
    parser.add_argument(
        "--complexity-rate-drift-threshold",
        type=float,
        default=0.25,
        help="absolute complexity-rate drift threshold",
    )

    parser.add_argument(
        "--fallback-threshold",
        type=float,
        default=float(os.getenv("PLAN_DISPATCH_FALLBACK_WARN_RATE", "0.30")),
        help="fallback_rate fail threshold",
    )
    parser.add_argument(
        "--fallback-min-samples",
        type=int,
        default=int(os.getenv("PLAN_DISPATCH_FALLBACK_MIN_SAMPLES", "20")),
        help="minimum metrics samples before fallback threshold applies",
    )

    parser.add_argument("--backfill-report", action="store_true", help="backfill replay history before summary")
    parser.add_argument("--backfill-limit", type=int, default=120, help="max tasks for replay backfill")
    parser.add_argument("--backfill-snapshots", type=int, default=12, help="number of backfill snapshots")
    parser.add_argument("--append-replay", action="store_true", help="append current replay summary to log")
    args = parser.parse_args()

    report = generate_report(
        source=Path(args.source),
        metrics_log=Path(args.metrics_log),
        replay_log=Path(args.replay_log),
        metrics_window=args.metrics_window,
        replay_limit=args.replay_limit,
        min_complexity=args.min_complexity,
        drift_window=args.drift_window,
        drift_min_samples=args.drift_min_samples,
        allowed_rate_drift_threshold=args.allowed_rate_drift_threshold,
        complexity_rate_drift_threshold=args.complexity_rate_drift_threshold,
        fallback_threshold=args.fallback_threshold,
        fallback_min_samples=args.fallback_min_samples,
        backfill_report=args.backfill_report,
        backfill_limit=args.backfill_limit,
        backfill_snapshots=args.backfill_snapshots,
        append_replay=args.append_replay,
    )

    output_path = _resolve_output_path(report["date_utc"], args.output, args.output_dir)
    report["report_path"] = str(output_path)

    if args.write:
        output_path.parent.mkdir(parents=True, exist_ok=True)
        output_path.write_text(
            json.dumps(report, ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )

    if args.json:
        print(json.dumps(report, ensure_ascii=False))
    else:
        print(
            f"plan_dispatch_daily_report status={report['health']['status']} "
            f"path={output_path}"
        )
    if args.fail_on_health and str(report.get("health", {}).get("status", "")).lower() == "fail":
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
