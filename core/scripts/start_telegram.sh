#!/bin/bash
# Start LAYER OS Telegram Executive Secretary
# v6 (JARVIS Plus Edition): NotebookLM Deep RAG + Multi-Agent + Premium UX

cd /Users/97layer/97layerOS
export PYTHONPATH=/Users/97layer/97layerOS

echo "=========================================="
echo "LAYER OS Telegram Executive Secretary"
echo "v6: NotebookLM RAG + Multi-Agent + Premium UX"
echo "=========================================="
echo ""

# Check environment
if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    echo "❌ TELEGRAM_BOT_TOKEN not set"
    echo ""
    echo "Please set it in .env file or export:"
    echo "  export TELEGRAM_BOT_TOKEN='your_token_here'"
    echo ""
    exit 1
fi

# Check if bot is already running
if [ -f /tmp/telegram_bot.pid ]; then
    PID=$(cat /tmp/telegram_bot.pid)
    if ps -p $PID > /dev/null 2>&1; then
        echo "⚠️  Bot is already running with PID: $PID"
        echo "   To stop it, run: kill $PID"
        echo ""
        exit 0
    else
        # Clean up stale PID file
        rm /tmp/telegram_bot.pid
    fi
fi

# Display features
echo "🤖 Features:"
echo "   • 10 Commands: /start, /status, /report, /analyze, /signal"
echo "                  /morning, /evening, /search, /memo, /sync"
echo "   • Auto Signal Capture: text + images + links"
echo "   • Multi-Agent Analysis: SA+AD → CE → CD"
echo "   • Ralph Loop Quality Control"
echo "   • Daily Automation: morning briefing + evening report"
echo "   • Google Drive Sync (if configured)"
echo ""

# Optional: Start monitor in background
if command -v tmux &> /dev/null; then
    echo "💡 Tip: Run './start_monitor.sh' in another terminal to see real-time status"
    echo ""
fi

# Start bot
echo "🚀 Starting Telegram Executive Secretary..."
echo "   Press Ctrl+C to stop"
echo ""

python3 core/daemons/telegram_secretary.py &
BOT_PID=$!
echo $BOT_PID > /tmp/telegram_bot.pid

echo "✅ Bot started with PID: $BOT_PID"
echo "   Logs: .infra/logs/telegram_secretary.log"
echo ""

# Wait for bot
wait $BOT_PID
