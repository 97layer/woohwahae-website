'use client';

import { useEffect, useRef, useState } from 'react';
import * as THREE from 'three';

import { formatRuntimeTimestamp } from './neural-graph-time.mjs';

const TONE = {
  core: { fill: 'rgba(250,250,250,0.96)', stroke: 'rgba(22,22,22,0.92)', glow: 'rgba(22,22,22,0.14)', ink: 'rgba(22,22,22,0.34)' },
  focus: { fill: 'rgba(243,243,243,0.18)', stroke: 'rgba(34,34,34,0.88)', glow: 'rgba(34,34,34,0.16)', ink: 'rgba(34,34,34,0.32)' },
  review: { fill: 'rgba(236,236,236,0.16)', stroke: 'rgba(56,56,56,0.82)', glow: 'rgba(56,56,56,0.14)', ink: 'rgba(56,56,56,0.28)' },
  proposal: { fill: 'rgba(240,240,240,0.14)', stroke: 'rgba(68,68,68,0.8)', glow: 'rgba(68,68,68,0.14)', ink: 'rgba(68,68,68,0.26)' },
  job: { fill: 'rgba(232,232,232,0.12)', stroke: 'rgba(82,82,82,0.8)', glow: 'rgba(82,82,82,0.14)', ink: 'rgba(82,82,82,0.28)' },
  observation: { fill: 'rgba(242,242,242,0.12)', stroke: 'rgba(96,96,96,0.78)', glow: 'rgba(96,96,96,0.14)', ink: 'rgba(96,96,96,0.24)' },
  verification: { fill: 'rgba(234,234,234,0.12)', stroke: 'rgba(108,108,108,0.78)', glow: 'rgba(108,108,108,0.14)', ink: 'rgba(108,108,108,0.22)' },
  alert: { fill: 'rgba(228,228,228,0.14)', stroke: 'rgba(18,18,18,0.9)', glow: 'rgba(18,18,18,0.18)', ink: 'rgba(18,18,18,0.34)' },
  good: { fill: 'rgba(244,244,244,0.12)', stroke: 'rgba(72,72,72,0.76)', glow: 'rgba(72,72,72,0.12)', ink: 'rgba(72,72,72,0.22)' },
  muted: { fill: 'rgba(236,236,236,0.1)', stroke: 'rgba(120,120,120,0.7)', glow: 'rgba(120,120,120,0.12)', ink: 'rgba(120,120,120,0.2)' },
};

const RING_ORDER = ['focus', 'risk', 'step', 'thread', 'review', 'proposal', 'job', 'verification', 'lesson', 'observation'];
const RING_INDEX = Object.fromEntries(RING_ORDER.map((kind, index) => [kind, index]));
const GRAPH_WIDTH = 1220;
const GRAPH_HEIGHT = 880;
const CENTER_X = GRAPH_WIDTH / 2;
const CENTER_Y = GRAPH_HEIGHT / 2;
const FRAME_INTERVAL_MS = 80;
const GRAPH_REFRESH_MS = 15000;

function clamp(value, min, max) {
  return Math.max(min, Math.min(max, value));
}

function hashString(value) {
  let hash = 2166136261;
  for (let i = 0; i < value.length; i += 1) {
    hash ^= value.charCodeAt(i);
    hash = Math.imul(hash, 16777619);
  }
  return Math.abs(hash >>> 0);
}

function wrapLabel(label, maxChars) {
  if (!label) return [''];
  const words = label.split(/\s+/);
  const lines = [];
  let current = '';
  words.forEach((word) => {
    const next = current ? `${current} ${word}` : word;
    if (next.length <= maxChars || current.length === 0) {
      current = next;
      return;
    }
    lines.push(current);
    current = word;
  });
  if (current) lines.push(current);
  return lines.slice(0, 3);
}

function edgePath(source, target) {
  const dx = target.x - source.x;
  const dy = target.y - source.y;
  const curve = clamp(Math.hypot(dx, dy) * 0.18, 24, 90);
  return `M ${source.x} ${source.y} C ${source.x} ${source.y + curve}, ${target.x} ${target.y - curve}, ${target.x} ${target.y}`;
}

function buildTargets(graph) {
  const targets = {};
  const grouped = new Map();
  graph.nodes.forEach((node) => {
    const group = grouped.get(node.kind) || [];
    group.push(node);
    grouped.set(node.kind, group);
  });

  graph.nodes.forEach((node) => {
    if (node.id === 'root') {
      targets[node.id] = { x: CENTER_X, y: CENTER_Y };
      return;
    }
    if (node.kind === 'focus') {
      targets[node.id] = { x: CENTER_X, y: CENTER_Y - 116 };
      return;
    }
    const list = grouped.get(node.kind) || [];
    const index = Math.max(list.findIndex((item) => item.id === node.id), 0);
    const ring = (RING_INDEX[node.kind] ?? RING_ORDER.length) + 1;
    const baseRadius = 122 + ring * 36;
    const angleStep = (Math.PI * 2) / Math.max(list.length, 1);
    const seed = hashString(node.id) % 1000;
    const angle = angleStep * index + seed * 0.0008;
    const stretch = 0.44 + ((seed % 7) * 0.022);
    const pad = node.radius + 70;
    targets[node.id] = {
      x: clamp(CENTER_X + Math.cos(angle) * baseRadius, pad, GRAPH_WIDTH - pad),
      y: clamp(CENTER_Y + Math.sin(angle) * (baseRadius * stretch), pad, GRAPH_HEIGHT - pad),
    };
  });

  return targets;
}

