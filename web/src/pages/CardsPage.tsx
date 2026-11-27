import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api";
import { CardImage } from "../components/CardImage";
import { Icon } from "../components/Icon";
import { Pinyin } from "../components/Pinyin";
import type { Category, Flashcard } from "../types";

/** The word library: search, filter by category, open a card to edit it. */
export function CardsPage() {
  const [cards, setCards] = useState<Flashcard[] | null>(null);
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
    const timer = window.setTimeout(() => {
      api<Flashcard[]>(`/api/flashcards${query ? `?${query}` : ""}`)
        .then(setCards)
        .catch((err: Error) => setError(err.message));
    }, 150);
    return () => window.clearTimeout(timer);
  }, [q, categoryID]);

  const filtering = q.trim() !== "" || categoryID !== "";

  return (
    <div>
      <div className="mb-4 flex items-center justify-between gap-3">
        <h1 className="text-2xl font-extrabold lg:text-3xl">
          Cards {cards && !filtering && <span className="text-lg font-semibold text-muted">· {cards.length}</span>}
        </h1>
        <Link to="/add" className="btn-primary h-11 lg:hidden">
          <Icon name="plus" /> Add
        </Link>
      </div>
      <div className="relative">
        <Icon name="search" className="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-lg text-muted" />
        <input className="field pl-11" placeholder="Search hanzi, pinyin or English" aria-label="Search cards" value={q} onChange={(event) => setQ(event.target.value)} />
      </div>
      {categories.length > 0 && (
        <div className="no-scrollbar -mx-5 mt-3 flex gap-2 overflow-x-auto px-5 pb-1 lg:mx-0 lg:flex-wrap lg:px-0">
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
      {error && <p className="mt-3 text-sm font-semibold text-crimson" role="alert">{error}</p>}
      {cards && cards.length === 0 && (
        <div className="surface mt-6 p-6 text-center">
          <p className="font-semibold">{filtering ? "No cards match." : "No cards yet."}</p>
          {!filtering && (
            <p className="mt-1 text-muted">
              <Link className="link" to="/add">Add a word</Link> or <Link className="link" to="/sets/generate">generate a set</Link>.
            </p>
          )}
        </div>
      )}
      <ul className="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
        {(cards ?? []).map((card) => (
          <li key={card.id} className="min-w-0">
            <Link className="surface block overflow-hidden transition hover:shadow-lift" to={`/cards/${card.id}`}>
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
    </div>
  );
}
