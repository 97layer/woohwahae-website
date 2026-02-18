"""
WOOHWAHAE CMS Backend
Simple Flask-based content management system
"""

from flask import Flask, request, jsonify, session
from flask_cors import CORS
from pathlib import Path
import json
import os
from datetime import datetime
from functools import wraps
import config

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


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=config.DEBUG)
