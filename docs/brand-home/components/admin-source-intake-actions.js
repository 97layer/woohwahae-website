'use client';

import { useRouter } from 'next/navigation';
import { useMemo, useState } from 'react';

function parseCSV(value) {
  return String(value || '')
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean);
}

function formatCSV(value) {
  return Array.isArray(value) ? value.join(', ') : '';
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

function toneClass(policyColor) {
  if (policyColor === 'green') return 'status-pill status-pill--good';
  if (policyColor === 'yellow') return 'status-pill status-pill--alert';
  return 'status-pill status-pill--muted';
}

function draftSeedList(item) {
  if (Array.isArray(item?.draftSeeds) && item.draftSeeds.length > 0) {
    return item.draftSeeds;
  }
  return item?.draftSeed ? [item.draftSeed] : [];
}

function prepLaneList(item) {
  if (Array.isArray(item?.prepLanes) && item.prepLanes.length > 0) {
    return item.prepLanes;
  }
  return item?.prepLane ? [item.prepLane] : [];
}

export default function AdminSourceIntakeActions({ canWrite, initialView }) {
  const router = useRouter();
  const [pending, setPending] = useState(false);
  const [prepPendingId, setPrepPendingId] = useState('');
  const [message, setMessage] = useState('');
  const [result, setResult] = useState(null);
  const [form, setForm] = useState({
    intakeClass: initialView?.defaults?.intakeClass || 'manual_drop',
    policyColor: initialView?.defaults?.policyColor || 'green',
    title: initialView?.defaults?.title || '',
    url: initialView?.defaults?.url || '',
    excerpt: initialView?.defaults?.excerpt || '',
    founderNote: initialView?.defaults?.founderNote || '',
    domainTags: formatCSV(initialView?.defaults?.domainTags),
    worldviewTags: formatCSV(initialView?.defaults?.worldviewTags),
    suggestedRoutes: initialView?.defaults?.suggestedRoutes || ['97layer'],
  });

  const selectedRoutes = useMemo(
    () => (initialView?.routeOptions || []).filter((item) => form.suggestedRoutes.includes(item.routeId)),
    [initialView?.routeOptions, form.suggestedRoutes],
  );
  const attention = result?.follow_up || result?.next_action || initialView?.attention || null;

  function updateField(key, value) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  function toggleRoute(routeId) {
    setForm((current) => {
      const next = current.suggestedRoutes.includes(routeId)
        ? current.suggestedRoutes.filter((item) => item !== routeId)
        : [...current.suggestedRoutes, routeId];
      return {
        ...current,
        suggestedRoutes: next.length ? next : ['97layer'],
      };
    });
  }

  async function submit(event) {
    event.preventDefault();
    setPending(true);
    setMessage('');
    setResult(null);
    try {
      const response = await fetch('/api/admin/runtime/source-intake', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          intake_class: form.intakeClass,
          policy_color: form.policyColor,
          title: form.title,
          url: form.url,
          excerpt: form.excerpt,
          founder_note: form.founderNote,
          domain_tags: parseCSV(form.domainTags),
          worldview_tags: parseCSV(form.worldviewTags),
          suggested_routes: form.suggestedRoutes,
        }),
      });
      const payload = await response.json();
      if (!response.ok) {
        throw new Error(payload?.error || 'source intake create failed');
      }
      setResult(payload);
      setMessage(payload?.next_action?.summary || `source intake 저장 완료 · ${payload?.created?.summary || 'new source unit created'}`);
      setForm((current) => ({
        ...current,
        title: '',
        url: '',
        excerpt: '',
        founderNote: '',
        domainTags: '',
        worldviewTags: '',
      }));
      router.refresh();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'source intake create failed');
    } finally {
      setPending(false);
    }
  }

  async function openThreadsPrep(draftSeed) {
    setPrepPendingId(draftSeed?.observationId || '');
    setMessage('');
    setResult(null);
    try {
      const response = await fetch('/api/admin/runtime/source-intake/drafts', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          draft_observation_id: draftSeed?.observationId,
          channel: 'threads',
        }),
      });
      const payload = await response.json();
      if (!response.ok) {
        throw new Error(payload?.error || 'source draft prep open failed');
      }
      setResult(payload);
      setMessage(
        payload?.follow_up?.summary ||
          `Threads 준비 열림 · ${payload?.lane?.targetAccountLabel || draftSeed?.targetAccountLabel || 'account'} · ${payload?.lane?.title || draftSeed?.title || 'draft'}`,
      );
      router.refresh();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'source draft prep open failed');
    } finally {
      setPrepPendingId('');
    }
  }

  return (
    <section className="admin-card admin-card--wide">
      <div className="quickwork-card__head">
        <div>
          <p className="section-eyebrow">Source intake</p>
          <h2 className="section-title">외부 소스 inbox</h2>
        </div>
        <div className="inline-list">
          <span>{initialView?.summaryMeta || ''}</span>
          <span>{form.suggestedRoutes.join(', ')}</span>
        </div>
      </div>

      <p className="section-body">
        {initialView?.summaryNote || ''}
      </p>

      <div className="signal-grid">
        <article className="signal-card">
          <span className="signal-card__source">다음 액션</span>
          <h3 className="signal-card__title">{attention?.summary || '링크나 메모를 먼저 하나 쌓아두세요.'}</h3>
          <p className="section-body">
            {attention?.detail || '지금은 raw source를 observation으로 쌓고, draft seed나 prep lane이 생기면 그때 다음 단계가 여기 먼저 뜹니다.'}
          </p>
          {attention?.ref ? <p className="section-body">ref · {attention.ref}</p> : null}
          {attention?.mode === 'open_threads_prep' && attention?.draftObservationId ? (
            <div className="inline-action-block__buttons">
              <button
                type="button"
                className="button-primary button-primary--compact"
                disabled={!canWrite || prepPendingId === attention.draftObservationId}
                onClick={() => openThreadsPrep({ observationId: attention.draftObservationId })}
              >
                {prepPendingId === attention.draftObservationId ? '여는 중…' : attention?.actionLabel || 'Threads 준비 열기'}
              </button>
            </div>
          ) : null}
        </article>
      </div>

      <form className="admin-action-grid" onSubmit={submit}>
        <section className="admin-card">
          <h3>새 source unit</h3>
          <div className="admin-form-grid">
            <select value={form.intakeClass} onChange={(event) => updateField('intakeClass', event.target.value)} disabled={!canWrite || pending}>
              {(initialView?.intakeClasses || []).map((item) => (
                <option key={item.value} value={item.value}>
                  {item.label}
                </option>
              ))}
            </select>
            <select value={form.policyColor} onChange={(event) => updateField('policyColor', event.target.value)} disabled={!canWrite || pending}>
              {(initialView?.policyColors || []).map((item) => (
                <option key={item.value} value={item.value}>
                  {item.label}
                </option>
              ))}
            </select>
            <input value={form.title} onChange={(event) => updateField('title', event.target.value)} placeholder="title or short source name" disabled={!canWrite || pending} />
            <input value={form.url} onChange={(event) => updateField('url', event.target.value)} placeholder="source url (optional)" disabled={!canWrite || pending} />
            <textarea value={form.excerpt} onChange={(event) => updateField('excerpt', event.target.value)} rows={7} placeholder="paste the raw text, key excerpt, or your note" disabled={!canWrite || pending} />
            <textarea value={form.founderNote} onChange={(event) => updateField('founderNote', event.target.value)} rows={3} placeholder="founder note: why this matters / what to do later" disabled={!canWrite || pending} />
            <input value={form.domainTags} onChange={(event) => updateField('domainTags', event.target.value)} placeholder="domain tags (comma separated)" disabled={!canWrite || pending} />
            <input value={form.worldviewTags} onChange={(event) => updateField('worldviewTags', event.target.value)} placeholder="worldview tags (comma separated)" disabled={!canWrite || pending} />
          </div>

          <div className="inline-list">
            {(initialView?.routeOptions || []).map((route) => (
              <label key={route.routeId}>
                <input
                  type="checkbox"
                  checked={form.suggestedRoutes.includes(route.routeId)}
                  onChange={() => toggleRoute(route.routeId)}
                  disabled={!canWrite || pending}
                />
                <span>{route.label}</span>
              </label>
            ))}
          </div>

          {selectedRoutes.length ? (
            <p className="section-body">
              route cue · {selectedRoutes.map((item) => `${item.label}: ${item.cue}`).join(' / ')}
            </p>
          ) : null}

          <button type="submit" className="button-primary button-primary--compact" disabled={!canWrite || pending}>
            {pending ? '저장 중…' : 'source intake 저장'}
          </button>
          {!canWrite ? <p className="inline-action-block__message">founder write 세션에서만 source intake를 저장할 수 있습니다.</p> : null}
          {message ? <p className="inline-action-block__message">{message}</p> : null}
        </section>

        <section className="admin-card">
          <h3>활성 intake</h3>
          <div className="admin-table">
            {(initialView?.items || []).map((item) => (
              <article className="admin-table__row" key={item.observationId}>
                <div>
                  <strong>{item.title || item.summary}</strong>
                  <p>{item.url || item.summary}</p>
                  <p>{item.excerpt}</p>
                  <p>origin · {item.originLabel}{item.feedSourceLabel ? ` · ${item.feedSourceLabel}` : ''}</p>
                  <p>priority · {item.priorityScore} · {item.dispositionLabel}{item.dispositionNote ? ` · ${item.dispositionNote}` : ''}</p>
                  {item.routeDecision ? (
                    <p>route opened · {item.routeDecision.decisionLabel} · {item.routeDecision.routeSourceLabel}</p>
                  ) : null}
                  {prepLaneList(item).map((prepLane) => (
                    <p key={`${prepLane.observationId}-${prepLane.channel}`}>prep lane · {prepLane.targetAccountLabel} · {prepLane.channelLabel}</p>
                  ))}
                  {draftSeedList(item).map((draftSeed) => (
                    <div key={draftSeed.observationId} className="inline-action-block">
                      <p>draft seed · {draftSeed.targetAccountLabel} · {draftSeed.targetTone}</p>
                      {draftSeed.revisionNote ? <p>revision · {draftSeed.revisionNote}</p> : null}
                      <p>{draftSeed.preview}</p>
                      {prepLaneList(item).some((prepLane) => prepLane.draftObservationId === draftSeed.observationId) ? (
                        <p>이미 {prepLaneList(item).find((prepLane) => prepLane.draftObservationId === draftSeed.observationId)?.channelLabel || 'prep'} 준비가 열려 있습니다.</p>
                      ) : (
                        <button
                          type="button"
                          className="button-primary button-primary--compact"
                          disabled={!canWrite || prepPendingId === draftSeed.observationId}
                          onClick={() => openThreadsPrep(draftSeed)}
                        >
                          {prepPendingId === draftSeed.observationId ? '여는 중…' : 'Threads 준비 열기'}
                        </button>
                      )}
                    </div>
                  ))}
                </div>
                <div>
                  <span className={toneClass(item.policyColor)}>{item.policyColorLabel}</span>
                  <p>{item.intakeClassLabel}</p>
                </div>
                <div>{item.suggestedRoutes?.join(', ') || '97layer'}</div>
                <div>{formatTimestamp(item.capturedAt)}</div>
              </article>
            ))}
            {initialView?.items?.length === 0 ? (
              <article className="admin-table__row">
                <div>
                  <strong>active action queue empty</strong>
                  <p>
                    {initialView?.quietCount
                      ? `급한 intake는 없고, quiet candidate ${initialView.quietCount}개가 축적 중입니다.`
                      : '링크나 텍스트를 하나 저장하면 여기부터 쌓입니다.'}
                  </p>
                </div>
                <div>--</div>
                <div>--</div>
                <div>--</div>
              </article>
            ) : null}
          </div>
          <p className="section-body">
            founder 기본 목록에는 review, prep, draft/prep lane이 열린 intake만 올립니다. observe disposition은 quiet candidate로 남겨두고 필요할 때만 당겨옵니다.
          </p>
          {initialView?.quietCount ? (
            <div className="admin-card admin-card--subtle">
              <h4>quiet candidates</h4>
              <p className="section-body">
                {initialView?.quietNote || ''}
              </p>
              <div className="admin-list">
                {(initialView?.quietItems || []).map((item) => (
                  <div key={item.observationId}>
                    <strong>{item.title || item.summary}</strong>
                    <span>
                      priority {item.priorityScore} · {item.dispositionLabel}
                      {item.feedSourceLabel ? ` · ${item.feedSourceLabel}` : ''}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          ) : null}
        </section>
      </form>

      {result?.created ? (
        <div className="signal-grid">
          <article className="signal-card">
            <span className="signal-card__source">Created</span>
            <h3 className="signal-card__title">{result.created.title || 'source intake'}</h3>
            <p className="section-body">{result.created.summary}</p>
            {result?.next_action?.summary ? <p className="section-body">next action · {result.next_action.summary}</p> : null}
            <div className="admin-list">
              <div><strong>observation</strong><span>{result.created.observationId}</span></div>
              <div><strong>policy</strong><span>{result.created.policyColorLabel}</span></div>
              <div><strong>priority</strong><span>{result.created.priorityScore} · {result.created.dispositionLabel}</span></div>
              <div><strong>class</strong><span>{result.created.intakeClassLabel}</span></div>
              <div><strong>routes</strong><span>{result.created.suggestedRoutes?.join(', ') || '97layer'}</span></div>
              <div><strong>domain tags</strong><span>{formatCSV(result.created.domainTags) || 'n/a'}</span></div>
              <div><strong>worldview tags</strong><span>{formatCSV(result.created.worldviewTags) || 'n/a'}</span></div>
            </div>
          </article>
        </div>
      ) : null}
    </section>
  );
}
