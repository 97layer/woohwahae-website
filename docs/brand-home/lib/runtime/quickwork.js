import 'server-only';

import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { execFile as execFileCallback } from 'node:child_process';
import { promisify } from 'node:util';

import { fetchLayerOs, safeLayerOs } from './layer-os';
import { getActiveLayerOsWriteToken } from './layer-os-token-context';

const execFile = promisify(execFileCallback);
const moduleDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(moduleDir, '../../../../');
const orchestratorPath = path.join(repoRoot, 'scripts', 'worker_orchestrator.sh');
const defaultWorkerStateDir = '/tmp/layer-os-worker-orchestrator';
const quickworkHealthTimeoutMs = 1500;
const quickworkRoleOrder = ['implementer', 'verifier', 'planner', 'designer'];

function normalizeStringList(items) {
  return Array.isArray(items)
    ? items.map((item) => String(item || '').trim()).filter(Boolean)
    : [];
}

function normalizeAllowedPaths(items) {
  return normalizeStringList(items);
}

function normalizeQuickworkRole(role) {
  switch (String(role || '').trim()) {
    case 'planner':
      return 'planner';
    case 'designer':
      return 'designer';
    case 'verifier':
      return 'verifier';
    default:
      return 'implementer';
  }
}

export function quickworkOrchestratorUpArgs(role) {
  switch (normalizeQuickworkRole(role)) {
    case 'planner':
      return ['up', '--roles', 'implementer,verifier,planner'];
    case 'designer':
      return ['up', '--roles', 'implementer,verifier,designer'];
    default:
      return ['up'];
  }
}

function parsePayloadJSON(raw) {
  const trimmed = String(raw || '').trim();
  if (!trimmed) {
    return {};
  }
  let parsed;
  try {
    parsed = JSON.parse(trimmed);
  } catch {
    throw new Error('payload_json must be valid JSON');
  }
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    throw new Error('payload_json must be a JSON object');
  }
  return parsed;
}

function defaultKindForRole(role) {
  switch (normalizeQuickworkRole(role)) {
    case 'verifier':
      return 'verify';
    case 'planner':
      return 'plan';
    case 'designer':
      return 'design';
    default:
      return 'implement';
  }
}

function defaultStageForRole(role) {
  switch (normalizeQuickworkRole(role)) {
    case 'verifier':
      return 'verify';
    case 'planner':
      return 'discover';
    case 'designer':
      return 'experience';
    default:
      return 'compose';
  }
}

async function runOrchestrator(args) {
  const { stdout } = await execFile(orchestratorPath, args, {
    cwd: repoRoot,
    env: process.env,
    maxBuffer: 1024 * 1024,
  });
  return stdout.trim();
}

function quickworkStateDir() {
  const raw = String(process.env.LAYER_OS_WORKER_STATE_DIR || '').trim();
  return raw || defaultWorkerStateDir;
}

async function readPidFile(pidFilePath) {
  try {
    const raw = await readFile(pidFilePath, 'utf8');
    const trimmed = String(raw || '').trim();
    return trimmed || '';
  } catch {
    return '';
  }
}

function pidIsRunning(pidValue) {
  const pid = Number.parseInt(String(pidValue || '').trim(), 10);
  if (!Number.isInteger(pid) || pid <= 0) {
    return false;
  }
  try {
    process.kill(pid, 0);
    return true;
  } catch {
    return false;
  }
}

async function readQuickworkBaseStatus() {
  const stateDir = quickworkStateDir();
  const daemonResult = await safeLayerOs('/healthz', { timeoutMs: quickworkHealthTimeoutMs });
  const pidEntries = await Promise.all(
    quickworkRoleOrder.map(async (role) => {
      const pid = await readPidFile(path.join(stateDir, `${role}.pid`));
      return {
        role,
        pid,
        running: pidIsRunning(pid),
      };
    }),
  );

  const raw = [
    `state_dir=${stateDir}`,
    `daemon=${daemonResult.ok ? 'reachable' : 'unreachable'}`,
    ...pidEntries.map((entry) => (entry.running ? `worker_${entry.role}=running pid=${entry.pid}` : `worker_${entry.role}=stopped`)),
  ].join('\n');

  const warnings = daemonResult.ok ? [] : [`daemon_status_unavailable: ${daemonResult.error}`];

  return {
    raw,
    warnings,
  };
}

