#!/usr/bin/env python3
"""
evidence_guard.py - Evidence log helper for anti-hallucination workflow.

Usage:
  python3 core/system/evidence_guard.py --check
  python3 core/system/evidence_guard.py --append \
      --claim "..." --evidence-type command --source "rg -n ..."
"""

from __future__ import annotations

import argparse
import json
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, List


PROJECT_ROOT = Path(__file__).resolve().parents[2]
DEFAULT_LOG = PROJECT_ROOT / "knowledge" / "system" / "evidence_log.jsonl"
REQUIRED_KEYS = ("timestamp", "claim", "evidence_type", "source")
ALLOWED_TYPES = {"file", "command", "url", "doc", "test", "inference"}


def _validate_entry(entry: Dict[str, object], lineno: int) -> List[str]:
    errors: List[str] = []
    for key in REQUIRED_KEYS:
        if key not in entry:
            errors.append(f"line {lineno}: missing key '{key}'")
    evidence_type = str(entry.get("evidence_type", ""))
    if evidence_type and evidence_type not in ALLOWED_TYPES:
        errors.append(
            f"line {lineno}: invalid evidence_type '{evidence_type}' "
            f"(allowed: {sorted(ALLOWED_TYPES)})"
        )
    return errors


def check_log(log_path: Path) -> int:
    if not log_path.exists():
        print(f"BLOCKED missing: {log_path}")
        return 1

    errors: List[str] = []
    for lineno, raw in enumerate(log_path.read_text(encoding="utf-8").splitlines(), start=1):
        line = raw.strip()
        if not line:
            continue
        try:
            entry = json.loads(line)
        except json.JSONDecodeError as exc:
            errors.append(f"line {lineno}: invalid json ({exc})")
            continue
        if not isinstance(entry, dict):
            errors.append(f"line {lineno}: entry must be object")
            continue
        errors.extend(_validate_entry(entry, lineno))

    if errors:
        for err in errors:
            print(f"BLOCKED {err}")
        print("BLOCKED")
        return 1

    print("READY")
    return 0


def append_entry(
    log_path: Path,
    claim: str,
    evidence_type: str,
    source: str,
    detail: str | None,
) -> int:
    if evidence_type not in ALLOWED_TYPES:
        print(
            f"ERROR invalid --evidence-type '{evidence_type}'. "
            f"Allowed: {', '.join(sorted(ALLOWED_TYPES))}"
        )
        return 1

    entry: Dict[str, object] = {
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "claim": claim,
        "evidence_type": evidence_type,
        "source": source,
    }
    if detail:
        entry["detail"] = detail

    log_path.parent.mkdir(parents=True, exist_ok=True)
    with log_path.open("a", encoding="utf-8") as fp:
        fp.write(json.dumps(entry, ensure_ascii=False) + "\n")

    print("OK")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description="Evidence log guard")
    parser.add_argument(
        "--log",
        default=str(DEFAULT_LOG),
        help="Path to evidence jsonl (default: knowledge/system/evidence_log.jsonl)",
    )
    parser.add_argument("--check", action="store_true", help="Validate evidence log format")
    parser.add_argument("--append", action="store_true", help="Append one evidence record")
    parser.add_argument("--claim", help="Claim text for --append")
    parser.add_argument("--evidence-type", help="Evidence type for --append")
    parser.add_argument("--source", help="Evidence source for --append")
    parser.add_argument("--detail", help="Optional detail for --append")
    args = parser.parse_args()

    log_path = Path(args.log)

    if args.check and args.append:
        print("ERROR choose one: --check or --append")
        return 1
    if not args.check and not args.append:
        print("ERROR specify --check or --append")
        return 1

    if args.check:
        return check_log(log_path)

    if not args.claim or not args.evidence_type or not args.source:
        print("ERROR --append requires --claim --evidence-type --source")
        return 1
    return append_entry(log_path, args.claim, args.evidence_type, args.source, args.detail)


if __name__ == "__main__":
    sys.exit(main())
