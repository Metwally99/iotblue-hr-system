'use client';

import { FormEvent, useCallback, useEffect, useState } from 'react';
import { useAuth } from '@/components/auth-provider';
import { PageState, PortalShell, StatusBadge } from '@/components/portal-shell';
import { api, ApiError } from '@/lib/api';
import type { LeaveBalance, LeaveRequest, LeaveStatus, LeaveType, ListResponse } from '@/lib/types';

export default function LeavePage() {
  const { identity, token } = useAuth();
  const [items, setItems] = useState<LeaveRequest[]>([]);
  const [balance, setBalance] = useState<LeaveBalance | null>(null);
  const [status, setStatus] = useState<LeaveStatus | 'all'>('all');
	const [form, setForm] = useState<{ start_date: string; end_date: string; leave_type: LeaveType }>({ start_date: '', end_date: '', leave_type: 'annual' });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');

  const load = useCallback(async () => {
    if (!token || !identity) return;
    setLoading(true); setError('');
    const suffix = status === 'all' ? '' : `&status=${status}`;
    try {
      const canReview = identity.role === 'admin' || identity.role === 'hr' || identity.role === 'manager';
      const response = await api<ListResponse<LeaveRequest>>(`/leave-requests?page_size=100${suffix}`, {}, token);
      setItems(response.data);
      const mine = await api<{ data: LeaveBalance }>(`/employees/${identity.id}/leave-balances`, {}, token).catch(() => null);
      setBalance(mine?.data ?? null);
    } catch (caught) { setError(caught instanceof ApiError ? caught.message : 'Leave requests could not be loaded.'); }
    finally { setLoading(false); }
  }, [identity, status, token]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  async function submit(event: FormEvent) {
    event.preventDefault(); if (!token) return;
    setSaving(true); setError(''); setNotice('');
	try { await api('/leave-requests', { method: 'POST', body: JSON.stringify(form) }, token); setForm({ start_date: '', end_date: '', leave_type: 'annual' }); setNotice('Leave request submitted for review.'); await load(); }
    catch (caught) { setError(caught instanceof ApiError ? caught.message : 'Leave request could not be submitted.'); }
    finally { setSaving(false); }
  }

  async function decide(id: string, decision: 'approve' | 'reject') {
    if (!token) return;
    const review_note = window.prompt(`Optional note for this ${decision} decision:`) || null;
    try { await api(`/leave-requests/${id}/${decision}`, { method: 'POST', body: JSON.stringify({ review_note }) }, token); setNotice(`Request ${decision}d.`); await load(); }
    catch (caught) { setError(caught instanceof ApiError ? caught.message : 'Decision could not be saved.'); }
  }

  const canReview = identity?.role === 'admin' || identity?.role === 'hr' || identity?.role === 'manager';
  return <PortalShell eyebrow="TIME AWAY" title={canReview ? 'Leave management' : 'My leave'}>
    <section className="leave-overview">
      <form className="panel resource-form" onSubmit={submit}>
        <p className="eyebrow">NEW REQUEST</p><h2>Plan time away</h2>
        {balance ? <div className="balance-card"><strong>{balance.available_days}</strong><span>days available of {balance.total_days}</span></div> : null}
		<div className="form-grid"><label>Start date<input required type="date" value={form.start_date} onChange={(e) => setForm({ ...form, start_date: e.target.value })}/></label><label>End date<input required min={form.start_date} type="date" value={form.end_date} onChange={(e) => setForm({ ...form, end_date: e.target.value })}/></label></div>
		<label>Leave type<select value={form.leave_type} onChange={(e) => setForm({ ...form, leave_type: e.target.value as LeaveType })}><option value="annual">Annual</option><option value="sick">Sick</option><option value="holiday">Holiday</option></select></label>
        <button className="primary form-submit" disabled={saving}>{saving ? 'Submitting…' : 'Submit request'}</button>
      </form>
      <section className="panel table-panel">
        <div className="panel-heading"><div><p className="eyebrow">REQUESTS</p><h2>{items.length} results</h2></div><select className="filter-select" value={status} onChange={(e) => setStatus(e.target.value as LeaveStatus | 'all')}><option value="all">All statuses</option><option value="pending">Pending</option><option value="approved">Approved</option><option value="rejected">Rejected</option></select></div>
        {notice ? <p className="notice success-notice">{notice}</p> : null}
        {loading || error ? <PageState loading={loading} error={error}/> : items.length === 0 ? <PageState empty="Submit a request or change the status filter."/> : <div className="leave-list">{items.map((item) => <article key={item.id}>
          <div className="leave-date"><strong>{new Date(`${item.start_date}T00:00:00`).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</strong><span>to {new Date(`${item.end_date}T00:00:00`).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</span></div>
		  <div className="leave-copy"><strong>{item.employee_name}</strong><p>{item.leave_type} leave</p>{item.review_note ? <small>Reviewer note: {item.review_note}</small> : null}</div>
          <StatusBadge value={item.status}/>
          {canReview && item.status === 'pending' ? <div className="table-actions"><button className="positive-action" onClick={() => void decide(item.id, 'approve')}>Approve</button><button className="danger-action" onClick={() => void decide(item.id, 'reject')}>Reject</button></div> : null}
        </article>)}</div>}
      </section>
    </section>
  </PortalShell>;
}
