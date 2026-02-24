# Filesystem Manifest — LAYER OS 서재 배치 규칙

> **Authority**: 모든 에이전트(Claude/Gemini/GPT)는 파일 생성 전 이 문서를 반드시 참조한다.
> **갱신 정책**: 구조 변경 시에만 수정. 덮어쓰기.
> **마지막 갱신**: 2026-02-24 (archive/ 삭제, core/system/ 분류 명시, website/ 정책 갱신)

---

## 디렉토리 목적 정의

### 경계 규칙: directives/ vs knowledge/docs/

> **판단 기준 한 줄**: "앞으로도 유효한 규범인가?" → `directives/` | "과거 기록이나 참고 자료인가?" → `knowledge/docs/`

| 넣을 것 | directives/ | knowledge/docs/ |
|--------|------------|----------------|
| 브랜드 철학, 에이전트 판단 기준 | ✅ | ❌ |
| 운영 규칙, 배치 규칙 (이 문서) | ✅ | ❌ |
| 기술 구현 가이드, 코딩 규칙 | ❌ | ✅ (`docs/system/`) |
| 배포 절차, 인프라 문서 | ❌ | ✅ (`docs/deployment/`) |
| 세션 기록, 리포트 | ❌ | ✅ (`docs/sessions/`, `reports/`) |
| 과거 문서 보관 | ❌ | ✅ (`docs/archive/`) |

**`knowledge/docs/system/` 파일명 규칙**: `UPPER_SNAKE_CASE.md` (예: `CODING_RULES.md`, `ENFORCEMENT_SYSTEM.md`)

---

### directives/ — 지침 (변경 빈도 낮음)

| 경로 | 용도 | 권한 |
|------|------|------|
| `directives/IDENTITY.md` | 브랜드 철학 핵심 선언 | FROZEN |
| `directives/system/SYSTEM.md` | 운영 규칙 | 덮어쓰기 |
| `directives/system/FILESYSTEM_MANIFEST.md` | 이 문서 (서재 배치 규칙) | 덮어쓰기 |
| `directives/agents/{SA,CE,AD,CD}.md` | 에이전트별 판단 기준 | PROPOSE |
| `directives/brand/` | Brand OS 상세 문서 11개 | 참조용 |
| `directives/README.md` | directives 인덱스 | 덮어쓰기 |

### core/ — 코드 (실행 가능한 Python만)

| 경로 | 용도 |
|------|------|
| `core/agents/` | 에이전트 코드. 1 에이전트 = 1 파일. **새 에이전트 = 여기** |
| `core/modules/` | 레이어별 모듈 (L4 Ritual, L5 Growth) |
| `core/system/` | 파이프라인, 큐, AI엔진, 유틸리티. **아래 카테고리 참조** |
| `core/daemons/` | 상주 프로세스 (24/7 실행). **새 데몬 = 여기** |
| `core/bridges/` | 외부 API 연동 (NotebookLM, GDrive 등) |
| `core/admin/` | 웹 대시보드 (Flask) |

#### core/system/ 카테고리 가이드 (파일 추가 시 참고)

| 카테고리 | 해당 파일 예시 | 새 파일 기준 |
|---------|--------------|-------------|
| 파이프라인 | pipeline_orchestrator, queue_manager, signal_router, handoff | 신호→에세이 흐름 관련 |
| AI 엔진 | gemini_engine, knowledge_rag, intent_classifier, conversation_engine | LLM 호출/RAG 관련 |
| 크롤러 | scout_crawler, youtube_analyzer, image_analyzer | 외부 데이터 수집 |
| 모니터링 | agent_logger, agent_watcher, heartbeat, token_tracker | 상태 감시/로깅 |
| 퍼블리셔 | content_publisher, auto_reporter | 콘텐츠 최종 발행 |
| 관리도구 | env_validator, directive_editor, corpus_manager | 시스템 관리 유틸 |

### knowledge/ — 지식 (데이터, 기록, 상태)

