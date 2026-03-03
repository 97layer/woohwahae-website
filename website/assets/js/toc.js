/**
 * LAYER OS — Thought Traces (Universal Mobile TOC)
 * 사유의 궤적: 모바일 환경에서의 최소주의적 네비게이션 엔진
 */

(function () {
    'use strict';

    function initThoughtTraces() {
        const sections = document.querySelectorAll('section[id], article[id]');
        if (sections.length < 2) return; // 섹션이 적으면 TOC 생성 안 함

        // 1. UI 생성
        const tocContainer = document.createElement('div');
        tocContainer.id = 'thought-traces';
        tocContainer.className = 'thought-traces';
        tocContainer.setAttribute('aria-hidden', 'true');

        const tocBar = document.createElement('div');
        tocBar.className = 'thought-traces__bar';

        const tocLabel = document.createElement('span');
        tocLabel.className = 'thought-traces__label';
        tocLabel.textContent = 'Traces';

        const tocProgress = document.createElement('div');
        tocProgress.className = 'thought-traces__progress';

        tocBar.appendChild(tocLabel);
        tocBar.appendChild(tocProgress);
        tocContainer.appendChild(tocBar);
        document.body.appendChild(tocContainer);

        // 2. Intersection Observer 설정
        const options = {
            rootMargin: '-20% 0px -70% 0px',
            threshold: 0
        };

        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    updateTOC(entry.target);
                }
            });
        }, options);

        sections.forEach(section => observer.observe(section));

        // 3. 상태 업데이트 및 햅틱 피드백
        function updateTOC(activeSection) {
            const label = activeSection.querySelector('.section-label, .about-section__label, h2');
            let text = 'Exploring';

            if (label) {
                // 불필요한 공백 및 하위 태그 텍스트 정제
                text = label.innerText.split('\n')[0].replace(/[0-9.]/g, '').trim();
                if (!text) text = label.innerText.split('\n')[0].trim();
            }

            tocLabel.textContent = text;

            // 햅틱 피드백 (진동)
            if ('vibrate' in navigator) {
                navigator.vibrate(10); // 미세한 진동 (10ms)
            }

            // 시각적 피드백 (애니메이션 트리거을 위한 클래스)
            tocBar.classList.remove('is-updating');
            void tocBar.offsetWidth; // 리플로우
            tocBar.classList.add('is-updating');
        }

        // 4. 인트로 제어 (헤더 영역에서는 숨김)
        const intro = document.querySelector('.home-hero, .about-intro, .archive-intro');
        if (intro) {
            const introObserver = new IntersectionObserver((entries) => {
                if (entries[0].isIntersecting) {
                    tocContainer.classList.remove('is-visible');
                } else {
                    tocContainer.classList.add('is-visible');
                }
            }, { threshold: 0.1 });
            introObserver.observe(intro);
        } else {
            tocContainer.classList.add('is-visible');
        }
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initThoughtTraces);
    } else {
        initThoughtTraces();
    }

})();
