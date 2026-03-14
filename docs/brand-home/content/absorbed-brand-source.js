export const absorbedBrandSources = [
  {
    id: 'absorbed-home-gate',
    kind: 'absorbed_snapshot',
    title: 'Absorbed home gate',
    summary: '소거(消去)는 이전 브랜드 입구에서 가장 먼저 보이던 언어였다.',
    body:
      'An earlier WOOHWAHAE home snapshot framed the brand as a place that removes what is layered on top, instead of adding more signals. That lens still matters even as the shell moves closer to Layer OS.',
    tags: ['absorbed_snapshot', 'identity', 'home'],
    provenance: {
      kind: 'absorbed_snapshot',
      label: 'Absorbed website snapshot',
      ref: 'absorbed:website/index.html#meta-description',
    },
  },
  {
    id: 'absorbed-about-lens',
    kind: 'absorbed_snapshot',
    title: 'Absorbed lens',
    summary: '소거와 느림은 브랜드의 이전 장문 서사에서 반복되던 축이다.',
    body:
      'The absorbed About snapshot kept returning to subtraction, patience, and recovering an original signal beneath noise. The rebuilt shell should stay quieter than the old prose, but it should not lose that lens.',
    tags: ['absorbed_snapshot', 'about', 'lens'],
    provenance: {
      kind: 'absorbed_snapshot',
      label: 'Absorbed website snapshot',
      ref: 'absorbed:website/about/index.html#intro',
    },
  },
  {
    id: 'absorbed-name-origin',
    kind: 'absorbed_snapshot',
    title: 'Absorbed name origin',
    summary: '우화해라는 이름은 드러남, 변환, 실행을 함께 붙여 설명했다.',
    body:
      'The older site explained the name as a movement from what was already present, through transformation, into daily practice. That origin story is still useful when social drafts need a stable identity anchor.',
    tags: ['absorbed_snapshot', 'name', 'origin'],
    provenance: {
      kind: 'absorbed_snapshot',
      label: 'Absorbed website snapshot',
      ref: 'absorbed:website/about/index.html#name-origin',
    },
  },
  {
    id: 'absorbed-route-archive',
    kind: 'absorbed_snapshot',
    title: 'Absorbed route split',
    summary: 'Archive, Works, About의 분리는 예전 사이트에서도 이미 중요한 구조였다.',
    body:
      'The old site already separated archive, works, and origin so visitors could move between record, practice, and identity without collapsing them into one marketing page. The new shell keeps that route discipline.',
    tags: ['absorbed_snapshot', 'route', 'archive'],
    provenance: {
      kind: 'absorbed_snapshot',
      label: 'Absorbed website snapshot',
      ref: 'absorbed:website/_components/nav.html',
    },
  },
];

export const absorbedBrandMedia = [
  {
    id: 'absorbed-symbol-mark',
    kind: 'mark',
    title: 'WOOHWAHAE symbol',
    src: '/assets/media/brand/symbol.png',
    alt: 'WOOHWAHAE symbol mark',
    caption: 'The mark kept at the center of the WOOHWAHAE face.',
    sourceIds: ['absorbed-home-gate'],
    provenance: {
      kind: 'absorbed_snapshot',
      label: 'Absorbed brand asset',
      ref: 'absorbed:website/assets/media/brand/symbol.png',
    },
  },
  {
    id: 'absorbed-hero-field',
    kind: 'hero',
    title: 'Absorbed field graphic',
    src: '/assets/media/brand/hero-graphic.svg',
    alt: 'Animated field graphic carried into the absorbed WOOHWAHAE shell',
    caption: 'A field study that still fits the quieter public shell.',
    sourceIds: ['absorbed-home-gate', 'absorbed-about-lens'],
    provenance: {
      kind: 'absorbed_snapshot',
      label: 'Absorbed brand asset',
      ref: 'absorbed:website/assets/media/brand/hero-graphic.svg',
    },
  },
  {
    id: 'absorbed-atelier-hero',
    kind: 'hero',
    title: 'Atelier silhouette',
    src: '/assets/media/brand/hair-atelier-hero.svg',
    alt: 'Atelier silhouette carried into the absorbed WOOHWAHAE work surface',
    caption: 'A quiet silhouette for the atelier-facing surface.',
    sourceIds: ['absorbed-route-archive'],
    provenance: {
      kind: 'absorbed_snapshot',
      label: 'Absorbed brand asset',
      ref: 'absorbed:website/assets/media/brand/hair-atelier-hero.svg',
    },
  },
  {
    id: 'absorbed-ritual-objects',
    kind: 'detail',
    title: 'Ritual object',
    src: '/assets/media/brand/objects/tea.svg',
    alt: 'One of the absorbed ritual object illustrations',
    caption: 'A small object that keeps the visual family grounded.',
    sourceIds: ['absorbed-about-lens', 'absorbed-name-origin'],
    provenance: {
      kind: 'absorbed_snapshot',
      label: 'Absorbed brand asset',
      ref: 'absorbed:website/assets/media/brand/objects/tea.svg',
    },
  },
];