function initializePositions(graph, previous) {
  const targets = buildTargets(graph);
  const next = {};
  graph.nodes.forEach((node) => {
    const target = targets[node.id];
    const prior = previous[node.id];
    const seed = hashString(node.id);
    next[node.id] = {
      x: prior?.x ?? target.x + ((seed % 23) - 11),
      y: prior?.y ?? target.y + (((seed >> 4) % 23) - 11),
      targetX: target.x,
      targetY: target.y,
      vx: 0,
      vy: 0,
      driftPhase: (seed % 360) * (Math.PI / 180),
      driftRadius: 2 + ((seed >> 3) % 7) * 0.9,
    };
  });
  return next;
}

function buildDegrees(graph) {
  const degrees = {};
  graph.nodes.forEach((node) => {
    degrees[node.id] = 0;
  });
  graph.edges.forEach((edge) => {
    degrees[edge.source] = (degrees[edge.source] || 0) + 1;
    degrees[edge.target] = (degrees[edge.target] || 0) + 1;
  });
  return degrees;
}

function buildTraces(graph) {
  const degrees = buildDegrees(graph);
  const traces = [];

  graph.nodes.forEach((node) => {
    const degree = degrees[node.id] || 1;
    const seed = hashString(node.id);
    const count = clamp(degree * 5 + (node.status === 'failed' ? 10 : 6), 8, 28);
    for (let index = 0; index < count; index += 1) {
      traces.push({
        id: `${node.id}:trace:${index}`,
        nodeId: node.id,
        tone: node.tone,
        size: 1.2 + ((seed + index) % 4) * 0.55,
        orbit: node.radius + 18 + ((seed + index * 17) % 78),
        twist: 0.0011 + (((seed >> 3) + index) % 11) * 0.00024,
        branch: 0.78 + (((seed >> 6) + index) % 9) * 0.12,
        phase: ((seed % 360) * Math.PI) / 180 + index * 0.41,
        membrane: 0.4 + (((seed >> 2) + index) % 5) * 0.18,
        wobble: 0.7 + (((seed >> 5) + index) % 8) * 0.16,
      });
    }
  });

  return traces;
}

function buildPulses(graph) {
  const nodeMap = Object.fromEntries(graph.nodes.map((node) => [node.id, node]));
  const pulses = [];

  graph.edges.forEach((edge, index) => {
    const source = nodeMap[edge.source];
    const target = nodeMap[edge.target];
    const sourceSeed = hashString(edge.source);
    const targetSeed = hashString(edge.target);
    const count = clamp(Math.round(edge.strength * 1.4) + (source?.status === 'failed' || target?.status === 'failed' ? 1 : 0), 1, 2);
    for (let pulse = 0; pulse < count; pulse += 1) {
      pulses.push({
        id: `${edge.id}:pulse:${pulse}`,
        sourceId: edge.source,
        targetId: edge.target,
        edgeKind: edge.kind,
        tone: target?.tone || source?.tone || 'muted',
        speed: 0.05 + (((sourceSeed + targetSeed + pulse + index) % 7) * 0.012),
        offset: (((sourceSeed >> 2) + targetSeed + pulse * 37) % 100) / 100,
        width: edge.strength > 1 ? 2.2 : 1.4,
      });
    }
  });

  return pulses;
}

function updateTraces(traces, positions, now) {
  traces.forEach((trace) => {
    const anchor = positions[trace.nodeId];
    if (!anchor) return;
    const angle = trace.phase + now * trace.twist;
    const orbit = trace.orbit + Math.sin(now * trace.twist * 3 + trace.phase) * 9;
    trace.x = anchor.x + Math.cos(angle) * orbit * trace.branch;
    trace.y = anchor.y + Math.sin(angle) * orbit * 0.58;
  });
}

function pointOnCurve(source, target, t) {
  const controlX = (source.x + target.x) / 2;
  const controlY = ((source.y + target.y) / 2) - clamp(Math.hypot(target.x - source.x, target.y - source.y) * 0.08, 16, 42);
  const oneMinus = 1 - t;
  return {
    x: oneMinus * oneMinus * source.x + 2 * oneMinus * t * controlX + t * t * target.x,
    y: oneMinus * oneMinus * source.y + 2 * oneMinus * t * controlY + t * t * target.y,
  };
}

function updatePulses(pulses, positions, now) {
  pulses.forEach((pulse) => {
    const source = positions[pulse.sourceId];
    const target = positions[pulse.targetId];
    if (!source || !target) return;
    const progress = (pulse.offset + now * 0.00012 * pulse.speed * 60) % 1;
    const point = pointOnCurve(source, target, progress);
    pulse.x = point.x;
    pulse.y = point.y;
    pulse.progress = progress;
  });
}

function organicDrift(entry, now, intensity = 1) {
  const driftX = Math.sin(now * 0.00016 + entry.driftPhase) * entry.driftRadius * intensity;
  const driftY = Math.cos(now * 0.00012 + entry.driftPhase * 1.3) * entry.driftRadius * 0.72 * intensity;
  return { driftX, driftY };
}

function confineNode(entry, node) {
  const padX = node.radius + 58;
  const padY = node.radius + 54;
  entry.x = clamp(entry.x, padX, GRAPH_WIDTH - padX);
  entry.y = clamp(entry.y, padY, GRAPH_HEIGHT - padY);
}

