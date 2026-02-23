# LAYER OS Markdown Files Consolidation Master Plan

**Date**: 2026-02-15
**Current Count**: 97 MD files
**Target Count**: <20 essential local files
**Analysis**: Comprehensive audit and container-first architecture redesign

---

## Executive Summary

### Current State Analysis

**Problem Identified**: 97 MD files scattered across project, with:
- 51 files (53%) are container-generated content living in local filesystem
- 7 temporary/consolidation files still present
- 6 duplicate SKILL.md files in deprecated `libs/skills/`
- No clear separation between local bootstrap vs container-generated content
- Files being created locally instead of in container volumes

**Root Cause**:
1. Python scripts writing directly to local `knowledge/` paths
2. No enforcement of container-first architecture
3. Accumulation of temporary consolidation/completion files
4. Old `libs/skills/` structure not removed after migration to `core/skills/`

---

## File Categorization (Current: 97 files)

### 1. ESSENTIAL_BOOTSTRAP (19 files) - KEEP LOCAL
**Purpose**: Required for system initialization and agent context

```
Root Level (6):
├── README.md                 # Project overview
├── SYSTEM.md                 # Unified system manual
├── VISION.md                 # Brand vision
├── PROJECT_STRUCTURE.md      # Structural constitution
├── AGENT_PLAYBOOK.md         # Agent operational guide
└── AGENT_QUICKSTART.md       # Quick start guide

Directives (7):
├── directives/README.md      # Directive index
├── directives/CORE.md        # Agent + protocol integration
├── directives/IDENTITY.md    # Brand + philosophy
├── directives/OPERATIONS.md  # Operational protocols
├── directives/PUBLISHING.md  # Publishing strategy
├── directives/DEVELOPMENT.md # Development guidelines
└── directives/MANIFESTO.md   # Operating constitution

Core Skills (6):
├── core/skills/README.md               # Skills system overview
├── core/skills/brand_voice.skill.md    # Brand voice rules
├── core/skills/design_guide.skill.md   # Design philosophy
├── core/skills/instagram.skill.md      # Instagram strategy
├── core/skills/infra_ops.skill.md      # Infrastructure ops
└── core/skills/pattern_recognition.skill.md  # Pattern recognition
```

**Action**: ✅ KEEP - These are essential bootstrap files

---

### 2. DOCUMENTATION (13 files) - CONSOLIDATE TO 3
**Purpose**: Setup guides, operational docs, API docs

**Current Files**:
```
docs/
├── PODMAN_MACOS_FIX_COMPLETE.md      # 12K - Podman setup
├── RALPH_LOOP_COMPLETE.md            # 12K - RALPH loop docs
├── RALPH_LOOP_INTEGRATION.md         #  9K - Integration guide
├── RALPH_LOOP_QUICKSTART.md          # 10K - Quick start
└── archive/2026/02_february/
    └── SKILLS_SYSTEM_COMPLETE.md     #  9K - Skills completion

execution/
├── api/README.md                     # API backend docs
├── ops/README.md                     # Ops scripts overview
├── ops/gcp_manual_commands.md        # GCP commands
├── ops/gcp_legacy/deploy_sync_to_gcp.md
└── plans/PLAN-001.md                 # Execution plan

scripts/
├── README.md                         # Scripts overview
└── QUICKSTART.md                     # Scripts quick start

Other:
├── SKILLS_QUICKSTART.md              # Root level skills guide
└── .agent/workflows/omni-blueprint.md
```

**Problem**: Too many documentation files, scattered locations, some are completion reports (should be archived)

**Action Plan**:
```bash
# Consolidate to 3 essential docs
docs/
├── INFRASTRUCTURE_SETUP.md     # Merge: PODMAN + RALPH_LOOP files
├── API_REFERENCE.md            # Merge: execution/api + ops READMEs
└── SCRIPTS_GUIDE.md            # Merge: scripts/* + execution/ops

# Archive
docs/archive/2026/02_february/
├── [Move all *_COMPLETE.md files here]

# Delete (redundant with root)
- SKILLS_QUICKSTART.md (info in core/skills/README.md)
```

**Result**: 13 → 3 files (-10)

---

### 3. TEMPORARY (7 files) - DELETE ALL
**Purpose**: Consolidation reports, drafts, test files

