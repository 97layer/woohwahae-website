"""
WOOHWAHAE CMS Backend
Simple Flask-based content management system
"""

import sys
from pathlib import Path

# LAYER OS core modules 접근 (프로젝트 루트 추가)
_PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(_PROJECT_ROOT))

from flask import Flask, request, jsonify, session, render_template, abort, redirect, url_for
from flask_cors import CORS
import json
import os
from datetime import datetime
from functools import wraps
import config

# Ritual / Growth 모듈
try:
    from core.modules.ritual import get_ritual_module
    from core.modules.growth import get_growth_module
    _MODULES_AVAILABLE = True
except ImportError:
    _MODULES_AVAILABLE = False

app = Flask(__name__)
app.config.from_object(config)
CORS(app, origins=config.CORS_ORIGINS)


# Authentication decorator
def require_auth(f):
    @wraps(f)
    def decorated(*args, **kwargs):
        if not session.get('authenticated'):
            return jsonify({'error': 'Unauthorized'}), 401
        return f(*args, **kwargs)
    return decorated


# Authentication routes
@app.route('/api/auth/login', methods=['POST'])
def login():
    """Simple password-based login"""
    data = request.get_json()
    password = data.get('password')

    if password == config.ADMIN_PASSWORD:
        session['authenticated'] = True
        return jsonify({'success': True, 'message': 'Authenticated'})

    return jsonify({'error': 'Invalid password'}), 401


@app.route('/api/auth/logout', methods=['POST'])
def logout():
    """Logout"""
    session.pop('authenticated', None)
    return jsonify({'success': True})


@app.route('/api/auth/check', methods=['GET'])
def check_auth():
    """Check authentication status"""
    return jsonify({'authenticated': session.get('authenticated', False)})


# Archive content management
@app.route('/api/archive', methods=['GET'])
def get_archive():
    """Get all archive posts"""
    try:
        with open(config.ARCHIVE_JSON, 'r', encoding='utf-8') as f:
            posts = json.load(f)
        return jsonify(posts)
    except FileNotFoundError:
        return jsonify([])


@app.route('/api/archive', methods=['POST'])
@require_auth
def create_post():
    """Create new archive post"""
    data = request.get_json()

    # Validate required fields
    required = ['slug', 'title', 'date', 'issue', 'preview', 'category']
    if not all(field in data for field in required):
        return jsonify({'error': 'Missing required fields'}), 400

    # Load existing posts
    try:
        with open(config.ARCHIVE_JSON, 'r', encoding='utf-8') as f:
            posts = json.load(f)
    except FileNotFoundError:
        posts = []

    # Check for duplicate slug
    if any(post['slug'] == data['slug'] for post in posts):
        return jsonify({'error': 'Slug already exists'}), 400

    # Add new post (at the beginning for newest first)
    posts.insert(0, {
        'slug': data['slug'],
        'title': data['title'],
        'date': data['date'],
        'issue': data['issue'],
        'preview': data['preview'],
        'category': data['category']
    })

    # Save
    with open(config.ARCHIVE_JSON, 'w', encoding='utf-8') as f:
        json.dump(posts, f, ensure_ascii=False, indent=2)

    # Create post directory
    post_dir = config.ARCHIVE_DIR / data['slug']
    post_dir.mkdir(exist_ok=True)

    # Create basic index.html if content provided
    if 'content' in data:
        html_content = generate_post_html(data)
        with open(post_dir / 'index.html', 'w', encoding='utf-8') as f:
            f.write(html_content)

    return jsonify({'success': True, 'post': data})


@app.route('/api/archive/<slug>', methods=['PUT'])
@require_auth
def update_post(slug):
    """Update existing archive post"""
    data = request.get_json()

    # Load existing posts
    try:
        with open(config.ARCHIVE_JSON, 'r', encoding='utf-8') as f:
            posts = json.load(f)
    except FileNotFoundError:
        return jsonify({'error': 'Archive not found'}), 404

    # Find and update post
    post_index = next((i for i, p in enumerate(posts) if p['slug'] == slug), None)
    if post_index is None:
        return jsonify({'error': 'Post not found'}), 404

    # Update fields
    for field in ['title', 'date', 'issue', 'preview', 'category']:
        if field in data:
            posts[post_index][field] = data[field]

    # Save
    with open(config.ARCHIVE_JSON, 'w', encoding='utf-8') as f:
        json.dump(posts, f, ensure_ascii=False, indent=2)

    return jsonify({'success': True, 'post': posts[post_index]})


@app.route('/api/archive/<slug>', methods=['DELETE'])
@require_auth
def delete_post(slug):
    """Delete archive post"""
    # Load existing posts
    try:
        with open(config.ARCHIVE_JSON, 'r', encoding='utf-8') as f:
            posts = json.load(f)
    except FileNotFoundError:
        return jsonify({'error': 'Archive not found'}), 404

    # Remove post
    posts = [p for p in posts if p['slug'] != slug]

    # Save
    with open(config.ARCHIVE_JSON, 'w', encoding='utf-8') as f:
        json.dump(posts, f, ensure_ascii=False, indent=2)

    return jsonify({'success': True})