function computeTissueLoad(graph, traces, pulses) {
  const nodeWeight = graph.nodes.length * 0.9;
  const edgeWeight = graph.edges.length * 1.15;
  const traceWeight = traces.length * 0.035;
  const pulseWeight = pulses.length * 0.18;
  const reviewWeight = (graph.metrics?.reviewOpen || 0) * 0.6;
  const riskWeight = (graph.metrics?.risks || 0) * 0.75;
  return clamp(nodeWeight + edgeWeight + traceWeight + pulseWeight + reviewWeight + riskWeight, 12, 120);
}

function drawTissueBloom(context, load, now) {
  const intensity = clamp(load / 120, 0.16, 1);
  const bloomRadius = 140 + intensity * 240;
  const verticalStretch = 0.58 + intensity * 0.12;

  const bloom = context.createRadialGradient(
    CENTER_X,
    CENTER_Y - 12,
    18,
    CENTER_X,
    CENTER_Y - 12,
    bloomRadius,
  );
  bloom.addColorStop(0, `rgba(18,18,18,${0.035 + intensity * 0.08})`);
  bloom.addColorStop(0.42, `rgba(44,44,44,${0.024 + intensity * 0.042})`);
  bloom.addColorStop(0.82, `rgba(88,88,88,${0.012 + intensity * 0.018})`);
  bloom.addColorStop(1, 'rgba(255,255,255,0)');
  context.fillStyle = bloom;
  context.beginPath();
  context.ellipse(CENTER_X, CENTER_Y - 8, bloomRadius, bloomRadius * verticalStretch, 0, 0, Math.PI * 2);
  context.fill();

  const stainCount = 3;
  for (let index = 0; index < stainCount; index += 1) {
    const phase = now * 0.00008 + index * 1.7;
    const offsetX = Math.sin(phase) * (24 + intensity * 34);
    const offsetY = Math.cos(phase * 1.2) * (18 + intensity * 22);
    const stainRadius = 72 + intensity * 92 + index * 18;
    context.fillStyle = `rgba(28,28,28,${0.018 + intensity * 0.016})`;
    context.beginPath();
    for (let point = 0; point <= 18; point += 1) {
      const theta = (Math.PI * 2 * point) / 18;
      const wobble = 1 + Math.sin(theta * 3 + phase * 6) * (0.08 + intensity * 0.05);
      const radius = stainRadius * wobble;
      const px = CENTER_X + offsetX + Math.cos(theta) * radius;
      const py = CENTER_Y + offsetY + Math.sin(theta) * radius * (0.52 + intensity * 0.08);
      if (point === 0) context.moveTo(px, py);
      else context.lineTo(px, py);
    }
    context.closePath();
    context.fill();
  }
}

function alpha(value, amount) {
  return value.replace(/rgba\(([^)]+),\s*[\d.]+\)/, `rgba($1, ${amount})`);
}

function drawOrganicSoma(context, x, y, radius, tone, active, now) {
  const wobble = active ? 1.8 : 1.2;
  context.save();
  context.translate(x, y);
  context.rotate(Math.sin(now * 0.00045 + radius) * 0.08);

  const halo = context.createRadialGradient(0, 0, radius * 0.2, 0, 0, radius * 2.4);
  halo.addColorStop(0, alpha(tone.glow, active ? 0.42 : 0.26));
  halo.addColorStop(0.7, alpha(tone.glow, active ? 0.12 : 0.06));
  halo.addColorStop(1, alpha(tone.glow, 0));
  context.fillStyle = halo;
  context.beginPath();
  context.arc(0, 0, radius * 2.45, 0, Math.PI * 2);
  context.fill();

  context.fillStyle = alpha(tone.fill, active ? 0.88 : 0.72);
  context.strokeStyle = alpha(tone.stroke, active ? 0.9 : 0.58);
  context.lineWidth = active ? 2 : 1.2;
  context.beginPath();
  for (let index = 0; index <= 14; index += 1) {
    const theta = (Math.PI * 2 * index) / 14;
    const pulse = 1 + Math.sin(now * 0.0015 + theta * 3.2 + radius) * 0.06 * wobble;
    const pointRadius = radius * pulse;
    const px = Math.cos(theta) * pointRadius;
    const py = Math.sin(theta) * pointRadius * 0.88;
    if (index === 0) context.moveTo(px, py);
    else context.lineTo(px, py);
  }
  context.closePath();
  context.fill();
  context.stroke();

  context.fillStyle = alpha(tone.stroke, active ? 0.92 : 0.72);
  context.beginPath();
  context.ellipse(0, 0, radius * 0.38, radius * 0.32, 0, 0, Math.PI * 2);
  context.fill();

  context.fillStyle = active ? 'rgba(255,255,255,0.34)' : 'rgba(255,255,255,0.18)';
  context.beginPath();
  context.ellipse(radius * -0.18, radius * -0.16, radius * 0.2, radius * 0.12, -0.5, 0, Math.PI * 2);
  context.fill();
  context.restore();
}

function drawBranchBundle(context, source, target, tone, strength, active, now, seed) {
  const dx = target.x - source.x;
  const dy = target.y - source.y;
  const distance = Math.max(Math.hypot(dx, dy), 1);
  const nx = -dy / distance;
  const ny = dx / distance;
  const fibers = clamp(Math.round(strength * 3) + (active ? 2 : 1), 2, 6);

  for (let fiber = 0; fiber < fibers; fiber += 1) {
    const jitter = ((seed + fiber * 17) % 9) - 4;
    const drift = jitter * 2.2;
    const bend = clamp(distance * 0.12, 18, 64) + Math.sin(now * 0.001 + fiber) * 6;
    const controlX = (source.x + target.x) / 2 + nx * drift;
    const controlY = (source.y + target.y) / 2 - bend + ny * drift * 0.4;
    context.strokeStyle = alpha(tone.ink, active ? 0.34 : 0.1 + fiber * 0.018);
    context.lineWidth = active ? 1.5 - fiber * 0.12 : 0.9 - fiber * 0.08;
    context.beginPath();
    context.moveTo(source.x + nx * drift * 0.2, source.y + ny * drift * 0.2);
    context.quadraticCurveTo(controlX, controlY, target.x - nx * drift * 0.2, target.y - ny * drift * 0.2);
    context.stroke();
  }
}

