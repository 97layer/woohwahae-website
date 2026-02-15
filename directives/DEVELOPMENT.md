# 🔧 DEVELOPMENT - 통합 개발 지침 v3.0

> **통합**: token_optimization_protocol + system_sop + infrastructure_sentinel + directive_lifecycle + skills_integration + data_asset_management + 97layerOS_Optimization_Directive + system_handshake + agent_instructions
> **버전**: 3.0
> **갱신**: 2026-02-15
> **철학**: "이 파일은 계속 성장한다. 파편화하지 말고 여기에 추가하라."

---

## 🎯 Development Philosophy

### Core Principle: Token-First Development
```python
# 모든 개발의 시작
def before_any_development():
    """토큰은 돈이다. 낭비는 죄다."""

    # Rule 1: Query Before Read
    search_first()

    # Rule 2: Cache Everything
    use_cache()

    # Rule 3: Batch Process
    batch_not_loop()

    return "60-80% token saved"
```

---

## 💰 Token Optimization Protocol

### Layer 1: Query Strategy

#### ❌ BAD Pattern (20,000+ tokens)
```python
# 전체 파일 읽기
content = Read("large_module.py")
# 전체 검색
all_results = search_everything()
# 반복 호출
for item in items:
    ai_call(item)
```

#### ✅ GOOD Pattern (1,500 tokens)
```python
# 1. Glob으로 후보 찾기
files = Glob("**/*target*.py")

# 2. Grep으로 정확한 위치
matches = Grep("specific_function", files[0])

# 3. Read with offset/limit
content = Read(file, offset=100, limit=20)

# 4. Batch 처리
results = ai_batch_call(items)
```

### Layer 2: Caching System

#### Implementation
```python
from functools import lru_cache
from execution.system.token_optimizer import cache_result

@cache_result(ttl_hours=24)
def expensive_operation(prompt):
    """24시간 캐시 - 같은 질문 반복 방지"""
    return ai_engine.generate(prompt)

@lru_cache(maxsize=100)
def frequent_lookup(key):
    """메모리 캐시 - 자주 쓰는 데이터"""
    return database.query(key)
```

#### Cache Strategy
```yaml
Cache Levels:
  L1_Memory: 100 items, TTL 1 hour
  L2_Disk: 1000 items, TTL 24 hours
  L3_Archive: Unlimited, TTL 7 days

Cache Keys:
  Pattern: {function}_{hash(params)}_{date}
  Example: generate_content_a3f4b2_20260215
```

### Layer 3: Batch Processing

```python
# ❌ Sequential (slow, expensive)
for i in range(100):
    result = api.call(data[i])

# ✅ Batch (fast, cheap)
results = api.batch_call(data[:100])

# ✅ Async Parallel (fastest)
async def parallel_process():
    tasks = [api.async_call(d) for d in data]
    return await asyncio.gather(*tasks)
```

---

## 🏗️ System Architecture

### 3-Layer Design
```
┌─────────────────────────────────────┐
│         Layer 1: Directives         │  ← 지침 (What)
├─────────────────────────────────────┤
│       Layer 2: Orchestration        │  ← 결정 (How)
├─────────────────────────────────────┤
│         Layer 3: Execution          │  ← 실행 (Do)
└─────────────────────────────────────┘
```

### File Organization
```
97layerOS/
├── directives/          # 5 core files only
│   ├── CORE.md
│   ├── IDENTITY.md
│   ├── OPERATIONS.md
│   ├── PUBLISHING.md
│   └── DEVELOPMENT.md   # This file
├── execution/
│   ├── system/          # Core systems
│   ├── api/            # API servers
│   ├── ops/            # Operations
│   ├── plans/          # PLAN-XXX.md (incremental)
│   └── utils/          # Utilities
├── knowledge/
│   ├── system/         # Single source configs
│   ├── magazines/      # Date-based outputs
│   └── archive/        # Historical records
└── libs/               # Shared libraries
```

---

## 🔄 Directive Lifecycle

