"""
WOOHWAHAE Admin Panel
Flask-based CMS for managing archive posts and reviewing 97layerOS pipeline output.
"""

import os
import sys
import json
import logging
import secrets
import functools
from collections import defaultdict
from datetime import datetime
from pathlib import Path

from flask import (
    Flask, render_template, request, redirect, url_for,
    session, flash, jsonify, abort
)
from werkzeug.security import check_password_hash, generate_password_hash
from werkzeug.utils import secure_filename

# Load .env
try:
    from dotenv import load_dotenv
    _BASE_DIR_FOR_ENV = Path(__file__).resolve().parent.parent.parent
    load_dotenv(_BASE_DIR_FOR_ENV / '.env')
except ImportError:
    pass

# ─── Paths ───
BASE_DIR = Path(__file__).resolve().parent.parent.parent  # 97layerOS root
sys.path.insert(0, str(BASE_DIR))  # core.modules 접근용

# Ritual / Growth 모듈
try:
    from core.modules.ritual import get_ritual_module
    from core.modules.growth import get_growth_module
    _MODULES_AVAILABLE = True
except ImportError:
    _MODULES_AVAILABLE = False
WEBSITE_DIR = BASE_DIR / 'website'
ARCHIVE_DIR = WEBSITE_DIR / 'archive'
PUBLISHED_DIR = BASE_DIR / 'knowledge' / 'assets' / 'published'
UPLOAD_DIR = WEBSITE_DIR / 'assets' / 'img' / 'uploads'
SIGNALS_DIR = BASE_DIR / 'knowledge' / 'signals'
MEMORY_FILE = BASE_DIR / 'knowledge' / 'long_term_memory.json'
AUDIT_LOG_FILE = BASE_DIR / 'knowledge' / 'reports' / 'audit.log'

# ─── App ───
app = Flask(
    __name__,
    template_folder=str(Path(__file__).parent / 'templates'),
    static_folder=str(Path(__file__).parent / 'static'),
    static_url_path='/admin-static'
)

# B1: Secret key — env var 필수, fallback 없음
_secret_key = os.getenv('ADMIN_SECRET_KEY')
if not _secret_key:
    raise RuntimeError("ADMIN_SECRET_KEY 환경변수 필수. .env에 설정하세요.")
app.secret_key = _secret_key

app.config['MAX_CONTENT_LENGTH'] = 16 * 1024 * 1024  # 16MB max upload
app.config['DEBUG'] = False
app.config['PROPAGATE_EXCEPTIONS'] = False

# B2: 쿠키 보안
app.config.update(
    SESSION_COOKIE_HTTPONLY=True,
    SESSION_COOKIE_SECURE=os.getenv('HTTPS_ENABLED', 'false').lower() == 'true',
    SESSION_COOKIE_SAMESITE='Strict',
    PERMANENT_SESSION_LIFETIME=3600,  # 1시간 세션 만료
)

ADMIN_PASSWORD_HASH = os.getenv(
    'ADMIN_PASSWORD_HASH',
    generate_password_hash('woohwahae2024', method='pbkdf2:sha256')
)

ALLOWED_EXTENSIONS = {'png', 'jpg', 'jpeg', 'gif', 'webp', 'svg'}
ALLOWED_MIME = {'image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/svg+xml'}

# ─── Logging ───
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# B10: Audit logger
audit_logger = logging.getLogger('audit')
try:
    AUDIT_LOG_FILE.parent.mkdir(parents=True, exist_ok=True)
    _audit_handler = logging.FileHandler(str(AUDIT_LOG_FILE))
    _audit_handler.setFormatter(logging.Formatter('%(asctime)s %(message)s'))
    audit_logger.addHandler(_audit_handler)
    audit_logger.setLevel(logging.INFO)
except Exception:
    pass  # 로그 파일 생성 실패해도 서버 동작에 영향 없음


def _audit(action: str, detail: str = '') -> None:
    """B10: 관리자 액션 감사 로그"""
    ip = request.remote_addr if request else 'system'
    audit_logger.info("ip=%s action=%s detail=%s", ip, action, detail)


# B8: 로그인 실패 카운터 (인메모리, 재시작 시 초기화)
_login_attempts: dict = defaultdict(int)
LOGIN_MAX_ATTEMPTS = 10


