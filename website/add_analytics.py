#!/usr/bin/env python3
"""
모든 HTML 파일에 Google Analytics 추가
"""

import os
import re
from pathlib import Path

# Analytics 스크립트 태그
analytics_tag = '  <script src="assets/js/analytics.js"></script>\n'

# HTML 파일 찾기
html_files = [
    'index.html',
    'about.html',
    'shop.html',
    'atelier.html',
    'playlist.html',
    'project.html',
    'photography.html',
    'contact.html'
]

for filename in html_files:
    filepath = Path(filename)
    if not filepath.exists():
        print(f"⚠️  {filename} not found")
        continue

    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    # 이미 analytics.js가 포함되어 있는지 확인
    if 'analytics.js' in content:
        print(f"✓ {filename} already has analytics")
        continue

    # </head> 태그 바로 전에 추가
    if '</head>' in content:
        content = content.replace('</head>', f'{analytics_tag}</head>')
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"✓ Added analytics to {filename}")
    else:
        print(f"⚠️  {filename} has no </head> tag")

print("\n✅ Google Analytics 추가 완료!")
print("⚠️  주의: GA_MEASUREMENT_ID를 실제 Google Analytics ID로 변경해야 합니다.")
print("Google Analytics에서 새 속성을 만들고 측정 ID(G-XXXXXXXXXX)를 받아 교체하세요.")