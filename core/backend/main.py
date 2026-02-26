#!/usr/bin/env python3
"""
WOOHWAHAE CMS Backend
간단한 인라인 텍스트 편집 시스템
"""

from fastapi import FastAPI, HTTPException, Depends, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from fastapi.responses import HTMLResponse, JSONResponse
from sqlalchemy import create_engine, Column, String, Text, DateTime, Integer
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, Session
from datetime import datetime
import json
import os
from typing import Optional, Dict
from pydantic import BaseModel

# FastAPI 앱 초기화
app = FastAPI(title="WOOHWAHAE CMS", version="1.0.0")

# CORS 설정 - 프론트엔드에서 API 호출 가능하도록
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:8081", "http://localhost:8000"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 데이터베이스 설정
DATABASE_URL = "sqlite:///./woohwahae_cms.db"
engine = create_engine(DATABASE_URL, connect_args={"check_same_thread": False})
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
Base = declarative_base()

# 데이터베이스 모델
class Content(Base):
    __tablename__ = "contents"

    id = Column(Integer, primary_key=True, index=True)
    page = Column(String, index=True)  # 페이지 이름 (index, about, shop 등)
    element_id = Column(String, index=True)  # HTML 요소 ID
    content = Column(Text)  # 실제 콘텐츠
    updated_at = Column(DateTime, default=datetime.utcnow)

# 테이블 생성
Base.metadata.create_all(bind=engine)

# Pydantic 모델
class ContentUpdate(BaseModel):
    page: str
    element_id: str
    content: str

class ContentResponse(BaseModel):
    id: int
    page: str
    element_id: str
    content: str
    updated_at: datetime

# 데이터베이스 세션 의존성
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

# 간단한 인증 (실제 환경에서는 JWT 사용 권장)
ADMIN_PASSWORD = os.getenv("ADMIN_PASSWORD", "changeme")

class AdminAuth(BaseModel):
    password: str

# 세션 저장소 (실제로는 Redis 사용 권장)
admin_sessions = {}

@app.post("/api/admin/login")
async def admin_login(auth: AdminAuth):
    """관리자 로그인"""
    if auth.password != ADMIN_PASSWORD:
        raise HTTPException(status_code=401, detail="Invalid password")

    # 간단한 세션 토큰 생성
    import uuid
    token = str(uuid.uuid4())
    admin_sessions[token] = datetime.utcnow()

    return {"token": token, "message": "Login successful"}

def check_admin(token: str = None):
    """관리자 권한 확인"""
    if not token or token not in admin_sessions:
        raise HTTPException(status_code=401, detail="Unauthorized")
    return True

# API 엔드포인트
@app.get("/api/contents/{page}")
async def get_page_contents(page: str, db: Session = Depends(get_db)):
    """특정 페이지의 모든 콘텐츠 가져오기"""
    contents = db.query(Content).filter(Content.page == page).all()
    return contents

@app.get("/api/content/{page}/{element_id}")
async def get_content(page: str, element_id: str, db: Session = Depends(get_db)):
    """특정 요소의 콘텐츠 가져오기"""
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
    token: str = None,
    db: Session = Depends(get_db)
):
    """콘텐츠 업데이트 (관리자만)"""
    check_admin(token)

    # 기존 콘텐츠 찾기
    content = db.query(Content).filter(
        Content.page == data.page,
        Content.element_id == data.element_id
    ).first()

    if content:
        # 업데이트
        content.content = data.content
        content.updated_at = datetime.utcnow()
    else:
        # 새로 생성
        content = Content(
            page=data.page,
            element_id=data.element_id,
            content=data.content
        )
        db.add(content)

    db.commit()
    db.refresh(content)

    return {
        "success": True,
        "content": content.content,
        "updated_at": content.updated_at
    }

@app.get("/api/pages")
async def get_all_pages(db: Session = Depends(get_db)):
    """모든 페이지 목록 가져오기"""
    pages = db.query(Content.page).distinct().all()
    return [p[0] for p in pages]

