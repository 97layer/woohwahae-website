import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const viewModelSource = await readFile(new URL('./runtime/view-model.js', import.meta.url), 'utf8');
const runtimeSource = await readFile(new URL('./runtime/layer-os.js', import.meta.url), 'utf8');
const publicProofSection = viewModelSource.slice(
  viewModelSource.indexOf('export async function getPublicProofView()'),
  viewModelSource.indexOf('export async function getPublicHomeView()'),
);

test('public proof view uses bounded runtime reads instead of the heavyweight cockpit surface', () => {
  assert.match(publicProofSection, /const timeoutMs = 3000/);
  assert.match(publicProofSection, /safeLayerOs\('\/api\/layer-os\/status', \{ timeoutMs \}\)/);
  assert.match(publicProofSection, /safeLayerOs\('\/api\/layer-os\/founder-summary', \{ timeoutMs \}\)/);
  assert.match(publicProofSection, /safeLayerOs\('\/api\/layer-os\/review-room', \{ timeoutMs \}\)/);
  assert.match(publicProofSection, /safeLayerOs\('\/api\/layer-os\/verifications', \{ timeoutMs \}\)/);
  assert.match(publicProofSection, /safeLayerOs\('\/api\/layer-os\/auth', \{ timeoutMs \}\)/);
  assert.match(publicProofSection, /safeLayerOs\('\/api\/layer-os\/observations\?limit=24', \{ timeoutMs \}\)/);
  assert.doesNotMatch(publicProofSection, /safeLayerOs\('\/api\/layer-os\/cockpit'\)/);
});

test('Layer OS runtime fetches can fail fast with request timeouts', () => {
  assert.match(runtimeSource, /function createRequestSignal\(path, timeoutMs, upstreamSignal\)/);
  assert.match(runtimeSource, /Layer OS request timed out for \$\{path\} after \$\{timeoutMs\}ms/);
  assert.match(runtimeSource, /signal: request\.signal/);
});
