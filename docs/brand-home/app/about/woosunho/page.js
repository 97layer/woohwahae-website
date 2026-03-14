import HomeShell from '../../../components/home-shell';
import PublicSurfacePage from '../../../components/public-surface-page';
import { publicSurfacePages } from '../../../content/public-surface-pages';

export default function AboutWoosunhoPage() {
  return (
    <HomeShell preset="architect" surface="about" headerTitle="about / woosunho">
      <PublicSurfacePage page={publicSurfacePages.about_woosunho} />
    </HomeShell>
  );
}
