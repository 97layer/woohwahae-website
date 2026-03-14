import { publicHomeLocalSourcePack } from '../content/public-home-source.js';
import { defaultSocialAccount, socialAccountProfile } from '../content/social-account-profiles.js';
import { normalizeBrandSourcePack } from './brand-source-pack.js';

function asArray(value) {
  return Array.isArray(value) ? value : [];
}

function asText(value, fallback = '') {
  return typeof value === 'string' ? value.trim() : fallback;
}

function uniqueStrings(values) {
  return Array.from(new Set(asArray(values).map((value) => asText(value)).filter(Boolean)));
}

function limitText(value, max) {
  const text = asText(value);
  if (!max || text.length <= max) {
    return text;
  }
  return `${text.slice(0, Math.max(0, max - 1)).trim()}…`;
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

function profileExamples(profile, limit = 2) {
  return asArray(profile?.examples)
    .filter((item) => asText(item?.excerpt))
    .slice(0, limit)
    .map((item) => ({
      exampleId: asText(item.exampleId),
      signalId: asText(item.signalId),
      publishedAt: asText(item.publishedAt),
      excerpt: asText(item.excerpt),
    }));
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
  const pack = normalizeBrandSourcePack(packInput);
  const hero = pack.sections.hero;
  const baseTitle = limitText(hero.title.replace(/\s*\n+\s*/g, ' ').trim(), 110);
  const sourceIds = resolveSourceIds(pack, hero.sourceIds);
  const media = heroMedia(pack);
  const mediaIds = media ? [media.mediaId] : [];
  const mediaCue = media ? `${media.title}: ${media.caption || media.alt}` : '';
  const threadsProfile = channelProfile(pack, 'threads');
  const xProfile = channelProfile(pack, 'x') || channelProfile(pack, 'instagram');
  const telegramProfile = channelProfile(pack, 'telegram');
  const threadsExamples = profileExamples(threadsProfile);
  const xExamples = profileExamples(xProfile);
  const telegramExamples = profileExamples(telegramProfile);
  const threadsAccount = socialAccountProfile(defaultSocialAccount('threads'));
  const xAccount = socialAccountProfile(defaultSocialAccount('x'));
  const telegramAccount = socialAccountProfile(defaultSocialAccount('telegram'));

  return [
    {
      presetId: 'threads-launch',
      label: 'Threads 초안',
      channel: 'threads',
      title: baseTitle,
      body: buildDraftBody(pack, 'threads'),
      sourceIds,
      mediaIds,
      mediaCue,
      targetAccount: threadsAccount.accountId,
      targetAccountLabel: threadsAccount.label,
      styleProfile: threadsProfile?.profileId || '',
      styleCue: profileCue(threadsProfile),
      styleExamples: threadsExamples,
      styleExampleIds: threadsExamples.map((item) => item.exampleId),
      sourceLabels: sourceLabels(pack, sourceIds),
      note: [threadsProfile?.summary, mediaCue].filter(Boolean).join(' · ') || '짧은 문단 중심 공개 초안입니다.',
    },
    {
      presetId: 'x-signal',
      label: 'X 짧은 공지',
      channel: 'x',
      title: limitText(baseTitle, 90),
      body: buildDraftBody(pack, 'x'),
      sourceIds,
      mediaIds,
      mediaCue,
      targetAccount: xAccount.accountId,
      targetAccountLabel: xAccount.label,
      styleProfile: xProfile?.profileId || '',
      styleCue: profileCue(xProfile),
      styleExamples: xExamples,
      styleExampleIds: xExamples.map((item) => item.exampleId),
      sourceLabels: sourceLabels(pack, sourceIds),
      note: [xProfile?.summary, mediaCue].filter(Boolean).join(' · ') || '짧고 바로 읽히는 상태 공지 톤입니다.',
    },
    {
      presetId: 'telegram-brief',
      label: 'Telegram 요약',
      channel: 'telegram',
      title: baseTitle,
      body: buildDraftBody(pack, 'telegram'),
      sourceIds,
      mediaIds,
      mediaCue,
      targetAccount: telegramAccount.accountId,
      targetAccountLabel: telegramAccount.label,
      styleProfile: telegramProfile?.profileId || '',
      styleCue: profileCue(telegramProfile),
      styleExamples: telegramExamples,
      styleExampleIds: telegramExamples.map((item) => item.exampleId),
      sourceLabels: sourceLabels(pack, sourceIds),
      note: [telegramProfile?.summary, mediaCue].filter(Boolean).join(' · ') || '조금 더 긴 운영/브랜드 요약용 초안입니다.',
    },
  ];
}
