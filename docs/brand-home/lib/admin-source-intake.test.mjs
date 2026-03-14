import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const runtimeSource = await readFile(new URL('./runtime/source-intake.js', import.meta.url), 'utf8');
const componentSource = await readFile(new URL('../components/admin-source-intake-actions.js', import.meta.url), 'utf8');

test('source intake runtime derives a founder-facing next action before detail', () => {
  assert.match(runtimeSource, /function buildSourceIntakeAttention\(\)/);
  assert.match(runtimeSource, /safeLayerOs\('\/api\/layer-os\/source-intake', \{ timeoutMs \}\)/);
  assert.match(runtimeSource, /mode:\s*'drop_source'/);
  assert.match(runtimeSource, /링크나 메모를 하나 저장해 source intake를 시작하세요\./);
  assert.match(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/source-intake'/);
  assert.match(runtimeSource, /function normalizeSourceIntakeAttention\(attention\)/);
  assert.match(runtimeSource, /attention:\s*normalizeSourceIntakeAttention\(payloadResult\.payload\?\.attention\)/);
  assert.match(runtimeSource, /next_action:\s*normalizeSourceIntakeAttention\(payload\?\.next_action\)/);
  assert.doesNotMatch(runtimeSource, /function normalizeSourceDraftSeedItem\(item\)/);
  assert.doesNotMatch(runtimeSource, /function normalizeSourceIntakeRouteDecisionItem\(item\)/);
  assert.doesNotMatch(runtimeSource, /function normalizeSourceIntakePrepItem\(item\)/);
  assert.doesNotMatch(runtimeSource, /function buildSourceIntakeFollowUp\(item\)/);
  assert.doesNotMatch(runtimeSource, /function buildDraftPrepFollowUp\(/);
  assert.doesNotMatch(runtimeSource, /function sortSourceIntakeItems\(items\)/);
  assert.doesNotMatch(runtimeSource, /function isSourceIntakeActionItem\(item\)/);
  assert.doesNotMatch(runtimeSource, /function deriveSourceIntakePriority/);
  assert.doesNotMatch(runtimeSource, /function buildSourceIntakeRawExcerpt/);
});

test('source intake admin surface shows the next action ahead of the form and result details', () => {
  assert.match(componentSource, /const attention = result\?\.follow_up \|\| result\?\.next_action \|\| initialView\?\.attention \|\| null/);
  assert.match(componentSource, /<span className=\"signal-card__source\">다음 액션<\/span>/);
  assert.match(componentSource, /attention\?\.mode === 'open_threads_prep'/);
  assert.match(componentSource, /origin ·/);
  assert.match(componentSource, /route opened ·/);
  assert.match(componentSource, /priority ·/);
  assert.match(componentSource, /prep lane ·/);
  assert.match(componentSource, /summaryMeta/);
  assert.match(componentSource, /summaryNote/);
  assert.match(componentSource, /quietNote/);
  assert.match(componentSource, /quiet candidates/);
  assert.match(componentSource, /active action queue empty/);
  assert.match(componentSource, /payload\?\.next_action\?\.summary/);
  assert.match(componentSource, /payload\?\.follow_up\?\.summary/);
});
