#!/bin/bash
# 97LAYER OS - 최소 실행 스크립트
cd "$(dirname "$0")"
pkill -f telegram_daemon.py 2>/dev/null
pkill -f mac_realtime_receiver.py 2>/dev/null
sleep 1
python3 execution/telegram_daemon.py > logs/telegram.log 2>&1 &
python3 execution/ops/mac_realtime_receiver.py > logs/sync.log 2>&1 &
echo "✅ Started (PID: $!)"