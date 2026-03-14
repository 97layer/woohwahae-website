import { NextResponse } from 'next/server';

export async function POST() {
  const response = NextResponse.json({ ok: true, loggedOut: true });
  response.cookies.set('wwh_session', '', {
    httpOnly: true,
    sameSite: 'lax',
    secure: process.env.NODE_ENV === 'production',
    path: '/',
    expires: new Date(0),
  });
  return response;
}
