import 'server-only';

import { publicHomeLocalSourcePack } from '../../content/public-home-source';
import { defaultSocialAccount, socialAccountProfile, socialAccountProfiles } from '../../content/social-account-profiles';
import { buildBrandPublishPresets as composeBrandPublishPresets } from '../brand-publish-presets';
import { buildThreadsPublishCandidates, buildThreadsPublishReceipts } from '../brand-publish-shape';
import { normalizeBrandSourcePack } from '../brand-source-pack';
import { fetchLayerOs, safeLayerOs } from './layer-os';

const BRAND_LANE_NOTE = 'lane:brand_publish';
const DEFAULT_PRIORITY = 'high';
const DEFAULT_RISK = 'medium';

function asArray(value) {
  return Array.isArray(value) ? value : [];
}

function asText(value, fallback = '') {
  return typeof value === 'string' ? value.trim() : fallback;
}

function uniqueStrings(values) {
  return Array.from(new Set(asArray(values).map((value) => asText(value)).filter(Boolean)));
}

function toMillis(value) {
  if (!value) return 0;
  const numeric = new Date(value).getTime();
  return Number.isFinite(numeric) ? numeric : 0;
}

function slugify(value, fallback = 'draft') {
  const slug = asText(value)
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 24);
  return slug || fallback;
}

function limitText(value, max) {
  const text = asText(value);
  if (!max || text.length <= max) {
    return text;
  }
  return `${text.slice(0, Math.max(0, max - 1)).trim()}…`;
}

function channelLabel(channel) {
  switch (asText(channel).toLowerCase()) {
    case 'threads':
      return 'Threads';
    case 'x':
      return 'X';
    case 'telegram':
      return 'Telegram';
    default:
      return asText(channel, 'channel');
  }
}

function sourceMapFromPack(pack) {
  return new Map(asArray(pack?.sources).map((source) => [source.sourceId, source]));
}

function resolveSourceIds(pack, requested = []) {
  const sourceMap = sourceMapFromPack(pack);
  const preferred = uniqueStrings(requested).filter((sourceId) => sourceMap.has(sourceId));
  if (preferred.length > 0) {
    return preferred;
  }

  const heroIds = asArray(pack?.sections?.hero?.sourceIds);
  const moduleIds = asArray(pack?.sections?.modules).flatMap((item) => asArray(item?.sourceIds));
  const noteIds = asArray(pack?.sections?.notes).flatMap((item) => asArray(item?.sourceIds));
  return uniqueStrings([...heroIds, ...moduleIds, ...noteIds]).filter((sourceId) => sourceMap.has(sourceId)).slice(0, 3);
}

function sourceLabels(pack, sourceIds) {
  const sourceMap = sourceMapFromPack(pack);
  return uniqueStrings(sourceIds)
    .map((sourceId) => sourceMap.get(sourceId))
    .filter(Boolean)
    .map((source) => source.title);
}

function channelProfile(pack, channel) {
  const profiles = pack?.channelProfiles || {};
  return profiles[asText(channel).toLowerCase()] || null;
}

function heroMedia(pack) {
  const mediaMap = new Map(asArray(pack?.media).map((item) => [item.mediaId, item]));
  const heroIds = asArray(pack?.sections?.hero?.mediaIds);
  return heroIds.map((mediaId) => mediaMap.get(mediaId)).find(Boolean) || asArray(pack?.media)[0] || null;
}

function profileCue(profile) {
  if (!profile) {
    return '';
  }
  const cue = asArray(profile.cues).slice(0, 3).join(' / ');
  return cue || asText(profile.summary);
}

function buildDraftBody(pack, channel) {
  const hero = pack?.sections?.hero || {};
  const modules = asArray(pack?.sections?.modules);
  const notes = asArray(pack?.sections?.notes);
  const headline = asText(hero.title).replace(/\s*\n+\s*/g, ' ').trim();
  const body = asText(hero.body);
  const proof = asText(modules[0]?.body || notes[0]?.text);
  const route = asText(modules[1]?.body || notes[1]?.text);

  if (channel === 'x') {
    return limitText(`${headline} ${body} ${proof}`.replace(/\s+/g, ' '), 250);
  }

  if (channel === 'telegram') {
    return [headline, body, proof, route].filter(Boolean).join('\n\n');
  }

  return [headline, body, proof].filter(Boolean).join('\n\n');
}

