---
type: pattern
frequency: 2
first_seen: 2026-02-12
last_seen: 2026-02-12
recommendation: Monitor, Directive 승격 고려
---

# Pattern: 워크플로우 연속성 확보

## 발견 경위

Gemini와 Claude 간 작업 인계 시, 컨텍스트 손실 없이 완벽하게 이어받음.
이는 체계적인 상태 기록과 핸드셰이크 프로토콜 덕분.

## 패턴 설명

### 구조

```
에이전트 A (Gemini) 작업 중
  ↓
상태 기록 (~/.gemini/brain/*/task.md)
  ↓
knowledge/status.json 업데이트
  ↓
에이전트 B (Claude) 투입
  ↓
상태 파일 읽기
  ↓
작업 완벽 인계
```

### 성공 요인

1. **Gemini Brain 시스템**
   - task.md: 체크리스트
   - implementation_plan.md: 구현 계획
   - walkthrough.md: 상세 로그
   - 총 287개 파일 축적

2. **System Handshake Protocol**
   - knowledge/status.json에 상태 저장
   - 표준화된 상태 객체 구조
   - 런타임 환경 정보 포함

3. **Knowledge Base**
   - 각 세션마다 상세 문서화
   - 이전 에이전트의 의도 명확히 기록

## 발생 사례

### Case 1: Gemini → Claude (2026-02-12)

**작업**: 스냅샷 자동화 완료

**Gemini 작업**:
- snapshot_daemon.py 생성
- task.md에 체크리스트 기록
- 일부 작업 미완료 (권한 오류)

**Claude 인계**:
1. Gemini Brain 파일 발견
2. task.md 읽고 미완료 항목 파악
3. 권한 오류 해결
4. 작업 완료
5. 새 문서 작성 (gemini_workflow_continuation.md)

**결과**: ✅ 100% 성공

### Case 2: Claude 자가 인계 (2026-02-12)

**작업**: Context7 MCP 설정 → 스냅샷 격리

**방법**:
- 이전 세션의 knowledge/ 파일 읽기
- status.json으로 컨텍스트 파악
- HANDOVER_2026-02-12.md 참조

**결과**: ✅ 자연스러운 연속성

## 적용 조건

### ✅ 이 패턴이 작동하는 경우

- 상태 파일이 최신으로 유지됨
- 에이전트가 문서를 읽는 습관이 있음
- 명확한 핸드셰이크 프로토콜 존재
- Knowledge Base가 구조화됨

### ❌ 이 패턴이 실패하는 경우

- 상태 파일 불일치 (task_status.json vs status.json)
- 문서가 모호하거나 불완전
- 에이전트가 프로토콜 무시
- 메모리가 분산되어 찾기 어려움

## 개선 사항 (2026-02-12 적용)

### 1. 상태 동기화 자동화
```python
# execution/system/sync_status.py
# 두 상태 파일을 자동 병합
```

### 2. Directive 생명주기 문서화
```markdown
# directives/directive_lifecycle.md
# 명확한 업데이트 규칙
```

### 3. Knowledge 구조화
```
knowledge/
├── sessions/   # 세션별 기록
├── patterns/   # 반복 패턴
├── decisions/  # 주요 결정
└── errors/     # 오류 해결책
```

### 4. 체크리스트 자동화
```python
# execution/system/update_checklist.py
# Markdown 체크박스 자동 업데이트
```

## Directive 승격 가능성

### 평가: **높음 (80%)**

**이유**:
- 2회 이상 검증됨 (Gemini→Claude, Claude 자가인계)
- Critical Path (팀 협업 필수)
- 재현 가능한 절차
- 명확한 입력/출력

**필요 조건**:
- 1회 더 검증 (3회 규칙)
- 자동화 도구 안정화
- 에이전트 피드백 수집

## 다음 단계

1. **자동화 검증**
   - sync_status.py 1주일 운영
   - update_checklist.py 실사용 테스트

2. **추가 인계 사례 수집**
   - Claude → Gemini
   - Cursor 참여 시

3. **Directive 작성**
   - `directives/agent_handover.md`
   - 표준 인계 절차 정의

## 관련 파일

- [knowledge/gemini_workflow_continuation.md](../sessions/gemini_workflow_continuation.md)
- [directives/system_handshake.md](../../directives/system_handshake.md)
- [execution/system/sync_status.py](../../execution/system/sync_status.py)
- [knowledge/workflow_continuity_assessment.md](../sessions/workflow_continuity_assessment.md)

## 결론

워크플로우 연속성 패턴은 검증되었으며, 자동화 도구로 더욱 강화됨.
1회 더 검증 후 Directive 승격 권장.
