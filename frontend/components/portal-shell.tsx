'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import { useAuth } from '@/components/auth-provider';
import { BrandMark, Icon, type IconName } from '@/components/icons';

const navigation = [
  { href: '/', label: 'Overview', icon: 'overview', adminOnly: false },
  { href: '/employees', label: 'Employees', icon: 'people', adminOnly: true },
  { href: '/departments', label: 'Departments', icon: 'departments', adminOnly: true },
  { href: '/leave', label: 'Leave requests', icon: 'leave', adminOnly: false },
  { href: '/attendance', label: 'Attendance', icon: 'attendance', adminOnly: false },
  { href: '/reports', label: 'Reports', icon: 'reports', adminOnly: true },
] satisfies Array<{ href: string; label: string; icon: IconName; adminOnly: boolean }>;

export function PortalShell({
  title,
  eyebrow,
  adminOnly = false,
  action,
  children,
}: {
  title: string;
  eyebrow: string;
  adminOnly?: boolean;
  action?: React.ReactNode;
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const router = useRouter();
  const { identity, loading, logout } = useAuth();
  const [menuOpen, setMenuOpen] = useState(false);

  useEffect(() => {
    if (!loading && !identity) router.replace('/login');
  }, [identity, loading, router]);

  if (loading || !identity) {
    return <main className="loading-screen"><span className="spinner"/>Opening your workspace…</main>;
  }

  const initials = identity.name.split(' ').map((word) => word[0]).join('').slice(0, 2).toUpperCase();
  const forbidden = adminOnly && identity.role !== 'admin' && identity.role !== 'hr';

  return (
    <main className="app-shell">
      <button
        aria-controls="portal-navigation"
        aria-expanded={menuOpen}
        aria-label="Open navigation"
        className="mobile-nav-toggle"
        onClick={() => setMenuOpen(true)}
      >
        <Icon name="menu"/>
      </button>
      <button aria-label="Close navigation" className={`sidebar-scrim ${menuOpen ? 'visible' : ''}`} onClick={() => setMenuOpen(false)}/>
      <aside className={`sidebar ${menuOpen ? 'sidebar-open' : ''}`} id="portal-navigation">
        <div className="sidebar-header">
          <Link aria-label="Humanly overview" className="brand" href="/">
            <BrandMark className="brand-mark"/>
            <span className="brand-copy"><strong>Humanly</strong><small>People intelligence</small></span>
          </Link>
          <button aria-label="Close navigation" className="sidebar-close" onClick={() => setMenuOpen(false)}><Icon name="close"/></button>
        </div>
        <nav aria-label="Main navigation">
          <p className="nav-label">Operations</p>
          {navigation.filter((item) => !item.adminOnly || identity.role === 'admin' || identity.role === 'hr').map((item) => (
            <Link
              aria-current={pathname === item.href ? 'page' : undefined}
              className={`nav-item ${pathname === item.href ? 'active' : ''}`}
              href={item.href}
              key={item.href}
              onClick={() => setMenuOpen(false)}
            >
              <Icon className="nav-icon" name={item.icon}/>
              <span>{item.label}</span>
              {pathname === item.href ? <span aria-hidden="true" className="nav-signal"/> : null}
            </Link>
          ))}
        </nav>
        <div className="sidebar-network" aria-hidden="true"><i/><i/><i/><i/><span/><span/><span/></div>
        <div className="sidebar-footer">
          <div className="avatar dark">{initials}</div>
          <div><strong>{identity.name}</strong><span>{identity.role === 'admin' ? 'Administrator' : identity.role === 'hr' ? 'HR' : identity.role === 'manager' ? 'Manager' : 'Employee'}</span></div>
          <button aria-label="Sign out" onClick={logout}><Icon name="logout"/></button>
        </div>
      </aside>
      <section className="workspace">
        <header className="topbar">
          <div><p className="eyebrow">{eyebrow}</p><h1>{title}</h1></div>
          <div className="topbar-right">
            <span className="workspace-status"><i/> Workspace connected</span>
            {action ? <div className="top-actions">{action}</div> : null}
          </div>
        </header>
        <div className="page-stage">
          {forbidden ? (
            <section className="empty-state"><strong>Administrator access required</strong><p>Your account cannot open this workspace.</p></section>
          ) : children}
        </div>
      </section>
    </main>
  );
}

export function PageState({ loading, error, empty }: { loading?: boolean; error?: string; empty?: string }) {
  if (loading) return <section className="empty-state"><span className="spinner"/><p>Loading data…</p></section>;
  if (error) return <section className="empty-state error-state"><strong>We could not load this page</strong><p>{error}</p></section>;
  if (empty) return <section className="empty-state"><strong>Nothing here yet</strong><p>{empty}</p></section>;
  return null;
}

export function StatusBadge({ value }: { value: string }) {
  return <span className={`status-badge status-${value.toLowerCase()}`}>{value.replace('_', ' ')}</span>;
}
