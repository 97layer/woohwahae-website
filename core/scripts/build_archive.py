#!/usr/bin/env python3
"""
build_archive.py — WOOHWAHAE 아카이브 SSG 빌드

사용:
  python scripts/build_archive.py           # 전체 빌드
  python scripts/build_archive.py --slug essay-010-work-and-essence  # 단일 글

입력:  website/_content/*.md
출력:  website/archive/{slug}/index.html
       website/archive/index.json  (자동 갱신)
"""

import os
import re
import json
import argparse
import markdown as md_lib
from pathlib import Path
from datetime import datetime

ROOT = Path(__file__).parent.parent
CONTENT_DIR = ROOT / 'website' / '_content'
TEMPLATE_FILE = ROOT / 'website' / '_templates' / 'article.html'
ARCHIVE_DIR = ROOT / 'website' / 'archive'
INDEX_JSON = ARCHIVE_DIR / 'index.json'


# ── frontmatter 파서 ──────────────────────────────────────────────────────────

def parse_frontmatter(text):
    """--- ... --- 블록 파싱 → (meta dict, body str)"""
    match = re.match(r'^---\s*\n(.*?)\n---\s*\n(.*)', text, re.DOTALL)
    if not match:
        return {}, text

    meta = {}
    for line in match.group(1).splitlines():
        if ':' in line:
            key, _, val = line.partition(':')
            meta[key.strip()] = val.strip()

    body = match.group(2).strip()
    return meta, body


# ── 본문 변환 ─────────────────────────────────────────────────────────────────

def convert_body(body_md):
    """마크다운 → HTML (fade-in 클래스 자동 부여)"""
    html = md_lib.markdown(
        body_md,
        extensions=['extra'],   # tables, fenced_code, footnotes 등
    )

    # p 태그에 fade-in 추가
    html = re.sub(r'<p>', '<p class="fade-in">', html)
    # blockquote에 fade-in 추가
    html = re.sub(r'<blockquote>', '<blockquote class="fade-in">', html)

    return html


# ── 이전 글 링크 ──────────────────────────────────────────────────────────────

def make_prev_link(all_meta, current_slug):
    """현재 글보다 한 단계 이전 글 링크 생성"""
    slugs = [m['slug'] for m in all_meta]
    try:
        idx = slugs.index(current_slug)
    except ValueError:
        return ''

    # all_meta는 최신순 정렬 → idx+1이 이전 글
    if idx + 1 < len(all_meta):
        prev = all_meta[idx + 1]
        return (
            f'<a href="../{prev["slug"]}/" class="article-nav-prev">'
            f'← {prev["issue"]} · {prev["title"]}</a>'
        )
    return ''


# ── 단일 글 빌드 ──────────────────────────────────────────────────────────────

def build_one(md_path, template, all_meta):
    text = md_path.read_text(encoding='utf-8')
    meta, body = parse_frontmatter(text)

    slug = meta.get('slug') or md_path.stem
    title = meta.get('title', '')
    issue = meta.get('issue', '')
    date = meta.get('date', '')
    category = meta.get('category', 'Essay')
    preview = meta.get('preview', '')
    read_min = meta.get('readMin', '2')
    dot_label = meta.get('dot_label', '10 yrs · 120 mo')

    content_html = convert_body(body)
    prev_link = make_prev_link(all_meta, slug)

    html = template
    replacements = {
        '{{slug}}': slug,
        '{{title}}': title,
        '{{issue}}': issue,
        '{{date}}': date,
        '{{category}}': category,
        '{{preview}}': preview,
        '{{readMin}}': str(read_min),
        '{{dot_label}}': dot_label,
        '{{prev_link}}': prev_link,
        '{{content}}': content_html,
    }
    for key, val in replacements.items():
        html = html.replace(key, val)

    out_dir = ARCHIVE_DIR / slug
    out_dir.mkdir(parents=True, exist_ok=True)
    out_file = out_dir / 'index.html'
    out_file.write_text(html, encoding='utf-8')

    print('[build] %s → %s' % (md_path.name, out_file.relative_to(ROOT)))
    return {
        'slug': slug,
        'title': title,
        'date': date,
        'issue': issue,
        'preview': preview,
        'category': category,
        'readMin': int(read_min) if str(read_min).isdigit() else 2,
    }


# ── index.json 갱신 ──────────────────────────────────────────────────────────

def update_index(built_entries):
    """빌드된 항목으로 index.json 갱신 (날짜 역순 정렬)"""
    # 기존 JSON 읽기 (Lab 카드 등 비-md 항목 보존)
    existing = []
    if INDEX_JSON.exists():
        existing = json.loads(INDEX_JSON.read_text(encoding='utf-8'))

    # slug 기준으로 빌드 결과로 덮어쓰기
    built_slugs = {e['slug'] for e in built_entries}
    preserved = [e for e in existing if e.get('slug') not in built_slugs]
    merged = built_entries + preserved

    # 날짜 역순 정렬 (YYYY.MM.DD 형식 처리)
    def sort_key(e):
        d = e.get('date', '0000.00.00').replace('.', '-')
        try:
            return datetime.strptime(d, '%Y-%m-%d')
        except ValueError:
            return datetime.min

    merged.sort(key=sort_key, reverse=True)

    INDEX_JSON.write_text(
        json.dumps(merged, ensure_ascii=False, indent=2),
        encoding='utf-8'
    )
    print('[index] %s 갱신 (%d entries)' % (INDEX_JSON.relative_to(ROOT), len(merged)))


# ── 메인 ─────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--slug', help='특정 글만 빌드 (slug 지정)')
    args = parser.parse_args()

    if not CONTENT_DIR.exists():
        print('[error] _content/ 디렉토리 없음: %s' % CONTENT_DIR)
        return

    template = TEMPLATE_FILE.read_text(encoding='utf-8')

    # 전체 .md 목록 (빌드 순서용 메타 미리 수집)
    md_files = sorted(CONTENT_DIR.glob('*.md'))
    if not md_files:
        print('[warn] _content/에 .md 파일 없음')
        return

    # 메타 미리 파싱 (prev_link 계산용)
    all_meta = []
    for f in md_files:
        text = f.read_text(encoding='utf-8')
        meta, _ = parse_frontmatter(text)
        slug = meta.get('slug') or f.stem
        all_meta.append({
            'slug': slug,
            'title': meta.get('title', ''),
            'issue': meta.get('issue', ''),
        })

    # 날짜 기준 역순 정렬 (최신이 앞)
    # _content 파일명이 issue-NNN 순이면 역순 정렬
    all_meta.reverse()

    # 빌드 대상 결정
    if args.slug:
        targets = [f for f in md_files if (f.stem == args.slug or
                   parse_frontmatter(f.read_text('utf-8'))[0].get('slug') == args.slug)]
        if not targets:
            print('[error] slug "%s" 에 해당하는 .md 파일 없음' % args.slug)
            return
    else:
        targets = md_files

    built = []
    for f in targets:
        entry = build_one(f, template, all_meta)
        built.append(entry)

    update_index(built)
    print('[done] 빌드 완료 (%d개)' % len(built))


if __name__ == '__main__':
    main()
