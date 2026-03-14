import { NextResponse } from 'next/server';
import { extractSessionToken, resolveSessionSecret, verifySessionToken } from './lib/auth/session-token';

const canonicalAdminHost = (process.env.LAYER_OS_ADMIN_HOST || 'admin.woohwahae.kr').trim().toLowerCase();
const canonicalPublicHosts = new Set(
  (process.env.LAYER_OS_PUBLIC_HOSTS || 'woohwahae.kr,www.woohwahae.kr')
    .split(',')
    .map((value) => value.trim().toLowerCase())
    .filter(Boolean),
);

function isProtectedApiPath(pathname) {
  return pathname.startsWith('/api/admin');
}

function isPublicAdminAuthPath(pathname) {
  return pathname === '/api/admin/session/login' || pathname === '/api/admin/session/bootstrap';
}

function isProtectedPagePath(pathname) {
  return pathname.startsWith('/admin') && pathname !== '/admin/login';
}

function isAdminSurfacePath(pathname) {
  return pathname.startsWith('/admin') || pathname.startsWith('/api/admin');
}

function resolveRequestHost(request) {
  const forwarded = request.headers.get('x-forwarded-host');
  const raw = forwarded || request.headers.get('host') || request.nextUrl.host || '';
  return raw.split(',')[0].trim().split(':')[0].toLowerCase();
}

function hasAdminSurfaceRole(session) {
  return session.roles.includes('founder') || session.roles.includes('admin');
}

function unauthorizedApi(message, status) {
  return NextResponse.json({ error: message }, { status });
}

function redirectToLogin(request, reason) {
  const url = request.nextUrl.clone();
  const next = `${request.nextUrl.pathname}${request.nextUrl.search}`;
  url.pathname = '/admin/login';
  url.search = '';
  url.searchParams.set('next', next);
  if (reason) {
    url.searchParams.set('reason', reason);
  }
  return NextResponse.redirect(url);
}

function redirectToAdminHost(request) {
  const url = request.nextUrl.clone();
  url.hostname = canonicalAdminHost;
  url.port = '';
  return NextResponse.redirect(url);
}

function redirectAdminRoot(request) {
  const url = request.nextUrl.clone();
  url.pathname = '/admin/login';
  url.search = '';
  return NextResponse.redirect(url);
}

export async function middleware(request) {
  const pathname = request.nextUrl.pathname;
  const hostname = resolveRequestHost(request);
  if (canonicalPublicHosts.has(hostname) && isAdminSurfacePath(pathname)) {
    return redirectToAdminHost(request);
  }
  if (hostname === canonicalAdminHost && pathname === '/') {
    return redirectAdminRoot(request);
  }
  if (isPublicAdminAuthPath(pathname)) {
    return NextResponse.next();
  }
  const protectedApi = isProtectedApiPath(pathname);
  const protectedPage = isProtectedPagePath(pathname);
  if (!protectedApi && !protectedPage) {
    return NextResponse.next();
  }

  const token = extractSessionToken(request);
  if (!token) {
    return protectedApi ? unauthorizedApi('authentication required', 401) : redirectToLogin(request, 'missing');
  }

  const secret = resolveSessionSecret();
  if (!secret) {
    return protectedApi ? unauthorizedApi('server auth misconfiguration', 500) : redirectToLogin(request, 'misconfigured');
  }

  const session = await verifySessionToken(token, secret);
  if (!session) {
    return protectedApi ? unauthorizedApi('invalid session', 401) : redirectToLogin(request, 'invalid');
  }

  if (!hasAdminSurfaceRole(session)) {
    return protectedApi ? unauthorizedApi('forbidden', 403) : redirectToLogin(request, 'forbidden');
  }

  const requestHeaders = new Headers(request.headers);
  requestHeaders.set('x-auth-user-id', session.userId);
  requestHeaders.set('x-auth-roles', session.roles.join(','));

  return NextResponse.next({
    request: {
      headers: requestHeaders,
    },
  });
}

export const config = {
  matcher: ['/', '/api/admin/:path*', '/admin/:path*'],
};
