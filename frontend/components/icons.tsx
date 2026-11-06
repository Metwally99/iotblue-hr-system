import type { ReactNode, SVGProps } from 'react';

export type IconName =
  | 'overview'
  | 'people'
  | 'departments'
  | 'leave'
  | 'attendance'
  | 'reports'
  | 'plus'
  | 'logout'
  | 'menu'
  | 'close'
  | 'arrow'
  | 'check'
  | 'calendar'
  | 'pulse';

const paths: Record<IconName, ReactNode> = {
  overview: <><rect x="3" y="3" width="7" height="7" rx="2"/><rect x="14" y="3" width="7" height="5" rx="2"/><rect x="14" y="12" width="7" height="9" rx="2"/><rect x="3" y="14" width="7" height="7" rx="2"/></>,
  people: <><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"/></>,
  departments: <><path d="M3 21h18M5 21V7l7-4 7 4v14M9 9h.01M15 9h.01M9 13h.01M15 13h.01M10 21v-4h4v4"/></>,
  leave: <><rect x="3" y="5" width="18" height="16" rx="2"/><path d="M16 3v4M8 3v4M3 11h18M8 15h.01M12 15h.01M16 15h.01"/></>,
  attendance: <><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2M7 3 4-2M17 3l-4-2"/></>,
  reports: <><path d="M4 19V9M10 19V5M16 19v-7M22 19H2"/></>,
  plus: <path d="M12 5v14M5 12h14"/>,
  logout: <><path d="M10 17l5-5-5-5M15 12H3M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/></>,
  menu: <><path d="M4 7h16M4 12h16M4 17h16"/></>,
  close: <><path d="M6 6l12 12M18 6 6 18"/></>,
  arrow: <><path d="M5 12h14M13 6l6 6-6 6"/></>,
  check: <path d="m5 12 4 4L19 6"/>,
  calendar: <><rect x="3" y="5" width="18" height="16" rx="2"/><path d="M16 3v4M8 3v4M3 11h18"/></>,
  pulse: <path d="M3 12h4l2.5-6 5 12 2.5-6h4"/>,
};

export function Icon({ name, ...props }: { name: IconName } & SVGProps<SVGSVGElement>) {
  return (
    <svg aria-hidden="true" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" {...props}>
      {paths[name]}
    </svg>
  );
}

export function BrandMark({ className = '' }: { className?: string }) {
  return (
    <svg aria-hidden="true" className={className} viewBox="0 0 40 40" fill="none">
      <rect width="40" height="40" rx="12" fill="currentColor"/>
      <path d="M12 11v18M28 11v18M12 20h16" stroke="white" strokeWidth="2.6" strokeLinecap="round"/>
      <circle cx="12" cy="11" r="2.4" fill="#7CC4FF"/>
      <circle cx="28" cy="11" r="2.4" fill="#7CC4FF"/>
      <circle cx="12" cy="29" r="2.4" fill="#7CC4FF"/>
      <circle cx="28" cy="29" r="2.4" fill="#7CC4FF"/>
      <circle cx="20" cy="20" r="2.8" fill="white"/>
    </svg>
  );
}
