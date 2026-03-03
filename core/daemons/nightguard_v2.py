#!/usr/bin/env python3
"""
Nightguard v2 - Quota Tracking Daemon
사용 로그 기반 실제 할당량 추적 데몬
"""

import os
import sys
import json
import logging
from pathlib import Path
from datetime import datetime, date
from collections import defaultdict
from typing import Dict, Any, Optional

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

# 경로 설정
LOGS_DIR = PROJECT_ROOT / "knowledge" / "logs"
QUOTA_STATE_FILE = PROJECT_ROOT / "knowledge" / "system" / "quota_state.json"
USAGE_LOG_GLOB = "usage_*.jsonl"

# 기본 할당량 (환경변수로 오버라이드 가능)
DEFAULT_QUOTAS: Dict[str, int] = {
    "api_calls_per_day": int(os.getenv("QUOTA_API_CALLS_PER_DAY", "500")),
    "tokens_per_day": int(os.getenv("QUOTA_TOKENS_PER_DAY", "100000")),
    "images_per_day": int(os.getenv("QUOTA_IMAGES_PER_DAY", "50")),
    "youtube_per_day": int(os.getenv("QUOTA_YOUTUBE_PER_DAY", "20")),
}

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger("nightguard_v2")


def _load_quota_state() -> Dict[str, Any]:
    """현재 할당량 상태 로드. 파일 없으면 빈 상태 반환."""
    if QUOTA_STATE_FILE.exists():
        try:
            with open(QUOTA_STATE_FILE, "r", encoding="utf-8") as f:
                return json.load(f)
        except Exception as e:
            logger.warning("quota_state 로드 실패: %s", e)
    return {}


def _save_quota_state(state: Dict[str, Any]) -> None:
    """할당량 상태 저장."""
    try:
        QUOTA_STATE_FILE.parent.mkdir(parents=True, exist_ok=True)
        state["updated_at"] = datetime.now().isoformat()
        with open(QUOTA_STATE_FILE, "w", encoding="utf-8") as f:
            json.dump(state, f, indent=2, ensure_ascii=False)
    except Exception as e:
        logger.error("quota_state 저장 실패: %s", e)


def _parse_usage_logs(target_date: Optional[str] = None) -> Dict[str, int]:
    """
    usage_*.jsonl 파일을 파싱해 오늘(또는 target_date)의 사용량 집계.

    각 로그 레코드 예시:
    {"ts": "2026-01-15T10:30:00", "type": "api_call", "tokens": 1200, "model": "claude-3"}
    {"ts": "2026-01-15T11:00:00", "type": "image", "count": 1}
    {"ts": "2026-01-15T12:00:00", "type": "youtube", "count": 1}
    """
    if target_date is None:
        target_date = date.today().isoformat()  # "2026-01-15"

    counters: Dict[str, int] = defaultdict(int)

    if not LOGS_DIR.exists():
        logger.info("logs 디렉토리 없음, 사용량 0으로 반환: %s", LOGS_DIR)
        return dict(counters)

    log_files = list(LOGS_DIR.glob(USAGE_LOG_GLOB))
    logger.info("사용 로그 파일 %d개 발견", len(log_files))

    for log_file in log_files:
        try:
            with open(log_file, "r", encoding="utf-8") as f:
                for lineno, line in enumerate(f, 1):
                    line = line.strip()
                    if not line:
                        continue
                    try:
                        record = json.loads(line)
                    except json.JSONDecodeError as e:
                        logger.debug("JSON 파싱 오류 %s:%d — %s", log_file.name, lineno, e)
                        continue

                    # 날짜 필터
                    ts = record.get("ts", "")
                    if not ts.startswith(target_date):
                        continue

                    record_type = record.get("type", "")

                    if record_type == "api_call":
                        counters["api_calls_per_day"] += 1
                        counters["tokens_per_day"] += int(record.get("tokens", 0))

                    elif record_type == "image":
                        counters["images_per_day"] += int(record.get("count", 1))

                    elif record_type == "youtube":
                        counters["youtube_per_day"] += int(record.get("count", 1))

                    else:
                        logger.debug("알 수 없는 record type: %s", record_type)

        except Exception as e:
            logger.warning("로그 파일 처리 실패 %s: %s", log_file.name, e)

    logger.info(
        "사용량 집계 완료 (날짜=%s): api=%d, tokens=%d, images=%d, youtube=%d",
        target_date,
        counters.get("api_calls_per_day", 0),
        counters.get("tokens_per_day", 0),
        counters.get("images_per_day", 0),
        counters.get("youtube_per_day", 0),
    )
    return dict(counters)


def _compute_quota_status(
    usage: Dict[str, int],
    quotas: Dict[str, int],
) -> Dict[str, Any]:
    """사용량 대비 할당량 상태 계산. 초과 항목 플래그 포함."""
    status: Dict[str, Any] = {}
    any_exceeded = False

    for key, limit in quotas.items():
        used = usage.get(key, 0)
        pct = round(used / limit * 100, 1) if limit > 0 else 0.0
        exceeded = used >= limit
        if exceeded:
            any_exceeded = True
        status[key] = {
            "used": used,
            "limit": limit,
            "pct": pct,
            "exceeded": exceeded,
        }

    status["any_exceeded"] = any_exceeded
    return status


def run_quota_check(target_date: Optional[str] = None) -> Dict[str, Any]:
    """
    사용 로그 기반 할당량 체크 실행.
    결과를 quota_state.json에 저장하고 반환.
    """
    if target_date is None:
        target_date = date.today().isoformat()

    logger.info("할당량 체크 시작: %s", target_date)

    usage = _parse_usage_logs(target_date)
    quotas = DEFAULT_QUOTAS.copy()

    # 기존 상태에서 사용자 정의 할당량 로드 (있으면)
    existing_state = _load_quota_state()
    custom_quotas = existing_state.get("custom_quotas", {})
    quotas.update({k: v for k, v in custom_quotas.items() if isinstance(v, int) and v > 0})

    quota_status = _compute_quota_status(usage, quotas)

    state = {
        "date": target_date,
        "quotas": quotas,
        "usage": usage,
        "status": quota_status,
        "custom_quotas": custom_quotas,
    }

    _save_quota_state(state)

    if quota_status.get("any_exceeded"):
        exceeded_keys = [
            k for k, v in quota_status.items()
            if isinstance(v, dict) and v.get("exceeded")
        ]
        logger.warning("할당량 초과 항목: %s", exceeded_keys)
    else:
        logger.info("모든 할당량 정상 범위")

    return state


def get_quota_summary() -> str:
    """현재 할당량 상태를 사람이 읽기 쉬운 문자열로 반환."""
    state = _load_quota_state()
    if not state:
        return "할당량 데이터 없음 (아직 체크가 실행되지 않음)"

    date_str = state.get("date", "unknown")
    status = state.get("status", {})
    lines = [f"[Quota Summary — {date_str}]"]

    for key, info in status.items():
        if not isinstance(info, dict):
            continue
        flag = "🔴 초과" if info.get("exceeded") else "🟢 정상"
        lines.append(
            f"  {key}: {info['used']}/{info['limit']} ({info['pct']}%) {flag}"
        )

    return "\n".join(lines)


if __name__ == "__main__":
    import time

    interval = int(os.getenv("NIGHTGUARD_INTERVAL", "600"))  # 기본 10분
    logger.info("Nightguard v2 시작 (interval=%ds)", interval)

    while True:
        try:
            run_quota_check()
        except Exception as e:
            logger.error("할당량 체크 중 오류: %s", e)
        time.sleep(interval)
