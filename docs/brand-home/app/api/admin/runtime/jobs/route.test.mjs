import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const routeSource = await readFile(new URL('./route.js', import.meta.url), 'utf8');
const runtimeSource = await readFile(new URL('../../../../../lib/runtime/control-tower.js', import.meta.url), 'utf8');

test('jobs admin route protects read and write paths with the expected roles', () => {
  assert.match(routeSource, /authorizeRequest\(request,\s*\['founder', 'admin'\]\)/);
  assert.match(routeSource, /authorizeRequest\(request,\s*\['founder'\]\)/);
  assert.match(routeSource, /readValidatedJson\(request,\s*adminJobsActionSchema\)/);
  assert.match(routeSource, /withRequestLayerOsWriteToken\(request,\s*async \(\) =>/);
  assert.match(routeSource, /request\.nextUrl\.searchParams\.get\('include_packet'\) === '1'/);
  assert.match(routeSource, /getControlTowerView\(selectedJobId,\s*\{\s*includePacket:\s*false\s*\}\)/);
});

test('control tower runtime helper stays on canonical Layer OS job surfaces', () => {
  assert.match(runtimeSource, /getQuickworkRuntimeStatus\(\)/);
  assert.match(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/jobs\?limit=36', \{ timeoutMs: controlTowerTimeoutMs \}\)/);
  assert.match(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/knowledge', \{ timeoutMs: controlTowerTimeoutMs \}\)/);
  assert.match(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/providers', \{ timeoutMs: controlTowerTimeoutMs \}\)/);
  assert.match(runtimeSource, /export async function getControlTowerPacket\(jobId,\s*selectedJob = null\)/);
  assert.match(runtimeSource, /fetchLayerOs\(`\/api\/layer-os\/jobs\/packet\?job_id=\$\{encodeURIComponent\(normalizedJobID\)\}`,\s*\{\s*timeoutMs: controlTowerTimeoutMs,\s*\}\)/);
  assert.match(runtimeSource, /getControlTowerPacket\(selectedJobId,\s*selectedJob\)/);
  assert.match(runtimeSource, /submitQuickwork\(\{/);
  assert.match(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/jobs\/dispatch'/);
  assert.match(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/jobs\/update'/);
  assert.match(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/jobs\/promote'/);
});
