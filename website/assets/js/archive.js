/**
 * WOOHWAHAE Archive Search & Filter
 * 검색 및 필터링 기능
 */

let allPosts = [];
let currentCategory = 'all';
let searchTerm = '';

// Archive 데이터 로드 및 초기화
function initArchive() {
  fetch('index.json')
    .then(r => r.json())
    .then(posts => {
      // 날짜 최신순 정렬
      allPosts = posts.sort((a, b) => new Date(b.date) - new Date(a.date));
      renderCategories();
      renderPosts(allPosts);
      setupEventListeners();
      updateCount(allPosts.length);
    })
    .catch((e) => {
      console.error('Archive load failed:', e);
      showError();
    });
}

// 카테고리 버튼 렌더링 (전체 포함)
function renderCategories() {
  const filtersContainer = document.getElementById('archive-filters');
  // 중복 없는 카테고리 목록
  const categories = ['all', ...new Set(allPosts.map(p => p.category))];

  filtersContainer.innerHTML = ''; // 기존 버튼 초기화

  categories.forEach(category => {
    const btn = document.createElement('button');
    btn.className = 'filter-btn';
    if (category === 'all') btn.classList.add('active');
    btn.dataset.category = category;
    // 'all' -> '전체', 나머지는 그대로
    btn.textContent = category === 'all' ? '전체' : category;
    filtersContainer.appendChild(btn);
  });
}

// 포스트 렌더링
function renderPosts(posts) {
  const list = document.getElementById('archive-list');
  const empty = document.getElementById('archive-empty');

  list.innerHTML = '';

  if (!posts.length) {
    empty.style.display = 'block';
    list.style.display = 'none';
    return;
  }

  empty.style.display = 'none';
  list.style.display = 'block';

  posts.forEach(post => {
    const li = document.createElement('li');
    li.className = 'archive-item fade-in';

    // 링크 래퍼로 전체 감싸기 (접근성 위해 블록 링크 허용)
    li.innerHTML = `
      <span class="archive-date">${post.date}</span>
      <div class="archive-content">
        <a href="${post.slug}/" class="archive-link-wrapper">
          <p class="archive-issue">${post.issue}</p>
          <h2 class="archive-title">${post.title}</h2>
          <p class="archive-preview">${post.preview}</p>
          <span class="archive-category">${post.category}</span>
        </a>
      </div>
    `;
    list.appendChild(li);
  });

  // Fade-in animation 적용
  animateFadeIn();
}

// 검색 및 필터 적용
function filterPosts() {
  let filtered = allPosts;

  // 카테고리 필터
  if (currentCategory !== 'all') {
    filtered = filtered.filter(p => p.category === currentCategory);
  }

  // 검색어 필터
  if (searchTerm) {
    const term = searchTerm.toLowerCase();
    filtered = filtered.filter(p =>
      p.title.toLowerCase().includes(term) ||
      p.preview.toLowerCase().includes(term) ||
      p.category.toLowerCase().includes(term)
    );
  }

  renderPosts(filtered);
  updateCount(filtered.length);
}

// 이벤트 리스너 설정
function setupEventListeners() {
  // 검색
  const searchInput = document.getElementById('archive-search');
  if (searchInput) {
    searchInput.addEventListener('input', (e) => {
      searchTerm = e.target.value.trim();
      filterPosts();
    });
  }

  // 필터 버튼 (이벤트 위임)
  const filtersContainer = document.getElementById('archive-filters');
  if (filtersContainer) {
    filtersContainer.addEventListener('click', (e) => {
      if (e.target.classList.contains('filter-btn')) {
        // 활성 상태 토글
        document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
        e.target.classList.add('active');

        // 카테고리 업데이트
        currentCategory = e.target.dataset.category;
        filterPosts();
      }
    });
  }
}

// 카운트 업데이트
function updateCount(count) {
  const countEl = document.getElementById('archive-count');
  const total = allPosts.length;

  if (!countEl) return;

  if (count === total) {
    countEl.textContent = `Total ${total}`;
  } else {
    countEl.textContent = `Showing ${count} of ${total}`;
  }
}

// 에러 표시
function showError() {
  const list = document.getElementById('archive-list');
  if (list) {
    list.innerHTML = `
      <li class="archive-item fade-in">
        <div class="archive-content">
          <p class="archive-empty-msg">일시적인 오류로 기록을 불러올 수 없습니다.</p>
        </div>
      </li>
    `;
  }
}

// Fade-in 애니메이션
function animateFadeIn() {
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible');
        observer.unobserve(entry.target);
      }
    });
  }, { threshold: 0.05 });

  document.querySelectorAll('.fade-in').forEach(el => observer.observe(el));
}

// 초기화
document.addEventListener('DOMContentLoaded', initArchive);

