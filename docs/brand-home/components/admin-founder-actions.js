'use client';

import { useRouter } from 'next/navigation';
import { useState } from 'react';

export default function AdminFounderActions({ canWrite, defaults = {} }) {
  const router = useRouter();
  const [pending, setPending] = useState(false);
  const [message, setMessage] = useState('');
  const attention = defaults.attention || null;
  const actionLabels = {
    start: '흐름 시작',
    approve: '승인 완료',
    release: '릴리스 진행',
    rollback: '롤백 기록',
  };
  const [startForm, setStartForm] = useState({
    flow_id: defaults.flowId || '',
    work_item_id: '',
    approval_id: defaults.approvalId || '',
    title: defaults.title || '창업자 흐름',
    intent: defaults.intent || '운영 화면에서 창업자 흐름을 닫습니다.',
  });
  const [approveForm, setApproveForm] = useState({ flow_id: defaults.flowId || '' });
  const [releaseForm, setReleaseForm] = useState({ flow_id: defaults.flowId || '', release_id: '', deploy_id: '', target: 'vm', channel: 'cockpit' });
  const [rollbackForm, setRollbackForm] = useState({ flow_id: defaults.flowId || '', rollback_id: '' });

  async function run(action, payload) {
    setPending(true);
    setMessage('');
    try {
      const response = await fetch(`/api/admin/runtime/founder/${action}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      const result = await response.json();
      if (!response.ok) {
        throw new Error(result?.error || `${actionLabels[action]} 실패`);
      }
      setMessage(`${actionLabels[action]} 완료`);
      router.refresh();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'founder 작업 실패');
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="admin-action-grid">
      <section className="admin-card admin-card--wide">
        <h3>지금 founder flow에서 볼 것</h3>
        <p className="section-body">{attention?.summary || '대표 액션이나 흐름 참조가 생기면 여기 먼저 뜹니다.'}</p>
        <div className="admin-list">
          <div><strong>ref</strong><span>{attention?.ref || defaults.flowId || '없음'}</span></div>
          <div><strong>waiting</strong><span>{attention?.waitingCount ?? 0}</span></div>
          <div><strong>risk</strong><span>{attention?.riskCount ?? 0}</span></div>
        </div>
        <p className="section-body">{attention?.detail || 'flow / approval / release / rollback은 아래 카드에서 바로 실행할 수 있습니다.'}</p>
      </section>

      <section className="admin-card">
        <h3>1. 흐름 열기</h3>
        <p className="section-body">새 흐름을 열고 work item과 승인 기준을 묶습니다.</p>
        <div className="admin-form-grid">
          <input value={startForm.flow_id} onChange={(event) => setStartForm((current) => ({ ...current, flow_id: event.target.value }))} placeholder="흐름 ID" disabled={!canWrite || pending} />
          <input value={startForm.work_item_id} onChange={(event) => setStartForm((current) => ({ ...current, work_item_id: event.target.value }))} placeholder="work item ID" disabled={!canWrite || pending} />
          <input value={startForm.approval_id} onChange={(event) => setStartForm((current) => ({ ...current, approval_id: event.target.value }))} placeholder="approval ID" disabled={!canWrite || pending} />
          <input value={startForm.title} onChange={(event) => setStartForm((current) => ({ ...current, title: event.target.value }))} placeholder="흐름 제목" disabled={!canWrite || pending} />
          <textarea value={startForm.intent} onChange={(event) => setStartForm((current) => ({ ...current, intent: event.target.value }))} rows={3} placeholder="이 흐름으로 무엇을 닫을지 적어두세요" disabled={!canWrite || pending} />
        </div>
        <button type="button" className="button-primary button-primary--compact" disabled={!canWrite || pending} onClick={() => run('start', startForm)}>흐름 시작</button>
      </section>

      <section className="admin-card">
        <h3>2. 승인 완료</h3>
        <p className="section-body">현재 흐름을 founder 승인 단계로 넘깁니다.</p>
        <div className="admin-form-grid">
          <input value={approveForm.flow_id} onChange={(event) => setApproveForm({ flow_id: event.target.value })} placeholder="흐름 ID" disabled={!canWrite || pending} />
        </div>
        <button type="button" className="button-primary button-primary--compact" disabled={!canWrite || pending} onClick={() => run('approve', approveForm)}>승인 완료</button>
      </section>

      <section className="admin-card">
        <h3>3. 릴리스 진행</h3>
        <p className="section-body">흐름을 실제 릴리스와 배포 정보에 연결합니다.</p>
        <div className="admin-form-grid">
          <input value={releaseForm.flow_id} onChange={(event) => setReleaseForm((current) => ({ ...current, flow_id: event.target.value }))} placeholder="흐름 ID" disabled={!canWrite || pending} />
          <input value={releaseForm.release_id} onChange={(event) => setReleaseForm((current) => ({ ...current, release_id: event.target.value }))} placeholder="릴리스 ID" disabled={!canWrite || pending} />
          <input value={releaseForm.deploy_id} onChange={(event) => setReleaseForm((current) => ({ ...current, deploy_id: event.target.value }))} placeholder="배포 ID" disabled={!canWrite || pending} />
          <input value={releaseForm.target} onChange={(event) => setReleaseForm((current) => ({ ...current, target: event.target.value }))} placeholder="대상" disabled={!canWrite || pending} />
          <input value={releaseForm.channel} onChange={(event) => setReleaseForm((current) => ({ ...current, channel: event.target.value }))} placeholder="채널" disabled={!canWrite || pending} />
        </div>
        <button type="button" className="button-primary button-primary--compact" disabled={!canWrite || pending} onClick={() => run('release', releaseForm)}>릴리스 진행</button>
      </section>

      <section className="admin-card">
        <h3>4. 롤백 기록</h3>
        <p className="section-body">문제 상황을 흐름에 다시 연결해 기록합니다.</p>
        <div className="admin-form-grid">
          <input value={rollbackForm.flow_id} onChange={(event) => setRollbackForm((current) => ({ ...current, flow_id: event.target.value }))} placeholder="흐름 ID" disabled={!canWrite || pending} />
          <input value={rollbackForm.rollback_id} onChange={(event) => setRollbackForm((current) => ({ ...current, rollback_id: event.target.value }))} placeholder="롤백 ID" disabled={!canWrite || pending} />
        </div>
        <button type="button" className="button-secondary button-secondary--compact" disabled={!canWrite || pending} onClick={() => run('rollback', rollbackForm)}>롤백 기록</button>
      </section>

      {message ? <p className="inline-action-block__message">{message}</p> : null}
    </div>
  );
}
