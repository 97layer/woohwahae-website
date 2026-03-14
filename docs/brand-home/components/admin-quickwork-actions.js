'use client';

import Link from 'next/link';
import { useEffect, useMemo, useState } from 'react';
import { useRouter } from 'next/navigation';

const roleOptions = [
  {
    value: 'implementer',
    label: '구현 작업',
    summary: '코드 수정과 테스트까지 맡는 기본 작업자입니다.',
  },
  {
    value: 'verifier',
    label: '검증 작업',
    summary: '기본 read-only로 확인, 재현, 검증에 집중합니다.',
  },
  {
    value: 'planner',
    label: '정리 작업',
    summary: '분석/오케스트레이션 중심으로 방향을 정리합니다.',
  },
  {
    value: 'designer',
    label: '경험 점검',
    summary: 'IA, visual fit, experience 검토를 같은 runtime에서 엽니다.',
  },
];

const defaultAllowedPathsByRole = {
  implementer: 'cmd/layer-osctl/,scripts/,docs/',
  verifier: '',
  planner: '',
  designer: 'docs/,skills/,cmd/',
};

const quickworkPresets = [
  {
    id: 'backend-ship',
    label: '백엔드 구현',
    note: '코드 변경 + 테스트까지 바로 태웁니다.',
    summary: '백엔드 로직 개선',
    role: 'implementer',
    allowedPaths: defaultAllowedPathsByRole.implementer,
    payloadJson: '',
  },
  {
    id: 'verify-lane',
    label: '검증 돌리기',
    note: '최근 변경 확인, 회귀 체크, 로그 검토에 맞습니다.',
    summary: '최근 변경 검증',
    role: 'verifier',
    allowedPaths: '',
    payloadJson: '',
  },
  {
    id: 'plan-next',
    label: '다음 일 정리',
    note: '다음 액션과 작업 순서를 짧게 정리합니다.',
    summary: '다음 작업 우선순위 정리',
    role: 'planner',
    allowedPaths: '',
    payloadJson: '',
  },
  {
    id: 'experience-review',
    label: '경험 검토',
    note: 'admin/public surface와 작업 흐름의 경험 적합성을 봅니다.',
    summary: '운영 경험과 surface 일관성 검토',
    role: 'designer',
    allowedPaths: defaultAllowedPathsByRole.designer,
    payloadJson: '',
  },
];

function normalizeWorkerState(value) {
  if (!value) return { tone: 'muted', label: '확인 중' };
  if (value.startsWith('running')) {
    return { tone: 'good', label: '가동 중', detail: value.replace(/^running\s*/, '') };
  }
  if (value === 'stopped') {
    return { tone: 'muted', label: '대기 중', detail: '필요하면 자동으로 깨웁니다.' };
  }
  return { tone: 'muted', label: value, detail: '' };
}

function normalizeDaemonState(value) {
  if (value === 'reachable') {
    return { tone: 'good', label: '연결됨', detail: 'daemon이 응답 중입니다.' };
  }
  if (value === 'unreachable') {
    return { tone: 'alert', label: '깨우는 중', detail: 'submit 시 worker orchestrator가 다시 올립니다.' };
  }
  return { tone: 'muted', label: '확인 중', detail: 'runtime 상태를 읽는 중입니다.' };
}

function makeStatusToneClass(tone) {
  if (tone === 'good') return 'status-pill status-pill--good';
  if (tone === 'alert') return 'status-pill status-pill--alert';
  return 'status-pill status-pill--muted';
}

function presetIdForRole(role) {
  switch (String(role || '').trim()) {
    case 'planner':
      return 'plan-next';
    case 'verifier':
      return 'verify-lane';
    case 'designer':
      return 'experience-review';
    default:
      return 'backend-ship';
  }
}

async function readRuntimeStatus() {
  const response = await fetch('/api/admin/runtime/quickwork', {
    method: 'GET',
    cache: 'no-store',
  });
  const payload = await response.json();
  if (!response.ok) {
    throw new Error(payload?.error || 'runtime status failed');
  }
  return payload;
}

function withRoleDefaults(current, nextRole) {
  const currentDefault = defaultAllowedPathsByRole[current.role] || '';
  const nextDefault = defaultAllowedPathsByRole[nextRole] || '';
  const allowedPaths = current.allowedPaths.trim() === currentDefault.trim()
    ? nextDefault
    : current.allowedPaths;
  return {
    ...current,
    role: nextRole,
    allowedPaths,
  };
}

