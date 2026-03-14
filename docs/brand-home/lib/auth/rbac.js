import { extractSessionToken, verifySessionToken } from './session-token';
import { getServerSecrets } from '../env/server';

export function hasAnyRole(session, allowedRoles) {
  if (!session) return false;
  return allowedRoles.some((role) => session.roles.includes(role));
}

export async function authenticateRequest(request) {
  const token = extractSessionToken(request);
  if (!token) {
    return { ok: false, status: 401, error: 'authentication required' };
  }
  const { sessionHmacSecret } = getServerSecrets();
  const session = await verifySessionToken(token, sessionHmacSecret);
  if (!session) {
    return { ok: false, status: 401, error: 'invalid session' };
  }
  return { ok: true, session };
}

export async function authorizeRequest(request, allowedRoles) {
  const auth = await authenticateRequest(request);
  if (!auth.ok) return auth;
  if (!hasAnyRole(auth.session, allowedRoles)) {
    return { ok: false, status: 403, error: 'forbidden' };
  }
  return auth;
}
