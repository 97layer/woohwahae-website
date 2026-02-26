/**
 * WOOHWAHAE Archive - Magazine Grid 
 * Clean Implementation
 */

document.addEventListener('DOMContentLoaded', () => {
  initArchive();
});

async function initArchive() {
  console.log('[Archive] Initializing...');
  const grid = document.querySelector('.archive-grid');
  if (!grid) return;

  try {
    let posts = [];
    const inlineData = document.getElementById('archive-data');

    if (inlineData) {
      console.log('[Archive] Loading inlined data...');
      posts = JSON.parse(inlineData.textContent);
    } else {
      console.log('[Archive] Fetching index.json...');
      const response = await fetch('index.json');
      if (!response.ok) throw new Error('Failed to load index.json');
      posts = await response.json();
    }

    console.log('[Archive] Posts loaded:', posts.length);
    renderGrid(posts, 'All');
    setupFilters(posts);

  } catch (error) {
    console.error('[Archive] Error during init:', error);
    grid.innerHTML = '<p class="error-msg">기록을 불러오는 중 오류가 발생했습니다.</p>';
  }
}

function setupFilters(posts) {
  const filterBtns = document.querySelectorAll('.filter-btn');
  filterBtns.forEach(btn => {
    btn.addEventListener('click', (e) => {
      // 버튼 활성화 상태 변경
      filterBtns.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');

      const category = btn.getAttribute('data-category');
      renderGrid(posts, category);
    });
  });
}

function renderGrid(posts, category) {
  const grid = document.querySelector('.archive-grid');
  grid.innerHTML = '';

  // 필터링 필링
  const filtered = category === 'All'
    ? posts
    : posts.filter(p => {
      if (category === 'Essays') return p.category === 'Essay';
      if (category === 'Features') return p.category !== 'Essay';
      return true;
    });

  if (filtered.length === 0) {
    grid.innerHTML = '<p class="empty-msg">해당하는 기록이 없습니다.</p>';
    return;
  }

  filtered.forEach(post => {
    const card = document.createElement('a');
    card.href = post.slug + '/'; // slug 폴더 내 index.html
    card.className = 'archive-card fade-in';

    // 이미지 경로 처리 (JSON에 기록된 경로 사용)
    const imageUrl = post.image_url || 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7';

    card.innerHTML = `
            <div class="archive-card-image">
                <img src="${imageUrl}" alt="${post.title}" loading="lazy">
            </div>
            <div class="archive-card-meta">
                <span class="essay-num">${post.essay || 'ESSAY'}</span>
                <span class="essay-date">${post.date}</span>
            </div>
            <h3 class="archive-card-title">${post.title}</h3>
            <p class="archive-card-preview">${post.preview}</p>
        `;

    grid.appendChild(card);
  });

  // 애니메이션 실행
  setTimeout(() => {
    document.querySelectorAll('.archive-card.fade-in').forEach(el => {
      el.classList.add('visible');
    });
  }, 100);
}
