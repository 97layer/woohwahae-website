"""
Filename: execution/system/update_checklist.py
Author: 97LAYER System
Date: 2026-02-12

Purpose: Markdown 체크리스트 자동 업데이트
- Gemini Brain의 task.md 파일 파싱
- 작업 완료 시 자동으로 [x] 체크
- 상태와 일관성 유지
"""

import os
import re
from pathlib import Path
from datetime import datetime

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent


def parse_checklist(content):
    """
    Markdown 체크리스트 파싱

    Returns:
        list: [{'checked': bool, 'text': str, 'line_num': int}]
    """
    items = []
    lines = content.split('\n')

    for i, line in enumerate(lines):
        # - [ ] or - [x] 패턴 매칭
        match = re.match(r'^(\s*)- \[([ x])\] (.+)$', line)
        if match:
            indent, status, text = match.groups()
            items.append({
                'line_num': i,
                'indent': indent,
                'checked': status.lower() == 'x',
                'text': text.strip(),
                'original': line
            })

    return items


def update_checklist_item(content, item_text, checked=True):
    """
    특정 체크리스트 항목 업데이트

    Args:
        content (str): Markdown 파일 내용
        item_text (str): 업데이트할 항목 텍스트 (부분 매칭)
        checked (bool): True면 [x], False면 [ ]

    Returns:
        str: 업데이트된 내용
    """
    lines = content.split('\n')
    updated = False

    for i, line in enumerate(lines):
        # 체크박스 패턴 매칭
        match = re.match(r'^(\s*- \[)([ x])(\] )(.+)$', line)
        if match:
            prefix, current_status, suffix, text = match.groups()

            # 텍스트 부분 매칭
            if item_text.lower() in text.lower():
                new_status = 'x' if checked else ' '
                lines[i] = f"{prefix}{new_status}{suffix}{text}"
                updated = True
                print(f"  ✓ Updated: {text[:50]}...")

    if not updated:
        print(f"  ✗ Not found: {item_text}")

    return '\n'.join(lines)


def check_all_items(content, keywords):
    """
    여러 항목을 한번에 체크

    Args:
        content (str): Markdown 내용
        keywords (list): 체크할 항목 키워드 리스트

    Returns:
        str: 업데이트된 내용
    """
    for keyword in keywords:
        content = update_checklist_item(content, keyword, checked=True)
    return content


def uncheck_all_items(content, keywords):
    """여러 항목을 한번에 체크 해제"""
    for keyword in keywords:
        content = update_checklist_item(content, keyword, checked=False)
    return content


def get_checklist_status(file_path):
    """
    체크리스트 현황 조회

    Returns:
        dict: {'total': int, 'completed': int, 'pending': int, 'items': list}
    """
    if not Path(file_path).exists():
        return None

    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    items = parse_checklist(content)
    completed = sum(1 for item in items if item['checked'])
    pending = len(items) - completed

    return {
        'total': len(items),
        'completed': completed,
        'pending': pending,
        'items': items,
        'progress': f"{completed}/{len(items)}" if len(items) > 0 else "0/0"
    }


def update_file(file_path, item_text, checked=True):
    """
    파일의 체크리스트 항목 업데이트 및 저장

    Args:
        file_path (str): 대상 파일 경로
        item_text (str): 업데이트할 항목
        checked (bool): 체크 상태

    Returns:
        bool: 성공 여부
    """
    file_path = Path(file_path)

    if not file_path.exists():
        print(f"[ERROR] File not found: {file_path}")
        return False

    try:
        # 파일 읽기
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # 업데이트
        updated_content = update_checklist_item(content, item_text, checked)

        # 변경사항 확인
        if content == updated_content:
            return False

        # 파일 쓰기
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(updated_content)

        print(f"[{datetime.now()}] Updated: {file_path.name}")
        return True

    except Exception as e:
        print(f"[ERROR] Failed to update {file_path}: {e}")
        return False


