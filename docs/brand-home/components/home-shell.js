'use client';

import Image from 'next/image';
import Script from 'next/script';
import { useEffect, useRef, useState } from 'react';

import {
  computeMagneticPull,
  computePullRefreshDistance,
  computeWhimsyTension,
  isPullRefreshReady,
  PULL_REFRESH_RELOAD_DELAY_MS,
  PULL_REFRESH_SETTLED_DISTANCE,
} from './home-shell-motion.mjs';

const navGroups = [
  {
    label: 'Archive',
    href: '/archive',
    items: [
      { label: 'Log', href: '/archive/log', desc: '사유 / 기록' },
      { label: 'Curation', href: '/archive/curation', desc: '시선 / 청음' },
    ],
  },
  {
    label: 'Works',
    href: '/works',
    items: [
      { label: 'Atelier', href: '/works/atelier', desc: '작업' },
      { label: 'Offering', href: '/works/offering', desc: '서비스 / 제품' },
      { label: 'Project', href: '/works/project', desc: '협업' },
    ],
  },
  {
    label: 'About',
    href: '/about',
    items: [
      { label: 'Root', href: '/about', desc: '근원' },
      { label: 'Woosunho', href: '/about/woosunho', desc: '기획자' },
    ],
  },
  {
    label: 'Lab',
    href: '/lab',
    items: [
      { label: 'Field Lab', href: '/field-lab', desc: '실험 / 구조' },
      { label: 'Neural', href: '/field-lab/neural', desc: '세포 단위 보기' },
    ],
  },
];

const fieldPresets = [
  { key: 'membrane', label: 'Membrane', note: '얇은 막 같은 존재가 조용히 호흡하는 가장 새로운 대안' },
  { key: 'monolith', label: 'Monolith', note: '정체 모를 기물 같은 존재감으로 가는 방향' },
  { key: 'constellation', label: 'Constellation', note: '선 대신 점과 잔향으로 구조를 드러내는 방향' },
  { key: 'observer', label: 'Observer', note: '관측되는 순간 구조가 드러나는 더 새로운 메인 후보' },
  { key: 'architect', label: 'Architect', note: '낯선 장력과 설계 도면 같은 축을 가진 메인 후보' },
  { key: 'quiet', label: 'Quiet', note: '현재 결을 유지한 기준 필드' },
  { key: 'ink', label: 'Ink', note: '먹 번짐 같은 코어와 잔향' },
  { key: 'orbit', label: 'Orbit', note: '외곽 궤도감이 더 도는 버전' },
];

function usePrefersReducedMotion() {
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false);

  useEffect(() => {
    const media = window.matchMedia('(prefers-reduced-motion: reduce)');
    setPrefersReducedMotion(media.matches);

    function onChange(event) {
      setPrefersReducedMotion(event.matches);
    }

    if (typeof media.addEventListener === 'function') media.addEventListener('change', onChange);
    else if (typeof media.addListener === 'function') media.addListener(onChange);

    return () => {
      if (typeof media.removeEventListener === 'function') media.removeEventListener('change', onChange);
      else if (typeof media.removeListener === 'function') media.removeListener(onChange);
    };
  }, []);

  return prefersReducedMotion;
}

