function asArray(value) {
  return Array.isArray(value) ? value : [];
}

function toMillis(value) {
  if (!value) return 0;
  const numeric = new Date(value).getTime();
  return Number.isFinite(numeric) ? numeric : 0;
}

function sortNewest(items) {
  return asArray(items)
    .slice()
    .sort((left, right) => toMillis(right.updated_at || right.created_at) - toMillis(left.updated_at || left.created_at));
}

export const STALE_REVIEW_WINDOW_HOURS = 6;

const STRATEGIC_SOURCES = new Set([
  'agent_assistant',
  'document_analysis',
  'founder_idea',
  'founder_research',
  'founder_report',
]);

const RUNTIME_RESIDUE_SOURCES = new Set([
  'agent_job.failed',
  'verification.failed',
]);

function countsBy(items, key) {
  const counts = new Map();
  for (const item of asArray(items)) {
    const value = String(item?.[key] || 'unknown').trim() || 'unknown';
    counts.set(value, (counts.get(value) || 0) + 1);
  }
  return Array.from(counts.entries())
    .map(([label, count]) => ({ label, count }))
    .sort((left, right) => right.count - left.count || left.label.localeCompare(right.label))
    .slice(0, 4);
}

function isStrategicBacklogItem(item) {
  if (!item) return false;
  if (STRATEGIC_SOURCES.has(item.source)) return true;
  return item.kind === 'proposal' || item.kind === 'insight';
}

function isRuntimeResidueItem(item) {
  return Boolean(item) && RUNTIME_RESIDUE_SOURCES.has(item.source);
}

function isHardBlockerItem(item) {
  if (!item) return false;
  if (item.source === 'codex_review') return true;
  if (typeof item.source === 'string' && item.source.startsWith('audit.')) return true;
  return item.kind === 'structure_drift' || item.kind === 'daemon_freshness_drift';
}

function isRecentItem(item, cutoff) {
  return toMillis(item?.updated_at || item?.created_at) >= cutoff;
}

export function formatReviewRoomTimestamp(value) {
  const numeric = toMillis(value);
  if (!numeric) {
    return '--';
  }
  return new Date(numeric).toLocaleString('ko-KR', {
    hour12: false,
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function triageReviewRoom(room, options = {}) {
  const staleWindowHours = Number.isFinite(options.staleWindowHours)
    ? options.staleWindowHours
    : STALE_REVIEW_WINDOW_HOURS;
  const now = Number.isFinite(options.now) ? options.now : Date.now();
  const cutoff = now - staleWindowHours * 60 * 60 * 1000;

  const activeBlockers = [];
  const runtimeResidue = [];
  const strategicBacklog = [];
  const otherUnresolved = [];

  for (const item of sortNewest(room?.open)) {
    if (isStrategicBacklogItem(item)) {
      strategicBacklog.push(item);
      continue;
    }
    if (isRuntimeResidueItem(item) && !isRecentItem(item, cutoff)) {
      runtimeResidue.push(item);
      continue;
    }
    if (isHardBlockerItem(item) || isRecentItem(item, cutoff)) {
      activeBlockers.push(item);
      continue;
    }
    otherUnresolved.push(item);
  }

  return {
    staleWindowHours,
    activeBlockers,
    runtimeResidue,
    strategicBacklog,
    otherUnresolved,
    metrics: {
      open: asArray(room?.open).length,
      blockers: activeBlockers.length,
      residue: runtimeResidue.length,
      strategic: strategicBacklog.length,
      other: otherUnresolved.length,
      deferred: asArray(room?.deferred).length,
      accepted: asArray(room?.accepted).length,
    },
    sourceBreakdown: {
      blockers: countsBy(activeBlockers, 'source'),
      residue: countsBy(runtimeResidue, 'source'),
      strategic: countsBy(strategicBacklog, 'source'),
      other: countsBy(otherUnresolved, 'source'),
    },
  };
}
