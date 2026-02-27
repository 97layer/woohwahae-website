# 🧠 INTELLIGENCE QUANTA - 지능 앵커

> **목적**: 어떤 모델/세션이 오더라도 사고 흐름이 끊기지 않도록 보장하는 물리적 앵커
> **갱신 정책**: 덮어쓰기 (최신 상태만 유지). session-stop 훅이 자동 갱신.
> **마지막 갱신**: 2026-02-27 (SAGE-ARCHITECT v4.0 전면 정렬)
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

- **Ver**: 12.0 — SAGE-ARCHITECT 대개편 (멀티에이전트 위계 재설계)
- **웹**: Cloudflare Pages (`git push` = 자동 배포, woohwahae.kr)
- **API**: `api.woohwahae.kr` → VM `136.109.201.201` (nginx reverse proxy)
- **VM 서비스 (5개 active)**: 97layer-telegram / 97layer-ecosystem / 97layer-gardener / woohwahae-backend / cortex-admin
- **배포**: 웹=`git push` / 코드=`git push` → VM `git fetch + reset --hard origin/main` + restart
- **파이프라인**: 신호 유입 → signal.schema.json → SA 분석 → Gardener 군집화 → CE 에세이 → git push → CF Pages 자동 발행
- **빌드**: `python3 core/scripts/build.py` — archive → components → cache bust 원커맨드

---

## 🎯 다음 작업

1. [NAMING] ~~**소문자 통일**~~ ✅ — 전 파일명 lowercase 리네임 + state.md 완료
2. [DESIGN] **디자인 레이아웃 전 페이지 확정** — 홈/archive/practice/about + 모바일
3. [DESIGN] **홈 히어로 SVG 리디자인** — dipole field 유지 중. 방향 보류.
4. [INFRA] **`<head>` 컴포넌트화** — 디자인 확정 후 진행
5. [INFRA] content_publisher.py — 에세이 발행 자동화 (CF Pages 연동)
6. [CONTENT] the_origin.md 언어 수정 — SAGE-ARCHITECT 발현 후
7. Ralph 피드백 루프 — STAP 자동 검증 + Gardener 수정 제안

**완료됨**:

- ✅ DNS A레코드 연결 (Cloudflare 경유)
- ✅ HTTPS/SSL (certbot, Let's Encrypt)
- ✅ VM git 초기화 + GitHub remote 연결
- ✅ 4축 구조 정렬 Ver 11.0 (d6a448b0)
- ✅ website HTML 리빌딩 — 네비/푸터 전체 통일, 깨진 참조 0건
- ✅ 비주얼 Phase 1 — 캔버스 복구, 패밀리룩 디자인 프레임, 배경색 통일, SVG 히어로
- ✅ 인프라 구조조정 Phase 2-3-5 — nginx API 분리, deploy.sh git pull 전환, 레거시 스크립트 삭제, 유령 프로세스 정리
- ✅ **THE ORIGIN 생태계 완전 구축** — 61파일 레거시 전면 청산 (a66f8df0)
- ✅ **Cloudflare Pages 연결 완료** — woohwahae.kr + www + api DNS 설정
- ✅ **컴포넌트 통일** — nav/footer/wave-bg 단일 소스 (6e104a64)
  - `website/_components/` 4개 프래그먼트 (nav, footer-contact, footer-archive, wave-bg)
  - `core/scripts/build_components.py` — 마커 기반 빌드 (--init, --dry-run, --file)
  - 26개 HTML 마커 삽입 + 표준 컴포넌트 주입
  - 전 경로 루트 상대(`/archive/`, `/assets/css/`) 통일
  - 에세이 bare nav → 표준 site-nav + overlay 통일
  - site.js active nav 로직 개선 (overlay 포함, section 기반)
  - `_templates/article.html` 마커 교체 + 루트 상대경로
  - `filesystem_validator.py` archive/index.html 예외 추가 (8f6107e9)
- ✅ **통합 빌드 파이프라인** — `core/scripts/build.py` (3c8c020c)
  - archive → components → cache bust 원커맨드
  - CSS MD5 해시 기반 자동 캐시 버스팅
- ✅ **AI_CONSTITUTION.md** — SSOT 이동, CLAUDE.md 참조화 (다른 에이전트)

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
# 웹사이트 빌드 (로컬)
python3 core/scripts/build.py              # 전체: archive → components → cache bust
python3 core/scripts/build.py --components # 컴포넌트만
python3 core/scripts/build.py --bust       # 캐시 버스팅만

# 웹 배포 (Cloudflare Pages)
git push origin main                    # 30초 내 woohwahae.kr 반영

# 코드 배포 (VM)
git push origin main                    # GitHub에 push
ssh 97layer-vm "cd /home/skyto5339_gmail_com/97layerOS && git fetch origin main && git reset --hard origin/main"
ssh 97layer-vm "sudo systemctl restart 97layer-telegram 97layer-ecosystem 97layer-gardener woohwahae-backend cortex-admin"

# 상태 확인
ssh 97layer-vm "for s in 97layer-telegram 97layer-ecosystem 97layer-gardener woohwahae-backend cortex-admin; do printf '%-25s %s\n' \$s \$(systemctl is-active \$s); done"
```

---

## 📍 현재 상태 (CURRENT STATE)

### [2026-02-27 23:30] SAGE-ARCHITECT 대개편 완료 — claude-opus-4-6

**커밋**: `2f0b6077` — 38파일 변경, +1312 -2429

**실행 완료 (6 Phase)**:

1. **sage_architect.md 생성** — language.md 격상 + CD.md §10 흡수 (FROZEN)
2. **KERNEL 적출** — the_origin.md → 순수 철학 에세이. KERNEL → system.md 이관
3. **3파일 통합** — glossary→sage_architect 부록, MANIFEST+AI_CONSTITUTION→system.md v10.0
4. **에이전트 동기화** — SA/CE/AD 인격 뿌리 경로 변경, 코드 5파일 수정
5. **파일 흡수** — experience→visual, audience→content, GEMINI→SYSTEM, origin_editor 삭제
6. **전역 스위프** — 중복 README 3개 제거, pre-commit hook 갱신, filesystem_cache 정리

**결과**: directives/ 17→9파일. 프로젝트 전체 34→23 md.

**최종 구조**:
```
directives/ (9파일)
├── the_origin.md        경전 (FROZEN)
├── sage_architect.md    창시자 + 용어 (FROZEN)
├── system.md            실행의 전부 (v10.0)
├── practice/
│   ├── visual.md        시각 + 웹 경험
│   ├── content.md       콘텐츠 + 공명 대상
│   └── service.md       아틀리에
└── agents/
    └── sa.md, ce.md, ad.md
```

**다음 태스크**:
1. [NAMING] ~~소문자 통일~~ ✅ — 전 파일명 lowercase + state.md 리네임 완료
2. [DESIGN] 디자인 레이아웃 전 페이지 확정
3. [INFRA] content_publisher.py 자동화
4. [CONTENT] the_origin.md 언어 수정 (SAGE-ARCHITECT 발현 후)

**주의사항**:
- 홈 index.html 에디토리얼 스타일 변경됨. canvas 확인 필요.
- `build_archive.py` ROOT 경로 버그 미수정 (`parents[2]`가 맞음).
- test_queue_manager 8개 실패 — 기존 이슈. 이번 변경과 무관.
- pre-commit hook MANIFEST.md→system.md 갱신 완료 (.git/hooks/pre-commit)

**업데이트 시간**: 2026-02-27T23:30:00

work_lock: unlocked
