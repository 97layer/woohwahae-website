"""
WOOHWAHAE Admin Panel
Flask-based CMS for managing archive posts and reviewing 97layerOS pipeline output.
"""

import os
import json
import functools
from datetime import datetime
from pathlib import Path

from flask import (
    Flask, render_template, request, redirect, url_for,
    session, flash, jsonify
)
from werkzeug.security import check_password_hash, generate_password_hash
from werkzeug.utils import secure_filename

# ─── Paths ───
BASE_DIR = Path(__file__).resolve().parent.parent.parent  # 97layerOS root
WEBSITE_DIR = BASE_DIR / 'website'
ARCHIVE_DIR = WEBSITE_DIR / 'archive'
PUBLISHED_DIR = BASE_DIR / 'knowledge' / 'assets' / 'published'
UPLOAD_DIR = WEBSITE_DIR / 'assets' / 'img' / 'uploads'
SIGNALS_DIR = BASE_DIR / 'knowledge' / 'signals'
MEMORY_FILE = BASE_DIR / 'knowledge' / 'long_term_memory.json'

# ─── App ───
app = Flask(
    __name__,
    template_folder=str(Path(__file__).parent / 'templates'),
    static_folder=str(Path(__file__).parent / 'static'),
    static_url_path='/admin-static'
)
app.secret_key = os.getenv('ADMIN_SECRET_KEY', 'woohwahae-dev-secret-change-me')
app.config['MAX_CONTENT_LENGTH'] = 16 * 1024 * 1024  # 16MB max upload

ADMIN_PASSWORD_HASH = os.getenv(
    'ADMIN_PASSWORD_HASH',
    generate_password_hash('woohwahae2024', method='pbkdf2:sha256')  # 개발용 기본 비밀번호
)

ALLOWED_EXTENSIONS = {'png', 'jpg', 'jpeg', 'gif', 'webp', 'svg'}


# ─── Auth ───
def login_required(f):
    @functools.wraps(f)
    def decorated(*args, **kwargs):
        if not session.get('logged_in'):
            return redirect(url_for('login'))
        return f(*args, **kwargs)
    return decorated


@app.route('/login', methods=['GET', 'POST'])
def login():
    if request.method == 'POST':
        password = request.form.get('password', '')
        if check_password_hash(ADMIN_PASSWORD_HASH, password):
            session['logged_in'] = True
            return redirect(url_for('dashboard'))
        flash('비밀번호가 올바르지 않습니다.')
    return render_template('login.html')


@app.route('/logout')
def logout():
    session.clear()
    return redirect(url_for('login'))


# ─── Dashboard ───
@app.route('/')
@login_required
def dashboard():
    posts = _load_index()
    pipeline_count = _count_pipeline_items()
    return render_template('dashboard.html',
                           post_count=len(posts),
                           pipeline_count=pipeline_count)


# ─── Archive CRUD ───
@app.route('/archive')
@login_required
def archive_list():
    posts = _load_index()
    return render_template('archive_list.html', posts=posts)


@app.route('/archive/new', methods=['GET', 'POST'])
@login_required
def archive_new():
    if request.method == 'POST':
        title = request.form.get('title', '').strip()
        slug = request.form.get('slug', '').strip()
        body = request.form.get('body', '').strip()
        preview = request.form.get('preview', '').strip()

        if not title or not slug or not body:
            flash('제목, 슬러그, 본문은 필수입니다.')
            return render_template('archive_edit.html', mode='new',
                                   post={'title': title, 'slug': slug,
                                         'body': body, 'preview': preview})

        # Sanitize slug
        slug = _sanitize_slug(slug)

        # Generate static HTML
        from core.admin.utils.post_generator import generate_post
        date_str = datetime.now().strftime('%Y.%m.%d')
        generate_post(slug, title, date_str, body, preview)

        flash(f'"{title}" 포스트가 생성되었습니다.')
        return redirect(url_for('archive_list'))

    return render_template('archive_edit.html', mode='new', post={})


