'use client';

import { useEffect, useMemo, useState } from 'react';
import {
  normalizeDeliveryState,
  normalizeInboundState,
  normalizeRouteStates,
  normalizeSendState,
  normalizeTelegramSurface,
} from '../lib/admin-telegram-state';

function makeToneClass(tone) {
  if (tone === 'good') return 'status-pill status-pill--good';
  if (tone === 'alert') return 'status-pill status-pill--alert';
  return 'status-pill status-pill--muted';
}

async function readTelegramStatus() {
  const response = await fetch('/api/admin/runtime/telegram', {
    method: 'GET',
    cache: 'no-store',
  });
  const payload = await response.json();
  if (!response.ok) {
    throw new Error(payload?.error || 'telegram preview failed');
  }
  return payload;
}

async function sendTelegram() {
  const response = await fetch('/api/admin/runtime/telegram', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({}),
  });
  const payload = await response.json();
  if (!response.ok) {
    throw new Error(payload?.error || 'telegram send failed');
  }
  return payload;
}

function renderList(items, emptyText) {
  if (!items?.length) {
    return <p className="section-body">{emptyText}</p>;
  }
  return (
    <ul className="inline-list">
      {items.map((item, index) => (
        <li key={`${item}-${index}`}>
          <span>{item}</span>
        </li>
      ))}
    </ul>
  );
}

