import { NextResponse } from 'next/server';
import { withRequestLayerOsWriteToken } from '../../../../../lib/auth/layer-os-write-token';
import { authorizeRequest } from '../../../../../lib/auth/rbac';
import { getAdminOverviewView } from '../../../../../lib/runtime/view-model';

export async function GET(request) {
  const auth = await authorizeRequest(request, ['founder', 'admin']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  return withRequestLayerOsWriteToken(request, async () => {
    try {
      return NextResponse.json({ overview: await getAdminOverviewView() });
    } catch (error) {
      return NextResponse.json({ error: error instanceof Error ? error.message : 'runtime unavailable' }, { status: 502 });
    }
  });
}
