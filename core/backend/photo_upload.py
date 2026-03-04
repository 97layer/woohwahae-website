#!/usr/bin/env python3
"""
WOOHWAHAE Photo Upload Extension
사진 업로드 기능 추가
"""

from __future__ import annotations

import json
import os
import re
import secrets
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional

from fastapi import FastAPI, File, Form, HTTPException, Request, UploadFile
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from fastapi.staticfiles import StaticFiles

_PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(_PROJECT_ROOT))

from core.system.security import load_cors_origins

# FastAPI 앱 초기화
app = FastAPI(title="WOOHWAHAE Photo Upload", version="1.0.0")

# CORS 설정
app.add_middleware(
    CORSMiddleware,
    allow_origins=load_cors_origins(default=["https://woohwahae.kr", "http://localhost:8082"]),
    allow_credentials=True,
    allow_methods=["GET", "POST", "DELETE"],
    allow_headers=["Authorization", "X-Admin-Token", "Content-Type"],
)

# 업로드 디렉토리 설정
UPLOAD_DIR = Path(__file__).resolve().parent.parent.parent / "website" / "assets" / "uploads"
UPLOAD_DIR.mkdir(exist_ok=True, parents=True)

# 카테고리별 디렉토리 생성
CATEGORIES = ["hair", "space", "detail", "moment", "product"]
for category in CATEGORIES:
    (UPLOAD_DIR / category).mkdir(exist_ok=True)

# 정적 파일 서빙
app.mount("/uploads", StaticFiles(directory=str(UPLOAD_DIR)), name="uploads")

# 이미지 메타데이터 저장 파일
METADATA_FILE = UPLOAD_DIR / "metadata.json"
ALLOWED_EXTENSIONS = {"jpg", "jpeg", "png", "gif", "webp"}
ALLOWED_MIME_TYPES = {
    "image/jpeg",
    "image/png",
    "image/gif",
    "image/webp",
}
MAX_UPLOAD_BYTES = int(os.getenv("PHOTO_UPLOAD_MAX_BYTES", str(10 * 1024 * 1024)))
MAX_UPLOAD_FILES = int(os.getenv("PHOTO_UPLOAD_MAX_FILES", "10"))
PHOTO_UPLOAD_TOKEN = os.getenv("PHOTO_UPLOAD_ADMIN_TOKEN", "").strip()


def _extract_admin_token(request: Request) -> str:
    auth = request.headers.get("authorization", "").strip()
    if auth.lower().startswith("bearer "):
        return auth.split(" ", 1)[1].strip()
    return request.headers.get("x-admin-token", "").strip()


def _require_admin_auth(request: Request) -> None:
    if not PHOTO_UPLOAD_TOKEN:
        raise HTTPException(
            status_code=503,
            detail="PHOTO_UPLOAD_ADMIN_TOKEN is not configured",
        )
    token = _extract_admin_token(request)
    if not token or not secrets.compare_digest(token, PHOTO_UPLOAD_TOKEN):
        raise HTTPException(status_code=401, detail="Unauthorized")


def _safe_filename_parts(original_name: str) -> tuple[str, str]:
    if not original_name:
        raise HTTPException(status_code=400, detail="Empty filename")
    basename = Path(original_name).name
    suffix = Path(basename).suffix.lower().lstrip(".")
    if suffix not in ALLOWED_EXTENSIONS:
        raise HTTPException(status_code=400, detail="Invalid file extension")
    stem = re.sub(r"[^a-zA-Z0-9._-]", "_", Path(basename).stem).strip("._")
    if not stem:
        stem = "upload"
    return stem[:80], suffix


def _matches_image_signature(raw: bytes, suffix: str) -> bool:
    if suffix in {"jpg", "jpeg"}:
        return raw.startswith(b"\xff\xd8\xff")
    if suffix == "png":
        return raw.startswith(b"\x89PNG\r\n\x1a\n")
    if suffix == "gif":
        return raw.startswith(b"GIF87a") or raw.startswith(b"GIF89a")
    if suffix == "webp":
        return len(raw) >= 12 and raw[:4] == b"RIFF" and raw[8:12] == b"WEBP"
    return False


def _safe_upload_path(relative_path: str) -> Path:
    target = (UPLOAD_DIR / relative_path).resolve()
    if not str(target).startswith(str(UPLOAD_DIR.resolve())):
        raise HTTPException(status_code=400, detail="Invalid path")
    return target


def load_metadata() -> list[dict]:
    """저장된 메타데이터 불러오기"""
    if METADATA_FILE.exists():
        try:
            return json.loads(METADATA_FILE.read_text(encoding="utf-8"))
        except (json.JSONDecodeError, OSError):
            return []
    return []


