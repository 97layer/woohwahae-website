#!/usr/bin/env python3
"""
plan_dispatch_classifier 단위 테스트

로컬 plan_council_reports.jsonl에서 반복된 태스크 문구를 회귀 케이스로 고정한다.
"""

from __future__ import annotations

import json
from pathlib import Path

import pytest

from core.system.plan_dispatch_classifier import classify_task


LOG_DERIVED_NONTRIVIAL_CASES = (
    "백엔드 코드 로직 미구현/결함 전수조사 및 안전화 기획 구축",
    "스킬 및 기본 툴 체계를 분석하고 업그레이드 항목을 구현한다",
    "로컬 저장소의 plan_council_reports.jsonl 샘플을 사용해 plan_dispatch 분류 규칙을 보정하고, harness_doctor --run-tests를 실행하며, 관련 회귀 테스트를 추가한다",
)


def test_simple_task_is_skipped_under_default_threshold():
    result = classify_task("ok", min_complexity="medium")
    assert result["complexity"] == "simple"
    assert result["allowed"] is False
    assert result["level"] == 0


@pytest.mark.parametrize("task", LOG_DERIVED_NONTRIVIAL_CASES)
def test_log_derived_nontrivial_cases_are_not_simple(task: str):
    result = classify_task(task, min_complexity="medium")
    assert result["complexity"] in {"medium", "high"}
    assert result["allowed"] is True


def test_high_threshold_rejects_medium_complexity():
    task = "스킬 및 기본 툴 체계를 분석하고 업그레이드 항목을 구현한다"
    result = classify_task(task, min_complexity="high")
    assert result["complexity"] in {"medium", "high"}
    if result["complexity"] == "medium":
        assert result["allowed"] is False


def test_reports_file_tasks_classify_without_error():
    reports = Path("knowledge/system/plan_council_reports.jsonl")
    assert reports.exists()

    tasks = []
    for raw in reports.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line:
            continue
        try:
            payload = json.loads(line)
        except Exception:  # noqa: BLE001
            continue
        task = str(payload.get("task", "")).strip()
        if task:
            tasks.append(task)
        if len(tasks) >= 5:
            break

    assert tasks, "plan_council_reports.jsonl에서 task를 읽지 못했습니다."
    for task in tasks:
        result = classify_task(task, min_complexity="medium")
        assert result["complexity"] in {"simple", "medium", "high"}
        assert isinstance(result["score"], int)
