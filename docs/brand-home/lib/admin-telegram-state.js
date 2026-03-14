function readStatus(runtime) {
  return runtime?.status && typeof runtime.status === 'object' ? runtime.status : {};
}

function sendAdapterName(runtime) {
  const status = readStatus(runtime);
  if (typeof status.sendAdapter === 'string' && status.sendAdapter.trim()) {
    return status.sendAdapter.trim();
  }
  if (typeof runtime?.adapter === 'string' && runtime.adapter.trim()) {
    return runtime.adapter.trim();
  }
  return 'noop';
}

function routeCatalog() {
  return [
    { routeId: 'founder', label: 'Founder' },
    { routeId: 'ops', label: 'Ops' },
    { routeId: 'brand', label: 'Brand' },
  ];
}

function findRoute(runtime, routeId) {
  const status = readStatus(runtime);
  const routes = Array.isArray(status.routes) ? status.routes : [];
  return routes.find((item) => item?.routeId === routeId) || null;
}

export function normalizeTelegramSurface(runtime) {
  const status = readStatus(runtime);
  const inboundMode = status.inboundMode || 'off';
  const adapter = sendAdapterName(runtime);

  if (status.sendConfigured) {
    return {
      tone: 'good',
      label: adapter,
      detail: inboundMode === 'assistant' ? 'founder delivery and assistant reply path are ready' : 'founder delivery is ready',
    };
  }

  if (status.founderDelivery === 'split_required') {
    return {
      tone: 'alert',
      label: 'split required',
      detail: 'founder room and founder DM share one chat id, so direct alerts stay paused',
    };
  }

  if (status.pollingConfigured) {
    return {
      tone: 'muted',
      label: 'polling only',
      detail: 'inbound polling works, but the founder route still needs a dedicated chat id',
    };
  }

  if (status.chatConfigured) {
    return {
      tone: 'alert',
      label: 'chat only',
      detail: 'chat id exists, but the bot token is missing so Telegram is still offline',
    };
  }

  return {
    tone: 'muted',
    label: adapter,
    detail: 'bot token and chat id not configured',
  };
}

export function normalizeSendState(canWrite, runtime, pending) {
  const status = readStatus(runtime);
  if (pending) {
    return { tone: 'muted', label: 'sending', detail: 'packet dispatch in progress' };
  }
  if (!canWrite) {
    return { tone: 'muted', label: 'read only', detail: 'founder session required for send' };
  }
  if (!status.sendConfigured) {
    if (status.founderDelivery === 'split_required') {
      return { tone: 'alert', label: 'split required', detail: 'founder room and founder DM must be separated before direct alerts can be trusted' };
    }
    return { tone: 'alert', label: 'not ready', detail: 'preview works, but founder delivery is still blocked' };
  }
  return { tone: 'good', label: 'ready', detail: 'canonical founder write path is available' };
}

export function normalizeInboundState(runtime) {
  const status = readStatus(runtime);
  const mode = status.inboundMode || 'off';
  if (mode === 'assistant') {
    return { tone: 'good', label: 'assistant', detail: 'polling and Gemini reply path are active' };
  }
  if (mode === 'command_only') {
    return { tone: 'muted', label: 'command only', detail: 'polling works, but free-text assistant replies stay off' };
  }
  return { tone: 'alert', label: 'off', detail: 'bot token is missing, so inbound polling is off' };
}

export function normalizeDeliveryState(runtime) {
  const status = readStatus(runtime);
  const delivery = status.founderDelivery || 'disabled';
  if (delivery === 'ready') {
    return { tone: 'good', label: 'delivery ready', detail: 'founder alert packets can be sent' };
  }
  if (delivery === 'chat_missing') {
    return { tone: 'alert', label: 'chat missing', detail: 'bot token exists, but TELEGRAM_FOUNDER_CHAT_ID is still missing' };
  }
  if (delivery === 'split_required') {
    return { tone: 'alert', label: 'split required', detail: 'founder room and founder DM share the same chat id' };
  }
  if (delivery === 'token_missing') {
    return { tone: 'alert', label: 'token missing', detail: 'chat id exists, but the bot token is missing' };
  }
  return { tone: 'muted', label: 'disabled', detail: 'bot token and chat id are both missing' };
}

export function normalizeRouteStates(runtime) {
  return routeCatalog().map((routeMeta) => {
    const route = findRoute(runtime, routeMeta.routeId);
    const delivery = route?.delivery || 'disabled';
    let tone = 'muted';
    let detail = 'route not configured yet';

    switch (delivery) {
      case 'ready':
        tone = 'good';
        if (routeMeta.routeId === 'founder') {
          detail = 'founder packets and direct alerts land here';
        } else if (routeMeta.routeId === 'ops') {
          detail = 'job, deploy, and runtime notices land here';
        } else {
          detail = 'brand review and publish traffic can land here';
        }
        break;
      case 'chat_missing':
        tone = 'alert';
        detail = 'set the dedicated chat id before treating this route as live';
        break;
      case 'split_required':
        tone = 'alert';
        detail = 'founder room and founder DM must use different chats before this route can go live';
        break;
      case 'token_missing':
        tone = 'alert';
        detail = 'chat id exists, but the bot token is missing';
        break;
      default:
        if (routeMeta.routeId === 'brand') {
          detail = 'brand route is optional until content review starts living in Telegram';
        } else if (routeMeta.routeId === 'ops') {
          detail = 'ops route is not configured yet';
        } else {
          detail = 'founder route is not configured yet';
        }
        break;
    }

    return {
      routeId: routeMeta.routeId,
      title: route?.label || routeMeta.label,
      tone,
      label: delivery.replaceAll('_', ' '),
      detail,
      notes: Array.isArray(route?.notes) ? route.notes : [],
    };
  });
}
