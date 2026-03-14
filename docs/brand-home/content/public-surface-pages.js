const archiveLinks = [
  { label: 'Log', href: '/archive/log', desc: '사유 / 기록' },
  { label: 'Curation', href: '/archive/curation', desc: '시선 / 청음' },
];

const worksLinks = [
  { label: 'Atelier', href: '/works/atelier', desc: '작업' },
  { label: 'Offering', href: '/works/offering', desc: '서비스 / 제품' },
  { label: 'Project', href: '/works/project', desc: '협업' },
];

const aboutLinks = [
  { label: 'Root', href: '/about', desc: '근원' },
  { label: 'Woosunho', href: '/about/woosunho', desc: '기획자' },
];

function withActive(links, activeHref) {
  return links.map((link) => ({
    ...link,
    active: link.href === activeHref,
  }));
}

export const publicSurfacePages = {
  archive_index: {
    slug: 'archive',
    eyebrow: 'Archive',
    title: 'Archive',
    body: 'Archive는 기록과 큐레이션 두 갈래를 나란히 두는 공용 분류입니다. 상단에서 바로 하위 축으로 이동할 수 있게 정렬합니다.',
    panel: {
      eyebrow: 'Children',
      title: 'Log / Curation',
      body: 'Archive 아래에는 Log와 Curation 두 축만 둡니다.',
      links: withActive(archiveLinks),
    },
    section: {
      eyebrow: 'Structure',
      title: 'Archive children',
      cards: [
        {
          eyebrow: 'Log',
          title: '기록과 사유',
          body: '작업 중 남는 문장, 메모, 관찰을 쌓는 자리입니다.',
        },
        {
          eyebrow: 'Curation',
          title: '시선과 청음',
          body: '남겨둘 레퍼런스와 선택 기준을 정리하는 자리입니다.',
        },
      ],
    },
  },
  archive_log: {
    slug: 'archive-log',
    eyebrow: 'Archive / Log',
    title: 'Log',
    body: 'Log는 작업 중 남는 기록과 사유를 모으는 자리입니다. 설명을 크게 늘리지 않고, 실제로 남겨야 할 문장 위주로 정리합니다.',
    panel: {
      eyebrow: 'Archive',
      title: 'Log / Curation',
      body: 'Archive 안에서 Log는 기록 쪽 축입니다.',
      links: withActive(archiveLinks, '/archive/log'),
    },
    section: {
      eyebrow: 'Log',
      title: 'What belongs in Log',
      cards: [
        {
          eyebrow: 'Record',
          title: 'Short notes from actual work',
          body: '작업 중 바로 남기는 짧은 기록을 둡니다.',
        },
        {
          eyebrow: 'Observation',
          title: 'Small shifts worth keeping',
          body: '방향을 바꾸는 작은 변화와 관찰을 남깁니다.',
        },
        {
          eyebrow: 'Rhythm',
          title: 'Slow accumulation',
          body: '한 번에 설명하기보다 천천히 축적되는 결을 봅니다.',
        },
      ],
    },
  },
  archive_curation: {
    slug: 'archive-curation',
    eyebrow: 'Archive / Curation',
    title: 'Curation',
    body: 'Curation은 시선과 청음을 기준으로 고른 자료를 정리하는 자리입니다. 양보다 선택 기준이 먼저 보이도록 둡니다.',
    panel: {
      eyebrow: 'Archive',
      title: 'Log / Curation',
      body: 'Archive 안에서 Curation은 선택과 배열 쪽 축입니다.',
      links: withActive(archiveLinks, '/archive/curation'),
    },
    section: {
      eyebrow: 'Curation',
      title: 'What belongs in Curation',
      cards: [
        {
          eyebrow: 'Sight',
          title: 'Visual references',
          body: '형태, 질감, 균형감을 남기는 시각 레퍼런스를 다룹니다.',
        },
        {
          eyebrow: 'Listening',
          title: 'Listening references',
          body: '리듬과 호흡을 남기는 청음 단서를 모읍니다.',
        },
        {
          eyebrow: 'Selection',
          title: 'Chosen, not crowded',
          body: '많이 쌓기보다 남겨야 할 것을 추립니다.',
        },
      ],
    },
  },
  works_index: {
    slug: 'works',
    eyebrow: 'Works',
    title: 'Works',
    body: 'Works는 작업, 오퍼링, 프로젝트 세 축을 정리하는 상위 분류입니다. 작업 성격이 섞이지 않도록 하위로 바로 나눕니다.',
    panel: {
      eyebrow: 'Children',
      title: 'Atelier / Offering / Project',
      body: 'Works 아래에는 세 갈래만 남기고 과한 소개 화면은 두지 않습니다.',
      links: withActive(worksLinks),
    },
    section: {
      eyebrow: 'Structure',
      title: 'Works children',
      cards: [
        {
          eyebrow: 'Atelier',
          title: '실제 작업',
          body: '헤어와 이미지 작업 자체를 다루는 축입니다.',
        },
        {
          eyebrow: 'Offering',
          title: '서비스와 제품',
          body: '고객에게 제안 가능한 형태를 정리하는 축입니다.',
        },
        {
          eyebrow: 'Project',
          title: '협업과 적용',
          body: '외부 협업과 적용형 작업을 남기는 축입니다.',
        },
      ],
    },
  },
  works_atelier: {
    slug: 'works-atelier',
    eyebrow: 'Works / Atelier',
    title: 'Atelier',
    body: 'Atelier는 우화해의 실제 작업 축입니다. 헤어와 이미지 작업을 중심에 두고, 사람에게 맞는 결을 찾는 쪽으로 정렬합니다.',
    panel: {
      eyebrow: 'Works',
      title: 'Atelier / Offering / Project',
      body: 'Works 안에서 Atelier는 가장 직접적인 작업 축입니다.',
      links: withActive(worksLinks, '/works/atelier'),
    },
    section: {
      eyebrow: 'Atelier',
      title: 'What belongs in Atelier',
      cards: [
        {
          eyebrow: 'Care',
          title: 'Hair and image practice',
          body: '헤어와 이미지의 실제 작업 축을 담습니다.',
        },
        {
          eyebrow: 'Approach',
          title: 'Subtraction before styling',
          body: '먼저 덜어내고, 나중에 더하는 순서로 접근합니다.',
        },
        {
          eyebrow: 'Fit',
          title: 'What suits the person',
          body: '유행보다 사람에게 맞는 균형을 우선합니다.',
        },
      ],
    },
  },
  works_offering: {
    slug: 'works-offering',
    eyebrow: 'Works / Offering',
    title: 'Offering',
    body: 'Offering은 실제로 제안 가능한 서비스와 제품 축입니다. 작업 경험을 어떤 형태로 제공할지 정리하는 자리입니다.',
    panel: {
      eyebrow: 'Works',
      title: 'Atelier / Offering / Project',
      body: 'Works 안에서 Offering은 제공 방식과 구성을 다룹니다.',
      links: withActive(worksLinks, '/works/offering'),
    },
    section: {
      eyebrow: 'Offering',
      title: 'What belongs in Offering',
      cards: [
        {
          eyebrow: 'Service',
          title: 'Customer-facing service shape',
          body: '예약, 응대, 작업 제안을 고객 기준으로 정리합니다.',
        },
        {
          eyebrow: 'Product',
          title: 'Product or package direction',
          body: '필요한 경우 제품과 패키지 방향을 다룹니다.',
        },
        {
          eyebrow: 'Clarity',
          title: 'Clear before broad',
          body: '종류를 넓히기보다 무엇을 제공하는지 먼저 또렷하게 둡니다.',
        },
      ],
    },
  },
  works_project: {
    slug: 'works-project',
    eyebrow: 'Works / Project',
    title: 'Project',
    body: 'Project는 외부 협업과 적용형 작업을 정리하는 축입니다. 소개보다 협업 구조와 역할이 먼저 보이게 둡니다.',
    panel: {
      eyebrow: 'Works',
      title: 'Atelier / Offering / Project',
      body: 'Works 안에서 Project는 협업 구조를 보여주는 자리입니다.',
      links: withActive(worksLinks, '/works/project'),
    },
    section: {
      eyebrow: 'Project',
      title: 'What belongs in Project',
      cards: [
        {
          eyebrow: 'Collaboration',
          title: 'Partnership structure',
          body: '누구와 어떤 방식으로 협업하는지 명확히 남깁니다.',
        },
        {
          eyebrow: 'Scope',
          title: 'Applied outcomes',
          body: '단순 소개보다 실제 적용 범위와 역할을 보여줍니다.',
        },
        {
          eyebrow: 'Discipline',
          title: 'Calm and exact',
          body: '프로젝트 축도 과장 없이 조용하고 정확하게 정렬합니다.',
        },
      ],
    },
  },
  about_root: {
    slug: 'about-root',
    eyebrow: 'About / Root',
    title: 'Root',
    body: 'Root는 우화해를 움직이는 근원 축입니다. 소거와 느림, 그리고 실제 작업으로 이어지는 태도를 여기서 정리합니다.',
    panel: {
      eyebrow: 'About',
      title: 'Root / Woosunho',
      body: 'About은 근원과 사람의 두 축으로 정리합니다.',
      links: withActive(aboutLinks, '/about'),
    },
    section: {
      eyebrow: 'Root',
      title: 'What belongs in Root',
      cards: [
        {
          eyebrow: 'Lens',
          title: 'Subtraction',
          body: '덧씌워진 것을 걷어내는 렌즈가 근원 축의 중심입니다.',
        },
        {
          eyebrow: 'Tempo',
          title: 'Slowness',
          body: '빠르게 채우기보다 충분히 기다리는 태도를 둡니다.',
        },
        {
          eyebrow: 'Direction',
          title: 'Daily practice',
          body: '철학을 문장에만 두지 않고 일상의 선택으로 연결합니다.',
        },
      ],
    },
  },
  about_woosunho: {
    slug: 'about-woosunho',
    eyebrow: 'About / Woosunho',
    title: 'Woosunho',
    body: 'Woosunho는 우화해의 기획과 운영, 실제 작업 흐름을 함께 붙드는 축입니다. 브랜드 언어와 실행이 따로 놀지 않게 정렬합니다.',
    panel: {
      eyebrow: 'About',
      title: 'Root / Woosunho',
      body: 'Woosunho 축은 사람과 역할을 보여주는 자리입니다.',
      links: withActive(aboutLinks, '/about/woosunho'),
    },
    section: {
      eyebrow: 'Woosunho',
      title: 'What belongs here',
      cards: [
        {
          eyebrow: 'Role',
          title: 'Founder-led direction',
          body: '브랜드 방향과 실제 작업 흐름을 founder 기준으로 붙듭니다.',
        },
        {
          eyebrow: 'Practice',
          title: 'Planning close to making',
          body: '사유, 기획, 실행이 너무 멀어지지 않게 정렬합니다.',
        },
        {
          eyebrow: 'Tone',
          title: 'Quiet, exact, non-performative',
          body: '사람 설명도 과장 없이 조용하고 정확한 톤을 유지합니다.',
        },
      ],
    },
  },
};
