import 'server-only';

import { fetchLayerOs } from './layer-os';

function asArray(value) {
  return Array.isArray(value) ? value : [];
}

function asText(value) {
  return typeof value === 'string' ? value.trim() : '';
}

function normalizeAdapterName(adapter, enabled) {
  if (typeof adapter === 'string' && adapter.trim()) {
    return adapter.trim();
  }
  return enabled ? 'configured' : 'noop';
}

function normalizeStatus(payload, enabled) {
  const status = payload?.status && typeof payload.status === 'object' ? payload.status : {};
  const routes = asArray(status.routes || status.Routes).map((item) => ({
    routeId: item?.route_id || item?.routeId || '',
    label: item?.label || '',
    chatConfigured: Boolean(item?.chat_configured ?? item?.chatConfigured),
    delivery: item?.delivery || 'disabled',
    notes: asArray(item?.notes).slice(0, 3),
  }));
  return {
    sendAdapter: status.send_adapter || status.sendAdapter || normalizeAdapterName(payload?.adapter, enabled),
    sendConfigured: Boolean(status.send_configured ?? status.sendConfigured ?? enabled),
    pollingConfigured: Boolean(status.polling_configured ?? status.pollingConfigured ?? enabled),
    chatConfigured: Boolean(status.chat_configured ?? status.chatConfigured ?? enabled),
    geminiConfigured: Boolean(status.gemini_configured ?? status.geminiConfigured),
    inboundMode: status.inbound_mode || status.inboundMode || 'off',
    founderDelivery: status.founder_delivery || status.founderDelivery || (enabled ? 'ready' : 'disabled'),
    routes,
    notes: asArray(status.notes).slice(0, 4),
  };
}

function buildTelegramAttention(status, packet) {
  const primaryAction = asText(packet.primary_action || packet.primaryAction);
  const primaryRef = asText(packet.primary_ref || packet.primaryRef);
  const recommendedMode = asText(packet.recommended_mode || packet.recommendedMode || 'observe');
  const founderNotice = asText(packet.founder_notice || packet.founderNotice);
  const currentFocus = asText(packet.current_focus || packet.currentFocus || packet.current_goal || packet.currentGoal);
  const nextSteps = asArray(packet.next_steps || packet.nextSteps).map((item) => asText(item)).filter(Boolean).slice(0, 3);
  const openRisks = asArray(packet.open_risks || packet.openRisks).map((item) => asText(item)).filter(Boolean).slice(0, 3);
  const reviewOpenCount = Number(packet.review_open_count ?? packet.reviewOpenCount ?? 0);
  const reviewTopOpen = asArray(packet.review_top_open || packet.reviewTopOpen);

  if (primaryAction) {
    return {
      mode: recommendedMode || 'review_packet',
      summary: primaryAction,
      detail: founderNotice || currentFocus || 'telegram founder packet이 가리키는 다음 액션입니다.',
      ref: primaryRef,
      nextSteps,
      openRisks,
    };
  }

  if (!status.sendConfigured) {
    return {
      mode: 'restore_delivery',
      summary: 'Founder Telegram 전달 경로를 먼저 복구하세요.',
      detail: founderNotice || 'bot token / founder chat id가 빠져 있으면 founder inbox packet이 실제 전달되지 않습니다.',
      ref: '',
      nextSteps,
      openRisks,
    };
  }

  if (reviewOpenCount > 0) {
    const firstAgenda = reviewTopOpen[0];
    const agendaText = asText(firstAgenda?.text || firstAgenda?.summary || firstAgenda?.title);
    return {
      mode: 'review_room',
      summary: `열린 검토 안건 ${reviewOpenCount}건을 먼저 확인하세요.`,
      detail: agendaText || founderNotice || 'review room에서 막혀 있는 안건이 founder 결정을 기다리고 있습니다.',
      ref: primaryRef,
      nextSteps,
      openRisks,
    };
  }

  if (nextSteps.length > 0) {
    return {
      mode: recommendedMode || 'continue_loop',
      summary: nextSteps[0],
      detail: founderNotice || currentFocus || 'packet이 남긴 다음 단계입니다.',
      ref: primaryRef,
      nextSteps,
      openRisks,
    };
  }

  return {
    mode: recommendedMode || 'observe',
    summary: founderNotice || currentFocus || 'Founder inbox packet은 현재 안정 상태입니다.',
    detail: '새 packet이 오면 여기서 다음 액션을 먼저 요약해 보여줍니다.',
    ref: primaryRef,
    nextSteps,
    openRisks,
  };
}

function normalizeTelegramAdminView(payload) {
  const enabled = Boolean(payload?.enabled);
  const status = normalizeStatus(payload, enabled);
  const packet = payload?.telegram && typeof payload.telegram === 'object' ? payload.telegram : {};
  return {
    generatedAt: new Date().toISOString(),
    enabled,
    adapter: normalizeAdapterName(payload?.adapter, enabled),
    status,
    packet: {
      headline: packet.headline || '',
      bodyLines: asArray(packet.body_lines || packet.bodyLines).slice(0, 6),
      primaryAction: packet.primary_action || packet.primaryAction || '',
      primaryRef: packet.primary_ref || packet.primaryRef || '',
      currentFocus: packet.current_focus || packet.currentFocus || '',
      currentGoal: packet.current_goal || packet.currentGoal || '',
      nextSteps: asArray(packet.next_steps || packet.nextSteps).slice(0, 4),
      openRisks: asArray(packet.open_risks || packet.openRisks).slice(0, 4),
      reviewOpenCount: packet.review_open_count ?? packet.reviewOpenCount ?? 0,
      reviewTopOpen: asArray(packet.review_top_open || packet.reviewTopOpen).slice(0, 4),
      founderNotice: packet.founder_notice || packet.founderNotice || '',
      recommendedMode: packet.recommended_mode || packet.recommendedMode || '',
    },
    attention: buildTelegramAttention(status, packet),
  };
}

export async function getAdminTelegramView() {
  const payload = await fetchLayerOs('/api/layer-os/telegram');
  return normalizeTelegramAdminView(payload);
}

export async function sendAdminTelegram() {
  const sendResult = await fetchLayerOs('/api/layer-os/telegram', {
    method: 'POST',
    json: {},
    requireWriteToken: true,
  });
  return {
    ...sendResult,
    telegram: await getAdminTelegramView(),
  };
}
