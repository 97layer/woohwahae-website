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

function limitText(value, max) {
  const text = asText(value);
  if (!max || text.length <= max) {
    return text;
  }
  return `${text.slice(0, Math.max(0, max - 1)).trim()}…`;
}

function isBrandObservation(item) {
  return asText(item?.topic).toLowerCase() === 'brand_publish_prep';
}

function isThreadsPublishObservation(item) {
  return asText(item?.topic).toLowerCase() === 'brand_publish_threads';
}

function refWithPrefix(refs, prefix) {
  return asArray(refs).find((ref) => asText(ref).startsWith(prefix)) || '';
}

export function parseBrandPublishObservationRaw(value) {
  const lines = String(value || '').replace(/\r\n/g, '\n').split('\n');
  const parsed = {
    channel: '',
    targetAccount: '',
    title: '',
    body: '',
    topicTag: '',
    sourceIds: [],
    mediaIds: [],
    styleExampleIds: [],
    notes: [],
    mediaCue: '',
    styleProfile: '',
    styleCue: '',
    creationId: '',
    threadId: '',
    publishedAt: '',
  };
  let bodyStart = -1;

  for (const [index, rawLine] of lines.entries()) {
    const line = asText(rawLine);
    if (!line) {
      continue;
    }
    if (line === 'draft:') {
      bodyStart = index + 1;
      continue;
    }
    if (line.startsWith('channel=')) {
      parsed.channel = asText(line.slice('channel='.length)).toLowerCase();
      continue;
    }
    if (line.startsWith('target_account=')) {
      parsed.targetAccount = asText(line.slice('target_account='.length));
      continue;
    }
    if (line.startsWith('title=')) {
      parsed.title = asText(line.slice('title='.length));
      continue;
    }
    if (line.startsWith('sources=')) {
      const rawSources = asText(line.slice('sources='.length));
      parsed.sourceIds = rawSources.toLowerCase() === 'none' ? [] : uniqueStrings(rawSources.split(','));
      continue;
    }
    if (line.startsWith('topic_tag=')) {
      parsed.topicTag = asText(line.slice('topic_tag='.length));
      continue;
    }
    if (line.startsWith('media=')) {
      const rawMedia = asText(line.slice('media='.length));
      parsed.mediaIds = rawMedia.toLowerCase() === 'none' ? [] : uniqueStrings(rawMedia.split(','));
      continue;
    }
    if (line.startsWith('notes=')) {
      const rawNotes = asText(line.slice('notes='.length));
      parsed.notes = rawNotes.toLowerCase() === 'none' ? [] : uniqueStrings(rawNotes.split(','));
      continue;
    }
    if (line.startsWith('media_cue=')) {
      parsed.mediaCue = asText(line.slice('media_cue='.length));
      continue;
    }
    if (line.startsWith('style_profile=')) {
      parsed.styleProfile = asText(line.slice('style_profile='.length));
      continue;
    }
    if (line.startsWith('style_cue=')) {
      parsed.styleCue = asText(line.slice('style_cue='.length));
      continue;
    }
    if (line.startsWith('style_examples=')) {
      const rawExamples = asText(line.slice('style_examples='.length));
      parsed.styleExampleIds = rawExamples.toLowerCase() === 'none' ? [] : uniqueStrings(rawExamples.split(','));
      continue;
    }
    if (line.startsWith('creation_id=')) {
      parsed.creationId = asText(line.slice('creation_id='.length));
      continue;
    }
    if (line.startsWith('thread_id=')) {
      parsed.threadId = asText(line.slice('thread_id='.length));
      continue;
    }
    if (line.startsWith('published_at=')) {
      parsed.publishedAt = asText(line.slice('published_at='.length));
    }
  }

  if (bodyStart >= 0 && bodyStart <= lines.length) {
    parsed.body = String(lines.slice(bodyStart).join('\n')).trim();
  }

  return parsed;
}

export function buildThreadsPublishReceipts(items = []) {
  return asArray(items)
    .filter(isThreadsPublishObservation)
    .sort((left, right) => toMillis(right.captured_at) - toMillis(left.captured_at))
    .map((item) => {
      const draft = parseBrandPublishObservationRaw(item.raw_excerpt);
      return {
        observationId: item.observation_id,
        approvalId: refWithPrefix(item.refs, 'approval_'),
        proposalId: refWithPrefix(item.refs, 'proposal_'),
        workItemId: refWithPrefix(item.refs, 'work_'),
        flowId: refWithPrefix(item.refs, 'flow_'),
        targetAccount: draft.targetAccount,
        title: draft.title || item.normalized_summary,
        creationId: draft.creationId,
        threadId: draft.threadId,
        topicTag: draft.topicTag,
        sourceIds: draft.sourceIds,
        mediaIds: draft.mediaIds,
        styleExampleIds: draft.styleExampleIds,
        styleProfile: draft.styleProfile,
        styleCue: draft.styleCue,
        publishedAt: draft.publishedAt || item.captured_at,
        summary: item.normalized_summary,
      };
    })
    .slice(0, 6);
}

export function buildThreadsPublishCandidates({ proposals = [], approvals = [], flows = [], observations = [], receipts = [] } = {}) {
  const proposalByWorkItemId = new Map(
    asArray(proposals)
      .filter((item) => asText(item?.promotedWorkItemId))
      .map((item) => [item.promotedWorkItemId, item]),
  );
  const flowByWorkItemId = new Map(
    asArray(flows)
      .filter((item) => asText(item?.workItemId))
      .map((item) => [item.workItemId, item]),
  );
  const observationByApprovalId = new Map();
  for (const item of asArray(observations)) {
    if (!isBrandObservation(item)) {
      continue;
    }
    const approvalId = refWithPrefix(item.refs, 'approval_');
    if (!approvalId || observationByApprovalId.has(approvalId)) {
      continue;
    }
    observationByApprovalId.set(approvalId, item);
  }
  const publishedApprovalIds = new Set(
    buildThreadsPublishReceipts(receipts)
      .map((item) => item.approvalId)
      .filter(Boolean),
  );

  return asArray(approvals)
    .filter((item) => asText(item?.status).toLowerCase() === 'approved')
    .map((item) => {
      const observation = observationByApprovalId.get(item.approvalId);
      const draft = parseBrandPublishObservationRaw(observation?.raw_excerpt);
      if (!observation || draft.channel !== 'threads') {
        return null;
      }
      const proposal = proposalByWorkItemId.get(item.workItemId);
      const flow = flowByWorkItemId.get(item.workItemId);
      return {
        approvalId: item.approvalId,
        workItemId: item.workItemId,
        flowId: flow?.flowId || refWithPrefix(observation.refs, 'flow_'),
        proposalId: proposal?.proposalId || refWithPrefix(observation.refs, 'proposal_'),
        observationId: observation.observation_id,
        targetAccount: draft.targetAccount,
        title: draft.title || proposal?.title || item.summary,
        bodyPreview: limitText(draft.body.replace(/\s+/g, ' '), 180),
        topicTag: draft.topicTag,
        sourceIds: draft.sourceIds,
        mediaIds: draft.mediaIds,
        styleExampleIds: draft.styleExampleIds,
        styleProfile: draft.styleProfile,
        styleCue: draft.styleCue,
        resolvedAt: item.resolvedAt,
        channel: draft.channel,
        published: publishedApprovalIds.has(item.approvalId),
      };
    })
    .filter(Boolean)
    .filter((item) => !item.published)
    .sort((left, right) => toMillis(right.resolvedAt) - toMillis(left.resolvedAt))
    .slice(0, 6);
}
