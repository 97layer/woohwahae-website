# Infrastructure Operations Skill

## Skill ID

`infra_ops_v1`

## Purpose

97layerOS의 하이브리드 인프라 (MacBook + Cloud)를 안정적으로 운영하며, Podman 기반 컨테이너 환경과 무료 클라우드 서비스를 최적화한다.

## Core Philosophy

> "Silent Guardian: 보이지 않게 작동하되, 항상 준비되어 있다"

### Quick Reference: macOS Podman Commands

**CRITICAL: ALWAYS use TMPDIR bypass on macOS**

```bash
# Method 1: Inline TMPDIR (for one-off commands)
export TMPDIR=/tmp && podman ps
export TMPDIR=/tmp && podman exec container_name command

# Method 2: Fix script (RECOMMENDED for regular use)
/Users/97layer/97layerOS/scripts/fix_podman_macos.sh health     # Check all containers
/Users/97layer/97layerOS/scripts/fix_podman_macos.sh ps        # List containers
/Users/97layer/97layerOS/scripts/fix_podman_macos.sh exec CONTAINER CMD

# Method 3: Set permanently (for shell sessions)
export TMPDIR=/tmp  # Add to ~/.zshrc for persistence
```

## Architecture Overview

### Hybrid Zero-Cost Model

```yaml
Local (MacBook):
  Role: Development + Heavy AI Processing
  Tools: Podman, Python venv
  Resources: 8-16GB RAM, M-series CPU
  Cost: $0 (owned hardware)

Cloud (Free Tier):
  GCP Cloud Run: API endpoints
  Cloudflare Workers: Edge functions
  Telegram Bot API: Message handling
  GitHub: Version control + CI/CD
  Cost: $0 (free tier limits)
```

### Service Distribution

```python
service_map = {
    "local": [
        "gemini_api",         # Gemini API calls
        "claude_api",         # Claude API calls
        "image_processing",   # Vision analysis
        "heavy_computation",  # Pattern recognition
        "development",        # Active development
    ],
    "cloud": [
        "telegram_webhook",   # 24/7 bot availability
        "api_gateway",        # Public endpoints
        "scheduled_tasks",    # Cron jobs
        "static_hosting",     # Documentation
    ]
}
```

## Rules

### 1. Container Management (Podman)

#### Podman vs Docker

```bash
# ✅ ALWAYS use Podman (rootless, daemon-less)
podman ps
podman build -t image_name .
podman run -d --name container_name image_name

# ❌ NEVER use Docker (requires root, daemon)
docker ps  # FORBIDDEN
```

#### macOS-Specific: TMPDIR Bypass (MANDATORY)

```bash
# CRITICAL: macOS sandbox blocks Podman's temp file creation
# SOLUTION: Set TMPDIR=/tmp for ALL Podman operations

# ✅ CORRECT: Use TMPDIR bypass
export TMPDIR=/tmp && podman ps
export TMPDIR=/tmp && podman exec container_name command

# ✅ BETTER: Use the fix script
/Users/97layer/97layerOS/scripts/fix_podman_macos.sh health
/Users/97layer/97layerOS/scripts/fix_podman_macos.sh exec container_name command

# ❌ WRONG: Direct Podman without TMPDIR (will fail on macOS)
podman ps  # ERROR: chmod /var/folders/... permission denied

# Why this works:
# - macOS sandbox restricts /var/folders/* access
# - /tmp is universally accessible
# - TMPDIR tells Podman where to create temporary files
# - Must be set for EVERY Podman invocation
```

#### Container-First Execution Model

```bash
# ALWAYS prefer container execution over local Python
# Benefits: Isolated, reproducible, no local environment issues

# ✅ CORRECT: Execute in container
export TMPDIR=/tmp && podman exec 97layer-workspace python3 script.py

# ✅ BETTER: Use fix script
/Users/97layer/97layerOS/scripts/fix_podman_macos.sh exec 97layer-workspace python3 script.py

# ❌ AVOID: Local execution (unless debugging)
python3 script.py  # Might have environment issues
```

