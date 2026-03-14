function asArray(value) {
  return Array.isArray(value) ? value : [];
}

function asText(value, fallback = '') {
  return typeof value === 'string' ? value.trim() : fallback;
}

function uniqueStrings(values) {
  return Array.from(new Set(asArray(values).map((value) => asText(value)).filter(Boolean)));
}

export function parseSourceIntakeRaw(value) {
  const lines = String(value || '').replace(/\r\n/g, '\n').split('\n');
  const parsed = {
    intakeClass: '',
    policyColor: '',
    title: '',
    url: '',
    excerpt: '',
    founderNote: '',
    priorityScore: 0,
    disposition: '',
    dispositionNote: '',
    domainTags: [],
    worldviewTags: [],
    suggestedRoutes: [],
  };
  let excerptStart = -1;

  for (const [index, rawLine] of lines.entries()) {
    const line = asText(rawLine);
    if (!line) {
      continue;
    }
    if (line === 'excerpt:') {
      excerptStart = index + 1;
      continue;
    }
    if (line.startsWith('intake_class=')) {
      parsed.intakeClass = asText(line.slice('intake_class='.length));
      continue;
    }
    if (line.startsWith('policy_color=')) {
      parsed.policyColor = asText(line.slice('policy_color='.length));
      continue;
    }
    if (line.startsWith('title=')) {
      parsed.title = asText(line.slice('title='.length));
      continue;
    }
    if (line.startsWith('url=')) {
      parsed.url = asText(line.slice('url='.length));
      continue;
    }
    if (line.startsWith('founder_note=')) {
      parsed.founderNote = asText(line.slice('founder_note='.length));
      continue;
    }
    if (line.startsWith('priority_score=')) {
      parsed.priorityScore = Number.parseInt(asText(line.slice('priority_score='.length)), 10) || 0;
      continue;
    }
    if (line.startsWith('disposition=')) {
      parsed.disposition = asText(line.slice('disposition='.length));
      continue;
    }
    if (line.startsWith('disposition_note=')) {
      parsed.dispositionNote = asText(line.slice('disposition_note='.length));
      continue;
    }
    if (line.startsWith('domain_tags=')) {
      parsed.domainTags = uniqueStrings(line.slice('domain_tags='.length).split(','));
      continue;
    }
    if (line.startsWith('worldview_tags=')) {
      parsed.worldviewTags = uniqueStrings(line.slice('worldview_tags='.length).split(','));
      continue;
    }
    if (line.startsWith('suggested_routes=')) {
      parsed.suggestedRoutes = uniqueStrings(line.slice('suggested_routes='.length).split(','));
    }
  }

  if (excerptStart >= 0 && excerptStart <= lines.length) {
    parsed.excerpt = String(lines.slice(excerptStart).join('\n')).trim();
  }

  return parsed;
}

export function parseSourceDraftSeedRaw(value) {
  const lines = String(value || '').replace(/\r\n/g, '\n').split('\n');
  const parsed = {
    targetAccount: '',
    targetTone: '',
    title: '',
    sourceObservationId: '',
    routeDecisionId: '',
    parentDraftObservationId: '',
    revisionNote: '',
    sourceTitle: '',
    sourceURL: '',
    founderNote: '',
    domainTags: [],
    worldviewTags: [],
    draft: '',
  };
  let draftStart = -1;

  for (const [index, rawLine] of lines.entries()) {
    const line = asText(rawLine);
    if (!line) {
      continue;
    }
    if (line === 'draft:') {
      draftStart = index + 1;
      continue;
    }
    if (line.startsWith('target_account=')) {
      parsed.targetAccount = asText(line.slice('target_account='.length));
      continue;
    }
    if (line.startsWith('target_tone=')) {
      parsed.targetTone = asText(line.slice('target_tone='.length));
      continue;
    }
    if (line.startsWith('title=')) {
      parsed.title = asText(line.slice('title='.length));
      continue;
    }
    if (line.startsWith('source_observation_id=')) {
      parsed.sourceObservationId = asText(line.slice('source_observation_id='.length));
      continue;
    }
    if (line.startsWith('route_decision_id=')) {
      parsed.routeDecisionId = asText(line.slice('route_decision_id='.length));
      continue;
    }
    if (line.startsWith('parent_draft_observation_id=')) {
      parsed.parentDraftObservationId = asText(line.slice('parent_draft_observation_id='.length));
      continue;
    }
    if (line.startsWith('revision_note=')) {
      parsed.revisionNote = asText(line.slice('revision_note='.length));
      continue;
    }
    if (line.startsWith('source_title=')) {
      parsed.sourceTitle = asText(line.slice('source_title='.length));
      continue;
    }
    if (line.startsWith('source_url=')) {
      parsed.sourceURL = asText(line.slice('source_url='.length));
      continue;
    }
    if (line.startsWith('founder_note=')) {
      parsed.founderNote = asText(line.slice('founder_note='.length));
      continue;
    }
    if (line.startsWith('domain_tags=')) {
      parsed.domainTags = uniqueStrings(line.slice('domain_tags='.length).split(',')).filter((tag) => tag !== 'none');
      continue;
    }
    if (line.startsWith('worldview_tags=')) {
      parsed.worldviewTags = uniqueStrings(line.slice('worldview_tags='.length).split(',')).filter((tag) => tag !== 'none');
    }
  }

  if (draftStart >= 0 && draftStart <= lines.length) {
    parsed.draft = String(lines.slice(draftStart).join('\n')).trim();
  }

  return parsed;
}

export function parseSourceIntakeRouteDecisionRaw(value) {
  const lines = String(value || '').replace(/\r\n/g, '\n').split('\n');
  const parsed = {
    sourceObservationId: '',
    decision: '',
    title: '',
    summary: '',
    routeSource: '',
  };

  for (const rawLine of lines) {
    const line = asText(rawLine);
    if (!line) {
      continue;
    }
    if (line.startsWith('source_observation_id=')) {
      parsed.sourceObservationId = asText(line.slice('source_observation_id='.length));
      continue;
    }
    if (line.startsWith('decision=')) {
      parsed.decision = asText(line.slice('decision='.length));
      continue;
    }
    if (line.startsWith('title=')) {
      parsed.title = asText(line.slice('title='.length));
      continue;
    }
    if (line.startsWith('summary=')) {
      parsed.summary = asText(line.slice('summary='.length));
      continue;
    }
    if (line.startsWith('route_source=')) {
      parsed.routeSource = asText(line.slice('route_source='.length));
    }
  }

  return parsed;
}