function normalizeQuickworkAuth(authStatus, writeTokenConfigured) {
  let writeAuthEnabled = null;
  if (typeof authStatus?.write_auth_enabled === 'boolean') {
    writeAuthEnabled = authStatus.write_auth_enabled;
  } else if (typeof authStatus?.writeAuthEnabled === 'boolean') {
    writeAuthEnabled = authStatus.writeAuthEnabled;
  }

  return {
    write_auth_enabled: writeAuthEnabled,
    write_token_configured: Boolean(writeTokenConfigured),
    write_ready: writeAuthEnabled === null ? null : !writeAuthEnabled || Boolean(writeTokenConfigured),
  };
}

function normalizeQuickworkProfile(item) {
  const payload = item && typeof item === 'object' ? item : {};
  return {
    role: String(payload.role || '').trim(),
    provider: String(payload.provider || '').trim(),
    model: String(payload.model || '').trim(),
    risk: String(payload.risk || '').trim(),
    novelty: String(payload.novelty || '').trim(),
    token_class: String(payload.token_class || payload.tokenClass || '').trim(),
    token_budget: Number.parseInt(String(payload.token_budget ?? payload.tokenBudget ?? 0), 10) || 0,
    dispatch_ready: Boolean(payload.dispatch_ready ?? payload.dispatchReady),
    notes: normalizeStringList(payload.notes),
  };
}

function normalizeQuickworkJob(item) {
  const payload = item && typeof item === 'object' ? item : {};
  const result = payload.result && typeof payload.result === 'object' ? payload.result : {};
  const jobPayload = payload.payload && typeof payload.payload === 'object' ? payload.payload : {};
  const followUp = result.follow_up && typeof result.follow_up === 'object' ? result.follow_up : {};

  return {
    job_id: String(payload.job_id || payload.jobId || '').trim(),
    kind: String(payload.kind || '').trim(),
    role: String(payload.role || '').trim(),
    stage: String(payload.stage || '').trim(),
    source: String(payload.source || '').trim(),
    status: String(payload.status || '').trim(),
    summary: String(payload.summary || '').trim(),
    updated_at: String(payload.updated_at || payload.updatedAt || '').trim(),
    dispatch_state: String(result.dispatch_state || result.dispatchState || '').trim(),
    dispatch_transport: String(result.dispatch_transport || result.dispatchTransport || '').trim(),
    allowed_paths: normalizeAllowedPaths(jobPayload.allowed_paths || jobPayload.allowedPaths),
    follow_up: {
      mode: String(followUp.mode || '').trim(),
      summary: String(followUp.summary || '').trim(),
      job_ids: normalizeStringList(followUp.job_ids || followUp.jobIds),
    },
  };
}

