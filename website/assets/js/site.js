/*!
 * WOOHWAHAE — site.js
 * Nav 토글 · Scroll reveal · Breath 시간대 변조
 */

(function () {
    'use strict';

    /* ─── Breath 시간대 변조 ─── */
    var h = new Date().getHours();
    var cls = h < 6 ? 'time-dawn'
        : h < 18 ? 'time-day'
            : h < 22 ? 'time-evening'
                : 'time-night';
    document.documentElement.className = cls;

    /* ─── Nav 토글 ─── */
    var toggle = document.getElementById('nav-toggle');
    var overlay = document.getElementById('nav-overlay');

    if (toggle && overlay) {
        toggle.addEventListener('click', function () {
            var isOpen = overlay.classList.contains('open');
            overlay.classList.toggle('open', !isOpen);
            toggle.setAttribute('aria-expanded', String(!isOpen));
            document.body.style.overflow = isOpen ? '' : 'hidden';
        });

        /* 오버레이 링크 클릭 시 닫기 */
        overlay.querySelectorAll('a').forEach(function (a) {
            a.addEventListener('click', function () {
                overlay.classList.remove('open');
                toggle.setAttribute('aria-expanded', 'false');
                document.body.style.overflow = '';
            });
        });

        /* ESC 닫기 */
        document.addEventListener('keydown', function (e) {
            if (e.key === 'Escape' && overlay.classList.contains('open')) {
                overlay.classList.remove('open');
                toggle.setAttribute('aria-expanded', 'false');
                document.body.style.overflow = '';
                toggle.focus();
            }
        });
    }

    /* ─── Scroll Reveal ─── */
    var revealEls = document.querySelectorAll('[data-reveal]');
    if (revealEls.length && 'IntersectionObserver' in window) {
        var observer = new IntersectionObserver(function (entries) {
            entries.forEach(function (entry) {
                if (entry.isIntersecting) {
                    entry.target.classList.add('is-visible');
                    observer.unobserve(entry.target);
                }
            });
        }, { threshold: 0.12, rootMargin: '0px 0px -32px 0px' });

        revealEls.forEach(function (el, i) {
            /* stagger: 카드류만 60ms 간격 */
            if (el.classList.contains('archive-card') ||
                el.classList.contains('practice-card') ||
                el.classList.contains('ws-item')) {
                el.style.transitionDelay = (i % 9 * 0.06) + 's';
            }
            observer.observe(el);
        });
    } else {
        /* IO 미지원 fallback */
        revealEls.forEach(function (el) { el.classList.add('is-visible'); });
    }

    /* ─── 내비게이션 숫자 태그 인터랙션 ─── */
    document.querySelectorAll('.nav-links a, .nav-overlay a').forEach(function (a) {
        var tag = a.querySelector('.numeric-tag');
        if (tag) {
            a.addEventListener('mouseenter', function () { tag.classList.add('numeric-tag--active'); });
            a.addEventListener('mouseleave', function () { 
                if (!a.classList.contains('active')) {
                    tag.classList.remove('numeric-tag--active');
                }
            });
        }
    });

    /* ─── Deconstructed Title Shuffle ─── */
    var deconTitle = document.querySelector('.decon-title');
    if (deconTitle) {
        setTimeout(function() {
            deconTitle.style.transition = 'opacity 1s var(--ease)';
            deconTitle.style.opacity = '1';
        }, 100);
    }

})();
