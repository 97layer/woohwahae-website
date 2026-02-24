#!/usr/bin/env python3
"""
token_tracker.py — 세션별 토큰 사용량 추적

Stop hook에서 stdin JSON을 받아 usage 데이터를 로그에 누적.

Usage (session-stop.sh에서):
    echo "$INPUT_JSON" | python3 core/system/token_tracker.py
"""
import json
import sys
from datetime import datetime
from pathlib import Path


PROJECT_ROOT = Path(__file__).parent.parent.parent
LOG_PATH = PROJECT_ROOT / "knowledge" / "system" / "token_usage_log.jsonl"


def parse_usage_from_stdin() -> dict:
    """stdin JSON에서 usage 데이터 추출."""
    try:
        raw = sys.stdin.read()
        if not raw.strip():
            return {}
        data = json.loads(raw)
        return data
    except Exception:
        return {}


def build_entry(data: dict) -> dict:
    """로그 엔트리 구성."""
    now = datetime.now()

    usage = data.get("usage", {})

    entry = {
        "ts": now.isoformat(),
        "date": now.strftime("%Y-%m-%d"),
        "session_id": data.get("session_id", "unknown"),
    }

    # usage 필드가 있으면 추출
    if usage:
        entry["input_tokens"] = usage.get("input_tokens", 0)
        entry["output_tokens"] = usage.get("output_tokens", 0)
        entry["cache_read"] = usage.get("cache_read_input_tokens", 0)
        entry["cache_write"] = usage.get("cache_creation_input_tokens", 0)
        total = entry["input_tokens"] + entry["output_tokens"]
        entry["total_tokens"] = total
    else:
        # usage 데이터 없음 — 세션 기록만
        entry["note"] = "usage_unavailable"

    return entry


def append_log(entry: dict) -> None:
    LOG_PATH.parent.mkdir(parents=True, exist_ok=True)
    with open(LOG_PATH, "a", encoding="utf-8") as f:
        f.write(json.dumps(entry, ensure_ascii=False) + "\n")


def print_summary(entry: dict) -> None:
    if "note" in entry:
        print(f"token_tracker: 세션 기록 완료 (usage 데이터 없음)")
        return

    inp = entry.get("input_tokens", 0)
    out = entry.get("output_tokens", 0)
    cache_r = entry.get("cache_read", 0)
    total = entry.get("total_tokens", 0)

    print(f"토큰 사용: input={inp:,} output={out:,} cache_hit={cache_r:,} total={total:,}")


def get_daily_summary(date=None) -> dict:
    """일별 누적 합산."""
    if date is None:
        date = datetime.now().strftime("%Y-%m-%d")

    if not LOG_PATH.exists():
        return {}

    totals = {"input": 0, "output": 0, "cache_read": 0, "sessions": 0}
    with open(LOG_PATH, encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                entry = json.loads(line)
                if entry.get("date") == date and "note" not in entry:
                    totals["input"] += entry.get("input_tokens", 0)
                    totals["output"] += entry.get("output_tokens", 0)
                    totals["cache_read"] += entry.get("cache_read", 0)
                    totals["sessions"] += 1
            except Exception:
                continue
    return totals


def main() -> None:
    data = parse_usage_from_stdin()
    entry = build_entry(data)
    append_log(entry)
    print_summary(entry)

    # 오늘 누적 출력
    daily = get_daily_summary()
    if daily.get("sessions", 0) > 1:
        print(
            f"오늘 누적: {daily['sessions']}세션 "
            f"input={daily['input']:,} output={daily['output']:,} "
            f"cache_hit={daily['cache_read']:,}"
        )


if __name__ == "__main__":
    main()
