#!/usr/bin/env python3
"""
97layerOS Telegram Secretary v6 - JARVIS Plus Edition
초고도화된 AI 비서: Deep RAG + Premium UX + Multi-Agent Visibility

Features:
- 🧠 Deep RAG: NotebookLM MCP 직접 연동 (Knowledge Base 심층 검색)
- 💎 Premium UX: 절제된 포맷팅 및 에이전트 상태 리포팅
- 🎥 YouTube Pro: 자동 분석 + 오디오 개요 + 마인드맵 생성
- 🤖 Auto-Pilot: 인사이트 자동 분류 및 에이전트 워크플로우 트리거
"""

import os
import sys
import re
import logging
import json
from pathlib import Path
from typing import Dict, Optional
from datetime import datetime
import asyncio

# Add project root to path
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

# Load .env
try:
    from dotenv import load_dotenv
    load_dotenv(PROJECT_ROOT / '.env')
except ImportError:
    pass

from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup, constants
from telegram.ext import (
    Application,
    CommandHandler,
    MessageHandler,
    CallbackQueryHandler,
    filters,
    ContextTypes
)

# Core Components
from core.bridges.notebooklm_bridge import get_bridge
from core.system.conversation_engine import get_conversation_engine
from core.system.intent_classifier import get_intent_classifier
from core.system.youtube_analyzer import YouTubeAnalyzer
from core.system.image_analyzer import ImageAnalyzer

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


def _escape_html(text: str) -> str:
    """Telegram HTML 모드용 이스케이프"""
    return (text
            .replace('&', '&amp;')
            .replace('<', '&lt;')
            .replace('>', '&gt;'))


