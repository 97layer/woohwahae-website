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
      allPosts = posts;
      renderCategories();
      renderPosts(posts);
      setupEventListeners();
      updateCount(posts.length);
    })
    .catch(() => {
      showError();
    });
}

// 카테고리 버튼 렌더링
function renderCategories() {
  const filtersContainer = document.getElementById('archive-filters');
  const categories = [...new Set(allPosts.map(p => p.category))];

  categories.forEach(category => {
    const btn = document.createElement('button');
    btn.className = 'filter-btn';
    btn.dataset.category = category;
    btn.textContent = category;
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
    li.innerHTML = `
      <span class="archive-date">${post.date}</span>
      <div class="archive-content">
        <a href="${post.slug}/">
          <p class="archive-issue">${post.issue}</p>
          <h2 class="archive-title">${post.title}</h2>
          <p class="archive-preview">${post.preview}</p>
          <span class="archive-category">${post.category}</span>
        </a>
      </div>
    `;
    list.appendChild(li);
  });

  // Fade-in animation
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
  searchInput.addEventListener('input', (e) => {
    searchTerm = e.target.value;
    filterPosts();
  });

  // 필터 버튼
  document.querySelectorAll('.filter-btn').forEach(btn => {
    btn.addEventListener('click', (e) => {
      // 활성 상태 토글
      document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
      e.target.classList.add('active');

      // 카테고리 업데이트
      currentCategory = e.target.dataset.category;
      filterPosts();
    });
  });
}

// 카운트 업데이트
function updateCount(count) {
  const countEl = document.getElementById('archive-count');
  const total = allPosts.length;

  if (count === total) {
    countEl.textContent = `총 ${total}개의 기록`;
  } else {
    countEl.textContent = `${count}개의 기록 (전체 ${total}개)`;
  }
}

// 에러 표시
function showError() {
  const list = document.getElementById('archive-list');
  list.innerHTML = `
    <li class="archive-item fade-in">
      <div class="archive-content">
        <p class="archive-title archive-empty-msg">콘텐츠를 불러올 수 없습니다.</p>
      </div>
    </li>
  `;
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
  }, { threshold: 0.1 });

  document.querySelectorAll('.fade-in').forEach(el => observer.observe(el));
}

// 초기화
document.addEventListener('DOMContentLoaded', initArchive);
