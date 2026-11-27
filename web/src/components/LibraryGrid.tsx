import { Link } from "react-router-dom";
import type { LibrarySet } from "../types";
import { Icon } from "./Icon";

/** Browsable library sets: cover photo, a few of the words, and whether you have it already. */
export function LibraryGrid({ sets, compact = false, from }: { sets: LibrarySet[]; compact?: boolean; from?: "welcome" }) {
  return (
    <ul className={`grid gap-3 ${compact ? "grid-cols-[minmax(0,1fr)] sm:grid-cols-2" : "grid-cols-[minmax(0,1fr)] sm:grid-cols-2 lg:grid-cols-3"}`}>
      {sets.map((set) => (
        <li key={set.id} className="min-w-0">
          <Link to={`/library/${set.id}${from ? `?${from}` : ""}`} className="surface flex h-full gap-3 p-2.5 transition hover:shadow-lift sm:flex-col">
            <div className="relative h-24 w-28 shrink-0 overflow-hidden rounded-2xl bg-[#F1EEEA] sm:aspect-[16/9] sm:h-auto sm:w-full">
              {set.cover_url ? (
                <img src={set.cover_url} alt="" className="h-full w-full object-cover" />
              ) : (
                <div className="grid h-full place-items-center font-hanzi text-3xl font-bold text-[#BDB5AC]">{set.preview[0]}</div>
              )}
              {set.imported_set_id && (
                <span className="absolute left-2 top-2 flex items-center gap-1 rounded-full bg-ink px-2 py-0.5 text-xs font-bold text-white">
                  <Icon name="check" className="text-gold" /> Yours
                </span>
              )}
            </div>
            <div className="min-w-0 flex-1 py-0.5 sm:px-1 sm:pb-1">
              <div className="flex items-start justify-between gap-2">
                <b className="block min-w-0 text-[17px] leading-snug">{set.title}</b>
                {set.level && <span className="shrink-0 rounded-full bg-[#F1EEEA] px-2 py-0.5 text-xs font-semibold text-muted">{set.level}</span>}
              </div>
              <p className="truncate font-hanzi text-lg text-ink">{set.preview.join(" · ")}</p>
              <p className="text-sm text-muted">{set.card_count} words</p>
            </div>
          </Link>
        </li>
      ))}
    </ul>
  );
}
