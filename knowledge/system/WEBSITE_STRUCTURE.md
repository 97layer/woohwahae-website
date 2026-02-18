# WOOHWAHAE Website Structure

> **버전**: 1.0
> **갱신**: 2026-02-17
> **목적**: 슬로우라이프·미니멀 라이프의 매거진 B — 웹사이트 세션 구조 정의

---

## I. 전체 구조

```
woohwahae.kr/
├── about/          — 브랜드 철학, 순호 이야기, 공간 소개
├── archive/        — 매거진 (크롤링 기반 브랜드 해석 + 자체 사유)
├── shop/           — 제품/굿즈 (미래)
├── service/        — 헤어 오프라인 예약
├── playlist/       — 큐레이션된 음악 (아틀리에 BGM)
├── project/        — 협업 아카이브 (브랜드/사진작가/공간)
└── photography/    — 비주얼 아카이브 (공간/제품/분위기)
```

---

## II. 세션별 상세

### 1. `/about/` — 브랜드 철학

**목적**: WOOHWAHAE가 누구인가, 왜 존재하는가

**콘텐츠**:
- **Manifesto**: "Remove the Noise, Reveal the Essence" 선언문
- **97layer 소개**: 순호, 반지하 8평, 슬로우라이프 실천
- **공간**: 아틀리에 사진 + 공간 철학
- **Contact**: 예약 방법, 위치, 운영 시간

**톤**: 1인칭 에세이. 개인적이지만 보편적.

**예시 구조**:
```
/about/
  index.html        — Manifesto + 브랜드 개요
  space.html        — 아틀리에 공간 소개
  contact.html      — 연락처 + 오시는 길
```

---

### 2. `/archive/` — 매거진 (핵심)

**목적**: 슬로우라이프 브랜드 관찰 저널리즘 + WOOHWAHAE 자체 사유

**포맷**: 이슈 단위 (Issue NN)
- 월간 발행 목표
- 한 이슈 = 한 브랜드 or 한 개념 깊이 탐구
- 5000-8000자 장문 에세이

**콘텐츠 소싱**:
1. **크롤링 기반** (Brand Scout 에이전트)
   - 타겟 브랜드 SNS/웹사이트 자동 분석
   - SA Agent → 전략 분석 → 테마 추출
   - CE Agent → 에세이 형태 재구성
2. **자체 사유** (순호의 텔레그램 신호)
   - 개인 경험 + 브랜드 관찰 → 에세이

**이슈 예시**:
- Issue 00: WOOHWAHAE Manifesto (파일럿)
- Issue 01: 제주 Mute — 고요함을 파는 숙소
- Issue 02: 도쿄 D&DEPARTMENT — 롱라이프 디자인
- Issue 03: 서울 어니언 — 독립 출판의 지속 가능성
- Issue 04: WOOHWAHAE — 본질주의 헤어 아틀리에

**구조**:
```
/archive/
  index.html              — 이슈 목록 (그리드 또는 리스트)
  issue-00/index.html     — Issue 00 본문
  issue-01/index.html     — Issue 01 본문
  ...
```

**메타데이터** (`archive/index.json`):
```json
[
  {
    "slug": "issue-00-manifesto",
    "title": "WOOHWAHAE Manifesto",
    "date": "2026-02-17",
    "preview": "소음을 제거하면 본질이 남는다.",
    "themes": ["브랜드 아이덴티티", "슬로우라이프"],
    "cover_image": "issue-00/cover.jpg"
  }
]
```

---

### 3. `/service/` — 헤어 오프라인 예약

**목적**: 아틀리에 방문 예약 시스템

**기능**:
- 예약 가능 시간 표시 (Calendly 통합 or 자체 시스템)
- 메뉴 + 가격
- 준비 사항 안내
- 취소 정책

**톤**: 정확하고 정중. 불필요한 수식 없이.

**구조**:
```
/service/
  index.html        — 메뉴 + 예약 안내
  book.html         — 예약 폼 (Calendly embed or 자체)
```

---

### 4. `/playlist/` — 음악 큐레이션

**목적**: 아틀리에 BGM, 순호의 음악 취향 공유

**형태**:
- Spotify/Apple Music 플레이리스트 embed
- 월별 플레이리스트 (계절/분위기별)
- 짧은 해설 (왜 이 곡인가)

**예시**:
- "2026.02 — 겨울 끝자락의 고요함"
- "WOOHWAHAE Atelier Essentials"

