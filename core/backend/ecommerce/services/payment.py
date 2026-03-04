"""
Payment Service — Stripe SDK 초기화 및 결제 처리.

환경변수:
    STRIPE_SECRET_KEY: Stripe 비밀 키 (필수)
    STRIPE_WEBHOOK_SECRET: 웹훅 서명 검증 키 (필수)
"""

import logging
from typing import Optional

import stripe

from core.system.security import require_env

logger = logging.getLogger(__name__)
_STRIPE_READY = False


def _init_stripe() -> None:
    """Stripe SDK를 환경변수 기반으로 초기화."""
    global _STRIPE_READY
    if _STRIPE_READY:
        return
    stripe.api_key = require_env("STRIPE_SECRET_KEY")
    _STRIPE_READY = True
    logger.info("Stripe SDK initialized: api_version=%s", stripe.api_version)


def create_payment_intent(
    amount: int,
    currency: str = "krw",
    metadata: Optional[dict] = None,
    idempotency_key: Optional[str] = None,
) -> stripe.PaymentIntent:
    """PaymentIntent 생성.

    Args:
        amount: 결제 금액 (최소 단위, 원화 기준 원)
        currency: ISO 4217 통화 코드 (기본값: krw)
        metadata: 주문 관련 추가 메타데이터

    Returns:
        생성된 PaymentIntent 객체
    """
    _init_stripe()
    create_kwargs = {
        "amount": amount,
        "currency": currency,
        "metadata": metadata or {},
    }
    if idempotency_key:
        create_kwargs["idempotency_key"] = idempotency_key

    intent = stripe.PaymentIntent.create(
        **create_kwargs,
    )
    logger.info(
        "PaymentIntent created: id=%s, amount=%s, currency=%s",
        intent.id,
        amount,
        currency,
    )
    return intent


def construct_webhook_event(payload: bytes, sig_header: str) -> stripe.Event:
    """웹훅 서명 검증 후 Event 객체 반환.

    Args:
        payload: 요청 raw body
        sig_header: Stripe-Signature 헤더 값

    Returns:
        검증된 stripe.Event 객체

    Raises:
        stripe.error.SignatureVerificationError: 서명 불일치
    """
    _init_stripe()
    webhook_secret = require_env("STRIPE_WEBHOOK_SECRET")
    event = stripe.Webhook.construct_event(
        payload=payload,
        sig_header=sig_header,
        secret=webhook_secret,
    )
    logger.info("Webhook event verified: type=%s, id=%s", event["type"], event["id"])
    return event
