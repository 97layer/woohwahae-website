# Design Tokens — 시각 정체성

> **상위**: THE_ORIGIN.md 제3장 "시각의 침묵"
> **역할**: 시각적 침묵을 구현하는 디자인 시스템(CSS SSOT) 실무 규격.
> **권한**: PROPOSE
> **참조 에이전트**: AD (필수), CE (이미지 선정 시)
> **주의**: 이 문서가 SSOT. style.css 변경 시 반드시 동기화.

---

## 0. 시각의 침묵 (Visual Silence) — 핵심 개념

사유를 담아내는 공간의 기본 컨셉은 **"낡은 종이 위에 0.5mm 연필선으로 그린 동심원의 파동"**입니다.

창백하게 낡아가는 회색빛(#E3E2E0) 종이와 먹처럼 단정한 검은 잉크(#1A1A1A)만이 허용됩니다. 우리는 덧칠하는 것이 아니라 긁어내듯 그립니다.

인간의 호흡에 가장 닿아 있는 **3.8초의 맥동**과 절대적으로 강제된 **60% 이상의 텅 빈 여백**. 이 속에서 화면은 화려한 캔버스가 아니라 당신이 잠시 거주하며 사유를 투기하기 위한 빈 그릇이 됩니다.

가벼움과 무의미한 유행을 경계하기 위해, 우리는 구도자의 도면을 닮은 극단적인 폰트(IBM Plex Mono)만을 건축 자재로 사용합니다.

---

## 1. Colors — 낡은 종이 위의 잉크

### 우선순위 체계

| Token | 값 | 용도 | 우선순위 |
|-------|-----|------|---------|
| `--bg` | #E3E2E0 | 메인 배경 (낡은 종이) | ★★★ PRIMARY |
| `--bg-dark` | #DDDCDA | 보조 배경 | ★★★ PRIMARY |
| `--text` | #1a1a1a | 본문 텍스트 (잉크) | ★★★ PRIMARY |
| `--text-sub` | #4a4a4a | 보조 텍스트 | ★★★ PRIMARY |
| `--text-faint` | #7A7A74 | 희미한 텍스트 (WCAG AA 4.5:1 확보) | ★★★ PRIMARY |
| `--white` | #FFFFFF | 흰색 (반전용) | ★★★ PRIMARY |
| `--line` | #D5D4CF | 구분선 | ★★★ PRIMARY |

### Stone Palette (SECONDARY — Warm Gray 포인트용)

| Token | 값 | 용도 | 우선순위 |
|-------|-----|------|---------|
| `--stone-dark` | #7A6E5A | 어두운 갈색 (그림자, 강조) | ★★ SECONDARY |
| `--stone-mid` | #8B7355 | 중간 갈색 (텍스트 포인트) | ★★ SECONDARY |
| `--stone-light` | #A89880 | 밝은 갈색 (배경 톤) | ★★ SECONDARY |

### Deep Navy (OPTIONAL — 조건부 최소 사용)

| Token | 값 | 용도 | 우선순위 |
|-------|-----|------|---------|
| `--navy` | #1B2D4F | ::selection, blockquote border | ★ OPTIONAL |

**원칙**:

- **Monochrome 기본** (90%+): 블랙(--text) / 그레이(--bg, --line) 계열로 구성
- **Warm Gray 포인트**: Stone 팔레트로 차분하게 강조 (10%)
- **Deep Navy 절제**: ::selection, blockquote border 등 극소수 영역에만 (10% 미만)
- **금지**: 비비드 계열의 과도한 색채 (단정함을 해치는 요소 배제)

**네이비 허용 용도** (예외):

1. `::selection` — 텍스트 선택 시 배경
2. `.art-body blockquote` — 인용문 좌측 테두리 (2px solid)
3. **사용 불가**: placeholder, 범용 border, button background, 기타 장식 요소

**근거**: "Deep Navy의 절제된 사용" 및 "블랙/그레이/우드톤 기반의 차분한 분위기 유지"

---

## 2. Typography

| Token | 값 | 용도 | 상태 |
|-------|-----|------|------|
| `--font-body` | Pretendard Variable | 본문 (한국어 가독성) | ✅ 적용 |
| `--font-mono` | IBM Plex Mono, DM Mono | 레이블, 코드 | ✅ 적용 |
| `--font-serif` | Crimson Text | 포인트 세리프 (최소 사용) | ⚠️ 미적용 (Pretendard 대체) |
| `--font-serif-slab` | Bitter | 슬랩 세리프 | ⚠️ 미적용 |

### Letter-spacing

| Token | 값 | 용도 |
|-------|-----|------|
| `--ls-label` | 0.10em | nav, section-label |
| `--ls-wide` | 0.16em | 카테고리, 필터 버튼 |
| `--ls-heading` | -0.03em | h1, h2 헤딩 |
| `--ls-tight` | -0.01em | 본문 타이틀 |

### Font-weight

| Token | 값 |
|-------|-----|
| `--fw-thin` | 200 |
| `--fw-light` | 300 |
| `--fw-regular` | 400 |
| `--fw-medium` | 500 |

**주의**: 과거 IDENTITY.md에 "본명조"로 기재되어 있었으나 실제 CSS는 Pretendard Variable이다. **Pretendard Variable이 정본.** (THE_ORIGIN.md §3.1에서도 동일하게 정의)

---

## 3. Spacing — 여백 60%+ 원칙

| Token | 값 | 용도 |
|-------|-----|------|
| `--space-xs` | 0.5rem | 최소 간격 |
| `--space-sm` | 1rem | 요소 내부 |
| `--space-md` | 2rem | 섹션 내부 |
| `--space-lg` | 4rem | 섹션 간 |
| `--space-xl` | 8rem | 대형 여백 |
| `--space-2xl` | 12rem | 히어로 여백 |

### Max widths

| Token | 값 |
|-------|-----|
| `--max-content` | 680px |
| `--max-wide` | 960px |

---

## 4. Breath System — 호흡 기반 모션 규칙

> **지향점**: 인간의 편안한 호흡 주기인 ~0.27Hz(3.8초)를 시각적 맥동의 기준으로 삼는다.

### 호흡 주기 토큰

| Token | 값 | 의미 |
|-------|-----|------|
| `--breath` | 3.8s | 한 호흡 주기 (낮 기준) |
| `--breath-half` | 1.9s | 반 호흡 |
| `--breath-quarter` | 0.95s | 1/4 호흡 |

### 시간대별 호흡 변조

방문자 로컬 시간에 따라 호흡 주기 자동 조정:

| 시간대 | 주기 | 특징 |
|--------|------|------|
| dawn (새벽) | 4.8s | 가장 느린 호흡 — 고요함 |
| day (낮) | 3.8s | 기본 호흡 |
| evening (저녁) | 4.2s | 느린 호흡 — 이완 |
| night (밤) | 5.2s | 가장 긴 호흡 — 깊은 사색 |

**적용 예시**:

- Hero 동심원 pulse: `--breath` 주기로 opacity 변화
- Manifesto 호흡: opacity `1 ↔ 0.94` (--breath 주기)
- Archive card stagger: `60ms` (~1/63 breath = 수학적 조화)
- TOC row stagger: `70ms` (~1/54 breath)

### Easing Curves

| Token | 값 | 성격 |
|-------|-----|------|
| `--ease` / `--ease-out` | cubic-bezier(0.16, 1, 0.3, 1) | Apple-style ease-out-expo. 내쉬는 곡선 |
| `--ease-in-out` | cubic-bezier(0.42, 0, 0.58, 1) | 들이쉬고 내쉬는 곡선 |
| `--ease-wave` | cubic-bezier(0.37, 0, 0.63, 1) | 정현파 근사 S곡선 (파동 철학) |

### Duration

| Token | 값 | 용도 |
|-------|-----|------|
| `--duration` | 0.6s | 기본 전환 (hover, click) |
| `--duration-slow` | 1.2s | 중요 요소 (히어로, 핵심 컨텐츠) |

**시간대 변조**: `--duration-slow`도 dawn 1.8s / day 1.2s / night 2.0s로 자동 조정.

**실무 지침 (Mantra)**:
> "장식하지 않고, 단정하게 보여준다."
> "덜어낼 수 있는 것은 모두 덜어낸다."

---

## 5. Glassmorphism

| Token | 값 |
|-------|-----|
| `--glass-surface` | rgba(255, 255, 255, 0) |
| `--glass-blur` | saturate(180%) blur(20px) |
| `--glass-border` | rgba(0, 0, 0, 0.05) |

---

## 6. Photography 규격 (시선의 기록)

- 톤: 채도를 낮춘 차분한 색감 (muted, desaturated)
- 질감: 아날로그 텍스처를 은유하는 35mm 필름 그레인
- 구도: 피사체보다 여백(Negative space)을 우선하는 삼분할법
- 조명: 인위적인 조작 없는 자연광, 깊이를 더하는 그림자
- 금지: 과포화, 튀는 플래시 조명, 시선을 분산시키는 배경, 작위적인 보정

---

## 7. AD 검증 체크리스트

산출물 비주얼 검증 시 AD가 사용하는 기준:

- [ ] 여백 60%+ 확보
- [ ] 컬러가 토큰 범위 내
- [ ] 서체가 지정된 4종 내
- [ ] 이미지가 Photography 기준 충족
- [ ] 과장된 시각 요소 부재

---

## 8. Family Look — 패밀리룩 디자인 프레임

> 모든 페이지가 같은 주파수에서 출발한다. 이 프레임이 WOOHWAHAE의 시각적 DNA.

### 8.1 키워드 (시각 언어의 근간)

| 키워드 | 시각적 번역 | 적용 |
|--------|-----------|------|
| **Essentialism** | 60%+ 여백. 요소를 뺄수록 완성. | 모든 페이지 공통 |
| **Resonance (공명)** | 극세선(0.5px), 순차 페이드, 선이 그려지는 드로잉 | SVG, 트랜지션 |
| **Self-care** | 호흡 리듬 (3.8s). 급하지 않은 트랜지션. | 모션, 타이밍 |

### 8.1.1 마이크로 디테일 감도 (패밀리룩의 핵심)

패밀리룩은 큰 요소가 아니라 **작은 것의 정밀함**에서 나온다:

| 요소 | 규격 | 느낌 |
|------|------|------|
| SVG 선 | `stroke-width: 0.5` | 종이 위의 연필선 |
| SVG 노드 | `r: 3.5px`, opacity 단계 (0.9→0.7→0.5→0.35) | 시간의 무게감 |
| SVG 레이블 | IBM Plex Mono 9px, 300, ls 0.1em | 도면의 주석 |
| 드로잉 모션 | `stroke-dashoffset` 2s ease-out | 선이 그려지는 과정 |
| 순차 페이드 | 0.4s 간격 stagger (1.2s→1.6s→2.0s→2.4s) | 하나씩 드러남 |
| blockquote | navy 1px 좌측 border + Crimson Text italic | 인용의 격 |
| 읽기 진행 | 우측 2px bar, stone-dark fill | 독서의 리듬 |
| 제목 이탈릭 | font-weight 200, italic, clamp(1.5-2.2rem) | 손글씨 무게 |
| 헤더 하단선 | `border-bottom: 1px solid var(--line)` | 장을 나누는 여백 |

**원칙**: 큰 요소를 추가하지 말고, 작은 요소의 두께·크기·속도를 정밀하게 조율한다.

### 8.2 배경 레이어 시스템

모든 페이지는 3겹 레이어로 구성:

```
z-index: 0  — 배경 (파동)
z-index: 1  — 콘텐츠
z-index: 200 — 네비게이션
```

| 페이지 유형 | z-0 배경 | 철학적 근거 | 구현 |
|------------|---------|------------|------|
| **Home** | Three.js 쌍극자 자기장 캔버스 | **공명의 물리적 은유**: 보이지 않는 파동장(Field of Resonance)을 시각화. 중심에서 퍼지는 자기력선 = THE_ORIGIN §1.2 "자발적 파동 전이". 극세선(0.5px) 대신 3D 곡선으로 깊이 렌더링. 홈은 "첫 인상 = 파동장 진입" 연출 필요. | `bg-field.js` + `<canvas class="field-bg">` |
| 내부 페이지 | SVG 동심원 파동 | 2D 극세선 파동 (호흡 주기 3.8s). 텍스트 방해하지 않는 정적 배경. | `<div class="wave-bg">` + SMIL animate |
| 에세이 읽기 | 없음 (텍스트 집중) | 여백 60%+ 극대화. 독서 몰입 우선. | 배경 비움 |

#### Three.js Field 규격 (Home 전용)

**파라미터**:
- **색상**: `#E3E2E0` (배경 톤과 동기화)
- **선 밀도**: 데스크톱 96개 / 모바일 32개
- **곡선 정밀도**: 데스크톱 500 포인트 / 모바일 120 포인트
- **안개(Fog)**: `0xE3E2E0` opacity 0.035 (깊이감)
- **마우스 인터랙션**: 관성 추적 (고요한 반응)
- **접근성**: prefers-reduced-motion 감지 시 비활성화

**허용 근거**:
1. 홈만 3D → 내부는 2D 동심원 (계층 분리)
2. 쌍극자 자기장 = 공명(Resonance) 물리 법칙 시각화
3. 모바일 최적화 (LINE_COUNT 96→32, POINTS 500→120)
4. 배경 톤 `#E3E2E0` 동기화 → "낡은 종이" 질감 유지

### 8.3 SVG 파동 (wave-bg) 규격

```html
<div class="wave-bg" aria-hidden="true"
  style="position:fixed;top:50%;left:50%;
         transform:translate(-50%,-50%);
         width:120vmax;height:120vmax;
         z-index:0;pointer-events:none;opacity:0.6">
  <svg viewBox="0 0 800 800" width="100%" height="100%">
    <g fill="none" stroke="#D5D4CF" stroke-width="0.5" opacity="0.4">
      <!-- 4개 동심원: 60 → 140 → 240 → 360 반지름 -->
      <!-- dur: 3.8s (호흡 주기) -->
      <!-- 바깥 원일수록 opacity 낮게 (0.4 → 0.12) -->
      <!-- begin: 0s / 0.4s / 0.8s / 1.2s (시차) -->
    </g>
  </svg>
</div>
```

**파라미터**:
- 색상: `#D5D4CF` (--line 토큰)
- stroke-width: 0.5 (극세선)
- 호흡 주기: 3.8s (--breath 토큰과 동기화)
- 반지름 변화: ±20px (미세한 맥동)
- 전체 opacity: 0.6 (배경에 녹아듦)

### 8.4 모션 패밀리

| 이름 | 용도 | CSS |
|------|------|-----|
| `wave-fade-in` | 페이지 진입 시 콘텐츠 등장 | `opacity 0→1, translateY 12→0, 2.4s ease-wave` |
| `line-breathe` | 구분선/CTA 호흡 | `opacity 1↔0.5, --breath 주기` |
| `home-breathe` | 홈 로고 호흡 | `opacity 1↔0.94, 시간대별 주기` |
| `hb-fade` | 초기 등장 | `opacity 0→1, 1.2s ease-out` |
| `[data-reveal]` | 스크롤 진입 | `opacity 0→1, translateY 16→0, IO threshold 0.12` |

**규칙**:
- 모든 트랜지션은 `--ease-wave` (정현파 S곡선) 사용
- 급격한 움직임 금지. 최소 0.6s duration.
- 카드류 stagger: 60ms 간격 (호흡의 1/63)

### 8.5 레이아웃 프레임

```
┌─────────────────────────────────────────────┐
│ nav-brand          Archive  Practice  About  │ ← 고정 상단
├─────────────────────────────────────────────┤
│                                             │
│         [wave-bg: 동심원 파동]               │ ← z-0, fixed
│                                             │
│    ┌──────── max 680px ────────┐            │
│    │  section-label            │            │ ← z-1
│    │  headline / manifesto     │            │
│    │                           │            │
│    │  body text                │            │
│    └───────────────────────────┘            │
│                                             │
│    ┌──────── max 960px ────────┐            │
│    │  card grid (2col)         │            │
│    └───────────────────────────┘            │
│                                             │
├─────────────────────────────────────────────┤
│ footer: brand · nav · legal · ©             │
└─────────────────────────────────────────────┘
```

### 8.6 새 페이지 체크리스트

- [ ] wave-bg SVG 삽입 (에세이 제외)
- [ ] page-main에 `position:relative; z-index:1`
- [ ] 텍스트 영역 `container--content` (680px)
- [ ] 그리드 영역 `container--wide` (960px)
- [ ] 진입 애니메이션: headline에 `wave-fade-in`
- [ ] 카드류에 `[data-reveal]` + stagger
- [ ] nav-overlay 모바일 메뉴 포함
- [ ] CSS cache buster 갱신

---

## 9. Symbol & Wordmark — 낙관(落款)의 규격

> 심볼은 전각(篆刻) 인장 형태의 우화(羽化) 나비 문양. 외부 디자이너 협업 산출물.

### 9.1 에셋 파일

| 파일 | 용도 | 비고 |
|------|------|------|
| `symbol.svg` | 벡터 원본 (인쇄, 확대) | SSOT |
| `symbol.png` | 웹 표시 (nav, OG) | 흰색 배경 |
| `symbol.jpg` | OG 이미지 (SNS 공유) | |
| `icon-192.png` | PWA 아이콘 | 192×192 |
| `icon-512.png` | PWA 아이콘 | 512×512 |

경로: `website/assets/media/brand/`

### 9.2 심볼 사용 규격

**최소 크기**: 24px (웹), 8mm (인쇄)

**보호 영역 (Clear Space)**: 심볼 높이의 50%를 사방에 확보. 다른 요소가 침범 금지.

**허용 배경**:
- `--bg` (#E3E2E0) — 기본. 심볼은 검정(#1A1A1A).
- `--white` (#FFFFFF) — 허용. 심볼은 검정.
- 다크 배경 — 심볼을 흰색(#FFFFFF) 반전 사용. (반전 에셋 미구축 — 필요 시 생성)

**금지 사용**:
- 심볼 비율 왜곡 (가로/세로 독립 스케일 금지)
- 심볼 위에 텍스트 오버레이
- 그림자, 글로우, 외곽선 추가
- 배경 대비 불충분한 상태에서 사용 (심볼이 읽히지 않는 조합)
- 컬러 변형 (모노톤 흑/백만 허용)

### 9.3 워드마크

텍스트 로고: `WOOHWAHAE` — IBM Plex Mono, 300 weight, letter-spacing 0.10em, uppercase.

**조합 규격**:
- 심볼 단독 사용 허용 (nav, favicon, 앱 아이콘)
- 워드마크 단독 사용 허용 (footer, 명함, 서명)
- 심볼 + 워드마크 병기 시: 심볼 좌측, 워드마크 우측. 수직 중앙 정렬. 간격 = 심볼 너비의 50%.

### 9.4 웹 적용 현황

| 위치 | 에셋 | 크기 |
|------|------|------|
| Nav 브랜드 | `symbol.png` (img) | height: 32px |
| Favicon | `symbol.png` | 브라우저 자동 |
| Apple Touch | `icon-192.png` | 192×192 |
| OG Image | `symbol.jpg` | SNS 썸네일 |
| Three.js 중심 | `symbol.png` (texture) | 1.2×1.2 plane |

---

---

## 9. 웹 경험 설계 (구 experience.md 흡수)

### 9.1 사이트 구조 (IA)

```text
woohwahae.kr/
├── about/          — 브랜드 서사 (철학, 이야기, 에디터)
├── archive/        — 매거진 (essay-NNN, magazine/, lookbook/)
├── practice/       — 수행의 실천 (아틀리에, 디렉션, 프로젝트, 프로덕트, 연락)
├── woosunho/       — 에디터 포트폴리오
└── lab/            — 실험 (네비 미노출)
```

네비: `Archive | Practice | About` — 고정 상단, 드롭다운 없음, 깊이 1단계.

### 9.2 섹션별 톤

| 섹션 | 톤 | 참조 |
|------|-----|------|
| `/about/` | 단정하고 묵직한 1인칭 서술 | THE_ORIGIN Part I |
| `/archive/` | 한다체. 깊이 있는 통찰 | content.md |
| `/practice/` | 정확하고 정중. 수식 배제 | service.md |

### 9.3 기술 스택

| 영역 | 선택 | 이유 |
| :--- | :--- | :--- |
| 프론트엔드 | 순수 HTML/CSS/JS | 오버엔지니어링 방지, 통제권 |
| 반응형 | 모바일 우선 | 독자 환경 대응 |
| 배포 | Cloudflare Pages | git push = 자동 배포 |
| 자동 발행 | content_publisher.py → git push | 파이프라인 자동화 |

### 9.4 레이아웃 원칙

- 화면의 60% 이상을 비워두는 여백 우선 확보
- 콘텐츠 최대 너비: 680px (텍스트), 960px (그리드)
- 색상/서체: 본 문서 §1-§3 디자인 시스템 참조

### 9.5 금지된 경험 (Anti-Patterns)

- 팝업, 모달, 강제 구독 배너
- 좋아요 수 표기, 강압적 공유 유도
- 조회수를 좇는 얕은 콘텐츠 배치
- 외부 광고 삽입
- 맥락 없이 유행을 쫓는 레이아웃 변경

---

**Last Updated**: 2026-02-27
