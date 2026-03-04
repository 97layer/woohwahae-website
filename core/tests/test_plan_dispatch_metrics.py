#!/usr/bin/env python3
"""
plan_dispatch_metrics 단위 테스트
"""

from __future__ import annotations

from pathlib import Path

from core.system.plan_dispatch_metrics import append_metric, summarize_metrics


def test_append_and_summary_counts(tmp_path: Path):
    log_path = tmp_path / "plan_dispatch_metrics.jsonl"

    append_metric(
        log_path=log_path,
        task="스킬 및 기본 툴 체계를 분석하고 업그레이드 항목을 구현한다",
        mode="auto",
        phase="smoke",
        reason="smoke_mode",
        executed=True,
        complexity="medium",
        score=3,
        fallback=False,
    )
    append_metric(
        log_path=log_path,
        task="ok",
        mode="auto",
        phase="skip",
        reason="simple_task",
        executed=False,
        complexity="simple",
        score=0,
        fallback=True,
    )

    summary = summarize_metrics(log_path=log_path, window=100)
    assert summary["total"] == 2
    assert summary["executed_count"] == 1
    assert summary["fallback_count"] == 1
    assert summary["phase_counts"]["smoke"] == 1
    assert summary["phase_counts"]["skip"] == 1
    assert summary["reason_counts"]["simple_task"] == 1
    assert summary["complexity_counts"]["medium"] == 1
    assert summary["complexity_counts"]["simple"] == 1
