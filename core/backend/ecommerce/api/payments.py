"""Payment API endpoints (Stripe intent + webhook)."""

from __future__ import annotations

from datetime import datetime, timezone
from decimal import Decimal

import stripe
from fastapi import APIRouter, Depends, HTTPException, Request, status
from pydantic import BaseModel, Field
from sqlalchemy.orm import Session

from ..models import Order, PaymentStatus, User, get_db
from ..services.payment import construct_webhook_event, create_payment_intent
from ..utils import get_current_active_admin, get_current_user

router = APIRouter(prefix="/payments", tags=["Payments"])


class PaymentIntentCreate(BaseModel):
    order_id: int = Field(..., gt=0)


@router.post("/intent")
def create_intent(
    payload: PaymentIntentCreate,
    current_user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """Create Stripe PaymentIntent for current user's order."""
    order = db.query(Order).filter(
        Order.id == payload.order_id,
        Order.user_id == current_user.id,
    ).first()
    if not order:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Order not found")
    if order.payment_status == PaymentStatus.PAID.value:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Order already paid")

    total = Decimal(order.total)
    amount = int(total)
    if amount <= 0:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Invalid order amount")

    try:
        intent = create_payment_intent(
            amount=amount,
            currency="krw",
            metadata={"order_id": str(order.id), "order_number": order.order_number},
        )
    except RuntimeError as exc:
        raise HTTPException(status_code=status.HTTP_503_SERVICE_UNAVAILABLE, detail=str(exc)) from exc
    except stripe.error.StripeError as exc:
        raise HTTPException(status_code=status.HTTP_502_BAD_GATEWAY, detail=f"Stripe error: {exc.user_message or exc.code}") from exc

    order.payment_transaction_id = intent.id
    order.payment_gateway = "stripe"
    db.commit()

    return {
        "order_id": order.id,
        "payment_intent_id": intent.id,
        "client_secret": intent.client_secret,
        "amount": amount,
        "currency": "krw",
        "status": order.payment_status,
    }


@router.get("/orders/{order_id}")
def get_payment_status(
    order_id: int,
    current_user: User = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    order = db.query(Order).filter(
        Order.id == order_id,
        Order.user_id == current_user.id,
    ).first()
    if not order:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Order not found")
    return {
        "order_id": order.id,
        "payment_status": order.payment_status,
        "payment_gateway": order.payment_gateway,
        "payment_transaction_id": order.payment_transaction_id,
        "paid_at": order.paid_at,
    }


@router.post("/webhook")
async def stripe_webhook(request: Request, db: Session = Depends(get_db)):
    """Stripe webhook endpoint."""
    body = await request.body()
    sig_header = request.headers.get("Stripe-Signature", "")
    if not sig_header:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Missing Stripe-Signature header")

    try:
        event = construct_webhook_event(body, sig_header)
    except RuntimeError as exc:
        raise HTTPException(status_code=status.HTTP_503_SERVICE_UNAVAILABLE, detail=str(exc)) from exc
    except stripe.error.SignatureVerificationError as exc:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=f"Invalid signature: {exc}") from exc
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=f"Invalid payload: {exc}") from exc

    event_type = event.get("type")
    event_obj = event.get("data", {}).get("object", {})
    payment_intent_id = event_obj.get("id")
    if not payment_intent_id:
        return {"received": True}

    order = db.query(Order).filter(Order.payment_transaction_id == payment_intent_id).first()
    if not order:
        return {"received": True}

    if event_type == "payment_intent.succeeded":
        order.payment_status = PaymentStatus.PAID.value
        if not order.paid_at:
            order.paid_at = datetime.now(timezone.utc)
    elif event_type == "payment_intent.payment_failed":
        order.payment_status = PaymentStatus.FAILED.value

    db.commit()
    return {"received": True}


@router.post("/orders/{order_id}/mark-paid")
def mark_paid_manually(
    order_id: int,
    _: User = Depends(get_current_active_admin),
    db: Session = Depends(get_db),
):
    """Admin fallback: mark payment as paid manually."""
    order = db.query(Order).filter(Order.id == order_id).first()
    if not order:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Order not found")

    order.payment_status = PaymentStatus.PAID.value
    if not order.paid_at:
        order.paid_at = datetime.now(timezone.utc)
    db.commit()
    db.refresh(order)
    return {
        "order_id": order.id,
        "payment_status": order.payment_status,
        "paid_at": order.paid_at,
    }
