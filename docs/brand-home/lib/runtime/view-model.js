import 'server-only';

import { publicHomeLocalSourcePack } from '../../content/public-home-source';
import { buildPublicHomeBrandView, normalizeBrandSourcePack } from '../brand-source-pack';
import { fetchLayerOs, safeLayerOs } from './layer-os';

function asArray(value) {
  return Array.isArray(value) ? value : [];
}

function asString(value) {
  return typeof value === 'string' && value.trim() ? value.trim() : '';
}

function toMillis(value) {
  if (!value) return 0;
  const numeric = new Date(value).getTime();
  return Number.isFinite(numeric) ? numeric : 0;
}

function latestBy(items, key) {
  return asArray(items).slice().sort((left, right) => toMillis(right?.[key]) - toMillis(left?.[key]))[0] || null;
}

function firstBy(items, predicate) {
  return asArray(items).find((item) => predicate(item)) || null;
}

function durationSeconds(startedAt, finishedAt) {
  const start = toMillis(startedAt);
  const finish = toMillis(finishedAt);
  if (!start || !finish || finish < start) {
    return null;
  }
  return Math.round((finish - start) / 1000);
}

function extractDeployFacts(notes) {
  const facts = {
    releaseStamp: '',
    remoteHost: '',
    remoteReleaseDir: '',
    service: '',
  };
  for (const note of asArray(notes)) {
    if (typeof note !== 'string' || note.length === 0) continue;
    const source = note.startsWith('output: ') ? note.slice('output: '.length) : note;
    for (const line of source.split('\n')) {
      const trimmed = line.trim();
      if (trimmed.startsWith('release_stamp=')) {
        facts.releaseStamp = trimmed.slice('release_stamp='.length);
      } else if (trimmed.startsWith('remote_host=')) {
        facts.remoteHost = trimmed.slice('remote_host='.length);
      } else if (trimmed.startsWith('remote_release_dir=')) {
        facts.remoteReleaseDir = trimmed.slice('remote_release_dir='.length);
      } else if (trimmed.startsWith('service=')) {
        facts.service = trimmed.slice('service='.length);
      }
    }
  }
  return facts;
}

function extractTargetHost(command) {
  const parts = asArray(command).map((part) => String(part || '').trim()).filter(Boolean);
  const index = parts.indexOf('--host');
  if (index >= 0 && parts[index + 1]) {
    return parts[index + 1];
  }
  return '';
}

function summarizeProgress(progress) {
  if (!progress) {
    return {
      overallPercent: 0,
      overallStatus: 'unknown',
      overallBar: '',
      axes: [],
    };
  }
  return {
    overallPercent: progress.overall_percent ?? 0,
    overallStatus: progress.overall_status ?? 'unknown',
    overallBar: progress.overall_bar ?? '',
    axes: asArray(progress.axes).slice(0, 6).map((axis) => ({
      axisId: axis.axis_id,
      label: axis.label,
      percent: axis.percent,
      bar: axis.bar,
      status: axis.status,
      signals: asArray(axis.signals).slice(0, 3),
    })),
  };
}

function summarizeReview(summary, room) {
  const openItems = asArray(summary?.top_open).length > 0 ? asArray(summary.top_open) : asArray(room?.open).slice(0, 5);
  return {
    openCount: summary?.open_count ?? asArray(room?.open).length,
    deferredCount: summary?.deferred_count ?? asArray(room?.deferred).length,
    acceptedCount: summary?.accepted_count ?? asArray(room?.accepted).length,
    topOpen: openItems.map((item) => ({
      text: item.text,
      severity: item.severity,
      source: item.source,
      updatedAt: item.updated_at,
    })),
  };
}

function summarizeJobs(items) {
  const normalized = asArray(items)
    .slice()
    .sort((left, right) => toMillis(right.updated_at || right.created_at) - toMillis(left.updated_at || left.created_at));
  const attentionSource = normalized.find((item) => asString(item?.result?.follow_up?.summary) && asString(item?.result?.follow_up?.mode) !== 'continue_loop')
    || normalized.find((item) => asString(item?.result?.follow_up?.summary))
    || null;
  return {
    total: normalized.length,
    openCount: normalized.filter((item) => item.status === 'queued' || item.status === 'running').length,
    recent: normalized.slice(0, 8).map((item) => ({
      jobId: item.job_id,
      summary: item.summary,
      role: item.role,
      kind: item.kind,
      status: item.status,
      stage: item.stage,
      updatedAt: item.updated_at,
      source: item.source,
      ref: item.ref || null,
      followUp: {
        mode: asString(item?.result?.follow_up?.mode),
        summary: asString(item?.result?.follow_up?.summary),
        jobIds: asArray(item?.result?.follow_up?.job_ids).map((value) => asString(value)).filter(Boolean),
      },
    })),
    attentionHint: attentionSource ? {
      jobId: attentionSource.job_id,
      role: attentionSource.role,
      summary: asString(attentionSource?.result?.follow_up?.summary),
      mode: asString(attentionSource?.result?.follow_up?.mode),
      nextJobIds: asArray(attentionSource?.result?.follow_up?.job_ids).map((value) => asString(value)).filter(Boolean),
    } : null,
  };
}

function summarizeJobCounts(snapshot) {
  return {
    open: Number(snapshot?.open ?? 0),
    queued: Number(snapshot?.queued ?? 0),
    running: Number(snapshot?.running ?? 0),
    terminal: Number(snapshot?.terminal ?? 0),
    summaryNote: asString(snapshot?.summary_note ?? snapshot?.summaryNote),
    summaryMeta: asString(snapshot?.summary_meta ?? snapshot?.summaryMeta),
  };
}

