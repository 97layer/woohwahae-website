import 'server-only';

import { fetchLayerOs } from './layer-os';
import { getQuickworkRuntimeStatus, submitQuickwork } from './quickwork';

const laneOrder = ['planner', 'implementer', 'verifier', 'designer'];

const laneLabels = {
  planner: '정리',
  implementer: '구현',
  verifier: '검증',
  designer: '경험',
};

function asArray(value) {
  return Array.isArray(value) ? value : [];
}

function asRecord(value) {
  return value && typeof value === 'object' && !Array.isArray(value) ? value : {};
}

function asString(value) {
  return typeof value === 'string' && value.trim() ? value.trim() : '';
}

function toMillis(value) {
  if (!value) return 0;
  const numeric = new Date(value).getTime();
  return Number.isFinite(numeric) ? numeric : 0;
}

function uniqueStrings(items) {
  return [...new Set(asArray(items).map((item) => asString(item)).filter(Boolean))];
}

function safeError(error, fallback) {
  return error instanceof Error ? error.message : fallback;
}

const controlTowerTimeoutMs = 1800;

function statusTone(status) {
  switch (status) {
    case 'running':
    case 'succeeded':
      return 'good';
    case 'failed':
    case 'canceled':
    case 'cancelled':
      return 'alert';
    default:
      return 'muted';
  }
}

function humanizeStatus(status) {
  switch (status) {
    case 'queued':
      return '대기';
    case 'running':
      return '실행 중';
    case 'succeeded':
      return '성공';
    case 'failed':
      return '실패';
    case 'canceled':
    case 'cancelled':
      return '취소';
    default:
      return status || '미확인';
  }
}

function normalizeProfile(item) {
  return {
    role: asString(item?.role),
    provider: asString(item?.provider),
    model: asString(item?.model),
    dispatchReady: Boolean(item?.dispatch_ready ?? item?.dispatchReady),
    notes: uniqueStrings(item?.notes),
  };
}

function previewText(result) {
  return asString(result?.response_preview)
    || asString(result?.summary)
    || asString(result?.response)
    || '';
}

function normalizeFollowUp(value) {
  const item = asRecord(value);
  const mode = asString(item?.mode);
  const summary = asString(item?.summary);
  const jobIds = uniqueStrings(item?.job_ids || item?.jobIds);
  if (!mode && !summary && jobIds.length === 0) {
    return null;
  }
  return {
    mode,
    summary,
    jobIds,
  };
}

function normalizeJob(item) {
  const payload = asRecord(item?.payload);
  const result = asRecord(item?.result);
  const council = asRecord(result?.council);

  return {
    jobId: asString(item?.job_id || item?.jobId),
    summary: asString(item?.summary),
    role: asString(item?.role),
    kind: asString(item?.kind),
    stage: asString(item?.stage),
    source: asString(item?.source),
    surface: asString(item?.surface),
    status: asString(item?.status),
    ref: asString(item?.ref),
    createdAt: asString(item?.created_at || item?.createdAt),
    updatedAt: asString(item?.updated_at || item?.updatedAt),
    notes: uniqueStrings(item?.notes),
    title: asString(payload?.title),
    description: asString(payload?.description),
    issueId: asString(payload?.issue_id),
    issueIdentifier: asString(payload?.issue_identifier),
    issueTitle: asString(payload?.issue_title),
    allowedPaths: uniqueStrings(payload?.allowed_paths || payload?.allowedPaths),
    dispatchState: asString(result?.dispatch_state || result?.dispatchState),
    dispatchTransport: asString(result?.dispatch_transport || result?.dispatchTransport),
    executionOrigin: asString(result?.execution_origin || result?.executionOrigin),
    provider: asString(result?.provider),
    model: asString(result?.model),
    responsePreview: previewText(result),
    error: asString(result?.last_error) || asString(result?.error),
    packetPath: asString(result?.job_packet_path) || `/api/layer-os/jobs/packet?job_id=${encodeURIComponent(asString(item?.job_id || item?.jobId))}`,
    changedPaths: uniqueStrings(result?.changed_paths || result?.changedPaths),
    artifacts: uniqueStrings(result?.artifacts),
    followUp: normalizeFollowUp(result?.follow_up || result?.followUp),
    council: {
      requestedCount: Number(council?.requested_count || 0),
      succeededCount: Number(council?.succeeded_count || 0),
      primaryProvider: asString(council?.primary_provider),
    },
    tone: statusTone(asString(item?.status)),
    statusLabel: humanizeStatus(asString(item?.status)),
    isOpen: ['queued', 'running'].includes(asString(item?.status)),
  };
}

