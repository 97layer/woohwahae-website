#!/bin/bash
# 97layerOS MCP-GDrive Server Runner
# Purpose: Run MCP server in Podman container with proper mounts
# Usage: ./run_mcp_server.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

echo "🚀 97layerOS MCP-GDrive Server"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check if container image exists
if ! podman images 97layer-mcp-gdrive | grep -q "97layer-mcp-gdrive"; then
    echo "❌ Container image not found. Run ./build_mcp_container.sh first."
    exit 1
fi

# Check if credentials exist
CREDENTIALS_PATH="$PROJECT_ROOT/credentials/gdrive_auth.json"
if [ ! -f "$CREDENTIALS_PATH" ]; then
    echo "⚠️  Warning: $CREDENTIALS_PATH not found."
    echo "   Google Drive authentication may fail."
    echo ""
fi

# Check if .env exists
if [ ! -f "$PROJECT_ROOT/.env" ]; then
    echo "❌ .env file not found. Create it with GOOGLE_DRIVE_FOLDER_ID."
    exit 1
fi

# Set TMPDIR to avoid macOS sandbox issues
export TMPDIR=/tmp

echo "📂 Credentials: $CREDENTIALS_PATH"
echo "🔧 Environment: $PROJECT_ROOT/.env"
echo ""
echo "🐳 Starting MCP server in container..."
echo "   (Press Ctrl+C to stop)"
echo ""

# Run container with:
# - Interactive mode (-i) for stdio communication
# - Remove on exit (--rm)
# - Environment from .env file
# - Read-only mount of credentials
podman run -i --rm \
    --env-file "$PROJECT_ROOT/.env" \
    -v "$PROJECT_ROOT/credentials:/app/credentials:ro" \
    97layer-mcp-gdrive:latest

echo ""
echo "✅ MCP server stopped."
