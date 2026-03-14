import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const nextConfigSource = await readFile(new URL('../next.config.mjs', import.meta.url), 'utf8');
const homeShellSource = await readFile(new URL('../components/home-shell.js', import.meta.url), 'utf8');
const notFoundSource = await readFile(new URL('../app/not-found.js', import.meta.url), 'utf8');
const robotsSource = await readFile(new URL('../app/robots.js', import.meta.url), 'utf8');
const sitemapSource = await readFile(new URL('../app/sitemap.js', import.meta.url), 'utf8');
const privacyPageSource = await readFile(new URL('../app/privacy/page.js', import.meta.url), 'utf8');
const termsPageSource = await readFile(new URL('../app/terms/page.js', import.meta.url), 'utf8');
const paymentSuccessPageSource = await readFile(new URL('../app/payment-success/page.js', import.meta.url), 'utf8');
const paymentFailPageSource = await readFile(new URL('../app/payment-fail/page.js', import.meta.url), 'utf8');
const labPageSource = await readFile(new URL('../app/lab/page.js', import.meta.url), 'utf8');

test('legacy html public routes redirect into current app routes', () => {
  assert.match(nextConfigSource, /source:\s*'\/privacy\.html'/);
  assert.match(nextConfigSource, /destination:\s*'\/privacy'/);
  assert.match(nextConfigSource, /source:\s*'\/terms\.html'/);
  assert.match(nextConfigSource, /destination:\s*'\/terms'/);
  assert.match(nextConfigSource, /source:\s*'\/payment-success\.html'/);
  assert.match(nextConfigSource, /destination:\s*'\/payment-success'/);
  assert.match(nextConfigSource, /source:\s*'\/payment-fail\.html'/);
  assert.match(nextConfigSource, /destination:\s*'\/payment-fail'/);
  assert.match(nextConfigSource, /source:\s*'\/works\/product'/);
  assert.match(nextConfigSource, /destination:\s*'\/works\/offering'/);
});

test('home shell keeps legal links and lab continuity inside the current app', () => {
  assert.match(homeShellSource, /label: 'Lab'/);
  assert.match(homeShellSource, /href: '\/lab'/);
  assert.match(homeShellSource, /href="\/privacy"/);
  assert.match(homeShellSource, /href="\/terms"/);
});

test('public continuity pages exist inside the current shell', () => {
  assert.match(privacyPageSource, /PublicUtilityPage/);
  assert.match(privacyPageSource, /개인정보/);
  assert.match(termsPageSource, /서비스/);
  assert.match(paymentSuccessPageSource, /결제/);
  assert.match(paymentFailPageSource, /결제/);
  assert.match(labPageSource, /redirect\('\/field-lab'\)/);
  assert.match(notFoundSource, /페이지를\\n찾을 수 없음/);
  assert.match(robotsSource, /sitemap:\s*'https:\/\/woohwahae\.kr\/sitemap\.xml'/);
  assert.match(sitemapSource, /https:\/\/woohwahae\.kr\/works\/offering/);
});
