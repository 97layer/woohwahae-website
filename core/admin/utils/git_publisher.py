"""
Git Publisher — website/ 변경사항을 GitHub Pages에 자동 배포
"""

import subprocess
from datetime import datetime
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent.parent.parent  # 97layerOS root


def publish_to_website() -> tuple[bool, str]:
    """
    website/ 디렉토리의 변경사항을 git commit + push.
    Returns (success: bool, message: str)
    """
    try:
        # 1. Stage website changes
        result = subprocess.run(
            ['git', 'add', 'website/'],
            cwd=str(BASE_DIR),
            capture_output=True, text=True, timeout=30
        )
        if result.returncode != 0:
            return False, f'git add 실패: {result.stderr}'

        # 2. Check if there are changes to commit
        result = subprocess.run(
            ['git', 'diff', '--cached', '--quiet', '--', 'website/'],
            cwd=str(BASE_DIR),
            capture_output=True, text=True, timeout=10
        )
        if result.returncode == 0:
            return True, '변경사항이 없습니다. 이미 최신 상태입니다.'

        # 3. Commit
        now = datetime.now().strftime('%Y-%m-%d %H:%M')
        commit_msg = f'publish: {now}'
        result = subprocess.run(
            ['git', 'commit', '-m', commit_msg],
            cwd=str(BASE_DIR),
            capture_output=True, text=True, timeout=30
        )
        if result.returncode != 0:
            return False, f'git commit 실패: {result.stderr}'

        # 4. Push to website branch
        result = subprocess.run(
            ['git', 'push', 'origin', 'main:website'],
            cwd=str(BASE_DIR),
            capture_output=True, text=True, timeout=60
        )
        if result.returncode != 0:
            return False, f'git push 실패: {result.stderr}'

        return True, f'발행 완료! ({now})'

    except subprocess.TimeoutExpired:
        return False, 'Git 명령 타임아웃. 네트워크를 확인해주세요.'
    except Exception as e:
        return False, f'오류: {str(e)}'
