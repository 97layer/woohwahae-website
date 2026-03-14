#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

REMOTE_HOST="${LAYER_OS_VM_HOST:-97layer-vm}"
REMOTE_USER="${LAYER_OS_VM_USER:-skyto5339_gmail_com}"
REMOTE_PORT="${LAYER_OS_VM_PORT:-22}"
REMOTE_SSH_KEY="${LAYER_OS_VM_SSH_KEY:-}"
NODE_ROOT="${LAYER_OS_VM_NODE_ROOT:-/srv/layer-os/node/current}"
INSTALL_PREFIX="${LAYER_OS_VM_AI_CLI_PREFIX:-}"
CHECK_ONLY=0

usage() {
  cat >&2 <<'EOF'
usage: install_vm_ai_clis.sh [--host <ssh-host>] [--user <remote-user>] [--port <port>] [--ssh-key <path>] [--node-root <path>] [--check]

Install Codex, Claude Code, and Gemini CLI onto the Layer OS VM by reusing the
bundled Layer OS Node runtime and a user-local npm prefix.
EOF
}

note() {
  printf '%s\n' "$1"
}

die() {
  printf 'error: %s\n' "$1" >&2
  exit 1
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    die "required command not found: $1"
  fi
}

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --host)
      [[ "$#" -ge 2 ]] || die "--host requires a value"
      REMOTE_HOST="$2"
      shift 2
      ;;
    --user)
      [[ "$#" -ge 2 ]] || die "--user requires a value"
      REMOTE_USER="$2"
      shift 2
      ;;
    --port)
      [[ "$#" -ge 2 ]] || die "--port requires a value"
      REMOTE_PORT="$2"
      shift 2
      ;;
    --ssh-key)
      [[ "$#" -ge 2 ]] || die "--ssh-key requires a value"
      REMOTE_SSH_KEY="$2"
      shift 2
      ;;
    --node-root)
      [[ "$#" -ge 2 ]] || die "--node-root requires a value"
      NODE_ROOT="$2"
      shift 2
      ;;
    --check)
      CHECK_ONLY=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

require_cmd ssh

SSH_ARGS=(-p "${REMOTE_PORT}" -o StrictHostKeyChecking=accept-new)
if [[ -n "${REMOTE_SSH_KEY}" ]]; then
  SSH_ARGS+=(-i "${REMOTE_SSH_KEY}")
fi
REMOTE_TARGET="${REMOTE_USER}@${REMOTE_HOST}"

note "remote_host=${REMOTE_HOST}"
note "remote_user=${REMOTE_USER}"
note "remote_port=${REMOTE_PORT}"
note "node_root=${NODE_ROOT}"
note "install_prefix=${INSTALL_PREFIX}"

if [[ "${CHECK_ONLY}" -eq 1 ]]; then
  ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "
    set -euo pipefail
    test -x '${NODE_ROOT}/bin/node'
    test -x '${NODE_ROOT}/lib/node_modules/npm/bin/npm-cli.js'
    printf 'node=%s\n' \"\$('${NODE_ROOT}/bin/node' -v)\"
    printf 'npm=%s\n' \"\$('${NODE_ROOT}/bin/node' '${NODE_ROOT}/lib/node_modules/npm/bin/npm-cli.js' -v)\"
  "
  note "vm ai cli install check complete"
  exit 0
fi

ssh "${SSH_ARGS[@]}" "${REMOTE_TARGET}" "
  set -euo pipefail

  NODE_ROOT='${NODE_ROOT}'
  INSTALL_PREFIX='${INSTALL_PREFIX}'
  if [[ -z \"\${INSTALL_PREFIX}\" ]]; then
    INSTALL_PREFIX=\"\$HOME/.local\"
  fi
  BIN_DIR=\"\${INSTALL_PREFIX}/bin\"
  PATH_SNIPPET_PATH=\"\$HOME/.config/layer-os/ai-cli-path.sh\"

  if [[ ! -x \"\${NODE_ROOT}/bin/node\" ]]; then
    printf 'error: missing node runtime at %s\n' \"\${NODE_ROOT}/bin/node\" >&2
    exit 1
  fi

  mkdir -p \"\$HOME/.config/layer-os\" \"\${BIN_DIR}\"

  cat > \"\${PATH_SNIPPET_PATH}\" <<'EOF'
export PATH=\"\$HOME/.local/bin:\$PATH\"
EOF
  chmod 600 \"\${PATH_SNIPPET_PATH}\"

  for shell_rc in \"\$HOME/.profile\" \"\$HOME/.bash_profile\" \"\$HOME/.bashrc\" \"\$HOME/.zprofile\" \"\$HOME/.zshrc\"; do
    touch \"\${shell_rc}\"
    if ! grep -Fq 'ai-cli-path.sh' \"\${shell_rc}\"; then
      printf '\n[ -f \"\$HOME/.config/layer-os/ai-cli-path.sh\" ] && . \"\$HOME/.config/layer-os/ai-cli-path.sh\"\n' >> \"\${shell_rc}\"
    fi
  done

  ln -sfn \"\${NODE_ROOT}/bin/node\" \"\${BIN_DIR}/node\"
  ln -sfn \"\${NODE_ROOT}/bin/npm\" \"\${BIN_DIR}/npm\"
  ln -sfn \"\${NODE_ROOT}/bin/npx\" \"\${BIN_DIR}/npx\"
  if [[ -e \"\${NODE_ROOT}/bin/corepack\" ]]; then
    ln -sfn \"\${NODE_ROOT}/bin/corepack\" \"\${BIN_DIR}/corepack\"
  fi

  export PATH=\"\${BIN_DIR}:\$PATH\"
  npm config set prefix \"\${INSTALL_PREFIX}\"

  npm install -g @openai/codex @anthropic-ai/claude-code @google/gemini-cli

  printf 'prefix=%s\n' \"\$(npm config get prefix)\"
  printf 'node=%s\n' \"\$(node -v)\"
  printf 'npm=%s\n' \"\$(npm -v)\"
  printf 'codex_path=%s\n' \"\$(command -v codex)\"
  printf 'claude_path=%s\n' \"\$(command -v claude)\"
  printf 'gemini_path=%s\n' \"\$(command -v gemini)\"
  printf 'codex_version=%s\n' \"\$(codex --version 2>/dev/null || codex -V 2>/dev/null || echo unavailable)\"
  printf 'claude_version=%s\n' \"\$(claude --version 2>/dev/null || echo unavailable)\"
  printf 'gemini_version=%s\n' \"\$(gemini --version 2>/dev/null || echo unavailable)\"
"

note "ai CLIs ready on ${REMOTE_HOST}"
note "open a fresh shell or run: source ~/.config/layer-os/ai-cli-path.sh"