# ─── B3: CSRF ───
_CSRF_EXEMPT_PATHS = {'/login'}


def _generate_csrf_token() -> str:
    if '_csrf_token' not in session:
        session['_csrf_token'] = secrets.token_hex(32)
    return session['_csrf_token']


app.jinja_env.globals['csrf_token'] = _generate_csrf_token


@app.before_request
def csrf_protect() -> None:
    """B3: POST 요청 CSRF 검증 (로그인 제외)"""
    if request.method != 'POST':
        return
    if request.path in _CSRF_EXEMPT_PATHS:
        return
    # JSON API: X-CSRF-Token 헤더 허용 (SameSite=Strict 보완)
    if request.is_json:
        token = request.headers.get('X-CSRF-Token')
    else:
        token = request.form.get('_csrf_token')
    expected = session.get('_csrf_token')
    if not expected or token != expected:
        _audit('csrf_violation', request.path)
        abort(403)


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
    ip = request.remote_addr
    if request.method == 'POST':
        # B8: rate limiting
        if _login_attempts[ip] >= LOGIN_MAX_ATTEMPTS:
            _audit('login_blocked', 'too_many_attempts')
            return jsonify({'error': '로그인 시도 횟수 초과. 잠시 후 다시 시도하세요.'}), 429

        password = request.form.get('password', '')
        if check_password_hash(ADMIN_PASSWORD_HASH, password):
            _login_attempts[ip] = 0
            session['logged_in'] = True
            session.permanent = True
            _audit('login_success')
            return redirect(url_for('dashboard'))

        _login_attempts[ip] += 1
        _audit('login_failed', 'attempt_%d' % _login_attempts[ip])
        flash('비밀번호가 올바르지 않습니다.')
    return render_template('login.html')


@app.route('/logout')
def logout():
    _audit('logout')
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

        _audit('archive_create', slug)
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

        _audit('archive_edit', slug)
        flash(f'"{title}" 포스트가 수정되었습니다.')
        return redirect(url_for('archive_list'))

    post['body'] = body
    return render_template('archive_edit.html', mode='edit', post=post)


@app.route('/archive/<slug>/delete', methods=['POST'])
@login_required
def archive_delete(slug):
    from core.admin.utils.post_generator import delete_post
    delete_post(slug)
    _audit('archive_delete', slug)
    flash('포스트가 삭제되었습니다.')
    return redirect(url_for('archive_list'))


# ─── Publish (git push) ───
@app.route('/publish', methods=['POST'])
@login_required
def publish():
    from core.admin.utils.git_publisher import publish_to_website
    success, message = publish_to_website()
    _audit('publish', 'success=%s' % success)
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
    # B9: Path traversal 방어
    item_path = (PUBLISHED_DIR / item_id).resolve()
    if not str(item_path).startswith(str(PUBLISHED_DIR.resolve())):
        _audit('path_traversal_blocked', item_id)
        abort(400)

    meta_path = item_path / 'meta.json'
    essay_path = item_path / 'archive_essay.txt'

    if not essay_path.exists():
        flash('에세이 파일을 찾을 수 없습니다.')
        return redirect(url_for('pipeline'))

    body = essay_path.read_text(encoding='utf-8')
    title = ''
    if meta_path.exists():
        try:
            meta = json.loads(meta_path.read_text(encoding='utf-8'))
            title = meta.get('title', meta.get('signal_id', item_id))
        except (json.JSONDecodeError, IOError):
            pass

    if not title:
        title = item_id.replace('/', ' — ')

    slug = _sanitize_slug(title)
    date_str = datetime.now().strftime('%Y.%m.%d')
    preview = body[:100] + '...' if len(body) > 100 else body

    from core.admin.utils.post_generator import generate_post
    generate_post(slug, title, date_str, body, preview)

    _audit('pipeline_adopt', slug)
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
        except (json.JSONDecodeError, IOError):
            pass

    # 랭킹 데이터 추출
    concepts = sorted(memory.get('concepts', {}).items(), key=lambda x: x[1], reverse=True)[:10]
    experiences = memory.get('experiences', [])[-5:]

    return render_template('cockpit.html', concepts=concepts, experiences=experiences)