**구조**:
```
/playlist/
  index.html        — 플레이리스트 목록
  2026-02.html      — 개별 플레이리스트 페이지
```

---

### 5. `/project/` — 협업 아카이브

**목적**: 브랜드/사진작가/공간과의 협업 기록

**콘텐츠 예시**:
- "OO 브랜드와의 팝업"
- "사진작가 AA와의 비주얼 작업"
- "BB 공간에서의 워크숍"

**구조**:
```
/project/
  index.html                — 프로젝트 목록
  2026-popup-store/         — 개별 프로젝트 페이지
```

---

### 6. `/photography/` — 비주얼 아카이브

**목적**: 공간/제품/분위기 사진 아카이브

**형태**:
- 그리드 레이아웃 (Pinterest 스타일)
- 캡션 최소화 (사진이 말하게)
- 테마별 필터 (공간/제품/순간)

**구조**:
```
/photography/
  index.html        — 사진 그리드
  [image-id].html   — 개별 사진 상세 (선택)
```

---

### 7. `/shop/` — 제품/굿즈 (미래)

**목적**: WOOHWAHAE 제품 판매 (나중에)

**예시**:
- 헤어 케어 제품
- 브랜드 굿즈 (노트, 포스터)
- 협업 제품

**보류 이유**: 초기엔 브랜드 구축에 집중. 제품은 신뢰 쌓인 후.

---

## III. 시각 정체성 (전 세션 공통)

### 레이아웃 원칙

- **여백**: 화면의 60% 이상
- **타이포그래피**:
  - 국문: 본명조 (정갈함)
  - 영문: Crimson Text (우아한 세리프)
- **컬러**:
  - 기본: Monochrome (흑백)
  - 포인트: Deep Navy (#1a1a2e), Warm Gray (#e8e8e8)
- **이미지**:
  - 자연광, 그림자, 시간의 흔적
  - 과도한 보정 금지

### 네비게이션

```
[WOOHWAHAE]  About  Archive  Service  Playlist  Project  Photography
```

- 고정 상단 네비게이션
- 모바일: 햄버거 메뉴
- 현재 세션 표시 (active 상태)

---

## IV. 기술 스택

### 프론트엔드

- **순수 HTML/CSS/JS** (오버엔지니어링 방지)
- **프레임워크 없음** — 속도 + 통제권
- **반응형** — 모바일 우선 설계

### 백엔드 (필요시)

- **정적 생성 우선** (GitHub Pages)
- **동적 필요 시**: Flask/FastAPI (예약 시스템)

### 배포

- **GitHub Pages** (woohwahae.kr CNAME)
- **자동 배포**: [content_publisher.py](../../core/system/content_publisher.py) → git push

---

## V. Archive → Magazine 전환 로드맵

### Phase 1: 파일럿 (현재 ~ 1개월)

- ✅ 웹사이트 구조 설계
- [ ] Issue 00: WOOHWAHAE Manifesto 작성
- [ ] `/about/`, `/archive/` 최소 구현
- [ ] 기존 [content_publisher.py](../../core/system/content_publisher.py) → `/archive/` 발행

### Phase 2: Brand Scout (1-3개월)

- [ ] [brand_scout.py](../../core/agents/brand_scout.py) 에이전트 구현
- [ ] 크롤링 → 브랜드 분석 파이프라인
- [ ] Issue 01-03: 외부 브랜드 3개 발행

### Phase 3: 구독 시스템 (3-6개월)

- [ ] 이메일 구독 (Substack/Ghost 통합)
- [ ] PDF 다운로드 기능
- [ ] 월간 발행 정규화

### Phase 4: 확장 (6개월~)

- [ ] `/service/` 예약 시스템
- [ ] `/playlist/`, `/project/`, `/photography/` 활성화
- [ ] 커뮤니티 기능 (댓글/반응)

---

## VI. 핵심 원칙 (Forbidden)

- ❌ 알고리즘 추종 — 발행 주기를 SEO에 맞추지 않음
- ❌ 광고 — 중립성 유지, 수익은 구독/제품
- ❌ 과도한 인터랙션 — 좋아요/공유 버튼 최소화
- ❌ 트렌드 추종 — 유행하는 브랜드 아닌, 본질적 브랜드

---

**Last Updated**: 2026-02-17
**Version**: 1.0
**소유자**: 97layer
**시스템**: 97layerOS

> "archive for slowlife" — WOOHWAHAE
