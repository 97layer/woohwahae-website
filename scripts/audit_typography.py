#!/usr/bin/env python3
"""
타이포그래피 일관성 감사
Font-family 직접 지정 → 토큰 변환 검증
"""

import os
import re
import json
from pathlib import Path

issues = {
    'font_direct': [],
    'font_size': [],
    'letter_spacing': []
}

# 허용 토큰
ALLOWED_TOKENS = [
    'var(--font-body)',
    'var(--font-mono)',
    'var(--font-serif)',
    'var(--font-serif-slab)'
]

# 허용 폰트 (fallback 포함)
ALLOWED_FONTS = [
    'Pretendard Variable',
    'IBM Plex Mono',
    'DM Mono',
    'Crimson Text',
    'Bitter',
    'sans-serif',  # fallback
    'serif',       # fallback
    'monospace',   # fallback
    'inherit'
]

def check_font_family(file_path: str, content: str):
    """font-family 직접 지정 검사"""
    # font-family 패턴 찾기
    font_pattern = r'font-family:\s*([^;]+);'
    matches = re.finditer(font_pattern, content, re.IGNORECASE)

    for match in matches:
        font_value = match.group(1).strip()
        line_num = content[:match.start()].count('\n') + 1

        # 토큰 사용 확인
        if any(token in font_value for token in ALLOWED_TOKENS):
            continue

        # :root 정의부 제외
        if ':root' in content[max(0, match.start() - 100):match.start()]:
            continue

        # 허용 폰트 조합 체크
        fonts_in_value = [f.strip().strip("'\"") for f in font_value.split(',')]

        if all(any(allowed in font for allowed in ALLOWED_FONTS) for font in fonts_in_value):
            # 모두 허용 범위 내 — pass
            continue

        # 토큰 사용 권장
        issues['font_direct'].append({
            'file': file_path,
            'line': line_num,
            'value': font_value,
            'message': '토큰 사용 권장'
        })

def check_font_size_consistency(file_path: str, content: str):
    """font-size 일관성 검사 (참고용)"""
    # 자주 등장하는 font-size 수집
    size_pattern = r'font-size:\s*([\d.]+(?:rem|px|em));'
    sizes = re.findall(size_pattern, content)

    # 통계
    size_counts = {}
    for size in sizes:
        size_counts[size] = size_counts.get(size, 0) + 1

    # 상위 5개
    top_sizes = sorted(size_counts.items(), key=lambda x: x[1], reverse=True)[:5]

    if top_sizes:
        issues['font_size'].append({
            'file': file_path,
            'top_sizes': top_sizes,
            'message': 'Font-size 사용 통계 (참고)'
        })

def check_letter_spacing(file_path: str, content: str):
    """letter-spacing 토큰 사용 검사"""
    ls_pattern = r'letter-spacing:\s*([^;]+);'
    matches = re.finditer(ls_pattern, content, re.IGNORECASE)

    allowed_ls_tokens = [
        'var(--ls-label)',
        'var(--ls-wide)',
        'var(--ls-heading)',
        'var(--ls-tight)'
    ]

    for match in matches:
        ls_value = match.group(1).strip()
        line_num = content[:match.start()].count('\n') + 1

        # 토큰 사용 확인
        if any(token in ls_value for token in allowed_ls_tokens):
            continue

        # :root 정의부 제외
        if ':root' in content[max(0, match.start() - 100):match.start()]:
            continue

        # inherit/normal은 허용
        if ls_value in ['inherit', 'normal', '0']:
            continue

        issues['letter_spacing'].append({
            'file': file_path,
            'line': line_num,
            'value': ls_value,
            'message': 'letter-spacing 토큰 사용 권장'
        })

def audit_file(file_path: str):
    """단일 파일 감사"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()

        check_font_family(file_path, content)
        check_letter_spacing(file_path, content)

        # font-size는 style.css만
        if 'style.css' in file_path:
            check_font_size_consistency(file_path, content)

    except Exception as e:
        print(f"  ✗ 오류: {file_path} — {str(e)}")

def main():
    website_root = Path('/Users/97layer/97layerOS/website')

    print("📐 타이포그래피 일관성 감사\n")
    print("=" * 60)

    # style.css 우선
    print("\n[1] style.css 감사")
    style_css = website_root / 'assets/css/style.css'
    if style_css.exists():
        audit_file(str(style_css))

    # 핵심 HTML 파일
    print("\n[2] 핵심 HTML 파일 감사")
    core_files = ['index.html', 'about.html', 'contact.html', 'offering.html', 'photography.html']

    for file in core_files:
        file_path = website_root / file
        if file_path.exists():
            print(f"  • {file}")
            audit_file(str(file_path))

    # Flask 템플릿
    print("\n[3] Flask 템플릿 감사")
    backend_dir = website_root / 'backend/templates'
    if backend_dir.exists():
        for template in backend_dir.glob('*.html'):
            print(f"  • {template.name}")
            audit_file(str(template))

    # 결과 출력
    print("\n" + "=" * 60)
    print(f"\n📊 감사 결과")
    print(f"  • font-family 직접 지정: {len(issues['font_direct'])}")
    print(f"  • letter-spacing 직접 값: {len(issues['letter_spacing'])}")
    print(f"  • font-size 통계: {len(issues['font_size'])} 파일")

    # JSON 저장
    output_file = '/tmp/typography_audit.json'
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(issues, f, ensure_ascii=False, indent=2)

    print(f"\n✅ 상세 보고서: {output_file}")

    # 주요 이슈 출력
    if issues['font_direct']:
        print(f"\n🔍 font-family 직접 지정 (상위 5개):")
        for issue in issues['font_direct'][:5]:
            print(f"  • {issue['file'].split('/')[-1]}:{issue['line']}")
            print(f"    | {issue['value']}")

    if issues['letter_spacing']:
        print(f"\n🔍 letter-spacing 직접 값 (상위 5개):")
        for issue in issues['letter_spacing'][:5]:
            print(f"  • {issue['file'].split('/')[-1]}:{issue['line']}")
            print(f"    | {issue['value']}")

if __name__ == '__main__':
    main()