def find_gemini_brain_tasks():
    """
    Gemini Brain 폴더에서 task.md 파일 찾기

    Returns:
        list: task.md 파일 경로 리스트
    """
    gemini_brain = Path.home() / ".gemini" / "antigravity" / "brain"

    if not gemini_brain.exists():
        return []

    task_files = []
    for brain_dir in gemini_brain.iterdir():
        if brain_dir.is_dir():
            task_file = brain_dir / "task.md"
            if task_file.exists():
                task_files.append(task_file)

    return sorted(task_files, key=lambda x: x.stat().st_mtime, reverse=True)


def sync_with_status():
    """
    status.json과 체크리스트 자동 동기화

    status.json의 completed_tasks를 읽어서
    해당하는 항목을 자동으로 체크
    """
    import json

    status_file = PROJECT_ROOT / "knowledge" / "status.json"

    if not status_file.exists():
        print("[WARN] status.json not found")
        return

    with open(status_file, 'r', encoding='utf-8') as f:
        status = json.load(f)

    completed = status.get('completed_tasks', [])

    # 최근 Gemini Brain task.md 찾기
    task_files = find_gemini_brain_tasks()

    if not task_files:
        print("[WARN] No Gemini Brain task.md found")
        return

    # 가장 최근 task.md 업데이트
    latest_task = task_files[0]
    print(f"\n[INFO] Syncing with: {latest_task}")

    with open(latest_task, 'r', encoding='utf-8') as f:
        content = f.read()

    # Completed tasks를 체크
    updated_content = check_all_items(content, completed)

    # 저장
    with open(latest_task, 'w', encoding='utf-8') as f:
        f.write(updated_content)

    print(f"[{datetime.now()}] Sync complete")


# CLI 인터페이스
if __name__ == "__main__":
    import sys

    if len(sys.argv) < 2:
        print("Usage:")
        print("  python3 update_checklist.py status <file>")
        print("  python3 update_checklist.py check <file> <item_text>")
        print("  python3 update_checklist.py uncheck <file> <item_text>")
        print("  python3 update_checklist.py sync")
        print("\nExamples:")
        print("  python3 update_checklist.py status ~/.gemini/.../task.md")
        print("  python3 update_checklist.py check task.md 'Create snapshot'")
        print("  python3 update_checklist.py sync")
        sys.exit(1)

    command = sys.argv[1]

    if command == "status":
        if len(sys.argv) < 3:
            print("[ERROR] File path required")
            sys.exit(1)

        file_path = sys.argv[2]
        status = get_checklist_status(file_path)

        if status:
            print(f"\n=== Checklist Status: {Path(file_path).name} ===")
            print(f"Progress: {status['progress']} ({status['completed']}/{status['total']})")
            print(f"Completed: {status['completed']}")
            print(f"Pending: {status['pending']}")
            print("\nItems:")
            for item in status['items']:
                status_icon = "✓" if item['checked'] else "○"
                print(f"  {status_icon} {item['text']}")
        else:
            print("[ERROR] Failed to get status")

    elif command == "check":
        if len(sys.argv) < 4:
            print("[ERROR] File path and item text required")
            sys.exit(1)

        file_path = sys.argv[2]
        item_text = sys.argv[3]

        if update_file(file_path, item_text, checked=True):
            print(f"✓ Checked: {item_text}")
        else:
            print(f"✗ Failed to check: {item_text}")

    elif command == "uncheck":
        if len(sys.argv) < 4:
            print("[ERROR] File path and item text required")
            sys.exit(1)

        file_path = sys.argv[2]
        item_text = sys.argv[3]

        if update_file(file_path, item_text, checked=False):
            print(f"○ Unchecked: {item_text}")
        else:
            print(f"✗ Failed to uncheck: {item_text}")

    elif command == "sync":
        sync_with_status()

    else:
        print(f"[ERROR] Unknown command: {command}")
        sys.exit(1)
