/**
 * WOOHWAHAE CMS - 인라인 편집 시스템
 *
 * 사용법:
 * 1. HTML 요소에 data-editable="true" 속성 추가
 * 2. data-page="페이지명" data-element="요소ID" 추가
 * 3. 관리자 모드에서만 편집 가능
 */

(function() {
    'use strict';

    const CMS_API = 'http://localhost:8082/api';
    let isEditMode = false;
    let adminToken = localStorage.getItem('woohwahae_admin_token');

    // 편집 모드 토글 버튼 생성
    function createEditToggle() {
        // URL 파라미터 체크
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.get('edit') !== 'true' && !adminToken) return;

        const toggle = document.createElement('div');
        toggle.id = 'cms-edit-toggle';
        toggle.innerHTML = `
            <style>
                #cms-edit-toggle {
                    position: fixed;
                    bottom: 30px;
                    right: 30px;
                    z-index: 10000;
                    background: #000;
                    color: white;
                    padding: 15px 25px;
                    border-radius: 30px;
                    cursor: pointer;
                    font-size: 14px;
                    box-shadow: 0 4px 15px rgba(0,0,0,0.3);
                    transition: all 0.3s ease;
                }
                #cms-edit-toggle:hover {
                    background: #333;
                    transform: translateY(-2px);
                }
                #cms-edit-toggle.active {
                    background: #4CAF50;
                }

                .cms-editable {
                    position: relative;
                    outline: 2px dashed transparent;
                    transition: all 0.3s ease;
                }
                .cms-editable:hover {
                    outline-color: #4CAF50;
                    background: rgba(76, 175, 80, 0.05);
                }
                .cms-editable[contenteditable="true"] {
                    outline: 2px solid #4CAF50;
                    background: rgba(76, 175, 80, 0.1);
                    padding: 5px;
                    cursor: text;
                }

                .cms-save-btn {
                    position: absolute;
                    top: -35px;
                    right: 0;
                    background: #4CAF50;
                    color: white;
                    border: none;
                    padding: 5px 15px;
                    border-radius: 3px;
                    font-size: 12px;
                    cursor: pointer;
                    z-index: 10001;
                    display: none;
                }
                .cms-editable[contenteditable="true"] + .cms-save-btn {
                    display: block;
                }

                .cms-status {
                    position: fixed;
                    top: 20px;
                    right: 20px;
                    background: #333;
                    color: white;
                    padding: 10px 20px;
                    border-radius: 5px;
                    z-index: 10002;
                    display: none;
                }
                .cms-status.show {
                    display: block;
                    animation: slideIn 0.3s ease;
                }
                .cms-status.success {
                    background: #4CAF50;
                }
                .cms-status.error {
                    background: #f44336;
                }

                @keyframes slideIn {
                    from {
                        transform: translateX(100%);
                        opacity: 0;
                    }
                    to {
                        transform: translateX(0);
                        opacity: 1;
                    }
                }
            </style>
            <span id="toggle-text">✏️ 편집 모드</span>
        `;
        document.body.appendChild(toggle);

        // 상태 메시지 영역
        const status = document.createElement('div');
        status.className = 'cms-status';
        status.id = 'cms-status';
        document.body.appendChild(status);

        // 토글 클릭 이벤트
        toggle.addEventListener('click', toggleEditMode);
    }

    // 편집 모드 토글
    function toggleEditMode() {
        isEditMode = !isEditMode;
        const toggle = document.getElementById('cms-edit-toggle');
        const toggleText = document.getElementById('toggle-text');

        if (isEditMode) {
            toggle.classList.add('active');
            toggleText.textContent = '💾 편집 종료';
            enableEditing();
        } else {
            toggle.classList.remove('active');
            toggleText.textContent = '✏️ 편집 모드';
            disableEditing();
        }
    }

    // 편집 가능 요소 찾기
    function findEditableElements() {
        // data-editable 속성이 있는 요소들
        let editables = document.querySelectorAll('[data-editable="true"]');

        // 없으면 기본 요소들을 편집 가능하게 만들기
        if (editables.length === 0) {
            // 주요 텍스트 요소들 선택
            const selectors = [
                '.hero-title',
                '.hero-subtitle',
                '.hero-platform-def',
                '.hero-platform-body',
                '.section-label',
                '.hub-item-label',
                '.hub-item-desc',
                '.about-body p',
                '.values-title',
                '.values-desc',
                'h1', 'h2', 'h3',
                '.content-card-title',
                '.content-card-preview'
            ];

            selectors.forEach(selector => {
                document.querySelectorAll(selector).forEach((el, index) => {
                    el.setAttribute('data-editable', 'true');
                    el.setAttribute('data-page', getCurrentPage());
                    el.setAttribute('data-element', `${selector.replace(/[.\s]/g, '-')}-${index}`);
                });
            });

            editables = document.querySelectorAll('[data-editable="true"]');
        }

        return editables;
    }

    // 현재 페이지 이름 가져오기
    function getCurrentPage() {
        const path = window.location.pathname;
        if (path === '/' || path === '/index.html') return 'index';
        return path.replace('/', '').replace('.html', '');
    }

    // 편집 활성화
    function enableEditing() {
        const editables = findEditableElements();

        editables.forEach(element => {
            element.classList.add('cms-editable');
            element.contentEditable = true;
            element.spellcheck = false;

            // 저장 버튼 추가
            const saveBtn = document.createElement('button');
            saveBtn.className = 'cms-save-btn';
            saveBtn.textContent = '저장';
            saveBtn.onclick = () => saveContent(element);
            element.parentNode.insertBefore(saveBtn, element.nextSibling);

            // 원본 콘텐츠 저장
            element.dataset.originalContent = element.innerHTML;

            // Enter 키로 저장
            element.addEventListener('keydown', (e) => {
                if (e.key === 'Enter' && e.ctrlKey) {
                    e.preventDefault();
                    saveContent(element);
                }
            });
        });
    }

    // 편집 비활성화
    function disableEditing() {
        const editables = document.querySelectorAll('.cms-editable');

        editables.forEach(element => {
            element.classList.remove('cms-editable');
            element.contentEditable = false;

            // 저장 버튼 제거
            const saveBtn = element.nextSibling;
            if (saveBtn && saveBtn.className === 'cms-save-btn') {
                saveBtn.remove();
            }
        });
    }

    // 콘텐츠 저장
    async function saveContent(element) {
        const page = element.dataset.page || getCurrentPage();
        const elementId = element.dataset.element;
        const content = element.innerHTML;

        // 관리자 토큰 확인
        if (!adminToken) {
            const password = prompt('관리자 비밀번호를 입력하세요:');
            if (!password) return;

            try {
                const response = await fetch(`${CMS_API}/admin/login`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ password })
                });

                if (response.ok) {
                    const data = await response.json();
                    adminToken = data.token;
                    localStorage.setItem('woohwahae_admin_token', adminToken);
                } else {
                    showStatus('비밀번호가 틀렸습니다.', 'error');
                    return;
                }
            } catch (error) {
                showStatus('로그인 실패', 'error');
                return;
            }
        }

        // 콘텐츠 저장
        try {
            const response = await fetch(`${CMS_API}/content/update`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': adminToken
                },
                body: JSON.stringify({
                    page: page,
                    element_id: elementId,
                    content: content
                })
            });

            if (response.ok) {
                showStatus('저장 완료!', 'success');
                element.dataset.originalContent = content;
            } else {
                throw new Error('저장 실패');
            }
        } catch (error) {
            showStatus('저장 실패', 'error');
            console.error('Save error:', error);
        }
    }

    // 상태 메시지 표시
    function showStatus(message, type) {
        const status = document.getElementById('cms-status');
        if (!status) return;

        status.textContent = message;
        status.className = `cms-status show ${type}`;

        setTimeout(() => {
            status.classList.remove('show');
        }, 3000);
    }

    // 저장된 콘텐츠 로드
    async function loadSavedContent() {
        const editables = findEditableElements();
        const page = getCurrentPage();

        for (const element of editables) {
            const elementId = element.dataset.element;
            if (!elementId) continue;

            try {
                const response = await fetch(`${CMS_API}/content/${page}/${elementId}`);
                const data = await response.json();

                if (data.content) {
                    element.innerHTML = data.content;
                }
            } catch (error) {
                console.error('Load error:', error);
            }
        }
    }

    // 초기화
    function init() {
        // 저장된 콘텐츠 로드
        loadSavedContent();

        // 편집 토글 버튼 생성
        createEditToggle();

        // 키보드 단축키
        document.addEventListener('keydown', (e) => {
            // Cmd/Ctrl + E: 편집 모드 토글
            if ((e.metaKey || e.ctrlKey) && e.key === 'e') {
                e.preventDefault();
                const toggle = document.getElementById('cms-edit-toggle');
                if (toggle) toggle.click();
            }
        });
    }

    // DOM 로드 완료 시 초기화
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();