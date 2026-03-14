function asArray(value) {
  return Array.isArray(value) ? value : [];
}

function asText(value, fallback = '') {
  return typeof value === 'string' ? value.trim() : fallback;
}

function uniqueStrings(values) {
  return Array.from(
    new Set(
      asArray(values)
        .map((value) => asText(value))
        .filter(Boolean),
    ),
  );
}

function normalizeProvenance(value, fallback = {}) {
  return {
    kind: asText(value?.kind, asText(fallback.kind, 'local')),
    label: asText(value?.label, asText(fallback.label, 'local source')),
    ref: asText(value?.ref) || null,
  };
}

function normalizeLink(value, index) {
  return {
    linkId: asText(value?.id || value?.linkId, `link-${index + 1}`),
    label: asText(value?.label, `Link ${index + 1}`),
    href: asText(value?.href, '#'),
  };
}

function normalizeAction(value, index) {
  return {
    actionId: asText(value?.id || value?.actionId, `action-${index + 1}`),
    label: asText(value?.label, `Action ${index + 1}`),
    href: asText(value?.href, '#'),
    kind: asText(value?.kind) === 'secondary' ? 'secondary' : 'primary',
  };
}

function normalizeSource(value, index, defaultProvenance) {
  return {
    sourceId: asText(value?.id || value?.sourceId, `source-${index + 1}`),
    kind: asText(value?.kind, 'note'),
    title: asText(value?.title, `Source ${index + 1}`),
    summary: asText(value?.summary),
    body: asText(value?.body),
    tags: uniqueStrings(value?.tags),
    links: asArray(value?.links).map(normalizeLink),
    provenance: normalizeProvenance(value?.provenance, defaultProvenance),
  };
}

function resolveSourceIds(values, sourceMap) {
  return uniqueStrings(values).filter((sourceId) => sourceMap.has(sourceId));
}

function normalizeStrip(value, index, sourceMap) {
  return {
    stripId: asText(value?.id || value?.stripId, `strip-${index + 1}`),
    label: asText(value?.label, `Strip ${index + 1}`),
    value: asText(value?.value),
    detail: asText(value?.detail),
    sourceIds: resolveSourceIds(value?.sourceIds || value?.source_ids, sourceMap),
  };
}

function normalizeModule(value, index, sourceMap) {
  return {
    moduleId: asText(value?.id || value?.moduleId, `module-${index + 1}`),
    eyebrow: asText(value?.eyebrow, 'Section'),
    title: asText(value?.title, `Module ${index + 1}`),
    body: asText(value?.body),
    sourceIds: resolveSourceIds(value?.sourceIds || value?.source_ids, sourceMap),
  };
}

function normalizeNote(value, index, sourceMap) {
  return {
    noteId: asText(value?.id || value?.noteId, `note-${index + 1}`),
    eyebrow: asText(value?.eyebrow, 'Note'),
    text: asText(value?.text),
    sourceIds: resolveSourceIds(value?.sourceIds || value?.source_ids, sourceMap),
  };
}

function normalizeMedia(value, index, sourceMap, defaultProvenance) {
  return {
    mediaId: asText(value?.id || value?.mediaId, `media-${index + 1}`),
    kind: asText(value?.kind, 'image'),
    title: asText(value?.title, `Media ${index + 1}`),
    src: asText(value?.src),
    alt: asText(value?.alt, asText(value?.title, `Media ${index + 1}`)),
    caption: asText(value?.caption),
    sourceIds: resolveSourceIds(value?.sourceIds || value?.source_ids, sourceMap),
    provenance: normalizeProvenance(value?.provenance, defaultProvenance),
  };
}

function normalizeChannelProfile(key, value, defaultProvenance) {
  return {
    channel: asText(key),
    profileId: asText(value?.profileId, `${asText(key)}-profile`),
    label: asText(value?.label, `${asText(key)} profile`),
    sourceMode: asText(value?.sourceMode, 'local'),
    summary: asText(value?.summary),
    cues: uniqueStrings(value?.cues),
    preferred: uniqueStrings(value?.preferred),
    avoid: uniqueStrings(value?.avoid),
    examples: asArray(value?.examples).map((item, index) => ({
      exampleId: asText(item?.exampleId || item?.id, `${asText(key)}-example-${index + 1}`),
      signalId: asText(item?.signalId || item?.signal_id),
      publishedAt: asText(item?.publishedAt || item?.published_at),
      excerpt: asText(item?.excerpt),
    })).filter((item) => item.excerpt),
    provenance: normalizeProvenance(value?.provenance, defaultProvenance),
  };
}

