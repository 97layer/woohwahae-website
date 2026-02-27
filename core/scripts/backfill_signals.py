#!/usr/bin/env python3
"""
Signal Backfill — 기존 signal JSON에 누락 필드 채우기

스키마 v2.0 기준으로:
- signal_id 없으면 파일명에서 추출
- content 없으면 transcript/analysis.description에서 추출
- from_user 없으면 "97layer" 기본값
- source_channel 없으면 "manual" 기본값
- 최상위 analysis/video_id/source → metadata 안으로 정리
- analyzed_at 없고 status=analyzed면 파일 mtime 사용

dry-run 기본. --apply로 실제 쓰기.
"""

import json
import logging
import sys
from datetime import datetime
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
SIGNALS_DIR = PROJECT_ROOT / "knowledge" / "signals"

logging.basicConfig(level=logging.INFO, format="%(message)s")
logger = logging.getLogger(__name__)


def _extract_signal_id(filepath: Path, data: dict) -> str:
    """파일명에서 signal_id 추출"""
    stem = filepath.stem  # e.g. "youtube_7HBhL7lltpU_20260217_141746"
    sig_type = data.get("type", "")

    if sig_type == "youtube_video":
        # youtube_{videoId}_{YYYYMMDD}_{HHMMSS} → signal_id 형태로
        parts = stem.split("_")
        if len(parts) >= 4:
            date_part = parts[-2]
            time_part = parts[-1]
            return "youtube_video_%s_%s" % (date_part, time_part)

    if sig_type == "image":
        parts = stem.split("_")
        if len(parts) >= 3:
            date_part = parts[-2]
            time_part = parts[-1]
            return "image_%s_%s" % (date_part, time_part)

    # text, url_content 등은 보통 이미 signal_id가 있음
    return stem


def _extract_content(data: dict) -> str:
    """content 필드 추출 — transcript, analysis.description 등에서"""
    if data.get("content"):
        return data["content"]

    # youtube: transcript에서
    transcript = data.get("transcript", "")
    if transcript:
        return transcript[:500]

    # image: analysis.description에서
    analysis = data.get("analysis", {})
    if isinstance(analysis, dict):
        desc = analysis.get("description", "")
        if desc:
            return desc[:500]
        caption = analysis.get("caption", "")
        if caption:
            return caption

    return ""


def backfill_signal(filepath: Path, apply: bool = False) -> list:
    """단일 signal JSON 백필. 변경사항 목록 반환."""
    changes = []

    try:
        raw = filepath.read_text(encoding="utf-8")
        data = json.loads(raw)
    except Exception as exc:
        return [("ERROR", str(exc))]

    # signal_id
    if not data.get("signal_id"):
        sid = _extract_signal_id(filepath, data)
        data["signal_id"] = sid
        changes.append(("ADD", "signal_id", sid))

    # content
    if not data.get("content"):
        content = _extract_content(data)
        if content:
            data["content"] = content
            changes.append(("ADD", "content", content[:60] + "..."))

    # from_user
    if not data.get("from_user"):
        data["from_user"] = "97layer"
        changes.append(("ADD", "from_user", "97layer"))

    # source_channel
    if not data.get("source_channel"):
        data["source_channel"] = "manual"
        changes.append(("ADD", "source_channel", "manual"))

    # analyzed_at (status=analyzed인데 없는 경우)
    if data.get("status") == "analyzed" and not data.get("analyzed_at"):
        mtime = datetime.fromtimestamp(filepath.stat().st_mtime).isoformat()
        data["analyzed_at"] = mtime
        changes.append(("ADD", "analyzed_at", mtime))

    # metadata 정리: 최상위 video_id, source, analysis → metadata로
    metadata = data.get("metadata", {})
    if not isinstance(metadata, dict):
        metadata = {}

    if "video_id" in data and "video_id" not in metadata:
        metadata["video_id"] = data.pop("video_id")
        changes.append(("MOVE", "video_id → metadata.video_id"))

    if "source" in data and data["type"] != "image":
        url = data.pop("source")
        if url and "source_url" not in metadata:
            metadata["source_url"] = url
            changes.append(("MOVE", "source → metadata.source_url"))

    if "transcript" in data:
        preview = data.pop("transcript")[:2000]
        if "transcript_preview" not in metadata:
            metadata["transcript_preview"] = preview
            changes.append(("MOVE", "transcript → metadata.transcript_preview"))

    if "full_transcript_length" in data:
        length = data.pop("full_transcript_length")
        metadata["transcript_length"] = length
        changes.append(("MOVE", "full_transcript_length → metadata.transcript_length"))

    # image: source → metadata.image_path
    if data.get("type") == "image" and "source" in data:
        src = data.pop("source")
        if "image_path" not in metadata:
            metadata["image_path"] = src
            changes.append(("MOVE", "source → metadata.image_path"))

    if "saved_image" in data:
        saved = data.pop("saved_image")
        if "image_path" not in metadata:
            metadata["image_path"] = saved
        changes.append(("MOVE", "saved_image → metadata.image_path"))

    # 최상위 analysis → metadata.analysis
    if "analysis" in data and "analysis" not in metadata:
        metadata["analysis"] = data.pop("analysis")
        changes.append(("MOVE", "analysis → metadata.analysis"))
    elif "analysis" in data and "analysis" in metadata:
        data.pop("analysis")
        changes.append(("DROP", "duplicate top-level analysis"))

    if metadata:
        data["metadata"] = metadata

    # 쓰기
    if changes and apply:
        filepath.write_text(
            json.dumps(data, ensure_ascii=False, indent=2),
            encoding="utf-8"
        )

    return changes


def main():
    apply = "--apply" in sys.argv

    if not SIGNALS_DIR.exists():
        logger.error("signals/ 디렉토리 없음")
        return

    total_files = 0
    total_changes = 0

    for sf in sorted(SIGNALS_DIR.glob("**/*.json")):
        changes = backfill_signal(sf, apply=apply)
        if changes:
            total_files += 1
            total_changes += len(changes)
            rel = sf.relative_to(PROJECT_ROOT)
            logger.info("\n📄 %s", rel)
            for c in changes:
                if len(c) == 3:
                    logger.info("  %s %s = %s", c[0], c[1], c[2])
                else:
                    logger.info("  %s %s", c[0], c[1])

    mode = "APPLIED" if apply else "DRY-RUN"
    logger.info("\n--- %s ---", mode)
    logger.info("파일: %d / 변경: %d", total_files, total_changes)
    if not apply and total_changes > 0:
        logger.info("실제 적용: python3 core/scripts/backfill_signals.py --apply")


if __name__ == "__main__":
    main()
