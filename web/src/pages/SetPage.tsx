import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { api, ApiError } from "../api";
import { CardImage } from "../components/CardImage";
import { Icon } from "../components/Icon";
import { Pinyin } from "../components/Pinyin";
import { SessionSheet } from "../components/SessionSheet";
import { Sheet } from "../components/Sheet";
import type { CardSet, Flashcard, SetDetail } from "../types";

export function SetPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const [detail, setDetail] = useState<SetDetail | null>(null);
  const [error, setError] = useState("");
  const [practice, setPractice] = useState(false);
  const [picking, setPicking] = useState(false);
  const [editing, setEditing] = useState(false);

  function load() {
    api<SetDetail>(`/api/sets/${id}`).then(setDetail).catch((err: Error) => setError(err.message));
  }
  useEffect(load, [id]);

  async function save(ids: string[]) {
    try {
      setDetail(await api<SetDetail>(`/api/sets/${id}/cards`, { method: "PUT", body: JSON.stringify({ flashcard_ids: ids }) }));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not update the set");
    }
  }

  async function rename(title: string) {
    if (!detail) return;
    try {
      await api<CardSet>(`/api/sets/${id}`, { method: "PUT", body: JSON.stringify({ title, description: detail.description }) });
      setDetail({ ...detail, title });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not rename the set");
    }
  }

  async function remove() {
    if (!confirm(`Delete the set "${detail?.title}"? Its cards stay in your library.`)) return;
    await api(`/api/sets/${id}`, { method: "DELETE" });
    navigate("/sets");
  }

  if (!detail) return <p className="text-muted">{error || "Loading…"}</p>;
  const ids = detail.cards.map((card) => card.id);
  const removeCard = (card: Flashcard) => save(ids.filter((cardID) => cardID !== card.id));

  return (
    <div>
      <Link to="/sets" className="mb-2 flex min-h-11 w-fit items-center gap-1 font-semibold text-muted">
        <Icon name="back" /> Sets
      </Link>
      <div className="mb-5 flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="break-words text-2xl font-extrabold lg:text-3xl">{detail.title}</h1>
          <p className="text-muted">{detail.cards.length} {detail.cards.length === 1 ? "card" : "cards"}</p>
        </div>
        <button type="button" aria-label="Edit set" className="grid h-11 w-11 shrink-0 place-items-center rounded-xl bg-white shadow-card" onClick={() => setEditing(true)}>
          <Icon name="edit" />
        </button>
      </div>

      <div className="grid gap-3 sm:grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)_minmax(0,1fr)]">
        <button type="button" className="btn-primary h-14 text-lg" disabled={detail.cards.length === 0} onClick={() => setPractice(true)}>
          <Icon name="play" /> Practice
        </button>
        <Link to={`/add?set=${id}`} className="btn-ghost h-14">
          <Icon name="plus" /> New words
        </Link>
        <Link to={`/sets/generate?set=${id}`} className="btn-ghost h-14">
          <Icon name="spark" /> Generate more
        </Link>
      </div>
      <button type="button" className="link tap mt-1 text-sm" onClick={() => setPicking(true)}>
        Add cards you already have
      </button>

      {error && <p className="mt-3 text-sm font-semibold text-crimson" role="alert">{error}</p>}
      {detail.cards.length === 0 && (
        <div className="surface mt-6 p-6 text-center text-muted">This set is empty. Add new words, generate some, or pick from your library.</div>
      )}
      <ul className="mt-5 grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
        {detail.cards.map((card) => (
          <li key={card.id} className="surface relative min-w-0 overflow-hidden">
            <Link to={`/cards/${card.id}`} className="block">
              <CardImage card={card} className="aspect-[4/3] w-full" />
              <div className="px-3 pb-3 pt-2">
                <p className="truncate font-hanzi text-2xl font-bold">{card.hanzi}</p>
                <Pinyin tonal={card.tonal_pinyin} plain={card.pinyin} className="block truncate text-sm font-semibold" />
                <p className="truncate text-sm text-muted">{card.english}</p>
              </div>
            </Link>
            <button
              type="button"
              aria-label={`Remove ${card.hanzi} from this set`}
              className="absolute right-0.5 top-0.5 grid h-11 w-11 place-items-center"
              onClick={() => removeCard(card)}
            >
              <span className="grid h-8 w-8 place-items-center rounded-full bg-white/90 text-sm text-muted shadow-card">
                <Icon name="close" />
              </span>
            </button>
          </li>
        ))}
      </ul>

      <SessionSheet open={practice} onClose={() => setPractice(false)} sets={[detail]} initialSetIds={[detail.id]} />
      <PickSheet open={picking} onClose={() => setPicking(false)} have={ids} onAdd={(more) => save([...ids, ...more])} />
      <EditSheet open={editing} onClose={() => setEditing(false)} title={detail.title} onRename={rename} onDelete={remove} />
    </div>
  );
}

