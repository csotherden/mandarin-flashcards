import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api";
import { useAuth } from "../auth";
import { Icon } from "../components/Icon";
import type { Home } from "../types";

export function MePage() {
  const { user, logout } = useAuth();
  const [home, setHome] = useState<Home | null>(null);

  useEffect(() => {
    api<Home>("/api/home").then(setHome).catch(() => undefined);
  }, []);

  return (
    <div className="mx-auto max-w-xl">
      <h1 className="mb-5 text-2xl font-extrabold lg:text-3xl">Me</h1>
      <div className="surface p-5">
        <p className="label">Signed in as</p>
        <p className="mt-1 break-all font-semibold">{user?.email}</p>
      </div>
      {home && (
        <div className="mt-3 grid grid-cols-2 gap-3">
          <Stat value={home.streak_days} label={home.streak_today || home.streak_days === 0 ? "day streak" : "day streak, practice today"} />
          <Stat value={home.streak_longest} label="longest streak" />
          <Stat value={home.xp_total} label="total XP" />
          <Stat value={home.total_cards} label="cards" />
        </div>
      )}
      <h2 className="mb-2 mt-7 font-bold">Keep learning</h2>
      <div className="surface divide-y divide-line">
        <MeLink to="/tones" glyph="声" title="Tone practice" detail={home ? (home.tones_next ? `Intro to tones · ${home.tones_completed} of ${home.tones_total} done` : "Revisit any tones lesson") : ""} />
        <MeLink to="/library" glyph="书" title="Library" detail="Ready-made sets to practice or add" />
        <MeLink to="/cards" glyph="卡" title="Your cards" detail={home ? `${home.total_cards} cards` : ""} />
      </div>
      <button type="button" className="btn-ghost mt-6 h-12 w-full" onClick={logout}>
        <Icon name="logout" /> Sign out
      </button>
    </div>
  );
}

function Stat({ value, label }: { value: number; label: string }) {
  return (
    <div className="surface p-4 text-center">
      <b className="block text-3xl font-extrabold">{value}</b>
      <span className="text-sm text-muted">{label}</span>
    </div>
  );
}

function MeLink({ to, glyph, title, detail }: { to: string; glyph: string; title: string; detail: string }) {
  return (
    <Link to={to} className="flex items-center gap-3 px-4 py-3">
      <span className="grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-ink font-hanzi text-lg font-bold text-gold" aria-hidden>
        {glyph}
      </span>
      <span className="min-w-0 flex-1">
        <b className="block">{title}</b>
        <span className="block truncate text-sm text-muted">{detail}</span>
      </span>
      <Icon name="right" className="shrink-0 text-lg text-muted" />
    </Link>
  );
}
