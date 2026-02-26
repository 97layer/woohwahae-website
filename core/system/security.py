"""
LAYER OS Shared Security Module
admin/app.py 보안 패턴 추출 — Flask CMS + FastAPI CMS 공유.
"""

import html
import logging
import os
import secrets
import time
from collections import defaultdict
from pathlib import Path
from typing import Optional

from werkzeug.security import check_password_hash, generate_password_hash


# ─── 환경변수 강제 ────────────────────────────────────────────

def require_env(name: str) -> str:
    """환경변수 필수 로드. 없으면 RuntimeError."""
    value = os.getenv(name)
    if not value:
        raise RuntimeError("%s 환경변수 필수. .env에 설정하세요." % name)
    return value


# ─── 비밀번호 해싱 ────────────────────────────────────────────

def hash_password(plain: str) -> str:
    """PBKDF2-SHA256 해싱."""
    return generate_password_hash(plain, method='pbkdf2:sha256')


def verify_password(plain: str, hashed: str) -> bool:
    """해시 비교. 타이밍 공격 방어는 werkzeug 내장."""
    return check_password_hash(hashed, plain)


# ─── 토큰 생성 ────────────────────────────────────────────────

def generate_token(nbytes: int = 32) -> str:
    """secrets 기반 암호학적 안전 토큰. uuid4 대체."""
    return secrets.token_hex(nbytes)


# ─── Rate Limiter ─────────────────────────────────────────────

class RateLimiter:
    """슬라이딩 윈도우 인메모리 레이트 리미터 (IP별)."""

    def __init__(self, max_requests: int = 10, window_seconds: int = 60):
        self.max_requests = max_requests
        self.window_seconds = window_seconds
        self._attempts: dict[str, list[float]] = defaultdict(list)

    def is_limited(self, key: str) -> bool:
        """True면 제한 중. 호출 시 자동 기록."""
        now = time.time()
        cutoff = now - self.window_seconds
        # 만료된 기록 제거
        self._attempts[key] = [t for t in self._attempts[key] if t > cutoff]
        if len(self._attempts[key]) >= self.max_requests:
            return True
        self._attempts[key].append(now)
        return False

    def reset(self, key: str) -> None:
        """특정 키의 기록 초기화."""
        self._attempts.pop(key, None)


# ─── 보안 헤더 (Flask) ────────────────────────────────────────

def apply_security_headers(response):
    """Flask @after_request 핸들러로 등록."""
    response.headers['Content-Security-Policy'] = (
        "default-src 'self'; "
        "script-src 'self' 'unsafe-inline'; "
        "style-src 'self' 'unsafe-inline'; "
        "img-src 'self' data: blob:; "
        "font-src 'self' data:; "
        "connect-src 'self'; "
        "frame-ancestors 'none'; "
        "base-uri 'self'; "
        "form-action 'self'"
    )
    response.headers['X-Frame-Options'] = 'DENY'
    response.headers['X-Content-Type-Options'] = 'nosniff'
    response.headers['Referrer-Policy'] = 'strict-origin-when-cross-origin'
    response.headers['X-XSS-Protection'] = '1; mode=block'
    response.headers['Permissions-Policy'] = (
        'camera=(), microphone=(), geolocation=(), payment=()'
    )
    response.headers['Strict-Transport-Security'] = (
        'max-age=31536000; includeSubDomains'
    )
    response.headers.pop('Server', None)
    return response


# ─── 보안 헤더 (FastAPI/Starlette) ────────────────────────────

class SecurityHeadersMiddleware:
    """Starlette ASGI 미들웨어."""

    HEADERS = {
        'Content-Security-Policy': (
            "default-src 'self'; "
            "script-src 'self' 'unsafe-inline'; "
            "style-src 'self' 'unsafe-inline'; "
            "img-src 'self' data: blob:; "
            "font-src 'self' data:; "
            "connect-src 'self'; "
            "frame-ancestors 'none'; "
            "base-uri 'self'; "
            "form-action 'self'"
        ),
        'X-Frame-Options': 'DENY',
        'X-Content-Type-Options': 'nosniff',
        'Referrer-Policy': 'strict-origin-when-cross-origin',
        'X-XSS-Protection': '1; mode=block',
        'Permissions-Policy': 'camera=(), microphone=(), geolocation=(), payment=()',
        'Strict-Transport-Security': 'max-age=31536000; includeSubDomains',
    }

    def __init__(self, app):
        self.app = app

    async def __call__(self, scope, receive, send):
        if scope['type'] != 'http':
            await self.app(scope, receive, send)
            return

        async def send_with_headers(message):
            if message['type'] == 'http.response.start':
                headers = dict(message.get('headers', []))
                for key, value in self.HEADERS.items():
                    headers[key.lower().encode()] = value.encode()
                message['headers'] = list(headers.items())
            await send(message)

        await self.app(scope, receive, send_with_headers)


# ─── HTML 이스케이프 ──────────────────────────────────────────

def sanitize_html_field(value: str) -> str:
    """사용자 입력 HTML 이스케이프. XSS 방어."""
    if not isinstance(value, str):
        return value
    return html.escape(value)


# ─── 경로 방어 ────────────────────────────────────────────────

def safe_path(user_input: str, base_dir: Path) -> Optional[Path]:
    """경로 순회 방어. base_dir 외부 접근 시 None 반환."""
    try:
        resolved = (base_dir / user_input).resolve()
        base_resolved = base_dir.resolve()
        if not str(resolved).startswith(str(base_resolved)):
            return None
        return resolved
    except (ValueError, OSError):
        return None


# ─── 감사 로거 ────────────────────────────────────────────────

def setup_audit_logger(
    name: str,
    log_file: Path,
) -> logging.Logger:
    """파일 기반 감사 로거 생성."""
    logger = logging.getLogger(name)
    if not logger.handlers:
        try:
            log_file.parent.mkdir(parents=True, exist_ok=True)
            handler = logging.FileHandler(str(log_file))
            handler.setFormatter(logging.Formatter('%(asctime)s %(message)s'))
            logger.addHandler(handler)
            logger.setLevel(logging.INFO)
        except Exception:
            pass  # 로그 파일 생성 실패해도 서버 중단 방지
    return logger
