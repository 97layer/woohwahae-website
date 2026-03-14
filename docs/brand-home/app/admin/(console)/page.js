import Link from 'next/link';
import { Suspense } from 'react';

import AdminQuickworkActions from '../../../components/admin-quickwork-actions';
import AdminTelegramActions from '../../../components/admin-telegram-actions';
import AdminFounderActions from '../../../components/admin-founder-actions';
import AdminBrandPublishActions from '../../../components/admin-brand-publish-actions';
import AdminSourceIntakeActions from '../../../components/admin-source-intake-actions';
import { formatRuntimeTimestamp } from '../../../components/neural-graph-time.mjs';
import { getRequestSessionMeta } from '../../../lib/auth/request-session';
import { getAdminOverviewView } from '../../../lib/runtime/view-model';
import { getAdminBrandPublishView } from '../../../lib/runtime/brand-publish';
import { getAdminSourceIntakeView } from '../../../lib/runtime/source-intake';

function statusToneClass(status) {
  if (status === 'ready' || status === 'reachable' || status === 'succeeded' || status === 'passed' || status === 'ok' || status === 'operational') {
    return 'status-pill status-pill--good';
  }
  if (status === 'degraded' || status === 'failed' || status === 'missing' || status === 'disabled') {
    return 'status-pill status-pill--alert';
  }
  return 'status-pill status-pill--muted';
}

function formatAdminTime(value) {
  return value ? formatRuntimeTimestamp(value) : '--:--:-- KST';
}

function formatDuration(seconds) {
  if (!Number.isFinite(seconds) || seconds <= 0) {
    return '미기록';
  }
  if (seconds < 60) {
    return `${seconds}초`;
  }
  const minutes = Math.floor(seconds / 60);
  const remainder = seconds % 60;
  return remainder > 0 ? `${minutes}분 ${remainder}초` : `${minutes}분`;
}

function humanizeStatus(value) {
  switch (value) {
    case 'ready':
    case 'reachable':
    case 'succeeded':
    case 'passed':
    case 'ok':
    case 'operational':
      return '정상';
    case 'degraded':
      return '주의';
    case 'failed':
      return '실패';
    case 'missing':
      return '미설정';
    case 'disabled':
      return '꺼짐';
    case 'pending':
      return '대기';
    case 'queued':
      return '대기';
    case 'running':
      return '진행 중';
    case 'unknown':
    case '':
    case null:
    case undefined:
      return '확인 필요';
    default:
      return String(value).replaceAll('_', ' ');
  }
}

