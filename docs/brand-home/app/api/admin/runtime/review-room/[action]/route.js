import { NextResponse } from 'next/server';
import { withRequestLayerOsWriteToken } from '../../../../../../lib/auth/layer-os-write-token';
import { authorizeRequest } from '../../../../../../lib/auth/rbac';
import { readValidatedJson } from '../../../../../../lib/http/request';
import { performReviewRoomAction } from '../../../../../../lib/runtime/view-model';
import { reviewRoomActionSchema } from '../../../../../../lib/validation/review-room-action';

const allowedActions = new Set(['accept', 'defer', 'resolve']);

export async function POST(request, context) {
  const auth = await authorizeRequest(request, ['founder']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  const params = await context.params;
  const action = typeof params?.action === 'string' ? params.action.trim() : '';
  if (!allowedActions.has(action)) {
    return NextResponse.json({ error: 'unsupported action' }, { status: 404 });
  }

  const parsed = await readValidatedJson(request, reviewRoomActionSchema);
  if (!parsed.ok) {
    return NextResponse.json({ error: parsed.error, details: parsed.details }, { status: parsed.status });
  }

  return withRequestLayerOsWriteToken(request, async () => {
    try {
      const payload = await performReviewRoomAction(action, {
        ...parsed.data,
        rule: parsed.data.rule || 'web-admin-founder',
        evidence: ['surface:web-admin', ...(parsed.data.evidence || [])],
      });
      return NextResponse.json({ ok: true, review: payload });
    } catch (error) {
      const message = error instanceof Error ? error.message : 'review room mutation failed';
      const status = message.includes('missing server write token') ? 503 : 502;
      return NextResponse.json({ error: message }, { status });
    }
  });
}
