# WOOHWAHAE 웹사이트 배포 가이드

## 개요
WOOHWAHAE 웹사이트는 정적 사이트 + Flask 백엔드로 구성되어 있습니다.

## 배포 전 체크리스트

### 1. 환경 변수 설정
```bash
# website/backend/.env 파일 생성
FLASK_SECRET_KEY=your-production-secret-key
FLASK_DEBUG=False
ADMIN_PASSWORD=your-secure-password
```

### 2. Google Analytics 설정
모든 HTML 파일의 `<head>`에 다음 추가:
```html
<meta name="ga-id" content="G-YOUR-MEASUREMENT-ID">
```

### 3. 도메인 설정
다음 파일에서 도메인 업데이트:
- `robots.txt` - Sitemap URL
- `sitemap.xml` - 모든 URL
- `manifest.webmanifest` - start_url

## 배포 방법

### 옵션 1: 정적 호스팅 (Netlify/Vercel/GitHub Pages)

**장점**: 무료, 간편, CDN 자동 제공
**단점**: CMS 백엔드 별도 배포 필요

#### Netlify 배포
```bash
# 1. Netlify CLI 설치
npm install -g netlify-cli

# 2. 로그인
netlify login

# 3. 배포
cd website
netlify deploy --prod
```

#### Vercel 배포
```bash
# 1. Vercel CLI 설치
npm install -g vercel

# 2. 배포
cd website
vercel --prod
```

### 옵션 2: 전체 스택 (VPS/Cloud Server)

**장점**: CMS 백엔드 포함, 완전한 제어
**단점**: 서버 관리 필요, 비용 발생

#### Nginx + Gunicorn 설정

1. **서버 설정**
```bash
# 패키지 설치
sudo apt update
sudo apt install nginx python3-pip python3-venv

# 프로젝트 클론
git clone https://github.com/yourusername/woohwahae.git
cd woohwahae/website
```

2. **백엔드 설정**
```bash
# 가상환경 생성
python3 -m venv venv
source venv/bin/activate

# 의존성 설치
cd backend
pip install -r requirements.txt
pip install gunicorn

# .env 파일 생성
cp .env.example .env
nano .env  # 환경 변수 입력
```

3. **Gunicorn 서비스 생성**
```bash
sudo nano /etc/systemd/system/woohwahae-backend.service
```

```ini
[Unit]
Description=WOOHWAHAE CMS Backend
After=network.target

[Service]
User=www-data
Group=www-data
WorkingDirectory=/path/to/woohwahae/website/backend
Environment="PATH=/path/to/woohwahae/website/venv/bin"
ExecStart=/path/to/woohwahae/website/venv/bin/gunicorn --workers 3 --bind unix:woohwahae.sock app:app

[Install]
WantedBy=multi-user.target
```

```bash
# 서비스 시작
sudo systemctl start woohwahae-backend
sudo systemctl enable woohwahae-backend
```

4. **Nginx 설정**
```bash
sudo nano /etc/nginx/sites-available/woohwahae
```

```nginx
server {
    listen 80;
    server_name woohwahae.kr www.woohwahae.kr;

    # 정적 파일
    root /path/to/woohwahae/website;
    index index.html;

    # Gzip 압축
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;

    # 정적 파일 캐싱
    location ~* \.(jpg|jpeg|png|gif|ico|css|js|svg|woff|woff2)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # 백엔드 API
    location /api {
        include proxy_params;
        proxy_pass http://unix:/path/to/woohwahae/website/backend/woohwahae.sock;
    }

    # HTML 파일
    location / {
        try_files $uri $uri/ =404;
    }

    # 404 페이지
    error_page 404 /404.html;
}
```

```bash
# Nginx 활성화
sudo ln -s /etc/nginx/sites-available/woohwahae /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

5. **SSL 인증서 (Let's Encrypt)**
```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d woohwahae.kr -d www.woohwahae.kr
```

### 옵션 3: Docker 배포

```dockerfile
# Dockerfile
FROM nginx:alpine

