#!/usr/bin/env python3
"""
web_consistency_lock.py — 웹 작업 일관성 보장 시스템

웹페이지 수정 전 반드시 실행하여 일관성 깨짐 방지.
에이전트들이 중복/충돌 작업하지 않도록 Lock 관리.

사용법:
    python core/system/web_consistency_lock.py --check     # 현재 잠금 상태 확인
    python core/system/web_consistency_lock.py --acquire   # 잠금 획득 (작업 시작)
    python core/system/web_consistency_lock.py --release   # 잠금 해제 (작업 완료)
    python core/system/web_consistency_lock.py --validate  # 변경사항 검증
"""

import hashlib
import json
import logging
import sys
from datetime import datetime
from pathlib import Path

logging.basicConfig(level=logging.INFO, format="%(message)s")
logger = logging.getLogger(__name__)

PROJECT_ROOT = Path(__file__).resolve().parents[2]
LOCK_FILE = PROJECT_ROOT / "knowledge" / "system" / "web_work_lock.json"
WEBSITE_DIR = PROJECT_ROOT / "website"
STYLE_CSS = WEBSITE_DIR / "assets" / "css" / "style.css"
PRACTICE_MD = PROJECT_ROOT / "directives" / "practice.md"

# 웹 작업 권한 매트릭스
WEB_PERMISSIONS = {
    "AD": ["style", "layout", "components", "visual"],  # Archive Designer 전담
    "CE": ["content", "text"],  # Corpus Editor는 텍스트만
    "SA": [],  # SAGE는 웹 직접 수정 금지
    "HUMAN": ["*"]  # 사람은 모든 권한
}

# 어조 규칙 (practice.md Part II-8 기반)
TONE_RULES = {
    "archive": "한다체",  # 에세이, 독백적
    "practice": "합니다체",  # 서비스, 고객 지향
    "about": "합니다체",  # 소개, 공식적
    "home": "한다체"  # 홈은 에세이 스타일
}


def get_file_hash(filepath: Path) -> str:
    """파일의 MD5 해시 계산."""
    if not filepath.exists():
        return ""
    content = filepath.read_bytes()
    return hashlib.md5(content).hexdigest()


def check_lock() -> dict:
    """현재 잠금 상태 확인."""
    if not LOCK_FILE.exists():
        return {"locked": False}

    with open(LOCK_FILE, "r") as f:
        return json.load(f)


def acquire_lock(agent_id: str, task: str) -> bool:
    """잠금 획득 시도."""
    current = check_lock()

    if current.get("locked"):
        # 2시간 이상 방치된 Lock은 자동 해제
        from datetime import timedelta
        started = datetime.fromisoformat(current["started_at"])
        if datetime.now() - started > timedelta(hours=2):
            logger.warning("⚠️  2시간 이상 방치된 Lock 자동 해제: %s", current['agent'])
            LOCK_FILE.unlink()
        else:
            logger.error("❌ 잠금 실패: %s가 작업 중", current['agent'])
            logger.error("   진행 중: %s", current['task'])
            logger.error("   시작: %s", current['started_at'])
            return False

    # 현재 상태 스냅샷
    lock_data = {
        "locked": True,
        "agent": agent_id,
        "task": task,
        "started_at": datetime.now().isoformat(),
        "initial_state": {
            "style_hash": get_file_hash(STYLE_CSS),
            "components": list(WEBSITE_DIR.glob("_components/*.html")),
            "practice_hash": get_file_hash(PRACTICE_MD)
        }
    }

    LOCK_FILE.parent.mkdir(parents=True, exist_ok=True)
    with open(LOCK_FILE, "w") as f:
        json.dump(lock_data, f, indent=2, default=str)

    logger.info("✅ 잠금 획득: %s", agent_id)
    logger.info("   작업: %s", task)
    return True


