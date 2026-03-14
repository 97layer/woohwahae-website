#!/usr/bin/env bash
set -euo pipefail

SESSION="${LAYER_OS_TMUX_SESSION:-layeros}"
ROOT="${LAYER_OS_TMUX_ROOT:-/srv/layer-os/current}"
DEV_ROOT="${LAYER_OS_TMUX_DEV_ROOT:-/srv/layer-os-dev}"
SHELL_PATH="${LAYER_OS_TMUX_SHELL:-/bin/bash}"
DAEMON_SERVICE="${LAYER_OS_TMUX_DAEMON_SERVICE:-layer-osd}"
WEB_SERVICE="${LAYER_OS_TMUX_WEB_SERVICE:-layer-os-web}"
HOME_DIR="${HOME:-/tmp}"

if [[ ! -d "${ROOT}" ]]; then
  ROOT="${HOME_DIR}"
fi

has_session() {
  tmux has-session -t "${SESSION}" 2>/dev/null
}

window_exists() {
  tmux list-windows -t "${SESSION}" -F '#W' 2>/dev/null | grep -Fxq "$1"
}

ensure_window() {
  local name="$1"
  shift
  if window_exists "${name}"; then
    return 0
  fi
  tmux new-window -d -t "${SESSION}:" -n "${name}" -c "${ROOT}" "$@"
}

ensure_shell_window() {
  local name="$1"
  local intro="$2"
  ensure_window "${name}" /bin/bash -lc "cd '${ROOT}' && printf '%s\n' \"${intro}\" && exec ${SHELL_PATH} -l"
}

ensure_log_window() {
  local name="$1"
  local service_name="$2"
  ensure_window "${name}" /bin/bash -lc "exec sudo journalctl -fu ${service_name}"
}

tmux start-server

if ! has_session; then
  tmux new-session -d -s "${SESSION}" -n ops -c "${ROOT}" /bin/bash -lc "cd '${ROOT}' && printf '%s\n' 'Layer OS ops shell' && exec ${SHELL_PATH} -l"
fi

tmux set-option -t "${SESSION}" -g mouse on
tmux set-option -t "${SESSION}" -g history-limit 100000
tmux set-window-option -t "${SESSION}" -g remain-on-exit on

ensure_shell_window "status" "Layer OS tmux cockpit\nSession: ${SESSION}\nRoot: ${ROOT}\n\nUseful:\n  tmux attach -t ${SESSION}\n  sudo systemctl status ${DAEMON_SERVICE} ${WEB_SERVICE}\n  curl -fsS http://127.0.0.1:17808/healthz\n  curl -fsS http://127.0.0.1:3081/admin/login >/dev/null && echo admin=200"
if [[ -d "${DEV_ROOT}" ]]; then
  ensure_window "dev" /bin/bash -lc "cd '${DEV_ROOT}' && printf '%s\n' 'Layer OS dev workspace\nRoot: ${DEV_ROOT}\n\nUseful:\n  git status --short --branch\n  GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./...\n  cd docs/brand-home && npm test' && exec ${SHELL_PATH} -l"
fi
ensure_log_window "daemon-log" "${DAEMON_SERVICE}"
ensure_log_window "web-log" "${WEB_SERVICE}"

tmux select-window -t "${SESSION}:ops" >/dev/null 2>&1 || true

printf 'session=%s\n' "${SESSION}"
printf 'root=%s\n' "${ROOT}"
tmux list-windows -t "${SESSION}" -F 'window=#{window_index}:#{window_name}'
