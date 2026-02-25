/* ══════════════════════════════════════════════════════
   FIELD BG — 전체 페이지 배경 쌍극자 자기장 (Three.js)
   스크롤에 따라 극 위치 · 씨앗 · 회전속도 변형
   출처: index.html Three.js IIFE 추출 (가장 완성도 높은 버전)
══════════════════════════════════════════════════════ */
(function () {
  var canvas = document.getElementById('field-bg');
  if (!canvas || typeof THREE === 'undefined') return;

  /* ── prefers-reduced-motion: 접근성 — 애니메이션 비활성화 ── */
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
    canvas.style.display = 'none';
    return;
  }

  /* ── 모바일 감지 — 렌더 품질 조절 ── */
  var isMobile = window.innerWidth < 768;

  /* ── Renderer ── */
  var renderer = new THREE.WebGLRenderer({
    canvas: canvas,
    alpha: true,
    antialias: true  /* 모바일 포함 antialias 활성화 — 선 품질 우선 */
  });
  renderer.setPixelRatio(window.devicePixelRatio);
  renderer.setClearColor(0x000000, 0);

  /* ── Scene & Camera ── */
  var scene    = new THREE.Scene();
  var fieldGroup = new THREE.Group();  /* 회전 그룹 — 씬에 먼저 등록 */
  scene.add(fieldGroup);

  var camera = new THREE.PerspectiveCamera(45, 1, 0.1, 1000);
  camera.position.set(0, 1.5, 20);
  camera.lookAt(0, 0, 0);

  /* ── 마우스 추적 ── */
  var mouse = { x: 0, y: 0 };
  document.addEventListener('mousemove', function (e) {
    mouse.x = (e.clientX / window.innerWidth  - 0.5) * 2;
    mouse.y = (e.clientY / window.innerHeight - 0.5) * 2;
  });

  /* ── 쌍극자 필드라인 파라미터 — 모바일/데스크탑 분기 ── */
  var isPortrait = window.innerHeight > window.innerWidth;
  var LINE_COUNT = isMobile ? 18 : 24;   /* 모바일: φ 슬라이스 3/4 */
  var SEEDS      = isMobile ? 8  : 10;   /* 모바일: 씨앗 수 80% */
  var POINTS_PER = isMobile ? 160 : 220; /* 모바일: 포인트 수 72% */
  var MAX_R      = 15.0;
  /* Portrait 모바일: 필드 클리핑 방지, 단 너무 위축되지 않도록 2.2 유지 */
  var SCALE      = (isMobile && isPortrait) ? 2.2 : 2.8;

  function buildFieldLine(r0, phi) {
    var pts   = new Float32Array(POINTS_PER * 3);
    var baseZ = new Float32Array(POINTS_PER);
    var tMin  = 0.04;
    var tMax  = Math.PI - 0.04;
    for (var i = 0; i < POINTS_PER; i++) {
      var theta = tMin + (i / (POINTS_PER - 1)) * (tMax - tMin);
      var sinT  = Math.sin(theta);
      var cosT  = Math.cos(theta);
      var r     = r0 * sinT * sinT;
      if (r > MAX_R) r = MAX_R;
      pts[i * 3]     = r * sinT * Math.cos(phi) * SCALE;
      pts[i * 3 + 1] = r * cosT * SCALE;
      pts[i * 3 + 2] = r * sinT * Math.sin(phi) * SCALE;
      baseZ[i]       = pts[i * 3 + 2];
    }
    return { pts: pts, baseZ: baseZ };
  }

  var fieldLines = [];

  /* r0 씨앗: 안쪽 조밀, 바깥 성긴 — 중심 수렴감 강화 */
  var seeds = [];
  for (var si = 0; si < SEEDS; si++) {
    seeds.push(0.8 + Math.pow((si + 1) / SEEDS, 1.8) * 11.0);
  }

  for (var li = 0; li < LINE_COUNT; li++) {
    var phi = (li / LINE_COUNT) * Math.PI * 2;
    for (var si2 = 0; si2 < SEEDS; si2++) {
      var r0 = seeds[si2];
      var frac = (si2 + 1) / SEEDS;
      /* 안쪽(소): 진함. 바깥(대): 연함 — 중심 집중감 */
      var baseOp = 0.75 - frac * 0.35;
      /* 색상: 안쪽 다크, 바깥 네이비 — 주파수/에너지 레이어 */
      var col = frac < 0.5 ? 0x1a1a1a : 0x1B2D4F;

      var mat = new THREE.LineBasicMaterial({
        color: col,
        opacity: baseOp,
        transparent: true
      });

      var built = buildFieldLine(r0, phi);
      var geo   = new THREE.BufferGeometry();
      geo.setAttribute('position', new THREE.BufferAttribute(built.pts, 3));
      fieldGroup.add(new THREE.Line(geo, mat));

      fieldLines.push({
        mat:    mat,
        baseOp: baseOp,
        phase:  li * (Math.PI * 2 / LINE_COUNT) + si2 * 0.35,
        geo:    geo,
        pts:    built.pts,
        baseZ:  built.baseZ,
        frac:   frac
      });
    }
  }

  /* ── 중심 극점: 발광 링 3개 (맥동) ── */
  var pulseRings = [];
  [0.18, 0.38, 0.65].forEach(function (r, idx) {
    var segs = 80;
    var pts  = [];
    for (var a = 0; a <= segs; a++) {
      var ang = (a / segs) * Math.PI * 2;
      pts.push(new THREE.Vector3(Math.cos(ang) * r, Math.sin(ang) * r, 0));
    }
    var geo = new THREE.BufferGeometry().setFromPoints(pts);
    var mat = new THREE.LineBasicMaterial({
      color: 0x1a1a1a, opacity: 0.5 - idx * 0.12, transparent: true
    });
    fieldGroup.add(new THREE.Line(geo, mat));
    pulseRings.push({ mat: mat, baseOp: 0.5 - idx * 0.12, phase: idx * 1.2 });
  });

  /* ── 중심 심볼: 나비 로고 텍스처 Plane ── */
  var symbolMesh = null;
  (function () {
    var loader = new THREE.TextureLoader();
    loader.load('/assets/media/brand/symbol.png', function (tex) {
      tex.minFilter = THREE.LinearMipmapLinearFilter;
      tex.magFilter = THREE.LinearFilter;
      tex.anisotropy = renderer.capabilities.getMaxAnisotropy();
      tex.generateMipmaps = true;
      var geo = new THREE.PlaneGeometry(1.2, 1.2);
      var mat = new THREE.MeshBasicMaterial({
        map:         tex,
        transparent: true,
        opacity:     0.45,
        depthTest:   false,
        alphaTest:   0.01
      });
      symbolMesh = new THREE.Mesh(geo, mat);
      symbolMesh.renderOrder = 10;
      fieldGroup.add(symbolMesh);
    });
  })();

  /* ── 리사이즈 ── */
  function onResize() {
    var W = window.innerWidth;
    var H = window.innerHeight;
    renderer.setSize(W, H, false);
    camera.aspect = W / H;
    camera.updateProjectionMatrix();
  }
  onResize();
  window.addEventListener('resize', onResize);

  /* ── 애니메이션 루프 ── */
  var clock   = new THREE.Clock();
  var camX    = 0;
  var camY    = isPortrait && isMobile ? 0 : 2;
  /* Portrait 모바일: 카메라를 뒤로 빼서 필드 전체가 화면 안에 들어오게 */
  var baseZ   = (isMobile && isPortrait) ? 28 : 20;

  function animate() {
    requestAnimationFrame(animate);
    var t = clock.getElapsedTime();

    /* 스크롤 상태 읽기 — _fieldState는 별도 IIFE에서 정의 */
    var sr = (window._fieldState && window._fieldState.ratio) || 0;

    /* 씬 Y축 회전 — 돌의 무거운 자전 (기존 대비 약 60% 감속) */
    var rotSpeed = 0.002 + sr * 0.003;
    fieldGroup.rotation.y = t * rotSpeed;

    /* X축 회전 — 돌의 최소 기울기 (±0.08rad), 느린 lerp */
    fieldGroup.rotation.x += (-mouse.y * 0.08 - fieldGroup.rotation.x) * 0.015;

    /* 카메라 z + FOV — 돌 무게감, 완만한 수렴 + 무거운 관성 */
    var targetZ = baseZ - sr * 8;
    camera.position.z += (targetZ - camera.position.z) * 0.015;
    camera.fov += (45 - sr * 9 - camera.fov) * 0.015;
    camera.updateProjectionMatrix();

    /* 각 필드라인: opacity 맥동 + z-wobble */
    for (var i = 0; i < fieldLines.length; i++) {
      var fl = fieldLines[i];

      /* opacity — 완만한 밝기 변화 */
      var waveBase = 0.38 + sr * 0.10;
      var wave     = 0.5 + 0.5 * Math.sin(t * 0.7 + fl.phase);  /* 느린 맥동 (~9s 주기) */
      /* scrollPulse — 중간(sr≈0.5)에서 최고조, 처음/끝에서 약해짐 (0.7~1.0) */
      var scrollPulse = 0.7 + 0.3 * Math.sin(sr * Math.PI);
      fl.mat.opacity = fl.baseOp * scrollPulse * (waveBase + 0.60 * wave);

      /* z-wobble — 돌의 억제된 진동 (진폭 최소화, 느린 지각 호흡 ~7s) */
      var wobbleMax = 0.15 + sr * 0.12;
      var pos = fl.pts;
      for (var p = 0; p < POINTS_PER; p++) {
        var frac = p / (POINTS_PER - 1);
        var amp  = Math.sin(frac * Math.PI) * wobbleMax;
        pos[p * 3 + 2] = fl.baseZ[p] +
          amp * Math.sin(t * 0.9 + fl.phase + frac * 6.2832); /* 느린 호흡 ~7s = 2π/0.9 */
      }
      fl.geo.attributes.position.needsUpdate = true;
    }

    /* 중심 링 맥동 — 돌의 느린 극점 진동 (~3.5s 주기) */
    for (var ri = 0; ri < pulseRings.length; ri++) {
      var ring = pulseRings[ri];
      ring.mat.opacity = ring.baseOp * (0.4 + 0.6 * Math.sin(t * 1.8 + ri * 2.094));
    }

    /* 나비 심볼 — 필드 호흡과 동기화 (느린 맥동) */
    if (symbolMesh) {
      symbolMesh.material.opacity = (0.35 + sr * 0.3) * (0.8 + 0.2 * Math.sin(t * 0.9)); /* 느린 호흡 동기화 */
      /* 필드 그룹 회전에 역방향 보정 — 심볼은 항상 정면을 향함 */
      symbolMesh.rotation.y = -fieldGroup.rotation.y;
    }

    /* 카메라 x/y — 돌의 둔감한 마우스 반응 */
    camX += (mouse.x * 1.2 - camX) * 0.012;
    camY += (1.5 - mouse.y * 1.0 - camY) * 0.012;
    camera.position.x = camX;
    camera.position.y = camY;
    camera.lookAt(0, 0, 0);

    renderer.render(scene, camera);
  }

  animate();
})();

/* ─── Scroll → 씬 상태 변형 ───
   스크롤 진행률에 따라 극 위치 · 씨앗 수 · 회전속도 lerp.
   상태는 Three.js 루프에서 읽어 fieldGroup / camera에 적용.
─── */
(function () {
  var scrollRatio = 0;  /* 0 ~ 1 */
  var lerpRatio   = 0;  /* 부드러운 추종 */

  function updateScroll() {
    var docH = document.documentElement.scrollHeight - window.innerHeight;
    scrollRatio = docH > 0 ? window.scrollY / docH : 0;
  }
  window.addEventListener('scroll', updateScroll, { passive: true });
  updateScroll();

  /* Three.js 루프에 공유할 상태 객체 */
  window._fieldState = {
    get ratio() { return lerpRatio; }
  };

  /* lerp 루프 */
  (function lerpLoop() {
    lerpRatio += (scrollRatio - lerpRatio) * 0.04;
    requestAnimationFrame(lerpLoop);
  })();
})();
