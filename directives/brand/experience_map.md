# Experience Map — 웹사이트 경험 설계

> **소스**: WEBSITE_STRUCTURE.md 흡수
> **권한**: PROPOSE
> **이 문서 생성 후**: WEBSITE_STRUCTURE.md는 이 문서를 참조하는 리다이렉트로 대체.

---

## 1. 사이트 구조 (IA)

```
woohwahae.kr/
├── about/          — 브랜드 철학, 순호 이야기, 공간 소개
├── archive/        — 매거진 (브랜드 해석 + 자체 사유)
├── service/        — 헤어 아틀리에 예약
├── playlist/       — 큐레이션된 음악
├── project/        — 협업 아카이브
├── photography/    — 비주얼 아카이브
└── shop/           — 제품/굿즈 (미래)
```

**WOOHWAHAE는 7개 섹션이 하나의 철학을 표현하는 다른 방식들이다.**

---

## 2. 섹션별 상세

### `/about/` — 철학 선언
- Manifesto: "Remove the Noise, Reveal the Essence"
- 97layer(순호) 소개 + 공간(반지하 8평) 철학
- 톤: 1인칭 에세이. 개인적이지만 보편적.

### `/archive/` — 매거진 (핵심 섹션)
- 이슈 단위 발행 (Issue NN)
- 한 이슈 = 한 브랜드 or 한 개념 깊이 탐구
- 5000-8000자 장문 에세이
- 소싱: Brand Scout(크롤링) + 개인 사유(텔레그램 신호)
- 파일 구조: `website/archive/issue-{NNN}-{slug}/index.html`

### `/service/` — 헤어 아틀리에 예약
- 예약 시간 표시, 메뉴/가격, 준비 사항, 취소 정책
- 톤: 정확하고 정중. 불필요한 수식 없이.

### `/playlist/` — 음악 큐레이션
- Spotify/Apple Music embed
- 월별/계절별 플레이리스트 + 짧은 해설

### `/project/` — 협업 아카이브
- 브랜드/사진작가/공간과의 협업 기록

### `/photography/` — 비주얼 아카이브
- 그리드 레이아웃. 캡션 최소화 — 사진이 말하게.
- 테마별 필터: 공간/제품/순간

### `/shop/` — 제품/굿즈 (보류)
- 초기엔 브랜드 구축 집중. 제품은 신뢰 쌓인 후.

---

## 3. 네비게이션

```
[WOOHWAHAE]  About  Archive  Service  Playlist  Project  Photography
```

- 고정 상단 네비게이션
- 모바일: 햄버거 메뉴
- 현재 섹션 active 표시
- 최소주의: 드롭다운 없음, 깊이 1단계

---

## 4. 기술 스택

| 영역 | 선택 | 이유 |
|------|------|------|
| 프론트엔드 | 순수 HTML/CSS/JS | 오버엔지니어링 방지, 속도, 통제권 |
| 프레임워크 | 없음 | 의도적 선택 |
| 반응형 | 모바일 우선 | 독자 환경 대응 |
| 배포 | GCP VM + nginx | 자체 호스팅, 통제권 |
| 자동 발행 | content_publisher.py → git push | 파이프라인 자동화 |

---

## 5. 레이아웃 원칙

- 여백: 화면의 60%+
- 최대 콘텐츠 너비: 680px (에세이), 960px (그리드)
- 색상: design_tokens.md 참조
- 서체: design_tokens.md 참조
- 이미지: 자연광, 그림자, 시간의 흔적. 과도한 보정 금지.

---

## 6. 금지 사항

- 알고리즘 추종 발행 주기
- 광고 삽입 (수익은 구독/제품)
- 과도한 인터랙션 (좋아요/공유 버튼 최소화)
- 트렌드 추종 콘텐츠
- 팝업, 모달, 뉴스레터 강제 구독 배너

---

**Last Updated**: 2026-02-24
