import assert from 'node:assert/strict';
import test from 'node:test';

import { parseBrandPublishObservationRaw } from './brand-publish-shape.js';
import { buildSourceDraftPrepShape } from './source-intake-prep-shape.js';
import { parseSourceDraftSeedRaw, parseSourceIntakeRaw, parseSourceIntakeRouteDecisionRaw } from './source-intake-shape.js';
import { mergeSourceIntakeDraftSeeds, mergeSourceIntakePrepLanes, mergeSourceIntakeRouteDecisions } from './source-intake-draft-merge.js';

test('source draft seed raw parser keeps account and draft preview fields intact', () => {
  const parsed = parseSourceDraftSeedRaw([
    'target_account=woosunhokr',
    'target_tone=refined',
    'title=우순호 draft · beauty practice note',
    'source_observation_id=observation_source_002',
    'route_decision_id=observation_route_001',
    'source_title=beauty practice note',
    'source_url=https://example.com/note',
    'founder_note=미용 태도 쪽으로 번역',
    'domain_tags=beauty,practice',
    'worldview_tags=identity,subtraction',
    'draft:',
    'beauty practice note를 보면서 다시 느낀 건 실무는 손기술보다 기준이 먼저라는 점이다.',
    '',
    '우순호 쪽에서는 이 소재를 미용사의 단상으로 더 정리해볼 수 있다.',
  ].join('\n'));

  assert.deepEqual(parsed, {
    targetAccount: 'woosunhokr',
    targetTone: 'refined',
    title: '우순호 draft · beauty practice note',
    sourceObservationId: 'observation_source_002',
    routeDecisionId: 'observation_route_001',
    parentDraftObservationId: '',
    revisionNote: '',
    sourceTitle: 'beauty practice note',
    sourceURL: 'https://example.com/note',
    founderNote: '미용 태도 쪽으로 번역',
    domainTags: ['beauty', 'practice'],
    worldviewTags: ['identity', 'subtraction'],
    draft:
      'beauty practice note를 보면서 다시 느낀 건 실무는 손기술보다 기준이 먼저라는 점이다.\n\n우순호 쪽에서는 이 소재를 미용사의 단상으로 더 정리해볼 수 있다.',
  });
});

test('source intake raw parser keeps priority and disposition fields intact', () => {
  const parsed = parseSourceIntakeRaw([
    'intake_class=public_collector',
    'policy_color=yellow',
    'title=ambiguous field note',
    'url=https://example.com/note',
    'founder_note=route cue가 비어 있어 founder 분류가 필요합니다.',
    'priority_score=72',
    'disposition=review',
    'disposition_note=route cue가 비어 있어 founder route 분류가 필요합니다.',
    'domain_tags=beauty,system',
    'worldview_tags=subtraction',
    'suggested_routes=97layer',
    'excerpt:',
    'A public signal arrived without a clear publishing lane.',
  ].join('\n'));

  assert.deepEqual(parsed, {
    intakeClass: 'public_collector',
    policyColor: 'yellow',
    title: 'ambiguous field note',
    url: 'https://example.com/note',
    excerpt: 'A public signal arrived without a clear publishing lane.',
    founderNote: 'route cue가 비어 있어 founder 분류가 필요합니다.',
    priorityScore: 72,
    disposition: 'review',
    dispositionNote: 'route cue가 비어 있어 founder route 분류가 필요합니다.',
    domainTags: ['beauty', 'system'],
    worldviewTags: ['subtraction'],
    suggestedRoutes: ['97layer'],
  });
});

test('mergeSourceIntakeDraftSeeds keeps the latest draft per source/account and preserves multiple accounts on one source', () => {
  const items = [{ observationId: 'observation_source_002', title: 'source' }];
  const drafts = [
    {
      observationId: 'draft_latest_97',
      sourceObservationId: 'observation_source_002',
      targetAccount: '97layer',
      targetAccountLabel: '97layer',
      targetTone: 'raw',
      preview: 'latest 97 preview',
      capturedAt: '2026-03-12T00:00:02Z',
    },
    {
      observationId: 'draft_old_97',
      sourceObservationId: 'observation_source_002',
      targetAccount: '97layer',
      targetAccountLabel: '97layer',
      targetTone: 'raw',
      preview: 'old 97 preview',
      capturedAt: '2026-03-12T00:00:01Z',
    },
    {
      observationId: 'draft_woo',
      sourceObservationId: 'observation_source_002',
      targetAccount: 'woosunhokr',
      targetAccountLabel: '우순호',
      targetTone: 'refined',
      preview: 'woo preview',
      capturedAt: '2026-03-12T00:00:03Z',
    },
  ];

  const merged = mergeSourceIntakeDraftSeeds(items, drafts);
  assert.equal(merged.length, 1);
  assert.equal(merged[0].draftSeeds.length, 2);
  assert.equal(merged[0].draftSeeds[0].observationId, 'draft_latest_97');
  assert.equal(merged[0].draftSeeds[1].observationId, 'draft_woo');
});

