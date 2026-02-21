/**
 * WOOHWAHAE 토스페이먼츠 결제 모듈
 * 백엔드 없이 프론트엔드에서 직접 결제 처리
 */

// 토스페이먼츠 클라이언트 키 (실제 운영 시 환경변수로 관리)
const TOSS_CLIENT_KEY = 'test_ck_D5GePWvyJnrK0W0k6q8gLzN97Eoq'; // 테스트 키

// 상품 정보
const PRODUCTS = {
  photography: {
    name: 'Snap Photography',
    price: 200000,
    description: '과장 없이 당신의 표정과 공기를 기록. 원본 및 보정본 10매 제공.'
  },
  manual_pdf: {
    name: 'Atelier Manual PDF',
    price: 85000,
    description: '1인 미용실 생존 가이드라인 전자책.'
  },
  incense: {
    name: 'Signature Incense',
    price: 32000,
    description: '우화해 공간의 향. 샌달우드와 베티버 베이스.'
  },
  slow_object_001: {
    name: 'Slow Object Report #001 — 가위',
    price: 28000,
    description: '도구의 본질에 대한 탐구. Magazine B 스타일 오브제 해부 전자책.'
  }
};

/**
 * 토스페이먼츠 결제 시작
 * @param {string} productId - 상품 ID
 * @param {number} amount - 결제 금액
 */
function initTossPayment(productId, amount) {
  const product = PRODUCTS[productId];

  if (!product) {
    alert('상품 정보를 찾을 수 없습니다.');
    return;
  }

  // 주문 ID 생성 (타임스탬프 기반)
  const orderId = `ORDER_${Date.now()}_${productId}`;

  // 결제 요청 데이터
  const paymentData = {
    amount: amount,
    orderId: orderId,
    orderName: product.name,
    customerName: '구매자', // 실제로는 폼에서 입력받아야 함
    successUrl: `${window.location.origin}/payment-success.html`,
    failUrl: `${window.location.origin}/payment-fail.html`,
  };

  // 토스페이먼츠 SDK 로드 후 결제창 호출
  if (typeof TossPayments === 'undefined') {
    alert('결제 모듈 로딩 중입니다. 잠시 후 다시 시도해주세요.');
    return;
  }

  const tossPayments = TossPayments(TOSS_CLIENT_KEY);

  tossPayments.requestPayment('카드', paymentData)
    .catch(function (error) {
      if (error.code === 'USER_CANCEL') {
        console.log('결제 취소');
      } else {
        console.error('결제 오류:', error);
        alert('결제 중 오류가 발생했습니다.');
      }
    });
}

/**
 * 결제 성공 처리 (success.html에서 호출)
 */
function handlePaymentSuccess() {
  const urlParams = new URLSearchParams(window.location.search);
  const orderId = urlParams.get('orderId');
  const amount = urlParams.get('amount');
  const paymentKey = urlParams.get('paymentKey');

  // 실제로는 백엔드로 전송해 승인 처리해야 하지만,
  // 백엔드 없이 프론트엔드만으로 처리하는 경우 제한적
  console.log('결제 성공:', { orderId, amount, paymentKey });

  // 상품 ID 추출 (orderId 형식: ORDER_timestamp_productId)
  const productId = orderId ? orderId.split('_')[2] : null;

  // PDF 다운로드 링크 제공
  const downloadInfo = document.getElementById('download-info');
  if (downloadInfo) {
    let downloadHTML = '';

    if (productId === 'manual_pdf') {
      downloadHTML = `
        <div style="margin-top: 2rem; padding: 1.5rem; background: #f5f5f5; border-left: 3px solid #2C2C2C;">
          <strong style="font-size: 1.1rem;">📥 다운로드</strong><br><br>
          <a href="../products/atelier-manual.html" target="_blank" style="display: inline-block; padding: 0.75rem 1.5rem; background: #2C2C2C; color: #FAFAF7; text-decoration: none; font-family: monospace; font-size: 0.9rem;">
            Atelier Manual 보기 →
          </a><br><br>
          <small style="color: #5A5A5A; line-height: 1.6;">
            브라우저에서 <strong>Cmd/Ctrl + P</strong>를 눌러<br>
            "PDF로 저장"을 선택하세요.
          </small>
        </div>
      `;
    } else if (productId === 'slow_object_001') {
      downloadHTML = `
        <div style="margin-top: 2rem; padding: 1.5rem; background: #f5f5f5; border-left: 3px solid #2C2C2C;">
          <strong style="font-size: 1.1rem;">📥 다운로드</strong><br><br>
          <a href="../products/slow-object-001-scissors.html" target="_blank" style="display: inline-block; padding: 0.75rem 1.5rem; background: #2C2C2C; color: #FAFAF7; text-decoration: none; font-family: monospace; font-size: 0.9rem;">
            Slow Object Report #001 보기 →
          </a><br><br>
          <small style="color: #5A5A5A; line-height: 1.6;">
            브라우저에서 <strong>Cmd/Ctrl + P</strong>를 눌러<br>
            "PDF로 저장"을 선택하세요.
          </small>
        </div>
      `;
    }

    downloadInfo.innerHTML = downloadHTML;
  }
}

/**
 * 결제 실패 처리
 */
function handlePaymentFail() {
  const urlParams = new URLSearchParams(window.location.search);
  const code = urlParams.get('code');
  const message = urlParams.get('message');

  console.error('결제 실패:', { code, message });
  alert(`결제에 실패했습니다.\n${message}`);
}
