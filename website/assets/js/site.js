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

    /* ─── 현재 섹션 Nav active ─── */
    var path = window.location.pathname;
    document.querySelectorAll('.site-nav .nav-links a').forEach(function (a) {
        var href = a.getAttribute('href') || '';
        /* 이미 active면 건드리지 않음 */
        if (a.classList.contains('active')) return;
        if (href && path.indexOf(href) !== -1 && href !== '/') {
            a.classList.add('active');
        }
    });

    /* ─── 홈 브랜드 breath (home-brand) ─── */
    var brand = document.querySelector('.home-brand[data-breath]');
    if (brand) {
        var breathMs = {
            'time-dawn': 4800,
            'time-day': 3800,
            'time-evening': 4200,
            'time-night': 5200
        };
        var dur = breathMs[cls] || 3800;
        brand.style.animation =
            'hb-fade 1.2s cubic-bezier(0.16,1,0.3,1) 0.1s both,' +
            'home-breathe ' + dur + 'ms ease-in-out 1.4s infinite';
    }

})();
