import { useEffect, useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api, ApiError } from "../api";
import { CardImage } from "../components/CardImage";
import { Icon } from "../components/Icon";
import { Sheet } from "../components/Sheet";
import type { CardSet, Home } from "../types";

/** Sets: generate one from a theme, start an empty one, or open one to practice. */
export function SetsPage() {
  const navigate = useNavigate();
  const [home, setHome] = useState<Home | null>(null);
  const [creating, setCreating] = useState(false);
  const [title, setTitle] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    // Home carries each set with a cover card, which is what this grid shows.
    api<Home>(`/api/home?tz_offset=${new Date().getTimezoneOffset()}`)
      .then(setHome)
      .catch((err: Error) => setError(err.message));
  }, []);

  async function create(event: FormEvent) {
    event.preventDefault();
    setError("");
    try {
      const set = await api<CardSet>("/api/sets", { method: "POST", body: JSON.stringify({ title, description: "" }) });
      navigate(`/sets/${set.id}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not create the set");
    }
  }

  return (
    <div>
      <h1 className="mb-4 text-2xl font-extrabold lg:text-3xl">Sets</h1>
      <div className="grid gap-3 sm:grid-cols-2">
        <Link to="/sets/generate" className="relative overflow-hidden rounded-[24px] bg-hero p-5 text-white">
          <Icon name="spark" className="absolute right-4 top-4 text-3xl text-gold" />
          <b className="block pr-10 text-xl font-extrabold">Generate a set</b>
          <span className="mt-1 block text-white/90">Describe a theme and get 5, 10 or 15 new cards.</span>
        </Link>
        <button type="button" onClick={() => setCreating(true)} className="surface p-5 text-left">
          <Icon name="plus" className="float-right text-2xl text-crimson" />
          <b className="block text-xl font-extrabold">New empty set</b>
          <span className="mt-1 block text-muted">Name it, then add cards you already have.</span>
        </button>
      </div>

      {error && <p className="mt-3 text-sm font-semibold text-crimson" role="alert">{error}</p>}
      {home && home.sets.length === 0 && <p className="mt-8 text-center text-muted">No sets yet.</p>}
      <ul className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
        {(home?.sets ?? []).map(({ set, cover }) => (
          <li key={set.id} className="min-w-0">
            <Link to={`/sets/${set.id}`} className="surface block p-2.5 transition hover:shadow-lift">
              {cover ? <CardImage card={cover} className="aspect-[4/3] w-full rounded-2xl" /> : <div className="grid aspect-[4/3] place-items-center rounded-2xl bg-[#F1EEEA] text-3xl text-[#BDB5AC]"><Icon name="sets" /></div>}
              <b className="mt-2.5 block truncate px-1">{set.title}</b>
              <span className="block px-1 pb-1 text-sm text-muted">{set.card_count} {set.card_count === 1 ? "card" : "cards"}</span>
            </Link>
          </li>
        ))}
      </ul>

      <Sheet open={creating} onClose={() => setCreating(false)} title="New set">
        <form onSubmit={create} className="grid gap-4">
          <input autoFocus className="field" placeholder="Section 3 sticky words" aria-label="Set name" value={title} onChange={(event) => setTitle(event.target.value)} />
          <button type="submit" className="btn-primary h-13 py-3.5" disabled={!title.trim()}>Create set</button>
        </form>
      </Sheet>
    </div>
  );
}
