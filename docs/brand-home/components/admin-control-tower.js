'use client';

import { useEffect, useMemo, useRef, useState } from 'react';

function toneClass(tone) {
  if (tone === 'good') return 'status-pill status-pill--good';
  if (tone === 'alert') return 'status-pill status-pill--alert';
  return 'status-pill status-pill--muted';
}

function formatStamp(value) {
  if (!value) return '--';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '--';
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  return `${month}.${day} ${hours}:${minutes}`;
}

function providerSummary(profile) {
  if (!profile?.provider && !profile?.model) return '미배정';
  return [profile.provider, profile.model].filter(Boolean).join(' / ');
}

function previewText(job) {
  return job?.responsePreview || job?.error || job?.description || 'no preview';
}

const taskRoleOptions = [
  { value: 'implementer', label: '구현' },
  { value: 'verifier', label: '검증' },
  { value: 'planner', label: '정리' },
  { value: 'designer', label: '경험' },
];

const defaultAllowedPathsByRole = {
  implementer: ['cmd/layer-osctl/', 'scripts/', 'docs/'],
  verifier: [],
  planner: [],
  designer: ['docs/', 'skills/', 'cmd/'],
};

async function readControlTower(jobId = '', includePacket = true) {
  const params = new URLSearchParams();
  if (jobId) {
    params.set('job_id', jobId);
  }
  if (!includePacket) {
    params.set('include_packet', '0');
  }
  const query = params.size > 0 ? `?${params.toString()}` : '';
  const response = await fetch(`/api/admin/runtime/jobs${query}`, {
    method: 'GET',
    cache: 'no-store',
  });
  const payload = await response.json();
  if (!response.ok) {
    throw new Error(payload?.error || 'control tower unavailable');
  }
  return payload.jobs;
}

async function readControlTowerPacket(jobId) {
  const params = new URLSearchParams();
  params.set('job_id', jobId);
  const response = await fetch(`/api/admin/runtime/jobs/packet?${params.toString()}`, {
    method: 'GET',
    cache: 'no-store',
  });
  const payload = await response.json();
  if (!response.ok) {
    throw new Error(payload?.error || 'job packet unavailable');
  }
  return payload.packet;
}

async function mutateControlTower(action, payload) {
  const response = await fetch('/api/admin/runtime/jobs', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ action, ...payload }),
  });
  const result = await response.json();
  if (!response.ok) {
    throw new Error(result?.error || 'control tower mutation failed');
  }
  return result;
}

async function readWriteTokenStatus() {
  const response = await fetch('/api/admin/runtime/token', {
    method: 'GET',
    cache: 'no-store',
  });
  const payload = await response.json();
  if (!response.ok) {
    throw new Error(payload?.error || 'write token status unavailable');
  }
  return payload;
}

async function saveWriteToken(token) {
  const response = await fetch('/api/admin/runtime/token', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token }),
  });
  const payload = await response.json();
  if (!response.ok) {
    throw new Error(payload?.error || 'write token save failed');
  }
  return payload;
}

async function clearWriteToken() {
  const response = await fetch('/api/admin/runtime/token', {
    method: 'DELETE',
  });
  const payload = await response.json();
  if (!response.ok) {
    throw new Error(payload?.error || 'write token clear failed');
  }
  return payload;
}

function syncJobIdToUrl(jobId) {
  if (typeof window === 'undefined') {
    return;
  }
  const url = new URL(window.location.href);
  if (jobId) {
    url.searchParams.set('job_id', jobId);
  } else {
    url.searchParams.delete('job_id');
  }
  window.history.replaceState({}, '', `${url.pathname}${url.search}`);
}