test('source intake route decision parser keeps decision and route source intact', () => {
  const parsed = parseSourceIntakeRouteDecisionRaw([
    'source_observation_id=observation_source_002',
    'decision=woohwahae',
    'title=slow note',
    'summary=Source intake · slow note -> 우화해',
    'route_source=feed_sensor',
  ].join('\n'));

  assert.deepEqual(parsed, {
    sourceObservationId: 'observation_source_002',
    decision: 'woohwahae',
    title: 'slow note',
    summary: 'Source intake · slow note -> 우화해',
    routeSource: 'feed_sensor',
  });
});

test('mergeSourceIntakeRouteDecisions keeps the latest route decision per source', () => {
  const items = [{ observationId: 'observation_source_002', title: 'source' }];
  const decisions = [
    {
      observationId: 'route_old',
      sourceObservationId: 'observation_source_002',
      decision: '97layer',
      routeSource: 'founder',
      capturedAt: '2026-03-12T00:00:01Z',
    },
    {
      observationId: 'route_latest',
      sourceObservationId: 'observation_source_002',
      decision: 'woohwahae',
      routeSource: 'feed_sensor',
      capturedAt: '2026-03-12T00:00:03Z',
    },
  ];

  const merged = mergeSourceIntakeRouteDecisions(items, decisions);
  assert.equal(merged.length, 1);
  assert.equal(merged[0].routeDecisions.length, 2);
  assert.equal(merged[0].routeDecision.observationId, 'route_latest');
  assert.equal(merged[0].routeDecision.routeSource, 'feed_sensor');
});

test('brand publish prep parser keeps source linkage intact', () => {
  const parsed = parseBrandPublishObservationRaw([
    'channel=threads',
    'target_account=woohwahae',
    'title=slow note',
    'sources=observation_source_002',
    'notes=lane:brand_publish,source_draft_seed:observation_draft_001',
    'draft:',
    '조용한 생활의 리듬을 다시 붙잡는 문장이다.',
  ].join('\n'));

  assert.deepEqual(parsed, {
    channel: 'threads',
    targetAccount: 'woohwahae',
    title: 'slow note',
    body: '조용한 생활의 리듬을 다시 붙잡는 문장이다.',
    topicTag: '',
    sourceIds: ['observation_source_002'],
    mediaIds: [],
    styleExampleIds: [],
    notes: ['lane:brand_publish', 'source_draft_seed:observation_draft_001'],
    mediaCue: '',
    styleProfile: '',
    styleCue: '',
    creationId: '',
    threadId: '',
    publishedAt: '',
  });
});

test('mergeSourceIntakePrepLanes keeps the latest prep lane per source/account/channel', () => {
  const items = [{ observationId: 'observation_source_002', title: 'source' }];
  const preps = [
    {
      observationId: 'prep_old',
      sourceObservationId: 'observation_source_002',
      targetAccount: 'woohwahae',
      targetAccountLabel: '우화해',
      channel: 'threads',
      channelLabel: 'Threads',
      capturedAt: '2026-03-12T00:00:01Z',
    },
    {
      observationId: 'prep_latest',
      sourceObservationId: 'observation_source_002',
      targetAccount: 'woohwahae',
      targetAccountLabel: '우화해',
      channel: 'threads',
      channelLabel: 'Threads',
      capturedAt: '2026-03-12T00:00:03Z',
    },
  ];

  const merged = mergeSourceIntakePrepLanes(items, preps);
  assert.equal(merged.length, 1);
  assert.equal(merged[0].prepLanes.length, 1);
  assert.equal(merged[0].prepLane.observationId, 'prep_latest');
  assert.equal(merged[0].prepLane.channelLabel, 'Threads');
});

