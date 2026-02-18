/**
 * WOOHWAHAE Photography Page
 * 업로드된 이미지 자동 로드 및 갤러리 표시
 */

(function() {
    'use strict';

    const API_URL = 'http://localhost:8083';
    let currentCategory = 'all';
    let allImages = [];

    // DOM 요소
    const filterLinks = document.querySelectorAll('.gallery-filter a');
    const galleryGrid = document.querySelector('.gallery-grid');

    // 이미지 로드
    async function loadImages() {
        try {
            const response = await fetch(`${API_URL}/api/images`);
            if (response.ok) {
                allImages = await response.json();
                displayImages();
            }
        } catch (error) {
            console.error('Failed to load images:', error);
            // 에러시 기본 플레이스홀더 표시
            displayPlaceholders();
        }
    }

    // 이미지 표시
    function displayImages() {
        if (!galleryGrid) return;

        // 카테고리 필터링
        let imagesToShow = allImages;
        if (currentCategory !== 'all') {
            imagesToShow = allImages.filter(img => img.category === currentCategory);
        }

        // 이미지가 없으면 플레이스홀더 표시
        if (imagesToShow.length === 0) {
            displayPlaceholders();
            return;
        }

        // 갤러리 그리드 업데이트
        galleryGrid.innerHTML = '';

        imagesToShow.forEach((image, index) => {
            const item = document.createElement('div');
            item.className = 'gallery-item fade-in';
            item.style.animationDelay = `${index * 0.05}s`;

            // 카테고리별 레이아웃 클래스
            if (image.category === 'hair' && index % 4 === 0) {
                item.classList.add('gallery-item--wide');
            }
            if (image.category === 'space' && index % 6 === 2) {
                item.classList.add('gallery-item--tall');
            }

            item.innerHTML = `
                <img src="${API_URL}/uploads/${image.filename}"
                     alt="${image.category}"
                     loading="lazy"
                     onclick="openLightbox('${API_URL}/uploads/${image.filename}')">
                <div class="gallery-overlay">
                    <span class="gallery-category">${formatCategory(image.category)}</span>
                </div>
            `;

            galleryGrid.appendChild(item);
        });

        // Fade in animation
        setTimeout(() => {
            document.querySelectorAll('.gallery-item').forEach(item => {
                item.classList.add('visible');
            });
        }, 100);
    }

    // 플레이스홀더 표시 (이미지가 없을 때)
    function displayPlaceholders() {
        if (!galleryGrid) return;

        const placeholders = [
            { color: '#f0f0f0', text: 'Hair Work' },
            { color: '#e8e8e8', text: 'Space' },
            { color: '#f5f5f5', text: 'Detail' },
            { color: '#ebebeb', text: 'Moment' },
            { color: '#f8f8f8', text: 'Coming Soon' },
            { color: '#e5e5e5', text: 'Archive' }
        ];

        galleryGrid.innerHTML = '';

        placeholders.forEach((placeholder, index) => {
            const item = document.createElement('div');
            item.className = 'gallery-item fade-in';
            item.style.animationDelay = `${index * 0.05}s`;

            if (index % 4 === 0) item.classList.add('gallery-item--wide');
            if (index % 6 === 2) item.classList.add('gallery-item--tall');

            item.innerHTML = `
                <div class="placeholder" style="background: ${placeholder.color};
                     display: flex; align-items: center; justify-content: center;
                     height: 100%; font-size: 14px; color: #999;">
                    ${placeholder.text}
                </div>
            `;

            galleryGrid.appendChild(item);
        });

        setTimeout(() => {
            document.querySelectorAll('.gallery-item').forEach(item => {
                item.classList.add('visible');
            });
        }, 100);
    }

    // 카테고리 포맷
    function formatCategory(category) {
        const categoryNames = {
            'hair': 'Hair Work',
            'space': 'Space',
            'detail': 'Detail',
            'moment': 'Moment',
            'product': 'Product'
        };
        return categoryNames[category] || category;
    }

    // 필터 클릭 이벤트
    filterLinks.forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();

            // active 클래스 토글
            filterLinks.forEach(l => l.classList.remove('active'));
            link.classList.add('active');

            // 카테고리 업데이트
            currentCategory = link.dataset.category || 'all';

            // 이미지 다시 표시
            displayImages();
        });
    });

    // 라이트박스 (간단한 구현)
    window.openLightbox = function(imageSrc) {
        const lightbox = document.createElement('div');
        lightbox.className = 'lightbox';
        lightbox.innerHTML = `
            <div class="lightbox-content">
                <span class="lightbox-close">&times;</span>
                <img src="${imageSrc}" alt="">
            </div>
        `;

        // 스타일 추가
        const style = document.createElement('style');
        style.textContent = `
            .lightbox {
                position: fixed;
                top: 0;
                left: 0;
                right: 0;
                bottom: 0;
                background: rgba(0,0,0,0.95);
                z-index: 10000;
                display: flex;
                align-items: center;
                justify-content: center;
                animation: fadeIn 0.3s ease;
            }
            .lightbox-content {
                position: relative;
                max-width: 90%;
                max-height: 90%;
            }
            .lightbox-content img {
                max-width: 100%;
                max-height: 90vh;
                display: block;
            }
            .lightbox-close {
                position: absolute;
                top: -40px;
                right: 0;
                color: white;
                font-size: 36px;
                cursor: pointer;
                transition: opacity 0.3s ease;
            }
            .lightbox-close:hover {
                opacity: 0.7;
            }
            @keyframes fadeIn {
                from { opacity: 0; }
                to { opacity: 1; }
            }
        `;

        document.head.appendChild(style);
        document.body.appendChild(lightbox);

        // 닫기 이벤트
        lightbox.addEventListener('click', () => {
            lightbox.remove();
            style.remove();
        });
    };

    // 페이지 로드시 이미지 불러오기
    document.addEventListener('DOMContentLoaded', () => {
        loadImages();

        // 5초마다 새로운 이미지 체크 (선택적)
        // setInterval(loadImages, 5000);
    });

    // 이미지 업로드 링크 추가 (관리자용)
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.get('admin') === 'true') {
        const uploadLink = document.createElement('a');
        uploadLink.href = '/upload.html';
        uploadLink.textContent = '+ Upload Photos';
        uploadLink.style.cssText = `
            position: fixed;
            bottom: 30px;
            right: 30px;
            background: #000;
            color: white;
            padding: 12px 24px;
            border-radius: 30px;
            text-decoration: none;
            z-index: 1000;
            font-size: 14px;
        `;
        document.body.appendChild(uploadLink);
    }

})();