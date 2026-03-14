'use client';

import { useRouter } from 'next/navigation';
import { useMemo, useState } from 'react';

function formatSourceIds(value) {
  return Array.isArray(value) ? value.join(', ') : '';
}

function formatExampleIds(value) {
  return Array.isArray(value) ? value.join(', ') : '';
}

function parseCSV(value) {
  return String(value || '')
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean);
}

function formatTimestamp(value) {
  if (!value) {
    return '--';
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return '--';
  }
  return date.toLocaleString('ko-KR', {
    hour12: false,
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function recentCount(view) {
  return {
    proposals: view?.recent?.proposals?.length || 0,
    approvals: view?.recent?.approvals?.length || 0,
    flows: view?.recent?.flows?.length || 0,
  };
}

function makeToneClass(tone) {
  if (tone === 'good') return 'status-pill status-pill--good';
  if (tone === 'alert') return 'status-pill status-pill--alert';
  return 'status-pill status-pill--muted';
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

function normalizeThreadsState(view) {
  const status = view?.threads?.status || {};
  if (status.publishConfigured) {
    return {
      tone: 'good',
      label: status.adapter || 'threads_api',
      detail: 'approved Threads draft를 바로 publish할 수 있습니다.',
    };
  }
  return {
    tone: 'alert',
    label: status.adapter || 'noop',
    detail: 'THREADS_ACCESS_TOKEN이 없어 live publish는 아직 잠겨 있습니다.',
  };
}

export default function AdminBrandPublishActions({ canWrite, initialView }) {
  const router = useRouter();
  const [pending, setPending] = useState(false);
  const [message, setMessage] = useState('');
  const [result, setResult] = useState(null);
  const [publishPendingId, setPublishPendingId] = useState('');
  const [publishMessage, setPublishMessage] = useState('');
  const [publishResult, setPublishResult] = useState(null);
  const [form, setForm] = useState({
    channel: initialView?.defaults?.channel || 'threads',
    targetAccount: initialView?.defaults?.targetAccount || '97layer',
    title: initialView?.defaults?.title || '',
    body: initialView?.defaults?.body || '',
    topicTag: initialView?.defaults?.topicTag || '',
    sourceIds: formatSourceIds(initialView?.defaults?.sourceIds),
    notes: '',
  });

  const counters = useMemo(() => recentCount(initialView), [initialView]);
  const threadsState = useMemo(() => normalizeThreadsState(initialView), [initialView]);
  const selectedAccount = useMemo(
    () => (initialView?.accounts || []).find((item) => item.accountId === form.targetAccount) || null,
    [initialView?.accounts, form.targetAccount],
  );
  const threadsEligible = initialView?.threads?.eligible || [];
  const threadsRecent = initialView?.threads?.recent || [];
  const attention = publishResult?.follow_up || result?.next_action || initialView?.attention || null;

  function updateField(key, value) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  function applyPreset(preset) {
    setForm({
      channel: preset.channel,
      targetAccount: preset.targetAccount || initialView?.defaults?.targetAccount || '97layer',
      title: preset.title,
      body: preset.body,
      topicTag: preset.topicTag || '',
      sourceIds: formatSourceIds(preset.sourceIds),
      notes: '',
    });
    setMessage(
      preset.styleExamples?.length
        ? `${preset.label} 초안을 채웠습니다. style ref ${preset.styleExamples.map((item) => item.signalId || item.exampleId).join(', ')}`
        : `${preset.label} 초안을 채웠습니다. 문구만 다듬고 열면 됩니다.`,
    );
    setResult(null);
  }

  async function submit(event) {
    event.preventDefault();
    setPending(true);
    setMessage('');
    setResult(null);
    try {
      const response = await fetch('/api/admin/runtime/brand-publish', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          channel: form.channel,
          target_account: form.targetAccount,
          title: form.title,
          body: form.body,
          topic_tag: form.topicTag,
          source_ids: parseCSV(form.sourceIds),
          notes: parseCSV(form.notes),
        }),
      });
      const payload = await response.json();
      if (!response.ok) {
        throw new Error(payload?.error || 'brand publish failed');
      }
      setResult(payload);
      setMessage(
        payload?.next_action?.summary ||
          `${payload?.lane?.targetAccountLabel || form.targetAccount} · ${payload?.lane?.label || form.channel} 초안을 승인 대기까지 올렸습니다.`,
      );
      router.refresh();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'brand publish failed');
    } finally {
      setPending(false);
    }
  }

  async function publishThreadsCandidate(approvalId) {
    setPublishPendingId(approvalId);
    setPublishMessage('');
    setPublishResult(null);
    try {
      const response = await fetch('/api/admin/runtime/brand-publish/threads', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ approval_id: approvalId }),
      });
      const payload = await response.json();
      if (!response.ok) {
        throw new Error(payload?.error || 'threads publish failed');
      }
      setPublishResult(payload);
      setPublishMessage(payload?.follow_up?.summary || `Threads publish receipt ${payload?.thread_id || payload?.threadId || 'created'}`);
      router.refresh();
    } catch (error) {
      setPublishMessage(error instanceof Error ? error.message : 'threads publish failed');
    } finally {
      setPublishPendingId('');
    }
  }

  return (
    <section className="admin-card admin-card--wide">
      <div className="quickwork-card__head">
        <div>
          <p className="section-eyebrow">Brand publish</p>
          <h2 className="section-title">브랜드 초안 {'->'} 승인 대기 열기</h2>
        </div>
        <div className="inline-list">
          <span>{counters.proposals} proposals</span>
          <span>{counters.approvals} approvals</span>
          <span>{counters.flows} flows</span>
        </div>
      </div>

      <p className="section-body">
        브랜드 소스팩에서 초안을 채우고, proposal {'->'} work {'->'} approval {'->'} flow까지 한 번에 엽니다.
        실제 게시는 다음 슬라이스에서 붙이고, 지금은 founder review corridor를 빠르게 닫는 데 집중합니다.
      </p>

      <div className="signal-grid">
        <article className="signal-card">
          <span className="signal-card__source">다음 액션</span>
          <h3 className="signal-card__title">{attention?.summary || '브랜드 초안을 하나 열어 승인 corridor를 시작하세요.'}</h3>
          <p className="section-body">
            {attention?.detail || '지금은 preset으로 초안을 열고 founder review corridor를 닫는 단계입니다.'}
          </p>
          {attention?.ref ? <p className="section-body">ref · {attention.ref}</p> : null}
          {attention?.mode === 'publish_threads' && attention?.approvalId ? (
            <div className="inline-action-block__buttons">
              <button
                type="button"
                className="button-primary button-primary--compact"
                disabled={!canWrite || !initialView?.threads?.status?.publishConfigured || publishPendingId === attention.approvalId}
                onClick={() => publishThreadsCandidate(attention.approvalId)}
              >
                {publishPendingId === attention.approvalId ? 'publishing…' : attention?.actionLabel || '바로 publish'}
              </button>
            </div>
          ) : null}
        </article>
      </div>

      {initialView?.sourcePack?.threadsProfile ? (
        <>
          <p className="section-body">
            Threads style profile: {initialView.sourcePack.threadsProfile.label}
            {' · '}
            {initialView.sourcePack.threadsProfile.summary}
          </p>
          <p className="section-body">
            Repo-local snapshot examples: {initialView.sourcePack.threadsProfile.exampleCount || 0}
            {initialView.sourcePack.threadsProfile.provenanceLabel ? ` · ${initialView.sourcePack.threadsProfile.provenanceLabel}` : ''}
          </p>
          {initialView.sourcePack.threadsProfile.examples?.length ? (
            <div className="admin-table">
              {initialView.sourcePack.threadsProfile.examples.map((item) => (
                <article className="admin-table__row" key={item.exampleId}>
                  <div>
                    <strong>{item.signalId || item.exampleId}</strong>
                    <p>{item.excerpt}</p>
                  </div>
                  <div>{item.exampleId}</div>
                  <div>{formatTimestamp(item.publishedAt)}</div>
                  <div>style reference</div>
                </article>
              ))}
            </div>
          ) : null}
        </>
      ) : null}

      {initialView?.sourcePack?.media ? (
        <p className="section-body">
          Current visual cue: {initialView.sourcePack.media.title}
          {initialView.sourcePack.media.caption ? ` · ${initialView.sourcePack.media.caption}` : ''}
        </p>
      ) : null}

      <div className="inline-action-block__buttons">
        {initialView?.presets?.map((preset) => (
          <button
            key={preset.presetId}
            type="button"
            className="button-secondary button-secondary--compact"
            onClick={() => applyPreset(preset)}
            disabled={pending}
          >
            {preset.label}
          </button>
        ))}
      </div>

      <form className="admin-action-grid" onSubmit={submit}>
        <section className="admin-card">
          <h3>새 초안</h3>
          <div className="admin-form-grid">
            <input value={form.channel} onChange={(event) => updateField('channel', event.target.value)} placeholder="channel" disabled={!canWrite || pending} />
            <select value={form.targetAccount} onChange={(event) => updateField('targetAccount', event.target.value)} disabled={!canWrite || pending}>
              {(initialView?.accounts || []).map((account) => (
                <option key={account.accountId} value={account.accountId}>
                  {account.label} · {account.toneLevel}
                </option>
              ))}
            </select>
            <input value={form.title} onChange={(event) => updateField('title', event.target.value)} placeholder="title" disabled={!canWrite || pending} />
            <input value={form.topicTag} onChange={(event) => updateField('topicTag', event.target.value)} placeholder="native topic tag (optional)" disabled={!canWrite || pending} />
            <textarea value={form.body} onChange={(event) => updateField('body', event.target.value)} rows={7} placeholder="draft body" disabled={!canWrite || pending} />
            <input value={form.sourceIds} onChange={(event) => updateField('sourceIds', event.target.value)} placeholder="source ids" disabled={!canWrite || pending} />
            <input value={form.notes} onChange={(event) => updateField('notes', event.target.value)} placeholder="extra notes (comma separated)" disabled={!canWrite || pending} />
          </div>
          {selectedAccount ? <p className="section-body">target account · {selectedAccount.summary}</p> : null}
          <button type="submit" className="button-primary button-primary--compact" disabled={!canWrite || pending}>
            {pending ? '승인 준비 여는 중…' : '승인 준비 열기'}
          </button>
          {!canWrite ? <p className="inline-action-block__message">founder write 세션에서만 실행됩니다.</p> : null}
          {message ? <p className="inline-action-block__message">{message}</p> : null}
        </section>

        <section className="admin-card">
          <h3>최근 브랜드 초안</h3>
          <div className="admin-table">
            {initialView?.recent?.proposals?.slice(0, 3).map((item) => (
              <article className="admin-table__row" key={item.proposalId}>
                <div>
                  <strong>{item.title}</strong>
                  <p>{item.proposalId}</p>
                </div>
                <div>{item.channel}</div>
                <div>{item.targetAccount ? `${item.targetAccount} · ${item.status}` : item.status}</div>
                <div>{formatTimestamp(item.updatedAt)}</div>
              </article>
            ))}
            {initialView?.recent?.proposals?.length === 0 ? (
              <article className="admin-table__row">
                <div>
                  <strong>brand queue empty</strong>
                  <p>첫 초안을 열면 여기부터 쌓입니다.</p>
                </div>
                <div>--</div>
                <div>--</div>
                <div>--</div>
              </article>
            ) : null}
          </div>
          <p className="section-body">
            source pack {initialView?.sourcePack?.packId || 'n/a'}
            {' · '}
            {initialView?.sourcePack?.voice?.join(' / ') || 'voice unavailable'}
          </p>
          {initialView?.recent?.proposals?.[0]?.styleProfile ? (
            <p className="section-body">recent style profile · {initialView.recent.proposals[0].styleProfile}</p>
          ) : null}
        </section>

        <section className="admin-card">
          <div className="inline-action-block__buttons">
            <span className={makeToneClass(threadsState.tone)}>{threadsState.label}</span>
            <span>{threadsEligible.length} approved drafts</span>
            <span>{threadsRecent.length} receipts</span>
          </div>
          <h3>Threads publish</h3>
          <p className="section-body">
            founder가 승인한 Threads 초안만 live로 내보냅니다. 게시는 daemon이 수행하고, 결과는 observation receipt로 남깁니다.
          </p>
          {renderList(initialView?.threads?.status?.notes, threadsState.detail)}

          <div className="admin-table">
            {threadsEligible.map((item) => (
              <article className="admin-table__row" key={item.approvalId}>
                <div>
                  <strong>{item.title}</strong>
                  <p>{item.bodyPreview || item.approvalId}</p>
                  {item.targetAccount ? <p>account · {item.targetAccount}</p> : null}
                  {item.topicTag ? <p>topic · {item.topicTag}</p> : null}
                </div>
                <div>{item.flowId || item.proposalId || item.approvalId}</div>
                <div>{item.styleExampleIds?.length ? formatExampleIds(item.styleExampleIds) : item.styleProfile || formatTimestamp(item.resolvedAt)}</div>
                <div>
                  <button
                    type="button"
                    className="button-primary button-primary--compact"
                    disabled={!canWrite || !initialView?.threads?.status?.publishConfigured || publishPendingId === item.approvalId}
                    onClick={() => publishThreadsCandidate(item.approvalId)}
                  >
                    {publishPendingId === item.approvalId ? 'publishing…' : 'publish'}
                  </button>
                </div>
              </article>
            ))}
            {threadsEligible.length === 0 ? (
              <article className="admin-table__row">
                <div>
                  <strong>no approved Threads drafts</strong>
                  <p>approval이 끝난 Threads 초안이 생기면 여기서 바로 publish할 수 있습니다.</p>
                </div>
                <div>--</div>
                <div>--</div>
                <div>--</div>
              </article>
            ) : null}
          </div>

          <p className="section-body">최근 publish receipts</p>
          <div className="admin-table">
            {threadsRecent.map((item) => (
              <article className="admin-table__row" key={item.observationId}>
                <div>
                  <strong>{item.title}</strong>
                  <p>{item.threadId || item.creationId || item.observationId}</p>
                  {item.targetAccount ? <p>account · {item.targetAccount}</p> : null}
                  {item.topicTag ? <p>topic · {item.topicTag}</p> : null}
                </div>
                <div>{item.approvalId || 'n/a'}</div>
                <div>{item.threadId || 'thread pending'}</div>
                <div>{formatTimestamp(item.publishedAt)}</div>
              </article>
            ))}
            {threadsRecent.length === 0 ? (
              <article className="admin-table__row">
                <div>
                  <strong>no publish receipts</strong>
                  <p>첫 Threads publish가 끝나면 여기부터 실제 영수증이 쌓입니다.</p>
                </div>
                <div>--</div>
                <div>--</div>
                <div>--</div>
              </article>
            ) : null}
          </div>

          {publishMessage ? <p className="inline-action-block__message">{publishMessage}</p> : null}
        </section>
      </form>

      {result ? (
        <div className="signal-grid">
          <article className="signal-card">
            <span className="signal-card__source">Created</span>
            <h3 className="signal-card__title">{result.lane?.label || 'brand publish'}</h3>
            <p className="section-body">{result.proposal?.summary || result.observation?.normalized_summary}</p>
            {result?.next_action?.summary ? <p className="section-body">next action · {result.next_action.summary}</p> : null}
            <div className="admin-list">
              <div><strong>proposal</strong><span>{result.proposal?.proposal_id || 'n/a'}</span></div>
              <div><strong>work</strong><span>{result.work_item?.id || 'n/a'}</span></div>
              <div><strong>approval</strong><span>{result.approval?.approval_id || 'n/a'}</span></div>
              <div><strong>account</strong><span>{result.lane?.targetAccountLabel || result.lane?.targetAccount || 'n/a'}</span></div>
              <div><strong>flow</strong><span>{result.flow?.flow_id || 'n/a'}</span></div>
              <div><strong>style refs</strong><span>{result.lane?.styleExamples?.length ? result.lane.styleExamples.map((item) => item.signalId || item.exampleId).join(', ') : 'n/a'}</span></div>
            </div>
          </article>
        </div>
      ) : null}

      {publishResult ? (
        <div className="signal-grid">
          <article className="signal-card">
            <span className="signal-card__source">Threads receipt</span>
            <h3 className="signal-card__title">{publishResult.title || 'Threads publish'}</h3>
            {publishResult?.follow_up?.summary ? <p className="section-body">{publishResult.follow_up.summary}</p> : null}
            <div className="admin-list">
              <div><strong>approval</strong><span>{publishResult.approval_id || publishResult.approvalId || 'n/a'}</span></div>
              <div><strong>account</strong><span>{publishResult.target_account || publishResult.targetAccount || 'n/a'}</span></div>
              <div><strong>thread</strong><span>{publishResult.thread_id || publishResult.threadId || 'n/a'}</span></div>
              <div><strong>creation</strong><span>{publishResult.creation_id || publishResult.creationId || 'n/a'}</span></div>
              <div><strong>topic</strong><span>{publishResult.topic_tag || publishResult.topicTag || 'n/a'}</span></div>
              <div><strong>observation</strong><span>{publishResult.observation_id || publishResult.observationId || 'n/a'}</span></div>
            </div>
          </article>
        </div>
      ) : null}
    </section>
  );
}
