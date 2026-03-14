#!/usr/bin/env python3
from __future__ import annotations

import json
import os
from datetime import datetime, timezone
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TARGET = ROOT / "docs/brand-home/content/social-style-source.json"


def as_text(value: object) -> str:
    return value.strip() if isinstance(value, str) else ""


def resolve_legacy_source() -> Path:
    explicit = as_text(os.getenv("LEGACY_SOCIAL_STYLE_SOURCE"))
    if explicit:
        return Path(explicit).expanduser()
    raise SystemExit(
        "LEGACY_SOCIAL_STYLE_SOURCE is required. "
        "The repo-local snapshot is already canonical, so this absorb step is only for explicit external legacy imports."
    )


def build_payload(source_path: Path) -> dict[str, object]:
    data = json.loads(source_path.read_text(encoding="utf-8"))
    rows: list[dict[str, str]] = []
    for item in data.get("published_content", []):
        excerpt = as_text(item.get("instagram_caption_preview"))
        if not excerpt:
            continue
        rows.append(
            {
                "signal_id": as_text(item.get("signal_id")),
                "published_at": as_text(item.get("published_at")),
                "instagram_caption_preview": excerpt,
            }
        )
    return {
        "source_mode": "legacy_absorbed_snapshot",
        "source_path": str(source_path),
        "generated_at": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
        "published_content": rows,
    }


def main() -> None:
    source_path = resolve_legacy_source()
    if not source_path.exists():
        raise SystemExit(f"social style source missing: {source_path}")
    payload = build_payload(source_path)
    TARGET.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"examples={len(payload['published_content'])}")
    print(f"target={TARGET}")


if __name__ == "__main__":
    main()