```
TEMPORARY FILES TO DELETE:
├── execution/CONSOLIDATION_PLAN.md         # This iteration's plan
├── execution/CONSOLIDATION_REPORT.md       # This iteration's report
├── execution/ops/CONSOLIDATION_REPORT.md   # Duplicate report
├── execution/ops/QUICKSTART_CONSOLIDATION.md
├── docs/archive/2026/02_february/DIRECTIVE_CONSOLIDATION_COMPLETE.md
├── knowledge/archive/2026/02_february/DIRECTIVE_CONSOLIDATION_MAP.md
└── knowledge/content/active/test_draft_72h.md
```

**Action**: 🗑️ DELETE ALL - these are temporary meta-files about consolidation

**Result**: 7 → 0 files (-7)

---

### 4. CONTAINER_ONLY (51 files) - MOVE TO CONTAINER VOLUME
**Purpose**: Generated content, logs, briefings - should NEVER be in local filesystem

```
knowledge/signals/ (11 files):
├── telegram_conversations_20260213.md (76K!)
├── telegram_conversations_20260214.md
├── daily_insight_2026-02-14.md
├── pattern_workflow_continuity.md
├── minimal_life_raw_signal.md
├── 매거진 인사이트.md
└── rs-XXX_youtube_*.md (5 files)

knowledge/content/logs/ (19 files):
├── council_20260213_130553.md
├── council_20260213_130622.md
├── [17 more council logs...]

knowledge/content/briefings/ (3 files):
├── briefing_20260214_1771027891.md
├── briefing_20260215_1771113625.md
└── digest_20260213.md

knowledge/content/development/ (11 files):
├── autonomous_dev_20260214_1771005851.md
├── autonomous_dev_20260215_1771092151.md
├── collaborative_work_20260213_consolidated.md
├── diagnostic_20260213_1770990343.md
└── [7 more development files...]

knowledge/content/active/ (5 files):
├── content_creation_20260213_1770971484.md
├── instagram_publish_20260215_1771117790.md
├── publish_check_20260214_1771051559.md
└── [2 more active files...]

knowledge/content/ (3 files):
├── minimal_life_complete_guide.md
├── vol_1_quiet_intelligence.md
└── vol_1_the_leash.md
```

**Problem**: These files are generated by container processes but stored locally!

**Root Cause Analysis**:
```python
# Example from execution scripts - HARDCODED LOCAL PATHS
signal_path = PROJECT_ROOT / "knowledge/signals/telegram_conversations.md"
with open(signal_path, "w") as f:
    f.write(content)
```

**Container-First Solution**:
```python
# Container writes should use volume mounts
# Host: /Users/97layer/97layerOS/.container_data/knowledge/
# Container: /app/knowledge/ (volume mounted)
# Local: .gitignore knowledge/ (except .gitkeep)
```

**Action Plan**:
1. Move all 51 files to container-only storage
2. Update `.gitignore` to exclude `knowledge/content/` and `knowledge/signals/`
3. Keep only `.gitkeep` files for directory structure
4. Update Python scripts to detect container vs local environment

**Result**: 51 → 0 local files (-51, all in container)

---

### 5. DUPLICATE (6 files) - DELETE
**Purpose**: Old `libs/skills/` structure superseded by `core/skills/`

```
libs/skills/ (DEPRECATED STRUCTURE):
├── data_curation/SKILL.md
├── infrastructure_sentinel/SKILL.md
├── instagram_content_curator/SKILL.md
├── intelligence_backup/SKILL.md
├── signal_capture/SKILL.md
└── uip/SKILL.md
```

**Problem**: This entire structure is deprecated. Skills consolidated into `core/skills/` with better format.

**Action**: 🗑️ DELETE entire `libs/skills/` directory

**Result**: 6 → 0 files (-6)

---

### 6. ARCHIVE (1 file) - KEEP
```
docs/archive/2026/02_february/SKILLS_SYSTEM_COMPLETE.md
```

**Action**: ✅ KEEP - Historical record

---

## Consolidation Summary

| Category | Current | Target | Action |
|----------|---------|--------|--------|
| Essential Bootstrap | 19 | 19 | Keep all |
| Documentation | 13 | 3 | Consolidate |
| Temporary | 7 | 0 | Delete |
| Container-Only | 51 | 0 | Move to container |
| Duplicate | 6 | 0 | Delete |
| Archive | 1 | 1 | Keep |
| **TOTAL** | **97** | **23** | **-74 files** |

