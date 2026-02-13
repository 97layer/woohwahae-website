"""Auto-promote Knowledge to Directive"""
import os
import json
from pathlib import Path
from datetime import datetime
from collections import Counter

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
KNOWLEDGE_DIR = PROJECT_ROOT / "knowledge"
DIRECTIVES_DIR = PROJECT_ROOT / "directives"

def scan_knowledge():
    """Scan knowledge for promotion candidates"""
    sessions = KNOWLEDGE_DIR / "sessions"
    patterns = KNOWLEDGE_DIR / "patterns"

    keywords = Counter()
    files = {}

    for md_file in sessions.glob("*.md"):
        with open(md_file, 'r', encoding='utf-8') as f:
            content = f.read()
            for word in ['snapshot', 'sync', 'scrape', 'backup', 'venv', 'daemon']:
                if word in content.lower():
                    keywords[word] += 1
                    if word not in files:
                        files[word] = []
                    files[word].append(md_file.name)

    candidates = []
    for word, count in keywords.items():
        if count >= 3:
            candidates.append({
                'keyword': word,
                'frequency': count,
                'files': files[word][:3],
                'recommendation': 'Directive 승격 권장'
            })

    return candidates

def create_directive_template(keyword):
    """Generate directive template"""
    return f"""---
type: directive
status: draft
priority: medium
created: {datetime.now().strftime("%Y-%m-%d")}
auto_generated: true
---

# {keyword.title()} Workflow

## 1. 목적
[자동 생성됨 - 수동 완성 필요]

## 2. 입력
-

## 3. 도구
- execution/{keyword}_*.py

## 4. 실행 절차
1.

## 5. 출력
-

## 6. 예외 처리
-

## 7. Self-Annealing 이력
- {datetime.now().strftime("%Y-%m-%d")}: Auto-generated from pattern analysis
"""

def promote(keyword):
    """Promote to directive"""
    directive_path = DIRECTIVES_DIR / f"{keyword}_workflow.md"
    if directive_path.exists():
        return False

    content = create_directive_template(keyword)
    directive_path.write_text(content, encoding='utf-8')
    return True

if __name__ == "__main__":
    candidates = scan_knowledge()
    print(f"=== Promotion Candidates ({len(candidates)}) ===")
    for c in candidates:
        print(f"  {c['keyword']}: {c['frequency']}x - {c['recommendation']}")
        if c['frequency'] >= 3:
            if promote(c['keyword']):
                print(f"    ✓ Created directive: {c['keyword']}_workflow.md")