export function buildBrandPublishPresets(packInput = publicHomeLocalSourcePack) {
  return composeBrandPublishPresets(packInput);
}

function buildIDs(channel, title) {
  const stamp = new Date().toISOString().replace(/\D/g, '').slice(0, 14);
  const slug = slugify(title, channel || 'draft');
  const base = `brand_${asText(channel, 'draft').toLowerCase()}_${stamp}_${slug}`;
  return {
    proposalId: `proposal_${base}`,
    workItemId: `work_${base}`,
    approvalId: `approval_${base}`,
    flowId: `flow_${base}`,
  };
}

function buildLaneNotes(pack, channel, sourceIds, extraNotes = [], mediaIds = [], profile = null) {
  return uniqueStrings([
    BRAND_LANE_NOTE,
    `channel:${asText(channel).toLowerCase()}`,
    `source_pack:${pack.packId}`,
    ...uniqueStrings(sourceIds).map((sourceId) => `source:${sourceId}`),
    ...uniqueStrings(mediaIds).map((mediaId) => `media:${mediaId}`),
    profile?.profileId ? `style_profile:${profile.profileId}` : '',
    ...uniqueStrings(extraNotes),
  ]);
}

function brandProposalSummary(channel, accountLabel, title, body) {
  return limitText(`${channelLabel(channel)} draft for ${asText(accountLabel, 'account')} is ready for founder review: ${title}. ${body.replace(/\s+/g, ' ')}`, 180);
}

function buildObservationRawExcerpt({ channel, targetAccount = '', title, body, topicTag = '', sourceIds, notes, mediaIds = [], mediaCue = '', profile = null, styleExamples = [] }) {
  const lines = [
    `channel=${asText(channel).toLowerCase()}`,
    `target_account=${asText(targetAccount || noteValue(notes, 'account:') || defaultSocialAccount(channel))}`,
    `title=${asText(title)}`,
    `sources=${uniqueStrings(sourceIds).join(',') || 'none'}`,
  ];
  if (asText(topicTag)) {
    lines.push(`topic_tag=${asText(topicTag)}`);
  }
  if (uniqueStrings(mediaIds).length > 0) {
    lines.push(`media=${uniqueStrings(mediaIds).join(',')}`);
  }
  if (asText(mediaCue)) {
    lines.push(`media_cue=${asText(mediaCue)}`);
  }
  if (profile?.profileId) {
    lines.push(`style_profile=${profile.profileId}`);
  }
  if (profileCue(profile)) {
    lines.push(`style_cue=${profileCue(profile)}`);
  }
  if (uniqueStrings(asArray(styleExamples).map((item) => asText(item?.exampleId))).length > 0) {
    lines.push(`style_examples=${uniqueStrings(asArray(styleExamples).map((item) => asText(item?.exampleId))).join(',')}`);
  }
  if (uniqueStrings(notes).length > 0) {
    lines.push(`notes=${uniqueStrings(notes).join(',')}`);
  }
  lines.push('draft:');
  lines.push(asText(body));
  return lines.join('\n');
}

function noteHasPrefix(notes, prefix) {
  return asArray(notes).some((note) => asText(note).startsWith(prefix));
}

function noteValue(notes, prefix) {
  const note = asArray(notes).find((item) => asText(item).startsWith(prefix));
  return note ? asText(note).slice(prefix.length) : '';
}

function isBrandProposal(item) {
  return noteHasPrefix(item?.notes, BRAND_LANE_NOTE);
}

function isBrandFlow(item) {
  return noteHasPrefix(item?.notes, BRAND_LANE_NOTE);
}

function isBrandApproval(item) {
  return asText(item?.summary).toLowerCase().startsWith('brand publish');
}

