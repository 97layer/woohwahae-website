import { NextResponse } from 'next/server';
import { withRequestLayerOsWriteToken } from '../../../../../../lib/auth/layer-os-write-token';
import { authorizeRequest } from '../../../../../../lib/auth/rbac';
import { readValidatedJson } from '../../../../../../lib/http/request';
import {
  founderApproveSchema,
  founderReleaseSchema,
  founderRollbackSchema,
  founderStartSchema,
} from '../../../../../../lib/validation/founder-action';
import { performFounderAction } from '../../../../../../lib/runtime/view-model';

const schemaByAction = {
  start: founderStartSchema,
  approve: founderApproveSchema,
  release: founderReleaseSchema,
  rollback: founderRollbackSchema,
};

export async function POST(request, context) {
  const auth = await authorizeRequest(request, ['founder']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  const params = await context.params;
  const action = typeof params?.action === 'string' ? params.action.trim() : '';
  const schema = schemaByAction[action];
  if (!schema) {
    return NextResponse.json({ error: 'unsupported action' }, { status: 404 });
  }

  const parsed = await readValidatedJson(request, schema);
  if (!parsed.ok) {
    return NextResponse.json({ error: parsed.error, details: parsed.details }, { status: parsed.status });
  }

  return withRequestLayerOsWriteToken(request, async () => {
    try {
      const notes = ['surface:web-admin', ...(parsed.data.notes || [])];
      const payload = await performFounderAction(action, { ...parsed.data, notes });
      return NextResponse.json({ ok: true, founder: payload });
    } catch (error) {
      const message = error instanceof Error ? error.message : 'founder mutation failed';
      const status = message.includes('missing server write token') ? 503 : 502;
      return NextResponse.json({ error: message }, { status });
    }
  });
}
