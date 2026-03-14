import { getRequestSessionMeta } from '../../../lib/auth/request-session';

import AdminShell from '../../../components/admin-shell';

export default async function AdminConsoleLayout({ children }) {
  const session = await getRequestSessionMeta();
  return <AdminShell userId={session.userId} roles={session.roles}>{children}</AdminShell>;
}
