import { absorbedBrandMedia, absorbedBrandSources } from './absorbed-brand-source.js';
import { socialChannelProfiles } from './social-style-corpus.js';

export const publicHomeLocalSourcePack = {
  packId: 'woohwahae-public-home-local',
  label: 'WOOHWAHAE public home local pack',
  version: '2026-03-m3',
  status: 'active',
  updatedAt: '2026-03-11T00:00:00.000Z',
  provenance: {
    kind: 'local',
    label: 'brand-home local content',
    ref: 'docs/brand-home/content/public-home-source.js',
  },
  voice: ['calm', 'direct', 'precise', 'non-performative'],
  channelProfiles: socialChannelProfiles,
  sources: [
    {
      id: 'origin-shell',
      kind: 'origin',
      title: 'Subtractive care',
      summary: 'What does not fit is removed first.',
      body:
        'WOOHWAHAE looks for what feels layered on top and starts by removing that pressure. The public face stays quiet so care, rhythm, and fit can come forward first.',
      tags: ['care', 'posture', 'public'],
      provenance: {
        kind: 'constitution',
        label: 'Brand',
        ref: 'constitution/brand.md',
      },
    },
    {
      id: 'voice-posture',
      kind: 'voice',
      title: 'Quiet by design',
      summary: 'Calm language. Fewer signals.',
      body:
        'The surface stays calm, direct, and exact. It avoids inflated promise, trend language, and decorative warmth that does not carry real care.',
      tags: ['voice', 'surface', 'tone'],
      provenance: {
        kind: 'constitution',
        label: 'Voice',
        ref: 'constitution/voice.md',
      },
    },
    {
      id: 'surface-corridor',
      kind: 'surface',
      title: 'One quiet route',
      summary: 'Archive, practice, and origin stay distinct.',
      body:
        'Archive keeps notes and references visible, practice keeps the work grounded, and about keeps the longer philosophy available without turning the front page into an essay.',
      tags: ['archive', 'practice', 'about'],
      provenance: {
        kind: 'constitution',
        label: 'Surface',
        ref: 'constitution/surface.md',
      },
    },
    {
      id: 'proof-rule',
      kind: 'proof',
      title: 'No performance before care',
      summary: 'Do not let system language overtake the public face.',
      body:
        'The public face should not feel like a control panel. Internal systems can stay behind the surface while the customer-facing page keeps attention on care, pace, and direct contact.',
      tags: ['public', 'restraint', 'clarity'],
      provenance: {
        kind: 'constitution',
        label: 'Brand',
        ref: 'constitution/brand.md',
      },
    },
    {
      id: 'contact-route',
      kind: 'route',
      title: 'Simple route when work is real',
      summary: 'Direct contact first.',
      body:
        'When the work is specific, the route stays simple. Direct contact opens the first conversation, Archive gives context, and the protected admin surface stays out of sight.',
      tags: ['contact', 'archive', 'route'],
      provenance: {
        kind: 'constitution',
        label: 'Voice',
        ref: 'constitution/voice.md',
      },
    },
    ...absorbedBrandSources,
  ],
  media: absorbedBrandMedia,
  sections: {
    hero: {
      eyebrow: 'WOOHWAHAE',
      title: 'A quiet atelier.\nSubtractive care.',
      body:
        'WOOHWAHAE looks at hair, image, and daily rhythm through subtraction. Rather than adding more, it removes what feels layered on top and leaves room for what fits naturally.',
      actions: [
        { id: 'open-contact', label: 'Contact', href: '#home-footer', kind: 'primary' },
        { id: 'open-about', label: 'About', href: '#about', kind: 'secondary' },
      ],
      sourceIds: ['origin-shell', 'voice-posture', 'proof-rule', 'absorbed-home-gate'],
      mediaIds: ['absorbed-atelier-hero', 'absorbed-symbol-mark'],
    },
    strips: [
      {
        id: 'posture-strip',
        label: 'Lens',
        value: 'subtraction',
        detail: 'remove what does not fit',
        sourceIds: ['voice-posture'],
      },
      {
        id: 'corridor-strip',
        label: 'Pace',
        value: 'quiet / attentive',
        detail: 'begin with listening',
        sourceIds: ['surface-corridor'],
      },
      {
        id: 'proof-strip',
        label: 'Practice',
        value: 'hair / image / rhythm',
        detail: 'daily sense over trend',
        sourceIds: ['proof-rule'],
      },
      {
        id: 'absorbed-strip',
        label: 'Route',
        value: 'direct contact',
        detail: 'appointment and conversation first',
        sourceIds: ['absorbed-about-lens', 'absorbed-route-archive'],
      },
    ],
    modules: [
      {
        id: 'practice',
        eyebrow: 'Practice',
        title: 'Care that begins by removing excess',
        body:
          'The first move is not to add more. It is to notice what feels heavy, noisy, or misplaced, and to let the work become clearer from there.',
        sourceIds: ['origin-shell', 'proof-rule'],
      },
      {
        id: 'rhythm',
        eyebrow: 'Rhythm',
        title: 'Hair, image, and daily pace',
        body:
          'Beauty work is held as practice, not spectacle. The pace stays human, the surface stays calm, and the result should feel like it belongs to the person wearing it.',
        sourceIds: ['surface-corridor'],
      },
      {
        id: 'archive',
        eyebrow: 'Archive',
        title: 'Notes, references, and small studies',
        body:
          'Archive holds fragments, references, and studies that shape the work. It lets the public face stay quiet without becoming empty.',
        sourceIds: ['contact-route', 'absorbed-route-archive'],
      },
      {
        id: 'about',
        eyebrow: 'About',
        title: 'The long-form story behind the name',
        body:
          'The longer philosophy still matters. About keeps the lens, the name, and the slower story available for people who want to read beyond the first impression.',
        sourceIds: ['absorbed-home-gate', 'absorbed-about-lens', 'absorbed-name-origin'],
      },
    ],
    notes: [
      {
        id: 'voice-note',
        eyebrow: 'Lens',
        text: 'Start by removing what feels layered on top.',
        sourceIds: ['voice-posture'],
      },
      {
        id: 'proof-note',
        eyebrow: 'Pace',
        text: 'Move slowly enough to notice what actually fits.',
        sourceIds: ['proof-rule'],
      },
      {
        id: 'contact-note',
        eyebrow: 'Surface',
        text: 'Keep the page quiet enough that care stays in front.',
        sourceIds: ['contact-route'],
      },
      {
        id: 'absorbed-note',
        eyebrow: 'Route',
        text: 'When the work is specific, use direct contact first.',
        sourceIds: ['absorbed-home-gate', 'absorbed-route-archive'],
      },
    ],
  },
};

export default publicHomeLocalSourcePack;
