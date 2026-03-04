#!/usr/bin/env python3
"""
progress_graph.py

Build compact progress trends (sparkline) from system logs.
"""

from __future__ import annotations

import argparse
import json
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict, List

PROJECT_ROOT = Path(__file__).resolve().parents[2]
DOCTOR_REPORT_PATH = PROJECT_ROOT / "knowledge" / "system" / "harness_doctor_reports.jsonl"
PLAN_METRICS_PATH = PROJECT_ROOT / "knowledge" / "system" / "plan_dispatch_metrics.jsonl"

_SPARK_CHARS = "▁▂▃▄▅▆▇█"


def _load_jsonl(path: Path) -> List[Dict[str, Any]]:
    if not path.exists():
        return []
    rows: List[Dict[str, Any]] = []
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line:
            continue
        try:
            obj = json.loads(line)
        except Exception:
            continue
        if isinstance(obj, dict):
            rows.append(obj)
    return rows


def _resample(values: List[float], width: int) -> List[float]:
    if not values:
        return []
    if width <= 0:
        return []
    if len(values) <= width:
        return values
    sampled: List[float] = []
    n = len(values)
    for i in range(width):
        idx = int(round(i * (n - 1) / max(1, width - 1)))
        sampled.append(values[idx])
    return sampled


def sparkline(values: List[float], width: int = 24) -> str:
    data = _resample(list(values), width=max(1, width))
    if not data:
        return "·"
    min_v = min(data)
    max_v = max(data)
    if max_v == min_v:
        return _SPARK_CHARS[0] * len(data)
    out = []
    for v in data:
        norm = (v - min_v) / (max_v - min_v)
        idx = int(round(norm * (len(_SPARK_CHARS) - 1)))
        idx = max(0, min(len(_SPARK_CHARS) - 1, idx))
        out.append(_SPARK_CHARS[idx])
    return "".join(out)


def _safe_float(value: Any, default: float = 0.0) -> float:
    try:
        return float(value)
    except Exception:
        return default


def _metric_series_from_events(events: List[Dict[str, Any]], key: str, *, limit: int) -> List[float]:
    rows = events[-max(1, limit):]
    values: List[float] = []
    for row in rows:
        if key == "fallback_rate":
            values.append(1.0 if bool(row.get("fallback")) else 0.0)
            continue
        if key == "blocked_rate":
            values.append(1.0 if str(row.get("phase", "")) == "blocked" else 0.0)
            continue
        if key == "executed_rate":
            values.append(1.0 if bool(row.get("executed")) else 0.0)
            continue
    return values


def _rolling_rate(values: List[float], window: int = 10) -> List[float]:
    if not values:
        return []
    out: List[float] = []
    for i in range(len(values)):
        start = max(0, i - window + 1)
        chunk = values[start : i + 1]
        out.append(sum(chunk) / len(chunk))
    return out


def _latest_and_delta(values: List[float]) -> Dict[str, float]:
    if not values:
        return {"latest": 0.0, "delta": 0.0}
    latest = float(values[-1])
    prev = float(values[-2]) if len(values) >= 2 else latest
    return {"latest": round(latest, 3), "delta": round(latest - prev, 3)}


def build_progress_payload(
    *,
    doctor_report_path: Path = DOCTOR_REPORT_PATH,
    plan_metrics_path: Path = PLAN_METRICS_PATH,
    limit: int = 30,
    graph_width: int = 24,
) -> Dict[str, Any]:
    doctor_rows = _load_jsonl(doctor_report_path)
    metric_rows = _load_jsonl(plan_metrics_path)

    score_values = [_safe_float(row.get("score", 0.0)) for row in doctor_rows[-max(1, limit):]]
    fallback_events = _metric_series_from_events(metric_rows, "fallback_rate", limit=max(1, limit))
    blocked_events = _metric_series_from_events(metric_rows, "blocked_rate", limit=max(1, limit))

    fallback_rates = _rolling_rate(fallback_events, window=10)
    blocked_rates = _rolling_rate(blocked_events, window=10)

    score_meta = _latest_and_delta(score_values)
    fallback_meta = _latest_and_delta(fallback_rates)
    blocked_meta = _latest_and_delta(blocked_rates)

    now = datetime.now(timezone.utc).isoformat()
    return {
        "generated_at": now,
        "window": limit,
        "graphs": {
            "score": sparkline(score_values, width=graph_width),
            "fallback_rate": sparkline(fallback_rates, width=graph_width),
            "blocked_rate": sparkline(blocked_rates, width=graph_width),
        },
        "metrics": {
            "score": score_meta,
            "fallback_rate": fallback_meta,
            "blocked_rate": blocked_meta,
        },
        "counts": {
            "doctor_reports": len(doctor_rows),
            "plan_metric_events": len(metric_rows),
        },
    }


def render_text(payload: Dict[str, Any]) -> str:
    g = payload.get("graphs", {})
    m = payload.get("metrics", {})
    w = payload.get("window", 0)

    def _line(label: str, key: str, fmt: str = "{latest:.3f}", pct: bool = False) -> str:
        metric = m.get(key, {})
        latest = float(metric.get("latest", 0.0))
        delta = float(metric.get("delta", 0.0))
        if pct:
            latest_txt = f"{latest*100:.1f}%"
            delta_txt = f"{delta*100:+.1f}%"
        else:
            latest_txt = fmt.format(latest=latest)
            delta_txt = f"{delta:+.3f}"
        graph = g.get(key, "·")
        return f"{label:<14} {graph}  latest={latest_txt}  delta={delta_txt}"

    lines = [
        f"=== Progress Trend (last {w}) ===",
        _line("Doctor Score", "score", fmt="{latest:.1f}", pct=False),
        _line("Fallback Rate", "fallback_rate", pct=True),
        _line("Blocked Rate", "blocked_rate", pct=True),
    ]
    return "\n".join(lines)


def main() -> int:
    parser = argparse.ArgumentParser(description="Generate compact progress trend graphs")
    parser.add_argument("--doctor-report", default=str(DOCTOR_REPORT_PATH), help="doctor jsonl path")
    parser.add_argument("--plan-metrics", default=str(PLAN_METRICS_PATH), help="plan metrics jsonl path")
    parser.add_argument("--limit", type=int, default=30, help="history window")
    parser.add_argument("--width", type=int, default=24, help="graph width")
    parser.add_argument("--json", action="store_true", help="print as JSON")
    args = parser.parse_args()

    payload = build_progress_payload(
        doctor_report_path=Path(args.doctor_report),
        plan_metrics_path=Path(args.plan_metrics),
        limit=max(1, args.limit),
        graph_width=max(8, args.width),
    )
    if args.json:
        print(json.dumps(payload, ensure_ascii=False))
    else:
        print(render_text(payload))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
