/**
 * Google Analytics 설정
 * WOOHWAHAE 웹사이트 트래킹
 */

// Google Analytics 4 설정 (측정 ID 필요)
// 환경 변수 또는 meta 태그에서 GA ID 가져오기
const GA_MEASUREMENT_ID = document.querySelector('meta[name="ga-id"]')?.content || 'G-XXXXXXXXXX';

// 개발 환경에서는 Analytics 비활성화
const isDevelopment = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// Google Analytics 스크립트 로드
(function() {
    // 개발 환경이거나 GA ID가 설정되지 않은 경우 건너뛰기
    if (isDevelopment || GA_MEASUREMENT_ID === 'G-XXXXXXXXXX') {
        console.log('Google Analytics: Disabled (development mode or no GA ID)');
        return;
    }

    // Google Tag (gtag.js) 추가
    const script = document.createElement('script');
    script.async = true;
    script.src = `https://www.googletagmanager.com/gtag/js?id=${GA_MEASUREMENT_ID}`;
    document.head.appendChild(script);

    // gtag 설정
    window.dataLayer = window.dataLayer || [];
    function gtag(){dataLayer.push(arguments);}
    gtag('js', new Date());
    gtag('config', GA_MEASUREMENT_ID, {
        'anonymize_ip': true, // IP 익명화
        'cookie_flags': 'SameSite=None;Secure'
    });

    // 커스텀 이벤트 트래킹 함수
    window.trackEvent = function(eventName, parameters) {
        if (typeof gtag !== 'undefined') {
            gtag('event', eventName, parameters);
        }
    };

    // 페이지뷰 트래킹
    window.trackPageView = function(pageName) {
        if (typeof gtag !== 'undefined') {
            gtag('event', 'page_view', {
                page_title: pageName,
                page_location: window.location.href,
                page_path: window.location.pathname
            });
        }
    };

    // 링크 클릭 트래킹
    document.addEventListener('DOMContentLoaded', function() {
        // 카카오톡 버튼 클릭 트래킹
        document.querySelectorAll('a[href*="kakao"]').forEach(link => {
            link.addEventListener('click', function() {
                trackEvent('contact', {
                    contact_method: 'kakao'
                });
            });
        });

        // 이메일 링크 클릭 트래킹
        document.querySelectorAll('a[href^="mailto"]').forEach(link => {
            link.addEventListener('click', function() {
                trackEvent('contact', {
                    contact_method: 'email'
                });
            });
        });

        // 네이버 예약 클릭 트래킹
        document.querySelectorAll('a[href*="naver.com"]').forEach(link => {
            link.addEventListener('click', function() {
                trackEvent('booking', {
                    booking_platform: 'naver'
                });
            });
        });
    });

    // 스크롤 깊이 트래킹
    let maxScrollPercentage = 0;
    window.addEventListener('scroll', function() {
        const scrollPercentage = Math.round(
            (window.scrollY / (document.documentElement.scrollHeight - window.innerHeight)) * 100
        );

        if (scrollPercentage > maxScrollPercentage) {
            maxScrollPercentage = scrollPercentage;

            // 25%, 50%, 75%, 100% 지점에서 이벤트 발생
            if ([25, 50, 75, 100].includes(maxScrollPercentage)) {
                trackEvent('scroll', {
                    scroll_depth: maxScrollPercentage
                });
            }
        }
    });

    // 세션 시간 트래킹
    const sessionStart = Date.now();
    window.addEventListener('beforeunload', function() {
        const sessionDuration = Math.round((Date.now() - sessionStart) / 1000);

        trackEvent('session_duration', {
            duration_seconds: sessionDuration
        });
    });

})();

/**
 * 사용 방법:
 *
 * 1. HTML <head>에 GA ID 메타 태그 추가:
 *    <meta name="ga-id" content="G-YOUR-MEASUREMENT-ID">
 *
 * 2. 커스텀 이벤트 트래킹:
 *    trackEvent('button_click', { button_name: 'cta_button' });
 *
 * 3. 페이지뷰 트래킹:
 *    trackPageView('Custom Page Name');
 */