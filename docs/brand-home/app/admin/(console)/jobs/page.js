import AdminControlTower from '../../../../components/admin-control-tower';
import AdminQuickworkActions from '../../../../components/admin-quickwork-actions';
import { getRequestSessionMeta } from '../../../../lib/auth/request-session';
import { getAdminOverviewView } from '../../../../lib/runtime/view-model';

export default async function AdminJobsPage({ searchParams }) {
  const params = await searchParams;
  const jobId = typeof params?.job_id === 'string' ? params.job_id : '';
  const [session, overview] = await Promise.all([
    getRequestSessionMeta(),
    getAdminOverviewView(),
  ]);
  const founderAttention = overview.founderAttention || null;

  return (
    <div className="admin-page-grid">
      <section className="admin-card admin-card--wide">
        <p className="section-eyebrow">과업 현황</p>
        <h2 className="section-title">과업 배정 · 진행 확인</h2>
        <p className="section-body">
          과업 배정, 다음 안건 실행, 진행 상황 확인을 한 번에 합니다.
          이 화면이 Layer OS 실무 운영 중심입니다.
        </p>
        <p className="section-body">
          첫 화면은 바로 열리고, 실시간 작업 상태는 아래 작업판이 즉시 이어서 불러옵니다.
        </p>
        <p className="section-body">
          {founderAttention?.summary || overview.jobs.attentionHint?.summary || '지금 가장 먼저 볼 후속 액션은 작업판에서 바로 이어집니다.'}
        </p>
      </section>

      <AdminControlTower canWrite={session.canWrite} initialData={null} initialJobId={jobId} />
      <AdminQuickworkActions canWrite={session.canWrite} compact />
    </div>
  );
}
