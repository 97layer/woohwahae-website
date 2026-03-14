#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import re
from collections import Counter
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
LOCAL_SOURCE = ROOT / "docs/brand-home/content/social-style-source.json"
TARGET = ROOT / "docs/brand-home/content/social-style-analysis.generated.js"

THEME_RULES = {
    "subtraction": ["덜어", "비워", "여백", "본질"],
    "stillness": ["고요", "침묵", "멈춘", "조용", "느리", "슬로우"],
    "observation": ["바라", "응시", "시선", "머무", "헤아", "드러"],
    "identity": ["삶", "존재", "이름", "나의", "자기"],
    "daily_ritual": ["아침", "샤워", "수면", "물건", "옷", "허기", "공간"],
    "questioning": ["무엇", "어디", "어떻게", "일까", "?"],
}

STOPWORDS = {
    "그리고",
    "그러나",
    "하지만",
    "우리는",
    "우리의",
    "오늘",
    "때로는",
    "가장",
    "모든",
    "다시",
    "이것",
    "그것",
    "무언가",
    "있다",
    "한다",
    "합니다",
    "하는",
    "하게",
}


def as_text(value: object) -> str:
    return value.strip() if isinstance(value, str) else ""


def normalize_excerpt(value: str) -> str:
    return " ".join(value.replace("\r", "\n").split())


def resolve_source() -> Path:
    explicit = os.getenv("SOCIAL_STYLE_SOURCE")
    if explicit:
        return Path(explicit).expanduser()
    return LOCAL_SOURCE


def load_rows(source_path: Path) -> list[dict[str, str]]:
    data = json.loads(source_path.read_text(encoding="utf-8"))
    rows: list[dict[str, str]] = []
    for index, item in enumerate(data.get("published_content", []), start=1):
        excerpt = normalize_excerpt(as_text(item.get("instagram_caption_preview")))
        if not excerpt:
            continue
        rows.append(
            {
                "exampleId": f"legacy-ig-{index:02d}",
                "signalId": as_text(item.get("signal_id")),
                "publishedAt": as_text(item.get("published_at")),
                "excerpt": excerpt,
            }
        )
    return rows


def match_theme(excerpt: str, needles: list[str]) -> bool:
    return any(needle in excerpt for needle in needles)


def theme_summary(rows: list[dict[str, str]]) -> list[dict[str, object]]:
    summary: list[dict[str, object]] = []
    for theme_id, needles in THEME_RULES.items():
        matches = [row for row in rows if match_theme(row["excerpt"], needles)]
        summary.append(
            {
                "themeId": theme_id,
                "hits": len(matches),
                "coverage": round(len(matches) / max(1, len(rows)), 3),
                "signals": needles,
                "examples": [
                    {
                        "exampleId": row["exampleId"],
                        "signalId": row["signalId"],
                        "excerpt": row["excerpt"][:160] + ("…" if len(row["excerpt"]) > 160 else ""),
                    }
                    for row in matches[:3]
                ],
            }
        )
    return sorted(summary, key=lambda item: int(item["hits"]), reverse=True)


def rhetorical_summary(rows: list[dict[str, str]]) -> dict[str, object]:
    question_like = 0
    average_length = 0.0
    if rows:
        question_like = sum(1 for row in rows if any(token in row["excerpt"] for token in ["?", "무엇", "어디", "일까"]))
        average_length = round(sum(len(row["excerpt"]) for row in rows) / len(rows), 1)
    return {
        "questionLikeRate": round(question_like / max(1, len(rows)), 3),
        "averageExcerptLength": average_length,
        "shortParagraphBias": True,
        "notes": [
            "Concrete observation tends to arrive before explanation.",
            "Many examples close with a question or suspended reflection.",
            "The voice returns to subtraction, quiet, and essence rather than momentum.",
        ],
    }


def keyword_summary(rows: list[dict[str, str]]) -> list[dict[str, object]]:
    counter: Counter[str] = Counter()
    for row in rows:
        for token in re.findall(r"[가-힣]{2,}", row["excerpt"]):
            if token in STOPWORDS:
                continue
            counter[token] += 1
    return [
        {"keyword": keyword, "count": count}
        for keyword, count in counter.most_common(12)
    ]


def build_payload(rows: list[dict[str, str]]) -> dict[str, object]:
    themes = theme_summary(rows)
    top_theme_ids = [item["themeId"] for item in themes[:4]]
    return {
        "summary": {
            "sourceExamples": len(rows),
            "sharedEssence": [
                "subtraction over accumulation",
                "quiet over performance",
                "observation before declaration",
                "a reflective question near the end",
            ],
            "dominantThemes": top_theme_ids,
        },
        "themes": themes,
        "rhetoric": rhetorical_summary(rows),
        "keywords": keyword_summary(rows),
    }


def render_module(payload: dict[str, object], source_path: Path) -> str:
    return (
        "// Generated from the repo-local social style snapshot.\n"
        "// Builder: scripts/analyze_legacy_social_style.py\n"
        f"// Source: {source_path}\n"
        f"export const socialStyleAnalysis = {json.dumps(payload, ensure_ascii=False, indent=2)};\n"
        "export default socialStyleAnalysis;\n"
    )


def main() -> None:
    source_path = resolve_source()
    if not source_path.exists():
        raise SystemExit(f"social style source missing: {source_path}")
    rows = load_rows(source_path)
    payload = build_payload(rows)
    TARGET.write_text(render_module(payload, source_path), encoding="utf-8")
    print(f"examples={len(rows)}")
    print(f"source={source_path}")
    print(f"target={TARGET}")
    print(f"dominant_themes={','.join(payload['summary']['dominantThemes'])}")


if __name__ == "__main__":
    main()
