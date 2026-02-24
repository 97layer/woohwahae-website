# Phase 30 — 웹사이트 완전 리빌딩 Master Plan

> **목표**: WOOHWAHAE 웹사이트를 브랜드 A-Z 매뉴얼 기반으로 전면 재설계
> **범위**: UI/UX/GUI/내러티브/모바일 최적화/검증/완전 배포
> **기준**: Magazine B 철학 + 동아시아 수행 문화 + Slow Life 브랜드
> **날짜**: 2026-02-25

---

## 구조 — 6개 레이어, 30개 Phase

### 🎯 Layer 1: Foundation (Phase 1-5) — ✅ 완료

**Phase 1**: 브랜드 메뉴얼 품질 업그레이드
- ✅ design_tokens.md 색상 우선순위 체계 (PRIMARY/SECONDARY/OPTIONAL)
- ✅ BRAND_MANUAL.md 동기화
- ✅ Breath System 시간대별 호흡 변조

**Phase 2**: CSS 토큰 시스템 정비
- ✅ style.css `:root` 주석 추가 (PRIMARY/SECONDARY/OPTIONAL)
- ✅ 5605라인 CSS 검토 완료

**Phase 3**: 전체 HTML 감사 스크립트
- ✅ audit_html_simple.py 작성
- ✅ 82개 파일 감사 완료

**Phase 4**: Navy 색상 금지 패턴 제거
- ✅ Flask 템플릿 3개 (portal.html, consult.html, consult_done.html)
- ✅ Lab 2개 (design-system.html, prototype-offering.html)
- ✅ style.css 4개 인스턴스
- ✅ photography.html placeholder

**Phase 5**: 전수조사 1차 완료
- ✅ 핵심 페이지 6개 + Archive 10개 감사 완료
- ✅ 실질적 이슈 0개 (Flask nav/og는 의도된 설계)

---

### 🎨 Layer 2: Visual Consistency (Phase 6-12)

**Phase 6**: 타이포그래피 통합
- [ ] font-family 직접 지정 → 토큰 변환
- [ ] font-size 일관성 검증 (hero/body/caption)
- [ ] letter-spacing 토큰 사용 검증
- [ ] line-height 일관성 (body 1.9, heading 1.2-1.4)

**Phase 7**: 색상 시스템 완성
- [ ] Hardcoded 색상 → 토큰 변환 (meta theme-color 제외)
- [ ] Stone palette 적용 범위 확대
- [ ] Dark mode 대비 CSS 변수 구조 설계

**Phase 8**: 스페이싱 시스템 검증
- [ ] padding/margin 하드코딩 → 토큰 변환
- [ ] 여백 60%+ 원칙 준수 검증
- [ ] section 간격 일관성 (--space-lg / --space-xl)

**Phase 9**: 반응형 Breakpoint 통일
- [ ] 480px / 768px / 1024px 3단계 통일
- [ ] 모바일 font-size 일관성 (-0.2rem 규칙)
- [ ] Grid/Flex 모바일 대체 구조

**Phase 10**: 애니메이션 시스템 정비
- [ ] Breath System 적용 확대 (fade-in, pulse, stagger)
- [ ] Easing curve 통일 (--ease / --ease-wave)
- [ ] Duration 토큰 사용 검증

**Phase 11**: 이미지 시스템 최적화
- [ ] lazy loading 전체 적용
- [ ] WebP 변환 + fallback
- [ ] aspect-ratio CSS 적용
- [ ] 35mm 필름 그레인 필터 표준화

**Phase 12**: Glassmorphism/투명도 정비
- [ ] --glass-surface / --glass-blur 사용 검증
- [ ] backdrop-filter 브라우저 호환성

---

### 🧩 Layer 3: Component Unification (Phase 13-18)

**Phase 13**: Nav 컴포넌트 통일
- [ ] 82개 파일 nav 구조 표준화
- [ ] Archive / Offering / About / Contact / Lab 순서 통일
- [ ] 모바일 햄버거 메뉴 일관성
- [ ] nav-toggle 애니메이션 통일

**Phase 14**: Footer 컴포넌트 통일
- [ ] footer-grid 구조 표준화
- [ ] Navigate / Connect 섹션 일관성
- [ ] 저작권 표기 통일 (2026 WOOHWAHAE)

**Phase 15**: Button/CTA 스타일 통일
- [ ] .btn / .btn-primary / .btn-ghost 3종 정의
- [ ] hover 상태 일관성 (stone-dark background)
- [ ] min-height 44px (터치 최적화)

**Phase 16**: Form Input 컴포넌트
- [ ] input/textarea/select 스타일 통일
- [ ] focus 상태 border-color: stone-mid
- [ ] placeholder 색상 통일 (--text-faint)

**Phase 17**: Card 컴포넌트 시스템
- [ ] .arc-row / .arc-card 스타일 정비
- [ ] hover elevation 통일
- [ ] stagger animation 일관성

**Phase 18**: Loading/Transition 컴포넌트
- [ ] Skeleton loader 디자인
- [ ] Page transition fade-in 통일
- [ ] Lazy load placeholder

---

### 📝 Layer 4: Content & Narrative (Phase 19-22)

