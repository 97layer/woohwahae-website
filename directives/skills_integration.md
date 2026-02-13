# Skills Integration Directive

## Overview

Skills are now fully integrated into 97layerOS as first-class execution units. This document describes the architecture, usage patterns, and how to work with skills.

## Architecture

### Components

1. **SkillEngine** ([libs/skill_engine.py](libs/skill_engine.py))
   - Loads skill definitions from `skills/*/SKILL.md`
   - Parses YAML frontmatter for skill metadata
   - Orchestrates skill execution
   - Routes to appropriate executors based on skill ID

2. **AgentRouter** ([libs/agent_router.py](libs/agent_router.py))
   - Integrates SkillEngine for pattern-based routing
   - Detects URLs and patterns that trigger skills
   - Routes to skills BEFORE agent routing (priority routing)
   - Returns `SKILL:<skill_id>` marker when skill is detected

3. **Technical Daemon** ([execution/technical_daemon.py](execution/technical_daemon.py))
   - Executes skill tasks via `_handle_skill_execution()`
   - Supports skill invocation as part of rituals or ad-hoc tasks
   - Task type: `"SKILL"` with `skill_id` and `context` fields

4. **Core Config** ([libs/core_config.py](libs/core_config.py))
   - SKILLS_STRUCTURE registry maps skill IDs to executors
   - Defines trigger patterns and output paths per skill

## Skill Definition Format

Skills are defined in `skills/<skill_name>/SKILL.md` with YAML frontmatter:

```markdown
---
id: skill-001
name: Unified Input Protocol (UIP)
description: Transforms external signals into 97layerOS ontology assets
version: 1.0.0
target_agents: ["Gemini_Web", "Cursor_Agent", "The_Curator"]
tools:
  - youtube_parser
  - ontology_transform
---

# Skill Documentation

Markdown description of the skill workflow, constraints, and usage.
```

## Current Skills

### 1. Unified Input Protocol (UIP)
- **ID**: `skill-001`
- **Purpose**: Converts YouTube links to raw signals in `knowledge/raw_signals/`
- **Trigger**: YouTube URL patterns (`youtube.com`, `youtu.be`)
- **Workflow**:
  1. Detect YouTube URL
  2. Execute `execution/youtube_parser.py` to extract metadata
  3. Generate structured markdown file with YAML frontmatter
  4. Save to `knowledge/raw_signals/rs-<timestamp>_youtube_<title>.md`

### 2. Signal Capture
- **ID**: `signal_capture`
- **Purpose**: Captures external web signals to inbox
- **Status**: Placeholder (not implemented)

### 3. Data Curation
- **ID**: `data_curation`
- **Purpose**: Transforms raw signals into curated knowledge
- **Status**: Placeholder (not implemented)

### 4. Instagram Content Curator
- **ID**: `instagram_content_curator`
- **Purpose**: Processes Instagram content
- **Trigger**: Instagram URL patterns
- **Status**: Placeholder (not implemented)

### 5. Intelligence Backup
- **ID**: `intelligence_backup`
- **Purpose**: System knowledge backup
- **Status**: Defined

### 6. Infrastructure Sentinel
- **ID**: `infrastructure_sentinel`
- **Purpose**: System health monitoring
- **Status**: Defined

## Usage Patterns

### Automatic Detection (Via AgentRouter)

When a message contains a skill trigger pattern, the AgentRouter automatically routes to the skill:

```python
from libs.agent_router import AgentRouter

router = AgentRouter()
route = router.route("https://youtube.com/watch?v=abc123")
# Returns: "SKILL:skill-001"
```

### Manual Execution

Execute a skill directly via SkillEngine:

```python
from libs.skill_engine import SkillEngine

engine = SkillEngine()
result = engine.execute_skill("skill-001", {
    "url": "https://youtube.com/watch?v=abc123"
})

# Result structure:
# {
#   "status": "success" | "error",
#   "message": "...",
#   "output_file": "/path/to/output.md",
#   "signal_id": "rs-123456789"
# }
```

### Via Technical Daemon

Create a skill task in `task_status.json`:

```json
{
  "id": "task_uip_001",
  "type": "SKILL",
  "skill_id": "skill-001",
  "context": {
    "url": "https://youtube.com/watch?v=abc123"
  }
}
```

## Adding New Skills

### 1. Create Skill Directory

```bash
mkdir -p skills/my_new_skill
```

### 2. Create SKILL.md

```markdown
---
id: my_new_skill
name: My New Skill
description: Description of what this skill does
version: 1.0.0
target_agents: ["Technical_Director"]
tools:
  - execution_script_name
---

# My New Skill

Workflow documentation here.
```

### 3. Implement Executor

Add executor method to `SkillEngine._execute_<skill_id>()`:

```python
def _execute_my_new_skill(self, context: Dict[str, Any]) -> Dict[str, Any]:
    # Implementation here
    return {
        "status": "success",
        "message": "Skill executed",
        "output": ...
    }
```

### 4. Register in execute_skill()

```python
elif skill_id == "my_new_skill":
    return self._execute_my_new_skill(context)
```

### 5. Update SKILLS_STRUCTURE (Optional)

Add entry to `libs/core_config.py` if needed:

```python
"my_new_skill": {
    "name": "My New Skill",
    "executor": ["execution/my_script.py"],
    "trigger_patterns": [r'some_pattern'],
    "output_path": "knowledge/output/"
}
```

## Testing

Run integration tests:

```bash
python3 tests/test_skill_integration.py
```

Test specific skill execution:

```bash
python3 tests/test_uip_execution.py
```

## Benefits of Skills Architecture

1. **Token Efficiency**: Skills avoid loading full agent personas for deterministic tasks
2. **Modularity**: Each skill is self-contained and independently updateable
3. **Self-Annealing**: Skills can be improved without touching core orchestration
4. **Cross-Channel**: Same skill runs in Telegram, Gemini Web, Cursor, or daemon
5. **Declarative**: Skill definitions are documentation + execution in one file

## Next Steps

1. Implement remaining placeholder skills (signal_capture, data_curation, instagram_curator)
2. Add skill execution metrics and logging
3. Create skill dependency graph for composite workflows
4. Integrate skills with ritual system for scheduled execution
5. Add skill versioning and rollback capability

## References

- UIP Protocol: [directives/uip_protocol.md](directives/uip_protocol.md)
- Skill Definitions: [skills/](skills/)
- Agent Router: [libs/agent_router.py](libs/agent_router.py)
- Skill Engine: [libs/skill_engine.py](libs/skill_engine.py)
