# 🧠 INTELLIGENCE QUANTA - 지능 앵커

> **목적**: 어떤 모델/세션이 오더라도 사고 흐름이 끊기지 않도록 보장하는 물리적 앵커
> **갱신 정책**: 덮어쓰기 (최신 상태만 유지). Gardener가 매일 자동 갱신.
> **마지막 갱신**: 2026-02-24 (auto-update by auto-session)

---

## 👤 이 사람에 대해 (세션 리셋 후 즉시 복원)

- **이름**: WOOSUNHO (순호). 반지하 8평 원룸. WOOHWAHAE 헤어 아틀리에 운영.
- **공개 정체성**: Editor & Chief Curator (@woosunhokr)
- **지향**: 슬로우라이프. 속도보다 본질. Magazine B 방식의 브랜드.
- **말투**: 짧고 직관적. "이거 낫지 않겠어?", "오케이" — 이게 대화 단위
- **기술 이해도**: 높음. 구조 설명 불필요. 거짓 구현에 민감.
- **싫어하는 것**: 감탄사("오!", "훌륭합니다"), 장황한 설명, 빈 공감

### 에이전트 포지션
- 비서가 아닌 동료. 받아쓰지 말고 판단하고 제안.
- 짧게 핵심만. 할 수 없으면 없다고 — 이게 신뢰의 기반.

---

## 🧬 개념 사고 수준 (Gardener 자동 갱신 — 세션 연속성 핵심)

> 이 섹션이 모델 비의존성의 핵심. 어떤 AI가 와도 이걸 읽으면 현재 사고 수준에서 출발.

*아직 Gardener가 충분한 Corpus entry를 분석하지 않음. 신호가 쌓이면 자동 갱신.*

---

## 🏗️ 시스템 아키텍처

**버전**: Ver 7.4 — woohwahae.kr 슈퍼앱 통합 완료 (고객 포털 + 사전상담 + Growth Dashboard)

```
신호 유입 (텔레그램/CLI/유튜브/URL/이미지/PDF/크롤러 — 전부 통합 스키마)
    ↓  signal.schema.json 통일
knowledge/signals/{type}_{YYYYMMDD}_{HHMMSS}.json
    ↓  SignalRouter → SA 큐 자동 전달
SA 분석 → knowledge/corpus/entries/
    ↓  Gardener 매일 군집 성숙도 점검
군집 성숙 (동일 테마 5개+ / 72시간+ 분포)
    ↓
CE Agent: corpus RAG → Magazine B 에세이
    ↓
woohwahae.kr/archive/ 발행
```

**파이프라인 재설계**: SA → CE → Ralph(인라인 QA) → AD → CD. Gardener 주도 발행.

### 인프라
- **GCP VM**: `97layer-vm` = `136.109.201.201` (Static IP)
- **앱 경로**: `/home/skyto5339_gmail_com/97layerOS/`
- **배포**: `scp [파일] 97layer-vm:/home/.../97layerOS/[경로]/`
- **서비스**: 97layer-telegram / 97layer-ecosystem / 97layer-gardener

### 환경변수
- TELEGRAM_BOT_TOKEN ✅ / GEMINI_API_KEY ✅ / ANTHROPIC_API_KEY ✅
- ADMIN_TELEGRAM_ID=7565534667 ✅

### 핵심 설계 원칙
- **신뢰 기반**: 할 수 없는 건 "못 한다". 거짓 구현 절대 금지
- **FROZEN 파일**: IDENTITY.md, CD.md, brand/story.md — `/confirm [token]` 확인 필요
- **모델 비의존성**: 모든 성장 기록이 파일로. 어떤 모델도 QUANTA만 읽으면 동일 수준 출발

---

## ✅ 완료된 작업 (누적)

