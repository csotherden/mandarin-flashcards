// Small stroke icons, sized by font-size (1em) unless a className sets width and height.
import type { ReactElement } from "react";
const paths: Record<string, ReactElement> = {
  home: <path d="M4 10.5 12 4l8 6.5V20a1 1 0 0 1-1 1h-4.5v-6h-5v6H5a1 1 0 0 1-1-1z" strokeLinejoin="round" />,
  cards: (
    <g strokeLinejoin="round">
      <rect x="7" y="3.5" width="12" height="15" rx="2.5" />
      <path d="M5 7v11.5A2.5 2.5 0 0 0 7.5 21H15" />
    </g>
  ),
  sets: (
    <g strokeLinejoin="round">
      <rect x="3.5" y="3.5" width="7" height="7" rx="2" />
      <rect x="13.5" y="3.5" width="7" height="7" rx="2" />
      <rect x="3.5" y="13.5" width="7" height="7" rx="2" />
      <rect x="13.5" y="13.5" width="7" height="7" rx="2" />
    </g>
  ),
  plus: <path d="M12 5v14M5 12h14" strokeLinecap="round" />,
  close: <path d="M6 6l12 12M18 6 6 18" strokeLinecap="round" />,
  check: <path d="m5 12.5 4.5 4.5L19 7.5" strokeLinecap="round" strokeLinejoin="round" />,
  left: <path d="m14.5 6-6 6 6 6" strokeLinecap="round" strokeLinejoin="round" />,
  right: <path d="m9.5 6 6 6-6 6" strokeLinecap="round" strokeLinejoin="round" />,
  mic: (
    <g strokeLinecap="round">
      <rect x="9" y="3" width="6" height="11" rx="3" />
      <path d="M5.5 11a6.5 6.5 0 0 0 13 0M12 17.5V21" />
    </g>
  ),
  volume: <path d="M4 9.5h3.5L12 5.5v13l-4.5-4H4zM16 9a4 4 0 0 1 0 6" strokeLinejoin="round" />,
  search: (
    <g strokeLinecap="round">
      <circle cx="11" cy="11" r="6.5" />
      <path d="m20 20-4.2-4.2" />
    </g>
  ),
  user: (
    <g strokeLinecap="round">
      <circle cx="12" cy="8.5" r="3.5" />
      <path d="M5 20a7 7 0 0 1 14 0" />
    </g>
  ),
  edit: <path d="M4 20h4L19 9l-4-4L4 16zM13.5 6.5l4 4" strokeLinejoin="round" />,
  trash: <path d="M5 7h14M10 11v6M14 11v6M6 7l1 13h10l1-13M9 7V4h6v3" strokeLinecap="round" strokeLinejoin="round" />,
  back: <path d="M15 5 8 12l7 7" strokeLinecap="round" strokeLinejoin="round" />,
};

const filled: Record<string, ReactElement> = {
  star: <path d="M12 2.5l2.9 6.2 6.8.8-5 4.6 1.3 6.7L12 17.5l-6 3.3 1.3-6.7-5-4.6 6.8-.8z" />,
  play: <path d="M8 5.5v13l10.5-6.5z" />,
  spark: <path d="M12 2l1.8 5.5L19 9.3l-5.2 1.8L12 16.5l-1.8-5.4L5 9.3l5.2-1.8z" />,
};

export type IconName = keyof typeof paths | keyof typeof filled;

export function Icon({ name, className = "", label }: { name: IconName; className?: string; label?: string }) {
  const solid = filled[name];
  return (
    <svg
      viewBox="0 0 24 24"
      width="1em"
      height="1em"
      className={`inline-block shrink-0 ${className}`}
      fill={solid ? "currentColor" : "none"}
      stroke={solid ? "none" : "currentColor"}
      strokeWidth={2}
      aria-hidden={label ? undefined : true}
      role={label ? "img" : undefined}
      aria-label={label}
    >
      {solid ?? paths[name]}
    </svg>
  );
}
