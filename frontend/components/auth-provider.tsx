'use client';

import { createContext, useContext, useEffect, useMemo, useState } from 'react';
import { api, Identity } from '@/lib/api';

type AuthContextValue = {
  identity: Identity | null;
  token: string | null;
  loading: boolean;
  login(email: string, password: string): Promise<void>;
  logout(): void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(null);
  const [identity, setIdentity] = useState<Identity | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      const stored = window.sessionStorage.getItem('hr_access_token');
      if (!stored) {
        setLoading(false);
        return;
      }
      setToken(stored);
      const storedIdentity = window.sessionStorage.getItem('hr_identity');
      if (!storedIdentity) {
        window.sessionStorage.removeItem('hr_access_token');
        setToken(null);
        setLoading(false);
        return;
      }
      try { setIdentity(JSON.parse(storedIdentity) as Identity); }
      catch { window.sessionStorage.removeItem('hr_access_token'); setToken(null); }
      setLoading(false);
    }, 0);
    return () => window.clearTimeout(timer);
  }, []);

  async function login(email: string, password: string) {
    const response = await api<{ data: { access_token: string; employee: Identity } }>('/auth/login', {
      method: 'POST', body: JSON.stringify({ email, password }),
    });
    window.sessionStorage.setItem('hr_access_token', response.data.access_token);
    window.sessionStorage.setItem('hr_identity', JSON.stringify(response.data.employee));
    setToken(response.data.access_token);
    setIdentity(response.data.employee);
  }

  function logout() {
    window.sessionStorage.removeItem('hr_access_token');
    window.sessionStorage.removeItem('hr_identity');
    setToken(null);
    setIdentity(null);
    window.location.assign('/login');
  }

  const value = useMemo(() => ({ identity, token, loading, login, logout }), [identity, token, loading]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) throw new Error('useAuth must be used inside AuthProvider');
  return value;
}
