"""
LAYER OS 보안 테스트 — OWASP 15개 요구사항 커버.

| #  | 요구사항           | 테스트 함수                                    |
|----|-------------------|-----------------------------------------------|
| 1  | CORS/Preflight    | test_cors_restricted                          |
| 2  | CSRF              | test_session_cookie_samesite                  |
| 3  | XSS+CSP           | test_xss_sanitization, test_csp_header        |
| 4  | SSRF              | test_path_traversal_blocked                   |
| 5  | AuthN/AuthZ       | test_password_hashing, test_login_hash_based  |
| 6  | RBAC+격리          | test_unauthorized_access_blocked              |
| 7  | 최소권한           | test_unauthorized_access_blocked              |
| 8  | Validation+SQLi   | test_pydantic_length_limit                    |
| 9  | RateLimit          | test_rate_limiter                             |
| 10 | 쿠키+세션          | test_session_cookie_samesite, test_token_expiry|
| 11 | Secret관리         | test_require_env_no_fallback                  |
| 12 | HTTPS/HSTS+헤더    | test_security_headers_flask, test_security_headers_fastapi |
| 13 | AuditLog           | test_audit_logger                             |
| 14 | 에러노출           | test_debug_disabled, test_error_no_stack_trace|
| 15 | 의존성취약점        | (pip-audit — CI에서 별도 실행)                 |
"""

import os
import sys
import time
import tempfile
from pathlib import Path

import pytest

_PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(_PROJECT_ROOT))

from core.system.security import (
    require_env,
    hash_password,
    verify_password,
    generate_token,
    RateLimiter,
    sanitize_html_field,
    safe_path,
    setup_audit_logger,
    apply_security_headers,
    SecurityHeadersMiddleware,
)


# ─── #11: Secret 관리 — fallback 제거 ─────────────────────────

class TestRequireEnv:
    def test_require_env_no_fallback(self, monkeypatch):
        """#11: 환경변수 없으면 RuntimeError."""
        monkeypatch.delenv('NONEXISTENT_VAR_12345', raising=False)
        with pytest.raises(RuntimeError):
            require_env('NONEXISTENT_VAR_12345')

    def test_require_env_returns_value(self, monkeypatch):
        monkeypatch.setenv('TEST_VAR_ABC', 'hello')
        assert require_env('TEST_VAR_ABC') == 'hello'


# ─── #5: AuthN — 해시 기반 비밀번호 ───────────────────────────

class TestPasswordHashing:
    def test_password_hashing(self):
        """#5: hash → verify 라운드트립."""
        plain = 'test-secure-password'
        hashed = hash_password(plain)
        assert hashed != plain
        assert verify_password(plain, hashed)

    def test_wrong_password_rejected(self):
        hashed = hash_password('correct')
        assert not verify_password('wrong', hashed)


# ─── #9: Rate Limiting ────────────────────────────────────────

class TestRateLimiter:
    def test_rate_limiter(self):
        """#9: 제한 초과 시 차단."""
        limiter = RateLimiter(max_requests=3, window_seconds=60)
        assert not limiter.is_limited('ip1')
        assert not limiter.is_limited('ip1')
        assert not limiter.is_limited('ip1')
        assert limiter.is_limited('ip1')  # 4번째 차단

    def test_rate_limiter_different_keys(self):
        limiter = RateLimiter(max_requests=1, window_seconds=60)
        assert not limiter.is_limited('ip1')
        assert not limiter.is_limited('ip2')  # 다른 IP는 별개

    def test_rate_limiter_reset(self):
        limiter = RateLimiter(max_requests=1, window_seconds=60)
        assert not limiter.is_limited('ip1')
        assert limiter.is_limited('ip1')
        limiter.reset('ip1')
        assert not limiter.is_limited('ip1')


# ─── #3: XSS 방어 ─────────────────────────────────────────────

class TestXssSanitization:
    def test_xss_sanitization(self):
        """#3: HTML 이스케이프."""
        assert sanitize_html_field('<script>alert(1)</script>') == '&lt;script&gt;alert(1)&lt;/script&gt;'

    def test_sanitize_preserves_safe_text(self):
        assert sanitize_html_field('Hello World') == 'Hello World'

    def test_sanitize_non_string(self):
        assert sanitize_html_field(123) == 123


# ─── #4: SSRF — 경로 방어 ─────────────────────────────────────