function summarizeFounderAttentionSnapshot(snapshot) {
  if (!snapshot || typeof snapshot !== 'object') {
    return null;
  }
  return {
    mode: asString(snapshot.mode),
    summary: asString(snapshot.summary),
    detail: asString(snapshot.detail),
    ref: asString(snapshot.ref),
    lane: asString(snapshot.lane),
    source: asString(snapshot.source),
    rule: asString(snapshot.rule),
    waitingCount: Number(snapshot.waiting_count ?? snapshot.waitingCount ?? 0),
    riskCount: Number(snapshot.risk_count ?? snapshot.riskCount ?? 0),
    reviewOpenCount: Number(snapshot.review_open_count ?? snapshot.reviewOpenCount ?? 0),
    nowCount: Number(snapshot.now_count ?? snapshot.nowCount ?? 0),
    nextJobIds: asArray(snapshot.next_job_ids ?? snapshot.nextJobIds).map((value) => asString(value)).filter(Boolean),
    lastReleaseRef: asString(snapshot.last_release_ref ?? snapshot.lastReleaseRef),
    lastRollbackRef: asString(snapshot.last_rollback_ref ?? snapshot.lastRollbackRef),
  };
}

function summarizeFounderFlowDefaultsSnapshot(snapshot) {
  if (!snapshot || typeof snapshot !== 'object') {
    return null;
  }
  const attention = snapshot.attention && typeof snapshot.attention === 'object'
    ? {
        summary: asString(snapshot.attention.summary),
        ref: asString(snapshot.attention.ref),
        waitingCount: Number(snapshot.attention.waiting_count ?? snapshot.attention.waitingCount ?? 0),
        riskCount: Number(snapshot.attention.risk_count ?? snapshot.attention.riskCount ?? 0),
        detail: asString(snapshot.attention.detail),
      }
    : null;
  return {
    flowId: asString(snapshot.flow_id ?? snapshot.flowId),
    approvalId: asString(snapshot.approval_id ?? snapshot.approvalId),
    title: asString(snapshot.title),
    intent: asString(snapshot.intent),
    attention,
  };
}

function founderActionLabel(action) {
  switch (asString(action)) {
    case 'review_room':
      return '리뷰룸 안건을 먼저 봐야 합니다.';
    case 'rollback_or_fix':
      return '위험 lane을 먼저 고쳐야 합니다.';
    case 'shape_or_promote':
      return 'proposal을 shaped work로 올릴 차례입니다.';
    case 'dispatch_job':
      return 'queued job을 먼저 dispatch하거나 triage해야 합니다.';
    case 'approve':
      return 'waiting approval을 founder가 먼저 결정해야 합니다.';
    case 'continue':
      return '이미 열린 실행 lane을 이어가면 됩니다.';
    case 'idle':
      return '긴급 blocker가 없어 다음 한 가지를 새로 열 수 있습니다.';
    default:
      return '';
  }
}

function summarizeFounderAttention(founderSummary, founderView, review, jobs) {
  const primaryAction = asString(founderSummary?.primary_action);
  const primaryRef = asString(founderSummary?.primary_ref);
  const rationale = founderSummary?.priority_rationale || null;
  const reviewText = asArray(founderSummary?.review_top_open).map((item) => asString(item)).filter(Boolean)[0]
    || review?.topOpen?.[0]?.text
    || '';
  const lastChangeSummary = asString(founderSummary?.last_change?.summary);
  const waitingCount = Number(founderSummary?.waiting_count || 0);
  const riskCount = Number(founderSummary?.risk_count || 0);
  const reviewOpenCount = Number(founderSummary?.review_open_count || 0);
  const nowCount = Number(founderSummary?.now_count || 0);
  const fallbackJobSummary = jobs?.attentionHint?.summary || jobs?.recent?.[0]?.followUp?.summary || jobs?.recent?.[0]?.summary || '';

  const summary = founderActionLabel(primaryAction)
    || reviewText
    || lastChangeSummary
    || fallbackJobSummary
    || '지금 연결된 founder 흐름이 없습니다.';

  const detailParts = [
    rationale?.reason ? `rule · ${rationale.reason}` : '',
    reviewText && reviewText !== summary ? `agenda · ${reviewText}` : '',
    lastChangeSummary && lastChangeSummary !== summary ? `change · ${lastChangeSummary}` : '',
  ].filter(Boolean);

  return {
    mode: primaryAction || 'idle',
    summary,
    ref: primaryRef || jobs?.attentionHint?.jobId || '',
    lane: asString(rationale?.lane),
    source: asString(rationale?.source),
    rule: asString(rationale?.rule),
    waitingCount,
    riskCount,
    reviewOpenCount,
    nowCount,
    detail: detailParts.join(' · ') || 'founder priority와 current runtime 상태를 합쳐 지금 먼저 볼 레인을 정리했습니다.',
    nextJobIds: asArray(jobs?.attentionHint?.nextJobIds).map((value) => asString(value)).filter(Boolean),
    lastReleaseRef: asString(founderView?.last_release?.ref),
    lastRollbackRef: asString(founderView?.last_rollback?.ref),
  };
}

function summarizeKnowledgePacket(packet) {
  if (!packet) {
    return {
      currentFocus: '',
      nextSteps: [],
      openRisks: [],
      corpusLessons: [],
      observationLinks: [],
      proposalCandidates: [],
      openThreads: [],
      actionHints: [],
    };
  }
  return {
    currentFocus: packet.current_focus || '',
    nextSteps: asArray(packet.next_steps).slice(0, 4),
    openRisks: asArray(packet.open_risks).slice(0, 4),
    corpusLessons: asArray(packet.corpus_lessons).slice(0, 4),
    observationLinks: asArray(packet.observation_links).slice(0, 4),
    proposalCandidates: asArray(packet.proposal_candidates).slice(0, 4),
    openThreads: asArray(packet.open_threads).slice(0, 4),
    actionHints: asArray(packet.action_hints).slice(0, 4),
  };
}

