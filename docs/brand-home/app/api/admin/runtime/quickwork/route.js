import { NextResponse } from 'next/server';

import { withRequestLayerOsWriteToken } from '../../../../../lib/auth/layer-os-write-token';
import { authorizeRequest } from '../../../../../lib/auth/rbac';
import { readValidatedJson } from '../../../../../lib/http/request';
import { getQuickworkRuntimeStatus, submitQuickwork } from '../../../../../lib/runtime/quickwork';
import { adminQuickworkSchema } from '../../../../../lib/validation/admin-quickwork';

export async function GET(request) {
  const auth = await authorizeRequest(request, ['founder', 'admin']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  return withRequestLayerOsWriteToken(request, async () => {
    try {
      const runtime = await getQuickworkRuntimeStatus();
      return NextResponse.json(runtime);
    } catch (error) {
      const message = error instanceof Error ? error.message : 'quickwork status failed';
      return NextResponse.json({ error: message }, { status: 502 });
    }
  });
}

export async function POST(request) {
  const auth = await authorizeRequest(request, ['founder']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  const parsed = await readValidatedJson(request, adminQuickworkSchema);
  if (!parsed.ok) {
    return NextResponse.json({ error: parsed.error, details: parsed.details }, { status: parsed.status });
  }

  return withRequestLayerOsWriteToken(request, async () => {
    try {
      const payload = await submitQuickwork(parsed.data);
      return NextResponse.json(payload);
    } catch (error) {
      const message = error instanceof Error ? error.message : 'quickwork failed';
      const status = message.includes('missing server write token') ? 503 : 502;
      return NextResponse.json({ error: message }, { status });
    }
  });
}
