# SYSTEM — 통합 운영 지침 v5.0

> **버전**: 5.0
> **갱신**: 2026-02-16
> **원칙**: 이 문서는 덮어쓰기(Overwrite) 방식으로 관리된다. 모든 AI 에이전트의 행동 기준.

---

## I. Architecture (현재 구조 — Ver 3.0)

```
97layerOS/
├── core/              실행 코드 (Python)
│   ├── agents/        5 에이전트 + 자산 관리
│   ├── system/        오케스트레이션 엔진
│   ├── daemons/       장기 실행 서비스 (Telegram Bot)
│   ├── bridges/       외부 연동 (GDrive, NotebookLM)
│   └── utils/         유틸리티
│
├── directives/        철학 및 규칙 (변경 빈도 낮음)
│   ├── IDENTITY.md    브랜드 철학 — SSOT
│   └── system/SYSTEM.md  운영 프로토콜 — 본 문서
│
├── knowledge/         데이터 레이어 (축적)
│   ├── agent_hub/     INTELLIGENCE_QUANTA.md, council_room.md
│   ├── signals/       입력 신호 (텔레그램 수신)
│   ├── reports/       자동 생성 보고서 (아침/저녁)
│   ├── system/        런타임 상태 (execution_context.json 등)
│   └── docs/          기술 문서 + 배포 가이드
│
├── archive/           레거시 + 완료 코드
└── .infra/            컨테이너 런타임 (gitignored)
```

**3환경 분리:**
- **로컬 MacBook**: 코드 작성, Git, Google Drive 동기화 소스
- **Podman 컨테이너**: Python 실행, Telegram Bot, MCP CLI, .venv
- **GCP VM**: 24/7 상시 운영 (systemd: 97layer-telegram)

---

## II. THE CYCLE — 핵심 운영 철학

모든 시스템 설계는 THE CYCLE을 구현하기 위해 존재한다:

```
입력 (텔레그램/신호)
  ↓  [signal_router.py — 신호 감지 → 큐 투입]
저장 (knowledge/signals/)
  ↓  [queue_manager.py — 에이전트 분기]
연결 (NotebookLM RAG — 패턴, 브랜드 컨텍스트)
  ↓  [SA → CE → AD → CD 에이전트 체인]
생성 (WOOHWAHAE 콘텐츠)
  ↓  [ralph_loop.py — 품질 검증]
발행 (승인된 자산)
  ↓  [gdrive_sync.py — Drive 동기화]
다시 입력
```

---

## III. 5-Agent Framework

### 역할 정의

| 에이전트 | 역할 | 판단 기준 |
|---|---|---|
| **SA (Strategy Analyst)** | 신호 분석, 패턴 인식, 인사이트 도출 | 데이터 기반, NotebookLM 쿼리 |
| **CE (Chief Editor)** | 텍스트 톤앤매너, 서사 구조화 | IDENTITY.md Voice 기준 |
| **AD (Art Director)** | 시각 컨셉, 여백/모노크롬 검증 | IDENTITY.md Visual 기준 |
| **CD (Creative Director)** | 최종 승인, 브랜드 일관성 | 72시간 규칙, WOOHWAHAE Pillars |
| **Ralph (Quality Gate)** | 결과물 품질 검증 | STAP 기준 (Stop-Task-Assess-Process) |

### CE 프롬프트 기준 (텍스트 생성 시 필수 참조)

```
- 길이: 50-100단어
- 톤: 성찰적, 절제된
- 허용 키워드: 본질, 기록, 순간, 흔적
- 금지: 과장, 감정 과잉, 트렌드 추종, 알고리즘 어휘
- NotebookLM에서 "97layer brand voice" 쿼리 후 생성
```

### AD 비주얼 컨셉 기준 (이미지 생성 프롬프트 시 필수 참조)

```
- 톤: muted, desaturated
- 질감: 35mm 필름 그레인
- 구도: rule of thirds, negative space
- 참조: Aesop 브랜드 시각 언어
- 금지: 과포화, 플래시 조명, 복잡한 배경
```

---

## IV. Operational Protocols

### 1. Session Start Protocol

매 세션 시작 시 반드시:
```bash
cat knowledge/agent_hub/INTELLIGENCE_QUANTA.md
```
이 파일이 시스템의 현재 상태다. 무시하면 CRITICAL VIOLATION.

### 2. File Creation Policy

| 파일 종류 | 정책 | 위치 |
|---|---|---|
| 상태(State) | 덮어쓰기 | INTELLIGENCE_QUANTA.md, IDENTITY.md, SYSTEM.md |
| 이력(History) | 추가(append) | council_room.md, feedback_loop.md |
| 산출물(Output) | 날짜별 생성 | knowledge/reports/ |
| 부산물(Byproduct) | **생성 금지** | SESSION_SUMMARY, WAKEUP_REPORT 등 |

**루트(/)에 .md 파일 생성 절대 금지.**

### 3. Quality Gate (Ralph Loop)

모든 콘텐츠 산출물은 Ralph Loop 통과 후 승인:
- Stop: 완성 조건 명확화
- Task: 실행
- Assess: 결과 평가 (품질 점수 60+ 기준)
- Process: 피드백 반영 또는 승인

### 4. Work Lock

작업 전:
```bash
cat knowledge/system/work_lock.json
```
잠금 존재 시 → 중단. 다른 에이전트 작업 중.

### 5. Handoff Protocol

세션 종료 시:
```bash
python core/system/handoff.py --handoff
```

---

## V. THE CYCLE 컴포넌트 실행 참조

| 컴포넌트 | 실행 명령 | 역할 |
|---|---|---|
| Telegram Bot | `./start_telegram.sh` | 신호 수신 + 에이전트 인터페이스 |
| Signal Router | `python core/system/signal_router.py --watch` | 신호→큐 자동 라우팅 |
| Daily Scheduler | `python core/system/daily_routine.py --scheduler` | 09:00/21:00 자동 브리핑 |
| Heartbeat | `python core/system/heartbeat.py` | Mac↔GCP 상태 감지 |
| Nightguard | `python core/system/nightguard_v2.py` | 쿠키/API/서비스 감시 |
| Drive Sync | `python core/bridges/gdrive_sync.py --all` | Google Drive 동기화 |

---

## VI. Agent Collaboration

### Synapse Bridge
모든 에이전트는 작업 시작/종료 시 `knowledge/agent_hub/synapse_bridge.json` 업데이트.

### Council Room
시스템 개선 제안은 `knowledge/agent_hub/council_room.md`에 append 방식으로 기록.
CD 승인 또는 에이전트 3인 이상 동의 시 TD가 구현.

### Shadow Review
- TD 코드 → SA 교차 검토
- AD 시각물 → CE 교차 검토

---

## VII. Technical Standards

- **언어**: Python 3.11+ (Podman 컨테이너 내)
- **로깅**: f-string 금지, lazy formatting `%` 사용 (pylint 준수)
- **의존성**: requirements.txt 기준 — 컨테이너 내 `pip install -r requirements.txt`
- **환경변수**: `.env` — gitignored, driveignored
- **Container-First**: 모든 Python 실행은 Podman 컨테이너 내부에서. 로컬은 관제/편집 전용.

---

**Last Updated**: 2026-02-16
**Version**: 5.0 (Ver 3.0 아키텍처 + THE CYCLE 통합)
**갱신 정책**: 덮어쓰기. 버전 이력은 Git에 위임.
**연계 문서**: [IDENTITY.md](../IDENTITY.md) — 브랜드 철학 SSOT
