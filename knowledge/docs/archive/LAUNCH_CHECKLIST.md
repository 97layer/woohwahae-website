# WOOHWAHAE 웹사이트 런칭 체크리스트

## 🚀 배포 전 필수 사항

### 1. 환경 변수 설정
- [ ] `backend/.env` 파일 생성 및 설정
  - [ ] `FLASK_SECRET_KEY` - 보안 키 설정
  - [ ] `FLASK_DEBUG=False` - 프로덕션 환경
  - [ ] `ADMIN_PASSWORD` - 강력한 비밀번호

### 2. Google Analytics 설정
- [ ] GA4 계정 생성
- [ ] 측정 ID 획득
- [ ] 모든 HTML 파일 `<head>`에 추가:
  ```html
  <meta name="ga-id" content="G-YOUR-MEASUREMENT-ID">
  ```

### 3. 도메인 및 URL 업데이트
- [ ] `robots.txt` - Sitemap URL 업데이트
- [ ] `sitemap.xml` - 모든 URL을 실제 도메인으로 변경
- [ ] `manifest.webmanifest` - start_url 업데이트
- [ ] `backend/config.py` - CORS_ORIGINS에 실제 도메인 추가

### 4. 콘텐츠 확인
- [ ] Archive에 최소 5-10개 게시글 준비
- [ ] About 페이지 내용 최종 검토
- [ ] Atelier 페이지 영업 정보 확인
- [ ] Contact 정보 확인 (이메일, Instagram, 예약 링크)

### 5. 이미지 및 에셋
- [ ] 로고/심볼 이미지 최적화
- [ ] PWA 아이콘 (192x192, 512x512)
- [ ] Open Graph 이미지 준비 (1200x630)
- [ ] 아틀리에 공간 사진 (선택)
- [ ] Photography 갤러리 이미지 (선택)

## 🧪 기능 테스트

### 페이지 로딩
- [ ] 홈페이지 (index.html)
- [ ] About
- [ ] Archive (목록 + 개별 게시글)
- [ ] Atelier
- [ ] Shop
- [ ] Playlist
- [ ] Project
- [ ] Photography
- [ ] 404 페이지
- [ ] 개인정보처리방침
- [ ] 이용약관

### 네비게이션
- [ ] 메인 메뉴 모든 링크 작동
- [ ] Footer 링크 작동
- [ ] 모바일 햄버거 메뉴 작동
- [ ] Breadcrumb (해당 시)

### Archive 기능
- [ ] 게시글 목록 로드
- [ ] 검색 기능 작동
- [ ] 카테고리 필터 작동
- [ ] 개별 게시글 페이지 로드
- [ ] 뒤로가기 버튼 작동

### Newsletter
- [ ] 이메일 입력 및 제출
- [ ] 유효성 검사 작동
- [ ] 성공 메시지 표시
- [ ] 에러 메시지 표시

### 외부 링크
- [ ] Instagram 링크
- [ ] 네이버 예약 링크
- [ ] 이메일 링크 (mailto:)
- [ ] 모든 외부 링크 `target="_blank"` 및 `rel="noopener"` 포함

## 📱 반응형 테스트

### 데스크톱
- [ ] Chrome (최신)
- [ ] Firefox (최신)
- [ ] Safari (최신)
- [ ] Edge (최신)

### 모바일
- [ ] iPhone (Safari)
- [ ] Android (Chrome)
- [ ] 태블릿 (iPad)

### 브레이크포인트
- [ ] 320px (모바일 S)
- [ ] 375px (모바일 M)
- [ ] 768px (태블릿)
- [ ] 1024px (데스크톱)
- [ ] 1440px (대형 데스크톱)

## 🔍 SEO 확인

### 메타 태그
- [ ] 모든 페이지에 `<title>` 태그
- [ ] 모든 페이지에 `<meta description>` 태그
- [ ] Open Graph 태그 (og:title, og:description, og:image)
- [ ] Twitter Card 태그 (선택)

### 구조화된 데이터
- [ ] Schema.org 마크업 (Organization, LocalBusiness)
- [ ] Breadcrumb 스키마 (해당 시)

### 파일
- [ ] `robots.txt` 접근 가능
- [ ] `sitemap.xml` 접근 가능 및 유효성
- [ ] `manifest.webmanifest` 접근 가능

