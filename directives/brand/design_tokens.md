# Design Tokens — 시각 정체성

> **소스**: website/assets/css/style.css `:root` 변수 문서화
> **권한**: PROPOSE
> **참조 에이전트**: AD (필수), CE (이미지 선정 시)
> **주의**: 이 문서가 SSOT. style.css 변경 시 반드시 동기화.

---

## 1. Colors — 낡은 종이 위의 잉크

| Token | 값 | 용도 |
|-------|-----|------|
| `--bg` | #E4E2DC | 메인 배경 (낡은 종이) |
| `--bg-dark` | #E8E7E2 | 보조 배경 |
| `--text` | #1a1a1a | 본문 텍스트 |
| `--text-sub` | #4a4a4a | 보조 텍스트 |
| `--text-faint` | #8E8E88 | 희미한 텍스트 |
| `--navy` | #1B2D4F | 포인트 컬러 (Deep Navy) |
| `--white` | #FFFFFF | 흰색 |
| `--line` | #D5D4CF | 구분선 |

### Stone Palette
| Token | 값 | 용도 |
|-------|-----|------|
| `--stone-dark` | #7A6E5A | 어두운 갈색 (그림자) |
| `--stone-mid` | #8B7355 | 중간 갈색 (텍스트) |
| `--stone-light` | #A89880 | 밝은 갈색 (배경 톤) |

**원칙**: Monochrome 기본. 컬러는 이유가 있을 때만. Aesop 시각 언어 70%+ 수준.

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

인간 호흡(~0.27Hz)에 기반한 파동 시스템.

| Token | 값 | 의미 |
|-------|-----|------|
| `--breath` | 3.8s | 한 호흡 주기 |
| `--breath-half` | 1.9s | 반 호흡 |
| `--breath-quarter` | 0.95s | 1/4 호흡 |

### Easing Curves
| Token | 값 | 용도 |
|-------|-----|------|
| `--ease-out` | cubic-bezier(0.16, 1, 0.3, 1) | 내쉬는 곡선 |
| `--ease-in-out` | cubic-bezier(0.42, 0, 0.58, 1) | 들이쉬고 내쉬는 곡선 |
| `--ease-wave` | cubic-bezier(0.37, 0, 0.63, 1) | 정현파 S곡선 |

### Transitions
| Token | 값 |
|-------|-----|
| `--duration` | 0.6s |
| `--duration-slow` | 1.2s |

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