function isBrandObservation(item) {
  return asText(item?.topic).toLowerCase() === 'brand_publish_prep';
}

function normalizeThreadsStatus(payload) {
  const status = payload?.status && typeof payload.status === 'object' ? payload.status : {};
  return {
    adapter: asText(status.adapter || payload?.adapter, 'noop'),
    publishConfigured: Boolean(status.publishConfigured ?? status.publish_configured ?? payload?.enabled),
    notes: asArray(status.notes),
  };
}

function buildBrandPublishAttention({ proposals = [], approvals = [], threadsEligible = [], threadsStatus = {} }) {
  if (threadsEligible.length > 0) {
    const candidate = threadsEligible[0];
    if (threadsStatus.publishConfigured) {
      return {
        mode: 'publish_threads',
        summary: `${candidate.title || '승인된 Threads 초안'}을 바로 publish할 수 있습니다.`,
        detail: `${candidate.targetAccount || 'threads'} · approval ${candidate.approvalId || 'n/a'}`,
        actionLabel: '바로 publish',
        approvalId: candidate.approvalId || '',
        ref: candidate.flowId || candidate.proposalId || candidate.approvalId || '',
      };
    }
    return {
      mode: 'restore_threads_delivery',
      summary: 'Threads 게시 자격을 먼저 복구하세요.',
      detail: `${candidate.title || '승인된 Threads 초안'}은 준비됐지만 access token이 없어 live publish가 막혀 있습니다.`,
      actionLabel: '설정 확인',
      approvalId: candidate.approvalId || '',
      ref: candidate.approvalId || '',
    };
  }

  const pendingApproval = approvals.find((item) => asText(item.status).toLowerCase() === 'pending');
  if (pendingApproval) {
    return {
      mode: 'review_brand_approval',
      summary: `${pendingApproval.summary || pendingApproval.approvalId} founder 승인을 먼저 결정하세요.`,
      detail: `approval ${pendingApproval.approvalId} · work ${pendingApproval.workItemId || 'n/a'}`,
      actionLabel: '승인 corridor 확인',
      approvalId: pendingApproval.approvalId || '',
      ref: pendingApproval.approvalId || '',
    };
  }

  if (proposals.length > 0) {
    const proposal = proposals[0];
    return {
      mode: 'review_brand_draft',
      summary: `${proposal.title || '최근 브랜드 초안'}을 먼저 읽고 founder review로 넘길지 결정하세요.`,
      detail: `${proposal.channel || 'draft'} · ${proposal.targetAccount || proposal.status || 'brand corridor'}`,
      actionLabel: '최근 초안 확인',
      approvalId: '',
      ref: proposal.proposalId || '',
    };
  }

  return {
    mode: 'open_brand_draft',
    summary: '브랜드 초안을 하나 열어 founder review corridor를 시작하세요.',
    detail: '지금은 preset으로 초안을 채우고 approval / flow를 여는 첫 단계입니다.',
    actionLabel: '승인 준비 열기',
    approvalId: '',
    ref: '',
  };
}

function buildBrandPublishFollowUp({ lane, proposal, approval, flow }) {
  return {
    mode: 'review_brand_approval',
    summary: `${lane?.targetAccountLabel || lane?.targetAccount || 'account'} ${lane?.label || lane?.channel || 'draft'} 초안을 founder 승인 corridor로 올렸습니다.`,
    detail: `${proposal?.proposal_id || 'proposal'} · ${approval?.approval_id || 'approval'} · ${flow?.flow_id || 'flow'}`,
    actionLabel: '승인 상태 확인',
    approvalId: approval?.approval_id || '',
    ref: flow?.flow_id || approval?.approval_id || proposal?.proposal_id || '',
  };
}

function buildThreadsPublishFollowUp(payload) {
  const title = asText(payload?.title, 'Threads publish');
  const threadId = asText(payload?.thread_id || payload?.threadId || 'pending');
  const approvalId = asText(payload?.approval_id || payload?.approvalId);
  return {
    mode: 'review_publish_receipt',
    summary: `${title} 게시 영수증이 생성됐습니다. 실제 thread id와 후속 반응을 확인하세요.`,
    detail: `thread ${threadId}${approvalId ? ` · approval ${approvalId}` : ''}`,
    actionLabel: '영수증 확인',
    approvalId,
    ref: threadId,
  };
}

