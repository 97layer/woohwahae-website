#!/usr/bin/env python3
"""
97layerOS Telegram Executive Secretary
Phase 2: 24/7 자동화된 비서 - 신호 포착, 명령어 처리, 멀티에이전트 협업

핵심 기능:
- 명령어: /status, /report, /analyze, /signal
- 자동 신호 포착: 텍스트 + 이미지 + 링크
- parallel_orchestrator.py 호출로 멀티에이전트 처리
- asset_manager.py로 결과 자산 등록
- handoff.py 세션 연속성 통합

Author: 97layerOS Technical Director
Created: 2026-02-16
"""

import os
import sys
import asyncio
import logging
from pathlib import Path
from datetime import datetime
from typing import Optional, Dict, Any

# Telegram bot imports
from telegram import Update
from telegram.ext import (
    Application,
    CommandHandler,
    MessageHandler,
    filters,
    ContextTypes
)

# Project setup
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from execution.system.handoff import HandoffEngine
from execution.system.parallel_orchestrator import ParallelOrchestrator
from execution.system.daily_routine import DailyRoutine
from execution.system.gdrive_sync import GDriveSync
from execution.system.notebooklm_bridge import NotebookLMBridge, anti_gravity_youtube
from system.libs.agents.asset_manager import AssetManager

