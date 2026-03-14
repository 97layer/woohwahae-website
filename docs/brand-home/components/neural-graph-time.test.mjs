import assert from 'node:assert/strict';
import test from 'node:test';

import { formatRuntimeTimestamp } from './neural-graph-time.mjs';

test('formatRuntimeTimestamp renders a stable KST label', () => {
  assert.equal(formatRuntimeTimestamp('2026-03-09T14:49:32.000Z'), '23:49:32 KST');
});

test('formatRuntimeTimestamp falls back for invalid values', () => {
  assert.equal(formatRuntimeTimestamp('not-a-date'), '--:--:-- KST');
});
