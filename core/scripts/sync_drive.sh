#!/bin/bash
# 97layerOS → Google Drive 동기화 (allowlist)
# knowledge/ + directives/ 만 올림. 나머지 전부 무시.

set -e

LOCAL="/Users/97layer/97layerOS"
REMOTE="gdrive:97layerOS"

echo "📤 knowledge/ 동기화..."
rclone sync "$LOCAL/knowledge" "$REMOTE/knowledge" \
  --exclude "__pycache__/**" \
  --exclude "*.pyc" \
  --progress

echo "📤 directives/ 동기화..."
rclone sync "$LOCAL/directives" "$REMOTE/directives" \
  --exclude "__pycache__/**" \
  --exclude "*.pyc" \
  --progress

echo "✅ 완료"
