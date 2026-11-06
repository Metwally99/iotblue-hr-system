'use client';

import { FormEvent, useMemo, useState } from 'react';
import { useAuth } from '@/components/auth-provider';
import { PageState, PortalShell } from '@/components/portal-shell';
import { api, ApiError } from '@/lib/api';
import type { AttendanceReportRow } from '@/lib/types';

export default function ReportsPage() {
  const { token } = useAuth();
  const today = useMemo(() => new Date().toISOString().slice(0, 10), []);
  const [range, setRange] = useState({ start: `${today.slice(0, 8)}01`, end: today });
  const [rows, setRows] = useState<AttendanceReportRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState('');

  async function generate(event?: FormEvent) {
    event?.preventDefault(); if (!token) return;
    setLoading(true); setError(''); setLoaded(true);
    try { setRows((await api<{ data: AttendanceReportRow[] }>(`/attendance/report?start_date=${range.start}&end_date=${range.end}`, {}, token)).data); }
    catch (caught) { setError(caught instanceof ApiError ? caught.message : 'The report could not be generated.'); }
    finally { setLoading(false); }
  }

  const totalHours = rows.reduce((sum, row) => sum + row.total_hours, 0);
  const totalDays = rows.reduce((sum, row) => sum + row.present_days, 0);
  return <PortalShell adminOnly eyebrow="INSIGHTS" title="Attendance reports">
    <form className="panel report-builder" onSubmit={generate}>
      <div><p className="eyebrow">REPORT PERIOD</p><h2>Choose a date range</h2><p>Aggregation is calculated by PostgreSQL from completed and open attendance records.</p></div>
      <label>Start date<input required type="date" value={range.start} onChange={(e) => setRange({ ...range, start: e.target.value })}/></label>
      <label>End date<input required min={range.start} type="date" value={range.end} onChange={(e) => setRange({ ...range, end: e.target.value })}/></label>
      <button className="primary" disabled={loading}>{loading ? 'Generating…' : 'Generate report'}</button>
    </form>
    {loaded && !loading && !error && rows.length > 0 ? <section className="attendance-summary report-summary"><article className="metric hero-metric"><p>Total hours</p><strong>{totalHours.toFixed(1)}</strong><span>Across selected employees</span></article><article className="metric"><p>Present days</p><strong>{totalDays}</strong><span>Attendance records</span></article><article className="metric"><p>Employees</p><strong>{rows.length}</strong><span>With attendance in range</span></article></section> : null}
    <section className="panel table-panel report-table">
      <div className="panel-heading"><div><p className="eyebrow">RESULTS</p><h2>Employee totals</h2></div>{rows.length ? <span className="export-note">Ready to print from your browser</span> : null}</div>
      {loading || error ? <PageState loading={loading} error={error}/> : !loaded ? <PageState empty="Choose a date range and generate your first report."/> : rows.length === 0 ? <PageState empty="No attendance data exists in this date range."/> : <div className="data-table-wrap"><table className="data-table"><thead><tr><th>Employee</th><th>Department</th><th>Present days</th><th>Completed days</th><th>Total hours</th></tr></thead><tbody>{rows.map((row) => <tr key={row.employee_id}><td><strong>{row.employee_name}</strong></td><td>{row.department_name}</td><td>{row.present_days}</td><td>{row.completed_days}</td><td><strong>{row.total_hours.toFixed(2)} h</strong></td></tr>)}</tbody></table></div>}
    </section>
  </PortalShell>;
}
