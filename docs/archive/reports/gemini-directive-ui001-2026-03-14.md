# Gemini 작업 지시서 — ui_001
**발행**: 2026-03-14 / 컨트롤 타워: Gemini (Staff Strategist)
**잡 ID**: job_ui_semantic_refactor
**우선순위**: high

---

## 1. 컨텍스트 (Context)
현재 Layer OS의 어드민(Cockpit)은 엔지니어링 용어(Intake, AgentJob, Dispatch 등)가 노출되어 비개발자 창업자의 '직관적 판단'을 방해하고 있음. «THE ORIGIN»의 소거(Subtraction) 원칙에 따라, 기술적 노이즈를 제거하고 비즈니스 언어로 인터페이스를 재정의함.

## 2. 목표 (Objectives)
- **언어적 소거**: 개발자 전용 용어를 비즈니스 직관 언어로 전면 치환.
- **흐름 중심 재배치**: 파편화된 섹션을 [외부 신호 -> 에이전트 가공 -> 최종 승인]의 파이프라인 뷰로 정렬.
- **판단(Judgment) 중심 UI**: 상태값 나열을 줄이고 창업자가 'Approve/Reject'를 결정하는 데 필요한 정보만 노출.

## 3. 상세 작업 가이드 (Detailed Steps)

### Step 1: 용어 매핑 및 치환 (Label Refactoring)
`docs/brand-home/` 하위의 모든 UI 컴포넌트에서 다음 용어를 치환할 것:
- `AgentJob` / `WorkItem` → **"진행 중인 과업"**
- `Source Intake` → **"포착된 신호 / 아이디어"**
- `Review Room` → **"최종 결재판"**
- `Observation` → **"시스템 기억"**
- `Dispatch` → **"수행 지시"**

### Step 2: 파이프라인 레이아웃 개편
- 현재의 대시보드 그리드를 수평적 타임라인(Lane) 구조로 변경.
- **Lane 1 (The Sensor)**: 포착된 신호(Source Intake) 목록.
- **Lane 2 (The Builder)**: 에이전트가 처리 중인 과업(Active Jobs) 상태.
- **Lane 3 (The Reviewer)**: 창업자의 최종 승인이 필요한 안건(Review Room Items).

### Step 3: 데이터 노이즈 소거
- 리스트 뷰에서 `JobID`, `Created At`, `Payload` 등 엔지니어링 메타데이터는 '상세 보기'나 툴팁으로 숨기고, **"요약(Summary)"**과 **"핵심 리스크(Risk)"**를 전면에 배치.

## 4. 완료 및 검증 기준 (Definition of Done)
- [ ] `docs/brand-home/` 내 주요 라벨들이 비즈니스 용어로 변경됨.
- [ ] 대시보드 진입 시 시스템의 현재 흐름(Pipeline)이 한눈에 파악됨.
- [ ] 비개발자가 1분 안에 현재 가장 시급한 결재 건이 무엇인지 식별 가능함.
- [ ] `npm run test` (프론트엔드 테스트) 통과.

---

## 5. 실행 에이전트 지침
본 지시서를 수령한 Implementer는 코드 수정 전 `Plan`을 먼저 작성하여 Review Room에 제출하고, 창업자의 승인 후 작업을 개시할 것.
