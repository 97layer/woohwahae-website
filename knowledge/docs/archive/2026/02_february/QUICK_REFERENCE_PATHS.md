# Quick Reference: Container-Aware Paths

## Usage in Python Scripts

```python
# Import at top of file
from libs.core_config import KNOWLEDGE_PATHS

# Use KNOWLEDGE_PATHS for all runtime data
signal_file = KNOWLEDGE_PATHS["signals"] / "my_signal.md"
draft_file = KNOWLEDGE_PATHS["content"] / "draft" / "my_draft.md"
state_file = KNOWLEDGE_PATHS["system"] / "state.json"
```

## Available Paths

| Key | Local Path | Container Path | Purpose |
|-----|-----------|----------------|---------|
| `signals` | `.container_data/knowledge/signals` | `/app/.container_data/knowledge/signals` | Raw signal capture |
| `content` | `.container_data/knowledge/content` | `/app/.container_data/knowledge/content` | Processed content |
| `archive` | `.container_data/knowledge/archive` | `/app/.container_data/knowledge/archive` | Time-based archival |
| `system` | `.container_data/knowledge/system` | `/app/.container_data/knowledge/system` | System state |
| `blueprints` | `knowledge/blueprints` | `/app/knowledge/blueprints` | Bootstrap only |
| `manuals` | `knowledge/manuals` | `/app/knowledge/manuals` | Bootstrap only |

## Common Subdirectories

```python
# Content subdirectories
KNOWLEDGE_PATHS["content"] / "draft"
KNOWLEDGE_PATHS["content"] / "published"
KNOWLEDGE_PATHS["content"] / "ready_to_publish"
KNOWLEDGE_PATHS["content"] / "notifications"
KNOWLEDGE_PATHS["content"] / "council_log"
KNOWLEDGE_PATHS["content"] / "daily_briefs"

# System subdirectories
KNOWLEDGE_PATHS["system"] / "task_board.json"
KNOWLEDGE_PATHS["system"] / "sync_state.json"
KNOWLEDGE_PATHS["system"] / "ralph_loop.json"
KNOWLEDGE_PATHS["system"] / "quality_results"
```

## Migration Pattern

### Before (Hardcoded)
```python
signal_dir = Path("knowledge/raw_signals")
draft_file = Path("knowledge/assets/draft/my_draft.md")
state_file = Path("knowledge/system/state.json")
```

### After (Container-Aware)
```python
from libs.core_config import KNOWLEDGE_PATHS

signal_dir = KNOWLEDGE_PATHS["signals"]
draft_file = KNOWLEDGE_PATHS["content"] / "draft" / "my_draft.md"
state_file = KNOWLEDGE_PATHS["system"] / "state.json"
```

## Verification

```bash
# Test import
python3 -c "from libs.core_config import KNOWLEDGE_PATHS; print(KNOWLEDGE_PATHS)"

# Check environment detection
python3 -c "from libs.core_config import ENVIRONMENT; print(f'Environment: {ENVIRONMENT}')"

# Verify directory creation
python3 -c "from libs.core_config import KNOWLEDGE_PATHS; [p.mkdir(parents=True, exist_ok=True) for p in KNOWLEDGE_PATHS.values()]"
```

## Why Container-Aware Paths?

1. **Isolation**: Container data separate from source code
2. **Portability**: Same code works locally and in containers
3. **Sync Efficiency**: `.container_data/` excluded from git
4. **Clean Deployment**: No runtime data mixed with source
5. **Zero Config**: Environment auto-detected

## Updated Files (Phase 5)

- execution/launchers/WORKING_BOT.py
- execution/workflows/junction_executor.py
- libs/skill_engine.py
- execution/system/hybrid_sync.py
- execution/ops/bots/LAUNCH_SYSTEM.py
- execution/system/task_manager.py
- execution/system/quality_gate.py
- libs/ralph_loop_engine.py

---

**Mercenary Standard**: Zero assumptions, maximum clarity.
