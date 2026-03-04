#!/usr/bin/env python3
"""
plan_dispatch_replay 단위 테스트
"""

from __future__ import annotations

import json
from pathlib import Path

from core.system.plan_dispatch_replay import (
    append_replay_summary,
    backfill_replay_history,
    build_summary,
    detect_drift,
    load_recent_tasks,
    replay_tasks,
)


def test_load_recent_tasks_deduplicates_and_limits(tmp_path: Path):
    source = tmp_path / "reports.jsonl"
    rows = [
        {"task": "ok"},
        {"task": "스킬 및 기본 툴 체계를 분석하고 업그레이드 항목을 구현한다"},
        {"task": "ok"},
        {"task": "백엔드 코드 로직 미구현/결함 전수조사 및 안전화 기획 구축"},
    ]
    source.write_text("\n".join(json.dumps(r, ensure_ascii=False) for r in rows) + "\n", encoding="utf-8")

    tasks = load_recent_tasks(source, limit=2)
    assert len(tasks) == 2
    assert "ok" in tasks
    assert "백엔드 코드 로직 미구현/결함 전수조사 및 안전화 기획 구축" in tasks


def test_replay_summary_has_expected_keys():
    tasks = [
        "ok",
        "스킬 및 기본 툴 체계를 분석하고 업그레이드 항목을 구현한다",
    ]
    rows = replay_tasks(tasks, min_complexity="medium")
    summary = build_summary(rows, min_complexity="medium", source=Path("dummy.jsonl"))
    assert summary["total"] == 2
    assert "complexity_counts" in summary
    assert "allowed_rate" in summary
    assert isinstance(summary["top_signals"], list)


def test_replay_on_real_reports_file_if_exists():
    source = Path("knowledge/system/plan_council_reports.jsonl")
    if not source.exists():
        return
    tasks = load_recent_tasks(source, limit=5)
    if not tasks:
        return
    rows = replay_tasks(tasks, min_complexity="medium")
    summary = build_summary(rows, min_complexity="medium", source=source)
    assert summary["total"] == len(rows)


def test_append_replay_summary_writes_jsonl(tmp_path: Path):
    source = tmp_path / "source.jsonl"
    source.write_text(json.dumps({"task": "ok"}, ensure_ascii=False) + "\n", encoding="utf-8")
    rows = replay_tasks(["ok"], min_complexity="medium")
    summary = build_summary(rows, min_complexity="medium", source=source)

    report_log = tmp_path / "replay_reports.jsonl"
    appended = append_replay_summary(report_log, summary)
    assert report_log.exists()
    lines = [line for line in report_log.read_text(encoding="utf-8").splitlines() if line.strip()]
    assert len(lines) == 1
    payload = json.loads(lines[0])
    assert payload["total"] == 1
    assert payload["min_complexity"] == "medium"
    assert appended["total"] == 1


def test_detect_drift_reports_threshold_exceeded(tmp_path: Path):
    report_log = tmp_path / "replay_reports.jsonl"
    baseline = {
        "source": "dummy",
        "min_complexity": "medium",
        "total": 20,
        "allowed_count": 16,
        "allowed_rate": 0.8,
        "complexity_counts": {"simple": 4, "medium": 10, "high": 6},
        "top_signals": [],
    }
    for _ in range(6):
        append_replay_summary(report_log, baseline)

    current = {
        "source": "dummy",
        "min_complexity": "medium",
        "total": 20,
        "allowed_count": 6,
        "allowed_rate": 0.3,
        "complexity_counts": {"simple": 16, "medium": 3, "high": 1},
        "top_signals": [],
    }

    drift = detect_drift(
        current,
        report_log,
        window=14,
        min_samples=5,
        allowed_rate_threshold=0.2,
        complexity_rate_threshold=0.25,
    )
    assert drift["detected"] is True
    assert drift["reason"] == "threshold_exceeded"
    assert drift["allowed_delta"] >= 0.2


def test_backfill_replay_history_appends_snapshots(tmp_path: Path):
    source = tmp_path / "reports.jsonl"
    rows = []
    for idx in range(1, 26):
        rows.append(
            {
                "task": (
                    f"스킬 및 기본 툴 체계를 분석하고 업그레이드 항목을 구현한다 {idx}"
                    if idx % 2 == 0
                    else f"ok-{idx}"
                )
            }
        )
    source.write_text(
        "\n".join(json.dumps(row, ensure_ascii=False) for row in rows) + "\n",
        encoding="utf-8",
    )

    report_log = tmp_path / "replay_reports.jsonl"
    result = backfill_replay_history(
        source=source,
        report_log=report_log,
        min_complexity="medium",
        limit=25,
        snapshots=6,
    )
    assert result["reason"] == "ok"
    assert result["appended"] >= 2
    assert report_log.exists()
    lines = [line for line in report_log.read_text(encoding="utf-8").splitlines() if line.strip()]
    assert len(lines) == result["appended"]
