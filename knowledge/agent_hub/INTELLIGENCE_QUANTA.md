# 🧠 INTELLIGENCE QUANTA - 지능 앵커

> **목적**: AI 세션이 바뀌어도 사고 흐름이 끊기지 않도록 보장하는 물리적 앵커
> **갱신 정책**: 덮어쓰기 (최신 상태만 유지)
> **마지막 갱신**: 2026-02-16 (하위 폴더 정리 + 통합 테스트 완료)

---

## 📍 현재 상태 (CURRENT STATE)

### [2026-02-16] Renewal PHASE 1-8 전체 완료 — Claude Code (Sonnet 4.5)

**아키텍처 버전**: Clean Architecture Ver 3.0 (Sanctuary Ver 3.0)

**진행률**: ✅ PHASE 1-8 전체 완료

- ✅ PHASE 1: 환경 정비 (requirements.txt 현행화, .driveignore 보완)
- ✅ PHASE 2: 규칙 동기화 (.ai_rules + GEMINI.md FILE CREATION POLICY)
- ✅ PHASE 3: md 파일 정리 (47개 → 25개, 루트 README.md 단일화)
- ✅ PHASE 4: core/ 구조 정리 (데몬 v6 단일화, bridges 최신화)
- ✅ PHASE 5: heartbeat.py (`core/system/heartbeat.py`)
- ✅ PHASE 6: signal_router.py (`core/system/signal_router.py`)
- ✅ PHASE 7: daily_routine.py APScheduler 연결 (`--scheduler`)
- ✅ PHASE 8: 철학 문서 리뉴얼 (IDENTITY.md v5.0 + SYSTEM.md v5.0)

---

## 🏗️ 현재 아키텍처

```
97layerOS/
├── core/
│   ├── agents/    (14개) — SA, CE, AD, CD, Ralph + 자산관리
│   ├── system/    (19개) — 핵심 엔진 (notebooklm_bridge 중복 제거 완료)
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
├── tests/                 (test_multi_agent_workflow.py 등)
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
- `ADMIN_TELEGRAM_ID` ❌ 미설정 — Nightguard V2 알림을 위해 필요
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

### 최우선: CE/AD 에이전트 프롬프트 정교화

SYSTEM.md v5.0에 CE/AD 기준이 명시됐지만 실제 에이전트 코드(`core/agents/ce_agent.py`, `core/agents/ad_agent.py`)에 NotebookLM 브랜드 가이드 쿼리 연동 필요:

```python
# ce_agent.py 개선 방향
brand_voice = notebooklm.query("97layer brand voice + WOOHWAHAE tone")
# ad_agent.py 개선 방향
visual_ref = notebooklm.query("WOOHWAHAE visual identity archival film")
```

### 장기 과제: Nightguard V2 Cookie Watchdog 활성화

`core/system/nightguard_v2.py`가 구현됐지만 실제 GCP VM에서 상시 실행 설정 필요.

---

## 🧭 장기 로드맵

```
[완료] Clean Architecture Ver 3.0
  ✅ Phase 1-4: 구조 정리
  ✅ Phase 5-7: Organic Ecosystem 코어 구현 (로컬 실행 검증 완료)
      - heartbeat.py: Mac↔GCP 상태 감지 ✅ 실행 확인
      - signal_router.py: 신호→큐 자동 라우팅 ✅ 7개 신호 처리
      - daily_routine.py: APScheduler 09:00/21:00 자동화 ✅ 브리핑 실행
  ✅ Phase 8: 철학 문서 리뉴얼 (IDENTITY.md v5.0 + SYSTEM.md v5.0)

[목표] THE CYCLE 완전 자동화
  텔레그램 입력 → 신호 저장 → 큐 라우팅 → 에이전트 처리
  → knowledge/ 저장 → Drive 동기화 → 텔레그램 보고 → 반복
```

## 🚀 실행 명령 (PYTHONPATH 필수)

```bash
# 전체 에코시스템 한번에 시작 (권장)
./start_ecosystem.sh

# 개별 실행 (PYTHONPATH 설정 필요)
export PYTHONPATH=/Users/97layer/97layerOS
python core/system/heartbeat.py             # heartbeat 데몬
python core/system/signal_router.py --watch # 신호 라우팅 감시
python core/system/daily_routine.py --scheduler  # 스케줄러

# 테스트 (1회 실행)
python core/system/heartbeat.py --once
python core/system/signal_router.py --once
python core/system/daily_routine.py --morning
```

---

## 📋 주요 파일 경로 레퍼런스

| 컴포넌트 | 경로 |
|---|---|
| 텔레그램 봇 실행 | `./start_telegram.sh` |
| 에코시스템 전체 시작 | `./start_ecosystem.sh` |
| 세션 핸드오프 | `core/system/handoff.py` |
| 큐 관리 | `core/system/queue_manager.py` |
| Nightguard | `core/system/nightguard_v2.py` |
| Drive 동기화 | `core/bridges/gdrive_sync.py` |
| NotebookLM | `core/system/notebooklm_bridge.py` (+ bridges/ 동기화됨) |
| 일일 루틴 + 스케줄러 | `core/system/daily_routine.py --scheduler` |
| Mac↔GCP 하트비트 | `core/system/heartbeat.py` |
| 신호→큐 라우팅 | `core/system/signal_router.py --watch` |
| 배포 가이드 | `knowledge/docs/deployment/DEPLOY.md` |
| 실행 컨텍스트 | `knowledge/system/execution_context.json` |
| 신호 처리 기록 | `knowledge/system/signal_router_processed.json` |

---

> "Remove the Noise, Reveal the Essence" — 97layerOS
