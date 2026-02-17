# WOOHWAHAE Website — 구조 맵

> 에이전트/개발자가 파일 찾을 때 이 문서를 먼저 읽을 것.
> 마지막 갱신: 2026-02 (로고 복원 + 카피 정리 완료)

---

## 배포

- **URL**: `woohwahae.kr` → GitHub Pages
- **레포**: `97layer/woohwahae-website` (branch: `website`)
- **배포 방법**: `git push origin main:website` → GitHub Actions 자동 빌드
- **워크플로**: `.github/workflows/deploy-pages.yml`

---

## 파일 구조

```
website/
├── index.html              ← 메인 (허브 페이지)
├── about.html              ← 브랜드 소개
├── atelier.html            ← 서비스 철학 + 예약 CTA
├── contact.html            ← 연락처
├── CNAME                   ← woohwahae.kr (GitHub Pages 도메인)
├── manifest.webmanifest    ← PWA 매니페스트
├── sw.js                   ← Service Worker (현재 v3)
│
├── archive/
│   ├── index.html          ← 아카이브 목록 (index.json에서 동적 로드)
│   └── index.json          ← 포스트 목록 [{slug, title, date, preview}]
│   └── {slug}/
│       └── index.html      ← 개별 포스트 (어드민에서 자동 생성)
│
└── assets/
    ├── css/
    │   └── style.css       ← 단일 CSS (CSS 변수 시스템)
    ├── js/
    │   └── main.js         ← 최소 JS (fade-in, nav toggle, SW 등록)
    └── img/
        ├── symbol.jpg      ← ⚠️ 오리지널 로고 (우화 도장 심볼). 이것만 사용.
        ├── icon-192.png    ← PWA 아이콘
        └── icon-512.png    ← PWA 아이콘
```

---

## 핵심 규칙

### 로고
- **반드시 `symbol.jpg`** 사용. SVG 버전은 삭제됨 (AI 생성 가짜였음).
- nav: `<img src="assets/img/symbol.jpg" class="nav-symbol" alt="WOOHWAHAE">`
- footer: `<img src="assets/img/symbol.jpg" width="28" class="footer-brand-symbol">`
- archive 서브디렉토리에서는 `../assets/img/symbol.jpg`

### CSS 변수 (style.css)
```css
--bg:        #FAFAF7   /* 메인 배경 */
--bg-dark:   #F2F1EC   /* 어두운 섹션 배경 */
--text:      #1A1A1A
--text-sub:  #6B6B6B
--text-faint:#AEAEA8
--navy:      #1B2D4F
--font-kr:   'Noto Serif KR'
--font-en:   'Crimson Text'
```

### 섹션 배경 패턴
- 교차 리듬: `section--light` (#FAFAF7) ↔ `section--dark` (#F2F1EC)
- 모든 `<footer>`에 반드시 `class="section--light"` 추가

### 페이지 공통 구조
```html
<nav> ... </nav>
<div class="container">
  <div class="page-header"> ... </div>  ← 서브 페이지 헤더
</div>
<section class="section--dark"> ... </section>
<section class="section--light"> ... </section>
<footer class="section--light"> ... </footer>
<script src="assets/js/main.js"></script>
```

### 카피 원칙
- 직접적이고 구체적으로. 추상 은유 최소화.
- 짧은 문장 우선.
- 금지 표현: "삶의 리듬", "오늘의 밀도", "우주적", "현상이자 아카이브"

---

## 어드민 연동

- **어드민**: `core/admin/app.py` (Flask)
- 포스트 발행 시 `website/archive/{slug}/index.html` + `website/archive/index.json` 자동 생성
- 발행 후 `git push origin main:website` → 자동 배포
- 포스트 개별 페이지 CSS 클래스: `.post-body` (style.css에 정의됨)

---

## Service Worker 캐시 버전

현재: `woohwahae-v3`
CSS/에셋 변경 시 `website/sw.js`의 `CACHE_VERSION` 올릴 것.
