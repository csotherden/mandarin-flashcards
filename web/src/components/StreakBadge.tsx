import { Icon } from "./Icon";

/**
 * Three looks: no streak; a streak that still needs today's practice (outlined, crimson ring);
 * and a streak that today already counts toward (solid ink with a gold star and check).
 */
export function StreakBadge({ days, today }: { days: number; today: boolean }) {
  if (days === 0 && !today) {
    return (
      <span className="flex shrink-0 items-center gap-1.5 whitespace-nowrap rounded-full bg-ink px-3.5 py-2 text-sm font-semibold text-white">
        <Icon name="star-outline" className="text-gold" />
        Start a streak
      </span>
    );
  }
  if (!today) {
    return (
      <span
        className="flex shrink-0 items-center gap-1.5 whitespace-nowrap rounded-full border-2 border-crimson bg-white px-3 py-1.5 text-sm font-semibold text-ink"
        title="Practice today to keep it"
        aria-label={`${days}-day streak, practice today to keep it`}
      >
        <Icon name="star-outline" className="text-crimson" />
        <b>{days}</b> day streak
        <span className="h-2 w-2 animate-pulse rounded-full bg-crimson" aria-hidden />
      </span>
    );
  }
  return (
    <span
      className="flex shrink-0 items-center gap-1.5 whitespace-nowrap rounded-full bg-ink px-3.5 py-2 text-sm font-semibold text-white"
      aria-label={`${days}-day streak, today done`}
    >
      <Icon name="star" className="text-gold" />
      <b className="text-gold">{days}</b> day streak
      <Icon name="check" className="text-gold" />
    </span>
  );
}