function WorkerLaneCard({ lane, canWrite, onFocus, onDispatch, onCancel, pending }) {
  const job = lane.job;
  const workerTone = lane.workerTone === 'good' ? 'good' : lane.workerTone === 'alert' ? 'alert' : 'muted';

  return (
    <article className="metric-card quickwork-runtime-card">
      <span className="metric-card__label">{lane.label}</span>
      <strong className="metric-card__value">{job ? job.statusLabel : '유휴'}</strong>
      <span className={toneClass(workerTone)}>{lane.workerState || 'stopped'}</span>
      <span className="metric-card__note">{providerSummary(lane.profile)}</span>
      {job ? (
        <>
          <p className="section-body">{job.summary}</p>
          <div className="inline-action-block__buttons">
            <button type="button" className="button-secondary button-secondary--compact" onClick={() => onFocus(job.jobId)}>
              detail
            </button>
            {canWrite && job.status === 'queued' ? (
              <button type="button" className="button-primary button-primary--compact" disabled={pending} onClick={() => onDispatch(job.jobId)}>
                dispatch
              </button>
            ) : null}
            {canWrite && (job.status === 'queued' || job.status === 'running') ? (
              <button type="button" className="button-secondary button-secondary--compact" disabled={pending} onClick={() => onCancel(job.jobId)}>
                cancel
              </button>
            ) : null}
          </div>
        </>
      ) : (
        <p className="section-body">현재 이 담당칸에 열린 작업이 없습니다.</p>
      )}
    </article>
  );
}

