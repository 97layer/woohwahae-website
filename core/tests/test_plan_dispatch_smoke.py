#!/usr/bin/env python3
"""
plan_dispatch --smoke 통합 테스트
"""

from __future__ import annotations

import json
import os
import subprocess
from pathlib import Path


PROJECT_ROOT = Path(__file__).resolve().parents[2]
PLAN_DISPATCH = PROJECT_ROOT / "core" / "scripts" / "plan_dispatch.sh"


def _run_plan_dispatch(task: str, mode: str = "auto", extra_env: dict | None = None) -> dict:
    assert mode in {"auto", "manual"}
    cmd = ["bash", str(PLAN_DISPATCH), task, f"--{mode}", "--smoke"]
    env = os.environ.copy()
    if extra_env:
        env.update(extra_env)
    proc = subprocess.run(
        cmd,
        cwd=str(PROJECT_ROOT),
        capture_output=True,
        text=True,
        env=env,
    )
    assert proc.returncode == 0, f"stderr={proc.stderr}\nstdout={proc.stdout}"
    payload = json.loads(proc.stdout)
    assert isinstance(payload, dict)
    return payload


def test_auto_simple_task_skips_dispatcher_execution():
    payload = _run_plan_dispatch("ok", mode="auto")
    dispatcher = payload["dispatcher"]
    assert dispatcher["executed"] is False
    assert dispatcher["reason"] == "simple_task"
    assert dispatcher["complexity"] == "simple"


def test_auto_nontrivial_task_executes_dispatcher():
    payload = _run_plan_dispatch(
        "스킬 및 기본 툴 체계를 분석하고 업그레이드 항목을 구현한다",
        mode="auto",
    )
    dispatcher = payload["dispatcher"]
    assert dispatcher["executed"] is True
    assert dispatcher["complexity"] in {"medium", "high"}
    consensus = payload["consensus"]
    assert consensus["status"] in {"smoke", "ready", "degraded"}


def test_manual_mode_executes_even_with_short_task():
    payload = _run_plan_dispatch("ok", mode="manual")
    dispatcher = payload["dispatcher"]
    assert dispatcher["executed"] is True
    assert dispatcher["reason"] == "smoke_mode"


def test_auto_fallback_when_classifier_output_is_invalid_json(tmp_path):
    bad_classifier = tmp_path / "bad_classifier.py"
    bad_classifier.write_text("print('not-json')\n", encoding="utf-8")

    payload = _run_plan_dispatch(
        "스킬 및 기본 툴 체계를 분석하고 업그레이드 항목을 구현한다",
        mode="auto",
        extra_env={"PLAN_DISPATCH_CLASSIFIER_SCRIPT": str(bad_classifier)},
    )
    dispatcher = payload["dispatcher"]
    assert dispatcher["executed"] is False
    assert dispatcher["reason"] == "simple_task"
    assert dispatcher["complexity"] == "simple"
