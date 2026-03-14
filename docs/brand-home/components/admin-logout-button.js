'use client';

import { useRouter } from 'next/navigation';
import { useState } from 'react';

export default function AdminLogoutButton() {
  const router = useRouter();
  const [pending, setPending] = useState(false);

  async function logout() {
    setPending(true);
    try {
      await fetch('/api/admin/session/logout', { method: 'POST' });
      router.push('/admin/login');
      router.refresh();
    } finally {
      setPending(false);
    }
  }

  return (
    <button type="button" className="button-secondary button-secondary--compact" onClick={logout} disabled={pending}>
      {pending ? '정리 중…' : '로그아웃'}
    </button>
  );
}
