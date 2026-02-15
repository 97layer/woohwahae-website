#!/usr/bin/env python3
"""
하드코딩된 절대 경로를 상대 경로로 전환
포드맨 환경 호환성 확보
"""

import re
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
HARDCODED = "/Users/97layer/97layerOS"

TARGET_FILES = [
    "execution/archive_daemon.py",
    "execution/clipboard_sentinel.py",
    "execution/create_snapshot.py",
    "execution/snapshot_daemon.py"
]

def fix_file(filepath: Path):
    """파일 내 하드코딩 경로를 동적 경로로 변환"""
    try:
        content = filepath.read_text(encoding='utf-8')

        # 하드코딩 경로 확인
        if HARDCODED not in content:
            print(f"✅ {filepath.name}: No hardcoded paths")
            return

        # PROJECT_ROOT 설정이 이미 있는지 확인
        has_project_root = "PROJECT_ROOT = Path(__file__)" in content

        # 경로 치환
        fixed_content = content.replace(
            f'"{HARDCODED}"',
            'str(PROJECT_ROOT)'
        ).replace(
            f"'{HARDCODED}'",
            'str(PROJECT_ROOT)'
        ).replace(
            f'"{HARDCODED}',
            'f"{PROJECT_ROOT}'
        ).replace(
            f"'{HARDCODED}",
            "f'{PROJECT_ROOT}"
        )

        # PROJECT_ROOT 정의 추가 (필요 시)
        if not has_project_root:
            lines = fixed_content.split('\n')
            import_idx = 0

            # import 섹션 찾기
            for i, line in enumerate(lines):
                if line.startswith('import ') or line.startswith('from '):
                    import_idx = i + 1
                elif import_idx > 0 and not line.strip().startswith(('import', 'from')):
                    break

            # Path import 확인 및 추가
            if 'from pathlib import Path' not in fixed_content:
                lines.insert(import_idx, 'from pathlib import Path')
                import_idx += 1

            # PROJECT_ROOT 정의 추가
            lines.insert(import_idx, '')
            lines.insert(import_idx + 1, '# 동적 경로 설정 (포드맨 호환)')
            lines.insert(import_idx + 2, 'PROJECT_ROOT = Path(__file__).resolve().parent.parent')
            lines.insert(import_idx + 3, '')

            fixed_content = '\n'.join(lines)

        # 파일 저장
        filepath.write_text(fixed_content, encoding='utf-8')
        print(f"🔧 {filepath.name}: Fixed hardcoded paths")

    except Exception as e:
        print(f"❌ Error fixing {filepath.name}: {e}")

def main():
    print("🔍 Fixing hardcoded paths for Podman compatibility...\n")

    for target in TARGET_FILES:
        filepath = PROJECT_ROOT / target
        if filepath.exists():
            fix_file(filepath)
        else:
            print(f"⚠️  {target}: File not found")

    print("\n✅ Path abstraction complete")

if __name__ == "__main__":
    main()