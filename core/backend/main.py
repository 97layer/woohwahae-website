#!/usr/bin/env python3
"""
WOOHWAHAE CMS Backend
FastAPI 인라인 텍스트 편집 시스템 — 보안 강화 버전
"""

import os
import sys
import time
from pathlib import Path
from datetime import datetime
from typing import Optional

from fastapi import FastAPI, HTTPException, Depends, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from fastapi.responses import HTMLResponse, JSONResponse
from sqlalchemy import create_engine, Column, String, Text, DateTime, Integer
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, Session
from pydantic import BaseModel, Field

# LAYER OS core modules
_PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(_PROJECT_ROOT))

from core.system.security import (
    require_env,
    verify_password,
    generate_token,
    sanitize_html_field,
    SecurityHeadersMiddleware,
    setup_audit_logger,
    RateLimiter,
)

# Load .env
try:
    from dotenv import load_dotenv
    load_dotenv(_PROJECT_ROOT / '.env')
except ImportError:
    pass

# ─── 환경변수 (fallback 없음) ─────────────────────────────────

ADMIN_PASSWORD_HASH = require_env('FASTAPI_ADMIN_PASSWORD_HASH')

# ─── FastAPI 앱 ───────────────────────────────────────────────

app = FastAPI(
    title="WOOHWAHAE CMS",
    version="1.0.0",
    docs_url=None,   # Swagger UI 비활성화 (프로덕션)
    redoc_url=None,   # ReDoc 비활성화
)

# 보안 헤더 미들웨어 (CORS보다 먼저 등록 = 나중에 실행)
app.add_middleware(SecurityHeadersMiddleware)

# CORS — 허용 오리진 명시, 메서드/헤더 제한
_cors_origins = os.getenv('FASTAPI_CORS_ORIGINS', 'https://woohwahae.kr').split(',')
app.add_middleware(
    CORSMiddleware,
    allow_origins=_cors_origins,
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["Content-Type", "Authorization"],
)

# ─── 감사 로거 ────────────────────────────────────────────────

_audit_log_path = _PROJECT_ROOT / 'knowledge' / 'reports' / 'fastapi_audit.log'
_audit_logger = setup_audit_logger('fastapi_audit', _audit_log_path)

# 로그인 레이트 리미터 (5회/분)
_login_limiter = RateLimiter(max_requests=5, window_seconds=60)


def _audit(request: Request, action: str, detail: str = '') -> None:
    ip = request.client.host if request.client else 'unknown'
    _audit_logger.info("ip=%s action=%s detail=%s", ip, action, detail)


# ─── 데이터베이스 ─────────────────────────────────────────────

DATABASE_URL = "sqlite:///./woohwahae_cms.db"
engine = create_engine(DATABASE_URL, connect_args={"check_same_thread": False})
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
Base = declarative_base()


class Content(Base):
    __tablename__ = "contents"
    id = Column(Integer, primary_key=True, index=True)
    page = Column(String(100), index=True)
    element_id = Column(String(200), index=True)
    content = Column(Text)
    updated_at = Column(DateTime, default=datetime.utcnow)


Base.metadata.create_all(bind=engine)


# ─── Pydantic 모델 (길이 제한) ────────────────────────────────

class ContentUpdate(BaseModel):
    page: str = Field(..., max_length=100)
    element_id: str = Field(..., max_length=200)
    content: str = Field(..., max_length=50000)


class ContentResponse(BaseModel):
    id: int
    page: str
    element_id: str
    content: str
    updated_at: datetime


class AdminAuth(BaseModel):
    password: str = Field(..., max_length=200)


# ─── 세션/토큰 ────────────────────────────────────────────────

# {token: {'created': timestamp}} — secrets 기반, 1시간 만료
_admin_sessions: dict[str, dict] = {}
SESSION_TTL = 3600  # 1시간


def _cleanup_expired_sessions() -> None:
    now = time.time()
    expired = [t for t, v in _admin_sessions.items() if now - v['created'] > SESSION_TTL]
    for t in expired:
        del _admin_sessions[t]


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


def check_admin(request: Request, token: str = None):
    """관리자 권한 확인 — 토큰 유효성 + 만료 검증."""
    _cleanup_expired_sessions()
    if not token or token not in _admin_sessions:
        raise HTTPException(status_code=401, detail="Unauthorized")
    return True


