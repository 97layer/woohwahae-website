import crypto from 'node:crypto';

function toBase64Url(input) {
  return Buffer.from(input)
    .toString('base64')
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/g, '');
}

function usage() {
  console.error('usage: SESSION_HMAC_SECRET=... node scripts/dev-session-token.mjs <user-id> <roles-comma-separated> [ttl-seconds]');
  process.exit(1);
}

const [, , userId, rolesRaw, ttlRaw] = process.argv;
if (!userId || !rolesRaw) usage();

const secret = process.env.SESSION_HMAC_SECRET;
if (!secret || !secret.trim()) {
  console.error('missing SESSION_HMAC_SECRET');
  process.exit(1);
}

const roles = rolesRaw.split(',').map((item) => item.trim()).filter(Boolean);
if (roles.length === 0) {
  console.error('at least one role is required');
  process.exit(1);
}

const ttlSeconds = Number.parseInt(ttlRaw || '3600', 10);
const payload = {
  userId,
  roles,
  exp: Math.floor(Date.now() / 1000) + (Number.isFinite(ttlSeconds) ? ttlSeconds : 3600),
};
const payloadSegment = toBase64Url(JSON.stringify(payload));
const signature = crypto.createHmac('sha256', secret).update(payloadSegment).digest();
const token = `${payloadSegment}.${toBase64Url(signature)}`;

console.log(token);
