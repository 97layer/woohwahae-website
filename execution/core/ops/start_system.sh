#!/bin/bash
# 97LAYER OS 자동 시작 스크립트

PROJECT_DIR="/Users/97layer/97layerOS"
LOG_DIR="$PROJECT_DIR/logs"

mkdir -p $LOG_DIR

# 기존 프로세스 정리
pkill -f "telegram_daemon.py" 2>/dev/null
pkill -f "async_telegram_daemon.py" 2>/dev/null
pkill -f "master_controller.py" 2>/dev/null
pkill -f "cycle_manager.py" 2>/dev/null

sleep 3

echo "🚀 97LAYER OS 자동화 시스템 시작..."
cd $PROJECT_DIR

# 1. Master Controller 시작
nohup python3 execution/ops/master_controller.py start > $LOG_DIR/master.log 2>&1 &
echo "✅ Master Controller 시작됨"

sleep 5

# 2. Cycle Manager 시작  
nohup python3 execution/cycle_manager.py > $LOG_DIR/cycle.log 2>&1 &
echo "✅ Cycle Manager 시작됨"

# 3. Async Telegram Daemon 시작
nohup python3 execution/async_telegram_daemon.py > $LOG_DIR/async_telegram.log 2>&1 &
echo "✅ Async Telegram Bot 시작됨"

echo "✨ 97LAYER OS 완전 자율 시스템 가동 완료!"