function renderInlineItems(items, emptyText) {
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

function renderDeployTimelineRow(kind, item) {
  const kindMeta = {
    release: {
      label: '최근 릴리스',
      empty: '릴리스 기록이 아직 없습니다.',
    },
    deploy: {
      label: '최근 배포',
      empty: '배포 기록이 아직 없습니다.',
    },
    rollback: {
      label: '최근 롤백',
      empty: '롤백 기록이 아직 없습니다.',
    },
  }[kind];

  if (!item) {
    return (
      <article className="admin-table__row" key={kind}>
        <div>
          <strong>{kindMeta.label}</strong>
          <p>{kindMeta.empty}</p>
        </div>
        <div>--</div>
        <div>--</div>
        <div>--:--:-- KST</div>
      </article>
    );
  }

  if (kind === 'release') {
    return (
      <article className="admin-table__row" key={kind}>
        <div>
          <strong>{item.releaseId}</strong>
          <p>{`${item.target} · ${item.channel}`}</p>
        </div>
        <div>{`${item.approvalCount}개 승인`}</div>
        <div>{`${item.artifactCount}개 산출물`}</div>
        <div>{formatAdminTime(item.releasedAt)}</div>
      </article>
    );
  }

  if (kind === 'deploy') {
    return (
      <article className="admin-table__row" key={kind}>
        <div>
          <strong>{item.deployId}</strong>
          <p>{item.releaseId}</p>
        </div>
        <div>{item.target}</div>
        <div>{humanizeStatus(item.status)}</div>
        <div>{formatAdminTime(item.finishedAt || item.startedAt)}</div>
      </article>
    );
  }

  return (
    <article className="admin-table__row" key={kind}>
      <div>
        <strong>{item.rollbackId}</strong>
        <p>{item.releaseId}</p>
      </div>
      <div>{item.target}</div>
      <div>{humanizeStatus(item.status)}</div>
      <div>{formatAdminTime(item.finishedAt || item.startedAt)}</div>
    </article>
  );
}

function humanizeRole(role) {
  switch (role) {
    case 'implementer': return '구현';
    case 'verifier': return '검증';
    case 'planner': return '기획';
    case 'designer': return '설계';
    case 'architect': return '설계사';
    default: return role || '미배정';
  }
}

function FounderPipelineBar({ signalCount, jobCount, reviewCount }) {
  return (
    <div className="founder-pipeline">
      <div className="founder-pipeline__step">
        <strong className="founder-pipeline__count">{signalCount ?? 0}</strong>
        <span className="founder-pipeline__label">포착된 신호</span>
      </div>
      <span className="founder-pipeline__arrow" aria-hidden="true">›</span>
      <div className="founder-pipeline__step">
        <strong className="founder-pipeline__count">{jobCount ?? 0}</strong>
        <span className="founder-pipeline__label">처리 중인 과업</span>
      </div>
      <span className="founder-pipeline__arrow" aria-hidden="true">›</span>
      <div className={`founder-pipeline__step${reviewCount > 0 ? ' founder-pipeline__step--action' : ''}`}>
        <strong className="founder-pipeline__count">{reviewCount ?? 0}</strong>
        <span className="founder-pipeline__label">결재 대기</span>
      </div>
    </div>
  );
}

function renderJumpLink(item) {
  const content = (
    <>
      <strong className="admin-jump-card__title">{item.label}</strong>
      <p className="admin-jump-card__note">{item.note}</p>
      {item.meta ? <span className="admin-jump-card__meta">{item.meta}</span> : null}
    </>
  );

  if (item.external) {
    return (
      <Link key={item.href} href={item.href} className="admin-jump-card">
        {content}
      </Link>
    );
  }

  return (
    <a key={item.href} href={item.href} className="admin-jump-card">
      {content}
    </a>
  );
}

function LoadingLaneCard({ eyebrow, title, body, anchorId }) {
  return (
    <div id={anchorId} className="admin-card admin-card--wide">
      <p className="section-eyebrow">{eyebrow}</p>
      <h2 className="section-title">{title}</h2>
      <p className="section-body">{body}</p>
      <p className="section-body">운영 카드를 백그라운드에서 이어서 불러오는 중입니다.</p>
    </div>
  );
}

async function AdminSourceIntakeLane({ canWrite }) {
  const sourceIntake = await getAdminSourceIntakeView();
  return (
    <div id="source-intake" className="admin-card--wide">
      <AdminSourceIntakeActions canWrite={canWrite} initialView={sourceIntake} />
    </div>
  );
}

async function AdminBrandPublishLane({ canWrite }) {
  const brandPublish = await getAdminBrandPublishView();
  return (
    <div id="brand-publish" className="admin-card--wide">
      <AdminBrandPublishActions canWrite={canWrite} initialView={brandPublish} />
    </div>
  );
}

export default async function AdminOverviewPage() {
  const [overview, session] = await Promise.all([
    getAdminOverviewView(),
    getRequestSessionMeta(),
  ]);
  const canWrite = session.canWrite;
  const jobCounts = overview.jobCounts || {};
  const founderAttention = overview.founderAttention || null;
  const primaryAttention = overview.primaryAttention || null;
  const founderFlowDefaults = overview.founderFlowDefaults || {};
  const sourceIntake = overview.sourceIntake || null;
  const continuity = overview.deployment.continuityHost;
  const daemon = overview.deployment.daemon;
  const latestVerification = overview.latestVerification || null;
  const runtimeDegraded = !overview.runtimeAvailable;
  const runtimeDegradedReason = overview.degradedReason || '';
  const jumpLinks = [
    {
      href: '#continuity',
      label: '배포와 런타임',
      note: continuity.host || 'VM 호스트 연결 정보를 먼저 확인하세요.',
      meta: continuity.latestDeployId ? `최근 배포 ${continuity.latestDeployId}` : '최근 배포 기록 없음',
    },
    {
      href: '#quickwork',
      label: '즉시 실행',
      note: primaryAttention?.summary || jobCounts.summaryNote || '바로 처리할 작업을 열 수 있습니다.',
      meta: jobCounts.summaryMeta || `열린 작업 ${overview.jobs.openCount}건`,
    },
    {
      href: '#source-intake',
      label: '포착된 신호',
      note: sourceIntake?.summaryNote || '링크·메모·아이디어를 쌓는 입구입니다.',
      meta: sourceIntake?.summaryMeta || '스트리밍으로 이어서 로드',
    },
    {
      href: '#brand-publish',
      label: '콘텐츠 발행',
      note: '초안부터 발행 대기까지 이어집니다.',
      meta: '스트리밍으로 이어서 로드',
    },
    {
      href: '#telegram',
      label: '텔레그램',
      note: 'founder / 운영 / 브랜드 라우트를 한 번에 확인합니다.',
      meta: canWrite ? '즉시 전송 가능' : '읽기 전용 세션',
    },
    {
      href: '/admin/review-room',
      label: '결재 대기 안건',
      note: overview.reviewNote || '창업자 최종 판단이 필요한 안건입니다.',
      meta: overview.reviewMeta || `결재 대기 ${overview.review.openCount}건`,
      external: true,
    },
  ];

  return (
    <div className="admin-page-grid">
      <section className="admin-card admin-card--wide admin-hero-card">
        <div className="admin-hero-card__head">
          <div>
            <p className="section-eyebrow">Founder operations</p>
            <h2 className="section-title">서비스 운영 한눈판</h2>
            <p className="section-body">
              Layer OS snapshot 하나를 기준으로 서비스 상태, 검토 안건, 작업 진행, 소스 인입, 텔레그램, Threads 발행 준비까지 한 화면에서 보고 바로 조치합니다.
            </p>
            {runtimeDegraded ? (
              <p className="section-body">
                runtime 응답이 느려 일부 카드를 fallback 상태로 보여주고 있습니다.
                {runtimeDegradedReason ? ` (${runtimeDegradedReason})` : ''}
              </p>
            ) : null}
          </div>
          <div className="inline-action-block__buttons">
            <span className={statusToneClass(daemon.status)}>{`daemon ${humanizeStatus(daemon.status)}`}</span>
            <span className={statusToneClass(continuity.status)}>{`VM ${humanizeStatus(continuity.status)}`}</span>
            <span className={statusToneClass(overview.security?.status)}>{`보안 ${humanizeStatus(overview.security?.status)}`}</span>
            <span className="status-pill">{canWrite ? '쓰기 가능' : '읽기 전용'}</span>
          </div>
        </div>

        <FounderPipelineBar
          signalCount={sourceIntake?.recentCount ?? 0}
          jobCount={overview.jobs.openCount}
          reviewCount={overview.review.openCount}
        />

        <div className="proof-metrics proof-metrics--admin">
          <article className="metric-card">
            <span className="metric-card__label">서비스 상태</span>
            <strong className="metric-card__value">{humanizeStatus(daemon.status)}</strong>
            <span className="metric-card__note">{daemon.address || 'daemon 주소 미기록'}</span>
          </article>
          <article className="metric-card">
            <span className="metric-card__label">결재 대기</span>
            <strong className="metric-card__value">{overview.review.openCount}</strong>
            <span className="metric-card__note">{overview.reviewNote || '결재판이 비어 있습니다.'}</span>
          </article>
          <article className="metric-card">
            <span className="metric-card__label">처리 중인 과업</span>
            <strong className="metric-card__value">{overview.jobs.openCount}</strong>
            <span className="metric-card__note">{jobCounts.summaryNote || '지금 열린 과업이 없습니다.'}</span>
          </article>
          <article className="metric-card">
            <span className="metric-card__label">쓰기 권한</span>
            <strong className="metric-card__value">{canWrite ? '가능' : '읽기'}</strong>
            <span className="metric-card__note">
              {overview.auth?.write_auth_enabled ? 'write auth가 켜져 있습니다.' : 'write auth 설정을 확인하세요.'}
            </span>
          </article>
          <article className="metric-card">
            <span className="metric-card__label">포착된 신호</span>
            <strong className="metric-card__value">{sourceIntake?.recentCount ?? 0}</strong>
            <span className="metric-card__note">{sourceIntake?.summaryNote || '신호 카드는 아래에서 이어서 불러옵니다.'}</span>
          </article>
          <article className="metric-card">
            <span className="metric-card__label">콘텐츠 발행</span>
            <strong className="metric-card__value">stream</strong>
            <span className="metric-card__note">발행 카드는 아래에서 이어서 불러옵니다.</span>
          </article>
        </div>

        <div className="admin-jump-grid">
          {jumpLinks.map(renderJumpLink)}
        </div>
      </section>

      <section className="admin-card">
        <p className="section-eyebrow">지금 당장</p>
        <h2 className="section-title">지금 먼저 볼 것</h2>
        <div className="admin-list">
          <div><strong>필요한 액션</strong><span>{primaryAttention?.mode || founderAttention?.mode || '없음'}</span></div>
          <div><strong>참조 안건</strong><span>{primaryAttention?.ref || founderAttention?.ref || '없음'}</span></div>
          <div><strong>결재 대기</strong><span>{founderAttention?.waitingCount ?? 0}건</span></div>
          <div><strong>리스크</strong><span>{founderAttention?.riskCount ?? 0}건</span></div>
        </div>
        <p className="section-body">
          {primaryAttention?.summary
            ? `가장 먼저 볼 후속 액션: ${primaryAttention.summary}`
            : overview.reviewNote
            ? `가장 먼저 볼 안건: ${overview.reviewNote}`
            : '지금 결재 대기 중인 급한 안건이 없습니다.'}
        </p>
        {primaryAttention?.detail ? <p className="section-body">{primaryAttention.detail}</p> : null}
      </section>

      <section className="admin-card">
        <p className="section-eyebrow">사실 기록</p>
        <h2 className="section-title">최근 운영 사실</h2>
        <div className="admin-list">
          <div><strong>마지막 배포</strong><span>{continuity.latestDeployId || '미기록'}</span></div>
          <div><strong>릴리스 스탬프</strong><span>{continuity.releaseStamp || '미기록'}</span></div>
          <div><strong>원격 경로</strong><span>{continuity.remoteReleaseDir || '미기록'}</span></div>
          <div><strong>마지막 검증</strong><span>{latestVerification?.record_id || '미기록'}</span></div>
        </div>
        <p className="section-body">
          {latestVerification
            ? `${humanizeStatus(latestVerification.status)} · ${formatAdminTime(latestVerification.started_at)}`
            : '최근 검증 기록이 아직 없습니다.'}
        </p>
      </section>

      <section id="continuity" className="admin-card admin-card--wide">
        <p className="section-eyebrow">Runtime / deploy</p>
        <h2 className="section-title">배포와 런타임</h2>
        <div className="proof-metrics proof-metrics--admin">
          <article className="metric-card">
            <span className="metric-card__label">daemon</span>
            <strong className="metric-card__value">{humanizeStatus(daemon.status)}</strong>
            <span className="metric-card__note">{daemon.address || '주소 미기록'}</span>
          </article>
          <article className="metric-card">
            <span className="metric-card__label">VM 호스트</span>
            <strong className="metric-card__value">{humanizeStatus(continuity.status)}</strong>
            <span className="metric-card__note">{continuity.host || 'VM 대상이 아직 없습니다.'}</span>
          </article>
          <article className="metric-card">
            <span className="metric-card__label">배포 상태</span>
            <strong className="metric-card__value">{humanizeStatus(daemon.deployHealth)}</strong>
            <span className="metric-card__note">company state release corridor</span>
          </article>
          <article className="metric-card">
            <span className="metric-card__label">VM 대상</span>
            <strong className="metric-card__value">{overview.deployment.vmTargetReady ? '연결됨' : '미설정'}</strong>
            <span className="metric-card__note">{`${overview.deployment.counts.targets}개 대상 등록됨`}</span>
          </article>
          <article className="metric-card">
            <span className="metric-card__label">배포 기록</span>
            <strong className="metric-card__value">{overview.deployment.counts.releases}</strong>
            <span className="metric-card__note">{`${overview.deployment.counts.deploys} deploy · ${overview.deployment.counts.rollbacks} rollback`}</span>
          </article>
        </div>

        <div className="signal-grid">
          <article className="signal-card">
            <span className="signal-card__source">VM host</span>
            <h3 className="signal-card__title">
              <span className={statusToneClass(continuity.status)}>{humanizeStatus(continuity.status)}</span>
            </h3>
            <p className="section-body">
              {continuity.host || 'VM 호스트가 아직 연결되지 않았습니다.'}
              {' · '}
              최근 배포 {formatAdminTime(continuity.latestDeployAt)}
            </p>
            <div className="admin-list">
              <div><strong>최근 deploy</strong><span>{continuity.latestDeployId || '미기록'}</span></div>
              <div><strong>최근 release</strong><span>{continuity.latestReleaseId || '미기록'}</span></div>
              <div><strong>걸린 시간</strong><span>{formatDuration(continuity.latestDeployDurationSeconds)}</span></div>
              <div><strong>보안 상태</strong><span>{humanizeStatus(continuity.securityStatus)}</span></div>
            </div>
            {renderInlineItems([
              continuity.releaseStamp ? `release stamp ${continuity.releaseStamp}` : '',
              continuity.remoteReleaseDir ? `remote ${continuity.remoteReleaseDir}` : '',
              continuity.writeAuthEnabled ? 'write auth 켜짐' : 'write auth 꺼짐',
              continuity.lastSecurityReviewAt ? `보안 점검 ${formatAdminTime(continuity.lastSecurityReviewAt)}` : '',
            ].filter(Boolean), '기록된 배포 증거가 아직 부족합니다.')}
          </article>

          <article className="signal-card">
            <span className="signal-card__source">Runtime</span>
            <h3 className="signal-card__title">
              <span className={statusToneClass(daemon.status)}>{humanizeStatus(daemon.status)}</span>
            </h3>
            <p className="section-body">
              architect 자동 분배 {daemon.architectAutoDispatch ? '켜짐' : '꺼짐'}
              {' · '}
              최근 실행 {formatAdminTime(daemon.architectLastRunAt)}
            </p>
            {renderInlineItems(daemon.degradedReasons, 'degraded 사유가 보고되지 않았습니다.')}
          </article>

          <article className="signal-card">
            <span className="signal-card__source">Targets</span>
            <h3 className="signal-card__title">{overview.deployment.vmTargetReady ? '배포 대상이 연결돼 있습니다.' : '배포 대상을 먼저 등록해야 합니다.'}</h3>
            {overview.deployment.targets.length > 0 ? (
              <ul className="inline-list">
                {overview.deployment.targets.map((target) => (
                  <li key={target.targetId}>
                    <span>{target.targetId}</span>
                    <p>{target.commandPreview || '명령 미기록'}</p>
                    {target.host ? <p>{`host ${target.host}`}</p> : null}
                    {target.workdir ? <p>{target.workdir}</p> : null}
                  </li>
                ))}
              </ul>
            ) : (
              <p className="section-body">등록된 deploy target이 아직 없습니다.</p>
            )}
          </article>
        </div>
      </section>

      <section className="admin-card admin-card--wide" id="founder-flow">
        <p className="section-eyebrow">Founder flow</p>
        <h2 className="section-title">승인 · 릴리스 · 롤백</h2>
        <p className="section-body">
          승인, 릴리스, 롤백을 각 카드로 분리해서 바로 실행합니다.
        </p>
        <AdminFounderActions
          canWrite={canWrite}
          defaults={founderFlowDefaults}
        />
      </section>

      <div id="quickwork" className="admin-card--wide">
        <AdminQuickworkActions canWrite={canWrite} />
      </div>

      <Suspense
        fallback={(
          <LoadingLaneCard
            anchorId="source-intake"
            eyebrow="포착된 신호"
            title="신호 · 아이디어 inbox"
            body="링크·메모·아이디어를 쌓는 운영 카드를 준비 중입니다."
          />
        )}
      >
        <AdminSourceIntakeLane canWrite={canWrite} />
      </Suspense>

      <Suspense
        fallback={(
          <LoadingLaneCard
            anchorId="brand-publish"
            eyebrow="콘텐츠 발행"
            title="초안 · 발행 준비"
            body="초안부터 발행 대기까지 잇는 운영 카드를 준비 중입니다."
          />
        )}
      >
        <AdminBrandPublishLane canWrite={canWrite} />
      </Suspense>

      <div id="telegram" className="admin-card--wide">
        <AdminTelegramActions canWrite={canWrite} />
      </div>

      <section className="admin-card admin-card--wide">
        <p className="section-eyebrow">Timeline</p>
        <h2 className="section-title">최근 배포 기록</h2>
        <div className="admin-table">
          {renderDeployTimelineRow('release', overview.deployment.latestRelease)}
          {renderDeployTimelineRow('deploy', overview.deployment.latestDeploy)}
          {renderDeployTimelineRow('rollback', overview.deployment.latestRollback)}
        </div>
      </section>

      <section className="admin-card admin-card--wide">
        <p className="section-eyebrow">과업</p>
        <h2 className="section-title">최근 과업 현황</h2>
        <div className="admin-table">
          {overview.jobs.recent.map((job) => (
            <article className="admin-table__row" key={job.jobId}>
              <div>
                <strong>{job.summary}</strong>
                <p>{job.followUp?.summary || job.jobId}</p>
              </div>
              <div>{humanizeRole(job.role)}</div>
              <div>{humanizeStatus(job.status)}</div>
              <div>{job.stage || '세부 단계 없음'}</div>
            </article>
          ))}
        </div>
      </section>
    </div>
  );
}
