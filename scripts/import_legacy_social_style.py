#!/usr/bin/env python3
from __future__ import annotations

import json
import os
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
LOCAL_SOURCE = ROOT / "docs/brand-home/content/social-style-source.json"
TARGET = ROOT / "docs/brand-home/content/social-style-examples.generated.js"


def as_text(value: object) -> str:
    return value.strip() if isinstance(value, str) else ""


def compact_excerpt(value: str, limit: int = 180) -> str:
    text = " ".join(value.replace("\r", "\n").split())
    return text[: limit - 1].rstrip() + "…" if len(text) > limit else text


def resolve_source() -> Path:
    explicit = os.getenv("SOCIAL_STYLE_SOURCE")
    if explicit:
        return Path(explicit).expanduser()
    return LOCAL_SOURCE


def load_examples(source_path: Path) -> list[dict[str, str]]:
    data = json.loads(source_path.read_text(encoding="utf-8"))
    rows: list[dict[str, str]] = []
    for index, item in enumerate(data.get("published_content", []), start=1):
        excerpt = as_text(item.get("instagram_caption_preview"))
        if not excerpt:
            continue
        rows.append(
            {
                "exampleId": f"legacy-ig-{index:02d}",
                "signalId": as_text(item.get("signal_id")),
                "publishedAt": as_text(item.get("published_at")),
                "excerpt": compact_excerpt(excerpt),
            }
        )
    return rows


def render_module(examples: list[dict[str, str]], source_path: Path) -> str:
    payload = {
        "instagram": examples,
        "threads": examples[:8],
    }
    return (
        "// Generated from the repo-local social style snapshot.\n"
        "// Builder: scripts/import_legacy_social_style.py\n"
        f"// Source: {source_path}\n"
        f"export const socialStyleExamples = {json.dumps(payload, ensure_ascii=False, indent=2)};\n"
        "export default socialStyleExamples;\n"
    )


def main() -> None:
    source_path = resolve_source()
    if not source_path.exists():
        raise SystemExit(f"social style source missing: {source_path}")
    examples = load_examples(source_path)
    TARGET.write_text(render_module(examples, source_path), encoding="utf-8")
    print(f"examples={len(examples)}")
    print(f"source={source_path}")
    print(f"target={TARGET}")


if __name__ == "__main__":
    main()
