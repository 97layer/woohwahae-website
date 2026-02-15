# 97layerOS Directives - Agent Constitution

## ⚠️ 신규 에이전트 필독 순서 (위반 시 시스템 파편화)

### 1단계: 시스템 헌법 (3-Layer Architecture)
1. [../CLAUDE.md](../CLAUDE.md) - 3-Layer 아키텍처 원칙
2. [directive_lifecycle.md](directive_lifecycle.md) ⭐ **핵심 헌법**
3. [system_handshake.md](system_handshake.md) - 에이전트 교대 프로토콜

### 2단계: 정체성 (누구를 돕는가)
4. [97layer_identity.md](97layer_identity.md) ⭐ Foundation
5. [woohwahae_identity.md](woohwahae_identity.md) 🔒 **Read-Only**
6. [brand_constitution.md](brand_constitution.md) 🔒 **Read-Only**

### 3단계: 역할 선택
에이전트별 역할 정의: [agents/README.md](agents/README.md)

### 4단계: 프로토콜 숙지 (역할별 필수)
| 역할 | 필수 Directive |
|------|---------------|
| SA (Strategy Analyst) | cycle_protocol.md, anti_algorithm_protocol.md |
| AD (Art Director) | visual_identity_guide.md, aesop_benchmark.md |
| CE (Chief Editor) | imperfect_publish_protocol.md, communication_protocol.md |
| CD (Creative Director) | brand_constitution.md, 97layer_identity.md, woohwahae_identity.md |
| TD (Technical Director) | cycle_protocol.md, daemon_workflow.md, sync_protocol.md |

## 🔒 브랜드 헌법 보호 (Read-Only)

다음 파일은 Gardener 자동 수정 금지:
- `woohwahae_identity.md` 🔒
- `brand_constitution.md` 🔒
- `97layer_identity.md` 🔒

**이유**: 브랜드 정체성은 AI가 "최적화"할 대상이 아님.
인간(97layer)만 수정 가능.

## 📜 Directive 생성/수정 규칙 (Lifecycle Protocol)

### 금지 사항 (❌ 헌법 위반)
1. Directive 3회 반복 규칙 무시
2. 테스트 없이 Critical Path directive 수정
3. Self-Annealing 이력 삭제
4. Git 커밋 없이 directive 변경
5. 루트에 임의 폴더 생성
6. Knowledge와 Directive 혼동

### 필수 준수 (✅ 헌법 준수)
1. 모든 반복 작업 3회 시 Knowledge→Directive 승격
2. Directive 수정 시 [directive_lifecycle.md](directive_lifecycle.md) 7조 Self-Annealing 이력 기록
3. 변경 사항 Git 커밋 (`git commit -m "directive: [이유]"`)
4. Gardener 패턴 스캔 결과 따르기

## 🌱 Gardener (정원사) 시스템

**위치**: `libs/gardener.py`

**역할**:
- Knowledge 폴더 자동 스캔 (매일)
- 3회 반복 패턴 감지
- Directive 승격 후보 추천
- 중복 파일/폴더 탐지 및 정리
- Self-Annealing 이력 분석

**실행**:
```bash
python3 -c "from libs.gardener import Gardener; from libs.ai_engine import AIEngine; from libs.memory_manager import MemoryManager; g = Gardener(AIEngine(), MemoryManager(), '.'); print(g.run_cycle(7))"
```

## 📂 표준 폴더 구조 (절대 변경 금지)

```
97layerOS/
├── directives/          ← 규범 (Normative) - 어떻게 해야 하는가
│   ├── README.md       ← 당신이 지금 읽는 파일
│   ├── agents/         ← 역할별 매뉴얼
│   └── *.md            ← 27개 프로토콜
├── knowledge/          ← 기록 (Descriptive) - 무엇이 일어났는가
│   ├── system/         ← 시스템 상태
│   ├── agent_hub/      ← 멀티모달 협업
│   └── sessions/       ← 작업 기록
├── execution/          ← Python 도구들
│   ├── launchers/      ← 런처 스크립트
│   └── ops/            ← 운영 스크립트
├── libs/               ← 공유 라이브러리 (멀티모달 포함)
├── deployment/         ← 배포 스크립트
├── docs/               ← 문서
│   ├── milestones/     ← 완료 보고서
│   └── dashboard/      ← 대시보드
└── .tmp/               ← 임시 파일만
    ├── cache/          ← AI 캐시
    └── drive/          ← Drive 동기화
```

## 🔄 상태 파일 위치 (표준)
- `knowledge/system_state.json` - 에이전트 실시간 상태
- `knowledge/system/task_status.json` - 작업 진행 (루트에 symlink)
- `knowledge/agent_hub/synapse_bridge.json` - 멀티모달 협업

## ⚡ 멀티모달 시스템 (절대 수정 금지)
- `libs/async_agent_hub.py` - 병렬 처리 허브
- `libs/claude_engine.py` - Claude Opus CD
- `libs/gemini_engine.py` - Gemini Flash SA/AD/CE
- `execution/async_five_agent_multimodal.py` - 5-Agent 시스템
- `execution/async_telegram_daemon.py` - 텔레그램 통합

## 📚 전체 Directive 색인

### 철학 & 정체성
- [97layer_identity.md](97layer_identity.md) ⭐ Foundation
- [97layerOS_Optimization_Directive.md](97layerOS_Optimization_Directive.md) - 시스템 최적화
- [woohwahae_identity.md](woohwahae_identity.md) 🔒 - 브랜드
- [woohwahae_brand_source.md](woohwahae_brand_source.md) - 브랜드 소스
- [brand_constitution.md](brand_constitution.md) 🔒 - 헌법

### 프로토콜 (운영 규칙)
- [cycle_protocol.md](cycle_protocol.md) ⭐ 5단계 사이클
- [imperfect_publish_protocol.md](imperfect_publish_protocol.md) ⭐ 72시간 규칙
- [anti_algorithm_protocol.md](anti_algorithm_protocol.md)
- [communication_protocol.md](communication_protocol.md)
- [efficiency_protocol.md](efficiency_protocol.md)
- [junction_protocol.md](junction_protocol.md)
- [directive_lifecycle.md](directive_lifecycle.md) ⭐ **헌법**

### 워크플로우 (실행 지침)
- [daemon_workflow.md](daemon_workflow.md)
- [snapshot_workflow.md](snapshot_workflow.md)
- [sync_workflow.md](sync_workflow.md)
- [venv_workflow.md](venv_workflow.md)

### 시스템 (기술)
- [system_handshake.md](system_handshake.md) ⭐ 필수
- [system_sop.md](system_sop.md)
- [infrastructure_sentinel.md](infrastructure_sentinel.md)
- [data_asset_management.md](data_asset_management.md)

### 스킬 & 통합
- [skills_integration.md](skills_integration.md)
- [skill_unified_input.md](skill_unified_input.md)
- [uip_protocol.md](uip_protocol.md)

### 디자인 & 비주얼
- [visual_identity_guide.md](visual_identity_guide.md)
- [aesop_benchmark.md](aesop_benchmark.md)
- [design/style_guide.md](design/style_guide.md)

---

**Last Updated**: 2026-02-15
**Purpose**: 신규 에이전트 온보딩 및 시스템 일관성 유지
**Authority**: directive_lifecycle.md가 이 시스템의 헌법임