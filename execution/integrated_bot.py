#!/usr/bin/env python3
"""
97LAYER OS Integrated Bot
Complete automation with Claude + Gemini dual AI
Implements Cycle Protocol with multimodal support
"""

import os
import sys
import json
import time
import logging
import base64
from pathlib import Path
from datetime import datetime
from typing import Optional, Dict, Any, List

# Add project root to path
PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.append(str(PROJECT_ROOT))

# Import our engines
from libs.ai_engine import AIEngine
from libs.claude_engine import DualAIEngine
from libs.memory_manager import MemoryManager

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(PROJECT_ROOT / 'logs' / 'integrated_bot.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

# Telegram imports
try:
    import telegram
    from telegram import Update, Bot
    from telegram.ext import Application, CommandHandler, MessageHandler, filters, ContextTypes
except ImportError:
    logger.error("telegram package not installed. Installing...")
    os.system("pip3 install --quiet python-telegram-bot")
    import telegram
    from telegram import Update, Bot
    from telegram.ext import Application, CommandHandler, MessageHandler, filters, ContextTypes


class IntegratedBot:
    """Complete 97LAYER OS Bot with all features"""

    def __init__(self):
        # Load config
        from libs.core_config import TELEGRAM_CONFIG
        self.token = TELEGRAM_CONFIG["BOT_TOKEN"]

        # Initialize engines
        logger.info("Initializing AI engines...")
        self.dual_ai = DualAIEngine()
        self.memory = MemoryManager()

        # Cycle protocol stages
        self.stages = {
            "capture": self.stage_capture,
            "connect": self.stage_connect,
            "meaning": self.stage_meaning,
            "manifest": self.stage_manifest,
            "cycle": self.stage_cycle
        }

        # Command handlers
        self.commands = {
            "status": self.cmd_status,
            "claude": self.cmd_claude,
            "gemini": self.cmd_gemini,
            "cycle": self.cmd_cycle,
            "help": self.cmd_help
        }

        # Statistics
        self.stats = {
            "messages_processed": 0,
            "images_analyzed": 0,
            "claude_calls": 0,
            "gemini_calls": 0,
            "start_time": datetime.now()
        }

    async def start(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /start command"""
        welcome_text = """
🦋 97LAYER OS - Operational

I am your second brain, implementing the Cycle Protocol:
1. Capture - Record all signals
2. Connect - Find patterns
3. Meaning - Generate insights
4. Manifest - Create content
5. Cycle - Continuous improvement

Commands:
/status - System status
/claude [text] - Sovereign judgment (Claude Opus)
/gemini [text] - Quick analysis (Gemini)
/cycle - Show cycle status
/help - Show this message

Send me text or images to capture as raw signals.
"""
        await update.message.reply_text(welcome_text)

    async def cmd_status(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Show system status"""
        uptime = datetime.now() - self.stats["start_time"]
        hours = int(uptime.total_seconds() // 3600)
        minutes = int((uptime.total_seconds() % 3600) // 60)

        # Get AI status
        ai_status = self.dual_ai.get_status()

        status_text = f"""
📊 System Status

Uptime: {hours}h {minutes}m
Messages: {self.stats['messages_processed']}
Images: {self.stats['images_analyzed']}

AI Engines:
• Claude calls: {self.stats['claude_calls']}/20 monthly
• Gemini calls: {self.stats['gemini_calls']}

{ai_status}

Memory:
• Raw signals: {len(list(Path('knowledge/raw_signals').glob('*.md'))) if Path('knowledge/raw_signals').exists() else 0}
• Drafts: {len(list(Path('knowledge/assets/draft').glob('*.md'))) if Path('knowledge/assets/draft').exists() else 0}
• Published: {len(list(Path('knowledge/assets/published').glob('*.md'))) if Path('knowledge/assets/published').exists() else 0}
"""
        await update.message.reply_text(status_text)

    async def cmd_claude(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Force Claude Opus for sovereign judgment"""
        if not context.args:
            await update.message.reply_text("Usage: /claude [your question]")
            return

        prompt = " ".join(context.args)
        await update.message.reply_text("🤔 Consulting Sovereign (Claude Opus)...")

        try:
            result = self.dual_ai.process(
                prompt=prompt,
                force_claude=True
            )

            self.stats["claude_calls"] += 1

            if isinstance(result["response"], dict):
                # Judgment response
                response = f"""
⚖️ Sovereign Judgment

Approved: {'✅ Yes' if result['response'].get('approved') else '❌ No'}
Score: {result['response'].get('score', 0)}/100

{result['response'].get('reason', '')}
"""
            else:
                # Text response
                response = f"🎯 Sovereign says:\n\n{result['response']}"

            await update.message.reply_text(response)

        except Exception as e:
            logger.error(f"Claude error: {e}")
            await update.message.reply_text(f"Error: {str(e)}")

    async def cmd_gemini(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Force Gemini for quick analysis"""
        if not context.args:
            await update.message.reply_text("Usage: /gemini [your question]")
            return

        prompt = " ".join(context.args)
        await update.message.reply_text("💭 Analyzing with Gemini...")

        try:
            result = self.dual_ai.process(
                prompt=prompt,
                force_gemini=True
            )

            self.stats["gemini_calls"] += 1
            response = f"💡 Analysis:\n\n{result['response']}"
            await update.message.reply_text(response[:4096])  # Telegram limit

        except Exception as e:
            logger.error(f"Gemini error: {e}")
            await update.message.reply_text(f"Error: {str(e)}")

    async def cmd_cycle(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Show cycle protocol status"""
        cycle_text = """
🦋 Cycle Protocol Status

Current implementation:
1. ✅ Capture - All messages saved to raw_signals
2. ⚠️ Connect - Manual pattern finding
3. ⚠️ Meaning - Semi-automated with AI
4. ❌ Manifest - Manual approval needed
5. ❌ Cycle - Not yet automated

Send any message to capture it as a raw signal.
Images are automatically analyzed with Gemini Vision.
"""
        await update.message.reply_text(cycle_text)

    async def cmd_help(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Show help message"""
        await self.start(update, context)

    async def handle_message(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle regular text messages"""
        message = update.message
        text = message.text
        chat_id = message.chat_id
        username = message.from_user.username or message.from_user.first_name

        logger.info(f"Message from {username}: {text[:100]}")
        self.stats["messages_processed"] += 1

        # Stage 1: Capture
        signal_id = await self.stage_capture(text, username)

        # Stage 2: Connect (find patterns)
        connections = await self.stage_connect(signal_id, text)

        # Respond to user
        response = f"✓ Captured signal #{signal_id}\n"

        if connections:
            response += f"📊 Found {len(connections)} connections\n"

        # Simple analysis with Gemini
        if len(text) > 50:  # Only analyze substantial messages
            try:
                analysis = self.dual_ai.process(
                    prompt=f"Brief insight about: {text[:500]}",
                    force_gemini=True
                )
                self.stats["gemini_calls"] += 1
                response += f"\n💭 Insight: {analysis['response'][:200]}..."
            except:
                pass

        await message.reply_text(response)

    async def handle_photo(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle photo messages with multimodal AI"""
        message = update.message
        chat_id = message.chat_id
        username = message.from_user.username or message.from_user.first_name

        logger.info(f"Photo from {username}")
        self.stats["images_analyzed"] += 1

        await message.reply_text("🖼️ Analyzing image...")

        try:
            # Download photo
            photo = message.photo[-1]  # Highest resolution
            file = await photo.get_file()
            photo_bytes = await file.download_as_bytearray()

            # Analyze with Gemini Vision
            caption = message.caption or "Describe this image"
            result = self.dual_ai.process(
                prompt=caption,
                image_data=bytes(photo_bytes)
            )

            # Save to raw_signals with vision analysis
            signal_id = await self.stage_capture(
                f"[IMAGE] {caption}\nAnalysis: {result['response']}",
                username
            )

            response = f"""
🖼️ Image Analysis

Signal ID: #{signal_id}

{result['response'][:1000]}

Image saved to knowledge/raw_signals/
"""
            await message.reply_text(response)

        except Exception as e:
            logger.error(f"Photo processing error: {e}")
            await message.reply_text(f"Error analyzing image: {str(e)}")

    # Cycle Protocol Implementation
    async def stage_capture(self, content: str, source: str) -> str:
        """Stage 1: Capture raw signal"""
        timestamp = datetime.now()
        signal_id = f"rs-{timestamp.strftime('%Y%m%d%H%M%S')}"

        # Save to raw_signals
        signal_dir = PROJECT_ROOT / "knowledge" / "raw_signals"
        signal_dir.mkdir(parents=True, exist_ok=True)

        signal_file = signal_dir / f"{signal_id}_{source}.md"
        with open(signal_file, "w", encoding="utf-8") as f:
            f.write(f"# Raw Signal {signal_id}\n\n")
            f.write(f"**Date**: {timestamp.isoformat()}\n")
            f.write(f"**Source**: {source}\n\n")
            f.write(f"---\n\n{content}\n")

        logger.info(f"Captured signal {signal_id}")
        return signal_id

    async def stage_connect(self, signal_id: str, content: str) -> List[str]:
        """Stage 2: Connect to patterns"""
        # Simple pattern matching for now
        patterns = []

        keywords = ["philosophy", "time", "solitude", "imperfection", "algorithm"]
        for keyword in keywords:
            if keyword in content.lower():
                patterns.append(keyword)

        return patterns

    async def stage_meaning(self, signal_id: str) -> Optional[str]:
        """Stage 3: Generate meaning"""
        # This would use AI to generate essay/content
        # For now, just placeholder
        return None

    async def stage_manifest(self, content: str) -> bool:
        """Stage 4: Manifest with Sovereign approval"""
        # This would use Claude Opus for final judgment
        return False

    async def stage_cycle(self):
        """Stage 5: Complete the cycle"""
        # Archive and prepare for next cycle
        pass

    def run(self):
        """Start the bot"""
        logger.info("Starting Integrated Bot...")

        # Create application
        application = Application.builder().token(self.token).build()

        # Add command handlers
        application.add_handler(CommandHandler("start", self.start))
        application.add_handler(CommandHandler("status", self.cmd_status))
        application.add_handler(CommandHandler("claude", self.cmd_claude))
        application.add_handler(CommandHandler("gemini", self.cmd_gemini))
        application.add_handler(CommandHandler("cycle", self.cmd_cycle))
        application.add_handler(CommandHandler("help", self.cmd_help))

        # Add message handlers
        application.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, self.handle_message))
        application.add_handler(MessageHandler(filters.PHOTO, self.handle_photo))

        # Start bot
        logger.info("Bot started. Polling for updates...")
        application.run_polling()


def main():
    """Main entry point"""
    # Kill any existing bots first
    os.system("pkill -f telegram_daemon.py")
    os.system("pkill -f integrated_bot.py")
    time.sleep(2)

    # Start the integrated bot
    bot = IntegratedBot()
    bot.run()


if __name__ == "__main__":
    main()