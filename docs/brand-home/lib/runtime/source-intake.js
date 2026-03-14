import 'server-only';

import { fetchLayerOs, safeLayerOs } from './layer-os';

function asArray(value) {
  return Array.isArray(value) ? value : [];
}

function asText(value, fallback = '') {
  return typeof value === 'string' ? value.trim() : fallback;
}

function uniqueStrings(values) {
  return Array.from(new Set(asArray(values).map((value) => asText(value)).filter(Boolean)));
}

function normalizeSourceIntakeAttention(attention) {
  if (!attention || typeof attention !== 'object') {
    return null;
  }
  return {
    mode: asText(attention.mode),
    summary: asText(attention.summary),
    detail: asText(attention.detail),
    ref: asText(attention.ref),
    sourceObservationId: asText(attention.sourceObservationId),
    draftObservationId: asText(attention.draftObservationId),
    actionLabel: asText(attention.actionLabel),
  };
}

function sourceIntakeRefs(input) {
  return uniqueStrings([
    ...uniqueStrings(input.suggestedRoutes).map((route) => `route:${route}`),
    ...uniqueStrings(input.domainTags).map((tag) => `domain:${tag}`),
    ...uniqueStrings(input.worldviewTags).map((tag) => `worldview:${tag}`),
  ]);
}

function buildSourceIntakeAttention() {
  return {
    mode: 'drop_source',
    summary: '링크나 메모를 하나 저장해 source intake를 시작하세요.',
    detail: '지금은 founder가 던진 raw source를 먼저 정규화해 쌓는 단계입니다.',
    actionLabel: 'source intake 저장',
    ref: '',
    draftObservationId: '',
    sourceObservationId: '',
  };
}

function buildDegradedSourceIntakeView(degradedReason) {
  return {
    generatedAt: new Date().toISOString(),
    runtimeAvailable: false,
    degradedReason,
    summaryNote: '링크나 텍스트를 하나 저장하면 source intake가 여기부터 쌓입니다.',
    summaryMeta: 'action 0 · quiet 0 · drafts 0 · prep 0',
    quietNote: 'quiet candidate는 아직 없습니다.',
    defaults: {
      intakeClass: 'manual_drop',
      policyColor: 'green',
      title: '',
      url: '',
      excerpt: '',
      founderNote: '',
      priorityScore: 0,
      disposition: 'observe',
      dispositionNote: '',
      domainTags: [],
      worldviewTags: [],
      suggestedRoutes: ['97layer'],
    },
    routeOptions: [
      { routeId: '97layer', label: '97layer', cue: 'raw intake / broad source hub' },
      { routeId: 'woosunhokr', label: '우순호', cue: 'beauty craft / refined translation' },
      { routeId: 'woohwahae', label: '우화해', cue: 'slow magazine / polished public shell' },
    ],
    intakeClasses: [
      { value: 'manual_drop', label: 'manual drop' },
      { value: 'authorized_connector', label: 'authorized connector' },
      { value: 'public_collector', label: 'public collector' },
    ],
    policyColors: [
      { value: 'green', label: 'green' },
      { value: 'yellow', label: 'yellow' },
      { value: 'red', label: 'red' },
    ],
    items: [],
    actionCount: 0,
    quietCount: 0,
    quietItems: [],
    drafts: [],
    routeDecisions: [],
    prepLanes: [],
    attention: buildSourceIntakeAttention(),
  };
}

function buildSourceIntakeRequestPayload(input) {
  const requestPayload = {
    intakeClass: asText(input?.intake_class || input?.intakeClass, 'manual_drop'),
    policyColor: asText(input?.policy_color || input?.policyColor, 'green'),
    title: asText(input?.title),
    url: asText(input?.url),
    excerpt: asText(input?.excerpt),
    founderNote: asText(input?.founder_note || input?.founderNote),
    domainTags: uniqueStrings(input?.domain_tags || input?.domainTags),
    worldviewTags: uniqueStrings(input?.worldview_tags || input?.worldviewTags),
    suggestedRoutes: uniqueStrings(input?.suggested_routes || input?.suggestedRoutes),
  };

  return {
    source_channel: 'cockpit',
    confidence: 'high',
    intake_class: requestPayload.intakeClass,
    policy_color: requestPayload.policyColor,
    title: requestPayload.title,
    url: requestPayload.url,
    excerpt: requestPayload.excerpt,
    founder_note: requestPayload.founderNote,
    domain_tags: requestPayload.domainTags,
    worldview_tags: requestPayload.worldviewTags,
    suggested_routes: requestPayload.suggestedRoutes,
    refs: sourceIntakeRefs(requestPayload),
  };
}

export async function getAdminSourceIntakeView() {
  const timeoutMs = 2000;
  const payloadResult = await safeLayerOs('/api/layer-os/source-intake', { timeoutMs });
  if (payloadResult.ok) {
    return {
      ...payloadResult.payload,
      attention: normalizeSourceIntakeAttention(payloadResult.payload?.attention),
    };
  }

  return buildDegradedSourceIntakeView(payloadResult.error);
}

export async function submitSourceIntake(input) {
  const requestPayload = buildSourceIntakeRequestPayload(input);
  const payload = await fetchLayerOs('/api/layer-os/source-intake', {
    method: 'POST',
    requireWriteToken: true,
    json: requestPayload,
  });
  return {
    ...payload,
    next_action: normalizeSourceIntakeAttention(payload?.next_action),
  };
}
