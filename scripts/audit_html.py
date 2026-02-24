#!/usr/bin/env python3
"""
HTML 전수조사 스크립트
WOOHWAHAE 웹사이트 UI/UX/텍스트 일관성 검증
"""

import os
import re
import json
from pathlib import Path
from bs4 import BeautifulSoup
from typing import Dict, List, Tuple

# 브랜드 기준
BRAND_TOKENS = {
    'colors': {
        'allowed': ['--bg', '--text', '--text-sub', '--text-faint', '--white', '--line', '--stone-dark', '--stone-mid', '--stone-light'],
        'prohibited': ['--navy'],  # ::selection, blockquote만 허용
        'hardcoded_patterns': [
            r'#[0-9A-Fa-f]{6}',  # hex colors
            r'rgb\(',
            r'rgba\('
        ]
    },
    'fonts': {
        'allowed': ['--font-body', '--font-mono', '--font-serif', '--font-serif-slab', 'Pretendard Variable', 'IBM Plex Mono', 'DM Mono', 'Crimson Text', 'Bitter'],
        'prohibited': ['Arial', 'sans-serif', 'serif']  # generic만 금지, fallback은 허용
    },
    'spacing': {
        'tokens': ['--space-xs', '--space-sm', '--space-md', '--space-lg', '--space-xl', '--space-2xl']
    },
    'tone': {
        'archive': '한다체',  # 사색적, 열린 결말
        'magazine': '합니다체'  # 독자 지향
    }
}

# 검사 결과 저장
issues = {
    'critical': [],  # 즉시 수정 필요
    'warning': [],   # 일관성 개선 권장
    'info': []       # 참고 사항
}

def check_hardcoded_colors(file_path: str, soup: BeautifulSoup):
    """hardcoded 색상 검사"""
    style_tags = soup.find_all('style')

    for style in style_tags:
        content = style.string
        if not content:
            continue

        # hex 색상 찾기
        hex_colors = re.findall(r'#[0-9A-Fa-f]{6}', content)
        for color in hex_colors:
            # 토큰 정의부는 제외
            if ':root' in content or 'CSS Custom Properties' in content:
                continue

            issues['warning'].append({
                'file': file_path,
                'type': 'hardcoded_color',
                'value': color,
                'message': f'Hardcoded color {color} — 토큰으로 변환 권장'
            })

def check_navy_usage(file_path: str, soup: BeautifulSoup):
    """navy 색상 금지 패턴 검사"""
    style_tags = soup.find_all('style')

    for style in style_tags:
        content = style.string
        if not content:
            continue

        if 'var(--navy)' in content:
            # 허용 패턴 체크
            if '::selection' in content or 'blockquote' in content:
                continue

            issues['critical'].append({
                'file': file_path,
                'type': 'prohibited_navy',
                'message': 'var(--navy) 금지 패턴 발견 — ::selection/blockquote만 허용'
            })

def check_nav_structure(file_path: str, soup: BeautifulSoup):
    """nav 구조 일관성 검사"""
    nav = soup.find('nav')
    if not nav:
        issues['warning'].append({
            'file': file_path,
            'type': 'missing_nav',
            'message': 'nav 요소 없음'
        })
        return

    # nav 링크 추출
    links = nav.find_all('a')
    nav_items = [link.get_text(strip=True) for link in links]

    # 표준 nav 구조: Archive / Offering / About / Contact / Lab
    expected = ['Archive', 'Offering', 'About', 'Contact', 'Lab']

    if nav_items != expected and 'WOOHWAHAE' not in nav_items[0]:  # 로고는 제외
        issues['info'].append({
            'file': file_path,
            'type': 'nav_structure',
            'current': nav_items,
            'expected': expected,
            'message': 'nav 구조가 표준과 다름'
        })

def check_meta_tags(file_path: str, soup: BeautifulSoup):
    """SEO/메타태그 검사"""
    # og:tags
    og_title = soup.find('meta', property='og:title')
    og_desc = soup.find('meta', property='og:description')
    og_image = soup.find('meta', property='og:image')

    if not og_title:
        issues['warning'].append({
            'file': file_path,
            'type': 'missing_og_title',
            'message': 'og:title 없음'
        })

    if not og_desc:
        issues['warning'].append({
            'file': file_path,
            'type': 'missing_og_description',
            'message': 'og:description 없음'
        })

