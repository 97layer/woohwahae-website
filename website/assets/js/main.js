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

  // ─── Smooth Nav Appearance ───
  // 스크롤 시 nav 배경 미세하게 변화
  const nav = document.querySelector('nav');
  if (nav) {
    window.addEventListener('scroll', () => {
      if (window.scrollY > 60) {
        nav.style.backdropFilter = 'blur(8px)';
        nav.style.backgroundColor = 'rgba(250, 250, 247, 0.85)';
      } else {
        nav.style.backdropFilter = '';
        nav.style.backgroundColor = '';
      }
    }, { passive: true });
  }

});
