#!/bin/bash
# LAYER OS Real-Time Monitoring Dashboard Launcher
# Usage: ./start_monitor.sh [refresh_interval]
# Example: ./start_monitor.sh 3  (refresh every 3 seconds)

REFRESH_INTERVAL=${1:-5}  # Default 5 seconds
export PYTHONPATH="$(cd "$(dirname "$0")" && pwd)"

echo "🖥️  Starting LAYER OS Real-Time Monitor..."
echo "   Refresh Interval: ${REFRESH_INTERVAL} seconds"
echo "   Press Ctrl+C to stop"
echo ""

python3 core/system/monitor_dashboard.py --refresh "$REFRESH_INTERVAL"
