import AdminLoginForm from '../../../components/admin-login-form';
import { getRuntimeConfig } from '../../../lib/env/server';

export default async function AdminLoginPage({ searchParams }) {
  const params = await searchParams;
  const nextPath = typeof params?.next === 'string' ? params.next : '/admin';
  const reason = typeof params?.reason === 'string' ? params.reason : '';
  const { localAdminBootstrapEnabled, passwordAdminLoginEnabled } = getRuntimeConfig();

  return (
    <main className="admin-login-page">
      <AdminLoginForm
        nextPath={nextPath}
        reason={reason}
        bootstrapEnabled={localAdminBootstrapEnabled}
        passwordEnabled={passwordAdminLoginEnabled}
      />
    </main>
  );
}