function summarizeKnowledgeAttention(knowledge) {
  const currentFocus = asString(knowledge?.currentFocus);
  const topRisk = asString(asArray(knowledge?.openRisks)[0]);
  const topStep = asString(asArray(knowledge?.nextSteps)[0]);
  const topThread = asArray(knowledge?.openThreads)[0] || null;

  if (topRisk) {
    return {
      mode: 'review_risk',
      summary: `지금 먼저 볼 지식 리스크: ${topRisk}`,
      detail: currentFocus ? `focus · ${currentFocus}` : '열린 risk를 먼저 정리한 뒤 다음 step으로 넘어가는 편이 안전합니다.',
      ref: asString(topThread?.thread_id),
    };
  }
  if (topStep) {
    return {
      mode: 'continue_step',
      summary: `지금 먼저 이어갈 step: ${topStep}`,
      detail: currentFocus ? `focus · ${currentFocus}` : 'current focus를 기준으로 가장 가까운 다음 step을 먼저 보면 됩니다.',
      ref: asString(topThread?.thread_id),
    };
  }
  if (currentFocus) {
    return {
      mode: 'review_focus',
      summary: `지금 붙들 current focus: ${currentFocus}`,
      detail: '열린 risk나 next step이 없어도 current focus를 잃지 않는 것이 중요합니다.',
      ref: asString(topThread?.thread_id),
    };
  }
  return {
    mode: 'idle',
    summary: '지금 바로 볼 knowledge hotspot은 없습니다.',
    detail: '새 source, corpus search, review-room 신호가 생기면 여기 먼저 뜹니다.',
    ref: '',
  };
}

function summarizeReviewAttention(summary, source, degradedReason = '') {
  const topOpen = asArray(summary?.topOpen);
  const topItem = topOpen[0] || null;
  if (topItem?.text) {
    return {
      mode: 'review_blocker',
      summary: `지금 먼저 볼 blocker: ${topItem.text}`,
      detail: topItem.severity ? `severity · ${topItem.severity}` : 'top open agenda를 먼저 닫는 것이 현재 우선입니다.',
      ref: asString(topItem.source),
    };
  }
  if (degradedReason) {
    return {
      mode: 'runtime_degraded',
      summary: 'review-room runtime 상태를 먼저 확인해야 합니다.',
      detail: degradedReason,
      ref: asString(source),
    };
  }
  if ((summary?.openCount || 0) > 0) {
    return {
      mode: 'review_open',
      summary: `열린 review 항목 ${summary.openCount}건이 남아 있습니다.`,
      detail: '상세 residue와 backlog는 뒤에서 이어서 봐도 됩니다.',
      ref: asString(source),
    };
  }
  return {
    mode: 'clear',
    summary: '지금 급한 review blocker는 없습니다.',
    detail: 'runtime residue와 strategic backlog만 가볍게 훑으면 됩니다.',
    ref: asString(source),
  };
}

function summarizeSourceIntakeDashboard(snapshot) {
  const attention = snapshot?.attention && typeof snapshot.attention === 'object'
    ? {
        mode: asString(snapshot.attention.mode),
        summary: asString(snapshot.attention.summary),
        detail: asString(snapshot.attention.detail),
        ref: asString(snapshot.attention.ref),
        sourceObservationId: asString(snapshot.attention.source_observation_id ?? snapshot.attention.sourceObservationId),
        draftObservationId: asString(snapshot.attention.draft_observation_id ?? snapshot.attention.draftObservationId),
      }
    : null;
  return {
    recentCount: Number(snapshot?.recent_count ?? snapshot?.recentCount ?? 0),
    actionCount: Number(snapshot?.action_count ?? snapshot?.actionCount ?? 0),
    quietCount: Number(snapshot?.quiet_count ?? snapshot?.quietCount ?? 0),
    feedCount: Number(snapshot?.feed_count ?? snapshot?.feedCount ?? 0),
    draftSeedCount: Number(snapshot?.draft_seed_count ?? snapshot?.draftSeedCount ?? 0),
    prepCount: Number(snapshot?.prep_count ?? snapshot?.prepCount ?? 0),
    reviewCount: Number(snapshot?.review_count ?? snapshot?.reviewCount ?? 0),
    prepReadyCount: Number(snapshot?.prep_ready_count ?? snapshot?.prepReadyCount ?? 0),
    autoRoutedCount: Number(snapshot?.auto_routed_count ?? snapshot?.autoRoutedCount ?? 0),
    summaryNote: asString(snapshot?.summary_note ?? snapshot?.summaryNote),
    summaryMeta: asString(snapshot?.summary_meta ?? snapshot?.summaryMeta),
    quietNote: asString(snapshot?.quiet_note ?? snapshot?.quietNote),
    attention,
  };
}

function summarizeOverviewPrimaryAttention(founderAttention, jobs, sourceIntake) {
  if (founderAttention?.mode && founderAttention.mode !== 'idle') {
    return founderAttention;
  }
  if (jobs?.attentionHint?.summary) {
    return {
      mode: jobs.attentionHint.mode || 'job_follow_up',
      summary: jobs.attentionHint.summary,
      detail: jobs?.recent?.[0]?.summary || '최근 작업 후속 액션을 먼저 확인하세요.',
      ref: jobs.attentionHint.jobId || '',
    };
  }
  if (sourceIntake?.attention?.summary) {
    return sourceIntake.attention;
  }
  return founderAttention || null;
}

function summarizeFounderFlowDefaults(founderSummary, founderAttention, primaryAttention, founderRef, jobs, review) {
  return {
    flowId: founderRef,
    approvalId: '',
    title: asString(founderSummary?.primary_action) || '운영 흐름',
    intent: '보호된 관리자 화면에서 founder 흐름을 닫습니다.',
    attention: {
      summary: primaryAttention?.summary || asString(founderSummary?.primary_action) || '지금 연결된 founder 흐름이 없습니다.',
      ref: primaryAttention?.ref || founderRef || jobs?.attentionHint?.jobId || '',
      waitingCount: founderAttention?.waitingCount ?? Number(founderSummary?.waiting_count || 0),
      riskCount: founderAttention?.riskCount ?? Number(founderSummary?.risk_count || 0),
      detail: primaryAttention?.detail || review?.topOpen?.[0]?.text || '대표 액션과 대표 참조를 먼저 확인한 뒤 아래 founder flow 카드를 실행하면 됩니다.',
    },
  };
}

