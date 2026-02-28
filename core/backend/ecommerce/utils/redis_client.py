"""Redis client for cart session management."""
import json
from typing import Optional, List, Dict, Any
from redis import Redis

# Redis configuration
REDIS_HOST = "localhost"
REDIS_PORT = 6379
REDIS_DB = 0
REDIS_PASSWORD = None

# Cart session TTL (7 days)
CART_TTL = 60 * 60 * 24 * 7


class RedisClient:
    """Redis client wrapper for cart operations."""

    def __init__(self):
        """Initialize Redis connection."""
        self.client = Redis(
            host=REDIS_HOST,
            port=REDIS_PORT,
            db=REDIS_DB,
            password=REDIS_PASSWORD,
            decode_responses=True
        )

    def get_cart(self, session_id: str) -> Optional[List[Dict[str, Any]]]:
        """Get cart items for session."""
        key = f"cart:{session_id}"
        data = self.client.get(key)
        if data:
            return json.loads(data)
        return []

    def set_cart(self, session_id: str, items: List[Dict[str, Any]]) -> bool:
        """Set cart items for session."""
        key = f"cart:{session_id}"
        self.client.setex(key, CART_TTL, json.dumps(items))
        return True

    def add_item(self, session_id: str, product_id: int, quantity: int) -> bool:
        """Add or update item in cart."""
        items = self.get_cart(session_id) or []

        # Check if item already exists
        found = False
        for item in items:
            if item["product_id"] == product_id:
                item["quantity"] += quantity
                found = True
                break

        if not found:
            items.append({"product_id": product_id, "quantity": quantity})

        return self.set_cart(session_id, items)

    def update_item(self, session_id: str, product_id: int, quantity: int) -> bool:
        """Update item quantity in cart."""
        items = self.get_cart(session_id) or []

        if quantity <= 0:
            # Remove item if quantity is 0
            items = [item for item in items if item["product_id"] != product_id]
        else:
            # Update quantity
            found = False
            for item in items:
                if item["product_id"] == product_id:
                    item["quantity"] = quantity
                    found = True
                    break

            if not found:
                return False

        return self.set_cart(session_id, items)

    def remove_item(self, session_id: str, product_id: int) -> bool:
        """Remove item from cart."""
        items = self.get_cart(session_id) or []
        items = [item for item in items if item["product_id"] != product_id]
        return self.set_cart(session_id, items)

    def clear_cart(self, session_id: str) -> bool:
        """Clear all items from cart."""
        key = f"cart:{session_id}"
        self.client.delete(key)
        return True

    def get_cart_count(self, session_id: str) -> int:
        """Get total number of items in cart."""
        items = self.get_cart(session_id) or []
        return sum(item["quantity"] for item in items)


# Global Redis client instance
redis_client = RedisClient()
