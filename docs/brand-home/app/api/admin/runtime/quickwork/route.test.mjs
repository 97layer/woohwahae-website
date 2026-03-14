import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const source = await readFile(new URL('../../../../../lib/runtime/quickwork.js', import.meta.url), 'utf8');

test('planner quickwork startup mirrors the CLI planner worker set', () => {
  assert.match(
    source,
    /export function quickworkOrchestratorUpArgs\(role\)\s*\{[\s\S]*case 'planner':[\s\S]*\['up', '--roles', 'implementer,verifier,planner'\]/,
  );
  assert.match(
    source,
    /export async function submitQuickwork\(input\)\s*\{[\s\S]*await runOrchestrator\(quickworkOrchestratorUpArgs\(role\)\);/,
  );
});

test('quickwork runtime status aggregates auth, dispatch profiles, and open jobs', () => {
  assert.match(source, /write_auth_enabled/);
  assert.match(source, /write_ready/);
  assert.match(source, /dispatch_profiles:/);
  assert.match(source, /open_jobs:/);
  assert.match(source, /attention = buildQuickworkAttention\(status\)/);
  assert.match(source, /follow_up:\s*buildQuickworkFollowUp\(job, dispatched\)/);
  assert.match(source, /safeLayerOs\('\/api\/layer-os\/auth', \{ timeoutMs: 4000 \}\)/);
  assert.match(source, /safeLayerOs\('\/api\/layer-os\/jobs\/profiles', \{ timeoutMs: 4000 \}\)/);
  assert.match(source, /safeLayerOs\('\/api\/layer-os\/jobs\?status=open&limit=6', \{ timeoutMs: 4000 \}\)/);
});
