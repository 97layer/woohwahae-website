#!/usr/bin/env python3
"""
97LAYER OS Unified System
Instant sharing between Claude and internal Gemini
Real-time integration with automatic execution
"""

import os
import sys
import json
import time
import asyncio
import logging
from pathlib import Path
from datetime import datetime
from typing import Dict, Any, Optional, List
from queue import Queue
from threading import Thread

PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.append(str(PROJECT_ROOT))

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(name)s] %(levelname)s: %(message)s'
)
logger = logging.getLogger("UnifiedSystem")


class UnifiedBridge:
    """Bridge for instant communication between Claude and internal systems"""

    def __init__(self):
        self.shared_memory = {}
        self.message_queue = Queue()
        self.context_buffer = []
        self.api_keys = self._load_api_keys()
        self.engines = self._initialize_engines()

    def _load_api_keys(self) -> Dict[str, str]:
        """Load all API keys from environment"""
        keys = {
            "gemini": os.getenv("GEMINI_API_KEY", "AIzaSyBHpQRFjdZRzzkYGR6eqBezyPteaHX_uMQ"),
            "anthropic": os.getenv("ANTHROPIC_API_KEY", ""),
            "telegram": "8501568801:AAE-3fBl-p6uZcmrdsWSRQuz_eg8yDADwjI"
        }
        logger.info(f"API Keys loaded: Gemini={bool(keys['gemini'])}, Claude={bool(keys['anthropic'] and 'your_' not in keys['anthropic'])}")
        return keys

    def _initialize_engines(self) -> Dict[str, Any]:
        """Initialize all AI engines"""
        engines = {}

        # Initialize Gemini
        try:
            from libs.ai_engine import AIEngine
            engines["gemini"] = AIEngine()
            logger.info("✅ Gemini engine initialized")
        except Exception as e:
            logger.error(f"❌ Gemini initialization failed: {e}")
            engines["gemini"] = None

        # Initialize Claude if key exists
        if self.api_keys["anthropic"] and "your_" not in self.api_keys["anthropic"]:
            try:
                from libs.claude_engine import DualAIEngine
                engines["dual"] = DualAIEngine()
                logger.info("✅ Claude engine initialized")
            except Exception as e:
                logger.error(f"❌ Claude initialization failed: {e}")
                engines["dual"] = None
        else:
            logger.warning("⚠️ Claude API key not configured - using Gemini only mode")
            engines["dual"] = None

        return engines

    def share_context(self, source: str, data: Dict[str, Any]):
        """Share context instantly between systems"""
        timestamp = datetime.now().isoformat()
        context = {
            "timestamp": timestamp,
            "source": source,
            "data": data
        }

        # Add to shared memory
        self.shared_memory[f"{source}_{timestamp}"] = context

        # Add to context buffer for AI engines
        self.context_buffer.append(context)
        if len(self.context_buffer) > 100:
            self.context_buffer.pop(0)

        # Add to message queue for processing
        self.message_queue.put(context)

        logger.info(f"📤 Context shared from {source}: {list(data.keys())}")

    def get_shared_context(self, source: Optional[str] = None, last_n: int = 10) -> List[Dict]:
        """Get shared context from all or specific source"""
        if source:
            contexts = [v for k, v in self.shared_memory.items() if k.startswith(source)]
        else:
            contexts = list(self.shared_memory.values())

        return sorted(contexts, key=lambda x: x['timestamp'], reverse=True)[:last_n]

    async def process_with_ai(self, prompt: str, use_claude: bool = False) -> Dict[str, Any]:
        """Process with appropriate AI engine"""
        # Add context to prompt
        recent_context = self.get_shared_context(last_n=5)
        if recent_context:
            context_str = "\n".join([f"[{c['source']}] {json.dumps(c['data'])}" for c in recent_context])
            full_prompt = f"Context:\n{context_str}\n\nTask: {prompt}"
        else:
            full_prompt = prompt

        result = {"success": False, "response": "", "engine": "none"}

        # Use Claude if requested and available
        if use_claude and self.engines.get("dual"):
            try:
                response = self.engines["dual"].process(
                    prompt=full_prompt,
                    force_claude=True
                )
                result = {
                    "success": True,
                    "response": response["response"],
                    "engine": "claude"
                }
                logger.info("🤖 Processed with Claude")
            except Exception as e:
                logger.error(f"Claude processing error: {e}")

        # Fallback to Gemini
        if not result["success"] and self.engines.get("gemini"):
            try:
                response = self.engines["gemini"].generate_response(full_prompt)
                result = {
                    "success": True,
                    "response": response,
                    "engine": "gemini"
                }
                logger.info("🤖 Processed with Gemini")
            except Exception as e:
                logger.error(f"Gemini processing error: {e}")
                result["response"] = f"Error: {e}"

        # Share AI response back to context
        self.share_context(f"ai_{result['engine']}", {
            "prompt": prompt[:100],
            "response": result["response"][:500],
            "success": result["success"]
        })

        return result


