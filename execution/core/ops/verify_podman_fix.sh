#!/bin/bash
#
# Comprehensive Podman Fix Verification Script
# Tests all aspects of the TMPDIR bypass solution
#

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}=== Podman Fix Verification ===${NC}"
echo ""

# Test 1: TMPDIR environment variable
echo -e "${YELLOW}[Test 1] TMPDIR Environment Variable${NC}"
export TMPDIR=/tmp
if [ "$TMPDIR" = "/tmp" ]; then
    echo -e "  ${GREEN}✓ TMPDIR correctly set to /tmp${NC}"
else
    echo -e "  ${RED}✗ TMPDIR not set correctly${NC}"
    exit 1
fi
echo ""

# Test 2: Podman ps command
echo -e "${YELLOW}[Test 2] Podman ps Command${NC}"
if export TMPDIR=/tmp && podman ps > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓ podman ps works with TMPDIR bypass${NC}"
    container_count=$(export TMPDIR=/tmp && podman ps --format '{{.Names}}' | wc -l)
    echo -e "  Running containers: $container_count"
else
    echo -e "  ${RED}✗ podman ps failed${NC}"
    exit 1
fi
echo ""

# Test 3: Fix script exists and is executable
echo -e "${YELLOW}[Test 3] Fix Script Availability${NC}"
FIX_SCRIPT="/Users/97layer/97layerOS/scripts/fix_podman_macos.sh"
if [ -f "$FIX_SCRIPT" ]; then
    echo -e "  ${GREEN}✓ Fix script exists${NC}"
    if [ -x "$FIX_SCRIPT" ]; then
        echo -e "  ${GREEN}✓ Fix script is executable${NC}"
    else
        echo -e "  ${RED}✗ Fix script is not executable${NC}"
        exit 1
    fi
else
    echo -e "  ${RED}✗ Fix script not found${NC}"
    exit 1
fi
echo ""

# Test 4: Container exec capability
echo -e "${YELLOW}[Test 4] Container Exec Capability${NC}"
test_container="97layer-workspace"
if export TMPDIR=/tmp && podman ps --format '{{.Names}}' | grep -q "$test_container"; then
    echo -e "  ${GREEN}✓ Test container ($test_container) is running${NC}"

    if export TMPDIR=/tmp && podman exec "$test_container" echo "OK" > /dev/null 2>&1; then
        echo -e "  ${GREEN}✓ Container exec works${NC}"

        # Test Python execution
        python_version=$(export TMPDIR=/tmp && podman exec "$test_container" python3 --version 2>&1)
        echo -e "  ${GREEN}✓ Python in container: $python_version${NC}"
    else
        echo -e "  ${RED}✗ Container exec failed${NC}"
        exit 1
    fi
else
    echo -e "  ${YELLOW}! Test container not running (not critical)${NC}"
fi
echo ""

# Test 5: Fix script health check
echo -e "${YELLOW}[Test 5] Fix Script Health Check${NC}"
if "$FIX_SCRIPT" health > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓ Health check command works${NC}"
else
    echo -e "  ${RED}✗ Health check command failed${NC}"
    exit 1
fi
echo ""

# Test 6: Images available
echo -e "${YELLOW}[Test 6] Podman Images${NC}"
image_count=$(export TMPDIR=/tmp && podman images --format '{{.Repository}}' | wc -l)
if [ "$image_count" -gt 0 ]; then
    echo -e "  ${GREEN}✓ $image_count image(s) available${NC}"
else
    echo -e "  ${YELLOW}! No images found (might need to build)${NC}"
fi
echo ""

# Test 7: Permission test on /tmp
echo -e "${YELLOW}[Test 7] /tmp Write Permissions${NC}"
test_file="/tmp/podman_test_$$"
if touch "$test_file" 2>/dev/null; then
    echo -e "  ${GREEN}✓ /tmp is writable${NC}"
    rm -f "$test_file"
else
    echo -e "  ${RED}✗ /tmp is not writable${NC}"
    exit 1
fi
echo ""

# Final Summary
echo -e "${GREEN}=== All Tests Passed ===${NC}"
echo ""
echo "Your Podman infrastructure is working correctly with the TMPDIR bypass."
echo ""
echo "Recommended usage:"
echo "  1. For one-off commands: export TMPDIR=/tmp && podman <command>"
echo "  2. For regular use: $FIX_SCRIPT <command>"
echo "  3. For persistence: Add 'export TMPDIR=/tmp' to ~/.zshrc"
echo ""
echo "Quick commands:"
echo "  $FIX_SCRIPT health"
echo "  $FIX_SCRIPT ps"
echo "  $FIX_SCRIPT exec CONTAINER COMMAND"
