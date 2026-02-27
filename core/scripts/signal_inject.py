#!/usr/bin/env python3
"""
signal_inject.py — CLI 수동 신호 입력 도구

텔레그램 없이도 모든 유형의 신호를 LAYER OS에 투입할 수 있다.

사용법:
    python scripts/signal_inject.py --text "메모 내용"
    python scripts/signal_inject.py --url "https://example.com"
    python scripts/signal_inject.py --image "/path/to/photo.jpg"
    python scripts/signal_inject.py --pdf "/path/to/document.pdf"

모든 신호는 knowledge/signals/에 통합 스키마 JSON으로 저장되고
SA 큐에 자동 전달된다.
"""

import argparse
import json
import logging
import shutil
import sys
from datetime import datetime
from pathlib import Path
from typing import Optional

logger = logging.getLogger(__name__)

PROJECT_ROOT = Path(__file__).parent.parent
SIGNALS_DIR = PROJECT_ROOT / "knowledge" / "signals"
FILES_DIR = SIGNALS_DIR / "files"
QUEUE_DIR = PROJECT_ROOT / ".infra" / "queue" / "tasks" / "pending"


def generate_signal_id(signal_type: str) -> str:
    """통합 스키마 signal_id 생성"""
    now = datetime.now()
    return "%s_%s_%s" % (signal_type, now.strftime("%Y%m%d"), now.strftime("%H%M%S"))


def create_signal(
    signal_type: str,
    content: str,
    source_channel: str = "cli",
    metadata: Optional[dict] = None,
) -> dict:
    """통합 스키마 신호 객체 생성"""
    signal_id = generate_signal_id(signal_type)
    return {
        "signal_id": signal_id,
        "type": signal_type,
        "status": "captured",
        "content": content,
        "captured_at": datetime.now().isoformat(),
        "from_user": "97layer",
        "source_channel": source_channel,
        "metadata": metadata or {},
    }


def save_signal(signal: dict) -> Path:
    """신호를 knowledge/signals/에 저장"""
    SIGNALS_DIR.mkdir(parents=True, exist_ok=True)
    filepath = SIGNALS_DIR / ("%s.json" % signal["signal_id"])
    with open(filepath, "w", encoding="utf-8") as f:
        json.dump(signal, f, ensure_ascii=False, indent=2)
    return filepath


def enqueue_for_sa(signal: dict) -> Optional[Path]:
    """SA 큐에 분석 태스크 전달"""
    QUEUE_DIR.mkdir(parents=True, exist_ok=True)
    task = {
        "task_id": "cli_%s" % signal["signal_id"],
        "agent": "SA",
        "task_type": "analyze_signal",
        "priority": "normal",
        "payload": {
            "signal_id": signal["signal_id"],
            "signal_path": str(SIGNALS_DIR / ("%s.json" % signal["signal_id"])),
        },
        "created_at": datetime.now().isoformat(),
        "status": "pending",
    }
    task_path = QUEUE_DIR / ("%s.json" % task["task_id"])
    with open(task_path, "w", encoding="utf-8") as f:
        json.dump(task, f, ensure_ascii=False, indent=2)
    return task_path


def copy_binary(source_path: str, signal_id: str, extension: str) -> str:
    """바이너리 파일을 signals/files/로 복사"""
    FILES_DIR.mkdir(parents=True, exist_ok=True)
    dest = FILES_DIR / ("%s%s" % (signal_id, extension))
    shutil.copy2(source_path, dest)
    return str(dest)


def inject_text(text: str) -> None:
    """텍스트 신호 입력"""
    signal = create_signal("text_insight", text)
    path = save_signal(signal)
    enqueue_for_sa(signal)
    print("저장: %s" % path)
    print("SA 큐 전달: %s" % signal["signal_id"])


def inject_url(url: str) -> None:
    """URL 신호 입력"""
    signal = create_signal(
        "url_content",
        "URL 수집: %s" % url,
        metadata={"source_url": url, "title": ""},
    )
    path = save_signal(signal)
    enqueue_for_sa(signal)
    print("저장: %s" % path)
    print("SA 큐 전달: %s (URL 콘텐츠 추출은 SA가 처리)" % signal["signal_id"])


def inject_image(image_path: str) -> None:
    """이미지 신호 입력"""
    source = Path(image_path)
    if not source.exists():
        logger.error("파일 없음: %s", image_path)
        sys.exit(1)

    signal_id = generate_signal_id("image")
    dest_path = copy_binary(image_path, signal_id, source.suffix)

    signal = create_signal(
        "image",
        "이미지 수집: %s" % source.name,
        metadata={"image_path": dest_path, "title": source.stem},
    )
    signal["signal_id"] = signal_id
    path = save_signal(signal)
    enqueue_for_sa(signal)
    print("파일 복사: %s" % dest_path)
    print("저장: %s" % path)
    print("SA 큐 전달: %s" % signal_id)


def inject_pdf(pdf_path: str) -> None:
    """PDF 신호 입력"""
    source = Path(pdf_path)
    if not source.exists():
        logger.error("파일 없음: %s", pdf_path)
        sys.exit(1)

    signal_id = generate_signal_id("pdf_document")
    dest_path = copy_binary(pdf_path, signal_id, ".pdf")

    # PDF 텍스트 추출 시도
    content = "PDF 수집: %s" % source.name
    try:
        import pdfplumber
        with pdfplumber.open(source) as pdf:
            pages_text = []
            for page in pdf.pages[:10]:  # 최대 10페이지
                text = page.extract_text()
                if text:
                    pages_text.append(text)
            if pages_text:
                content = "\n".join(pages_text)[:3000]  # 3000자 제한
                print("PDF 텍스트 추출: %d페이지, %d자" % (len(pages_text), len(content)))
    except ImportError:
        logger.warning("pdfplumber 미설치 — 텍스트 추출 생략 (pip install pdfplumber)")
    except Exception as e:
        logger.error("PDF 추출 실패: %s", e)

    signal = create_signal(
        "pdf_document",
        content,
        metadata={"pdf_path": dest_path, "title": source.stem},
    )
    signal["signal_id"] = signal_id
    path = save_signal(signal)
    enqueue_for_sa(signal)
    print("파일 복사: %s" % dest_path)
    print("저장: %s" % path)
    print("SA 큐 전달: %s" % signal_id)


def main():
    parser = argparse.ArgumentParser(
        description="LAYER OS 수동 신호 입력 도구",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="모든 신호는 knowledge/signals/에 통합 스키마로 저장됩니다.",
    )
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("--text", "-t", help="텍스트 메모 입력")
    group.add_argument("--url", "-u", help="URL 입력")
    group.add_argument("--image", "-i", help="이미지 파일 경로")
    group.add_argument("--pdf", "-p", help="PDF 파일 경로")

    args = parser.parse_args()

    if args.text:
        inject_text(args.text)
    elif args.url:
        inject_url(args.url)
    elif args.image:
        inject_image(args.image)
    elif args.pdf:
        inject_pdf(args.pdf)


if __name__ == "__main__":
    main()
