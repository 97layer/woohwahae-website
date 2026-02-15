# Ralph Loop Self-Validation Engine Integration Guide

## Overview

The Ralph Loop Engine implements STAP (Stop, Task, Assess, Process) validation to prevent false success reports. Agents must pass all validation checks before the `[SYSTEM_STABLE]` token is issued.

## Core Concept

**Problem**: Agents often report task completion without proper validation, leading to broken systems.

**Solution**: Ralph Loop forces agents to self-validate through multiple checks and automatic retries before allowing completion.

## STAP Process

### 1. STOP
Pause before claiming success. No premature completion reports.

### 2. TASK
Review what was supposed to be accomplished and validation requirements.

### 3. ASSESS
Run comprehensive validation checks:
- Build verification (npm/Python)
- Skill compliance (directive adherence)
- Log analysis (error detection)
- Syntax validation
- Banned pattern detection
- Output verification

### 4. PROCESS
Determine outcome:
- **All checks pass** → Issue `[SYSTEM_STABLE]` token
- **Any check fails** → Retry (max 5 attempts)
- **Max retries exceeded** → Report failure, no token

## Usage

### From Python

```python
from libs.ralph_loop_engine import RalphLoopEngine

engine = RalphLoopEngine()

success, report = engine.validate_task(
    task_id="TASK-001",
    agent_id="TD",
    task_type="backend",  # frontend, backend, system, content
    files_modified=[
        "/Users/97layer/97layerOS/execution/new_feature.py",
        "/Users/97layer/97layerOS/libs/helper.py"
    ],
    output_path="/Users/97layer/97layerOS/knowledge/output.md"  # Optional
)

if success:
    print("Task validated - [SYSTEM_STABLE] token issued")
else:
    print("Task failed validation")
    print(f"Failed checks: {report['failed_checks']}")
```

### From Command Line

```bash
# Validate a task
python libs/ralph_loop_engine.py validate \
    TASK-001 \
    TD \
    backend \
    /path/to/file1.py \
    /path/to/file2.py

# Check statistics
python libs/ralph_loop_engine.py stats
```

### Integration with Task Manager

```python
from execution.system.task_manager import TaskManager
from libs.ralph_loop_engine import RalphLoopEngine

manager = TaskManager()
ralph = RalphLoopEngine()

# Agent claims task
manager.claim_task("TD", "TASK-001")

# Agent performs work...
# ...

# Validate before marking complete
success, report = ralph.validate_task(
    task_id="TASK-001",
    agent_id="TD",
    task_type="backend",
    files_modified=["execution/new_file.py"]
)

if success:
    manager.update_status("TD", "TASK-001", "completed")
else:
    manager.update_status("TD", "TASK-001", "failed")
```

## Configuration

Edit `/Users/97layer/97layerOS/core/ralph_loop_config.yaml`:

```yaml
# Maximum retry attempts
max_iterations: 5

# Enable/disable specific checks
verification_rules:
  build: true
  skill_compliance: true
  logs: true
  syntax: true

# Add banned patterns
banned_patterns:
  - 'TODO\s*:'
  - 'hardcoded.*token'

# Completion token
completion_token: '[SYSTEM_STABLE]'
```

## Validation Checks

### 1. Build Verification
- **Frontend**: `npm run build` must succeed
- **Backend**: All Python files must compile
- **System**: Core services must be functional

### 2. Skill Compliance
- Type hints present (Python)
- No emojis in code (Zero Noise protocol)
- Proper encoding (utf-8)
- Directive adherence

### 3. Log Analysis
- Scan recent logs for error keywords
- Check notification logs
- Verify system health

### 4. Syntax Validation
- Python: `py_compile` check
- JSON: Parse validation
- YAML: Safe load validation

### 5. Banned Pattern Detection
- No TODO/FIXME markers
- No hardcoded credentials
- No debug statements (print, console.log)
- No placeholder text

### 6. Output Verification (if applicable)
- File exists
- File not empty
- No placeholder text
- Valid format

## Retry Logic

When validation fails:
1. Log the failure
2. Wait 2 seconds
3. Retry validation
4. Repeat up to 5 times total
5. If all retries fail, log final failure and exit without token

## Logs

All validation runs are logged to: `/Users/97layer/97layerOS/knowledge/system/ralph_loop.json`

```json
{
  "total_runs": 42,
  "successful_runs": 38,
  "failed_runs": 4,
  "runs": [
    {
      "task_id": "TASK-001",
      "agent_id": "TD",
      "task_type": "backend",
      "started_at": "2026-02-15T10:30:00",
      "iterations": [
        {
          "iteration": 1,
          "checks": {
            "build": {"passed": true},
            "syntax": {"passed": true},
            "banned_patterns": {"passed": true}
          },
          "failed_checks": [],
          "passed": true
        }
      ],
      "final_status": "success",
      "completion_token_issued": true
    }
  ]
}
```

