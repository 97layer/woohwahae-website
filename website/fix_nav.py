#!/usr/bin/env python3
import re
import os

# 7개 세션 네비게이션 템플릿
nav_template = """    <ul class="nav-links">
      <li><a href="about.html"{about_active}>About</a></li>
      <li><a href="archive/"{archive_active}>Archive</a></li>
      <li><a href="shop.html"{shop_active}>Shop</a></li>
      <li><a href="atelier.html"{atelier_active}>Atelier</a></li>
      <li><a href="playlist.html"{playlist_active}>Playlist</a></li>
      <li><a href="project.html"{project_active}>Project</a></li>
      <li><a href="photography.html"{photography_active}>Photography</a></li>
    </ul>"""

footer_nav = """        <ul class="footer-nav-list">
          <li><a href="about.html">About</a></li>
          <li><a href="archive/">Archive</a></li>
          <li><a href="shop.html">Shop</a></li>
          <li><a href="atelier.html">Atelier</a></li>
          <li><a href="playlist.html">Playlist</a></li>
          <li><a href="project.html">Project</a></li>
          <li><a href="photography.html">Photography</a></li>
        </ul>"""

# 페이지별 active 클래스 매핑
pages = {
    'shop.html': 'shop',
    'atelier.html': 'atelier',
    'playlist.html': 'playlist',
    'project.html': 'project',
    'photography.html': 'photography',
    'contact.html': None,  # contact는 7개 세션에 없음
}

# 메인 페이지 네비게이션도 수정
main_nav = """    <ul class="nav-links">
      <li><a href="about.html">About</a></li>
      <li><a href="archive/">Archive</a></li>
      <li><a href="shop.html">Shop</a></li>
      <li><a href="atelier.html">Atelier</a></li>
      <li><a href="playlist.html">Playlist</a></li>
      <li><a href="project.html">Project</a></li>
      <li><a href="photography.html">Photography</a></li>
    </ul>"""

def update_file(filename, active_page=None):
    """파일의 네비게이션과 푸터 업데이트"""
    try:
        with open(filename, 'r', encoding='utf-8') as f:
            content = f.read()

        # 네비게이션 생성
        nav = nav_template
        actives = {
            'about_active': '',
            'archive_active': '',
            'shop_active': '',
            'atelier_active': '',
            'playlist_active': '',
            'project_active': '',
            'photography_active': ''
        }

        if active_page:
            actives[f'{active_page}_active'] = ' class="active"'

        nav = nav.format(**actives)

        # nav-links 교체
        content = re.sub(
            r'<ul class="nav-links">.*?</ul>',
            nav,
            content,
            flags=re.DOTALL
        )

        # footer-nav-list 교체
        content = re.sub(
            r'<ul class="footer-nav-list">.*?</ul>',
            footer_nav,
            content,
            flags=re.DOTALL
        )

        with open(filename, 'w', encoding='utf-8') as f:
            f.write(content)

        print(f"✓ Updated {filename}")
        return True
    except Exception as e:
        print(f"✗ Error updating {filename}: {e}")
        return False

# 각 페이지 업데이트
for page, active in pages.items():
    if os.path.exists(page):
        update_file(page, active)

# index.html 업데이트 (active 없음)
if os.path.exists('index.html'):
    with open('index.html', 'r', encoding='utf-8') as f:
        content = f.read()

    # 상단 네비게이션 업데이트
    content = re.sub(
        r'<ul class="nav-links">.*?</ul>',
        main_nav,
        content,
        flags=re.DOTALL
    )

    # 푸터 네비게이션 업데이트
    content = re.sub(
        r'<ul class="footer-nav-list">.*?</ul>',
        footer_nav,
        content,
        flags=re.DOTALL
    )

    with open('index.html', 'w', encoding='utf-8') as f:
        f.write(content)

    print("✓ Updated index.html")

# archive/index.html 업데이트
if os.path.exists('archive/index.html'):
    with open('archive/index.html', 'r', encoding='utf-8') as f:
        content = f.read()

    archive_nav = """    <ul class="nav-links">
      <li><a href="../about.html">About</a></li>
      <li><a href="../archive/" class="active">Archive</a></li>
      <li><a href="../shop.html">Shop</a></li>
      <li><a href="../atelier.html">Atelier</a></li>
      <li><a href="../playlist.html">Playlist</a></li>
      <li><a href="../project.html">Project</a></li>
      <li><a href="../photography.html">Photography</a></li>
    </ul>"""

    archive_footer = """        <ul class="footer-nav-list">
          <li><a href="../about.html">About</a></li>
          <li><a href="../archive/">Archive</a></li>
          <li><a href="../shop.html">Shop</a></li>
          <li><a href="../atelier.html">Atelier</a></li>
          <li><a href="../playlist.html">Playlist</a></li>
          <li><a href="../project.html">Project</a></li>
          <li><a href="../photography.html">Photography</a></li>
        </ul>"""

    content = re.sub(
        r'<ul class="nav-links">.*?</ul>',
        archive_nav,
        content,
        flags=re.DOTALL
    )

    content = re.sub(
        r'<ul class="footer-nav-list">.*?</ul>',
        archive_footer,
        content,
        flags=re.DOTALL
    )

    with open('archive/index.html', 'w', encoding='utf-8') as f:
        f.write(content)

    print("✓ Updated archive/index.html")

print("\n✓ Navigation update complete!")