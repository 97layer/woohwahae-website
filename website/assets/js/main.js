/* =============================================
   WOOHWAHAE — Interactions
   Apple 무드: 부드러운 페이드인, 최소한의 JS
   ============================================= */

document.addEventListener('DOMContentLoaded', () => {

  // ─── Fade-in on Scroll ───
  const observerOptions = {
    threshold: 0.12,
    rootMargin: '0px 0px -40px 0px'
  };

  const observer = new IntersectionObserver((entries) => {
    entries.forEach((entry, i) => {
      if (entry.isIntersecting) {
        // 순차 딜레이 (같은 뷰포트 내 요소들)
        const delay = entry.target.dataset.delay || 0;
        setTimeout(() => {
          entry.target.classList.add('visible');
        }, delay * 100);
        observer.unobserve(entry.target);
      }
    });
  }, observerOptions);

  document.querySelectorAll('.fade-in').forEach((el, i) => {
    // 형제 요소들에 순차 딜레이 부여
    if (!el.dataset.delay) {
      const siblings = el.parentElement?.querySelectorAll('.fade-in');
      if (siblings) {
        let idx = Array.from(siblings).indexOf(el);
        el.dataset.delay = idx;
      }
    }
    observer.observe(el);
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
      .catch(() => {});
  });
}