test('buildSourceDraftPrepShape keeps account voice while removing internal draft framing', () => {
  const rawPrep = buildSourceDraftPrepShape({
    targetAccount: '97layer',
    title: '97layer raw draft · operating surface rebuild',
    sourceTitle: 'operating surface rebuild',
    founderNote: '브랜드 구축 과정에 더 가깝게',
    domainTags: ['system', 'brand'],
    worldviewTags: ['identity'],
    draft: [
      '요즘 붙들고 있는 건 operating surface rebuild 안에 들어 있는 기능보다 구조와 순서다.',
      '홈페이지를 다시 세우고 운영 페이지를 분리하면서, 결국 마지막에 남는 건 기능보다 구조라는 걸 다시 느낀다.',
      '브랜드 구축 과정에 더 가깝게',
      '97layer에서는 이걸 아직 정리보다 기록에 가까운 메모로 둔다.',
    ].join('\n\n'),
  });

  assert.equal(rawPrep.title, 'operating surface rebuild');
  assert.match(rawPrep.body, /요즘 자꾸 다시 보게 되는 건 operating surface rebuild 안에 들어 있는 기능보다 구조와 순서 쪽이다\./);
  assert.match(rawPrep.body, /홈페이지를 다시 세우고 운영 페이지를 분리하면서/);
  assert.match(rawPrep.body, /브랜드 구축 과정에 더 가깝게라는 축을 더 또렷하게 적어둔다\./);
  assert.doesNotMatch(rawPrep.body, /97layer에서는/);
  assert.equal(rawPrep.prepShapeId, 'source_draft_threads_v1');

  const refinedPrep = buildSourceDraftPrepShape({
    targetAccount: 'woosunhokr',
    title: '우순호 draft · quiet beauty note',
    sourceTitle: 'quiet beauty note',
    founderNote: '미용사의 단상에 더 가깝게',
    revisionNote: '조금 더 구체적으로',
    domainTags: ['beauty', 'practice'],
    worldviewTags: ['subtraction'],
    draft: [
      'quiet beauty note를 우순호 쪽으로 옮기면 결국 기준과 손기술 사이의 감각이 사람에게 어떻게 닿는지가 먼저 보인다.',
      '실무에서는 결국 손기술보다 기준이 먼저 남고, 그 기준이 손의 리듬을 바꾼다는 생각이 든다.',
      '미용사의 단상에 더 가깝게',
      '우순호에서는 이 소재를 미용사의 단상, 태도, 실무 감각이 같이 보이는 문장으로 더 정리해본다.',
    ].join('\n\n'),
  });

  assert.equal(refinedPrep.title, 'quiet beauty note');
  assert.match(refinedPrep.body, /quiet beauty note를 보다 보면 결국 기준과 손기술 사이의 감각에 가까운 순간이 사람에게 어떻게 닿는지가 먼저 남는다\./);
  assert.match(refinedPrep.body, /실무에서는 결국 손기술보다 기준이 먼저 남고/);
  assert.match(refinedPrep.body, /이번 버전은 조금 더 구체적으로라는 요청을 따라 더 정제해본다\./);
  assert.doesNotMatch(refinedPrep.body, /우순호에서는/);

  const polishedPrep = buildSourceDraftPrepShape({
    targetAccount: 'woohwahae',
    title: '우화해 draft · slow note',
    sourceTitle: 'slow note',
    founderNote: '슬로우 라이프 문장으로',
    domainTags: ['brand'],
    worldviewTags: ['subtraction'],
    draft: [
      'slow note를 우화해 쪽으로 옮기면 결국 기준을 덜어내며 선명하게 만드는 감각을 덜어낸 뒤 남는 생활의 리듬이 더 중요해진다.',
      '조용한 생활의 리듬을 다시 붙잡는 문장은 결국 덜어낸 뒤에야 보인다고 느낀다.',
      '슬로우 라이프 문장으로',
      '우화해에서는 이 소재를 슬로우 라이프, 태도, 조용한 공적 문장 쪽으로 더 다듬는다.',
    ].join('\n\n'),
  });

  assert.equal(polishedPrep.title, 'slow note');
  assert.match(polishedPrep.body, /slow note에서 덜어내고 남는 건 결국 기준을 덜어내며 선명하게 만드는 감각에 가까운 리듬이 생활 안으로 번지는 순간이다\./);
  assert.match(polishedPrep.body, /조용한 생활의 리듬을 다시 붙잡는 문장은 결국 덜어낸 뒤에야 보인다고 느낀다\./);
  assert.match(polishedPrep.body, /이 소재는 슬로우 라이프 문장으로라는 결로 더 덜어낼 만하다\./);
  assert.doesNotMatch(polishedPrep.body, /우화해에서는/);
});
