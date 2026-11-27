import { useCallback, useEffect, useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { api, ApiError } from "../api";
import { CardImage } from "../components/CardImage";
import { Icon } from "../components/Icon";
import { Pinyin } from "../components/Pinyin";
import { Sheet } from "../components/Sheet";
import type { CardSet, Category, Flashcard } from "../types";

const sorts = [
  { id: "newest", label: "Newest" },
  { id: "oldest", label: "Oldest" },
  { id: "az", label: "A–Z (pinyin)" },
  { id: "weakest", label: "Weakest first" },
];

function scoreClass(score: number) {
  return score >= 85 ? "bg-tone3 text-white" : score >= 60 ? "bg-gold text-ink" : "bg-crimson text-white";
}

/** The word library: search, filter by set or category, sort, and select cards to act on together. */
export function CardsPage() {
  const [cards, setCards] = useState<Flashcard[] | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);
  const [sets, setSets] = useState<CardSet[]>([]);
  const [q, setQ] = useState("");
  const [categoryID, setCategoryID] = useState("");
  const [setID, setSetID] = useState("");
  const [sort, setSort] = useState("newest");
  const [error, setError] = useState("");
  const [selecting, setSelecting] = useState(false);
  const [selected, setSelected] = useState<string[]>([]);
  const [sheet, setSheet] = useState<"" | "add" | "delete">("");
  const [note, setNote] = useState("");

  const loadLists = useCallback(() => {
    api<Category[]>("/api/categories").then(setCategories).catch(() => setCategories([]));
    api<CardSet[]>("/api/sets").then(setSets).catch(() => setSets([]));
  }, []);
  useEffect(loadLists, [loadLists]);

  const loadCards = useCallback(() => {
    const params = new URLSearchParams();
    if (q.trim()) params.set("q", q.trim());
    if (categoryID) params.set("category_id", categoryID);
    if (setID) params.set("set_id", setID);
    if (sort !== "newest") params.set("sort", sort);
    const query = params.toString();
    return api<Flashcard[]>(`/api/flashcards${query ? `?${query}` : ""}`)
      .then(setCards)
      .catch((err: Error) => setError(err.message));
  }, [q, categoryID, setID, sort]);

  useEffect(() => {
    const timer = window.setTimeout(loadCards, 150);
    return () => window.clearTimeout(timer);
  }, [loadCards]);

  useEffect(() => {
    if (!note) return;
    const timer = window.setTimeout(() => setNote(""), 5000);
    return () => window.clearTimeout(timer);
  }, [note]);

  const filtering = q.trim() !== "" || categoryID !== "" || setID !== "";
  const shown = cards ?? [];
  const allSelected = shown.length > 0 && shown.every((card) => selected.includes(card.id));

  function toggle(id: string) {
    setSelected((current) => (current.includes(id) ? current.filter((x) => x !== id) : [...current, id]));
  }

  function stopSelecting() {
    setSelecting(false);
    setSelected([]);
  }

  return (
    <div className={selecting ? "pb-24" : ""}>
      <div className="mb-4 flex items-center justify-between gap-3">
        <h1 className="text-2xl font-extrabold lg:text-3xl">
          Cards {cards && !filtering && <span className="text-lg font-semibold text-muted">· {cards.length}</span>}
        </h1>
        <div className="flex gap-2">
          {shown.length > 0 && (
            <button type="button" className="btn-ghost h-11 px-4" aria-pressed={selecting} onClick={() => (selecting ? stopSelecting() : setSelecting(true))}>
              {selecting ? "Done" : "Select"}
            </button>
          )}
          {!selecting && (
            <Link to="/add" className="btn-primary h-11 lg:hidden">
              <Icon name="plus" /> Add
            </Link>
          )}
        </div>
      </div>
      <div className="relative">
        <Icon name="search" className="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-lg text-muted" />
        <input className="field pl-11" placeholder="Search hanzi, pinyin or English" aria-label="Search cards" value={q} onChange={(event) => setQ(event.target.value)} />
      </div>
      <div className="mt-3 grid grid-cols-2 gap-2 sm:flex">
        <label className="min-w-0 sm:w-56">
          <span className="sr-only">Set</span>
          <select className="field h-11 py-0 font-semibold" value={setID} onChange={(event) => setSetID(event.target.value)}>
            <option value="">All sets</option>
            {sets.map((set) => (
              <option key={set.id} value={set.id}>
                {set.title}
              </option>
            ))}
          </select>
        </label>
        <label className="min-w-0 sm:w-56">
          <span className="sr-only">Sort</span>
          <select className="field h-11 py-0 font-semibold" value={sort} onChange={(event) => setSort(event.target.value)}>
            {sorts.map((s) => (
              <option key={s.id} value={s.id}>
                {s.label}
              </option>
            ))}
          </select>
        </label>
      </div>
      {categories.length > 0 && (
        <div className="no-scrollbar -mx-5 mt-3 flex gap-2 overflow-x-auto px-5 pb-1 lg:mx-0 lg:flex-wrap lg:px-0" role="group" aria-label="Category">
          <button type="button" aria-pressed={categoryID === ""} className={`chip shrink-0 ${categoryID === "" ? "chip-on" : "bg-white"}`} onClick={() => setCategoryID("")}>
            All
          </button>
          {categories.map((category) => (
            <button
              key={category.id}
              type="button"
              aria-pressed={categoryID === category.id}
              className={`chip shrink-0 ${categoryID === category.id ? "chip-on" : "bg-white"}`}
              onClick={() => setCategoryID(categoryID === category.id ? "" : category.id)}
            >
              {category.name}
            </button>
          ))}
        </div>
      )}
      {selecting && shown.length > 0 && (
        <button type="button" className="link tap mt-2" onClick={() => setSelected(allSelected ? [] : shown.map((card) => card.id))}>
          {allSelected ? "Clear selection" : `Select all ${shown.length}`}
        </button>
      )}
      {note && (
        <p className="mt-3 flex items-center gap-2 rounded-2xl bg-ink px-4 py-3 text-sm text-white" role="status">
          <Icon name="check" className="shrink-0 text-gold" /> {note}
        </p>
      )}
      {error && <p className="mt-3 text-sm font-semibold text-crimson" role="alert">{error}</p>}
      {cards && cards.length === 0 && (
        <div className="surface mt-6 p-6 text-center">
          <p className="font-semibold">{filtering ? "No cards match." : "No cards yet."}</p>
          {!filtering && (
            <p className="mt-1 text-muted">
              <Link className="link" to="/add">Add a word</Link> or <Link className="link" to="/library">pick a set from the library</Link>.
            </p>
          )}
        </div>
      )}
      <ul className="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
        {shown.map((card) => {
          const on = selected.includes(card.id);
          const body = (
            <>
              <div className="relative">
                <CardImage card={card} className="aspect-[4/3] w-full" />
                {card.recent_score !== undefined && (
                  <span className={`absolute left-2 top-2 rounded-full px-2 py-0.5 text-xs font-bold ${scoreClass(card.recent_score)}`} title="Average over the last 30 days">
                    {Math.round(card.recent_score)}
                  </span>
                )}
                {selecting && (
                  <span className={`absolute right-2 top-2 grid h-8 w-8 place-items-center rounded-full border-2 ${on ? "border-crimson bg-crimson text-white" : "border-white bg-black/20"}`} aria-hidden>
                    {on && <Icon name="check" />}
                  </span>
                )}
              </div>
              <div className="px-3 pb-3 pt-2 text-left">
                <p className="truncate font-hanzi text-2xl font-bold">{card.hanzi}</p>
                <Pinyin tonal={card.tonal_pinyin} plain={card.pinyin} className="block truncate text-sm font-semibold" />
                <p className="truncate text-sm text-muted">{card.english}</p>
              </div>
            </>
          );
          return (
            <li key={card.id} className="min-w-0">
              {selecting ? (
                <button
                  type="button"
                  aria-pressed={on}
                  aria-label={`${card.hanzi} ${card.english}`}
                  onClick={() => toggle(card.id)}
                  className={`surface block w-full overflow-hidden transition ${on ? "ring-[3px] ring-crimson" : ""}`}
                >
                  {body}
                </button>
              ) : (
                <Link className="surface block overflow-hidden transition hover:shadow-lift" to={`/cards/${card.id}`}>
                  {body}
                </Link>
              )}
            </li>
          );
        })}
      </ul>

      {selecting && (
        <div className="fixed inset-x-0 bottom-[88px] z-30 flex justify-center px-4 lg:bottom-6 lg:pl-64" role="region" aria-label="Selected cards">
          <div className="flex w-full max-w-xl items-center gap-2 rounded-2xl bg-ink p-2 pl-4 text-white shadow-float">
            <span className="min-w-0 flex-1 font-semibold" aria-live="polite">
              {selected.length} selected
            </span>
            <button type="button" className="min-h-11 rounded-xl bg-white/10 px-3 font-semibold disabled:opacity-50" disabled={!selected.length} onClick={() => setSheet("add")}>
              Add to set
            </button>
            <button type="button" className="min-h-11 rounded-xl bg-crimson px-3 font-semibold disabled:opacity-50" disabled={!selected.length} onClick={() => setSheet("delete")}>
              Delete
            </button>
          </div>
        </div>
      )}

      <AddToSetSheet
        open={sheet === "add"}
        onClose={() => setSheet("")}
        sets={sets}
        ids={selected}
        onDone={(message) => {
          setSheet("");
          setNote(message);
          stopSelecting();
          loadLists();
        }}
      />
      <DeleteSheet
        open={sheet === "delete"}
        onClose={() => setSheet("")}
        ids={selected}
        onDone={(message) => {
          setSheet("");
          setNote(message);
          stopSelecting();
          void loadCards();
        }}
      />
    </div>
  );
}

