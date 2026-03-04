"""
Backend hardening regression tests.
"""

from __future__ import annotations

import json
import importlib
import sys
from pathlib import Path

from core.system import growth as growth_module


def test_generate_order_number_has_random_suffix(monkeypatch):
    monkeypatch.setenv("JWT_SECRET_KEY", "test-jwt-secret-key-not-for-production")
    for mod_name in list(sys.modules.keys()):
        if mod_name.startswith("core.backend.ecommerce"):
            del sys.modules[mod_name]
    orders_api = importlib.import_module("core.backend.ecommerce.api.orders")

    values = iter([1, 2])
    monkeypatch.setattr(orders_api.secrets, "randbelow", lambda _: next(values))

    first = orders_api.generate_order_number()
    second = orders_api.generate_order_number()

    assert first != second
    assert first.startswith("WH") and second.startswith("WH")
    assert first.endswith("0001")
    assert second.endswith("0002")


def test_growth_monthly_report_preserves_json_metrics(monkeypatch, tmp_path):
    growth_dir = tmp_path / "growth"
    signals_dir = tmp_path / "signals"
    corpus_dir = tmp_path / "corpus"
    clients_dir = tmp_path / "clients"
    archive_dir = tmp_path / "archive"

    for d in [growth_dir, signals_dir, corpus_dir, clients_dir, archive_dir]:
        d.mkdir(parents=True, exist_ok=True)

    monkeypatch.setattr(growth_module, "GROWTH_DIR", growth_dir)
    monkeypatch.setattr(growth_module, "SIGNALS_DIR", signals_dir)
    monkeypatch.setattr(growth_module, "CORPUS_DIR", corpus_dir)
    monkeypatch.setattr(growth_module, "CLIENTS_DIR", clients_dir)
    monkeypatch.setattr(growth_module, "ARCHIVE_DIR", archive_dir)

    gm = growth_module.GrowthModule()
    gm.record_revenue("2026-03", atelier=10000, consulting=20000, products=30000)
    out_path = Path(gm.generate_monthly_report("2026-03"))

    payload = json.loads(out_path.read_text(encoding="utf-8"))
    assert payload["period"] == "2026-03"
    assert payload["revenue"]["total"] == 60000
    assert "monthly_report_markdown" in payload
    assert payload["monthly_report_markdown"].startswith("# Growth Report")
