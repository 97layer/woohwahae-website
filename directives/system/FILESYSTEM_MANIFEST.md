# Filesystem Manifest — LAYER OS 서재 배치 규칙

> **Authority**: 모든 에이전트(Claude/Gemini/GPT)는 파일 생성 전 이 문서를 반드시 참조한다.
> **갱신 정책**: 구조 변경 시에만 수정. 덮어쓰기.
> **마지막 갱신**: 2026-02-24

---

## 디렉토리 목적 정의

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
| `core/agents/` | 에이전트 코드. 1 에이전트 = 1 파일 |
| `core/modules/` | 레이어별 모듈 (L4 Ritual, L5 Growth) |
| `core/system/` | 파이프라인, 큐, 유틸리티 |
| `core/daemons/` | 상주 프로세스 (텔레그램, 감시자) |
| `core/bridges/` | 외부 API 연동 (NotebookLM 등) |
| `core/admin/` | 웹 대시보드 |

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

### website/ — 웹사이트 (정적 HTML/CSS/JS)

| 경로 | 용도 |
|------|------|
| `website/archive/` | 발행된 에세이 (issue-NNN-slug/) |
| `website/offering/` | 서비스 페이지 |
| `website/lab/` | 실험/프로토타입 (본 사이트 미포함) |
| `website/assets/` | CSS, JS, 이미지 |

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

### archive/ — 과거 코드 백업

| 경로 | 용도 |
|------|------|
| `archive/` | 레거시 코드 보관. 실행 대상 아님 |

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