### Google 도구
- [ ] Google Search Console에 사이트 등록
- [ ] Sitemap 제출
- [ ] Google Analytics 작동 확인
- [ ] PageSpeed Insights 점수 확인 (목표: 90+)

## ⚡ 성능 최적화

### 이미지
- [ ] 모든 이미지 최적화 (WebP 변환 권장)
- [ ] 적절한 이미지 크기 (과도한 해상도 방지)
- [ ] Lazy loading 적용 (`loading="lazy"`)
- [ ] Alt 텍스트 추가

### CSS/JS
- [ ] CSS 압축 (프로덕션)
- [ ] JavaScript 압축 (프로덕션)
- [ ] 불필요한 코드 제거
- [ ] 폰트 최적화 (서브셋 사용 고려)

### 캐싱
- [ ] Service Worker 작동 확인
- [ ] 브라우저 캐시 설정 (Nginx/Apache)
- [ ] CDN 사용 고려 (Cloudflare 권장)

### 성능 점수
- [ ] Lighthouse Performance > 90
- [ ] Lighthouse Accessibility > 90
- [ ] Lighthouse Best Practices > 90
- [ ] Lighthouse SEO > 90

## 🔒 보안

### HTTPS
- [ ] SSL 인증서 설치 (Let's Encrypt)
- [ ] HTTP → HTTPS 리다이렉트 설정
- [ ] Mixed Content 경고 없음

### 환경 변수
- [ ] `.env` 파일 Git에서 제외 (.gitignore)
- [ ] 민감한 정보 코드에 하드코딩 안 함
- [ ] 프로덕션 환경 변수 안전하게 관리

### 헤더
- [ ] Content-Security-Policy 설정 (선택)
- [ ] X-Frame-Options 설정
- [ ] X-Content-Type-Options 설정

## ♿ 접근성

### 스크린 리더
- [ ] 모든 이미지에 alt 텍스트
- [ ] ARIA 레이블 적절히 사용
- [ ] 폼 요소에 label 연결
- [ ] 링크에 의미 있는 텍스트

### 키보드 네비게이션
- [ ] Tab 키로 모든 요소 접근 가능
- [ ] Focus 상태 시각적으로 표시
- [ ] Skip to content 링크 (선택)

### 색상 대비
- [ ] WCAG AA 기준 충족 (최소 4.5:1)
- [ ] WAVE 도구로 검증

## 📊 분석 및 모니터링

### Google Analytics
- [ ] 페이지뷰 트래킹 작동
- [ ] 이벤트 트래킹 작동
- [ ] 실시간 사용자 확인

### 에러 모니터링
- [ ] 브라우저 콘솔 에러 없음
- [ ] 404 에러 처리
- [ ] 500 에러 처리
- [ ] 에러 로깅 시스템 (선택)

## 🚢 배포

### 배포 방법 선택
- [ ] 옵션 1: 정적 호스팅 (Netlify/Vercel) ← 권장
- [ ] 옵션 2: VPS (Nginx + Gunicorn)
- [ ] 옵션 3: Docker

### 배포 후 확인
- [ ] 사이트 정상 접근
- [ ] 모든 페이지 로딩 확인
- [ ] SSL 인증서 작동
- [ ] 리다이렉트 작동 (www → non-www 등)

## 📝 문서화

### README
- [ ] 프로젝트 설명
- [ ] 설치 방법
- [ ] 사용 방법
- [ ] 기여 가이드 (선택)

### 배포 가이드
- [ ] DEPLOYMENT.md 작성 완료
- [ ] 환경 변수 문서화
- [ ] 백업 전략 명시

## 🎉 런칭 후

### 즉시
- [ ] 모든 기능 재확인
- [ ] 성능 모니터링 시작
- [ ] Google Search Console에서 색인 요청
- [ ] 소셜 미디어에 공유 (Instagram 등)

### 첫 주
- [ ] 사용자 피드백 수집
- [ ] Analytics 데이터 검토
- [ ] 버그/이슈 수정
- [ ] 콘텐츠 정기 업데이트 시작

### 첫 달
- [ ] SEO 순위 확인
- [ ] 트래픽 분석
- [ ] 사용자 행동 패턴 분석
- [ ] 개선 사항 도출

## ✅ 최종 확인

- [ ] 모든 체크리스트 항목 완료
- [ ] 팀원/고객 최종 승인
- [ ] 백업 완료
- [ ] 롤백 계획 수립

---

**런칭 준비 완료! 🚀**

문의: hello@woohwahae.kr
