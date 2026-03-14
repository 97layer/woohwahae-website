import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const runtimeSource = await readFile(new URL('./runtime/brand-publish.js', import.meta.url), 'utf8');
const componentSource = await readFile(new URL('../components/admin-brand-publish-actions.js', import.meta.url), 'utf8');

test('brand publish runtime derives a next action before founder reviews tables and receipts', () => {
  assert.match(runtimeSource, /function buildBrandPublishAttention\(/);
  assert.match(runtimeSource, /mode:\s*'publish_threads'/);
  assert.match(runtimeSource, /mode:\s*'review_brand_approval'/);
  assert.match(runtimeSource, /mode:\s*'open_brand_draft'/);
});

test('brand publish admin surface shows next action ahead of the draft form', () => {
  assert.match(componentSource, /const attention = publishResult\?\.follow_up \|\| result\?\.next_action \|\| initialView\?\.attention \|\| null/);
  assert.match(componentSource, /<span className="signal-card__source">다음 액션<\/span>/);
  assert.match(componentSource, /attention\?\.mode === 'publish_threads'/);
  assert.match(componentSource, /payload\?\.next_action\?\.summary/);
  assert.match(componentSource, /payload\?\.follow_up\?\.summary/);
});
