import assert from 'node:assert/strict';
import test from 'node:test';

import {
  normalizeDeliveryState,
  normalizeInboundState,
  normalizeRouteStates,
  normalizeSendState,
  normalizeTelegramSurface,
} from './admin-telegram-state.js';

test('telegram surface reports polling-only instead of noop when inbound is alive without founder delivery', () => {
  const runtime = {
    adapter: 'noop',
    status: {
      sendConfigured: false,
      pollingConfigured: true,
      chatConfigured: false,
      inboundMode: 'command_only',
      founderDelivery: 'chat_missing',
    },
  };

  assert.deepEqual(normalizeTelegramSurface(runtime), {
    tone: 'muted',
    label: 'polling only',
    detail: 'inbound polling works, but the founder route still needs a dedicated chat id',
  });
  assert.equal(normalizeInboundState(runtime).label, 'command only');
  assert.equal(normalizeDeliveryState(runtime).label, 'chat missing');
  assert.equal(normalizeSendState(true, runtime, false).label, 'not ready');
});

test('telegram surface reports full readiness when founder delivery and assistant replies are both available', () => {
  const runtime = {
    adapter: 'telegram_bot',
    status: {
      sendConfigured: true,
      pollingConfigured: true,
      chatConfigured: true,
      inboundMode: 'assistant',
      founderDelivery: 'ready',
    },
  };

  assert.deepEqual(normalizeTelegramSurface(runtime), {
    tone: 'good',
    label: 'telegram_bot',
    detail: 'founder delivery and assistant reply path are ready',
  });
  assert.equal(normalizeInboundState(runtime).label, 'assistant');
  assert.equal(normalizeDeliveryState(runtime).label, 'delivery ready');
  assert.equal(normalizeSendState(true, runtime, false).label, 'ready');
});

test('telegram surface reports split conflict when founder room and founder DM share one chat id', () => {
  const runtime = {
    adapter: 'telegram_bot',
    status: {
      sendConfigured: false,
      pollingConfigured: true,
      chatConfigured: false,
      inboundMode: 'assistant',
      founderDelivery: 'split_required',
    },
  };

  assert.deepEqual(normalizeTelegramSurface(runtime), {
    tone: 'alert',
    label: 'split required',
    detail: 'founder room and founder DM share one chat id, so direct alerts stay paused',
  });
  assert.equal(normalizeDeliveryState(runtime).label, 'split required');
  assert.equal(normalizeSendState(true, runtime, false).label, 'split required');
});

test('telegram route states explain founder, ops fallback, and optional brand chat', () => {
  const runtime = {
    adapter: 'telegram_bot',
    status: {
      routes: [
        { routeId: 'founder', label: 'Founder', delivery: 'ready', notes: [] },
        { routeId: 'ops', label: 'Ops', delivery: 'chat_missing', notes: [] },
        { routeId: 'brand', label: 'Brand', delivery: 'disabled', notes: [] },
      ],
    },
  };

  assert.deepEqual(normalizeRouteStates(runtime), [
    {
      routeId: 'founder',
      title: 'Founder',
      tone: 'good',
      label: 'ready',
      detail: 'founder packets and direct alerts land here',
      notes: [],
    },
    {
      routeId: 'ops',
      title: 'Ops',
      tone: 'alert',
      label: 'chat missing',
      detail: 'set the dedicated chat id before treating this route as live',
      notes: [],
    },
    {
      routeId: 'brand',
      title: 'Brand',
      tone: 'muted',
      label: 'disabled',
      detail: 'brand route is optional until content review starts living in Telegram',
      notes: [],
    },
  ]);
});
