import HomeShell from '../../components/home-shell';
import PublicUtilityPage from '../../components/public-utility-page';

const cards = [
  {
    eyebrow: 'Scope',
    title: 'Only the minimum needed',
    body: '문의와 운영 응답에 필요한 최소 정보만 다룹니다. 과도한 추적이나 장식적 수집을 기본값으로 두지 않습니다.',
  },
  {
    eyebrow: 'Contact',
    title: 'Human follow-up stays available',
    body: '개인정보 처리에 대한 문의는 hello@woohwahae.kr로 직접 이어질 수 있어야 합니다.',
  },
  {
    eyebrow: 'Principle',
    title: 'Clarity before complexity',
    body: '정책 문장도 고객이 이해할 수 있어야 하며, 운영 편의를 이유로 불명확한 표현을 남기지 않습니다.',
  },
];

const actions = [
  { label: 'Contact', href: '/#home-footer', desc: '문의와 응대 경로 보기' },
  { label: 'Offering', href: '/works/offering', desc: '현재 서비스 방향 보기' },
];

export default function PrivacyPage() {
  return (
    <HomeShell preset="architect" surface="privacy" headerTitle="privacy">
      <PublicUtilityPage
        eyebrow="Privacy"
        title={'개인정보\n처리 기준'}
        body="WOOHWAHAE는 B2C 서비스 운영에 필요한 최소 범위 안에서만 정보를 다룹니다. 실제 수집 범위와 응답 경로는 고객이 읽고 이해할 수 있는 수준으로 유지합니다."
        note="정식 법률 문안은 서비스 운영과 결제 흐름이 안정되는 시점에 더 구체화하되, 현재 기본 원칙은 최소 수집과 명확한 응대입니다."
        cards={cards}
        actions={actions}
      />
    </HomeShell>
  );
}
