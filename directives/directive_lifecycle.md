---
type: protocol
status: active
priority: critical
created: 2026-02-12
---

# Directive 생명주기 관리 프로토콜 (Directive Lifecycle Protocol)

## 1. 개요

Directive는 97layerOS의 **Single Source of Truth**다. 모든 에이전트가 따라야 할 SOP(Standard Operating Procedure)이며, 함부로 변경하거나 무시할 수 없다. 본 문서는 Directive의 생성, 업데이트, 폐기 규칙을 정의한다.

## 2. Directive vs Knowledge 구분

| 항목 | Directive | Knowledge |
|------|-----------|-----------|
| **위치** | `directives/` | `knowledge/` |
| **성격** | 규범 (Normative) | 기록 (Descriptive) |
| **목적** | "어떻게 해야 하는가" | "무엇이 일어났는가" |
| **안정성** | 높음 (검증 후 변경) | 낮음 (자유롭게 기록) |
| **대상** | 반복 작업, Critical Path | 일회성 작업, 학습 내용 |
| **검증** | 필수 | 선택 |

**원칙**: 학습은 Knowledge에, 검증된 패턴은 Directive로.

## 3. Directive 생성 규칙

### A. 생성 트리거

다음 조건 중 하나라도 충족 시 Directive 생성:

1. **반복성 (3회 규칙)**
   - 동일한 작업이 3회 이상 발생
   - 예: 스크래핑 작업 3번 수행 → `directives/scrape_workflow.md` 생성

2. **Critical Path**
   - 시스템 안정성에 필수적인 작업
   - 예: venv 설정, 데몬 관리, 상태 동기화

3. **Self-Annealing 패턴 발견**
   - 동일한 오류가 2회 이상 발생하고 해결책이 명확함
   - 예: Google Drive 권한 오류 → 외부 경로 사용 패턴

4. **Cross-Agent 작업**
   - 여러 에이전트가 협업해야 하는 복잡한 워크플로우
   - 예: 데이터 수집 → 분석 → 리포트 생성

### B. 생성 프로세스

```
1. Knowledge 기록
   ↓
2. 패턴 검증 (3회 이상 or Critical)
   ↓
3. Directive 초안 작성 (에이전트)
   ↓
4. Gardener 검토 (자동 or 수동)
   ↓
5. Directive 승격
   ↓
6. Git 커밋 (버전 관리)
```

### C. Directive 템플릿

```markdown
---
type: directive
status: active | draft | deprecated
priority: critical | high | medium | low
dependencies: [파일1, 파일2]
created: YYYY-MM-DD
last_updated: YYYY-MM-DD
---

# [Directive 제목]

## 1. 목적 (Objective)
이 directive가 해결하는 문제

## 2. 입력 (Inputs)
- 필요한 데이터
- 환경 변수
- 의존성

## 3. 도구 (Tools)
- 사용할 스크립트: `execution/xxx.py`
- MCP 서버: context7, fetch 등
- 외부 API

## 4. 실행 절차 (Procedure)
1. 단계별 명령어
2. 예상 결과
3. 검증 방법

## 5. 출력 (Outputs)
- 생성되는 파일
- 업데이트되는 상태
- 클라우드 결과물

## 6. 예외 처리 (Edge Cases)
- 오류 시나리오
- 복구 절차
- Fallback 전략

## 7. Self-Annealing 이력
- YYYY-MM-DD: [학습 내용]
```

## 4. Directive 업데이트 규칙

### A. 업데이트 트리거

1. **Self-Annealing 발생**
   - 오류 해결 후 즉시 업데이트
   - `## 7. Self-Annealing 이력`에 추가

2. **API 변경**
   - 외부 API 스펙 변경 감지
   - 즉시 Directive 수정 및 테스트

3. **성능 개선**
   - 더 나은 방법 발견 시
   - 기존 방법과 비교 후 업데이트

4. **에이전트 피드백**
   - 에이전트가 Directive 불명확하다고 판단 시
   - 명확성 개선 목적 업데이트