class UnifiedBot:
    """Unified Telegram bot with instant Claude-Gemini integration"""

    def __init__(self, bridge: UnifiedBridge):
        self.bridge = bridge
        self.token = bridge.api_keys["telegram"]
        self.running = False

    async def start(self):
        """Start the unified bot"""
        logger.info("🚀 Starting Unified Bot...")

        import telegram
        from telegram import Update
        from telegram.ext import Application, CommandHandler, MessageHandler, filters, ContextTypes

        # Create application
        self.app = Application.builder().token(self.token).build()

        # Command handlers
        async def cmd_status(update: Update, context: ContextTypes.DEFAULT_TYPE):
            """System status command"""
            status = {
                "gemini": "✅" if self.bridge.engines.get("gemini") else "❌",
                "claude": "✅" if self.bridge.engines.get("dual") else "❌",
                "shared_contexts": len(self.bridge.shared_memory),
                "queue_size": self.bridge.message_queue.qsize()
            }

            text = f"""📊 System Status

AI Engines:
• Gemini: {status['gemini']}
• Claude: {status['claude']}

Shared Memory:
• Contexts: {status['shared_contexts']}
• Queue: {status['queue_size']} messages

Recent Activity:
"""
            recent = self.bridge.get_shared_context(last_n=3)
            for ctx in recent:
                text += f"• [{ctx['source']}] {ctx['timestamp'][-8:]}\n"

            await update.message.reply_text(text)

        async def cmd_claude(update: Update, context: ContextTypes.DEFAULT_TYPE):
            """Force Claude processing"""
            if not context.args:
                await update.message.reply_text("Usage: /claude [prompt]")
                return

            prompt = " ".join(context.args)

            # Share to bridge
            self.bridge.share_context("telegram_command", {
                "user": update.message.from_user.username,
                "command": "claude",
                "prompt": prompt
            })

            # Process with Claude
            await update.message.reply_text("🤔 Processing with Claude...")
            result = await self.bridge.process_with_ai(prompt, use_claude=True)

            if result["success"]:
                await update.message.reply_text(f"🎯 Claude:\n{result['response'][:2000]}")
            else:
                await update.message.reply_text(f"❌ Error: {result['response']}")

        async def cmd_gemini(update: Update, context: ContextTypes.DEFAULT_TYPE):
            """Force Gemini processing"""
            if not context.args:
                await update.message.reply_text("Usage: /gemini [prompt]")
                return

            prompt = " ".join(context.args)

            # Share to bridge
            self.bridge.share_context("telegram_command", {
                "user": update.message.from_user.username,
                "command": "gemini",
                "prompt": prompt
            })

            # Process with Gemini
            await update.message.reply_text("💭 Processing with Gemini...")
            result = await self.bridge.process_with_ai(prompt, use_claude=False)

            if result["success"]:
                await update.message.reply_text(f"💡 Gemini:\n{result['response'][:2000]}")
            else:
                await update.message.reply_text(f"❌ Error: {result['response']}")

        async def handle_message(update: Update, context: ContextTypes.DEFAULT_TYPE):
            """Handle regular messages"""
            text = update.message.text
            user = update.message.from_user.username or update.message.from_user.first_name

            # Share to bridge
            self.bridge.share_context("telegram_message", {
                "user": user,
                "text": text,
                "chat_id": update.message.chat_id
            })

            # Auto-process with available AI
            if len(text) > 20:
                result = await self.bridge.process_with_ai(
                    f"Brief response to: {text}",
                    use_claude=False  # Use Gemini for regular messages
                )

                if result["success"]:
                    response = f"✓ Captured\n💭 {result['response'][:500]}"
                else:
                    response = "✓ Message captured"

                await update.message.reply_text(response)

        async def handle_photo(update: Update, context: ContextTypes.DEFAULT_TYPE):
            """Handle photo messages"""
            user = update.message.from_user.username or update.message.from_user.first_name

            # Share to bridge
            self.bridge.share_context("telegram_photo", {
                "user": user,
                "caption": update.message.caption or "No caption",
                "chat_id": update.message.chat_id
            })

            await update.message.reply_text("📷 Photo captured for analysis")

            # Download and analyze
            try:
                photo = update.message.photo[-1]
                file = await photo.get_file()
                photo_bytes = await file.download_as_bytearray()

                # Process with Gemini Vision
                if self.bridge.engines.get("gemini"):
                    from libs.ai_engine import AIEngine
                    ai = self.bridge.engines["gemini"]
                    analysis = ai.generate_multimodal(
                        update.message.caption or "Describe this image",
                        bytes(photo_bytes)
                    )

                    # Share analysis
                    self.bridge.share_context("vision_analysis", {
                        "analysis": analysis[:500],
                        "user": user
                    })

                    await update.message.reply_text(f"🖼️ Analysis:\n{analysis[:1000]}")

            except Exception as e:
                logger.error(f"Photo processing error: {e}")
                await update.message.reply_text(f"Error analyzing photo: {e}")

        # Register handlers
        self.app.add_handler(CommandHandler("status", cmd_status))
        self.app.add_handler(CommandHandler("claude", cmd_claude))
        self.app.add_handler(CommandHandler("gemini", cmd_gemini))
        self.app.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, handle_message))
        self.app.add_handler(MessageHandler(filters.PHOTO, handle_photo))

        # Start polling
        await self.app.run_polling()


