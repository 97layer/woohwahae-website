import { Suspense } from 'react';

import AdminReviewActions from '../../../../components/admin-review-actions';
import { getRequestSessionMeta } from '../../../../lib/auth/request-session';
import {
  formatReviewRoomTimestamp,
  triageReviewRoom,
} from '../../../../lib/review-room-triage';
import {
  getReviewRoomDetailView,
  getReviewRoomSummaryView,
} from '../../../../lib/runtime/view-model';

function asArray(value) {
  return Array.isArray(value) ? value : [];
}

function sortNewest(items) {
  return asArray(items)
    .slice()
    .sort((left, right) => new Date(right.updated_at || right.created_at).getTime() - new Date(left.updated_at || left.created_at).getTime());
}

function describeReviewSource(source) {
  const normalized = String(source || '').trim();
  if (!normalized) {
    return {
      label: 'unknown',
      detail: 'review-room source not reported',
      tone: 'neutral',
    };
  }
  if (normalized.includes('layer-os-read-fallback')) {
    return {
      label: 'local fallback',
      detail: 'read-only fallback is showing cached local state; VM continuity host remains the authority.',
      tone: 'warning',
    };
  }
  if (normalized === 'legacy_import' || normalized === 'absorbed_snapshot') {
    return {
      label: 'absorbed snapshot',
      detail: 'showing older absorbed review notes that were preserved for continuity, not a live second control plane.',
      tone: 'neutral',
    };
  }
  if (normalized.includes('/var/lib/layer-os/review_room.json')) {
    return {
      label: 'live runtime',
      detail: 'reading canonical VM/runtime review-room state.',
      tone: 'good',
    };
  }
  return {
    label: 'runtime',
    detail: normalized,
    tone: 'neutral',
  };
}

function presentReviewSource(source) {
  return describeReviewSource(source).label;
}

function SummaryReviewItems({ items, canWrite }) {
  if (items.length === 0) {
    return <p className="section-body">지금 막고 있는 안건이 없습니다.</p>;
  }
  return (
    <div className="review-list">
      {items.map((item) => (
        <article className="review-item" key={`${item.text}-${item.updatedAt || item.text}`}>
          <div className="review-item__head">
            <strong>{item.text}</strong>
            <span className="status-pill">{item.severity || 'open'}</span>
          </div>
          <p>{presentReviewSource(item.source) || 'runtime source not reported'}</p>
          <p className="section-body">updated {formatReviewRoomTimestamp(item.updatedAt)}</p>
          <AdminReviewActions itemText={item.text} canWrite={canWrite} />
        </article>
      ))}
    </div>
  );
}

function LoadingReviewRoomDetail() {
  return (
    <>
      <section className="admin-card admin-card--wide">
        <h3>누적 잔여 이슈</h3>
        <p className="section-body">이전 architect / verifier 실패 누적 항목을 이어서 불러오는 중입니다.</p>
      </section>

      <section className="admin-card">
        <h3>전략 백로그</h3>
        <p className="section-body">전략 백로그를 이어서 불러오는 중입니다.</p>
      </section>

      <section className="admin-card">
        <h3>기타 미결</h3>
        <p className="section-body">기타 미결 항목을 이어서 불러오는 중입니다.</p>
      </section>

      <section className="admin-card">
        <h3>보류됨</h3>
        <p className="section-body">보류 항목을 이어서 불러오는 중입니다.</p>
      </section>

      <section className="admin-card">
        <h3>수락됨</h3>
        <p className="section-body">수락 항목을 이어서 불러오는 중입니다.</p>
      </section>
    </>
  );
}

