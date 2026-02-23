# LAYER OS 정식 도메인 배포 계획

> **작성일**: 2026-02-18
> **목적**: woohwahae.kr 정식 도메인 연결 시 비용 및 배포 전략
> **현재 상태**: TryCloudflare 임시 URL 사용 중

---

## 🌐 현재 상황

### **임시 배포 (현재)**
```
https://busy-title-legislative-buck.trycloudflare.com/archive/
└─ Cloudflare TryCloudflare (무료)
   └─ GCP VM (136.109.201.201)
      └─ WOOHWAHAE Website
```

**문제점**:
- ❌ URL이 랜덤 생성 (재시작 시 변경)
- ❌ 브랜딩 불가능 (busy-title-legislative-buck?)
- ❌ 인증 없음 (Public 노출)
- ❌ SSL 인증서 관리 불가

---

## 💰 Cloudflare 가격 정책 (2026년 기준)

### **Cloudflare Plans**

| Plan | 월 비용 | 주요 기능 | WOOHWAHAE 적합성 |
|---|---|---|---|
| **Free** | $0 | DNS, CDN, SSL, DDoS 방어 (무제한) | ✅ **충분함** |
| **Pro** | $20 | + Image Optimization, WAF | △ 선택적 |
| **Business** | $200 | + PCI 준수, 고급 분석 | ❌ 과함 |
| **Enterprise** | 협의 | + 전담 지원, SLA 100% | ❌ 불필요 |

---

## ✅ Cloudflare Free Plan으로 충분한 이유

### **1. DNS + CDN (무료)**

