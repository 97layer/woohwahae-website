/* =============================================
   WOOHWAHAE — Hero Animation System v2
   우화(羽化) 컨셉: 고치 → 나비 변태 과정
   단일 책임 원칙: hero 요소 애니메이션은 이 파일만 제어
   ─────────────────────────────────────────────
   구조:
   1. Particle burst (심볼 파티클 → 타이틀 등장) ← 최초 방문
   2. Typewriter (우측 텍스트 순차 출력)
   3. Scroll parallax (좌우 역방향 이동)
   4. Cursor magnetic (데스크탑)
   ============================================= */

(function () {
  'use strict';

  if (!document.querySelector('.hero-platform-def')) return;

  // ─── 요소 참조 ───
  const heroSection = document.querySelector('.hero');
  const heroInner  = document.querySelector('.hero-inner');
  const defEl      = document.querySelector('.hero-platform-def');

  // ─── hero 요소 등장 트리거 ───
  function showHeroElements(delayOffset) {
    delayOffset = delayOffset || 0;
    setTimeout(() => {
      if (heroInner) heroInner.classList.add('hero-show');
      initTypewriter();
    }, delayOffset);
  }

  // ─── 1. PARTICLE BURST ───
  function initParticleBurst() {
    const canvas = document.createElement('canvas');
    canvas.id = 'hero-canvas';
    Object.assign(canvas.style, {
      position: 'absolute',
      top: '0', left: '0',
      width: '100%', height: '100%',
      pointerEvents: 'none',
      zIndex: '1',
      opacity: '0',
    });
    if (!heroSection) return;
    heroSection.style.position = 'relative';
    heroSection.style.overflow = 'hidden';
    heroSection.prepend(canvas);

    const ctx = canvas.getContext('2d');
    let particles = [];
    let animFrame;
    let phase = 'gather'; // gather → hold → burst → done

    function resize() {
      canvas.width  = heroSection.offsetWidth;
      canvas.height = heroSection.offsetHeight;
    }
    resize();
    window.addEventListener('resize', resize, { passive: true });

    const img = new Image();
    img.src = document.querySelector('.nav-symbol')?.src || '/assets/img/symbol.png';

    img.onload = function () {
      const offscreen = document.createElement('canvas');
      const size = 80;
      offscreen.width  = size;
      offscreen.height = size;
      const offCtx = offscreen.getContext('2d');
      offCtx.drawImage(img, 0, 0, size, size);

      const imageData  = offCtx.getImageData(0, 0, size, size).data;
      const sampleRate = 4;
      const cx    = canvas.width  / 2;
      const cy    = canvas.height / 2;
      const scale = Math.min(canvas.width, canvas.height) * 0.35 / size;

      for (let y = 0; y < size; y += sampleRate) {
        for (let x = 0; x < size; x += sampleRate) {
          const i = (y * size + x) * 4;
          const r = imageData[i], g = imageData[i+1], b = imageData[i+2], a = imageData[i+3];
          if (a > 128 && r < 80 && g < 80 && b < 80) {
            const targetX = cx + (x - size / 2) * scale;
            const targetY = cy + (y - size / 2) * scale;
            const angle = Math.random() * Math.PI * 2;
            const dist  = Math.max(canvas.width, canvas.height) * 0.6;
            particles.push({
              x: cx + Math.cos(angle) * dist,
              y: cy + Math.sin(angle) * dist,
              tx: targetX, ty: targetY,
              vx: 0, vy: 0,
              size: Math.random() * 2 + 1,
              opacity: 0,
            });
          }
        }
      }

      canvas.style.opacity = '1';
      let startTime = null;
      const gatherDuration = 900;
      const holdDuration   = 300;
      const burstDuration  = 600;

      function animate(ts) {
        if (!startTime) startTime = ts;
        const elapsed = ts - startTime;
        ctx.clearRect(0, 0, canvas.width, canvas.height);

        if (phase === 'gather') {
          const progress = Math.min(elapsed / gatherDuration, 1);
          const ease     = 1 - Math.pow(1 - progress, 3);
          particles.forEach(p => {
            p.x += (p.tx - p.x) * ease * 0.15;
            p.y += (p.ty - p.y) * ease * 0.15;
            p.opacity = Math.min(progress * 2, 1);
            ctx.beginPath();
            ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2);
            ctx.fillStyle = `rgba(44,44,44,${p.opacity})`;
            ctx.fill();
          });
          if (progress >= 1) { phase = 'hold'; startTime = ts; }

        } else if (phase === 'hold') {
          particles.forEach(p => {
            ctx.beginPath();
            ctx.arc(p.tx, p.ty, p.size, 0, Math.PI * 2);
            ctx.fillStyle = 'rgba(44,44,44,1)';
            ctx.fill();
          });
          if (elapsed > holdDuration) {
            phase = 'burst'; startTime = ts;
            particles.forEach(p => {
              const angle = Math.atan2(p.ty - cy, p.tx - cx);
              const speed = Math.random() * 8 + 4;
              p.vx = Math.cos(angle) * speed;
              p.vy = Math.sin(angle) * speed;
            });
            // 파티클 버스트 시작 → hero 요소 등장
            showHeroElements(0);
          }

        } else if (phase === 'burst') {
          const progress = Math.min(elapsed / burstDuration, 1);
          particles.forEach(p => {
            p.x += p.vx * (1 - progress) * 2;
            p.y += p.vy * (1 - progress) * 2;
            ctx.beginPath();
            ctx.arc(p.x, p.y, p.size * (1 + progress), 0, Math.PI * 2);
            ctx.fillStyle = `rgba(44,44,44,${1 - progress})`;
            ctx.fill();
          });
          if (progress >= 1) {
            canvas.style.transition = 'opacity 0.3s';
            canvas.style.opacity    = '0';
            cancelAnimationFrame(animFrame);
            return;
          }
        }
        animFrame = requestAnimationFrame(animate);
      }
      animFrame = requestAnimationFrame(animate);
    };

    img.onerror = function () {
      // 이미지 로드 실패 → 즉시 등장
      showHeroElements(0);
    };
  }

  // ─── 2. TYPEWRITER — character-by-character ───
  function initTypewriter() {
    if (!defEl) return;

    const fullText = 'Archive for Slow Life.';
    const CHAR_MS  = 72;   // 글자당 간격 (ms)
    const INIT_MS  = 320;  // 시작 전 딜레이

    defEl.innerHTML = '';
    defEl.style.animation = 'none';
    defEl.style.opacity = '1';
    defEl.style.whiteSpace = 'nowrap';

    // 커서 스팬
    const cursor = document.createElement('span');
    cursor.textContent = '|';
    cursor.style.cssText = [
      'display:inline-block',
      'margin-left:1px',
      'opacity:1',
      'animation:tw-blink 0.9s step-end infinite',
    ].join(';');
    defEl.appendChild(cursor);

    // 커서 깜빡임 keyframe (이미 없을 경우만 추가)
    if (!document.getElementById('tw-style')) {
      const s = document.createElement('style');
      s.id = 'tw-style';
      s.textContent = '@keyframes tw-blink{0%,100%{opacity:1}50%{opacity:0}}';
      document.head.appendChild(s);
    }

    let i = 0;
    setTimeout(function type() {
      if (i < fullText.length) {
        const ch = document.createTextNode(fullText[i]);
        defEl.insertBefore(ch, cursor);
        i++;
        setTimeout(type, CHAR_MS);
      } else {
        // 타이핑 완료 → 커서 1.2초 후 fade out
        setTimeout(() => {
          cursor.style.transition = 'opacity 0.5s';
          cursor.style.animation  = 'none';
          cursor.style.opacity    = '0';
          setTimeout(() => cursor.remove(), 500);
        }, 1200);
      }
    }, INIT_MS);
  }

  // ─── 3. SCROLL PARALLAX ───
  function initParallax() {
    if (!heroInner) return;
    let ticking = false;
    window.addEventListener('scroll', () => {
      if (!ticking) {
        requestAnimationFrame(() => {
          const scrollY = window.scrollY;
          const heroH   = heroSection?.offsetHeight || window.innerHeight;
          if (scrollY < heroH) {
            heroInner.style.transform = `translateY(${scrollY * 0.04}px)`;
          }
          ticking = false;
        });
        ticking = true;
      }
    }, { passive: true });
  }

  // ─── 4. CURSOR MAGNETIC ───
  function initMagnetic() {
    if (window.matchMedia('(hover: none)').matches) return;
    const target = document.querySelector('.hero-symbol-large');
    if (!target) return;

    document.addEventListener('mousemove', (e) => {
      const rect  = target.getBoundingClientRect();
      const cx    = rect.left + rect.width  / 2;
      const cy    = rect.top  + rect.height / 2;
      const dx    = e.clientX - cx;
      const dy    = e.clientY - cy;
      const dist  = Math.sqrt(dx * dx + dy * dy);
      const maxDist = 300;

      if (dist < maxDist) {
        const force = (1 - dist / maxDist) * 8;
        target.style.transition = 'transform 0.3s cubic-bezier(0.16,1,0.3,1)';
        target.style.transform  = `translate(${(dx/dist)*force}px, ${(dy/dist)*force}px)`;
      } else {
        target.style.transition = 'transform 0.6s cubic-bezier(0.16,1,0.3,1)';
        target.style.transform  = 'translate(0,0)';
      }
    }, { passive: true });
  }

  // ─── 실행 ───
  // localStorage 제거: 항상 일관된 애니메이션 시퀀스
  // 각 페이지 로드마다 동일한 히어로 프레임 보장

  // main.js IntersectionObserver가 hero 요소를 건드리지 않도록
  // hero 내부의 .fade-in, .fade-in-slow는 없지만 혹시라도 있으면 제거
  document.querySelectorAll('.hero .fade-in, .hero .fade-in-slow').forEach(el => {
    el.classList.remove('fade-in', 'fade-in-slow');
  });

  // 패럴랙스 + 자기장 항상 실행
  setTimeout(() => {
    initParallax();
    initMagnetic();
  }, 200);

  // 항상 일정한 지연 후 hero 요소 등장 (파티클 버스트 없음, 깔끔한 애니메이션)
  showHeroElements(200);

})();
