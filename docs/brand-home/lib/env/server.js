import 'server-only';

import { execFileSync } from 'node:child_process';

const secretCache = new Map();

function required(name, fallback = null) {
  const value = process.env[name];
  if (!value || !value.trim()) {
    if (fallback !== null) {
      return fallback;
    }
    throw new Error(`Missing required server env: ${name}`);
  }
  return value.trim();
}

function optional(name, fallback = null) {
  const value = process.env[name];
  if (!value || !value.trim()) {
    return fallback;
  }
  return value.trim();
}

function readKeychainSecret(name) {
  if (process.platform !== 'darwin') {
    return null;
  }
  if (secretCache.has(name)) {
    return secretCache.get(name);
  }
  try {
    const value = execFileSync('security', ['find-generic-password', '-a', 'layer-os', '-s', name, '-w'], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    }).trim();
    const normalized = value || null;
    secretCache.set(name, normalized);
    return normalized;
  } catch {
    secretCache.set(name, null);
    return null;
  }
}

function optionalSecret(name) {
  const explicit = optional(name);
  if (explicit) {
    return { value: explicit, source: 'env' };
  }
  const keychain = readKeychainSecret(name);
  if (keychain) {
    return { value: keychain, source: 'keychain' };
  }
  return { value: null, source: 'missing' };
}

function localBootstrapDefault() {
  if (process.env.NODE_ENV === 'production') {
    return false;
  }
  const override = process.env.LAYER_OS_LOCAL_ADMIN_BOOTSTRAP;
  if (!override || !override.trim()) {
    return true;
  }
  return !['0', 'false', 'off', 'no'].includes(override.trim().toLowerCase());
}

export function getServerSecrets() {
  const layerOsWriteToken = optionalSecret('LAYER_OS_WRITE_TOKEN');
  return {
    sessionHmacSecret: required('SESSION_HMAC_SECRET', process.env.NODE_ENV !== 'production' ? 'layer-os-local-dev-session-secret' : null),
    layerOsWriteToken: layerOsWriteToken.value || 'woohwahae2024',
    layerOsWriteTokenSource: layerOsWriteToken.source !== 'missing' ? layerOsWriteToken.source : 'fallback',
    adminPassword: optional('LAYER_OS_ADMIN_PASSWORD', 'admin'),
    adminPasswordSha256: optional('LAYER_OS_ADMIN_PASSWORD_SHA256'),
  };
}

export function getRuntimeConfig() {
  const passwordAdminLoginEnabled = Boolean(optional('LAYER_OS_ADMIN_PASSWORD', 'admin') || optional('LAYER_OS_ADMIN_PASSWORD_SHA256'));
  return {
    layerOsBaseUrl: optional('LAYER_OS_BASE_URL', 'http://127.0.0.1:17808'),
    localAdminBootstrapEnabled: localBootstrapDefault(),
    passwordAdminLoginEnabled,
  };
}
