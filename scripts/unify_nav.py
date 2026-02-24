#!/usr/bin/env python3
"""
Nav 컴포넌트 통일 스크립트
82개 HTML 파일의 nav 구조를 표준화
"""

import os
import re
from pathlib import Path

# 표준 Nav 구조
STANDARD_NAV = '''<nav>
    <a href="/" class="nav-logo">
      <img src="assets/img/symbol.png" class="nav-symbol" alt="WOOHWAHAE">
    </a>
    <ul class="nav-links">
      <li><a href="archive/">Archive</a></li>
      <li><a href="offering.html">Offering</a></li>
      <li><a href="about.html">About</a></li>
      <li><a href="contact.html">Contact</a></li>
      <li><a href="lab/">Lab</a></li>
    </ul>
    <button class="nav-toggle" aria-label="Menu">
      <span></span><span></span><span></span>
    </button>
  </nav>'''

def extract_nav(content: str):
    """HTML에서 <nav>...</nav> 추출"""
    nav_match = re.search(r'<nav[^>]*>.*?</nav>', content, re.DOTALL)
    if nav_match:
        return nav_match.group(0)
    return None

def replace_nav(content: str, new_nav: str):
    """기존 nav를 새 nav로 교체"""
    # 기존 nav 패턴 찾기
    nav_pattern = r'<nav[^>]*>.*?</nav>'

    if re.search(nav_pattern, content, re.DOTALL):
        # nav 교체
        new_content = re.sub(nav_pattern, new_nav, content, count=1, flags=re.DOTALL)
        return new_content
    else:
        return None

def unify_nav_in_file(file_path: str, dry_run=True):
    """단일 파일의 nav 통일"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Flask 템플릿은 제외 (nav 없음)
        if 'backend/templates' in file_path:
            return False, "Skip (Flask template)"

        # 기존 nav 추출
        old_nav = extract_nav(content)

        if not old_nav:
            return False, "No nav found"

        # 이미 표준화되어 있는지 확인
        if '<li><a href="lab/">Lab</a></li>' in old_nav and '<button class="nav-toggle"' in old_nav:
            return False, "Already standard"

        # nav 교체
        new_content = replace_nav(content, STANDARD_NAV)

        if new_content:
            if not dry_run:
                with open(file_path, 'w', encoding='utf-8') as f:
                    f.write(new_content)

            return True, "Updated"
        else:
            return False, "Failed to replace"

    except Exception as e:
        return False, f"Error: {str(e)}"

def main():
    website_root = Path('/Users/97layer/97layerOS/website')

    print("🔧 Nav 컴포넌트 통일 스크립트\n")
    print("=" * 60)

    # 모든 HTML 파일 찾기
    html_files = list(website_root.rglob('*.html'))

    print(f"\n총 {len(html_files)}개 HTML 파일 발견")
    print("\n[DRY RUN] 변경 사항 미리보기...\n")

    updated = []
    skipped = []
    errors = []

    for file_path in html_files:
        result, message = unify_nav_in_file(str(file_path), dry_run=True)

        if result:
            updated.append(str(file_path))
            print(f"  ✓ {file_path.relative_to(website_root)}")
        elif "Error" in message:
            errors.append((str(file_path), message))
        else:
            skipped.append((str(file_path), message))

    print("\n" + "=" * 60)
    print(f"\n📊 결과:")
    print(f"  • 업데이트 필요: {len(updated)}")
    print(f"  • 건너뜀: {len(skipped)}")
    print(f"  • 오류: {len(errors)}")

    if updated:
        print(f"\n✅ 실제 적용하려면: python3 scripts/unify_nav.py --apply")

    # 건너뜀 상세 (상위 5개)
    if skipped:
        print(f"\n건너뜀 이유 (상위 5개):")
        for path, reason in skipped[:5]:
            file_name = Path(path).relative_to(website_root)
            print(f"  • {file_name}: {reason}")

    # 오류 상세
    if errors:
        print(f"\n⚠️  오류 발생:")
        for path, error in errors:
            file_name = Path(path).relative_to(website_root)
            print(f"  • {file_name}: {error}")

if __name__ == '__main__':
    import sys

    if '--apply' in sys.argv:
        print("\n⚠️  실제 파일 수정 중...\n")
        website_root = Path('/Users/97layer/97layerOS/website')
        html_files = list(website_root.rglob('*.html'))

        updated_count = 0
        for file_path in html_files:
            result, message = unify_nav_in_file(str(file_path), dry_run=False)
            if result:
                updated_count += 1
                print(f"  ✓ {file_path.relative_to(website_root)}")

        print(f"\n✅ {updated_count}개 파일 업데이트 완료")
    else:
        main()