def save_metadata(metadata: list[dict]) -> None:
    """메타데이터 저장"""
    METADATA_FILE.write_text(
        json.dumps(metadata, indent=2, ensure_ascii=False),
        encoding="utf-8",
    )

@app.post("/api/upload")
async def upload_photos(
    request: Request,
    photos: list[UploadFile] = File(...),
    category: str = Form(...),
):
    """사진 업로드"""
    _require_admin_auth(request)
    if category not in CATEGORIES:
        raise HTTPException(status_code=400, detail="Invalid category")
    if not photos:
        raise HTTPException(status_code=400, detail="No files provided")
    if len(photos) > MAX_UPLOAD_FILES:
        raise HTTPException(
            status_code=400,
            detail=f"Too many files (max {MAX_UPLOAD_FILES})",
        )

    uploaded_files = []
    metadata = load_metadata()
    next_id = (max((m.get("id", 0) for m in metadata), default=0) + 1)

    for photo in photos:
        stem, suffix = _safe_filename_parts(photo.filename or "")
        content_type = (photo.content_type or "").lower()
        if content_type not in ALLOWED_MIME_TYPES:
            raise HTTPException(status_code=400, detail="Invalid content type")

        raw = await photo.read(MAX_UPLOAD_BYTES + 1)
        await photo.close()
        if len(raw) > MAX_UPLOAD_BYTES:
            raise HTTPException(status_code=400, detail="File too large")
        if not _matches_image_signature(raw, suffix):
            raise HTTPException(status_code=400, detail="Invalid image signature")

        # 파일명 생성 (UTC 타임스탬프 + 랜덤 토큰)
        now_utc = datetime.now(timezone.utc)
        timestamp = now_utc.strftime("%Y%m%d_%H%M%S")
        filename = f"{timestamp}_{secrets.token_hex(6)}_{stem}.{suffix}"
        filepath = UPLOAD_DIR / category / filename

        # 파일 저장
        filepath.write_bytes(raw)

        # 메타데이터 추가
        file_info = {
            "id": next_id,
            "filename": f"{category}/{filename}",
            "original_name": Path(photo.filename or "").name,
            "category": category,
            "uploaded_at": now_utc.isoformat(),
            "size": filepath.stat().st_size,
            "content_type": content_type,
        }
        next_id += 1
        metadata.append(file_info)
        uploaded_files.append(file_info)

    # 메타데이터 저장
    save_metadata(metadata)

    return {
        "success": True,
        "count": len(uploaded_files),
        "files": uploaded_files
    }

@app.get("/api/images")
async def get_images(request: Request, category: Optional[str] = None):
    """업로드된 이미지 목록 조회"""
    _require_admin_auth(request)
    metadata = load_metadata()

    if category:
        filtered = [img for img in metadata if img["category"] == category]
        return filtered

    return metadata

@app.get("/api/images/{category}")
async def get_category_images(request: Request, category: str):
    """특정 카테고리 이미지 조회"""
    _require_admin_auth(request)
    if category not in CATEGORIES:
        raise HTTPException(status_code=400, detail="Invalid category")

    metadata = load_metadata()
    filtered = [img for img in metadata if img["category"] == category]
    return filtered

@app.delete("/api/images/{image_id}")
async def delete_image(request: Request, image_id: int):
    """이미지 삭제"""
    _require_admin_auth(request)
    metadata = load_metadata()

    # 이미지 찾기
    image = None
    for img in metadata:
        if img["id"] == image_id:
            image = img
            break

    if not image:
        raise HTTPException(status_code=404, detail="Image not found")

    # 파일 삭제
    filepath = _safe_upload_path(image["filename"])
    if filepath.exists():
        filepath.unlink()

    # 메타데이터에서 제거
    metadata = [img for img in metadata if img["id"] != image_id]
    save_metadata(metadata)

    return {"success": True, "message": "Image deleted"}

@app.get("/api/gallery/{category}")
async def get_gallery_html(request: Request, category: str):
    """갤러리 HTML 생성"""
    _require_admin_auth(request)
    if category not in CATEGORIES:
        raise HTTPException(status_code=400, detail="Invalid category")

    metadata = load_metadata()
    images = [img for img in metadata if img["category"] == category]

    html = f"""
    <div class="gallery-grid">
        {"".join([f'<img src="/uploads/{img["filename"]}" alt="{img["category"]}" loading="lazy">' for img in images])}
    </div>
    """

    return JSONResponse(content={"html": html, "count": len(images)})

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8083)
