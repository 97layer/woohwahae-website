import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const layoutSource = await readFile(new URL('../app/layout.js', import.meta.url), 'utf8');
const manifestSource = await readFile(new URL('../app/manifest.js', import.meta.url), 'utf8');
const serviceWorkerSource = await readFile(new URL('../public/sw.js', import.meta.url), 'utf8');
const adminShellSource = await readFile(new URL('../components/admin-shell.js', import.meta.url), 'utf8');

test('root layout declares Layer OS as a single installable shell', () => {
  assert.match(layoutSource, /applicationName:\s*'Layer OS'/);
  assert.match(layoutSource, /manifest:\s*'\/manifest\.webmanifest'/);
  assert.match(layoutSource, /appleWebApp:\s*\{/);
  assert.match(layoutSource, /<PwaRegistration \/>/);
  assert.match(layoutSource, /themeColor:\s*'#f4f4f4'/);
});

test('manifest anchors the single PWA shell on admin login and provides app shortcuts', () => {
  assert.match(manifestSource, /start_url:\s*'\/admin\/login'/);
  assert.match(manifestSource, /display:\s*'standalone'/);
  assert.match(manifestSource, /short_name:\s*'Layer OS'/);
  assert.match(manifestSource, /url:\s*'\/admin'/);
  assert.match(manifestSource, /url:\s*'\/'/);
  assert.match(manifestSource, /icon-192\.png/);
  assert.match(manifestSource, /icon-512\.png/);
});

test('service worker caches the shell entry points and keeps admin login as the offline fallback', () => {
  assert.match(serviceWorkerSource, /const CACHE_NAME = 'layer-os-shell-v1'/);
  assert.match(serviceWorkerSource, /'\/admin\/login'/);
  assert.match(serviceWorkerSource, /'\/manifest\.webmanifest'/);
  assert.match(serviceWorkerSource, /self\.addEventListener\('install'/);
  assert.match(serviceWorkerSource, /self\.addEventListener\('fetch'/);
  assert.match(serviceWorkerSource, /if \(url\.pathname\.startsWith\('\/admin'\)\)/);
});

test('admin shell exposes the single-PWA posture and install affordance', () => {
  assert.match(adminShellSource, /import PwaInstallButton from '\.\/pwa-install-button'/);
  assert.match(adminShellSource, /<PwaInstallButton \/>/);
  assert.match(adminShellSource, /같은 단일 PWA 안에서 보고/);
  assert.match(adminShellSource, /href: '\/'/);
});
