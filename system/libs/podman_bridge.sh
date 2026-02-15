#!/bin/bash
# 97layerOS Podman Collaboration Bridge
# 로컬 Podman CLI의 권한 장애를 우회하여 Podman Desktop VM에 직접 명령을 전달합니다.

SOCKET="/tmp/97layer_podman.sock"
SSH_KEY="/Users/97layer/.local/share/containers/podman/machine/machine"
SSH_PORT=53674

# SSH 터널 활성 여부 확인 및 재기동
if [ ! -S "$SOCKET" ]; then
    rm -f "$SOCKET"
    ssh -i "$SSH_KEY" -p "$SSH_PORT" -o StrictHostKeyChecking=no -f -N -L "$SOCKET":/run/user/501/podman/podman.sock core@localhost
    sleep 1
fi

# 명령 인자가 없는 경우 도움말 출력
if [ $# -eq 0 ]; then
    echo "Usage: $0 [podman commands...]"
    echo "Example: $0 ps"
    exit 0
fi

# VM 내부에서 직접 podman 명령 수행 (로컬 CLI 장애 우회)
# rootful: true 설정에 맞추어 sudo 사용
ssh -i "$SSH_KEY" -p "$SSH_PORT" -o StrictHostKeyChecking=no core@localhost "sudo podman $(printf '%q ' "$@")"