def admin_only(func):
    """ADMIN_TELEGRAM_ID만 커맨드 실행 가능"""
    async def wrapper(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        admin_id = os.getenv('ADMIN_TELEGRAM_ID')
        if admin_id and str(update.effective_user.id) != str(admin_id):
            await update.message.reply_text("권한 없음")
            return
        return await func(self, update, context)
    wrapper.__name__ = func.__name__
    return wrapper


class TelegramSecretaryV6:
    def __init__(self, bot_token: str):
        self.bot_token = bot_token
        self.notebooklm = get_bridge()
        self.engine = get_conversation_engine()
        self.classifier = get_intent_classifier()
        self.youtube = YouTubeAnalyzer()
        self.image = ImageAnalyzer()

        # Gardener (선택적 — 없어도 봇 동작)
        try:
            from core.agents.gardener import Gardener
            self.gardener = Gardener()
        except Exception as e:
            self.gardener = None
            logger.warning("Gardener 비활성: %s", e)

        # UI Settings
        self.loading_emojis = ["🔘", "⚪", "⚫"]

        logger.info("✅ Telegram Secretary V6 (JARVIS Plus) initialized")

    async def start_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        user = update.effective_user
        welcome_msg = (
            f"<b>97layerOS</b>\n\n"
            f"안녕하세요, {_escape_html(user.first_name)}님.\n\n"
            f"<b>명령어</b>\n"
            f"/status — 파이프라인 현황\n"
            f"/publish — 콘텐츠 즉시 발행\n"
            f"/publish [테마] — 테마 지정 발행\n"
            f"/report — 오늘 처리 요약\n"
            f"/growth [YYYY-MM] — 성장 지표\n"
            f"/signal [텍스트] — 신호 직접 투입\n"
            f"/client list|add|info|due — 고객 관리\n"
            f"/visit <이름> <서비스> [만족도] — 방문 기록\n"
            f"/pending — 가드너 제안 목록\n\n"
            f"<b>자동 처리</b>\n"
            f"텍스트 → 신호 수집\n"
            f"YouTube 링크 → 영상 분석\n"
            f"이미지 → 브랜드 인사이트 추출"
        )
        await update.message.reply_text(welcome_msg, parse_mode=constants.ParseMode.HTML)

    @admin_only
    async def growth_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """월별 성장 지표 리포트 (Growth Module + 시스템 현황)"""
        from core.modules.growth import get_growth_module

        # 기간 인자 파싱
        period = context.args[0] if context.args else datetime.now().strftime('%Y-%m')

        try:
            gm = get_growth_module()
            gm.auto_count_content(period)
            gm.auto_count_service(period)
            data = gm.get_month(period)

            revenue = data.get('revenue', {})
            content = data.get('content', {})
            service = data.get('service', {})

            # 시스템 현황 (누적)
            signals_dir = PROJECT_ROOT / 'knowledge' / 'signals'
            total_signals = len(list(signals_dir.glob('**/*.json'))) if signals_dir.exists() else 0

            msg = (
                f"<b>📈 Growth Report — {_escape_html(period)}</b>\n\n"
                f"<b>수익</b>\n"
                f"아틀리에: {revenue.get('atelier', 0):,}원\n"
                f"컨설팅: {revenue.get('consulting', 0):,}원\n"
                f"제품: {revenue.get('products', 0):,}원\n"
                f"합계: <b>{revenue.get('total', 0):,}원</b>\n\n"
                f"<b>콘텐츠</b>\n"
                f"신호 수집: {content.get('signals_captured', 0)}건\n"
                f"발행 에세이: {content.get('essays_published', 0)}개\n"
                f"성숙 군집: {content.get('clusters_ripe', 0)}개\n\n"
                f"<b>서비스</b>\n"
                f"총 방문: {service.get('total_visits', 0)}건\n"
                f"신규 고객: {service.get('new_clients', 0)}명\n"
                f"재방문 고객: {service.get('returning_clients', 0)}명\n\n"
                f"<b>시스템 누적</b>\n"
                f"전체 신호: {total_signals}개"
            )
            await update.message.reply_text(msg, parse_mode=constants.ParseMode.HTML)

        except Exception as e:
            logger.error("growth_command error: %s", e)
            await update.message.reply_text("지표 조회 중 오류가 발생했습니다.")

    @admin_only
    async def client_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Ritual Module 고객 관리 (/client list|add|info|due)"""
        from core.modules.ritual import get_ritual_module

        args = context.args
        subcmd = args[0] if args else 'list'

        try:
            rm = get_ritual_module()

            if subcmd == 'list':
                clients = rm.list_clients()
                if not clients:
                    await update.message.reply_text("등록된 고객 없음")
                    return
                lines = ["<b>고객 목록</b>\n"]
                for c in clients:
                    visits = len(c.get('visits', []))
                    lines.append(
                        f"{_escape_html(c['client_id'])} | {_escape_html(c['name'])} | "
                        f"방문 {visits}회 | 리듬: {_escape_html(c.get('rhythm', '보통'))}"
                    )
                await update.message.reply_text('\n'.join(lines), parse_mode=constants.ParseMode.HTML)

            elif subcmd == 'add':
                if len(args) < 2:
                    await update.message.reply_text("사용법: /client add <이름>")
                    return
                name = args[1]
                client = rm.create_client(name)
                msg = (
                    f"✅ 고객 등록\n"
                    f"ID: {_escape_html(client['client_id'])}\n"
                    f"이름: {_escape_html(client['name'])}"
                )
                await update.message.reply_text(msg)

            elif subcmd == 'info':
                if len(args) < 2:
                    await update.message.reply_text("사용법: /client info <이름>")
                    return
                name = args[1]
                client = rm.find_client(name)
                if not client:
                    await update.message.reply_text("고객을 찾을 수 없음")
                    return
                revisit = rm.check_revisit_due(client['client_id'])
                visits = client.get('visits', [])
                last_visit = client.get('last_visit') or '없음'
                due_text = f"⚠️ 재방문 필요 ({revisit['days_since']}일 경과)" if revisit.get('due') else "정상"
                msg = (
                    f"<b>{_escape_html(client['name'])}</b>\n"
                    f"ID: {_escape_html(client['client_id'])}\n"
                    f"모발: {_escape_html(client.get('hair_type', '미기입'))}\n"
                    f"리듬: {_escape_html(client.get('rhythm', '보통'))}\n"
                    f"방문: {len(visits)}회 | 마지막: {_escape_html(last_visit)}\n"
                    f"재방문: {_escape_html(due_text)}"
                )
                await update.message.reply_text(msg, parse_mode=constants.ParseMode.HTML)

            elif subcmd == 'due':
                due_list = rm.get_due_clients()
                if not due_list:
                    await update.message.reply_text("재방문 대상 없음")
                    return
                lines = ["<b>재방문 대상</b>\n"]
                for d in due_list:
                    lines.append(
                        f"{_escape_html(d['client_name'])} | {d['days_since']}일 경과 | "
                        f"임계값: {d['threshold']}일"
                    )
                await update.message.reply_text('\n'.join(lines), parse_mode=constants.ParseMode.HTML)

            else:
                await update.message.reply_text(
                    "사용법: /client list | add <이름> | info <이름> | due"
                )

        except Exception as e:
            logger.error("client_command error: %s", e)
            await update.message.reply_text("고객 조회 중 오류가 발생했습니다.")

    @admin_only
    async def visit_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """방문 기록 (/visit <이름> <서비스> [만족도])"""
        from core.modules.ritual import get_ritual_module

        args = context.args
        if len(args) < 2:
            await update.message.reply_text("사용법: /visit <이름> <서비스> [만족도(1-5)]")
            return

        name = args[0]
        service = args[1]
        satisfaction = None
        if len(args) >= 3:
            try:
                satisfaction = int(args[2])
            except ValueError:
                pass

        try:
            rm = get_ritual_module()
            client = rm.find_client(name)
            if not client:
                await update.message.reply_text(f"고객 없음: {_escape_html(name)}")
                return

            visit = rm.add_visit(
                client['client_id'],
                service=service,
                satisfaction=satisfaction,
            )
            sat_text = f" | 만족도: {satisfaction}/5" if satisfaction else ""
            msg = (
                f"✅ 방문 기록\n"
                f"{_escape_html(client['name'])} | {_escape_html(service)}"
                f"{_escape_html(sat_text)}\n"
                f"날짜: {_escape_html(visit['date'])}"
            )
            await update.message.reply_text(msg)

        except Exception as e:
            logger.error("visit_command error: %s", e)
            await update.message.reply_text("방문 기록 중 오류가 발생했습니다.")

    async def handle_message(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        message = update.message
        if not message.text and not message.photo and not message.document:
            return

        # 1. YouTube 전역 감지
        youtube_match = re.search(r'(https?://(?:www\.)?(youtube\.com|youtu\.be)/[\w-]+)', message.text or '')
        if youtube_match:
            await self.process_youtube(update, context, youtube_match.group(1))
            return

        # 2. 이미지 처리
        if message.photo:
            await self.process_image(update, context)
            return

        # 3. PDF 문서 처리
        if message.document:
            await self.process_document(update, context)
            return

        # 4. 텍스트 의도 분석 및 처리
        await self.process_text(update, context)

    async def process_youtube(self, update: Update, context: ContextTypes.DEFAULT_TYPE, url: str):
        status_msg = await update.message.reply_text(
            "🛸 YouTube 분석 중...",
            parse_mode=constants.ParseMode.HTML
        )

        try:
            # 1. 자막 수집 + Gemini 분석 + 로컬 JSON 저장
            await status_msg.edit_text("🛸 <code>Step 1/2</code>: 영상 자막 수집 중...", parse_mode=constants.ParseMode.HTML)
            result = self.youtube.process_url(url, source_channel="telegram")

            if not result['success']:
                await status_msg.edit_text(
                    f"❌ 분석 실패: {_escape_html(str(result.get('error', '')))}",
                    parse_mode=constants.ParseMode.HTML
                )
                return

            video_id = result['video_id']
            transcript_len = len(result.get('transcript', ''))
            analysis = result.get('analysis', {})

            # 2. NotebookLM 저장 시도 (실패해도 전체 흐름 중단 안 함)
            nlm_saved = False
            if self.notebooklm and self.notebooklm.authenticated:
                await status_msg.edit_text("🛸 <code>Step 2/2</code>: NotebookLM 저장 중...", parse_mode=constants.ParseMode.HTML)
                try:
                    content_text = (
                        f"YouTube: {url}\n"
                        f"Video ID: {video_id}\n\n"
                        f"분석:\n{json.dumps(analysis, ensure_ascii=False, indent=2)}\n\n"
                        f"자막 요약:\n{result.get('transcript', '')[:3000]}"
                    )
                    nb_id = self.notebooklm.get_or_create_notebook("97layerOS: Signal Archive")
                    title = f"[YouTube] {analysis.get('title', video_id)[:60]}"
                    self.notebooklm.add_source_text(nb_id, content_text, title)
                    nlm_saved = True
                except Exception as nlm_e:
                    logger.warning("NotebookLM 저장 실패: %s", nlm_e)

            # 결과 메시지 — 실제 완료된 것만 표시
            lines = [f"✅ <b>YouTube 분석 완료</b>", ""]
            lines.append(f"ID: <code>{_escape_html(video_id)}</code>")
            lines.append(f"자막: {transcript_len}자 수집")
            lines.append(f"로컬 저장: ✅")
            lines.append(f"NotebookLM: {'✅ 저장됨' if nlm_saved else '⚠️ 저장 실패 (로컬만)'}")

            await status_msg.edit_text("\n".join(lines), parse_mode=constants.ParseMode.HTML)

        except Exception as e:
            logger.error("YouTube processing error: %s", e)
            try:
                await status_msg.edit_text(f"❌ 시스템 오류: {_escape_html(str(e))}", parse_mode=constants.ParseMode.HTML)
            except Exception:
                pass

    async def process_image(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """이미지 처리 — 통합 스키마로 signals/ 저장 + SA 큐 전달"""
        caption = update.message.caption or ""
        status_msg = await update.message.reply_text("🖼️ 이미지 분석 중...")
        try:
            photo = update.message.photo[-1]
            file = await context.bot.get_file(photo.file_id)
            import tempfile
            import shutil
            with tempfile.NamedTemporaryFile(suffix='.jpg', delete=False) as tmp:
                tmp_path = tmp.name
                await file.download_to_drive(tmp_path)

            # Gemini Vision 분석
            result = self.image.analyze_image(tmp_path)

            # 통합 스키마로 저장
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            signal_id = "image_%s" % timestamp
            files_dir = PROJECT_ROOT / 'knowledge' / 'signals' / 'files'
            files_dir.mkdir(parents=True, exist_ok=True)

            # 바이너리 복사
            dest_path = files_dir / ("%s.jpg" % signal_id)
            shutil.copy2(tmp_path, dest_path)
            os.unlink(tmp_path)

            # 통합 신호 JSON
            description = result.get('full_analysis') or result.get('description', '')
            content = description[:3000]
            if caption:
                content = "[메모] %s\n\n%s" % (caption, content)

            signal_data = {
                'signal_id': signal_id,
                'type': 'image',
                'status': 'captured',
                'content': content,
                'captured_at': datetime.now().isoformat(),
                'from_user': update.effective_user.username or update.effective_user.first_name,
                'source_channel': 'telegram',
                'metadata': {
                    'image_path': str(dest_path),
                    'title': caption[:100] if caption else 'telegram_image',
                },
            }

            signals_dir = PROJECT_ROOT / 'knowledge' / 'signals'
            signals_dir.mkdir(parents=True, exist_ok=True)
            with open(signals_dir / ("%s.json" % signal_id), 'w', encoding='utf-8') as f:
                json.dump(signal_data, f, ensure_ascii=False, indent=2)

            # SA 큐 전달
            try:
                from core.system.queue_manager import QueueManager
                qm = QueueManager()
                qm.create_task(
                    agent_type='SA',
                    task_type='analyze_signal',
                    payload={
                        'signal_id': signal_id,
                        'signal_path': str(signals_dir / ("%s.json" % signal_id)),
                    }
                )
                logger.info("SA 큐 전달 (이미지): %s", signal_id)
            except Exception as q_e:
                logger.warning("SA 큐 전달 실패: %s", q_e)

            # 응답
            if description and '분석 실패' not in description:
                combined = description
                if caption:
                    combined = "📝 **메모**: %s\n\n%s" % (caption, description)
                if len(combined) > 4000:
                    combined = combined[:4000] + "\n\n..."
                await status_msg.edit_text(combined)
            else:
                err = result.get('description', '알 수 없는 오류')
                await status_msg.edit_text(
                    "❌ 이미지 분석 실패: %s" % _escape_html(err),
                    parse_mode=constants.ParseMode.HTML
                )

        except Exception as e:
            logger.error("Image processing error: %s", e)
            try:
                await status_msg.edit_text(
                    "❌ 이미지 처리 오류: %s" % _escape_html(str(e)),
                    parse_mode=constants.ParseMode.HTML
                )
            except Exception:
                pass

    async def process_document(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """PDF/문서 처리 — 통합 스키마로 signals/ 저장 + SA 큐 전달"""
        doc = update.message.document
        file_name = doc.file_name or "unknown"
        mime = doc.mime_type or ""

        # PDF만 지원
        if not (mime == "application/pdf" or file_name.lower().endswith(".pdf")):
            await update.message.reply_text(
                "현재 PDF 파일만 지원합니다. (%s)" % _escape_html(file_name),
                parse_mode=constants.ParseMode.HTML,
            )
            return

        status_msg = await update.message.reply_text("📄 PDF 수집 중...")
        try:
            import tempfile
            import shutil

            file = await context.bot.get_file(doc.file_id)
            with tempfile.NamedTemporaryFile(suffix=".pdf", delete=False) as tmp:
                tmp_path = tmp.name
                await file.download_to_drive(tmp_path)

            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            signal_id = "pdf_document_%s" % timestamp

            # 바이너리 복사
            files_dir = PROJECT_ROOT / "knowledge" / "signals" / "files"
            files_dir.mkdir(parents=True, exist_ok=True)
            dest_path = files_dir / ("%s.pdf" % signal_id)
            shutil.copy2(tmp_path, dest_path)

            # 텍스트 추출 시도
            content = "PDF 수집: %s" % file_name
            pages_extracted = 0
            try:
                import pdfplumber
                with pdfplumber.open(tmp_path) as pdf:
                    pages_text = []
                    for page in pdf.pages[:10]:
                        text = page.extract_text()
                        if text:
                            pages_text.append(text)
                    if pages_text:
                        content = "\n".join(pages_text)[:3000]
                        pages_extracted = len(pages_text)
            except ImportError:
                logger.info("pdfplumber 미설치 — 텍스트 추출 생략")
            except Exception as pdf_e:
                logger.warning("PDF 추출 실패: %s", pdf_e)

            os.unlink(tmp_path)

            # 통합 신호 JSON
            signal_data = {
                "signal_id": signal_id,
                "type": "pdf_document",
                "status": "captured",
                "content": content,
                "captured_at": datetime.now().isoformat(),
                "from_user": update.effective_user.username or update.effective_user.first_name,
                "source_channel": "telegram",
                "metadata": {
                    "pdf_path": str(dest_path),
                    "title": Path(file_name).stem,
                },
            }

            signals_dir = PROJECT_ROOT / "knowledge" / "signals"
            signals_dir.mkdir(parents=True, exist_ok=True)
            with open(signals_dir / ("%s.json" % signal_id), "w", encoding="utf-8") as f:
                json.dump(signal_data, f, ensure_ascii=False, indent=2)

            # SA 큐 전달
            try:
                from core.system.queue_manager import QueueManager
                qm = QueueManager()
                qm.create_task(
                    agent_type="SA",
                    task_type="analyze_signal",
                    payload={
                        "signal_id": signal_id,
                        "signal_path": str(signals_dir / ("%s.json" % signal_id)),
                    },
                )
                logger.info("SA 큐 전달 (PDF): %s", signal_id)
            except Exception as q_e:
                logger.warning("SA 큐 전달 실패: %s", q_e)

            extract_info = ""
            if pages_extracted:
                extract_info = "\n텍스트 추출: %d페이지, %d자" % (pages_extracted, len(content))
            await status_msg.edit_text(
                "📄 <b>PDF 수집 완료</b>\n\n"
                "파일: %s%s\n"
                "SA 분석 큐 전달됨." % (_escape_html(file_name), extract_info),
                parse_mode=constants.ParseMode.HTML,
            )

        except Exception as e:
            logger.error("Document processing error: %s", e)
            try:
                await status_msg.edit_text(
                    "❌ PDF 처리 오류: %s" % _escape_html(str(e)),
                    parse_mode=constants.ParseMode.HTML,
                )
            except Exception:
                pass

    async def process_text(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        text = update.message.text
        try:
            # 의도 분류
            intent_data = self.classifier.classify(text)
            intent = intent_data['intent']

            if intent == 'insight':
                # 먼저 저장 실행
                self._save_insight(text, update.effective_user)

                # 저장 완료 후 응답 (실제로 한 것만 표시)
                timestamp = datetime.now().strftime('%H:%M:%S')
                preview = _escape_html(text[:150])
                await update.message.reply_text(
                    f"💾 <b>Captured</b> (<code>{timestamp}</code>)\n\n"
                    f"\"{preview}\"\n\n"
                    f"signals/ 저장 완료. SA가 분석 중입니다.",
                    parse_mode=constants.ParseMode.HTML
                )
            else:
                # 대화 및 질문 (Deep RAG)
                placeholder = await update.message.reply_text("💭 사유 중...")
                try:
                    response = self.engine.chat(str(update.effective_user.id), text)
                    # Gemini 응답은 parse_mode 없이 순수 텍스트로 전송
                    # (마크다운 파싱 오류 방지)
                    await placeholder.edit_text(response)
                except Exception as chat_e:
                    logger.error("Chat engine error: %s", chat_e)
                    await placeholder.edit_text("죄송합니다. 응답 생성 중 오류가 발생했습니다.")

        except Exception as e:
            logger.error("process_text error: %s", e)
            try:
                await update.message.reply_text("처리 중 오류가 발생했습니다. 다시 시도해주십시오.")
            except Exception:
                pass

    def _save_insight(self, text: str, user):
        """
        인사이트 저장 + SA 에이전트 분석 큐에 전달.
        signals/ 파일 저장은 항상 성공. 큐 전달은 실패해도 조용히 스킵.
        """
        signals_dir = PROJECT_ROOT / 'knowledge' / 'signals'
        signals_dir.mkdir(parents=True, exist_ok=True)
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        signal_id = f"text_{timestamp}"

        signal_data = {
            'signal_id': signal_id,
            'type': 'text_insight',
            'status': 'captured',
            'content': text,
            'captured_at': datetime.now().isoformat(),
            'from_user': user.username or user.first_name,
            'source_channel': 'telegram',
            'metadata': {},
        }

        # 1. signals/ 저장 (항상)
        with open(signals_dir / f"{signal_id}.json", 'w', encoding='utf-8') as f:
            json.dump(signal_data, f, ensure_ascii=False, indent=2)

        # 2. SA 에이전트 큐에 분석 요청 (THE CYCLE 트리거)
        try:
            from core.system.queue_manager import QueueManager
            qm = QueueManager()
            qm.create_task(
                agent_type='SA',
                task_type='analyze',
                payload={
                    'signal_id': signal_id,
                    'content': text,
                    'source': 'telegram_insight',
                }
            )
            logger.info("SA 큐 전달: %s", signal_id)
        except Exception as q_e:
            logger.warning("SA 큐 전달 실패 (signals/ 저장은 완료): %s", q_e)

    @admin_only
    async def status_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """/status — 파이프라인 현황 스냅샷"""
        try:
            signals_dir = PROJECT_ROOT / 'knowledge' / 'signals'
            corpus_index = PROJECT_ROOT / 'knowledge' / 'corpus' / 'index.json'
            queue_completed = PROJECT_ROOT / '.infra' / 'queue' / 'tasks' / 'completed'
            queue_pending = PROJECT_ROOT / '.infra' / 'queue' / 'tasks' / 'pending'

            sig_stats: Dict = {}
            for f in signals_dir.glob('*.json'):
                try:
                    d = json.loads(f.read_text())
                    s = d.get('status', 'unknown')
                    sig_stats[s] = sig_stats.get(s, 0) + 1
                except Exception:
                    pass

            clusters_total = 0
            published_count = 0
            if corpus_index.exists():
                ci = json.loads(corpus_index.read_text())
                clusters_total = len(ci.get('clusters', {}))
                published_count = len(ci.get('published', []))

            pending_cnt = len(list(queue_pending.glob('*.json'))) if queue_pending.exists() else 0
            completed_cnt = len(list(queue_completed.glob('*.json'))) if queue_completed.exists() else 0

            from datetime import datetime as _dt
            now = _dt.now()
            next_g = now.replace(hour=3, minute=0, second=0, microsecond=0)
            if next_g <= now:
                from datetime import timedelta
                next_g += timedelta(days=1)
            delta = next_g - now
            h, rem = divmod(int(delta.total_seconds()), 3600)
            m = rem // 60

            total_sigs = sum(sig_stats.values())
            msg = (
                f"<b>⚙️ 파이프라인 현황</b>\n\n"
                f"<b>신호</b>\n"
                f"총 {total_sigs}개 | 수집 {sig_stats.get('captured',0)} | "
                f"분석완료 {sig_stats.get('analyzed',0)}\n\n"
                f"<b>Corpus</b>\n"
                f"군집 {clusters_total}개 | 발행됨 {published_count}개\n\n"
                f"<b>태스크 큐</b>\n"
                f"대기 {pending_cnt}개 | 완료 {completed_cnt}개\n\n"
                f"<b>Gardener</b>\n"
                f"다음 실행: {h}시간 {m}분 후 (03:00)"
            )
            await update.message.reply_text(msg, parse_mode=constants.ParseMode.HTML)

        except Exception as e:
            logger.error("status_command error: %s", e)
            await update.message.reply_text(f"상태 조회 오류: {_escape_html(str(e))}", parse_mode=constants.ParseMode.HTML)

    @admin_only
    async def publish_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """/publish [테마] — Corpus 군집 즉시 발행 트리거"""
        theme_arg = ' '.join(context.args).strip() if context.args else None
        try:
            from core.system.corpus_manager import CorpusManager
            corpus = CorpusManager()
            corpus_index_path = PROJECT_ROOT / 'knowledge' / 'corpus' / 'index.json'

            if theme_arg:
                index = json.loads(corpus_index_path.read_text()) if corpus_index_path.exists() else {}
                clusters = index.get('clusters', {})
                matched = next((v for k, v in clusters.items() if theme_arg in k), None)
                if not matched:
                    available = ', '.join(clusters.keys()) or '없음'
                    await update.message.reply_text(
                        f"테마 <b>{_escape_html(theme_arg)}</b> 군집 없음.\n"
                        f"현재 군집: {_escape_html(available)}",
                        parse_mode=constants.ParseMode.HTML
                    )
                    return
                ripe = [matched]
                forced = True
            else:
                ripe = corpus.get_ripe_clusters()
                forced = False

            if not ripe:
                index = json.loads(corpus_index_path.read_text()) if corpus_index_path.exists() else {}
                clusters = index.get('clusters', {})
                if clusters:
                    lines = ["<b>발행 가능한 군집 없음</b> (성숙도 미달)\n", "<b>현재 군집:</b>"]
                    for theme, c in clusters.items():
                        cnt = len(c.get('entry_ids', []))
                        lines.append(f"  • {_escape_html(theme)}: {cnt}개 신호")
                    lines.append("\n조건: 5개+ 신호, 72시간+ 분포")
                    await update.message.reply_text('\n'.join(lines), parse_mode=constants.ParseMode.HTML)
                else:
                    await update.message.reply_text(
                        "아직 Corpus 군집이 없습니다.\n신호가 분석되면 자동으로 군집이 형성됩니다."
                    )
                return

            from core.system.queue_manager import QueueManager
            qm = QueueManager()
            triggered = []
            for cluster in ripe:
                theme = cluster.get('theme', 'unknown')
                qm.create_task(
                    agent_type='CE',
                    task_type='write_corpus_essay',
                    payload={
                        'cluster': cluster,
                        'forced': forced,
                        'triggered_by': 'telegram_publish_command',
                    }
                )
                triggered.append(theme)

            forced_label = " (강제)" if forced else ""
            themes_text = ', '.join(_escape_html(t) for t in triggered)
            await update.message.reply_text(
                f"🚀 <b>발행 트리거됨{forced_label}</b>\n\n"
                f"테마: {themes_text}\n"
                f"CE 에이전트가 에세이를 작성합니다.\n"
                f"완료 시 {os.getenv('SITE_ARCHIVE_PATH', 'website/archive')}/ 에 파일로 저장됩니다.\n"
                f"(도메인 연결 후 {os.getenv('SITE_BASE_URL', 'https://woohwahae.kr')}/archive/ 에 노출)",
                parse_mode=constants.ParseMode.HTML
            )

        except Exception as e:
            logger.error("publish_command error: %s", e)
            await update.message.reply_text(f"발행 오류: {_escape_html(str(e))}", parse_mode=constants.ParseMode.HTML)

    @admin_only
    async def signal_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """/signal [텍스트] — 신호 직접 투입"""
        text = ' '.join(context.args).strip() if context.args else ''
        if not text:
            await update.message.reply_text(
                "사용법: <code>/signal 투입할 내용</code>",
                parse_mode=constants.ParseMode.HTML
            )
            return
        try:
            self._save_insight(text, update.effective_user)
            preview = _escape_html(text[:100])
            await update.message.reply_text(
                f"📥 <b>신호 투입 완료</b>\n\n\"{preview}\"\n\nSA 분석 큐 전달됨.",
                parse_mode=constants.ParseMode.HTML
            )
        except Exception as e:
            logger.error("signal_command error: %s", e)
            await update.message.reply_text(f"신호 투입 오류: {_escape_html(str(e))}", parse_mode=constants.ParseMode.HTML)

    @admin_only
    async def report_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """/report — 오늘 처리 요약"""
        try:
            from datetime import date as _date
            today = _date.today().isoformat()
            signals_dir = PROJECT_ROOT / 'knowledge' / 'signals'
            queue_completed = PROJECT_ROOT / '.infra' / 'queue' / 'tasks' / 'completed'

            today_sigs = []
            for f in signals_dir.glob('*.json'):
                try:
                    d = json.loads(f.read_text())
                    if d.get('captured_at', '').startswith(today):
                        today_sigs.append(d)
                except Exception:
                    pass

            today_sa = 0
            scores = []
            if queue_completed.exists():
                for f in queue_completed.glob('*.json'):
                    try:
                        d = json.loads(f.read_text())
                        if d.get('agent_type') == 'SA' and (d.get('completed_at') or '').startswith(today):
                            today_sa += 1
                            score = d.get('result', {}).get('result', {}).get('strategic_score', 0)
                            if score:
                                scores.append(score)
                    except Exception:
                        pass
            avg_score = int(sum(scores) / len(scores)) if scores else 0

            corpus_index = PROJECT_ROOT / 'knowledge' / 'corpus' / 'index.json'
            clusters_total = published_count = 0
            if corpus_index.exists():
                ci = json.loads(corpus_index.read_text())
                clusters_total = len(ci.get('clusters', {}))
                published_count = len(ci.get('published', []))

            sig_types: Dict = {}
            for s in today_sigs:
                t = s.get('type', 'unknown')
                sig_types[t] = sig_types.get(t, 0) + 1
            sig_type_text = ', '.join(f"{t} {n}개" for t, n in sig_types.items()) or '없음'

            msg = (
                f"<b>📋 일일 리포트 — {today}</b>\n\n"
                f"<b>오늘 수집</b>\n"
                f"신규 신호: {len(today_sigs)}개 ({sig_type_text})\n\n"
                f"<b>오늘 분석</b>\n"
                f"SA 완료: {today_sa}건"
                + (f" | 평균 점수: {avg_score}" if avg_score else "") + "\n\n"
                f"<b>Corpus 누적</b>\n"
                f"군집: {clusters_total}개 | 발행: {published_count}개\n\n"
                f"/publish — 즉시 발행  |  /status — 상세 현황"
            )
            await update.message.reply_text(msg, parse_mode=constants.ParseMode.HTML)

        except Exception as e:
            logger.error("report_command error: %s", e)
            await update.message.reply_text(f"리포트 오류: {_escape_html(str(e))}", parse_mode=constants.ParseMode.HTML)

    @admin_only
    async def approve_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Gardener 제안 승인 — /approve [id]"""
        if not self.gardener:
            await update.message.reply_text("Gardener가 비활성 상태입니다.")
            return

        pending = self.gardener.pending
        if not pending:
            await update.message.reply_text("대기 중인 제안이 없습니다.")
            return

        # ID 지정 없으면 첫 번째 제안
        args = context.args
        proposal_id = args[0] if args else pending[0]['id']

        success, msg = self.gardener.approve_proposal(proposal_id)
        await update.message.reply_text(msg, parse_mode=constants.ParseMode.HTML)

    @admin_only
    async def reject_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Gardener 제안 거절 — /reject [id]"""
        if not self.gardener:
            await update.message.reply_text("Gardener가 비활성 상태입니다.")
            return

        pending = self.gardener.pending
        if not pending:
            await update.message.reply_text("대기 중인 제안이 없습니다.")
            return

        args = context.args
        proposal_id = args[0] if args else pending[0]['id']
        proposal = next((p for p in pending if p['id'] == proposal_id), None)
        label = f"{proposal['target_file']} — {proposal['reason']}" if proposal else proposal_id

        self.gardener.reject_proposal(proposal_id)
        await update.message.reply_text(f"❌ 거절됨: {label}")

    @admin_only
    async def pending_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """대기 중인 Gardener 제안 목록 — /pending"""
        if not self.gardener or not self.gardener.pending:
            await update.message.reply_text("대기 중인 제안이 없습니다.")
            return

        lines = ["<b>📋 대기 중인 제안</b>", ""]
        for p in self.gardener.pending:
            lines.append(
                f"<code>{p['id']}</code>\n"
                f"파일: {p['target_file']}\n"
                f"이유: {p['reason']}\n"
                f"내용: {p['proposed_addition'][:80]}...\n"
            )
        lines.append("/approve [id] 또는 /reject [id]")
        await update.message.reply_text(
            "\n".join(lines), parse_mode=constants.ParseMode.HTML
        )

    @admin_only
    async def draft_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """/draft [테마] — 에세이 초안 생성 후 승인 대기"""
        theme_arg = ' '.join(context.args).strip() if context.args else None
        if not theme_arg:
            await update.message.reply_text(
                "사용법: <code>/draft 테마명</code>\n예: /draft 슬로우라이프",
                parse_mode=constants.ParseMode.HTML
            )
            return

        status_msg = await update.message.reply_text(
            f"✍️ <b>{_escape_html(theme_arg)}</b> 초안 작성 중...",
            parse_mode=constants.ParseMode.HTML
        )
        try:
            # 지식 베이스에서 테마 관련 신호 검색
            knowledge = self.engine._search_knowledge(theme_arg)
            draft_prompt = (
                f"woohwahae 브랜드 아카이브 에세이 초안을 써줘.\n"
                f"테마: {theme_arg}\n"
                f"관련 자료:\n{knowledge[:2000] if knowledge else '없음'}\n\n"
                f"조건: 800~1200자, 한국어, 명사형 제목, Magazine B 스타일, "
                f"슬로우라이프 철학 반영. 제목과 본문만 출력."
            )
            import requests as _req
            _api_key = os.getenv('GOOGLE_API_KEY') or os.getenv('GEMINI_API_KEY')
            if not _api_key:
                raise RuntimeError("GOOGLE_API_KEY 환경변수 없음")
            _url = (
                "https://generativelanguage.googleapis.com/v1beta/models/"
                f"gemini-2.5-flash:generateContent?key={_api_key}"
            )
            _resp = _req.post(
                _url,
                json={"contents": [{"parts": [{"text": draft_prompt}]}]},
                timeout=90
            )
            _resp.raise_for_status()
            draft_text = _resp.json()["candidates"][0]["content"]["parts"][0]["text"]

            if not draft_text:
                await status_msg.edit_text("초안 생성 실패.")
                return

            # 초안 임시 저장
            draft_id = f"draft_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
            draft_path = PROJECT_ROOT / '.infra' / 'drafts'
            draft_path.mkdir(parents=True, exist_ok=True)
            (draft_path / f"{draft_id}.json").write_text(
                json.dumps({'id': draft_id, 'theme': theme_arg, 'content': draft_text,
                            'created_at': datetime.now().isoformat()}, ensure_ascii=False, indent=2)
            )

            # 초안 미리보기 + 승인 버튼
            preview = draft_text[:800] + ("..." if len(draft_text) > 800 else "")
            keyboard = InlineKeyboardMarkup([[
                InlineKeyboardButton("✅ 발행", callback_data=f"draft_publish:{draft_id}"),
                InlineKeyboardButton("❌ 폐기", callback_data=f"draft_discard:{draft_id}"),
            ]])
            await status_msg.edit_text(
                f"<b>📝 초안 — {_escape_html(theme_arg)}</b>\n\n"
                f"{_escape_html(preview)}",
                parse_mode=constants.ParseMode.HTML,
                reply_markup=keyboard
            )
        except Exception as e:
            logger.error("draft_command error: %s", e)
            await status_msg.edit_text(f"초안 오류: {_escape_html(str(e))}", parse_mode=constants.ParseMode.HTML)

    @admin_only
    async def corpus_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """/corpus — 군집 현황 미리보기"""
        try:
            corpus_index = PROJECT_ROOT / 'knowledge' / 'corpus' / 'index.json'
            if not corpus_index.exists():
                await update.message.reply_text("Corpus가 아직 초기화되지 않았습니다.")
                return

            ci = json.loads(corpus_index.read_text())
            clusters = ci.get('clusters', {})
            published = ci.get('published', [])

            if not clusters:
                await update.message.reply_text(
                    "아직 군집 없음.\nSA 분석이 완료되면 자동으로 군집이 형성됩니다."
                )
                return

            lines = [f"<b>🗂 Corpus 군집 현황</b> (발행됨: {len(published)}개)\n"]
            for theme, c in sorted(clusters.items(), key=lambda x: len(x[1].get('entry_ids', [])), reverse=True):
                cnt = len(c.get('entry_ids', []))
                last = c.get('last_seen', '')[:10]
                # 성숙도 판단 (5개+ = 발행 가능)
                ready = "🟢 발행가능" if cnt >= 5 else f"🟡 축적중 ({5-cnt}개 더 필요)"
                lines.append(f"<b>{_escape_html(theme)}</b>\n  {cnt}개 신호 | {last} | {ready}\n")

            lines.append("/publish [테마] 로 즉시 발행")
            await update.message.reply_text('\n'.join(lines), parse_mode=constants.ParseMode.HTML)

        except Exception as e:
            logger.error("corpus_command error: %s", e)
            await update.message.reply_text(f"오류: {_escape_html(str(e))}", parse_mode=constants.ParseMode.HTML)

    async def handle_draft_callback(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """초안 승인/폐기 인라인 버튼 처리"""
        query = update.callback_query
        await query.answer()
        data = query.data  # "draft_publish:draft_id" or "draft_discard:draft_id"

        try:
            action, draft_id = data.split(':', 1)
            draft_path = PROJECT_ROOT / '.infra' / 'drafts' / f"{draft_id}.json"

            if not draft_path.exists():
                await query.edit_message_text("초안을 찾을 수 없습니다.")
                return

            draft = json.loads(draft_path.read_text())

            if action == 'draft_discard':
                draft_path.unlink()
                await query.edit_message_text(f"❌ 폐기됨: {draft['theme']}")
                return

            # 발행: CE 태스크로 전달
            from core.system.queue_manager import QueueManager
            qm = QueueManager()
            qm.create_task(
                agent_type='CE',
                task_type='write_corpus_essay',
                payload={
                    'draft_content': draft['content'],
                    'theme': draft['theme'],
                    'forced': True,
                    'triggered_by': 'telegram_draft_approve',
                }
            )
            draft_path.unlink()
            await query.edit_message_text(
                f"🚀 <b>발행 승인됨</b>\n\n"
                f"테마: {_escape_html(draft['theme'])}\n"
                f"CE 에이전트가 최종 포맷으로 처리합니다.",
                parse_mode=constants.ParseMode.HTML
            )
        except Exception as e:
            logger.error("draft callback error: %s", e)
            await query.edit_message_text(f"처리 오류: {str(e)[:100]}")

    async def send_daily_briefing(self, app):
        """매일 08:00 자동 브리핑 푸시"""
        admin_id = os.getenv('ADMIN_TELEGRAM_ID')
        if not admin_id:
            return
        try:
            from datetime import date as _date
            today = _date.today().isoformat()
            signals_dir = PROJECT_ROOT / 'knowledge' / 'signals'
            corpus_index = PROJECT_ROOT / 'knowledge' / 'corpus' / 'index.json'

            today_sigs = sum(
                1 for f in signals_dir.glob('*.json')
                if json.loads(f.read_text()).get('captured_at', '').startswith(today)
            )
            clusters_total = published = 0
            if corpus_index.exists():
                ci = json.loads(corpus_index.read_text())
                clusters_total = len(ci.get('clusters', {}))
                published = len(ci.get('published', []))
                ripe = sum(
                    1 for c in ci.get('clusters', {}).values()
                    if len(c.get('entry_ids', [])) >= 5
                )
            else:
                ripe = 0

            msg = (
                f"☀️ <b>일일 브리핑 — {today}</b>\n\n"
                f"어젯밤 수집: {today_sigs}개 신호\n"
                f"Corpus 군집: {clusters_total}개 (발행가능 {ripe}개)\n"
                f"누적 발행: {published}개\n\n"
                + (f"💡 <b>{ripe}개 군집이 발행 준비 완료</b>\n/publish 로 발행하세요." if ripe > 0
                   else "Gardener가 03:00에 군집을 점검합니다.")
            )
            await app.bot.send_message(chat_id=int(admin_id), text=msg, parse_mode=constants.ParseMode.HTML)
        except Exception as e:
            logger.error("daily briefing error: %s", e)

    async def notify_publish_complete(self, theme: str, url: str = ""):
        """발행 완료 시 텔레그램 알림 (CE agent에서 호출)"""
        admin_id = os.getenv('ADMIN_TELEGRAM_ID')
        if not admin_id or not hasattr(self, '_app'):
            return
        try:
            link_text = f"\n🔗 {url}" if url else ""
            msg = (
                f"✅ <b>발행 완료</b>\n\n"
                f"테마: {_escape_html(theme)}{link_text}\n"
                f"website/archive/ 에 파일 저장됨\n(도메인 연결 후 웹에서 확인 가능)"
            )
            await self._app.bot.send_message(
                chat_id=int(admin_id), text=msg, parse_mode=constants.ParseMode.HTML
            )
        except Exception as e:
            logger.error("notify_publish_complete error: %s", e)

    def run(self):
        from telegram.ext import JobQueue
        application = Application.builder().token(self.bot_token).build()
        self._app = application

        application.add_handler(CommandHandler("start", self.start_command))
        application.add_handler(CommandHandler("growth", self.growth_command))
        application.add_handler(CommandHandler("status", self.status_command))
        application.add_handler(CommandHandler("publish", self.publish_command))
        application.add_handler(CommandHandler("signal", self.signal_command))
        application.add_handler(CommandHandler("report", self.report_command))
        application.add_handler(CommandHandler("draft", self.draft_command))
        application.add_handler(CommandHandler("corpus", self.corpus_command))
        application.add_handler(CommandHandler("approve", self.approve_command))
        application.add_handler(CommandHandler("reject", self.reject_command))
        application.add_handler(CommandHandler("pending", self.pending_command))
        application.add_handler(CommandHandler("client", self.client_command))
        application.add_handler(CommandHandler("visit", self.visit_command))
        application.add_handler(CallbackQueryHandler(self.handle_draft_callback, pattern=r'^draft_'))
        application.add_handler(MessageHandler(filters.TEXT | filters.PHOTO | filters.Document.ALL, self.handle_message))

        # 매일 08:00 브리핑 자동 푸시
        from datetime import time as _time
        job_queue = application.job_queue
        if job_queue:
            job_queue.run_daily(
                lambda ctx: asyncio.create_task(self.send_daily_briefing(application)),
                time=_time(hour=8, minute=0, second=0),
                name="daily_briefing"
            )

        logger.info("🚀 V6 Secretary Service Started")
        application.run_polling()


if __name__ == "__main__":
    from core.system.env_validator import validate_env
    validate_env("telegram_secretary")

    token = os.getenv('TELEGRAM_BOT_TOKEN')
    bot = TelegramSecretaryV6(token)
    bot.run()
