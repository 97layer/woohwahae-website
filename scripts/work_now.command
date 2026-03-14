#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}/.."

"${SCRIPT_DIR}/work_now.sh" "$@"

printf '\nPress Enter to close...'
read -r _
