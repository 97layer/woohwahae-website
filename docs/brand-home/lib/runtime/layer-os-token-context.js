import 'server-only';

import { AsyncLocalStorage } from 'node:async_hooks';

import { getServerSecrets } from '../env/server';

const writeTokenStorage = new AsyncLocalStorage();

function normalizeToken(value) {
  return typeof value === 'string' && value.trim() ? value.trim() : '';
}

function normalizeSource(value) {
  return typeof value === 'string' && value.trim() ? value.trim() : '';
}

export function getActiveLayerOsWriteToken() {
  const scoped = writeTokenStorage.getStore();
  if (scoped?.token) {
    return {
      token: scoped.token,
      source: scoped.source || 'cookie',
    };
  }
  const { layerOsWriteToken, layerOsWriteTokenSource } = getServerSecrets();
  return {
    token: normalizeToken(layerOsWriteToken),
    source: layerOsWriteToken ? normalizeSource(layerOsWriteTokenSource) || 'env' : 'missing',
  };
}

export async function withLayerOsWriteToken(context, work) {
  const token = normalizeToken(context?.token ?? context);
  if (!token) {
    return work();
  }
  const source = normalizeSource(context?.source) || 'cookie';
  return writeTokenStorage.run({ token, source }, work);
}
