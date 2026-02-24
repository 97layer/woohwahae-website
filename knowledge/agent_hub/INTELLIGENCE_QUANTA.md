# 🧠 INTELLIGENCE QUANTA - 지능 앵커

> **목적**: 어떤 모델/세션이 오더라도 사고 흐름이 끊기지 않도록 보장하는 물리적 앵커
> **갱신 정책**: 덮어쓰기 (최신 상태만 유지). session-stop 훅이 자동 갱신.
> **마지막 갱신**: 2026-02-24 (QUANTA 구조 최적화 — 178줄 → 65줄)
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

### [2026-02-24] QUANTA 구조 최적화

**완료**:
- ✅ QUANTA 178줄 → 65줄 (-63%) — 완료 이력 분리, placeholder 제거, 아키텍처 압축
- ✅ COMPLETED_WORK.md 신규 생성 (완료 이력 30개 보존)
- ✅ session-start.sh 전체 로드로 전환 (QUANTA 경량화로 불필요)

**다음**: 위 🎯 참조

**업데이트 시간**: 2026-02-24T16:45:00