export async function getAdminBrandPublishView() {
  const timeoutMs = 2000;
  const pack = normalizeBrandSourcePack(publicHomeLocalSourcePack);
  const presets = buildBrandPublishPresets(pack);
  const selectedMedia = heroMedia(pack);
  const threadsProfile = channelProfile(pack, 'threads');
  const [proposalsResult, approvalsResult, flowsResult, observationsResult, threadsResult, threadsReceiptsResult] = await Promise.all([
    safeLayerOs('/api/layer-os/proposals', { timeoutMs }),
    safeLayerOs('/api/layer-os/approval-inbox', { timeoutMs }),
    safeLayerOs('/api/layer-os/flows', { timeoutMs }),
    safeLayerOs('/api/layer-os/observations?topic=brand_publish_prep&limit=12', { timeoutMs }),
    safeLayerOs('/api/layer-os/social/threads', { timeoutMs }),
    safeLayerOs('/api/layer-os/observations?topic=brand_publish_threads&limit=12', { timeoutMs }),
  ]);
  const proposalsPayload = proposalsResult.ok ? proposalsResult.payload : {};
  const approvalsPayload = approvalsResult.ok ? approvalsResult.payload : {};
  const flowsPayload = flowsResult.ok ? flowsResult.payload : {};
  const observationsPayload = observationsResult.ok ? observationsResult.payload : {};
  const threadsPayload = threadsResult.ok ? threadsResult.payload : {};
  const threadsReceiptsPayload = threadsReceiptsResult.ok ? threadsReceiptsResult.payload : {};
  const runtimeFailures = [proposalsResult, approvalsResult, flowsResult, observationsResult, threadsResult, threadsReceiptsResult].filter((result) => !result.ok);

  const proposals = asArray(proposalsPayload?.items)
    .filter(isBrandProposal)
    .sort((left, right) => toMillis(right.updated_at || right.updatedAt || right.created_at) - toMillis(left.updated_at || left.updatedAt || left.created_at))
    .slice(0, 6)
    .map((item) => ({
      proposalId: item.proposal_id,
      title: item.title,
      status: item.status,
      summary: item.summary,
      updatedAt: item.updated_at || item.created_at || null,
      channel: noteValue(item.notes, 'channel:') || 'draft',
      targetAccount: noteValue(item.notes, 'account:') || '',
      promotedWorkItemId: item.promoted_work_item_id || null,
      sourceIds: asArray(item.notes)
        .filter((note) => asText(note).startsWith('source:'))
        .map((note) => asText(note).slice('source:'.length)),
      styleProfile: noteValue(item.notes, 'style_profile:') || '',
    }));

  const approvals = asArray(approvalsPayload?.items)
    .filter(isBrandApproval)
    .sort((left, right) => toMillis(right.resolved_at || right.requested_at) - toMillis(left.resolved_at || left.requested_at))
    .slice(0, 6)
    .map((item) => ({
      approvalId: item.approval_id,
      workItemId: item.work_item_id,
      status: item.status,
      summary: item.summary,
      requestedAt: item.requested_at,
      resolvedAt: item.resolved_at || null,
    }));

  const flows = asArray(flowsPayload?.items)
    .filter(isBrandFlow)
    .sort((left, right) => toMillis(right.updated_at) - toMillis(left.updated_at))
    .slice(0, 6)
    .map((item) => ({
      flowId: item.flow_id,
      workItemId: item.work_item_id,
      approvalId: item.approval_id || null,
      status: item.status,
      updatedAt: item.updated_at,
      channel: noteValue(item.notes, 'channel:') || 'draft',
      targetAccount: noteValue(item.notes, 'account:') || '',
    }));

  const observations = asArray(observationsPayload?.items)
    .filter(isBrandObservation)
    .sort((left, right) => toMillis(right.captured_at) - toMillis(left.captured_at))
    .slice(0, 6)
    .map((item) => ({
      observationId: item.observation_id,
      summary: item.normalized_summary,
      capturedAt: item.captured_at,
      refs: asArray(item.refs),
      rawExcerpt: item.raw_excerpt || '',
    }));

  const threadsReceipts = buildThreadsPublishReceipts(asArray(threadsReceiptsPayload?.items));
  const threadsEligible = buildThreadsPublishCandidates({
    proposals,
    approvals,
    flows,
    observations: asArray(observationsPayload?.items),
    receipts: asArray(threadsReceiptsPayload?.items),
  });

  return {
    generatedAt: new Date().toISOString(),
    runtimeAvailable: runtimeFailures.length === 0,
    degradedReason: runtimeFailures[0]?.error || '',
    sourcePack: {
      packId: pack.packId,
      label: pack.label,
      updatedAt: pack.updatedAt,
      voice: pack.voice,
      media: selectedMedia
        ? {
            mediaId: selectedMedia.mediaId,
            title: selectedMedia.title,
            src: selectedMedia.src,
            caption: selectedMedia.caption,
          }
        : null,
      threadsProfile: threadsProfile
        ? {
            profileId: threadsProfile.profileId,
            label: threadsProfile.label,
            summary: threadsProfile.summary,
            sourceMode: threadsProfile.sourceMode,
            exampleCount: asArray(threadsProfile.examples).length,
            examples: asArray(threadsProfile.examples)
              .slice(0, 2)
              .map((item) => ({
                exampleId: item.exampleId,
                signalId: item.signalId,
                publishedAt: item.publishedAt,
                excerpt: item.excerpt,
              })),
            provenanceLabel: threadsProfile.provenance?.label || '',
          }
        : null,
      sources: asArray(pack.sources).map((source) => ({
        sourceId: source.sourceId,
        title: source.title,
      })),
    },
    defaults: {
      channel: presets[0]?.channel || 'threads',
      targetAccount: presets[0]?.targetAccount || defaultSocialAccount(presets[0]?.channel || 'threads'),
      title: presets[0]?.title || '',
      body: presets[0]?.body || '',
      topicTag: presets[0]?.topicTag || '',
      sourceIds: presets[0]?.sourceIds || [],
      mediaIds: presets[0]?.mediaIds || [],
      styleProfile: presets[0]?.styleProfile || '',
      styleCue: presets[0]?.styleCue || '',
      styleExamples: presets[0]?.styleExamples || [],
      notes: [],
    },
    accounts: socialAccountProfiles,
    presets,
    recent: {
      proposals,
      approvals,
      flows,
      observations,
    },
    attention: buildBrandPublishAttention({
      proposals,
      approvals,
      threadsEligible,
      threadsStatus: normalizeThreadsStatus(threadsPayload),
    }),
    threads: {
      status: normalizeThreadsStatus(threadsPayload),
      eligible: threadsEligible,
      recent: threadsReceipts,
    },
  };
}

