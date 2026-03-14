package api

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"
)

var cockpitPage = template.Must(template.New("cockpit").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Layer OS Cockpit</title>
  <style>
    :root { color-scheme: light; --bg:#f3f4f6; --panel:#ffffff; --ink:#111827; --line:#d1d5db; --muted:#6b7280; }
    * { box-sizing:border-box; }
    body { margin:0; font-family:ui-sans-serif, -apple-system, BlinkMacSystemFont, sans-serif; background:var(--bg); color:var(--ink); }
    header { padding:24px; border-bottom:1px solid var(--line); background:var(--panel); position:sticky; top:0; }
    main { padding:24px; display:grid; gap:16px; grid-template-columns:repeat(auto-fit,minmax(280px,1fr)); }
    section { background:var(--panel); border:1px solid var(--line); border-radius:16px; padding:16px; }
    h1,h2 { margin:0 0 12px; }
    h1 { font-size:24px; }
    h2 { font-size:16px; }
    p,li,pre,button,input,textarea { font:inherit; }
    .muted { color:var(--muted); }
    .row { display:grid; gap:8px; margin-top:8px; }
    .grid { display:grid; gap:8px; grid-template-columns:repeat(2,minmax(0,1fr)); }
    input, textarea, button { width:100%; padding:10px 12px; border-radius:10px; border:1px solid var(--line); background:#fff; }
    button { cursor:pointer; background:#111827; color:#fff; border-color:#111827; }
    pre { margin:0; white-space:pre-wrap; word-break:break-word; background:#0f172a; color:#e5e7eb; padding:12px; border-radius:12px; min-height:120px; }
    code { font-family:ui-monospace, SFMono-Regular, Menlo, monospace; }
  </style>
</head>
<body>
  <header>
    <h1>Layer OS Cockpit</h1>
    <p class="muted">Founder console on top of the kernel contracts.</p>
  </header>
  <main>
    <section>
      <h2>Founder Read Model</h2>
      <p class="muted">Read-first founder surface over kernel state, session bootstrap, and live evidence.</p>
      <div class="row">
        <button data-load="status">Refresh status</button>
        <button data-load="knowledge">Refresh knowledge</button>
        <button data-load="founder-summary">Refresh founder summary</button>
        <button data-load="founder-view">Refresh founder view</button>
        <button data-load="session-bootstrap">Refresh session bootstrap</button>
        <button data-load="review-room">Refresh review room</button>
        <button data-load="capabilities">Refresh capabilities</button>
        <button data-load="providers">Refresh providers</button>
        <button data-load="proposals">Refresh proposals</button>
        <button data-load="jobs">Refresh jobs</button>
        <button data-load="writer">Refresh writer lease</button>
        <button data-load="handoff">Refresh handoff</button>
        <button data-load="cockpit">Refresh full dashboard dump</button>
        <button data-load="events">Refresh recent events</button>
        <pre id="status"></pre>
        <pre id="knowledge"></pre>
        <pre id="founder-summary"></pre>
        <pre id="founder-view"></pre>
        <pre id="session-bootstrap"></pre>
        <pre id="review-room"></pre>
        <pre id="capabilities"></pre>
        <pre id="providers"></pre>
        <pre id="proposals"></pre>
        <pre id="jobs"></pre>
        <pre id="writer-lease"></pre>
        <pre id="handoff"></pre>
        <pre id="cockpit-snapshot"></pre>
        <p id="events-stream-status" class="muted">Live refresh connecting...</p>
        <pre id="events-output"></pre>
      </div>
    </section>
    <section>
      <h2>Founder Actions</h2>
      <div class="grid">
        <input id="founder-flow-id" placeholder="flow_001">
        <input id="founder-title" placeholder="Founder loop">
      </div>
      <div class="grid">
        <input id="founder-work-id" placeholder="work_001">
        <input id="founder-approval-id" placeholder="approval_001">
      </div>
      <div class="grid">
        <input id="founder-release-id" placeholder="release_001">
        <input id="founder-deploy-id" placeholder="deploy_001">
      </div>
      <div class="grid">
        <input id="founder-rollback-id" placeholder="rollback_001">
        <input id="founder-intent" placeholder="Close founder loop">
      </div>
      <div class="row">
        <button id="founder-start">Start</button>
        <button id="founder-approve">Approve</button>
        <button id="founder-release">Release</button>
        <button id="founder-rollback">Rollback</button>
      </div>
      <pre id="founder-action-output"></pre>
    </section>
    <section>
      <h2>Work + Approval</h2>
      <div class="grid">
        <input id="work-id" placeholder="work_001">
        <input id="work-title" placeholder="Bootstrap founder loop">
      </div>
      <div class="grid">
        <input id="work-intent" placeholder="seed execute lane">
        <input id="approval-id" placeholder="approval_001">
      </div>
      <div class="row">
        <button id="create-work">Create work</button>
        <button id="create-approval">Create approval</button>
        <button id="resolve-approval">Resolve approval</button>
      </div>
      <pre id="work-approval-output"></pre>
    </section>
    <section>
      <h2>Flow Bundle</h2>
      <div class="grid">
        <input id="flow-id" placeholder="flow_001">
        <input id="flow-work-id" placeholder="work_001">
      </div>
      <div class="grid">
        <input id="flow-approval-id" placeholder="approval_001">
        <input id="flow-policy-id" placeholder="policy_001">
      </div>
      <div class="grid">
        <input id="flow-execute-id" placeholder="execute_001">
        <input id="flow-verify-id" placeholder="verify_001">
      </div>
      <div class="grid">
        <input id="flow-release-id" placeholder="release_001">
        <input id="flow-deploy-id" placeholder="deploy_001">
      </div>
      <div class="grid">
        <input id="flow-rollback-id" placeholder="rollback_001">
        <input id="flow-note" placeholder="founder-loop">
      </div>
      <div class="row">
        <button id="create-flow">Create flow</button>
        <button id="sync-flow">Sync flow</button>
        <button id="load-flows">Load flows</button>
      </div>
      <pre id="flow-output"></pre>
    </section>
    <section>
      <h2>Policy + Execute</h2>
      <div class="grid">
        <input id="policy-id" placeholder="policy_001">
        <input id="execute-id" placeholder="execute_001">
      </div>
      <div class="grid">
        <input id="policy-risk" placeholder="medium">
        <input id="policy-novelty" placeholder="low">
      </div>
      <div class="row">
        <button id="evaluate-policy">Evaluate policy</button>
        <button id="run-execute">Run execute</button>
      </div>
      <pre id="policy-execute-output"></pre>
    </section>
    <section>
      <h2>Release + Deploy</h2>
      <div class="grid">
        <input id="release-id" placeholder="release_001">
        <input id="deploy-id" placeholder="deploy_001">
      </div>
      <div class="grid">
        <input id="rollback-id" placeholder="rollback_001">
        <input id="rollback-deploy-id" placeholder="deploy_001">
      </div>
      <div class="row">
        <button id="create-release">Create release</button>
        <button id="run-deploy">Run deploy</button>
        <button id="run-rollback">Run rollback</button>
      </div>
      <pre id="release-deploy-output"></pre>
    </section>
    <section>
      <h2>Target + Gateway + Verify</h2>
      <div class="grid">
        <input id="target-id" placeholder="vm">
        <input id="target-command" placeholder="/usr/bin/true">
      </div>
      <div class="grid">
        <input id="gateway-id" placeholder="gateway_001">
        <input id="verify-id" placeholder="verify_001">
      </div>
      <div class="row">
        <button id="put-target">Put target</button>
        <button id="create-gateway">Create gateway call</button>
        <button id="run-verify">Run verify</button>
      </div>
      <pre id="target-gateway-verify-output"></pre>
    </section>
    <section>
      <h2>Memory</h2>
      <textarea id="memory-focus" rows="3" placeholder="Current focus"></textarea>
      <div class="row">
        <button id="set-memory">Update memory</button>
      </div>
      <pre id="memory-output"></pre>
    </section>
    <section>
      <h2>Audit</h2>
      <div class="row">
        <button id="load-adapters">Load adapters</button>
        <button id="load-structure-audit">Load structure audit</button>
        <button id="load-contract-audit">Load contract audit</button>
        <button id="load-residue-audit">Load residue audit</button>
        <button id="load-gemini-audit">Load Gemini audit</button>
        <button id="load-snapshot">Load snapshot</button>
      </div>
      <pre id="audit-output"></pre>
    </section>
    <section>
      <h2>Write Token</h2>
      <input id="write-token" placeholder="Bearer token for POST actions">
      <input id="set-token" placeholder="New write token for auth enable">
      <div class="row">
        <button id="auth-enable">Enable write auth</button>
        <button id="auth-disable">Disable write auth</button>
      </div>
      <p class="muted">Leave empty while write auth is disabled.</p>
      <pre id="auth-output"></pre>
    </section>
  </main>
  <script>
    const tokenInput = document.getElementById('write-token');
    const readOnlyLoads = [
      ['/api/layer-os/status', 'status'],
      ['/api/layer-os/knowledge', 'knowledge'],
      ['/api/layer-os/founder-summary', 'founder-summary'],
      ['/api/layer-os/founder-view', 'founder-view'],
      ['/api/layer-os/session/bootstrap', 'session-bootstrap'],
      ['/api/layer-os/review-room', 'review-room'],
      ['/api/layer-os/capabilities', 'capabilities'],
      ['/api/layer-os/providers', 'providers'],
      ['/api/layer-os/proposals', 'proposals'],
      ['/api/layer-os/writer', 'writer-lease'],
      ['/api/layer-os/handoff', 'handoff'],
      ['/api/layer-os/events', 'events-output'],
    ];
    let readOnlyRefreshTimer = 0;
    function headers() {
      const headers = {'Content-Type':'application/json'};
      const token = tokenInput.value.trim();
      if (token) headers['Authorization'] = 'Bearer ' + token;
      return headers;
    }
    async function load(path, target) {
      const res = await fetch(path);
      const body = await res.text();
      document.getElementById(target).textContent = body;
    }
    async function loadMany(items) {
      await Promise.all(items.map(([path, target]) => load(path, target)));
    }
    async function refreshReadOnlyPanels() {
      await loadMany(readOnlyLoads);
    }
    function scheduleReadOnlyRefresh() {
      if (readOnlyRefreshTimer) window.clearTimeout(readOnlyRefreshTimer);
      readOnlyRefreshTimer = window.setTimeout(() => {
        readOnlyRefreshTimer = 0;
        refreshReadOnlyPanels();
      }, 120);
    }
    function setEventsStreamStatus(text) {
      document.getElementById('events-stream-status').textContent = text;
    }
    function parseStreamPayload(streamEvent) {
      const raw = typeof streamEvent.data === 'string' ? streamEvent.data.trim() : '';
      if (!raw) return null;
      try {
        return JSON.parse(raw);
      } catch (_err) {
        return {kind: raw};
      }
    }
    function streamKind(streamEvent, payload) {
      const candidates = [];
      if (typeof streamEvent.type === 'string' && streamEvent.type !== 'message') candidates.push(streamEvent.type);
      if (payload && typeof payload === 'object') {
        candidates.push(payload.kind, payload.event_kind, payload.type);
        if (payload.event && typeof payload.event === 'object') {
          candidates.push(payload.event.kind, payload.event.event_kind, payload.event.type);
        }
      }
      for (const candidate of candidates) {
        if (typeof candidate === 'string' && candidate.trim()) return candidate.trim().toLowerCase();
      }
      return '';
    }
    function shouldRefreshReadOnlyPanels(streamEvent) {
      const kind = streamKind(streamEvent, parseStreamPayload(streamEvent));
      if (!kind) return true;
      return !kind.includes('heartbeat') && !kind.includes('keepalive') && kind !== 'connected' && kind !== 'ping' && kind !== 'ready';
    }
    function connectEventsStream() {
      if (typeof EventSource === 'undefined') {
        setEventsStreamStatus('Live refresh unavailable in this browser.');
        return;
      }
      setEventsStreamStatus('Live refresh connecting...');
      const stream = new EventSource('/api/layer-os/events/stream');
      const handleStreamEvent = (streamEvent) => {
        if (!shouldRefreshReadOnlyPanels(streamEvent)) return;
        scheduleReadOnlyRefresh();
      };
      stream.onopen = () => setEventsStreamStatus('Live refresh connected.');
      stream.onerror = () => setEventsStreamStatus('Live refresh reconnecting...');
      stream.onmessage = handleStreamEvent;
    }
    async function post(path, payload, target) {
      const res = await fetch(path, {method:'POST', headers:headers(), body:JSON.stringify(payload)});
      const body = await res.text();
      document.getElementById(target).textContent = body;
      await refreshReadOnlyPanels();
    }
    async function auth(method, payload, target) {
      const res = await fetch('/api/layer-os/auth', {method, headers:headers(), body: payload ? JSON.stringify(payload) : undefined});
      const body = await res.text();
      document.getElementById(target).textContent = body;
      await refreshReadOnlyPanels();
      await load('/api/layer-os/founder-view', 'founder-view');
    }
    async function founder(path, payload) {
      await post(path, payload, 'founder-action-output');
      await load('/api/layer-os/founder-view', 'founder-view');
    }
    document.querySelector('[data-load="status"]').onclick = () => load('/api/layer-os/status', 'status');
    document.querySelector('[data-load="knowledge"]').onclick = () => load('/api/layer-os/knowledge', 'knowledge');
    document.querySelector('[data-load="founder-summary"]').onclick = () => load('/api/layer-os/founder-summary', 'founder-summary');
    document.querySelector('[data-load="founder-view"]').onclick = () => load('/api/layer-os/founder-view', 'founder-view');
    document.querySelector('[data-load="session-bootstrap"]').onclick = () => load('/api/layer-os/session/bootstrap', 'session-bootstrap');
    document.querySelector('[data-load="review-room"]').onclick = () => load('/api/layer-os/review-room', 'review-room');
    document.querySelector('[data-load="capabilities"]').onclick = () => load('/api/layer-os/capabilities', 'capabilities');
    document.querySelector('[data-load="proposals"]').onclick = () => load('/api/layer-os/proposals', 'proposals');
    document.querySelector('[data-load="writer"]').onclick = () => load('/api/layer-os/writer', 'writer-lease');
    document.querySelector('[data-load="handoff"]').onclick = () => load('/api/layer-os/handoff', 'handoff');
    document.querySelector('[data-load="cockpit"]').onclick = () => load('/api/layer-os/cockpit?full=1', 'cockpit-snapshot');
    document.querySelector('[data-load="events"]').onclick = () => load('/api/layer-os/events', 'events-output');
    document.getElementById('founder-start').onclick = () => founder('/api/layer-os/founder-actions/start', {
      flow_id: document.getElementById('founder-flow-id').value.trim() || document.getElementById('flow-id').value.trim(),
      work_item_id: document.getElementById('founder-work-id').value.trim() || document.getElementById('work-id').value.trim(),
      approval_id: document.getElementById('founder-approval-id').value.trim() || document.getElementById('approval-id').value.trim(),
      title: document.getElementById('founder-title').value.trim() || document.getElementById('work-title').value.trim() || 'Founder loop',
      intent: document.getElementById('founder-intent').value.trim() || document.getElementById('work-intent').value.trim() || 'Close founder loop',
      notes: ['cockpit-founder-start']
    });
    document.getElementById('founder-approve').onclick = () => founder('/api/layer-os/founder-actions/approve', {
      flow_id: document.getElementById('founder-flow-id').value.trim() || document.getElementById('flow-id').value.trim(),
      notes: ['cockpit-founder-approve']
    });
    document.getElementById('founder-release').onclick = () => founder('/api/layer-os/founder-actions/release', {
      flow_id: document.getElementById('founder-flow-id').value.trim() || document.getElementById('flow-id').value.trim(),
      release_id: document.getElementById('founder-release-id').value.trim() || document.getElementById('release-id').value.trim(),
      deploy_id: document.getElementById('founder-deploy-id').value.trim() || document.getElementById('deploy-id').value.trim(),
      target: document.getElementById('target-id').value.trim() || 'vm',
      channel: 'cockpit',
      notes: ['cockpit-founder-release']
    });
    document.getElementById('founder-rollback').onclick = () => founder('/api/layer-os/founder-actions/rollback', {
      flow_id: document.getElementById('founder-flow-id').value.trim() || document.getElementById('flow-id').value.trim(),
      rollback_id: document.getElementById('founder-rollback-id').value.trim() || document.getElementById('rollback-id').value.trim(),
      notes: ['cockpit-founder-rollback']
    });
    document.getElementById('create-work').onclick = () => post('/api/layer-os/work-items', {
      id: document.getElementById('work-id').value.trim(),
      title: document.getElementById('work-title').value.trim(),
      intent: document.getElementById('work-intent').value.trim(),
      stage: 'discover',
      surface: 'cockpit',
      pack: 'founder',
      priority: 'high',
      risk: document.getElementById('policy-risk').value.trim() || 'medium',
      requires_approval: true,
      payload: {},
      correlation_id: document.getElementById('work-id').value.trim()
    }, 'work-approval-output');
    document.getElementById('create-approval').onclick = () => post('/api/layer-os/approval-inbox', {
      approval_id: document.getElementById('approval-id').value.trim(),
      work_item_id: document.getElementById('work-id').value.trim(),
      stage: 'verify',
      summary: 'Founder approval',
      risks: ['founder gate required'],
      rollback_plan: 'stop release',
      decision_surface: 'cockpit',
      status: 'pending'
    }, 'work-approval-output');
    document.getElementById('resolve-approval').onclick = () => post('/api/layer-os/approval-inbox/resolve', {
      approval_id: document.getElementById('approval-id').value.trim(),
      status: 'approved'
    }, 'work-approval-output');
    document.getElementById('create-flow').onclick = () => post('/api/layer-os/flows', {
      flow_id: document.getElementById('flow-id').value.trim(),
      work_item_id: document.getElementById('flow-work-id').value.trim() || document.getElementById('work-id').value.trim(),
      status: 'active',
      notes: [document.getElementById('flow-note').value.trim() || 'cockpit']
    }, 'flow-output');
    document.getElementById('sync-flow').onclick = () => post('/api/layer-os/flows/sync', {
      flow_id: document.getElementById('flow-id').value.trim(),
      work_item_id: document.getElementById('flow-work-id').value.trim() || document.getElementById('work-id').value.trim(),
      approval_id: document.getElementById('flow-approval-id').value.trim() || document.getElementById('approval-id').value.trim(),
      policy_decision_id: document.getElementById('flow-policy-id').value.trim() || document.getElementById('policy-id').value.trim(),
      execute_id: document.getElementById('flow-execute-id').value.trim() || document.getElementById('execute-id').value.trim(),
      verification_id: document.getElementById('flow-verify-id').value.trim() || document.getElementById('verify-id').value.trim(),
      release_id: document.getElementById('flow-release-id').value.trim() || document.getElementById('release-id').value.trim(),
      deploy_id: document.getElementById('flow-deploy-id').value.trim() || document.getElementById('deploy-id').value.trim(),
      rollback_id: document.getElementById('flow-rollback-id').value.trim() || document.getElementById('rollback-id').value.trim(),
      notes: [document.getElementById('flow-note').value.trim() || 'cockpit']
    }, 'flow-output');
    document.getElementById('load-flows').onclick = () => load('/api/layer-os/flows', 'flow-output');
    document.getElementById('evaluate-policy').onclick = () => post('/api/layer-os/policies/evaluate', {
      decision_id: document.getElementById('policy-id').value.trim(),
      intent: document.getElementById('work-intent').value.trim(),
      scope: 'kernel',
      risk: document.getElementById('policy-risk').value.trim() || 'medium',
      novelty: document.getElementById('policy-novelty').value.trim() || 'low',
      token_class: 'small',
      requires_approval: true
    }, 'policy-execute-output');
    document.getElementById('run-execute').onclick = () => post('/api/layer-os/execute-runs/run', {
      execute_id: document.getElementById('execute-id').value.trim(),
      work_item_id: document.getElementById('work-id').value.trim(),
      policy_decision_id: document.getElementById('policy-id').value.trim(),
      notes: ['cockpit']
    }, 'policy-execute-output');
    document.getElementById('create-release').onclick = () => post('/api/layer-os/releases', {
      release_id: document.getElementById('release-id').value.trim(),
      work_item_id: document.getElementById('work-id').value.trim(),
      target: 'vm',
      channel: 'cockpit',
      artifacts: ['founder-loop'],
      metrics: {},
      rollback_plan: 'hold release',
      approval_refs: [document.getElementById('approval-id').value.trim()]
    }, 'release-deploy-output');
    document.getElementById('run-deploy').onclick = () => post('/api/layer-os/deploys/execute', {
      deploy_id: document.getElementById('deploy-id').value.trim(),
      release_id: document.getElementById('release-id').value.trim(),
      notes: ['cockpit']
    }, 'release-deploy-output');
    document.getElementById('run-rollback').onclick = () => post('/api/layer-os/rollbacks/execute', {
      rollback_id: document.getElementById('rollback-id').value.trim(),
      release_id: document.getElementById('release-id').value.trim(),
      deploy_id: document.getElementById('rollback-deploy-id').value.trim(),
      notes: ['cockpit']
    }, 'release-deploy-output');
    document.getElementById('put-target').onclick = () => post('/api/layer-os/deploy-targets', {
      target_id: document.getElementById('target-id').value.trim() || 'vm',
      command: [document.getElementById('target-command').value.trim() || '/usr/bin/true']
    }, 'target-gateway-verify-output');
    document.getElementById('create-gateway').onclick = () => post('/api/layer-os/gateway-calls', {
      call_id: document.getElementById('gateway-id').value.trim(),
      decision_id: document.getElementById('policy-id').value.trim(),
      provider: 'openai',
      model: 'gpt-5.4',
      request_kind: 'verify',
      status: 'recorded',
      token_budget: 8000,
      notes: ['cockpit']
    }, 'target-gateway-verify-output');
    document.getElementById('run-verify').onclick = () => post('/api/layer-os/verifications/run', {
      record_id: document.getElementById('verify-id').value.trim(),
      scope: 'kernel',
      notes: ['cockpit']
    }, 'target-gateway-verify-output');
    document.getElementById('set-memory').onclick = () => post('/api/layer-os/memory', {
      current_focus: document.getElementById('memory-focus').value.trim() || 'founder loop',
      next_steps: [],
      open_risks: []
    }, 'memory-output');
    document.getElementById('load-adapters').onclick = () => load('/api/layer-os/adapters', 'audit-output');
    document.getElementById('load-structure-audit').onclick = () => load('/api/layer-os/audit/structure', 'audit-output');
    document.getElementById('load-contract-audit').onclick = () => load('/api/layer-os/audit/contracts', 'audit-output');
    document.getElementById('load-residue-audit').onclick = () => load('/api/layer-os/audit/residue', 'audit-output');
    document.getElementById('load-gemini-audit').onclick = () => load('/api/layer-os/audit/gemini', 'audit-output');
    document.getElementById('load-snapshot').onclick = () => load('/api/layer-os/snapshot', 'audit-output');
    document.getElementById('auth-enable').onclick = () => auth('POST', {token: document.getElementById('set-token').value.trim()}, 'auth-output');
    document.getElementById('auth-disable').onclick = () => auth('DELETE', null, 'auth-output');
    load('/api/layer-os/cockpit?full=1', 'cockpit-snapshot');
    refreshReadOnlyPanels();
    connectEventsStream();
  </script>
</body>
</html>`))

func serveCockpit(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	_ = cockpitPage.Execute(w, nil)
}

func SmokeCockpit() []string {
	var buf bytes.Buffer
	_ = cockpitPage.Execute(&buf, nil)
	html := buf.String()
	required := []string{
		"Layer OS Cockpit",
		"Founder Read Model",
		"Founder Actions",
		"Refresh knowledge",
		"Refresh capabilities",
		"Refresh session bootstrap",
		"Load contract audit",
		"Load Gemini audit",
		"Load snapshot",
	}
	issues := []string{}
	for _, needle := range required {
		if !strings.Contains(html, needle) {
			issues = append(issues, "missing cockpit marker: "+needle)
		}
	}
	return issues
}
