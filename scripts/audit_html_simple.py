#!/usr/bin/env python3
"""
HTML 전수조사 스크립트 (Simple Regex 기반)
WOOHWAHAE 웹사이트 UI/UX/텍스트 일관성 검증
"""

import os
import re
import json
from pathlib import Path
from typing import Dict, List

# 검사 결과 저장
issues = {
    'critical': [],
    'warning': [],
    'info': []
}

def check_navy_usage(file_path: str, content: str):
    """navy 색상 금지 패턴 검사"""
    if 'var(--navy)' not in content:
        return

    lines = content.split('\n')
    for i, line in enumerate(lines, 1):
        if 'var(--navy)' in line:
            # 허용 패턴 체크
            if '::selection' in line or 'blockquote' in line:
                continue

            # context 추출 (이전 5줄)
            context_start = max(0, i - 5)
            context = '\n'.join(lines[context_start:i])

            issues['critical'].append({
                'file': file_path,
                'line': i,
                'type': 'prohibited_navy',
                'snippet': line.strip(),
                'message': 'var(--navy) 금지 패턴 — ::selection/blockquote만 허용'
            })

def check_hardcoded_colors(file_path: str, content: str):
    """hardcoded 색상 검사"""
    # :root 정의부는 제외
    if ':root' in content or '/* ━━━ PRIMARY PALETTE' in content:
        # 토큰 정의 섹션 제외
        content_parts = content.split(':root')
        if len(content_parts) > 1:
            # :root 블록 이후만 검사
            content = content_parts[-1]

    # hex 색상 패턴
    hex_pattern = r'#[0-9A-Fa-f]{6}(?![0-9A-Fa-f])'  # 6자리만, 8자리 제외
    matches = re.finditer(hex_pattern, content)

    for match in matches:
        color = match.group()
        # 줄 번호 계산
        line_num = content[:match.start()].count('\n') + 1

        # 토큰 정의부 제외 체크
        line_start = content.rfind('\n', 0, match.start())
        line_end = content.find('\n', match.start())
        line_content = content[line_start:line_end].strip()

        if '--' in line_content and ':' in line_content:  # CSS 변수 정의
            continue

        issues['warning'].append({
            'file': file_path,
            'line': line_num,
            'type': 'hardcoded_color',
            'value': color,
            'snippet': line_content[:80],
            'message': f'Hardcoded color {color} → 토큰 변환 권장'
        })

def check_nav_consistency(file_path: str, content: str):
    """nav 구조 일관성 검사"""
    if '<nav>' not in content and '<nav ' not in content:
        issues['warning'].append({
            'file': file_path,
            'type': 'missing_nav',
            'message': 'nav 요소 없음'
        })
        return

    # nav 링크 추출
    nav_match = re.search(r'<nav[^>]*>(.*?)</nav>', content, re.DOTALL)
    if nav_match:
        nav_content = nav_match.group(1)
        links = re.findall(r'<a[^>]*>([^<]+)</a>', nav_content)

        # 표준 구조: Archive, Offering, About, Contact, Lab
        expected = {'Archive', 'Offering', 'About', 'Contact', 'Lab'}
        found = set(link.strip() for link in links if link.strip() and link.strip() != 'WOOHWAHAE')

        missing = expected - found
        extra = found - expected

        if missing or extra:
            issues['info'].append({
                'file': file_path,
                'type': 'nav_structure',
                'missing': list(missing),
                'extra': list(extra),
                'message': f'nav 구조 차이: missing={missing}, extra={extra}'
            })

def check_meta_tags(file_path: str, content: str):
    """SEO/메타태그 검사"""
    if 'og:title' not in content:
        issues['warning'].append({
            'file': file_path,
            'type': 'missing_og_title',
            'message': 'og:title 메타태그 없음'
        })

    if 'og:description' not in content:
        issues['warning'].append({
            'file': file_path,
            'type': 'missing_og_description',
            'message': 'og:description 메타태그 없음'
        })

    if 'og:image' not in content:
        issues['info'].append({
            'file': file_path,
            'type': 'missing_og_image',
            'message': 'og:image 메타태그 없음'
        })

