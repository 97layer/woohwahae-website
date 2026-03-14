import { NextResponse } from 'next/server';

import { withRequestLayerOsWriteToken } from '../../../../../lib/auth/layer-os-write-token';
import { authorizeRequest } from '../../../../../lib/auth/rbac';
import { readValidatedJson } from '../../../../../lib/http/request';
import { getAdminBrandPublishView, submitBrandPublish } from '../../../../../lib/runtime/brand-publish';
import { adminBrandPublishSchema } from '../../../../../lib/validation/admin-brand-publish';

export async function GET(request) {
  const auth = await authorizeRequest(request, ['founder', 'admin']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  return withRequestLayerOsWriteToken(request, async () => {
    try {
      return NextResponse.json(await getAdminBrandPublishView());
    } catch (error) {
      const message = error instanceof Error ? error.message : 'brand publish view failed';
      return NextResponse.json({ error: message }, { status: 502 });
    }
  });
}

export async function POST(request) {
  const auth = await authorizeRequest(request, ['founder']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  const parsed = await readValidatedJson(request, adminBrandPublishSchema);
  if (!parsed.ok) {
    return NextResponse.json({ error: parsed.error, details: parsed.details }, { status: parsed.status });
  }

  return withRequestLayerOsWriteToken(request, async () => {
    try {
      return NextResponse.json(await submitBrandPublish({
        ...parsed.data,
        notes: ['surface:web-admin', ...(parsed.data.notes || [])],
      }));
    } catch (error) {
      const message = error instanceof Error ? error.message : 'brand publish failed';
      const status = message.includes('missing server write token') ? 503 : 502;
      return NextResponse.json({ error: message }, { status });
    }
  });
}
