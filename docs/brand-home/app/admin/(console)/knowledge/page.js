import { Suspense } from 'react';

import { formatRuntimeTimestamp } from '../../../../components/neural-graph-time.mjs';
import { getKnowledgeDetailView, getKnowledgeSummaryView } from '../../../../lib/runtime/view-model';

function formatEntryTime(value) {
  if (!value) return '';
  try {
    return formatRuntimeTimestamp(value);
  } catch {
    return value.slice(0, 10);
  }
}

function extractEntryLabel(entryId) {
  if (!entryId) return '';
  // cap_job_ctx_2bc85ba2:agent_job.succeeded:20260313T153506 → 뒤 날짜 제거해서 짧게
  const parts = entryId.split(':');
  if (parts.length >= 2) return parts.slice(0, 2).join(':');
  return entryId;
}

function LoadingKnowledgeDetail({ query }) {
  return (
    <>
      <section className="admin-card admin-card--wide">
        <h3>최근 기억 기록</h3>
        <p className="section-body">최근 기록과 검색 결과를 백그라운드에서 이어서 불러오는 중입니다.</p>
      </section>

      {query ? (
        <section className="admin-card admin-card--wide">
          <h3>검색 결과</h3>
          <p className="section-body">검색 결과를 이어서 불러오는 중입니다.</p>
        </section>
      ) : null}
    </>
  );
}

async function KnowledgeDetailSection({ query }) {
  const detail = await getKnowledgeDetailView(query);
  return (
    <>
      <section className="admin-card admin-card--wide">
        <h3>최근 기억 기록</h3>
        <div className="admin-table">
          {detail.entries.map((entry) => (
            <article className="admin-table__row" key={entry.entry_id}>
              <div>
                <strong>{entry.situation?.summary || '상황 요약 없음'}</strong>
                <p className="knowledge-entry-id">{extractEntryLabel(entry.entry_id)}</p>
              </div>
              <div className="knowledge-entry-time">{formatEntryTime(entry.created_at)}</div>
            </article>
          ))}
        </div>
      </section>

      {detail.search ? (
        <section className="admin-card admin-card--wide">
          <h3>검색 결과</h3>
          <div className="admin-table">
            {(detail.search.results || []).map((result) => (
              <article className="admin-table__row" key={result.entry.entry_id}>
                <div>
                  <strong>{result.entry.situation?.summary || extractEntryLabel(result.entry.entry_id)}</strong>
                  <p className="knowledge-entry-id">{(result.match_fields || []).join(' · ')}</p>
                </div>
                <div className="knowledge-entry-time">{result.match_count}건 일치</div>
              </article>
            ))}
          </div>
        </section>
      ) : null}
    </>
  );
}

export default async function AdminKnowledgePage({ searchParams }) {
  const params = await searchParams;
  const query = typeof params?.q === 'string' ? params.q : '';
  const summary = await getKnowledgeSummaryView();

  return (
    <div className="admin-page-grid">
      <section className="admin-card admin-card--wide">
        <p className="section-eyebrow">기억 창고</p>
        <h2 className="section-title">집중할 것 · 리스크 · 기록 검색</h2>
        <form className="knowledge-search" action="/admin/knowledge" method="get">
          <input type="text" name="q" defaultValue={query} placeholder="기억 창고 검색..." />
          <button type="submit" className="button-primary button-primary--compact">검색</button>
        </form>
        <p className="section-body">
          지금 중요한 신호를 먼저 보고, 최근 기록과 검색 결과는 뒤에서 이어서 불러옵니다.
        </p>
        <p className="section-body">
          {summary.attention?.summary || '지금 먼저 볼 긴급 신호가 없습니다.'}
        </p>
        {summary.attention?.detail ? <p className="section-body">{summary.attention.detail}</p> : null}
      </section>

      <section className="admin-card">
        <h3>지금 집중할 것</h3>
        <p>{summary.knowledge.currentFocus || '집중 항목이 아직 없습니다.'}</p>
        <ul className="inline-list">
          {summary.knowledge.nextSteps.map((item) => <li key={item}>{item}</li>)}
        </ul>
      </section>

      <section className="admin-card">
        <h3>열린 리스크</h3>
        {summary.knowledge.openRisks.length === 0 ? (
          <p className="section-body">현재 열린 리스크가 없습니다.</p>
        ) : (
          <ul className="inline-list">
            {summary.knowledge.openRisks.map((item) => <li key={item}>{item}</li>)}
          </ul>
        )}
      </section>

      <Suspense fallback={<LoadingKnowledgeDetail query={query} />}>
        <KnowledgeDetailSection query={query} />
      </Suspense>
    </div>
  );
}