function readCockpitDashboard(cockpit) {
  return cockpit?.dashboard && typeof cockpit.dashboard === 'object' ? cockpit.dashboard : {};
}

function summarizeDeploymentOverview(cockpit, daemonPayload) {
  const companyState = cockpit?.company_state || {};
  const daemon = daemonPayload?.daemon || {};
  const security = cockpit?.security || {};
  const auth = cockpit?.auth || {};
  const targets = asArray(cockpit?.targets)
    .slice()
    .sort((left, right) => String(left?.target_id || '').localeCompare(String(right?.target_id || '')));
  const releases = asArray(cockpit?.releases)
    .slice()
    .sort((left, right) => toMillis(right.released_at) - toMillis(left.released_at));
  const deploys = asArray(cockpit?.deploys)
    .slice()
    .sort((left, right) => toMillis(right.finished_at || right.started_at) - toMillis(left.finished_at || left.started_at));
  const rollbacks = asArray(cockpit?.rollbacks)
    .slice()
    .sort((left, right) => toMillis(right.finished_at || right.started_at) - toMillis(left.finished_at || left.started_at));
  const latestRelease = releases[0] || null;
  const latestDeploy = deploys[0] || null;
  const latestRollback = rollbacks[0] || null;
  const vmTarget = firstBy(targets, (item) => item.target_id === 'vm');
  const latestVmRelease = firstBy(releases, (item) => item.target === 'vm');
  const latestVmDeploy = firstBy(deploys, (item) => item.target === 'vm');
  const latestVmDeployFacts = extractDeployFacts(latestVmDeploy?.notes);
  const continuityHostStatus = !vmTarget
    ? 'missing'
    : daemon.status !== 'ok' || security.status === 'degraded'
      ? 'degraded'
      : latestVmDeploy?.status === 'succeeded'
        ? 'operational'
        : latestVmDeploy?.status || 'pending';

  return {
    daemon: {
      status: daemon.status || 'unknown',
      address: daemon.address || '',
      deployHealth: daemon.deploy_health || companyState.deploy_health || 'unknown',
      degradedReasons: asArray(daemon.degraded_reasons).slice(0, 4),
      architectAutoDispatch: Boolean(daemon.architect?.auto_dispatch),
      architectLastRunAt: daemon.architect?.last_run_at || '',
    },
    vmTargetReady: targets.some((item) => item.target_id === 'vm'),
    counts: {
      targets: targets.length,
      releases: releases.length,
      deploys: deploys.length,
      rollbacks: rollbacks.length,
    },
    targets: targets.slice(0, 4).map((item) => ({
      targetId: item.target_id,
      commandPreview: asArray(item.command).join(' '),
      host: extractTargetHost(item.command),
      workdir: item.workdir || '',
    })),
    continuityHost: {
      status: continuityHostStatus,
      targetId: vmTarget?.target_id || 'vm',
      host: extractTargetHost(vmTarget?.command) || latestVmDeployFacts.remoteHost || '',
      commandPreview: vmTarget ? asArray(vmTarget.command).join(' ') : '',
      latestReleaseId: latestVmRelease?.release_id || latestVmDeploy?.release_id || '',
      latestDeployId: latestVmDeploy?.deploy_id || '',
      latestDeployStatus: latestVmDeploy?.status || 'not_recorded',
      latestDeployAt: latestVmDeploy?.finished_at || latestVmDeploy?.started_at || '',
      latestDeployDurationSeconds: durationSeconds(latestVmDeploy?.started_at, latestVmDeploy?.finished_at),
      releaseStamp: latestVmDeployFacts.releaseStamp || '',
      remoteReleaseDir: latestVmDeployFacts.remoteReleaseDir || '',
      remoteService: latestVmDeployFacts.service || 'layer-osd',
      securityStatus: security.status || 'unknown',
      writeAuthEnabled: Boolean(auth.write_auth_enabled),
      lastSecurityReviewAt: security.last_review_at || '',
    },
    latestRelease: latestRelease
      ? {
          releaseId: latestRelease.release_id,
          workItemId: latestRelease.work_item_id,
          target: latestRelease.target,
          channel: latestRelease.channel,
          releasedAt: latestRelease.released_at,
          artifactCount: asArray(latestRelease.artifacts).length,
          approvalCount: asArray(latestRelease.approval_refs).length,
        }
      : null,
    latestDeploy: latestDeploy
      ? {
          deployId: latestDeploy.deploy_id,
          releaseId: latestDeploy.release_id,
          target: latestDeploy.target,
          status: latestDeploy.status,
          startedAt: latestDeploy.started_at,
          finishedAt: latestDeploy.finished_at || null,
        }
      : null,
    latestRollback: latestRollback
      ? {
          rollbackId: latestRollback.rollback_id,
          releaseId: latestRollback.release_id,
          deployId: latestRollback.deploy_id || null,
          target: latestRollback.target,
          status: latestRollback.status,
          startedAt: latestRollback.started_at,
          finishedAt: latestRollback.finished_at || null,
        }
      : null,
  };
}

