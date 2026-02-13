# Data Asset Management Directive (SOP)

## 1. 개요

이 지시서는 97layerOS의 모든 데이터 자산이 '지능적 가치'를 유지하도록 관리하는 표준 운영 절차(SOP)를 정의합니다.

## 2. 파일 명명 및 배치 규칙

- **Naming**: `[scope]_[subject]_[type].md` 형식을 권장합니다. (예: `minimal_life_strategy_insight.md`)
- **Placement**:
  - `directives/`: 운영 절차 및 행동 강령 (Deterministic)
  - `knowledge/`: 추출된 인사이트 및 학습 데이터 (Intelligence)
  - `execution/`: 실행 스크립트 (Deterministic)

## 3. 지능 메타데이터 (Ontology Tags)

모든 핵심 지능 파일은 상단에 다음과 같은 YAML 태그를 포함해야 합니다.

```yaml
---
id: [UUID 또는 고유 ID]
context: [이 데이터가 생성된 맥락]
links: [관련된 다른 파일들의 경로 목록]
density: [1-10, 정보의 밀도 및 중요도]
last_curated: [YYYY-MM-DD]
---
```

## 4. 금지 사항

- **No Placeholders**: "나중에 작성"과 같은 자리 표시자 사용을 엄격히 금지합니다.
- **No Duplicate Venvs**: 프로젝트 루트 내의 가상 환경 생성을 금지하며, 발견 시 Sentinel에 의해 즉각 소거됩니다.
- **Minimizing Noise**: 불필요한 수식어, 이모지, 중복된 문장을 배제하고 명사 중심의 고밀도 텍스트를 유지합니다.

## 5. 지능 배양 워크플로우 (Incubation Workflow)

- **Step 1: Capture**: `signal_capture` 스킬로 `knowledge/inbox/`에 원시 데이터 저장.
- **Step 2: Incubate**: 보관된 데이터를 숙성시키며 보충 정보 수집.
- **Step 3: Curate**: `data_curation` 스킬로 정제 및 온톨로지 태깅.
- **Step 4: Asset**: 최종 지능 자산으로 `knowledge/` 또는 프로젝트 루트로 승격.

## 6. 큐레이션 주기

- 모든 지능 자산은 에이전트 오케스트레이션 과정에서 실시간으로 업데이트되거나, 주 1회 Sentinel에 의해 무결성 검사를 받습니다.