- ✅ NotebookLM 브리지 재구축 (subprocess → HTTP API)
- ✅ conversation_engine 로컬 RAG (12초 → 0.01초)
- ✅ telegram_secretary hallucination 전수 제거
- ✅ Gardener 3단계 권한 (FROZEN/PROPOSE/AUTO)
- ✅ 양방향 소통 구현 (DirectiveEditor + `/confirm` 토큰)
- ✅ GCP Static IP 고정 (136.109.201.201)
- ✅ Drive 동기화 구축 (rclone + gdrive)
- ✅ Magazine B 방향 전환 + Brand Scout 에이전트
- ✅ **하니스 엔지니어링** (2026-02-18): CLAUDE.md 헌법 인라인, 4-Layer Enforcement, skills/
- ✅ **Pipeline 데드락 수정** (2026-02-18): _scan_new_signals() 구현
- ✅ **Corpus 아키텍처** (2026-02-18): 즉시발행 → 군집 기반 발행 전환
- ✅ **개념 진화 기록** (2026-02-18): Gardener _evolve_concept_memory() 구현
- ✅ **venv notebooklm-py 설치** (2026-02-19): GCP VM .venv에 notebooklm-py 패키지 설치. 시스템 Python이 아닌 venv 기준으로 설치
- ✅ **NotebookLM Essay Archive 연동** (2026-02-19): storage_state.json VM 직접 scp 배포. Issue 013 "충만의 조건"부터 에세이 자동 저장 확인
- ✅ **CE content_type 분기 설계** (2026-02-19): archive(한다체/사색적)/magazine(합니다체/독자지향) 어조+구조 분기 플랜 확정. 미구현 상태
- ✅ **CE content_type 분기 구현** (2026-02-20): gardener.py payload에 content_type 추가 + ce_agent.py _write_corpus_essay() 어조 분기 로직 구현. VM 배포 + ecosystem 재시작 완료.
- ✅ **WOOHWAHAE 대규모 업데이트 + nginx 도메인 배포 준비** (2026-02-20): nginx 80포트/server_name/root 수정, style.css v36 전체 통일(24개 파일), 레거시 CSS 제거, OG태그 보완, CDN 통일, 375px 미디어쿼리, 전체 VM 재배포. DNS BLOCKER 남음(아임웹 A레코드 136.109.201.201).
- ✅ **LAYER OS Rebuild Phase -1~1** (2026-02-24): Claude Code 인프라(Memory 4개, 커맨드 4개, Hooks, Rules), 레거시 10파일 삭제, 배포스크립트 이동, 에이전트 기능화(persona→role: JOON→SA, MIA→AD, RAY→CE, CD_SUNHO→CD), OS 리브랜딩(97layerOS→LAYER OS), FILESYSTEM_MANIFEST.md 서재 맵 구축.
- ✅ **LAYER OS Rebuild Phase 2A~4** (2026-02-24): 2차 파편제거(빈폴더 20개/worktree 7개/이벤트 479개 삭제, 파편 4건 통합), Brand OS 11개 문서(directives/brand/), IDENTITY v7(brand/ 참조 체계), SYSTEM v6(5-Layer 매핑), agent_router v2(AGENT_REGISTRY 수정+brand/ 로딩), CE/AD/SA brand/ 문서 로딩, 파이프라인 재설계(SA→CE→Ralph→AD→CD), 통합 신호 스키마(signal/ritual/growth 3종).
- ✅ **통합 신호 수집 코드** (2026-02-24): 7개 파일 수정. telegram_secretary(source_channel+이미지 통합스키마+PDF 핸들러), youtube_analyzer(signal_id+from_user+source_channel+SA큐연결), image_analyzer(signals/images/→signals/+signals/files/), signal_router(5개 통합타입+레거시 호환), scout_crawler(.md→.json 통합스키마), pipeline_orchestrator(새 타입 호환), scripts/signal_inject.py(CLI 수동 입력 도구).
- ✅ **Claude Code 인프라 강화** (2026-02-24): 보안 hooks(output-secret-filter+command-guard), 세션 라이프사이클(session-start+session-stop), 규칙(security.md+git-workflow.md), 품질 게이트(code-quality-check+/verify+validate-path 보강). deploy.sh PROJECT_ROOT 버그 수정+3서비스 재시작+누락 경로 추가.
- ✅ **VM 배포** (2026-02-24): deploy.sh → 3서비스 모두 active (telegram/ecosystem/gardener). Brand OS + 통합 신호 수집 + 스키마 전체 반영.
- ✅ **Ritual Module** (2026-02-24): core/modules/ritual.py — 고객 CRUD, 방문 기록, 리듬 자동 계산(빠른/보통/느린), 재방문 알림, CLI. 스키마 기반.
- ✅ **Growth Module** (2026-02-24): core/modules/growth.py — 수익 수동입력, 콘텐츠/서비스 자동 집계, 월간 리포트 마크다운 생성, 추세 분석. 스키마 기반.
- ✅ **레거시 신호 마이그레이션** (2026-02-24): .md 11개 → JSON 4개 변환 + 7개 archive. wellness/ 폴더 삭제. signals/ 100% JSON.
- ✅ **Ritual/Growth Telegram 연동** (2026-02-24): /client(list|add|info|due), /visit 신규 커맨드. /growth → Growth Module 월별 지표. @admin_only 데코레이터 12개 커맨드 전체.
- ✅ **Gardener Growth 자동 집계** (2026-02-24): _record_growth_snapshot() + run_cycle() step5 연동. 매일 새벽 3시 Growth 스냅샷 자동 갱신.
- ✅ **전면 보안 강화** (2026-02-24): B1 Secret강제/B2 Cookie/B3 CSRF/B4 Telegram인증/B5 CORS/B6 SSRF/B7 보안헤더/B8 RateLimit/B9 PathTraversal/B10 AuditLog/B11 ErrorHandler/B12 HTTPS준비. 자가검증 전수 통과. nginx security headers VM 라이브 확인.
- ✅ **Sprint 6 — 슈퍼앱 통합** (2026-02-24): nginx /api/+/me/+/consult/ 프록시 추가. Ritual Module portal_token/email/find_client_by_token()/add_visit() 확장. website/backend Flask → core/modules 연결. /me/<token> 고객 시술일지 포털. /consult/<token> 사전상담 폼 + 사진 업로드. style_matcher.py 무드 키워드→레퍼런스 이미지 알고리즘. portal.html/consult.html/consult_done.html 신규. Component Standard Layer v1.0 (style.css + admin.css 포팅). Growth Dashboard /admin/growth + /admin/growth/revenue. growth.html 신규. base.html Growth 네비. VM woohwahae-backend 서비스 신규 + ADMIN_SECRET_KEY .env 추가. 3개 라우트 smoke test 통과.

