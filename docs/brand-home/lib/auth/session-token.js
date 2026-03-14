const encoder = new TextEncoder();

const devSessionSecret = 'layer-os-local-dev-session-secret';

function toBase64Url(input) {
  return Buffer.from(input)
    .toString('base64')
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/g, '');
}

function fromBase64Url(input) {
  const normalized = input.replace(/-/g, '+').replace(/_/g, '/');
  const padding = normalized.length % 4 === 0 ? '' : '='.repeat(4 - (normalized.length % 4));
  return Buffer.from(normalized + padding, 'base64').toString('utf8');
}

async function signPayload(payloadSegment, secret) {
  const key = await crypto.subtle.importKey(
    'raw',
    encoder.encode(secret),
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign']
  );
  const signature = await crypto.subtle.sign('HMAC', key, encoder.encode(payloadSegment));
  return toBase64Url(Buffer.from(signature));
}

export function resolveSessionSecret() {
  const explicit = process.env.SESSION_HMAC_SECRET;
  if (explicit && explicit.trim()) {
    return explicit.trim();
  }
  if (process.env.NODE_ENV !== 'production') {
    return devSessionSecret;
  }
  return null;
}

export function localAdminBootstrapEnabled() {
  if (process.env.NODE_ENV === 'production') {
    return false;
  }
  const override = process.env.LAYER_OS_LOCAL_ADMIN_BOOTSTRAP;
  if (!override || !override.trim()) {
    return true;
  }
  return !['0', 'false', 'off', 'no'].includes(override.trim().toLowerCase());
}

export function extractSessionToken(request) {
  const authHeader = request.headers.get('authorization');
  if (authHeader && authHeader.startsWith('Bearer ')) {
    return authHeader.slice('Bearer '.length).trim();
  }
  return request.cookies.get('wwh_session')?.value || null;
}

export async function createSessionToken(session, secret) {
  if (!session || typeof session !== 'object') {
    throw new Error('session payload is required');
  }
  const payload = {
    userId: typeof session.userId === 'string' && session.userId.trim() ? session.userId.trim() : 'founder',
    roles: Array.isArray(session.roles) ? session.roles.map((item) => String(item).trim()).filter(Boolean) : [],
    exp: session.exp ?? null,
  };
  if (payload.roles.length === 0) {
    throw new Error('at least one role is required');
  }
  const payloadSegment = toBase64Url(JSON.stringify(payload));
  const signatureSegment = await signPayload(payloadSegment, secret);
  return `${payloadSegment}.${signatureSegment}`;
}

export async function createFounderAdminSession(secret, ttlSeconds = 12 * 60 * 60) {
  const exp = Math.floor(Date.now() / 1000) + ttlSeconds;
  const session = {
    userId: 'founder',
    roles: ['founder', 'admin'],
    exp,
  };
  return {
    token: await createSessionToken(session, secret),
    session,
  };
}

export async function verifySessionToken(token, secret) {
  if (!token || !secret) return null;
  const segments = token.split('.');
  if (segments.length !== 2) return null;
  const [payloadSegment, signatureSegment] = segments;
  const expected = await signPayload(payloadSegment, secret);
  if (expected !== signatureSegment) return null;
  try {
    const payload = JSON.parse(fromBase64Url(payloadSegment));
    if (!payload || typeof payload !== 'object') return null;
    if (typeof payload.userId !== 'string' || !payload.userId.trim()) return null;
    if (!Array.isArray(payload.roles) || payload.roles.some((role) => typeof role !== 'string')) return null;
    if (payload.exp && Date.now() >= payload.exp * 1000) return null;
    return {
      userId: payload.userId,
      roles: payload.roles,
      exp: payload.exp ?? null,
    };
  } catch {
    return null;
  }
}
