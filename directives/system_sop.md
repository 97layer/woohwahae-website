# 97LAYER System SOP (Standard Operating Procedure)

이 문서는 97LAYER 시스템 통합 운영을 위한 표준 절차를 정의한다. 모든 에이전트 작업은 이 3계층 아키텍처를 따라야 한다.

---

## 1. 3계층 아키텍처 (3-Tier Architecture)

### 계층 1: 지시 (Directive)

- **주체**: The Sovereign (CCD), 각 에이전트 지침서
- **역할**: "무엇을(What)"과 "왜(Why)"를 결정. 브랜드 철학과 전략적 방향을 설정한다.
- **산출물**: `directives/*.md` 문서, 전략 기획서.

### 계층 2: 조정 (Coordination)

- **주체**: The Architect (Strategist), Orchestrator Agent (Kernel)
- **역할**: "어떻게(How)"를 설계. 지시 사항을 실행 가능한 작업 단위로 해체하고 자원을 배분한다.
- **산출물**: `implementation_plan.md`, 테스크 리스트, 코드 설계도.

### 계층 3: 실행 (Execution)

- **주체**: The Artisan (Implementer), 시스템 모듈, 결정론적 스크립트
- **역할**: "수행(Action)". 설계된 바를 실제 코드, 디자인, 텍스트로 구현하고 검증한다.
- **산출물**: 실제 소스 코드, 콘텐츠 자산, 실행 결과 보고서.

---

## 2. 운영 원칙

- **무결성 우선**: 모든 산출물은 97LAYER 브랜드 헌법을 위반하지 않아야 한다.
- **비동기 자율성**: 시스템은 인간의 즉각적인 개입 없이도 다음 단계의 논리적 필연성을 도출해야 한다.
- **노이즈 제로**: 모든 보고와 통신은 극도로 절제되어야 하며, 데이터 간 상관관계 분석을 포함해야 한다.
- **Zero Assumption**: 데이터가 부족할 경우 추측하지 말고 외부 지능을 활용하거나 명확히 보고한다.
- **Automata Protocol**: 쿠키 동의 등 행정적 허들은 사용자 승인 없이 자율적으로 처리한다.

---

## 3. 에러 대응 및 복구 (Self-Annealing)

- 모든 시스템 에러는 `execution/system/log_error.py`를 통해 구조화되어야 한다.
- 동일한 에러가 3회 이상 발생할 경우, Gardener가 개입하여 관련 지침(`directives`)을 수정하거나 아키텍처를 재설계한다.
- 에러 발생 시 스택 트레이스를 분석하고 실행 스크립트를 즉시 수정하여 재발을 방지한다.