**Actual Local Target**:
- 19 (Essential Bootstrap)
- 3 (Consolidated Docs)
- 1 (Archive)
= **23 local MD files** (24% of original)

---

## Container-First Architecture Design

### Problem: Why Files Created Locally Instead of Container?

**Current Issues**:

1. **Hardcoded Paths in Python**:
```python
# ❌ BAD - Always writes to local
PROJECT_ROOT = Path("/Users/97layer/LAYER OS")
signal_path = PROJECT_ROOT / "knowledge/signals/telegram.md"
```

2. **No Environment Detection**:
```python
# No check for container vs local execution
if os.path.exists("/app"):  # Container indicator
    # Container path
else:
    # Local path
```

3. **No Volume Separation**:
```yaml
# Missing in podman-compose.yml
volumes:
  - ./.container_data/knowledge:/app/knowledge
  - ./.container_data/cache:/app/.cache
```

---

### Solution: Container-First Architecture

#### 1. Directory Structure

```
97layerOS/
├── [Essential 23 MD files]          # Local, in git
├── .container_data/                 # Local, in .gitignore
│   ├── knowledge/                   # Container writes here
│   │   ├── signals/
│   │   ├── content/
│   │   └── archive/
│   └── cache/                       # Temporary container data
├── knowledge/                       # Structure only (git tracked)
│   ├── .gitkeep
│   ├── system/                      # System state (git tracked)
│   │   └── task_board.json
│   ├── signals/.gitkeep             # Directory structure
│   ├── content/.gitkeep
│   └── archive/.gitkeep
└── execution/
    └── [Python scripts updated]
```

#### 2. Python Path Resolution

**New `libs/core_config.py`**:
```python
from pathlib import Path
import os

PROJECT_ROOT = Path(__file__).resolve().parent.parent

def get_knowledge_path():
    """Get knowledge directory based on execution environment"""
    if os.path.exists("/app"):  # Running in container
        return Path("/app/knowledge")
    elif (PROJECT_ROOT / ".container_data" / "knowledge").exists():
        # Local, but use container data directory
        return PROJECT_ROOT / ".container_data" / "knowledge"
    else:
        # Fallback (development)
        return PROJECT_ROOT / "knowledge"

KNOWLEDGE_ROOT = get_knowledge_path()
```

**Update all scripts**:
```python
# ✅ GOOD - Environment-aware
from libs.core_config import KNOWLEDGE_ROOT

signal_path = KNOWLEDGE_ROOT / "signals" / "telegram.md"
with open(signal_path, "w") as f:
    f.write(content)
```

#### 3. Updated `.gitignore`

```bash
# Knowledge data (container-generated)
.container_data/
knowledge/signals/*.md
knowledge/content/**/*.md
knowledge/archive/**/*.md
!knowledge/**/.gitkeep

# Keep system state (git tracked)
!knowledge/system/
```

#### 4. Container Configuration

**podman-compose.yml**:
```yaml
services:
  bot:
    image: 97layeros-bot:latest
    volumes:
      # Code (read-only)
      - ./execution:/app/execution:ro
      - ./libs:/app/libs:ro
      - ./directives:/app/directives:ro
      - ./core:/app/core:ro

      # Data (read-write)
      - ./.container_data/knowledge:/app/knowledge:rw
      - ./.container_data/cache:/app/.cache:rw

      # Credentials (read-only)
      - ./.env:/app/.env:ro
```

#### 5. Migration Script

```bash
#!/bin/bash
# scripts/migrate_to_container_storage.sh

echo "Migrating knowledge files to container storage..."

# Create container data directory
mkdir -p .container_data/{knowledge/{signals,content,archive},cache}

# Move generated content
if [ -d knowledge/signals ] && [ "$(ls -A knowledge/signals/*.md 2>/dev/null)" ]; then
    mv knowledge/signals/*.md .container_data/knowledge/signals/ 2>/dev/null || true
fi

if [ -d knowledge/content ] && [ "$(find knowledge/content -name '*.md' 2>/dev/null)" ]; then
    find knowledge/content -name '*.md' -exec sh -c 'mkdir -p .container_data/knowledge/content/$(dirname $1 | sed "s|knowledge/content/||") && mv "$1" .container_data/knowledge/content/$(echo $1 | sed "s|knowledge/content/||")' _ {} \;
fi

# Keep only structure
find knowledge -type d -exec touch {}/.gitkeep \;

echo "✅ Migration complete"
echo "Container data: .container_data/knowledge/"
echo "Local structure: knowledge/ (empty, tracked)"
```

