import { NextResponse } from 'next/server';

import { withRequestLayerOsWriteToken } from '../../../../../../lib/auth/layer-os-write-token';
import { authorizeRequest } from '../../../../../../lib/auth/rbac';
import { readValidatedJson } from '../../../../../../lib/http/request';
import { publishBrandThreads } from '../../../../../../lib/runtime/brand-publish';
import { adminBrandPublishThreadsSchema } from '../../../../../../lib/validation/admin-brand-publish-threads';

export async function POST(request) {
  const auth = await authorizeRequest(request, ['founder']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  const parsed = await readValidatedJson(request, adminBrandPublishThreadsSchema);
  if (!parsed.ok) {
    return NextResponse.json({ error: parsed.error, details: parsed.details }, { status: parsed.status });
  }

  return withRequestLayerOsWriteToken(request, async () => {
    try {
      return NextResponse.json(await publishBrandThreads(parsed.data));
    } catch (error) {
      const message = error instanceof Error ? error.message : 'threads publish failed';
      const normalized = message.toLowerCase();
      let status = 502;
      if (normalized.includes('not configured')) {
        status = 503;
      } else if (normalized.includes('already published')) {
        status = 409;
      } else if (normalized.includes('approval') || normalized.includes('draft')) {
        status = 400;
      }
      return NextResponse.json({ error: message }, { status });
    }
  });
}
