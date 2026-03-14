import 'server-only';

import { withLayerOsWriteToken } from '../runtime/layer-os-token-context';
import { getServerSecrets } from '../env/server';

export const layerOsWriteTokenCookieName = 'wwh_layer_os_write_token';

function normalizeToken(value) {
  return typeof value === 'string' && value.trim() ? value.trim() : '';
}

function cookieEntry(store, name) {
  if (!store || typeof store.get !== 'function') {
    return null;
  }
  const value = store.get(name);
  if (!value) return null;
  if (typeof value === 'string') {
    return value;
  }
  return typeof value.value === 'string' ? value.value : null;
}

export function readLayerOsWriteTokenFromCookieStore(store) {
  return normalizeToken(cookieEntry(store, layerOsWriteTokenCookieName));
}

export function resolveLayerOsWriteToken(request = null) {
  const cookieToken = readLayerOsWriteTokenFromCookieStore(request?.cookies);
  if (cookieToken) {
    return {
      token: cookieToken,
      source: 'cookie',
      configured: true,
    };
  }
  const { layerOsWriteToken, layerOsWriteTokenSource } = getServerSecrets();
  return {
    token: normalizeToken(layerOsWriteToken),
    source: layerOsWriteToken ? layerOsWriteTokenSource || 'env' : 'missing',
    configured: Boolean(layerOsWriteToken),
  };
}

export async function withRequestLayerOsWriteToken(request, work) {
  return withLayerOsWriteToken(resolveLayerOsWriteToken(request), work);
}

function cookieOptions() {
  return {
    httpOnly: true,
    sameSite: 'lax',
    secure: process.env.NODE_ENV === 'production',
    path: '/api/admin/runtime',
  };
}

export function attachLayerOsWriteTokenCookie(response, token) {
  response.cookies.set(layerOsWriteTokenCookieName, normalizeToken(token), {
    ...cookieOptions(),
    maxAge: 12 * 60 * 60,
  });
  return response;
}

export function clearLayerOsWriteTokenCookie(response) {
  response.cookies.set(layerOsWriteTokenCookieName, '', {
    ...cookieOptions(),
    maxAge: 0,
    expires: new Date(0),
  });
  return response;
}
