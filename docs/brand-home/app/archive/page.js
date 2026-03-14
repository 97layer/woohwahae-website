import HomeShell from '../../components/home-shell';
import PublicSurfacePage from '../../components/public-surface-page';
import { publicSurfacePages } from '../../content/public-surface-pages';

export default function ArchiveIndexPage() {
  return (
    <HomeShell preset="architect" surface="archive" headerTitle="archive">
      <PublicSurfacePage page={publicSurfacePages.archive_index} />
    </HomeShell>
  );
}
