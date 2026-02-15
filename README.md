# 97LAYER OS (Sanctuary Ver 3.0)

> **상태**: Clean Architecture (Refactored)
> **최종 갱신**: 2026-02-16

## 📂 핵심 구조 (Core Architecture)

```
97layerOS/
├── core/                    # 🎯 실행 코드 (Unified)
│   ├── agents/             # AssetManager, AsyncAgentHub
│   ├── system/             # handoff, orchestrator, ralph_loop
│   ├── daemons/            # telegram_secretary (main)
│   ├── bridges/            # notebooklm, gdrive
│   └── utils/              # parsers, helpers
│
├── directives/              # 📜 철학 및 규칙
│   ├── IDENTITY.md         # Slow Life 브랜드 철학
│   └── system/SYSTEM.md    # 운영 프로토콜
│
├── knowledge/               # 📚 데이터 레이어
│   ├── signals/            # 입력 신호
│   ├── assets/             # 생성 자산 (registry.json)
│   ├── reports/            # 일일/주간 보고서
│   └── docs/               # 기술 문서
│
├── .infra/                  # 🔧 Container-only (gitignored)
│   ├── cache/
│   ├── logs/
│   └── tmp/
│
└── archive/                 # 📦 백업 및 레거시
```

## 🚀 실행 (Execution)

### 1. Telegram Bot 시작
```bash
./start_telegram.sh
```

### 2. 실시간 모니터링
```bash
./start_monitor.sh
```

### 3. Python Import
```python
from core.system.handoff import HandoffEngine
from core.agents.asset_manager import AssetManager
from core.bridges.notebooklm_bridge import NotebookLMBridge
from core.daemons.telegram_secretary import TelegramSecretary
```

## 🎯 주요 기능

### Phase 1-2: 기반 인프라
- ✅ Session Handoff (세션 연속성)
- ✅ Parallel Orchestrator (멀티에이전트 병렬)
- ✅ Ralph Loop (STAP 품질 검증)
- ✅ Asset Manager (자산 생명주기)
- ✅ Daily Automation (아침 브리핑, 저녁 리포트)

### Phase 3-4: Anti-Gravity Protocol
- ✅ YouTube Analyzer (NotebookLM 기반)
- ✅ NotebookLM MCP Integration (28 tools)
- ✅ RAG 질의 (요약, 인사이트, 브랜드 연결)
- ✅ Audio Overview (Google Gemini)

### Phase 5: Clean Architecture
- ✅ execution+system → core 통합
- ✅ Container-First 명확화
- ✅ Google Drive 동기화 준비
- ✅ Legacy dependency 제거

## 📡 Telegram Commands

```
/start       - 시스템 소개
/status      - 현재 상태
/report      - 오늘의 작업 보고
/analyze     - 마지막 신호 분석
/signal      - 새 신호 입력
/morning     - 아침 브리핑 (09:00 권장)
/evening     - 저녁 리포트 (21:00 권장)
/search      - 과거 지식 베이스 검색
/memo        - 빠른 메모 저장
/sync        - 클라우드 동기화
/youtube     - YouTube Anti-Gravity 분석 (NotebookLM)
```

## 🔄 Container-First 원칙

- **macOS 호스트**: 코드 작성, Git 관리, NotebookLM 인증
- **Podman 컨테이너**: Python 실행, Telegram Bot, MCP CLI

## 📝 세션 연속성

모든 작업 전후 `INTELLIGENCE_QUANTA.md` 자동 업데이트:
- 현재 상태
- 완료된 작업
- 다음 단계
- 문제 해결 기록

## 🎨 Slow Life 철학

- **속도보다 방향**: 급하게 하지 않고 구조부터 고민
- **효율보다 본질**: 당장 되는 것보다 장기적 유지보수성
- **완벽보다 진행**: 100% 아니어도 점진적으로 개선

## 🔐 환경 설정

`.env` 파일 필요:
```bash
TELEGRAM_BOT_TOKEN=your_token_here
ANTHROPIC_API_KEY=your_key_here
```

## 📚 Documentation

- [IDENTITY.md](directives/IDENTITY.md) - 브랜드 철학
- [SYSTEM.md](directives/system/SYSTEM.md) - 운영 프로토콜
- [INTELLIGENCE_QUANTA.md](knowledge/agent_hub/INTELLIGENCE_QUANTA.md) - 세션 연속성
- [TECHNICAL_SPEC.md](knowledge/docs/TECHNICAL_SPEC.md) - 기술 명세

---

> "Remove the Noise, Reveal the Essence" — 97layerOS
