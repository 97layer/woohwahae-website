#!/usr/bin/env python3
"""
Night Guard Daemon (정찰기)
GCP VM 전용: 맥북 부재 시 24/7 트렌드 감시 및 상태 유지

역할:
- 맥북 오프라인 감지 (10분 타임아웃)
- 주권 획득 시 트렌드 크롤링, 모니터링 실행
- 맥북 복귀 시 관찰 모드로 전환
"""

import time
import sys
from pathlib import Path
from datetime import datetime
import logging

# Path Setup
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
if str(PROJECT_ROOT) not in sys.path:
    sys.path.append(str(PROJECT_ROOT))
system_path = str(PROJECT_ROOT / "system")
if system_path not in sys.path:
    sys.path.append(system_path)

# Imports
try:
    from core.system.hybrid_sync import HybridSync
    from libs.ai_engine import AIEngine
    from libs.notifier import Notifier
    from libs.core_config import ENVIRONMENT, PROCESSING_MODE
except ImportError as e:
    print(f"[CRITICAL] Import failed: {e}")
    sys.exit(1)

# Logging Setup
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - [Night Guard] - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class NightGuard:
    """Night Guard: VM 24/7 정찰기"""

    def __init__(self):
        logger.info("🛰️ Night Guard 초기화 중...")
        self.handshake = HybridSync()
        self.node_type = self.handshake.get_node_type()

        # VM 환경 확인
        if ENVIRONMENT != "GCP_VM":
            logger.warning(f"⚠️ Night Guard는 GCP_VM 환경 전용입니다 (현재: {ENVIRONMENT})")
            logger.warning("   맥북에서는 실행하지 마세요.")

        logger.info(f"✅ Night Guard 준비 완료 (노드: {self.node_type}, 모드: {PROCESSING_MODE})")

        # AI 및 Notifier (주권 획득 시에만 초기화)
        self.ai = None
        self.notifier = None

    def _init_ai_services(self):
        """AI 서비스 초기화 (메모리 절약을 위해 필요할 때만)"""
        if self.ai is None:
            logger.info("AI Engine 및 Notifier 초기화 중...")
            try:
                self.ai = AIEngine()
                self.notifier = Notifier()
                logger.info("✅ AI 서비스 초기화 완료")
            except Exception as e:
                logger.error(f"❌ AI 서비스 초기화 실패: {e}")

    def run_surveillance(self):
        """트렌드 감시 작업 (예시)"""
        logger.info("🔍 트렌드 크롤링 시작...")

        # TODO: 실제 트렌드 크롤링 로직 구현
        # 예시:
        # - RSS 피드 수집
        # - 뉴스 API 호출
        # - SNS 트렌드 분석

        trends = [
            "헤어 트렌드: 2026년 봄 '실버 애시' 컬러 인기",
            "미용실 경영: 온라인 예약 시스템 필수 전환",
            "WOOHWAHAE 브랜드: Slow 철학 공명 증가"
        ]

        # Drive 저장 (임시 로직)
        trends_path = PROJECT_ROOT / ".tmp" / "nightguard" / "trends"
        trends_path.mkdir(parents=True, exist_ok=True)

        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        output_file = trends_path / f"trends_{timestamp}.md"

        with open(output_file, "w", encoding="utf-8") as f:
            f.write(f"# Night Guard 트렌드 리포트\n\n")
            f.write(f"**수집 시각**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")
            for trend in trends:
                f.write(f"- {trend}\n")

        logger.info(f"✅ 트렌드 크롤링 완료: {len(trends)}개 항목")

        # Telegram 알림 (선택)
        if self.notifier:
            try:
                self.notifier.send_message_to_admin(
                    f"🛰️ Night Guard 보고\n\n"
                    f"트렌드 {len(trends)}개 감지\n"
                    f"시각: {datetime.now().strftime('%H:%M')}"
                )
            except Exception as e:
                logger.error(f"❌ Telegram 알림 실패: {e}")

        return True

    def run(self):
        """Main Loop: 5분마다 주권 확인 및 작업 실행"""
        logger.info("🛰️ Night Guard 가동 시작 (24/7 모드)")
        logger.info("   맥북 오프라인 시 자동 승격됩니다...")

        cycle_count = 0

        while True:
            try:
                cycle_count += 1
                now = datetime.now().strftime("%Y-%m-%d %H:%M:%S")

                # 주권 확인 (10분 타임아웃)
                has_ownership = self.handshake.claim_ownership(
                    node=self.node_type,
                    timeout_minutes=10
                )

                if has_ownership:
                    logger.info(f"[{now}] ✓ Night Guard 활성화 (Cycle #{cycle_count})")

                    # AI 서비스 초기화 (처음 한 번만)
                    self._init_ai_services()

                    # 트렌드 감시 실행
                    self.run_surveillance()

                    logger.info("   다음 주기까지 대기 (5분)...")
                else:
                    logger.info(f"[{now}] ○ 관찰 모드 (맥북 활성) (Cycle #{cycle_count})")

                # 5분 대기
                time.sleep(300)

            except KeyboardInterrupt:
                logger.info("\n🛑 Night Guard 중지 요청 감지")
                break
            except Exception as e:
                logger.error(f"❌ Night Guard 오류: {e}")
                logger.info("   10초 후 재시도...")
                time.sleep(10)

        logger.info("✅ Night Guard 종료")

if __name__ == "__main__":
    guard = NightGuard()
    guard.run()
