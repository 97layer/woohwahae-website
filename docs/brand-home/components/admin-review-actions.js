'use client';

import { useRouter } from 'next/navigation';
import { useState } from 'react';

export default function AdminReviewActions({ itemText, canWrite }) {
  const router = useRouter();
  const [reason, setReason] = useState('founder reviewed in web admin');
  const [pending, setPending] = useState(false);
  const [message, setMessage] = useState('');

  async function transition(action) {
    setPending(true);
    setMessage('');
    try {
      const response = await fetch(`/api/admin/runtime/review-room/${action}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ item: itemText, reason }),
      });
      const payload = await response.json();
      if (!response.ok) {
        throw new Error(payload?.error || `${action} failed`);
      }
      setMessage(`${action} 완료`);
      router.refresh();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'transition failed');
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="inline-action-block">
      <input
        type="text"
        value={reason}
        onChange={(event) => setReason(event.target.value)}
        maxLength={240}
        disabled={!canWrite || pending}
        placeholder="reason"
      />
      <div className="inline-action-block__buttons">
        <button type="button" className="button-secondary button-secondary--compact" disabled={!canWrite || pending} onClick={() => transition('accept')}>accept</button>
        <button type="button" className="button-secondary button-secondary--compact" disabled={!canWrite || pending} onClick={() => transition('defer')}>defer</button>
        <button type="button" className="button-primary button-primary--compact" disabled={!canWrite || pending} onClick={() => transition('resolve')}>resolve</button>
      </div>
      {message ? <p className="inline-action-block__message">{message}</p> : null}
    </div>
  );
}
