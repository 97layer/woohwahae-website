# Design Tokens — 시각 정체성

> **소스**: website/assets/css/style.css `:root` 변수 문서화
> **권한**: PROPOSE
> **참조 에이전트**: AD (필수), CE (이미지 선정 시)
> **주의**: 이 문서가 SSOT. style.css 변경 시 반드시 동기화.

---

## 1. Colors — 낡은 종이 위의 잉크

### 우선순위 체계
| Token | 값 | 용도 | 우선순위 |
|-------|-----|------|---------|
| `--bg` | #E4E2DC | 메인 배경 (낡은 종이) | ★★★ PRIMARY |
| `--bg-dark` | #E8E7E2 | 보조 배경 | ★★★ PRIMARY |
| `--text` | #1a1a1a | 본문 텍스트 (잉크) | ★★★ PRIMARY |
| `--text-sub` | #4a4a4a | 보조 텍스트 | ★★★ PRIMARY |
| `--text-faint` | #8E8E88 | 희미한 텍스트 (캡션, 날짜) | ★★★ PRIMARY |
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
- **Monochrome 기본** (90%+): 블랙(--text) / 그레이(--bg, --line) 계열
- **Warm Gray 포인트**: Stone 팔레트로 강조 (10%)
- **Deep Navy 절제**: ::selection, blockquote border만 (10% 미만)
- **절대 금지**: 비비드 계열 (FOUNDATION 인터뷰 §I 기준)

**네이비 허용 용도** (예외):
1. `::selection` — 텍스트 선택 시 배경
2. `.art-body blockquote` — 인용문 좌측 border (2px solid)
3. **금지**: placeholder, 범용 border, button background, 기타 장식 요소

**근거**: 브랜드 v7.0 "Deep Navy 절제 사용" + FOUNDATION 인터뷰 "블랙/그레이/우드톤 차분한 계열"

---

## 2. Typography

| Token | 값 | 용도 |
|-------|-----|------|
| `--font-body` | Pretendard Variable | 본문 (한국어 가독성) |
| `--font-mono` | IBM Plex Mono, DM Mono | 레이블, 코드 |
| `--font-serif` | Crimson Text | 포인트 세리프 (최소 사용) |
| `--font-serif-slab` | Bitter | 슬랩 세리프 |

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

**주의**: IDENTITY.md에 "본명조"로 기재되어 있으나 실제 CSS는 Pretendard Variable이다. **Pretendard Variable이 정본.**

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

## 4. Breath System — 호흡 기반 애니메이션

> **브랜드 차별점**: 인간 호흡 주기 ~0.27Hz (3.8초)를 브랜드 파동 철학으로 전환.

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

**Daily Mantra**:
> "조용히, 확실하게."
> "소음을 제거하면 본질이 남는다."

---

## 5. Glassmorphism

| Token | 값 |
|-------|-----|
| `--glass-surface` | rgba(255, 255, 255, 0) |
| `--glass-blur` | saturate(180%) blur(20px) |
| `--glass-border` | rgba(0, 0, 0, 0.05) |

---

## 6. Photography 기준

- 톤: muted, desaturated
- 질감: 35mm 필름 그레인
- 구도: rule of thirds, negative space
- 조명: 자연광, 그림자, 시간의 흔적
- 금지: 과포화, 플래시 조명, 복잡한 배경, 과도한 보정

---

## 7. AD 검증 체크리스트

산출물 비주얼 검증 시 AD가 사용하는 기준:

- [ ] 여백 60%+ 확보
- [ ] 컬러가 토큰 범위 내
- [ ] 서체가 지정된 4종 내
- [ ] 이미지가 Photography 기준 충족
- [ ] 과장된 시각 요소 부재

---

**Last Updated**: 2026-02-24
