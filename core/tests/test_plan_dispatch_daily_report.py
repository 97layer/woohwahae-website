#!/usr/bin/env python3
"""
plan_dispatch_daily_report 통합 테스트
"""

from __future__ import annotations

import json
import subprocess
from pathlib import Path

from core.system.plan_dispatch_metrics import append_metric


PROJECT_ROOT = Path(__file__).resolve().parents[2]
DAILY_REPORT_SCRIPT = PROJECT_ROOT / "core" / "scripts" / "plan_dispatch_daily_report.py"


def _parse_json_from_stdout(raw: str) -> dict:
    text = (raw or "").strip()
    if not text:
        raise ValueError("empty stdout")
    for line in reversed(text.splitlines()):
        candidate = line.strip()
        if not candidate:
            continue
        try:
            payload = json.loads(candidate)
        except Exception:
            continue
        if isinstance(payload, dict):
            return payload
    raise ValueError("json payload not found")


def test_daily_report_writes_output_file(tmp_path: Path):
    metrics_log = tmp_path / "metrics.jsonl"
    append_metric(
        log_path=metrics_log,
        task="스킬 및 기본 툴 체계를 분석하고 업그레이드 항목을 구현한다",
        mode="auto",
        phase="execute",
        reason="executed",
        executed=True,
        complexity="medium",
        score=3,
        fallback=False,
    )

    source = tmp_path / "plan_reports.jsonl"
    source.write_text(
        json.dumps({"task": "스킬 및 기본 툴 체계를 분석하고 업그레이드 항목을 구현한다"}, ensure_ascii=False)
        + "\n",
        encoding="utf-8",
    )

    replay_log = tmp_path / "replay_reports.jsonl"
    out_file = tmp_path / "daily_report.json"
    proc = subprocess.run(
        [
            "python3",
            str(DAILY_REPORT_SCRIPT),
            "--source",
            str(source),
            "--metrics-log",
            str(metrics_log),
            "--replay-log",
            str(replay_log),
            "--output",
            str(out_file),
            "--write",
            "--json",
        ],
        cwd=str(PROJECT_ROOT),
        capture_output=True,
        text=True,
        check=False,
    )
    assert proc.returncode == 0, proc.stderr or proc.stdout
    payload = _parse_json_from_stdout(proc.stdout)
    assert payload["type"] == "plan_dispatch_daily"
    assert payload["report_path"] == str(out_file)
    assert payload["health"]["status"] in {"pass", "warn", "fail"}
    assert out_file.exists()


def test_daily_report_marks_fail_when_fallback_exceeds_threshold(tmp_path: Path):
    metrics_log = tmp_path / "metrics.jsonl"
    append_metric(
        log_path=metrics_log,
        task="ok",
        mode="auto",
        phase="skip",
        reason="simple_task",
        executed=False,
        complexity="simple",
        score=0,
        fallback=True,
    )
    append_metric(
        log_path=metrics_log,
        task="ok-2",
        mode="auto",
        phase="skip",
        reason="simple_task",
        executed=False,
        complexity="simple",
        score=0,
        fallback=True,
    )

    source = tmp_path / "plan_reports.jsonl"
    source.write_text(json.dumps({"task": "ok"}, ensure_ascii=False) + "\n", encoding="utf-8")
    replay_log = tmp_path / "replay_reports.jsonl"
    out_file = tmp_path / "daily_report_fail.json"

    proc = subprocess.run(
        [
            "python3",
            str(DAILY_REPORT_SCRIPT),
            "--source",
            str(source),
            "--metrics-log",
            str(metrics_log),
            "--replay-log",
            str(replay_log),
            "--fallback-min-samples",
            "2",
            "--fallback-threshold",
            "0.40",
            "--output",
            str(out_file),
            "--write",
            "--json",
        ],
        cwd=str(PROJECT_ROOT),
        capture_output=True,
        text=True,
        check=False,
    )
    assert proc.returncode == 0, proc.stderr or proc.stdout
    payload = _parse_json_from_stdout(proc.stdout)
    assert payload["health"]["status"] == "fail"
    issues = payload["health"]["issues"]
    assert any("fallback_rate" in str(item) for item in issues)


def test_daily_report_fail_on_health_returns_nonzero(tmp_path: Path):
    metrics_log = tmp_path / "metrics.jsonl"
    append_metric(
        log_path=metrics_log,
        task="ok",
        mode="auto",
        phase="skip",
        reason="simple_task",
        executed=False,
        complexity="simple",
        score=0,
        fallback=True,
    )
    append_metric(
        log_path=metrics_log,
        task="ok-2",
        mode="auto",
        phase="skip",
        reason="simple_task",
        executed=False,
        complexity="simple",
        score=0,
        fallback=True,
    )
    source = tmp_path / "plan_reports.jsonl"
    source.write_text(json.dumps({"task": "ok"}, ensure_ascii=False) + "\n", encoding="utf-8")
    replay_log = tmp_path / "replay_reports.jsonl"
    out_file = tmp_path / "daily_report_fail_strict.json"

    proc = subprocess.run(
        [
            "python3",
            str(DAILY_REPORT_SCRIPT),
            "--source",
            str(source),
            "--metrics-log",
            str(metrics_log),
            "--replay-log",
            str(replay_log),
            "--fallback-min-samples",
            "2",
            "--fallback-threshold",
            "0.40",
            "--output",
            str(out_file),
            "--write",
            "--json",
            "--fail-on-health",
        ],
        cwd=str(PROJECT_ROOT),
        capture_output=True,
        text=True,
        check=False,
    )
    assert proc.returncode == 1
    payload = _parse_json_from_stdout(proc.stdout)
    assert payload["health"]["status"] == "fail"