function summarizePacket(packet, selectedJob) {
  if (!packet) return null;

  const prompting = asRecord(packet.prompting);
  const knowledge = asRecord(packet.knowledge);
  const runtime = asRecord(packet.runtime);
  const handoffSummary = asRecord(packet.handoff_summary);
  const reviewTopOpen = asArray(handoffSummary.review_top_open)
    .map((item) => asString(item?.text))
    .filter(Boolean)
    .slice(0, 4);

  return {
    source: asString(packet.source),
    reportPath: asString(runtime.report_path),
    reportCommand: asString(runtime.report_command),
    writeAuthRequired: Boolean(runtime.write_auth_required),
    sourceOfTruth: asString(runtime.source_of_truth),
    dispatchTransport: asString(runtime.dispatch_transport),
    cognitionMode: asString(prompting.cognition_mode),
    decisionScope: asString(prompting.decision_scope),
    autonomyBudget: asString(prompting.autonomy_budget),
    mutationPolicy: asString(prompting.mutation_policy),
    directive: asString(prompting.directive),
    openQuestions: uniqueStrings(prompting.open_questions),
    currentFocus: asString(handoffSummary.current_focus) || asString(knowledge.current_focus) || selectedJob?.summary || '',
    nextSteps: uniqueStrings(handoffSummary.next_steps || knowledge.next_steps).slice(0, 6),
    openRisks: uniqueStrings(handoffSummary.open_risks || knowledge.open_risks).slice(0, 6),
    reviewTopOpen,
    handoffMode: packet.handoff ? 'full' : packet.handoff_summary ? 'summary' : 'none',
  };
}

export async function getControlTowerPacket(jobId, selectedJob = null) {
  const normalizedJobID = asString(jobId);
  if (!normalizedJobID) {
    return null;
  }
  const packet = await fetchLayerOs(`/api/layer-os/jobs/packet?job_id=${encodeURIComponent(normalizedJobID)}`, {
    timeoutMs: controlTowerTimeoutMs,
  });
  return summarizePacket(packet, selectedJob);
}

function buildLane(role, jobs, profiles, runtimeWorkers) {
  const openJob = jobs.find((item) => item.role === role && item.isOpen) || null;
  const latestJob = openJob || jobs.find((item) => item.role === role) || null;
  const profile = profiles.find((item) => item.role === role) || null;
  const workerState = asString(runtimeWorkers?.[role]) || 'stopped';

  return {
    role,
    label: laneLabels[role] || role,
    workerState,
    workerTone: workerState.startsWith('running') ? 'good' : workerState === 'stopped' ? 'muted' : 'alert',
    profile,
    job: latestJob,
    live: Boolean(openJob),
  };
}

function buildAttentionHint(jobs) {
  const candidate = jobs.find((item) => item.followUp?.summary && item.followUp.mode !== 'continue_loop')
    || jobs.find((item) => item.followUp?.summary)
    || null;
  if (!candidate) {
    return null;
  }
  return {
    jobId: candidate.jobId,
    role: candidate.role,
    statusLabel: candidate.statusLabel,
    summary: candidate.followUp.summary,
    mode: candidate.followUp.mode,
    nextJobIds: candidate.followUp.jobIds,
  };
}

async function settle(label, work, fallback) {
  try {
    return { ok: true, payload: await work(), error: '' };
  } catch (error) {
    return { ok: false, payload: fallback, error: `${label}: ${safeError(error, `${label} unavailable`)}` };
  }
}

