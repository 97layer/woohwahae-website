"""Payment gateway integration stubs (Stripe & TossPayments)."""
from typing import Optional, Dict, Any
from decimal import Decimal
from enum import Enum


class PaymentGateway(str, Enum):
    """Supported payment gateways."""
    STRIPE = "stripe"
    TOSSPAYMENTS = "tosspayments"


class PaymentResult:
    """Payment processing result."""

    def __init__(
        self,
        success: bool,
        transaction_id: Optional[str] = None,
        error_message: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None
    ):
        self.success = success
        self.transaction_id = transaction_id
        self.error_message = error_message
        self.metadata = metadata or {}


class StripePaymentService:
    """Stripe payment integration stub.

    Implementation notes:
    1. Install: pip install stripe
    2. Set API key: stripe.api_key = "sk_test_..."
    3. Implement payment intents API
    4. Handle webhooks for payment confirmation
    """

    def __init__(self, api_key: str):
        """Initialize Stripe service."""
        self.api_key = api_key
        # TODO: Initialize Stripe SDK
        # import stripe
        # stripe.api_key = api_key

    def create_payment_intent(
        self,
        amount: Decimal,
        currency: str = "krw",
        metadata: Optional[Dict[str, Any]] = None
    ) -> PaymentResult:
        """Create Stripe payment intent.

        Args:
            amount: Payment amount in smallest currency unit (e.g., cents for USD, won for KRW)
            currency: ISO currency code (default: krw)
            metadata: Additional metadata to attach to payment

        Returns:
            PaymentResult with client_secret for frontend integration
        """
        # TODO: Implement Stripe payment intent creation
        # try:
        #     intent = stripe.PaymentIntent.create(
        #         amount=int(amount),
        #         currency=currency,
        #         metadata=metadata
        #     )
        #     return PaymentResult(
        #         success=True,
        #         transaction_id=intent.id,
        #         metadata={"client_secret": intent.client_secret}
        #     )
        # except stripe.error.StripeError as e:
        #     return PaymentResult(success=False, error_message=str(e))

        return PaymentResult(
            success=True,
            transaction_id="pi_stub_123456",
            metadata={"client_secret": "pi_stub_secret_123456"}
        )

    def verify_payment(self, payment_intent_id: str) -> PaymentResult:
        """Verify payment status.

        Args:
            payment_intent_id: Stripe payment intent ID

        Returns:
            PaymentResult with verification status
        """
        # TODO: Implement payment verification
        # try:
        #     intent = stripe.PaymentIntent.retrieve(payment_intent_id)
        #     return PaymentResult(
        #         success=intent.status == "succeeded",
        #         transaction_id=intent.id
        #     )
        # except stripe.error.StripeError as e:
        #     return PaymentResult(success=False, error_message=str(e))

        return PaymentResult(success=True, transaction_id=payment_intent_id)


class TossPaymentsService:
    """TossPayments integration stub.

    Implementation notes:
    1. Get API credentials from https://developers.tosspayments.com
    2. Install requests: pip install requests
    3. Implement payment widget or API integration
    4. Handle payment confirmation callback
    """

    def __init__(self, secret_key: str, client_key: str):
        """Initialize TossPayments service."""
        self.secret_key = secret_key
        self.client_key = client_key
        self.base_url = "https://api.tosspayments.com/v1"

    def create_payment(
        self,
        amount: Decimal,
        order_id: str,
        order_name: str,
        customer_email: Optional[str] = None,
        customer_name: Optional[str] = None
    ) -> PaymentResult:
        """Create TossPayments payment.

        Args:
            amount: Payment amount in KRW
            order_id: Unique order identifier
            order_name: Order description
            customer_email: Customer email
            customer_name: Customer name

        Returns:
            PaymentResult with payment URL or widget data
        """
        # TODO: Implement TossPayments API call
        # import requests
        # import base64
        #
        # auth = base64.b64encode(f"{self.secret_key}:".encode()).decode()
        # headers = {
        #     "Authorization": f"Basic {auth}",
        #     "Content-Type": "application/json"
        # }
        # data = {
        #     "amount": int(amount),
        #     "orderId": order_id,
        #     "orderName": order_name,
        #     "customerEmail": customer_email,
        #     "customerName": customer_name
        # }
        # response = requests.post(
        #     f"{self.base_url}/payments",
        #     headers=headers,
        #     json=data
        # )
        # if response.status_code == 200:
        #     result = response.json()
        #     return PaymentResult(
        #         success=True,
        #         transaction_id=result.get("paymentKey"),
        #         metadata={"checkout_url": result.get("checkoutUrl")}
        #     )
        # return PaymentResult(success=False, error_message=response.text)

        return PaymentResult(
            success=True,
            transaction_id="toss_stub_123456",
            metadata={
                "checkout_url": "https://checkout.tosspayments.com/stub",
                "client_key": self.client_key
            }
        )

    def confirm_payment(
        self,
        payment_key: str,
        order_id: str,
        amount: Decimal
    ) -> PaymentResult:
        """Confirm payment after customer approval.

        Args:
            payment_key: Payment key from TossPayments
            order_id: Order identifier
            amount: Payment amount for verification

        Returns:
            PaymentResult with confirmation status
        """
        # TODO: Implement payment confirmation
        # import requests
        # import base64
        #
        # auth = base64.b64encode(f"{self.secret_key}:".encode()).decode()
        # headers = {
        #     "Authorization": f"Basic {auth}",
        #     "Content-Type": "application/json"
        # }
        # data = {
        #     "orderId": order_id,
        #     "amount": int(amount)
        # }
        # response = requests.post(
        #     f"{self.base_url}/payments/{payment_key}/confirm",
        #     headers=headers,
        #     json=data
        # )
        # if response.status_code == 200:
        #     result = response.json()
        #     return PaymentResult(
        #         success=True,
        #         transaction_id=result.get("paymentKey")
        #     )
        # return PaymentResult(success=False, error_message=response.text)

        return PaymentResult(success=True, transaction_id=payment_key)


class PaymentService:
    """Unified payment service facade."""

    def __init__(self):
        """Initialize payment services."""
        # TODO: Load from environment variables
        self.stripe = StripePaymentService(api_key="sk_test_stub")
        self.tosspayments = TossPaymentsService(
            secret_key="test_sk_stub",
            client_key="test_ck_stub"
        )

    def process_payment(
        self,
        gateway: PaymentGateway,
        amount: Decimal,
        order_id: str,
        **kwargs
    ) -> PaymentResult:
        """Process payment through specified gateway."""
        if gateway == PaymentGateway.STRIPE:
            return self.stripe.create_payment_intent(
                amount=amount,
                metadata={"order_id": order_id, **kwargs}
            )
        elif gateway == PaymentGateway.TOSSPAYMENTS:
            return self.tosspayments.create_payment(
                amount=amount,
                order_id=order_id,
                order_name=kwargs.get("order_name", f"Order #{order_id}"),
                customer_email=kwargs.get("customer_email"),
                customer_name=kwargs.get("customer_name")
            )
        else:
            return PaymentResult(
                success=False,
                error_message=f"Unsupported payment gateway: {gateway}"
            )


# Global payment service instance
payment_service = PaymentService()
