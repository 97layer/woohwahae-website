#!/bin/bash
# Set timezone to Asia/Seoul for all 97layer containers
# This script ensures all containers use Korean Standard Time (KST)

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "🕐 Setting timezone to Asia/Seoul (KST) for all containers..."
echo

# Function to set timezone for a container
set_container_timezone() {
    local container=$1

    # Check if container is running
    if ! export TMPDIR=/tmp && podman ps --format "{{.Names}}" | grep -q "^${container}$"; then
        echo -e "${YELLOW}⚠️  ${container}: Not running${NC}"
        return 1
    fi

    # Set timezone
    export TMPDIR=/tmp && podman exec -u root $container bash -c "
        # Install tzdata if needed (for Debian/Ubuntu based containers)
        if command -v apt-get &> /dev/null; then
            apt-get update > /dev/null 2>&1 && apt-get install -y tzdata > /dev/null 2>&1
        fi

        # Set timezone
        ln -sf /usr/share/zoneinfo/Asia/Seoul /etc/localtime
        echo 'Asia/Seoul' > /etc/timezone

        # For Alpine Linux containers
        if command -v apk &> /dev/null; then
            apk add --no-cache tzdata > /dev/null 2>&1
            cp /usr/share/zoneinfo/Asia/Seoul /etc/localtime
            echo 'Asia/Seoul' > /etc/timezone
        fi
    " 2>/dev/null

    if [ $? -eq 0 ]; then
        # Verify the time
        local current_time=$(export TMPDIR=/tmp && podman exec $container date 2>/dev/null)
        echo -e "${GREEN}✅ ${container}: $current_time${NC}"
        return 0
    else
        echo -e "${RED}❌ ${container}: Failed to set timezone${NC}"
        return 1
    fi
}

# List of containers to update
CONTAINERS=(
    "97layer-workspace"
    "97layer-snapshot"
    "97layer-gcp-mgmt"
    "97layer-receiver"
)

# Set timezone for each container
SUCCESS_COUNT=0
TOTAL_COUNT=0

for container in "${CONTAINERS[@]}"; do
    TOTAL_COUNT=$((TOTAL_COUNT + 1))
    if set_container_timezone "$container"; then
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    fi
done

echo
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Completed: $SUCCESS_COUNT/$TOTAL_COUNT containers updated"

# Show current time in KST
echo
echo "🇰🇷 Current time in Seoul:"
export TMPDIR=/tmp && podman exec 97layer-workspace date 2>/dev/null || date

# Create docker-compose override for persistent timezone setting
echo
echo "💡 To make timezone persistent across container restarts, add this to your docker-compose.yml or podman-compose.yml:"
echo
echo "  environment:"
echo "    - TZ=Asia/Seoul"
echo "  volumes:"
echo "    - /etc/localtime:/etc/localtime:ro"
echo "    - /etc/timezone:/etc/timezone:ro"