function buildQuickworkAttention(status) {
  if (status.write_auth_enabled && !status.write_ready) {
    return {
      mode: 'restore_write',
      summary: '쓰기 토큰을 먼저 준비해야 quickwork를 실제로 실행할 수 있습니다.',
      detail: 'founder write 세션이나 LAYER_OS_WRITE_TOKEN이 준비되지 않으면 submit이 막힙니다.',
      action_label: '쓰기 세션 확인',
      recommended_role: '',
      ref: '',
      open_count: status.open_jobs.length,
    };
  }

  if (status.daemon !== 'reachable') {
    return {
      mode: 'restore_runtime',
      summary: 'daemon 연결을 먼저 복구한 뒤 quickwork를 여는 편이 안전합니다.',
      detail: 'layer-osd가 unreachable이면 새 job create/dispatch가 안정적으로 이어지지 않습니다.',
      action_label: 'runtime 새로고침',
      recommended_role: '',
      ref: '',
      open_count: status.open_jobs.length,
    };
  }

  const openJob = status.open_jobs.find((item) => item.follow_up?.summary) || status.open_jobs[0] || null;
  if (openJob) {
    return {
      mode: 'review_open_job',
      summary: openJob.follow_up?.summary || `${openJob.role || 'lane'} 작업 ${openJob.job_id}가 이미 열려 있습니다.`,
      detail: `열린 작업 ${status.open_jobs.length}건 중 ${openJob.job_id}를 먼저 확인하는 편이 새 quickwork보다 낫습니다.`,
      action_label: 'jobs 보기',
      recommended_role: openJob.role || '',
      ref: openJob.job_id,
      open_count: status.open_jobs.length,
    };
  }

  const plannerProfile = status.dispatch_profiles.find((item) => item.dispatch_ready && item.role === 'planner') || null;
  if (plannerProfile) {
    return {
      mode: 'seed_planner',
      summary: '열린 작업이 없으니 planner lane으로 다음 일을 먼저 정리하세요.',
      detail: '현재 harness는 planner가 다음 액션을 짧게 정리한 뒤 implementer / verifier로 넘기는 흐름과 잘 맞습니다.',
      action_label: '정리 작업 프리셋 채우기',
      recommended_role: 'planner',
      ref: '',
      open_count: 0,
    };
  }

  const implementerProfile = status.dispatch_profiles.find((item) => item.dispatch_ready && item.role === 'implementer') || null;
  return {
    mode: 'seed_implementer',
    summary: '바로 구현 작업을 열 수 있습니다.',
    detail: implementerProfile
      ? `${implementerProfile.provider || 'provider'} / ${implementerProfile.model || 'model'} binding이 준비된 상태입니다.`
      : 'dispatch profile이 충분히 읽히지 않아도 quickwork는 기본 구현 레인으로 열 수 있습니다.',
    action_label: '백엔드 구현 프리셋 채우기',
    recommended_role: 'implementer',
    ref: '',
    open_count: 0,
  };
}

function buildQuickworkFollowUp(job, dispatch) {
  const jobId = String(job?.job_id || job?.jobId || '').trim();
  const role = String(job?.role || '').trim();
  const status = String(dispatch?.job?.status || job?.status || 'queued').trim();
  const dispatchState = String(dispatch?.job?.result?.dispatch_state || dispatch?.job?.result?.dispatchState || 'dispatched').trim();
  return {
    mode: 'monitor_job',
    summary: `${role || 'quickwork'} 작업 ${jobId || ''}을 열었습니다. jobs 탭에서 dispatch 상태를 먼저 확인하세요.`.trim(),
    detail: `status ${status}${dispatchState ? ` · ${dispatchState}` : ''}`,
    action_label: 'jobs 보기',
    recommended_role: role,
    ref: jobId,
    open_count: 1,
  };
}

export function buildQuickworkRuntimeStatus({
  raw,
  authStatus = null,
  profiles = [],
  jobs = [],
  writeTokenConfigured = false,
  writeTokenSource = 'missing',
  warnings = [],
}) {
  const auth = normalizeQuickworkAuth(authStatus, writeTokenConfigured);
  const status = {
    raw: String(raw || '').trim(),
    daemon: 'unknown',
    state_dir: '',
    workers: {},
    auth,
    write_token_configured: auth.write_token_configured,
    write_token_source: writeTokenSource,
    write_auth_enabled: auth.write_auth_enabled,
    write_ready: auth.write_ready,
    dispatch_profiles: Array.isArray(profiles) ? profiles.map(normalizeQuickworkProfile) : [],
    open_jobs: Array.isArray(jobs) ? jobs.map(normalizeQuickworkJob) : [],
    warnings: normalizeStringList(warnings),
  };

  for (const line of status.raw.split('\n')) {
    const trimmed = line.trim();
    if (!trimmed) continue;
    if (trimmed.startsWith('state_dir=')) {
      status.state_dir = trimmed.slice('state_dir='.length);
      continue;
    }
    if (trimmed.startsWith('daemon=')) {
      status.daemon = trimmed.slice('daemon='.length);
      continue;
    }
    if (trimmed.startsWith('worker_')) {
      const [key, value] = trimmed.split('=', 2);
      const role = key.replace(/^worker_/, '');
      status.workers[role] = value || 'unknown';
    }
  }

  status.attention = buildQuickworkAttention(status);
  return status;
}

