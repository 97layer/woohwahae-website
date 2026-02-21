/* editor-works-scroll.js — GSAP ScrollTrigger 스크롤 진입 stagger
 * 의존성: gsap@3.12.5, ScrollTrigger (CDN, defer 로드)
 * 폴백: GSAP 없으면 즉시 종료. CSS 틸트 (인라인) 단독 작동.
 */
(function () {
  'use strict';

  // GSAP CDN 로드 실패 시 안전 종료
  if (typeof gsap === 'undefined' || typeof ScrollTrigger === 'undefined') return;

  var isMobile = window.matchMedia('(max-width: 768px)').matches;
  var prefersReduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  gsap.registerPlugin(ScrollTrigger);

  var cards = gsap.utils.toArray('[data-works-card]');
  if (!cards.length) return;

  // 초기 숨김 set (FOUC 방지)
  gsap.set(cards, {
    opacity: 0,
    rotationX: prefersReduced ? 0 : 14,
    y: prefersReduced ? 0 : 36,
    transformOrigin: 'top center',
    transformPerspective: 900,
  });

  // 스크롤 진입 stagger 애니메이션
  gsap.to(cards, {
    scrollTrigger: {
      trigger: '#works-scroll-zone',
      start: 'top 78%',
    },
    opacity: 1,
    rotationX: 0,
    y: 0,
    duration: prefersReduced ? 0.3 : 0.85,
    ease: 'expo.out',
    stagger: {
      amount: prefersReduced ? 0.15 : 1.0,
      from: 'start',
      grid: [Math.ceil(cards.length / 2), 2],
    },
  });

  // 섹션 헤더 미세 패럴랙스 (데스크탑 + 모션 허용 시만)
  if (!isMobile && !prefersReduced) {
    gsap.to('.editor-works-header', {
      scrollTrigger: {
        trigger: '#works-scroll-zone',
        start: 'top bottom',
        end: 'top top',
        scrub: 1.5,
      },
      y: -18,
      ease: 'none',
    });
  }

})();
