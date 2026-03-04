#!/usr/bin/env bash
# signal_quick.sh — 최소 입력으로 신호 저장 + evidence 자동 기록
# 사용 예:
#   core/scripts/signal_quick.sh "오늘 회의 메모"
#   core/scripts/signal_quick.sh https://example.com
#   core/scripts/signal_quick.sh --pdf /path/doc.pdf --claim "보고서 캡처"
# 기본 동작: 신호 저장 후 evidence_log.jsonl에 기록 (--no-evidence로 비활성화)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
INJECT="${ROOT}/core/scripts/signal_inject.py"

usage() {
  cat <<'EOF'
사용:
  signal_quick.sh "텍스트" [--claim "..."]
  signal_quick.sh https://url [--claim "..."]
  signal_quick.sh --image /path/file.jpg [--claim "..."]
  signal_quick.sh --pdf /path/file.pdf [--claim "..."]
옵션:
  --no-evidence   evidence_log.jsonl 기록 생략
  --claim TEXT    evidence_log claim 덮어쓰기
  --help          이 도움말
EOF
}

TYPE=""
PAYLOAD=""
CLAIM=""
LOG_EVIDENCE=1

while [[ $# -gt 0 ]]; do
  case "$1" in
    --text)   TYPE="text";  PAYLOAD="${2:-}"; shift 2 ;;
    --url)    TYPE="url";   PAYLOAD="${2:-}"; shift 2 ;;
    --image|--img|--file)
              TYPE="image"; PAYLOAD="${2:-}"; shift 2 ;;
    --pdf)    TYPE="pdf";   PAYLOAD="${2:-}"; shift 2 ;;
    --claim)  CLAIM="${2:-}"; shift 2 ;;
    --no-evidence)
              LOG_EVIDENCE=0; shift ;;
    --help)   usage; exit 0 ;;
    *)        # 첫 번째 비옵션 인자를 텍스트/URL로 자동 감지
              if [[ -z "${PAYLOAD}" ]]; then
                PAYLOAD="$1"
              else
                echo "불필요한 인자: $1" >&2; usage; exit 1
              fi
              shift ;;
  esac
done

if [[ -z "${PAYLOAD}" ]]; then
  echo "입력이 없습니다." >&2
  usage
  exit 1
fi

if [[ -z "${TYPE}" ]]; then
  if [[ "${PAYLOAD}" =~ ^https?:// ]]; then
    TYPE="url"
  else
    TYPE="text"
  fi
fi

cd "${ROOT}"

LOG_FLAG=()
ETYPE="file"
case "${TYPE}" in
  text)  ETYPE="doc" ;;
  url)   ETYPE="url" ;;
  image|pdf) ETYPE="file" ;;
esac

if [[ "${LOG_EVIDENCE}" -eq 1 ]]; then
  LOG_FLAG=(--log-evidence --evidence-type "${ETYPE}")
  [[ -n "${CLAIM}" ]] && LOG_FLAG+=(--claim "${CLAIM}")
fi

case "${TYPE}" in
  text)  exec python3 "${INJECT}" --text "${PAYLOAD}" "${LOG_FLAG[@]:-}" ;;
  url)   exec python3 "${INJECT}" --url "${PAYLOAD}" "${LOG_FLAG[@]:-}" ;;
  image) exec python3 "${INJECT}" --image "${PAYLOAD}" "${LOG_FLAG[@]:-}" ;;
  pdf)   exec python3 "${INJECT}" --pdf "${PAYLOAD}" "${LOG_FLAG[@]:-}" ;;
  *)     echo "알 수 없는 타입: ${TYPE}" >&2; exit 1 ;;
esac
