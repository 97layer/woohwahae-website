import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const shellSource = await readFile(new URL('./home-shell.js', import.meta.url), 'utf8');
const routeSource = await readFile(new URL('../content/public-surface-pages.js', import.meta.url), 'utf8');
const archivePageSource = await readFile(new URL('../app/archive/page.js', import.meta.url), 'utf8');
const worksPageSource = await readFile(new URL('../app/works/page.js', import.meta.url), 'utf8');

test('public IA stays on the original archive works about family', () => {
  assert.doesNotMatch(shellSource, /Runtime proof/);
  assert.match(shellSource, /label: 'Archive'/);
  assert.match(shellSource, /label: 'Works'/);
  assert.match(shellSource, /label: 'About'/);
  assert.doesNotMatch(shellSource, /label: 'Practice'/);
  assert.match(shellSource, /\/archive\/log/);
  assert.match(shellSource, /\/archive\/curation/);
  assert.match(shellSource, /\/works\/atelier/);
  assert.match(shellSource, /\/works\/offering/);
  assert.match(shellSource, /\/works\/project/);
  assert.match(shellSource, /\/about\/woosunho/);
  assert.match(shellSource, /shellOnly/);
  assert.match(routeSource, /archive_index/);
  assert.match(routeSource, /archive_log/);
  assert.match(routeSource, /works_index/);
  assert.match(routeSource, /works_atelier/);
  assert.match(routeSource, /about_woosunho/);
  assert.match(archivePageSource, /archive_index/);
  assert.doesNotMatch(archivePageSource, /redirect\(/);
  assert.match(worksPageSource, /works_index/);
  assert.doesNotMatch(worksPageSource, /redirect\(/);
});
