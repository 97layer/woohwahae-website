# Self-Maintenance Protocol

**Version**: 1.0.0
**Created**: 2026-02-15
**Owner**: Technical Director
**Scope**: System-wide automatic maintenance and self-healing

---

## Objective

97layerOS 자가 유지보수 프로토콜. 시스템이 스스로 환경을 점검하고, 업데이트를 적용하며, 에러를 자동으로 수정하여 무중단 운영을 보장한다.

---

## Core Principles

1. **Silent Operation**: 사용자 개입 없이 자동으로 실행
2. **Fail-Safe**: 업데이트 실패 시 롤백 가능
3. **Logged Transparency**: 모든 자가 치유 활동을 로그로 기록
4. **Proactive Healing**: 문제가 반복되면 자동으로 수정 시도

---

## Components

### 1. Environment Integrity Check

**Script**: `execution/system/check_environment.py`

**Purpose**: 모든 실행 전 필수 패키지 및 시스템 요구사항 확인

**Checks**:
- Python 버전 (>= 3.9)
- 필수 패키지 설치 여부
- 필수 디렉토리 존재
- 환경변수 검증
- 시스템 리소스 (메모리, 디스크)
- 핵심 파일 존재

**Auto-Fix**:
- 누락 패키지 즉시 설치 (Silent Install)
- 누락 디렉토리 자동 생성

**Usage**:
```bash
# 모든 스크립트 실행 전
python3 execution/system/check_environment.py

# JSON 출력
python3 execution/system/check_environment.py --json

# 자동 설치 비활성화
python3 execution/system/check_environment.py --no-install
```

**Integration**:
```python
from execution.system.check_environment import EnvironmentChecker

checker = EnvironmentChecker(silent_install=True)
passed, report = checker.run_full_check()

if not passed:
    logger.error(f"Environment check failed: {report['issues']}")
    sys.exit(1)
```

---

### 2. Auto Update Protocol

**Script**: `execution/system/auto_update.py`

**Purpose**: Google Drive Core 폴더의 새 스크립트를 자동으로 시스템에 반영

**Process**:
1. Google Drive `Core/` 폴더 스캔
2. 신규/수정 파일 감지 (SHA256 해시 비교)
3. 기존 파일 `Archive/` 백업 (타임스탬프 포함)
4. 즉시 시스템 반영 (Apply)
5. 변경 이력 `knowledge/system/update_log.json`에 기록

**Schedule**: 매일 03:00 AM (Night Guard Ritual)

**Usage**:
```bash
# 업데이트 실행
python3 execution/system/auto_update.py

# Dry-run (스캔만)
python3 execution/system/auto_update.py --dry-run

# 이력 조회
python3 execution/system/auto_update.py --history
```

**Integration**:
```python
from execution.system.auto_update import AutoUpdater

updater = AutoUpdater()
result = updater.run_update_cycle()

if result["status"] == "COMPLETED":
    logger.info(f"Applied {result['success_count']} updates")
```

---

### 3. Self-Annealing Engine

**Script**: `execution/system/self_annealing_engine.py`

**Purpose**: 연산 에러 3회 반복 시 자동 코드 분석 및 수정 제안

**Process**:
1. 에러 추적 (`knowledge/system/error_tracking.json`)
2. 동일 에러 3회 반복 감지
3. 에러 패턴 분석 (FileNotFoundError, KeyError, etc.)
4. 수정 제안 생성 (try-except, 타입 체크, 재시도 로직)
5. `directives/self_annealing_log.md`에 사후 보고

**Auto-Fix Patterns**:
- `FileNotFoundError` → 파일 존재 확인 로직 추가
- `KeyError` → `.get()` 안전 접근 제안
- `IndexError` → 리스트 범위 체크
- `ConnectionError` → 재시도 로직 추가
- `TypeError` → 타입 체크 추가

**Usage**:
```bash
# 테스트 실행
python3 execution/system/self_annealing_engine.py
```

**Integration**:
```python
from execution.system.self_annealing_engine import SelfAnnealingEngine

engine = SelfAnnealingEngine(threshold=3)

# 재시도 메커니즘
result = engine.retry_with_annealing(
    risky_function,
    max_retries=3,
    context={"function": "risky_function", "file": __file__}
)

# 에러 핸들링
try:
    operation()
except Exception as e:
    annealing_result = engine.handle_error(
        error=e,
        context={"function": "operation", "file": __file__}
    )

    if annealing_result["needs_annealing"]:
        logger.warning("Self-annealing triggered - check directives/self_annealing_log.md")
```

---