# Image upload
@app.route('/api/upload', methods=['POST'])
@require_auth
def upload_image():
    """Upload image file"""
    if 'file' not in request.files:
        return jsonify({'error': 'No file provided'}), 400

    file = request.files['file']
    if file.filename == '':
        return jsonify({'error': 'No file selected'}), 400

    # Check file extension
    ext = file.filename.rsplit('.', 1)[1].lower() if '.' in file.filename else ''
    if ext not in config.ALLOWED_EXTENSIONS:
        return jsonify({'error': 'Invalid file type'}), 400

    # Generate unique filename
    timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
    filename = f"{timestamp}_{file.filename}"
    filepath = config.UPLOADS_DIR / filename

    # Save file
    file.save(filepath)

    # Return URL
    url = f"/assets/uploads/{filename}"
    return jsonify({'success': True, 'url': url})


# Utility: Generate post HTML
def generate_post_html(data):
    """Generate basic HTML for archive post"""
    return f"""<!DOCTYPE html>
<html lang="ko">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="description" content="{data.get('preview', '')}">
  <title>{data.get('title', '')} — WOOHWAHAE</title>
  <link rel="stylesheet" href="../../assets/css/style.css">
</head>
<body>

  <nav>
    <a href="/" class="nav-logo">
      <img src="../../assets/img/symbol.jpg" class="nav-symbol" alt="WOOHWAHAE">
    </a>
    <ul class="nav-links">
      <li><a href="../../about.html">About</a></li>
      <li><a href="../">Archive</a></li>
      <li><a href="../../shop.html">Shop</a></li>
      <li><a href="../../atelier.html">Atelier</a></li>
      <li><a href="../../playlist.html">Playlist</a></li>
      <li><a href="../../project.html">Project</a></li>
      <li><a href="../../photography.html">Photography</a></li>
    </ul>
    <button class="nav-toggle" aria-label="Menu">
      <span></span><span></span><span></span>
    </button>
  </nav>

  <article class="container">
    <header class="page-header">
      <p class="section-label fade-in">{data.get('issue', '')}</p>
      <h1 class="fade-in">{data.get('title', '')}</h1>
      <p class="page-header-desc fade-in">{data.get('date', '')}</p>
    </header>

    <section class="article-content fade-in">
      {data.get('content', '<p>Content coming soon...</p>')}
    </section>

    <footer class="article-footer">
      <a href="../" class="cta">← Back to Archive</a>
    </footer>
  </article>

  <script src="../../assets/js/main.js"></script>
</body>
</html>
"""


# ─── 고객 시술일지 포털 (/me/{token}) ──────────────────────────

@app.route('/me/<token>')
def client_portal(token):
    """고객 전용 시술일지 페이지 — 로그인 없음, 토큰 기반 접근."""
    if not _MODULES_AVAILABLE:
        abort(503)
    rm = get_ritual_module()
    client = rm.find_client_by_token(token)
    if not client:
        abort(404)

    # 내부 전용 필드 제거 (notes, color_formula 고객에게 비공개)
    visits = []
    for v in reversed(client.get('visits', [])):
        visits.append({
            'date': v.get('date', ''),
            'service': v.get('service', ''),
            'public_note': v.get('public_note', ''),
            'next_visit_weeks': v.get('next_visit_weeks', 0),
            'satisfaction': v.get('satisfaction'),
            'amount': v.get('amount', 0),
        })

    total_paid = sum(v.get('amount', 0) for v in client.get('visits', []))

    from silhouette_renderer import generate_silhouette
    silhouette = generate_silhouette(client)

    return render_template('portal.html',
        client=client,
        visits=visits,
        total_paid=total_paid,
        silhouette=silhouette,
        token=token,
    )


# ─── 사전상담 폼 (/consult/{token}) ────────────────────────────

@app.route('/consult/<token>', methods=['GET', 'POST'])
def consultation(token):
    """예약 확정 후 고객에게 전송하는 사전상담 폼."""
    if not _MODULES_AVAILABLE:
        abort(503)
    rm = get_ritual_module()
    client = rm.find_client_by_token(token)
    if not client:
        abort(404)

    if request.method == 'POST':
        # 상담 내용 저장
        design_memo = request.form.get('design_memo', '')
        lifestyle = request.form.get('lifestyle', '')
        length_pref = request.form.get('length_pref', '')
        mood_keywords = request.form.getlist('mood')
        expected_duration = request.form.get('expected_duration', '')

        preference_notes = (
            f"[{datetime.now().strftime('%Y-%m-%d')} 사전상담] "
            f"길이: {length_pref} / 무드: {', '.join(mood_keywords)} / "
            f"라이프스타일: {lifestyle} / 예상시간: {expected_duration}시간\n"
            f"{design_memo}"
        )
        rm.update_client(client['client_id'], preference_notes=preference_notes)

        # 사진 업로드 처리
        if 'design_photo' in request.files:
            photo = request.files['design_photo']
            if photo and photo.filename:
                ext = photo.filename.rsplit('.', 1)[-1].lower()
                if ext in {'jpg', 'jpeg', 'png', 'webp', 'heic'}:
                    save_dir = _PROJECT_ROOT / 'website' / 'assets' / 'uploads' / 'consult'
                    save_dir.mkdir(parents=True, exist_ok=True)
                    ts = datetime.now().strftime('%Y%m%d_%H%M%S')
                    photo.save(save_dir / f"{client['client_id']}_{ts}.{ext}")

        return redirect(f'/me/{token}')

    return render_template('consult.html', client=client)


if __name__ == '__main__':
    app.run(host='127.0.0.1', port=5000, debug=config.DEBUG)