#### Container Principles

```yaml
Stateless:
  - No data in containers
  - Mount volumes for persistence

Ephemeral:
  - Can be destroyed and recreated
  - Configuration in environment variables

Isolated:
  - One service per container
  - Clear boundaries
```

#### Podman Compose Example

```yaml
# podman-compose.yml
version: "3"

services:
  telegram_bot:
    image: 97layer/telegram-bot:latest
    environment:
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
      - GEMINI_API_KEY=${GEMINI_API_KEY}
    volumes:
      - ./knowledge:/app/knowledge:Z
    restart: unless-stopped

  api_server:
    image: 97layer/api:latest
    ports:
      - "8000:8000"
    environment:
      - ENVIRONMENT=production
    depends_on:
      - telegram_bot
    restart: unless-stopped
```

### 2. Environment Management

#### Virtual Environment Strategy

```python
# Project structure
97layerOS/
├── venv/              # Python virtual environment
│   ├── bin/
│   ├── lib/
│   └── pyvenv.cfg
├── requirements.txt   # Pinned dependencies
└── .env              # Secrets (gitignored)
```

#### Setup Workflow

```bash
# Create venv
python3 -m venv venv

# Activate
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Verify isolation
which python  # Should point to venv/bin/python
```

#### Dependency Management

```bash
# Pin current versions
pip freeze > requirements.txt

# Audit security
pip install safety
safety check

# Update dependencies (controlled)
pip install --upgrade package_name
pip freeze > requirements.txt
```

### 3. Resource Optimization

#### CPU/Memory Limits

```yaml
# Container resource constraints
resources:
  telegram_bot:
    memory: 512Mi      # Prevent memory leaks
    cpu: "0.5"         # 50% of one core

  api_server:
    memory: 1Gi
    cpu: "1.0"
```

#### Free Tier Limits

```python
GCP_LIMITS = {
    "cloud_run": {
        "requests": 2_000_000,  # per month
        "cpu_seconds": 360_000,
        "memory_gb_seconds": 180_000,
        "build_time": 120,      # minutes per day
    },
    "cloudflare": {
        "requests": 100_000,    # per day
        "cpu_time": 10,         # ms per request
    }
}

def check_quota_usage():
    """Monitor free tier usage"""
    if usage > 0.8 * limit:
        alert("Approaching quota limit")
        optimize_or_defer()
```

### 4. Secrets Management

#### Environment Variables

```bash
# .env file (NEVER commit)
TELEGRAM_BOT_TOKEN=xxxxx
GEMINI_API_KEY=xxxxx
CLAUDE_API_KEY=xxxxx
OPENAI_API_KEY=xxxxx

ENVIRONMENT=production
LOG_LEVEL=INFO
```

#### Secrets Loading

```python
from dotenv import load_dotenv
import os

load_dotenv()  # Load from .env

def get_secret(key):
    value = os.getenv(key)
    if not value:
        raise ValueError(f"Missing secret: {key}")
    return value

# Usage
TELEGRAM_TOKEN = get_secret("TELEGRAM_BOT_TOKEN")
```

#### Podman Secrets

```bash
# Create secret
echo "secret_value" | podman secret create secret_name -

# Use in container
podman run --secret secret_name image_name

# In container: read from /run/secrets/secret_name
```

### 5. Service Health Monitoring

#### Health Check Script

