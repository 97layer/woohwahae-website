import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const routeSource = await readFile(new URL('./route.js', import.meta.url), 'utf8');
const runtimeSource = await readFile(new URL('../../../../../../lib/runtime/brand-publish.js', import.meta.url), 'utf8');

test('threads brand publish admin route keeps founder-only write auth and input validation', () => {
  assert.match(routeSource, /authorizeRequest\(request,\s*\['founder'\]\)/);
  assert.match(routeSource, /adminBrandPublishThreadsSchema/);
});

test('threads brand publish helper stays on the canonical daemon publish route', () => {
  assert.match(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/social\/threads'/);
  assert.match(runtimeSource, /follow_up:\s*buildThreadsPublishFollowUp\(payload\)/);
});