export default function HomeShell({
  preset = 'architect',
  lab = false,
  surface = 'home',
  shellOnly = false,
  headerTitle: headerTitleProp = '',
  children = null,
}) {
  const [open, setOpen] = useState(false);
  const [labOpen, setLabOpen] = useState(false);
  const [footerOpen, setFooterOpen] = useState(false);
  const [timeLabel, setTimeLabel] = useState('');
  const [introReady, setIntroReady] = useState(false);
  const footerRef = useRef(null);
  const pullStartYRef = useRef(0);
  const pullDistanceRef = useRef(0);
  const pullActiveRef = useRef(false);
  const pullRefreshingRef = useRef(false);
  const [cursorPos, setCursorPos] = useState({ x: 0, y: 0 });
  const [magneticOffset, setMagneticOffset] = useState({ symbol: { x: 0, y: 0 }, toggle: { x: 0, y: 0 } });
  const prefersReducedMotion = usePrefersReducedMotion();

  useEffect(() => {
    if (prefersReducedMotion || open) {
      setCursorPos({ x: 0, y: 0 });
      return undefined;
    }

    function onPointerMove(e) {
      setCursorPos({ x: e.clientX, y: e.clientY });
    }
    window.addEventListener('pointermove', onPointerMove, { passive: true });
    return () => window.removeEventListener('pointermove', onPointerMove);
  }, [prefersReducedMotion, open]);

  useEffect(() => {
    if (prefersReducedMotion || open) {
      setMagneticOffset({ symbol: { x: 0, y: 0 }, toggle: { x: 0, y: 0 } });
      return undefined;
    }

    let rafId;
    function updateMagnetic() {
      const symbolEl = document.querySelector('.shell-brand');
      const toggleEl = document.querySelector('.nav-toggle');

      setMagneticOffset((prev) => {
        const next = { ...prev };
        if (symbolEl) {
          const target = computeMagneticPull(cursorPos.x, cursorPos.y, symbolEl.getBoundingClientRect(), 0.12, 120);
          next.symbol.x = computeWhimsyTension(prev.symbol.x, target.x, 0.12);
          next.symbol.y = computeWhimsyTension(prev.symbol.y, target.y, 0.12);
        }
        if (toggleEl) {
          const target = computeMagneticPull(cursorPos.x, cursorPos.y, toggleEl.getBoundingClientRect(), 0.18, 140);
          next.toggle.x = computeWhimsyTension(prev.toggle.x, target.x, 0.14);
          next.toggle.y = computeWhimsyTension(prev.toggle.y, target.y, 0.14);
        }
        return next;
      });
      rafId = requestAnimationFrame(updateMagnetic);
    }
    rafId = requestAnimationFrame(updateMagnetic);
    return () => cancelAnimationFrame(rafId);
  }, [cursorPos, prefersReducedMotion, open]);

  useEffect(() => {
    document.body.classList.toggle('nav-open', open);
    if (open) setFooterOpen(false);
    return () => document.body.classList.remove('nav-open');
  }, [open]);

  useEffect(() => {
    function onKeyDown(event) {
      if (event.key === 'Escape') {
        setOpen(false);
        setFooterOpen(false);
      }
    }
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, []);

  useEffect(() => {
    const raf = window.requestAnimationFrame(() => {
      window.setTimeout(() => setIntroReady(true), 240);
    });
    return () => window.cancelAnimationFrame(raf);
  }, []);

  useEffect(() => {
    function onDocumentClick(event) {
      if (!footerRef.current || !footerOpen) return;
      if (footerRef.current.contains(event.target)) return;
      setFooterOpen(false);
    }
    document.addEventListener('click', onDocumentClick);
    return () => document.removeEventListener('click', onDocumentClick);
  }, [footerOpen]);

  useEffect(() => {
    const formatter = new Intl.DateTimeFormat('ko-KR', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
      timeZone: 'Asia/Seoul',
    });

    function updateClock() {
      setTimeLabel(`${formatter.format(new Date())} KST`);
    }

    updateClock();
    const timer = window.setInterval(updateClock, 60000);
    return () => window.clearInterval(timer);
  }, []);

  useEffect(() => {
    const root = document.documentElement;
    const body = document.body;

    function resetPullState() {
      pullActiveRef.current = false;
      pullDistanceRef.current = 0;
      root.style.setProperty('--pull-distance', '0px');
      body.classList.remove('pull-refresh-active', 'pull-refresh-ready');
      if (!pullRefreshingRef.current) {
        body.classList.remove('pull-refreshing');
      }
    }

    if (open) {
      resetPullState();
      return;
    }

    const mobileQuery = window.matchMedia('(hover: none) and (pointer: coarse)');
    if (!mobileQuery.matches) {
      resetPullState();
      return;
    }

    function onTouchStart(event) {
      if (open || labOpen || pullRefreshingRef.current) return;
      if (event.touches.length !== 1) return;
      if (window.scrollY > 0) return;
      pullStartYRef.current = event.touches[0].clientY;
      pullActiveRef.current = true;
      pullDistanceRef.current = 0;
    }

    function onTouchMove(event) {
      if (!pullActiveRef.current || pullRefreshingRef.current) return;
      if (event.touches.length !== 1) return;

      const deltaY = event.touches[0].clientY - pullStartYRef.current;
      if (deltaY <= 0) {
        resetPullState();
        return;
      }
      if (window.scrollY > 0) {
        resetPullState();
        return;
      }

      const distance = computePullRefreshDistance(deltaY);
      pullDistanceRef.current = distance;
      root.style.setProperty('--pull-distance', `${distance}px`);
      body.classList.add('pull-refresh-active');
      if (isPullRefreshReady(distance)) body.classList.add('pull-refresh-ready');
      else body.classList.remove('pull-refresh-ready');
      event.preventDefault();
    }

    function onTouchEnd() {
      if (!pullActiveRef.current) return;
      pullActiveRef.current = false;

      if (isPullRefreshReady(pullDistanceRef.current) && !pullRefreshingRef.current) {
        pullRefreshingRef.current = true;
        body.classList.add('pull-refresh-active', 'pull-refresh-ready', 'pull-refreshing');
        root.style.setProperty('--pull-distance', `${PULL_REFRESH_SETTLED_DISTANCE}px`);
        window.setTimeout(() => {
          window.location.reload();
        }, PULL_REFRESH_RELOAD_DELAY_MS);
        return;
      }

      resetPullState();
    }

    window.addEventListener('touchstart', onTouchStart, { passive: true });
    window.addEventListener('touchmove', onTouchMove, { passive: false });
    window.addEventListener('touchend', onTouchEnd, { passive: true });
    window.addEventListener('touchcancel', onTouchEnd, { passive: true });

    return () => {
      window.removeEventListener('touchstart', onTouchStart);
      window.removeEventListener('touchmove', onTouchMove);
      window.removeEventListener('touchend', onTouchEnd);
      window.removeEventListener('touchcancel', onTouchEnd);
      pullRefreshingRef.current = false;
      resetPullState();
    };
  }, [open, labOpen]);

  const headerEyebrow = 'WOOHWAHAE';
  const headerTitle = lab ? 'Field lab' : headerTitleProp || (surface === 'home' ? 'home shell' : surface);

  return (
    <>
      <div className={`intro-veil${introReady ? ' is-hidden' : ''}`} aria-hidden="true" />

      <div className="pull-refresh-indicator" aria-hidden="true">
        <span className="pull-refresh-indicator__label">Reload</span>
      </div>

      <header className={`shell-header${introReady ? ' is-intro-ready' : ''}`} id="shell-header">
        <div className="shell-header__inner">
          <a
            className="shell-brand"
            href="/"
            aria-label="WOOHWAHAE"
            style={{
              transform: `translate3d(${magneticOffset.symbol.x}px, ${magneticOffset.symbol.y}px, 0)`,
              willChange: 'transform',
            }}
          >
            <Image
              className="shell-brand__symbol"
              src="/assets/media/brand/symbol.png"
              alt="WOOHWAHAE"
              width={28}
              height={28}
              priority
            />
          </a>
          <div className="shell-header__title-group" aria-hidden="true">
            <span className="shell-header__eyebrow">{headerEyebrow}</span>
            <span className="shell-header__title">{headerTitle}</span>
          </div>
          <button
            className="nav-toggle"
            id="nav-toggle"
            type="button"
            aria-label="메뉴"
            aria-controls="nav-overlay"
            aria-expanded={open ? 'true' : 'false'}
            onClick={() => setOpen((value) => !value)}
            style={{
              transform: `translate3d(${magneticOffset.toggle.x}px, ${magneticOffset.toggle.y}px, 0)`,
              willChange: 'transform',
            }}
          >
            <span />
            <span />
          </button>
        </div>
      </header>

      <div
        className="nav-overlay"
        id="nav-overlay"
        aria-hidden={open ? 'false' : 'true'}
        onClick={(event) => {
          if (event.target === event.currentTarget) setOpen(false);
        }}
      >
        <div className="nav-overlay__panel">
          <div className="nav-overlay__inner">
            {navGroups.map((group) => (
              <section className="nav-sector" key={group.label}>
                <h2 className="nav-sector__label">
                  <a href={group.href} onClick={() => setOpen(false)}>{group.label}</a>
                </h2>
                <div className="nav-sector__items">
                  {group.items.map((item) => (
                    <a key={`${group.label}-${item.label}-${item.href}`} href={item.href} className="nav-item" onClick={() => setOpen(false)}>
                      <span className="nav-item__label">{item.label}</span>
                      <span className="nav-item__desc">{item.desc}</span>
                    </a>
                  ))}
                </div>
              </section>
            ))}
          </div>
        </div>
      </div>

      <canvas className="field-bg" id="field-bg" data-field-preset={preset} aria-hidden="true" />

      <main className={`home-shell${introReady ? ' is-intro-ready' : ''}`} aria-label="WOOHWAHAE Home">
        {shellOnly || !children ? (
          <section className="home-hero" aria-hidden="true">
            <div className="home-hero__fog" aria-hidden="true" />
          </section>
        ) : (
          children
        )}
        <section className="home-field-buffer" aria-hidden="true" />
      </main>

      {lab ? (
        <aside className={`field-lab-panel${labOpen ? ' is-open' : ''}`} aria-label="Field preset lab">
          <button
            type="button"
            className="field-lab-panel__toggle"
            aria-expanded={labOpen ? 'true' : 'false'}
            onClick={() => setLabOpen((value) => !value)}
          >
            <span className="field-lab-panel__toggle-label">Field Lab</span>
            <span className="field-lab-panel__toggle-current">{fieldPresets.find((item) => item.key === preset)?.label || preset}</span>
          </button>
          <div className="field-lab-panel__content">
            <p className="field-lab-panel__eyebrow">Field Lab</p>
            <h1 className="field-lab-panel__title">Signature field presets</h1>
            <p className="field-lab-panel__body">이번엔 파동 계열 밖으로도 꺼내봤습니다. 막, 기물, 성좌처럼 존재 자체가 전면에 오는 대안들입니다.</p>
            <a href="/field-lab/neural" className="field-lab-panel__feature-link">
              <span className="field-lab-panel__feature-kicker">Neural Field</span>
              <strong>세포 단위 구조 보기</strong>
              <span>실시간 런타임 구조를 세포 조직처럼 보는 전용 화면으로 이동합니다.</span>
            </a>
            <div className="field-lab-panel__list">
              {fieldPresets.map((item) => {
                const active = item.key === preset;
                return (
                  <a
                    key={item.key}
                    href={`/field-lab?p=${item.key}`}
                    className={`field-lab-preset${active ? ' is-active' : ''}`}
                    aria-current={active ? 'page' : undefined}
                  >
                    <span className="field-lab-preset__label">{item.label}</span>
                    <span className="field-lab-preset__note">{item.note}</span>
                  </a>
                );
              })}
            </div>
          </div>
        </aside>
      ) : null}

      <footer className={`shell-footer${introReady ? ' is-intro-ready' : ''}`} id="home-footer">
        <div className="shell-footer__inner" ref={footerRef}>
          <div className="shell-footer__mid">
            <button
              type="button"
              className="shell-footer__contact-toggle"
              aria-expanded={footerOpen ? 'true' : 'false'}
              aria-controls="shell-footer-contact"
              onClick={() => setFooterOpen((value) => !value)}
            >
              <span>Contact</span>
              <span className="shell-footer__toggle-icon">+</span>
            </button>
            <div
              className={`shell-footer__contact-panel${footerOpen ? ' is-open' : ''}`}
              id="shell-footer-contact"
              aria-hidden={footerOpen ? 'false' : 'true'}
            >
              <a href="tel:050713932075">0507-1393-2075</a>
              <a href="mailto:hello@woohwahae.kr">hello@woohwahae.kr</a>
              <a href="https://instagram.com/woohwahae" target="_blank" rel="noreferrer">Instagram</a>
            </div>
          </div>

          <div className="shell-footer__base">
            <p className="shell-footer__biz">CEO SUNHO JO · BRN 532-05-02854 · 32, Bangujeong 20-gil, Jung-gu, Ulsan</p>
            <div className="shell-footer__legal">
              <span>© 2026 WOOHWAHAE</span>
              <a href="/privacy">Privacy</a>
              <a href="/terms">Terms</a>
              <span className="shell-footer__time">{timeLabel || 'KST'}</span>
            </div>
          </div>
        </div>
      </footer>

      <Script src="/assets/vendor/three.min.js" strategy="afterInteractive" />
      <Script src="/assets/js/bg-field.js" strategy="afterInteractive" />
    </>
  );
}
