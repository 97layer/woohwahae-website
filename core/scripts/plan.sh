#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 1 ]; then
  echo "Usage: bash core/scripts/plan.sh \"<task>\" [mode]" >&2
  echo "mode: preflight|hook (default: preflight)" >&2
  exit 2
fi

TASK="$1"
MODE="${2:-preflight}"

python3 core/system/plan_council.py --task "$TASK" --mode "$MODE"