### Version Control Philosophy
```python
class DirectiveManagement:
    """지시서는 성장한다, 파편화하지 않는다"""

    def update_directive(self, file_path, new_content):
        # 1. 백업은 Git이 담당
        git.add(file_path)
        git.commit(f"Before update: {reason}")

        # 2. 덮어쓰기로 업데이트
        file.write(new_content)  # Overwrite

        # 3. 새 버전 커밋
        git.add(file_path)
        git.commit(f"Updated v{version}: {changes}")

        # 4. 파일명은 그대로
        return file_path  # 같은 이름 유지
```

### Growth Pattern
```markdown
## Section Name
(Initial content from v1.0)

### Added in v1.1
(New insights)

### Enhanced in v1.2
(Improvements)

### Refined in v2.0
(Major update)
```

---

## 🛠️ Infrastructure Management

### System Health Monitoring
```python
class SystemGuardian:
    def __init__(self):
        self.checks = {
            "disk_space": lambda: check_disk() > 1_000_000_000,  # 1GB
            "memory": lambda: check_memory() > 500_000_000,      # 500MB
            "api_keys": lambda: validate_env_vars(),
            "git_clean": lambda: is_git_clean(),
            "services": lambda: check_services_running()
        }

    def health_check(self):
        for name, check in self.checks.items():
            if not check():
                self.auto_fix(name)
```

### Service Management
```bash
# Core Services
97layer-telegram.service    # Telegram bot
97layer-dashboard.service   # Web dashboard
97layer-api.service        # API server
97layer-shadow.service     # Shadow logic

# Management Commands
systemctl status 97layer-*
systemctl restart 97layer-telegram
journalctl -u 97layer-* -f
```

### Backup Strategy
```python
BACKUP_SCHEDULE = {
    "hourly": ["task_board.json"],           # Critical
    "daily": ["knowledge/system/*"],         # System state
    "weekly": ["knowledge/", "directives/"], # Full backup
    "monthly": ["*"]                         # Complete
}

def backup(schedule_type):
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    for path in BACKUP_SCHEDULE[schedule_type]:
        backup_path = f".backups/{schedule_type}/{timestamp}/"
        copy(path, backup_path)
```

---

## 🔌 Skills Integration

### Skill Architecture
```python
class Skill:
    """재사용 가능한 능력 단위"""

    def __init__(self, name):
        self.name = name
        self.path = f"libs/skills/{name}/"
        self.config = load_config(f"{self.path}/SKILL.md")

    def execute(self, params):
        # 1. Validate input
        self.validate(params)

        # 2. Run skill logic
        result = self.run(params)

        # 3. Cache result
        cache.set(self.cache_key(params), result)

        return result
```

### Available Skills
```yaml
Core Skills:
  signal_capture:     # 신호 수집
  data_curation:     # 데이터 정제
  content_creation:  # 콘텐츠 생성
  visual_selection:  # 이미지 선택
  publish_check:     # 퍼블리싱 검증

Utility Skills:
  instagram_publisher:  # 인스타 발행
  telegram_responder:  # 텔레그램 응답
  backup_manager:      # 백업 관리
  token_optimizer:     # 토큰 최적화
```

---

## 📊 Data Asset Management

### Asset Types
```python
ASSET_TYPES = {
    "raw_signals": {
        "path": "knowledge/raw_signals/",
        "format": "rs-{id}_{source}.md",
        "retention": "30 days"
    },
    "patterns": {
        "path": "knowledge/patterns/",
        "format": "pattern_{date}.md",
        "retention": "90 days"
    },
    "content": {
        "path": "knowledge/content/",
        "format": "vol_{n}_{title}.md",
        "retention": "permanent"
    },
    "magazines": {
        "path": "knowledge/magazines/",
        "format": "{date}_{title}.md",
        "retention": "permanent"
    }
}
```

### Asset Lifecycle
```python
def asset_lifecycle(asset):
    # 1. Create
    asset_id = create_asset(asset)

    # 2. Process
    process_asset(asset_id)

    # 3. Archive (after retention)
    if age(asset) > retention_period:
        archive_asset(asset)

    # 4. Purge (if needed)
    if should_purge(asset):
        purge_asset(asset)
```