export default function AdminQuickworkActions({ canWrite, compact = false }) {
  const router = useRouter();
  const [pending, setPending] = useState(false);
  const [statusPending, setStatusPending] = useState(false);
  const [message, setMessage] = useState('');
  const [result, setResult] = useState(null);
  const [runtime, setRuntime] = useState(null);
  const [form, setForm] = useState({
    summary: '',
    role: 'implementer',
    allowedPaths: defaultAllowedPathsByRole.implementer,
    payloadJson: '',
  });

  const helperText = useMemo(() => {
    if (form.role === 'verifier') {
      return '검증 작업은 기본 read-only입니다. 범위를 더 좁히고 싶을 때만 allowed paths를 적으세요.';
    }
    if (form.role === 'planner') {
      return '정리 작업은 보통 summary만으로 충분합니다. payload는 정말 필요한 경우에만 여세요.';
    }
    return '구현 작업은 기본적으로 cmd/layer-osctl/, scripts/, docs/ 안에서만 빠르게 손대도록 제한됩니다.';
  }, [form.role]);

  const daemonState = normalizeDaemonState(runtime?.runtime?.daemon || runtime?.daemon);
  const attention = result?.follow_up || runtime?.attention || null;
  const workerEntries = useMemo(() => {
    const workers = runtime?.runtime?.workers || runtime?.workers || {};
    return ['implementer', 'verifier', 'planner', 'designer'].map((role) => ({
      role,
      ...normalizeWorkerState(workers[role] || 'stopped'),
    }));
  }, [runtime]);
  const recommendedPreset = useMemo(
    () => quickworkPresets.find((preset) => preset.id === presetIdForRole(attention?.recommended_role)),
    [attention],
  );

  async function refreshStatus() {
    setStatusPending(true);
    try {
      const payload = await readRuntimeStatus();
      setRuntime(payload);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'runtime status failed');
    } finally {
      setStatusPending(false);
    }
  }

  useEffect(() => {
    refreshStatus();
  }, []);

  function updateField(key, value) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  function applyPreset(preset) {
    setForm({
      summary: preset.summary,
      role: preset.role,
      allowedPaths: preset.allowedPaths,
      payloadJson: preset.payloadJson,
    });
    setMessage(`${preset.label} 프리셋을 채웠습니다. summary만 다듬고 submit 하세요.`);
    setResult(null);
  }

  async function submit(event) {
    event.preventDefault();
    setPending(true);
    setMessage('');
    setResult(null);
    try {
      const response = await fetch('/api/admin/runtime/quickwork', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          summary: form.summary,
          role: form.role,
          allowed_paths: form.allowedPaths
            .split(',')
            .map((item) => item.trim())
            .filter(Boolean),
          payload_json: form.payloadJson,
        }),
      });
      const payload = await response.json();
      if (!response.ok) {
        throw new Error(payload?.error || 'quickwork failed');
      }
      setRuntime(payload.runtime || null);
      setResult(payload);
      setMessage(payload?.follow_up?.summary || '빠른 실행 제출 완료');
      setForm((current) => ({ ...current, summary: '', payloadJson: '' }));
      router.refresh();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'quickwork failed');
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="admin-action-grid">
      <section className={`admin-card ${compact ? '' : 'admin-card--wide'}`}>
        <div className="quickwork-card__head">
          <div>
            <p className="section-eyebrow">Quick work</p>
            <h2 className="section-title">터미널 없이 바로 작업 열기</h2>
          </div>
          <div className="inline-action-block__buttons">
            <button
              type="button"
              className="button-secondary button-secondary--compact"
              onClick={refreshStatus}
              disabled={statusPending}
            >
              {statusPending ? '상태 읽는 중…' : 'runtime 새로고침'}
            </button>
            <Link href="/admin/jobs" className="button-secondary button-secondary--compact">
              jobs 보기
            </Link>
          </div>
        </div>

        <p className="section-body">
          summary만 적으면 작업을 만들고 canonical job create + dispatch까지 이어집니다.
          {canWrite ? ' 자주 쓰는 작업은 아래 프리셋으로 더 빨리 열 수 있습니다.' : ' founder write 세션에서만 submit 가능합니다.'}
        </p>

        <div className="signal-grid">
          <article className="signal-card">
            <span className="signal-card__source">다음 액션</span>
            <h3 className="signal-card__title">{attention?.summary || '빠른 실행은 한 번에 하나의 lane만 여는 것이 좋습니다.'}</h3>
            <p className="section-body">
              {attention?.detail || '이미 열린 작업이 있으면 그 흐름을 먼저 보고, 비어 있으면 planner나 implementer 프리셋으로 새 quickwork를 여는 것이 현재 스윗스팟입니다.'}
            </p>
            {attention?.ref ? <p className="section-body">ref · {attention.ref}</p> : null}
            <div className="inline-action-block__buttons">
              {attention?.mode === 'review_open_job' ? (
                <Link href="/admin/jobs" className="button-secondary button-secondary--compact">
                  {attention?.action_label || 'jobs 보기'}
                </Link>
              ) : recommendedPreset ? (
                <button
                  type="button"
                  className="button-primary button-primary--compact"
                  onClick={() => applyPreset(recommendedPreset)}
                  disabled={pending}
                >
                  {attention?.action_label || `${recommendedPreset.label} 프리셋 채우기`}
                </button>
              ) : null}
            </div>
          </article>
        </div>

        <div className="proof-metrics proof-metrics--admin">
          <article className="metric-card quickwork-runtime-card">
            <span className="metric-card__label">Daemon</span>
            <strong className="metric-card__value">{daemonState.label}</strong>
            <span className={makeStatusToneClass(daemonState.tone)}>{runtime?.write_token_configured ? 'write token ready' : 'write token check'}</span>
            <span className="metric-card__note">{daemonState.detail}</span>
          </article>
          {workerEntries.map((worker) => (
            <article className="metric-card quickwork-runtime-card" key={worker.role}>
              <span className="metric-card__label">{worker.role}</span>
              <strong className="metric-card__value">{worker.label}</strong>
              <span className={makeStatusToneClass(worker.tone)}>{worker.role}</span>
              <span className="metric-card__note">{worker.detail || 'submit 하면 자동으로 이 담당 작업을 사용합니다.'}</span>
            </article>
          ))}
        </div>

        <div className="quickwork-preset-grid">
          {quickworkPresets.map((preset) => (
            <button
              key={preset.id}
              type="button"
              className="quickwork-preset-card"
              onClick={() => applyPreset(preset)}
              disabled={pending}
            >
              <span className="quickwork-preset-card__title">{preset.label}</span>
              <span className="quickwork-preset-card__note">{preset.note}</span>
            </button>
          ))}
        </div>

        <form className="admin-form-grid" onSubmit={submit}>
          <label>
            <span>무엇을 시킬까요?</span>
            <textarea
              value={form.summary}
              onChange={(event) => updateField('summary', event.target.value)}
              rows={3}
              placeholder="예: 백엔드 응답 지연 원인 정리"
              disabled={!canWrite || pending}
            />
          </label>

          <div className="quickwork-role-grid">
            {roleOptions.map((option) => (
              <button
                key={option.value}
                type="button"
                className={`quickwork-role-card ${form.role === option.value ? 'is-active' : ''}`}
                onClick={() => setForm((current) => withRoleDefaults(current, option.value))}
                disabled={!canWrite || pending}
              >
                <strong>{option.label}</strong>
                <span>{option.summary}</span>
              </button>
            ))}
          </div>

          <p className="section-body">{helperText}</p>

          <details className="quickwork-advanced">
            <summary>고급 설정 열기</summary>
            <div className="admin-form-grid">
              <label>
                <span>allowed paths</span>
                <input
                  value={form.allowedPaths}
                  onChange={(event) => updateField('allowedPaths', event.target.value)}
                  placeholder="cmd/layer-osctl/,scripts/,docs/"
                  disabled={!canWrite || pending}
                />
              </label>
              <label>
                <span>payload JSON (optional)</span>
                <textarea
                  value={form.payloadJson}
                  onChange={(event) => updateField('payloadJson', event.target.value)}
                  rows={4}
                  placeholder='{"chain_rules":{"rules":[]}}'
                  disabled={!canWrite || pending}
                />
              </label>
            </div>
          </details>

          <div className="inline-action-block__buttons">
            <button type="submit" className="button-primary button-primary--compact" disabled={!canWrite || pending || !form.summary.trim()}>
              {pending ? '작업 여는 중…' : '지금 던지기'}
            </button>
          </div>
        </form>

        {message ? <p className="inline-action-block__message">{message}</p> : null}

        {result ? (
          <div className="quickwork-result-card">
            <div className="admin-list">
              <div><strong>job</strong><span>{result.job?.job_id || result.job?.jobId || 'n/a'}</span></div>
              <div><strong>status</strong><span>{result.dispatch?.job?.status || result.job?.status || 'queued'}</span></div>
              <div><strong>dispatch</strong><span>{result.dispatch?.job?.result?.dispatch_state || 'dispatched'}</span></div>
              <div><strong>daemon</strong><span>{result.runtime?.daemon || result.runtime?.runtime?.daemon || 'unknown'}</span></div>
            </div>
            <p className="section-body">{result?.follow_up?.summary || '제출은 끝났습니다. 진행 상태는 Jobs 탭에서 계속 볼 수 있습니다.'}</p>
          </div>
        ) : null}
      </section>
    </div>
  );
}