# 정적 파일 복사
COPY website /usr/share/nginx/html

# Nginx 설정
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
```

```bash
# 빌드 및 실행
docker build -t woohwahae-web .
docker run -d -p 80:80 woohwahae-web
```

## 배포 후 확인사항

### 1. 기능 테스트
- [ ] 모든 페이지 정상 로드
- [ ] 네비게이션 작동
- [ ] Archive 검색/필터 작동
- [ ] 이미지 로드 확인
- [ ] 모바일 반응형 확인

### 2. SEO 확인
```bash
# robots.txt 확인
curl https://woohwahae.kr/robots.txt

# sitemap.xml 확인
curl https://woohwahae.kr/sitemap.xml

# 구조화된 데이터 테스트
# https://search.google.com/test/rich-results
```

### 3. 성능 테스트
- [Google PageSpeed Insights](https://pagespeed.web.dev/)
- [GTmetrix](https://gtmetrix.com/)
- [WebPageTest](https://www.webpagetest.org/)

### 4. 접근성 테스트
- [WAVE](https://wave.webaim.org/)
- [axe DevTools](https://www.deque.com/axe/devtools/)

## 모니터링 설정

### Google Search Console
1. [Search Console](https://search.google.com/search-console) 접속
2. 속성 추가 → woohwahae.kr
3. Sitemap 제출: https://woohwahae.kr/sitemap.xml

### Google Analytics
1. [Analytics](https://analytics.google.com/) 계정 생성
2. 측정 ID 획득
3. HTML 파일에 메타 태그 추가

## 업데이트 워크플로우

### 콘텐츠 추가
```bash
# 1. 백엔드 API 사용
curl -X POST https://woohwahae.kr/api/archive \
  -H "Content-Type: application/json" \
  -d '{
    "slug": "new-post",
    "title": "새 글",
    "date": "2026.03.01",
    "issue": "Issue 005",
    "preview": "미리보기...",
    "category": "Essay"
  }'

# 2. 또는 수동으로 index.json 수정 후 Git push
```

### 사이트 업데이트
```bash
# Git 업데이트
git pull origin main

# 정적 사이트: 자동 배포 (Netlify/Vercel)
# VPS: Nginx 재시작
sudo systemctl reload nginx

# Docker: 이미지 재빌드
docker build -t woohwahae-web .
docker restart woohwahae
```

## 백업 전략

### 1. 코드 백업
- GitHub/GitLab에 저장소 유지
- 정기적으로 push

### 2. 콘텐츠 백업
```bash
# archive/index.json 백업
cp website/archive/index.json website/archive/index.json.backup

# 전체 백업 (매일 자동화 권장)
tar -czf woohwahae-$(date +%Y%m%d).tar.gz website/
```

### 3. 데이터베이스 백업 (향후 필요 시)
```bash
# PostgreSQL 예시
pg_dump woohwahae > backup.sql
```

## 문제 해결

### 404 오류
- `404.html` 파일 존재 확인
- Nginx 설정에서 error_page 확인

### 느린 로딩
- 이미지 최적화 (WebP 변환)
- CDN 사용 (Cloudflare)
- Gzip 압축 활성화

### CMS 백엔드 오류
```bash
# 로그 확인
sudo journalctl -u woohwahae-backend -f

# 서비스 재시작
sudo systemctl restart woohwahae-backend
```

## 보안 체크리스트
- [ ] HTTPS 활성화 (SSL 인증서)
- [ ] 환경 변수 보호 (.env 파일 git ignore)
- [ ] 강력한 관리자 비밀번호
- [ ] CORS 정책 설정
- [ ] Rate limiting 구현 (향후)
- [ ] 정기적인 의존성 업데이트

## 지원

문의사항: hello@woohwahae.kr
