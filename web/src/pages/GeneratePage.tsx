import { useEffect, useState, type FormEvent } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { api, ApiError } from "../api";
import { CardImage } from "../components/CardImage";
import { Icon } from "../components/Icon";
import { Pinyin } from "../components/Pinyin";
import { SessionSheet } from "../components/SessionSheet";
import type { CardSet, GeneratedSet } from "../types";

// Themes the starter library doesn't already cover; basics like family or food live there ready-made.
const ideas = ["At the market", "In the kitchen", "Feelings", "Hobbies", "At work", "At the doctor"];

/** Describe a theme, pick a size, and get a finished set of new words (none you already have). */
export function GeneratePage() {
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const setId = params.get("set");
  const [target, setTarget] = useState<CardSet | null>(null);
  const [theme, setTheme] = useState("");
  const [count, setCount] = useState(10);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [result, setResult] = useState<GeneratedSet | null>(null);
  const [practice, setPractice] = useState(false);

  useEffect(() => {
    if (!setId) return;
    api<CardSet>(`/api/sets/${setId}`)
      .then((set) => {
        setTarget(set);
        setTheme(set.title);
      })
      .catch(() => setTarget(null));
  }, [setId]);

  async function generate(event: FormEvent) {
    event.preventDefault();
    setBusy(true);
    setError("");
    try {
      setResult(
        await api<GeneratedSet>("/api/sets/generate", {
          method: "POST",
          body: JSON.stringify({ theme: theme.trim(), count, set_id: setId || undefined }),
        }),
      );
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not generate the set");
    } finally {
      setBusy(false);
    }
  }

  if (result) {
    return (
      <div>
        <div className="rise relative overflow-hidden rounded-[28px] bg-hero p-6 text-white lg:p-8">
          <Icon name="spark" className="absolute right-5 top-5 text-4xl text-gold" />
          <p className="text-sm font-semibold uppercase tracking-[0.08em] text-white/85">{target ? "Added to" : "New set"}</p>
          <h1 className="mt-1 pr-12 text-3xl font-extrabold leading-tight">{result.set.title}</h1>
          <p className="mt-2 text-white/90">
            {result.cards.length} new {result.cards.length === 1 ? "card" : "cards"}
            {result.skipped.length > 0 && ` · skipped ${result.skipped.length} you already had`}
          </p>
          <button type="button" onClick={() => setPractice(true)} className="mt-5 flex h-14 w-full items-center justify-center gap-2 rounded-2xl bg-white text-lg font-extrabold text-maroon sm:w-auto sm:px-8">
            <Icon name="play" /> Practice this set
          </button>
        </div>
        <ul className="mt-5 grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
          {result.cards.map((card) => (
            <li key={card.id} className="min-w-0">
              <Link to={`/cards/${card.id}`} className="surface block overflow-hidden">
                <CardImage card={card} className="aspect-[4/3] w-full" />
                <div className="px-3 pb-3 pt-2">
                  <p className="truncate font-hanzi text-2xl font-bold">{card.hanzi}</p>
                  <Pinyin tonal={card.tonal_pinyin} plain={card.pinyin} className="block truncate text-sm font-semibold" />
                  <p className="truncate text-sm text-muted">{card.english}</p>
                </div>
              </Link>
            </li>
          ))}
        </ul>
        <div className="mt-5 flex flex-wrap gap-3">
          <Link to={`/sets/${result.set.id}`} className="btn-ghost h-12">Open the set</Link>
          <button type="button" className="btn-soft h-12" onClick={() => { setResult(null); setTheme(""); }}>
            Generate another
          </button>
        </div>
        <SessionSheet open={practice} onClose={() => setPractice(false)} sets={[{ ...result.set, card_count: Math.max(result.set.card_count, result.cards.length) }]} initialSetIds={[result.set.id]} />
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-xl">
      <button type="button" className="mb-2 flex min-h-11 items-center gap-1 font-semibold text-muted" onClick={() => navigate(-1)}>
        <Icon name="back" /> Back
      </button>
      <h1 className="text-2xl font-extrabold lg:text-3xl">{target ? `More for ${target.title}` : "Generate a set"}</h1>
      <p className="mt-1 text-muted">Describe a theme. You get everyday words you don't have yet, with photos.</p>
      <form onSubmit={generate} className="mt-5 grid gap-5">
        <div>
          <label className="label" htmlFor="theme">Theme</label>
          <input id="theme" autoFocus className="field mt-2 text-lg" placeholder="Things in the kitchen" value={theme} onChange={(event) => setTheme(event.target.value)} maxLength={80} />
          {!target && (
            <>
              <div className="mt-2 flex flex-wrap gap-2">
                {ideas.map((idea) => (
                  <button key={idea} type="button" className="chip bg-white" onClick={() => setTheme(idea)}>{idea}</button>
                ))}
              </div>
              <p className="mt-3 text-sm text-muted">
                Looking for basics like family, food or animals?{" "}
                <Link to="/library" className="link">The library has them ready-made.</Link>
              </p>
            </>
          )}
        </div>
        <div>
          <p className="label">How many cards</p>
          <div className="mt-2 grid grid-cols-3 gap-2">
            {[5, 10, 15].map((n) => (
              <button key={n} type="button" aria-pressed={count === n} onClick={() => setCount(n)} className={`h-14 rounded-2xl border-2 text-xl font-extrabold ${count === n ? "border-crimson bg-crimson text-white" : "border-line bg-white"}`}>
                {n}
              </button>
            ))}
          </div>
        </div>
        {error && <p className="text-sm font-semibold text-crimson" role="alert">{error}</p>}
        <button type="submit" className="btn-primary h-14 text-lg" disabled={busy || theme.trim().length < 2}>
          <Icon name="spark" /> {busy ? `Making ${count} cards…` : "Generate"}
        </button>
        {busy && <p className="-mt-2 text-center text-sm text-muted">Choosing words, filling pinyin and finding photos. This takes a little while.</p>}
      </form>
    </div>
  );
}
