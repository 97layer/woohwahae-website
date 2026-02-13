---
description: 97LAYER 에이전트 통합 협업 워크플로우 - 비즈니스 인텔리전스 및 실행 설계도(Blueprint) 생성
---

# 워크플로우명: 97LAYER_Omni_Blueprint

이 워크플로우는 새로운 프로젝트나 비즈니스 아이디어를 5인의 에이전트가 협업하여 97LAYER의 철학에 부합하는 정교한 실행 설계도로 전환하는 프로세스입니다.

## 1. 단계별 실행 가이드

### Step 1: Market & Knowledge Scouting (Strategy Analyst)

- **Input**: 프로젝트 키워드 또는 초기 아이디어
- **Action**:
  - `external_intelligence (context7)`를 사용하여 실시간 시장 트렌드 및 레퍼런스 수집.
  - `knowledge_retrieval (filesystem)`로 기존 내부 지식과의 상관관계를 분석하여 **Raw Signal** 리포트 작성.
- **Output**: `[아이디어명]_Raw_Signal.md`

### Step 2: Philosophical Alignment & Strategy (Creative Director)

- **Input**: SA의 Raw Signal 리포트
- **Action**:
  - `cognitive_processing (sequential-thinking)`을 가동하여 아이디어의 비즈니스 가치와 97LAYER 브랜드 철학(미니멀리즘, 본질 주의)의 합치 여부 검증.
  - 핵심 전략 방향성(Strategy Insight) 확정.
- **Output**: `[아이디어명]_Strategy_Insight.md`

### Step 3: Architecture & Logic Design (Technical Director)

- **Input**: CD의 Strategy Insight
- **Action**:
  - 전략을 구현하기 위한 논리 구조 설계(`Logical Blueprint`).
  - 필요한 기술 도구 선정 및 자동화 프로세스 정의.
  - `quality_control (TestSprite)`을 통한 에지 케이스 선제 검토.
- **Output**: `[아이디어명]_Tech_Blueprint.md`

### Step 4: Narrative & Brand Voice Refinement (Chief Editor)

- **Input**: TD의 Tech Blueprint 및 CD의 인사이트
- **Action**:
  - 기술적 언어를 문학적이고 절제된 97LAYER의 서사로 번역.
  - `knowledge_retrieval`을 통해 기존 브랜드 보이스와의 정합성 검토.
- **Output**: `[아이디어명]_Narrative.md`

### Step 5: Visual Identity & Asset Guide (Art Director)

- **Input**: CE의 Narrative 및 전체 기획안
- **Action**:
  - 서사에 부합하는 비주얼 컨셉 설계.
  - 여백과 비례를 고려한 레이아웃 그리드 및 에셋 가이드 작성.
  - `external_intelligence`로 최신 미학적 레퍼런스 매칭.
- **Output**: `[아이디어명]_Visual_Guide.md`

## 2. 최종 결과물

- 파이브 인 크루의 합의가 요약된 **[아이디어명]_Final_Omni_Blueprint.md** 파일 생성.
- 구글 드라이브(97layerOS_Sync) 동기화를 통해 웹 제미나이와 실시간 공유.

## 3. 실행 명령어 예시

```bash
# 이 워크플로우를 시작하려면 오케스트레이터에게 다음 명령을 하달하십시오.
/run omni-blueprint "프로젝트명/아이디어"
```
