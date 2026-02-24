#!/usr/bin/env python3
"""
LAYER OS Filesystem Guard — 자가 진화 하네스

규칙 소스: knowledge/system/guard_rules.json (Gardener가 자동 업데이트)
동작:
  1. guard_rules.json에서 규칙 로드 (15분마다 핫리로드)
  2. 15초마다 프로젝트 전체 스캔
  3. 위반 파일 → quarantine 이동
  4. 위반 패턴 빈도를 guard_rules.json에 기록 → Gardener가 룰 진화에 활용

Usage:
    python3 core/system/filesystem_guard.py
"""

import json
import logging
import os
import shutil
import sys
import tempfile
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional

logger = logging.getLogger("filesystem_guard")
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [guard] %(levelname)s %(message)s",
    handlers=[logging.StreamHandler(sys.stdout)],
)

PROJECT_ROOT = Path(__file__).parent.parent.parent.resolve()
RULES_PATH = PROJECT_ROOT / "knowledge/system/guard_rules.json"
QUARANTINE_DIR = PROJECT_ROOT / "knowledge/docs/archive/quarantine"
LOG_PATH = PROJECT_ROOT / ".infra/logs/filesystem_guard.log"

SCAN_INTERVAL = 15       # 초 — 파일시스템 스캔 주기
RELOAD_INTERVAL = 900    # 초 — 규칙 파일 핫리로드 주기 (15분)

# ── 파일 로거 초기화 ──────────────────────────────────────────────
try:
    LOG_PATH.parent.mkdir(parents=True, exist_ok=True)
    file_handler = logging.FileHandler(LOG_PATH)
    file_handler.setFormatter(logging.Formatter(
        "%(asctime)s [guard] %(levelname)s %(message)s"
    ))
    logger.addHandler(file_handler)
except Exception:
    pass  # 로그 파일 실패해도 stdout은 유지


class GuardRules:
    """guard_rules.json 래퍼 — 로드, 검증, 패턴 기록 담당."""

    def __init__(self):
        self._data: dict = {}
        self._loaded_at: float = 0.0
        self.load()

    def load(self) -> bool:
        """규칙 파일 로드. 실패 시 기존 규칙 유지. 성공 여부 반환."""
        try:
            raw = RULES_PATH.read_text(encoding="utf-8")
            data = json.loads(raw)
            # 최소 키 검증
            required = {"allowed_top_dirs", "allowed_root_files",
                        "forbidden_name_prefixes", "allowed_reports_prefixes"}
            if not required.issubset(data.keys()):
                raise ValueError("필수 키 누락")
            self._data = data
            self._loaded_at = time.time()
            return True
        except Exception as exc:
            logger.error("guard_rules.json 로드 실패 (기존 규칙 유지): %s", exc)
            return False

    def reload_if_stale(self) -> None:
        if time.time() - self._loaded_at > RELOAD_INTERVAL:
            if self.load():
                logger.info("guard_rules.json 핫리로드 완료")

    @property
    def allowed_top_dirs(self) -> set:
        return set(self._data.get("allowed_top_dirs", []))

    @property
    def allowed_root_files(self) -> set:
        return set(self._data.get("allowed_root_files", []))

    @property
    def forbidden_prefixes(self) -> tuple:
        return tuple(self._data.get("forbidden_name_prefixes", []))

    @property
    def allowed_reports_prefixes(self) -> tuple:
        return tuple(self._data.get("allowed_reports_prefixes", []))

    def record_violation(self, pattern: str) -> None:
        """위반 패턴 빈도 기록 — Gardener가 룰 진화에 활용."""
        try:
            # 최신 파일 읽기 → 패턴 카운트 업데이트 → atomic 저장
            raw = RULES_PATH.read_text(encoding="utf-8")
            data = json.loads(raw)
            patterns = data.setdefault("violation_patterns", {})
            entry = patterns.setdefault(pattern, {"count": 0, "first_seen": "", "last_seen": ""})
            entry["count"] += 1
            entry["last_seen"] = datetime.now(timezone.utc).isoformat()
            if not entry["first_seen"]:
                entry["first_seen"] = entry["last_seen"]

            # atomic write (tmp → rename)
            tmp = RULES_PATH.with_suffix(".tmp")
            tmp.write_text(json.dumps(data, ensure_ascii=False, indent=2), encoding="utf-8")
            tmp.replace(RULES_PATH)
        except Exception as exc:
            logger.warning("violation_patterns 기록 실패: %s", exc)


