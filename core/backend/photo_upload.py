#!/usr/bin/env python3
"""
WOOHWAHAE Photo Upload Extension
사진 업로드 기능 추가
"""

from fastapi import FastAPI, File, UploadFile, Form, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from fastapi.responses import JSONResponse
from typing import List
import os
import shutil
from datetime import datetime
import json
from pathlib import Path

# FastAPI 앱 초기화
app = FastAPI(title="WOOHWAHAE Photo Upload", version="1.0.0")

# CORS 설정
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:8081", "http://localhost:8082", "http://localhost:8000"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
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

def load_metadata():
    """저장된 메타데이터 불러오기"""
    if METADATA_FILE.exists():
        with open(METADATA_FILE, 'r') as f:
            return json.load(f)
    return []

def save_metadata(metadata):
    """메타데이터 저장"""
    with open(METADATA_FILE, 'w') as f:
        json.dump(metadata, f, indent=2, ensure_ascii=False)

@app.post("/api/upload")
async def upload_photos(
    photos: List[UploadFile] = File(...),
    category: str = Form(...)
):
    """사진 업로드"""
    if category not in CATEGORIES:
        raise HTTPException(status_code=400, detail="Invalid category")

    uploaded_files = []
    metadata = load_metadata()

    for photo in photos:
        # 파일명 생성 (타임스탬프 추가)
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        filename = f"{timestamp}_{photo.filename}"
        filepath = UPLOAD_DIR / category / filename

        # 파일 저장
        with open(filepath, "wb") as buffer:
            shutil.copyfileobj(photo.file, buffer)

        # 메타데이터 추가
        file_info = {
            "id": len(metadata) + 1,
            "filename": f"{category}/{filename}",
            "original_name": photo.filename,
            "category": category,
            "uploaded_at": datetime.now().isoformat(),
            "size": filepath.stat().st_size
        }
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
async def get_images(category: str = None):
    """업로드된 이미지 목록 조회"""
    metadata = load_metadata()

    if category:
        filtered = [img for img in metadata if img["category"] == category]
        return filtered

    return metadata

@app.get("/api/images/{category}")
async def get_category_images(category: str):
    """특정 카테고리 이미지 조회"""
    if category not in CATEGORIES:
        raise HTTPException(status_code=400, detail="Invalid category")

    metadata = load_metadata()
    filtered = [img for img in metadata if img["category"] == category]
    return filtered

@app.delete("/api/images/{image_id}")
async def delete_image(image_id: int):
    """이미지 삭제"""
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
    filepath = UPLOAD_DIR / image["filename"]
    if filepath.exists():
        filepath.unlink()

    # 메타데이터에서 제거
    metadata = [img for img in metadata if img["id"] != image_id]
    save_metadata(metadata)

    return {"success": True, "message": "Image deleted"}

@app.get("/api/gallery/{category}")
async def get_gallery_html(category: str):
    """갤러리 HTML 생성"""
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