| 경로 | 용도 |
|------|------|
| `knowledge/agent_hub/` | 시스템 상태 앵커 (QUANTA, council_room) |
| `knowledge/signals/` | 원시 신호 (텔레그램/유튜브/텍스트) |
| `knowledge/corpus/entries/` | 구조화된 지식 풀 (SA 분석 결과) |
| `knowledge/brands/` | 브랜드 도시에 (Brand Scout 결과) |
| `knowledge/sources/` | 원문 소스 (web/pdf/gdrive/youtube) |
| `knowledge/reports/` | 날짜별 리포트 |
| `knowledge/docs/` | 기술 문서, 세션 기록 |
| `knowledge/docs/sessions/` | 세션별 기록 |
| `knowledge/docs/archive/` | 과거 문서 보관 |
| `knowledge/system/` | 런타임 상태 (work_lock, cache, registry) |
| `knowledge/system/schemas/` | JSON 스키마 정의 |
| `knowledge/clients/` | CRM 클라이언트 데이터 (Ritual Module) |
| `knowledge/reports/growth/` | 월별 성장 지표 JSON (Growth Module) |
| `knowledge/long_term_memory.json` | 장기 기억 |

### website/ — 웹사이트 (정적 HTML/CSS/JS만)

| 경로 | 용도 |
|------|------|
| `website/` 루트 | 주요 공개 페이지 (index, about, contact 등) + 리다이렉터 |
| `website/archive/` | 발행된 에세이 (issue-NNN-slug/) |
| `website/offering/` | 서비스 상세 페이지. **새 서비스 페이지 = 여기** |
| `website/lab/` | 실험/프로토타입 (본 사이트 미포함) |
| `website/assets/` | CSS, JS, 이미지 |

**❌ website/ 내 .md 파일 생성 금지** → `knowledge/docs/`에 넣을 것

### scripts/ — 자동화 스크립트

| 경로 | 용도 |
|------|------|
| `scripts/session_bootstrap.sh` | 세션 시작 자동화 |
| `scripts/session_handoff.sh` | 세션 종료 핸드오프 |
| `scripts/deploy/` | 배포 스크립트 모음 |

### skills/ — 에이전트 스킬 정의

| 경로 | 용도 |
|------|------|
| `skills/<skill_name>/SKILL.md` | 스킬별 매뉴얼 |

### knowledge/signals/ — 신호 형식 규칙

**허용**: `*.json` (명명규칙: `{type}_{YYYYMMDD}_{HHMMSS}.json`)
**금지**: `*.md` 형식 신호 파일 (구버전 형식. `knowledge/docs/archive/legacy_signals/`에 보관)

---

## 산출물 배치 규칙

| 산출물 유형 | 저장 경로 | 명명 규칙 |
|---|---|---|
| 신호 (원시) | `knowledge/signals/` | `{type}_{YYYYMMDD}_{HHMMSS}.json` |
| 신호 (분석) | `knowledge/corpus/entries/` | `entry_{signal_id}.json` |
| 에세이 HTML | `website/archive/issue-{NNN}-{slug}/` | `index.html` |
| 브랜드 도시에 | `knowledge/brands/{slug}/` | `profile.json` |
| 리포트 | `knowledge/reports/` | `{type}_{YYYYMMDD}.md` |
| 수익 데이터 | `knowledge/reports/growth/` | `growth_{YYYYMM}.json` |
| 세션 기록 | `knowledge/docs/sessions/` | `{YYYYMMDD}_{agent_id}.md` |
| 자산 등록 | `knowledge/system/asset_registry.json` | append |
| 에이전트 제안 | `knowledge/agent_hub/council_room.md` | append |
| 클라이언트 | `knowledge/clients/` | `{client_id}.json` |
| JSON 스키마 | `knowledge/system/schemas/` | `{name}.schema.json` |

---

## 금지 사항

- **❌ 루트(/)에 .md 파일 생성** (CLAUDE.md, README.md 제외)
- **❌ 위 테이블에 없는 경로에 파일 생성**
- **❌ filesystem_cache.json 미확인 상태로 파일 생성**
- **❌ 임시 파일명** (`temp_`, `untitled_`, `무제` 등)
- **❌ 금지 파일명 패턴**: `SESSION_SUMMARY_*`, `WAKEUP_REPORT*`, `DEEP_WORK_PROGRESS*`, `DEPLOY_*`, `NEXT_STEPS*`

---

## 강제 메커니즘

| Layer | 메커니즘 | 위치 |
|-------|---------|------|
| 1 | CLAUDE.md / .ai_rules 규칙 | 루트 |
| 2 | Claude Code Hooks (validate-path.sh) | `.claude/hooks/` |
| 3 | Claude Code Rules | `.claude/rules/` |
| 4 | Git Pre-Commit Hook | `.git/hooks/pre-commit` |
| 5 | Bootstrap Script | `scripts/session_bootstrap.sh` |
