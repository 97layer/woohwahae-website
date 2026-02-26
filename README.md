# 97LAYER OS (Ver 11.0 — THE ORIGIN Enforcement)

> **상태**: THE ORIGIN 기반 강제 구조 완성 (Filesystem Validator + 3-Layer Defense)
> **최종 갱신**: 2026-02-26

---

## 📂 4축 구조 (The Quadrant)

```
97layerOS/
├── directives/              # 🧠 뇌 — 철학, 규칙, 규격
│   ├── THE_ORIGIN.md       # SSOT (Single Source of Truth)
│   ├── SYSTEM.md           # 운영 프로토콜
│   ├── MANIFEST.md         # Filesystem 배치 규칙
│   ├── practice/           # 실행 규격 (visual, language, content, audience, experience, offering)
│   └── agents/             # 에이전트 판단 기준 (SA, CE, AD, CD)
│
├── knowledge/               # 📚 기억 — 데이터, 신호, 상태
│   ├── agent_hub/          # INTELLIGENCE_QUANTA.md, council_room.md
│   ├── signals/            # 원시 신호 (텔레그램/유튜브/메모)
│   ├── corpus/             # 구조화 지식 (SA 분석 결과)
│   ├── clients/            # CRM 클라이언트 (Ritual Module)
│   ├── service/            # 서비스 아이템 카탈로그
│   ├── reports/            # 리포트 (morning, evening, audit)
│   └── system/             # 런타임 상태 (work_lock, cache, schemas)
│
├── core/                    # ⚙️  엔진 — 코드, 스크립트, 스킬, 테스트
│   ├── agents/             # SA, CE, AD, CD, Code, Gardener
│   ├── system/             # 파이프라인 + AI엔진 + bridges + modules (23개)
│   ├── daemons/            # 상주 서비스 (telegram, dashboard, nightguard)
│   ├── admin/              # 웹 대시보드 (Flask)
│   ├── scripts/            # 자동화 (deploy/, session, sync)
│   ├── skills/             # 에이전트 스킬 (deploy, signal_capture, data_curation, etc.)
│   └── tests/              # 테스트
│
└── website/                 # 🌐 얼굴 — HTML/CSS/JS, 네비: Archive | Practice | About
    ├── index.html          # 홈
    ├── about/              # About — 철학, 서사, 에디터
    ├── archive/            # Archive — essay-NNN-slug/, magazine/, lookbook/
    ├── practice/           # Practice — atelier, direction, project, product, contact
    ├── woosunho/           # 에디터 포트폴리오
    ├── lab/                # 실험 (네비 미노출)
    └── assets/             # CSS, JS, 이미지
```

---

## 🚀 실행 (Execution)

### VM (Production)
```bash
ssh 97layer-vm
systemctl status 97layer-telegram 97layer-ecosystem 97layer-gardener
sudo journalctl -u 97layer-ecosystem -n 50 --no-pager
```

### 로컬 (Development)
```bash
# Telegram Bot
python3 core/daemons/telegram_secretary.py

# Dashboard
python3 core/admin/server.py

# Filesystem Guard (Daemon)
python3 core/system/filesystem_guard.py
```

---

## 🎯 핵심 시스템

### 1. THE ORIGIN Guidance Route
철학적 순환망: THE_ORIGIN → practice/ → agents/ → 산출물 → THE_ORIGIN 회귀

### 2. 3-Layer Filesystem Defense
- **Layer 1**: Python API Wrapper (`filesystem_validator.py::safe_write()`)
- **Layer 2**: Pre-commit Hook (Git staged 파일 MANIFEST 검증)
- **Layer 3**: Daemon (`filesystem_guard.py` — 15초 스캔 → 격리)

### 3. THE CYCLE (구심점으로 돌아오는 반복)
```
점 (Signal) → 지층 (Archive) → 결속 (Synapse) → 세공 (Craft) → 발현 (Manifestation) → 회귀 (Return)
```

### 4. STAP 품질 게이트 (5 Pillars)
Ralph Agent 검증: 진정성, 실효적 파동, 단호한 여백, 정밀한 조율, 주권의 획득 (90점 임계치)

---

## 📡 Telegram Commands

```
/start       시스템 소개
/status      파이프라인 현황 (신호/Corpus/고객/Growth/VM)
/signal      새 신호 입력 (URL, 텍스트, 유튜브)
/morning     아침 리포트 생성 (morning_YYYYMMDD.md)
/doctor      시스템 상태 진단 (QUANTA + work_lock + cache)
/handoff     세션 종료 핸드오프 (다음 에이전트용)
```

---

## 🔐 슬래시 커맨드 (Claude Code)

```
/doctor       시스템 상태 진단
/deploy       GCP VM 배포 (전체 or 특정 파일/서비스)
/audit        폴더 구조 감사 (중복/orphan/금지 파일)
/status       LAYER OS 파이프라인 현황
/brand        Brand OS 핵심 규칙 참조
/quanta       INTELLIGENCE_QUANTA.md 갱신
/handoff      세션 종료 + 상태 기록
```

---

## 🛡️  강제 메커니즘

| Layer | 메커니즘 | 위치 |
|-------|---------|------|
| 1 | CLAUDE.md / .ai_rules | 루트 |
| 2 | Claude Code Hooks | `.claude/hooks/` |
| 3 | Claude Code Rules | `.claude/rules/` |
| 4 | Git Pre-Commit Hook | `.git/hooks/pre-commit` |
| 5 | Bootstrap Script | `core/scripts/session_bootstrap.sh` |

---

## 🎨 Slow Life 철학 (THE ORIGIN)

- **본질주의**: 조용히 덜어내어 투명한 뼈대만 렌더링
- **공명**: 넓게 흩뿌리는 확장이 아닌 수직으로 꽂히는 하강
- **자기긍정**: 생존 관리가 아닌 미학적 구원을 향한 의식

→ 상세: [THE_ORIGIN.md](directives/THE_ORIGIN.md)

---

## 📚 Documentation

- [THE_ORIGIN.md](directives/THE_ORIGIN.md) — 브랜드 철학 SSOT
- [SYSTEM.md](directives/SYSTEM.md) — 운영 프로토콜
- [MANIFEST.md](directives/MANIFEST.md) — Filesystem 배치 규칙
- [INTELLIGENCE_QUANTA.md](knowledge/agent_hub/INTELLIGENCE_QUANTA.md) — 세션 연속성

---

## 🔄 최근 주요 업데이트

- ✅ Ver 11.0 (2026-02-26): THE ORIGIN 기반 파일시스템 강제 구조 구축
- ✅ 3-Layer Defense: Python validator + Pre-commit hook + Daemon
- ✅ 레거시 위반 파일 59개 청산 (11개 archive 이동)
- ✅ 4축 구조 정렬 (directives/knowledge/core/website)
- ✅ Website HTML 리빌딩 (네비/푸터 통일, Phase 1 비주얼 완성)

---

> "소음이 걷힌 진공에 다다라서야 명징한 본질이 나선다." — THE ORIGIN
