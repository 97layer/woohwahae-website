# 🧠 INTELLIGENCE QUANTA - 지능 앵커

> **목적**: 어떤 모델/세션이 오더라도 사고 흐름이 끊기지 않도록 보장하는 물리적 앵커
> **갱신 정책**: 덮어쓰기 (최신 상태만 유지). Gardener가 매일 자동 갱신.
> **마지막 갱신**: 2026-02-24 (통합 신호 수집 코드 완료: 5개 소스 통일 + CLI 입력 + SA 큐 자동 연결)

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

**버전**: Ver 7.1 — 통합 신호 수집 코드 완료 (5개 소스 × 통합 스키마 + CLI 입력 + SA 큐 자동 연결)

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

## 🎯 다음 작업

1. [BLOCKER] 아임웹 DNS A레코드 `136.109.201.201` 설정 (사용자 직접)
2. VM 배포: Brand OS + 통합 신호 수집 코드 전체 배포 + 서비스 재시작
3. Ritual/Growth 코드 구현 (스키마 기반)
4. 레거시 신호 마이그레이션 (signals/wellness/*.md → JSON 또는 archive)

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

### [2026-02-24] Session Update - claude-opus-rebuild (Phase 2)

**완료한 작업**:
- ✅ Phase 2A: 2차 물리 정리 — 빈폴더 20개, worktree 7개, 이벤트 479개 삭제. 파편 4건 통합.
- ✅ Phase 2B: Brand OS 11개 문서 생성 (directives/brand/). IDENTITY v7.0, SYSTEM v6.0 업그레이드.
- ✅ Phase 2B: agent_router.py CRITICAL FIX — AGENT_REGISTRY 파일명 수정, brand/ 로딩 추가.
- ✅ Phase 2B: 에이전트 directive 기능화 완료 — persona 이름/성격 서술 0건.
- ✅ Phase 3: CE _load_brand_directives(), AD _load_design_tokens(), SA audience.md 로딩.
- ✅ Phase 3: 파이프라인 재설계 — SA→CE→Ralph(인라인)→AD→CD.
- ✅ Phase 4: 통합 신호 스키마 + Ritual + Growth (JSON Schema 3종).

**다음 단계**:
- ⏳ VM 배포 (기능화된 코드 + Brand OS 문서)
- ⏳ 통합 신호 수집 코드 구현 (이미지/PDF/URL 핸들러)
- ⏳ CLI signal_inject.py
- ⏳ Ritual/Growth 코드 구현

**업데이트 시간**: 2026-02-24T08:00:00