@app.route('/archive/<slug>/edit', methods=['GET', 'POST'])
@login_required
def archive_edit(slug):
    posts = _load_index()
    post = next((p for p in posts if p['slug'] == slug), None)

    if not post:
        flash('포스트를 찾을 수 없습니다.')
        return redirect(url_for('archive_list'))

    # Load markdown source
    md_path = ARCHIVE_DIR / slug / 'source.md'
    body = md_path.read_text(encoding='utf-8') if md_path.exists() else ''

    if request.method == 'POST':
        title = request.form.get('title', '').strip()
        body = request.form.get('body', '').strip()
        preview = request.form.get('preview', '').strip()

        if not title or not body:
            flash('제목과 본문은 필수입니다.')
            return render_template('archive_edit.html', mode='edit',
                                   post={**post, 'body': body, 'preview': preview})

        from core.admin.utils.post_generator import generate_post
        generate_post(slug, title, post['date'], body, preview)

        flash(f'"{title}" 포스트가 수정되었습니다.')
        return redirect(url_for('archive_list'))

    post['body'] = body
    return render_template('archive_edit.html', mode='edit', post=post)


@app.route('/archive/<slug>/delete', methods=['POST'])
@login_required
def archive_delete(slug):
    from core.admin.utils.post_generator import delete_post
    delete_post(slug)
    flash('포스트가 삭제되었습니다.')
    return redirect(url_for('archive_list'))


# ─── Publish (git push) ───
@app.route('/publish', methods=['POST'])
@login_required
def publish():
    from core.admin.utils.git_publisher import publish_to_website
    success, message = publish_to_website()
    flash(message)
    return redirect(url_for('dashboard'))


# ─── Pipeline Review ───
@app.route('/pipeline')
@login_required
def pipeline():
    items = _load_pipeline_items()
    return render_template('pipeline.html', items=items)


@app.route('/pipeline/<path:item_id>/adopt', methods=['POST'])
@login_required
def pipeline_adopt(item_id):
    """파이프라인 콘텐츠를 아카이브 포스트로 채택"""
    item_path = PUBLISHED_DIR / item_id
    meta_path = item_path / 'meta.json'
    essay_path = item_path / 'archive_essay.txt'

    if not essay_path.exists():
        flash('에세이 파일을 찾을 수 없습니다.')
        return redirect(url_for('pipeline'))

    body = essay_path.read_text(encoding='utf-8')
    title = ''
    if meta_path.exists():
        meta = json.loads(meta_path.read_text(encoding='utf-8'))
        title = meta.get('title', meta.get('signal_id', item_id))

    if not title:
        title = item_id.replace('/', ' — ')

    slug = _sanitize_slug(title)
    date_str = datetime.now().strftime('%Y.%m.%d')
    preview = body[:100] + '...' if len(body) > 100 else body

    from core.admin.utils.post_generator import generate_post
    generate_post(slug, title, date_str, body, preview)

    flash(f'"{title}" 파이프라인 콘텐츠가 채택되었습니다.')
    return redirect(url_for('archive_list'))


# ─── Cockpit (Second Brain UI) ───
@app.route('/cockpit')
@login_required
def cockpit():
    """세컨드 브레인 주입 및 소통 창구"""
    memory = {}
    if MEMORY_FILE.exists():
        try:
            memory = json.loads(MEMORY_FILE.read_text(encoding='utf-8'))
        except:
            pass
    
    # 랭킹 데이터 추출
    concepts = sorted(memory.get('concepts', {}).items(), key=lambda x: x[1], reverse=True)[:10]
    experiences = memory.get('experiences', [])[-5:]
    
    return render_template('cockpit.html', concepts=concepts, experiences=experiences)


