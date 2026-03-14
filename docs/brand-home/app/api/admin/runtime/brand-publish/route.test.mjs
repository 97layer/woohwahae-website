import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const routeSource = await readFile(new URL('./route.js', import.meta.url), 'utf8');
const runtimeSource = await readFile(new URL('../../../../../lib/runtime/brand-publish.js', import.meta.url), 'utf8');

test('brand publish admin route protects view and mutation with the expected roles', () => {
  assert.match(routeSource, /authorizeRequest\(request,\s*\['founder', 'admin'\]\)/);
  assert.match(routeSource, /authorizeRequest\(request,\s*\['founder'\]\)/);
  assert.match(routeSource, /adminBrandPublishSchema/);
});

test('brand publish runtime helper stays on canonical proposal, approval, flow, and observation routes', () => {
  assert.match(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/proposals'/);
  assert.match(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/proposals\/promote'/);
  assert.match(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/approval-inbox'/);
  assert.match(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/flows\/sync'/);
  assert.match(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/observations'/);
  assert.match(runtimeSource, /lane:brand_publish/);
  assert.match(runtimeSource, /attention:\s*buildBrandPublishAttention\(/);
  assert.match(runtimeSource, /next_action:\s*buildBrandPublishFollowUp\(/);
});
