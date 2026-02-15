#!/bin/bash
#
# Podman macOS Fix Script
# Bypass macOS sandbox restrictions for Podman operations
#
# Issue: macOS sandbox prevents Podman from creating temporary files
# Solution: Set TMPDIR=/tmp for all Podman operations
#
# Usage: ./fix_podman_macos.sh [command]
#   - No args: Show container status
#   - ps: List running containers
#   - exec <container> <cmd>: Execute command in container
#   - health: Check all container health
#   - restart <container>: Restart specific container
#

set -e

# CRITICAL: Set TMPDIR to bypass macOS sandbox
export TMPDIR=/tmp

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to handle errors gracefully
handle_error() {
    local exit_code=$?
    local error_msg="$1"
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}ERROR:${NC} $error_msg (exit code: $exit_code)" >&2
        return $exit_code
    fi
}

# Function to check container health
check_health() {
    local container="$1"
    echo -e "${YELLOW}Checking $container...${NC}"

    # Get container status
    local status=$(podman inspect "$container" --format '{{.State.Status}}' 2>/dev/null)
    if [ $? -ne 0 ]; then
        echo -e "  ${RED}NOT FOUND${NC}"
        return 1
    fi

    # Get uptime
    local started=$(podman inspect "$container" --format '{{.State.StartedAt}}' 2>/dev/null)

    echo -e "  Status: ${GREEN}$status${NC}"
    echo -e "  Started: $started"

    # Test exec capability
    if [ "$status" = "running" ]; then
        if podman exec "$container" echo "OK" > /dev/null 2>&1; then
            echo -e "  Exec: ${GREEN}OK${NC}"
        else
            echo -e "  Exec: ${RED}FAILED${NC}"
            return 1
        fi
    fi

    return 0
}

# Function to list containers
list_containers() {
    echo -e "${YELLOW}=== Running Containers ===${NC}"
    podman ps || handle_error "Failed to list containers"
    echo ""
    echo -e "${YELLOW}=== All Containers ===${NC}"
    podman ps -a || handle_error "Failed to list all containers"
}

# Function to check all known containers
check_all_health() {
    echo -e "${YELLOW}=== Container Health Check ===${NC}"

    local containers=(
        "97layer-workspace"
        "97layer-snapshot"
        "97layer-gcp-mgmt"
        "97layer-receiver"
        "97layer-nightguard"
    )

    local healthy=0
    local unhealthy=0

    for container in "${containers[@]}"; do
        if check_health "$container"; then
            ((healthy++))
        else
            ((unhealthy++))
        fi
        echo ""
    done

    echo -e "${YELLOW}=== Summary ===${NC}"
    echo -e "Healthy: ${GREEN}$healthy${NC}"
    echo -e "Unhealthy/Missing: ${RED}$unhealthy${NC}"
}

# Function to execute command in container
exec_in_container() {
    local container="$1"
    shift
    local cmd="$@"

    echo -e "${YELLOW}Executing in $container:${NC} $cmd"
    podman exec "$container" $cmd || handle_error "Failed to execute command"
}

# Function to restart container
restart_container() {
    local container="$1"

    echo -e "${YELLOW}Restarting $container...${NC}"
    podman restart "$container" || handle_error "Failed to restart container"

    # Wait a moment and check health
    sleep 2
    check_health "$container"
}

# Main script logic
main() {
    echo -e "${GREEN}=== Podman macOS Fix Script ===${NC}"
    echo -e "TMPDIR: $TMPDIR"
    echo ""

    case "${1:-status}" in
        status|ps)
            list_containers
            ;;
        health)
            check_all_health
            ;;
        exec)
            if [ -z "$2" ]; then
                echo -e "${RED}Usage:${NC} $0 exec <container> <command>"
                exit 1
            fi
            exec_in_container "$2" "${@:3}"
            ;;
        restart)
            if [ -z "$2" ]; then
                echo -e "${RED}Usage:${NC} $0 restart <container>"
                exit 1
            fi
            restart_container "$2"
            ;;
        *)
            echo "Usage: $0 [status|ps|health|exec|restart]"
            echo ""
            echo "Commands:"
            echo "  status/ps  - List all containers (default)"
            echo "  health     - Check health of all known containers"
            echo "  exec       - Execute command in container"
            echo "  restart    - Restart a container"
            exit 1
            ;;
    esac
}

main "$@"
