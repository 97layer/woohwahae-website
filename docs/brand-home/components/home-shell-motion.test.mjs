import assert from 'node:assert/strict';
import test from 'node:test';

import {
  PULL_REFRESH_MAX_DISTANCE,
  PULL_REFRESH_READY_DISTANCE,
  PULL_REFRESH_RELOAD_DELAY_MS,
  PULL_REFRESH_SETTLED_DISTANCE,
  computePullRefreshDistance,
  isPullRefreshReady,
} from './home-shell-motion.mjs';

test('pull refresh ramps resistance as the drag grows', () => {
  const shortPull = computePullRefreshDistance(40);
  const longPull = computePullRefreshDistance(120);

  assert.ok(shortPull > 0);
  assert.ok(longPull > shortPull);
  assert.ok(shortPull < 40);
  assert.ok(longPull < 120);
});

test('pull refresh never goes negative and caps long drags', () => {
  assert.equal(computePullRefreshDistance(-12), 0);
  assert.equal(computePullRefreshDistance(0), 0);
  assert.equal(computePullRefreshDistance(800), PULL_REFRESH_MAX_DISTANCE);
});

test('pull refresh only triggers past the sticky reload threshold', () => {
  assert.equal(isPullRefreshReady(PULL_REFRESH_READY_DISTANCE - 0.01), false);
  assert.equal(isPullRefreshReady(PULL_REFRESH_READY_DISTANCE), true);
  assert.ok(PULL_REFRESH_READY_DISTANCE > 0);
  assert.ok(PULL_REFRESH_SETTLED_DISTANCE < PULL_REFRESH_READY_DISTANCE);
  assert.ok(PULL_REFRESH_RELOAD_DELAY_MS > 0);
});