rules = GuardRules()


def _extract_pattern(name: str) -> str:
    """파일명에서 패턴 키 추출 (접두사 최대 3단어)."""
    parts = name.replace("-", "_").split("_")
    return "_".join(parts[:3]) + "_"


def is_violation(path: Path) -> tuple[bool, str]:
    """FILESYSTEM_MANIFEST 규칙 위반 여부 검사.

    Returns:
        (violated: bool, reason: str)
    """
    try:
        rel = path.relative_to(PROJECT_ROOT)
    except ValueError:
        return False, ""

    parts = rel.parts
    if not parts:
        return False, ""

    name = path.name

    # 숨김 파일/디렉토리는 허용
    if name.startswith(".") or parts[0].startswith("."):
        return False, ""

    # 디렉토리 자체는 검사 안 함
    if path.is_dir():
        return False, ""

    # ── 루트 파일 검사 ───────────────────────────────────────────
    if len(parts) == 1:
        if name in rules.allowed_root_files:
            return False, ""
        return True, "루트 파일 생성 금지"

    top_dir = parts[0]

    # ── 허용 최상위 디렉토리 ──────────────────────────────────────
    if top_dir not in rules.allowed_top_dirs:
        return True, "허용 외 최상위 디렉토리: %s" % top_dir

    # ── knowledge/reports/ 허용 type 검사 ───────────────────────
    if (parts[0] == "knowledge" and len(parts) >= 3
            and parts[1] == "reports"):
        if parts[2] == "growth":
            return False, ""
        if name.endswith(".md"):
            if not name.startswith(rules.allowed_reports_prefixes):
                return True, "reports/ 허용 type 아님 (%s)" % str(rules.allowed_reports_prefixes)

    # ── 금지 파일명 접두사 ────────────────────────────────────────
    if name.startswith(rules.forbidden_prefixes):
        return True, "금지 파일명 패턴: %s" % name

    return False, ""


def quarantine(path: Path, reason: str) -> None:
    """위반 파일 quarantine 이동 + 패턴 기록."""
    QUARANTINE_DIR.mkdir(parents=True, exist_ok=True)
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    dest = QUARANTINE_DIR / ("%s__%s" % (timestamp, path.name))
    try:
        shutil.move(str(path), str(dest))
        logger.warning(
            "QUARANTINE | %s | %s",
            path.relative_to(PROJECT_ROOT), reason,
        )
        # 패턴 빈도 기록
        pattern = _extract_pattern(path.name)
        rules.record_violation(pattern)
    except Exception as exc:
        logger.error("quarantine 이동 실패: %s — %s", path, exc)


def scan_once() -> int:
    """전체 스캔 1회. 위반 건수 반환."""
    violations = 0
    skip_dirs = {".git", "__pycache__", ".pytest_cache", ".venv"}
    for root, dirs, files in os.walk(PROJECT_ROOT):
        dirs[:] = [d for d in dirs if d not in skip_dirs]
        for fname in files:
            fpath = Path(root) / fname
            violated, reason = is_violation(fpath)
            if violated:
                quarantine(fpath, reason)
                violations += 1
    return violations


def watch() -> None:
    """메인 루프 — 스캔 + 핫리로드."""
    logger.info(
        "Filesystem Guard 시작 | scan=%ds reload=%ds | root=%s",
        SCAN_INTERVAL, RELOAD_INTERVAL, PROJECT_ROOT,
    )
    while True:
        try:
            rules.reload_if_stale()
            count = scan_once()
            if count:
                logger.info("위반 %d건 격리", count)
        except Exception as exc:
            logger.error("스캔 오류: %s", exc)
        time.sleep(SCAN_INTERVAL)


if __name__ == "__main__":
    initial = scan_once()
    if initial:
        logger.info("초기 스캔 — 위반 %d건 격리", initial)
    watch()
