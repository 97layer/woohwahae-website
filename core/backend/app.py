"""
WOOHWAHAE CMS Backend
Flask-based content management system — 보안 강화 버전
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
from werkzeug.utils import secure_filename
import config

from core.system.security import (
    verify_password,
    apply_security_headers,
    sanitize_html_field,
    safe_path,
    setup_audit_logger,
    RateLimiter,
)

# Ritual / Growth 모듈
try:
    from core.system.ritual import get_ritual_module
    from core.system.growth import get_growth_module
    _MODULES_AVAILABLE = True
except ImportError:
    _MODULES_AVAILABLE = False

app = Flask(__name__)
app.config.from_object(config)

# CORS — 명시적 오리진만 허용, 메서드/헤더 제한
CORS(app, origins=config.CORS_ORIGINS, methods=['GET', 'POST', 'PUT', 'DELETE'],
     allow_headers=['Content-Type', 'Authorization'])

# 보안 쿠키 설정
app.config.update(
    SESSION_COOKIE_HTTPONLY=True,
    SESSION_COOKIE_SECURE=True,
    SESSION_COOKIE_SAMESITE='Strict',
    PERMANENT_SESSION_LIFETIME=3600,  # 1시간
)

# 보안 헤더
app.after_request(apply_security_headers)

# 감사 로거
_audit_logger = setup_audit_logger('cms_audit', config.AUDIT_LOG_FILE)

# 로그인 레이트 리미터 (5회/분)
_login_limiter = RateLimiter(max_requests=5, window_seconds=60)


def _audit(action: str, detail: str = '') -> None:
    ip = request.remote_addr if request else 'system'
    _audit_logger.info("ip=%s action=%s detail=%s", ip, action, detail)


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
    """해시 기반 비밀번호 인증 + 레이트 리미팅"""
    ip = request.remote_addr or 'unknown'

    if _login_limiter.is_limited(ip):
        _audit('login_rate_limited', ip)
        return jsonify({'error': 'Too many attempts. Try later.'}), 429

    data = request.get_json()
    if not data:
        return jsonify({'error': 'Invalid request'}), 400

    password = data.get('password', '')

    if verify_password(password, config.ADMIN_PASSWORD_HASH):
        session['authenticated'] = True
        session.permanent = True
        _audit('login_success', ip)
        return jsonify({'success': True, 'message': 'Authenticated'})

    _audit('login_failed', ip)
    return jsonify({'error': 'Invalid password'}), 401


@app.route('/api/auth/logout', methods=['POST'])
def logout():
    _audit('logout')
    session.pop('authenticated', None)
    return jsonify({'success': True})


@app.route('/api/auth/check', methods=['GET'])
def check_auth():
    return jsonify({'authenticated': session.get('authenticated', False)})


# Archive content management
@app.route('/api/archive', methods=['GET'])
def get_archive():
    try:
        with open(config.ARCHIVE_JSON, 'r', encoding='utf-8') as f:
            posts = json.load(f)
        return jsonify(posts)
    except FileNotFoundError:
        return jsonify([])


@app.route('/api/archive', methods=['POST'])
@require_auth
def create_post():
    data = request.get_json()

    required = ['slug', 'title', 'date', 'issue', 'preview', 'category']
    if not all(field in data for field in required):
        return jsonify({'error': 'Missing required fields'}), 400

    # XSS 방어 — 사용자 입력 이스케이프
    sanitized = {}
    for field in required:
        sanitized[field] = sanitize_html_field(str(data[field])[:500])

    # slug 경로 방어
    if safe_path(sanitized['slug'], config.ARCHIVE_DIR) is None:
        return jsonify({'error': 'Invalid slug'}), 400

    try:
        with open(config.ARCHIVE_JSON, 'r', encoding='utf-8') as f:
            posts = json.load(f)
    except FileNotFoundError:
        posts = []

    if any(post['slug'] == sanitized['slug'] for post in posts):
        return jsonify({'error': 'Slug already exists'}), 400

    posts.insert(0, sanitized)

    with open(config.ARCHIVE_JSON, 'w', encoding='utf-8') as f:
        json.dump(posts, f, ensure_ascii=False, indent=2)

    # 디렉토리 생성 (safe_path로 검증 완료)
    post_dir = config.ARCHIVE_DIR / sanitized['slug']
    post_dir.mkdir(exist_ok=True)

    if 'content' in data:
        sanitized_content = sanitize_html_field(str(data['content']))
        html_content = generate_post_html({**sanitized, 'content': sanitized_content})
        with open(post_dir / 'index.html', 'w', encoding='utf-8') as f:
            f.write(html_content)

    _audit('create_post', sanitized['slug'])
    return jsonify({'success': True, 'post': sanitized})


@app.route('/api/archive/<slug>', methods=['PUT'])
@require_auth
def update_post(slug):
    data = request.get_json()

    try:
        with open(config.ARCHIVE_JSON, 'r', encoding='utf-8') as f:
            posts = json.load(f)
    except FileNotFoundError:
        return jsonify({'error': 'Archive not found'}), 404

    post_index = next((i for i, p in enumerate(posts) if p['slug'] == slug), None)
    if post_index is None:
        return jsonify({'error': 'Post not found'}), 404

    for field in ['title', 'date', 'issue', 'preview', 'category']:
        if field in data:
            posts[post_index][field] = sanitize_html_field(str(data[field])[:500])

    with open(config.ARCHIVE_JSON, 'w', encoding='utf-8') as f:
        json.dump(posts, f, ensure_ascii=False, indent=2)

    _audit('update_post', slug)
    return jsonify({'success': True, 'post': posts[post_index]})


@app.route('/api/archive/<slug>', methods=['DELETE'])
@require_auth
def delete_post(slug):
    try:
        with open(config.ARCHIVE_JSON, 'r', encoding='utf-8') as f:
            posts = json.load(f)
    except FileNotFoundError:
        return jsonify({'error': 'Archive not found'}), 404

    posts = [p for p in posts if p['slug'] != slug]

    with open(config.ARCHIVE_JSON, 'w', encoding='utf-8') as f:
        json.dump(posts, f, ensure_ascii=False, indent=2)

    _audit('delete_post', slug)
    return jsonify({'success': True})


# Image upload
@app.route('/api/upload', methods=['POST'])
@require_auth
def upload_image():
    if 'file' not in request.files:
        return jsonify({'error': 'No file provided'}), 400

    file = request.files['file']
    if file.filename == '':
        return jsonify({'error': 'No file selected'}), 400

    # secure_filename으로 파일명 정규화
    safe_name = secure_filename(file.filename)
    if not safe_name:
        return jsonify({'error': 'Invalid filename'}), 400

    ext = safe_name.rsplit('.', 1)[1].lower() if '.' in safe_name else ''
    if ext not in config.ALLOWED_EXTENSIONS:
        return jsonify({'error': 'Invalid file type'}), 400

    timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
    filename = "%s_%s" % (timestamp, safe_name)
    filepath = config.UPLOADS_DIR / filename

    file.save(filepath)

    _audit('upload_image', filename)
    url = "/assets/uploads/%s" % filename
    return jsonify({'success': True, 'url': url})


def generate_post_html(data):
    """HTML 생성 — 입력은 사전 이스케이프됨."""
    return """<!DOCTYPE html>
