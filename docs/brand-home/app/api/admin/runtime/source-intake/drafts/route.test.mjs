import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const routeSource = await readFile(new URL('./route.js', import.meta.url), 'utf8');
const runtimeSource = await readFile(new URL('../../../../../../lib/runtime/source-intake.js', import.meta.url), 'utf8');

test('source draft lane admin route keeps founder-only write auth and schema', () => {
  assert.match(routeSource, /authorizeRequest\(request,\s*\['founder'\]\)/);
  assert.match(routeSource, /adminSourceDraftLaneSchema/);
});

test('source draft lane helper defers to the canonical backend prep route', () => {
  assert.match(routeSource, /fetchLayerOs\('\/api\/layer-os\/source-intake\/drafts\/prep'/);
  assert.match(routeSource, /follow_up:\s*normalizeAttention\(payload\?\.follow_up\)/);
  assert.match(routeSource, /draftObservationId:\s*asText\(openedFromDraft\?\.draft_observation_id\)/);
  assert.match(routeSource, /targetAccountLabel:\s*asText\(lane\?\.target_account_label\)/);
  assert.doesNotMatch(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/source-intake\/drafts\/prep'/);
});