**Phase 19**: 톤앤보이스 일관성
- [ ] Archive (한다체) / Magazine (합니다체) 명확히 분리
- [ ] INFP 톤 검증 (느린 속도, 열린 결말, 은유)
- [ ] 과장 표현 제거 ("최고", "완벽", "혁신" 금지)

**Phase 20**: 텍스트 품질 검증
- [ ] 오타/맞춤법 전수조사
- [ ] 날짜 형식 통일 (YYYY.MM.DD)
- [ ] 레이블 대소문자 (UPPERCASE for mono font)

**Phase 21**: Placeholder/Empty State
- [ ] placeholder-block 5종 디자인 (gradient/stone/lines/typo)
- [ ] Empty state 문구 통일
- [ ] Loading state 메시지

**Phase 22**: Meta Description/SEO
- [ ] og:title / og:description / og:image 전체 검증
- [ ] meta description 150자 이내 최적화
- [ ] title 형식 통일 ("제목 — WOOHWAHAE")

---

### 🔧 Layer 5: Technical Excellence (Phase 23-27)

**Phase 23**: 접근성 (A11y)
- [ ] img alt 속성 전수조사
- [ ] ARIA 레이블 (button, nav, form)
- [ ] Semantic HTML 검증 (header/main/section/article)
- [ ] 색상 대비 4.5:1 이상

**Phase 24**: 성능 최적화
- [ ] Font preconnect/preload
- [ ] CSS 파일 minify + cache 버전 (?v=)
- [ ] JS bundle 크기 검증
- [ ] Critical CSS 추출

**Phase 25**: Flask 템플릿 상속 구조
- [ ] base.html 베이스 템플릿 생성
- [ ] block 구조 정의 (head/nav/content/footer)
- [ ] portal/consult 템플릿 상속 리팩토링

**Phase 26**: Error 페이지
- [ ] 404 페이지 디자인 (브랜드 톤 반영)
- [ ] 500 에러 페이지
- [ ] Flask error handler 등록

**Phase 27**: 보안 헤더
- [ ] Content-Security-Policy
- [ ] X-Frame-Options
- [ ] X-Content-Type-Options
- [ ] Referrer-Policy

---

### ✅ Layer 6: Validation & Deployment (Phase 28-30)

**Phase 28**: 핵심 User Flow E2E 테스트
- [ ] `/me/{token}` — Ritual Module 포털
  - 실제 고객 토큰으로 접근
  - 실루엣 렌더링 검증
  - 방문 기록 표시 검증
- [ ] `/consult/{token}` — 사전 상담 폼
  - 폼 제출 테스트
  - 이미지 업로드 검증
  - consult_done 리다이렉트
- [ ] `archive/` — 에세이 목록/개별 페이지
  - 10개 에세이 로딩 검증
  - TOC 내비게이션
  - 모바일 반응형

**Phase 29**: Cross-Browser/Device 검증
- [ ] Desktop (Chrome/Safari/Firefox)
- [ ] Mobile (iOS Safari/Android Chrome)
- [ ] Tablet (iPad)
- [ ] 480px / 768px / 1024px breakpoint 실제 테스트

**Phase 30**: Git Commit + VM 배포 + 검증
- [ ] Git commit (모든 변경사항)
- [ ] VM 배포 (`/deploy` 커맨드)
- [ ] systemctl restart 97layer-ecosystem / woohwahae-backend
- [ ] 라이브 사이트 검증 (https://woohwahae.kr)
- [ ] Lighthouse 점수 (Performance 90+, A11y 95+)

---

## 우선순위 매트릭스

### 🔴 CRITICAL (Phase 1-5, 13-14, 28, 30)
즉시 수정 없이는 배포 불가능한 항목
- ✅ Phase 1-5: Foundation 완료
- Phase 13-14: Nav/Footer 통일 (브랜드 일관성 핵심)
- Phase 28: E2E 테스트 (기능 작동 검증)
- Phase 30: 배포 + 검증

### 🟡 HIGH (Phase 6-12, 19-22)
사용자 경험에 직접 영향
- Phase 6-12: Visual Consistency (일관성 = 브랜드 신뢰)
- Phase 19-22: Content/Narrative (WOOHWAHAE 차별점)

### 🟢 MEDIUM (Phase 15-18, 23-27)
품질 향상 + 기술 부채 해소
- Phase 15-18: Component 시스템 (유지보수성)
- Phase 23-27: Technical Excellence (SEO/A11y/성능)

### 🔵 OPTIONAL (Phase 26-27)
Nice-to-have
- Phase 26: Error 페이지 (우선순위 낮음)
- Phase 27: 보안 헤더 (이미 기본 설정 있음)

---

## 실행 전략

### 자동화 우선
- 스크립트로 해결 가능한 것 (색상 변환, 메타태그 검증)은 Python/Bash 도구 작성
- 수동 검토는 최소화

### Batch 처리
- 같은 패턴 반복 작업은 한 번에 (예: 82개 파일 nav 통일)

### 단계별 커밋
- 5-Phase 단위로 커밋 (rollback 용이)

### 에이전트 협업
- AD (Art Director) — Visual/Typography 검증
- CE (Content Editor) — 텍스트/톤앤보이스
- Code Agent (본인) — 구현 + 배포

---

**Last Updated**: 2026-02-25
**Status**: Phase 1-5 완료 (17% 진행)
**Next**: Phase 6 (타이포그래피 통합) 시작
