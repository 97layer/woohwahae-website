"""E-commerce payment service tests."""

from __future__ import annotations

import importlib
import sys


def test_create_payment_intent_passes_idempotency_key(monkeypatch):
    monkeypatch.setenv("STRIPE_SECRET_KEY", "sk_test_not_for_production")

    for mod_name in list(sys.modules.keys()):
        if mod_name == "core.backend.ecommerce.services.payment":
            del sys.modules[mod_name]

    payment_service = importlib.import_module("core.backend.ecommerce.services.payment")
    payment_service._STRIPE_READY = False

    captured: dict[str, object] = {}

    class DummyIntent:
        id = "pi_test_123"

    def fake_create(**kwargs):
        captured.update(kwargs)
        return DummyIntent()

    monkeypatch.setattr(payment_service.stripe.PaymentIntent, "create", fake_create)

    payment_service.create_payment_intent(
        amount=1000,
        currency="krw",
        metadata={"order_id": "1"},
        idempotency_key="order-1-user-1-total-1000",
    )

    assert captured["amount"] == 1000
    assert captured["currency"] == "krw"
    assert captured["metadata"] == {"order_id": "1"}
    assert captured["idempotency_key"] == "order-1-user-1-total-1000"
