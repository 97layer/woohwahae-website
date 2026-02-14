#!/bin/bash
# Filename: auto_fetch_gcp_memory.sh
# Purpose: GCP chat_memory를 자동으로 가져오는 스크립트
# Usage: Cron으로 10분마다 실행 (GCP에서)

set -e

MEMORY_FILE="$HOME/97layerOS/knowledge/chat_memory/7565534667.json"
OUTPUT_FILE="/tmp/chat_memory_export.txt"

# chat_memory를 텍스트 파일로 export
cat "$MEMORY_FILE" > "$OUTPUT_FILE"

# 파일 정보
echo "# Exported at: $(date)" >> "$OUTPUT_FILE"
echo "# Size: $(wc -l < "$MEMORY_FILE") lines" >> "$OUTPUT_FILE"

# 로그
echo "[$(date)] ✅ chat_memory exported to $OUTPUT_FILE" >> /tmp/auto_fetch.log

# Telegram으로 최근 대화 요약 전송 (선택사항)
# python3 ~/97layerOS/libs/notifier.py "GCP chat_memory updated"
