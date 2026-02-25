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

1. ✅ VM 재배포 완료 — core/ 구조 + nginx redirect + systemd 경로 수정
2. [CRITICAL] website HTML 리빌딩 — Archive|Practice|About 네비 + 깨진 링크 전면 수정 + media/ 경로 반영
3. content_publisher.py — essay-NNN 타입 접두사 패턴 적용
4. [NEW] Ralph 피드백 루프 구현 — STAP 자동 검증 + Gardener practice/ 수정 제안 + CD 승인 사이클
5. 첫 고객 Ritual Module 등록 → `/me/{token}` URL 실사용 검증
6. Growth Dashboard 첫 수익 입력 (`/admin/growth`, 2026-02 데이터)

**완료됨**:

- ✅ DNS A레코드 연결 (Cloudflare 경유, 104.21.51.203)
- ✅ HTTPS/SSL (certbot, Let's Encrypt)
- ✅ VM git 초기화
- ✅ 4축 구조 정렬 Ver 11.0 (d6a448b0)

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

### [2026-02-25 20:35] Auto-Update — auto-session

**미커밋 변경**:
- ⚠️  .ai_rules
- ⚠️  .claude/commands/audit.md
- ⚠️  .claude/commands/brand.md
- ⚠️  .claude/commands/manifest.md
- ⚠️  .claude/rules/brand-content.md
- ⚠️  .driveignore
- ⚠️  CLAUDE.md
- ⚠️  README.md
- ⚠️  core/agents/agent_router.py
- ⚠️  core/agents/brand_scout.py
- ⚠️  core/agents/cd_agent.py
- ⚠️  core/agents/code_agent.py
- ⚠️  core/agents/gardener.py
- ⚠️  core/system/conversation_engine.py
- ⚠️  core/system/directive_editor.py
- ⚠️  core/system/filesystem_guard.py
- ⚠️  directives/IDENTITY.md
- ⚠️  directives/README.md
- ⚠️  directives/agents/CD.md
- ⚠️  directives/brand/BRAND_MANUAL.md
- ⚠️  directives/brand/README.md
- ⚠️  directives/brand/audience.md
- ⚠️  directives/brand/content_system.md
- ⚠️  directives/brand/design_tokens.md
- ⚠️  directives/brand/experience_map.md
- ⚠️  directives/brand/foundation.md
- ⚠️  directives/brand/philosophy.md
- ⚠️  directives/brand/roadmap.md
- ⚠️  directives/brand/service_ritual.md
- ⚠️  directives/brand/story.md
- ⚠️  directives/brand/teaching.md
- ⚠️  directives/brand/voice_tone.md
- ⚠️  directives/system/FILESYSTEM_MANIFEST.md
- ⚠️  directives/system/SYSTEM.md
- ⚠️  knowledge/agent_hub/INTELLIGENCE_QUANTA.md
- ⚠️  knowledge/docs/archive/CONTAINER_STRATEGY_ANALYSIS.md
- ⚠️  knowledge/docs/archive/FINAL_VALIDATION.md
- ⚠️  knowledge/docs/archive/LAUNCH_CHECKLIST.md
- ⚠️  knowledge/docs/archive/NOTEBOOKLM_MCP_INTEGRATION_PLAN.md
- ⚠️  knowledge/docs/deployment/phase_30_master_plan.md
- ⚠️  knowledge/docs/system/WEBSITE_STRUCTURE.md
- ⚠️  knowledge/offering/items.json
- ⚠️  knowledge/system/asset_registry.json
- ⚠️  knowledge/system/filesystem_cache.json
- ⚠️  nginx.conf
- ⚠️  scripts/audit_html.py
- ⚠️  scripts/audit_html_simple.py
- ⚠️  scripts/audit_typography.py
- ⚠️  scripts/generate_hair_assets.py
- ⚠️  scripts/unify_nav.py
- ⚠️  tests/test_handoff.py
- ⚠️  website/404.html
- ⚠️  website/_templates/article.html
- ⚠️  website/about.html
- ⚠️  website/archive/index.html
- ⚠️  website/archive/issue-00/index.html
- ⚠️  website/archive/issue-001-beginning/index.html
- ⚠️  website/archive/issue-002-slow-life/index.html
- ⚠️  website/archive/issue-003-hair-and-daily/index.html
- ⚠️  website/archive/issue-004-art-of-waiting/index.html
- ⚠️  website/archive/issue-005-72h-rule/index.html
- ⚠️  website/archive/issue-006-8pyeong/index.html
- ⚠️  website/archive/issue-007-noise-removal/index.html
- ⚠️  website/archive/issue-008-raw-materiality/index.html
- ⚠️  website/archive/issue-009-inner-world/index.html
- ⚠️  website/archive/issue-010-work-and-essence/index.html
- ⚠️  website/archive/issue-010-work-and-essence/proto.html
- ⚠️  website/archive/issue-010-work-and-essence/proto_equilibrium.html
- ⚠️  website/archive/issue-010-work-and-essence/proto_v3.html
- ⚠️  website/assets/css/style.css
- ⚠️  website/atelier.html
- ⚠️  website/backend/templates/consult.html
- ⚠️  website/backend/templates/consult_done.html
- ⚠️  website/backend/templates/portal.html
- ⚠️  website/brand-audit-form.html
- ⚠️  website/contact.html
- ⚠️  website/cut-anatomy.html
- ⚠️  website/detail/brand-consulting.html
- ⚠️  website/detail/cut-program.html
- ⚠️  website/detail/editor.html
- ⚠️  website/detail/hair-atelier.html
- ⚠️  website/detail/hair-project.html
- ⚠️  website/detail/objects.html
- ⚠️  website/index.html
- ⚠️  website/lab/agent-office.html
- ⚠️  website/lab/demo-agent-visualizer.html
- ⚠️  website/lab/demo-home-blueprint.html
- ⚠️  website/lab/demo-salon-blueprint.html
- ⚠️  website/lab/design-system.html
- ⚠️  website/lab/hyper-archive.html
- ⚠️  website/lab/index.html
- ⚠️  website/lab/production-index.html
- ⚠️  website/lab/prototype-avant-garde.html
- ⚠️  website/lab/prototype-external-full.html
- ⚠️  website/lab/prototype-master-cut.html
- ⚠️  website/lab/prototype-nexus.html
- ⚠️  website/lab/prototype-offering.html
- ⚠️  website/lab/prototype-omni.html
- ⚠️  website/lab/prototype-ultimate.html
- ⚠️  website/lab/prototype-unified.html
- ⚠️  website/lab/renewal-core.html
- ⚠️  website/lab/renewal-premium.html
- ⚠️  website/lab/system-build-core.html
- ⚠️  website/lab/system-offering-detail.html
- ⚠️  website/offering.html
- ⚠️  website/offering/atelier.html
- ⚠️  website/offering/consulting.html
- ⚠️  website/offering/project-form.html
- ⚠️  website/offering/project.html
- ⚠️  website/offering/shop.html
- ⚠️  website/payment-fail.html
- ⚠️  website/payment-success.html
- ⚠️  website/photography.html
- ⚠️  website/privacy.html
- ⚠️  website/service.html
- ⚠️  website/shop.html
- ⚠️  website/terms.html
- ⚠️  directives/MANIFEST.md
- ⚠️  directives/SYSTEM.md
- ⚠️  directives/THE_ORIGIN.md
- ⚠️  directives/practice/
- ⚠️  knowledge/docs/archive/2026/02_february/CONTAINER_STRATEGY_ANALYSIS.md
- ⚠️  knowledge/docs/archive/2026/02_february/FINAL_VALIDATION.md
- ⚠️  knowledge/docs/archive/2026/02_february/LAUNCH_CHECKLIST.md
- ⚠️  knowledge/docs/archive/2026/02_february/NOTEBOOKLM_MCP_INTEGRATION_PLAN.md
- ⚠️  knowledge/docs/archive/legacy/
- ⚠️  knowledge/service/
- ⚠️  website/service/

**업데이트 시간**: 2026-02-25T20:35:28.059601