function drawLivingField(canvas, graph, positions, activeNodeId, traces, pulses, now) {
  const context = canvas.getContext('2d');
  if (!context) return;

  context.clearRect(0, 0, GRAPH_WIDTH, GRAPH_HEIGHT);
  const tissueLoad = computeTissueLoad(graph, traces, pulses);

  const fieldGradient = context.createRadialGradient(CENTER_X, CENTER_Y, 20, CENTER_X, CENTER_Y, GRAPH_WIDTH * 0.52);
  fieldGradient.addColorStop(0, 'rgba(18,18,18,0.18)');
  fieldGradient.addColorStop(0.24, 'rgba(72,72,72,0.09)');
  fieldGradient.addColorStop(0.54, 'rgba(118,118,118,0.06)');
  fieldGradient.addColorStop(0.8, 'rgba(180,180,180,0.04)');
  fieldGradient.addColorStop(1, 'rgba(255,255,255,0)');
  context.fillStyle = fieldGradient;
  context.fillRect(0, 0, GRAPH_WIDTH, GRAPH_HEIGHT);
  drawTissueBloom(context, tissueLoad, now);

  const nodeMap = Object.fromEntries(graph.nodes.map((node) => [node.id, node]));

  context.save();
  graph.edges.forEach((edge, index) => {
    const source = positions[edge.source];
    const target = positions[edge.target];
    if (!source || !target) return;
    const active = edge.source === activeNodeId || edge.target === activeNodeId;
    const sourceTone = TONE[nodeMap[edge.source]?.tone] || TONE.muted;
    drawBranchBundle(context, source, target, sourceTone, edge.strength, active, now, index * 13 + hashString(edge.id));
  });
  context.restore();

  context.save();
  traces.forEach((trace) => {
    const anchor = positions[trace.nodeId];
    if (!anchor || !Number.isFinite(trace.x) || !Number.isFinite(trace.y)) return;
    const active = trace.nodeId === activeNodeId;
    const tone = TONE[trace.tone] || TONE.muted;
    const bend = ((hashString(trace.id) % 21) - 10) * 0.9;
    const controlX = (anchor.x + trace.x) / 2 + bend;
    const controlY = (anchor.y + trace.y) / 2 - 8 - bend * 0.3;
    context.strokeStyle = active ? alpha(tone.ink, 0.18) : alpha(tone.ink, 0.08);
    context.lineWidth = active ? 0.85 : 0.45;
    context.beginPath();
    context.moveTo(anchor.x, anchor.y);
    context.quadraticCurveTo(controlX, controlY, trace.x, trace.y);
    context.stroke();

    const membrane = trace.size * (2.1 + Math.sin(now * 0.0018 + trace.phase) * 0.25 * trace.membrane);
    context.fillStyle = alpha(tone.fill, active ? 0.32 : 0.16);
    context.beginPath();
    context.ellipse(trace.x, trace.y, membrane, membrane * (0.72 + trace.wobble * 0.03), trace.phase, 0, Math.PI * 2);
    context.fill();
    context.strokeStyle = alpha(tone.stroke, active ? 0.34 : 0.18);
    context.lineWidth = 0.7;
    context.stroke();

    context.fillStyle = alpha(tone.stroke, active ? 0.82 : 0.52);
    context.beginPath();
    context.arc(trace.x, trace.y, trace.size, 0, Math.PI * 2);
    context.fill();
  });
  context.restore();

  context.save();
  pulses.forEach((pulse) => {
    if (!Number.isFinite(pulse.x) || !Number.isFinite(pulse.y)) return;
    const tone = TONE[pulse.tone] || TONE.muted;
    const shimmer = 0.72 + Math.sin(now * 0.004 + pulse.progress * Math.PI * 2) * 0.18;
    context.fillStyle = tone.stroke.replace(/0\.[0-9]+\)/, `${clamp(shimmer, 0.42, 0.92)})`);
    context.beginPath();
    context.arc(pulse.x, pulse.y, pulse.width, 0, Math.PI * 2);
    context.fill();
  });
  context.restore();

  context.save();
  graph.nodes.forEach((node) => {
    const point = positions[node.id];
    if (!point) return;
    drawOrganicSoma(context, point.x, point.y, node.radius, TONE[node.tone] || TONE.muted, node.id === activeNodeId, now);
  });
  context.restore();

  context.save();
  graph.nodes.forEach((node) => {
    const point = positions[node.id];
    if (!point) return;
    const active = node.id === activeNodeId;
    const showLabel = active || node.kind === 'root' || node.kind === 'focus';
    if (!showLabel) return;
    const lines = wrapLabel(node.label, node.kind === 'root' ? 14 : 22);
    context.font = `${node.kind === 'root' ? '11px' : '10px'} var(--font-body)`;
    context.textAlign = 'center';
    context.textBaseline = 'middle';
    context.lineJoin = 'round';
    context.strokeStyle = active ? 'rgba(243,241,237,0.58)' : 'rgba(243,241,237,0.34)';
    context.fillStyle = active ? 'rgba(20,20,20,0.78)' : 'rgba(20,20,20,0.52)';
    context.lineWidth = 3;
    lines.forEach((line, index) => {
      const y = point.y + (index - (lines.length - 1) / 2) * 12;
      context.strokeText(line, point.x, y);
      context.fillText(line, point.x, y);
    });
  });
  context.restore();
}

