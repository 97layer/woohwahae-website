# WOOHWAHAE — Archive for Slow Life

슬로우 라이프를 기록하고 실천하는 플랫폼

## 📖 프로젝트 소개

WOOHWAHAE는 **Archive for Slow Life**를 컨셉으로 하는 브랜드 웹사이트입니다.
가속화된 세상에서 주체적으로 살아가기 위한 생각과 실천을 기록하며,
울산 남구에서 운영하는 헤어 아틀리에의 철학을 담고 있습니다.

**핵심 가치**:
- **실용적 미학**: 삶의 맥락을 해치지 않는 간결함
- **무언의 교감**: 언어를 넘어선 감각적 소통
- **자기 긍정**: 타인의 인정으로부터 독립된 내면의 질서

## 🎯 주요 기능

### 1. **About** - 철학의 선언
브랜드의 정체성과 슬로우 라이프에 대한 정의

### 2. **Archive** - 생각의 기록
- 슬로우 라이프 관련 에세이, 실천 방법
- 검색 및 카테고리 필터 기능
- Essay 기반 시리즈 형태

### 3. **Atelier** - 실천의 공간
- 울산 헤어 아틀리에 소개
- 예약 시스템 연동 (네이버 예약)
- 서비스 철학 및 접근 방식

### 4. **Playlist** - 공간의 소리
아틀리에 공간을 위한 음악 큐레이션

### 5. **Project** - 협업의 실험
브랜드 협업 및 프로젝트 (Coming Soon)

### 6. **Photography** - 순간의 포착
일상의 순간들을 담은 사진

## 🛠 기술 스택

### 프론트엔드
- **HTML5** - 시맨틱 마크업
- **CSS3** - 커스텀 디자인 시스템 (Pretendard 폰트)
- **Vanilla JavaScript** - 의존성 없는 경량 구현
- **PWA** - 오프라인 지원, 앱 설치 가능

### 백엔드 (CMS)
- **Flask** - Python 웹 프레임워크
- **JSON 기반 콘텐츠 관리** - 데이터베이스 없는 경량 구조

### 인프라
- **Service Worker** - 오프라인 캐싱
- **Google Analytics** - 사용자 행동 분석
- **SEO 최적화** - sitemap, robots.txt, Open Graph

## 📁 프로젝트 구조

```
website/
├── index.html              # 메인 페이지
├── /about/              # About 페이지
├── /practice/           # 아틀리에 소개
├── playlist.html           # Playlist
├── project.html            # Project (Coming Soon)
├── /photography/        # Photography
├── 404.html                # 404 에러 페이지
├── privacy.html            # 개인정보처리방침
├── terms.html              # 이용약관
├── robots.txt              # 검색 엔진 크롤러 설정
├── sitemap.xml             # 사이트맵
├── manifest.webmanifest    # PWA 매니페스트
├── sw.js                   # Service Worker
│
├── archive/                # Archive 섹션
│   ├── index.html          # Archive 메인
│   ├── index.json          # Archive 콘텐츠 데이터
│   ├── essay-00/           # Essay 00 - Manifesto
│   ├── essay-001-beginning/
│   ├── essay-002-slow-life/
│   └── essay-003-hair-and-daily/
│
├── assets/
│   ├── css/
│   │   ├── style.css       # 메인 스타일시트
│   │   └── _archive/       # 백업 CSS 파일들
│   ├── js/
│   │   ├── main.js         # 메인 JavaScript
│   │   ├── archive.js      # Archive 검색/필터
│   │   ├── analytics.js    # Google Analytics
│   │   └── photography.js  # Photography 갤러리
│   ├── img/
│   │   ├── symbol.jpg      # 로고/심볼
│   │   ├── icon-192.png    # PWA 아이콘
│   │   └── icon-512.png
│   └── uploads/            # 업로드된 이미지
│
└── backend/                # CMS 백엔드
    ├── app.py              # Flask 애플리케이션
    ├── config.py           # 설정 파일
    ├── requirements.txt    # Python 의존성
    ├── .env.example        # 환경 변수 예시
    └── README.md           # 백엔드 문서
```

## 🚀 로컬 개발 환경 설정

### 1. 프로젝트 클론
```bash
git clone https://github.com/yourusername/woohwahae.git
cd woohwahae/website
```

