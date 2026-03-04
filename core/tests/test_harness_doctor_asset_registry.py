#!/usr/bin/env python3
"""
harness_doctor asset registry integrity checks
"""

from __future__ import annotations

import json
from pathlib import Path

from core.scripts import harness_doctor as hd


def _write_registry(path: Path, payload: dict) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(payload, ensure_ascii=False), encoding="utf-8")


def test_asset_registry_check_pass(monkeypatch, tmp_path: Path):
    monkeypatch.setattr(hd, "PROJECT_ROOT", tmp_path)
    registry_path = tmp_path / "knowledge" / "system" / "asset_registry.json"
    _write_registry(
        registry_path,
        {
            "assets": [{"id": "AST-2026-03-001"}, {"id": "signal-1"}],
            "stats": {"total": 2},
        },
    )

    result = hd.check_asset_registry_integrity()
    assert result.status == "pass"


def test_asset_registry_check_fails_invalid_json(monkeypatch, tmp_path: Path):
    monkeypatch.setattr(hd, "PROJECT_ROOT", tmp_path)
    registry_path = tmp_path / "knowledge" / "system" / "asset_registry.json"
    registry_path.parent.mkdir(parents=True, exist_ok=True)
    registry_path.write_text("{invalid-json", encoding="utf-8")

    result = hd.check_asset_registry_integrity()
    assert result.status == "fail"
    assert "invalid json" in result.detail


def test_asset_registry_check_fails_duplicate_ast(monkeypatch, tmp_path: Path):
    monkeypatch.setattr(hd, "PROJECT_ROOT", tmp_path)
    registry_path = tmp_path / "knowledge" / "system" / "asset_registry.json"
    _write_registry(
        registry_path,
        {
            "assets": [{"id": "AST-2026-03-001"}, {"id": "AST-2026-03-001"}],
            "stats": {"total": 2},
        },
    )

    result = hd.check_asset_registry_integrity()
    assert result.status == "fail"
    assert "duplicate AST ids" in result.detail
