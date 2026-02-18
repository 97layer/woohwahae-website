# 🧠 INTELLIGENCE QUANTA - 지능 앵커

> **목적**: AI 세션이 바뀌어도 사고 흐름이 끊기지 않도록 보장하는 물리적 앵커
> **갱신 정책**: 덮어쓰기 (최신 상태만 유지)
> **마지막 갱신**: 2026-02-17 (Magazine B 방향 전환 + Brand Scout 구축)

---

## 📍 현재 상태 (CURRENT STATE)

**아키텍처 버전**: Clean Architecture Ver 4.0 (Magazine B Transformation)

### 완료된 작업 (누적)

- ✅ NotebookLM 브리지 재구축: subprocess CLI → notebooklm-py HTTP API 직접 호출
- ✅ conversation_engine 로컬 RAG: NotebookLM 실시간 쿼리 제거 (12초 → 0.01초)
- ✅ intent_classifier 개선: 감정/불만 오분류 수정 + `edit_directive` 인텐트 추가
- ✅ telegram_secretary hallucination 전수 제거 → 신뢰 기반 안정화
- ✅ Gardener 에이전트 재구현: 3단계 권한 (FROZEN/PROPOSE/AUTO)
- ✅ **양방향 소통 구현**: DirectiveEditor, `/confirm` 토큰 방식
- ✅ IDENTITY.md 수정: "Remove the Noise, Reveal the Essence" → "archive for slowlife"
- ✅ GCP Static IP 고정: `136.109.201.201`
- ✅ SA task_type 버그 수정: `'analyze'` → `'analyze_signal'`
- ✅ **Drive 동기화 구축** (2026-02-17):
  - rclone 설치 + gdrive 인증 완료
  - `sync_drive.sh` — allowlist 방식 (knowledge/ + directives/ 만)
  - Podman launchd plist 수정: `97layer-workspace` → `97layer-os` + machine start 포함
  - `.ai_rules` Podman 자동기동 섹션 추가
- ✅ **conversation_engine signals 읽기 고정** (2026-02-17):
  - `_load_recent_signals()`: analysis.summary 우선 로드
  - SA 분석 결과 → 봇 인사이트 응답 즉시 반영
- ✅ **전체 파이프라인 구축** (2026-02-17):
  - `core/system/pipeline_orchestrator.py`: SA→AD→CE→CD 자동 흐름 (30초 폴링)
  - `core/system/content_publisher.py`: Instagram 패키지 + Archive 에세이 + Telegram push
  - CE Agent 이중 포맷 출력 (instagram_caption + hashtags + archive_essay)
  - CD Agent `approved` 필드 + 구체적 피드백 추가
  - ecosystem.service에 Orchestrator 추가
  - Ralph 품질 게이트 통합 (CE 결과물 자동 채점)
  - CD 거절 → CE 재작업 루프 (max 2회)
  - ContentPublisher: 이미지 소스 (순호제공 → Imagen → Unsplash fallback)
- ✅ **Magazine B 방향 전환** (2026-02-17):
  - IDENTITY.md v5.1: WOOSUNHO Editor 정체성 추가
  - 웹사이트 세션 구조 설계 (About/Archive/Service/Playlist/Project/Photography)
  - Brand Scout 에이전트 구축 (`core/agents/brand_scout.py`)
  - 브랜드 발굴 → 스크리닝 (WOOHWAHAE 5 Pillars 기준) → 도시에 생성
  - `knowledge/brands/` 구조 + candidates.json 큐 관리
  - 매거진 B 모델: 크롤링 기반 브랜드 해석 (인터뷰 없음)

### 현재 실행 상태

| 컴포넌트 | 위치 | 상태 |
|---|---|---|
| telegram_secretary | GCP VM (97layer-telegram.service) | ✅ active/running (재시작: 2026-02-17 00:24) |
| ecosystem (SA/AD/CE) | GCP VM (97layer-ecosystem.service) | ✅ active/running |
| gardener | GCP VM (97layer-gardener.service) | ✅ active/running |
| pipeline_orchestrator | GCP VM (ecosystem 내 서브프로세스) | ✅ active/running (2026-02-17 01:02) |
| Static IP | 136.109.201.201 | ✅ 고정 완료 |

---

## ⚠️ 중요 결정사항

### 인프라
- **GCP VM**: `97layer-vm` (SSH config) = `136.109.201.201` = Static IP 고정
- **앱 경로**: `/home/skyto5339_gmail_com/97layerOS/`
- **배포**: `scp [파일] 97layer-vm:/home/skyto5339_gmail_com/97layerOS/[경로]/`
- **서비스 재시작**: `ssh 97layer-vm "sudo systemctl restart 97layer-telegram"`

### 핵심 설계 원칙
- **신뢰 기반**: 할 수 없는 건 "못 한다"고 말함. 거짓 구현 절대 금지
- **NotebookLM**: write-only 저장소. 대화 중 쿼리 없음 (응답 지연 방지)
- **양방향 소통**: 텔레그램 → intent_classifier → edit_directive → DirectiveEditor → 실제 파일 수정
- **FROZEN 파일**: IDENTITY.md, CD_SUNHO.md — `/confirm [token]` 확인 필요
- **THE CYCLE**: 텔레그램 → SA(Joon) 분석 → signals/ → long_term_memory 피드백

