import HomeShell from '../../components/home-shell';
import PublicUtilityPage from '../../components/public-utility-page';

const cards = [
  {
    eyebrow: 'Status',
    title: 'Payment received',
    body: '결제 완료 이후에는 확인 메일 또는 후속 응답 경로가 조용하지만 명확하게 이어져야 합니다.',
  },
  {
    eyebrow: 'Support',
    title: 'Human contact remains open',
    body: '문제가 있거나 확인이 필요하면 hello@woohwahae.kr로 직접 이어질 수 있어야 합니다.',
  },
  {
    eyebrow: 'Flow',
    title: 'Proof over mystery',
    body: '결제 성공 뒤에는 무엇이 진행되고 무엇이 다음인지가 보이는 편이 좋습니다.',
  },
];

const actions = [
  { label: 'Offering', href: '/works/offering', desc: '서비스 축으로 돌아가기' },
  { label: 'Contact', href: '/#home-footer', desc: '문의 경로 보기' },
];

export default function PaymentSuccessPage() {
  return (
    <HomeShell preset="architect" surface="payment-success" headerTitle="payment / success">
      <PublicUtilityPage
        eyebrow="Payment Success"
        title={'결제\n완료'}
        body="결제가 정상적으로 접수되었습니다. 후속 안내와 진행 상황은 등록된 연락 경로 또는 수동 응대를 통해 이어집니다."
        note="디지털 제공이나 후속 응답이 필요한 경우, 같은 흐름 안에서 빠르게 확인되도록 서비스 루프를 계속 정리하는 중입니다."
        cards={cards}
        actions={actions}
      />
    </HomeShell>
  );
}
