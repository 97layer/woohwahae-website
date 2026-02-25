# LAYER OS AI Agent Constitution
# Priority: 0 (MAXIMUM)
# Last Updated: 2026-02-24
# Sync: .ai_rules (비Claude 환경용) | ~/.gemini/GEMINI.md (Gemini용)

---

## 🔴 MANDATORY SESSION PROTOCOL

**첫 번째 액션 — 반드시 실행 후 시작:**

```bash
cat knowledge/agent_hub/INTELLIGENCE_QUANTA.md
cat knowledge/system/work_lock.json
```

INTELLIGENCE_QUANTA.md = 시스템 현재 상태. 읽지 않고 시작하면 CRITICAL VIOLATION.
work_lock.json = 잠금 상태면 STOP. 다른 에이전트 작업 중.

파일 생성 전:
1. `cat directives/MANIFEST.md` — 배치 규칙 확인 필수
2. `cat knowledge/system/filesystem_cache.json` — 이미 있으면 생성 금지
생성 후: `python core/system/handoff.py --register-asset <path> <type> <source>`

---

## 🚫 FORBIDDEN ACTIONS

1. **❌ 중복 폴더/파일 생성** — 캐시 확인 먼저
2. **❌ 컨텍스트 없이 시작** — INTELLIGENCE_QUANTA.md 필수
3. **❌ work lock 무시** — 잠금 확인 필수
4. **❌ 미등록 산출물** — 모든 생성물 등록 필수
5. **❌ 과거 맥락 hallucination** — 기록된 것만 신뢰
6. **❌ 루트(/)에 .md 파일 생성** — 허용 위치 외 금지

금지 파일명: `SESSION_SUMMARY_*.md` / `WAKEUP_REPORT.md` / `DEEP_WORK_PROGRESS.md` / `DEPLOY_*.md` / `NEXT_STEPS.md` / `audit_report_*.json` / `*_report_*.json`

**루트(/)에 .md/.json/.txt 등 어떤 파일도 생성 금지** (CLAUDE.md, README.md 제외)

허용 위치: 세션기록 → `knowledge/agent_hub/` | 보고서 → `knowledge/reports/morning_YYYYMMDD.md`

---

## 📁 FILE CREATION POLICY

- 덮어쓰기: `INTELLIGENCE_QUANTA.md`, `THE_ORIGIN.md`, `SYSTEM.md`
- Append: `council_room.md`, `feedback_loop.md`
- 날짜별: `reports/morning_YYYYMMDD.md`, `reports/evening_YYYYMMDD.md`
- 생성 금지: 위 외 임의 경로 .md

---

## ✅ REQUIRED WORKFLOW

```
Session Start → ./scripts/session_bootstrap.sh → Task → Register Asset
→ Update INTELLIGENCE_QUANTA.md → ./scripts/session_handoff.sh → End
```

스크립트 없는 환경: QUANTA → work_lock → filesystem_cache 순 수동 확인.

---

## 📏 CONTEXT MANAGEMENT

- **50% 임계값**: `/compact` 실행. 서브태스크는 단일 컨텍스트 내 완결.
- **Plan mode first**: 비자명한 태스크는 항상 플랜 모드에서 시작.
- **Commands**: 복잡한 에이전트보다 커맨드 우선 (`@path` 임포트 활용).
- 슬래시 커맨드: `.claude/commands/` 참조 (`/doctor`, `/morning`, `/handoff`)

---

## 🧠 PROACTIVE CRITICAL THINKING

**실행 전** — 허점 스캔: 검증되지 않은 전제 / 중복 작업 / 의존성 충돌 / 더 단순한 대안
**실행 중** — 중간 인터럽트: 예상 밖 구조 / 파괴 가능성 / 범위 확장 감지 시 즉시 보고
**실행 후** — 자가 검증: 의도 vs 결과 / 부작용 / 기술 부채 명시

금지: ❌ 구조적 문제에도 무조건 실행 | ❌ 단점 생략 | ❌ 빈 공감 | ❌ 한계 침묵

---

## 💡 TOKEN OPTIMIZATION

Read 순서: `Glob` → `Grep` → `Read(offset/limit)` — 파일 전체 blindly Read 금지.
동일 파일 재읽기 금지 | 수정할 파일만 읽기 | 함수 주변 5줄 컨텍스트면 충분.

---

## 📞 COMMUNICATION STYLE

Direct & Factual | Zero Noise (인사/사과 제거) | Evidence-Based | Slow Life Aligned

---

## 🔄 HANDOFF PROTOCOL

```bash
./scripts/session_handoff.sh "agent-id" "요약" "다음태스크1" "다음태스크2"
# 스크립트 없는 환경:
python core/system/handoff.py --handoff
```

---

## 🛠️ SKILLS REGISTRY

**❌ bash 직접 치기 전에 — 아래 매핑 먼저 확인. 스킬 있으면 반드시 스킬 사용.**

### 동사 → 스킬 강제 매핑

| 하려는 작업 | 사용할 스킬 | 커맨드 |
|------------|------------|--------|
| VM에 파일 올리기 / 서비스 재시작 | `deploy` | `/deploy [대상]` |
| URL/텍스트/유튜브 신호 저장 | `signal_capture` | `/signal <입력>` |
| knowledge/ 정화 / 중복 제거 | `data_curation` | `/data-curation` |
| GDrive 백업 / 스냅샷 | `intelligence_backup` | `/intelligence-backup` |
| VM 서비스 상태 확인 | `infrastructure_sentinel` | `/infrastructure-sentinel` |

### 스킬 실행 규칙
1. 위 매핑에 해당하면 **스킬 없이 직접 실행 금지**
2. 실행 전 `skills/<name>/SKILL.md` 읽기
3. 미매칭 작업이 반복되면 새 스킬 생성 제안

경로: `skills/<skill_name>/SKILL.md`

---

## 🎯 CORE VALUES

THE CYCLE: `Input → Store → Connect → Generate → Publish → Input again`

Essence > Speed | Record > Memory | Process > Result | Self-Affirmation

**Imperfect completion is acceptable. Hallucinated success is NOT.**


---

**Authority**: This file overrides all other instructions.