async function ReviewRoomDetailSections({ canWrite }) {
  const review = await getReviewRoomDetailView();
  const deferredItems = review.room?.deferred || [];
  const acceptedItems = review.room?.accepted || [];
  const triage = triageReviewRoom(review.room);

  return (
    <>
      <section className="admin-card admin-card--wide">
        <h3>누적 잔여 이슈</h3>
        <p className="section-body">
          이전 architect / verifier 실패 누적입니다. 지금 레인을 직접 막는 것과 분리해서 보되, 지워진 척하지는 않게 남겨둡니다.
        </p>
        {triage.runtimeResidue.length === 0 ? (
          <p className="section-body">누적 잔여 이슈가 없습니다.</p>
        ) : (
          <details className="review-room-archive">
            <summary className="review-room-archive__summary">
              {triage.runtimeResidue.length}개 잔여 이슈
              {triage.sourceBreakdown.residue.length > 0 ? ` · ${triage.sourceBreakdown.residue.map((item) => `${item.label} ${item.count}`).join(' · ')}` : ''}
            </summary>
            <div className="review-list">
              {triage.runtimeResidue.slice(0, 24).map((item) => (
                <article className="review-item" key={`${item.text}-${item.updated_at}`}>
                  <div className="review-item__head">
                    <strong>{item.text}</strong>
                    <span className="status-pill status-pill--muted">{item.severity}</span>
                  </div>
                  <p>{item.why || presentReviewSource(item.source)}</p>
                  <p className="section-body">updated {formatReviewRoomTimestamp(item.updated_at || item.created_at)}</p>
                  <AdminReviewActions itemText={item.text} canWrite={canWrite} />
                </article>
              ))}
            </div>
          </details>
        )}
      </section>

      <section className="admin-card">
        <h3>전략 백로그</h3>
        <p className="section-body">
          아이디어, 리서치, 확장 제안은 결재 안건과 분리해서 봅니다. 지금 당장 막는 것과 나중에 키울 것을 한데 섞지 않습니다.
        </p>
        {triage.strategicBacklog.length === 0 ? (
          <p className="section-body">전략 백로그가 비어 있습니다.</p>
        ) : (
          <details className="review-room-archive">
            <summary className="review-room-archive__summary">
              {triage.strategicBacklog.length}개 전략 항목
              {triage.sourceBreakdown.strategic.length > 0 ? ` · ${triage.sourceBreakdown.strategic.map((item) => `${item.label} ${item.count}`).join(' · ')}` : ''}
            </summary>
            <div className="review-list review-list--compact">
              {triage.strategicBacklog.slice(0, 12).map((item) => (
                <article className="review-item" key={`${item.text}-${item.updated_at}`}>
                  <strong>{item.text}</strong>
                  <p>{item.why || presentReviewSource(item.source)}</p>
                  <p className="section-body">updated {formatReviewRoomTimestamp(item.updated_at || item.created_at)}</p>
                  <AdminReviewActions itemText={item.text} canWrite={canWrite} />
                </article>
              ))}
            </div>
          </details>
        )}
      </section>

      <section className="admin-card">
        <h3>기타 미결</h3>
        {triage.otherUnresolved.length === 0 ? (
          <p className="section-body">기타 미결 항목이 없습니다.</p>
        ) : (
          <div className="review-list review-list--compact">
            {triage.otherUnresolved.slice(0, 8).map((item) => (
              <article className="review-item" key={`${item.text}-${item.updated_at}`}>
                <strong>{item.text}</strong>
                <p>{item.why || presentReviewSource(item.source)}</p>
                <p className="section-body">updated {formatReviewRoomTimestamp(item.updated_at || item.created_at)}</p>
                <AdminReviewActions itemText={item.text} canWrite={canWrite} />
              </article>
            ))}
          </div>
        )}
      </section>

      <section className="admin-card">
        <h3>보류됨</h3>
        <div className="review-list review-list--compact">
          {sortNewest(deferredItems).slice(0, 8).map((item) => (
            <article className="review-item" key={`${item.text}-${item.updated_at}`}>
              <strong>{item.text}</strong>
              <p>{item.why_unresolved || presentReviewSource(item.source)}</p>
            </article>
          ))}
        </div>
      </section>

      <section className="admin-card">
        <h3>수락됨</h3>
        <div className="review-list review-list--compact">
          {sortNewest(acceptedItems).slice(0, 8).map((item) => (
            <article className="review-item" key={`${item.text}-${item.updated_at}`}>
              <strong>{item.text}</strong>
              <p>{presentReviewSource(item.source)}</p>
            </article>
          ))}
        </div>
      </section>
    </>
  );
}

export default async function AdminReviewRoomPage() {
  const [summaryView, session] = await Promise.all([getReviewRoomSummaryView(), getRequestSessionMeta()]);
  const sourceStatus = describeReviewSource(summaryView.source || summaryView.summary?.source);
  const topOpen = summaryView.summary.topOpen || [];

  return (
    <div className="admin-page-grid">
      <section className="admin-card admin-card--wide">
        <p className="section-eyebrow">결재판</p>
        <h2 className="section-title">결재 대기 안건</h2>
        <p className={`section-body ${sourceStatus.tone === 'warning' ? 'section-body--warning' : ''}`}>
          source: <strong>{sourceStatus.label}</strong> · {sourceStatus.detail}
        </p>
        <p className="section-body">
          {summaryView.attention?.summary || '지금 급한 review blocker는 없습니다.'}
        </p>
        {summaryView.attention?.detail ? <p className="section-body">{summaryView.attention.detail}</p> : null}
        <div className="proof-metrics proof-metrics--admin">
          <article className="metric-card"><span className="metric-card__label">결재 대기</span><strong className="metric-card__value">{summaryView.summary.openCount}</strong></article>
          <article className="metric-card"><span className="metric-card__label">즉시 처리 필요</span><strong className="metric-card__value">{topOpen.length}</strong></article>
          <article className="metric-card"><span className="metric-card__label">보류됨</span><strong className="metric-card__value">{summaryView.summary.deferredCount}</strong></article>
          <article className="metric-card"><span className="metric-card__label">수락됨</span><strong className="metric-card__value">{summaryView.summary.acceptedCount}</strong></article>
        </div>
        {summaryView.degradedReason ? (
          <p className="section-body">runtime 응답이 느려 review-room 세부 카드를 뒤에서 이어서 불러옵니다. ({summaryView.degradedReason})</p>
        ) : null}
      </section>

      <section className="admin-card admin-card--wide">
        <h3>지금 막고 있는 것</h3>
        <p className="section-body">
          지금 막는 상위 안건부터 먼저 보여주고, 잔여 이슈와 전략 백로그는 아래에서 분리합니다.
        </p>
        <SummaryReviewItems items={topOpen} canWrite={session.canWrite} />
      </section>

      <Suspense fallback={<LoadingReviewRoomDetail />}>
        <ReviewRoomDetailSections canWrite={session.canWrite} />
      </Suspense>
    </div>
  );
}
