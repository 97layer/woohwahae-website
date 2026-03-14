import { NextResponse } from 'next/server';

import { withRequestLayerOsWriteToken } from '../../../../../../lib/auth/layer-os-write-token';
import { authorizeRequest } from '../../../../../../lib/auth/rbac';
import { readValidatedJson } from '../../../../../../lib/http/request';
import { fetchLayerOs } from '../../../../../../lib/runtime/layer-os';
import { adminSourceDraftLaneSchema } from '../../../../../../lib/validation/admin-source-draft-lane';

function asText(value, fallback = '') {
  return typeof value === 'string' ? value.trim() : fallback;
}

function asArray(value) {
  return Array.isArray(value) ? value : [];
}

function uniqueStrings(values) {
  return Array.from(new Set(asArray(values).map((value) => asText(value)).filter(Boolean)));
}

function normalizeAttention(attention) {
  if (!attention || typeof attention !== 'object') {
    return null;
  }
  return {
    mode: asText(attention.mode),
    summary: asText(attention.summary),
    detail: asText(attention.detail),
    ref: asText(attention.ref),
    sourceObservationId: asText(attention.sourceObservationId || attention.source_observation_id),
    draftObservationId: asText(attention.draftObservationId || attention.draft_observation_id),
    actionLabel: asText(attention.actionLabel || attention.action_label),
  };
}

function normalizeDraftPrepPayload(payload) {
  const openedFromDraft = payload?.opened_from_draft || {};
  const lane = payload?.lane || {};
  return {
    ...payload,
    follow_up: normalizeAttention(payload?.follow_up),
    opened_from_draft: {
      draftObservationId: asText(openedFromDraft?.draft_observation_id),
      targetAccount: asText(openedFromDraft?.target_account),
      targetAccountLabel: asText(openedFromDraft?.target_account_label),
      targetTone: asText(openedFromDraft?.target_tone),
      title: asText(openedFromDraft?.title),
      bodyPreview: asText(openedFromDraft?.body_preview),
      revisionNote: asText(openedFromDraft?.revision_note),
    },
    lane: {
      channel: asText(lane?.channel),
      label: asText(lane?.label),
      targetAccount: asText(lane?.target_account),
      targetAccountLabel: asText(lane?.target_account_label),
      targetToneLevel: asText(lane?.target_tone_level),
      title: asText(lane?.title),
      bodyPreview: asText(lane?.body_preview),
      sourceIds: uniqueStrings(lane?.source_ids),
      externalRefs: uniqueStrings(lane?.external_refs),
      prepShapeId: asText(lane?.prep_shape_id),
    },
    reused: Boolean(payload?.reused),
  };
}

export async function POST(request) {
  const auth = await authorizeRequest(request, ['founder']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  const parsed = await readValidatedJson(request, adminSourceDraftLaneSchema);
  if (!parsed.ok) {
    return NextResponse.json({ error: parsed.error, details: parsed.details }, { status: parsed.status });
  }

  return withRequestLayerOsWriteToken(request, async () => {
    try {
      const payload = await fetchLayerOs('/api/layer-os/source-intake/drafts/prep', {
        method: 'POST',
        requireWriteToken: true,
        json: {
          draft_observation_id: parsed.data.draft_observation_id,
          channel: asText(parsed.data.channel, 'threads'),
        },
      });
      return NextResponse.json(normalizeDraftPrepPayload(payload));
    } catch (error) {
      const message = error instanceof Error ? error.message : 'source draft prep open failed';
      const status = message.includes('missing server write token') ? 503 : 502;
      return NextResponse.json({ error: message }, { status });
    }
  });
}
