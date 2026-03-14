import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const routeSource = await readFile(new URL('./route.js', import.meta.url), 'utf8');
const runtimeSource = await readFile(new URL('../../../../../lib/runtime/telegram.js', import.meta.url), 'utf8');

test('telegram admin route protects preview and send with the expected roles', () => {
  assert.match(routeSource, /authorizeRequest\(request,\s*\['founder', 'admin'\]\)/);
  assert.match(routeSource, /authorizeRequest\(request,\s*\['founder'\]\)/);
});

test('telegram runtime helper stays on the canonical daemon path', () => {
  assert.match(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/telegram'\)/);
  assert.match(runtimeSource, /requireWriteToken:\s*true/);
  assert.match(runtimeSource, /telegram:\s*await getAdminTelegramView\(\)/);
  assert.match(runtimeSource, /return enabled \? 'configured' : 'noop';/);
  assert.match(runtimeSource, /const status = normalizeStatus\(payload, enabled\)/);
  assert.match(runtimeSource, /attention:\s*buildTelegramAttention\(status, packet\)/);
});