**제공 기능**:
- ✅ **무제한 대역폭** (트래픽 제한 없음)
- ✅ **글로벌 CDN** (전 세계 200+ 데이터센터)
- ✅ **자동 SSL/TLS** (Let's Encrypt 무료 인증서)
- ✅ **DDoS 방어** (무료, 자동)
- ✅ **DNS 관리** (A, CNAME, MX 등 무제한)

**WOOHWAHAE 사용 시나리오**:
```
woohwahae.kr (도메인)
    ↓ Cloudflare DNS (무료)
    ↓ Cloudflare CDN (무료)
    ↓ SSL/TLS (무료)
GCP VM (136.109.201.201)
```

**비용**: **$0/월** ✅

---

### **2. Cloudflare Tunnel (무료)**

**현재 사용 중**: TryCloudflare (임시)
**정식 전환**: Named Tunnel (무료)

```bash
# 무료 Named Tunnel 생성
cloudflared tunnel create woohwahae-production

# 도메인 연결
cloudflared tunnel route dns woohwahae-production woohwahae.kr

# 실행
cloudflared tunnel run woohwahae-production
```

**장점**:
- ✅ **고정 URL** (`woohwahae.kr`)
- ✅ **자동 SSL** (Cloudflare가 관리)
- ✅ **방화벽 불필요** (Outbound 연결만 사용)
- ✅ **DDoS 방어** 자동

**비용**: **$0/월** ✅

---

### **3. Cloudflare Access (무료, 제한적)**

**Admin Panel 보호용**:
```
woohwahae.kr/admin → Cloudflare Access 인증
woohwahae.kr/archive → Public (인증 없음)
```

**Free Tier 제한**:
- ✅ **50 users까지 무료**
- ✅ Google/GitHub OAuth 인증
- ✅ Email OTP 인증

**WOOHWAHAE는 1명만 사용** → 완전 무료 ✅

**비용**: **$0/월** ✅

---

## 💵 실제 비용 구조

### **시나리오 1: Cloudflare Free (권장) ✅**

| 항목 | 월 비용 | 비고 |
|---|---|---|
| **도메인 등록** | ~$1-2 | woohwahae.kr (연간 $12~24) |
| **Cloudflare Free** | $0 | DNS, CDN, SSL, Tunnel 모두 무료 |
| **GCP VM** | ~$10-20 | e2-micro (24/7 운영) |
| **총 비용** | **$11-22/월** | **Cloudflare는 무료** |

**Cloudflare로 인한 추가 비용: $0** ✅

---

### **시나리오 2: Cloudflare Pro (선택적)**

**Pro Plan이 필요한 경우**:
- 이미지 최적화 자동화 (Shopify처럼)
- WAF (Web Application Firewall)
- 고급 DDoS 방어

| 항목 | 월 비용 |
|---|---|
| Cloudflare Pro | $20 |
| 기타 (도메인+VM) | $11-22 |
| **총 비용** | **$31-42/월** |

**WOOHWAHAE 필요성**: ❌ **불필요**
- 이미지는 로컬 최적화 후 업로드 가능
- DDoS는 Free Plan으로 충분
- 트래픽 초기 단계 (Pro 기능 과함)

---

## 🚀 정식 도메인 배포 계획

### **Phase 1: 도메인 구매 (1일)**

**1-1. 도메인 구매**
```
woohwahae.kr (권장)
또는
woohwahae.com (국제)
```

**구매처 옵션**:
- **가비아** (국내, 한국어 지원) — .kr 도메인 $12-15/년
- **Cloudflare Registrar** (최저가) — .com $9.77/년
- **Namecheap** (해외) — .com $10-12/년

**권장**: **Cloudflare Registrar** (비용 투명, 통합 관리)

---

### **Phase 2: Cloudflare 설정 (1시간)**

**2-1. Cloudflare 계정 생성**
```
1. https://dash.cloudflare.com 가입 (무료)
2. "Add Site" → woohwahae.kr 입력
3. Plan 선택 → Free ($0) 선택
```

**2-2. DNS 설정**
```
Type  Name             Content               Proxy
A     @                136.109.201.201       Proxied (CDN)
A     www              136.109.201.201       Proxied
CNAME archive          @                     Proxied
CNAME shop             @                     Proxied
```

**2-3. Nameserver 변경**
```
도메인 등록업체에서 Nameserver 변경:
→ Cloudflare가 제공하는 NS (예: ns1.cloudflare.com)
```

**소요 시간**: 1-24시간 (DNS 전파)

---

### **Phase 3: Cloudflare Tunnel 전환 (30분)**

**3-1. Named Tunnel 생성**
```bash
# GCP VM에서 실행
cd ~/LAYER OS

# cloudflared 설치 (없으면)
wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64
sudo mv cloudflared-linux-amd64 /usr/local/bin/cloudflared
sudo chmod +x /usr/local/bin/cloudflared

# Tunnel 생성
cloudflared tunnel login  # 브라우저 인증
cloudflared tunnel create woohwahae-prod
# → Tunnel ID 발급 (예: abc123-def456)
```

**3-2. DNS 연결**
```bash
# Cloudflare DNS에 자동 등록
cloudflared tunnel route dns woohwahae-prod woohwahae.kr
cloudflared tunnel route dns woohwahae-prod www.woohwahae.kr
```

**3-3. Config 파일 작성**
```yaml
# ~/.cloudflared/config.yml
tunnel: woohwahae-prod
credentials-file: /home/skyto5339_gmail_com/.cloudflared/abc123-def456.json

ingress:
  # Public Website
  - hostname: woohwahae.kr
    service: http://localhost:8080
  - hostname: www.woohwahae.kr
    service: http://localhost:8080

  # Admin Panel (Cloudflare Access 보호)
  - hostname: admin.woohwahae.kr
    service: http://localhost:5001

  # Catch-all
  - service: http_status:404
```

**3-4. Systemd 서비스 등록**
```bash
sudo cloudflared service install
sudo systemctl enable cloudflared
sudo systemctl start cloudflared
```

**결과**:
```
https://woohwahae.kr          → Website (Public)
https://www.woohwahae.kr      → Website (Public)
https://admin.woohwahae.kr    → Admin Panel (인증 필요)
```

---

### **Phase 4: Cloudflare Access 설정 (30분)**

**4-1. Access Application 생성**
```
Cloudflare Dashboard
→ Zero Trust
→ Access
→ Applications
→ Add an Application

Name: WOOHWAHAE Admin Panel
Domain: admin.woohwahae.kr
```

**4-2. Access Policy**
```
Policy Name: Admin Only
Action: Allow
Include:
  - Emails: your@email.com (순호 이메일)
```

**결과**:
- `woohwahae.kr` → 인증 없음 (Public)
- `admin.woohwahae.kr` → 이메일 인증 필요

**비용**: **$0** (1 user는 무료)

---

### **Phase 5: SSL/TLS 자동 활성화 (자동)**

**Cloudflare가 자동 처리**:
```
1. Let's Encrypt 인증서 발급 (자동)
2. HTTPS 리다이렉트 활성화
3. 인증서 자동 갱신 (90일마다)
```

**사용자 액션**: **없음** (자동)

---

## 📊 최종 비용 요약

### **초기 비용 (1회)**
- 도메인 등록: $12-24 (연간)
- 설정 시간: 2-3시간 (무료, 본인 작업)

### **월 운영 비용**
| 항목 | 비용 | 비고 |
|---|---|---|
| 도메인 갱신 | $1-2/월 | 연간 $12-24 ÷ 12 |
| GCP VM | $10-20/월 | e2-micro |
| **Cloudflare** | **$0** | **DNS+CDN+SSL+Tunnel 모두 무료** |
| **총계** | **$11-22/월** | |

### **트래픽 증가 시**
- 월 10만 방문자: Cloudflare Free **$0** (무제한)
- 월 100만 방문자: Cloudflare Free **$0** (여전히 무료)
- 월 1000만 방문자: Cloudflare Free **$0** (계속 무료)

**Cloudflare는 대역폭 제한 없음** ✅

---

## 🆚 대안 비교

### **Option 1: Cloudflare (권장) ✅**
- DNS + CDN + SSL: **무료**
- Tunnel: **무료**
- Access 인증: **무료** (50 users까지)
- 총 비용: **$0/월**

### **Option 2: 직접 구축**
- GCP Load Balancer: **$18/월**
- SSL 인증서 (Let's Encrypt): 무료 (수동 갱신)
- CDN 없음: 속도 느림
- 총 비용: **$18/월** + 관리 비용

### **Option 3: AWS CloudFront**
- CloudFront CDN: **$1-5/월** (트래픽 기반)
- Route 53 DNS: **$0.5/월**
- SSL 인증서: 무료 (ACM)
- 총 비용: **$1.5-5.5/월**

**결론**: Cloudflare Free가 **최적** ✅

---

## ⚠️ Cloudflare Free Plan 제약사항

### **제한되는 기능** (무료 플랜)
1. ❌ **이미지 최적화** 자동 (WebP 변환 등)
   - **해결**: 로컬에서 이미지 최적화 후 업로드
2. ❌ **WAF (Web Application Firewall)**
   - **해결**: Nginx 레벨에서 기본 보안 규칙 설정
3. ❌ **고급 분석** (방문자 상세 로그)
   - **해결**: Google Analytics 사용
4. ❌ **Load Balancing**
   - **영향 없음**: 단일 서버 운영

### **제한되지 않는 기능** (무료지만 무제한)
- ✅ **대역폭**: 무제한
- ✅ **요청 수**: 무제한
- ✅ **SSL/TLS**: 무제한
- ✅ **DDoS 방어**: 무제한
- ✅ **DNS 쿼리**: 무제한

**WOOHWAHAE 영향**: ✅ **없음** (Free Plan으로 충분)

---

## 🎯 실행 체크리스트

### **도메인 연결 전 (준비)**
- [ ] 도메인 구매 완료 (woohwahae.kr)
- [ ] Cloudflare 계정 생성
- [ ] GCP VM에 cloudflared 설치
- [ ] Website 파일 최종 점검

### **도메인 연결 (1일차)**
- [ ] Cloudflare에 도메인 추가
- [ ] DNS 레코드 설정 (A, CNAME)
- [ ] Nameserver 변경 (도메인 등록업체)
- [ ] DNS 전파 대기 (1-24시간)

### **Tunnel 설정 (2일차)**
- [ ] Named Tunnel 생성
- [ ] DNS 연결 (`cloudflared tunnel route dns`)
- [ ] Config 파일 작성
- [ ] Systemd 서비스 등록
- [ ] Tunnel 실행 확인

### **보안 설정 (2일차)**
- [ ] Cloudflare Access 설정
- [ ] Admin Panel 인증 테스트
- [ ] SSL/TLS 모드 "Full (strict)" 설정
- [ ] HTTPS 강제 리다이렉트 활성화

### **최종 검증 (2일차)**
- [ ] `https://woohwahae.kr` 접속 확인
- [ ] SSL 인증서 유효성 확인
- [ ] 모든 페이지 로딩 테스트
- [ ] Admin Panel 인증 확인
- [ ] 모바일 접속 테스트

---

## 💡 추가 권장사항

### **1. 이메일 설정 (선택적)**

**Cloudflare Email Routing (무료)**:
```
hello@woohwahae.kr → your-gmail@gmail.com
```

**설정**:
1. Cloudflare Dashboard → Email Routing
2. MX 레코드 자동 생성
3. 전달 주소 설정

**비용**: **$0** ✅

---

### **2. 페이지 규칙 (무료)**

**Cloudflare Page Rules (3개 무료)**:
```
Rule 1: woohwahae.kr/admin/*
  → Cache Level: Bypass

Rule 2: woohwahae.kr/assets/*
  → Cache Level: Cache Everything
  → Edge Cache TTL: 1 month

Rule 3: www.woohwahae.kr/*
  → Forwarding URL (301): https://woohwahae.kr/$1
```

**비용**: **$0** ✅

---

### **3. 분석 도구 연동**

**Google Analytics 4** (무료):
- 방문자 추적
- 페이지 조회수
- 전환율 분석

**Cloudflare Web Analytics** (무료):
- 개인정보 보호 중심
- 쿠키 없음
- 실시간 트래픽

**둘 다 사용 권장** → 비용: **$0** ✅

---

## 🏆 결론

### **질문: "정식 도메인 달고 웹에 올리면 클라우드플레어는 유료 아닌가?"**

### **답변: 아니오, 무료입니다. ✅**

**근거**:
1. Cloudflare Free Plan이 제공하는 기능:
   - DNS, CDN, SSL, DDoS 방어 **모두 무료**
   - Cloudflare Tunnel **무료**
   - Cloudflare Access (50 users) **무료**

2. WOOHWAHAE 사용 케이스:
   - 트래픽: 초기 단계 (Free Plan 충분)
   - 사용자: 1명 (Access 무료 범위)
   - 기능: Free Plan으로 100% 커버

3. 실제 비용 구조:
   - 도메인: $1-2/월
   - GCP VM: $10-20/월
   - **Cloudflare: $0/월**
   - **총: $11-22/월**

**Cloudflare로 인한 추가 비용: $0** ✅

---

**최종 권고**:
1. ✅ Cloudflare Free Plan 사용
2. ✅ Named Tunnel로 전환 (TryCloudflare 중단)
3. ✅ Cloudflare Access로 Admin 보호
4. ✅ 트래픽 증가해도 Free Plan 유지 가능

**Pro Plan ($20/월) 필요 시점**:
- 월 100만+ 방문자 & 이미지 자동 최적화 필요
- 고급 WAF 규칙 필요
- PCI DSS 준수 필요 (결제 시스템)

**현재 WOOHWAHAE**: ✅ **Free Plan으로 충분**

---

**작성**: 2026-02-18
**작성자**: System Infrastructure Architect
**유효기간**: 2026년 내 (가격 정책 변경 시 재검토)