export default function AdminControlTower({ canWrite, initialData, initialJobId = '' }) {
  const [data, setData] = useState(initialData);
  const [message, setMessage] = useState('');
  const [pending, setPending] = useState(false);
  const [packetPending, setPacketPending] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [writeTokenInput, setWriteTokenInput] = useState('');
  const [writeTokenPending, setWriteTokenPending] = useState(false);
  const [writeTokenStatus, setWriteTokenStatus] = useState({
    configured: Boolean(initialData?.runtime?.writeTokenConfigured),
    source: initialData?.runtime?.writeTokenSource || 'missing',
  });
  const [selectedJobId, setSelectedJobId] = useState(initialData?.selectedJob?.jobId || initialJobId || '');
  const [taskForm, setTaskForm] = useState({
    summary: '',
    role: 'implementer',
  });
  const refreshLockRef = useRef(false);

  useEffect(() => {
    setData(initialData);
    setSelectedJobId(initialData?.selectedJob?.jobId || initialJobId || '');
    setWriteTokenStatus({
      configured: Boolean(initialData?.runtime?.writeTokenConfigured),
      source: initialData?.runtime?.writeTokenSource || 'missing',
    });
  }, [initialData, initialJobId]);

  async function refreshWriteTokenStatus() {
    try {
      const payload = await readWriteTokenStatus();
      setWriteTokenStatus({
        configured: Boolean(payload?.configured),
        source: payload?.source || 'missing',
      });
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'write token status unavailable');
    }
  }

  async function refresh(jobId = selectedJobId, options = {}) {
    if (refreshLockRef.current && !options.force) {
      return;
    }
    refreshLockRef.current = true;
    setRefreshing(true);
    try {
      const payload = await readControlTower(jobId, options.includePacket === true);
      setData((current) => {
        if (
          payload?.selectedJob
          && !payload.selectedJob.packet
          && current?.selectedJob?.jobId === payload.selectedJob.jobId
          && current.selectedJob.packet
        ) {
          return {
            ...payload,
            selectedJob: {
              ...payload.selectedJob,
              packet: current.selectedJob.packet,
            },
          };
        }
        return payload;
      });
      const nextSelectedJobId = payload?.selectedJob?.jobId || jobId || '';
      setSelectedJobId(nextSelectedJobId);
      syncJobIdToUrl(nextSelectedJobId);
      setMessage('');
      void refreshWriteTokenStatus();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'control tower unavailable');
    } finally {
      refreshLockRef.current = false;
      setRefreshing(false);
    }
  }

  useEffect(() => {
    void refreshWriteTokenStatus();
  }, []);

  useEffect(() => {
    if (data?.generatedAt || refreshLockRef.current) {
      return;
    }
    void refresh(initialJobId || selectedJobId, { force: true, includePacket: false });
  }, [data?.generatedAt, initialJobId, selectedJobId]);

  useEffect(() => {
    const interval = window.setInterval(() => {
      if (!refreshLockRef.current) {
        void refresh(selectedJobId, { includePacket: false });
      }
    }, 15000);
    return () => window.clearInterval(interval);
  }, [selectedJobId]);

  async function loadSelectedPacket(jobId = selectedJobId) {
    const normalizedJobId = String(jobId || '').trim();
    if (!normalizedJobId) {
      setMessage('packet을 읽을 작업을 먼저 선택해 주세요');
      return;
    }
    setPacketPending(true);
    try {
      const packet = await readControlTowerPacket(normalizedJobId);
      setData((current) => {
        if (current?.selectedJob?.jobId !== normalizedJobId) {
          return current;
        }
        return {
          ...current,
          selectedJob: {
            ...current.selectedJob,
            packet,
          },
        };
      });
      setMessage('작업 규칙과 다음 단계를 불러왔습니다');
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'job packet unavailable');
    } finally {
      setPacketPending(false);
    }
  }

  async function runAction(action, payload, successMessage) {
    setPending(true);
    try {
      const result = await mutateControlTower(action, payload);
      setData(result.jobs);
      const nextSelectedJobId = result.jobs?.selectedJob?.jobId || selectedJobId;
      setSelectedJobId(nextSelectedJobId);
      syncJobIdToUrl(nextSelectedJobId);
      setMessage(result?.result?.follow_up?.summary || result?.jobs?.attentionHint?.summary || successMessage);
      void refreshWriteTokenStatus();
      return result;
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'mutation failed');
      return null;
    } finally {
      setPending(false);
    }
  }

  async function submitWriteToken() {
    if (!writeTokenInput.trim()) {
      setMessage('write token을 입력해 주세요');
      return;
    }
    setWriteTokenPending(true);
    try {
      const payload = await saveWriteToken(writeTokenInput);
      setWriteTokenStatus({
        configured: Boolean(payload?.configured),
        source: payload?.source || 'cookie',
      });
      setWriteTokenInput('');
      setMessage('write token을 `/admin` runtime에 연결했습니다');
      await refresh(selectedJobId, { force: true, includePacket: false });
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'write token save failed');
    } finally {
      setWriteTokenPending(false);
    }
  }

  async function removeWriteToken() {
    setWriteTokenPending(true);
    try {
      const payload = await clearWriteToken();
      setWriteTokenStatus({
        configured: Boolean(payload?.configured),
        source: payload?.source || 'missing',
      });
      setMessage('runtime write token 연결을 해제했습니다');
      await refresh(selectedJobId, { force: true, includePacket: false });
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'write token clear failed');
    } finally {
      setWriteTokenPending(false);
    }
  }

  function updateTaskField(key, value) {
    setTaskForm((current) => ({
      ...current,
      [key]: value,
    }));
  }

  async function assignTask(event) {
    event.preventDefault();
    if (!taskForm.summary.trim()) {
      setMessage('assign할 작업 요약을 적어 주세요');
      return;
    }
    const result = await runAction(
      'assign',
      {
        summary: taskForm.summary,
        role: taskForm.role,
        allowed_paths: defaultAllowedPathsByRole[taskForm.role] || [],
        payload_json: '',
      },
      `${taskForm.role} 작업으로 배정했습니다`,
    );
    if (result) {
      setTaskForm((current) => ({ ...current, summary: '' }));
    }
  }

  async function runHeartbeat() {
    await runAction('heartbeat', { limit: 1, dispatch: true }, '다음 안건을 정리 작업으로 올렸습니다');
  }

  async function pauseSelectedJob() {
    if (!selectedJob?.jobId || !['queued', 'running'].includes(selectedJob.status)) {
      setMessage('pause할 열린 run을 먼저 선택해 주세요');
      return;
    }
    await runAction('pause', { job_id: selectedJob.jobId }, `${selectedJob.jobId} run을 일시 정지했습니다`);
  }

  function focusJob(jobId) {
    setSelectedJobId(jobId);
    syncJobIdToUrl(jobId);
    void refresh(jobId, { force: true, includePacket: false });
  }

  const selectedJob = data?.selectedJob || null;
  const warnings = useMemo(() => Array.isArray(data?.warnings) ? data.warnings.filter(Boolean) : [], [data?.warnings]);
  const loading = !data?.generatedAt;

  return (
    <section className="admin-card admin-card--wide">
      <div className="quickwork-card__head">
        <div>
          <p className="section-eyebrow">Founder operations</p>
          <h2 className="section-title">실시간 작업판</h2>
          <p className="section-body">
            지금 어떤 일이 돌아가고 있는지, 어디서 막혔는지, 다음에 무엇을 눌러야 하는지를 이 화면 하나에서 봅니다.
          </p>
        </div>
        <div className="inline-action-block__buttons">
          <span className={toneClass(data?.runtimeAvailable ? 'good' : 'alert')}>
            {data?.runtimeAvailable ? '실시간 연결' : '주의 필요'}
          </span>
          <span className={toneClass(data?.runtime?.writeReady ? 'good' : 'muted')}>
            {data?.runtime?.writeReady ? '조작 가능' : '보기만 가능'}
          </span>
          <button type="button" className="button-secondary button-secondary--compact" onClick={() => refresh(selectedJobId, { force: true, includePacket: false })} disabled={refreshing || pending}>
            {refreshing ? '불러오는 중' : '새로고침'}
          </button>
          {canWrite ? (
            <button type="button" className="button-primary button-primary--compact" onClick={() => runAction('promote', { limit: 1, dispatch: true }, 'context를 정리 작업으로 올렸습니다')} disabled={pending}>
              다음 안건 실행
            </button>
          ) : null}
        </div>
      </div>

      {loading ? (
        <div className="inline-action-block">
          <div>
            <p className="section-eyebrow">Loading</p>
            <h3 className="section-title">실시간 작업 상태를 불러오는 중입니다</h3>
            <p className="section-body">
              화면은 먼저 열고, 실제 작업 상태는 백그라운드에서 이어서 가져옵니다.
            </p>
          </div>
        </div>
      ) : null}

      <div className="inline-action-block">
        <div>
          <p className="section-eyebrow">Live actions</p>
          <h3 className="section-title">작업 배정 · 다음 안건 실행 · 일시 정지</h3>
          <p className="section-body">
            새 작업을 바로 넣고, 다음 안건을 올리고, 현재 작업을 멈출 수 있습니다. 따로 다른 화면으로 갈 필요가 없습니다.
          </p>
        </div>
        <form className="admin-form-grid admin-form-grid--inline" onSubmit={assignTask}>
          <label>
            <span>작업 요약</span>
            <input
              type="text"
              value={taskForm.summary}
              onChange={(event) => updateTaskField('summary', event.target.value)}
              placeholder="예: founder 대시보드 응답 지연 원인 정리"
            />
          </label>
          <label>
            <span>담당 레인</span>
            <select value={taskForm.role} onChange={(event) => updateTaskField('role', event.target.value)}>
              {taskRoleOptions.map((option) => (
                <option key={option.value} value={option.value}>{option.label}</option>
              ))}
            </select>
          </label>
          <div className="inline-action-block__buttons">
            <button type="submit" className="button-primary button-primary--compact" disabled={pending || !canWrite}>
              작업 배정
            </button>
            <button type="button" className="button-secondary button-secondary--compact" onClick={runHeartbeat} disabled={pending || !canWrite}>
              다음 안건 실행
            </button>
            <button
              type="button"
              className="button-secondary button-secondary--compact"
              onClick={pauseSelectedJob}
              disabled={pending || !canWrite || !selectedJob?.jobId || !['queued', 'running'].includes(selectedJob?.status)}
            >
              일시 정지
            </button>
          </div>
        </form>
      </div>

      <div className="proof-metrics proof-metrics--admin">
        <article className="metric-card metric-card--compact">
          <span className="metric-card__label">열린 작업</span>
          <strong className="metric-card__value">{data?.metrics?.open || 0}</strong>
          <span className="metric-card__note">대기 + 진행 중</span>
        </article>
        <article className="metric-card metric-card--compact">
          <span className="metric-card__label">실행 중</span>
          <strong className="metric-card__value">{data?.metrics?.running || 0}</strong>
          <span className="metric-card__note">지금 돌아가는 레인</span>
        </article>
        <article className="metric-card metric-card--compact">
          <span className="metric-card__label">실패 작업</span>
          <strong className="metric-card__value">{data?.metrics?.failed || 0}</strong>
          <span className="metric-card__note">검토가 필요한 항목</span>
        </article>
        <article className="metric-card metric-card--compact">
          <span className="metric-card__label">리뷰 안건</span>
          <strong className="metric-card__value">{data?.metrics?.reviewOpen || 0}</strong>
          <span className="metric-card__note">{data?.queueHint?.primaryAction || '다음 행동 대기 중'}</span>
        </article>
      </div>

      <div className="inline-action-block">
        <div>
          <p className="section-eyebrow">Runtime auth</p>
          <h3 className="section-title">조작 권한 연결</h3>
          <p className="section-body">
            한 번만 토큰을 연결하면 이 화면에서 바로 실행, 취소, 전송 같은 조작을 할 수 있습니다.
          </p>
        </div>
        <div className="inline-action-block__buttons">
          <span className={toneClass(writeTokenStatus.configured ? 'good' : 'alert')}>
            {writeTokenStatus.configured ? '토큰 연결됨' : '토큰 필요'}
          </span>
          <span className="metric-card__note">source: {writeTokenStatus.source || 'missing'}</span>
        </div>
        {canWrite ? (
          <div className="admin-form-grid admin-form-grid--inline">
            <label>
              <span>Layer OS write token</span>
              <input
                type="password"
                placeholder="현재 daemon write token"
                value={writeTokenInput}
                onChange={(event) => setWriteTokenInput(event.target.value)}
              />
            </label>
            <div className="inline-action-block__buttons">
              <button type="button" className="button-primary button-primary--compact" disabled={writeTokenPending || !writeTokenInput.trim()} onClick={submitWriteToken}>
                {writeTokenPending ? '저장 중' : '토큰 연결'}
              </button>
              <button type="button" className="button-secondary button-secondary--compact" disabled={writeTokenPending || !writeTokenStatus.configured} onClick={removeWriteToken}>
                연결 해제
              </button>
            </div>
          </div>
        ) : null}
      </div>

      <div className="proof-metrics proof-metrics--admin">
        {Array.isArray(data?.lanes) ? data.lanes.map((lane) => (
          <WorkerLaneCard
            key={lane.role}
            lane={lane}
            canWrite={canWrite}
            pending={pending}
            onFocus={focusJob}
            onDispatch={(jobId) => runAction('dispatch', { job_id: jobId }, `${jobId} dispatch 완료`)}
            onCancel={(jobId) => runAction('cancel', { job_id: jobId }, `${jobId} 취소 기록 완료`)}
          />
        )) : null}
      </div>

      <div className="signal-grid">
        <article className="signal-card">
          <span className="signal-card__source">{data?.attentionHint ? '보고 후속' : '지금 할 일'}</span>
          <h3 className="signal-card__title">{data?.attentionHint?.summary || data?.queueHint?.currentFocus || '지금 할 일이 아직 없습니다'}</h3>
          <p className="section-body">
            {data?.attentionHint
              ? `최근 완료된 ${data.attentionHint.role || 'lane'} 작업 ${data.attentionHint.jobId || ''}의 후속 액션입니다.`
              : data?.queueHint?.primaryRef
                ? `기준 항목: ${data.queueHint.primaryRef}`
                : '지금 바로 집어갈 기준 항목이 없습니다'}
          </p>
          {data?.attentionHint?.jobId ? (
            <div className="inline-action-block__buttons">
              <button type="button" className="button-secondary button-secondary--compact" onClick={() => focusJob(data.attentionHint.jobId)}>
                해당 작업 보기
              </button>
            </div>
          ) : null}
          {data?.attentionHint?.nextJobIds?.length ? (
            <ul className="inline-list">
              {data.attentionHint.nextJobIds.map((item) => (
                <li key={item}>
                  <button type="button" className="button-secondary button-secondary--compact" onClick={() => focusJob(item)}>
                    {item}
                  </button>
                </li>
              ))}
            </ul>
          ) : null}
          <ul className="inline-list">
            {(data?.queueHint?.nextSteps || []).map((item) => (
              <li key={item}><span>{item}</span></li>
            ))}
          </ul>
        </article>

        <article className="signal-card">
          <span className="signal-card__source">진행 중인 작업</span>
          <h3 className="signal-card__title">지금 돌아가는 작업</h3>
          <div className="admin-table">
            {(data?.openJobs || []).map((job) => (
              <article className="admin-table__row" key={job.jobId}>
                <div>
                  <button type="button" className="button-secondary button-secondary--compact" onClick={() => focusJob(job.jobId)}>
                    {job.summary}
                  </button>
                  <p>{job.jobId}</p>
                </div>
                <div>{job.role}</div>
                <div>{job.statusLabel}</div>
                <div>{formatStamp(job.updatedAt || job.createdAt)}</div>
              </article>
            ))}
            {(data?.openJobs || []).length === 0 ? <p className="section-body">열린 작업이 없습니다.</p> : null}
          </div>
        </article>

        <article className="signal-card">
          <span className="signal-card__source">최근 작업 기록</span>
          <h3 className="signal-card__title">방금 끝난 작업까지 확인</h3>
          <div className="admin-table">
            {(data?.recentJobs || []).slice(0, 8).map((job) => (
              <article className="admin-table__row" key={job.jobId}>
                <div>
                  <button type="button" className="button-secondary button-secondary--compact" onClick={() => focusJob(job.jobId)}>
                    {job.summary}
                  </button>
                  <p>{job.followUp?.summary || job.provider || job.dispatchTransport || job.source || 'runtime'}</p>
                </div>
                <div>{job.role}</div>
                <div>{job.statusLabel}</div>
                <div>{formatStamp(job.updatedAt || job.createdAt)}</div>
              </article>
            ))}
          </div>
        </article>
      </div>

      <section className="admin-card">
        <div className="quickwork-card__head">
          <div>
            <p className="section-eyebrow">선택한 작업</p>
            <h3 className="section-title">{selectedJob?.summary || '선택된 job 없음'}</h3>
          </div>
          {selectedJob ? (
            <div className="inline-action-block__buttons">
              <button
                type="button"
                className="button-secondary button-secondary--compact"
                onClick={() => loadSelectedPacket(selectedJob.jobId)}
                disabled={packetPending}
              >
                {packetPending ? '규칙 읽는 중' : selectedJob.packet ? '작업 규칙 다시 읽기' : '작업 규칙 불러오기'}
              </button>
            </div>
          ) : null}
        </div>
        {selectedJob ? (
          <>
            <div className="proof-metrics proof-metrics--admin">
              <article className="metric-card metric-card--compact">
                <span className="metric-card__label">Role / kind</span>
                <strong className="metric-card__value metric-card__value--compact">{selectedJob.role} · {selectedJob.kind}</strong>
                <span className={toneClass(selectedJob.tone)}>{selectedJob.statusLabel}</span>
              </article>
              <article className="metric-card metric-card--compact">
                <span className="metric-card__label">Dispatch</span>
                <strong className="metric-card__value metric-card__value--compact">{selectedJob.provider || selectedJob.dispatchTransport || 'n/a'}</strong>
                <span className="metric-card__note">{selectedJob.model || selectedJob.dispatchState || 'dispatch state unavailable'}</span>
              </article>
              <article className="metric-card metric-card--compact">
                <span className="metric-card__label">Updated</span>
                <strong className="metric-card__value metric-card__value--compact">{formatStamp(selectedJob.updatedAt || selectedJob.createdAt)}</strong>
                <span className="metric-card__note">{selectedJob.jobId}</span>
              </article>
              <article className="metric-card metric-card--compact">
                <span className="metric-card__label">Packet</span>
                <strong className="metric-card__value metric-card__value--compact">
                  {packetPending ? 'loading' : selectedJob.packet?.handoffMode || 'on-demand'}
                </strong>
                <span className="metric-card__note">
                  {selectedJob.packet ? selectedJob.packetPath : '필요할 때만 packet을 읽습니다'}
                </span>
              </article>
            </div>

            <div className="signal-grid">
              <article className="signal-card">
                <span className="signal-card__source">작업 내용</span>
                <h3 className="signal-card__title">{selectedJob.issueTitle || selectedJob.title || selectedJob.summary}</h3>
                <p className="section-body">{previewText(selectedJob)}</p>
                {selectedJob.error ? <p className="section-body">{selectedJob.error}</p> : null}
                {selectedJob.allowedPaths?.length ? (
                  <ul className="inline-list">
                    {selectedJob.allowedPaths.map((item) => <li key={item}><span>{item}</span></li>)}
                  </ul>
                ) : null}
              </article>

              <article className="signal-card">
                <span className="signal-card__source">작업 방식</span>
                <h3 className="signal-card__title">{selectedJob.packet?.cognitionMode || '작업 규칙 정보 없음'}</h3>
                <p className="section-body">
                  {selectedJob.packet?.directive || '초기 로드에서는 job 요약만 읽습니다. 이 작업의 규칙이 필요할 때만 위 버튼으로 packet을 불러옵니다.'}
                </p>
                {selectedJob.packet?.openQuestions?.length ? (
                  <ul className="inline-list">
                    {(selectedJob.packet?.openQuestions || []).map((item) => <li key={item}><span>{item}</span></li>)}
                  </ul>
                ) : null}
              </article>

              <article className="signal-card">
                <span className="signal-card__source">다음 액션</span>
                <h3 className="signal-card__title">{selectedJob.followUp?.summary || selectedJob.packet?.currentFocus || '다음 액션 정보 없음'}</h3>
                {selectedJob.followUp ? (
                  <>
                    <p className="section-body">
                      {selectedJob.followUp.mode || 'continue_loop'}
                    </p>
                    {selectedJob.followUp.jobIds?.length ? (
                      <ul className="inline-list">
                        {selectedJob.followUp.jobIds.map((item) => (
                          <li key={item}>
                            <button type="button" className="button-secondary button-secondary--compact" onClick={() => focusJob(item)}>
                              {item}
                            </button>
                          </li>
                        ))}
                      </ul>
                    ) : null}
                  </>
                ) : null}
                {selectedJob.packet ? (
                  <>
                    <ul className="inline-list">
                      {(selectedJob.packet?.nextSteps || []).map((item) => <li key={item}><span>{item}</span></li>)}
                    </ul>
                    <ul className="inline-list">
                      {(selectedJob.packet?.openRisks || []).map((item) => <li key={item}><span>{item}</span></li>)}
                    </ul>
                  </>
                ) : (
                  <p className="section-body">report가 남긴 다음 액션을 먼저 보고, 더 넓은 맥락이 필요할 때만 packet을 읽습니다.</p>
                )}
              </article>

              <article className="signal-card">
                <span className="signal-card__source">실행 결과</span>
                <h3 className="signal-card__title">{selectedJob.executionOrigin || selectedJob.provider || 'runtime execution'}</h3>
                <p className="section-body">
                  {selectedJob.council?.requestedCount > 0
                    ? `council ${selectedJob.council.succeededCount}/${selectedJob.council.requestedCount} · primary ${selectedJob.council.primaryProvider || selectedJob.provider || 'n/a'}`
                    : `single provider · ${selectedJob.provider || selectedJob.dispatchTransport || 'n/a'}`}
                </p>
                <ul className="inline-list">
                  {(selectedJob.changedPaths || []).slice(0, 6).map((item) => <li key={item}><span>{item}</span></li>)}
                </ul>
                <ul className="inline-list">
                  {(selectedJob.artifacts || []).slice(0, 4).map((item) => <li key={item}><span>{item}</span></li>)}
                </ul>
              </article>
            </div>
          </>
        ) : (
          <p className="section-body">job을 선택하면 packet과 실행 메타를 여기서 바로 봅니다.</p>
        )}
      </section>

      {warnings.length > 0 ? (
        <div className="inline-action-block__message">
          {warnings[0]}
        </div>
      ) : null}
      {message ? <p className="inline-action-block__message">{message}</p> : null}
    </section>
  );
}
