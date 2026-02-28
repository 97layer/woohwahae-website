/* ══════════════════════════════════════════════════════
   FIELD BG — 전체 페이지 배경 쌍극자 자기장 (Three.js)
   품질 압도적 향상: 초정밀 밀도, 깊이 렌더링, 정교한 포그, 부드러운 감쇄
   출처: index.html
══════════════════════════════════════════════════════ */
(function () {
  var canvas = document.getElementById('field-bg');
  if (!canvas || typeof THREE === 'undefined') return;

  /* ── prefers-reduced-motion: 접근성 — 애니메이션 비활성화 ── */
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
    canvas.style.display = 'none';
    return;
  }

  /* ── 모바일 감지 — 부하 경량화 모드 전환용 ── */
  var isMobile = window.innerWidth < 768;
  
  /* ── Renderer ── */
  var renderer = new THREE.WebGLRenderer({
    canvas: canvas,
    alpha: true,
    antialias: true
  });
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
  renderer.setClearColor(0x000000, 0);
 
  /* ── Scene & Camera ── */
  var scene = new THREE.Scene();
  /* 안개 밀도 대폭 감소 (가시성 확보 - 0.035 -> 0.02) */
  scene.fog = new THREE.FogExp2(0xE3E2E0, 0.02);
 
  var fieldGroup = new THREE.Group();
  scene.add(fieldGroup);
 
  var camera = new THREE.PerspectiveCamera(42, window.innerWidth / window.innerHeight, 0.1, 1000);
  camera.position.set(0, 1.2, 22);
  camera.lookAt(0, 0, 0);
 
  /* ── 마우스 추적 (관성 적용) ── */
  var mouse = { x: 0, y: 0 };
  document.addEventListener('mousemove', function (e) {
    mouse.x = (e.clientX / window.innerWidth - 0.5) * 2;
    mouse.y = (e.clientY / window.innerHeight - 0.5) * 2;
  });
 
  /* ── 쌍극자 필드라인 파라미터 (V25 리밸런싱: 선명한 밀도 + 고속 렌더링) ── */
  var isPortrait = window.innerHeight > window.innerWidth;
  var LINE_COUNT = isMobile ? 60 : 180;    /* 선명한 밀도 (160 -> 180) */
  var SEEDS = isMobile ? 12 : 42;         /* 선명한 결 (36 -> 42) */
  var POINTS_PER = isMobile ? 120 : 360;   /* 연산 경량화 (400 -> 360) */
  var MAX_R = 18.0;
  var SCALE = (isMobile && isPortrait) ? 2.8 : 3.6;
 
  function buildFieldLine(r0, phi) {
    var pts = new Float32Array(POINTS_PER * 3);
    var baseZ = new Float32Array(POINTS_PER);
    var tMin = 0.01;
    var tMax = Math.PI - 0.01;
    for (var i = 0; i < POINTS_PER; i++) {
      var theta = tMin + (i / (POINTS_PER - 1)) * (tMax - tMin);
      var sinT = Math.sin(theta);
      var cosT = Math.cos(theta);
      var r = r0 * sinT * sinT;
      if (r > MAX_R) r = MAX_R;
      pts[i * 3] = r * sinT * Math.cos(phi) * SCALE;
      pts[i * 3 + 1] = r * cosT * SCALE;
      pts[i * 3 + 2] = r * sinT * Math.sin(phi) * SCALE;
      baseZ[i] = pts[i * 3 + 2];
    }
    return { pts: pts, baseZ: baseZ };
  }
 
  var fieldLines = [];
 
  /* 지수적 분포 씨앗 */
  var seeds = [];
  for (var si = 0; si < SEEDS; si++) {
    seeds.push(0.3 + Math.pow((si + 1) / SEEDS, 2.6) * 14.0);
  }
 
  for (var li = 0; li < LINE_COUNT; li++) {
    var phi = (li / LINE_COUNT) * Math.PI * 2;
    for (var si2 = 0; si2 < SEEDS; si2++) {
      var r0 = seeds[si2];
      var frac = (si2 + 1) / SEEDS;
      
      /* 선의 가시성 확보를 위해 투명도 범위를 살짝 상향 (0.28 ~ 0.12) */
      var baseOp = 0.28 - frac * 0.16;
      var col = frac < 0.4 ? 0x000000 : (frac < 0.8 ? 0x111111 : 0x333333);
 
      var mat = new THREE.LineBasicMaterial({
        color: col,
        opacity: baseOp,
        transparent: true,
        depthWrite: false,
        linewidth: 1
      });
 
      var built = buildFieldLine(r0, phi);
      var geo = new THREE.BufferGeometry();
      geo.setAttribute('position', new THREE.BufferAttribute(built.pts, 3));
      fieldGroup.add(new THREE.Line(geo, mat));
 
      fieldLines.push({
        mat: mat,
        baseOp: baseOp,
        phase: li * (Math.PI * 2 / LINE_COUNT) + si2 * 0.45,
        geo: geo,
        pts: built.pts,
        baseZ: built.baseZ,
        frac: frac
      });
    }
  }
 
  /* ── 분진 레이어 (Particle Dust) — 경량 질감 ── */
  var dustCount = isMobile ? 1200 : 3000;
  var dustGeo = new THREE.BufferGeometry();
  var dustPos = new Float32Array(dustCount * 3);
  for (var di = 0; di < dustCount; di++) {
    var r = Math.random() * 15 * SCALE;
    var theta = Math.random() * Math.PI;
    var phi = Math.random() * Math.PI * 2;
    dustPos[di * 3] = r * Math.sin(theta) * Math.cos(phi);
    dustPos[di * 3 + 1] = r * Math.cos(theta);
    dustPos[di * 3 + 2] = r * Math.sin(theta) * Math.sin(phi);
  }
  dustGeo.setAttribute('position', new THREE.BufferAttribute(dustPos, 3));
  var dustMat = new THREE.PointsMaterial({
    color: 0x000000,
    size: 0.05,
    transparent: true,
    opacity: 0.08,
    sizeAttenuation: true
  });
  var dustSystem = new THREE.Points(dustGeo, dustMat);
  fieldGroup.add(dustSystem);

  var lastW = window.innerWidth;
  function onResize() {
    var W = window.innerWidth;
    var H = window.innerHeight;

    /* 모바일 주소창 토글에 의한 높이 변화는 무시 (진정한 해상도 조정만 수행) */
    if (isMobile && W === lastW && Math.abs(H - renderer.domElement.height / renderer.getPixelRatio()) < 100) {
      return;
    }

    lastW = W;
    renderer.setSize(W, H, false);
    camera.aspect = W / H;
    camera.updateProjectionMatrix();
    isPortrait = H > W;
  }
  onResize();
  window.addEventListener('resize', onResize, { passive: true });

  /* ── 정교한 애니메이션 루프 ── */
  var clock = new THREE.Clock();
  var camX = 0;
  var camY = (isPortrait && isMobile) ? 0.5 : 1.5;
  var baseZ = (isPortrait && isMobile) ? 32 : 25;

  function animate() {
    requestAnimationFrame(animate);
    var t = clock.getElapsedTime();
    var sr = (window._fieldState && window._fieldState.ratio) || 0;

    var rotSpeed = 0.0008 + sr * 0.002;
    fieldGroup.rotation.y = t * rotSpeed;
    fieldGroup.rotation.x += (-mouse.y * 0.04 - fieldGroup.rotation.x) * 0.01;

    var targetZ = baseZ - sr * 12;
    camera.position.z += (targetZ - camera.position.z) * 0.01;
    camera.fov += (38 - sr * 10 - camera.fov) * 0.01;
    camera.updateProjectionMatrix();

    for (var i = 0; i < fieldLines.length; i++) {
      var fl = fieldLines[i];
      var scrollPulse = 0.82 + 0.18 * Math.sin(sr * Math.PI);
      var wave = 0.4 + 0.6 * Math.sin(t * 0.4 + fl.phase);
      fl.mat.opacity = fl.baseOp * scrollPulse * wave;

      var wobbleMax = 0.1 + sr * 0.12;
      var pos = fl.pts;
      for (var p = 0; p < POINTS_PER; p++) {
        var frac = p / (POINTS_PER - 1);
        var amp = Math.sin(frac * Math.PI) * wobbleMax;
        pos[p * 3 + 2] = fl.baseZ[p] + amp * Math.sin(t * 0.5 + fl.phase + frac * 6.28);
      }
      fl.geo.attributes.position.needsUpdate = true;
    }

    /* 입자 레이어 미세 맥동 */
    dustMat.opacity = 0.04 + 0.04 * Math.sin(t * 0.3);
    dustSystem.rotation.y = t * 0.002;

    camX += (mouse.x * 3.0 - camX) * 0.006;
    camY += (1.5 - mouse.y * 2.5 - camY) * 0.006;
    camera.position.x = camX;
    camera.position.y = camY;
    camera.lookAt(0, 0, 0);

    renderer.render(scene, camera);
  }

  animate();
})();

(function () {
  var scrollRatio = 0;
  var lerpRatio = 0;

  function updateScroll() {
    /* window.innerHeight 대신 clientHeight를 사용하여 주소창 변화에 의한 튕김 방지 */
    var viewportH = document.documentElement.clientHeight;
    var docH = document.documentElement.scrollHeight - viewportH;
    scrollRatio = docH > 0 ? window.scrollY / docH : 0;
  }
  window.addEventListener('scroll', updateScroll, { passive: true });
  window.addEventListener('resize', updateScroll, { passive: true });
  updateScroll();

  window._fieldState = {
    get ratio() { return lerpRatio; }
  };

  (function lerpLoop() {
    lerpRatio += (scrollRatio - lerpRatio) * 0.035; /* 관성 증가 */
    requestAnimationFrame(lerpLoop);
  })();
})();
