import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const routeSource = await readFile(new URL('./route.js', import.meta.url), 'utf8');

test('runtime token route exposes founder/admin status and founder-only mutation', () => {
  assert.match(routeSource, /authorizeRequest\(request,\s*\['founder', 'admin'\]\)/);
  assert.match(routeSource, /authorizeRequest\(request,\s*\['founder'\]\)/);
  assert.match(routeSource, /adminRuntimeTokenSchema/);
  assert.match(routeSource, /attachLayerOsWriteTokenCookie/);
  assert.match(routeSource, /clearLayerOsWriteTokenCookie/);
});
