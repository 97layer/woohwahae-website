# 🧠 INTELLIGENCE QUANTA - 지능 앵커

> **목적**: 어떤 모델/세션이 오더라도 사고 흐름이 끊기지 않도록 보장하는 물리적 앵커
> **갱신 정책**: 덮어쓰기 (최신 상태만 유지). session-stop 훅이 자동 갱신.
> **마지막 갱신**: 2026-02-24 (auto-update by auto-session)
> **완료 이력**: `knowledge/agent_hub/COMPLETED_WORK.md` 참조

---

## 👤 이 사람에 대해

- **이름**: WOOSUNHO (순호). 반지하 8평 원룸. WOOHWAHAE 헤어 아틀리에 운영.
- **공개 정체성**: Editor & Chief Curator (@woosunhokr)
- **지향**: 슬로우라이프. 속도보다 본질. Magazine B 방식의 브랜드.
- **말투**: 짧고 직관적. "이거 낫지 않겠어?", "오케이" — 이게 대화 단위
- **기술 이해도**: 높음. 구조 설명 불필요. 거짓 구현에 민감.
- **싫어하는 것**: 감탄사("오!", "훌륭합니다"), 장황한 설명, 빈 공감
- **에이전트 포지션**: 비서가 아닌 동료. 받아쓰지 말고 판단하고 제안.

---

## 🛠️ 스킬 트리거 (작업 전 확인)

| 작업 | 커맨드 |
|------|--------|
| VM 배포 / 서비스 재시작 | `/deploy [대상]` |
| 신호 저장 (URL/텍스트/유튜브) | `/signal <입력>` |
| knowledge/ 정화 | `/data-curation` |
| GDrive 백업 | `/intelligence-backup` |
| VM 서비스 상태 | `/infrastructure-sentinel` |

> bash 직접 치기 전에 위 스킬 확인. 해당되면 스킬 먼저.

---

## 🏗️ 인프라 핵심

- **Ver**: 7.4 — woohwahae.kr 슈퍼앱 통합 완료 (고객 포털 + 사전상담 + Growth Dashboard)
- **GCP VM**: `97layer-vm` = `136.109.201.201` | 앱 경로: `/home/skyto5339_gmail_com/97layerOS/`
- **서비스**: 97layer-telegram / 97layer-ecosystem / 97layer-gardener / woohwahae-backend (5000) / cortex-admin (5001)
- **파이프라인**: 신호 유입 → signal.schema.json → SA 분석 → Gardener 군집화 → CE 에세이 → 발행

---

## 🎯 다음 작업

1. [BLOCKER] 아임웹 DNS A레코드 `136.109.201.201` 설정 (사용자 직접)
2. [BLOCKER] 첫 고객 Ritual Module 등록 → `/me/{token}` URL 실사용 검증
3. `/consult/{token}` 카톡 전송 → 실제 폼 제출 → consult_done 확인
4. Growth Dashboard 첫 수익 입력 (`/admin/growth`, 2026-02 데이터)
5. DNS 연결 후: certbot + HTTPS/HSTS 활성화
6. 재방문 알림 자동화 — Gardener `get_due_clients()` → 카카오 Alimtalk or 텔레그램

---

## 📐 콘텐츠 전략

- **단일 렌즈**: WOOHWAHAE = "슬로우라이프"라는 렌즈로 세상을 읽는다
- **어조 분기**: archive(한다체, 사색적) / magazine(합니다체, 독자 지향) — 사람이 명시 지정
- **현재 상태**: 에세이 13개, 신호 38개, 군집 20개 (ripe 1개)
- **수익화**: 전자책 PDF → 구독화 (에세이 50개 이후)

---

## 🚀 실행 명령

```bash
ssh 97layer-vm "systemctl is-active 97layer-telegram 97layer-ecosystem 97layer-gardener"
ssh 97layer-vm "sudo journalctl -u 97layer-ecosystem -n 50 --no-pager"
scp <file> 97layer-vm:/home/skyto5339_gmail_com/97layerOS/<path>/
ssh 97layer-vm "sudo systemctl restart 97layer-ecosystem"
```

---

## 📍 현재 상태 (CURRENT STATE)

### [2026-02-24 16:58] Auto-Update — auto-session

**이번 세션 커밋**:
- ✅ feat: /deploy 스킬 + 커맨드 추가 — 전체/서비스/파일 타겟 배포
- ✅ fix: AgentWatcher 시작 시 stale processing 태스크 자동 회수 (30분 임계값)
- ✅ refactor: QUANTA 구조 최적화 — 178줄 → 65줄, 완료 이력 분리
- ✅ refactor: 훅 최적화 — 중복 JSON 파싱 통합, QUANTA 체크 제거, compact-reminder 삭제
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
- ⚠️  knowledge/agent_hub/INTELLIGENCE_QUANTA.md
- ⚠️  knowledge/system/token_usage_log.jsonl
- ⚠️  website/assets/css/style.css
- ⚠️  website/index.html

**업데이트 시간**: 2026-02-24T16:58:10.842996
