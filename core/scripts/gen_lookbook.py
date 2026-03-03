#!/usr/bin/env python3
"""
gen_lookbook.py — WOOHWAHAE 룩북 Imagen 4.0 이미지 생성

사용법:
    python3 core/scripts/gen_lookbook.py
    python3 core/scripts/gen_lookbook.py --prompt-index 0
    python3 core/scripts/gen_lookbook.py --all
"""

import argparse
import os
import sys
from pathlib import Path

from dotenv import load_dotenv

load_dotenv(Path(__file__).resolve().parents[2] / ".env")

import google.genai as genai
from google.genai import types

OUTPUT_DIR = Path(__file__).resolve().parents[2] / "website/archive/lookbook/assets/images"
OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

# 브랜드 기반 공통 접두 프롬프트
BASE_STYLE = (
    "editorial fashion photography, slow life aesthetics, "
    "muted desaturated palette, soft natural light from a single window, "
    "Japanese-Korean minimalism, quiet atmosphere, analog grain, "
    "medium format film look, no text, no logo"
)

PROMPTS = [
    {
        "filename": "lookbook-01-stillness.jpg",
        "prompt": (
            f"A pair of hands resting on a white ceramic bowl, morning light, "
            f"linen tablecloth, minimal composition, close-up. {BASE_STYLE}"
        ),
    },
    {
        "filename": "lookbook-02-texture.jpg",
        "prompt": (
            f"Extreme close-up of freshly cut hair strands on a pale stone floor, "
            f"geometric shadow pattern, soft focus background. {BASE_STYLE}"
        ),
    },
    {
        "filename": "lookbook-03-space.jpg",
        "prompt": (
            f"Empty atelier interior, bare white walls, single wooden stool, "
            f"afternoon window light casting long shadow, high ceiling, wide angle. {BASE_STYLE}"
        ),
    },
    {
        "filename": "lookbook-04-ritual.jpg",
        "prompt": (
            f"Hairdresser's hands delicately combing through dark hair, "
            f"bokeh background, intimate close-up, calm focus, warm shadow. {BASE_STYLE}"
        ),
    },
    {
        "filename": "lookbook-05-material.jpg",
        "prompt": (
            f"Flat lay of haircutting scissors, a small glass jar, dried botanicals "
            f"on rough linen, top-down shot, negative space dominant. {BASE_STYLE}"
        ),
    },
    {
        "filename": "lookbook-06-portrait.jpg",
        "prompt": (
            f"Side profile of a person with a precise short haircut, eyes closed, "
            f"against a soft neutral wall, late afternoon light, film portrait. {BASE_STYLE}"
        ),
    },
    {
        "filename": "lookbook-07-moment.jpg",
        "prompt": (
            f"A single cup of tea on a window ledge, rain blur outside, "
            f"condensation on glass, hands cradling the cup, crop mid-frame. {BASE_STYLE}"
        ),
    },
    {
        "filename": "lookbook-08-light.jpg",
        "prompt": (
            f"Abstract: a sheer white curtain diffusing late morning light, "
            f"soft creases, almost no detail, peaceful emptiness. {BASE_STYLE}"
        ),
    },
]


def generate_image(client: genai.Client, prompt_data: dict, dry_run: bool = False) -> bool:
    out_path = OUTPUT_DIR / prompt_data["filename"]

    if out_path.exists():
        print("skip (exists): %s" % prompt_data["filename"])
        return True

    if dry_run:
        print("[dry-run] would generate: %s" % prompt_data["filename"])
        print("  prompt: %s" % prompt_data["prompt"][:80])
        return True

    print("generating: %s" % prompt_data["filename"])
    try:
        response = client.models.generate_images(
            model="imagen-4.0-generate-001",
            prompt=prompt_data["prompt"],
            config=types.GenerateImagesConfig(
                number_of_images=1,
                aspect_ratio="3:4",  # 세로 룩북 포맷
                safety_filter_level="BLOCK_LOW_AND_ABOVE",
            ),
        )
        img = response.generated_images[0].image
        out_path.write_bytes(img.image_bytes)
        print("  saved: %s (%d KB)" % (out_path.name, len(img.image_bytes) // 1024))
        return True
    except Exception as e:
        print("  ERROR: %s" % e)
        return False


def main():
    parser = argparse.ArgumentParser(description="WOOHWAHAE 룩북 이미지 생성")
    parser.add_argument("--prompt-index", type=int, help="단일 프롬프트 인덱스 (0-based)")
    parser.add_argument("--all", action="store_true", help="전체 생성")
    parser.add_argument("--dry-run", action="store_true", help="실제 생성 없이 프롬프트 확인")
    args = parser.parse_args()

    api_key = os.getenv("GEMINI_API_KEY") or os.getenv("GOOGLE_API_KEY")
    if not api_key:
        print("ERROR: GEMINI_API_KEY or GOOGLE_API_KEY not set")
        sys.exit(1)

    client = genai.Client(api_key=api_key)

    if args.prompt_index is not None:
        if args.prompt_index >= len(PROMPTS):
            print("ERROR: index %d out of range (0-%d)" % (args.prompt_index, len(PROMPTS) - 1))
            sys.exit(1)
        targets = [PROMPTS[args.prompt_index]]
    else:
        targets = PROMPTS

    print("target: %d images → %s" % (len(targets), OUTPUT_DIR))
    ok = 0
    for p in targets:
        if generate_image(client, p, dry_run=args.dry_run):
            ok += 1

    print("\ndone: %d/%d" % (ok, len(targets)))


if __name__ == "__main__":
    main()
