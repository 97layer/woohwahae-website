# 🧠 INTELLIGENCE QUANTA - 지능 앵커

> **목적**: 어떤 모델/세션이 오더라도 사고 흐름이 끊기지 않도록 보장하는 물리적 앵커
> **갱신 정책**: 덮어쓰기 (최신 상태만 유지). session-stop 훅이 자동 갱신.
> **마지막 갱신**: 2026-02-25 (auto-update by auto-session)
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
- 구조적 문제 보이면 먼저 말할 것. 할 수 없으면 없다고.

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

- **Ver**: 11.0 — 4축 통합: directives(뇌) / knowledge(기억) / core(엔진: agents+system+daemons+admin+scripts+skills+tests) / website(얼굴). offering→service 통일. bridges/modules→system 통합. orphan 12개 아카이브.
- **GCP VM**: `97layer-vm` = `136.109.201.201` | 앱 경로: `/home/skyto5339_gmail_com/97layerOS/`
- **서비스**: 97layer-telegram / 97layer-ecosystem / 97layer-gardener / woohwahae-backend (5000) / cortex-admin (5001)
- **파이프라인**: 신호 유입 → signal.schema.json → SA 분석 → Gardener 군집화 → CE 에세이 → 발행

---

## 🎯 다음 작업

1. [CRITICAL] 컴포넌트 통일 — nav/footer/wave-bg를 _templates/ 기준으로 정의 후 전 페이지 일괄 적용. 현재 페이지마다 구조 불일치.
2. [DESIGN] 비주얼 Phase 2 — Practice 서비스 상세, 에세이 읽기 경험, 모바일 최적화
2. [DESIGN] About 카피 확정 — 매니페스토/본문/Philosophy/Journey/Editor 텍스트 순호 검토
3. content_publisher.py — essay-NNN 타입 접두사 패턴 적용
4. Ralph 피드백 루프 구현 — STAP 자동 검증 + Gardener practice/ 수정 제안 + CD 승인 사이클
5. 첫 고객 Ritual Module 등록 → `/me/{token}` URL 실사용 검증
6. Growth Dashboard 첫 수익 입력 (`/admin/growth`, 2026-02 데이터)

**완료됨**:

- ✅ DNS A레코드 연결 (Cloudflare 경유, 104.21.51.203)
- ✅ HTTPS/SSL (certbot, Let's Encrypt)
- ✅ VM git 초기화
- ✅ 4축 구조 정렬 Ver 11.0 (d6a448b0)
- ✅ website HTML 리빌딩 — 네비/푸터 전체 통일 (Archive|Practice|About), 깨진 참조 0건
- ✅ 비주얼 Phase 1 — 캔버스 복구, 패밀리룩 디자인 프레임, 배경색 통일, SVG 히어로

---

## 📐 콘텐츠 전략

- **단일 렌즈**: WOOHWAHAE = "슬로우라이프"라는 렌즈로 세상을 읽는다
- **어조 분기**: archive(한다체, 사색적) / magazine(합니다체, 독자 지향) — 사람이 명시 지정
- **현재 상태**: 에세이 13개, 신호 38개, 군집 20개 (ripe 1개)
- **수익화**: 전자책 PDF → 구독화 (에세이 50개 이후)
- **디자인 검수 지침**: 행동 유도 버튼(CTA, 링크 등)이나 주요 설명 텍스트에 `--text-faint` 등 극단적 저대비 색상 사용 금지. (최소 대비 `--text-sub` 사용 유지)

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

### [2026-02-26 03:30] Manual-Update — claude-opus

**커밋 14건** (2581f4d0 → b31f4bd5):
- HTML 리빌딩 + 네비/푸터 전체 통일 (Archive|Practice|About)
- 비주얼 Phase 1: 캔버스 복구, SVG 히어로, 패밀리룩 디자인 프레임
- 배경색 톤다운 (#E4E2DC → #E3E2E0), Story→Journey 통일
- 홈 히어로 리빌딩 (로고 심볼 + 하단 세션 + 푸터 Contact)
- 레거시 경로 전수조사 + 일괄 수정 (.ai_rules, CLAUDE.md, ce_agent.py 등)
- ⚠️  website/service/

**업데이트 시간**: 2026-02-25T20:35:28.059601
