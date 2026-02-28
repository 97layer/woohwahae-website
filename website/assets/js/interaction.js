/*!
 * WOOHWAHAE — interaction.js (V18 - Blueprint Edition)
 * High-Fidelity Recording Engine (GSAP + Lenis + ScrollTrigger)
 */

(function () {
    'use strict';

    // ─── 00. Lenis Inertial Scroll ───
    const lenis = new Lenis({
        duration: 1.4,
        easing: (t) => Math.min(1, 1.001 - Math.pow(2, -10 * t)),
        smoothWheel: true
    });

    function raf(time) {
        lenis.raf(time);
        requestAnimationFrame(raf);
    }
    requestAnimationFrame(raf);

    lenis.on('scroll', ScrollTrigger.update);
    gsap.ticker.add((time) => {
        lenis.raf(time * 1000);
    });
    gsap.ticker.lagSmoothing(0);

    // ─── 01. Reveal System (V18 Blueprint) ───
    // data-reveal 속성을 가진 요소들을 고요하게 리빌
    const revealElements = gsap.utils.toArray('[data-reveal]');

    revealElements.forEach((el) => {
        gsap.set(el, { opacity: 0, y: 15 });

        ScrollTrigger.create({
            trigger: el,
            start: 'top 92%',
            onEnter: () => {
                gsap.to(el, {
                    opacity: 1,
                    y: 0,
                    duration: 1.2,
                    ease: 'power2.out',
                    overwrite: 'auto'
                });
            },
            onLeaveBack: () => {
                gsap.to(el, {
                    opacity: 0,
                    y: 15,
                    duration: 0.8,
                    ease: 'power2.in'
                });
            }
        });
    });

    // ─── 02. Timeline SVG Animation ───
    const timelineSvg = document.querySelector('.journey-timeline__svg');
    if (timelineSvg) {
        const axis = timelineSvg.querySelector('.tl-axis');
        const nodes = timelineSvg.querySelectorAll('.tl-node-g');

        if (axis) {
            gsap.set(axis, { strokeDasharray: 300, strokeDashoffset: 300 });
            gsap.to(axis, {
                strokeDashoffset: 0,
                duration: 2,
                ease: 'power2.inOut',
                scrollTrigger: {
                    trigger: timelineSvg,
                    start: 'top 80%'
                }
            });
        }

        nodes.forEach((node, i) => {
            gsap.set(node, { opacity: 0, x: -10 });
            gsap.to(node, {
                opacity: 1,
                x: 0,
                duration: 0.8,
                delay: 0.5 + (i * 0.3),
                ease: 'expo.out',
                scrollTrigger: {
                    trigger: timelineSvg,
                    start: 'top 80%'
                }
            });
        });
    }

    // ─── 03. Section Navigation Tracking ───
    const sections = gsap.utils.toArray('.about-section');
    function setActiveNav(id) {
        document.querySelectorAll('.nav-links a').forEach(a => {
            a.classList.toggle('is-active', a.getAttribute('href').includes(id));
        });
    }

    sections.forEach((section) => {
        const id = section.id;
        if (!id) return;
        ScrollTrigger.create({
            trigger: section,
            start: 'top center',
            end: 'bottom center',
            onEnter: () => setActiveNav(id),
            onEnterBack: () => setActiveNav(id),
        });
    });

})();
