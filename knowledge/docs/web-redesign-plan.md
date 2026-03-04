# WOOHWAHAE 웹사이트 리디자인 기획안
> 작성: 2026-03-04 | 상태: **확정** (사용자 승인 완료)

---

## 1. 핵심 원칙

| 원칙 | 내용 |
|------|------|
| **절대 보존** | `website/about/index.html` 텍스트 내용 수정 금지 |
| **절대 보존** | `website/index.html` Three.js 메인 필드 배경 |
| **패밀리룩** | 필드 감성 연장 — 종이 질감, 먹 잉크, 호흡 주기 |
| **폰트** | Pretendard Variable + IBM Plex Mono 만 |
| **색상** | 모노크롬만 (`--bg` `--text` `--text-sub` `--text-faint` `--line`) |
| **타이포** | 대문짝 금지 — 최대 28px (clamp 기반) |
| **전환** | 300ms 이상, fade only |

---

## 2. IA (Information Architecture)

```
woohwahae.kr/
│
├── /                        홈
│   └── Three.js 필드 + 브랜드 선언(최소 텍스트) + 섹션 네비
│
├── /archive/                아카이브
│   ├── 에세이   (essay-000 ~ 현재, 역시간순)
│   ├── 저널     (journal-001 ~ 현재)
│   ├── 룩북     (lookbook/)
│   └── 플레이리스트  (Spotify 임베드)
│
├── /practice/               프랙티스
│   ├── 아틀리에  (헤어 아틀리에 — CTO 서비스)
│   ├── 디렉팅    (상품화/확장 방향)
│   ├── 프로젝트  (프로젝트 1, 2... 확장 폼 형태)
│   └── 프로덕트  (상품화 확장 페이지)
│
├── /about/                  어바웃
│   ├── THE ORIGIN 텍스트 (001~006 챕터) ← 절대 수정 금지
│   └── 편집자의 궤적 섹션 → /about/woosunho/ 연동
│
└── /lab/                    랩 (자율, nav 미노출 유지)

Nav:    Archive  ·  Practice  ·  About
Footer: Contact (이메일 + 인스타) · 저작권
```

---

## 3. 디자인 토큰 (기존 style.css 기반 조정)

### 색상 — 변경 없음
```css
--bg:         #E3E2E0;
--text:       #1A1A1A;
--text-sub:   #4A4A4A;
--text-faint: #7A7A74;
--line:       #EFEEED;
```

### 타이포그래피 스케일 — 최대 28px로 제한
```css
/* 기존 clamp 값 상한 조정 */
--type-xl:   clamp(22px, 2vw + 16px, 28px);   /* 섹션 헤딩 최대 */
--type-lg:   clamp(18px, 1.5vw + 14px, 22px); /* 소제목 */
--type-md:   clamp(15px, 1vw + 12px, 18px);   /* 본문 */
--type-sm:   14px;                              /* 캡션 */
--type-xs:   12px;                              /* 레이블 (Mono) */
```

### 스페이싱 — 픽셀 기반으로 정비
```css
--space-xs:  8px;
--space-sm:  16px;
--space-md:  24px;
--space-lg:  32px;
--space-xl:  48px;
--space-2xl: 64px;
--space-3xl: 96px;
```

### 레이아웃
```css
--content-width:   720px;   /* 본문 최대폭 */
--content-padding: 32px;    /* 좌우 여백 */
/* Mobile: 100% - 32px padding */
/* Tablet(768px+): max 640px */
/* Desktop(1024px+): max 720px */
```

---

## 4. 컴포넌트

### Nav (`website/_components/nav.html`)
```
[WOOHWAHAE 로고/심볼] ————————————— [Archive] [Practice] [About]
높이: 56px | sticky | transparent → bg 스크롤 시
```

### Footer (`website/_components/footer.html`)
```
[이메일] · [인스타그램] · [© 2026 WOOHWAHAE]
높이: 96px 이상 | contact 링크 포함
```

---

## 5. 페이지별 골조

### Home (`/`)
```
[Three.js 필드 — fullscreen canvas, z-index 0]
[Nav — sticky, transparent]

[Hero — 화면 중앙]
  브랜드 선언 한 줄 (최대 22px, Pretendard 300)
  ↓ scroll indicator

[섹션 네비 — 필드 위 오버레이]
  01 Archive
  02 Practice
  03 About

[Footer]
```

