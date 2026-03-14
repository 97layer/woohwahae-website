export const socialAccountProfiles = [
  {
    accountId: '97layer',
    label: '97layer',
    toneLevel: 'raw',
    summary: '원천 소스 허브. 만드는 과정, 외부 텍스트 반응, 일지, 비하인드가 먼저 쌓이는 자리.',
    sourceBoundary: '주식, 인문학, 개발, 미용, 시스템, 일상까지 넓게 받아들인다.',
    outputCue: '메이커 일지 / 비개발자 일지 / 브랜드 구축 비하인드',
  },
  {
    accountId: 'woosunhokr',
    label: '우순호',
    toneLevel: 'refined',
    summary: '같은 사유를 미용 실무와 개인 브랜드 문장으로 눌러쓴 자리.',
    sourceBoundary: '미용 실무, 미용사의 단상, 손기술과 태도의 번역에 집중한다.',
    outputCue: '미용사의 단상 / 실무적 통찰 / 더 정제된 개인 브랜드 노트',
  },
  {
    accountId: 'woohwahae',
    label: '우화해',
    toneLevel: 'polished',
    summary: '여러 원천에서 덜어낸 뒤 남는 공적인 브랜드 셸.',
    sourceBoundary: '슬로우 라이프, 조용한 매거진 톤, 공적인 브랜드 세계관으로 번역한다.',
    outputCue: '매거진형 브랜드 노트 / 슬로우 라이프 / 공적인 브랜드 문장',
  },
];

export function socialAccountProfile(accountId, fallback = '97layer') {
  const match = socialAccountProfiles.find((item) => item.accountId === String(accountId || '').trim());
  if (match) {
    return match;
  }
  return socialAccountProfiles.find((item) => item.accountId === fallback) || socialAccountProfiles[0];
}

export function defaultSocialAccount(channel) {
  switch (String(channel || '').trim().toLowerCase()) {
    case 'threads':
      return '97layer';
    case 'telegram':
      return 'woohwahae';
    case 'x':
      return 'woohwahae';
    default:
      return '97layer';
  }
}
