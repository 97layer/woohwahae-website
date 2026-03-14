import assert from 'node:assert/strict';
import test from 'node:test';

import { publicHomeLocalSourcePack } from '../content/public-home-source.js';
import { buildPublicHomeBrandView, normalizeBrandSourcePack } from './brand-source-pack.js';

test('normalizeBrandSourcePack keeps a stable local pack shape and drops unknown source refs', () => {
  const pack = normalizeBrandSourcePack({
    packId: 'test-pack',
    voice: ['calm', 'calm', 'direct'],
    provenance: { kind: 'local', label: 'fixture' },
    channelProfiles: {
      threads: {
        profileId: 'threads-fixture',
        cues: ['one', 'one', 'two'],
        examples: [{ id: 'sample', excerpt: 'quiet sample' }],
      },
    },
    sources: [{ id: 'alpha', title: 'Alpha', tags: ['shell', 'shell'] }],
    media: [{ id: 'hero', src: '/hero.svg', sourceIds: ['alpha', 'missing'] }],
    sections: {
      hero: { sourceIds: ['alpha', 'missing'], mediaIds: ['hero', 'missing'] },
      strips: [{ id: 'strip', sourceIds: ['alpha', 'missing'] }],
      modules: [{ id: 'module', sourceIds: ['missing', 'alpha'] }],
      notes: [{ id: 'note', sourceIds: ['missing'] }],
    },
  });

  assert.equal(pack.packId, 'test-pack');
  assert.deepEqual(pack.voice, ['calm', 'direct']);
  assert.equal(pack.sources[0].sourceId, 'alpha');
  assert.deepEqual(pack.sources[0].tags, ['shell']);
  assert.equal(pack.channelProfiles.threads.profileId, 'threads-fixture');
  assert.deepEqual(pack.channelProfiles.threads.cues, ['one', 'two']);
  assert.equal(pack.channelProfiles.threads.examples[0].excerpt, 'quiet sample');
  assert.equal(pack.media[0].mediaId, 'hero');
  assert.deepEqual(pack.media[0].sourceIds, ['alpha']);
  assert.deepEqual(pack.sections.hero.sourceIds, ['alpha']);
  assert.deepEqual(pack.sections.hero.mediaIds, ['hero']);
  assert.deepEqual(pack.sections.strips[0].sourceIds, ['alpha']);
  assert.deepEqual(pack.sections.modules[0].sourceIds, ['alpha']);
  assert.deepEqual(pack.sections.notes[0].sourceIds, []);
});

test('buildPublicHomeBrandView attaches referenced sources for the public home sections', () => {
  const view = buildPublicHomeBrandView(publicHomeLocalSourcePack);

  assert.equal(view.packId, 'woohwahae-public-home-local');
  assert.equal(view.sourceCount, 9);
  assert.equal(view.mediaCount, 4);
  assert.equal(view.channelProfiles.threads.profileId, 'threads-repo-local-snapshot-v1');
  assert.ok(view.channelProfiles.threads.examples.length > 0);
  assert.equal(view.channelProfiles.instagram.profileId, 'instagram-repo-local-snapshot-v1');
  assert.ok(view.channelProfiles.instagram.examples.length >= view.channelProfiles.threads.examples.length);
  assert.equal(view.hero.actions.length, 2);
  assert.equal(view.hero.sources.length, 4);
  assert.equal(view.hero.media.length, 2);
  assert.equal(view.hero.title, 'A quiet atelier.\nSubtractive care.');
  assert.doesNotMatch(view.hero.body, /runtime|proof|Layer OS/i);
  assert.equal(view.hero.media[0].src, '/assets/media/brand/hair-atelier-hero.svg');
  assert.equal(view.hero.media[0].sources[0].sourceId, 'absorbed-route-archive');
  assert.equal(view.hero.sources[0].provenance.ref, 'constitution/brand.md');
  assert.equal(view.modules.length, 4);
  assert.equal(view.modules[0].sources.length, 2);
  assert.equal(view.notes.length, 4);
  assert.equal(view.notes[1].sources[0].sourceId, 'proof-rule');
});