### B. 업데이트 프로세스

```python
# execution/system/update_directive.py 사용

update_directive(
    directive_path="directives/scrape_workflow.md",
    section="Self-Annealing 이력",
    content="2026-02-12: Rate limit 회피를 위해 sleep(2) 추가"
)
```

### C. 업데이트 금지 사항

❌ **절대 하지 말 것**:
1. 테스트 없이 Critical Path directive 수정
2. 이전 버전 삭제 (Git 이력 유지)
3. Self-Annealing 이력 제거
4. 여러 directive를 한 번에 대규모 수정

✅ **해야 할 것**:
1. 작은 단위로 업데이트
2. 변경 이유를 명확히 기록
3. 테스트 후 커밋
4. `last_updated` 타임스탬프 갱신

## 5. Directive 폐기 규칙

### A. 폐기 조건

1. **대체 방법 확립**
   - 더 나은 directive가 생성됨
   - 기존 방법이 obsolete

2. **도구 지원 중단**
   - 외부 API 종료
   - 스크립트 의존성 제거

3. **요구사항 변경**
   - 비즈니스 로직 변경으로 불필요해짐

### B. 폐기 프로세스

```
1. 상태를 "deprecated"로 변경
   ↓
2. 대체 directive 링크 추가
   ↓
3. 3개월 대기 (혹시 모를 롤백 대비)
   ↓
4. `directives/archive/`로 이동
   ↓
5. Git 이력 보존
```

**예시**:
```markdown
---
type: directive
status: deprecated
deprecated_date: 2026-02-12
replacement: directives/new_workflow.md
---

# [Deprecated] Old Workflow

⚠️ 이 directive는 폐기되었습니다.
새 방법: [new_workflow.md](new_workflow.md)
```

## 6. Knowledge → Directive 승격 기준

### A. 승격 체크리스트

```markdown
- [ ] 작업이 3회 이상 반복되었는가?
- [ ] 작업이 Critical Path인가?
- [ ] 명확한 입력/출력이 정의되는가?
- [ ] 재현 가능한 절차가 있는가?
- [ ] 오류 처리 로직이 포함되는가?
- [ ] 다른 에이전트가 이해 가능한가?
```

### B. 승격 프로세스

```python
# execution/system/promote_to_directive.py

# Knowledge 파일 분석
knowledge_file = "knowledge/20260212_snapshot_workflow.md"

# Directive 형식으로 변환
directive = convert_to_directive(knowledge_file)

# Gardener 검토 요청
if gardener.approve(directive):
    save_directive("directives/snapshot_workflow.md", directive)
    git_commit(f"Add directive: snapshot_workflow")
```

### C. 승격 거부 사유

다음 경우 Directive로 승격하지 않고 Knowledge에 유지:

1. 일회성 작업 (반복 가능성 없음)
2. 컨텍스트가 너무 특수함 (일반화 불가)
3. 아직 검증 중인 실험적 방법
4. 개인 선호도 기반 (표준화 불가)

## 7. Gardener와의 연동

### A. Gardener의 역할

**자동 작업**:
1. Knowledge 폴더 스캔 (매일 1회)
2. 반복 패턴 감지
3. Directive 후보 추출
4. 중복 Directive 통합 제안

**수동 작업**:
1. Critical Path directive 검토
2. 승격 최종 승인
3. 구조 개선 제안

### B. Gardener 실행

```python
# libs/gardener.py의 analyze_knowledge() 메서드

gardener = Gardener(ai_engine, memory_manager, project_root)

# Knowledge 분석
candidates = gardener.analyze_knowledge()

# 승격 후보 출력
for candidate in candidates:
    print(f"[CANDIDATE] {candidate['title']}")
    print(f"  Frequency: {candidate['frequency']} times")
    print(f"  Recommendation: {candidate['recommendation']}")
```

## 8. 버전 관리 (Git Integration)

### A. Git 커밋 규칙

