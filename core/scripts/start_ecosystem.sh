#!/bin/bash
# LAYER OS Ecosystem — SA + AD + CE + CD + Orchestrator
# systemd 서비스(97layer-ecosystem)에서 호출. SIGTERM 수신 시 자식 프로세스 정리.

VENV="/home/skyto5339_gmail_com/97layerOS/.venv/bin/python3"
LOGS=".infra/logs"

# 로컬 실행 시 venv 경로 fallback
if [ ! -f "$VENV" ]; then
    VENV="python3"
fi

mkdir -p "$LOGS"

# SIGTERM / SIGINT → 자식 프로세스 전체 종료 후 exit
cleanup() {
    kill "$GUARD_PID" "$SA_PID" "$AD_PID" "$CE_PID" "$CD_PID" "$ORCH_PID" 2>/dev/null
    wait "$GUARD_PID" "$SA_PID" "$AD_PID" "$CE_PID" "$CD_PID" "$ORCH_PID" 2>/dev/null
    exit 0
}
trap cleanup SIGTERM SIGINT

"$VENV" -u core/system/filesystem_guard.py   >> "$LOGS/filesystem_guard.log" 2>&1 & GUARD_PID=$!
"$VENV" -u core/agents/sa_agent.py           >> "$LOGS/sa_agent.log"     2>&1 & SA_PID=$!
"$VENV" -u core/agents/ad_agent.py           >> "$LOGS/ad_agent.log"     2>&1 & AD_PID=$!
"$VENV" -u core/agents/ce_agent.py           >> "$LOGS/ce_agent.log"     2>&1 & CE_PID=$!
"$VENV" -u core/agents/cd_agent.py           >> "$LOGS/cd_agent.log"     2>&1 & CD_PID=$!
"$VENV" -u core/system/pipeline_orchestrator.py >> "$LOGS/orchestrator.log" 2>&1 & ORCH_PID=$!

wait
