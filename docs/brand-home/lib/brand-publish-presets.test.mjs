import assert from 'node:assert/strict';
import test from 'node:test';

import { publicHomeLocalSourcePack } from '../content/public-home-source.js';
import { buildBrandPublishPresets } from './brand-publish-presets.js';

test('brand publish presets carry media and imported style cues separately from the draft body', () => {
  const presets = buildBrandPublishPresets(publicHomeLocalSourcePack);
  const threadsPreset = presets.find((item) => item.channel === 'threads');

  assert.ok(threadsPreset);
  assert.equal(threadsPreset.styleProfile, 'threads-repo-local-snapshot-v1');
  assert.ok(threadsPreset.styleCue.length > 0);
  assert.ok(Array.isArray(threadsPreset.styleExamples));
  assert.ok(threadsPreset.styleExamples.length > 0);
  assert.deepEqual(threadsPreset.styleExampleIds, threadsPreset.styleExamples.map((item) => item.exampleId));
  assert.deepEqual(threadsPreset.mediaIds, ['absorbed-atelier-hero']);
  assert.match(threadsPreset.mediaCue, /Atelier silhouette/);
  assert.doesNotMatch(threadsPreset.body, /Style cue:/);
});
