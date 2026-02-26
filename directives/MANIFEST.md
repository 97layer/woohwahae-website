# Filesystem Manifest — LAYER OS 서재 배치 규칙

> **Authority**: 모든 에이전트(Claude/Gemini/GPT)는 파일 생성 전 이 문서를 반드시 참조한다.
> **갱신 정책**: 구조 변경 시에만 수정. 덮어쓰기.
> **마지막 갱신**: 2026-02-26 (레거시 청산 완료 + knowledge/docs 추가)

---

## 4축 구조

```
97layerOS/
├── directives/     뇌 — 철학, 규칙, 규격
├── knowledge/      기억 — 데이터, 신호, 상태
├── core/           엔진 — 코드, 스크립트, 스킬, 테스트
├── website/        얼굴 — HTML/CSS/JS, 네비: Archive | Practice | About
└── .infra/         런타임 — 로그, 큐, nginx 설정 (.gitignore 대상)
```

---

## directives/ — 지침 (FROZEN / PROPOSE)

| 경로 | 용도 | 권한 |
|------|------|------|
| `THE_ORIGIN.md` | 브랜드 철학 SSOT | FROZEN |
| `AI_CONSTITUTION.md` | **AI 헌법 SSOT (모든 모델 공통)** | FROZEN |
| `SYSTEM.md` | 운영 프로토콜 | 덮어쓰기 |
| `MANIFEST.md` | 이 문서 | 덮어쓰기 |
| `README.md` | 인덱스 | 덮어쓰기 |
| `agents/{SA,CE,AD,CD,GEMINI}.md` | 에이전트별 판단 기준 (GEMINI.md = Gemini 전용 특수 규칙) | PROPOSE |
| `practice/` | 실행 규격 7개 (visual, language, content, audience, experience, service, references) | 참조용 |

---

## knowledge/ — 데이터 (8폴더)

| 경로 | 용도 |
|------|------|
| `agent_hub/` | 시스템 상태 (QUANTA, council_room, feedback_loop) |
| `signals/` | 원시 신호 (텔레그램/유튜브/메모). 하위: images/, wellness/ |
| `corpus/entries/` | 구조화 지식 (SA 분석 결과) |
| `clients/` | CRM 클라이언트 데이터 (Ritual Module) |
| `service/` | 서비스 아이템 카탈로그 (items.json) |
| `docs/` | 시스템 문서 (검증 보고서, 아키텍처 기록) |
| `reports/` | 리포트. 하위: daily/, growth/ |
| `system/` | 런타임 상태 + schemas/. work_lock, cache, registry 등 |

### 신호 형식 규칙
**허용**: `*.json` (명명: `{type}_{YYYYMMDD}_{HHMMSS}.json`)
**금지**: `*.md` 형식 신호

---

## core/ — 엔진 (8폴더)

| 경로 | 용도 |
|------|------|
| `agents/` | 에이전트 코드 (SA, AD, CE, CD, Code, Gardener) |
| `system/` | 파이프라인 + AI엔진 + bridges + modules 통합 (23개) |
| `daemons/` | 상주 서비스 (telegram, dashboard, nightguard) |
| `admin/` | 웹 대시보드 (Flask) |
| `backend/` | Flask 백엔드 서버 (VM 24/7 서비스, 포트 5000) |
| `scripts/` | 자동화 (deploy/, session, sync) |
| `skills/` | 에이전트 스킬 (`<name>/SKILL.md`) |
| `tests/` | 테스트 |

---

## website/ — 얼굴 (IA: Archive | Practice | About)

| 경로 | 용도 |
|------|------|
| `index.html` | 홈 |
| `about/` | About — 철학(philosophy), 서사(journey), 에디터(editor) |
| `archive/` | Archive — essay-NNN-slug/, magazine/, lookbook/ |
| `practice/` | Practice — atelier, direction, project, product, contact |
| `woosunho/` | 에디터 포트폴리오 (works/ 하위에 작업물) |
| `lab/` | 실험 (네비 미노출) |
| `assets/css/` | 스타일 |
| `assets/js/` | 스크립트 |
| `assets/media/` | 이미지 통합: brand/, archive/, reference/, uploads/, instagram/ |
| `backend/` | Flask 템플릿 |
| `_templates/` | 빌드 템플릿 |

### 에세이 명명 규칙
`{type}-{NNN}-{slug}/index.html` — type: essay, magazine, lookbook

**❌ website/ 내 .md 파일 생성 금지**

---

## 산출물 배치 규칙

| 산출물 | 경로 | 명명 |
|--------|------|------|
| 신호 (원시) | `knowledge/signals/` | `{type}_{YYYYMMDD}_{HHMMSS}.json` |
| 신호 (분석) | `knowledge/corpus/entries/` | `entry_{signal_id}.json` |
| 에세이 HTML | `website/archive/essay-{NNN}-{slug}/` | `index.html` |
| 리포트 | `knowledge/reports/` | morning/evening/audit만 허용 |
| 수익 데이터 | `knowledge/reports/growth/` | `growth_{YYYYMM}.json` |
| 클라이언트 | `knowledge/clients/` | `{client_id}.json` |
| 스키마 | `knowledge/system/schemas/` | `{name}.schema.json` |
| 에이전트 제안 | `knowledge/agent_hub/council_room.md` | append |

---

## 에이전트 Write Path

| 에이전트 | 허용 | 금지 |
|---------|------|------|
| **Claude Code** | 전체 | — |
| **Gemini CLI** | `knowledge/signals/` `knowledge/corpus/` `council_room.md`(append) | `core/` `directives/` `knowledge/system/` |
| **Gardener** | `knowledge/agent_hub/` `knowledge/system/guard_rules.json` | `core/` |
| **SA** | `knowledge/corpus/entries/` `.infra/queue/` | `core/` `website/` |
| **CE** | `website/archive/` `.infra/queue/` | `core/` `directives/` |
| **AD** | `.infra/queue/` | `core/` `website/` |
| **CD** | `council_room.md`(append) `.infra/queue/` | `core/` `directives/` |

---

## 금지 사항

- ❌ 루트(/)에 .md 생성 (CLAUDE.md, README.md 제외)
- ❌ 위 테이블에 없는 경로에 파일 생성
- ❌ filesystem_cache.json 미확인 상태로 파일 생성
- ❌ 임시 파일명 (temp_, untitled_, 무제)
- ❌ 금지 패턴: SESSION_SUMMARY_*, WAKEUP_REPORT*, DEPLOY_*, NEXT_STEPS*

---

## 강제 메커니즘

| Layer | 메커니즘 | 위치 |
|-------|---------|------|
| 1 | CLAUDE.md / .ai_rules | 루트 |
| 2 | Claude Code Hooks | `.claude/hooks/` |
| 3 | Claude Code Rules | `.claude/rules/` |
| 4 | Git Pre-Commit Hook | `.git/hooks/pre-commit` |
| 5 | Bootstrap Script | `core/scripts/session_bootstrap.sh` |
