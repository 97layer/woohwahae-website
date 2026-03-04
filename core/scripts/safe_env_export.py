#!/usr/bin/env python3
"""Print shell-safe export lines from .env style files."""

from __future__ import annotations

import argparse
import re
import shlex
from pathlib import Path

_ENV_RE = re.compile(r"^([A-Za-z_][A-Za-z0-9_]*)=(.*)$")


def parse_env_file(env_path: Path) -> dict[str, str]:
    values: dict[str, str] = {}
    for raw in env_path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        match = _ENV_RE.match(line)
        if not match:
            continue
        key = match.group(1)
        value = match.group(2).strip()
        if len(value) >= 2 and value[0] == value[-1] and value[0] in {"'", '"'}:
            value = value[1:-1]
        values[key] = value
    return values


def to_shell_exports(values: dict[str, str], selected_keys: list[str] | None = None) -> str:
    keys = selected_keys or list(values.keys())
    lines: list[str] = []
    for key in keys:
        if key not in values:
            continue
        lines.append(f"export {key}={shlex.quote(values[key])}")
    return "\n".join(lines)


def main() -> int:
    parser = argparse.ArgumentParser(description="Safe .env to shell export converter")
    parser.add_argument("--file", default=".env", help="Path to .env file")
    parser.add_argument(
        "--keys",
        nargs="*",
        help="Optional key whitelist. If omitted, prints all keys",
    )
    args = parser.parse_args()

    env_file = Path(args.file)
    if not env_file.exists():
        return 0

    values = parse_env_file(env_file)
    print(to_shell_exports(values, args.keys))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
