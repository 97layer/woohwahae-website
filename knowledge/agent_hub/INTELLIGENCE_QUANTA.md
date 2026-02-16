# 🧠 INTELLIGENCE QUANTA - 지능 앵커

> **목적**: AI 세션이 바뀌어도 사고 흐름이 끊기지 않도록 보장하는 물리적 앵커
> **갱신 정책**: 덮어쓰기 (최신 상태만 유지)
> **마지막 갱신**: 2026-02-16 (PHASE 9 완료 + SDK 마이그레이션 + systemd 서비스 파일)

---

## 📍 현재 상태 (CURRENT STATE)

### [2026-02-16] PHASE 1-9 전체 완료 — Claude Code (Sonnet 4.5)

**아키텍처 버전**: Clean Architecture Ver 3.0 (Sanctuary Ver 3.0)

**진행률**: ✅ THE CYCLE 완전 연결 완료

- ✅ PHASE 1: 환경 정비 (requirements.txt 현행화, .driveignore 보완)
- ✅ PHASE 2: 규칙 동기화 (.ai_rules + GEMINI.md FILE CREATION POLICY)
- ✅ PHASE 3: md 파일 정리 (47개 → 25개, 루트 README.md 단일화)
- ✅ PHASE 4: core/ 구조 정리 (데몬 v6 단일화, bridges 최신화)
- ✅ PHASE 5: heartbeat.py (`core/system/heartbeat.py`)
- ✅ PHASE 6: signal_router.py (`core/system/signal_router.py`)
- ✅ PHASE 7: daily_routine.py APScheduler 연결 (`--scheduler`)
- ✅ PHASE 8: 철학 문서 리뉴얼 (IDENTITY.md v5.0 + SYSTEM.md v5.0)
- ✅ PHASE 9: CE/AD NotebookLM RAG 연동 + 발행 단계 연결 (에이전트 → 텔레그램)

### THE CYCLE 연결 상태

```
입력    텔레그램 메시지 수신
  ↓     telegram_secretary.py → knowledge/signals/*.json
저장
  ↓     signal_router.py (10s polling) → QueueManager.create_task()
라우팅
  ↓     .infra/queue/tasks/pending/*.json
큐
  ↓     AgentWatcher (5s polling) → SA/AD/CE process_task()
에이전트 처리
  ↓     Gemini API + NotebookLM RAG (브랜드 보이스/시각 레퍼런스)
생성
  ↓     AgentWatcher._notify_admin() → Telegram Bot API
발행    ← ADMIN_TELEGRAM_ID 설정 시 자동 알림 ✅
  ↓
반복    signal_router 계속 대기 중
```

---

## 🏗️ 현재 아키텍처

```
97layerOS/
├── core/
│   ├── agents/    (14개) — SA, CE, AD, CD, Ralph + 자산관리
│   │               CE/AD: NotebookLM 브랜드 RAG 연동 (Phase 6.3)
│   ├── system/    (19개) — 핵심 엔진
│   │               agent_watcher.py: 완료 시 텔레그램 알림 (Phase 9)
│   ├── daemons/   (5개)  — telegram_secretary.py (v6 기반 단일화)
│   ├── bridges/   (3개)  — gdrive_sync, notebooklm_bridge (공식 단일 위치)
│   └── utils/     (5개)  — parsers, progress_analyzer
│
├── directives/
│   ├── IDENTITY.md        (v5.0 — WOOHWAHAE 철학 + THE CYCLE 완료)
│   └── system/SYSTEM.md   (v5.0 — Clean Arch Ver 3.0 + THE CYCLE 완료)
│
├── knowledge/
│   ├── agent_hub/         (QUANTA, council_room, feedback_loop)
│   ├── signals/           (텔레그램 신호 축적)
│   ├── reports/           (아침/저녁 자동 보고서)
│   ├── system/            (execution_context.json, signal_router_processed.json 등)
│   ├── docs/
│   │   ├── deployment/    (DEPLOY.md + 97layer-telegram.service)
│   │   ├── sessions/      (세션 기록 저장 위치)
│   │   └── archive/       (완료된 문서 보관)
│   └── assets/            (미디어 파일)
│
├── archive/
│   └── 2026-02-pre-refactor/  (레거시 코드 + telegram v1-v6)
│
├── tests/
│
└── .infra/                (컨테이너 런타임, logs/ — gitignored)
```

---

## ⚠️ 중요 결정사항

### Container-First 원칙 (확정)
- **로컬 MacBook**: 코드 작성, Git 관리, Google Drive 동기화 소스
- **Podman 컨테이너**: Python 실행, Telegram Bot, MCP CLI, .venv 관리
- **GCP VM**: 24/7 운영 (systemd로 상시 기동)
- `.venv/` 로컬에 절대 생성 금지 — Google Drive 동기화 대상이기 때문

### 환경변수 상태 (.env)
- `TELEGRAM_BOT_TOKEN` ✅ 설정됨
- `GEMINI_API_KEY` / `GOOGLE_API_KEY` ✅ 설정됨 (동일 키)
- `ANTHROPIC_API_KEY` ⚠️ 손상된 패턴 — 실제 키로 교체 필요
- `ADMIN_TELEGRAM_ID` ❌ 미설정 — **에이전트 완료 알림에 필수** (설정 시 THE CYCLE 발행 단계 활성화)
- `GOOGLE_DRIVE_FOLDER_ID` ❌ 미설정 — gdrive_sync.py를 위해 필요

