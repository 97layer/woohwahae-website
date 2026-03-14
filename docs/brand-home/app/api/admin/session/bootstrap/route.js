import { NextResponse } from 'next/server';

import { createFounderAdminSession, localAdminBootstrapEnabled } from '../../../../../lib/auth/session-token';
import { getServerSecrets } from '../../../../../lib/env/server';

function isLocalHostname(hostname) {
  const value = String(hostname || '').trim().toLowerCase();
  return value === 'localhost' || value === '127.0.0.1' || value === '::1' || value === '[::1]';
}

export async function POST(request) {
  if (!localAdminBootstrapEnabled()) {
    return NextResponse.json({ error: 'local bootstrap disabled' }, { status: 404 });
  }
  if (!isLocalHostname(request.nextUrl.hostname)) {
    return NextResponse.json({ error: 'local bootstrap only works on localhost' }, { status: 403 });
  }

  const { sessionHmacSecret } = getServerSecrets();
  const { token, session } = await createFounderAdminSession(sessionHmacSecret);

  const response = NextResponse.json({
    ok: true,
    next: '/admin',
    session,
  });
  response.cookies.set('wwh_session', token, {
    httpOnly: true,
    sameSite: 'lax',
    secure: process.env.NODE_ENV === 'production',
    path: '/',
    expires: new Date(session.exp * 1000),
  });
  return response;
}