```python
#!/usr/bin/env python3
"""System health monitor"""

import subprocess
import requests
from datetime import datetime

class HealthMonitor:
    def check_services(self):
        """Check all critical services"""
        checks = {
            "podman": self.check_podman(),
            "telegram": self.check_telegram(),
            "api": self.check_api(),
            "disk": self.check_disk(),
            "memory": self.check_memory(),
        }
        return all(checks.values()), checks

    def check_podman(self):
        """Verify Podman is running"""
        result = subprocess.run(
            ["podman", "ps"],
            capture_output=True
        )
        return result.returncode == 0

    def check_telegram(self):
        """Verify Telegram bot responds"""
        try:
            url = "https://api.telegram.org/botTOKEN/getMe"
            response = requests.get(url, timeout=5)
            return response.status_code == 200
        except:
            return False

    def auto_heal(self, failed_check):
        """Attempt automatic recovery"""
        heal_actions = {
            "podman": self.restart_podman,
            "telegram": self.restart_telegram_bot,
            "api": self.restart_api_server,
            "disk": self.cleanup_disk,
            "memory": self.restart_memory_intensive,
        }

        if failed_check in heal_actions:
            heal_actions[failed_check]()
```

### 6. Backup & Recovery

#### Backup Strategy

```yaml
Automated Backups:
  Hourly:
    - knowledge/system/task_board.json
    - knowledge/system/sync_state.json

  Daily:
    - knowledge/ (full)
    - .env (encrypted)

  Weekly:
    - Complete system snapshot
    - Git repository

  Monthly:
    - Full archive to cloud
```

#### Backup Script

```python
#!/usr/bin/env python3
"""Automated backup system"""

import shutil
from pathlib import Path
from datetime import datetime

class BackupManager:
    BACKUP_ROOT = Path("/Users/97layer/.backups")

    def hourly_backup(self):
        """Critical files only"""
        timestamp = datetime.now().strftime("%Y%m%d_%H0000")
        backup_dir = self.BACKUP_ROOT / "hourly" / timestamp
        backup_dir.mkdir(parents=True, exist_ok=True)

        critical_files = [
            "knowledge/system/task_board.json",
            "knowledge/system/sync_state.json",
        ]

        for file in critical_files:
            src = Path(file)
            dst = backup_dir / src.name
            shutil.copy2(src, dst)

    def daily_backup(self):
        """Full knowledge base"""
        timestamp = datetime.now().strftime("%Y%m%d")
        backup_dir = self.BACKUP_ROOT / "daily" / timestamp

        shutil.copytree("knowledge", backup_dir)

    def restore(self, backup_path):
        """Restore from backup"""
        if not Path(backup_path).exists():
            raise FileNotFoundError(f"Backup not found: {backup_path}")

        # Safety: create pre-restore backup
        self.emergency_backup()

        # Restore
        shutil.rmtree("knowledge")
        shutil.copytree(backup_path, "knowledge")
```

### 7. Deployment Protocol

#### Pre-Deployment Checks

```python
def pre_deploy_checks():
    """Run before any deployment"""
    checks = [
        check_git_clean(),
        check_tests_pass(),
        check_secrets_set(),
        check_dependencies_updated(),
        check_no_hardcoded_secrets(),
    ]

    if not all(checks):
        raise DeploymentBlockedError("Pre-checks failed")
```

#### Zero-Downtime Deployment

```bash
#!/bin/bash
# Blue-green deployment for containers

# 1. Build new version
podman build -t app:v2 .

# 2. Start new container (green)
podman run -d --name app_green app:v2

# 3. Health check
if health_check app_green; then
    # 4. Stop old container (blue)
    podman stop app_blue
    podman rm app_blue

    # 5. Rename green to blue
    podman rename app_green app_blue
else
    # Rollback
    podman stop app_green
    podman rm app_green
    echo "Deployment failed, rollback complete"
fi
```

### 8. Log Management

#### Logging Strategy

```python
import logging
from logging.handlers import RotatingFileHandler

def setup_logging():
    """Structured logging setup"""

    # File handler (rotating)
    file_handler = RotatingFileHandler(
        "knowledge/logs/system.log",
        maxBytes=10_000_000,  # 10MB
        backupCount=5
    )

    # Format
    formatter = logging.Formatter(
        '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    file_handler.setFormatter(formatter)

    # Root logger
    logger = logging.getLogger()
    logger.setLevel(logging.INFO)
    logger.addHandler(file_handler)

    return logger
```

#### Log Analysis

