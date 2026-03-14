import { NextResponse } from 'next/server';

import { authorizeRequest } from '../../../../../lib/auth/rbac';
import {
  attachLayerOsWriteTokenCookie,
  clearLayerOsWriteTokenCookie,
  resolveLayerOsWriteToken,
} from '../../../../../lib/auth/layer-os-write-token';
import { readValidatedJson } from '../../../../../lib/http/request';
import { adminRuntimeTokenSchema } from '../../../../../lib/validation/admin-runtime-token';

export async function GET(request) {
  const auth = await authorizeRequest(request, ['founder', 'admin']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  const token = resolveLayerOsWriteToken(request);
  return NextResponse.json({
    configured: token.configured,
    source: token.source,
  });
}

export async function POST(request) {
  const auth = await authorizeRequest(request, ['founder']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  const parsed = await readValidatedJson(request, adminRuntimeTokenSchema);
  if (!parsed.ok) {
    return NextResponse.json({ error: parsed.error, details: parsed.details }, { status: parsed.status });
  }

  const response = NextResponse.json({
    ok: true,
    configured: true,
    source: 'cookie',
  });
  return attachLayerOsWriteTokenCookie(response, parsed.data.token);
}

export async function DELETE(request) {
  const auth = await authorizeRequest(request, ['founder']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  const response = NextResponse.json({
    ok: true,
    configured: false,
    source: 'missing',
  });
  return clearLayerOsWriteTokenCookie(response);
}
