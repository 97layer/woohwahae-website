#!/bin/bash
# 97layerOS MCP-GDrive Container Build Script
# Purpose: Build and tag MCP Google Drive server container
# Usage: ./build_mcp_container.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

echo "🐳 97layerOS MCP-GDrive Container Builder"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📂 Project Root: $PROJECT_ROOT"
echo "📦 Containerfile: $SCRIPT_DIR/Containerfile"
echo ""

# Check if Podman is installed
if ! command -v podman &> /dev/null; then
    echo "❌ Podman not found. Please install Podman first."
    exit 1
fi

# Set TMPDIR to avoid macOS sandbox issues
export TMPDIR=/tmp

# Build container
echo "🔨 Building 97layer-mcp-gdrive container..."
podman build \
    -t 97layer-mcp-gdrive:latest \
    -f "$SCRIPT_DIR/Containerfile" \
    "$SCRIPT_DIR"

if [ $? -eq 0 ]; then
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "✅ MCP-GDrive container built successfully"
    echo ""
    echo "📋 Image Details:"
    podman images 97layer-mcp-gdrive
    echo ""
    echo "🚀 Next Steps:"
    echo "   1. Set up Google Drive credentials in credentials/gdrive_auth.json"
    echo "   2. Add GOOGLE_DRIVE_FOLDER_ID to .env"
    echo "   3. Run: ./run_mcp_server.sh"
    echo ""
    echo "💡 Slow Life Reminder: Container-First protocol ensures clean separation."
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
else
    echo ""
    echo "❌ Build failed. Check the error messages above."
    exit 1
fi
