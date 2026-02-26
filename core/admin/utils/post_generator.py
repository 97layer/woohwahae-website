"""
Post Generator — 마크다운 → 정적 HTML 변환 + index.json 관리
"""

import json
import shutil
from pathlib import Path

import markdown

BASE_DIR = Path(__file__).resolve().parent.parent.parent.parent  # LAYER OS root
WEBSITE_DIR = BASE_DIR / 'website'
ARCHIVE_DIR = WEBSITE_DIR / 'archive'
INDEX_PATH = ARCHIVE_DIR / 'index.json'

# ─── Post HTML Template ───
POST_TEMPLATE = """<!DOCTYPE html>
<html lang="ko">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="description" content="{preview}">
  <meta name="theme-color" content="#1A1A1A">
  <meta name="apple-mobile-web-app-capable" content="yes">
  <link rel="manifest" href="/manifest.webmanifest">
  <link rel="apple-touch-icon" href="/assets/img/icon-192.png">
  <title>{title} — WOOHWAHAE</title>
  <link rel="stylesheet" href="../../assets/css/style.css">
</head>
<body>

  <nav>
    <a href="/" class="nav-logo">
      <img src="../../assets/img/symbol.svg" class="nav-symbol" alt="">
      WOOHWAHAE
    </a>
    <ul class="nav-links">
      <li><a href="../../about.html">About</a></li>
      <li><a href="../../practice/">Practice</a></li>
      <li><a href="../../archive/" class="active">Archive</a></li>
      <li><a href="../../contact.html">Contact</a></li>
    </ul>
    <button class="nav-toggle" aria-label="Menu">
      <span></span><span></span><span></span>
    </button>
  </nav>

  <article>
    <div class="container">
      <header class="article-header">
        <p class="article-meta fade-in">Archive · {date}</p>
        <h1 class="article-title fade-in">{title}</h1>
        <p class="article-lead fade-in">{preview}</p>
      </header>
    </div>
    <div class="container">
      <div class="article-body fade-in">
        {body_html}
      </div>
    </div>
  </article>

  <footer>
    <div class="footer-grid">
      <div class="footer-brand">
        <img src="../../assets/img/symbol.svg" alt="WOOHWAHAE" width="28" style="opacity:0.3; margin-bottom:0.5rem;">
        <p class="footer-brand-name">WOOHWAHAE</p>
        <p class="footer-brand-desc">Slow Life Atelier</p>
      </div>
      <div>
        <p class="footer-nav-title">Navigate</p>
        <ul class="footer-nav-list">
          <li><a href="../../about.html">About</a></li>
          <li><a href="../../practice/">Practice</a></li>
          <li><a href="../../archive/">Archive</a></li>
          <li><a href="../../contact.html">Contact</a></li>
        </ul>
      </div>
      <div>
        <p class="footer-connect-title">Connect</p>
        <ul class="footer-connect-list">
          <li><a href="https://instagram.com/woohwahae" target="_blank" rel="noopener">Instagram</a></li>
          <li><a href="https://map.naver.com/p/entry/place/1017153611?lng=129.3497445&lat=35.5592547&placePath=/stylist" target="_blank" rel="noopener">Booking</a></li>
          <li><a href="mailto:hello@woohwahae.kr">Email</a></li>
        </ul>
      </div>
    </div>
    <div class="footer-bottom">
      <p class="footer-bottom-copy">&copy; 2026 WOOHWAHAE &middot; Based in Ulsan</p>
    </div>
  </footer>

  <script src="../../assets/js/main.js"></script>
</body>
</html>"""


def generate_post(slug: str, title: str, date: str, body_md: str, preview: str):
    """마크다운 본문으로 정적 HTML 생성 + index.json 업데이트"""

    # 1. 마크다운 → HTML
    body_html = markdown.markdown(
        body_md,
        extensions=['extra', 'smarty', 'sane_lists']
    )

    # 2. HTML 파일 생성
    post_dir = ARCHIVE_DIR / slug
    post_dir.mkdir(parents=True, exist_ok=True)

    html = POST_TEMPLATE.format(
        title=_escape_html(title),
        date=date,
        preview=_escape_html(preview),
        body_html=body_html
    )
    (post_dir / 'index.html').write_text(html, encoding='utf-8')

    # 3. 마크다운 소스 저장 (편집용)
    (post_dir / 'source.md').write_text(body_md, encoding='utf-8')

    # 4. index.json 업데이트
    _update_index(slug, title, date, preview)


def delete_post(slug: str):
    """포스트 삭제 + index.json에서 제거"""
    post_dir = ARCHIVE_DIR / slug
    if post_dir.exists():
        shutil.rmtree(post_dir)
    _remove_from_index(slug)


def _update_index(slug: str, title: str, date: str, preview: str):
    """index.json에 포스트 추가/업데이트 (날짜 역순 정렬)"""
    posts = _load_index()

    # 기존 포스트 업데이트 or 새로 추가
    existing = next((p for p in posts if p['slug'] == slug), None)
    if existing:
        existing['title'] = title
        existing['date'] = date
        existing['preview'] = preview
    else:
        posts.append({
            'slug': slug,
            'title': title,
            'date': date,
            'preview': preview
        })

    # 날짜 역순 정렬
    posts.sort(key=lambda p: p['date'], reverse=True)
    INDEX_PATH.write_text(
        json.dumps(posts, ensure_ascii=False, indent=2),
        encoding='utf-8'
    )


def _remove_from_index(slug: str):
    """index.json에서 포스트 제거"""
    posts = _load_index()
    posts = [p for p in posts if p['slug'] != slug]
    INDEX_PATH.write_text(
        json.dumps(posts, ensure_ascii=False, indent=2),
        encoding='utf-8'
    )


def _load_index():
    if not INDEX_PATH.exists():
        return []
    try:
        return json.loads(INDEX_PATH.read_text(encoding='utf-8'))
    except (json.JSONDecodeError, IOError):
        return []


def _escape_html(text: str) -> str:
    return text.replace('&', '&amp;').replace('<', '&lt;').replace('>', '&gt;').replace('"', '&quot;')