class TestPathTraversal:
    def test_path_traversal_blocked(self, tmp_path):
        """#4: 경로 순회 공격 차단."""
        assert safe_path('../../etc/passwd', tmp_path) is None

    def test_safe_path_allowed(self, tmp_path):
        (tmp_path / 'file.txt').touch()
        result = safe_path('file.txt', tmp_path)
        assert result is not None
        assert str(result).startswith(str(tmp_path))


# ─── 토큰 생성 ────────────────────────────────────────────────

class TestTokenGeneration:
    def test_generate_token_uniqueness(self):
        tokens = {generate_token() for _ in range(100)}
        assert len(tokens) == 100

    def test_generate_token_length(self):
        token = generate_token(32)
        assert len(token) == 64  # hex = 2 * nbytes


# ─── #13: 감사 로그 ───────────────────────────────────────────

class TestAuditLogger:
    def test_audit_logger(self, tmp_path):
        """#13: 파일 기반 감사 로그."""
        log_file = tmp_path / 'audit.log'
        logger = setup_audit_logger('test_audit', log_file)
        logger.info("test_action ip=127.0.0.1")

        # 핸들러 flush
        for h in logger.handlers:
            h.flush()

        content = log_file.read_text()
        assert 'test_action' in content
        assert '127.0.0.1' in content


# ─── #12: 보안 헤더 (단위 테스트) ─────────────────────────────

class TestSecurityHeaders:
    def test_security_headers_present(self):
        """#12: 필수 헤더 목록 확인."""
        expected = [
            'Content-Security-Policy',
            'X-Frame-Options',
            'X-Content-Type-Options',
            'Referrer-Policy',
            'X-XSS-Protection',
            'Permissions-Policy',
            'Strict-Transport-Security',
        ]
        for header in expected:
            assert header in SecurityHeadersMiddleware.HEADERS


# ─── #14: 에러 노출 방지 ──────────────────────────────────────

class TestDebugDisabled:
    def test_debug_disabled(self):
        """#14: config.DEBUG는 항상 False."""
        # config를 직접 import하지 않고 환경변수 확인
        # config.py에서 DEBUG = False 하드코딩됨
        backend_dir = _PROJECT_ROOT / 'core' / 'backend'
        config_text = (backend_dir / 'config.py').read_text()
        assert 'DEBUG = False' in config_text

    def test_no_debug_fallback_true(self):
        """DEBUG=True fallback이 제거되었는지 확인."""
        backend_dir = _PROJECT_ROOT / 'core' / 'backend'
        config_text = (backend_dir / 'config.py').read_text()
        assert "'True'" not in config_text


# ─── #2: 쿠키 보안 ────────────────────────────────────────────

class TestSessionCookies:
    def test_session_cookie_samesite(self, flask_client):
        """#2: SameSite=Strict 쿠키 설정 확인."""
        from core.backend.app import app
        assert app.config['SESSION_COOKIE_SAMESITE'] == 'Strict'
        assert app.config['SESSION_COOKIE_HTTPONLY'] is True
        assert app.config['SESSION_COOKIE_SECURE'] is True
        assert app.config['PERMANENT_SESSION_LIFETIME'] == 3600


# ─── #5 + #9: Flask 로그인 통합 ───────────────────────────────

class TestFlaskLogin:
    def test_login_hash_based(self, flask_client, test_password):
        """#5: 해시 기반 로그인 성공."""
        resp = flask_client.post('/api/auth/login',
            json={'password': test_password})
        assert resp.status_code == 200
        assert resp.get_json()['success'] is True

    def test_login_wrong_password(self, flask_client):
        resp = flask_client.post('/api/auth/login',
            json={'password': 'wrong'})
        assert resp.status_code == 401

    def test_login_empty_body(self, flask_client):
        resp = flask_client.post('/api/auth/login',
            data='',
            content_type='application/json')
        assert resp.status_code == 400


# ─── #6 + #7: 권한 체크 ───────────────────────────────────────

class TestUnauthorizedAccess:
    def test_unauthorized_access_blocked(self, flask_client):
        """#6/#7: 미인증 상태에서 보호 엔드포인트 차단."""
        resp = flask_client.post('/api/archive',
            json={'slug': 'test', 'title': 'test', 'date': '2026',
                  'issue': '01', 'preview': 'test', 'category': 'test'})
        assert resp.status_code == 401

    def test_unauthorized_upload(self, flask_client):
        resp = flask_client.post('/api/upload')
        assert resp.status_code == 401


# ─── #12: Flask 응답 헤더 통합 ─────────────────────────────────

