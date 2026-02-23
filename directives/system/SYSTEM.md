# SYSTEM — 통합 운영 지침 v6.0

> **버전**: 6.0
> **갱신**: 2026-02-24
> **원칙**: 이 문서는 덮어쓰기(Overwrite) 방식으로 관리된다. 모든 AI 에이전트의 행동 기준.

---

## I. Architecture (LAYER OS)

```
97layerOS/
├── core/              실행 코드 (Python)
│   ├── agents/        에이전트 코드 (1 에이전트 = 1 파일)
│   ├── system/        오케스트레이션 엔진
│   ├── daemons/       장기 실행 서비스 (Telegram Bot)
│   ├── bridges/       외부 연동 (GDrive, NotebookLM)
│   └── admin/         웹 대시보드
│
├── directives/        철학 및 규칙 (변경 빈도 낮음)
│   ├── IDENTITY.md    브랜드 철학 — SSOT
│   ├── system/        운영 프로토콜 (본 문서, FILESYSTEM_MANIFEST)
│   ├── agents/        에이전트별 판단 기준 (SA/CE/AD/CD.md)
│   └── brand/         Brand OS 상세 규칙 (10개 문서)
│
├── knowledge/         데이터 레이어 (축적)
│   ├── agent_hub/     INTELLIGENCE_QUANTA.md, council_room.md
│   ├── signals/       입력 신호 (모든 소스 → 통합 JSON)
│   ├── corpus/entries/ 구조화된 지식 풀 (SA 분석 결과)
│   ├── brands/        브랜드 도시에 (Brand Scout)
│   ├── reports/       자동 생성 보고서
│   ├── system/        런타임 상태 + JSON 스키마
│   ├── clients/       CRM 데이터 (Ritual Module)
│   └── docs/          기술 문서 + 세션 기록
│
├── website/           웹사이트 (정적 HTML/CSS/JS)
│   ├── archive/       발행된 에세이 (issue-NNN-slug/)
│   ├── offering/      서비스 페이지
│   └── assets/        CSS, JS, 이미지
│
├── scripts/           자동화 스크립트
│   ├── deploy/        배포 스크립트
│   └── session_*.sh   세션 시작/종료
│
└── archive/           레거시 + 완료 코드
```

### 5-Layer OS 매핑

| Layer | 이름 | 에이전트 | directives | brand/ |
|-------|------|----------|------------|--------|
| L1 | Philosophy | CD | CD.md | foundation.md |
| L2 | Design | AD | AD.md | design_tokens.md |
| L3 | Content | CE + SA | CE.md, SA.md | content_system.md, voice_tone.md |
| L4 | Service | Ritual Module | — | service_ritual.md |
| L5 | Business | Growth Module | — | roadmap.md |
| X | Quality | Ralph (QA) | — | content_system.md §4 |
| X | Evolution | Gardener | — | — |

### 3환경 분리
- **로컬 MacBook**: 코드 작성, Git, Google Drive 동기화 소스
- **GCP VM**: 24/7 상시 운영 (systemd: 97layer-telegram)
- **CLI/에이전트**: Claude Code, Gemini, GPT — 모델 독립

---

## II. THE CYCLE — 핵심 운영 철학

모든 시스템 설계는 THE CYCLE을 구현하기 위해 존재한다:

```
입력 (텔레그램/신호)
  ↓  [signal_router.py — 신호 감지 → 큐 투입]
저장 (knowledge/signals/)
  ↓  [queue_manager.py — 에이전트 분기]
연결 (NotebookLM RAG — 패턴, 브랜드 컨텍스트)
  ↓  [SA → CE → Ralph → AD → CD 에이전트 체인]
생성 (WOOHWAHAE 콘텐츠)
  ↓  [content_publisher.py — 웹 발행]
발행 (승인된 자산)
  ↓
다시 입력
```

---

## III. 에이전트 프레임워크

### 역할 정의

