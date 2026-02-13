# Knowledge Base 구조

본 폴더는 97layerOS의 학습 및 기록을 저장하는 통합 메모리 시스템입니다.

## 폴더 구조

### 📁 sessions/
세션별 작업 기록. 에이전트가 수행한 작업의 상세 로그.

**파일명 규칙**: `YYYYMMDD_작업명.md`

**예시**:
- `20260212_snapshot_isolation.md` - 스냅샷 격리 작업
- `20260212_mcp_integration.md` - MCP 서버 연동
- `20260212_gemini_continuation.md` - Gemini 워크플로우 이어받기

**용도**:
- 작업 내역 추적
- 에이전트 인계 자료
- Self-Annealing 기록

### 📁 patterns/
반복적으로 발견되는 패턴. Directive 승격 후보.

**파일명 규칙**: `pattern_설명.md`

**예시**:
- `pattern_google_drive_permissions.md` - Drive 권한 오류 패턴
- `pattern_api_rate_limiting.md` - API Rate Limit 회피 패턴
- `pattern_venv_isolation.md` - 가상환경 격리 패턴

**용도**:
- 반복 작업 식별
- Directive 생성 준비
- Gardener 분석 소스

### 📁 decisions/
주요 아키텍처 결정 및 이유.

**파일명 규칙**: `decision_YYYYMMDD_주제.md`

**예시**:
- `decision_20260212_snapshot_external_storage.md` - 스냅샷 외부 저장 결정
- `decision_20260210_3layer_architecture.md` - 3-Layer 아키텍처 채택

**용도**:
- 의사결정 이력
- 아키텍처 변경 추적
- 왜 이렇게 했는지 기록

### 📁 errors/
오류 및 해결책 데이터베이스.

**파일명 규칙**: `error_오류유형.md`

**예시**:
- `error_permission_denied.md` - 권한 오류 해결책 모음
- `error_rate_limit_exceeded.md` - Rate Limit 해결책
- `error_import_failed.md` - Import 오류 트러블슈팅

**용도**:
- 트러블슈팅 가이드
- Self-Annealing 참조
- FAQ 구축

### 📁 memory/
레거시 메모리 파일 (이전 구조).

**상태**: 마이그레이션 대기 중

### 📄 status.json
현재 시스템 상태 (통합 상태 객체).

**동기화**: `task_status.json`과 자동 동기화됨 (`execution/system/sync_status.py`)

## 사용 규칙

### 1. 새 세션 시작 시

```bash
# 현재 상태 확인
cat knowledge/status.json

# 최근 3개 세션 확인
ls -lt knowledge/sessions/ | head -4
```

### 2. 작업 완료 시

```bash
# 세션 기록 생성
# knowledge/sessions/YYYYMMDD_작업명.md

# 상태 업데이트
python3 execution/system/sync_status.py
```

### 3. 패턴 발견 시

작업이 3회 이상 반복되면:
1. `knowledge/patterns/pattern_설명.md` 생성
2. Gardener에게 Directive 승격 요청
3. 검증 후 `directives/`로 이동

### 4. 오류 해결 시

```markdown
# knowledge/errors/error_유형.md에 추가

## [오류 제목]

**발생 날짜**: YYYY-MM-DD
**증상**: [오류 메시지]
**원인**: [분석 결과]
**해결**: [해결 방법]
**예방**: [재발 방지 조치]
```

## 통합 원칙

### Knowledge vs Directive

| Knowledge | Directive |
|-----------|-----------|
| 기록 (Descriptive) | 규범 (Normative) |
| 자유롭게 작성 | 검증 후 작성 |
| 학습 단계 | 표준화 단계 |
| "무엇이 일어났나" | "어떻게 해야 하나" |

### 승격 기준

Knowledge → Directive 승격 조건:
- ✅ 3회 이상 반복
- ✅ Critical Path
- ✅ 재현 가능한 절차
- ✅ 명확한 입력/출력

## 관련 도구

- [execution/system/sync_status.py](../execution/system/sync_status.py) - 상태 동기화
- [libs/gardener.py](../libs/gardener.py) - 패턴 분석 및 승격
- [directives/directive_lifecycle.md](../directives/directive_lifecycle.md) - Directive 관리 프로토콜

---

**Knowledge는 시스템의 집단 기억이다. 기록하고, 패턴을 찾고, 진화한다.**
