import HomeShell from '../../../components/home-shell';
import ArchiveLogLayout from '../../../components/archive-log-layout';

export default function ArchiveLogPage() {
  return (
    <HomeShell preset="architect" surface="archive" headerTitle="archive / log">
      <ArchiveLogLayout />
    </HomeShell>
  );
}
