#!/bin/bash
# Filename: sync_from_gdrive_to_mac.sh
# Purpose: Google Drive에서 Mac으로 자동 동기화 (GCP 변경사항 가져오기)
# Usage: Run on Mac via LaunchAgent every 5 minutes

set -e

PROJECT_DIR="/Users/97layer/97layerOS"
GDRIVE_DIR="$HOME/내 드라이브(skyto5339@gmail.com)/97layerOS"

cd "$PROJECT_DIR"

echo "[$(date '+%Y-%m-%d %H:%M:%S')] 🔄 Google Drive → Mac 동기화 시작..."

# Sync knowledge directory (most important for GCP updates)
rsync -a --delete \
    --exclude=".DS_Store" \
    --exclude="*.pyc" \
    --exclude="__pycache__/" \
    "$GDRIVE_DIR/knowledge/" "$PROJECT_DIR/knowledge/"

# Sync other directories if they exist
if [ -d "$GDRIVE_DIR/directives" ]; then
    rsync -a --delete \
        --exclude=".DS_Store" \
        "$GDRIVE_DIR/directives/" "$PROJECT_DIR/directives/"
fi

if [ -d "$GDRIVE_DIR/execution" ]; then
    rsync -a --delete \
        --exclude=".DS_Store" \
        --exclude="*.pyc" \
        --exclude="__pycache__/" \
        "$GDRIVE_DIR/execution/" "$PROJECT_DIR/execution/"
fi

if [ -d "$GDRIVE_DIR/libs" ]; then
    rsync -a --delete \
        --exclude=".DS_Store" \
        --exclude="*.pyc" \
        --exclude="__pycache__/" \
        "$GDRIVE_DIR/libs/" "$PROJECT_DIR/libs/"
fi

# Sync critical files
if [ -f "$GDRIVE_DIR/task_status.json" ]; then
    cp "$GDRIVE_DIR/task_status.json" "$PROJECT_DIR/"
fi

echo "[$(date '+%Y-%m-%d %H:%M:%S')] ✅ 동기화 완료"