export async function getPublicProofView() {
  const timeoutMs = 3000;
  const [statusResult, founderSummaryResult, reviewRoomResult, verificationsResult, authResult, entriesResult, observationsResult] = await Promise.all([
    safeLayerOs('/api/layer-os/status', { timeoutMs }),
    safeLayerOs('/api/layer-os/founder-summary', { timeoutMs }),
    safeLayerOs('/api/layer-os/review-room', { timeoutMs }),
    safeLayerOs('/api/layer-os/verifications', { timeoutMs }),
    safeLayerOs('/api/layer-os/auth', { timeoutMs }),
    safeLayerOs('/api/layer-os/corpus/entries', { timeoutMs }),
    safeLayerOs('/api/layer-os/observations?limit=24', { timeoutMs }),
  ]);
  const failures = [statusResult, founderSummaryResult, reviewRoomResult, verificationsResult, authResult, entriesResult, observationsResult]
    .filter((result) => !result.ok);
  const status = statusResult.ok ? statusResult.payload : {};
  const founderSummary = founderSummaryResult.ok ? founderSummaryResult.payload?.founder_summary : {};
  const reviewRoom = reviewRoomResult.ok ? reviewRoomResult.payload?.review_room : {};
  const reviewSummary = reviewRoomResult.ok ? reviewRoomResult.payload?.summary : {};
  const verifications = verificationsResult.ok ? verificationsResult.payload?.items : [];
  const auth = authResult.ok ? authResult.payload?.auth : {};
  const entries = entriesResult.ok ? asArray(entriesResult.payload?.entries) : [];
  const observations = observationsResult.ok ? asArray(observationsResult.payload?.items) : [];
  const progress = summarizeProgress(status.company_state?.progress);
  const latestVerification = latestBy(verifications, 'started_at');
  const review = summarizeReview(
    reviewSummary?.open_count !== undefined || asArray(reviewSummary?.top_open).length > 0
      ? reviewSummary
      : {
          open_count: founderSummary?.review_open_count,
          top_open: founderSummary?.review_top_open,
        },
    reviewRoom,
  );
  const runtimeAvailable = failures.length === 0;
  const hasProgress = progress.overallStatus !== 'unknown' || progress.axes.length > 0 || progress.overallBar.length > 0;
  const hasKnowledge = entriesResult.ok || observationsResult.ok;
  const sourceLabel = runtimeAvailable
    ? 'live runtime'
    : failures.length < 7
      ? 'degraded runtime'
      : 'runtime unavailable';

  return {
    runtimeAvailable,
    degradedReason: failures[0]?.error || '',
    sourceLabel,
    generatedAt: new Date().toISOString(),
    status: status.status || status.company_state?.status || 'unknown',
    metrics: [
      {
        label: 'Progress',
        value: hasProgress ? `${progress.overallPercent}%` : 'n/a',
        note: hasProgress ? progress.overallStatus : sourceLabel,
      },
      {
        label: 'Review',
        value: reviewRoomResult.ok || founderSummaryResult.ok ? String(review.openCount) : 'n/a',
        note: reviewRoomResult.ok || founderSummaryResult.ok ? 'open agenda items' : sourceLabel,
      },
      {
        label: 'Knowledge',
        value: hasKnowledge ? `${entries.length} / ${observations.length}` : 'n/a',
        note: hasKnowledge ? 'corpus entries / observations' : sourceLabel,
      },
      {
        label: 'Security',
        value: status.security?.status || 'unknown',
        note: typeof auth?.write_auth_enabled === 'boolean'
          ? auth.write_auth_enabled ? 'write auth enabled' : 'write auth disabled'
          : sourceLabel,
      },
    ],
    axes: progress.axes.slice(0, 4),
    latestVerification: latestVerification
      ? {
          scope: latestVerification.scope,
          status: latestVerification.status,
          startedAt: latestVerification.started_at,
          notes: asArray(latestVerification.notes).slice(0, 3),
        }
      : null,
    review,
  };
}

export async function getPublicHomeView() {
  const proof = await getPublicProofView();
  const brandPack = normalizeBrandSourcePack(publicHomeLocalSourcePack);

  return {
    generatedAt: new Date().toISOString(),
    brandPack,
    brand: buildPublicHomeBrandView(brandPack),
    proof,
  };
}

export async function getAdminOverviewView() {
  const timeoutMs = 2000;
  const cockpitResult = await safeLayerOs('/api/layer-os/cockpit', { timeoutMs });
  const cockpit = cockpitResult.ok ? cockpitResult.payload : {};
  const dashboard = readCockpitDashboard(cockpit);
  const jobs = summarizeJobs(asArray(dashboard.open_jobs));
  const jobCounts = summarizeJobCounts(dashboard.job_counts || {});
  const review = summarizeReview(dashboard.review_room || cockpit.review_summary, null);
  const founderSummary = dashboard.founder_summary || cockpit.founder_summary;
  const founderView = dashboard.founder_view || cockpit.founder_view;
  const sourceIntake = summarizeSourceIntakeDashboard(dashboard.source_intake || cockpit.source_intake);
  const founderAttentionSnapshot = summarizeFounderAttentionSnapshot(dashboard.founder_attention);
  const founderAttentionFallback = summarizeFounderAttention(founderSummary, founderView, review, jobs);
  const founderAttention = cockpitResult.ok ? (founderAttentionSnapshot || null) : (founderAttentionSnapshot || founderAttentionFallback);
  const primaryAttentionSnapshot = summarizeFounderAttentionSnapshot(dashboard.primary_attention);
  const primaryAttentionFallback = summarizeOverviewPrimaryAttention(founderAttention || founderAttentionFallback, jobs, sourceIntake);
  const primaryAttention = cockpitResult.ok ? (primaryAttentionSnapshot || founderAttention || null) : (primaryAttentionSnapshot || primaryAttentionFallback);
  const founderRef = typeof founderSummary?.primary_ref === 'string' && founderSummary.primary_ref.startsWith('flow_')
    ? founderSummary.primary_ref
    : '';
  const founderFlowSnapshot = summarizeFounderFlowDefaultsSnapshot(dashboard.founder_flow_defaults);
  const founderFlowFallback = summarizeFounderFlowDefaults(
    founderSummary,
    founderAttention || founderAttentionFallback,
    primaryAttention || primaryAttentionFallback,
    founderRef,
    jobs,
    review,
  );
  const founderFlowDefaults = cockpitResult.ok ? (founderFlowSnapshot || {}) : (founderFlowSnapshot || founderFlowFallback);
  return {
    generatedAt: new Date().toISOString(),
    runtimeAvailable: cockpitResult.ok,
    degradedReason: cockpitResult.ok ? '' : cockpitResult.error,
    companyState: dashboard.company_state || cockpit.company_state,
    progress: summarizeProgress(dashboard.company_state?.progress || cockpit.company_state?.progress),
    founderSummary,
    founderView,
    founderAttention,
    primaryAttention,
    founderFlowDefaults,
    sourceIntake,
    security: dashboard.security || cockpit.security,
    auth: dashboard.auth || cockpit.auth,
    adapters: dashboard.adapters || cockpit.adapters,
    review,
    jobs,
    jobCounts,
    reviewNote: asString(dashboard.review_note),
    reviewMeta: asString(dashboard.review_meta),
    deployment: summarizeDeploymentOverview(cockpit, cockpit),
    latestVerification: latestBy(cockpit.verifications, 'started_at'),
  };
}

