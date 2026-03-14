import HomeShell from '../components/home-shell';
import PublicUtilityPage from '../components/public-utility-page';

const cards = [
  {
    eyebrow: 'Route',
    title: 'This page is not here',
    body: '찾으시는 페이지가 현재 public shell 안에 없거나, 이전 주소가 새 경로로 정리된 상태일 수 있습니다.',
  },
  {
    eyebrow: 'Return',
    title: 'Use the live shell',
    body: '홈, Works, About, Archive 안에서 현재 살아 있는 경로로 다시 들어오는 편이 가장 빠릅니다.',
  },
];

const actions = [
  { label: 'Home', href: '/', desc: '공개 메인으로 돌아가기' },
  { label: 'Works', href: '/works', desc: '현재 서비스 축 보기' },
];

export default function NotFoundPage() {
  return (
    <HomeShell preset="architect" surface="not-found" headerTitle="404">
      <PublicUtilityPage
        eyebrow="404"
        title={'페이지를\n찾을 수 없음'}
        body="현재 WOOHWAHAE shell 안에서 이 주소는 더 이상 쓰이지 않거나, 새 경로로 이동되었습니다."
        note="예전 website 주소 일부는 새 shell route로 계속 붙이는 중입니다."
        cards={cards}
        actions={actions}
      />
    </HomeShell>
  );
}
