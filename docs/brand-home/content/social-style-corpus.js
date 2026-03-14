import socialStyleExamples from './social-style-examples.generated.js';

function asExamples(value) {
  return Array.isArray(value) ? value.filter((item) => item?.excerpt) : [];
}

const instagramExamples = asExamples(socialStyleExamples.instagram);
const threadsExamples = asExamples(socialStyleExamples.threads);

export const socialChannelProfiles = {
  threads: {
    profileId: 'threads-repo-local-snapshot-v1',
    label: 'Threads repo-local snapshot profile',
    sourceMode: 'repo_local_snapshot',
    summary: '짧은 관찰에서 시작해, 여백을 남기고, 마지막에 조용한 질문을 건다.',
    cues: [
      'open with one concrete observation',
      'keep paragraphs short',
      'let one reflective question linger near the end',
      'end with a quiet continuation instead of a hard CTA',
    ],
    preferred: ['observation', 'restraint', 'one lingering question'],
    avoid: ['marketing crescendo', 'generic inspiration', 'too many claims'],
    examples: threadsExamples,
    provenance: {
      kind: 'repo_local_snapshot',
      label: 'Repo-local social style snapshot',
      ref: 'docs/brand-home/content/social-style-examples.generated.js',
    },
  },
  instagram: {
    profileId: 'instagram-repo-local-snapshot-v1',
    label: 'Instagram repo-local snapshot profile',
    sourceMode: 'repo_local_snapshot',
    summary: '이미지가 먼저 열리고, 캡션은 사유를 한 번만 더 눌러 적는다.',
    cues: [
      'let the image carry first contact',
      'use one anchoring phrase, not three',
      'leave room for the visual texture',
      'close with one reflective line instead of explanation',
    ],
    preferred: ['visual-first', 'single anchor phrase', 'spare caption'],
    avoid: ['explainer tone', 'dense paragraph stack', 'broad CTA'],
    examples: instagramExamples,
    provenance: {
      kind: 'repo_local_snapshot',
      label: 'Repo-local social style snapshot',
      ref: 'docs/brand-home/content/social-style-examples.generated.js',
    },
  },
  telegram: {
    profileId: 'telegram-operator-v1',
    label: 'Telegram operator profile',
    sourceMode: 'local',
    summary: '조금 더 직접적이되, 결정과 상태가 먼저 보이게 쓴다.',
    cues: [
      'state first, detail second',
      'keep the opening line operational',
      'preserve quiet tone even when urgent',
    ],
    preferred: ['state', 'decision', 'specific next step'],
    avoid: ['decorative lead-in', 'long setup', 'ambiguous ask'],
    provenance: {
      kind: 'local',
      label: 'Operator model',
      ref: 'docs/telegram_operating_model.md',
    },
  },
};
