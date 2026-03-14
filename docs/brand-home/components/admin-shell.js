import Link from 'next/link';

import AdminLogoutButton from './admin-logout-button';
import PwaInstallButton from './pwa-install-button';

const navItems = [
  { href: '/', label: '홈' },
  { href: '/admin', label: '운영판' },
  { href: '/admin/review-room', label: '결재판' },
  { href: '/admin/jobs', label: '과업 현황' },
  { href: '/admin/knowledge', label: '기억 창고' },
];

export default function AdminShell({ userId, roles, children }) {
  const canWrite = roles.includes('founder');

  return (
    <div className="admin-app">
      <header className="admin-header">
        <div className="admin-header__inner">
          <div>
            <p className="section-eyebrow">Admin</p>
            <h1 className="admin-header__title">Layer OS 운영센터</h1>
            <p className="section-body admin-header__summary">
              포착된 신호부터 과업 처리, 최종 결재까지 — founder 판단에 필요한 것만 한 화면에서 봅니다.
            </p>
          </div>
          <div className="admin-header__meta">
            <PwaInstallButton />
            <span className="status-pill">{canWrite ? '쓰기 가능' : '읽기 전용'}</span>
            <span className="admin-header__actor">{userId}</span>
            <AdminLogoutButton />
          </div>
        </div>
      </header>
      <div className="admin-layout">
        <aside className="admin-nav">
          <nav>
            {navItems.map((item) => (
              <Link key={item.href} href={item.href} className="admin-nav__link">
                {item.label}
              </Link>
            ))}
          </nav>
        </aside>
        <main className="admin-main">{children}</main>
      </div>
    </div>
  );
}
