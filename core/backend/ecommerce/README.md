# WOOHWAHAE E-commerce Backend API

FastAPI-based REST API for WOOHWAHAE product catalog, shopping cart, and order management.

## Architecture

```
core/backend/ecommerce/
├── api/              # FastAPI route handlers
│   ├── auth.py       # Authentication endpoints
│   ├── products.py   # Product catalog endpoints
│   ├── cart.py       # Shopping cart endpoints
│   └── orders.py     # Order management endpoints
├── models/           # SQLAlchemy database models
│   ├── user.py       # User and Address models
│   ├── product.py    # Product catalog model
│   ├── order.py      # Order and OrderItem models
│   └── base.py       # Database session and base
├── schemas/          # Pydantic validation schemas
│   ├── user.py       # User request/response schemas
│   ├── product.py    # Product schemas
│   ├── order.py      # Order schemas
│   └── cart.py       # Cart schemas
├── services/         # Business logic and integrations
│   └── payment.py    # Payment gateway integrations
├── utils/            # Utilities and helpers
│   ├── auth.py       # JWT authentication utilities
│   └── redis_client.py  # Redis cart session manager
├── main.py           # FastAPI application entry point
├── config.py         # Configuration management
└── requirements.txt  # Python dependencies
```

## Database Schema

### Users
- `id`, `email`, `phone`, `full_name`, `hashed_password`
- `is_active`, `is_verified`, `is_admin`
- `created_at`, `updated_at`

### Addresses
- `id`, `user_id`, `recipient_name`, `phone`
- `postal_code`, `address_line1`, `address_line2`, `city`
- `is_default`, `created_at`

### Products
- `id`, `sku`, `name`, `slug`, `description`
- `price`, `compare_at_price`, `cost_price`
- `stock_quantity`, `low_stock_threshold`, `track_inventory`
- `category`, `tags`, `images`
- `is_published`, `is_featured`
- `created_at`, `updated_at`

### Orders
- `id`, `order_number`, `user_id`
- `status`, `payment_status`
- `subtotal`, `shipping_fee`, `discount_amount`, `total`
- `shipping_address`, `shipping_method`, `tracking_number`
- `payment_method`, `payment_transaction_id`, `payment_gateway`
- `created_at`, `paid_at`, `shipped_at`, `delivered_at`

### OrderItems
- `id`, `order_id`, `product_id`
- `product_name`, `product_sku`, `product_image`
- `unit_price`, `quantity`, `subtotal`

## API Endpoints

### Authentication (`/api/v1/auth`)
- `POST /register` - Register new user
- `POST /login` - Login and get JWT token
- `GET /me` - Get current user info
- `POST /refresh` - Refresh access token

### Products (`/api/v1/products`)
- `GET /` - List products (pagination, filters)
- `GET /{product_id}` - Get product by ID
- `GET /slug/{slug}` - Get product by slug
- `POST /` - Create product (admin)
- `PATCH /{product_id}` - Update product (admin)
- `DELETE /{product_id}` - Delete product (admin)

### Cart (`/api/v1/cart`)
- `GET /` - Get cart contents
- `POST /items` - Add item to cart
- `PATCH /items/{product_id}` - Update item quantity
- `DELETE /items/{product_id}` - Remove item from cart
- `DELETE /` - Clear cart

**Note:** Cart endpoints require `X-Session-ID` header.

### Orders (`/api/v1/orders`)
- `POST /` - Create order from cart
- `GET /` - List user orders (pagination)
- `GET /{order_id}` - Get order details
- `PATCH /{order_id}` - Update order status (admin)
- `DELETE /{order_id}` - Cancel order (user)

## Setup

### 1. Install Dependencies

```bash
cd /Users/97layer/97layerOS/core/backend/ecommerce
pip install -r requirements.txt
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 3. Setup PostgreSQL

```bash
# Create database
createdb woohwahae_ecommerce

# Or using psql
psql -U postgres
CREATE DATABASE woohwahae_ecommerce;
```

### 4. Setup Redis

```bash
# Install Redis (macOS)
brew install redis

# Start Redis
redis-server

# Or using Docker
docker run -d -p 6379:6379 redis:7-alpine
```

### 5. Initialize Database

```bash
python -c "from models import init_db; init_db()"

# Or use Alembic for migrations
alembic init migrations
alembic revision --autogenerate -m "Initial migration"
alembic upgrade head
```

### 6. Run Development Server

```bash
uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

API documentation available at:
- Swagger UI: http://localhost:8000/api/docs
- ReDoc: http://localhost:8000/api/redoc

## Production Deployment

### Environment Variables

Set all secrets in production environment:

```bash
DATABASE_URL=postgresql://user:pass@host:5432/db
REDIS_HOST=redis.example.com
JWT_SECRET_KEY=long-random-secret-key
STRIPE_API_KEY=sk_live_...
TOSSPAYMENTS_SECRET_KEY=live_sk_...
TOSSPAYMENTS_CLIENT_KEY=live_ck_...
DEBUG=False
```

### Run with Gunicorn

```bash
pip install gunicorn
gunicorn main:app -w 4 -k uvicorn.workers.UvicornWorker -b 0.0.0.0:8000
```

### Systemd Service

```ini
[Unit]
Description=WOOHWAHAE E-commerce API
After=network.target

[Service]
User=www-data
WorkingDirectory=/path/to/97layerOS/core/backend/ecommerce
Environment="PATH=/path/to/venv/bin"
ExecStart=/path/to/venv/bin/gunicorn main:app -w 4 -k uvicorn.workers.UvicornWorker -b 0.0.0.0:8000

[Install]
WantedBy=multi-user.target
```

## Payment Gateway Integration

### Stripe

1. Get API keys from https://dashboard.stripe.com/apikeys
2. Set `STRIPE_API_KEY` environment variable
3. Uncomment `stripe` in requirements.txt
4. Implement webhooks for payment confirmation

### TossPayments

1. Get credentials from https://developers.tosspayments.com
2. Set `TOSSPAYMENTS_SECRET_KEY` and `TOSSPAYMENTS_CLIENT_KEY`
3. Uncomment `requests` in requirements.txt (if not already installed)
4. Implement payment confirmation callback

## Security Considerations

1. **JWT Secret**: Use strong random secret in production
2. **Database**: Use environment variables, never hardcode credentials
3. **CORS**: Restrict origins to actual domains in production
4. **HTTPS**: Always use HTTPS in production
5. **Rate Limiting**: Implement rate limiting for API endpoints
6. **Input Validation**: Pydantic schemas provide automatic validation
7. **Password Hashing**: bcrypt used for secure password storage

## Testing

```bash
# Run tests
pytest

# With coverage
pytest --cov=. --cov-report=html
```

## Next Steps

1. Implement database migrations with Alembic
2. Complete payment gateway integrations
3. Add email notifications for orders
4. Implement inventory management
5. Add product reviews and ratings
6. Implement discount codes and promotions
7. Add admin dashboard endpoints
8. Set up monitoring and logging
