"""Cart schemas for Redis session management."""
from typing import List, Optional
from decimal import Decimal
from pydantic import BaseModel, Field


class CartItemBase(BaseModel):
    """Base cart item schema."""
    product_id: int
    quantity: int = Field(..., gt=0)


class CartItemAdd(CartItemBase):
    """Add item to cart schema."""
    pass


class CartItemUpdate(BaseModel):
    """Update cart item quantity."""
    quantity: int = Field(..., ge=0)


class CartItemResponse(BaseModel):
    """Cart item response with product details."""
    product_id: int
    product_name: str
    product_sku: str
    product_image: Optional[str]
    price: Decimal
    quantity: int
    subtotal: Decimal


class CartResponse(BaseModel):
    """Full cart response."""
    items: List[CartItemResponse]
    subtotal: Decimal
    total_items: int
