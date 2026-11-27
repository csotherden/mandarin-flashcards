import { useEffect, useState } from "react";
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
        <div className="mt-3 grid grid-cols-3 gap-3">
          <Stat value={home.streak_days} label={home.streak_today || home.streak_days === 0 ? "day streak" : "day streak, practice today"} />
          <Stat value={home.streak_longest} label="longest streak" />
          <Stat value={home.total_cards} label="cards" />
        </div>
      )}
      <button type="button" className="btn-ghost mt-6 h-12 w-full" onClick={logout}>
        <Icon name="back" /> Sign out
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
