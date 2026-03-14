import { NextResponse } from 'next/server';

import { authorizeRequest } from '../../../../../../lib/auth/rbac';
import { getControlTowerPacket } from '../../../../../../lib/runtime/control-tower';

export async function GET(request) {
  const auth = await authorizeRequest(request, ['founder', 'admin']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  try {
    const jobId = request.nextUrl.searchParams.get('job_id') || '';
    if (!jobId.trim()) {
      return NextResponse.json({ error: 'job_id is required' }, { status: 400 });
    }
    const packet = await getControlTowerPacket(jobId);
    return NextResponse.json({ packet });
  } catch (error) {
    return NextResponse.json({ error: error instanceof Error ? error.message : 'job packet unavailable' }, { status: 502 });
  }
}
