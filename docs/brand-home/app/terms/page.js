import HomeShell from '../../components/home-shell';
import PublicUtilityPage from '../../components/public-utility-page';

const cards = [
  {
    eyebrow: 'Service',
    title: 'Customer-facing scope first',
    body: '무엇을 제공하는지와 어디까지 응답하는지를 고객 기준으로 먼저 설명해야 합니다.',
  },
  {
    eyebrow: 'Expectation',
    title: 'Quiet but explicit',
    body: '작업 범위, 응대 방식, 후속 연락 경로는 차분하되 모호하지 않게 둡니다.',
  },
  {
    eyebrow: 'Change',
    title: 'Terms must follow the actual service',
    body: '약관은 인프라 설명서가 아니라 현재 실제 서비스 흐름을 반영해야 합니다.',
  },
];

const actions = [
  { label: 'Works', href: '/works', desc: '현재 제공 구조 보기' },
  { label: 'About', href: '/about', desc: '브랜드와 운영 방향 보기' },
];

export default function TermsPage() {
  return (
    <HomeShell preset="architect" surface="terms" headerTitle="terms">
      <PublicUtilityPage
        eyebrow="Terms"
        title={'서비스\n이용 기준'}
        body="이용 기준은 고객이 무엇을 요청할 수 있고, 어떤 방식으로 안내와 후속 응답을 받는지 이해하게 만드는 문서여야 합니다. 현재는 B2C 서비스 운영 흐름에 맞춘 기본 기준을 우선 둡니다."
        note="정식 약관 세부 문안은 결제/예약/응대 플로우가 더 안정되면 현재 서비스 정의에 맞춰 추가 정리합니다."
        cards={cards}
        actions={actions}
      />
    </HomeShell>
  );
}
