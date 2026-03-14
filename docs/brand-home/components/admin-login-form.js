'use client';

import { useRouter } from 'next/navigation';
import { useState } from 'react';

export default function AdminLoginForm({
  nextPath = '/admin',
  reason = '',
  bootstrapEnabled = false,
  passwordEnabled = false,
}) {
  const router = useRouter();
  const [token, setToken] = useState('');
  const [password, setPassword] = useState('');
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');

  async function finishLogin(payload) {
    const response = await fetch('/api/admin/session/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ...payload, next: nextPath }),
    });
    const body = await response.json();
    if (!response.ok) {
      throw new Error(body?.error || 'login failed');
    }
    router.push(body.next || nextPath || '/admin');
    router.refresh();
  }

  async function submitPassword(event) {
    event.preventDefault();
    setPending(true);
    setError('');
    try {
      await finishLogin({ password });
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : 'login failed');
    } finally {
      setPending(false);
    }
  }

  async function submitToken(event) {
    event.preventDefault();
    setPending(true);
    setError('');
    try {
      await finishLogin({ token });
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : 'login failed');
    } finally {
      setPending(false);
    }
  }

  async function bootstrap() {
    setPending(true);
    setError('');
    try {
      const response = await fetch('/api/admin/session/bootstrap', { method: 'POST' });
      const payload = await response.json();
      if (!response.ok) {
        throw new Error(payload?.error || 'bootstrap failed');
      }
      router.push(nextPath || payload.next || '/admin');
      router.refresh();
    } catch (bootstrapError) {
      setError(bootstrapError instanceof Error ? bootstrapError.message : 'bootstrap failed');
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="admin-login-form">
      <div className="admin-login-form__copy">
        <p className="section-eyebrow">Founder session</p>
        <h1 className="section-title">보호된 관리자 셸 로그인</h1>
        <p className="section-body">
          토스처럼 바로 비밀번호로 들어가고, 필요하면 로컬 세션 버튼이나 기존 토큰 로그인도 계속 쓸 수 있게 바꿨습니다.
        </p>
        <div className="inline-action-block__buttons">
          {passwordEnabled ? <span className="status-pill status-pill--good">password ready</span> : <span className="status-pill status-pill--muted">password not set</span>}
          {bootstrapEnabled ? <span className="status-pill status-pill--muted">local bootstrap</span> : null}
          <span className="status-pill status-pill--muted">token fallback</span>
        </div>
        {reason ? <p className="admin-login-form__reason">reason: {reason}</p> : null}
      </div>

      {passwordEnabled ? (
        <form className="admin-login-form__panel" onSubmit={submitPassword}>
          <div>
            <p className="section-eyebrow">Password</p>
            <h2 className="section-title">비밀번호로 바로 입장</h2>
          </div>
          <label>
            <span>Admin password</span>
            <input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete="current-password"
              placeholder="서버에 설정한 관리자 비밀번호"
              disabled={pending}
            />
          </label>
          <button type="submit" className="button-primary" disabled={pending || !password.trim()}>
            {pending ? '로그인 중…' : '비밀번호로 로그인'}
          </button>
        </form>
      ) : null}

      {bootstrapEnabled ? (
        <div className="admin-login-form__panel admin-login-form__quick">
          <div>
            <p className="section-eyebrow">Local shortcut</p>
            <h2 className="section-title">로컬 founder 세션 바로 열기</h2>
          </div>
          <p className="section-body">개발 환경 localhost에서만 동작합니다.</p>
          <button type="button" className="button-secondary" onClick={bootstrap} disabled={pending}>
            {pending ? '세션 여는 중…' : '로컬 founder 세션 열기'}
          </button>
        </div>
      ) : null}

      <details className="admin-login-form__panel admin-login-form__advanced">
        <summary>고급 로그인 열기</summary>
        <form className="admin-form-grid" onSubmit={submitToken}>
          <label>
            <span>Session token</span>
            <textarea
              value={token}
              onChange={(event) => setToken(event.target.value)}
              rows={6}
              placeholder="이미 발급된 founder/admin 세션 토큰"
              disabled={pending}
            />
          </label>
          <button type="submit" className="button-secondary" disabled={pending || !token.trim()}>
            {pending ? '검증 중…' : '토큰으로 로그인'}
          </button>
        </form>
      </details>

      {error ? <p className="contact-form__status is-error">{error}</p> : null}
    </div>
  );
}
