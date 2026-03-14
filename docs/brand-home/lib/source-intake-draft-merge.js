function asArray(value) {
  return Array.isArray(value) ? value : [];
}

function asText(value, fallback = '') {
  return typeof value === 'string' ? value.trim() : fallback;
}

function toMillis(value) {
  if (!value) return 0;
  const numeric = new Date(value).getTime();
  return Number.isFinite(numeric) ? numeric : 0;
}

function draftAccountRank(accountId) {
  switch (asText(accountId)) {
    case '97layer':
      return 0;
    case 'woosunhokr':
      return 1;
    case 'woohwahae':
      return 2;
    default:
      return 9;
  }
}

export function mergeSourceIntakeDraftSeeds(items = [], drafts = []) {
  const grouped = new Map();
  const seen = new Set();

  for (const item of asArray(drafts)) {
    const sourceObservationId = asText(item?.sourceObservationId);
    const targetAccount = asText(item?.targetAccount);
    if (!sourceObservationId || !targetAccount) {
      continue;
    }
    const key = `${sourceObservationId}:${targetAccount}`;
    if (seen.has(key)) {
      continue;
    }
    seen.add(key);
    const current = grouped.get(sourceObservationId) || [];
    current.push(item);
    grouped.set(sourceObservationId, current);
  }

  return asArray(items).map((item) => {
    const draftSeeds = asArray(grouped.get(item.observationId)).sort((left, right) => {
      const rankGap = draftAccountRank(left.targetAccount) - draftAccountRank(right.targetAccount);
      if (rankGap !== 0) {
        return rankGap;
      }
      return toMillis(right.capturedAt) - toMillis(left.capturedAt);
    });
    return {
      ...item,
      draftSeed: draftSeeds[0] || null,
      draftSeeds,
    };
  });
}

export function mergeSourceIntakeRouteDecisions(items = [], decisions = []) {
  const grouped = new Map();

  for (const item of asArray(decisions)) {
    const sourceObservationId = asText(item?.sourceObservationId);
    if (!sourceObservationId) {
      continue;
    }
    const current = grouped.get(sourceObservationId) || [];
    current.push(item);
    grouped.set(sourceObservationId, current);
  }

  return asArray(items).map((item) => {
    const routeDecisions = asArray(grouped.get(item.observationId)).sort((left, right) => toMillis(right.capturedAt) - toMillis(left.capturedAt));
    return {
      ...item,
      routeDecision: routeDecisions[0] || null,
      routeDecisions,
    };
  });
}

export function mergeSourceIntakePrepLanes(items = [], preps = []) {
  const grouped = new Map();

  for (const item of asArray(preps)) {
    const sourceObservationId = asText(item?.sourceObservationId);
    const targetAccount = asText(item?.targetAccount);
    const channel = asText(item?.channel);
    if (!sourceObservationId || !targetAccount || !channel) {
      continue;
    }
    const key = `${sourceObservationId}:${targetAccount}:${channel}`;
    const current = grouped.get(sourceObservationId) || new Map();
    const existing = current.get(key);
    if (!existing || toMillis(item.capturedAt) > toMillis(existing.capturedAt)) {
      current.set(key, item);
    }
    grouped.set(sourceObservationId, current);
  }

  return asArray(items).map((item) => {
    const prepLanes = Array.from((grouped.get(item.observationId) || new Map()).values()).sort((left, right) => {
      const rankGap = draftAccountRank(left.targetAccount) - draftAccountRank(right.targetAccount);
      if (rankGap !== 0) {
        return rankGap;
      }
      return toMillis(right.capturedAt) - toMillis(left.capturedAt);
    });
    return {
      ...item,
      prepLane: prepLanes[0] || null,
      prepLanes,
    };
  });
}
