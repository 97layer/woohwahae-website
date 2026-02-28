"""Product schemas for request/response validation."""
from datetime import datetime
from typing import Optional, List, Dict
from decimal import Decimal
from pydantic import BaseModel, Field


class ProductBase(BaseModel):
    """Base product schema."""
    name: str = Field(..., max_length=200)
    slug: str = Field(..., max_length=200)
    description: Optional[str] = None
    short_description: Optional[str] = Field(None, max_length=500)
    price: Decimal = Field(..., gt=0, decimal_places=2)
    compare_at_price: Optional[Decimal] = Field(None, decimal_places=2)
    category: Optional[str] = Field(None, max_length=100)
    tags: Optional[List[str]] = None
    images: Optional[List[Dict[str, str]]] = None


class ProductCreate(ProductBase):
    """Product creation schema."""
    sku: str = Field(..., max_length=50)
    stock_quantity: int = Field(default=0, ge=0)
    track_inventory: bool = True
    is_published: bool = False


class ProductUpdate(BaseModel):
    """Product update schema."""
    name: Optional[str] = None
    description: Optional[str] = None
    short_description: Optional[str] = None
    price: Optional[Decimal] = None
    compare_at_price: Optional[Decimal] = None
    stock_quantity: Optional[int] = None
    category: Optional[str] = None
    tags: Optional[List[str]] = None
    images: Optional[List[Dict[str, str]]] = None
    is_published: Optional[bool] = None
    is_featured: Optional[bool] = None


class ProductResponse(ProductBase):
    """Product response schema."""
    id: int
    sku: str
    stock_quantity: int
    is_published: bool
    is_featured: bool
    is_in_stock: bool
    is_low_stock: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ProductListResponse(BaseModel):
    """Product list with pagination."""
    items: List[ProductResponse]
    total: int
    page: int
    page_size: int
    total_pages: int
