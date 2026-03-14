import { NextResponse } from 'next/server';
import { readValidatedJson } from '../../../../lib/http/request';
import { publicContactSchema } from '../../../../lib/validation/public-contact';

export async function POST(request) {
  const parsed = await readValidatedJson(request, publicContactSchema);
  if (!parsed.ok) {
    return NextResponse.json({ error: parsed.error, details: parsed.details }, { status: parsed.status });
  }

  return NextResponse.json({
    ok: true,
    accepted: true,
    payload: {
      name: parsed.data.name,
      email: parsed.data.email,
      messageLength: parsed.data.message.length,
    },
  });
}
