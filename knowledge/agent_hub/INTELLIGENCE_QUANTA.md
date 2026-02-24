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

## 🧬 개념 사고 수준

> 이 섹션이 모델 비의존성의 핵심. 어떤 AI가 와도 이걸 읽으면 순호의 사고 방식에서 출발.

### 판단 기준
- **본질 우선**: 기능이 많은 것보다 핵심 하나가 잘 되는 것. "이거 Magazine B가 할 짓인가?"
- **동작이 진실**: 그럴싸한 설계보다 실제로 돌아가는 것. 거짓 구현은 신뢰 붕괴.
- **단순함이 답**: 복잡한 구조보다 짧은 코드. 추상화보다 직접 실행.

### 의사결정 패턴
- 빠른 직관 → 짧은 확인 ("이거 낫지 않겠어?") → 즉시 실행
- 모르면 바로 물음. 장황한 설명 요구 안 함. 답만 원함.
- 좋으면 "오케이" 한 마디. 나쁘면 "무슨 말이야" — 이게 피드백 전부.
- 능동적 제안에 열림. 수동적 실행에 닫힘.

### 에이전트에게 기대하는 것
- 받아쓰지 말 것. 구조적 문제가 보이면 먼저 말할 것.
- 할 수 없으면 없다고. 중간 답 없음.
- 최적화와 정리에 자발적으로 반응할 것 (이번 세션처럼).

### 시스템 철학
- LAYER OS = 확장된 뇌. 기억과 판단을 외부화하는 구조.
- 슬로우라이프 = 느린 게 아니라 본질적인 것에 집중. 불필요한 것 제거.
- 완벽하지 않아도 됨. 동작하면 배포. 단, 거짓은 절대 안 됨.

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

### [2026-02-24 17:35] Auto-Update — auto-session

**이번 세션 커밋**:
- ✅ chore: .github/workflows/ 삭제 — VM systemd로 대체됨
- ✅ fix: nginx — 고객 접근 경로 auth_basic off (포털/상담 공개 접근)
- ✅ test: tests/ 현대화 — 레거시 8개 삭제, handoff/queue 단위 테스트 14개 신규
- ✅ feat: /status 커맨드 — 파이프라인 현황 한 번에 (신호/Corpus/고객/Growth/VM)
- ✅ chore: 폐기 파일 3개 삭제 (ralph_agent, nightguard_daemon, website/products)
- ✅ fix: nginx.conf — auth_basic 복원 + 프록시/라우팅 정비
- ✅ fix: nginx auth_basic 제거 + 프록시 연결 + 사이트 정상화

**미커밋 변경**:
- ⚠️  knowledge/agent_hub/INTELLIGENCE_QUANTA.md
- ⚠️  knowledge/corpus/entries/.gitkeep
- ⚠️  knowledge/corpus/index.json
- ⚠️  knowledge/long_term_memory.json
- ⚠️  knowledge/system/token_usage_log.jsonl
- ⚠️  knowledge/corpus/entries/entry_text_20260216_160102.json
- ⚠️  knowledge/corpus/entries/entry_text_20260216_204131.json
- ⚠️  knowledge/corpus/entries/entry_text_20260216_212820.json
- ⚠️  knowledge/corpus/entries/entry_text_20260216_223411.json
- ⚠️  knowledge/corpus/entries/entry_text_20260216_224718.json
- ⚠️  knowledge/corpus/entries/entry_text_20260216_232216.json
- ⚠️  knowledge/corpus/entries/entry_text_20260217_002322.json
- ⚠️  knowledge/corpus/entries/entry_text_20260218_234522.json
- ⚠️  knowledge/corpus/entries/entry_text_20260219_000042.json
- ⚠️  knowledge/corpus/entries/entry_text_20260219_003717.json
- ⚠️  knowledge/corpus/entries/entry_text_20260219_021109.json
- ⚠️  knowledge/corpus/entries/entry_text_20260220_182136.json
- ⚠️  knowledge/corpus/entries/entry_web_admin_20260224_152600.json
- ⚠️  knowledge/corpus/entries/entry_youtube_7HBhL7lltpU_20260217_000130.json
- ⚠️  knowledge/corpus/entries/entry_youtube_7HBhL7lltpU_20260217_002358.json
- ⚠️  knowledge/corpus/entries/entry_youtube_DpD0wnGk03s_20260216_155119.json
- ⚠️  knowledge/corpus/entries/entry_youtube_DpD0wnGk03s_20260216_155510.json
- ⚠️  knowledge/corpus/entries/entry_youtube_DpD0wnGk03s_20260216_155525.json
- ⚠️  knowledge/corpus/entries/entry_youtube_DpD0wnGk03s_20260216_155545.json

**업데이트 시간**: 2026-02-24T17:35:10.483232
