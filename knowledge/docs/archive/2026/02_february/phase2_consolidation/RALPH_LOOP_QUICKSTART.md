# Ralph Loop Quick Start

## What is Ralph Loop?

Ralph Loop is a self-validation engine that prevents agents from reporting false success. It implements STAP (Stop, Task, Assess, Process) validation with automatic retry.

**Core Rule**: No `[SYSTEM_STABLE]` token until ALL checks pass.

## 30-Second Setup

```bash
# 1. Install dependencies (if needed)
pip install pyyaml

# 2. Test the engine
python tests/test_ralph_loop.py

# 3. Check stats
python libs/ralph_loop_engine.py stats
```

## Basic Usage

### Python API

```python
from libs.ralph_loop_engine import RalphLoopEngine

engine = RalphLoopEngine()

# Validate a task
success, report = engine.validate_task(
    task_id="TASK-001",
    agent_id="TD",
    task_type="backend",
    files_modified=["execution/script.py"]
)

if success:
    print("✅ Validated - [SYSTEM_STABLE]")
else:
    print(f"❌ Failed: {report['failed_checks']}")
```

### Command Line

```bash
# Validate files
python libs/ralph_loop_engine.py validate \
    TASK-001 TD backend \
    execution/script.py libs/helper.py

# View statistics
python libs/ralph_loop_engine.py stats
```

## What Gets Checked?

1. **Build** - npm/Python compilation
2. **Syntax** - All files must be valid
3. **Banned Patterns** - No TODO, FIXME, hardcoded secrets
4. **Logs** - No recent errors
5. **Skill Compliance** - Follows directives
6. **Output** - File exists and is valid

## Retry Behavior

- ❌ Check fails → Wait 2s → Retry (max 5 attempts)
- ✅ All pass → Issue `[SYSTEM_STABLE]` token
- ❌ Max retries → Log failure, no token

## Configuration

Edit `core/ralph_loop_config.yaml`:

```yaml
max_iterations: 5          # Max retry attempts

verification_rules:
  build: true              # Enable build check
  skill_compliance: true   # Enable directive check
  logs: true              # Enable log check
  syntax: true            # Enable syntax check

completion_token: '[SYSTEM_STABLE]'
```

## Common Patterns

### Pattern 1: Task Manager Integration

```python
from execution.system.task_manager import TaskManager
from libs.ralph_loop_engine import RalphLoopEngine

manager = TaskManager()
ralph = RalphLoopEngine()

# Before completing task
success, _ = ralph.validate_task(...)
if success:
    manager.update_status("TD", "TASK-001", "completed")
```

### Pattern 2: Pre-commit Hook

```python
# Add to pre-commit hook
from libs.ralph_loop_engine import RalphLoopEngine
import subprocess

# Get changed files
result = subprocess.run(["git", "diff", "--name-only", "--cached"],
                       capture_output=True, text=True)
files = result.stdout.strip().split('\n')

# Validate
engine = RalphLoopEngine()
success, _ = engine.validate_task(
    task_id="PRE-COMMIT",
    agent_id="HOOK",
    task_type="system",
    files_modified=files
)

sys.exit(0 if success else 1)
```

### Pattern 3: CI/CD Pipeline

```bash
#!/bin/bash
# In CI/CD script

# Run Ralph Loop validation
python libs/ralph_loop_engine.py validate \
    CI-BUILD \
    CI \
    backend \
    $(git diff --name-only origin/main...HEAD)

if [ $? -eq 0 ]; then
    echo "✅ Validation passed - deploying..."
else
    echo "❌ Validation failed - blocking deployment"
    exit 1
fi
```

## Viewing Logs

```bash
# Quick stats
python libs/ralph_loop_engine.py stats

# Detailed log
cat knowledge/system/ralph_loop.json | jq .
```

## Troubleshooting

### Problem: Build check always fails
**Solution**: Disable if not applicable
```yaml
verification_rules:
  build: false
```

### Problem: Too many banned pattern alerts
**Solution**: Refine patterns in config
```yaml
banned_patterns:
  - '(?<!#\s)TODO\s*:'  # Only non-comment TODOs
```

### Problem: Validation timeout
**Solution**: Increase timeout
```yaml
timeout_seconds: 600  # 10 minutes
```

## Success Token

Only issued when ALL checks pass:

```
[SYSTEM_STABLE]
```

No token = validation failed or incomplete.

## Testing

```bash
# Run full test suite
python tests/test_ralph_loop.py

# Expected output:
# Test 1 PASSED: Success scenario
# Test 2 PASSED: Syntax failure detection
# Test 3 PASSED: Banned pattern detection
# Test 4 PASSED: Retry logic
# Test 5 PASSED: Completion token issuance
# Test 6 PASSED: Statistics tracking
```

## Key Principle

**Don't report success until Ralph Loop validates it.**

```python
# ❌ WRONG
do_work()
print("Task complete!")  # Premature claim

# ✅ RIGHT
do_work()
success, _ = ralph.validate_task(...)
if success:
    print("Task complete! [SYSTEM_STABLE]")
else:
    print("Validation failed - investigating")
```

## Next Steps

1. Read full documentation: `docs/RALPH_LOOP_INTEGRATION.md`
2. Review config options: `core/ralph_loop_config.yaml`
3. Integrate with your workflow
4. Monitor statistics regularly

## Support

- **Config**: `core/ralph_loop_config.yaml`
- **Engine**: `libs/ralph_loop_engine.py`
- **Tests**: `tests/test_ralph_loop.py`
- **Logs**: `knowledge/system/ralph_loop.json`
- **Docs**: `docs/RALPH_LOOP_INTEGRATION.md`
