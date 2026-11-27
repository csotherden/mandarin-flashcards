import { useEffect, useState } from "react";
import { Link, useNavigate, useParams, useSearchParams } from "react-router-dom";
import { api, ApiError } from "../api";
import { CardImage } from "../components/CardImage";
import { Icon } from "../components/Icon";
import { Pinyin } from "../components/Pinyin";
import { practiceModes, reviewPath, startSession, type PracticeMode } from "../session";
import type { Flashcard, LibraryCard, LibraryImport, LibrarySetDetail } from "../types";

/** CardImage and friends take a Flashcard; a library card has everything they read. */
function asCard(card: LibraryCard): Flashcard {
  return { ...card, has_local_image: false, category_ids: [], set_ids: [], created_at: "" };
}

/**
 * One library set: its words, and two ways in. "Start lesson" adds it to your library (if it
 * isn't already) and launches a session straight away; "Add to my library" only adds it.
 */
export function LibrarySetPage() {
  const { id = "" } = useParams();
  const [params] = useSearchParams();
  const welcome = params.has("welcome");
  const navigate = useNavigate();
  const [set, setSet] = useState<LibrarySetDetail | null>(null);
  const [mode, setMode] = useState<PracticeMode>("image_speak");
  const [busy, setBusy] = useState<"" | "start" | "add">("");
  const [error, setError] = useState("");

  useEffect(() => {
    api<LibrarySetDetail>(`/api/library/sets/${id}`).then(setSet).catch((err: Error) => setError(err.message));
  }, [id]);

  async function hydrate(): Promise<string> {
    if (set?.imported_set_id) return set.imported_set_id;
    const res = await api<LibraryImport>(`/api/library/sets/${id}/import`, { method: "POST" });
    setSet((current) => (current ? { ...current, imported_set_id: res.set.id } : current));
    return res.set.id;
  }

  async function start() {
    setBusy("start");
    setError("");
    try {
      const setId = await hydrate();
      if (mode === "review") navigate(reviewPath({ setIds: [setId] }, 0, `/sets/${setId}`));
      else navigate(`/drill/${await startSession({ kind: "sets", setIds: [setId] }, mode, 0)}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not start the lesson");
      setBusy("");
    }
  }

  async function add() {
    setBusy("add");
    setError("");
    try {
      navigate(`/sets/${await hydrate()}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not add the set");
      setBusy("");
    }
  }

  if (!set) return <p className="text-muted">{error || "Loading…"}</p>;
  const have = set.cards.filter((card) => card.owned).length;

  return (
    <div>
      <Link to={welcome ? "/welcome" : "/library"} className="mb-2 flex min-h-11 w-fit items-center gap-1 font-semibold text-muted">
        <Icon name="back" /> {welcome ? "Back" : "Library"}
      </Link>

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.1fr)] lg:items-start">
        <section className="overflow-hidden rounded-[28px] bg-hero text-white lg:sticky lg:top-8">
          {set.cover_url && <img src={set.cover_url} alt="" className="aspect-[16/9] w-full object-cover" />}
          <div className="p-6">
            <div className="flex flex-wrap items-center gap-2 text-sm font-semibold">
              {set.level && <span className="rounded-full bg-white/15 px-3 py-1">{set.level}</span>}
              <span className="rounded-full bg-white/15 px-3 py-1">{set.card_count} words</span>
              {set.imported_set_id && (
                <span className="flex items-center gap-1 rounded-full bg-ink px-3 py-1">
                  <Icon name="check" className="text-gold" /> In your library
                </span>
              )}
            </div>
            <h1 className="mt-3 text-3xl font-extrabold leading-tight">{set.title}</h1>
            {set.description && <p className="mt-1 text-white/90">{set.description}</p>}

            <p className="mt-5 text-sm font-semibold uppercase tracking-[0.08em] text-white/85">How to practice</p>
            <div className="mt-2 grid gap-2">
              {practiceModes.map((item) => (
                <button
                  key={item.id}
                  type="button"
                  aria-pressed={mode === item.id}
                  onClick={() => setMode(item.id)}
                  className={`flex items-center justify-between gap-3 rounded-2xl border-2 px-4 py-2.5 text-left transition ${mode === item.id ? "border-white bg-white/15" : "border-white/25"}`}
                >
                  <span>
                    <b className="block">{item.title}</b>
                    <span className="block text-sm text-white/85">{item.detail}</span>
                  </span>
                  {mode === item.id && <Icon name="check" className="shrink-0 text-xl text-gold" />}
                </button>
              ))}
            </div>

            {error && <p className="mt-3 rounded-xl bg-white px-3 py-2 text-sm font-semibold text-crimson" role="alert">{error}</p>}
            <button type="button" disabled={!!busy} onClick={start} className="mt-5 flex h-14 w-full items-center justify-center gap-2 rounded-2xl bg-white text-lg font-extrabold text-maroon disabled:opacity-70">
              <Icon name="play" /> {busy === "start" ? "Starting…" : mode === "review" ? "Start review" : "Start lesson"}
            </button>
            {set.imported_set_id ? (
              <Link to={`/sets/${set.imported_set_id}`} className="mt-1 flex min-h-11 items-center justify-center font-semibold text-white underline underline-offset-4">
                Open your copy
              </Link>
            ) : (
              <button type="button" disabled={!!busy} onClick={add} className="mt-1 block min-h-11 w-full text-center font-semibold text-white underline underline-offset-4">
                {busy === "add" ? "Adding…" : "Just add it to my library"}
              </button>
            )}
          </div>
        </section>

        <section className="min-w-0">
          <div className="mb-3 flex items-baseline justify-between gap-2">
            <h2 className="font-bold">Words in this set</h2>
            {have > 0 && <span className="text-sm text-muted">You have {have} already</span>}
          </div>
          <ul className="grid grid-cols-2 gap-3 sm:grid-cols-3">
            {set.cards.map((card) => (
              <li key={card.id} className="surface relative min-w-0 overflow-hidden">
                <CardImage card={asCard(card)} className="aspect-[4/3] w-full" />
                {card.owned && (
                  <span className="absolute right-2 top-2 grid h-7 w-7 place-items-center rounded-full bg-ink text-sm text-gold" title="Already in your cards" aria-label="Already in your cards">
                    <Icon name="check" />
                  </span>
                )}
                <div className="px-3 pb-3 pt-2">
                  <p className="truncate font-hanzi text-2xl font-bold">{card.hanzi}</p>
                  <Pinyin tonal={card.tonal_pinyin} plain={card.pinyin} className="block truncate text-sm font-semibold" />
                  <p className="truncate text-sm text-muted">{card.english}</p>
                </div>
              </li>
            ))}
          </ul>
        </section>
      </div>
    </div>
  );
}