## 🎯 다음 작업

1. [BLOCKER] 아임웹 DNS A레코드 `136.109.201.201` 설정 (사용자 직접)
2. [BLOCKER] 실제 고객 데이터 입력 — Ritual Module에 첫 고객 등록 후 portal_token 생성 → `/me/{token}` URL 테스트
3. 사전상담 URL 첫 실사용 — `/consult/{token}` 카톡 전송 → 제출 → consult_done 확인
4. Growth Dashboard 첫 수익 입력 — `/admin/growth`에서 2026-02 수익 기록
5. DNS 연결 후: certbot + HTTPS/HSTS 활성화 (nginx 준비 블록 해제)
6. 재방문 알림 자동화 — Gardener가 `get_due_clients()` 실행 → 카카오 Alimtalk or 텔레그램

## 📐 콘텐츠 전략 (2026-02-19 확정)

- **단일 렌즈**: WOOHWAHAE = "슬로우라이프"라는 렌즈로 세상을 읽는다
- **카테고리 없음**: 헤어/오브제/에세이 모두 같은 질문("어떻게 살 것인가")으로 귀결
- **어조 분기**: archive(한다체, 사색적) / magazine(합니다체, 독자 지향) — 사람이 명시 지정
- **수익화**: 전자책 PDF → 구독화 (에세이 50개 이후)
- **피드백 루프**: 에세이 50개 이후 설계
- **현재 상태**: 에세이 13개, 신호 38개, 군집 20개 (ripe 1개)

