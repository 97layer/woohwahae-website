/* ══════════════════════════════════════════════════════
   FIELD BG — signature field presets for Next.js
   wave presets + non-wave object presets
══════════════════════════════════════════════════════ */
(function () {
  var RETRY_MAX = 120;
  var RETRY_DELAY = 60;
  var activeFieldRuntime = null;
  var queuedRestart = 0;

  function getViewportSize() {
    var viewport = window.visualViewport;
    var width = viewport && viewport.width ? viewport.width : (window.innerWidth || document.documentElement.clientWidth || 1);
    var height = viewport && viewport.height ? viewport.height : (window.innerHeight || document.documentElement.clientHeight || 1);
    return {
      width: Math.max(1, Math.round(width)),
      height: Math.max(1, Math.round(height))
    };
  }

  function getRendererPixelRatio(width) {
    // Slightly lift the ceiling for crisper lines on high-DPI displays while
    // still capping to avoid runaway GPU cost on very dense panels.
    return width < 768
      ? Math.min(window.devicePixelRatio || 1, 2.0)
      : Math.min(window.devicePixelRatio || 1, 3.0);
  }

  function getMotionLayoutKey(viewport) {
    return (viewport.width < 768 ? 'mobile' : 'desktop') + '|' + (viewport.height > viewport.width ? 'portrait' : 'landscape');
  }

  function disposeMaterial(material) {
    if (material && typeof material.dispose === 'function') material.dispose();
  }

  function disposeObject(node) {
    if (!node) return;
    if (node.geometry && typeof node.geometry.dispose === 'function') node.geometry.dispose();
    if (Array.isArray(node.material)) {
      for (var i = 0; i < node.material.length; i++) disposeMaterial(node.material[i]);
      return;
    }
    disposeMaterial(node.material);
  }

  function queueRestart() {
    if (queuedRestart) return;
    queuedRestart = window.setTimeout(function () {
      queuedRestart = 0;
      if (activeFieldRuntime && typeof activeFieldRuntime.destroy === 'function') {
        activeFieldRuntime.destroy();
      }
      activeFieldRuntime = null;
      var canvas = document.getElementById('field-bg');
      if (canvas) delete canvas.dataset.fieldInit;
      init(0);
    }, 0);
  }

  function isNavOverlayOpen() {
    return !!(document.body && document.body.classList.contains('nav-open'));
  }

  function getPresetConfig(name, isMobile, isPortrait) {
    var base = {
      mode: 'wave',
      initZ: (isMobile && isPortrait) ? 28 : 18.5,
      initFov: (isMobile && isPortrait) ? 46 : 31,
      fogDensity: (isMobile ? 0.022 : 0.0165),
      fieldColor: 0x090909,
      fogHex: 0xECECEA,
      lineCount: isMobile ? 28 : 104,
      seeds: isMobile ? 7 : 14,
      pointsPer: isMobile ? 72 : 360,
      maxR: 24.5,
      scale: (isMobile && isPortrait) ? 3.35 : 4.6,
      opacityMobileBase: 0.135,
      opacityMobileFalloff: 0.06,
      opacityDesktopBase: 0.325,
      opacityDesktopFalloff: 0.18,
      dustCount: isMobile ? 0 : 1400,
      dustOpacity: 0.042,
      rotationSpeed: 0.014,
      breatheAmp: 0.003,
      breatheSpeed: 0.20,
      pulseMin: 0.80,
      pulseRange: 0.20,
      pulseSpeed: 0.09,
      zoomRange: isMobile ? 3 : 5,
      tiltX: -0.4,
      tiltY: 0.65,
      tiltLerp: 0.03,
      coreBias: 2.2,
      seedMax: 15.0,
      guideLoops: 0,
      guideMeridians: 0,
      guideOpacity: 0.06,
      guideScaleX: 1,
      guideScaleY: 1,
      guideScaleZ: 1,
      guidePulse: 0.10,
      coreDustCount: 0,
      coreDustOpacity: 0.08,
      coreDustRadius: 2.4,
      meshScale: isMobile ? 2.35 : 3.15,
      meshOpacity: 0.16,
      meshOpacity2: 0.08,
      meshDrift: 0.010,
      meshColor: 0x111111,
      meshColor2: 0xF2F0EC,
      particleCount: isMobile ? 600 : 1800,
      particleOpacity: 0.18
    };

    if (name === 'ink') {
      base.fogDensity *= 0.88;
      base.fieldColor = 0x060606;
      base.fogHex = 0xE8E7E4;
      base.lineCount = isMobile ? 32 : 128;
      base.seeds = isMobile ? 8 : 17;
      base.pointsPer = isMobile ? 84 : 420;
      base.scale = (isMobile && isPortrait) ? 3.15 : 4.35;
      base.opacityMobileBase = 0.16;
      base.opacityDesktopBase = 0.37;
      base.opacityDesktopFalloff = 0.20;
      base.dustCount = isMobile ? 0 : 1900;
      base.dustOpacity = 0.056;
      base.rotationSpeed = 0.011;
      base.breatheAmp = 0.0046;
      base.pulseMin = 0.84;
      base.pulseRange = 0.16;
      base.zoomRange = isMobile ? 2.4 : 4.2;
      base.coreBias = 2.6;
      base.seedMax = 13.8;
      base.coreDustCount = isMobile ? 36 : 160;
      base.coreDustOpacity = 0.10;
      base.coreDustRadius = 1.8;
      return base;
    }

    if (name === 'orbit') {
      base.fogDensity *= 0.78;
      base.fieldColor = 0x0B0B0B;
      base.fogHex = 0xEFEDEB;
      base.lineCount = isMobile ? 30 : 116;
      base.seeds = isMobile ? 7 : 15;
      base.pointsPer = isMobile ? 88 : 400;
      base.scale = (isMobile && isPortrait) ? 3.5 : 4.95;
      base.opacityMobileBase = 0.14;
      base.opacityDesktopBase = 0.34;
      base.opacityDesktopFalloff = 0.16;
      base.dustCount = isMobile ? 0 : 1200;
      base.dustOpacity = 0.03;
      base.rotationSpeed = 0.020;
      base.breatheAmp = 0.0024;
      base.pulseMin = 0.78;
      base.pulseRange = 0.22;
      base.zoomRange = isMobile ? 3.2 : 5.6;
      base.tiltY = 0.72;
      base.coreBias = 1.95;
      base.seedMax = 16.0;
      base.guideLoops = isMobile ? 1 : 2;
      base.guideOpacity = 0.045;
      return base;
    }

    if (name === 'architect') {
      base.fogDensity *= 0.72;
      base.fieldColor = 0x050505;
      base.fogHex = 0xEFEDE9;
      base.lineCount = isMobile ? 28 : 124;
      base.seeds = isMobile ? 7 : 16;
      base.pointsPer = isMobile ? 90 : 440;
      base.scale = (isMobile && isPortrait) ? 3.0 : 4.15;
      base.opacityMobileBase = 0.17;
      base.opacityDesktopBase = 0.39;
      base.opacityDesktopFalloff = 0.23;
      base.dustCount = isMobile ? 0 : 900;
      base.dustOpacity = 0.024;
      base.rotationSpeed = 0.008;
      base.breatheAmp = 0.0018;
      base.breatheSpeed = 0.16;
      base.pulseMin = 0.88;
      base.pulseRange = 0.12;
      base.pulseSpeed = 0.07;
      base.zoomRange = isMobile ? 2.4 : 3.8;
      base.tiltX = -0.22;
      base.tiltY = 0.42;
      base.tiltLerp = 0.024;
      base.coreBias = 2.9;
      base.seedMax = 12.2;
      base.guideLoops = isMobile ? 2 : 4;
      base.guideMeridians = isMobile ? 2 : 5;
      base.guideOpacity = 0.072;
      base.guideScaleX = 1.04;
      base.guideScaleY = 0.86;
      base.guideScaleZ = 1.04;
      base.guidePulse = 0.08;
      base.coreDustCount = isMobile ? 40 : 220;
      base.coreDustOpacity = 0.12;
      base.coreDustRadius = 1.35;
      return base;
    }

    if (name === 'observer') {
      base.fogDensity *= 0.64;
      base.fieldColor = 0x040404;
      base.fogHex = 0xF0EFEB;
      base.lineCount = isMobile ? 20 : 92;
      base.seeds = isMobile ? 6 : 12;
      base.pointsPer = isMobile ? 96 : 460;
      base.scale = (isMobile && isPortrait) ? 2.9 : 3.95;
      base.opacityMobileBase = 0.13;
      base.opacityDesktopBase = 0.29;
      base.opacityDesktopFalloff = 0.19;
      base.dustCount = isMobile ? 0 : 520;
      base.dustOpacity = 0.015;
      base.rotationSpeed = 0.004;
      base.breatheAmp = 0.0012;
      base.breatheSpeed = 0.12;
      base.pulseMin = 0.90;
      base.pulseRange = 0.08;
      base.pulseSpeed = 0.05;
      base.zoomRange = isMobile ? 1.8 : 2.6;
      base.tiltX = -0.10;
      base.tiltY = 0.18;
      base.tiltLerp = 0.018;
      base.coreBias = 3.2;
      base.seedMax = 10.6;
      base.guideLoops = isMobile ? 1 : 3;
      base.guideMeridians = isMobile ? 1 : 3;
      base.guideOpacity = 0.048;
      base.guideScaleX = 1.08;
      base.guideScaleY = 0.72;
      base.guideScaleZ = 1.08;
      base.guidePulse = 0.05;
      base.coreDustCount = isMobile ? 22 : 120;
      base.coreDustOpacity = 0.065;
      base.coreDustRadius = 0.92;
      return base;
    }

    if (name === 'membrane') {
      base.mode = 'membrane';
      base.initZ = (isMobile && isPortrait) ? 16.5 : 12.8;
      base.initFov = (isMobile && isPortrait) ? 42 : 26;
      base.fogDensity *= 0.56;
      base.fogHex = 0xF0EFEB;
      base.zoomRange = isMobile ? 1.1 : 1.8;
      base.tiltX = -0.06;
      base.tiltY = 0.10;
      base.tiltLerp = 0.016;
      base.meshScale = isMobile ? 2.45 : 3.45;
      base.meshOpacity = 0.14;
      base.meshOpacity2 = 0.065;
      base.meshDrift = 0.014;
      base.meshColor = 0x111111;
      base.meshColor2 = 0xF6F3EE;
      return base;
    }

    if (name === 'monolith') {
      base.mode = 'monolith';
      base.initZ = (isMobile && isPortrait) ? 15.2 : 11.9;
      base.initFov = (isMobile && isPortrait) ? 38 : 24;
      base.fogDensity *= 0.52;
      base.fogHex = 0xF2F1ED;
      base.zoomRange = isMobile ? 0.9 : 1.6;
      base.tiltX = -0.04;
      base.tiltY = 0.08;
      base.tiltLerp = 0.014;
      base.meshScale = isMobile ? 1.85 : 2.65;
      base.meshOpacity = 0.18;
      base.meshOpacity2 = 0.10;
      base.meshDrift = 0.006;
      base.meshColor = 0x0B0B0B;
      base.meshColor2 = 0x2A2A2A;
      return base;
    }

    if (name === 'constellation') {
      base.mode = 'constellation';
      base.initZ = (isMobile && isPortrait) ? 18.0 : 13.6;
      base.initFov = (isMobile && isPortrait) ? 40 : 24;
      base.fogDensity *= 0.46;
      base.fogHex = 0xF3F1ED;
      base.zoomRange = isMobile ? 1.4 : 2.1;
      base.tiltX = -0.08;
      base.tiltY = 0.14;
      base.tiltLerp = 0.015;
      base.particleCount = isMobile ? 800 : 2600;
      base.particleOpacity = 0.14;
      return base;
    }

    return base;
  }

  function buildGuideLoop(THREE, radiusX, radiusY, radiusZ, segments, axis) {
    var pts = new Float32Array((segments + 1) * 3);
    for (var i = 0; i <= segments; i++) {
      var t = (i / segments) * Math.PI * 2;
      var x = Math.cos(t) * radiusX;
      var y = Math.sin(t) * radiusY;
      var z = Math.sin(t) * radiusZ;
      if (axis === 'xz') {
        pts[i * 3] = x;
        pts[i * 3 + 1] = 0;
        pts[i * 3 + 2] = z;
      } else if (axis === 'yz') {
        pts[i * 3] = 0;
        pts[i * 3 + 1] = y;
        pts[i * 3 + 2] = z;
      } else {
        pts[i * 3] = x;
        pts[i * 3 + 1] = y;
        pts[i * 3 + 2] = 0;
      }
    }
    var geo = new THREE.BufferGeometry();
    geo.setAttribute('position', new THREE.BufferAttribute(pts, 3));
    return geo;
  }

  function addGuideLines(THREE, fieldGroup, preset) {
    var guides = [];
    var segments = 180;
    for (var i = 0; i < preset.guideLoops; i++) {
      var radius = preset.scale * (1.4 + i * 0.92);
      var mat = new THREE.LineBasicMaterial({
        color: preset.fieldColor,
        transparent: true,
        opacity: preset.guideOpacity - i * 0.01,
        depthWrite: false
      });
      var geo = buildGuideLoop(THREE, radius * preset.guideScaleX, radius * 0.18, radius * preset.guideScaleZ, segments, 'xz');
      var line = new THREE.Line(geo, mat);
      fieldGroup.add(line);
      guides.push({ mesh: line, mat: mat, baseOp: mat.opacity, drift: 0.003 + i * 0.0015, mode: 'loop' });
    }

    for (var j = 0; j < preset.guideMeridians; j++) {
      var radiusY = preset.scale * (1.35 + j * 0.75);
      var radiusZ = preset.scale * (1.55 + j * 0.64);
      var meridianMat = new THREE.LineBasicMaterial({
        color: preset.fieldColor,
        transparent: true,
        opacity: Math.max(0.026, preset.guideOpacity - j * 0.011),
        depthWrite: false
      });
      var meridianGeo = buildGuideLoop(THREE, radiusY * 0.2, radiusY * preset.guideScaleY, radiusZ * 0.42, segments, 'yz');
      var meridian = new THREE.Line(meridianGeo, meridianMat);
      meridian.rotation.y = (j / Math.max(preset.guideMeridians, 1)) * Math.PI;
      fieldGroup.add(meridian);
      guides.push({ mesh: meridian, mat: meridianMat, baseOp: meridianMat.opacity, drift: 0.0022 + j * 0.0011, mode: 'meridian' });
    }

    return guides;
  }

  function addCoreDust(THREE, fieldGroup, preset) {
    if (!preset.coreDustCount) return null;
    var geo = new THREE.BufferGeometry();
    var pos = new Float32Array(preset.coreDustCount * 3);
    for (var i = 0; i < preset.coreDustCount; i++) {
      var r = Math.random() * preset.coreDustRadius;
      var t = Math.random() * Math.PI;
      var p = Math.random() * Math.PI * 2;
      pos[i * 3] = r * Math.sin(t) * Math.cos(p);
      pos[i * 3 + 1] = r * Math.cos(t) * 0.7;
      pos[i * 3 + 2] = r * Math.sin(t) * Math.sin(p);
    }
    geo.setAttribute('position', new THREE.BufferAttribute(pos, 3));
    var mat = new THREE.PointsMaterial({
      color: preset.fieldColor,
      size: 0.042,
      transparent: true,
      opacity: preset.coreDustOpacity,
      sizeAttenuation: true
    });
    var points = new THREE.Points(geo, mat);
    fieldGroup.add(points);
    return { mesh: points, mat: mat, baseOp: mat.opacity };
  }

  function addSceneLights(THREE, scene) {
    var ambient = new THREE.AmbientLight(0xffffff, 0.92);
    var key = new THREE.DirectionalLight(0xffffff, 0.58);
    key.position.set(4, 6, 8);
    var rim = new THREE.DirectionalLight(0xffffff, 0.26);
    rim.position.set(-5, -1, -6);
    scene.add(ambient);
    scene.add(key);
    scene.add(rim);
  }

  function createMembraneField(THREE, scene, fieldGroup, preset, isMobile) {
    addSceneLights(THREE, scene);
    var segments = isMobile ? 52 : 96;
    var geo = new THREE.SphereGeometry(preset.meshScale, segments, segments);
    var positions = geo.attributes.position.array.slice(0);
    geo.userData.basePositions = positions;

    var shell = new THREE.Mesh(
      geo,
      new THREE.MeshPhysicalMaterial({
        color: preset.meshColor2,
        roughness: 0.18,
        metalness: 0.02,
        transparent: true,
        opacity: preset.meshOpacity,
        transmission: 0.08,
        thickness: 0.5,
        side: THREE.DoubleSide
      })
    );
    shell.scale.set(1.0, 0.72, 1.12);

    var core = new THREE.Mesh(
      geo.clone(),
      new THREE.MeshStandardMaterial({
        color: preset.meshColor,
        transparent: true,
        opacity: preset.meshOpacity2,
        roughness: 0.44,
        metalness: 0.06,
        side: THREE.BackSide
      })
    );
    core.scale.set(0.86, 0.62, 0.95);

    fieldGroup.add(shell);
    fieldGroup.add(core);
    return { type: 'membrane', shell: shell, core: core, base1: positions, base2: core.geometry.attributes.position.array.slice(0) };
  }

  function createMonolithField(THREE, scene, fieldGroup, preset, isMobile) {
    addSceneLights(THREE, scene);
    var geo = new THREE.CapsuleGeometry(isMobile ? 0.92 : 1.08, isMobile ? 4.2 : 5.6, 12, 28);
    var body = new THREE.Mesh(
      geo,
      new THREE.MeshPhysicalMaterial({
        color: preset.meshColor2,
        roughness: 0.34,
        metalness: 0.04,
        transparent: true,
        opacity: preset.meshOpacity,
        transmission: 0.02,
        side: THREE.DoubleSide
      })
    );
    body.scale.set(0.8, 1.0, 0.54);

    var shadowCore = new THREE.Mesh(
      geo.clone(),
      new THREE.MeshStandardMaterial({
        color: preset.meshColor,
        roughness: 0.55,
        metalness: 0.08,
        transparent: true,
        opacity: preset.meshOpacity2,
        side: THREE.BackSide
      })
    );
    shadowCore.scale.set(0.66, 0.92, 0.42);

    var haloGeo = new THREE.RingGeometry(1.8, 2.15, 96);
    var halo = new THREE.Mesh(
      haloGeo,
      new THREE.MeshBasicMaterial({
        color: 0x121212,
        transparent: true,
        opacity: 0.045,
        side: THREE.DoubleSide
      })
    );
    halo.rotation.x = Math.PI / 2;
    halo.position.y = -0.1;
    halo.scale.set(1.3, 1.0, 1.0);

    fieldGroup.add(halo);
    fieldGroup.add(body);
    fieldGroup.add(shadowCore);
    return { type: 'monolith', body: body, core: shadowCore, halo: halo };
  }

  function createConstellationField(THREE, scene, fieldGroup, preset, isMobile) {
    var geo = new THREE.BufferGeometry();
    var pos = new Float32Array(preset.particleCount * 3);
    var scale = isMobile ? 4.2 : 5.6;
    for (var i = 0; i < preset.particleCount; i++) {
      var theta = Math.random() * Math.PI * 2;
      var phi = Math.acos(2 * Math.random() - 1);
      var radius = Math.pow(Math.random(), 1.9) * scale;
      pos[i * 3] = radius * Math.sin(phi) * Math.cos(theta);
      pos[i * 3 + 1] = radius * Math.cos(phi) * 0.68;
      pos[i * 3 + 2] = radius * Math.sin(phi) * Math.sin(theta);
    }
    geo.setAttribute('position', new THREE.BufferAttribute(pos, 3));

    var cloud = new THREE.Points(
      geo,
      new THREE.PointsMaterial({
        color: 0x090909,
        size: isMobile ? 0.055 : 0.042,
        transparent: true,
        opacity: preset.particleOpacity,
        sizeAttenuation: true
      })
    );

    var coreGeo = new THREE.BufferGeometry();
    var corePos = new Float32Array((isMobile ? 80 : 220) * 3);
    for (var j = 0; j < corePos.length / 3; j++) {
      var r = Math.random() * 0.9;
      var t = Math.random() * Math.PI * 2;
      var p = Math.random() * Math.PI;
      corePos[j * 3] = r * Math.sin(p) * Math.cos(t);
      corePos[j * 3 + 1] = r * Math.cos(p) * 0.72;
      corePos[j * 3 + 2] = r * Math.sin(p) * Math.sin(t);
    }
    coreGeo.setAttribute('position', new THREE.BufferAttribute(corePos, 3));
    var core = new THREE.Points(
      coreGeo,
      new THREE.PointsMaterial({
        color: 0x000000,
        size: isMobile ? 0.07 : 0.052,
        transparent: true,
        opacity: 0.24,
        sizeAttenuation: true
      })
    );

    fieldGroup.add(cloud);
    fieldGroup.add(core);
    return { type: 'constellation', cloud: cloud, core: core };
  }

  function createWaveField(THREE, fieldGroup, preset, isMobile) {
    function buildFieldLine(r0, phi) {
      var pts = new Float32Array(preset.pointsPer * 3);
      var tMin = 0.01;
      var tMax = Math.PI - 0.01;
      for (var i = 0; i < preset.pointsPer; i++) {
        var theta = tMin + (i / (preset.pointsPer - 1)) * (tMax - tMin);
        var sinT = Math.sin(theta);
        var cosT = Math.cos(theta);
        var r = r0 * Math.pow(sinT, preset.coreBias);
        if (r > preset.maxR) r = preset.maxR;
        pts[i * 3] = r * sinT * Math.cos(phi) * preset.scale;
        pts[i * 3 + 1] = r * cosT * preset.scale;
        pts[i * 3 + 2] = r * sinT * Math.sin(phi) * preset.scale;
      }
      return pts;
    }

    var seeds = [];
    for (var si = 0; si < preset.seeds; si++) {
      seeds.push(0.4 + Math.pow(si / (preset.seeds - 1), preset.coreBias) * preset.seedMax);
    }

    var fieldLines = [];
    for (var li = 0; li < preset.lineCount; li++) {
      var phi = (li / preset.lineCount) * Math.PI * 2;
      for (var sj = 0; sj < preset.seeds; sj++) {
        var frac = (sj + 1) / preset.seeds;
        var baseOp = isMobile
          ? (preset.opacityMobileBase - frac * preset.opacityMobileFalloff)
          : (preset.opacityDesktopBase - frac * preset.opacityDesktopFalloff);
        var mat = new THREE.LineBasicMaterial({
          color: preset.fieldColor,
          opacity: baseOp,
          transparent: true,
          depthWrite: false,
          linewidth: 1
        });
        var geo = new THREE.BufferGeometry();
        geo.setAttribute('position', new THREE.BufferAttribute(buildFieldLine(seeds[sj], phi), 3));
        var line = new THREE.Line(geo, mat);
        fieldGroup.add(line);
        fieldLines.push({
          mesh: line,
          mat: mat,
          baseOp: baseOp,
          phase: li * (Math.PI * 2 / preset.lineCount) + sj * 0.45,
          depth: frac
        });
      }
    }

    return {
      type: 'wave',
      fieldLines: fieldLines,
      guideLines: addGuideLines(THREE, fieldGroup, preset),
      coreDust: addCoreDust(THREE, fieldGroup, preset)
    };
  }

  function init(attempt) {
    if (activeFieldRuntime) return;

    var canvas = document.getElementById('field-bg');
    if (!canvas) {
      if (attempt < RETRY_MAX) window.setTimeout(function () { init(attempt + 1); }, RETRY_DELAY);
      return;
    }
    if (canvas.dataset.fieldInit === 'ready') return;
    if (typeof THREE === 'undefined') {
      if (attempt < RETRY_MAX) window.setTimeout(function () { init(attempt + 1); }, RETRY_DELAY);
      return;
    }

    var reducedMotionQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    var reduceMotion = reducedMotionQuery.matches;

    canvas.style.display = '';
    canvas.dataset.fieldInit = 'ready';

    var viewport = getViewportSize();
    var isMobile = viewport.width < 768;
    var isPortrait = viewport.height > viewport.width;
    var layoutKey = getMotionLayoutKey(viewport);
    var presetName = canvas.dataset.fieldPreset || 'quiet';
    var preset = getPresetConfig(presetName, isMobile, isPortrait);
    if (reduceMotion) {
      preset.rotationSpeed = 0;
      preset.breatheAmp = 0;
      preset.pulseRange = 0;
      preset.zoomRange = 0;
      preset.tiltLerp = 0;
      preset.opacityMobileBase *= 0.82;
      preset.opacityDesktopBase *= 0.82;
    }

    var renderer = new THREE.WebGLRenderer({
      canvas: canvas,
      alpha: true,
      antialias: !isMobile,
      powerPreference: 'high-performance'
    });
    renderer.setPixelRatio(getRendererPixelRatio(viewport.width));
    renderer.setClearColor(0x000000, 0);
    if (THREE.SRGBColorSpace) renderer.outputColorSpace = THREE.SRGBColorSpace;
    renderer.setSize(viewport.width, viewport.height, true);

    var scene = new THREE.Scene();
    scene.fog = new THREE.FogExp2(0xE3E2E0, preset.fogDensity * (18.5 / preset.initZ));
    scene.fog.color.setHex(preset.fogHex);

    var fieldGroup = new THREE.Group();
    scene.add(fieldGroup);

    var camera = new THREE.PerspectiveCamera(preset.initFov, viewport.width / viewport.height, 0.1, 1000);
    camera.position.set(0, 1.2, preset.initZ);
    camera.lookAt(0, 0, 0);

    var mouse = { x: 0, y: 0 };
    var moveEventName = window.PointerEvent ? 'pointermove' : 'mousemove';
    var frameId = 0;
    var resizeTimer = 0;
    var destroyed = false;
    var contextLost = false;

    function onPointerMove(e) {
      if (e.pointerType === 'touch') return;
      var pointerViewport = getViewportSize();
      mouse.x = (e.clientX / pointerViewport.width - 0.5) * 2;
      mouse.y = (e.clientY / pointerViewport.height - 0.5) * 2;
    }

    function resetPointer() {
      mouse.x = 0;
      mouse.y = 0;
    }

    document.addEventListener(moveEventName, onPointerMove, { passive: true });
    window.addEventListener('blur', resetPointer, { passive: true });

    var field;
    if (preset.mode === 'membrane') {
      field = createMembraneField(THREE, scene, fieldGroup, preset, isMobile);
    } else if (preset.mode === 'monolith') {
      field = createMonolithField(THREE, scene, fieldGroup, preset, isMobile);
    } else if (preset.mode === 'constellation') {
      field = createConstellationField(THREE, scene, fieldGroup, preset, isMobile);
    } else {
      field = createWaveField(THREE, fieldGroup, preset, isMobile);
    }

    var lastViewport = viewport;
    var cachedDocH = document.documentElement.scrollHeight - viewport.height;
    var scrollLerp = 0;
    var targetCamX = 1.2;
    var targetCamY = 0;
    var clock = new THREE.Clock();

    function syncViewport(nextViewport) {
      lastViewport = nextViewport;
      renderer.setPixelRatio(getRendererPixelRatio(nextViewport.width));
      renderer.setSize(nextViewport.width, nextViewport.height, true);
      camera.aspect = nextViewport.width / nextViewport.height;
      camera.updateProjectionMatrix();
      cachedDocH = document.documentElement.scrollHeight - nextViewport.height;
    }

    function onResize() {
      if (destroyed) return;

      var nextViewport = getViewportSize();
      var nextLayoutKey = getMotionLayoutKey(nextViewport);
      if (nextLayoutKey !== layoutKey) {
        queueRestart();
        return;
      }

      if (
        nextViewport.width === lastViewport.width &&
        Math.abs(nextViewport.height - lastViewport.height) / Math.max(lastViewport.height, 1) < 0.02
      ) {
        cachedDocH = document.documentElement.scrollHeight - nextViewport.height;
        return;
      }

      syncViewport(nextViewport);
    }

    function scheduleResize() {
      clearTimeout(resizeTimer);
      resizeTimer = window.setTimeout(onResize, 120);
    }

    function onContextLost(event) {
      event.preventDefault();
      contextLost = true;
      if (frameId) {
        window.cancelAnimationFrame(frameId);
        frameId = 0;
      }
    }

    function onContextRestored() {
      contextLost = false;
      queueRestart();
    }

    function onReducedMotionChange(event) {
      reduceMotion = event.matches;
      queueRestart();
    }

    function destroy() {
      if (destroyed) return;
      destroyed = true;
      delete canvas.dataset.fieldInit;
      clearTimeout(resizeTimer);
      if (frameId) window.cancelAnimationFrame(frameId);
      document.removeEventListener(moveEventName, onPointerMove);
      window.removeEventListener('blur', resetPointer);
      window.removeEventListener('resize', scheduleResize);
      if (window.visualViewport) {
        window.visualViewport.removeEventListener('resize', scheduleResize);
        window.visualViewport.removeEventListener('scroll', scheduleResize);
      }
      canvas.removeEventListener('webglcontextlost', onContextLost);
      canvas.removeEventListener('webglcontextrestored', onContextRestored);
      if (typeof reducedMotionQuery.removeEventListener === 'function') {
        reducedMotionQuery.removeEventListener('change', onReducedMotionChange);
      } else if (typeof reducedMotionQuery.removeListener === 'function') {
        reducedMotionQuery.removeListener(onReducedMotionChange);
      }
      scene.traverse(disposeObject);
      renderer.dispose();
      if (activeFieldRuntime && activeFieldRuntime.canvas === canvas) activeFieldRuntime = null;
    }

    window.addEventListener('resize', scheduleResize, { passive: true });
    if (window.visualViewport) {
      window.visualViewport.addEventListener('resize', scheduleResize, { passive: true });
      window.visualViewport.addEventListener('scroll', scheduleResize, { passive: true });
    }
    canvas.addEventListener('webglcontextlost', onContextLost, false);
    canvas.addEventListener('webglcontextrestored', onContextRestored, false);
    if (typeof reducedMotionQuery.addEventListener === 'function') {
      reducedMotionQuery.addEventListener('change', onReducedMotionChange);
    } else if (typeof reducedMotionQuery.addListener === 'function') {
      reducedMotionQuery.addListener(onReducedMotionChange);
    }

    activeFieldRuntime = { canvas: canvas, destroy: destroy };

    function getRawScroll() {
      return cachedDocH > 10 ? Math.min(1, Math.max(0, window.pageYOffset / cachedDocH)) : 0;
    }

    function animateWaveField(t) {
      fieldGroup.rotation.y = t * preset.rotationSpeed;
      var breathe = 1 + Math.sin(t * preset.breatheSpeed) * preset.breatheAmp;
      fieldGroup.scale.set(breathe, breathe, breathe);

      for (var fi = 0; fi < field.fieldLines.length; fi++) {
        var fl = field.fieldLines[fi];
        fl.mat.opacity = fl.baseOp * (preset.pulseMin + preset.pulseRange * Math.sin(t * preset.pulseSpeed + fl.phase));
        if (presetName === 'architect') {
          fl.mesh.rotation.z = Math.sin(t * 0.018 + fl.phase) * 0.01 * (1 - fl.depth);
        } else if (presetName === 'observer') {
          fl.mesh.rotation.z = Math.sin(t * 0.012 + fl.phase) * 0.004 * (1 - fl.depth);
          fl.mesh.position.y = Math.sin(t * 0.016 + fl.phase) * 0.018 * (1 - fl.depth);
        }
      }

      for (var gi = 0; gi < field.guideLines.length; gi++) {
        var guide = field.guideLines[gi];
        guide.mat.opacity = guide.baseOp * (0.92 + preset.guidePulse * Math.sin(t * 0.11 + gi));
        if (guide.mode === 'loop') {
          guide.mesh.rotation.x = Math.sin(t * guide.drift + gi) * (presetName === 'observer' ? 0.03 : 0.08);
          guide.mesh.rotation.z = Math.cos(t * guide.drift + gi) * (presetName === 'observer' ? 0.018 : 0.04);
          if (presetName === 'observer') guide.mesh.position.y = Math.sin(t * 0.02 + gi) * 0.03;
        } else {
          guide.mesh.rotation.y += guide.drift * 0.016;
        }
      }

      if (field.coreDust) {
        field.coreDust.mat.opacity = field.coreDust.baseOp * (0.84 + 0.16 * Math.sin(t * 0.34));
        if (presetName === 'observer') {
          field.coreDust.mesh.rotation.y = t * 0.01;
          field.coreDust.mesh.rotation.x = Math.sin(t * 0.06) * 0.08;
        }
      }
    }

    function animateMembraneField(t) {
      var shellGeo = field.shell.geometry;
      var pos = shellGeo.attributes.position.array;
      var base1 = field.base1;
      for (var i = 0; i < pos.length; i += 3) {
        var ox = base1[i];
        var oy = base1[i + 1];
        var oz = base1[i + 2];
        var drift = Math.sin(t * 0.55 + ox * 0.9 + oz * 0.75) * 0.08 + Math.cos(t * 0.42 + oy * 1.4) * 0.05;
        pos[i] = ox + (ox / preset.meshScale) * drift;
        pos[i + 1] = oy + (oy / preset.meshScale) * drift * 0.7;
        pos[i + 2] = oz + (oz / preset.meshScale) * drift;
      }
      shellGeo.attributes.position.needsUpdate = true;
      shellGeo.computeVertexNormals();

      var coreGeo = field.core.geometry;
      var pos2 = coreGeo.attributes.position.array;
      var base2 = field.base2;
      for (var j = 0; j < pos2.length; j += 3) {
        var ix = base2[j];
        var iy = base2[j + 1];
        var iz = base2[j + 2];
        var wave = Math.sin(t * 0.46 + ix * 1.2 + iz * 0.9) * 0.03;
        pos2[j] = ix + wave;
        pos2[j + 1] = iy + wave * 0.55;
        pos2[j + 2] = iz + wave;
      }
      coreGeo.attributes.position.needsUpdate = true;
      coreGeo.computeVertexNormals();

      field.shell.rotation.y = t * 0.07;
      field.shell.rotation.x = Math.sin(t * 0.12) * 0.12;
      field.core.rotation.y = -t * 0.05;
      field.core.rotation.z = Math.cos(t * 0.08) * 0.08;
    }

    function animateMonolithField(t) {
      field.body.rotation.y = Math.sin(t * 0.08) * 0.12;
      field.body.rotation.x = Math.sin(t * 0.06) * 0.03;
      field.body.position.y = Math.sin(t * 0.16) * 0.08;
      field.core.rotation.y = -Math.sin(t * 0.06) * 0.09;
      field.core.position.y = Math.sin(t * 0.16) * 0.06;
      field.halo.rotation.z = t * 0.035;
      field.halo.position.y = -0.08 + Math.sin(t * 0.1) * 0.03;
    }

    function animateConstellationField(t) {
      field.cloud.rotation.y = t * 0.03;
      field.cloud.rotation.x = Math.sin(t * 0.05) * 0.18;
      field.core.rotation.y = -t * 0.045;
      field.core.material.opacity = 0.22 + Math.sin(t * 0.42) * 0.04;
      field.cloud.material.opacity = preset.particleOpacity * (0.92 + Math.sin(t * 0.21) * 0.08);
    }

    function animate() {
      if (destroyed) return;
      if (!document.body.contains(canvas)) {
        destroy();
        return;
      }

      frameId = window.requestAnimationFrame(animate);
      if (contextLost || document.hidden) return;
      if (isNavOverlayOpen()) return;

      if (reduceMotion) {
        renderer.render(scene, camera);
        return;
      }

      var t = clock.getElapsedTime();
      scrollLerp += (getRawScroll() - scrollLerp) * 0.04;
      var targetZ = preset.initZ - scrollLerp * (preset.zoomRange * 0.62);
      camera.position.z += (targetZ - camera.position.z) * 0.06;
      camera.fov += (preset.initFov - camera.fov) * 0.06;
      camera.updateProjectionMatrix();

      if (!isMobile) {
        targetCamX += (1.2 + mouse.y * preset.tiltX - targetCamX) * preset.tiltLerp;
        targetCamY += (mouse.x * preset.tiltY - targetCamY) * preset.tiltLerp;
        camera.position.x = targetCamY;
        camera.position.y = targetCamX;
      }

      if (field.type === 'wave') animateWaveField(t);
      if (field.type === 'membrane') animateMembraneField(t);
      if (field.type === 'monolith') animateMonolithField(t);
      if (field.type === 'constellation') animateConstellationField(t);

      camera.lookAt(0, 0, 0);
      renderer.render(scene, camera);
    }

    animate();
  }

  init(0);
})();