export async function getControlTowerView(requestedJobId = '', options = {}) {
  const includePacket = options.includePacket !== false;
  const [runtimeResult, jobsResult, knowledgeResult, providersResult] = await Promise.all([
    settle('quickwork_runtime', () => getQuickworkRuntimeStatus(), null),
    settle('jobs', () => fetchLayerOs('/api/layer-os/jobs?limit=36', { timeoutMs: controlTowerTimeoutMs }), { items: [] }),
    settle('knowledge', () => fetchLayerOs('/api/layer-os/knowledge', { timeoutMs: controlTowerTimeoutMs }), { knowledge: {} }),
    settle('providers', () => fetchLayerOs('/api/layer-os/providers', { timeoutMs: controlTowerTimeoutMs }), { providers: [] }),
  ]);

  const jobs = asArray(jobsResult.payload?.items)
    .map((item) => normalizeJob(item))
    .sort((left, right) => toMillis(right.updatedAt || right.createdAt) - toMillis(left.updatedAt || left.createdAt));
  const runtime = runtimeResult.payload;
  const profiles = asArray(runtime?.dispatch_profiles).map((item) => normalizeProfile(item));
  const providers = asArray(providersResult.payload?.providers);
  const openJobs = jobs.filter((item) => item.isOpen);
  const recentJobs = jobs.slice(0, 18);
  const selectedJobId = asString(requestedJobId) || openJobs[0]?.jobId || recentJobs[0]?.jobId || '';
  const selectedJob = jobs.find((item) => item.jobId === selectedJobId) || null;

  const packetResult = selectedJobId && includePacket
    ? await settle(
      'job_packet',
      () => getControlTowerPacket(selectedJobId, selectedJob),
      null,
    )
    : { ok: true, payload: null, error: '' };

  const warnings = [
    ...(runtimeResult.ok ? [] : [runtimeResult.error]),
    ...(jobsResult.ok ? [] : [jobsResult.error]),
    ...(knowledgeResult.ok ? [] : [knowledgeResult.error]),
    ...(providersResult.ok ? [] : [providersResult.error]),
    ...(packetResult.ok || !selectedJobId ? [] : [packetResult.error]),
    ...uniqueStrings(runtime?.warnings),
  ];

  return {
    generatedAt: new Date().toISOString(),
    runtimeAvailable: warnings.length === 0,
    degradedReason: warnings[0] || '',
    warnings,
    metrics: {
      total: jobs.length,
      open: openJobs.length,
      running: jobs.filter((item) => item.status === 'running').length,
      queued: jobs.filter((item) => item.status === 'queued').length,
      failed: jobs.filter((item) => item.status === 'failed').length,
      succeeded: jobs.filter((item) => item.status === 'succeeded').length,
      reviewOpen: Number(knowledgeResult.payload?.knowledge?.review_open_count || 0),
    },
    runtime: runtime ? {
      daemon: asString(runtime.daemon),
      stateDir: asString(runtime.state_dir),
      writeReady: Boolean(runtime.write_ready),
      writeTokenConfigured: Boolean(runtime.write_token_configured),
      writeTokenSource: asString(runtime.write_token_source),
      writeAuthEnabled: runtime.write_auth_enabled,
      workers: asRecord(runtime.workers),
    } : null,
    laneProfiles: profiles,
    providers,
    lanes: laneOrder.map((role) => buildLane(role, recentJobs, profiles, runtime?.workers)),
    attentionHint: buildAttentionHint(recentJobs),
    queueHint: {
      primaryAction: asString(knowledgeResult.payload?.knowledge?.primary_action),
      primaryRef: asString(knowledgeResult.payload?.knowledge?.primary_ref),
      currentFocus: asString(knowledgeResult.payload?.knowledge?.current_focus),
      nextSteps: uniqueStrings(knowledgeResult.payload?.knowledge?.next_steps).slice(0, 4),
    },
    openJobs: openJobs.slice(0, 12),
    recentJobs,
    selectedJob: selectedJob ? {
      ...selectedJob,
      packet: packetResult.payload,
    } : null,
  };
}

export async function performControlTowerAction(action, payload = {}) {
  const jobId = asString(payload.job_id || payload.jobId);

  switch (action) {
    case 'assign':
      return submitQuickwork({
        summary: asString(payload.summary),
        role: asString(payload.role) || 'implementer',
        allowed_paths: asArray(payload.allowed_paths),
        payload_json: asString(payload.payload_json),
      });
    case 'heartbeat':
      return fetchLayerOs('/api/layer-os/jobs/promote', {
        method: 'POST',
        requireWriteToken: true,
        json: {
          limit: Math.max(1, Number(payload.limit || 1)),
          dispatch: payload.dispatch !== false,
        },
      });
    case 'dispatch':
      if (!jobId) throw new Error('job_id is required');
      return fetchLayerOs('/api/layer-os/jobs/dispatch', {
        method: 'POST',
        requireWriteToken: true,
        json: {
          job_id: jobId,
        },
      });
    case 'pause':
    case 'cancel':
      if (!jobId) throw new Error('job_id is required');
      return fetchLayerOs('/api/layer-os/jobs/update', {
        method: 'POST',
        requireWriteToken: true,
        json: {
          job_id: jobId,
          status: 'canceled',
          notes: uniqueStrings(['surface:web-admin', 'cancelled_from_control_tower', ...(payload.notes || [])]),
          result: {
            canceled_by: 'web_admin',
          },
        },
      });
    case 'promote':
      return fetchLayerOs('/api/layer-os/jobs/promote', {
        method: 'POST',
        requireWriteToken: true,
        json: {
          limit: Math.max(1, Number(payload.limit || 1)),
          dispatch: payload.dispatch !== false,
        },
      });
    default:
      throw new Error('unsupported control tower action');
  }
}