async function readQuickworkRuntimeExtras(status) {
  if (status.daemon !== 'reachable') {
    return {
      authStatus: null,
      profiles: [],
      jobs: [],
      warnings: [],
    };
  }

  const warnings = [];
  const [authResult, profilesResult, jobsResult] = await Promise.all([
    safeLayerOs('/api/layer-os/auth', { timeoutMs: 4000 }),
    safeLayerOs('/api/layer-os/jobs/profiles', { timeoutMs: 4000 }),
    safeLayerOs('/api/layer-os/jobs?status=open&limit=6', { timeoutMs: 4000 }),
  ]);

  const authStatus = authResult.ok ? authResult.payload?.auth || authResult.payload : null;
  const profiles = profilesResult.ok && Array.isArray(profilesResult.payload?.items) ? profilesResult.payload.items : [];
  const jobs = jobsResult.ok && Array.isArray(jobsResult.payload?.items) ? jobsResult.payload.items : [];

  if (!authResult.ok) {
    warnings.push(`auth_status_unavailable: ${authResult.error}`);
  }
  if (!profilesResult.ok) {
    warnings.push(`dispatch_profiles_unavailable: ${profilesResult.error}`);
  }
  if (!jobsResult.ok) {
    warnings.push(`open_jobs_unavailable: ${jobsResult.error}`);
  }

  return {
    authStatus,
    profiles,
    jobs,
    warnings,
  };
}

export async function getQuickworkRuntimeStatus() {
  const writeToken = getActiveLayerOsWriteToken();
  const writeTokenConfigured = Boolean(writeToken.token);
  const base = await readQuickworkBaseStatus();
  const baseStatus = buildQuickworkRuntimeStatus({ raw: base.raw, writeTokenConfigured, warnings: base.warnings });
  const extras = await readQuickworkRuntimeExtras(baseStatus);
  return buildQuickworkRuntimeStatus({
    raw: base.raw,
    authStatus: extras.authStatus,
    profiles: extras.profiles,
    jobs: extras.jobs,
    writeTokenConfigured,
    writeTokenSource: writeToken.source,
    warnings: [...base.warnings, ...extras.warnings],
  });
}

export async function submitQuickwork(input) {
  const role = normalizeQuickworkRole(input.role);
  const allowedPaths = normalizeAllowedPaths(input.allowed_paths);
  const payload = parsePayloadJSON(input.payload_json);

  await runOrchestrator(quickworkOrchestratorUpArgs(role));

  const created = await fetchLayerOs('/api/layer-os/jobs', {
    method: 'POST',
    requireWriteToken: true,
    json: {
      kind: defaultKindForRole(role),
      role,
      summary: String(input.summary || '').trim(),
      status: 'queued',
      source: 'founder.manual',
      surface: 'api',
      stage: defaultStageForRole(role),
      payload: {
        ...(allowedPaths.length > 0 ? { allowed_paths: allowedPaths } : {}),
        ...payload,
      },
    },
  });

  const job = created.job || created;
  const jobId = job.job_id || job.jobId;
  if (!jobId) {
    throw new Error('quickwork did not return a job id');
  }

  const dispatched = await fetchLayerOs('/api/layer-os/jobs/dispatch', {
    method: 'POST',
    requireWriteToken: true,
    json: { job_id: jobId },
  });

  return {
    ok: true,
    job,
    dispatch: dispatched,
    follow_up: buildQuickworkFollowUp(job, dispatched),
    runtime: await getQuickworkRuntimeStatus(),
  };
}
