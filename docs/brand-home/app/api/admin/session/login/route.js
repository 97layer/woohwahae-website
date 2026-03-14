import { NextResponse } from 'next/server';
import { readValidatedJson } from '../../../../../lib/http/request';
import { createFounderAdminSession, verifySessionToken } from '../../../../../lib/auth/session-token';
import { passwordLoginEnabled, verifyAdminPassword } from '../../../../../lib/auth/password-login';
import { getServerSecrets } from '../../../../../lib/env/server';
import { adminLoginSchema } from '../../../../../lib/validation/admin-login';

function normalizeNextPath(raw) {
  if (typeof raw !== 'string') return '/admin';
  const value = raw.trim();
  if (!value.startsWith('/admin')) {
    return '/admin';
  }
  return value;
}

function hasAdminRole(session) {
  return session.roles.includes('founder') || session.roles.includes('admin');
}

function sessionResponse(token, session, nextPath) {
  const response = NextResponse.json({
    ok: true,
    next: nextPath,
    session: {
      userId: session.userId,
      roles: session.roles,
      exp: session.exp,
    },
  });
  response.cookies.set('wwh_session', token, {
    httpOnly: true,
    sameSite: 'lax',
    secure: process.env.NODE_ENV === 'production',
    path: '/',
    ...(session.exp ? { expires: new Date(session.exp * 1000) } : {}),
  });
  return response;
}

export async function POST(request) {
  const parsed = await readValidatedJson(request, adminLoginSchema);
  if (!parsed.ok) {
    return NextResponse.json({ error: parsed.error, details: parsed.details }, { status: parsed.status });
  }

  const nextPath = normalizeNextPath(parsed.data.next);
  const password = typeof parsed.data.password === 'string' ? parsed.data.password.trim() : '';
  if (password) {
    if (!passwordLoginEnabled()) {
      return NextResponse.json({ error: 'password login is not configured on this server' }, { status: 503 });
    }
    if (!verifyAdminPassword(password)) {
      return NextResponse.json({ error: 'invalid password' }, { status: 401 });
    }

    const { sessionHmacSecret } = getServerSecrets();
    const { token, session } = await createFounderAdminSession(sessionHmacSecret);
    return sessionResponse(token, session, nextPath);
  }

  const { sessionHmacSecret } = getServerSecrets();
  const session = await verifySessionToken(parsed.data.token, sessionHmacSecret);
  if (!session) {
    return NextResponse.json({ error: 'invalid session token' }, { status: 401 });
  }
  if (!hasAdminRole(session)) {
    return NextResponse.json({ error: 'forbidden' }, { status: 403 });
  }

  return sessionResponse(parsed.data.token, session, nextPath);
}