export function normalizeBrandSourcePack(input = {}) {
  const defaultProvenance = normalizeProvenance(input?.provenance);
  const sources = asArray(input?.sources).map((source, index) => normalizeSource(source, index, defaultProvenance));
  const sourceMap = new Map(sources.map((source) => [source.sourceId, source]));
  const media = asArray(input?.media)
    .map((item, index) => normalizeMedia(item, index, sourceMap, defaultProvenance))
    .filter((item) => item.src);
  const channelProfiles = Object.fromEntries(
    Object.entries(input?.channelProfiles || input?.channel_profiles || {}).map(([key, value]) => [
      key,
      normalizeChannelProfile(key, value, defaultProvenance),
    ]),
  );
  const hero = input?.sections?.hero || {};
  const mediaMap = new Map(media.map((item) => [item.mediaId, item]));

  return {
    packId: asText(input?.packId || input?.id, 'brand-source-pack'),
    label: asText(input?.label, 'Brand source pack'),
    version: asText(input?.version, 'local'),
    status: asText(input?.status, 'active'),
    updatedAt: asText(input?.updatedAt || input?.updated_at, new Date(0).toISOString()),
    provenance: defaultProvenance,
    voice: uniqueStrings(input?.voice),
    sources,
    media,
    channelProfiles,
    sections: {
      hero: {
        eyebrow: asText(hero?.eyebrow, 'Public surface'),
        title: asText(hero?.title, 'WOOHWAHAE'),
        body: asText(hero?.body),
        actions: asArray(hero?.actions).map(normalizeAction),
        sourceIds: resolveSourceIds(hero?.sourceIds || hero?.source_ids, sourceMap),
        mediaIds: uniqueStrings(hero?.mediaIds || hero?.media_ids).filter((mediaId) => mediaMap.has(mediaId)),
      },
      strips: asArray(input?.sections?.strips).map((item, index) => normalizeStrip(item, index, sourceMap)),
      modules: asArray(input?.sections?.modules).map((item, index) => normalizeModule(item, index, sourceMap)),
      notes: asArray(input?.sections?.notes).map((item, index) => normalizeNote(item, index, sourceMap)),
    },
  };
}

function attachSources(item, sourceMap) {
  return {
    ...item,
    sources: asArray(item?.sourceIds).map((sourceId) => sourceMap.get(sourceId)).filter(Boolean),
  };
}

function attachMedia(item, mediaMap, sourceMap) {
  return {
    ...item,
    media: asArray(item?.mediaIds)
      .map((mediaId) => mediaMap.get(mediaId))
      .filter(Boolean)
      .map((media) => attachSources(media, sourceMap)),
  };
}

export function buildPublicHomeBrandView(input = {}) {
  const pack = normalizeBrandSourcePack(input);
  const sourceMap = new Map(pack.sources.map((source) => [source.sourceId, source]));
  const mediaMap = new Map(pack.media.map((media) => [media.mediaId, media]));

  return {
    packId: pack.packId,
    label: pack.label,
    version: pack.version,
    status: pack.status,
    updatedAt: pack.updatedAt,
    provenance: pack.provenance,
    voice: pack.voice,
    sourceCount: pack.sources.length,
    mediaCount: pack.media.length,
    channelProfiles: pack.channelProfiles,
    media: pack.media.map((item) => attachSources(item, sourceMap)),
    hero: attachMedia(attachSources(pack.sections.hero, sourceMap), mediaMap, sourceMap),
    strips: pack.sections.strips.map((item) => attachSources(item, sourceMap)),
    modules: pack.sections.modules.map((item) => attachSources(item, sourceMap)),
    notes: pack.sections.notes.map((item) => attachSources(item, sourceMap)),
  };
}
