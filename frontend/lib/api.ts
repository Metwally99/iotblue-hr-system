export type Role = 'admin' | 'hr' | 'manager' | 'employee';

export type Identity = {
  id: string;
  name: string;
  email: string;
  role: Role;
};

export class ApiError extends Error {
  constructor(public status: number, public code: string, message: string, public details?: unknown) {
    super(message);
  }
}

const apiBase = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/api/v1';

export async function api<T>(path: string, init: RequestInit = {}, token?: string | null): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set('Accept', 'application/json');
  if (init.body) headers.set('Content-Type', 'application/json');
  if (token) headers.set('Authorization', `Bearer ${token}`);
  const response = await fetch(`${apiBase}${path}`, { ...init, headers, cache: 'no-store' });
  const body = await response.json().catch(() => null) as { error?: { code?: string; message?: string; details?: unknown } } | null;
  if (!response.ok) {
    throw new ApiError(response.status, body?.error?.code ?? 'REQUEST_FAILED', body?.error?.message ?? 'The request failed', body?.error?.details);
  }
  return body as T;
}
