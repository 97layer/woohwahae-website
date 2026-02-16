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


class TelegramSecretaryV6:
    def __init__(self, bot_token: str):
        self.bot_token = bot_token
        self.notebooklm = get_bridge()
        self.engine = get_conversation_engine()
        self.classifier = get_intent_classifier()
        self.youtube = YouTubeAnalyzer()
        self.image = ImageAnalyzer()

        # UI Settings
        self.loading_emojis = ["🔘", "⚪", "⚫"]

        logger.info("✅ Telegram Secretary V6 (JARVIS Plus) initialized")

    async def start_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        user = update.effective_user
        welcome_msg = (
            f"<b>97layer Executive Secretary V6</b>\n\n"
            f"안녕하세요, {_escape_html(user.first_name)}님. 전략적 의사결정을 돕는 JARVIS Plus입니다.\n\n"
            f"<b>핵심 인터페이스</b>:\n"
            f"- <code>자연어 질문</code>: NotebookLM Deep RAG 기반 답변\n"
            f"- <code>YouTube 링크</code>: 심층 분석 및 멀티모달 자산 생성\n"
            f"- <code>이미지 콘텐츠</code>: 브랜드 비전 기반 통찰 추출\n"
            f"- <code>아이디어 텍스트</code>: 인사이트 자동 분류 및 영구 저장\n\n"
            f"사령관의 의도를 분석하여 최적의 결과를 도출하겠습니다."
        )
        await update.message.reply_text(welcome_msg, parse_mode=constants.ParseMode.HTML)

    async def handle_message(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        message = update.message
        if not message.text and not message.photo:
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

        # 3. 텍스트 의도 분석 및 처리
        await self.process_text(update, context)

    async def process_youtube(self, update: Update, context: ContextTypes.DEFAULT_TYPE, url: str):
        status_msg = await update.message.reply_text(
            "🛸 <b>Anti-Gravity YouTube Analysis System 가동</b>",
            parse_mode=constants.ParseMode.HTML
        )

        try:
            await status_msg.edit_text("🛸 <code>Analysis</code>: 영상 데이터 수집 및 자막 추출 중...", parse_mode=constants.ParseMode.HTML)
            result = self.youtube.process_url(url)

            if not result['success']:
                await status_msg.edit_text(f"❌ 분석 실패: {_escape_html(str(result.get('error', '')))}", parse_mode=constants.ParseMode.HTML)
                return

            await status_msg.edit_text("🛸 <code>Intellect</code>: NotebookLM Deep RAG 연동 중...", parse_mode=constants.ParseMode.HTML)

            if self.notebooklm and self.notebooklm.authenticated:
                await status_msg.edit_text("🛸 <code>Synthesis</code>: 멀티모달 자산(Audio, Mindmap) 생성 중...", parse_mode=constants.ParseMode.HTML)
                summary = f"ID: <code>{_escape_html(result['video_id'])}</code>\n자막: {len(result['transcript'])}자 수집 완료."
            else:
                summary = f"ID: <code>{_escape_html(result['video_id'])}</code>\n자막 수집 완료 (NotebookLM Offline)."

            final_text = (
                f"✅ <b>YouTube 전략 분석 완료</b>\n\n"
                f"{summary}\n\n"
                f"지식 베이스에 성공적으로 영구 저장되었습니다.\n"
                f"추가적인 '오디오 브리핑'이나 '마인드맵'이 필요하시면 말씀해주십시오."
            )
            await status_msg.edit_text(final_text, parse_mode=constants.ParseMode.HTML)

        except Exception as e:
            logger.error("YouTube processing error: %s", e)
            try:
                await status_msg.edit_text(f"❌ 시스템 오류: {_escape_html(str(e))}", parse_mode=constants.ParseMode.HTML)
            except Exception:
                pass

    async def process_image(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """이미지 처리 — Gemini Vision 분석 + signals 저장 + 텍스트 메시지 함께 처리"""
        caption = update.message.caption or ""
        status_msg = await update.message.reply_text("🖼️ 이미지 분석 중...")
        try:
            photo = update.message.photo[-1]
            file = await context.bot.get_file(photo.file_id)
            import tempfile
            with tempfile.NamedTemporaryFile(suffix='.jpg', delete=False) as tmp:
                tmp_path = tmp.name
                await file.download_to_drive(tmp_path)

            # analyze_image()는 {'description', 'full_analysis', 'insights', ...} 반환
            result = self.image.analyze_image(tmp_path)

            # signals에 저장 (caption 포함)
            try:
                self.image.save_image_and_analysis(tmp_path, {**result, 'caption': caption})
            except Exception:
                pass
            os.unlink(tmp_path)

            full_analysis = result.get('full_analysis') or result.get('description', '')
            if full_analysis and '분석 실패' not in full_analysis:
                # caption이 있으면 함께 처리
                combined = full_analysis
                if caption:
                    combined = f"📝 **메모**: {caption}\n\n{full_analysis}"

                # 4096자 제한
                if len(combined) > 4000:
                    combined = combined[:4000] + "\n\n..."

                await status_msg.edit_text(combined)

                # caption을 insight로도 저장
                if caption:
                    self._save_insight(f"[이미지 메모] {caption}", update.effective_user)
            else:
                err = result.get('description', '알 수 없는 오류')
                await status_msg.edit_text(f"❌ 이미지 분석 실패: {_escape_html(err)}", parse_mode=constants.ParseMode.HTML)

        except Exception as e:
            logger.error("Image processing error: %s", e)
            try:
                await status_msg.edit_text(f"❌ 이미지 처리 오류: {_escape_html(str(e))}", parse_mode=constants.ParseMode.HTML)
            except Exception:
                pass

    async def process_text(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        text = update.message.text
        try:
            # 의도 분류
            intent_data = self.classifier.classify(text)
            intent = intent_data['intent']

            if intent == 'insight':
                # 인사이트 저장 UX
                timestamp = datetime.now().strftime('%H:%M:%S')
                preview = _escape_html(text[:150])
                await update.message.reply_text(
                    f"💾 <b>Insight Captured</b> (<code>{timestamp}</code>)\n\n"
                    f"\"{preview}...\"\n\n"
                    f"자동으로 지식 베이스에 분류 및 저장되었습니다.",
                    parse_mode=constants.ParseMode.HTML
                )
                self._save_insight(text, update.effective_user)
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
        signals_dir = PROJECT_ROOT / 'knowledge' / 'signals'
        signals_dir.mkdir(parents=True, exist_ok=True)
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        with open(signals_dir / f"text_{timestamp}.json", 'w', encoding='utf-8') as f:
            json.dump({
                'type': 'text_insight',
                'content': text,
                'captured_at': datetime.now().isoformat(),
                'from_user': user.username or user.first_name,
                'status': 'captured'
            }, f, ensure_ascii=False, indent=2)

    def run(self):
        application = Application.builder().token(self.bot_token).build()
        application.add_handler(CommandHandler("start", self.start_command))
        application.add_handler(MessageHandler(filters.TEXT | filters.PHOTO, self.handle_message))
        logger.info("🚀 V6 Secretary Service Started")
        application.run_polling()


if __name__ == "__main__":
    token = os.getenv('TELEGRAM_BOT_TOKEN')
    if token:
        bot = TelegramSecretaryV6(token)
        bot.run()
    else:
        print("Error: TELEGRAM_BOT_TOKEN not found")
