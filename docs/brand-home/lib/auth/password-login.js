import 'server-only';

import { createHash, timingSafeEqual } from 'node:crypto';

import { getServerSecrets } from '../env/server';

function normalize(value) {
  return String(value || '').trim();
}

function safeEqual(left, right) {
  const a = Buffer.from(String(left), 'utf8');
  const b = Buffer.from(String(right), 'utf8');
  if (a.length !== b.length) {
    return false;
  }
  return timingSafeEqual(a, b);
}

export function passwordLoginEnabled() {
  const { adminPassword, adminPasswordSha256 } = getServerSecrets();
  return Boolean(normalize(adminPassword) || normalize(adminPasswordSha256));
}

export function verifyAdminPassword(password) {
  const candidate = normalize(password);
  if (!candidate) {
    return false;
  }

  const { adminPassword, adminPasswordSha256 } = getServerSecrets();
  const plain = normalize(adminPassword);
  if (plain && safeEqual(candidate, plain)) {
    return true;
  }

  const sha256 = normalize(adminPasswordSha256).toLowerCase();
  if (!sha256) {
    return false;
  }

  const digest = createHash('sha256').update(candidate).digest('hex');
  return safeEqual(digest, sha256);
}