### 2. 정적 사이트 실행
```bash
# Python 내장 서버 사용
python3 -m http.server 8000

# 또는 Node.js 사용
npx http-server -p 8000
```

브라우저에서 `http://localhost:8000` 접속

### 3. CMS 백엔드 실행 (선택 사항)
```bash
cd backend

# 가상환경 생성
python3 -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate

# 의존성 설치
pip install -r requirements.txt

# 환경 변수 설정
cp .env.example .env
nano .env  # 환경 변수 편집

# 서버 실행
python app.py
```

백엔드는 `http://localhost:5000`에서 실행됩니다.

## 📝 콘텐츠 추가 방법

### 방법 1: CMS API 사용
```bash
curl -X POST http://localhost:5000/api/archive \
  -H "Content-Type: application/json" \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -d '{
    "slug": "essay-005-new-topic",
    "title": "새로운 주제",
    "date": "2026.03.01",
    "issue": "Issue 005",
    "preview": "이 글은...",
    "category": "Essay",
    "content": "<p>전체 내용...</p>"
  }'
```

### 방법 2: 수동으로 추가
1. `archive/index.json`에 새 항목 추가
2. `archive/essay-XXX/` 폴더 생성
3. `index.html` 파일 작성
4. Git commit 및 push

## 🎨 디자인 시스템

### 색상 팔레트
```css
--bg:         #FAFAF7   /* 베이지 화이트 */
--text:       #1A1A1A   /* 다크 그레이 */
--text-sub:   #6B6B6B   /* 중간 그레이 */
--navy:       #1B2D4F   /* 네이비 (강조) */
--line:       #E4E3DC   /* 라인 색상 */
```

### 타이포그래피
- **폰트**: Pretendard Variable (한글), System fonts (영문)
- **크기**: 18px 기본, 유동적 타이포그래피 (clamp 사용)
- **행간**: 2.0 (넉넉한 가독성)

### 레이아웃
- **최대 너비**: 680px (본문), 960px (와이드)
- **간격**: 0.5rem ~ 12rem (일관된 spacing scale)

## 🔧 유지보수

### CSS 업데이트
주 스타일시트: `assets/css/style.css`
- 백업 파일은 `assets/css/_archive/`에 보관

### Archive 콘텐츠 관리
- `archive/index.json` - 모든 게시글 메타데이터
- 새 글 추가 시 최상단에 추가 (최신순)
- `sitemap.xml` 업데이트 필수

### SEO 업데이트
```bash
# Sitemap 업데이트 (새 페이지 추가 시)
nano sitemap.xml

# Google Search Console에서 재제출
```

## 📊 분석 및 모니터링

### Google Analytics 설정
1. GA4 계정 생성
2. 측정 ID 획득
3. 모든 HTML `<head>`에 추가:
```html
<meta name="ga-id" content="G-YOUR-MEASUREMENT-ID">
```

### 추적 이벤트
- 페이지뷰
- 스크롤 깊이 (25%, 50%, 75%, 100%)
- 버튼 클릭 (예약, 문의)
- 세션 시간

## 🔒 보안

- 환경 변수는 `.env` 파일에 보관 (Git 제외)
- HTTPS 필수 (프로덕션)
- CORS 정책 설정
- 관리자 비밀번호 강력하게 설정

## 📱 브라우저 지원

- Chrome/Edge (최신 2버전)
- Firefox (최신 2버전)
- Safari (최신 2버전)
- iOS Safari (최신 2버전)
- Android Chrome (최신 2버전)

## 🚢 배포

자세한 배포 방법은 [DEPLOYMENT.md](DEPLOYMENT.md) 참고

**간단 배포**:
```bash
# Netlify
netlify deploy --prod

# Vercel
vercel --prod
```

## 📄 라이선스

Copyright © 2026 WOOHWAHAE. All rights reserved.

## 📞 문의

- **이메일**: hello@woohwahae.kr
- **Instagram**: [@woohwahae](https://instagram.com/woohwahae)
- **예약**: [네이버 예약](https://map.naver.com/p/entry/place/1017153611)

---

**Editor & Curator**: 우순호
**Location**: 울산광역시 남구
**Platform**: Archive for Slow Life
