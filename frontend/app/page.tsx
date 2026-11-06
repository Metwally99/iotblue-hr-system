'use client';

import Link from 'next/link';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useAuth } from '@/components/auth-provider';
import { PageState, PortalShell, StatusBadge } from '@/components/portal-shell';
import { Icon } from '@/components/icons';
import { SignalField } from '@/components/signal-field';
import { api, ApiError } from '@/lib/api';
import type { AttendanceRecord, Employee, LeaveBalance, LeaveRequest, ListResponse } from '@/lib/types';

type DashboardData = { employees: Employee[]; leave: LeaveRequest[]; attendance: AttendanceRecord[]; balance: LeaveBalance | null };
const emptyData: DashboardData = { employees: [], leave: [], attendance: [], balance: null };

export default function Home() {
  const { identity, token } = useAuth();
  const [data, setData] = useState<DashboardData>(emptyData);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const today = useMemo(() => new Date().toISOString().slice(0, 10), []);

  const load = useCallback(async () => {
    if (!identity || !token) return;
    setLoading(true);
    setError('');
    try {
      if (identity.role === 'admin') {
        const [employees, leave, attendance] = await Promise.all([
          api<ListResponse<Employee>>('/employees?page_size=100&is_active=true', {}, token),
          api<ListResponse<LeaveRequest>>('/leave-requests?page_size=100', {}, token),
          api<ListResponse<AttendanceRecord>>(`/attendance?page_size=100&start_date=${today}&end_date=${today}`, {}, token),
        ]);
        setData({ employees: employees.data, leave: leave.data, attendance: attendance.data, balance: null });
      } else {
        const [leave, attendance, balance] = await Promise.all([
          api<ListResponse<LeaveRequest>>('/leave-requests?page_size=100', {}, token),
          api<ListResponse<AttendanceRecord>>(`/attendance?page_size=100&start_date=${today}&end_date=${today}`, {}, token),
          api<{ data: LeaveBalance }>(`/employees/${identity.id}/leave-balances`, {}, token).catch(() => ({ data: null as unknown as LeaveBalance })),
        ]);
        setData({ employees: [], leave: leave.data, attendance: attendance.data, balance: balance.data });
      }
    } catch (caught) {
      setError(caught instanceof ApiError ? caught.message : 'The API is unavailable. Check that the backend is running.');
    } finally { setLoading(false); }
  }, [identity, token, today]);

  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  async function decide(id: string, decision: 'approve' | 'reject') {
    if (!token) return;
    try {
      await api(`/leave-requests/${id}/${decision}`, { method: 'POST', body: JSON.stringify({}) }, token);
      await load();
    } catch (caught) { setError(caught instanceof ApiError ? caught.message : 'The decision could not be saved.'); }
  }

  async function attendanceAction(action: 'check-in' | 'check-out') {
    if (!token) return;
    try {
      await api(`/attendance/${action}`, { method: 'POST' }, token);
      await load();
    } catch (caught) { setError(caught instanceof ApiError ? caught.message : 'Attendance could not be updated.'); }
  }

  const pending = data.leave.filter((item) => item.status === 'pending');
  const myToday = data.attendance.find((item) => item.employee_id === identity?.id);
  const action = identity?.role === 'admin'
    ? <Link className="primary button-link" href="/employees"><Icon name="plus"/> Add employee</Link>
    : <Link className="primary button-link" href="/leave"><Icon name="plus"/> Request leave</Link>;

  return (
    <PortalShell eyebrow={new Date().toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' })} title={`Good morning, ${identity?.name.split(' ')[0] ?? 'there'}`} action={action}>
      {loading || error ? <PageState loading={loading} error={error} /> : <>
        <section className="metrics" aria-label="Company overview">
          <article className="metric hero-metric">
            <SignalField compact/>
            <div className="metric-content"><p>{identity?.role === 'admin' ? 'Active employees' : 'Available leave'}</p><strong>{identity?.role === 'admin' ? data.employees.length : data.balance?.available_days ?? '—'}</strong><span>{identity?.role === 'admin' ? 'Connected across the organisation' : `of ${data.balance?.total_days ?? '—'} days this year`}</span></div>
            <span className="hero-status"><i/> People signal live</span>
          </article>
          <article className="metric"><div className="metric-head"><span className="symbol blue"><Icon name="attendance"/></span><em>Today</em></div><p>{identity?.role === 'admin' ? 'Present today' : 'Attendance'}</p><strong>{identity?.role === 'admin' ? data.attendance.length : myToday ? 'In' : '—'}</strong><span>{myToday?.check_out ? 'Shift completed' : myToday ? 'Currently checked in' : 'No check-in yet'}</span></article>
          <article className="metric"><div className="metric-head"><span className="symbol cyan"><Icon name="leave"/></span><em>Open</em></div><p>Leave requests</p><strong>{pending.length}</strong><span>{identity?.role === 'admin' ? 'Need administrator review' : 'Awaiting a decision'}</span></article>
          <article className="metric"><div className="metric-head"><span className="symbol violet"><Icon name="check"/></span><em>This year</em></div><p>Approved leave</p><strong>{data.leave.filter((item) => item.status === 'approved').length}</strong><span>Requests approved</span></article>
        </section>
        <section className="content-grid">
          <article className="panel requests-panel">
            <div className="panel-heading"><div><p className="eyebrow">RECENT ACTIVITY</p><h2>Leave requests</h2></div><Link className="inline-link" href="/leave">View all <Icon name="arrow"/></Link></div>
            {data.leave.length === 0 ? <p className="inline-empty">No leave requests yet.</p> : <div className="request-list">
              {data.leave.slice(0, 4).map((request) => <div className="request-row" key={request.id}>
                <div className="avatar amber">{request.employee_name.split(' ').map((word) => word[0]).join('').slice(0, 2)}</div>
				<div className="request-person"><strong>{request.employee_name}</strong><span>{request.leave_type} leave</span></div>
                <div className="request-date"><strong>{request.start_date} → {request.end_date}</strong><StatusBadge value={request.status}/></div>
                {identity?.role === 'admin' && request.status === 'pending' ? <div className="row-actions"><button className="approve" onClick={() => void decide(request.id, 'approve')}><Icon name="check"/> Approve</button><button className="more" onClick={() => void decide(request.id, 'reject')} aria-label={`Reject ${request.employee_name}`}><Icon name="close"/></button></div> : null}
              </div>)}
            </div>}
          </article>
          <article className="panel attendance-panel">
            <div className="panel-heading"><div><p className="eyebrow">MY ATTENDANCE</p><h2>Today</h2></div><span className="live"><i/> Live</span></div>
            <div className="clock"><strong>{myToday ? new Date(myToday.check_in).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '--:--'}</strong><span>{myToday ? 'Checked in' : 'Not checked in'}</span></div>
            <button className="checkout" onClick={() => void attendanceAction(myToday && !myToday.check_out ? 'check-out' : 'check-in')} disabled={Boolean(myToday?.check_out)}>{myToday?.check_out ? 'Shift completed' : myToday ? 'Check out' : 'Check in'}</button>
            <p className="timezone">Server time · Africa/Cairo</p>
          </article>
        </section>
        <article className="panel team-panel">
          <div className="panel-heading"><div><p className="eyebrow">QUICK VIEW</p><h2>{identity?.role === 'admin' ? 'Organisation status' : 'My leave status'}</h2></div></div>
          <div className="summary-strip">
            <div><strong>{data.leave.length}</strong><span>Total requests</span></div><div><strong>{pending.length}</strong><span>Pending</span></div><div><strong>{data.leave.filter((item) => item.status === 'approved').length}</strong><span>Approved</span></div><div><strong>{data.leave.filter((item) => item.status === 'rejected').length}</strong><span>Rejected</span></div>
          </div>
        </article>
      </>}
    </PortalShell>
  );
}
