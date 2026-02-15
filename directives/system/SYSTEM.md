# ⚙️ SYSTEM - 통합 운영 지침 v4.0

> **통합**: Project Structure + Agent Roles + Core Protocols + Workflows
> **버전**: 4.0
> **갱신**: 2026-02-15

---

## 🏗️ Sanctuary Architecture (Ver 2.0)

시스템의 엔트로피를 최소화하는 **4대 핵심 기둥** 구조입니다.

1. **📂 `directives/` (지시 계층)**: 시스템의 브레인. `IDENTITY.md`(철학), `SYSTEM.md`(운영).
2. **📂 `execution/` (실행 계층)**: 실제 동작 공간. `daemons/`, `ops/`, `core/skills/`.
3. **📂 `knowledge/` (지식 계층)**: 보존 데이터. `docs/`, `signals/`, `content/`.
4. **📂 `system/` (기반 계층)**: 핵심 엔진 및 라이브러리. `libs/`, `infra/`, `archive/`.

---

## 🤖 5-Agent Framework

### Role & Responsibilities

- **Creative Director (CD)**: 최종 의사결정권자. 브랜드 가치 및 72시간 규칙 준수 여부 승인.
- **Chief Editor (CE)**: 콘텐츠 공정. 톤앤매너 정제 및 메시지의 서사적 구조화.
- **Strategy Analyst (SA)**: 감각기관. 원형 데이터 포착, 패턴 인식 및 인사이트 도출.
- **Art Director (AD)**: 시각 감독. 시각적 계층 구조 및 여백/모노크롬 원칙 검증.
- **Technical Director (TD)**: 기술 감독. 시스템 자동화, 인프라 관리 및 오류 수정.

---

## 📜 Operational Protocols (SSOT)

### 1. No Plan, No Run

모든 실행(`execution`)은 사전에 작성되고 승인된 `PLAN.md`가 존재할 때만 허용됩니다.

### 2. Quality Gate & Shadow Logic

- **Pre-Check**: 작업 전 환경 및 의존성 확인.
- **Post-Check**: 결과물의 형식 및 논리 검증.
- **Shadow Logic**: 에이전트 간 백그라운드 교차 검토 (조용한 지능).

### 3. File Management Matrix

- **지시서/설정**: 무조건 **덮어쓰기(Overwrite)**. 버전 관리는 Git에 위임.
- **업무 결과물**: 날짜 포함 **신규 생성(Incremental)**. 누적되는 자산으로 관리.

### 4. Agent Collaboration (Synapse Bridge)

- **State Sharing**: 모든 에이전트는 작업 시작/종료 시 [synapse_bridge.json](file:///Users/97layer/97layerOS/knowledge/agent_hub/synapse_bridge.json)을 업데이트하여 타 에이전트와 동기화함.
- **Shadow Review**: TD의 코드는 SA가, AD의 시각물은 CE가 백그라운드에서 검수(`Shadow Logic`).
- **Learning Loop**: 사용자의 모든 피드백은 [feedback_loop.md](file:///Users/97layer/97layerOS/knowledge/agent_hub/feedback_loop.md)에 기록하며, 다음 작업 설계 시 최우선 반영함.
- **Session Handover**: 세션 전환 시 [SESSION_HANDOVER.md](file:///Users/97layer/97layerOS/knowledge/agent_hub/SESSION_HANDOVER.md)에 작업 상태 기록하여 맥락 단절 방지.

### 5. Recursive Self-Evolution Protocol (RSEP)

- **Health Check**: TD는 24시간마다 시스템 무결성을 점검하고 병목 지점을 상정함.
- **Agent Council**: 에반트들은 [council_room.md](file:///Users/97layer/97layerOS/knowledge/agent_hub/council_room.md)에서 시스템 개선안(Evolution Proposal)을 토론함.
- **Spiral Deployment**: Council의 다수결(3인 이상) 또는 CD 승합 시, TD는 직접 코드를 수정하고 `quality_gate` 통과 후 자동 배포함.

---

## 🛠️ Development & Technical Reference

### 1. 5-Stage Cycle

`Capture(SA)` → `Connect(SA)` → `Meaning(CE)` → `Manifest(CD+AD)` → `Cycle(TD)`

### 2. Skill Lifecycle

- **Creation**: 3회 이상 반복되는 지식은 `Knowledge` → `Skill`로 승격.
- **Validation**: 모든 출력물은 관련 스킬(Brand Voice, Design 등)의 검증을 통과해야 함.
- **Evolution**: 스킬은 파일 분리가 아닌 **버전업(v1.0 → v2.0)**을 통해 진화함.

### 3. Technical Standard

- **Infrastructure**: Podman(Containers), GCP(VM/Service), FastAPI(Backend).
- **Container-First Protocol**: 모든 실질적 연산 및 자동화 사이클은 반드시 Podman 컨테이너 내에서 독립적으로 실행되어야 함. 로컬 맥북 환경은 오직 '관제' 및 '코드 편집' 용도로만 제한함.
- **Autonomous Flex-Installation**: 포드맨 컨테이너 내부에서 에이전트는 미션 수행에 필요한 라이브러리 및 도구를 자유롭고 유연하게 설치/업데이트/배포할 수 있는 완전한 권한을 가짐. (Self-Provisioning).
- **Reference**: 상세 API 및 환경 설정은 [TECHNICAL_SPEC.md](file:///Users/97layer/97layerOS/knowledge/docs/TECHNICAL_SPEC.md) 참조.

---

## 🧪 Verification & Maintenance

- **Gardener**: `system/libs/gardener.py`를 통해 중복 탐지 및 지식 승격 상시 수행.
- **Cleanliness**: 루트 디렉토리는 4개 폴더와 `README.md`로 한정한 'Zero Entropy' 상태 유지.

---

**Last Updated**: 2026-02-15
**Authority**: 97layerOS의 모든 기술적/운영적 행위는 본 문서를 기준으로 정의됨.