/** Multi-select from the library; one save adds every picked card. */
function PickSheet({ open, onClose, have, onAdd }: { open: boolean; onClose: () => void; have: string[]; onAdd: (ids: string[]) => Promise<void> }) {
  const [cards, setCards] = useState<Flashcard[]>([]);
  const [q, setQ] = useState("");
  const [picked, setPicked] = useState<string[]>([]);

  useEffect(() => {
    if (!open) return;
    setPicked([]);
    setQ("");
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const timer = window.setTimeout(() => {
      api<Flashcard[]>(`/api/flashcards${q.trim() ? `?q=${encodeURIComponent(q.trim())}` : ""}`)
        .then(setCards)
        .catch(() => setCards([]));
    }, 150);
    return () => window.clearTimeout(timer);
  }, [open, q]);

  const choices = cards.filter((card) => !have.includes(card.id));

  return (
    <Sheet open={open} onClose={onClose} title="Add cards">
      <input className="field" placeholder="Search your cards" aria-label="Search your cards" value={q} onChange={(event) => setQ(event.target.value)} />
      <ul className="mt-3 grid max-h-[50vh] gap-2 overflow-y-auto">
        {choices.length === 0 && <li className="py-6 text-center text-muted">{q ? "No matches." : "Every card is already in this set."}</li>}
        {choices.map((card) => {
          const on = picked.includes(card.id);
          return (
            <li key={card.id}>
              <button
                type="button"
                aria-pressed={on}
                onClick={() => setPicked((current) => (on ? current.filter((v) => v !== card.id) : [...current, card.id]))}
                className={`flex w-full items-center gap-3 rounded-2xl border-2 bg-white p-2 text-left ${on ? "border-crimson" : "border-transparent"}`}
              >
                <CardImage card={card} className="h-12 w-12 shrink-0 rounded-xl" />
                <span className="min-w-0 flex-1">
                  <b className="block truncate font-hanzi text-lg">{card.hanzi}</b>
                  <span className="block truncate text-sm text-muted">{card.english}</span>
                </span>
                <span className={`grid h-7 w-7 shrink-0 place-items-center rounded-full ${on ? "bg-crimson text-white" : "border-2 border-line"}`}>{on && <Icon name="check" />}</span>
              </button>
            </li>
          );
        })}
      </ul>
      <button
        type="button"
        className="btn-primary mt-4 h-14 w-full text-lg"
        disabled={picked.length === 0}
        onClick={async () => {
          await onAdd(picked);
          onClose();
        }}
      >
        {picked.length ? `Add ${picked.length} ${picked.length === 1 ? "card" : "cards"}` : "Pick cards"}
      </button>
    </Sheet>
  );
}

function EditSheet({ open, onClose, title, onRename, onDelete }: { open: boolean; onClose: () => void; title: string; onRename: (t: string) => Promise<void>; onDelete: () => void }) {
  const [value, setValue] = useState(title);
  useEffect(() => {
    if (open) setValue(title);
  }, [open, title]);
  return (
    <Sheet open={open} onClose={onClose} title="Edit set">
      <form
        className="grid gap-4"
        onSubmit={async (event) => {
          event.preventDefault();
          await onRename(value.trim());
          onClose();
        }}
      >
        <input className="field" aria-label="Set name" value={value} onChange={(event) => setValue(event.target.value)} />
        <button type="submit" className="btn-primary h-13 py-3.5" disabled={!value.trim()}>Save</button>
        <button type="button" className="mx-auto flex items-center gap-1.5 text-sm font-semibold text-muted" onClick={onDelete}>
          <Icon name="trash" /> Delete set
        </button>
      </form>
    </Sheet>
  );
}
