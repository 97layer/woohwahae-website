import { NextResponse } from 'next/server';
import { authorizeRequest } from '../../../../../lib/auth/rbac';
import { getKnowledgeView } from '../../../../../lib/runtime/view-model';

export async function GET(request) {
  const auth = await authorizeRequest(request, ['founder', 'admin']);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.error }, { status: auth.status });
  }

  const query = request.nextUrl.searchParams.get('q') || '';
  try {
    return NextResponse.json({ knowledge: await getKnowledgeView(query) });
  } catch (error) {
    return NextResponse.json({ error: error instanceof Error ? error.message : 'runtime unavailable' }, { status: 502 });
  }
}
