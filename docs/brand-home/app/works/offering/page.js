import HomeShell from '../../../components/home-shell';
import PublicSurfacePage from '../../../components/public-surface-page';
import { publicSurfacePages } from '../../../content/public-surface-pages';

export default function WorksOfferingPage() {
  return (
    <HomeShell preset="architect" surface="works" headerTitle="works / offering">
      <PublicSurfacePage page={publicSurfacePages.works_offering} />
    </HomeShell>
  );
}
