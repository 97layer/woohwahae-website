"""Order schemas for request/response validation."""
from datetime import datetime
from typing import Optional, List, Dict
from decimal import Decimal
from pydantic import BaseModel, Field


class OrderItemBase(BaseModel):
    """Base order item schema."""
    product_id: int
    quantity: int = Field(..., gt=0)


class OrderItemCreate(OrderItemBase):
    """Order item creation schema."""
    pass


class OrderItemResponse(BaseModel):
    """Order item response schema."""
    id: int
    product_id: int
    product_name: str
    product_sku: str
    product_image: Optional[str]
    unit_price: Decimal
    quantity: int
    subtotal: Decimal

    class Config:
        from_attributes = True


class OrderCreate(BaseModel):
    """Order creation schema."""
    items: List[OrderItemCreate] = Field(..., min_items=1)
    shipping_address: Dict[str, str]
    shipping_method: Optional[str] = None
    payment_method: str = Field(..., max_length=50)
    customer_note: Optional[str] = None


class OrderUpdate(BaseModel):
    """Order update schema (admin)."""
    status: Optional[str] = None
    payment_status: Optional[str] = None
    tracking_number: Optional[str] = None
    admin_note: Optional[str] = None


class OrderResponse(BaseModel):
    """Order response schema."""
    id: int
    order_number: str
    user_id: int
    status: str
    payment_status: str
    subtotal: Decimal
    shipping_fee: Decimal
    discount_amount: Decimal
    total: Decimal
    shipping_address: Dict[str, str]
    shipping_method: Optional[str]
    tracking_number: Optional[str]
    payment_method: Optional[str]
    customer_note: Optional[str]
    created_at: datetime
    items: List[OrderItemResponse]

    class Config:
        from_attributes = True


class OrderListResponse(BaseModel):
    """Order list with pagination."""
    items: List[OrderResponse]
    total: int
    page: int
    page_size: int
    total_pages: int