class UnifiedOrchestrator:
    """Main orchestrator for the unified system"""

    def __init__(self):
        self.bridge = UnifiedBridge()
        self.bot = UnifiedBot(self.bridge)
        self.processing_thread = None

    def background_processor(self):
        """Background thread for processing queued messages"""
        logger.info("📡 Background processor started")

        while True:
            try:
                if not self.bridge.message_queue.empty():
                    context = self.bridge.message_queue.get()
                    logger.info(f"Processing queued message from {context['source']}")

                    # Auto-process certain types
                    if context['source'] == 'telegram_message':
                        # Could trigger automatic responses or analysis
                        pass

                time.sleep(1)

            except Exception as e:
                logger.error(f"Background processor error: {e}")
                time.sleep(5)

    def run(self):
        """Run the unified system"""
        logger.info("=" * 60)
        logger.info("97LAYER OS UNIFIED SYSTEM")
        logger.info("=" * 60)

        # Start background processor
        self.processing_thread = Thread(target=self.background_processor, daemon=True)
        self.processing_thread.start()

        # Test connections
        logger.info("\n🔍 Testing connections...")

        # Test Gemini
        if self.bridge.engines.get("gemini"):
            try:
                response = self.bridge.engines["gemini"].generate_response("Say 'OK'")
                logger.info(f"✅ Gemini test: {response}")
                self.bridge.share_context("system_test", {"gemini": "operational"})
            except Exception as e:
                logger.error(f"❌ Gemini test failed: {e}")

        # Test Claude if available
        if self.bridge.engines.get("dual"):
            try:
                from libs.claude_engine import DualAIEngine
                response = self.bridge.engines["dual"].process(
                    "Say 'OK'",
                    force_gemini=True  # Test with Gemini first to save Claude calls
                )
                logger.info(f"✅ Dual engine test: {response['response']}")
                self.bridge.share_context("system_test", {"dual_engine": "operational"})
            except Exception as e:
                logger.error(f"❌ Dual engine test failed: {e}")

        logger.info("\n🚀 Starting Telegram bot...")

        # Run the bot (blocking)
        try:
            asyncio.run(self.bot.start())
        except KeyboardInterrupt:
            logger.info("\n👋 Shutting down...")


def main():
    """Main entry point"""
    # Kill existing bots
    os.system("pkill -f telegram_daemon")
    os.system("pkill -f integrated_bot")
    os.system("launchctl unload ~/Library/LaunchAgents/com.97layer.os.plist 2>/dev/null")
    time.sleep(2)

    # Run unified system
    orchestrator = UnifiedOrchestrator()
    orchestrator.run()


if __name__ == "__main__":
    main()