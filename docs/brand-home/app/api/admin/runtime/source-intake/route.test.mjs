import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const routeSource = await readFile(new URL('./route.js', import.meta.url), 'utf8');
const runtimeSource = await readFile(new URL('../../../../../lib/runtime/source-intake.js', import.meta.url), 'utf8');

test('source intake admin route protects view and mutation with the expected roles and schema', () => {
  assert.match(routeSource, /authorizeRequest\(request,\s*\['founder', 'admin'\]\)/);
  assert.match(routeSource, /authorizeRequest\(request,\s*\['founder'\]\)/);
  assert.match(routeSource, /adminSourceIntakeSchema/);
});

test('source intake runtime helper reads the canonical source-intake view and creates through the canonical source-intake route', () => {
  assert.match(runtimeSource, /safeLayerOs\('\/api\/layer-os\/source-intake', \{ timeoutMs \}\)/);
  assert.match(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/source-intake'/);
  assert.doesNotMatch(runtimeSource, /fetchLayerOs\('\/api\/layer-os\/observations'/);
  assert.doesNotMatch(runtimeSource, /topic:\s*SOURCE_INTAKE_TOPIC/);
  assert.match(runtimeSource, /source_channel:\s*'cockpit'/);
  assert.match(runtimeSource, /confidence:\s*'high'/);
  assert.match(runtimeSource, /refs:\s*sourceIntakeRefs\(requestPayload\)/);
  assert.match(runtimeSource, /runtimeAvailable:\s*false/);
  assert.match(runtimeSource, /attention:\s*buildSourceIntakeAttention\(\)/);
  assert.match(runtimeSource, /function buildDegradedSourceIntakeView\(degradedReason\)/);
  assert.match(runtimeSource, /function buildSourceIntakeRequestPayload\(input\)/);
  assert.match(runtimeSource, /attention:\s*normalizeSourceIntakeAttention\(payloadResult\.payload\?\.attention\)/);
  assert.match(runtimeSource, /next_action:\s*normalizeSourceIntakeAttention\(payload\?\.next_action\)/);
  assert.doesNotMatch(runtimeSource, /buildSourceIntakeFollowUp\(normalized\)/);
});