**Directive 생성**:
```bash
git add directives/new_workflow.md
git commit -m "directive: Add new_workflow (3회 반복 패턴)"
```

**Directive 업데이트**:
```bash
git add directives/scrape_workflow.md
git commit -m "directive: Update scrape_workflow (self-anneal: rate limit 처리)"
```

**Directive 폐기**:
```bash
git mv directives/old_workflow.md directives/archive/
git commit -m "directive: Deprecate old_workflow (replaced by new_workflow)"
```

### B. 변경 이력 추적

```bash
# Directive 변경 이력 확인
git log --oneline directives/scrape_workflow.md

# 특정 날짜의 Directive 복원
git show HEAD@{2026-02-10}:directives/scrape_workflow.md
```

## 9. 에이전트별 책임

### A. 모든 에이전트 (공통)

✅ **해야 할 것**:
- Directive 준수
- Self-Annealing 이력 기록
- 불명확한 부분 질문

❌ **하지 말아야 할 것**:
- Directive 무시
- 임의 수정 (테스트 없이)

### B. Orchestrator (Claude, Gemini 등)

✅ **책임**:
- Directive 읽고 실행
- Knowledge 기록
- 승격 후보 제안

### C. Gardener (자가 진화 시스템)

✅ **책임**:
- 패턴 감지
- 승격 관리
- 중복 제거
- 구조 최적화

## 10. 실전 예시

### 예시 1: 스크래핑 작업 (3회 반복)

**1회차**: Knowledge 기록
```markdown
# knowledge/sessions/20260210_scrape_news.md
뉴스 사이트 스크래핑 완료. requests + BeautifulSoup 사용.
```

**2회차**: 패턴 발견
```markdown
# knowledge/sessions/20260211_scrape_blog.md
블로그 스크래핑 완료. 동일한 방법 사용. Rate limit 주의 필요.
```

**3회차**: Directive 생성
```markdown
# directives/scrape_website.md
---
type: directive
status: active
priority: high
---
# 웹사이트 스크래핑 표준 절차
...
```

### 예시 2: Self-Annealing (Google Drive 권한 오류)

**오류 발생**:
```
OSError: [Errno 1] Operation not permitted: '/Users/97layer/내 드라이브/...'
```

**해결 및 Directive 업데이트**:
```markdown
# directives/snapshot_backup.md

## 7. Self-Annealing 이력

- 2026-02-12: Google Drive 권한 오류 발생
  - 원인: macOS 파일 접근 권한
  - 해결: 외부 경로 `/Users/97layer/97layerOS_Snapshots/` 사용
  - 교훈: 클라우드 동기화 폴더를 Primary로 사용하지 말 것
```

## 11. 체크리스트 (새 에이전트용)

### Directive 생성 전

```markdown
- [ ] Knowledge에 3회 이상 기록되었는가?
- [ ] 재현 가능한 절차가 있는가?
- [ ] 입력/출력이 명확한가?
- [ ] 오류 처리가 포함되는가?
- [ ] 템플릿 구조를 따르는가?
```

### Directive 업데이트 전

```markdown
- [ ] 변경 이유가 명확한가? (Self-Annealing, API 변경 등)
- [ ] 테스트를 완료했는가?
- [ ] last_updated 타임스탬프를 갱신했는가?
- [ ] Git 커밋 메시지가 명확한가?
```

## 12. 관련 파일

- [CLAUDE.md](../CLAUDE.md) - 3-Layer Architecture 정의
- [directives/system_handshake.md](system_handshake.md) - 핸드셰이크 프로토콜
- [directives/agent_instructions.md](agent_instructions.md) - 에이전트 운영 지침
- [execution/system/sync_status.py](../execution/system/sync_status.py) - 상태 동기화 도구
- [libs/gardener.py](../libs/gardener.py) - 자가 진화 시스템

---

**Directive는 살아있는 문서다. 학습하고, 진화하고, 시스템을 강화한다.**