### 파일 생성 정책 (.ai_rules에 명시됨)
- 루트(/)에 .md 생성 절대 금지
- 상태 파일 → 덮어쓰기 (QUANTA, IDENTITY, SYSTEM)
- 이력 파일 → 추가(append) (council_room, feedback_loop)
- 산출물 → 날짜별 (knowledge/reports/)
- 부산물 → 생성 금지 (SESSION_SUMMARY, WAKEUP_REPORT 등)

### Claude ↔ Antigravity 충돌 방지
- `.ai_rules`와 `GEMINI.md` 양쪽에 동일 FILE CREATION POLICY 적용됨
- 공통 SSOT: `INTELLIGENCE_QUANTA.md` (이 파일)

---

## 🔒 작업 잠금 상태

**현재 잠금**: None

---

## 🎯 다음 세션 작업

### 최우선: .env 값 채우기 (사용자 직접)

THE CYCLE 발행 단계 완전 활성화를 위해:
```
ADMIN_TELEGRAM_ID=✅ 설정됨 (에이전트 완료 알림 활성화됨)
GOOGLE_DRIVE_FOLDER_ID=❌ 미설정 — gdrive_sync.py를 위해 필요
ANTHROPIC_API_KEY=⚠️ 손상된 패턴 — 실제 키로 교체 필요
```

### 중기: Nightguard V2 GCP systemd 등록 (서비스 파일 완성)

`knowledge/docs/deployment/97layer-nightguard.service` ✅ 작성 완료.
`97layer-ecosystem.service`도 작성 완료 (THE CYCLE 전체 스택용).
GCP VM에서 USERNAME_PLACEHOLDER 치환 후 `systemctl enable/start` 실행만 남음.
→ `DEPLOY.md` 6번 섹션 참고.

### 장기: Python 3.9 → 3.11 컨테이너 업그레이드

google.generativeai → google.genai SDK 마이그레이션 ✅ 완료.
남은 FutureWarning은 Python 3.9 EOL 문제. GCP VM .venv 재생성 시 python3.11 사용 권장.

---

## 🧭 장기 로드맵

```
[완료] Clean Architecture Ver 3.0
  ✅ Phase 1-4: 구조 정리
  ✅ Phase 5-7: Organic Ecosystem 코어 구현
  ✅ Phase 8: 철학 문서 리뉴얼 (IDENTITY.md v5.0 + SYSTEM.md v5.0)
  ✅ Phase 6.3: CE/AD NotebookLM 브랜드 RAG 연동
  ✅ Phase 9: THE CYCLE 완전 연결
      - agent_watcher: 완료 시 텔레그램 알림 (_notify_admin + _build_summary)
      - start_ecosystem.sh: SA+AD+CE 에이전트 자동 시작 포함

[현재 상태] THE CYCLE 코드 완전 연결 ✅ + SDK 마이그레이션 ✅
  ✅ google.generativeai → google.genai (SA/AD/CE 전체)
  ✅ Nightguard V2 systemd 서비스 파일 작성 완료
  ✅ ADMIN_TELEGRAM_ID 설정됨 → 에이전트 완료 알림 즉시 활성

[다음 목표] GCP 24/7 배포
  로컬 동작 검증 완료 → GCP VM에서 systemd enable/start (DEPLOY.md §6 참고)
  → ./start_telegram.sh + ./start_ecosystem.sh 실행 → 텔레그램 메시지 → 알림 수신
```

## 🚀 실행 명령

```bash
# 전체 에코시스템 한번에 시작 (권장)
# heartbeat + signal_router + scheduler + SA + AD + CE 자동 기동
./start_ecosystem.sh

# 텔레그램 봇만 실행
./start_telegram.sh

# 개별 테스트
export PYTHONPATH=/Users/97layer/97layerOS
python core/system/heartbeat.py --once
python core/system/signal_router.py --once
python core/system/daily_routine.py --morning
python core/agents/sa_agent.py --test
python core/agents/ad_agent.py --test
python core/agents/ce_agent.py --test
```

---

## 📋 주요 파일 경로 레퍼런스

| 컴포넌트 | 경로 |
|---|---|
| THE CYCLE 전체 시작 | `./start_ecosystem.sh` |
| 텔레그램 봇 실행 | `./start_telegram.sh` |
| 에이전트 완료 알림 | `core/system/agent_watcher.py` (_notify_admin) |
| 세션 핸드오프 | `core/system/handoff.py` |
| 큐 관리 | `core/system/queue_manager.py` |
| Nightguard | `core/system/nightguard_v2.py` |
| Drive 동기화 | `core/bridges/gdrive_sync.py` |
| NotebookLM | `core/bridges/notebooklm_bridge.py` |
| 일일 루틴 + 스케줄러 | `core/system/daily_routine.py --scheduler` |
| Mac↔GCP 하트비트 | `core/system/heartbeat.py` |
| 신호→큐 라우팅 | `core/system/signal_router.py --watch` |
| 배포 가이드 | `knowledge/docs/deployment/DEPLOY.md` |
| Nightguard systemd | `knowledge/docs/deployment/97layer-nightguard.service` |
| Ecosystem systemd | `knowledge/docs/deployment/97layer-ecosystem.service` |
| 실행 컨텍스트 | `knowledge/system/execution_context.json` |

---

> "Remove the Noise, Reveal the Essence" — 97layerOS