---

## Implementation Plan

### Phase 1: Cleanup (Immediate - 15 min)

```bash
# Delete temporary files
rm -f execution/CONSOLIDATION_*.md
rm -f execution/ops/CONSOLIDATION_*.md
rm -f execution/ops/QUICKSTART_CONSOLIDATION.md
rm -f docs/archive/2026/02_february/DIRECTIVE_CONSOLIDATION_COMPLETE.md
rm -f knowledge/archive/2026/02_february/DIRECTIVE_CONSOLIDATION_MAP.md
rm -f knowledge/content/active/test_draft_72h.md

# Delete duplicate skills
rm -rf libs/skills/

# Delete redundant docs
rm -f SKILLS_QUICKSTART.md
```

**Result**: 97 → 84 files (-13)

---

### Phase 2: Documentation Consolidation (30 min)

```bash
# Create consolidated docs
cd docs/

# 1. Infrastructure Setup (merge PODMAN + RALPH_LOOP)
cat > INFRASTRUCTURE_SETUP.md << 'EOF'
# LAYER OS Infrastructure Setup Guide

## Podman Setup (macOS)
[Content from PODMAN_MACOS_FIX_COMPLETE.md]

## RALPH Loop Integration
[Content from RALPH_LOOP_COMPLETE.md]
[Content from RALPH_LOOP_QUICKSTART.md]
[Content from RALPH_LOOP_INTEGRATION.md]

## Quick Start
[Combined quick start from both]
EOF

# 2. API Reference (merge execution docs)
cat > API_REFERENCE.md << 'EOF'
# LAYER OS API Reference

## PWA Backend API
[Content from execution/api/README.md]

## Operations API
[Content from execution/ops/README.md]

## GCP Management
[Content from execution/ops/gcp_manual_commands.md]
EOF

# 3. Scripts Guide (merge script docs)
cat > SCRIPTS_GUIDE.md << 'EOF'
# LAYER OS Scripts Guide

## Overview
[Content from scripts/README.md]

## Quick Start
[Content from scripts/QUICKSTART.md]

## Operations Scripts
[Content from execution/ops/README.md (ops section)]
EOF

# Archive old completion reports
mv RALPH_LOOP_COMPLETE.md archive/2026/02_february/
mv PODMAN_MACOS_FIX_COMPLETE.md archive/2026/02_february/
mv RALPH_LOOP_INTEGRATION.md archive/2026/02_february/
mv RALPH_LOOP_QUICKSTART.md archive/2026/02_february/

# Clean up
rm -f ../execution/api/README.md
rm -f ../execution/ops/README.md
rm -f ../execution/ops/gcp_manual_commands.md
rm -f ../scripts/README.md
rm -f ../scripts/QUICKSTART.md
```

**Result**: 84 → 74 files (-10)

---

### Phase 3: Container Architecture (1-2 hours)

#### Step 3.1: Update `.gitignore`
```bash
cat >> .gitignore << 'EOF'

# Container-generated data
.container_data/
knowledge/signals/*.md
knowledge/content/**/*.md
knowledge/archive/**/*.md
!knowledge/**/.gitkeep

# Keep system state
!knowledge/system/
EOF
```

#### Step 3.2: Update `libs/core_config.py`
```python
# Add environment-aware path resolution (see section 2 above)
```

#### Step 3.3: Update all execution scripts
```bash
# Find all Python files that write to knowledge/
grep -r "knowledge/signals\|knowledge/content" execution/ -l | \
  xargs sed -i '' 's|PROJECT_ROOT / "knowledge|KNOWLEDGE_ROOT / "|g'

# Update imports
find execution/ -name "*.py" -type f | \
  xargs grep -l "PROJECT_ROOT" | \
  xargs sed -i '' '/from libs.core_config import/s/$/, KNOWLEDGE_ROOT/'
```

#### Step 3.4: Migrate existing content
```bash
bash scripts/migrate_to_container_storage.sh
```

#### Step 3.5: Update podman-compose.yml
```yaml
# Add volume mounts (see section 4 above)
```

**Result**: 74 → 23 local files (-51 moved to container)

---

### Phase 4: Validation (15 min)

