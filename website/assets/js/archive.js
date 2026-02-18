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
  async function loadArchive() {
    try {
      const response = await fetch('index.json');
      const data = await response.json();

      // 초기 렌더링 (All)
      renderArchive(data, 'All');

      // 필터 버튼 이벤트
      document.querySelectorAll('.filter-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
          // Active 클래스 갱신
          document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
          e.target.classList.add('active');

          const category = e.target.dataset.category;
          renderArchive(data, category);
        });
      });

    } catch (error) {
      console.error('Error loading archive:', error);
    }
  }

  function renderArchive(posts, category) {
    const grid = document.querySelector('.archive-grid');
    grid.innerHTML = '';

    const filteredPosts = category === 'All'
      ? posts
      : posts.filter(post => {
        // 매핑: 화면 카테고리 -> 데이터 카테고리
        if (category === 'Essays') return post.category === 'Essay';
        if (category === 'Features') return post.category !== 'Essay'; // 나머지는 Feature 취급
        return true;
      });

    filteredPosts.forEach(post => {
      const card = document.createElement('a');
      card.href = post.url || `issue/${post.slug}`; // url 필드가 있으면 사용, 없으면 slug 기반
      card.className = 'archive-card fade-in';

      // 이미지 처리 (없으면 플레이스홀더)
      // 실제 구현 시에는 post.image_url이 있어야 함. 
      // 현재는 Issue 008만 이미지가 있으므로 체크 필요.
      let imageUrl = post.image_url || 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7'; // 투명
      if (!post.image_url) {
        // 임시: Issue 번호에 따라 다른 회색 톤 적용 (가상 이미지 대신)
        imageUrl = '../../assets/images/placeholder_grain.jpg'; // or similar
      }

      // Issue 008 특수 처리 (방금 생성한 이미지)
      if (post.slug === 'issue-008-raw-materiality') {
        // Artifact 경로를 절대 경로로 쓸 수 없으므로, assets 폴더로 이동시키는 로직이 필요함.
        // 여기서는 일단 하드코딩된 경로(예시)를 가정하거나, 나중에 파일 이동을 수행해야 함.
        // 임시로 로컬 경로 가정:
        imageUrl = 'ticket_image_path_placeholder';
      }

      card.innerHTML = `
                <div class="archive-card-image">
                    <img src="${imageUrl}" alt="${post.title}" loading="lazy">
                </div>
                <div class="archive-card-meta">
                    <span class="issue-num">${post.issue || 'ISSUE'}</span>
                    <span class="issue-date">${post.date}</span>
                </div>
                <h2 class="archive-card-title">${post.title}</h2>
                <p class="archive-card-preview">${post.preview}</p>
            `;

      grid.appendChild(card);
    });

    // Trigger animations
    setTimeout(() => {
      document.querySelectorAll('.fade-in').forEach(el => el.classList.add('visible'));
    }, 50);
  }

  loadArchive();
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

