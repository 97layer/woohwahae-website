import { NextResponse } from 'next/server';
import { authorizeRequest } from '../../../../lib/auth/rbac';
import { readValidatedJson } from '../../../../lib/http/request';
import { adminSessionCommandSchema } from '../../../../lib/validation/admin-session';

export async function GET(request) {
  const auth = await authorizeRequest(request, ['founder', 'admin']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  return NextResponse.json({
    ok: true,
    session: {
      userId: auth.session.userId,
      roles: auth.session.roles,
      exp: auth.session.exp,
    },
  });
}

export async function POST(request) {
  const auth = await authorizeRequest(request, ['founder', 'admin']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  const parsed = await readValidatedJson(request, adminSessionCommandSchema);
  if (!parsed.ok) {
    return NextResponse.json({ error: parsed.error, details: parsed.details }, { status: parsed.status });
  }

  return NextResponse.json({
    ok: true,
    actor: auth.session.userId,
    roles: auth.session.roles,
    command: parsed.data,
  });
}
