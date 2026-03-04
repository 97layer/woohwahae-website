#!/usr/bin/env python3
"""
monitor_dashboard plan_dispatch daily status tests
"""

from __future__ import annotations

import json
from pathlib import Path

from core.system.monitor_dashboard import MonitorDashboard


def test_plan_dispatch_daily_status_unavailable(tmp_path: Path):
    dashboard = MonitorDashboard()
    dashboard.plan_dispatch_daily_dir = tmp_path / "missing"
    status = dashboard.get_plan_dispatch_daily_status()
    assert status["available"] is False


def test_plan_dispatch_daily_status_reads_latest_report(tmp_path: Path):
    dashboard = MonitorDashboard()
    dashboard.plan_dispatch_daily_dir = tmp_path

    older = tmp_path / "plan_dispatch_20260303.json"
    older.write_text(
        json.dumps(
            {
                "generated_at": "2026-03-03T08:00:00+00:00",
                "health": {"status": "warn"},
                "metrics": {"fallback_rate": 0.2},
                "replay": {"allowed_rate": 0.6},
            },
            ensure_ascii=False,
        ),
        encoding="utf-8",
    )
    latest = tmp_path / "plan_dispatch_20260304.json"
    latest.write_text(
        json.dumps(
            {
                "generated_at": "2026-03-04T08:00:00+00:00",
                "health": {"status": "pass"},
                "metrics": {"fallback_rate": 0.05},
                "replay": {"allowed_rate": 0.75},
            },
            ensure_ascii=False,
        ),
        encoding="utf-8",
    )

    status = dashboard.get_plan_dispatch_daily_status()
    assert status["available"] is True
    assert status["status"] == "pass"
    assert status["fallback_rate"] == 0.05
    assert status["allowed_rate"] == 0.75
    assert status["file"].endswith("plan_dispatch_20260304.json")


def test_plan_dispatch_daily_status_invalid_json(tmp_path: Path):
    dashboard = MonitorDashboard()
    dashboard.plan_dispatch_daily_dir = tmp_path
    broken = tmp_path / "plan_dispatch_20260304.json"
    broken.write_text("{invalid-json", encoding="utf-8")

    status = dashboard.get_plan_dispatch_daily_status()
    assert status["available"] is True
    assert status["status"] == "invalid"