# ─── 인증 엔드포인트 ──────────────────────────────────────────

@app.post("/api/admin/login")
async def admin_login(auth: AdminAuth, request: Request):
    ip = request.client.host if request.client else 'unknown'

    if _login_limiter.is_limited(ip):
        _audit(request, 'login_rate_limited', ip)
        raise HTTPException(status_code=429, detail="Too many attempts. Try later.")

    if not verify_password(auth.password, ADMIN_PASSWORD_HASH):
        _audit(request, 'login_failed', ip)
        raise HTTPException(status_code=401, detail="Invalid password")

    token = generate_token(32)
    _admin_sessions[token] = {'created': time.time()}
    _audit(request, 'login_success', ip)
    return {"token": token, "message": "Login successful"}


# ─── API 엔드포인트 ───────────────────────────────────────────

@app.get("/api/contents/{page}")
async def get_page_contents(page: str, db: Session = Depends(get_db)):
    if len(page) > 100:
        raise HTTPException(status_code=400, detail="Page name too long")
    contents = db.query(Content).filter(Content.page == page).all()
    return contents


@app.get("/api/content/{page}/{element_id}")
async def get_content(page: str, element_id: str, db: Session = Depends(get_db)):
    if len(page) > 100 or len(element_id) > 200:
        raise HTTPException(status_code=400, detail="Input too long")
    content = db.query(Content).filter(
        Content.page == page,
        Content.element_id == element_id
    ).first()

    if not content:
        return {"content": None}

    return {
        "content": content.content,
        "updated_at": content.updated_at
    }


@app.post("/api/content/update")
async def update_content(
    data: ContentUpdate,
    request: Request,
    token: str = None,
    db: Session = Depends(get_db)
):
    check_admin(request, token)

    # XSS 방어
    safe_content = sanitize_html_field(data.content)
    safe_page = sanitize_html_field(data.page)
    safe_element = sanitize_html_field(data.element_id)

    content = db.query(Content).filter(
        Content.page == safe_page,
        Content.element_id == safe_element
    ).first()

    if content:
        content.content = safe_content
        content.updated_at = datetime.utcnow()
    else:
        content = Content(
            page=safe_page,
            element_id=safe_element,
            content=safe_content
        )
        db.add(content)

    db.commit()
    db.refresh(content)

    _audit(request, 'content_update', "%s/%s" % (safe_page, safe_element))
    return {
        "success": True,
        "content": content.content,
        "updated_at": content.updated_at
    }


@app.get("/api/pages")
async def get_all_pages(db: Session = Depends(get_db)):
    pages = db.query(Content.page).distinct().all()
    return [p[0] for p in pages]


# ─── 관리자 패널 ──────────────────────────────────────────────