## Best Practices

### For Agents

1. **Always use Ralph Loop before reporting completion**
   ```python
   # ❌ BAD
   manager.update_status("TD", "TASK-001", "completed")

   # ✅ GOOD
   success, _ = ralph.validate_task(...)
   if success:
       manager.update_status("TD", "TASK-001", "completed")
   ```

2. **Fix issues before retrying**
   - Ralph Loop will auto-retry, but fix root causes
   - Review failed checks in the report
   - Don't rely on retries to mask problems

3. **Monitor validation logs**
   ```bash
   python libs/ralph_loop_engine.py stats
   ```

### For System Operators

1. **Review failure patterns**
   - Check `ralph_loop.json` for recurring failures
   - Identify problematic agents or task types
   - Adjust configuration if needed

2. **Tune banned patterns**
   - Add project-specific patterns to config
   - Keep patterns relevant and specific

3. **Set appropriate timeouts**
   - Default: 300 seconds per iteration
   - Adjust based on task complexity

## Testing

Run the test suite:

```bash
python tests/test_ralph_loop.py
```

Tests cover:
- ✅ Success scenario
- ❌ Syntax failure detection
- 🚫 Banned pattern detection
- 🔄 Retry logic
- 🎯 Completion token issuance
- 📊 Statistics tracking

## Troubleshooting

### Issue: Validation always fails on build

**Solution**: Check if build commands are correct for your environment.
```yaml
# Edit config to disable build check if not applicable
verification_rules:
  build: false
```

### Issue: Too many false positives on banned patterns

**Solution**: Refine patterns in config:
```yaml
banned_patterns:
  - '(?<!#\s)TODO\s*:'  # Only flag TODOs not in comments
```

### Issue: Validation takes too long

**Solution**: Increase timeout or reduce file count:
```yaml
timeout_seconds: 600  # 10 minutes
```

### Issue: [SYSTEM_STABLE] never issued

**Solution**: Check logs to see which checks are failing:
```bash
python libs/ralph_loop_engine.py stats
# Review recent_runs to identify failing checks
```

## Integration Points

### With Task Manager
- `execution/system/task_manager.py`
- Call Ralph Loop before marking tasks complete

### With Quality Gate
- `execution/system/quality_gate.py`
- Ralph Loop can use Quality Gate for pre/post checks

### With Shadow Logic
- `execution/system/shadow_logic.py`
- Peer review can trigger Ralph Loop validation

## Success Criteria

A task is considered validated when:
1. ✅ All enabled verification rules pass
2. ✅ No banned patterns detected
3. ✅ No syntax errors
4. ✅ Logs clean of errors
5. ✅ Output valid (if applicable)
6. ✅ Build successful (if applicable)

**Only then** is `[SYSTEM_STABLE]` issued.

## Example: Full Workflow

```python
#!/usr/bin/env python3
"""
Example: Complete task validation workflow
"""

from execution.system.task_manager import TaskManager
from libs.ralph_loop_engine import RalphLoopEngine

def complete_task_safely(task_id: str, agent_id: str, task_type: str, files: list):
    """Complete a task with proper validation"""

    manager = TaskManager()
    ralph = RalphLoopEngine()

    # 1. Claim task
    if not manager.claim_task(agent_id, task_id):
        print("❌ Could not claim task")
        return False

    # 2. Perform work
    print("🔧 Performing task...")
    # ... do actual work ...

    # 3. Validate with Ralph Loop
    print("🔍 Validating with Ralph Loop...")
    success, report = ralph.validate_task(
        task_id=task_id,
        agent_id=agent_id,
        task_type=task_type,
        files_modified=files
    )

    # 4. Update task status based on validation
    if success:
        manager.update_status(agent_id, task_id, "completed")
        print("✅ Task completed and validated")
        return True
    else:
        manager.update_status(agent_id, task_id, "failed")
        print(f"❌ Task failed validation: {report['failed_checks']}")
        return False

# Usage
if __name__ == "__main__":
    complete_task_safely(
        task_id="TASK-001",
        agent_id="TD",
        task_type="backend",
        files=["execution/new_feature.py"]
    )
```

## Summary

Ralph Loop ensures system stability by:
- 🛑 Stopping premature success reports
- 📋 Reviewing task requirements
- 🔍 Assessing with comprehensive checks
- ⚙️  Processing with auto-retry or token issuance

**No more false positives. No more broken systems. Only validated completions.**
