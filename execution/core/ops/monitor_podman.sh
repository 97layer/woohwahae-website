#!/bin/bash

# 97layerOS Podman Container Monitor
# 실시간 컨테이너 상태 모니터링 스크립트

clear
echo "========================================="
echo "   97layerOS Podman Container Monitor   "
echo "========================================="
echo ""

while true; do
    # 화면 지우고 헤더 출력
    tput cup 5 0

    # 현재 시간
    echo "📅 Last Updated: $(date '+%Y-%m-%d %H:%M:%S')"
    echo ""

    # 컨테이너 상태
    echo "📦 Container Status:"
    echo "-------------------"
    podman ps --format "table {{.Names}}\t{{.Status}}\t{{.Created}}"
    echo ""

    # 리소스 사용량
    echo "📊 Resource Usage:"
    echo "-----------------"
    podman stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"
    echo ""

    # 최근 로그 (각 컨테이너별 마지막 2줄)
    echo "📝 Recent Logs:"
    echo "--------------"
    for container in 97layer-snapshot 97layer-gcp-mgmt 97layer-receiver; do
        if podman ps --format "{{.Names}}" | grep -q "$container"; then
            echo "[$container]"
            podman logs "$container" --tail 2 2>/dev/null | sed 's/^/  /'
            echo ""
        fi
    done

    # 포트 상태
    echo "🌐 Port Status:"
    echo "--------------"
    echo "  8081: GCP Management"
    curl -s http://localhost:8081/health 2>/dev/null && echo "    ✅ Healthy" || echo "    ⚠️  No response"
    echo "  9876: Realtime Receiver"
    curl -s http://localhost:9876/health 2>/dev/null && echo "    ✅ Healthy" || echo "    ⚠️  No response"
    echo ""

    echo "-------------------"
    echo "Press Ctrl+C to exit"

    # 5초마다 업데이트
    sleep 5
done