### 4. Security Update Scheduler

**Script**: `execution/system/security_updater.py`

**Purpose**: 주 1회 종속성 라이브러리 보안 업데이트 및 무결성 테스트

**Process**:
1. 구버전 패키지 확인 (`pip list --outdated`)
2. 보안 취약점 검사 (`pip-audit`)
3. 취약점 패키지 우선 업데이트
4. `requirements.txt` 버전 갱신
5. 시스템 무결성 테스트 (import, environment, critical files)
6. 업데이트 이력 `knowledge/system/security_update_log.json`에 기록

**Schedule**: 매주 일요일 02:00 AM

**Usage**:
```bash
# 보안 업데이트 실행
python3 execution/system/security_updater.py

# 강제 실행 (주 1회 제한 무시)
python3 execution/system/security_updater.py --force

# 취약점만 업데이트 (구버전 제외)
python3 execution/system/security_updater.py --no-auto-update

# 체크만 (업데이트 안 함)
python3 execution/system/security_updater.py --check-only
```

**Integration**:
```python
from execution.system.security_updater import SecurityUpdater

updater = SecurityUpdater(auto_update=True)

# 업데이트 필요 여부 확인
should_update, reason = updater.should_update()

if should_update:
    result = updater.run_security_update()

    if result["status"] == "SUCCESS":
        logger.info("Security update completed")
```

---

### 5. Self-Healing Monitor (Existing)

**Script**: `execution/self_healing_monitor.py`

**Purpose**: 순환 참조, 메모리 누수, 에이전트 상태 실시간 모니터링

**Features**:
- 순환 참조 감지 및 차단
- 메모리 누수 방지 (임계값 500MB)
- 자동 캐시 정리
- 파일 시스템 건강도 체크
- 시스템 상태 `knowledge/system_state.json`에 기록

**Integration**: Technical Daemon에서 자동 실행 중

---

## Rituals Integration

자가 유지보수는 `libs/core_config.py:RITUALS_CONFIG`에 통합되어 자동 실행됩니다.

### Daily Rituals

**DAILY_AUTONOMOUS_DEVELOPMENT** (03:00 AM):
```python
{
    "trigger_hour": 3,
    "agent": "Technical_Director",
    "task_type": "AUTONOMOUS_DEV",
    "instruction": "현재 프로젝트 상황과 비전을 분석하여 시스템 고도화를 위한 차세대 태스크를 스스로 생성하십시오.",
    "council": False
}
```

**Action**:
1. Environment Check 실행
2. Auto Update 실행 (Google Drive 동기화)
3. System Health Check 실행

### Weekly Rituals

**WEEKLY_SECURITY_UPDATE** (일요일 02:00 AM):
```python
{
    "trigger_weekday": 6,  # Sunday
    "trigger_hour": 2,
    "agent": "Technical_Director",
    "task_type": "SECURITY_UPDATE",
    "instruction": "주간 보안 업데이트를 실행하십시오. (1) pip 패키지 취약점 검사, (2) 보안 업데이트 적용, (3) 시스템 무결성 테스트, (4) requirements.txt 갱신",
    "council": False
}
```

**Action**:
1. Security Updater 실행
2. 무결성 테스트 통과 확인
3. 실패 시 Telegram 알림

---

## Error Handling Workflow

### Standard Error Flow

```
1. Error Occurs
   ↓
2. Self-Annealing Engine Captures
   ↓
3. Error Count Incremented
   ↓
4. Count >= 3?
   ├─ No → Retry (exponential backoff)
   └─ Yes → Trigger Annealing
       ↓
5. Analyze Error Pattern
   ↓
6. Generate Fix Suggestions
   ↓
7. Log to directives/self_annealing_log.md
   ↓
8. Alert Technical Director (Telegram)
   ↓
9. Manual Review & Code Patch
```

### Critical Error Flow

```
1. Critical Error (CRITICAL severity)
   ↓
2. Self-Healing Monitor Detects
   ↓
3. Immediate Alert (Agent Hub Broadcast)
   ↓
4. Attempt Auto-Fix:
   - Missing Directory → Create
   - Cache Overflow → Clear
   - Memory Leak → GC + Cache Clear
   ↓
5. Log to knowledge/health_log.json
   ↓
6. If Auto-Fix Fails → Escalate to Admin
```

---

## Logging Structure

### 1. Update Log
**Path**: `knowledge/system/update_log.json`

