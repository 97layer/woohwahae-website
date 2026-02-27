#!/usr/bin/env python3
"""
LAYER OS Signal Processor — DEPRECATED

⚠️ 이 모듈은 폐기됨. 사용하지 마시오.
대체: pipeline_orchestrator.py + signal_router.py

이유:
- watchdog 기반 → 폴링 기반 (pipeline_orchestrator.run_forever)
- cortex_edge 의존 → 큐 기반 에이전트 체인으로 전환
- 직접 에이전트 실행 → 큐 태스크 생성 방식으로 전환

Author: LAYER OS Technical Director
Created: 2026-02-16
Deprecated: 2026-02-28
"""

import os
import sys
import json
import time
import logging
from pathlib import Path
from typing import Dict, List, Optional
from datetime import datetime
import asyncio
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler
from core.system.cortex_edge import get_cortex

# Add project root to path
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

# Load .env
try:
    from dotenv import load_dotenv
    load_dotenv(PROJECT_ROOT / '.env')
except ImportError:
    pass

logger = logging.getLogger(__name__)


class SignalHandler(FileSystemEventHandler):
    """
    신호 파일 생성 감지 및 처리
    """

    def __init__(self, processor):
        self.processor = processor
        super().__init__()

    def on_created(self, event):
        """파일 생성 이벤트"""
        if event.is_directory:
            return

        # JSON 파일만 처리
        if event.src_path.endswith('.json'):
            logger.info("🔔 New signal detected: %s", event.src_path)
            # 파일이 완전히 쓰여질 때까지 잠시 대기
            time.sleep(0.5)
            self.processor.process_signal(event.src_path)