function AddToSetSheet({ open, onClose, sets, ids, onDone }: { open: boolean; onClose: () => void; sets: CardSet[]; ids: string[]; onDone: (message: string) => void }) {
  const [title, setTitle] = useState("");
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (open) {
      setTitle("");
      setError("");
      setBusy("");
    }
  }, [open]);

  async function addTo(set: CardSet) {
    setBusy(set.id);
    setError("");
    try {
      const res = await api<{ added: number }>(`/api/sets/${set.id}/cards/add`, { method: "POST", body: JSON.stringify({ flashcard_ids: ids }) });
      const skipped = ids.length - res.added;
      onDone(`Added ${res.added} ${res.added === 1 ? "card" : "cards"} to ${set.title}${skipped ? ` (${skipped} already there)` : ""}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not add the cards");
      setBusy("");
    }
  }

  async function create(event: FormEvent) {
    event.preventDefault();
    setBusy("new");
    try {
      const set = await api<CardSet>("/api/sets", { method: "POST", body: JSON.stringify({ title: title.trim(), description: "" }) });
      await addTo(set);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not create the set");
      setBusy("");
    }
  }

  return (
    <Sheet open={open} onClose={onClose} title={`Add ${ids.length} ${ids.length === 1 ? "card" : "cards"} to a set`}>
      <ul className="grid max-h-[45vh] gap-2 overflow-y-auto">
        {sets.map((set) => (
          <li key={set.id}>
            <button type="button" disabled={!!busy} onClick={() => addTo(set)} className="flex min-h-14 w-full items-center gap-3 rounded-2xl bg-white px-4 text-left shadow-card disabled:opacity-60">
              <span className="min-w-0 flex-1">
                <b className="block truncate">{set.title}</b>
                <span className="text-sm text-muted">{set.card_count} cards</span>
              </span>
              {busy === set.id ? <span className="text-sm text-muted">Adding…</span> : <Icon name="plus" className="text-crimson" />}
            </button>
          </li>
        ))}
      </ul>
      <form onSubmit={create} className="mt-4 flex gap-2">
        <input className="field min-w-0 flex-1" placeholder="New set name" aria-label="New set name" value={title} onChange={(event) => setTitle(event.target.value)} />
        <button type="submit" className="btn-primary min-h-11 shrink-0" disabled={!title.trim() || !!busy}>
          Create
        </button>
      </form>
      {error && <p className="mt-3 text-sm font-semibold text-crimson" role="alert">{error}</p>}
    </Sheet>
  );
}

function DeleteSheet({ open, onClose, ids, onDone }: { open: boolean; onClose: () => void; ids: string[]; onDone: (message: string) => void }) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const n = ids.length;

  async function remove() {
    setBusy(true);
    setError("");
    try {
      const res = await api<{ deleted: number }>("/api/flashcards/bulk-delete", { method: "POST", body: JSON.stringify({ ids }) });
      onDone(`Deleted ${res.deleted} ${res.deleted === 1 ? "card" : "cards"}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not delete the cards");
    } finally {
      setBusy(false);
    }
  }

  return (
    <Sheet open={open} onClose={onClose} title={`Delete ${n} ${n === 1 ? "card" : "cards"}?`}>
      <p>They'll be removed from every set, along with their practice history. This can't be undone.</p>
      {error && <p className="mt-3 text-sm font-semibold text-crimson" role="alert">{error}</p>}
      <div className="mt-5 grid gap-3">
        <button type="button" className="min-h-14 rounded-2xl bg-crimson text-lg font-bold text-white disabled:opacity-60" disabled={busy} onClick={remove}>
          {busy ? "Deleting…" : `Delete ${n} ${n === 1 ? "card" : "cards"}`}
        </button>
        <button type="button" className="btn-ghost min-h-12" onClick={onClose}>
          Keep them
        </button>
      </div>
    </Sheet>
  );
}