@app.route('/api/insight', methods=['POST'])
@login_required
def api_insight():
    """인사이트 주입 API"""
    data = request.json
    text = data.get('text', '').strip()
    if not text:
        return jsonify({'error': '내용이 없습니다.'}), 400
    
    from core.system.cortex_edge import get_cortex
    cortex = get_cortex()
    result = cortex.inject_signal(text, source="web_admin")
    
    return jsonify(result)


@app.route('/api/chat', methods=['POST'])
@login_required
def api_chat():
    """AIEngine 대화 API (Deep RAG)"""
    data = request.json
    text = data.get('text', '').strip()
    if not text:
        return jsonify({'error': '내용이 없습니다.'}), 400
    
    try:
        from core.system.cortex_edge import get_cortex
        cortex = get_cortex()
        result = cortex.query("admin", text)
        return jsonify({'response': result['response']})
    except Exception as e:
        return jsonify({'error': str(e)}), 500


# ─── Image Upload ───
@app.route('/upload', methods=['POST'])
@login_required
def upload_image():
    if 'file' not in request.files:
        return jsonify({'error': '파일이 없습니다.'}), 400

    file = request.files['file']
    if not file.filename:
        return jsonify({'error': '파일명이 없습니다.'}), 400

    if not _allowed_file(file.filename):
        return jsonify({'error': '허용되지 않는 파일 형식입니다.'}), 400

    UPLOAD_DIR.mkdir(parents=True, exist_ok=True)
    filename = secure_filename(file.filename)
    timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
    filename = f"{timestamp}_{filename}"
    file.save(str(UPLOAD_DIR / filename))

    return jsonify({
        'url': f'/assets/img/uploads/{filename}',
        'filename': filename
    })


# ─── Helpers ───
def _load_index():
    index_path = ARCHIVE_DIR / 'index.json'
    if not index_path.exists():
        return []
    try:
        return json.loads(index_path.read_text(encoding='utf-8'))
    except (json.JSONDecodeError, IOError):
        return []


def _count_pipeline_items():
    if not PUBLISHED_DIR.exists():
        return 0
    count = 0
    for date_dir in PUBLISHED_DIR.iterdir():
        if date_dir.is_dir():
            for item_dir in date_dir.iterdir():
                if item_dir.is_dir() and (item_dir / 'archive_essay.txt').exists():
                    count += 1
    return count


def _load_pipeline_items():
    items = []
    if not PUBLISHED_DIR.exists():
        return items

    for date_dir in sorted(PUBLISHED_DIR.iterdir(), reverse=True):
        if not date_dir.is_dir():
            continue
        for item_dir in sorted(date_dir.iterdir(), reverse=True):
            if not item_dir.is_dir():
                continue

            essay_path = item_dir / 'archive_essay.txt'
            caption_path = item_dir / 'instagram_caption.txt'
            meta_path = item_dir / 'meta.json'

            if not essay_path.exists():
                continue

            item = {
                'id': f"{date_dir.name}/{item_dir.name}",
                'date': date_dir.name,
                'essay': essay_path.read_text(encoding='utf-8')[:300],
                'caption': '',
                'meta': {}
            }

            if caption_path.exists():
                item['caption'] = caption_path.read_text(encoding='utf-8')
            if meta_path.exists():
                try:
                    item['meta'] = json.loads(meta_path.read_text(encoding='utf-8'))
                except json.JSONDecodeError:
                    pass

            items.append(item)

    return items[:20]  # Latest 20


def _sanitize_slug(text):
    import re
    slug = text.lower().strip()
    slug = re.sub(r'[^\w\s-]', '', slug)
    slug = re.sub(r'[\s_]+', '-', slug)
    slug = re.sub(r'-+', '-', slug)
    return slug[:60].strip('-')


def _allowed_file(filename):
    return '.' in filename and \
           filename.rsplit('.', 1)[1].lower() in ALLOWED_EXTENSIONS


# ─── Run ───
if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5001, debug=True)
