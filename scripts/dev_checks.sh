#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<'EOF' >&2
Usage: dev_checks.sh [--lint-only | --test-only]

Runs golangci-lint and go test with sandbox-friendly caches under /tmp.
Override caches via env:
  GOCACHE, GOMODCACHE, GOLANGCI_LINT_CACHE, GOLANGCI_LINT_TEMP_DIR
EOF
}

LINT_ONLY=0
TEST_ONLY=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --lint-only) LINT_ONLY=1 ;;
    --test-only) TEST_ONLY=1 ;;
    -h|--help) usage; exit 0 ;;
    *) usage; exit 2 ;;
  esac
  shift
done

export GOCACHE="${GOCACHE:-/tmp/gocache}"
export GOMODCACHE="${GOMODCACHE:-/tmp/gomodcache}"
export GOLANGCI_LINT_CACHE="${GOLANGCI_LINT_CACHE:-/tmp/golangci-lint-cache}"
export GOLANGCI_LINT_TEMP_DIR="${GOLANGCI_LINT_TEMP_DIR:-/tmp/golangci-lint-temp}"
export GOLANGCI_LINT_SKIP_FILES="${GOLANGCI_LINT_SKIP_FILES:-}"

run_lint() {
  if ! command -v golangci-lint >/dev/null 2>&1; then
    echo "golangci-lint not found; install with:" >&2
    echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" >&2
    exit 1
  fi
  echo "==> golangci-lint run"
  (cd "${ROOT_DIR}" && golangci-lint run --verbose)
}

run_tests() {
  echo "==> go test -cover ./..."
  (cd "${ROOT_DIR}" && go test -cover ./...)
}

if [[ "${LINT_ONLY}" -eq 1 ]]; then
  run_lint
  exit 0
fi

if [[ "${TEST_ONLY}" -eq 1 ]]; then
  run_tests
  exit 0
fi

run_lint
run_tests