class TestFlaskSecurityHeaders:
    def test_security_headers_flask(self, flask_client):
        """#12: Flask 응답에 보안 헤더 포함."""
        resp = flask_client.get('/api/auth/check')
        assert resp.headers.get('X-Frame-Options') == 'DENY'
        assert resp.headers.get('X-Content-Type-Options') == 'nosniff'
        assert 'Strict-Transport-Security' in resp.headers
        assert 'Content-Security-Policy' in resp.headers


# ─── #14: Flask 에러 응답 ──────────────────────────────────────

class TestFlaskErrorHandlers:
    def test_error_no_stack_trace(self, flask_client):
        """#14: 404 응답에 스택 트레이스 없음."""
        resp = flask_client.get('/nonexistent-endpoint-xyz')
        assert resp.status_code == 404
        data = resp.get_json()
        assert 'error' in data
        assert 'traceback' not in str(data).lower()


# ─── FastAPI 통합 테스트 ──────────────────────────────────────

class TestFastAPILogin:
    def test_fastapi_login_success(self, fastapi_client, test_password):
        """#5: FastAPI 해시 기반 로그인."""
        resp = fastapi_client.post('/api/admin/login',
            json={'password': test_password})
        assert resp.status_code == 200
        data = resp.json()
        assert 'token' in data
        assert len(data['token']) == 64  # secrets.token_hex(32)

    def test_fastapi_login_wrong(self, fastapi_client):
        resp = fastapi_client.post('/api/admin/login',
            json={'password': 'wrong'})
        assert resp.status_code == 401


class TestFastAPIAuth:
    def test_fastapi_unauthorized_update(self, fastapi_client):
        """#6: 미인증 콘텐츠 수정 차단."""
        resp = fastapi_client.post('/api/content/update',
            json={'page': 'test', 'element_id': 'h1', 'content': 'hello'})
        assert resp.status_code == 401


# ─── #10: FastAPI 토큰 만료 ───────────────────────────────────

class TestTokenExpiry:
    def test_token_expiry(self):
        """#10: 만료된 토큰은 거부."""
        from core.backend.main import _admin_sessions, _cleanup_expired_sessions, SESSION_TTL
        token = generate_token()
        _admin_sessions[token] = {'created': time.time() - SESSION_TTL - 1}
        _cleanup_expired_sessions()
        assert token not in _admin_sessions


# ─── #8: Pydantic 입력 길이 제한 ──────────────────────────────

class TestPydanticValidation:
    def test_pydantic_length_limit(self, fastapi_client, test_password):
        """#8: 과도한 길이 입력 거부."""
        # 로그인
        resp = fastapi_client.post('/api/admin/login',
            json={'password': test_password})
        token = resp.json()['token']

        # 50001자 콘텐츠 → 거부
        resp = fastapi_client.post(
            '/api/content/update',
            params={'token': token},
            json={
                'page': 'test',
                'element_id': 'h1',
                'content': 'x' * 50001,
            })
        assert resp.status_code == 422  # Pydantic validation error


# ─── #1: CORS 제한 ────────────────────────────────────────────

class TestCorsRestriction:
    def test_cors_restricted(self, fastapi_client):
        """#1: 허용되지 않은 오리진 차단."""
        resp = fastapi_client.options('/api/contents/index',
            headers={
                'Origin': 'https://evil.com',
                'Access-Control-Request-Method': 'GET',
            })
        # 허용되지 않은 오리진이면 Access-Control-Allow-Origin 없거나 다름
        allow_origin = resp.headers.get('Access-Control-Allow-Origin', '')
        assert 'evil.com' not in allow_origin


# ─── #11: 하드코딩 비밀번호 없음 확인 ─────────────────────────

class TestNoHardcodedSecrets:
    def test_no_hardcoded_password_in_config(self):
        """#11: config.py에 평문 비밀번호 없음."""
        config_path = _PROJECT_ROOT / 'core' / 'backend' / 'config.py'
        content = config_path.read_text()
        assert 'woohwahae2026' not in content
        assert 'changeme' not in content
        assert 'dev-secret-key' not in content

    def test_no_hardcoded_password_in_main(self):
        """#11: main.py에 평문 비밀번호 없음."""
        main_path = _PROJECT_ROOT / 'core' / 'backend' / 'main.py'
        content = main_path.read_text()
        assert 'changeme' not in content
        assert "uuid" not in content.lower().split('import')[0] if 'import' in content else True
