import { afterEach, describe, expect, it, vi } from 'vitest';
import { api, ApiError } from '@/lib/api';

describe('api client', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('adds JSON and bearer headers and returns the response body', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ data: { id: 'employee-1' } }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }));
    vi.stubGlobal('fetch', fetchMock);

    const response = await api<{ data: { id: string } }>('/employees/123', {}, 'signed-token');

    expect(response.data.id).toBe('employee-1');
    const [url, request] = fetchMock.mock.calls[0];
    expect(url).toBe('http://localhost:8080/api/v1/employees/123');
    expect((request.headers as Headers).get('Authorization')).toBe('Bearer signed-token');
    expect((request.headers as Headers).get('Accept')).toBe('application/json');
  });

  it('turns the API error envelope into a typed ApiError', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
      error: { code: 'INVALID_CREDENTIALS', message: 'Email or password is incorrect' },
    }), { status: 401, headers: { 'Content-Type': 'application/json' } })));

    await expect(api('/auth/login', { method: 'POST', body: '{}' })).rejects.toEqual(
      expect.objectContaining<ApiError>({ status: 401, code: 'INVALID_CREDENTIALS', message: 'Email or password is incorrect' }),
    );
  });

  it('uses a safe fallback when a failed response is not JSON', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('gateway down', { status: 502 })));
    await expect(api('/attendance')).rejects.toMatchObject({ status: 502, code: 'REQUEST_FAILED' });
  });
});