def check_font_usage(file_path: str, soup: BeautifulSoup):
    """폰트 토큰 사용 검사"""
    style_tags = soup.find_all('style')

    for style in style_tags:
        content = style.string
        if not content:
            continue

        # font-family 직접 지정 찾기 (토큰 사용 권장)
        font_declarations = re.findall(r'font-family:\s*([^;]+);', content)

        for font in font_declarations:
            if 'var(--font-' not in font and 'inherit' not in font:
                # fallback 체크 (Pretendard, sans-serif 같은 구조는 허용)
                if ',' in font and 'sans-serif' in font:
                    continue

                issues['info'].append({
                    'file': file_path,
                    'type': 'direct_font',
                    'value': font.strip(),
                    'message': 'font-family 직접 지정 — 토큰 사용 권장'
                })

def check_accessibility(file_path: str, soup: BeautifulSoup):
    """접근성 검사"""
    # img alt 체크
    images = soup.find_all('img')
    for img in images:
        if not img.get('alt'):
            issues['warning'].append({
                'file': file_path,
                'type': 'missing_alt',
                'src': img.get('src', 'unknown'),
                'message': 'img alt 속성 없음'
            })

    # button aria-label 체크
    buttons = soup.find_all('button')
    for btn in buttons:
        text = btn.get_text(strip=True)
        if not text and not btn.get('aria-label'):
            issues['warning'].append({
                'file': file_path,
                'type': 'missing_aria_label',
                'message': 'button에 텍스트/aria-label 없음'
            })

def audit_html_file(file_path: str):
    """단일 HTML 파일 감사"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()

        soup = BeautifulSoup(content, 'html.parser')

        # 검사 실행
        check_hardcoded_colors(file_path, soup)
        check_navy_usage(file_path, soup)
        check_nav_structure(file_path, soup)
        check_meta_tags(file_path, soup)
        check_font_usage(file_path, soup)
        check_accessibility(file_path, soup)

    except Exception as e:
        issues['critical'].append({
            'file': file_path,
            'type': 'parse_error',
            'message': f'파일 파싱 오류: {str(e)}'
        })

def main():
    website_root = Path('/Users/97layer/97layerOS/website')

    # 우선순위 1 — 핵심 페이지
    priority_files = [
        'index.html',
        'about.html',
        'contact.html',
        'offering.html',
        'photography.html',
        'archive/index.html'
    ]

    print("🔍 WOOHWAHAE 웹사이트 전수조사 시작\n")
    print("=" * 60)

    # 1차 우선순위 파일 감사
    print("\n[Phase 5] 핵심 페이지 감사 (6개)")
    for file in priority_files:
        file_path = website_root / file
        if file_path.exists():
            print(f"  • {file}")
            audit_html_file(str(file_path))

    # Archive 에세이 감사
    print("\n[Phase 5] Archive 에세이 감사")
    archive_dir = website_root / 'archive'
    essay_dirs = sorted([d for d in archive_dir.iterdir() if d.is_dir() and d.name.startswith('issue-')])

    for essay_dir in essay_dirs[:10]:  # 최대 10개
        index_file = essay_dir / 'index.html'
        if index_file.exists():
            print(f"  • {essay_dir.name}/index.html")
            audit_html_file(str(index_file))

    # 결과 출력
    print("\n" + "=" * 60)
    print(f"\n📊 감사 결과 요약")
    print(f"  • CRITICAL (즉시 수정): {len(issues['critical'])}")
    print(f"  • WARNING (권장 수정): {len(issues['warning'])}")
    print(f"  • INFO (참고): {len(issues['info'])}")

    # JSON 저장
    output_file = '/tmp/html_audit_report.json'
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(issues, f, ensure_ascii=False, indent=2)

    print(f"\n✅ 상세 보고서: {output_file}")

    # Critical 이슈 즉시 출력
    if issues['critical']:
        print("\n🚨 CRITICAL 이슈:")
        for issue in issues['critical'][:5]:
            print(f"  • [{issue['type']}] {issue['file']}")
            print(f"    → {issue['message']}")

if __name__ == '__main__':
    main()
