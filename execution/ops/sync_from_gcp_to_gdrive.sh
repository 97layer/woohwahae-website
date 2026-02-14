#!/bin/bash
# Filename: sync_from_gcp_to_gdrive.sh
# Purpose: GCP에서 Google Drive로 자동 동기화
# Usage: Run on GCP server via cron every 5 minutes

set -e

PROJECT_DIR="$HOME/97layerOS"
GDRIVE_REMOTE="gdrive:97layerOS"

cd "$PROJECT_DIR"

echo "[$(date '+%Y-%m-%d %H:%M:%S')] 🔄 GCP → Google Drive 동기화 시작..."

# Sync key directories and files
rclone sync "$PROJECT_DIR/knowledge/" "$GDRIVE_REMOTE/knowledge/" \
    --exclude ".DS_Store" \
    --exclude "*.pyc" \
    --exclude "__pycache__/" \
    --quiet

rclone sync "$PROJECT_DIR/directives/" "$GDRIVE_REMOTE/directives/" \
    --exclude ".DS_Store" \
    --quiet

rclone sync "$PROJECT_DIR/execution/" "$GDRIVE_REMOTE/execution/" \
    --exclude ".DS_Store" \
    --exclude "*.pyc" \
    --exclude "__pycache__/" \
    --quiet

rclone sync "$PROJECT_DIR/libs/" "$GDRIVE_REMOTE/libs/" \
    --exclude ".DS_Store" \
    --exclude "*.pyc" \
    --exclude "__pycache__/" \
    --quiet

# Sync critical files
rclone copy "$PROJECT_DIR/task_status.json" "$GDRIVE_REMOTE/" --quiet
rclone copy "$PROJECT_DIR/CLAUDE.md" "$GDRIVE_REMOTE/" --quiet
rclone copy "$PROJECT_DIR/AGENTS.md" "$GDRIVE_REMOTE/" --quiet
rclone copy "$PROJECT_DIR/GEMINI.md" "$GDRIVE_REMOTE/" --quiet

echo "[$(date '+%Y-%m-%d %H:%M:%S')] ✅ 동기화 완료"
