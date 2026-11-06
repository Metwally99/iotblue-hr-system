'use client';

import { FormEvent, useCallback, useEffect, useState } from 'react';
import { useAuth } from '@/components/auth-provider';
import { PageState, PortalShell, StatusBadge } from '@/components/portal-shell';
import { api, ApiError } from '@/lib/api';
import type { Department, Employee, ListResponse } from '@/lib/types';

const blank = { name: '', email: '', password: '', department_id: '', role: 'employee', total_leave_days: '21' };

export default function EmployeesPage() {
  const { token } = useAuth();
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [form, setForm] = useState(blank);
  const [editing, setEditing] = useState<Employee | null>(null);
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');

  const load = useCallback(async () => {
    if (!token) return;
    setLoading(true); setError('');
    try {
      const [people, teams] = await Promise.all([
        api<ListResponse<Employee>>(`/employees?page_size=100&search=${encodeURIComponent(search)}`, {}, token),
        api<{ data: Department[] }>('/departments', {}, token),
      ]);
      setEmployees(people.data); setDepartments(teams.data);
    } catch (caught) { setError(caught instanceof ApiError ? caught.message : 'Employees could not be loaded.'); }
    finally { setLoading(false); }
  }, [search, token]);

  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  async function submit(event: FormEvent) {
    event.preventDefault(); if (!token) return;
    setSaving(true); setError(''); setNotice('');
    try {
      const body = editing ? {
        name: form.name, email: form.email, department_id: form.department_id, role: form.role,
      } : { ...form, total_leave_days: Number(form.total_leave_days) };
      await api(editing ? `/employees/${editing.id}` : '/employees', { method: editing ? 'PATCH' : 'POST', body: JSON.stringify(body) }, token);
      setNotice(editing ? 'Employee updated.' : 'Employee and yearly leave balance created.');
      setEditing(null); setForm(blank); await load();
    } catch (caught) { setError(caught instanceof ApiError ? caught.message : 'Employee could not be saved.'); }
    finally { setSaving(false); }
  }

  function beginEdit(employee: Employee) {
    setEditing(employee);
    setForm({ name: employee.name, email: employee.email, password: '', department_id: employee.department_id, role: employee.role, total_leave_days: '21' });
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }

  async function deactivate(employee: Employee) {
    if (!token || !window.confirm(`Deactivate ${employee.name}? Their history will be preserved.`)) return;
    try { await api(`/employees/${employee.id}/deactivate`, { method: 'PATCH' }, token); setNotice('Employee deactivated safely.'); await load(); }
    catch (caught) { setError(caught instanceof ApiError ? caught.message : 'Employee could not be deactivated.'); }
  }

  return <PortalShell adminOnly eyebrow="PEOPLE DIRECTORY" title="Employees">
    <section className="split-layout">
      <form className="panel resource-form" onSubmit={submit}>
        <div className="panel-heading"><div><p className="eyebrow">{editing ? 'EDIT RECORD' : 'NEW TEAMMATE'}</p><h2>{editing ? 'Update employee' : 'Add an employee'}</h2></div>{editing ? <button type="button" className="text-button" onClick={() => { setEditing(null); setForm(blank); }}>Cancel</button> : null}</div>
        <label>Full name<input required minLength={2} value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })}/></label>
        <label>Email address<input required type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })}/></label>
        {!editing ? <label>Temporary password<input required minLength={8} type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })}/></label> : null}
        <div className="form-grid"><label>Department<select required value={form.department_id} onChange={(e) => setForm({ ...form, department_id: e.target.value })}><option value="">Select…</option>{departments.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label><label>Role<select value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value })}><option value="employee">Employee</option><option value="admin">Administrator</option></select></label></div>
        {!editing ? <label>Annual leave allowance<input required min={1} max={365} type="number" value={form.total_leave_days} onChange={(e) => setForm({ ...form, total_leave_days: e.target.value })}/></label> : null}
        <button className="primary form-submit" disabled={saving}>{saving ? 'Saving…' : editing ? 'Save changes' : 'Create employee'}</button>
      </form>
      <section className="panel table-panel">
        <div className="panel-heading"><div><p className="eyebrow">DIRECTORY</p><h2>{employees.length} employees</h2></div><input className="search-input" aria-label="Search employees" placeholder="Search name or email" value={search} onChange={(e) => setSearch(e.target.value)}/></div>
        {notice ? <p className="notice success-notice">{notice}</p> : null}
        {loading || error ? <PageState loading={loading} error={error}/> : employees.length === 0 ? <PageState empty="Create the first employee using the form."/> : <div className="data-table-wrap"><table className="data-table"><thead><tr><th>Employee</th><th>Department</th><th>Role</th><th>Status</th><th>Actions</th></tr></thead><tbody>{employees.map((employee) => <tr key={employee.id}><td><strong>{employee.name}</strong><span>{employee.email}</span></td><td>{departments.find((item) => item.id === employee.department_id)?.name ?? '—'}</td><td><StatusBadge value={employee.role}/></td><td><StatusBadge value={employee.is_active ? 'active' : 'inactive'}/></td><td className="table-actions"><button onClick={() => beginEdit(employee)}>Edit</button>{employee.is_active ? <button className="danger-action" onClick={() => void deactivate(employee)}>Deactivate</button> : null}</td></tr>)}</tbody></table></div>}
      </section>
    </section>
  </PortalShell>;
}