| 에이전트 | 역할 | 판단 기준 | brand/ 참조 |
|---|---|---|---|
| **SA** | 신호 분석, 패턴 인식 | 데이터 기반, NotebookLM | audience.md |
| **CE** | 텍스트 톤앤매너, 서사 구조화 | voice_tone.md, content_system.md | voice_tone.md, content_system.md |
| **AD** | 시각 컨셉, 여백/모노크롬 검증 | design_tokens.md | design_tokens.md |
| **CD** | 최종 승인, 브랜드 일관성 | 5 Pillars, 72시간 규칙 | foundation.md |
| **Ralph** | 결과물 품질 검증 (STAP) | content_system.md §4 | content_system.md |
| **Gardener** | 군집 성숙도 점검, 개념 진화 | corpus 통계 기반 | — |

→ CE/AD 인라인 프롬프트는 brand/ 문서로 대체됨. 에이전트는 brand/ 문서를 직접 로드.

---

## IV. Operational Protocols

### 1. Session Start Protocol

매 세션 시작 시 반드시:
```bash
cat knowledge/agent_hub/INTELLIGENCE_QUANTA.md
cat knowledge/system/work_lock.json
cat directives/system/FILESYSTEM_MANIFEST.md  # 파일 생성 전 필수
```

### 2. File Creation Policy

| 파일 종류 | 정책 | 위치 |
|---|---|---|
| 상태(State) | 덮어쓰기 | INTELLIGENCE_QUANTA.md, IDENTITY.md, SYSTEM.md |
| 이력(History) | 추가(append) | council_room.md, feedback_loop.md |
| 산출물(Output) | 날짜별 생성 | knowledge/reports/ |
| 부산물(Byproduct) | **생성 금지** | SESSION_SUMMARY, WAKEUP_REPORT 등 |

**배치 규칙**: `directives/system/FILESYSTEM_MANIFEST.md` 참조 필수.
**루트(/)에 .md 파일 생성 절대 금지.**

### 3. Quality Gate (Ralph/QA)

모든 콘텐츠 산출물은 Ralph STAP 검증 통과 후 승인:
→ 상세: `brand/content_system.md` §4

### 4. Work Lock

작업 전:
```bash
cat knowledge/system/work_lock.json
```
잠금 존재 시 → 중단. 다른 에이전트 작업 중.

### 5. Handoff Protocol

세션 종료 시:
```bash
./scripts/session_handoff.sh "agent-id" "요약" "다음태스크"
```

---

## V. THE CYCLE 컴포넌트

| 컴포넌트 | 실행 명령 | 역할 |
|---|---|---|
| Telegram Bot | `./start_telegram.sh` | 신호 수신 + 에이전트 인터페이스 |
| Signal Router | `python core/system/signal_router.py --watch` | 신호→큐 자동 라우팅 |
| Heartbeat | `python core/system/heartbeat.py` | Mac↔GCP 상태 감지 |
| Nightguard | `python core/system/nightguard_v2.py` | 쿠키/API/서비스 감시 |
| Drive Sync | `python core/bridges/gdrive_sync.py --all` | Google Drive 동기화 |

---

## VI. Agent Collaboration

### Council Room
시스템 개선 제안은 `knowledge/agent_hub/council_room.md`에 append 방식으로 기록.
CD 승인 또는 에이전트 3인 이상 동의 시 구현.

### Shadow Review
- AD 시각물 → CE 교차 검토
- CE 텍스트 → Ralph 검증

---

## VII. Technical Standards

- **언어**: Python 3.11+
- **로깅**: f-string 금지, lazy formatting `%` 사용
- **의존성**: requirements.txt 기준
- **환경변수**: `.env` — gitignored
- **모델 독립**: 상태는 파일 기반. 어떤 AI든 QUANTA + MANIFEST 읽으면 시작 가능.

---

**Last Updated**: 2026-02-24
**Version**: 6.0 (Brand OS 참조 체계 + 5-Layer 매핑 + Ralph 삽입)
**갱신 정책**: 덮어쓰기. 버전 이력은 Git에 위임.
**연계 문서**: [IDENTITY.md](../IDENTITY.md) — 브랜드 철학 SSOT | [FILESYSTEM_MANIFEST.md](FILESYSTEM_MANIFEST.md) — 배치 규칙
