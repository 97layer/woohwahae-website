#!/usr/bin/env python3
"""
Podman Secrets 읽기 헬퍼 스크립트
환경변수에 *_FILE 경로가 있으면 파일에서 읽어서 환경변수로 설정

Usage:
    # Dockerfile CMD 앞에 추가
    CMD ["python3", "/usr/local/bin/read_secrets.py", "&&", "python3", "execution/system/nightguard_daemon.py"]

    # 또는 Python 코드에서 직접 호출
    from deployment.read_secrets import load_secrets
    load_secrets()
"""

import os
from pathlib import Path


def load_secrets():
    """
    *_FILE 환경변수를 읽어서 실제 값을 환경변수로 설정

    Example:
        TELEGRAM_BOT_TOKEN_FILE=/run/secrets/telegram_bot_token
        → TELEGRAM_BOT_TOKEN=<파일 내용>
    """
    for key, value in os.environ.items():
        if key.endswith("_FILE"):
            # 원본 키 이름 (_FILE 제거)
            original_key = key[:-5]  # "_FILE" 제거

            # 파일 경로
            secret_file = Path(value)

            if secret_file.exists():
                try:
                    # 파일에서 값 읽기
                    secret_value = secret_file.read_text().strip()

                    # 환경변수로 설정
                    os.environ[original_key] = secret_value

                    print(f"✓ Loaded secret: {original_key} from {secret_file}")

                except Exception as e:
                    print(f"⚠️ Failed to load secret {original_key}: {e}")
            else:
                print(f"⚠️ Secret file not found: {secret_file}")


if __name__ == "__main__":
    load_secrets()
