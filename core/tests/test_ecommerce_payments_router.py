"""
E-commerce payment router regression tests.
"""

from __future__ import annotations

import importlib
import sys

import pytest
from fastapi.testclient import TestClient


@pytest.fixture
def ecommerce_client(monkeypatch, tmp_path):
    monkeypatch.setenv("JWT_SECRET_KEY", "test-jwt-secret-key-not-for-production")
    monkeypatch.setenv("DATABASE_URL", f"sqlite:///{tmp_path / 'ecommerce.db'}")

    for mod_name in list(sys.modules.keys()):
        if mod_name.startswith("core.backend.ecommerce"):
            del sys.modules[mod_name]

    app_module = importlib.import_module("core.backend.ecommerce.main")
    with TestClient(app_module.app) as client:
        yield client, app_module.app


def test_payment_routes_registered(ecommerce_client):
    _, app = ecommerce_client
    paths = {route.path for route in app.router.routes}
    assert "/api/v1/payments/intent" in paths
    assert "/api/v1/payments/webhook" in paths
    assert "/api/v1/payments/orders/{order_id}" in paths


def test_payment_webhook_requires_signature(ecommerce_client):
    client, _ = ecommerce_client
    resp = client.post("/api/v1/payments/webhook", content=b"{}")
    assert resp.status_code == 400


def test_payment_webhook_fails_without_stripe_secret(ecommerce_client):
    client, _ = ecommerce_client
    resp = client.post(
        "/api/v1/payments/webhook",
        content=b"{}",
        headers={"Stripe-Signature": "t=1,v1=fake"},
    )
    assert resp.status_code == 503