@app.route('/api/insight', methods=['POST'])
@login_required
def api_insight():
    """인사이트 주입 API"""
    data = request.json or {}
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
    data = request.json or {}
    text = data.get('text', '').strip()
    if not text:
        return jsonify({'error': '내용이 없습니다.'}), 400

    try:
        from core.system.cortex_edge import get_cortex
        cortex = get_cortex()
        result = cortex.query("admin", text)
        return jsonify({'response': result['response']})
    except Exception as e:
        logger.error("api_chat error: %s", e)
        return jsonify({'error': '내부 오류'}), 500


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

    # B9: MIME type 검증 (python-magic 없을 때 graceful fallback)
    try:
        import magic
        file_bytes = file.read(2048)
        file.seek(0)
        mime = magic.from_buffer(file_bytes, mime=True)
        if mime not in ALLOWED_MIME:
            return jsonify({'error': '허용되지 않는 파일 형식입니다.'}), 400
    except ImportError:
        pass  # python-magic 미설치 시 확장자 검증만 사용

    UPLOAD_DIR.mkdir(parents=True, exist_ok=True)
    filename = secure_filename(file.filename)
    timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
    filename = f"{timestamp}_{filename}"
    file.save(str(UPLOAD_DIR / filename))

    _audit('upload', filename)
    return jsonify({
        'url': f'/assets/img/uploads/{filename}',
        'filename': filename
    })


# ─── Growth Dashboard ───

@app.route('/growth')
@login_required
def growth_dashboard():
    if not _MODULES_AVAILABLE:
        flash('모듈 로드 실패. core/modules 경로 확인.')
        return redirect(url_for('dashboard'))

    gm = get_growth_module()
    rm = get_ritual_module()
    period = request.args.get('period', datetime.now().strftime('%Y-%m'))

    try:
        gm.auto_count_content(period)
        gm.auto_count_service(period)
    except Exception as e:
        logger.error("growth auto_count error: %s", e)

    try:
        data = gm.get_month(period)
        trend = gm.get_trend(months=6)
        periods = gm.list_periods()
        stats = rm.get_stats()
    except Exception as e:
        logger.error("growth data fetch error: %s", e)
        data, trend, periods, stats = {}, [], [], {}

    _audit('growth_view', 'period=%s' % period)
    return render_template('growth.html',
        data=data,
        trend=trend,
        periods=periods,
        stats=stats,
        current_period=period,
    )


@app.route('/growth/revenue', methods=['POST'])
@login_required
def growth_revenue():
    if not _MODULES_AVAILABLE:
        return jsonify({'error': '모듈 없음'}), 503

    period = request.form.get('period', datetime.now().strftime('%Y-%m'))
    gm = get_growth_module()

    try:
        gm.record_revenue(
            period,
            atelier=int(request.form.get('atelier', 0) or 0),
            consulting=int(request.form.get('consulting', 0) or 0),
            products=int(request.form.get('products', 0) or 0),
        )
    except Exception as e:
        logger.error("growth record_revenue error: %s", e)
        flash('수익 저장 실패')
        return redirect(url_for('growth_dashboard', period=period))

    _audit('revenue_recorded', 'period=%s' % period)
    return redirect(url_for('growth_dashboard', period=period))


# ─── B11: 에러 핸들러 ───
@app.errorhandler(400)
def bad_request(e):
    return jsonify({'error': '잘못된 요청'}), 400


@app.errorhandler(403)
def forbidden(e):
    return jsonify({'error': '접근 거부'}), 403


@app.errorhandler(404)
def not_found(e):
    if request.path.startswith('/api/'):
        return jsonify({'error': '없는 경로'}), 404
    return render_template('dashboard.html', post_count=0, pipeline_count=0), 404


@app.errorhandler(429)
def rate_limited(e):
    return jsonify({'error': '요청 과다. 잠시 후 다시 시도하세요.'}), 429


@app.errorhandler(Exception)
def handle_error(e):
    logger.error("unhandled error: %s", e, exc_info=True)
    if request.path.startswith('/api/'):
        return jsonify({'error': '내부 오류'}), 500
    flash('내부 오류가 발생했습니다.')
    return redirect(url_for('dashboard'))


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
                except (json.JSONDecodeError, IOError):
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
    app.run(host='127.0.0.1', port=5001, debug=False)
