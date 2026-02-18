/**
 * WOOHWAHAE Newsletter Subscription
 * 간단한 이메일 구독 기능
 */

class Newsletter {
  constructor() {
    this.form = document.getElementById('newsletter-form');
    this.input = document.getElementById('newsletter-email');
    this.message = document.getElementById('newsletter-message');

    if (this.form) {
      this.init();
    }
  }

  init() {
    this.form.addEventListener('submit', (e) => {
      e.preventDefault();
      this.handleSubmit();
    });
  }

  async handleSubmit() {
    const email = this.input.value.trim();

    // 이메일 유효성 검사
    if (!this.validateEmail(email)) {
      this.showMessage('유효한 이메일 주소를 입력해주세요.', 'error');
      return;
    }

    // 로딩 상태
    this.setLoading(true);

    try {
      // 실제 프로덕션에서는 백엔드 API 호출
      // 현재는 로컬 스토리지에 저장
      await this.subscribe(email);

      this.showMessage('구독해주셔서 감사합니다. 곧 소식을 전해드리겠습니다.', 'success');
      this.input.value = '';

      // Google Analytics 이벤트
      if (typeof trackEvent !== 'undefined') {
        trackEvent('newsletter_subscription', {
          email_hash: this.hashEmail(email)
        });
      }
    } catch (error) {
      this.showMessage('구독 처리 중 오류가 발생했습니다. 다시 시도해주세요.', 'error');
      console.error('Newsletter subscription error:', error);
    } finally {
      this.setLoading(false);
    }
  }

  async subscribe(email) {
    // 프로덕션 환경: 백엔드 API 호출
    // await fetch('/api/newsletter/subscribe', {
    //   method: 'POST',
    //   headers: { 'Content-Type': 'application/json' },
    //   body: JSON.stringify({ email })
    // });

    // 개발 환경: 로컬 스토리지에 저장
    return new Promise((resolve) => {
      setTimeout(() => {
        const subscribers = JSON.parse(localStorage.getItem('newsletter_subscribers') || '[]');

        if (subscribers.includes(email)) {
          throw new Error('Already subscribed');
        }

        subscribers.push(email);
        localStorage.setItem('newsletter_subscribers', JSON.stringify(subscribers));
        resolve();
      }, 1000);
    });
  }

  validateEmail(email) {
    const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return re.test(email);
  }

  hashEmail(email) {
    // 간단한 해싱 (프라이버시 보호)
    let hash = 0;
    for (let i = 0; i < email.length; i++) {
      const char = email.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash;
    }
    return hash.toString(36);
  }

  setLoading(loading) {
    const button = this.form.querySelector('button[type="submit"]');
    if (loading) {
      button.disabled = true;
      button.textContent = '처리 중...';
    } else {
      button.disabled = false;
      button.textContent = '구독하기';
    }
  }

  showMessage(text, type) {
    this.message.textContent = text;
    this.message.className = `newsletter-message newsletter-message--${type}`;
    this.message.style.display = 'block';

    setTimeout(() => {
      this.message.style.display = 'none';
    }, 5000);
  }
}

// Initialize
document.addEventListener('DOMContentLoaded', () => {
  new Newsletter();
});
