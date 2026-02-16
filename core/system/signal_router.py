#!/usr/bin/env python3
"""
Signal Router — 97layerOS

역할:
  - knowledge/signals/ 디렉토리를 감시
  - 새 신호 파일(*.json) 감지 → QueueManager.create_task()로 큐에 투입
  - 신호 타입에 따라 에이전트 분기 (SA: 분석, AD: 비주얼, urgent: 즉시 처리)
  - 이미 처리한 신호는 재처리하지 않음 (processed 추적)

실행 방법:
  python core/system/signal_router.py [--once] [--watch]

THE CYCLE에서의 위치:
  텔레그램 입력 → knowledge/signals/ 저장 → [Signal Router] → Queue → Agent 처리
"""

import json
import time
import logging
from pathlib import Path
from datetime import datetime
from typing import Optional

from core.system.queue_manager import QueueManager

# 경로 설정
PROJECT_ROOT = Path(__file__).parent.parent.parent
SIGNALS_DIR = PROJECT_ROOT / "knowledge" / "signals"
PROCESSED_INDEX = PROJECT_ROOT / "knowledge" / "system" / "signal_router_processed.json"

# 설정
WATCH_INTERVAL = 10   # 초 (새 신호 폴링 간격)

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [SignalRouter] %(levelname)s: %(message)s",
    handlers=[logging.StreamHandler()],
)
logger = logging.getLogger(__name__)


# ─── 신호 분류 ─────────────────────────────────────────────────────────────

def classify_signal(signal: dict) -> tuple[str, str]:
    """
    신호를 분석하여 (agent_type, task_type) 튜플 반환.

    신호 타입별 라우팅:
      youtube_video  → SA / analyze_youtube
      text           → SA / analyze_text
      image          → AD / analyze_image
      urgent         → SA / urgent_analyze
      default        → SA / analyze
    """
    sig_type = signal.get("type", "unknown")
    status = signal.get("status", "captured")

    # 긴급 마킹된 신호
    if signal.get("urgent") or status == "urgent":
        return ("SA", "urgent_analyze")

    routing = {
        "youtube_video": ("SA", "analyze_youtube"),
        "text":          ("SA", "analyze_text"),
        "image":         ("AD", "analyze_image"),
        "link":          ("SA", "analyze_text"),
        "memo":          ("SA", "analyze_text"),
    }

    return routing.get(sig_type, ("SA", "analyze"))


# ─── 처리 기록 ─────────────────────────────────────────────────────────────

def load_processed() -> set:
    """이미 큐에 투입한 신호 파일명 집합 반환."""
    if PROCESSED_INDEX.exists():
        try:
            data = json.loads(PROCESSED_INDEX.read_text(encoding="utf-8"))
            return set(data.get("processed", []))
        except (json.JSONDecodeError, IOError):
            pass
    return set()


def save_processed(processed: set) -> None:
    """처리 기록 저장."""
    PROCESSED_INDEX.parent.mkdir(parents=True, exist_ok=True)
    data = {
        "processed": sorted(processed),
        "last_updated": datetime.now().isoformat(),
        "count": len(processed),
    }
    PROCESSED_INDEX.write_text(
        json.dumps(data, indent=2, ensure_ascii=False), encoding="utf-8"
    )


# ─── 핵심 라우팅 로직 ──────────────────────────────────────────────────────

def route_signal(signal_path: Path, queue: QueueManager) -> Optional[str]:
    """
    단일 신호 파일을 읽어 큐에 투입.

    반환값: task_id (성공) | None (실패 또는 이미 처리됨)
    """
    try:
        signal = json.loads(signal_path.read_text(encoding="utf-8"))
    except (json.JSONDecodeError, IOError) as e:
        logger.warning("신호 파일 읽기 실패 (%s): %s", signal_path.name, e)
        return None

    agent_type, task_type = classify_signal(signal)

    payload = {
        "signal_path": str(signal_path),
        "signal_type": signal.get("type", "unknown"),
        "captured_at": signal.get("captured_at", ""),
        "source": signal.get("source", ""),
        "data": signal,
    }

    try:
        task_id = queue.create_task(
            agent_type=agent_type,
            task_type=task_type,
            payload=payload,
        )
        logger.info(
            "큐 투입 → [%s/%s] %s (task_id=%s)",
            agent_type, task_type, signal_path.name, task_id,
        )
        return task_id
    except Exception as e:
        logger.error("큐 투입 실패 (%s): %s", signal_path.name, e)
        return None


def process_pending_signals(queue: Optional[QueueManager] = None) -> int:
    """
    signals/ 디렉토리의 미처리 신호 전체를 큐에 투입.

    반환값: 새로 투입된 신호 수
    """
    if not SIGNALS_DIR.exists():
        logger.warning("signals/ 디렉토리 없음: %s", SIGNALS_DIR)
        return 0

    if queue is None:
        queue = QueueManager()

    processed = load_processed()
    new_count = 0

    signal_files = sorted(SIGNALS_DIR.glob("*.json"))
    for path in signal_files:
        if path.name in processed:
            continue  # 이미 처리됨

        task_id = route_signal(path, queue)
        if task_id:
            processed.add(path.name)
            new_count += 1

    if new_count > 0:
        save_processed(processed)
        logger.info("총 %d개 신호 큐 투입 완료.", new_count)
    else:
        logger.debug("새 신호 없음.")

    return new_count


# ─── 실행 모드 ─────────────────────────────────────────────────────────────

def run_once() -> None:
    """미처리 신호 1회 스캔 후 종료."""
    queue = QueueManager()
    count = process_pending_signals(queue)
    print(f"처리된 신호: {count}개")


def run_watch() -> None:
    """
    signals/ 디렉토리를 주기적으로 감시하며 새 신호 자동 투입.
    Ctrl+C로 종료.
    """
    logger.info(
        "Signal Router 감시 시작 (interval=%ds, signals_dir=%s)",
        WATCH_INTERVAL,
        SIGNALS_DIR,
    )
    queue = QueueManager()

    try:
        while True:
            process_pending_signals(queue)
            time.sleep(WATCH_INTERVAL)
    except KeyboardInterrupt:
        logger.info("Signal Router 종료.")


# ─── CLI ────────────────────────────────────────────────────────────────────

if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="97layerOS Signal Router")
    parser.add_argument(
        "--once",
        action="store_true",
        help="미처리 신호 1회 스캔 후 종료",
    )
    parser.add_argument(
        "--watch",
        action="store_true",
        help=f"signals/ 디렉토리 감시 (폴링 {WATCH_INTERVAL}초 간격)",
    )
    parser.add_argument(
        "--reset-processed",
        action="store_true",
        help="처리 기록 초기화 (모든 신호 재처리)",
    )
    args = parser.parse_args()

    if args.reset_processed:
        if PROCESSED_INDEX.exists():
            PROCESSED_INDEX.unlink()
        logger.info("처리 기록 초기화 완료.")

    if args.once:
        run_once()
    else:
        # --watch 또는 기본값
        run_watch()