const THREE_TONE = {
  core: '#161616',
  focus: '#2a2a2a',
  review: '#3a3a3a',
  proposal: '#464646',
  job: '#555555',
  observation: '#646464',
  verification: '#727272',
  alert: '#141414',
  good: '#5c5c5c',
  muted: '#7b7b7b',
};

function hexToRgbArray(hex) {
  const color = new THREE.Color(hex);
  return [color.r, color.g, color.b];
}

function createThreeField(container) {
  const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
  renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
  renderer.setClearColor(0x000000, 0);
  renderer.setSize(container.clientWidth || 1, container.clientHeight || 1, false);
  container.innerHTML = '';
  container.appendChild(renderer.domElement);

  const scene = new THREE.Scene();
  const camera = new THREE.OrthographicCamera(0, GRAPH_WIDTH, GRAPH_HEIGHT, 0, -1000, 1000);
  camera.position.z = 50;

  const edgeGeometry = new THREE.BufferGeometry();
  const edgeMaterial = new THREE.LineBasicMaterial({ transparent: true, opacity: 0.16, vertexColors: true });
  const edgeLines = new THREE.LineSegments(edgeGeometry, edgeMaterial);
  scene.add(edgeLines);

  const traceGeometry = new THREE.BufferGeometry();
  const traceMaterial = new THREE.PointsMaterial({
    size: 2.2,
    transparent: true,
    opacity: 0.82,
    vertexColors: true,
    sizeAttenuation: false,
  });
  const tracePoints = new THREE.Points(traceGeometry, traceMaterial);
  scene.add(tracePoints);

  const pulseGeometry = new THREE.BufferGeometry();
  const pulseMaterial = new THREE.PointsMaterial({
    size: 4.4,
    transparent: true,
    opacity: 0.95,
    vertexColors: true,
    sizeAttenuation: false,
  });
  const pulsePoints = new THREE.Points(pulseGeometry, pulseMaterial);
  scene.add(pulsePoints);

  return {
    renderer,
    scene,
    camera,
    edgeGeometry,
    traceGeometry,
    pulseGeometry,
    resize() {
      renderer.setSize(container.clientWidth || 1, container.clientHeight || 1, false);
    },
    dispose() {
      edgeGeometry.dispose();
      edgeMaterial.dispose();
      traceGeometry.dispose();
      traceMaterial.dispose();
      pulseGeometry.dispose();
      pulseMaterial.dispose();
      renderer.dispose();
      if (renderer.domElement.parentNode === container) {
        container.removeChild(renderer.domElement);
      }
    },
  };
}

function updateThreeField(field, graph, positions, activeNodeId, traces, pulses) {
  if (!field) return;

  const edgePositions = new Float32Array(graph.edges.length * 2 * 3);
  const edgeColors = new Float32Array(graph.edges.length * 2 * 3);
  graph.edges.forEach((edge, index) => {
    const source = positions[edge.source];
    const target = positions[edge.target];
    if (!source || !target) return;
    const base = index * 6;
    edgePositions[base] = source.x;
    edgePositions[base + 1] = GRAPH_HEIGHT - source.y;
    edgePositions[base + 2] = 0;
    edgePositions[base + 3] = target.x;
    edgePositions[base + 4] = GRAPH_HEIGHT - target.y;
    edgePositions[base + 5] = 0;
    const color = hexToRgbArray((edge.source === activeNodeId || edge.target === activeNodeId) ? '#0f172a' : '#94a3b8');
    edgeColors.set(color, base);
    edgeColors.set(color, base + 3);
  });
  field.edgeGeometry.setAttribute('position', new THREE.BufferAttribute(edgePositions, 3));
  field.edgeGeometry.setAttribute('color', new THREE.BufferAttribute(edgeColors, 3));
  field.edgeGeometry.computeBoundingSphere();

  const tracePositions = new Float32Array(traces.length * 3);
  const traceColors = new Float32Array(traces.length * 3);
  traces.forEach((trace, index) => {
    const base = index * 3;
    tracePositions[base] = trace.x || 0;
    tracePositions[base + 1] = GRAPH_HEIGHT - (trace.y || 0);
    tracePositions[base + 2] = 0;
    const color = hexToRgbArray(THREE_TONE[trace.tone] || THREE_TONE.muted);
    traceColors.set(color, base);
  });
  field.traceGeometry.setAttribute('position', new THREE.BufferAttribute(tracePositions, 3));
  field.traceGeometry.setAttribute('color', new THREE.BufferAttribute(traceColors, 3));
  field.traceGeometry.computeBoundingSphere();

  const pulsePositions = new Float32Array(pulses.length * 3);
  const pulseColors = new Float32Array(pulses.length * 3);
  pulses.forEach((pulse, index) => {
    const base = index * 3;
    pulsePositions[base] = pulse.x || 0;
    pulsePositions[base + 1] = GRAPH_HEIGHT - (pulse.y || 0);
    pulsePositions[base + 2] = 0;
    const color = hexToRgbArray(THREE_TONE[pulse.tone] || THREE_TONE.muted);
    pulseColors.set(color, base);
  });
  field.pulseGeometry.setAttribute('position', new THREE.BufferAttribute(pulsePositions, 3));
  field.pulseGeometry.setAttribute('color', new THREE.BufferAttribute(pulseColors, 3));
  field.pulseGeometry.computeBoundingSphere();

  field.renderer.render(field.scene, field.camera);
}

