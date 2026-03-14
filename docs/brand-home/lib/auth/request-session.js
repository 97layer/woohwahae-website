import 'server-only';

import { headers } from 'next/headers';

export async function getRequestSessionMeta() {
  const requestHeaders = await headers();
  const userId = requestHeaders.get('x-auth-user-id') || 'founder';
  const roles = (requestHeaders.get('x-auth-roles') || '')
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean);

  return {
    userId,
    roles,
    canWrite: roles.includes('founder'),
  };
}
