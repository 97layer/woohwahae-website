import { NextResponse } from 'next/server';
import { withRequestLayerOsWriteToken } from '../../../../../lib/auth/layer-os-write-token';
import { authorizeRequest } from '../../../../../lib/auth/rbac';
import { readValidatedJson } from '../../../../../lib/http/request';
import {
  getControlTowerView,
  performControlTowerAction,
} from '../../../../../lib/runtime/control-tower';
import { adminJobsActionSchema } from '../../../../../lib/validation/admin-jobs';

export async function GET(request) {
  const auth = await authorizeRequest(request, ['founder', 'admin']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  return withRequestLayerOsWriteToken(request, async () => {
    try {
      const jobId = request.nextUrl.searchParams.get('job_id') || '';
      const includePacket = request.nextUrl.searchParams.get('include_packet') === '1';
      return NextResponse.json({ jobs: await getControlTowerView(jobId, { includePacket }) });
    } catch (error) {
      return NextResponse.json({ error: error instanceof Error ? error.message : 'runtime unavailable' }, { status: 502 });
    }
  });
}

export async function POST(request) {
  const auth = await authorizeRequest(request, ['founder']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  const parsed = await readValidatedJson(request, adminJobsActionSchema);
  if (!parsed.ok) {
    return NextResponse.json({ error: parsed.error, details: parsed.details }, { status: parsed.status });
  }

  return withRequestLayerOsWriteToken(request, async () => {
    try {
      const result = await performControlTowerAction(parsed.data.action, parsed.data);
      let selectedJobId = 'job_id' in parsed.data ? parsed.data.job_id : '';
      if (!selectedJobId && parsed.data.action === 'assign') {
        selectedJobId = result?.job?.job_id || result?.job?.jobId || '';
      }
      if (!selectedJobId && (parsed.data.action === 'promote' || parsed.data.action === 'heartbeat')) {
        selectedJobId = result?.items?.find((item) => item?.job?.job_id)?.job?.job_id || '';
      }
      const jobs = await getControlTowerView(selectedJobId, { includePacket: false });
      return NextResponse.json({ ok: true, action: parsed.data.action, result, jobs });
    } catch (error) {
      const message = error instanceof Error ? error.message : 'control tower mutation failed';
      const status = message.includes('missing server write token') ? 503 : 502;
      return NextResponse.json({ error: message }, { status });
    }
  });
}