def release_lock(agent_id: str) -> bool:
    """잠금 해제."""
    current = check_lock()

    if not current.get("locked"):
        logger.warning("⚠️  잠금이 없음")
        return True

    if current["agent"] != agent_id:
        logger.error("❌ 해제 권한 없음: %s의 잠금", current['agent'])
        return False

    # 변경사항 기록
    final_state = {
        "style_hash": get_file_hash(STYLE_CSS),
        "practice_hash": get_file_hash(PRACTICE_MD),
        "completed_at": datetime.now().isoformat()
    }

    # 히스토리 저장
    history_file = LOCK_FILE.parent / "web_work_history.jsonl"
    with open(history_file, "a") as f:
        history_entry = {
            **current,
            "final_state": final_state
        }
        f.write(json.dumps(history_entry, default=str) + "\n")

    # 잠금 해제
    LOCK_FILE.unlink()
    logger.info("✅ 잠금 해제: %s", agent_id)
    return True


def validate_changes(agent_id: str) -> bool:
    """변경사항 검증."""
    current = check_lock()

    if not current.get("locked"):
        logger.error("❌ 검증 실패: 잠금 없이 작업함")
        return False

    if current["agent"] != agent_id:
        logger.error("❌ 검증 실패: 다른 에이전트(%s)의 작업", current['agent'])
        return False

    # 스타일 변경 검증
    initial_hash = current["initial_state"]["style_hash"]
    current_hash = get_file_hash(STYLE_CSS)

    if initial_hash != current_hash:
        if agent_id not in ["AD", "HUMAN"]:
            logger.error("❌ 권한 위반: %s는 style.css 수정 불가", agent_id)
            return False
        logger.info("✓ style.css 변경됨 (권한 있음)")

    # 어조 일관성 검증
    for section, expected_tone in TONE_RULES.items():
        section_files = list(WEBSITE_DIR.glob("%s/**/*.html" % section))
        if not section_files:
            section_files = [WEBSITE_DIR / ("%s.html" % section)]

        for file in section_files:
            if not file.exists():
                continue
            content = file.read_text(encoding="utf-8")

            if expected_tone == "한다체" and "합니다" in content:
                logger.warning("⚠️  어조 불일치: %s에 합니다체 발견 (기대: 한다체)", file.name)
            elif expected_tone == "합니다체" and "한다." in content:
                logger.warning("⚠️  어조 불일치: %s에 한다체 발견 (기대: 합니다체)", file.name)

    logger.info("✅ 검증 완료")
    return True


def main():
    """CLI 엔트리포인트."""
    import argparse

    parser = argparse.ArgumentParser(description="웹 작업 일관성 Lock 시스템")
    parser.add_argument("--check", action="store_true", help="잠금 상태 확인")
    parser.add_argument("--acquire", metavar="AGENT", help="잠금 획득 (에이전트 ID)")
    parser.add_argument("--release", metavar="AGENT", help="잠금 해제 (에이전트 ID)")
    parser.add_argument("--validate", metavar="AGENT", help="변경사항 검증")
    parser.add_argument("--task", help="작업 설명 (acquire와 함께 사용)")

    args = parser.parse_args()

    if args.check:
        status = check_lock()
        if status.get("locked"):
            logger.info("🔒 잠금 중: %s", status['agent'])
            logger.info("   작업: %s", status['task'])
            logger.info("   시작: %s", status['started_at'])
        else:
            logger.info("🔓 잠금 없음 (작업 가능)")

    elif args.acquire:
        task = args.task or "웹페이지 수정"
        if acquire_lock(args.acquire, task):
            sys.exit(0)
        else:
            sys.exit(1)

    elif args.release:
        if release_lock(args.release):
            sys.exit(0)
        else:
            sys.exit(1)

    elif args.validate:
        if validate_changes(args.validate):
            sys.exit(0)
        else:
            sys.exit(1)

    else:
        parser.print_help()


if __name__ == "__main__":
    main()
