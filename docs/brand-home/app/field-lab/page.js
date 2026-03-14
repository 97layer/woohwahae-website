import HomeShell from '../../components/home-shell';

const presets = new Set(['membrane', 'monolith', 'constellation', 'observer', 'architect', 'quiet', 'ink', 'orbit']);

export default async function FieldLabPage({ searchParams }) {
  const params = await searchParams;
  const preset = typeof params?.p === 'string' && presets.has(params.p) ? params.p : 'membrane';
  return <HomeShell preset={preset} lab />;
}