```bash
# Recent errors
tail -f knowledge/logs/system.log | grep ERROR

# Error patterns
grep ERROR knowledge/logs/system.log | cut -d'-' -f4 | sort | uniq -c

# Performance issues
grep "slow_query" knowledge/logs/system.log
```

## Validation Criteria

### System Health Checks

- [ ] **All services running**: `podman ps` shows all containers
- [ ] **No memory leaks**: Memory usage < 80%
- [ ] **Disk space**: > 10GB free
- [ ] **API responsiveness**: < 500ms p95
- [ ] **Error rate**: < 1%

### Security Checks

- [ ] **No secrets in code**: `git grep -i "api_key"` returns nothing
- [ ] **Environment variables set**: All required vars present
- [ ] **Secrets rotated**: Last rotation < 90 days
- [ ] **Dependencies updated**: No critical CVEs

### Backup Validation

- [ ] **Backup exists**: Most recent < 24h old
- [ ] **Backup integrity**: Can be restored successfully
- [ ] **Backup size**: Within expected range

## Tools & Scripts

### macOS Podman Fix Script (NEW - MANDATORY)

```bash
# Location: /Users/97layer/97layerOS/scripts/fix_podman_macos.sh

# Check all container health
/Users/97layer/97layerOS/scripts/fix_podman_macos.sh health

# Output:
# === Container Health Check ===
# Checking 97layer-workspace...
#   Status: running
#   Started: 2026-02-15 19:42:29
#   Exec: OK
#
# Checking 97layer-snapshot...
#   Status: running
#   Started: 2026-02-15 19:50:00
#   Exec: OK
#
# === Summary ===
# Healthy: 4
# Unhealthy/Missing: 1

# List containers
/Users/97layer/97layerOS/scripts/fix_podman_macos.sh ps

# Execute command in container
/Users/97layer/97layerOS/scripts/fix_podman_macos.sh exec 97layer-workspace python3 --version

# Restart container
/Users/97layer/97layerOS/scripts/fix_podman_macos.sh restart 97layer-workspace
```

### Health Check

```bash
# System-wide health check
python execution/ops/health_check.py

# Output:
# ✓ Podman: Running (3 containers)
# ✓ Telegram: Responsive
# ✓ API: Healthy (200 OK)
# ✓ Disk: 45GB free (OK)
# ✓ Memory: 4.2GB/16GB (OK)
# Overall: HEALTHY
```

### Resource Monitor

```bash
# Real-time resource monitoring
python execution/ops/resource_monitor.py --interval 5

# Output (every 5 seconds):
# CPU: 32% | Memory: 6.1GB | Disk I/O: 12 MB/s
```

### Deployment Script

```bash
# Deploy with safety checks
python execution/ops/deploy.py --environment production

# Steps:
# 1. Pre-deployment checks
# 2. Build new containers
# 3. Run tests
# 4. Blue-green swap
# 5. Post-deployment validation
```

### Backup & Restore

```bash
# Create backup
python execution/ops/backup.py --type full

# List backups
python execution/ops/backup.py --list

# Restore backup
python execution/ops/backup.py --restore /path/to/backup
```

## Integration Points

### For TD (Technical Director)

```python
from libs.skill_loader import SkillLoader

infra = SkillLoader.load("infra_ops_v1")

# Pre-task infrastructure check
if not infra.health_check()['healthy']:
    infra.auto_heal()

# Deploy new service
infra.deploy_container("new_service")

# Monitor resources
metrics = infra.get_resource_metrics()
```

### For Quality Gate

```python
def infra_quality_gate():
    """Infrastructure validation before deploy"""

    infra = SkillLoader.load("infra_ops_v1")

    checks = {
        "health": infra.health_check(),
        "secrets": infra.validate_secrets(),
        "resources": infra.check_resource_availability(),
        "backups": infra.verify_recent_backup(),
    }

    return all(check['passed'] for check in checks.values())
```

## Common Issues & Solutions

### Issue 0: macOS Podman Permission Errors (MOST COMMON)

