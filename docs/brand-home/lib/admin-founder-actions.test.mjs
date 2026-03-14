import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const componentSource = await readFile(new URL('../components/admin-founder-actions.js', import.meta.url), 'utf8');
const pageSource = await readFile(new URL('../app/admin/(console)/page.js', import.meta.url), 'utf8');

test('founder actions surface shows the current founder-flow attention before raw forms', () => {
  assert.match(componentSource, /const attention = defaults\.attention \|\| null/);
  assert.match(componentSource, /지금 founder flow에서 볼 것/);
  assert.match(componentSource, /flow \/ approval \/ release \/ rollback은 아래 카드에서 바로 실행할 수 있습니다/);
});

test('admin overview passes founder-flow attention into the founder actions lane', () => {
  assert.match(pageSource, /const founderFlowDefaults = overview\.founderFlowDefaults \|\| \{\}/);
  assert.match(pageSource, /<AdminFounderActions[\s\S]*defaults=\{founderFlowDefaults\}/);
});