export async function getReviewRoomView() {
  const payload = await fetchLayerOs('/api/layer-os/review-room');
  return {
    generatedAt: new Date().toISOString(),
    room: payload.review_room,
    summary: summarizeReview(payload.summary, payload.review_room),
  };
}

export async function getReviewRoomSummaryView() {
  const timeoutMs = 2000;
  const cockpitResult = await safeLayerOs('/api/layer-os/cockpit', { timeoutMs });
  const cockpit = cockpitResult.ok ? cockpitResult.payload : {};
  const dashboard = readCockpitDashboard(cockpit);
  const reviewSummary = dashboard.review_room || cockpit.review_summary || {};
  const summary = summarizeReview(reviewSummary, null);
  return {
    generatedAt: new Date().toISOString(),
    runtimeAvailable: cockpitResult.ok,
    degradedReason: cockpitResult.ok ? '' : cockpitResult.error,
    summary,
    attention: summarizeReviewAttention(summary, reviewSummary?.source || '', cockpitResult.ok ? '' : cockpitResult.error),
    source: reviewSummary?.source || '',
    issues: asArray(reviewSummary?.issues).slice(0, 6),
  };
}

export async function getReviewRoomDetailView() {
  const payload = await fetchLayerOs('/api/layer-os/review-room');
  return {
    generatedAt: new Date().toISOString(),
    room: payload.review_room,
    summary: summarizeReview(payload.summary, payload.review_room),
  };
}

export async function getJobsView() {
  const [openJobs, allJobs] = await Promise.all([
    fetchLayerOs('/api/layer-os/jobs?status=open&limit=12'),
    fetchLayerOs('/api/layer-os/jobs?limit=24'),
  ]);
  return {
    generatedAt: new Date().toISOString(),
    open: summarizeJobs(openJobs.items),
    all: summarizeJobs(allJobs.items),
  };
}

export async function getKnowledgeSummaryView() {
  const knowledgePayload = await fetchLayerOs('/api/layer-os/knowledge');
  const knowledge = summarizeKnowledgePacket(knowledgePayload.knowledge);
  return {
    generatedAt: new Date().toISOString(),
    knowledge,
    attention: summarizeKnowledgeAttention(knowledge),
  };
}

export async function getKnowledgeDetailView(query = '') {
  const [entriesPayload, searchPayload] = await Promise.all([
    fetchLayerOs('/api/layer-os/corpus/entries'),
    query.trim() ? fetchLayerOs(`/api/layer-os/knowledge/search?q=${encodeURIComponent(query.trim())}`) : Promise.resolve(null),
  ]);
  return {
    generatedAt: new Date().toISOString(),
    query: query.trim(),
    entries: asArray(entriesPayload.entries)
      .slice()
      .sort((left, right) => toMillis(right.created_at) - toMillis(left.created_at))
      .slice(0, 12),
    search: searchPayload,
  };
}

export async function getKnowledgeView(query = '') {
  const [summary, detail] = await Promise.all([
    getKnowledgeSummaryView(),
    getKnowledgeDetailView(query),
  ]);
  return {
    generatedAt: new Date().toISOString(),
    query: detail.query,
    knowledge: summary.knowledge,
    entries: detail.entries,
    search: detail.search,
  };
}

function nodeRadiusByKind(kind) {
  switch (kind) {
    case 'root':
      return 36;
    case 'focus':
      return 22;
    case 'risk':
      return 11;
    case 'step':
      return 10;
    case 'thread':
      return 11;
    case 'lesson':
      return 10;
    case 'review':
      return 16;
    case 'job':
      return 14;
    case 'proposal':
      return 15;
    case 'observation':
      return 12;
    case 'verification':
      return 13;
    default:
      return 12;
  }
}

function nodeToneByKind(kind, status) {
  if (status === 'failed') return 'alert';
  if (status === 'succeeded' || status === 'passed' || status === 'promoted') return 'good';
  switch (kind) {
    case 'root':
      return 'core';
    case 'focus':
      return 'focus';
    case 'risk':
      return 'alert';
    case 'step':
      return 'good';
    case 'thread':
      return 'review';
    case 'lesson':
      return 'observation';
    case 'review':
      return 'review';
    case 'job':
      return 'job';
    case 'proposal':
      return 'proposal';
    case 'observation':
      return 'observation';
    case 'verification':
      return 'verification';
    default:
      return 'muted';
  }
}

function makeGraphNode({ id, label, kind, status = 'active', meta = '', weight = 1, ref = null }) {
  return {
    id,
    label,
    kind,
    status,
    meta,
    weight,
    radius: nodeRadiusByKind(kind),
    tone: nodeToneByKind(kind, status),
    ref,
  };
}

function makeGraphEdge(source, target, kind = 'signal', strength = 1) {
  return {
    id: `${source}->${target}:${kind}`,
    source,
    target,
    kind,
    strength,
  };
}

