#!/usr/bin/env bash
# LAYER OS — Legacy MANIFEST Violations Cleanup
# Created: 2026-02-26
# Purpose: 검증 전 생성된 MANIFEST 위반 파일 정리

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"

echo "🧹 LAYER OS Legacy Violations Cleanup"
echo "======================================"
echo ""

# 백업 디렉토리
ARCHIVE="knowledge/docs/archive/legacy_violations_20260226"
mkdir -p "$ARCHIVE"

moved_count=0

# ── 1. knowledge/signals/*.md → archive ─────────────────────────
echo "[1/4] Cleaning knowledge/signals/*.md..."
if ls knowledge/signals/*.md >/dev/null 2>&1; then
    mkdir -p "$ARCHIVE/signals"
    for f in knowledge/signals/*.md; do
        [ -f "$f" ] || continue
        mv "$f" "$ARCHIVE/signals/"
        echo "  ✓ $(basename "$f")"
        ((moved_count++))
    done
fi

if ls knowledge/signals/wellness/*.md >/dev/null 2>&1; then
    mkdir -p "$ARCHIVE/signals_wellness"
    for f in knowledge/signals/wellness/*.md; do
        [ -f "$f" ] || continue
        mv "$f" "$ARCHIVE/signals_wellness/"
        echo "  ✓ $(basename "$f")"
        ((moved_count++))
    done
fi

# ── 2. knowledge/reports/*.md (비규격) → archive ────────────────
echo ""
echo "[2/4] Cleaning knowledge/reports/*.md (illegal patterns)..."

ILLEGAL_REPORTS=(
    "deep_scan_"
    "wellness_report_"
    "evening_summary_"
    "morning_briefing_"
    "WEBSITE_STATUS_REPORT_"
    "action_"
    "structure_audit_"
    "valuation_"
    "gui_ux_improvement_plan_"
    "monetization_strategy_"
    "strategy_"
    "update_"
    "validation_report"
    "additional_risks_"
    "web_interface_consistency_audit_"
)

mkdir -p "$ARCHIVE/reports"
for pattern in "${ILLEGAL_REPORTS[@]}"; do
    for f in knowledge/reports/${pattern}*.md; do
        [ -f "$f" ] || continue
        mv "$f" "$ARCHIVE/reports/"
        echo "  ✓ $(basename "$f")"
        ((moved_count++))
    done
done

# ── 3. knowledge/brands/ → archive (MANIFEST 미정의) ────────────
echo ""
echo "[3/4] Archiving knowledge/brands/ (not in MANIFEST)..."
if [ -d "knowledge/brands" ]; then
    mv knowledge/brands "$ARCHIVE/"
    echo "  ✓ knowledge/brands/ → $ARCHIVE/"
    ((moved_count++))
fi

# ── 4. knowledge/offering/ → knowledge/service/ (통합) ──────────
echo ""
echo "[4/4] Merging knowledge/offering/ → knowledge/service/..."
if [ -d "knowledge/offering" ]; then
    mkdir -p knowledge/service
    if [ -f "knowledge/offering/items.json" ]; then
        mv knowledge/offering/items.json knowledge/service/
        echo "  ✓ items.json → knowledge/service/"
        ((moved_count++))
    fi
    # 나머지 파일 있으면 archive
    if [ "$(ls -A knowledge/offering)" ]; then
        mv knowledge/offering "$ARCHIVE/"
        echo "  ✓ offering/ 잔여 파일 → $ARCHIVE/"
    else
        rmdir knowledge/offering
    fi
fi

# ── 완료 ─────────────────────────────────────────────────────
echo ""
echo "======================================"
echo "✅ Cleanup complete: $moved_count items moved"
echo "📦 Archive location: $ARCHIVE"
echo ""
echo "Next steps:"
echo "1. git status — 변경 사항 확인"
echo "2. python3 core/system/filesystem_validator.py --all — 검증"
echo "3. git add . && git commit -m \"chore: clean up legacy MANIFEST violations\""
echo ""