export default function AdminTelegramActions({ canWrite }) {
  const [pending, setPending] = useState(false);
  const [statusPending, setStatusPending] = useState(false);
  const [message, setMessage] = useState('');
  const [runtime, setRuntime] = useState(null);

  const adapter = useMemo(() => normalizeTelegramSurface(runtime), [runtime]);
  const sendState = useMemo(() => normalizeSendState(canWrite, runtime, pending), [canWrite, pending, runtime]);
  const inboundState = useMemo(() => normalizeInboundState(runtime), [runtime]);
  const deliveryState = useMemo(() => normalizeDeliveryState(runtime), [runtime]);
  const routeStates = useMemo(() => normalizeRouteStates(runtime), [runtime]);
  const packet = runtime?.packet || {};
  const attention = runtime?.attention || null;

  async function refreshStatus() {
    setStatusPending(true);
    try {
      const payload = await readTelegramStatus();
      setRuntime(payload);
      setMessage('');
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'telegram preview failed');
    } finally {
      setStatusPending(false);
    }
  }

  async function submitTelegram() {
    setPending(true);
    setMessage('');
    try {
      const payload = await sendTelegram();
      setRuntime(payload.telegram || null);
      setMessage(payload?.telegram?.attention?.summary || 'telegram packet sent');
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'telegram send failed');
    } finally {
      setPending(false);
    }
  }

  useEffect(() => {
    refreshStatus();
  }, []);

  return (
    <section className="admin-card admin-card--wide">
      <div className="quickwork-card__head">
        <div>
          <p className="section-eyebrow">Telegram</p>
          <h2 className="section-title">Founder inbox packet</h2>
          <p className="section-body">daemon preview를 그대로 읽고, founder 권한일 때만 같은 경로로 전송합니다.</p>
        </div>
        <div className="inline-action-block__buttons">
          <span className={makeToneClass(adapter.tone)}>{adapter.label}</span>
          <span className={makeToneClass(inboundState.tone)}>{inboundState.label}</span>
          <span className={makeToneClass(deliveryState.tone)}>{deliveryState.label}</span>
          <span className={makeToneClass(sendState.tone)}>{sendState.label}</span>
          <button type="button" className="button-secondary button-secondary--compact" onClick={refreshStatus} disabled={statusPending || pending}>
            {statusPending ? '새로 읽는 중' : '새로고침'}
          </button>
          <button type="button" className="button-primary button-primary--compact" onClick={submitTelegram} disabled={!canWrite || !runtime?.status?.sendConfigured || pending}>
            {pending ? '보내는 중' : '지금 보내기'}
          </button>
        </div>
      </div>

      <div className="proof-metrics proof-metrics--admin">
        <article className="metric-card metric-card--compact">
          <span className="metric-card__label">Headline</span>
          <strong className="metric-card__value metric-card__value--compact">{packet.headline || 'n/a'}</strong>
          <span className="metric-card__note">{adapter.detail}</span>
        </article>
        <article className="metric-card metric-card--compact">
          <span className="metric-card__label">Inbound mode</span>
          <strong className="metric-card__value metric-card__value--compact">{inboundState.label}</strong>
          <span className="metric-card__note">{inboundState.detail}</span>
        </article>
        <article className="metric-card metric-card--compact">
          <span className="metric-card__label">Delivery</span>
          <strong className="metric-card__value metric-card__value--compact">{deliveryState.label}</strong>
          <span className="metric-card__note">{deliveryState.detail}</span>
        </article>
        <article className="metric-card metric-card--compact">
          <span className="metric-card__label">Review pressure</span>
          <strong className="metric-card__value">{packet.reviewOpenCount || 0}</strong>
          <span className="metric-card__note">open agenda items</span>
        </article>
        <article className="metric-card metric-card--compact">
          <span className="metric-card__label">Recommended mode</span>
          <strong className="metric-card__value metric-card__value--compact">{packet.recommendedMode || 'observe'}</strong>
          <span className="metric-card__note">{sendState.detail}</span>
        </article>
        <article className="metric-card metric-card--compact">
          <span className="metric-card__label">Primary ref</span>
          <strong className="metric-card__value metric-card__value--compact">{packet.primaryRef || 'n/a'}</strong>
          <span className="metric-card__note">{packet.primaryAction || 'founder action not set'}</span>
        </article>
      </div>

      <div className="signal-grid">
        <article className="signal-card">
          <span className="signal-card__source">다음 액션</span>
          <h3 className="signal-card__title">{attention?.summary || 'Founder inbox packet을 먼저 읽어야 합니다.'}</h3>
          <p className="section-body">
            {attention?.detail || 'Telegram preview는 founder에게 지금 가장 먼저 볼 액션을 한 줄로 올려주는 운영면입니다.'}
          </p>
          {attention?.ref ? <p className="section-body">ref · {attention.ref}</p> : null}
          {renderList(attention?.nextSteps, '아직 packet이 남긴 다음 단계는 없습니다.')}
        </article>
        <article className="signal-card">
          <span className="signal-card__source">Preview body</span>
          <h3 className="signal-card__title">{packet.currentFocus || packet.currentGoal || 'Current runtime focus'}</h3>
          {renderList(packet.bodyLines, 'No telegram body lines yet. Founder summary and review pressure will populate this preview.')}
        </article>
        <article className="signal-card">
          <span className="signal-card__source">Founder inbox</span>
          <p className="section-body">{packet.founderNotice || 'No environment advisory in the packet.'}</p>
          {renderList(runtime?.status?.notes, 'No Telegram readiness notes yet.')}
          {renderList(packet.nextSteps, 'No next steps attached to the packet.')}
          {renderList(packet.openRisks, 'No open risks attached to the packet.')}
        </article>
        <article className="signal-card">
          <span className="signal-card__source">One bot, many chats</span>
          <h3 className="signal-card__title">Founder, ops, and brand routes</h3>
          {routeStates.map((route) => (
            <div key={route.routeId}>
              <div className="inline-action-block__buttons">
                <span className={makeToneClass('muted')}>{route.title}</span>
                <span className={makeToneClass(route.tone)}>{route.label}</span>
              </div>
              <p className="section-body">{route.detail}</p>
              {route.notes?.length ? renderList(route.notes, '') : null}
            </div>
          ))}
        </article>
      </div>

      {message ? <p className="inline-action-block__message">{message}</p> : null}
    </section>
  );
}