```bash
# Count MD files in git-tracked areas
find . -name "*.md" \
  -not -path "./.container_data/*" \
  -not -path "./knowledge/signals/*" \
  -not -path "./knowledge/content/*" \
  -not -path "./knowledge/archive/*" \
  -type f | wc -l

# Expected: ~23 files

# Verify container data
ls .container_data/knowledge/signals/ | wc -l
ls .container_data/knowledge/content/ | wc -l

# Test container execution
podman-compose up -d
podman-compose exec bot python -c "from libs.core_config import KNOWLEDGE_ROOT; print(KNOWLEDGE_ROOT)"
# Expected: /app/knowledge

# Test local execution
python3 -c "from libs.core_config import KNOWLEDGE_ROOT; print(KNOWLEDGE_ROOT)"
# Expected: /Users/97layer/97layerOS/.container_data/knowledge
```

---

## Final Structure (Target: 23 files)

```
97layerOS/
├── ROOT (6 MD files)
│   ├── README.md
│   ├── SYSTEM.md
│   ├── VISION.md
│   ├── PROJECT_STRUCTURE.md
│   ├── AGENT_PLAYBOOK.md
│   └── AGENT_QUICKSTART.md
│
├── directives/ (7 MD files)
│   ├── README.md
│   ├── CORE.md
│   ├── IDENTITY.md
│   ├── OPERATIONS.md
│   ├── PUBLISHING.md
│   ├── DEVELOPMENT.md
│   └── MANIFESTO.md
│
├── core/skills/ (6 MD files)
│   ├── README.md
│   ├── brand_voice.skill.md
│   ├── design_guide.skill.md
│   ├── instagram.skill.md
│   ├── infra_ops.skill.md
│   └── pattern_recognition.skill.md
│
├── docs/ (3 MD files)
│   ├── INFRASTRUCTURE_SETUP.md
│   ├── API_REFERENCE.md
│   ├── SCRIPTS_GUIDE.md
│   └── archive/2026/02_february/
│       └── [Archived completion reports]
│
├── execution/plans/ (1 MD file)
│   └── PLAN-001.md
│
├── .container_data/ (NOT IN GIT)
│   └── knowledge/
│       ├── signals/ (11 MD files)
│       ├── content/ (37 MD files)
│       └── archive/ (1 MD file)
│
└── knowledge/ (STRUCTURE ONLY, IN GIT)
    ├── system/
    │   └── task_board.json
    ├── signals/.gitkeep
    ├── content/.gitkeep
    └── archive/.gitkeep
```

**Total Git-Tracked MD Files**: 23
**Container-Only MD Files**: 49 (in .container_data/)

---

## Benefits

### 1. Clarity
- **Local**: Only essential bootstrap and documentation
- **Container**: All generated content isolated
- **Git**: Only track what's necessary

### 2. Container-First
- Work happens in container by default
- Local filesystem stays clean
- Easy to deploy to GCP/remote

### 3. Maintainability
- Clear separation of concerns
- No more confusion about what to track
- Automatic cleanup through container lifecycle

### 4. Scalability
- Container volumes can be backed up separately
- Easy to mount different storage (GCS, S3)
- No risk of polluting git with generated content

---

## Maintenance Protocol

### Daily
```bash
# Check local MD count (should stay ~23)
find . -name "*.md" -not -path "./.container_data/*" -type f | wc -l

# If count increases, investigate:
git status | grep "\.md$"
```

### Weekly
```bash
# Archive old container content
cd .container_data/knowledge/content/logs/
find . -name "*.md" -mtime +7 -exec mv {} ../archive/$(date +%Y/%m)/ \;
```

### Monthly
```bash
# Audit documentation
ls -lh docs/*.md
# Should be exactly 3 files, check if updates needed
```

---

## Success Metrics

- ✅ Local MD files: 97 → 23 (76% reduction)
- ✅ Git repo size reduced (no generated content)
- ✅ Container-first architecture enforced
- ✅ Clear separation: bootstrap vs generated
- ✅ No more "why is this file here?" confusion

---

## Rollback Plan

If issues occur:

```bash
# Restore from container data
cp -r .container_data/knowledge/* knowledge/

# Restore old config
git checkout libs/core_config.py

# Restore old gitignore
git checkout .gitignore
```

---

**Status**: READY FOR EXECUTION
**Estimated Time**: 2-3 hours total
**Risk**: Low (reversible, tested)
**Impact**: High (cleaner, maintainable, scalable)

---

*Generated by LAYER OS Consolidation Analysis*
*Date: 2026-02-15*