@app.get("/admin", response_class=HTMLResponse)
async def admin_panel():
    return """
<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>WOOHWAHAE CMS Admin</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Helvetica Neue", sans-serif; background: #f5f5f5; padding: 40px; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { font-size: 28px; margin-bottom: 30px; color: #333; }
        .login-form { max-width: 400px; margin: 100px auto; }
        input[type="password"] { width: 100%; padding: 12px; font-size: 16px; border: 1px solid #ddd; border-radius: 4px; margin-bottom: 20px; }
        button { background: #000; color: white; padding: 12px 24px; font-size: 16px; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #333; }
        .admin-panel { display: none; }
        .admin-panel.active { display: block; }
        .page-selector { margin-bottom: 30px; }
        select { padding: 10px; font-size: 16px; border: 1px solid #ddd; border-radius: 4px; margin-right: 10px; }
        .content-editor { margin-top: 30px; }
        .editable-item { background: #f9f9f9; padding: 20px; margin-bottom: 20px; border-radius: 4px; border: 1px solid #e0e0e0; }
        .editable-item h3 { font-size: 14px; color: #666; margin-bottom: 10px; text-transform: uppercase; letter-spacing: 1px; }
        .editable-item textarea { width: 100%; padding: 12px; font-size: 15px; border: 1px solid #ddd; border-radius: 4px; min-height: 100px; font-family: inherit; resize: vertical; }
        .save-btn { background: #4CAF50; margin-top: 10px; }
        .status { margin-left: 10px; padding: 5px 10px; border-radius: 3px; font-size: 14px; display: inline-block; }
        .status.success { background: #d4edda; color: #155724; }
        .status.error { background: #f8d7da; color: #721c24; }
    </style>
</head>
<body>
    <div class="container">
        <h1>WOOHWAHAE CMS</h1>
        <div id="loginForm" class="login-form">
            <h2>관리자 로그인</h2>
            <input type="password" id="password" placeholder="비밀번호 입력">
            <button onclick="login()">로그인</button>
        </div>
        <div id="adminPanel" class="admin-panel">
            <div class="page-selector">
                <label>페이지 선택:</label>
                <select id="pageSelect" onchange="loadPageContent()">
                    <option value="">페이지를 선택하세요</option>
                    <option value="index">메인 페이지</option>
                    <option value="about">About</option>
                    <option value="practice">Practice</option>
                </select>
                <button onclick="openPage()">페이지 미리보기</button>
            </div>
            <div id="contentEditor" class="content-editor"></div>
        </div>
    </div>
    <script>
        let adminToken = sessionStorage.getItem('adminToken');
        if (adminToken) showAdminPanel();

        async function login() {
            const password = document.getElementById('password').value;
            try {
                const response = await fetch('/api/admin/login', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({password})
                });
                if (response.ok) {
                    const data = await response.json();
                    adminToken = data.token;
                    sessionStorage.setItem('adminToken', adminToken);
                    showAdminPanel();
                } else if (response.status === 429) {
                    alert('너무 많은 시도. 잠시 후 다시 시도하세요.');
                } else {
                    alert('비밀번호가 틀렸습니다.');
                }
            } catch (error) {
                alert('로그인 중 오류가 발생했습니다.');
            }
        }

        function showAdminPanel() {
            document.getElementById('loginForm').style.display = 'none';
            document.getElementById('adminPanel').classList.add('active');
        }

        async function loadPageContent() {
            const page = document.getElementById('pageSelect').value;
            if (!page) return;
            const editor = document.getElementById('contentEditor');
            editor.innerHTML = '<p>로딩 중...</p>';
            try {
                const response = await fetch('/api/contents/' + encodeURIComponent(page));
                const contents = await response.json();
                if (contents.length === 0) {
                    editor.innerHTML = '<p>아직 편집 가능한 콘텐츠가 없습니다.</p>';
                } else {
                    editor.innerHTML = contents.map(function(item) {
                        return '<div class="editable-item">' +
                            '<h3>' + escapeHtml(item.element_id) + '</h3>' +
                            '<textarea id="content-' + item.id + '" data-page="' + escapeHtml(item.page) + '" data-element="' + escapeHtml(item.element_id) + '">' + escapeHtml(item.content) + '</textarea>' +
                            '<button class="save-btn" onclick="saveContent(' + item.id + ')">저장</button>' +
                            '<span id="status-' + item.id + '" class="status"></span></div>';
                    }).join('');
                }
            } catch (error) {
                editor.innerHTML = '<p>콘텐츠를 불러오는데 실패했습니다.</p>';
            }
        }

        function escapeHtml(text) {
            var div = document.createElement('div');
            div.appendChild(document.createTextNode(text));
            return div.innerHTML;
        }

        async function saveContent(id) {
            const textarea = document.getElementById('content-' + id);
            const page = textarea.dataset.page;
            const elementId = textarea.dataset.element;
            const content = textarea.value;
            const status = document.getElementById('status-' + id);
            try {
                const response = await fetch('/api/content/update?token=' + encodeURIComponent(adminToken), {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({ page: page, element_id: elementId, content: content })
                });
                if (response.ok) {
                    status.className = 'status success';
                    status.textContent = '저장 완료!';
                    setTimeout(function() { status.textContent = ''; }, 3000);
                } else { throw new Error('저장 실패'); }
            } catch (error) {
                status.className = 'status error';
                status.textContent = '저장 실패';
            }
        }

        function openPage() {
            const page = document.getElementById('pageSelect').value;
            if (!page) return;
            const url = page === 'index' ? '/' : '/' + page + '/';
            window.open(url, '_blank');
        }
    </script>
</body>
</html>
    """


# ─── 에러 핸들러 (디버그 정보 노출 방지) ──────────────────────

@app.exception_handler(500)
async def internal_error_handler(request: Request, exc):
    return JSONResponse(status_code=500, content={"error": "Internal server error"})


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="127.0.0.1", port=8082)