# 관리자 페이지
@app.get("/admin", response_class=HTMLResponse)
async def admin_panel():
    """관리자 패널 HTML"""
    return """
<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>WOOHWAHAE CMS Admin</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Helvetica Neue", sans-serif;
            background: #f5f5f5;
            padding: 40px;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }

        h1 {
            font-size: 28px;
            margin-bottom: 30px;
            color: #333;
        }

        .login-form {
            max-width: 400px;
            margin: 100px auto;
        }

        input[type="password"] {
            width: 100%;
            padding: 12px;
            font-size: 16px;
            border: 1px solid #ddd;
            border-radius: 4px;
            margin-bottom: 20px;
        }

        button {
            background: #000;
            color: white;
            padding: 12px 24px;
            font-size: 16px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
        }

        button:hover {
            background: #333;
        }

        .admin-panel {
            display: none;
        }

        .admin-panel.active {
            display: block;
        }

        .page-selector {
            margin-bottom: 30px;
        }

        select {
            padding: 10px;
            font-size: 16px;
            border: 1px solid #ddd;
            border-radius: 4px;
            margin-right: 10px;
        }

        .content-editor {
            margin-top: 30px;
        }

        .editable-item {
            background: #f9f9f9;
            padding: 20px;
            margin-bottom: 20px;
            border-radius: 4px;
            border: 1px solid #e0e0e0;
        }

        .editable-item h3 {
            font-size: 14px;
            color: #666;
            margin-bottom: 10px;
            text-transform: uppercase;
            letter-spacing: 1px;
        }

        .editable-item textarea {
            width: 100%;
            padding: 12px;
            font-size: 15px;
            border: 1px solid #ddd;
            border-radius: 4px;
            min-height: 100px;
            font-family: inherit;
            resize: vertical;
        }

        .save-btn {
            background: #4CAF50;
            margin-top: 10px;
        }

        .status {
            margin-left: 10px;
            padding: 5px 10px;
            border-radius: 3px;
            font-size: 14px;
            display: inline-block;
        }

        .status.success {
            background: #d4edda;
            color: #155724;
        }

        .status.error {
            background: #f8d7da;
            color: #721c24;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>WOOHWAHAE CMS</h1>

        <!-- 로그인 폼 -->
        <div id="loginForm" class="login-form">
            <h2>관리자 로그인</h2>
            <input type="password" id="password" placeholder="비밀번호 입력">
            <button onclick="login()">로그인</button>
        </div>

        <!-- 관리자 패널 -->
        <div id="adminPanel" class="admin-panel">
            <div class="page-selector">
                <label>페이지 선택:</label>
                <select id="pageSelect" onchange="loadPageContent()">
                    <option value="">페이지를 선택하세요</option>
                    <option value="index">메인 페이지</option>
                    <option value="about">About</option>
                    <option value="practice">Practice</option>
                    <option value="playlist">Playlist</option>
                    <option value="project">Project</option>
                    <option value="photography">Photography</option>
                </select>

                <button onclick="openPage()">페이지 미리보기</button>
            </div>

            <div id="contentEditor" class="content-editor"></div>
        </div>
    </div>

    <script>
        let adminToken = localStorage.getItem('adminToken');

        // 토큰이 있으면 자동 로그인
        if (adminToken) {
            showAdminPanel();
        }

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
                    localStorage.setItem('adminToken', adminToken);
                    showAdminPanel();
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
                const response = await fetch(`/api/contents/${page}`);
                const contents = await response.json();

                if (contents.length === 0) {
                    editor.innerHTML = `
                        <p>아직 편집 가능한 콘텐츠가 없습니다.</p>
                        <p>페이지에서 직접 편집 모드를 활성화하세요.</p>
                    `;
                } else {
                    editor.innerHTML = contents.map(item => `
                        <div class="editable-item">
                            <h3>${item.element_id}</h3>
                            <textarea id="content-${item.id}" data-page="${item.page}" data-element="${item.element_id}">${item.content}</textarea>
                            <button class="save-btn" onclick="saveContent(${item.id})">저장</button>
                            <span id="status-${item.id}" class="status"></span>
                        </div>
                    `).join('');
                }
            } catch (error) {
                editor.innerHTML = '<p>콘텐츠를 불러오는데 실패했습니다.</p>';
            }
        }

        async function saveContent(id) {
            const textarea = document.getElementById(`content-${id}`);
            const page = textarea.dataset.page;
            const elementId = textarea.dataset.element;
            const content = textarea.value;
            const status = document.getElementById(`status-${id}`);

            try {
                const response = await fetch('/api/content/update', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': adminToken
                    },
                    body: JSON.stringify({
                        page: page,
                        element_id: elementId,
                        content: content
                    })
                });

                if (response.ok) {
                    status.className = 'status success';
                    status.textContent = '저장 완료!';
                    setTimeout(() => {
                        status.textContent = '';
                    }, 3000);
                } else {
                    throw new Error('저장 실패');
                }
            } catch (error) {
                status.className = 'status error';
                status.textContent = '저장 실패';
            }
        }

        function openPage() {
            const page = document.getElementById('pageSelect').value;
            if (!page) return;

            const url = page === 'index' ? '/' : `/${page}.html`;
            window.open(`http://localhost:8081${url}`, '_blank');
        }
    </script>
</body>
</html>
    """

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8082)