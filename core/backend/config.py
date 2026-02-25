"""
WOOHWAHAE CMS Backend Configuration
Simple Flask-based backend for content management
"""

import os
from pathlib import Path

# Base paths
BASE_DIR = Path(__file__).resolve().parent.parent
ARCHIVE_DIR = BASE_DIR / "archive"
ARCHIVE_JSON = ARCHIVE_DIR / "index.json"
UPLOADS_DIR = BASE_DIR / "assets" / "uploads"

# Ensure directories exist
UPLOADS_DIR.mkdir(parents=True, exist_ok=True)

# Flask config
SECRET_KEY = os.environ.get('FLASK_SECRET_KEY', 'dev-secret-key-change-in-production')
DEBUG = os.environ.get('FLASK_DEBUG', 'True') == 'True'

# Authentication (simple password-based for now)
ADMIN_PASSWORD = os.environ.get('ADMIN_PASSWORD', 'woohwahae2026')

# Content settings
MAX_CONTENT_LENGTH = 16 * 1024 * 1024  # 16MB max upload
ALLOWED_EXTENSIONS = {'png', 'jpg', 'jpeg', 'gif', 'webp', 'svg'}

# CORS settings
CORS_ORIGINS = ['http://localhost:3000', 'https://woohwahae.kr']