export async function submitBrandPublish(input) {
  const pack = normalizeBrandSourcePack(publicHomeLocalSourcePack);
  const channel = asText(input?.channel, 'threads').toLowerCase();
  const targetAccount = asText(input?.target_account || input?.targetAccount || defaultSocialAccount(channel));
  const title = asText(input?.title);
  const body = asText(input?.body);
  const topicTag = asText(input?.topic_tag || input?.topicTag);
  const sourceIds = resolveSourceIds(pack, input?.source_ids || input?.sourceIds);
  const externalRefs = uniqueStrings(input?.external_refs || input?.externalRefs);
  const preset = buildBrandPublishPresets(pack).find((item) => item.channel === channel) || null;
  const mediaIds = uniqueStrings(input?.media_ids || input?.mediaIds || preset?.mediaIds);
  const mediaCue = asText(input?.media_cue || input?.mediaCue || preset?.mediaCue);
  const profile = channelProfile(pack, channel) || channelProfile(pack, channel === 'x' ? 'instagram' : '');
  const account = socialAccountProfile(targetAccount);
  const styleExamples = asArray(preset?.styleExamples);
  const notes = buildLaneNotes(pack, channel, sourceIds, [...asArray(input?.notes), `account:${account.accountId}`, `tone:${account.toneLevel}`], mediaIds, profile);
  const ids = buildIDs(channel, title);
  const summary = brandProposalSummary(channel, account.label, title, body);
  const proposalTitle = limitText(`Brand publish · ${account.label} · ${channelLabel(channel)} · ${title}`, 160);

  const proposal = await fetchLayerOs('/api/layer-os/proposals', {
    method: 'POST',
    requireWriteToken: true,
    json: {
      proposal_id: ids.proposalId,
      title: proposalTitle,
      intent: limitText(`prepare ${channel} publication from the current brand pack`, 180),
      summary,
      surface: 'cockpit',
      priority: DEFAULT_PRIORITY,
      risk: DEFAULT_RISK,
      notes,
    },
  });

  const promotion = await fetchLayerOs('/api/layer-os/proposals/promote', {
    method: 'POST',
    requireWriteToken: true,
    json: {
      proposal_id: ids.proposalId,
      work_item_id: ids.workItemId,
    },
  });

  const approval = await fetchLayerOs('/api/layer-os/approval-inbox', {
    method: 'POST',
    requireWriteToken: true,
    json: {
      approval_id: ids.approvalId,
      work_item_id: ids.workItemId,
      stage: 'verify',
      summary: proposalTitle,
      risks: [
        `${channelLabel(channel)} wording needs founder approval`,
        'publish only after the draft matches brand proof and tone',
      ],
      rollback_plan: 'do not publish; revise the draft and reopen approval',
      decision_surface: 'cockpit',
      status: 'pending',
    },
  });

  const flow = await fetchLayerOs('/api/layer-os/flows/sync', {
    method: 'POST',
    requireWriteToken: true,
    json: {
      flow_id: ids.flowId,
      work_item_id: ids.workItemId,
      approval_id: ids.approvalId,
      notes,
    },
  });

  const observation = await fetchLayerOs('/api/layer-os/observations', {
    method: 'POST',
    requireWriteToken: true,
    json: {
      source_channel: 'cockpit',
      actor: 'admin',
      topic: 'brand_publish_prep',
      refs: uniqueStrings([ids.proposalId, ids.workItemId, ids.approvalId, ids.flowId, ...sourceIds, ...externalRefs]),
      confidence: 'high',
      raw_excerpt: buildObservationRawExcerpt({ channel, targetAccount: account.accountId, title, body, topicTag, sourceIds, notes, mediaIds, mediaCue, profile, styleExamples }),
      normalized_summary: summary,
    },
  });

  return {
    lane: {
      channel,
      label: channelLabel(channel),
      targetAccount: account.accountId,
      targetAccountLabel: account.label,
      targetToneLevel: account.toneLevel,
      targetSummary: account.summary,
      title,
      topicTag,
      sourceIds,
      externalRefs,
      mediaIds,
      mediaCue,
      styleProfile: profile?.profileId || '',
      styleCue: profileCue(profile),
      styleExamples,
      sourceLabels: sourceLabels(pack, sourceIds),
    },
    proposal,
    work_item: promotion?.work_item || null,
    approval,
    flow,
    observation,
    next_action: buildBrandPublishFollowUp({
      lane: {
        label: channelLabel(channel),
        channel,
        targetAccount: account.accountId,
        targetAccountLabel: account.label,
      },
      proposal,
      approval,
      flow,
    }),
  };
}

export async function publishBrandThreads(input) {
  const payload = await fetchLayerOs('/api/layer-os/social/threads', {
    method: 'POST',
    requireWriteToken: true,
    json: {
      approval_id: asText(input?.approval_id || input?.approvalId),
    },
  });
  return {
    ...payload,
    follow_up: buildThreadsPublishFollowUp(payload),
  };
}
