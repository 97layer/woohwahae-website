import { NextResponse } from 'next/server';

import { withRequestLayerOsWriteToken } from '../../../../../lib/auth/layer-os-write-token';
import { authorizeRequest } from '../../../../../lib/auth/rbac';
import { getAdminTelegramView, sendAdminTelegram } from '../../../../../lib/runtime/telegram';

export async function GET(request) {
  const auth = await authorizeRequest(request, ['founder', 'admin']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  return withRequestLayerOsWriteToken(request, async () => {
    try {
      return NextResponse.json(await getAdminTelegramView());
    } catch (error) {
      const message = error instanceof Error ? error.message : 'telegram preview failed';
      return NextResponse.json({ error: message }, { status: 502 });
    }
  });
}

export async function POST(request) {
  const auth = await authorizeRequest(request, ['founder']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  return withRequestLayerOsWriteToken(request, async () => {
    try {
      return NextResponse.json(await sendAdminTelegram());
    } catch (error) {
      const message = error instanceof Error ? error.message : 'telegram send failed';
      const status = message.includes('missing server write token') ? 503 : 502;
      return NextResponse.json({ error: message }, { status });
    }
  });
}
