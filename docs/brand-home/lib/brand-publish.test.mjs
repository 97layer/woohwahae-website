import assert from 'node:assert/strict';
import test from 'node:test';

import {
  buildThreadsPublishCandidates,
  buildThreadsPublishReceipts,
  parseBrandPublishObservationRaw,
} from './brand-publish-shape.js';

test('brand publish raw parser keeps draft text and receipt fields intact', () => {
  const parsed = parseBrandPublishObservationRaw([
    'channel=threads',
    'target_account=97layer',
    'title=Launch note',
    'sources=brand_story, proof_launch',
    'topic_tag=vibe coding',
    'media=absorbed-hero-field',
    'media_cue=Absorbed field graphic: A field study that still fits the quieter public shell.',
    'style_profile=threads-provisional-v1',
    'style_cue=open with one concrete observation / keep paragraphs short / avoid stacked slogans',
    'style_examples=legacy-ig-01,legacy-ig-05',
    'creation_id=creation_001',
    'thread_id=thread_001',
    'published_at=2026-03-11T01:02:03Z',
    'draft:',
    'We rebuilt the operating surface.',
  ].join('\n'));

  assert.deepEqual(parsed, {
    channel: 'threads',
    targetAccount: '97layer',
    title: 'Launch note',
    body: 'We rebuilt the operating surface.',
    topicTag: 'vibe coding',
    sourceIds: ['brand_story', 'proof_launch'],
    mediaIds: ['absorbed-hero-field'],
    styleExampleIds: ['legacy-ig-01', 'legacy-ig-05'],
    notes: [],
    mediaCue: 'Absorbed field graphic: A field study that still fits the quieter public shell.',
    styleProfile: 'threads-provisional-v1',
    styleCue: 'open with one concrete observation / keep paragraphs short / avoid stacked slogans',
    creationId: 'creation_001',
    threadId: 'thread_001',
    publishedAt: '2026-03-11T01:02:03Z',
  });
});

test('threads candidates only surface approved unpublished drafts', () => {
  const candidates = buildThreadsPublishCandidates({
    proposals: [
      {
        proposalId: 'proposal_001',
        promotedWorkItemId: 'work_001',
      },
    ],
    approvals: [
      {
        approvalId: 'approval_001',
        workItemId: 'work_001',
        status: 'approved',
        resolvedAt: '2026-03-11T02:00:00Z',
        summary: 'ready',
      },
      {
        approvalId: 'approval_002',
        workItemId: 'work_002',
        status: 'pending',
        resolvedAt: null,
        summary: 'pending',
      },
    ],
    flows: [
      {
        flowId: 'flow_001',
        workItemId: 'work_001',
      },
    ],
    observations: [
      {
        topic: 'brand_publish_prep',
        observation_id: 'observation_001',
        refs: ['proposal_001', 'work_001', 'approval_001', 'flow_001'],
        raw_excerpt: 'channel=threads\ntarget_account=97layer\ntitle=Launch note\ndraft:\nWe rebuilt the operating surface.',
      },
    ],
    receipts: [],
  });

  assert.deepEqual(candidates, [
    {
      approvalId: 'approval_001',
      workItemId: 'work_001',
        flowId: 'flow_001',
      proposalId: 'proposal_001',
        observationId: 'observation_001',
        targetAccount: '97layer',
        title: 'Launch note',
        bodyPreview: 'We rebuilt the operating surface.',
        topicTag: '',
        sourceIds: [],
        mediaIds: [],
        styleExampleIds: [],
        styleProfile: '',
        styleCue: '',
        resolvedAt: '2026-03-11T02:00:00Z',
        channel: 'threads',
        published: false,
      },
  ]);
});

test('threads receipts keep the latest publish evidence together', () => {
  const receipts = buildThreadsPublishReceipts([
    {
      topic: 'brand_publish_threads',
      observation_id: 'observation_101',
      normalized_summary: 'Published approved Threads draft: Launch note.',
      captured_at: '2026-03-11T03:00:00Z',
      refs: ['approval_001', 'proposal_001', 'work_001', 'flow_001'],
      raw_excerpt: 'channel=threads\ntarget_account=97layer\ntitle=Launch note\ncreation_id=creation_001\nthread_id=thread_001\npublished_at=2026-03-11T03:00:00Z\ndraft:\nhello',
    },
  ]);

  assert.deepEqual(receipts, [
    {
      observationId: 'observation_101',
      approvalId: 'approval_001',
      proposalId: 'proposal_001',
      workItemId: 'work_001',
      flowId: 'flow_001',
        targetAccount: '97layer',
        title: 'Launch note',
        creationId: 'creation_001',
        threadId: 'thread_001',
        topicTag: '',
        sourceIds: [],
        mediaIds: [],
        styleExampleIds: [],
        styleProfile: '',
        styleCue: '',
        publishedAt: '2026-03-11T03:00:00Z',
        summary: 'Published approved Threads draft: Launch note.',
      },
    ]);
});
