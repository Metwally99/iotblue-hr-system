'use client';

import { FormEvent, useState } from 'react';
import { useAuth } from '@/components/auth-provider';
import { ApiError } from '@/lib/api';
import { BrandMark, Icon } from '@/components/icons';
import { SignalField } from '@/components/signal-field';

export default function LoginPage() {
  const { login } = useAuth();
  const [email, setEmail] = useState('admin@example.com');
  const [password, setPassword] = useState('ChangeMe123!');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setSubmitting(true); setError('');
    try {
      await login(email, password);
      window.location.assign('/');
    } catch (cause) {
      setError(cause instanceof ApiError ? cause.message : 'Could not reach the HR API.');
    } finally { setSubmitting(false); }
  }

  return (
    <main className="login-page">
      <section className="login-story">
        <SignalField/>
        <div className="login-story-glow" aria-hidden="true"/>
        <div className="brand login-brand">
          <BrandMark className="brand-mark"/>
          <span className="brand-copy"><strong>Humanly</strong><small>People intelligence</small></span>
        </div>
        <div className="login-thesis">
          <p className="eyebrow">CONNECTED PEOPLE OPERATIONS</p>
          <h1>Every person.<br/><span>One clear signal.</span></h1>
          <p>Turn employee activity, time, and requests into decisions your team can trust.</p>
        </div>
        <div className="login-sequence" aria-label="Humanly workflow">
          <div><span>01</span><strong>People</strong><small>One trusted directory</small></div>
          <div><span>02</span><strong>Time</strong><small>Attendance and leave</small></div>
          <div><span>03</span><strong>Decisions</strong><small>Clear, accountable action</small></div>
        </div>
      </section>
      <section className="login-form-wrap">
        <form className="login-form" onSubmit={submit}>
          <div className="login-form-head">
            <span className="form-signal"><i/> Secure local workspace</span>
            <p className="eyebrow">WELCOME BACK</p>
            <h2>Sign in to Humanly</h2>
            <p>Use your work account to open the operations workspace.</p>
          </div>
          <label>Email<input autoComplete="email" onChange={(e) => setEmail(e.target.value)} type="email" value={email} /></label>
          <label>Password<input autoComplete="current-password" onChange={(e) => setPassword(e.target.value)} type="password" value={password} /></label>
          {error && <p className="form-error" role="alert">{error}</p>}
          <button className="primary login-button" disabled={submitting} type="submit">
            <span>{submitting ? 'Opening workspace…' : 'Open workspace'}</span>
            {!submitting ? <Icon name="arrow"/> : null}
          </button>
          <p className="demo-note"><Icon name="check"/> Demo credentials are prefilled for tomorrow&apos;s presentation.</p>
        </form>
      </section>
    </main>
  );
}
