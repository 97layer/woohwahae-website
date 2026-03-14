import { socialAccountProfile } from '../content/social-account-profiles.js';

function asArray(value) {
  return Array.isArray(value) ? value : [];
}

function asText(value, fallback = '') {
  return typeof value === 'string' ? value.trim() : fallback;
}

function limitText(value, max) {
  const text = asText(value);
  if (!max || text.length <= max) {
    return text;
  }
  return `${text.slice(0, Math.max(0, max - 1)).trim()}…`;
}

function collapseSpaces(value) {
  return asText(value).replace(/\s+/g, ' ').trim();
}

function uniqueStrings(values) {
  return Array.from(new Set(asArray(values).map((value) => asText(value)).filter(Boolean)));
}

function splitParagraphs(value) {
  return String(value || '')
    .replace(/\r\n/g, '\n')
    .split(/\n{2,}/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function containsCue(text, ...needles) {
  return needles.some((needle) => text.includes(needle));
}

function prepSubject(draft) {
  const sourceTitle = collapseSpaces(draft?.sourceTitle);
  if (sourceTitle) {
    return limitText(sourceTitle, 92);
  }
  const title = collapseSpaces(draft?.title)
    .replace(/^(97layer raw draft|우순호 draft|우화해 draft|draft)\s*·\s*/i, '')
    .trim();
  if (title) {
    return limitText(title, 92);
  }
  return limitText(collapseSpaces(draft?.sourceURL), 92) || 'source note';
}

function prepFocusCue(draft) {
  const combined = collapseSpaces([
    ...asArray(draft?.domainTags),
    ...asArray(draft?.worldviewTags),
  ].join(' ')).toLowerCase();
  switch (true) {
    case containsCue(combined, 'beauty', 'hair', 'aesthetic', 'style', 'practice', 'craft', '미용'):
      return '기준과 손기술 사이의 감각';
    case containsCue(combined, 'system', 'dev', 'development', 'build', 'product', 'automation', 'operator', '개발', '시스템'):
      return '기능보다 구조와 순서';
    case containsCue(combined, 'finance', 'stock', 'market', 'money', 'investment', '주식', '금융'):
      return '숫자보다 기준과 리듬';
    case containsCue(combined, 'brand', 'identity', 'subtraction', 'worldview', '브랜드', '정체성', '소거'):
      return '기준을 덜어내며 선명하게 만드는 감각';
    default:
      return '기준과 태도';
  }
}

function isMetaParagraph(paragraph) {
  const text = collapseSpaces(paragraph);
  return (
    text.startsWith('요즘 붙들고 있는 건') ||
    text.includes('우순호 쪽으로 옮기면 결국') ||
    text.includes('우화해 쪽으로 옮기면 결국') ||
    text.startsWith('97layer에서는') ||
    text.startsWith('우순호에서는') ||
    text.startsWith('우화해에서는') ||
    text.includes('기록에 가까운 메모로 둔다') ||
    text.includes('더 정리해본다') ||
    text.includes('더 다듬는다')
  );
}

function pickExcerptParagraph(draft) {
  const founderNote = collapseSpaces(draft?.founderNote);
  const revisionNote = collapseSpaces(draft?.revisionNote);
  const paragraphs = splitParagraphs(draft?.draft)
    .filter((paragraph) => !isMetaParagraph(paragraph))
    .filter((paragraph) => collapseSpaces(paragraph) !== founderNote)
    .filter((paragraph) => revisionNote === '' || !collapseSpaces(paragraph).includes(revisionNote));

  return paragraphs.sort((left, right) => right.length - left.length)[0] || '';
}

function prepLead(accountId, subject, focusCue) {
  switch (asText(accountId)) {
    case '97layer':
      return `요즘 자꾸 다시 보게 되는 건 ${subject} 안에 들어 있는 ${focusCue} 쪽이다.`;
    case 'woosunhokr':
      return `${subject}를 보다 보면 결국 ${focusCue}에 가까운 순간이 사람에게 어떻게 닿는지가 먼저 남는다.`;
    case 'woohwahae':
      return `${subject}에서 덜어내고 남는 건 결국 ${focusCue}에 가까운 리듬이 생활 안으로 번지는 순간이다.`;
    default:
      return `${subject}를 다시 보면 결국 ${focusCue}가 먼저 남는다.`;
  }
}

function founderCue(draft) {
  const note = collapseSpaces(draft?.founderNote);
  if (!note) {
    return '';
  }
  switch (asText(draft?.targetAccount)) {
    case '97layer':
      return limitText(`이번엔 ${note}라는 축을 더 또렷하게 적어둔다.`, 160);
    case 'woosunhokr':
      return limitText(`이 소재는 ${note}라는 결로 더 눌러볼 만하다.`, 160);
    case 'woohwahae':
      return limitText(`이 소재는 ${note}라는 결로 더 덜어낼 만하다.`, 160);
    default:
      return limitText(note, 160);
  }
}

function revisionCue(draft) {
  const note = collapseSpaces(draft?.revisionNote);
  if (!note) {
    return '';
  }
  switch (asText(draft?.targetAccount)) {
    case '97layer':
      return limitText(`이번 메모는 ${note}라는 요청을 따라 다시 눌러본 버전이다.`, 170);
    case 'woosunhokr':
      return limitText(`이번 버전은 ${note}라는 요청을 따라 더 정제해본다.`, 170);
    case 'woohwahae':
      return limitText(`이번 버전은 ${note}라는 요청을 따라 한 번 더 덜어낸다.`, 170);
    default:
      return limitText(`이번 버전은 ${note}라는 요청을 따라 다시 잡는다.`, 170);
  }
}

function joinParagraphs(items) {
  return uniqueStrings(items.map((item) => asText(item)).filter(Boolean)).join('\n\n');
}

export function buildSourceDraftPrepShape(draft, options = {}) {
  const targetAccount = asText(draft?.targetAccount, '97layer');
  const account = socialAccountProfile(targetAccount);
  const subject = prepSubject(draft);
  const focusCue = prepFocusCue(draft);
  const excerpt = pickExcerptParagraph(draft);
  const body = joinParagraphs([
    prepLead(targetAccount, subject, focusCue),
    excerpt,
    founderCue(draft),
    revisionCue(draft),
  ]);

  return {
    channel: asText(options?.channel, 'threads'),
    title: subject,
    body,
    bodyPreview: limitText(collapseSpaces(body), 180),
    prepShapeId: 'source_draft_threads_v1',
    targetAccount,
    targetAccountLabel: account.label,
  };
}
