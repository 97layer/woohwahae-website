import HomeShell from '../../../components/home-shell';
import PublicSurfacePage from '../../../components/public-surface-page';
import { publicSurfacePages } from '../../../content/public-surface-pages';

export default function WorksProjectPage() {
  return (
    <HomeShell preset="architect" surface="works" headerTitle="works / project">
      <PublicSurfacePage page={publicSurfacePages.works_project} />
    </HomeShell>
  );
}
