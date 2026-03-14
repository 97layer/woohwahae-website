import 'server-only';

import { getRuntimeConfig } from '../env/server';
import { getActiveLayerOsWriteToken } from './layer-os-token-context';

function buildUrl(path) {
  const { layerOsBaseUrl } = getRuntimeConfig();
  return new URL(path, layerOsBaseUrl).toString();
}

async function decodeJson(response) {
  const text = await response.text();
  if (!text) {
    return null;
  }
  try {
    return JSON.parse(text);
  } catch {
    return { error: text };
  }
}

function responseError(path, response, payload) {
  if (payload && typeof payload.error === 'string' && payload.error.trim()) {
    return new Error(`${response.status} ${payload.error.trim()}`);
  }
  return new Error(`${response.status} Layer OS request failed for ${path}`);
}

function createRequestSignal(path, timeoutMs, upstreamSignal) {
  if ((!Number.isFinite(timeoutMs) || timeoutMs <= 0) && !upstreamSignal) {
    return {
      signal: undefined,
      cleanup() {},
    };
  }

  const controller = new AbortController();
  let timeoutId;
  let upstreamListener;

  const abortWithReason = (reason) => {
    if (!controller.signal.aborted) {
      controller.abort(reason);
    }
  };

  if (upstreamSignal) {
    if (upstreamSignal.aborted) {
      abortWithReason(upstreamSignal.reason);
    } else {
      upstreamListener = () => abortWithReason(upstreamSignal.reason);
      upstreamSignal.addEventListener('abort', upstreamListener, { once: true });
    }
  }

  if (Number.isFinite(timeoutMs) && timeoutMs > 0) {
    timeoutId = setTimeout(() => {
      abortWithReason(new Error(`Layer OS request timed out for ${path} after ${timeoutMs}ms`));
    }, timeoutMs);
  }

  return {
    signal: controller.signal,
    cleanup() {
      if (timeoutId) {
        clearTimeout(timeoutId);
      }
      if (upstreamSignal && upstreamListener) {
        upstreamSignal.removeEventListener('abort', upstreamListener);
      }
    },
  };
}

export async function fetchLayerOs(path, options = {}) {
  const {
    method = 'GET',
    json,
    requireWriteToken = false,
    writeToken = '',
    headers: extraHeaders,
    cache = 'no-store',
    signal: upstreamSignal,
    timeoutMs = 0,
  } = options;
  const headers = new Headers(extraHeaders || {});
  headers.set('Accept', 'application/json');
  let body;
  if (json !== undefined) {
    headers.set('Content-Type', 'application/json');
    body = JSON.stringify(json);
  }
  if (requireWriteToken) {
    const scoped = getActiveLayerOsWriteToken();
    const candidate = typeof writeToken === 'string' && writeToken.trim() ? writeToken.trim() : scoped.token;
    if (!candidate) {
      throw new Error('missing server write token');
    }
    headers.set('Authorization', `Bearer ${candidate}`);
  }
  const request = createRequestSignal(path, timeoutMs, upstreamSignal);
  try {
    const response = await fetch(buildUrl(path), {
      method,
      headers,
      body,
      cache,
      signal: request.signal,
    });
    const payload = await decodeJson(response);
    if (!response.ok) {
      throw responseError(path, response, payload);
    }
    return payload;
  } catch (error) {
    if (request.signal?.aborted && Number.isFinite(timeoutMs) && timeoutMs > 0 && error instanceof Error && error.name === 'AbortError') {
      throw new Error(`Layer OS request timed out for ${path} after ${timeoutMs}ms`);
    }
    throw error;
  } finally {
    request.cleanup();
  }
}

export async function safeLayerOs(path, options = {}) {
  try {
    const payload = await fetchLayerOs(path, options);
    return { ok: true, payload };
  } catch (error) {
    return {
      ok: false,
      error: error instanceof Error ? error.message : 'runtime request failed',
    };
  }
}
