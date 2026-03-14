import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const css = await readFile(new URL('../app/globals.css', import.meta.url), 'utf8');

function escapeRegex(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function getBlock(selector) {
  const match = css.match(new RegExp(`${escapeRegex(selector)}\\s*\\{([^}]*)\\}`, 'm'));
  assert.ok(match, `Expected CSS block for ${selector}`);
  return match[1];
}

function getIntroDelay(selector) {
  const match = css.match(
    new RegExp(
      `${escapeRegex(selector)}\\s*\\{\\s*transition:\\s*opacity\\s+\\d+ms\\s+cubic-bezier\\([\\d.,\\s]+\\)\\s+(\\d+)ms`,
      'm',
    ),
  );
  assert.ok(match, `Expected intro delay for ${selector}`);
  return Number.parseInt(match[1], 10);
}

test('intro reveal keeps a staged sequence without pinning exact timing values', () => {
  const brandDelay = getIntroDelay('.shell-header__inner > .shell-brand');
  const homeDelay = getIntroDelay('.home-shell');
  const navDelay = getIntroDelay('.shell-header__inner > .nav-toggle');
  const footerDelay = getIntroDelay('.shell-footer__inner');

  assert.ok(brandDelay < homeDelay);
  assert.ok(homeDelay <= navDelay);
  assert.ok(navDelay < footerDelay);
});

test('menu hover keeps position fixed and only changes tone', () => {
  const hoverBlock = getBlock('.nav-item:hover');

  assert.match(hoverBlock, /color:\s*var\(--text\)/);
  assert.doesNotMatch(hoverBlock, /transform\s*:/);
});

test('menu label and description keep split hierarchy without full mono treatment', () => {
  const labelBlock = getBlock('.nav-item__label');
  const descBlock = getBlock('.nav-item__desc');

  assert.match(labelBlock, /font-family:\s*var\(--font-mono\)/);
  assert.match(labelBlock, /font-size:\s*0\.58rem/);
  assert.match(descBlock, /font-size:\s*0\.5rem/);
  assert.match(descBlock, /font-weight:\s*300/);
  assert.doesNotMatch(descBlock, /font-family:\s*var\(--font-mono\)/);
});

test('interactive elements keep a dedicated active press transform', () => {
  assert.match(
    css,
    /\.nav-toggle:active,\s*\.nav-item:active,\s*\.shell-footer__contact-toggle:active,\s*\.field-lab-panel__toggle:active,\s*\.field-lab-preset:active\s*\{[^}]*transform\s*:/s,
  );
});

test('footer stays close to the legacy contact-toggle pattern without ghostmark clutter', () => {
  assert.doesNotMatch(css, /\.shell-footer__ghostmark\s*\{/);
  assert.match(css, /\.shell-footer__contact-toggle\s*\{/);
  assert.match(css, /\.shell-footer__contact-panel\s*\{/);
});

test('public home keeps quiet supporting metadata in the same visual family', () => {
  const fogBlock = getBlock('.home-hero__fog');
  const metaBlock = getBlock('.hero-copy__meta');
  const timestampBlock = getBlock('.hero-panel__timestamp');

  assert.match(fogBlock, /radial-gradient/);
  assert.match(metaBlock, /font-family:\s*var\(--font-mono\)/);
  assert.match(metaBlock, /text-transform:\s*uppercase/);
  assert.match(timestampBlock, /font-family:\s*var\(--font-mono\)/);
  assert.match(css, /\.hero-copy__visual-image\s*\{[^}]*object-fit:\s*contain/s);
  assert.match(css, /\.hero-copy__visual-meta\s*\{[^}]*font-family:\s*var\(--font-mono\)/s);
  assert.match(css, /\.proof-note__source\s*\{[^}]*font-size:\s*0\.64rem/s);
  assert.match(css, /\.hero-title__line\s*\{\s*display:\s*block;/);
});