# Logging setup
log_dir = PROJECT_ROOT / 'logs'
log_dir.mkdir(parents=True, exist_ok=True)

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(name)s] %(levelname)s: %(message)s',
    handlers=[
        logging.FileHandler(log_dir / 'telegram_secretary.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger("TelegramSecretary")


class TelegramSecretary:
    """
    97layerOS Executive Secretary

    Responsibilities:
    1. 24/7 신호 포착 및 분류
    2. 명령어 처리 및 보고
    3. 멀티에이전트 협업 조율
    4. 자산 생명주기 관리
    """

    def __init__(self, bot_token: str):
        """Initialize Telegram Secretary"""
        self.bot_token = bot_token
        self.handoff = HandoffEngine()
        self.orchestrator = ParallelOrchestrator()
        self.asset_manager = AssetManager()
        self.daily_routine = DailyRoutine()

        # Google Drive sync (optional - only if credentials exist)
        try:
            self.gdrive = GDriveSync()
            logger.info("✅ Google Drive sync enabled")
        except Exception as e:
            self.gdrive = None
            logger.warning(f"⚠️  Google Drive sync disabled: {e}")

        # Session setup
        logger.info("🤖 Telegram Secretary 초기화 중...")
        self.handoff.onboard()  # 세션 연속성 복구

        # Acquire work lock
        if not self.handoff.acquire_work_lock(
            agent_id="TelegramSecretary",
            task="Telegram Bot Operation",
            resources=["knowledge/signals/", "knowledge/system/"],
            timeout_minutes=60  # 1시간 자동 갱신
        ):
            logger.warning("⚠️  Work lock 획득 실패 - 다른 작업 진행 중")

        logger.info("✅ Telegram Secretary 준비 완료")

    async def start_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """
        /start - Secretary 시작 인사
        """
        user = update.effective_user
        logger.info(f"📱 /start from {user.first_name} ({user.id})")

        await update.message.reply_text(
            "🤖 97layerOS Executive Secretary\n\n"
            "슬로우 라이프 아카이브의 비서입니다.\n\n"
            "**명령어**:\n"
            "/status - 시스템 현재 상태\n"
            "/report - 오늘의 작업 보고\n"
            "/analyze - 신호 멀티에이전트 분석\n"
            "/signal <텍스트> - 새 신호 입력\n"
            "/morning - 아침 브리핑 (09:00 권장)\n"
            "/evening - 저녁 리포트 (21:00 권장)\n\n"
            "**비서 기능** (Phase 2.4):\n"
            "/search <검색어> - 과거 지식 베이스 검색\n"
            "/memo <메모> - 빠른 메모 저장\n"
            "/sync - 클라우드 동기화 (수동)\n\n"
            "**자동 포착**:\n"
            "메시지, 이미지, 링크를 보내면 자동으로 분류하고 처리합니다."
        )

    async def status_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """
        /status - 시스템 현재 상태 보고
        """
        user = update.effective_user
        logger.info(f"📊 /status from {user.first_name} ({user.id})")

        # Work lock 상태
        lock_status = self.handoff.check_work_lock()

        # Asset 통계
        registry = self.asset_manager._load_registry()
        stats = registry.get('stats', {})

        # 상태 메시지 구성
        status_msg = f"📊 **97layerOS 시스템 상태**\n\n"
        status_msg += f"⏰ {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n"

        # 작업 잠금
        if lock_status['locked']:
            status_msg += f"🔒 **작업 중**: {lock_status['agent']}\n"
            status_msg += f"   └─ {lock_status['task']}\n\n"
        else:
            status_msg += f"🔓 **대기 중** (작업 가능)\n\n"

        # 자산 통계
        total = stats.get('total', 0)
        by_status = stats.get('by_status', {})

        status_msg += f"📦 **자산 현황** (총 {total}개)\n"
        for status, count in by_status.items():
            status_msg += f"   • {status}: {count}개\n"

        await update.message.reply_text(status_msg)

    async def report_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """
        /report - 오늘의 작업 보고서
        """
        user = update.effective_user
        logger.info(f"📋 /report from {user.first_name} ({user.id})")

        # Asset 보고서 생성
        report_path = PROJECT_ROOT / 'knowledge' / 'system' / 'daily_report.md'
        self.asset_manager.generate_report(str(report_path))

        await update.message.reply_text(
            f"📋 **오늘의 작업 보고서**\n\n"
            f"보고서가 생성되었습니다:\n"
            f"`{report_path}`\n\n"
            f"자세한 내용은 파일을 확인하세요."
        )

    async def analyze_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """
        /analyze - 마지막 신호 멀티에이전트 분석
        """
        user = update.effective_user
        logger.info(f"🔍 /analyze from {user.first_name} ({user.id})")

        await update.message.reply_text(
            "🔍 멀티에이전트 분석을 시작합니다...\n"
            "SA (전략) + AD (비주얼) 병렬 → CE (정제) → CD (승인)"
        )

        # 마지막 신호 파일 찾기
        signals_dir = PROJECT_ROOT / 'knowledge' / 'signals'
        if not signals_dir.exists():
            await update.message.reply_text(
                "⚠️  신호 폴더가 없습니다.\n"
                "/signal <텍스트> 또는 메시지를 보내서 신호를 입력하세요."
            )
            return

        # 가장 최근 신호 파일 가져오기
        signal_files = sorted(signals_dir.glob('*.md'), key=lambda p: p.stat().st_mtime, reverse=True)
        if not signal_files:
            await update.message.reply_text(
                "⚠️  분석할 신호가 없습니다.\n"
                "/signal <텍스트> 또는 메시지를 보내서 신호를 입력하세요."
            )
            return

        latest_signal = signal_files[0]
        logger.info(f"📄 분석 대상: {latest_signal.name}")

        # 이미지 신호인지 확인
        image_path = None
        if 'image_' in latest_signal.name:
            # 메타데이터에서 이미지 경로 추출
            with open(latest_signal, 'r', encoding='utf-8') as f:
                for line in f:
                    if line.startswith('**File**:'):
                        image_path = Path(line.split(':', 1)[1].strip())
                        break

        try:
            # 멀티에이전트 병렬 처리 실행
            await update.message.reply_text("⏳ 처리 중... (약 30초 소요)")

            result = await self.orchestrator.process_signal(
                signal_path=str(latest_signal),
                image_path=str(image_path) if image_path else None
            )

            # 결과 전송
            if result['status'] == 'success':
                final_asset = result['final_asset']
                quality = result['quality_score']

                response = f"✅ **분석 완료**\n\n"
                response += f"📄 신호: `{latest_signal.name}`\n"
                response += f"⭐ 품질: {quality}/100\n\n"
                response += f"**최종 자산**:\n`{final_asset}`\n\n"
                response += f"전체 결과는 파일을 확인하세요."

                await update.message.reply_text(response)

            else:
                await update.message.reply_text(
                    f"❌ 분석 실패\n\n"
                    f"오류: {result.get('error', 'Unknown')}\n"
                    f"로그를 확인하세요."
                )

        except Exception as e:
            logger.error(f"❌ 분석 오류: {e}")
            await update.message.reply_text(
                f"❌ 분석 중 오류 발생:\n{str(e)}\n\n"
                f"로그를 확인하세요."
            )

    async def signal_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """
        /signal <텍스트> - 새 신호 수동 입력
        """
        user = update.effective_user
        text = ' '.join(context.args) if context.args else None

        if not text:
            await update.message.reply_text(
                "사용법: /signal <텍스트>\n"
                "예: /signal 슬로우 라이프 콘텐츠 아이디어"
            )
            return

        logger.info(f"📥 /signal from {user.first_name}: {text[:50]}...")

        # 신호 파일 저장
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        signal_path = PROJECT_ROOT / 'knowledge' / 'signals' / f'signal_{timestamp}.md'
        signal_path.parent.mkdir(parents=True, exist_ok=True)

        with open(signal_path, 'w', encoding='utf-8') as f:
            f.write(f"# Signal {timestamp}\n\n")
            f.write(f"**From**: {user.first_name} (@{user.username or 'unknown'})\n")
            f.write(f"**Time**: {datetime.now().isoformat()}\n\n")
            f.write(f"## Content\n\n{text}\n")

        await update.message.reply_text(
            f"✅ 신호가 저장되었습니다.\n"
            f"ID: `signal_{timestamp}`\n\n"
            f"멀티에이전트 분석을 시작하려면 /analyze를 입력하세요."
        )

    async def handle_message(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """
        일반 메시지 자동 포착 및 분류

        특수 패턴 자동 감지:
        - YouTube URL → Anti-Gravity 분석 자동 실행
        - 일반 텍스트 → Signal로 저장
        """
        user = update.effective_user
        text = update.message.text

        logger.info(f"💬 Message from {user.first_name}: {text[:50]}...")

        # YouTube URL 자동 감지
        youtube_patterns = [
            'youtube.com/watch?v=',
            'youtu.be/',
            'm.youtube.com/watch?v='
        ]

        if any(pattern in text.lower() for pattern in youtube_patterns):
            logger.info(f"🛸 YouTube URL 자동 감지: {text}")
            await update.message.reply_text(
                "🛸 YouTube URL 감지!\n"
                "Anti-Gravity 프로토콜을 자동 실행합니다..."
            )

            # YouTube command로 위임
            context.args = [text.strip()]
            await self.youtube_command(update, context)
            return

        # 일반 신호로 자동 저장
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        signal_path = PROJECT_ROOT / 'knowledge' / 'signals' / f'auto_{timestamp}.md'
        signal_path.parent.mkdir(parents=True, exist_ok=True)

        with open(signal_path, 'w', encoding='utf-8') as f:
            f.write(f"# Auto Signal {timestamp}\n\n")
            f.write(f"**From**: {user.first_name} (@{user.username or 'unknown'})\n")
            f.write(f"**Time**: {datetime.now().isoformat()}\n")
            f.write(f"**Type**: text\n\n")
            f.write(f"## Content\n\n{text}\n")

        await update.message.reply_text(
            f"📥 신호 포착 완료: `auto_{timestamp}`\n"
            f"분석하려면 /analyze를 입력하세요."
        )

    async def morning_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """
        /morning - 아침 브리핑 (09:00 권장)
        """
        user = update.effective_user
        logger.info(f"🌅 /morning from {user.first_name} ({user.id})")

        await update.message.reply_text(
            "🌅 아침 브리핑을 생성하는 중입니다...\n"
            "잠시만 기다려주세요."
        )

        try:
            briefing = self.daily_routine.morning_briefing()

            # 요약 메시지 구성
            summary = briefing['summary']
            response = f"🌅 **좋은 아침입니다**\n\n"
            response += f"📊 현황:\n"
            response += f"   • 대기 중: {summary['pending_count']}개\n"
            response += f"   • 재작업 필요: {summary['refined_count']}개\n"
            response += f"   • 어제 완료: {summary['completed_yesterday']}개\n\n"

            if summary['pending_count'] > 0:
                response += f"🎯 오늘은 Pending 자산 처리에 집중해보세요.\n\n"

            response += f"💡 **슬로우 라이프 리마인더**\n"
            response += f"속도보다 방향, 효율보다 본질을 기억하세요.\n"
            response += f"오늘도 나다운 속도로 나아갑니다.\n\n"

            # 보고서 경로
            date_str = datetime.now().strftime('%Y%m%d')
            report_path = f"knowledge/reports/daily/morning_{date_str}.json"
            response += f"📄 상세 브리핑: `{report_path}`"

            await update.message.reply_text(response)

        except Exception as e:
            logger.error(f"❌ 아침 브리핑 오류: {e}")
            await update.message.reply_text(
                f"❌ 브리핑 생성 중 오류 발생:\n{str(e)}"
            )

    async def evening_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """
        /evening - 저녁 리포트 (21:00 권장)
        """
        user = update.effective_user
        logger.info(f"🌙 /evening from {user.first_name} ({user.id})")

        await update.message.reply_text(
            "🌙 저녁 리포트를 생성하는 중입니다...\n"
            "잠시만 기다려주세요."
        )

        try:
            report = self.daily_routine.evening_report()

            # 요약 메시지 구성
            summary = report['summary']
            ralph_stats = summary.get('ralph_stats', {})

            response = f"🌙 **하루를 마무리합니다**\n\n"
            response += f"📊 오늘의 성과:\n"
            response += f"   • 완료: {summary['approved_today']}개\n"
            response += f"   • 아카이브: {summary['archived_today']}개\n\n"

            if ralph_stats:
                response += f"🔍 품질 관리:\n"
                response += f"   • 총 검증: {ralph_stats.get('total', 0)}회\n"
                response += f"   • 통과율: {ralph_stats.get('pass_rate', 0)}%\n"
                response += f"   • 평균 점수: {ralph_stats.get('avg_score', 0)}/100\n\n"

            response += f"💭 **하루 마무리**\n"
            response += f"완벽하지 않아도 괜찮습니다.\n"
            response += f"과정의 흔적을 남긴 것만으로도 충분합니다.\n\n"

            # 보고서 경로
            date_str = datetime.now().strftime('%Y%m%d')
            report_path = f"knowledge/reports/daily/evening_{date_str}.json"
            response += f"📄 상세 리포트: `{report_path}`"

            await update.message.reply_text(response)

        except Exception as e:
            logger.error(f"❌ 저녁 리포트 오류: {e}")
            await update.message.reply_text(
                f"❌ 리포트 생성 중 오류 발생:\n{str(e)}"
            )

    async def search_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """
        /search <검색어> - 과거 지식 베이스 검색 (Google Drive)
        """
        user = update.effective_user
        query = ' '.join(context.args) if context.args else None

        if not query:
            await update.message.reply_text(
                "사용법: /search <검색어>\n"
                "예: /search 슬로우 라이프 전략"
            )
            return

        logger.info(f"🔍 /search from {user.first_name}: {query}")

        if not self.gdrive:
            await update.message.reply_text(
                "⚠️  Google Drive 연동이 비활성화되어 있습니다.\n"
                "credentials/gdrive_auth.json 및 .env 설정을 확인하세요."
            )
            return

        await update.message.reply_text(
            f"🔍 '{query}' 검색 중...\n"
            "Google Drive 지식 베이스를 검색합니다."
        )

        try:
            # Search in Google Drive
            results = self.gdrive.search_files(f"name contains '{query}'")

            if not results:
                await update.message.reply_text(
                    f"🤷 '{query}'에 대한 결과를 찾지 못했습니다.\n\n"
                    f"💡 Tip: 다른 키워드를 시도하거나 NotebookLM에 직접 질문해보세요."
                )
                return

            # Format results
            response = f"🔍 **검색 결과**: '{query}'\n\n"
            response += f"총 {len(results)}개 파일 발견:\n\n"

            for idx, file in enumerate(results[:10], 1):  # 최대 10개
                modified = file.get('modifiedTime', 'Unknown')[:10]
                response += f"{idx}. {file['name']}\n"
                response += f"   📅 {modified} | ID: {file['id'][:8]}...\n\n"

            if len(results) > 10:
                response += f"... 외 {len(results) - 10}개 더 있습니다.\n\n"

            response += "💡 특정 파일 내용이 필요하면 알려주세요."

            await update.message.reply_text(response)

        except Exception as e:
            logger.error(f"❌ 검색 오류: {e}")
            await update.message.reply_text(
                f"❌ 검색 중 오류 발생:\n{str(e)}"
            )

    async def memo_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """
        /memo <메모> - 빠른 메모 저장 및 Drive 동기화
        """
        user = update.effective_user
        memo_text = ' '.join(context.args) if context.args else None

        if not memo_text:
            await update.message.reply_text(
                "사용법: /memo <메모 내용>\n"
                "예: /memo 내일 WOOHWAHAE 미팅 준비"
            )
            return

        logger.info(f"📝 /memo from {user.first_name}: {memo_text[:50]}...")

        # 메모 파일 저장
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        memo_dir = PROJECT_ROOT / 'knowledge' / 'memos'
        memo_dir.mkdir(parents=True, exist_ok=True)

        memo_path = memo_dir / f'memo_{timestamp}.md'
        with open(memo_path, 'w', encoding='utf-8') as f:
            f.write(f"# Memo {timestamp}\n\n")
            f.write(f"**From**: {user.first_name} (@{user.username or 'unknown'})\n")
            f.write(f"**Time**: {datetime.now().isoformat()}\n")
            f.write(f"**Via**: Telegram\n\n")
            f.write(f"## Content\n\n{memo_text}\n")

        response = f"✅ 메모가 저장되었습니다.\n"
        response += f"ID: `memo_{timestamp}`\n\n"

        # Google Drive 동기화 (선택적)
        if self.gdrive:
            try:
                file_id = self.gdrive.upload_file(memo_path, drive_folder="memos")
                if file_id:
                    response += f"☁️  Google Drive 동기화 완료\n"
                    response += f"Drive ID: `{file_id[:12]}...`"
                else:
                    response += f"⚠️  Drive 동기화 실패 (로컬 저장은 완료)"
            except Exception as e:
                logger.error(f"❌ Drive 동기화 오류: {e}")
                response += f"⚠️  Drive 동기화 실패 (로컬 저장은 완료)"
        else:
            response += f"ℹ️  로컬 저장만 완료 (Drive 연동 비활성화)"

        await update.message.reply_text(response)

    async def sync_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """
        /sync - 수동 클라우드 동기화 (INTELLIGENCE_QUANTA + 리포트)
        """
        user = update.effective_user
        logger.info(f"☁️  /sync from {user.first_name} ({user.id})")

        if not self.gdrive:
            await update.message.reply_text(
                "⚠️  Google Drive 연동이 비활성화되어 있습니다.\n"
                "credentials/gdrive_auth.json 및 .env 설정을 확인하세요."
            )
            return

        await update.message.reply_text(
            "☁️  클라우드 동기화 시작...\n"
            "잠시만 기다려주세요."
        )

        try:
            results = []

            # 1. INTELLIGENCE_QUANTA.md 동기화
            await update.message.reply_text("📤 1/2: INTELLIGENCE_QUANTA.md 동기화 중...")
            quanta_success = self.gdrive.sync_intelligence_quanta()
            results.append(("INTELLIGENCE_QUANTA.md", quanta_success))

            # 2. 일일 리포트 동기화
            await update.message.reply_text("📤 2/2: 일일 리포트 동기화 중...")
            report_results = self.gdrive.sync_daily_reports()
            results.append(("Daily Reports", len(report_results) > 0))

            # 결과 요약
            response = "☁️  **클라우드 동기화 완료**\n\n"
            response += "📊 동기화 결과:\n"

            for item, success in results:
                status = "✅" if success else "❌"
                response += f"   {status} {item}\n"

            if report_results:
                success_count = sum(report_results.values())
                total_count = len(report_results)
                response += f"\n   • 리포트: {success_count}/{total_count}개 성공\n"

            response += f"\n💡 슬로우 라이프 리마인더:\n"
            response += f"과정의 흔적이 클라우드에도 보존되었습니다."

            await update.message.reply_text(response)

        except Exception as e:
            logger.error(f"❌ 동기화 오류: {e}")
            await update.message.reply_text(
                f"❌ 동기화 중 오류 발생:\n{str(e)}"
            )

    async def youtube_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """
        /youtube <URL> - Anti-Gravity YouTube 분석 (NotebookLM)

        NotebookLM 기반 RAG 분석:
        1. 3줄 요약
        2. 핵심 인사이트 (Aesop 스타일)
        3. 브랜드 연결 (5 Pillars)
        4. Audio Overview (Deep Dive Podcast)
        """
        user = update.effective_user
        url = ' '.join(context.args) if context.args else None

        if not url:
            await update.message.reply_text(
                "🛸 **Anti-Gravity YouTube Analyzer** (NotebookLM)\n\n"
                "사용법: /youtube <YouTube URL>\n\n"
                "예시:\n"
                "  /youtube https://youtu.be/xxxxx\n"
                "  /youtube https://www.youtube.com/watch?v=xxxxx\n\n"
                "📦 생성 결과:\n"
                "  📝 3줄 요약\n"
                "  💡 핵심 인사이트 (Aesop 스타일)\n"
                "  🎯 브랜드 연결 (5 Pillars)\n"
                "  🎙️  Audio Overview (Google Gemini)\n\n"
                "💡 Source Grounding:\n"
                "  실제 YouTube Transcript 기반, 환각 제로"
            )
            return

        logger.info(f"🛸 /youtube from {user.first_name}: {url}")

        # 진행 상황 알림
        progress_message = await update.message.reply_text(
            "🛸 **Anti-Gravity 프로토콜 시작** (NotebookLM)\n\n"
            "⏳ [1/4] 노트북 생성 중...\n"
            "⏳ [2/4] YouTube 소스 추가 대기\n"
            "⏳ [3/4] RAG 분석 대기\n"
            "⏳ [4/4] Audio Overview 생성 대기\n\n"
            "예상 소요 시간: 2-3분"
        )

        try:
            # Step 1: 노트북 생성
            await progress_message.edit_text(
                "🛸 **Anti-Gravity 프로토콜 진행 중**\n\n"
                "✅ [1/4] 노트북 생성 중...\n"
                "⏳ [2/4] YouTube 소스 추가 대기\n"
                "⏳ [3/4] RAG 분석 대기\n"
                "⏳ [4/4] Audio Overview 생성 대기"
            )

            # Anti-Gravity 분석 실행 (NotebookLM)
            result = await asyncio.to_thread(
                anti_gravity_youtube,
                url
            )

            if not result or not result.get('notebook_id'):
                await progress_message.edit_text(
                    "❌ **분석 실패**\n\n"
                    "NotebookLM 분석에 실패했습니다.\n\n"
                    "가능한 원인:\n"
                    "  • 잘못된 YouTube URL\n"
                    "  • 자막이 없는 영상\n"
                    "  • NotebookLM 인증 만료\n\n"
                    "다른 영상으로 시도해보세요."
                )
                return

            # Step 2-4: 소스 추가 및 분석 완료
            await progress_message.edit_text(
                "🛸 **Anti-Gravity 프로토콜 진행 중**\n\n"
                "✅ [1/4] 노트북 생성 완료\n"
                "✅ [2/4] YouTube 소스 추가 완료\n"
                "✅ [3/4] RAG 분석 완료\n"
                "✅ [4/4] Audio Overview 생성 중...\n\n"
                "결과 전송 중..."
            )

            # 최종 결과 전송
            response = "✅ **Anti-Gravity 분석 완료** (NotebookLM)\n\n"
            response += f"🔗 Source: `{url}`\n"
            response += f"📓 Notebook: https://notebooklm.google.com/notebook/{result['notebook_id']}\n\n"

            # 3줄 요약
            if result.get('summary'):
                response += f"📝 **3줄 요약**:\n{result['summary']}\n\n"

            # 핵심 인사이트
            if result.get('insights'):
                response += f"💡 **핵심 인사이트** (Aesop 스타일):\n{result['insights']}\n\n"

            # 브랜드 연결
            if result.get('brand_connection'):
                response += f"🎯 **브랜드 연결** (5 Pillars):\n{result['brand_connection']}\n\n"

            # Audio Overview
            response += f"🎙️  **Audio Overview**: 생성 중 (비동기)\n"
            response += f"   위 Notebook 링크에서 확인 가능\n\n"

            response += "💡 **Anti-Gravity 원칙**:\n"
            response += "   ✅ Source Grounding (YouTube Transcript)\n"
            response += "   ✅ Multi-modal Synthesis (Text + Audio)\n"
            response += "   ✅ MCP Connector (NotebookLM)\n\n"

            response += "📂 NotebookLM에서 추가 질문 가능:\n"
            response += "   • 화자의 주요 주장은?\n"
            response += "   • 실용적 적용 방법은?\n"
            response += "   • 다른 개념과의 연결점은?"

            await progress_message.edit_text(response)

        except Exception as e:
            logger.error(f"❌ YouTube 분석 오류: {e}")
            await progress_message.edit_text(
                f"❌ **분석 중 오류 발생**\n\n"
                f"오류: {str(e)}\n\n"
                f"💡 문제 해결:\n"
                f"  • URL 형식 확인\n"
                f"  • NotebookLM 인증 확인 (macOS에서 nlm login)\n"
                f"  • 로그 확인: logs/telegram_secretary.log"
            )

    async def handle_photo(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """
        이미지 자동 포착
        """
        user = update.effective_user
        photo = update.message.photo[-1]  # 가장 큰 사이즈
        caption = update.message.caption or ""

        logger.info(f"🖼️  Photo from {user.first_name}")

        # 이미지 다운로드
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        photo_dir = PROJECT_ROOT / 'knowledge' / 'signals' / 'images'
        photo_dir.mkdir(parents=True, exist_ok=True)

        photo_path = photo_dir / f'photo_{timestamp}.jpg'
        file = await photo.get_file()
        await file.download_to_drive(photo_path)

        # 신호 메타데이터 저장
        signal_path = PROJECT_ROOT / 'knowledge' / 'signals' / f'image_{timestamp}.md'
        with open(signal_path, 'w', encoding='utf-8') as f:
            f.write(f"# Image Signal {timestamp}\n\n")
            f.write(f"**From**: {user.first_name} (@{user.username or 'unknown'})\n")
            f.write(f"**Time**: {datetime.now().isoformat()}\n")
            f.write(f"**Type**: image\n")
            f.write(f"**File**: {photo_path}\n\n")
            if caption:
                f.write(f"## Caption\n\n{caption}\n")

        await update.message.reply_text(
            f"🖼️  이미지 신호 포착: `image_{timestamp}`\n"
            f"비주얼 분석을 위해 /analyze를 입력하세요."
        )

    def run(self):
        """
        Secretary 실행 (Blocking)
        """
        logger.info("🚀 Telegram Secretary 시작...")

        # Application 생성
        application = (
            Application.builder()
            .token(self.bot_token)
            .build()
        )

        # Command handlers
        application.add_handler(CommandHandler("start", self.start_command))
        application.add_handler(CommandHandler("status", self.status_command))
        application.add_handler(CommandHandler("report", self.report_command))
        application.add_handler(CommandHandler("analyze", self.analyze_command))
        application.add_handler(CommandHandler("signal", self.signal_command))
        application.add_handler(CommandHandler("morning", self.morning_command))
        application.add_handler(CommandHandler("evening", self.evening_command))

        # Phase 2.4: Secretary functions (Google Drive integration)
        application.add_handler(CommandHandler("search", self.search_command))
        application.add_handler(CommandHandler("memo", self.memo_command))
        application.add_handler(CommandHandler("sync", self.sync_command))

        # Phase 3: Anti-Gravity Protocol (YouTube Analyzer)
        application.add_handler(CommandHandler("youtube", self.youtube_command))

        # Message handlers
        application.add_handler(
            MessageHandler(filters.TEXT & ~filters.COMMAND, self.handle_message)
        )
        application.add_handler(
            MessageHandler(filters.PHOTO, self.handle_photo)
        )

        # Error handler
        async def error_handler(update: Update, context: ContextTypes.DEFAULT_TYPE):
            logger.error(f"❌ Error: {context.error}")
            if update and update.message:
                await update.message.reply_text(
                    f"⚠️  오류가 발생했습니다:\n{context.error}"
                )

        application.add_error_handler(error_handler)

        # Start polling
        logger.info("✅ Telegram Secretary 준비 완료 - 신호 대기 중...")
        application.run_polling(allowed_updates=Update.ALL_TYPES)

    def cleanup(self):
        """종료 시 정리"""
        logger.info("🔄 Telegram Secretary 종료 중...")

        # Work lock 해제
        self.handoff.release_work_lock("TelegramSecretary")

        # Handoff 저장
        self.handoff.handoff(
            agent_id="TelegramSecretary",
            summary="Telegram Secretary 정상 종료",
            next_steps=["재시작 시 세션 연속성 복구"]
        )

        logger.info("✅ 정리 완료")


def main():
    """Main entry point"""
    # Bot token 확인
    bot_token = os.getenv("TELEGRAM_BOT_TOKEN")
    if not bot_token:
        logger.error("❌ TELEGRAM_BOT_TOKEN이 설정되지 않았습니다.")
        logger.error("   .env 파일에 TELEGRAM_BOT_TOKEN을 추가하세요.")
        sys.exit(1)

    # Logs 디렉토리 생성
    log_dir = PROJECT_ROOT / 'logs'
    log_dir.mkdir(parents=True, exist_ok=True)

    # Secretary 시작
    secretary = TelegramSecretary(bot_token)

    try:
        secretary.run()
    except KeyboardInterrupt:
        logger.info("\n⏹️  사용자 중단 (Ctrl+C)")
    except Exception as e:
        logger.error(f"❌ 예상치 못한 오류: {e}")
        import traceback
        traceback.print_exc()
    finally:
        secretary.cleanup()


if __name__ == "__main__":
    main()
