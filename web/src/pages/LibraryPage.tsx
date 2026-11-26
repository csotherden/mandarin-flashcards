import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api";
import { CardImage } from "../components/CardImage";
import type { Category, Flashcard } from "../types";

export function LibraryPage() {
  const [cards, setCards] = useState<Flashcard[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [q, setQ] = useState("");
  const [categoryID, setCategoryID] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    api<Category[]>("/api/categories").then(setCategories).catch(() => setCategories([]));
  }, []);

  useEffect(() => {
    const params = new URLSearchParams();
    if (q.trim()) params.set("q", q.trim());
    if (categoryID) params.set("category_id", categoryID);
    const query = params.toString();
    api<Flashcard[]>(`/api/flashcards${query ? `?${query}` : ""}`)
      .then(setCards)
      .catch((err: Error) => setError(err.message));
  }, [q, categoryID]);

  return (
    <div>
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-xl font-semibold">Library</h2>
        <Link className="rounded-full bg-cinnabar px-4 py-2 text-sm font-semibold text-white" to="/cards/new">
          Add card
        </Link>
      </div>
      <input
        className="mt-3 w-full rounded-2xl border border-stone-300 bg-white px-3 py-3 text-base"
        placeholder="Search hanzi, pinyin, or English"
        value={q}
        onChange={(event) => setQ(event.target.value)}
      />
      <div className="mt-3 flex gap-2 overflow-x-auto pb-1">
        <button type="button" className={`shrink-0 rounded-full px-3 py-1 text-sm ${categoryID === "" ? "bg-ink text-paper" : "bg-white"}`} onClick={() => setCategoryID("")}>
          All
        </button>
        {categories.map((category) => (
          <button
            key={category.id}
            type="button"
            className={`shrink-0 rounded-full px-3 py-1 text-sm ${categoryID === category.id ? "bg-ink text-paper" : "bg-white"}`}
            onClick={() => setCategoryID(category.id)}
          >
            <span className="mr-1 inline-block h-2 w-2 rounded-full" style={{ background: category.color }} />
            {category.name}
          </button>
        ))}
      </div>
      {error && <p className="mt-3 text-sm text-cinnabar">{error}</p>}
      {cards.length === 0 && <p className="mt-8 text-stone-500">No cards yet. Add the words that keep slipping.</p>}
      <ul className="mt-4 grid gap-3">
        {cards.map((card) => (
          <li key={card.id}>
            <Link className="flex gap-3 rounded-3xl bg-white p-3 shadow-sm" to={`/cards/${card.id}`}>
              <CardImage card={card} className="h-20 w-20 shrink-0 rounded-2xl" />
              <div className="min-w-0">
                <p className="font-hanzi text-2xl font-semibold">{card.hanzi}</p>
                <p className="text-sm text-stone-500">{card.pinyin}</p>
                <p className="truncate text-sm">{card.english}</p>
              </div>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