---

## 🚀 실행 명령

```bash
# VM 상태
ssh 97layer-vm "systemctl is-active 97layer-telegram 97layer-ecosystem 97layer-gardener"

# 로그
ssh 97layer-vm "sudo journalctl -u 97layer-ecosystem -n 50 --no-pager"

# 배포 + 재시작
scp core/agents/gardener.py 97layer-vm:/home/skyto5339_gmail_com/97layerOS/core/agents/
ssh 97layer-vm "sudo systemctl restart 97layer-ecosystem"
```

---

## 🌱 Gardener 자동 업데이트

*미실행 — 다음 Gardener 사이클 시 자동 갱신*



---

## 📍 현재 상태 (CURRENT STATE)

### [2026-02-24 16:39] Auto-Update — auto-session

**이번 세션 커밋**:
- ✅ fix: CE published 상태 체크 수정 + telegram_sent 추적
- ✅ feat: 세션 연속성 고도화 — QUANTA 자동갱신 + 선택 로드 + 토큰 추적
- ✅ fix: ecosystem 좀비 프로세스 제거 — trap 핸들러 + 서비스 스크립트 전환
- ✅ fix: orchestrator 이중 로그 제거 — FileHandler 삭제 (StreamHandler 단일화)
- ✅ chore: 전수 조사 기반 파일시스템 구조 정리
- ✅ refactor: start_*.sh + sync*.sh 루트 → scripts/ 이동 (루트 체계화)
- ✅ fix: Gemini가 삭제한 tools.html + /tools 라우트 + 사이드바 링크 복원
- ✅ chore: Gemini 잔재 plan_dispatcher.py 삭제 (미사용, 문법 오류)
- ✅ fix: copyright year 2026 → 2024 복원
- ✅ style: 웹사이트 일관성 패치 — 모바일 nav slide-out, footer 통일, CSS 버전 bump
- ✅ fix: gardener 트리거 플래그 --once → --run-now 수정
- ✅ feat: Admin 지휘소 통합 — 사이드바 + SSE 실시간 + 4개 신규 패널
- ✅ feat: Admin Ritual 패널 신설 — 고객 관리 웹 UI
- ✅ feat: /client add 링크 자동 출력 + phone 필드 + /client link 커맨드
- ✅ feat: Sprint 6 — woohwahae.kr 슈퍼앱 통합 구조 구축
- ✅ chore: QUANTA v7.3 갱신 — Sprint 4+5 완료 상태 반영
- ✅ security: CSRF+SSRF+AuthZ+Cookie+Headers+AuditLog+RateLimit 전면 적용
- ✅ feat: Ritual/Growth Telegram 연동 + Gardener 자동 집계
- ✅ feat: 미추적 신규 파일 5개 추적 시작
- ✅ feat: Ritual Module (L4) + Growth Module (L5) + VM 배포 + 레거시 마이그레이션
- ✅ feat: Claude Code 인프라 강화 — 보안 hooks + 세션 라이프사이클 + 품질 게이트

**미커밋 변경**:
- ⚠️  .claude/hooks/session-start.sh
- ⚠️  .claude/hooks/validate-path.sh
- ⚠️  knowledge/agent_hub/INTELLIGENCE_QUANTA.md
- ⚠️  .claude/hooks/compact-reminder.sh
- ⚠️  knowledge/system/token_usage_log.jsonl

**업데이트 시간**: 2026-02-24T16:39:51.227061