### Archive (`/archive/`)
```
[Nav]

[페이지 헤더]
  "Archive" — label (Mono 12px, ls 0.16em)

[탭 네비]
  에세이 · 저널 · 룩북 · 플레이리스트
  (Mono, 12px, 밑줄 active)

[콘텐츠 리스트]
  에세이/저널: 날짜 · 제목 · 1줄 요약 · 구분선
  룩북: 이미지 그리드 (2col Desktop, 1col Mobile)
  플레이리스트: Spotify iframe embed (allow="autoplay; clipboard-write; encrypted-media; fullscreen; picture-in-picture")

[Footer]
```

### Practice (`/practice/`)
```
[Nav]

[페이지 헤더]
  "Practice" — label

[섹션 탭 or 앵커 네비]
  아틀리에 · 디렉팅 · 프로젝트 · 프로덕트

[아틀리에 섹션]
  헤어 아틀리에 소개 (CTO)
  예약/연락 정보

[디렉팅 섹션]
  방향성 텍스트 (상품화/확장)

[프로젝트 섹션] — JSON 리피터 방식
  projects.json에 항목 추가 → JS가 카드 자동 렌더링
  카드 구조: 번호 · 제목 · 연도 · 한 줄 설명 · 링크
  → 새 프로젝트: projects.json에 한 줄 추가만으로 확장

[프로덕트 섹션]
  상품 목록 (확장 가능)
  현재: 워크북 등 기존 product/

[Footer (Contact 포함)]
```

### About (`/about/`)
```
[Nav]

[챕터 네비 — 좌측 고정 (Desktop), 상단 스크롤 (Mobile)]
  001 · 002 · 003 · 004 · 005 · 006

[본문 — 기존 텍스트 100% 보존]
  001. 시선 · 서(序)
  002. 원리 · 소거(消去)
  003. 태도 · 느림
  004. 이름 · 우화해(羽化解)
  005. 실천 · 표현 방식
  006. 순환 · 발(跋)

[편집자의 궤적 섹션 — 신규]
  "편집자의 궤적" label
  한 줄 소개 → /about/woosunho/ 링크

[Footer]
```

---

## 6. 인터랙션 규칙

| 요소 | 동작 |
|------|------|
| 링크 hover | border-bottom 1px solid var(--text), transition 300ms |
| 페이지 전환 | fade opacity 0→1, 300ms |
| 스크롤 | smooth-scroll, behavior: smooth |
| Nav 스크롤 | transparent → bg(--bg), transition 300ms |
| 이미지 | lazy loading, 최대폭 100% |
| Spotify | iframe, responsive (padding-bottom: 56.25%) |

---

## 7. 구현 순서

| 순서 | 작업 | 파일 |
|------|------|------|
| 1 | about 텍스트 백업 | `knowledge/docs/about-backup.html` |
| 2 | CSS 토큰 타이포 조정 | `website/assets/css/style.css` |
| 3 | nav/footer 컴포넌트 업데이트 | `website/_components/` |
| 4 | 홈 골조 (필드 유지) | `website/index.html` |
| 5 | archive 골조 (4탭) | `website/archive/index.html` |
| 6 | practice 골조 (4섹션) | `website/practice/index.html` |
| 7 | about 편집자의 궤적 섹션 추가 | `website/about/index.html` |
| 8 | build_components 실행 | `python3 core/scripts/build.py --components` |
| 9 | 시각 검증 (Puppeteer) | 전체 페이지 스크린샷 |
| 10 | git push → CF Pages | 배포 |

---

## 8. 검증 체크리스트

- [ ] about 텍스트 원본과 diff 0 확인
- [ ] Three.js 필드 정상 렌더링
- [ ] 전체 페이지 Mono + Pretendard 외 폰트 없음
- [ ] 타이포 최대 28px 초과 없음
- [ ] 모노크롬 외 색상 없음
- [ ] Spotify iframe 재생 확인
- [ ] 모바일 320px~767px 레이아웃 정상
- [ ] Nav Contact 없음, Footer Contact 있음
- [ ] 모든 링크 정상 동작

---

## 9. 절대 금지

- `website/about/index.html` 텍스트 내용 수정
- Three.js 필드 코드 제거 또는 교체
- Pretendard / IBM Plex Mono 외 폰트 추가
- 28px 초과 타이포
- 컬러(비-모노크롬) 추가
- 11ty 전환 (별도 마이그레이션 프로젝트로 분리)
