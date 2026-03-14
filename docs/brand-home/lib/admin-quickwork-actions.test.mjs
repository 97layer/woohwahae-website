import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const runtimeSource = await readFile(new URL('./runtime/quickwork.js', import.meta.url), 'utf8');
const componentSource = await readFile(new URL('../components/admin-quickwork-actions.js', import.meta.url), 'utf8');

test('quickwork runtime derives a founder-facing attention summary and submit follow-up', () => {
  assert.match(runtimeSource, /function buildQuickworkAttention\(status\)/);
  assert.match(runtimeSource, /mode:\s*'review_open_job'/);
  assert.match(runtimeSource, /mode:\s*'seed_planner'/);
  assert.match(runtimeSource, /mode:\s*'seed_implementer'/);
  assert.match(runtimeSource, /function buildQuickworkFollowUp\(job, dispatch\)/);
});

test('quickwork admin surface shows next action before presets and forms', () => {
  assert.match(componentSource, /const attention = result\?\.follow_up \|\| runtime\?\.attention \|\| null/);
  assert.match(componentSource, /<span className="signal-card__source">다음 액션<\/span>/);
  assert.match(componentSource, /attention\?\.mode === 'review_open_job'/);
  assert.match(componentSource, /payload\?\.follow_up\?\.summary \|\| '빠른 실행 제출 완료'/);
});
