/* =============================================
   WOOHWAHAE — Interactions
   Apple 무드: 부드러운 페이드인, 최소한의 JS
   ============================================= */

document.addEventListener('DOMContentLoaded', () => {

  // ─── Fade-in on Scroll ───
  // rootMargin 제거: 뷰포트 내 요소가 즉시 트리거되도록
  const observerOptions = {
    threshold: 0.05
  };

  const observer = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        const delay = entry.target.dataset.delay || 0;
        setTimeout(() => {
          entry.target.classList.add('visible');
        }, delay * 80);
        observer.unobserve(entry.target);
      }
    });
  }, observerOptions);

  document.querySelectorAll('.fade-in').forEach((el, i) => {
    // 형제 요소들에 순차 딜레이 부여
    if (!el.dataset.delay) {
      const siblings = el.parentElement?.querySelectorAll('.fade-in');
      if (siblings) {
        const idx = Array.from(siblings).indexOf(el);
        el.dataset.delay = idx;
      }
    }
    observer.observe(el);
  });

  // 페이지 로드 직후 이미 뷰포트 안에 있는 요소 즉시 visible 처리
  // (h1, hero 등 스크롤 없이 보여야 하는 요소)
  requestAnimationFrame(() => {
    document.querySelectorAll('.fade-in').forEach((el) => {
      const rect = el.getBoundingClientRect();
      if (rect.top < window.innerHeight && rect.bottom > 0) {
        const delay = el.dataset.delay || 0;
        setTimeout(() => {
          el.classList.add('visible');
        }, delay * 80);
      }
    });
  });

  // ─── Nav Active State ───
  const currentPath = window.location.pathname;
  document.querySelectorAll('.nav-links a').forEach(link => {
    const href = link.getAttribute('href');
    if (href === currentPath ||
      (href !== '/' && currentPath.startsWith(href.replace('.html', '')))) {
      link.classList.add('active');
    }
  });

  // ─── Mobile Nav Toggle ───
  const toggle = document.querySelector('.nav-toggle');
  const navLinks = document.querySelector('.nav-links');
  if (toggle && navLinks) {
    toggle.addEventListener('click', () => {
      toggle.classList.toggle('active');
      navLinks.classList.toggle('open');
    });
    // 링크 클릭 시 메뉴 닫기
    navLinks.querySelectorAll('a').forEach(link => {
      link.addEventListener('click', () => {
        toggle.classList.remove('active');
        navLinks.classList.remove('open');
      });
    });
  }

  // ─── Magazine B Style Nav on Scroll ───
  const nav = document.querySelector('nav');
  if (nav) {
    let lastScroll = 0;
    window.addEventListener('scroll', () => {
      const currentScroll = window.scrollY;

      // 스크롤 다운: nav에 scrolled 클래스 추가
      if (currentScroll > 100) {
        nav.classList.add('scrolled');
      } else {
        nav.classList.remove('scrolled');
      }

      lastScroll = currentScroll;
    }, { passive: true });
  }

});

// ─── Service Worker Registration ───
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js')
      .catch(() => { });
  });
}