---

## 🔐 Security & Credentials

### Environment Variables
```bash
# .env file structure
TELEGRAM_BOT_TOKEN=xxx
GEMINI_API_KEY=xxx
CLAUDE_API_KEY=xxx
OPENAI_API_KEY=xxx
GOOGLE_CREDENTIALS_PATH=credentials.json
```

### Credential Management
```python
class CredentialManager:
    def __init__(self):
        self.load_env()
        self.validate_all()

    def get(self, key):
        value = os.getenv(key)
        if not value:
            raise CredentialError(f"Missing: {key}")
        return value

    def rotate(self, key):
        # Automatic rotation for security
        new_value = generate_new_token()
        update_env(key, new_value)
        reload_services()
```

---

## 🚀 Deployment Protocol

### Local Development
```bash
# Setup
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# Run
python execution/unified_system.py --dev
```

### Production Deployment
```bash
# Pre-deployment checks
python execution/system/quality_gate.py pre

# Deploy
git push origin main
ssh production "cd 97layerOS && git pull && systemctl restart 97layer-*"

# Post-deployment validation
python execution/system/quality_gate.py post production
```

### Rollback Procedure
```python
def emergency_rollback():
    # 1. Stop services
    stop_all_services()

    # 2. Restore backup
    restore_latest_backup()

    # 3. Restart services
    start_all_services()

    # 4. Validate
    run_health_checks()
```

---

## 🧪 Testing Strategy

### Test Levels
```python
# Unit Tests
def test_individual_functions():
    pytest tests/unit/

# Integration Tests
def test_component_interaction():
    pytest tests/integration/

# System Tests
def test_end_to_end():
    pytest tests/e2e/

# Smoke Tests (after deployment)
def smoke_test():
    assert api.health() == "ok"
    assert telegram.ping() == "pong"
    assert dashboard.status() == "running"
```

---

## 📈 Performance Optimization

### Optimization Targets
```yaml
Response Time:
  p50: < 100ms
  p95: < 500ms
  p99: < 1000ms

Resource Usage:
  CPU: < 50% average
  Memory: < 2GB
  Disk I/O: < 100 MB/s

Token Efficiency:
  Reduction: 60-80%
  Cache Hit: > 70%
  Batch Rate: > 50%
```

### Profiling Tools
```python
# Performance profiling
from cProfile import Profile
from memory_profiler import profile

@profile
def memory_intensive_function():
    # Memory usage tracked
    pass

with Profile() as pr:
    cpu_intensive_function()
    pr.print_stats()
```

---

## 🔄 Continuous Improvement

### Self-Annealing Process
```python
while True:
    try:
        execute()
    except Exception as e:
        # 1. Capture error
        error_id = log_error(e)

        # 2. Analyze root cause
        cause = analyze_error(error_id)

        # 3. Fix automatically if possible
        if can_auto_fix(cause):
            apply_fix(cause)

        # 4. Update this document
        update_development_md(lesson_learned)

        # 5. Retry
        continue
```

---

## 📚 Version History

- **v3.0** (2026-02-15): 대통합 - 9개 파일 → 1개
  - token_optimization_protocol.md
  - system_sop.md
  - infrastructure_sentinel.md
  - directive_lifecycle.md
  - skills_integration.md
  - data_asset_management.md
  - 97layerOS_Optimization_Directive.md
  - system_handshake.md
  - agent_instructions.md

- **v2.0** (2026-02-01): 토큰 최적화 추가
- **v1.0** (2026-01-15): 초기 개발 지침

---

## 🌱 Future Growth Areas

이 섹션에 새로운 개발 인사이트를 추가하세요:
- [ ] AI Model Management
- [ ] Multi-cloud Strategy
- [ ] Edge Computing Integration
- [ ] Real-time Processing Pipeline

---

> "파편화는 혼돈이다. 통합은 힘이다. 버전을 올리며 성장하라." — 97layerOS