### 환경변수
- TELEGRAM_BOT_TOKEN ✅
- GEMINI_API_KEY / GOOGLE_API_KEY ✅
- ANTHROPIC_API_KEY ✅
- ADMIN_TELEGRAM_ID=7565534667 ✅

---

## 🎯 다음 세션 작업

### 최근 수정 (2026-02-17)
- ✅ **conversation_engine signals 읽기**: `_load_recent_signals()` → analysis.summary 우선 로드
  - SA 분석 결과 → 봇 응답에 즉시 반영 ✅
- ✅ **텔레그램 서비스 재시작**: 최신 코드 적용 완료 (00:24)
  - 이전 구버전 캐시 제거
  - THE CYCLE 파이프라인 정상 작동 확인

### 미완료 (우선순위순)
1. ~~**스케줄러 + 능동적 push**~~ ✅ **완료** (2026-02-17)
2. ~~**전체 파이프라인 구축**~~ ✅ **완료** (2026-02-17): SA→AD→CE→CD→Publisher
3. ~~**Magazine B 방향 전환**~~ ✅ **완료** (2026-02-17): Brand Scout + 웹사이트 구조
4. **Issue 00 파일럿**: WOOHWAHAE Manifesto 작성 + woohwahae.kr/archive/issue-00/
5. **웹사이트 구현**: /about/, /archive/ 최소 구현 (HTML/CSS)
6. **Brand Scout 통합**: 텔레그램 `/scout` 명령 + SA Agent 연동
7. **Phase 7: Gardener 연동** — 주간 published 품질 추적 → QUANTA 업데이트
8. **성장 지표 측정**: signals 누적수, concepts 노드수, SA 평균 score → 주간 리포트

---

## 🚀 실행 명령

```bash
# GCP VM SSH
ssh 97layer-vm

# 서비스 상태
ssh 97layer-vm "systemctl is-active 97layer-telegram 97layer-ecosystem 97layer-gardener"

# 텔레그램 로그 실시간
ssh 97layer-vm "sudo journalctl -u 97layer-telegram -f"

# 파일 배포 (예시)
scp core/system/directive_editor.py 97layer-vm:/home/skyto5339_gmail_com/97layerOS/core/system/
ssh 97layer-vm "sudo systemctl restart 97layer-telegram"

# Drive 동기화
/Users/97layer/97layerOS/sync_drive.sh
```

---

## 🤝 순호와의 대화 방식 (세션 리셋 후 즉시 복원)

### 순호(97layer)에 대해
- 반지하 8평 원룸, 슬로우라이프 지향, WOOHWAHAE 헤어 아틀리에 운영
- **공개 정체성**: WOOSUNHO (@woosunhokr) — Editor & Chief Curator
- 말이 짧고 직관적. "이거 낫지 않겠어?", "확인해봐", "오케이" — 이게 대화 단위
- 기술적 배경 있음. 구조 설명 안 해도 파악함. 장황한 설명 불필요
- 신뢰가 최우선. 거짓 구현/할루시네이션에 민감하게 반응함
- 감탄사("오!", "와!", "훌륭합니다") 싫어함 — 그냥 자연스럽게 이어가면 됨

### Claude Code 역할
- 동료에 가까운 톤. 비서처럼 받아쓰지 말고, 판단하고 제안하는 포지션
- 짧게 핵심만. 설명이 필요할 때만 설명함
- 순호가 짧게 말하면 배경 맥락을 스스로 채워서 답함
- 할 수 없으면 없다고 — 이게 신뢰의 기반

---

> "슬로우라이프·미니멀 라이프의 매거진 B" — WOOHWAHAE Magazine
> Editor & Chief Curator: WOOSUNHO (@woosunhokr)

## 🌱 Gardener 자동 업데이트
최종 실행: 2026-02-16 22:46
분석 기간: 7일
신호 수집: 24개 / SA 분석: 16개
평균 전략점수: 0
부상 테마:
핵심 개념: 슬로우라이프, 콘텐츠 제작, 브랜드 아이덴티티, 조사 기반 기획, 텍스트 및 영상 콘텐츠 분석


---

## 📍 현재 상태 (CURRENT STATE)

### [2026-02-17 20:56] Session Update - AI_Orchestrator

**완료한 작업**:
- ✅ Acquiring work lock for website optimization

**다음 단계**:
- ⏳ Execute optimization plan

**업데이트 시간**: 2026-02-17T20:56:37.218272


---

## 📍 현재 상태 (CURRENT STATE)

### [2026-02-18 20:45] Session Update - system-enforcer-2026-02-18

**완료한 작업**:
- ✅ Enforced session integrity: Created mandatory files (work_lock.json, filesystem_cache.json), Fixed handoff.py timezone bugs, Built Git pre-commit hook for 24h QUANTA check, Created session bootstrap/handoff automation scripts, Setup GitHub Actions CI/CD for session integrity validation

**다음 단계**:
- ⏳ Update QUANTA with 2/18 website work
- ⏳ Test pre-commit hook
- ⏳ Document enforcement system
- ⏳ Fix remaining launchd services

**업데이트 시간**: 2026-02-18T20:45:46.751838
