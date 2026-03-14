import HomeShell from '../../components/home-shell';
import PublicUtilityPage from '../../components/public-utility-page';

const cards = [
  {
    eyebrow: 'Retry',
    title: 'Try again calmly',
    body: '실패는 즉시 재시도 경로나 대체 문의 경로가 함께 보여야 불필요한 이탈이 줄어듭니다.',
  },
  {
    eyebrow: 'Support',
    title: 'Manual help stays available',
    body: '자동 결제가 매끄럽지 않아도 사람 응대 경로가 열려 있어야 합니다.',
  },
  {
    eyebrow: 'Signal',
    title: 'Failure should still be legible',
    body: '무슨 문제가 있었는지 완벽하지 않더라도, 고객 입장에서 다음 선택지가 보여야 합니다.',
  },
];

const actions = [
  { label: 'Try again', href: '/works/offering', desc: '서비스 페이지로 돌아가기' },
  { label: 'Contact', href: '/#home-footer', desc: '직접 문의하기' },
];

export default function PaymentFailPage() {
  return (
    <HomeShell preset="architect" surface="payment-fail" headerTitle="payment / fail">
      <PublicUtilityPage
        eyebrow="Payment Failed"
        title={'결제\n실패'}
        body="결제가 완료되지 않았습니다. 다시 시도하거나, 필요한 경우 바로 문의 경로로 이어질 수 있어야 합니다."
        note="문제가 반복되면 hello@woohwahae.kr로 알려주시면 수동 확인 경로로 이어집니다."
        cards={cards}
        actions={actions}
      />
    </HomeShell>
  );
}
