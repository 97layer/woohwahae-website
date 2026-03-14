#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGET_DIR="${BRAND_HOME_ASSET_TARGET_DIR:-$ROOT_DIR/docs/brand-home/public/assets/media/brand}"
SOURCE_DIR="${BRAND_HOME_ASSET_SOURCE_DIR:-$TARGET_DIR}"

CHECK_ONLY=false
if [[ "${1:-}" == "--check" ]]; then
  CHECK_ONLY=true
fi

FILES=(
  "symbol.png"
  "symbol.svg"
  "symbol-nav.png"
  "hero-graphic.svg"
  "hair-atelier-hero.svg"
  "icon-192.png"
  "icon-512.png"
  "objects/ceramic.svg"
  "objects/incense.svg"
  "objects/pendant.svg"
  "objects/snap.svg"
  "objects/tea.svg"
)

if [[ ! -d "$SOURCE_DIR" ]]; then
  echo "brand source missing: $SOURCE_DIR" >&2
  exit 1
fi

echo "source=$SOURCE_DIR"
echo "target=$TARGET_DIR"

missing=0
for rel in "${FILES[@]}"; do
  if [[ ! -f "$SOURCE_DIR/$rel" ]]; then
    echo "missing=$rel" >&2
    missing=1
  fi
done

if [[ "$missing" -ne 0 ]]; then
  exit 1
fi

if [[ "$CHECK_ONLY" == true ]]; then
  printf 'ready=%s\n' "${#FILES[@]}"
  exit 0
fi

if [[ "$SOURCE_DIR" == "$TARGET_DIR" ]]; then
  printf 'already_local=%s\n' "${#FILES[@]}"
  exit 0
fi

mkdir -p "$TARGET_DIR/objects"

for rel in "${FILES[@]}"; do
  mkdir -p "$(dirname "$TARGET_DIR/$rel")"
  cp "$SOURCE_DIR/$rel" "$TARGET_DIR/$rel"
  echo "copied=$rel"
done

printf 'synced=%s\n' "${#FILES[@]}"