**Format**:
```json
[
  {
    "type": "MODIFIED",
    "relative_path": "execution/instagram_generator.py",
    "drive_hash": "a1b2c3...",
    "local_hash": "d4e5f6...",
    "archive_path": ".tmp/archive/instagram_generator_20260215_030000.py",
    "timestamp": "2026-02-15T03:00:00",
    "status": "SUCCESS"
  }
]
```

---

### 2. Error Tracking
**Path**: `knowledge/system/error_tracking.json`

**Format**:
```json
{
  "error_counts": {
    "FileNotFoundError:fetch_trends": 3,
    "ConnectionError:api_call": 2
  },
  "error_history": [
    {
      "timestamp": "2026-02-15T10:30:00",
      "error_key": "FileNotFoundError:fetch_trends",
      "error_type": "FileNotFoundError",
      "error_message": "[Errno 2] No such file or directory: '/path/to/trends.json'",
      "count": 3,
      "context": {
        "function": "fetch_trends",
        "file": "execution/trend_crawler.py",
        "line": 42
      }
    }
  ]
}
```

---

### 3. Self-Annealing Log
**Path**: `directives/self_annealing_log.md`

**Format**:
```markdown
# Self-Annealing Log

자가 치유 활동 기록

---

## [2026-02-15 10:30:00] FileNotFoundError

**에러 메시지**: [Errno 2] No such file or directory: '/path/to/trends.json'

**컨텍스트**:
- Function: fetch_trends
- File: execution/trend_crawler.py
- Line: 42

**발생 횟수**: 3

**자동 수정 제안**:

1. **FILE_CHECK**: 파일 존재 확인 로직 추가

\`\`\`python
if not Path(file_path).exists():
    logger.error(f'File not found: {file_path}')
    return None
\`\`\`

**상태**: 분석 완료 (수동 검토 필요)

---
```

---

### 4. Security Update Log
**Path**: `knowledge/system/security_update_log.json`

**Format**:
```json
[
  {
    "timestamp": "2026-02-15T02:00:00",
    "status": "SUCCESS",
    "outdated_packages": 5,
    "vulnerabilities": 2,
    "packages_to_update": 2,
    "update_results": {
      "requests": true,
      "urllib3": true
    },
    "integrity_test": {
      "status": "PASS",
      "tests": {
        "import_test": true,
        "environment_check": true,
        "critical_files": true
      }
    }
  }
]
```

---

## Manual Interventions

### When to Intervene

자동 치유가 실패하거나 수동 검토가 필요한 경우:

1. **Self-Annealing Log에 3회 이상 동일 에러**
   - Action: `directives/self_annealing_log.md` 확인 후 코드 수정

2. **Security Update 실패**
   - Action: `knowledge/system/security_update_log.json` 확인 후 수동 업데이트

3. **Critical Alert (Agent Hub Broadcast)**
   - Action: 즉시 시스템 상태 확인 (`knowledge/system_state.json`)

4. **Integrity Test 실패**
   - Action: Environment Check 실행 후 누락 파일/패키지 복구

---

## Testing

### Manual Test Commands

```bash
# 1. Environment Check
python3 execution/system/check_environment.py

# 2. Auto Update (Dry-run)
python3 execution/system/auto_update.py --dry-run

# 3. Security Update (Check-only)
python3 execution/system/security_updater.py --check-only

# 4. Self-Annealing Engine Test
python3 execution/system/self_annealing_engine.py

# 5. Self-Healing Monitor
python3 execution/self_healing_monitor.py
```

### Expected Outputs

- ✅ All checks PASS
- 📊 Logs updated in `knowledge/system/`
- 🔒 No vulnerabilities
- 💚 System integrity intact

---

## Future Enhancements

1. **AI-Powered Auto-Fix**
   - Gemini 2.0 Flash Thinking 모델로 에러 분석 → 자동 코드 패치 생성

2. **Rollback Mechanism**
   - 업데이트 실패 시 이전 버전으로 자동 롤백

3. **Performance Monitoring**
   - 성능 저하 감지 시 자동 최적화 제안

4. **Predictive Maintenance**
   - 에러 패턴 학습 → 사전 예방

---

## Summary

97layerOS Self-Maintenance Protocol은 다음을 보장합니다:

- ✅ **무중단 운영**: 자동 환경 점검 및 복구
- ✅ **보안 강화**: 주간 취약점 검사 및 업데이트
- ✅ **자가 치유**: 반복 에러 자동 분석 및 수정 제안
- ✅ **투명성**: 모든 활동 로그 기록
- ✅ **확장성**: Google Drive 동기화로 원격 업데이트

**Result**: 시스템이 스스로를 유지하며, 인간은 전략적 의사결정에만 집중.