<html lang="ko">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="description" content="{preview}">
  <title>{title} — WOOHWAHAE</title>
  <link rel="stylesheet" href="../../assets/css/style.css">
</head>
<body>

  <nav>
    <a href="/" class="nav-logo">
      <img src="../../assets/img/symbol.jpg" class="nav-symbol" alt="WOOHWAHAE">
    </a>
    <ul class="nav-links">
      <li><a href="/about/">About</a></li>
      <li><a href="/archive/">Archive</a></li>
      <li><a href="/practice/">Practice</a></li>
    </ul>
    <button class="nav-toggle" aria-label="Menu">
      <span></span><span></span><span></span>
    </button>
  </nav>

  <article class="container">
    <header class="page-header">
      <p class="section-label fade-in">{issue}</p>
      <h1 class="fade-in">{title}</h1>
      <p class="page-header-desc fade-in">{date}</p>
    </header>

    <section class="article-content fade-in">
      {content}
    </section>

    <footer class="article-footer">
      <a href="../" class="cta">&larr; Back to Archive</a>
    </footer>
  </article>

  <script src="../../assets/js/main.js"></script>
</body>
</html>
""".format(**data)


# ─── 고객 시술일지 포털 (/me/{token}) ──────────────────────────

@app.route('/me/<token>')
def client_portal(token):
    if not _MODULES_AVAILABLE:
        abort(503)
    rm = get_ritual_module()
    client = rm.find_client_by_token(token)
    if not client:
        abort(404)

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

    from silhouette_renderer import get_hair_style
    hair_style = get_hair_style(client)

    return render_template('portal.html',
        client=client,
        visits=visits,
        total_paid=total_paid,
        hair_style=hair_style,
        token=token,
    )


# ─── 사전상담 폼 (/consult/{token}) ────────────────────────────

@app.route('/consult/<token>', methods=['GET', 'POST'])
def consultation(token):
    if not _MODULES_AVAILABLE:
        abort(503)
    rm = get_ritual_module()
    client = rm.find_client_by_token(token)
    if not client:
        abort(404)

    if request.method == 'POST':
        design_memo = sanitize_html_field(request.form.get('design_memo', ''))
        lifestyle = sanitize_html_field(request.form.get('lifestyle', ''))
        length_pref = sanitize_html_field(request.form.get('length_pref', ''))
        mood_keywords = request.form.getlist('mood')
        expected_duration = sanitize_html_field(request.form.get('expected_duration', ''))

        preference_notes = (
            "[%s 사전상담] "
            "길이: %s / 무드: %s / "
            "라이프스타일: %s / 예상시간: %s시간\n"
            "%s"
        ) % (
            datetime.now().strftime('%Y-%m-%d'),
            length_pref,
            ', '.join(sanitize_html_field(m) for m in mood_keywords),
            lifestyle,
            expected_duration,
            design_memo,
        )
        rm.update_client(client['client_id'], preference_notes=preference_notes)

        if 'design_photo' in request.files:
            photo = request.files['design_photo']
            if photo and photo.filename:
                safe_name = secure_filename(photo.filename)
                ext = safe_name.rsplit('.', 1)[-1].lower() if '.' in safe_name else ''
                if ext in {'jpg', 'jpeg', 'png', 'webp', 'heic'}:
                    save_dir = _PROJECT_ROOT / 'website' / 'assets' / 'uploads' / 'consult'
                    save_dir.mkdir(parents=True, exist_ok=True)
                    ts = datetime.now().strftime('%Y%m%d_%H%M%S')
                    photo.save(save_dir / ("%s_%s.%s" % (client['client_id'], ts, ext)))

        _audit('consultation_submit', client.get('client_id', 'unknown'))
        return redirect('/me/%s' % token)

    return render_template('consult.html', client=client)


# ─── 에러 핸들러 (디버그 정보 노출 방지) ──────────────────────

@app.errorhandler(500)
def internal_error(e):
    return jsonify({'error': 'Internal server error'}), 500


@app.errorhandler(404)
def not_found(e):
    return jsonify({'error': 'Not found'}), 404


if __name__ == '__main__':
    app.run(host='127.0.0.1', port=5000, debug=False)
