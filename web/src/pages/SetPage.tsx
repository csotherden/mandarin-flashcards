import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api, ApiError } from "../api";
import type { Flashcard, SetDetail } from "../types";

export function SetPage() {
  const { id = "" } = useParams();
  const [detail, setDetail] = useState<SetDetail | null>(null);
  const [query, setQuery] = useState("");
  const [matches, setMatches] = useState<Flashcard[]>([]);
  const [error, setError] = useState("");

  function load() {
    api<SetDetail>(`/api/sets/${id}`).then(setDetail).catch((err: Error) => setError(err.message));
  }
  useEffect(load, [id]);

  useEffect(() => {
    if (!query.trim()) {
      setMatches([]);
      return;
    }
    api<Flashcard[]>(`/api/flashcards?q=${encodeURIComponent(query.trim())}`).then(setMatches).catch(() => setMatches([]));
  }, [query]);

  async function save(ids: string[]) {
    try {
      setDetail(await api<SetDetail>(`/api/sets/${id}/cards`, { method: "PUT", body: JSON.stringify({ flashcard_ids: ids }) }));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not update the set");
    }
  }

  if (!detail) return <p className="text-stone-500">{error || "Loading…"}</p>;
  const ids = detail.cards.map((card) => card.id);

  function move(index: number, delta: number) {
    const next = [...ids];
    const target = index + delta;
    if (target < 0 || target >= next.length) return;
    const [item] = next.splice(index, 1);
    next.splice(target, 0, item);
    void save(next);
  }

  return (
    <div>
      <Link className="text-sm text-stone-500" to="/sets">← Sets</Link>
      <h2 className="mt-2 text-xl font-semibold">{detail.title}</h2>
      <input className="mt-3 w-full rounded-2xl border border-stone-300 bg-white px-3 py-3 text-base" placeholder="Add a card from your library" value={query} onChange={(event) => setQuery(event.target.value)} />
      {matches.length > 0 && (
        <ul className="mt-2 rounded-2xl bg-white">
          {matches.filter((card) => !ids.includes(card.id)).slice(0, 6).map((card) => (
            <li key={card.id}>
              <button type="button" className="block w-full px-3 py-2 text-left" onClick={() => { setQuery(""); void save([...ids, card.id]); }}>
                <span className="font-hanzi">{card.hanzi}</span> <span className="text-stone-500">{card.english}</span>
              </button>
            </li>
          ))}
        </ul>
      )}
      {error && <p className="mt-3 text-sm text-cinnabar">{error}</p>}
      <ol className="mt-4 grid gap-2">
        {detail.cards.map((card, index) => (
          <li key={card.id} className="flex items-center gap-2 rounded-2xl bg-white px-3 py-2">
            <span className="w-6 text-sm text-stone-400">{index + 1}</span>
            <Link className="min-w-0 flex-1" to={`/cards/${card.id}`}>
              <span className="font-hanzi text-lg">{card.hanzi}</span>
              <span className="ml-2 text-sm text-stone-500">{card.english}</span>
            </Link>
            <button type="button" className="px-2" onClick={() => move(index, -1)} aria-label="Move up">↑</button>
            <button type="button" className="px-2" onClick={() => move(index, 1)} aria-label="Move down">↓</button>
            <button type="button" className="px-2 text-stone-400" onClick={() => save(ids.filter((cardID) => cardID !== card.id))} aria-label="Remove">✕</button>
          </li>
        ))}
      </ol>
    </div>
  );
}
