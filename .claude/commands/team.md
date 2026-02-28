---
description: 팀 모드 — 리더 기억 보존 + 워커(서브에이전트) 순차 실행
---

# /team — Team Mode

리더(메인 세션)가 state.md로 기억 유지하면서 워커(Task 서브에이전트)를 순차 실행.

## 루프

### 1. 컨텍스트 추출
```bash
python3 core/scripts/context_snippet.py [keys...]
# keys: infra / pipeline / web / content / design / agent / state
```

### 2. 워커 스폰
Task 도구 → subagent_type 선택 → 프롬프트:
```
목표: [구체적 태스크]

컨텍스트:
[context_snippet.py 출력 붙이기]

완료 후 반드시:
- 결과 한 줄 요약
- 리더에게 전달할 인사이트
- 수정된 파일 목록
```

### 3. 결과 반영
워커 완료 → 인사이트 중 시스템 영향 있는 것만 state.md 반영 → 다음 태스크로.

### 4. 종료
모든 태스크 완료 → `/handoff`

## 병렬 실행
의존성 없는 태스크 → Task 도구 동시 호출.