def check_font_loading(file_path: str, content: str):
    """폰트 로딩 최적화 검사"""
    if 'fonts.googleapis.com' in content or 'cdn.jsdelivr.net' in content:
        if 'preconnect' not in content:
            issues['info'].append({
                'file': file_path,
                'type': 'missing_preconnect',
                'message': 'font preconnect 없음 — 로딩 최적화 권장'
            })

def check_accessibility(file_path: str, content: str):
    """접근성 검사"""
    # img without alt
    img_pattern = r'<img[^>]*>'
    imgs = re.finditer(img_pattern, content)

    for img_match in imgs:
        img_tag = img_match.group()
        if 'alt=' not in img_tag:
            line_num = content[:img_match.start()].count('\n') + 1
            issues['warning'].append({
                'file': file_path,
                'line': line_num,
                'type': 'missing_alt',
                'snippet': img_tag[:60],
                'message': 'img alt 속성 없음'
            })

def check_css_cache_version(file_path: str, content: str):
    """CSS 캐시 버전 검사"""
    css_links = re.findall(r'<link[^>]*href=["\']([^"\']*\.css[^"\']*)["\'[^>]*>', content)

    for css_link in css_links:
        if '/assets/css/style.css' in css_link:
            if '?v=' not in css_link:
                issues['info'].append({
                    'file': file_path,
                    'type': 'missing_css_version',
                    'message': 'style.css 캐시 버전 쿼리 없음'
                })

def audit_html_file(file_path: str):
    """단일 HTML 파일 감사"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # 검사 실행
        check_navy_usage(file_path, content)
        check_hardcoded_colors(file_path, content)
        check_nav_consistency(file_path, content)
        check_meta_tags(file_path, content)
        check_font_loading(file_path, content)
        check_accessibility(file_path, content)
        check_css_cache_version(file_path, content)

    except Exception as e:
        issues['critical'].append({
            'file': file_path,
            'type': 'parse_error',
            'message': f'파일 읽기 오류: {str(e)}'
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
        else:
            print(f"  ✗ {file} (없음)")

    # Archive 에세이 감사
    print("\n[Phase 5] Archive 에세이 감사")
    archive_dir = website_root / 'archive'
    if archive_dir.exists():
        essay_dirs = sorted([d for d in archive_dir.iterdir() if d.is_dir() and d.name.startswith('issue-')])

        for essay_dir in essay_dirs[:10]:
            index_file = essay_dir / 'index.html'
            if index_file.exists():
                print(f"  • {essay_dir.name}/index.html")
                audit_html_file(str(index_file))

    # Flask 템플릿 (이미 완료했지만 재검증)
    print("\n[Phase 5] Flask 템플릿 재검증")
    backend_dir = website_root / 'backend' / 'templates'
    if backend_dir.exists():
        for template in ['portal.html', 'consult.html', 'consult_done.html']:
            template_path = backend_dir / template
            if template_path.exists():
                print(f"  • backend/templates/{template}")
                audit_html_file(str(template_path))

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
        print("\n🚨 CRITICAL 이슈 (최대 10개):")
        for issue in issues['critical'][:10]:
            print(f"\n  [{issue['type']}] {issue['file']}:{issue.get('line', '?')}")
            print(f"  → {issue['message']}")
            if 'snippet' in issue:
                print(f"  | {issue['snippet']}")

    # Warning 상위 5개
    if issues['warning']:
        print("\n⚠️  WARNING 상위 5개:")
        for issue in issues['warning'][:5]:
            print(f"  • [{issue['type']}] {issue['file']}")
            print(f"    → {issue['message']}")

if __name__ == '__main__':
    main()
