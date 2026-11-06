import type { CSSProperties } from 'react';

export function SignalField({ compact = false }: { compact?: boolean }) {
  return (
    <svg
      aria-hidden="true"
      className={`signal-field ${compact ? 'signal-field-compact' : ''}`}
      viewBox="0 0 640 520"
      preserveAspectRatio="xMidYMid slice"
    >
      <defs>
        <linearGradient id="signal-line" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0" stopColor="#1E6BFF" stopOpacity=".12"/>
          <stop offset=".48" stopColor="#56B7FF" stopOpacity=".82"/>
          <stop offset="1" stopColor="#8BD8FF" stopOpacity=".16"/>
        </linearGradient>
        <radialGradient id="signal-node">
          <stop offset="0" stopColor="#fff"/>
          <stop offset=".22" stopColor="#8DDBFF"/>
          <stop offset=".58" stopColor="#238BFF"/>
          <stop offset="1" stopColor="#146EF5" stopOpacity="0"/>
        </radialGradient>
        <filter id="signal-glow" x="-200%" y="-200%" width="400%" height="400%">
          <feGaussianBlur stdDeviation="7"/>
        </filter>
      </defs>

      <g className="signal-grid" opacity=".22">
        <path d="M0 104h640M0 208h640M0 312h640M0 416h640"/>
        <path d="M128 0v520M256 0v520M384 0v520M512 0v520"/>
      </g>

      <g className="signal-links" stroke="url(#signal-line)" strokeWidth="1.2">
        <path pathLength="1" d="M96 334 190 245 286 292 375 174 472 231 554 135"/>
        <path pathLength="1" d="M190 245 235 139 375 174 440 342 548 384"/>
        <path pathLength="1" d="M96 334 215 408 440 342 472 231"/>
      </g>

      <g className="signal-nodes">
        {[
          [96,334,20],[190,245,28],[286,292,20],[375,174,30],[472,231,23],
          [554,135,18],[235,139,16],[440,342,27],[215,408,17],[548,384,20],
        ].map(([cx, cy, r], index) => (
          <g key={`${cx}-${cy}`} style={{ '--node-delay': `${index * 130}ms` } as CSSProperties}>
            <circle cx={cx} cy={cy} r={r} fill="url(#signal-node)" opacity=".35" filter="url(#signal-glow)"/>
            <circle className="signal-node-ring" cx={cx} cy={cy} r="7" fill="none" stroke="#69C5FF" strokeOpacity=".42"/>
            <circle cx={cx} cy={cy} r="3.4" fill="#BCEBFF"/>
          </g>
        ))}
      </g>
    </svg>
  );
}
