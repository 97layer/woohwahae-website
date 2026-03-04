#!/usr/bin/env python3
"""
progress_graph tests
"""

from __future__ import annotations

import json
from pathlib import Path

from core.system.progress_graph import build_progress_payload, sparkline


def test_sparkline_handles_empty_values():
    assert sparkline([], width=12) == "·"


def test_sparkline_respects_width():
    line = sparkline([1, 2, 3, 4, 5, 6], width=4)
    assert len(line) == 4


def test_build_progress_payload_from_temp_logs(tmp_path: Path):
    doctor = tmp_path / "doctor.jsonl"
    metrics = tmp_path / "metrics.jsonl"

    doctor_rows = [
        {"timestamp": "2026-03-04T00:00:00+00:00", "score": 80},
        {"timestamp": "2026-03-04T01:00:00+00:00", "score": 90},
        {"timestamp": "2026-03-04T02:00:00+00:00", "score": 100},
    ]
    metric_rows = [
        {"fallback": False, "phase": "smoke", "executed": True},
        {"fallback": True, "phase": "blocked", "executed": False},
        {"fallback": False, "phase": "execute", "executed": True},
    ]
    doctor.write_text(
        "\n".join(json.dumps(row, ensure_ascii=False) for row in doctor_rows) + "\n",
        encoding="utf-8",
    )
    metrics.write_text(
        "\n".join(json.dumps(row, ensure_ascii=False) for row in metric_rows) + "\n",
        encoding="utf-8",
    )

    payload = build_progress_payload(
        doctor_report_path=doctor,
        plan_metrics_path=metrics,
        limit=10,
        graph_width=10,
    )
    assert payload["counts"]["doctor_reports"] == 3
    assert payload["counts"]["plan_metric_events"] == 3
    assert "score" in payload["graphs"]
    assert "fallback_rate" in payload["graphs"]
    assert "blocked_rate" in payload["graphs"]
