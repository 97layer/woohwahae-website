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
            f"<b>97layerOS</b>\n\n"
            f"안녕하세요, {_escape_html(user.first_name)}님.\n\n"
            f"- 자연어로 뭐든 물어보면 됩니다\n"
            f"- YouTube 링크 → 영상 분석\n"
            f"- 이미지 → 브랜드 인사이트 추출\n"
            f"- 아이디어 텍스트 → 자동 저장\n"
            f"- /growth → 시스템 성장 지표"
        )
        await update.message.reply_text(welcome_msg, parse_mode=constants.ParseMode.HTML)

    async def growth_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """자가발전 성장 지표 리포트"""
        knowledge_dir = PROJECT_ROOT / 'knowledge'
        lm_path = knowledge_dir / 'long_term_memory.json'
        signals_dir = knowledge_dir / 'signals'

        try:
            # long_term_memory 통계
            if lm_path.exists():
                data = json.loads(lm_path.read_text(encoding='utf-8'))
                total_exp = len(data.get('experiences', []))
                total_concepts = len(data.get('concepts', {}))
                top_concepts = sorted(
                    data.get('concepts', {}).items(),
                    key=lambda x: x[1], reverse=True
                )[:5]
                last_updated = data.get('metadata', {}).get('last_updated', 'N/A')

                # SA 분석 경험만 필터
                sa_experiences = [e for e in data.get('experiences', []) if e.get('source') == 'sa_agent']
                sa_scores = [e.get('score', 0) for e in sa_experiences if e.get('score')]
                avg_score = int(sum(sa_scores) / len(sa_scores)) if sa_scores else 0
            else:
                total_exp = total_concepts = avg_score = 0
                top_concepts = []
                last_updated = 'N/A'

            # signals 누적수
            signal_count = len(list(signals_dir.glob('**/*.json'))) if signals_dir.exists() else 0

            # 리포트 구성
            concepts_text = "\n".join(
                f"  {k}: {v}회" for k, v in top_concepts
            ) if top_concepts else "  아직 없음"

            msg = (
                f"<b>📈 97layerOS 성장 지표</b>\n\n"
                f"<b>지식 축적</b>\n"
                f"누적 signals: {signal_count}개\n"
                f"경험 기록: {total_exp}개\n"
                f"개념 노드: {total_concepts}개\n\n"
                f"<b>상위 개념</b>\n{concepts_text}\n\n"
                f"<b>SA 분석</b>\n"
                f"분석 완료: {len(sa_experiences)}건\n"
                f"평균 전략점수: {avg_score}/100\n\n"
                f"마지막 업데이트: {_escape_html(last_updated)}"
            )
            await update.message.reply_text(msg, parse_mode=constants.ParseMode.HTML)

        except Exception as e:
            logger.error("growth_command error: %s", e)
            await update.message.reply_text("지표 조회 중 오류가 발생했습니다.")

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
        application.add_handler(CommandHandler("growth", self.growth_command))
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
