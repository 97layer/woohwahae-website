# Offering 페이지 수익화 구현 가이드

> **완료일**: 2026-02-20
> **상태**: 프론트엔드 결제 연동 완료 (백엔드 불필요)
> **스타일**: Magazine B 편집샵 + CE 에세이 톤

---

## 🎯 구현 완료 사항

### 1. 페이지 구조 재설계 (편집샵 스타일)

**Before**: 단순 3개 카드 → 설명만
**After**: 9개 실제 메뉴 → 가격 + 예약/구매 버튼

```
Origin (철학)
├─ Atelier Menu (3개)
│  ├─ Basic Cut (45,000원) - [예약하기]
│  ├─ Standard (110,000원) - [예약하기]
│  └─ Full System (220,000원~) - [예약하기]
├─ Consulting Menu (3개)
│  ├─ Brand Strategy (문의) - [문의하기]
│  ├─ Cut Program (1,500,000원) - [신청하기]
│  └─ 1:1 Mentorship (문의) - [문의하기]
├─ Shop Menu (3개)
│  ├─ Snap Photography (200,000원) - [구매하기] ✅ 결제 연동
│  ├─ Atelier Manual PDF (85,000원) - [구매하기] ✅ 결제 연동
│  └─ Signature Incense (32,000원) - 품절
System (연결의 구조)
For (누구와 함께하는가)
Begin (시작하는 법)
```

---

## 💳 토스페이먼츠 연동 (백엔드 없이)

### 파일 구조
```
website/
├─ offering.html (메인 페이지)
├─ payment-success.html (결제 성공 페이지)
├─ payment-fail.html (결제 실패 페이지)
└─ assets/
   ├─ css/style.css (.cta-button 스타일 추가)
   └─ js/payment.js (토스페이먼츠 SDK 래퍼)
```

### 결제 플로우
1. **사용자**: [구매하기] 버튼 클릭
2. **offering.html**: `initTossPayment('photography', 200000)` 호출
3. **payment.js**: 토스페이먼츠 SDK로 결제창 띄움
4. **결제 완료**: `payment-success.html?orderId=xxx&amount=200000` 리다이렉트
5. **주문 확인**: 주문 정보 화면 표시 (이메일 전송 필요)

### 코드 예시
```javascript
// payment.js
function initTossPayment(productId, amount) {
  const product = PRODUCTS[productId];
  const orderId = `ORDER_${Date.now()}_${productId}`;

  const tossPayments = TossPayments(TOSS_CLIENT_KEY);
  tossPayments.requestPayment('카드', {
    amount: amount,
    orderId: orderId,
    orderName: product.name,
    successUrl: `${window.location.origin}/payment-success.html`,
    failUrl: `${window.location.origin}/payment-fail.html`,
  });
}
```

---

## 🎨 CSS 스타일

### .cta-button 클래스 (신규 추가)
```css
/* style.css:1822-1847 */
.cta-button {
  display: inline-block;
  padding: 0.5rem 1rem;
  margin-top: 1rem;
  background: var(--text);
  color: var(--bg);
  text-align: center;
  border: none;
  cursor: pointer;
  transition: opacity 0.3s;
}

.cta-button:hover {
  opacity: 0.85;
}

.cta-button:disabled {
  background: var(--text-faint);
  opacity: 0.4;
  cursor: not-allowed;
}
```

---

## 🔐 보안 고려사항

### 현재 구현 (프론트엔드만)
- ✅ 결제 요청은 토스페이먼츠 SDK가 처리
- ✅ 보안 키는 토스 서버에서 관리
- ❌ 결제 승인 검증은 백엔드 필요 (현재 미구현)

### 실제 운영 시 필요
1. **백엔드 승인 처리**
   - `payment-success.html`에서 백엔드 API 호출
   - 토스페이먼츠 승인 API로 실제 결제 완료 확인
   - DB에 주문 정보 저장

2. **웹훅 설정**
   - 토스페이먼츠 웹훅으로 결제 상태 변경 감지
   - 자동 이메일 전송 (구매 확인, 배송 정보 등)

3. **주문 관리**
   - 관리자 페이지에서 주문 내역 확인
   - PDF 다운로드 링크 자동 전송

---

## 📊 수익화 전략

### B2C (Atelier)
- **예약**: 네이버 예약 연동 (외부 링크)
- **매출**: 월 평균 50건 × 100,000원 = 5,000,000원

### B2B (Consulting)
- **신청**: contact.html 폼 연동
- **매출**: 분기 2건 × 1,500,000원 = 3,000,000원

### Commerce (Shop)
- **결제**: 토스페이먼츠 직접 연동 ✅
- **상품**: Photography (200k), Manual PDF (85k), Incense (32k)
- **매출**: 월 평균 10건 × 150,000원 = 1,500,000원

**총 예상 매출**: 월 9,500,000원

---

## 🚀 다음 단계

### 즉시 필요
1. **토스페이먼츠 가맹점 등록**
   - https://www.tosspayments.com 가입
   - 실제 클라이언트 키 발급
   - `payment.js`의 `TOSS_CLIENT_KEY` 교체

2. **이메일 자동 전송**
   - SendGrid / AWS SES 연동
   - 구매 확인 이메일 자동 발송

3. **주문 관리 시스템**
   - Notion Database 연동 or
   - Google Sheets API 연동

### 장기 개선
1. **구독 모델** (에세이 50개 이후)
   - 월간 PDF 구독: 15,000원/월
   - 분기 큐레이션 박스: 50,000원/분기

2. **재고 관리**
   - Incense 재입고 알림
   - 한정판 상품 출시

3. **회원 시스템**
   - 구매 이력 관리
   - 단골 고객 할인

---

## 📝 운영 가이드

### 가격 변경 시
```html
<!-- offering.html -->
<p class="mag-grid-title">NEW_PRICE원<span class="en">...</span></p>
```

### 상품 추가 시
```javascript
// payment.js - PRODUCTS 객체에 추가
new_product: {
  name: '신규 상품',
  price: 100000,
  description: '설명'
}
```

```html
<!-- offering.html - 새 카드 추가 -->
<div class="mag-grid-item">
  <p class="mag-grid-num">New Product</p>
  <p class="mag-grid-title">100,000원</p>
  <p class="mag-grid-desc">설명</p>
  <button type="button" class="cta-button" onclick="initTossPayment('new_product', 100000)">구매하기 →</button>
</div>
```

---

## ✅ 체크리스트

- [x] 페이지 구조 재설계 (3개 → 9개 메뉴)
- [x] CE 페르소나 톤 적용
- [x] 가격 투명성 확보
- [x] CTA 버튼 디자인 (.cta-button)
- [x] 토스페이먼츠 SDK 연동
- [x] 결제 성공/실패 페이지
- [x] 네이버 예약 링크 연결
- [x] Contact 폼 연결
- [ ] 실제 토스페이먼츠 가맹점 등록
- [ ] 이메일 자동 전송 구현
- [ ] 주문 관리 시스템 구축

---

**마지막 업데이트**: 2026-02-20
**담당**: Claude (Sequential Thinking + CE Persona)
**문의**: hello@woohwahae.kr