class SignalProcessor:
    """
    신호 처리 및 Multi-Agent 실행
    """

    def __init__(self, telegram_bot=None):
        """
        Initialize Signal Processor

        Args:
            telegram_bot: Telegram bot instance for notifications
        """
        self.signals_dir = PROJECT_ROOT / 'knowledge' / 'signals'
        self.telegram_bot = telegram_bot

        # Multi-Agent 스크립트 경로
        self.agents = {
            'sa': PROJECT_ROOT / 'core' / 'agents' / 'strategy_analyst.py',
            'ad': PROJECT_ROOT / 'core' / 'agents' / 'art_director.py',
            'ce': PROJECT_ROOT / 'core' / 'agents' / 'chief_editor.py',
            'ralph': PROJECT_ROOT / 'core' / 'agents' / 'ralph.py'
        }

        # 처리 큐
        self.processing_queue = []
        self.is_processing = False
        self.cortex = get_cortex()

        logger.info("✅ Signal Processor initialized with Cortex Integration")

    def process_signal(self, signal_path: str):
        """
        신호 파일 처리

        Args:
            signal_path: 신호 JSON 파일 경로
        """
        try:
            # 신호 파일 읽기
            with open(signal_path, 'r', encoding='utf-8') as f:
                signal_data = json.load(f)

            signal_type = signal_data.get('type', 'unknown')
            status = signal_data.get('status', 'unknown')

            # 이미 처리된 신호는 스킵
            if status != 'captured':
                logger.info("⏭️  Signal already processed: %s", signal_path)
                return

            logger.info("📊 Processing signal: %s", signal_type)

            # 신호 타입별 처리
            if signal_type == 'youtube_video':
                self._process_youtube_signal(signal_path, signal_data)
            elif signal_type == 'image':
                self._process_image_signal(signal_path, signal_data)
            elif signal_type == 'text_insight':
                self._process_text_signal(signal_path, signal_data)
            else:
                logger.warning("⚠️  Unknown signal type: %s", signal_type)

        except Exception as e:
            logger.error("❌ Error processing signal %s: %s", signal_path, e)

    def _process_youtube_signal(self, signal_path: str, signal_data: Dict):
        """YouTube 신호 처리"""
        logger.info("🎥 Processing YouTube signal...")

        video_id = signal_data.get('video_id', 'unknown')
        transcript_length = signal_data.get('full_transcript_length', 0)

        # Multi-Agent 실행 (간단한 버전)
        logger.info("🤖 Starting Multi-Agent analysis...")

        # 상태 업데이트
        signal_data['status'] = 'processing'
        signal_data['processed_at'] = datetime.now().isoformat()

        with open(signal_path, 'w', encoding='utf-8') as f:
            json.dump(signal_data, f, ensure_ascii=False, indent=2)

        # TODO: 실제 Multi-Agent 실행
        # 현재는 시뮬레이션
        time.sleep(2)

        # 완료 상태로 변경
        signal_data['status'] = 'completed'
        signal_data['completed_at'] = datetime.now().isoformat()
        signal_data['agent_results'] = {
            'sa': 'Strategy analysis completed',
            'ad': 'Visual direction completed',
            'ce': 'Content editing completed'
        }

        with open(signal_path, 'w', encoding='utf-8') as f:
            json.dump(signal_data, f, ensure_ascii=False, indent=2)

        logger.info("✅ YouTube signal processed: %s", video_id)

    def _process_image_signal(self, signal_path: str, signal_data: Dict):
        """이미지 신호 처리"""
        logger.info("📷 Processing image signal...")

        # 상태 업데이트
        signal_data['status'] = 'processing'
        signal_data['processed_at'] = datetime.now().isoformat()

        with open(signal_path, 'w', encoding='utf-8') as f:
            json.dump(signal_data, f, ensure_ascii=False, indent=2)

        # TODO: 실제 Multi-Agent 실행
        time.sleep(1)

        # 완료
        signal_data['status'] = 'completed'
        signal_data['completed_at'] = datetime.now().isoformat()

        with open(signal_path, 'w', encoding='utf-8') as f:
            json.dump(signal_data, f, ensure_ascii=False, indent=2)

        logger.info("✅ Image signal processed")

    def _process_text_signal(self, signal_path: str, signal_data: Dict):
        """텍스트 신호 처리"""
        logger.info("💬 Processing text signal...")

        content = signal_data.get('content', '')

        # 간단한 처리: 저장 및 메모리 반영
        signal_data['status'] = 'stored'
        signal_data['stored_at'] = datetime.now().isoformat()

        with open(signal_path, 'w', encoding='utf-8') as f:
            json.dump(signal_data, f, ensure_ascii=False, indent=2)

        # Cortex 메모리에 반영 (비동기성 분석 결과로 취급)
        self.cortex._update_long_term_memory(
            f"신규 텍스트 신호 감지: {content[:50]}...",
            f"시스템에 텍스트 인텔리전스로 저장되었습니다. ID: {signal_data.get('signal_id')}"
        )

        logger.info("✅ Text signal stored and indexed by Cortex")

    def start_monitoring(self):
        """신호 디렉토리 모니터링 시작"""
        logger.info("👁️  Monitoring directory: %s", self.signals_dir)

        # 기존 미처리 신호 처리
        self._process_existing_signals()

        # 실시간 모니터링 시작
        event_handler = SignalHandler(self)
        observer = Observer()
        observer.schedule(event_handler, str(self.signals_dir), recursive=True)
        observer.start()

        logger.info("✅ Signal monitoring started")

        try:
            while True:
                time.sleep(10)
        except KeyboardInterrupt:
            observer.stop()
            logger.info("🛑 Signal monitoring stopped")

        observer.join()

    def _process_existing_signals(self):
        """기존에 처리되지 않은 신호들 처리"""
        logger.info("🔍 Checking for existing unprocessed signals...")

        json_files = list(self.signals_dir.glob('**/*.json'))

        for json_file in json_files:
            try:
                with open(json_file, 'r', encoding='utf-8') as f:
                    signal_data = json.load(f)

                if signal_data.get('status') == 'captured':
                    logger.info("📌 Found unprocessed signal: %s", json_file.name)
                    self.process_signal(str(json_file))

            except Exception as e:
                logger.error("Error checking %s: %s", json_file, e)


def main():
    """Main entry point"""
    # Configure logging
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )

    logger.info("🚀 Starting Signal Processor...")

    processor = SignalProcessor()
    processor.start_monitoring()


if __name__ == "__main__":
    main()
