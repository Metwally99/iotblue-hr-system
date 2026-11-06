'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { useAuth } from '@/components/auth-provider';
import { PageState, PortalShell, StatusBadge } from '@/components/portal-shell';
import { api, ApiError } from '@/lib/api';
import type { AttendanceRecord, ListResponse } from '@/lib/types';

function duration(record: AttendanceRecord) {
  if (!record.check_out) return 'In progress';
  const hours = (new Date(record.check_out).getTime() - new Date(record.check_in).getTime()) / 3_600_000;
  return `${hours.toFixed(2)} h`;
}

export default function AttendancePage() {
  const { identity, token } = useAuth();
  const today = useMemo(() => new Date().toISOString().slice(0, 10), []);
  const monthStart = useMemo(() => `${today.slice(0, 8)}01`, [today]);
  const [range, setRange] = useState({ start: monthStart, end: today });
  const [items, setItems] = useState<AttendanceRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [working, setWorking] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');

  const load = useCallback(async () => {
    if (!token || !identity) return;
    setLoading(true); setError('');
    try {
      const path = '/attendance';
      const response = await api<ListResponse<AttendanceRecord>>(`${path}?page_size=100&start_date=${range.start}&end_date=${range.end}`, {}, token);
      setItems(response.data);
    } catch (caught) { setError(caught instanceof ApiError ? caught.message : 'Attendance could not be loaded.'); }
    finally { setLoading(false); }
  }, [identity, range.end, range.start, token]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  const myToday = items.find((item) => item.employee_id === identity?.id && item.attendance_date === today);
  async function act(action: 'check-in' | 'check-out') {
    if (!token) return;
    setWorking(true); setError(''); setNotice('');
    try { await api(`/attendance/${action}`, { method: 'POST' }, token); setNotice(action === 'check-in' ? 'You are checked in.' : 'Your shift is complete.'); await load(); }
    catch (caught) { setError(caught instanceof ApiError ? caught.message : 'Attendance could not be updated.'); }
    finally { setWorking(false); }
  }

  const action = <button className="primary" disabled={working || Boolean(myToday?.check_out)} onClick={() => void act(myToday && !myToday.check_out ? 'check-out' : 'check-in')}>{working ? 'Saving…' : myToday?.check_out ? 'Shift complete' : myToday ? 'Check out' : 'Check in'}</button>;
  return <PortalShell eyebrow="WORKDAY" title="Attendance" action={action}>
    <section className="attendance-summary">
      <article className="metric hero-metric"><p>Today</p><strong>{myToday ? new Date(myToday.check_in).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '--:--'}</strong><span>{myToday?.check_out ? `Out ${new Date(myToday.check_out).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}` : myToday ? 'Currently working' : 'No check-in recorded'}</span></article>
      <article className="metric"><p>Records in range</p><strong>{items.length}</strong><span>{range.start} → {range.end}</span></article>
      <article className="metric"><p>Completed shifts</p><strong>{items.filter((item) => item.check_out).length}</strong><span>With check-out time</span></article>
    </section>
    <section className="panel table-panel">
      <div className="panel-heading"><div><p className="eyebrow">HISTORY</p><h2>{identity?.role === 'admin' ? 'Organisation attendance' : 'My attendance'}</h2></div><div className="date-filters"><label>From<input type="date" value={range.start} onChange={(e) => setRange({ ...range, start: e.target.value })}/></label><label>To<input type="date" min={range.start} value={range.end} onChange={(e) => setRange({ ...range, end: e.target.value })}/></label></div></div>
      {notice ? <p className="notice success-notice">{notice}</p> : null}
      {loading || error ? <PageState loading={loading} error={error}/> : items.length === 0 ? <PageState empty="No attendance records exist for this date range."/> : <div className="data-table-wrap"><table className="data-table"><thead><tr><th>Employee</th><th>Date</th><th>Check in</th><th>Check out</th><th>Duration</th><th>Status</th></tr></thead><tbody>{items.map((item) => <tr key={item.id}><td><strong>{item.employee_name}</strong><span>{item.department_name}</span></td><td>{item.attendance_date}</td><td>{new Date(item.check_in).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</td><td>{item.check_out ? new Date(item.check_out).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '—'}</td><td>{duration(item)}</td><td><StatusBadge value={item.check_out ? 'complete' : 'active'}/></td></tr>)}</tbody></table></div>}
    </section>
  </PortalShell>;
}
