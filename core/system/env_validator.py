"""
97layerOS Environment Validator
================================
모든 에이전트/서비스 시작 시 호출.
필수 환경변수 누락 → 즉시 FATAL 출력 후 종료.
조용히 None으로 넘어가는 것 금지.

사용법:
    from core.system.env_validator import validate_env
    validate_env("telegram_secretary", REQUIRED_KEYS)
"""

import os
import sys
from typing import List, Optional


# 서비스별 필수 환경변수 목록
REQUIRED = {
    "telegram_secretary": [
        "TELEGRAM_BOT_TOKEN",
        "ADMIN_TELEGRAM_ID",
        "GOOGLE_API_KEY",
    ],
    "sa_agent": [
        "GOOGLE_API_KEY",
    ],
    "ce_agent": [
        "GOOGLE_API_KEY",
    ],
    "orchestrator": [
        "GOOGLE_API_KEY",
    ],
    "gardener": [
        "GOOGLE_API_KEY",
    ],
    # 공통 최소 요구
    "base": [
        "GOOGLE_API_KEY",
    ],
}


def validate_env(service_name: str, extra_keys: Optional[List[str]] = None) -> None:
    """
    필수 환경변수를 검증한다.
    누락된 키가 있으면 FATAL 로그 출력 후 sys.exit(1).

    Args:
        service_name: REQUIRED 딕셔너리의 키 (또는 "base")
        extra_keys:   서비스별 추가 검증이 필요한 키 목록
    """
    required_keys = REQUIRED.get(service_name, []) + REQUIRED.get("base", [])
    if extra_keys:
        required_keys = list(set(required_keys + extra_keys))

    missing = [k for k in required_keys if not os.getenv(k)]

    if missing:
        print(f"\n{'='*60}", flush=True)
        print(f"FATAL [{service_name}]: 필수 환경변수 누락", flush=True)
        for k in missing:
            print(f"  - {k} is not set", flush=True)
        print(f"\n.env 파일을 확인하세요: {os.getenv('PROJECT_ROOT', '~/97layerOS')}/.env", flush=True)
        print(f"{'='*60}\n", flush=True)
        sys.exit(1)

    # 정상 — 로드된 키 목록 출력 (디버깅 지원)
    print(f"[{service_name}] ENV OK — {len(required_keys)}개 변수 확인됨", flush=True)


def get_env(key: str, default: Optional[str] = None, required: bool = False) -> Optional[str]:
    """
    환경변수 안전 조회.
    required=True이면 없을 때 즉시 FATAL.
    """
    value = os.getenv(key, default)
    if required and not value:
        print(f"FATAL: 환경변수 {key} 가 설정되지 않았습니다.", flush=True)
        sys.exit(1)
    return value


def get_project_root() -> str:
    """PROJECT_ROOT 환경변수 반환. 없으면 파일 기반 추론."""
    from pathlib import Path
    env_root = os.getenv("PROJECT_ROOT")
    if env_root:
        return env_root
    # fallback: 이 파일 기준으로 3단계 위
    return str(Path(__file__).resolve().parent.parent.parent)


def get_site_base_url() -> str:
    """SITE_BASE_URL. 도메인 연결 전/후 단일 지점 관리."""
    return os.getenv("SITE_BASE_URL", "https://woohwahae.kr")


def get_archive_path() -> str:
    """SITE_ARCHIVE_PATH. website/archive 상대경로."""
    return os.getenv("SITE_ARCHIVE_PATH", "website/archive")
