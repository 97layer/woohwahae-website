import { NextResponse } from 'next/server';
import { getPublicProofView } from '../../../../lib/runtime/view-model';

export async function GET() {
  const proof = await getPublicProofView();
  return NextResponse.json({ proof });
}
