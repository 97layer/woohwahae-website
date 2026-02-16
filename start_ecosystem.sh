#!/bin/bash
# Start 97layerOS Organic Ecosystem Daemons
# Runs heartbeat + signal_router + daily_routine scheduler in parallel

cd "$(dirname "$0")"
export PYTHONPATH="$(pwd)"

# Load .env if exists
if [ -f .env ]; then
    set -a
    source .env
    set +a
fi

echo "=========================================="
echo "97layerOS Organic Ecosystem"
echo "heartbeat + signal_router + scheduler"
echo "=========================================="
echo ""

# 1. Heartbeat (MacBook ↔ GCP 상태 감지)
echo "💓 Starting Heartbeat daemon..."
python3 core/system/heartbeat.py &
HEARTBEAT_PID=$!
echo "   PID: $HEARTBEAT_PID"

# 2. Signal Router (signals/ → Queue 자동 라우팅)
echo "🔀 Starting Signal Router (watch mode)..."
python3 core/system/signal_router.py --watch &
ROUTER_PID=$!
echo "   PID: $ROUTER_PID"

# 3. Daily Routine Scheduler (09:00 / 21:00)
echo "⏰ Starting Daily Routine Scheduler..."
python3 core/system/daily_routine.py --scheduler &
SCHEDULER_PID=$!
echo "   PID: $SCHEDULER_PID"

echo ""
echo "✅ Ecosystem started"
echo "   Heartbeat:     PID $HEARTBEAT_PID (30s interval)"
echo "   Signal Router: PID $ROUTER_PID   (10s polling)"
echo "   Scheduler:     PID $SCHEDULER_PID (09:00 / 21:00)"
echo ""
echo "   To stop all: kill $HEARTBEAT_PID $ROUTER_PID $SCHEDULER_PID"
echo "   Press Ctrl+C to stop all"
echo ""

# Trap Ctrl+C → kill all children
trap "echo ''; echo 'Stopping ecosystem...'; kill $HEARTBEAT_PID $ROUTER_PID $SCHEDULER_PID 2>/dev/null; exit 0" INT TERM

wait
