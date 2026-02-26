#!/usr/bin/env bash
# LAYER OS Session Bootstrap - MANDATORY execution at session start
# This script MUST be run by every AI agent before starting work

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "╔════════════════════════════════════════════════════════════════╗"
echo "║  LAYER OS Session Bootstrap - Enforced Protocol               ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""

# 1. Verify Python environment
if ! command -v python3 &> /dev/null; then
    echo "❌ FATAL: python3 not found in PATH"
    exit 1
fi

echo "✓ Python 3 detected: $(python3 --version)"
echo ""

# 2. Check mandatory files and create if missing
echo "🔍 Step 1: Checking mandatory files..."
echo "─────────────────────────────────────────"

WORK_LOCK="knowledge/system/work_lock.json"
FS_CACHE="knowledge/system/filesystem_cache.json"

if [ ! -f "$WORK_LOCK" ]; then
    echo "⚠️  Creating missing: $WORK_LOCK"
    mkdir -p "$(dirname "$WORK_LOCK")"
    cat > "$WORK_LOCK" <<'EOF'
{
  "locked": false,
  "agent": null,
  "task": null,
  "started_at": null,
  "expires_at": null,
  "metadata": {
    "created": "AUTO_GENERATED",
    "version": "1.0",
    "enforcement": "mandatory"
  }
}
EOF
fi

if [ ! -f "$FS_CACHE" ]; then
    echo "⚠️  Creating missing: $FS_CACHE"
    mkdir -p "$(dirname "$FS_CACHE")"
    cat > "$FS_CACHE" <<'EOF'
{
  "files": [],
  "directories": [],
  "last_scan": "1970-01-01T00:00:00+00:00",
  "scan_count": 0,
  "metadata": {
    "created": "AUTO_GENERATED",
    "version": "1.0",
    "enforcement": "mandatory"
  }
}
EOF
fi

echo "✅ Mandatory files verified"
echo ""

# 3. Run handoff.py --onboard
echo "🔄 Step 2: Session Handoff - Onboarding"
echo "─────────────────────────────────────────"

if python3 core/system/handoff.py --onboard; then
    echo "✅ Handoff onboard completed"
else
    echo "❌ FATAL: Handoff onboard failed"
    exit 1
fi

echo ""

# 4. Display INTELLIGENCE_QUANTA summary
echo "📖 Step 3: Intelligence Quanta Summary"
echo "─────────────────────────────────────────"

QUANTA_FILE="knowledge/agent_hub/INTELLIGENCE_QUANTA.md"
if [ -f "$QUANTA_FILE" ]; then
    # Show last update timestamp
    LAST_MOD=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M:%S" "$QUANTA_FILE" 2>/dev/null || stat -c "%y" "$QUANTA_FILE" 2>/dev/null | cut -d'.' -f1)
    QUANTA_AGE=$(( $(date +%s) - $(stat -f %m "$QUANTA_FILE" 2>/dev/null || stat -c %Y "$QUANTA_FILE") ))
    QUANTA_AGE_HOURS=$(( QUANTA_AGE / 3600 ))

    echo "Last updated: $LAST_MOD ($QUANTA_AGE_HOURS hours ago)"
    echo ""

    # Show first 50 lines
    head -50 "$QUANTA_FILE"

    # Warning if stale
    if [ $QUANTA_AGE_HOURS -gt 24 ]; then
        echo ""
        echo "⚠️  WARNING: QUANTA is ${QUANTA_AGE_HOURS}h old (>24h)"
        echo "   Previous session may not have completed handoff properly"
    fi
else
    echo "❌ FATAL: INTELLIGENCE_QUANTA.md not found"
    exit 1
fi

echo ""
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║  ✅ Session Bootstrap Complete - Ready to Work                ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""
echo "📝 Next actions:"
echo "   1. Read the QUANTA summary above"
echo "   2. Check work_lock.json for ongoing tasks"
echo "   3. Begin your work"
echo ""
echo "🚨 REMEMBER: Run handoff before session ends:"
echo "   python3 core/system/handoff.py --handoff \\"
echo "     --agent-id 'your-name' \\"
echo "     --summary 'What you did' \\"
echo "     --next-steps 'task1' 'task2'"
echo ""
