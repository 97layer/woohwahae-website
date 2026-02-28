"""Application configuration.

Environment variables should be set in production:
- DATABASE_URL: PostgreSQL connection string
- REDIS_URL: Redis connection string
- JWT_SECRET_KEY: Secret key for JWT token signing
- STRIPE_API_KEY: Stripe API key
- TOSSPAYMENTS_SECRET_KEY: TossPayments secret key
- TOSSPAYMENTS_CLIENT_KEY: TossPayments client key
"""
import os
from typing import Optional


class Settings:
    """Application settings."""

    # Database
    DATABASE_URL: str = os.getenv(
        "DATABASE_URL",
        "postgresql://user:password@localhost:5432/woohwahae_ecommerce"
    )

    # Redis
    REDIS_HOST: str = os.getenv("REDIS_HOST", "localhost")
    REDIS_PORT: int = int(os.getenv("REDIS_PORT", "6379"))
    REDIS_DB: int = int(os.getenv("REDIS_DB", "0"))
    REDIS_PASSWORD: Optional[str] = os.getenv("REDIS_PASSWORD", None)

    # JWT
    JWT_SECRET_KEY: str = os.getenv(
        "JWT_SECRET_KEY",
        "your-secret-key-here-replace-in-production"
    )
    JWT_ALGORITHM: str = "HS256"
    JWT_EXPIRE_MINUTES: int = int(os.getenv("JWT_EXPIRE_MINUTES", str(60 * 24 * 7)))

    # Payment Gateways
    STRIPE_API_KEY: str = os.getenv("STRIPE_API_KEY", "sk_test_stub")
    TOSSPAYMENTS_SECRET_KEY: str = os.getenv("TOSSPAYMENTS_SECRET_KEY", "test_sk_stub")
    TOSSPAYMENTS_CLIENT_KEY: str = os.getenv("TOSSPAYMENTS_CLIENT_KEY", "test_ck_stub")

    # Application
    APP_NAME: str = "WOOHWAHAE E-commerce"
    APP_VERSION: str = "1.0.0"
    DEBUG: bool = os.getenv("DEBUG", "False").lower() == "true"

    # CORS
    CORS_ORIGINS: list = [
        "https://woohwahae.kr",
        "https://www.woohwahae.kr",
        "http://localhost:3000",
        "http://localhost:5173",
    ]


settings = Settings()