async function readGraph() {
  const response = await fetch('/api/public/neural-map', { cache: 'no-store' });
  const payload = await response.json();
  if (!response.ok) {
    throw new Error(payload?.error || 'graph unavailable');
  }
  return payload.graph;
}

export default function NeuralGraphDemo({ initialGraph }) {
  const [graph, setGraph] = useState(initialGraph);
  const [error, setError] = useState('');
  const [activeNodeId, setActiveNodeId] = useState('focus:current');
  const [liveState, setLiveState] = useState('live');
  const positionsRef = useRef(initializePositions(initialGraph, {}));
  const canvasRef = useRef(null);
  const threeHostRef = useRef(null);
  const threeFieldRef = useRef(null);
  const tracesRef = useRef(buildTraces(initialGraph));
  const pulsesRef = useRef(buildPulses(initialGraph));
  const lastFrameAtRef = useRef(0);

  useEffect(() => {
    positionsRef.current = initializePositions(graph, positionsRef.current);
    tracesRef.current = buildTraces(graph);
    pulsesRef.current = buildPulses(graph);
  }, [graph]);

  useEffect(() => {
    if (!threeHostRef.current) return undefined;
    const field = createThreeField(threeHostRef.current);
    threeFieldRef.current = field;

    function onResize() {
      field.resize();
    }

    window.addEventListener('resize', onResize);
    onResize();
    return () => {
      window.removeEventListener('resize', onResize);
      field.dispose();
      threeFieldRef.current = null;
    };
  }, []);

  useEffect(() => {
    if (!canvasRef.current) return undefined;
    const canvas = canvasRef.current;

    function pickNode(event) {
      const rect = canvas.getBoundingClientRect();
      if (!rect.width || !rect.height) return;
      const scaleX = GRAPH_WIDTH / rect.width;
      const scaleY = GRAPH_HEIGHT / rect.height;
      const x = (event.clientX - rect.left) * scaleX;
      const y = (event.clientY - rect.top) * scaleY;
      let bestId = null;
      let bestDist = Infinity;

      graph.nodes.forEach((node) => {
        const point = positionsRef.current[node.id];
        if (!point) return;
        const dist = Math.hypot(point.x - x, point.y - y);
        const hitRadius = node.radius + 18;
        if (dist <= hitRadius && dist < bestDist) {
          bestDist = dist;
          bestId = node.id;
        }
      });

      if (bestId) setActiveNodeId(bestId);
    }

    canvas.addEventListener('click', pickNode);
    return () => canvas.removeEventListener('click', pickNode);
  }, [graph]);

  useEffect(() => {
    let mounted = true;
    let timeoutId;

    async function refresh() {
      try {
        const next = await readGraph();
        if (!mounted) return;
        setGraph(next);
        setLiveState('live');
        setError('');
      } catch (nextError) {
        if (!mounted) return;
        setLiveState('degraded');
        setError(nextError instanceof Error ? nextError.message : 'runtime unavailable');
      } finally {
        if (mounted) timeoutId = window.setTimeout(refresh, GRAPH_REFRESH_MS);
      }
    }

    timeoutId = window.setTimeout(refresh, GRAPH_REFRESH_MS);
    return () => {
      mounted = false;
      if (timeoutId) window.clearTimeout(timeoutId);
    };
  }, []);

  useEffect(() => {
    let frameId;

    function tick(now) {
      if (document.visibilityState === 'hidden') {
        frameId = window.requestAnimationFrame(tick);
        return;
      }
      if (now - lastFrameAtRef.current < FRAME_INTERVAL_MS) {
        frameId = window.requestAnimationFrame(tick);
        return;
      }
      lastFrameAtRef.current = now;

      const next = { ...positionsRef.current };
      graph.nodes.forEach((node) => {
        const entry = next[node.id];
        if (!entry) return;
        const drift = organicDrift(entry, now, node.id === 'root' ? 0.35 : node.kind === 'focus' ? 0.55 : 1);
        const targetX = entry.targetX + drift.driftX;
        const targetY = entry.targetY + drift.driftY;
        const spring = node.id === 'root' ? 0.018 : node.kind === 'focus' ? 0.028 : node.status === 'failed' ? 0.034 : 0.022;
        const damping = node.id === 'root' ? 0.82 : node.kind === 'focus' ? 0.84 : 0.8;
        entry.vx = (entry.vx + (targetX - entry.x) * spring) * damping;
        entry.vy = (entry.vy + (targetY - entry.y) * spring) * damping;
        if (node.status === 'failed') {
          entry.vx += Math.sin(now * 0.0012 + hashString(node.id)) * 0.006;
          entry.vy += Math.cos(now * 0.001 + hashString(node.id)) * 0.006;
        }
        entry.x += entry.vx;
        entry.y += entry.vy;
      });

      for (let i = 0; i < graph.nodes.length; i += 1) {
        for (let j = i + 1; j < graph.nodes.length; j += 1) {
          const leftNode = graph.nodes[i];
          const rightNode = graph.nodes[j];
          const left = next[leftNode.id];
          const right = next[rightNode.id];
          if (!left || !right) continue;
          const dx = right.x - left.x;
          const dy = right.y - left.y;
          const dist = Math.max(Math.hypot(dx, dy), 1);
          const minDist = leftNode.radius + rightNode.radius + 24;
          if (dist >= minDist) continue;
          const push = (minDist - dist) * 0.0048;
          const offsetX = (dx / dist) * push;
          const offsetY = (dy / dist) * push;
          left.x -= offsetX;
          left.y -= offsetY;
          right.x += offsetX;
          right.y += offsetY;
        }
      }

      graph.nodes.forEach((node) => {
        const entry = next[node.id];
        if (!entry) return;
        confineNode(entry, node);
      });

      positionsRef.current = next;
      updateTraces(tracesRef.current, next, now);
      updatePulses(pulsesRef.current, next, now);
      updateThreeField(threeFieldRef.current, graph, next, activeNodeId, tracesRef.current, pulsesRef.current);
      if (canvasRef.current) {
        drawLivingField(canvasRef.current, graph, next, activeNodeId, tracesRef.current, pulsesRef.current, now);
      }
      frameId = window.requestAnimationFrame(tick);
    }

    frameId = window.requestAnimationFrame(tick);
    return () => window.cancelAnimationFrame(frameId);
  }, [activeNodeId, graph]);

  const activeNode = graph.nodes.find((node) => node.id === activeNodeId) || graph.nodes[0];
  const positions = positionsRef.current;
  const estimatedCells = tracesRef.current.length;
  const synapseCount = pulsesRef.current.length;
  const runtimeDescriptor = graph.runtimeAvailable ? 'Live runtime tissue' : 'Synthetic rehearsal';
  const runtimeNote = graph.runtimeAvailable
    ? `${graph.sourceLabel || 'live runtime'} signal map`
    : graph.degradedReason || 'daemon unavailable';

  return (
    <main className="neural-page">
      <section className="neural-hero">
        <div>
          <p className="neural-kicker">Field Lab / Live Runtime Topology</p>
          <h1>Layer OS Neural Field</h1>
          <p className="neural-copy">
            이제 단순 점과 선이 아니라, 각 런타임 객체가 세포체와 주변 조직으로 번지며 엮이도록 바꿨습니다.
            살아 있는 것은 장식이 아니라 review-room, proposal, job, verification, observation, risk, step의 실제 결속입니다.
          </p>
        </div>
        <div className="neural-status-card">
          <span className={`neural-live-pill is-${liveState}`}>{liveState === 'live' ? 'Live' : 'Degraded'}</span>
          <strong className="neural-status-title">{runtimeDescriptor}</strong>
          <p>Updated {formatRuntimeTimestamp(graph.generatedAt)}</p>
          <p className="neural-status-note">{runtimeNote}</p>
          <dl>
            <div><dt>Open agenda</dt><dd>{graph.metrics.reviewOpen}</dd></div>
            <div><dt>Failed jobs</dt><dd>{graph.metrics.failedJobs}</dd></div>
            <div><dt>Open risks</dt><dd>{graph.metrics.risks}</dd></div>
            <div><dt>Next steps</dt><dd>{graph.metrics.steps}</dd></div>
            <div><dt>Cells</dt><dd>{estimatedCells}</dd></div>
            <div><dt>Synapses</dt><dd>{synapseCount}</dd></div>
          </dl>
        </div>
      </section>

      <section className="neural-stage">
        <div className="neural-canvas-wrap">
          <div ref={threeHostRef} className="neural-three-layer" aria-hidden="true" />
          <canvas
            ref={canvasRef}
            className="neural-particle-layer"
            width={GRAPH_WIDTH}
            height={GRAPH_HEIGHT}
            role="img"
            aria-label="Layer OS live graph"
          />
        </div>

        <aside className="neural-sidecar">
          <section className="neural-panel">
            <p className="neural-panel-kicker">Focused node</p>
            <h2>{activeNode?.label}</h2>
            <p>{activeNode?.meta || 'meta unavailable'}</p>
            <dl className="neural-node-meta">
              <div><dt>Kind</dt><dd>{activeNode?.kind}</dd></div>
              <div><dt>Status</dt><dd>{activeNode?.status}</dd></div>
              <div><dt>Ref</dt><dd>{activeNode?.ref || 'none'}</dd></div>
            </dl>
          </section>

          <section className="neural-panel">
            <p className="neural-panel-kicker">Live topology</p>
            <ul className="neural-list">
              {graph.nodes.slice(0, 10).map((node) => (
                <li key={node.id}>
                  <button type="button" onClick={() => setActiveNodeId(node.id)}>
                    <strong>{node.kind}</strong>
                    <span>{node.label}</span>
                  </button>
                </li>
              ))}
            </ul>
          </section>

          <section className="neural-panel">
            <p className="neural-panel-kicker">Read model</p>
            <p>
              가지 섬유는 실제 edge를 따라 묶음으로 자라고, 주변 세포 군집은 node degree에 따라 밀도가 달라집니다.
              선 위를 흐르는 pulse는 실제 관계를 따라 순환하며, 데몬이 없을 때도 synthetic rehearsal로 구조 감각을 유지합니다.
            </p>
            {error ? <p className="neural-error">{error}</p> : null}
          </section>
        </aside>
      </section>

      <style jsx>{`
        .neural-page {
          min-height: 100vh;
          padding: 2rem 1.25rem 2.5rem;
          background:
            radial-gradient(circle at 18% 12%, rgba(22, 22, 22, 0.08), transparent 26%),
            radial-gradient(circle at 82% 18%, rgba(90, 90, 90, 0.05), transparent 24%),
            linear-gradient(180deg, #eceae6 0%, #f3f1ed 46%, #e8e5df 100%);
          color: #171717;
        }

        .neural-hero,
        .neural-stage {
          max-width: 1440px;
          margin: 0 auto;
        }

        .neural-hero {
          display: grid;
          grid-template-columns: minmax(0, 1.2fr) minmax(280px, 360px);
          gap: 1rem;
          align-items: end;
          margin-bottom: 1rem;
        }

        .neural-kicker,
        .neural-panel-kicker {
          margin: 0 0 0.5rem;
          font-family: var(--font-mono);
          font-size: 0.7rem;
          letter-spacing: 0.14em;
          text-transform: uppercase;
          color: rgba(23, 23, 23, 0.54);
        }

        h1 {
          margin: 0;
          font-size: clamp(2.3rem, 5vw, 4.8rem);
          line-height: 0.94;
          letter-spacing: -0.05em;
          font-weight: 600;
        }

        .neural-copy {
          max-width: 60ch;
          margin: 0.95rem 0 0;
          line-height: 1.6;
          color: rgba(23, 23, 23, 0.68);
        }

        .neural-status-card,
        .neural-panel {
          border: 1px solid rgba(23, 23, 23, 0.08);
          background: rgba(244, 242, 238, 0.72);
          box-shadow: 0 18px 60px rgba(23, 23, 23, 0.06);
          backdrop-filter: blur(14px);
        }

        .neural-status-card {
          padding: 1rem 1rem 0.85rem;
          border-radius: 24px;
        }

        .neural-status-card p {
          margin: 0.5rem 0 0.3rem;
          color: rgba(23, 23, 23, 0.58);
        }

        .neural-status-title {
          display: block;
          margin-top: 0.8rem;
          font-size: 1.1rem;
          line-height: 1.2;
        }

        .neural-status-note {
          font-size: 0.86rem;
          color: rgba(23, 23, 23, 0.48) !important;
          margin-bottom: 0.8rem !important;
        }

        .neural-status-card dl,
        .neural-node-meta {
          margin: 0;
          display: grid;
          grid-template-columns: repeat(2, minmax(0, 1fr));
          gap: 0.7rem;
        }

        .neural-status-card div,
        .neural-node-meta div {
          display: grid;
          gap: 0.15rem;
        }

        dt {
          font-family: var(--font-mono);
          font-size: 0.68rem;
          letter-spacing: 0.1em;
          text-transform: uppercase;
          color: rgba(23, 23, 23, 0.44);
        }

        dd {
          margin: 0;
          font-size: 1.05rem;
          font-weight: 600;
        }

        .neural-live-pill {
          display: inline-flex;
          align-items: center;
          gap: 0.4rem;
          padding: 0.38rem 0.72rem;
          border-radius: 999px;
          font-family: var(--font-mono);
          font-size: 0.7rem;
          letter-spacing: 0.1em;
          text-transform: uppercase;
        }

        .neural-live-pill.is-live {
          background: rgba(28, 28, 28, 0.08);
          color: rgba(28, 28, 28, 0.78);
        }

        .neural-live-pill.is-degraded {
          background: rgba(96, 96, 96, 0.1);
          color: rgba(28, 28, 28, 0.72);
        }

        .neural-stage {
          display: grid;
          grid-template-columns: minmax(0, 1fr) 320px;
          gap: 1rem;
          align-items: start;
        }

        .neural-canvas-wrap {
          position: relative;
          min-height: min(78vh, 980px);
          padding: 0.55rem;
          border-radius: 32px;
          background: rgba(233, 229, 223, 0.34);
          border: 1px solid rgba(23, 23, 23, 0.06);
          box-shadow: 0 20px 80px rgba(23, 23, 23, 0.06);
        }

        .neural-three-layer,
        .neural-particle-layer {
          width: 100%;
          height: 100%;
          display: block;
          border-radius: 28px;
        }

        .neural-three-layer,
        .neural-particle-layer {
          position: absolute;
          inset: 0.55rem;
          width: calc(100% - 1.1rem);
          height: calc(100% - 1.1rem);
        }

        .neural-three-layer {
          z-index: 0;
        }

        .neural-particle-layer {
          z-index: 1;
          cursor: pointer;
        }

        .neural-sidecar {
          display: grid;
          gap: 0.9rem;
        }

        .neural-panel {
          padding: 1rem;
          border-radius: 24px;
        }

        .neural-panel h2 {
          margin: 0 0 0.5rem;
          font-size: 1.2rem;
          line-height: 1.2;
        }

        .neural-panel p {
          margin: 0;
          line-height: 1.55;
          color: rgba(23, 23, 23, 0.66);
        }

        .neural-list {
          list-style: none;
          padding: 0;
          margin: 0;
          display: grid;
          gap: 0.55rem;
        }

        .neural-list button {
          width: 100%;
          text-align: left;
          padding: 0.72rem 0.8rem;
          border-radius: 16px;
          border: 1px solid rgba(23, 23, 23, 0.08);
          background: rgba(246, 243, 238, 0.8);
          display: grid;
          gap: 0.2rem;
        }

        .neural-list strong {
          font-family: var(--font-mono);
          font-size: 0.68rem;
          letter-spacing: 0.1em;
          text-transform: uppercase;
          color: rgba(23, 23, 23, 0.5);
        }

        .neural-list span {
          color: rgba(23, 23, 23, 0.82);
        }

        .neural-error {
          margin-top: 0.7rem !important;
          color: rgba(23, 23, 23, 0.7) !important;
        }

        @media (max-width: 1120px) {
          .neural-hero,
          .neural-stage {
            grid-template-columns: 1fr;
          }
        }

        @media (max-width: 720px) {
          .neural-page {
            padding: 1rem 0.8rem 1.4rem;
          }
        }
      `}</style>
    </main>
  );
}
