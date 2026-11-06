'use client';

import { FormEvent, useCallback, useEffect, useState } from 'react';
import { useAuth } from '@/components/auth-provider';
import { PageState, PortalShell } from '@/components/portal-shell';
import { api, ApiError } from '@/lib/api';
import type { Department } from '@/lib/types';

export default function DepartmentsPage() {
  const { token } = useAuth();
  const [departments, setDepartments] = useState<Department[]>([]);
  const [name, setName] = useState('');
  const [editing, setEditing] = useState<Department | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');

  const load = useCallback(async () => {
    if (!token) return;
    setLoading(true); setError('');
    try { setDepartments((await api<{ data: Department[] }>('/departments', {}, token)).data); }
    catch (caught) { setError(caught instanceof ApiError ? caught.message : 'Departments could not be loaded.'); }
    finally { setLoading(false); }
  }, [token]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  async function submit(event: FormEvent) {
    event.preventDefault(); if (!token) return;
    setError(''); setNotice('');
    try {
      await api(editing ? `/departments/${editing.id}` : '/departments', { method: editing ? 'PATCH' : 'POST', body: JSON.stringify({ name }) }, token);
      setNotice(editing ? 'Department renamed.' : 'Department created.'); setName(''); setEditing(null); await load();
    } catch (caught) { setError(caught instanceof ApiError ? caught.message : 'Department could not be saved.'); }
  }

  return <PortalShell adminOnly eyebrow="ORGANISATION" title="Departments">
    <section className="split-layout departments-layout">
      <form className="panel resource-form" onSubmit={submit}>
        <p className="eyebrow">{editing ? 'EDIT TEAM' : 'NEW TEAM'}</p><h2>{editing ? 'Rename department' : 'Create a department'}</h2>
        <label>Department name<input required minLength={2} maxLength={100} value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. Product Engineering"/></label>
        <button className="primary form-submit">{editing ? 'Save name' : 'Create department'}</button>
        {editing ? <button type="button" className="text-button form-cancel" onClick={() => { setEditing(null); setName(''); }}>Cancel</button> : null}
      </form>
      <section className="panel table-panel">
        <div className="panel-heading"><div><p className="eyebrow">TEAMS</p><h2>{departments.length} departments</h2></div></div>
        {notice ? <p className="notice success-notice">{notice}</p> : null}
        {loading || error ? <PageState loading={loading} error={error}/> : departments.length === 0 ? <PageState empty="Create the first department before adding employees."/> : <div className="department-cards">{departments.map((department, index) => <article key={department.id}><span>{String(index + 1).padStart(2, '0')}</span><div><strong>{department.name}</strong><small>Created {new Date(department.created_at).toLocaleDateString()}</small></div><button onClick={() => { setEditing(department); setName(department.name); }}>Rename</button></article>)}</div>}
      </section>
    </section>
  </PortalShell>;
}