function buildFallbackNeuralRuntime() {
  return {
    knowledge: {
      current_focus: 'Restore the living neural field so runtime structure feels sentient at a glance',
      next_steps: [
        'Tune signal density so live topology reads before text',
        'Expose layer filters for focus, review, jobs, and verification',
        'Keep the field usable even when the daemon is offline',
      ],
      open_risks: [
        'Runtime bridge may be unavailable during local design work',
        'Too many equal-weight nodes can flatten the neural feeling',
      ],
      open_threads: [
        {
          thread_id: 'field-depth',
          question: 'How should the field reveal hierarchy without becoming a dashboard?',
          source: 'design thread',
        },
        {
          thread_id: 'runtime-ghost',
          question: 'What remains visible when live runtime data disappears?',
          source: 'resilience thread',
        },
      ],
      corpus_lessons: [
        'When motion reflects actual runtime links, the field feels trustworthy instead of decorative.',
        'A single strong focus spine helps the rest of the graph feel alive rather than noisy.',
      ],
    },
    review_room: {
      open: [
        {
          ref: 'rr-neural-density',
          text: 'Increase neural contrast between active focus and surrounding lanes',
          severity: 'high',
          source: 'design review',
        },
        {
          ref: 'rr-offline-fallback',
          text: 'Provide graceful offline topology when Layer OS daemon is unreachable',
          severity: 'medium',
          source: 'runtime review',
        },
      ],
      accepted: [],
      deferred: [],
    },
    jobs: {
      items: [
        {
          job_id: 'job-neural-implement',
          summary: 'Implement layered neural field controls',
          role: 'implementer',
          kind: 'designer',
          stage: 'experience',
          status: 'running',
          updated_at: new Date().toISOString(),
          ref: 'proposal-neural-field',
        },
        {
          job_id: 'job-neural-verify',
          summary: 'Verify topology remains legible on mobile widths',
          role: 'verifier',
          kind: 'designer',
          stage: 'verify',
          status: 'queued',
          updated_at: new Date().toISOString(),
        },
      ],
    },
    proposals: {
      items: [
        {
          proposal_id: 'proposal-neural-field',
          title: 'Neural field continuity pass',
          status: 'active',
          priority: 'founder',
          updated_at: new Date().toISOString(),
        },
      ],
    },
    observations: {
      items: [
        {
          observation_id: 'obs-neural-vibe',
          normalized_summary: 'The neural map should feel like a living lattice, not a static admin graph',
          source_channel: 'thread',
          refs: ['proposal-neural-field'],
          created_at: new Date().toISOString(),
        },
      ],
    },
    cockpit: {
      company_state: { status: 'degraded' },
      verifications: [
        {
          record_id: 'vrf-neural-rehearsal',
          scope: 'Fallback rehearsal graph',
          status: 'passed',
          started_at: new Date().toISOString(),
        },
      ],
    },
  };
}

function buildNeuralGraphView({
  knowledge = {},
  reviewRoom = {},
  jobs = [],
  proposals = [],
  observations = [],
  verifications = [],
  companyStatus = 'active',
  runtimeAvailable = true,
  degradedReason = '',
  sourceLabel = 'live runtime',
}) {
  const normalizedJobs = asArray(jobs);
  const normalizedProposals = asArray(proposals);
  const normalizedObservations = asArray(observations);
  const normalizedVerifications = asArray(verifications);

  const nodes = [];
  const edges = [];
  const seen = new Set();

  function pushNode(node) {
    if (seen.has(node.id)) return;
    seen.add(node.id);
    nodes.push(node);
  }

  pushNode(makeGraphNode({
    id: 'root',
    label: 'Layer OS',
    kind: 'root',
    status: companyStatus || 'active',
    meta: `${asArray(reviewRoom.open).length} open agenda`,
    weight: 3,
  }));

  const currentFocus = typeof knowledge.current_focus === 'string' && knowledge.current_focus.trim()
    ? knowledge.current_focus.trim()
    : 'Current focus unavailable';

  pushNode(makeGraphNode({
    id: 'focus:current',
    label: currentFocus,
    kind: 'focus',
    meta: `next ${asArray(knowledge.next_steps).length} / risks ${asArray(knowledge.open_risks).length}`,
    weight: 2,
  }));
  edges.push(makeGraphEdge('root', 'focus:current', 'spine', 1.6));

  asArray(knowledge.open_risks)
    .slice(0, 4)
    .forEach((item, index) => {
      const id = `risk:${index}`;
      pushNode(makeGraphNode({
        id,
        label: item,
        kind: 'risk',
        status: 'failed',
        meta: 'open risk',
      }));
      edges.push(makeGraphEdge('focus:current', id, 'risk', 1));
    });

  asArray(knowledge.next_steps)
    .slice(0, 4)
    .forEach((item, index) => {
      const id = `step:${index}`;
      pushNode(makeGraphNode({
        id,
        label: item,
        kind: 'step',
        status: 'queued',
        meta: 'next step',
      }));
      edges.push(makeGraphEdge('focus:current', id, 'step', 0.92));
    });

  asArray(knowledge.open_threads)
    .slice(0, 4)
    .forEach((item, index) => {
      const id = `thread:${item.thread_id || index}`;
      pushNode(makeGraphNode({
        id,
        label: item.question || item.text || `Open thread ${index + 1}`,
        kind: 'thread',
        status: 'active',
        meta: item.source || 'open thread',
        ref: item.thread_id || null,
      }));
      edges.push(makeGraphEdge('focus:current', id, 'thread', 0.88));
    });

  asArray(knowledge.corpus_lessons)
    .slice(0, 4)
    .forEach((item, index) => {
      const text = typeof item === 'string' ? item : item?.summary || item?.text || `Lesson ${index + 1}`;
      const id = `lesson:${index}`;
      pushNode(makeGraphNode({
        id,
        label: text,
        kind: 'lesson',
        status: 'active',
        meta: 'corpus lesson',
      }));
      edges.push(makeGraphEdge('root', id, 'lesson', 0.72));
      edges.push(makeGraphEdge(id, 'focus:current', 'lesson-focus', 0.7));
    });

  asArray(reviewRoom.open)
    .slice(0, 6)
    .forEach((item, index) => {
      const id = item.ref ? `review:${item.ref}` : `review:open:${index}`;
      pushNode(makeGraphNode({
        id,
        label: item.text || 'Open agenda',
        kind: 'review',
        status: item.severity || 'active',
        meta: item.source || 'review-room',
        weight: 1.2,
        ref: item.ref || null,
      }));
      edges.push(makeGraphEdge('focus:current', id, 'review', 1.15));
    });

  normalizedProposals
    .slice()
    .sort((left, right) => toMillis(right.updated_at || right.created_at) - toMillis(left.updated_at || left.created_at))
    .slice(0, 5)
    .forEach((item) => {
      const id = `proposal:${item.proposal_id}`;
      pushNode(makeGraphNode({
        id,
        label: item.title || item.summary || item.proposal_id,
        kind: 'proposal',
        status: item.status || 'active',
        meta: item.priority || item.surface || 'proposal',
        ref: item.proposal_id,
      }));
      edges.push(makeGraphEdge('root', id, 'proposal', 1.1));
    });

  normalizedJobs
    .slice()
    .sort((left, right) => toMillis(right.updated_at || right.created_at) - toMillis(left.updated_at || left.created_at))
    .slice(0, 8)
    .forEach((item) => {
      const id = `job:${item.job_id}`;
      pushNode(makeGraphNode({
        id,
        label: item.summary || item.job_id,
        kind: 'job',
        status: item.status || 'active',
        meta: `${item.role || 'agent'} / ${item.stage || item.kind || 'lane'}`,
        ref: item.ref || item.job_id,
      }));
      edges.push(makeGraphEdge('focus:current', id, 'job', 1));
      if (item.ref && seen.has(`proposal:${item.ref}`)) {
        edges.push(makeGraphEdge(`proposal:${item.ref}`, id, 'handoff', 0.95));
      }
    });

  normalizedObservations
    .slice()
    .sort((left, right) => toMillis(right.captured_at || right.created_at) - toMillis(left.captured_at || left.created_at))
    .slice(0, 6)
    .forEach((item) => {
      const id = `observation:${item.observation_id}`;
      pushNode(makeGraphNode({
        id,
        label: item.normalized_summary || item.topic || item.observation_id,
        kind: 'observation',
        status: 'active',
        meta: item.source_channel || 'observation',
        ref: item.observation_id,
      }));
      edges.push(makeGraphEdge('root', id, 'observation', 0.85));
      asArray(item.refs)
        .filter((ref) => typeof ref === 'string' && ref.trim())
        .forEach((ref) => {
          const proposalId = `proposal:${ref}`;
          const jobId = `job:${ref}`;
          if (seen.has(proposalId)) edges.push(makeGraphEdge(id, proposalId, 'reference', 0.8));
          if (seen.has(jobId)) edges.push(makeGraphEdge(id, jobId, 'reference', 0.8));
        });
    });

  normalizedVerifications
    .slice()
    .sort((left, right) => toMillis(right.started_at) - toMillis(left.started_at))
    .slice(0, 4)
    .forEach((item) => {
      const id = `verification:${item.record_id}`;
      pushNode(makeGraphNode({
        id,
        label: item.scope || item.record_id,
        kind: 'verification',
        status: item.status || 'active',
        meta: item.started_at ? new Date(item.started_at).toLocaleTimeString('ko-KR', { hour: '2-digit', minute: '2-digit' }) : 'verification',
        ref: item.record_id,
      }));
      edges.push(makeGraphEdge('root', id, 'verification', 0.9));
    });

  return {
    generatedAt: new Date().toISOString(),
    runtimeAvailable,
    degradedReason,
    sourceLabel,
    metrics: {
      reviewOpen: asArray(reviewRoom.open).length,
      accepted: asArray(reviewRoom.accepted).length,
      activeJobs: normalizedJobs.filter((item) => item.status === 'queued' || item.status === 'running').length,
      failedJobs: normalizedJobs.filter((item) => item.status === 'failed').length,
      observations: normalizedObservations.length,
      verifications: normalizedVerifications.length,
      risks: asArray(knowledge.open_risks).length,
      steps: asArray(knowledge.next_steps).length,
    },
    nodes,
    edges,
  };
}

