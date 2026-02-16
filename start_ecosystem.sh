#!/bin/bash
# Start 97layerOS Organic Ecosystem — THE CYCLE 완전 자동화
# Runs heartbeat + signal_router + scheduler + SA + AD + CE in parallel

cd "$(dirname "$0")"
export PYTHONPATH="$(pwd)"

# Load .env if exists
if [ -f .env ]; then
    set -a
    source .env
    set +a
fi

echo "=========================================="
echo "97layerOS Organic Ecosystem (THE CYCLE)"
echo "heartbeat + signal_router + scheduler"
echo "+ SA + AD + CE agents"
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

# 4. SA Agent (Strategy Analyst — 신호 분석)
echo "🔍 Starting SA Agent (Strategy Analyst)..."
python3 core/agents/sa_agent.py &
SA_PID=$!
echo "   PID: $SA_PID"

# 5. AD Agent (Art Director — 비주얼 컨셉)
echo "🎨 Starting AD Agent (Art Director)..."
python3 core/agents/ad_agent.py &
AD_PID=$!
echo "   PID: $AD_PID"

# 6. CE Agent (Chief Editor — 콘텐츠 작성)
echo "✍️  Starting CE Agent (Chief Editor)..."
python3 core/agents/ce_agent.py &
CE_PID=$!
echo "   PID: $CE_PID"

echo ""
echo "✅ THE CYCLE 에코시스템 시작 완료"
echo "   Heartbeat:      PID $HEARTBEAT_PID (30s interval)"
echo "   Signal Router:  PID $ROUTER_PID   (10s polling)"
echo "   Scheduler:      PID $SCHEDULER_PID (09:00 / 21:00)"
echo "   SA Agent:       PID $SA_PID        (5s polling)"
echo "   AD Agent:       PID $AD_PID        (5s polling)"
echo "   CE Agent:       PID $CE_PID        (5s polling)"
echo ""
echo "   THE CYCLE: 텔레그램 입력 → 신호 저장 → 라우팅 → 큐"
echo "              → 에이전트 처리 → 텔레그램 알림 → 반복"
echo ""
echo "   To stop all: kill $HEARTBEAT_PID $ROUTER_PID $SCHEDULER_PID $SA_PID $AD_PID $CE_PID"
echo "   Press Ctrl+C to stop all"
echo ""

# Trap Ctrl+C → kill all children
trap "echo ''; echo 'Stopping ecosystem...'; kill $HEARTBEAT_PID $ROUTER_PID $SCHEDULER_PID $SA_PID $AD_PID $CE_PID 2>/dev/null; exit 0" INT TERM

wait