```bash
# Symptom:
# "chmod /var/folders/xx/xx/xx/libpod_tmp_xxx: operation not permitted"

# Root Cause:
# macOS sandbox blocks Podman from creating temp files in /var/folders/*

# Solution 1: Use TMPDIR bypass (IMMEDIATE FIX)
export TMPDIR=/tmp && podman ps
export TMPDIR=/tmp && podman exec container_name command

# Solution 2: Use fix script (RECOMMENDED)
/Users/97layer/97layerOS/scripts/fix_podman_macos.sh health
/Users/97layer/97layerOS/scripts/fix_podman_macos.sh exec 97layer-workspace python3 --version

# Solution 3: Set permanently in shell profile
echo 'export TMPDIR=/tmp' >> ~/.zshrc
source ~/.zshrc

# Verification:
echo $TMPDIR  # Should show: /tmp
export TMPDIR=/tmp && podman ps  # Should work without errors
```

### Issue 1: Container Won't Start

```bash
# Diagnosis
export TMPDIR=/tmp && podman logs container_name
export TMPDIR=/tmp && podman inspect container_name

# Common fixes
export TMPDIR=/tmp && podman restart container_name
export TMPDIR=/tmp && podman rm container_name && podman run ...

# Nuclear option
export TMPDIR=/tmp && podman system prune -a
```

### Issue 2: Memory Leak

```python
# Detect
df = get_memory_usage_over_time()
if df['memory'].is_increasing():
    alert("Memory leak detected")

# Fix
restart_container("leaky_service")
investigate_logs()
```

### Issue 3: Out of Disk Space

```bash
# Quick cleanup
python execution/ops/cleanup.py --aggressive

# Manual cleanup
podman system prune -a
rm -rf knowledge/archive/202401  # Old archives
```

### Issue 4: API Key Expired

```python
def handle_api_error(error):
    if "invalid_api_key" in str(error):
        alert_user("API key needs rotation")
        use_backup_api_key()
        schedule_key_rotation()
```

## Emergency Procedures

### DEFCON 1: System Down

```bash
# 1. Stop all services
podman stop --all

# 2. Restore from backup
python execution/ops/backup.py --restore latest

# 3. Restart services
podman start --all

# 4. Verify health
python execution/ops/health_check.py
```

### DEFCON 2: Data Corruption

```bash
# 1. Stop writes
podman stop $(podman ps -q)

# 2. Assess damage
python execution/ops/integrity_check.py

# 3. Restore affected files
python execution/ops/backup.py --restore-selective

# 4. Resume operations
podman start --all
```

### DEFCON 3: Security Breach

```bash
# 1. Rotate all secrets immediately
python execution/ops/rotate_secrets.py --all --force

# 2. Audit access logs
grep "unauthorized" knowledge/logs/*.log

# 3. Rebuild containers
podman build --no-cache -t app:latest .

# 4. Update security measures
python execution/ops/security_hardening.py
```

## Performance Optimization

### Container Optimization

```dockerfile
# Multi-stage build (reduce image size)
FROM python:3.11-slim AS builder
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

FROM python:3.11-slim
WORKDIR /app
COPY --from=builder /usr/local/lib/python3.11/site-packages /usr/local/lib/python3.11/site-packages
COPY . .

CMD ["python", "main.py"]
```

### Resource Tuning

```yaml
# Podman resource limits
services:
  high_cpu_task:
    cpus: "2.0"
    memory: 2Gi
    pids_limit: 100

  low_priority:
    cpus: "0.25"
    memory: 256Mi
    cpu_shares: 512  # Lower priority
```

## Version History

- **v1.0** (2026-02-15): Initial skill creation
  - Podman container management
  - Hybrid cloud architecture
  - Zero-cost optimization
  - Health monitoring & auto-healing
  - Backup & recovery procedures

---

> "인프라는 보이지 않아야 완벽하다. 문제 없이 작동할 때 그 존재를 잊게 만들어라." — 97layerOS