export async function getNeuralGraphView() {
  const [knowledgeResult, reviewResult, jobsResult, proposalsResult, observationsResult, cockpitResult] = await Promise.all([
    safeLayerOs('/api/layer-os/knowledge'),
    safeLayerOs('/api/layer-os/review-room'),
    safeLayerOs('/api/layer-os/jobs?limit=18'),
    safeLayerOs('/api/layer-os/proposals'),
    safeLayerOs('/api/layer-os/observations?limit=16'),
    safeLayerOs('/api/layer-os/cockpit'),
  ]);

  const fallback = buildFallbackNeuralRuntime();
  const results = [
    knowledgeResult,
    reviewResult,
    jobsResult,
    proposalsResult,
    observationsResult,
    cockpitResult,
  ];
  const failures = results.filter((result) => !result.ok).map((result) => result.error);
  const runtimeAvailable = failures.length === 0;
  const degradedReason = failures[0] || '';

  return buildNeuralGraphView({
    knowledge: (knowledgeResult.ok ? knowledgeResult.payload?.knowledge : fallback.knowledge) || {},
    reviewRoom: (reviewResult.ok ? reviewResult.payload?.review_room : fallback.review_room) || {},
    jobs: (jobsResult.ok ? jobsResult.payload?.items : fallback.jobs.items) || [],
    proposals: (proposalsResult.ok ? proposalsResult.payload?.items : fallback.proposals.items) || [],
    observations: (observationsResult.ok ? observationsResult.payload?.items : fallback.observations.items) || [],
    verifications: (cockpitResult.ok ? cockpitResult.payload?.verifications : fallback.cockpit.verifications) || [],
    companyStatus: runtimeAvailable
      ? cockpitResult.payload?.company_state?.status || 'active'
      : fallback.cockpit.company_state?.status || 'degraded',
    runtimeAvailable,
    degradedReason,
    sourceLabel: runtimeAvailable ? 'live runtime' : 'synthetic rehearsal',
  });
}

export async function performReviewRoomAction(action, input) {
  return fetchLayerOs('/api/layer-os/review-room', {
    method: 'POST',
    requireWriteToken: true,
    json: {
      action,
      item: input.item,
      reason: input.reason,
      rule: input.rule,
      evidence: input.evidence,
    },
  });
}

export async function performFounderAction(action, input) {
  return fetchLayerOs(`/api/layer-os/founder-actions/${action}`, {
    method: 'POST',
    requireWriteToken: true,
    json: input,
  });
}
