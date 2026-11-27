import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { api, ApiError } from "../api";
import { Attribution, CardImage } from "../components/CardImage";
import type { Autofill, CardSet, Category, Flashcard, ImageCandidate } from "../types";

const empty = {
  hanzi: "",
  pinyin: "",
  tonal_pinyin: "",
  english: "",
  example_sentence_hanzi: "",
  example_sentence_pinyin: "",
  image_url: "",
  category_ids: [] as string[],
  set_ids: [] as string[],
};

export function CardPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [form, setForm] = useState(empty);
  const [attribution, setAttribution] = useState<Flashcard["unsplash_attribution"]>();
  const [categories, setCategories] = useState<Category[]>([]);
  const [sets, setSets] = useState<CardSet[]>([]);
  const [photos, setPhotos] = useState<ImageCandidate[]>([]);
  const [search, setSearch] = useState("");
  const [preview, setPreview] = useState<Flashcard | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [newCategory, setNewCategory] = useState("");

  useEffect(() => {
    api<Category[]>("/api/categories").then(setCategories).catch(() => undefined);
    api<CardSet[]>("/api/sets").then(setSets).catch(() => undefined);
    if (!id) return;
    api<Flashcard>(`/api/flashcards/${id}`)
      .then((card) => {
        setForm({
          hanzi: card.hanzi,
          pinyin: card.pinyin,
          tonal_pinyin: card.tonal_pinyin,
          english: card.english,
          example_sentence_hanzi: card.example_sentence_hanzi,
          example_sentence_pinyin: card.example_sentence_pinyin,
          image_url: card.image_url,
          category_ids: card.category_ids,
          set_ids: card.set_ids,
        });
        setAttribution(card.unsplash_attribution);
        setPreview(card);
      })
      .catch((err: Error) => setError(err.message));
  }, [id]);

  function toggle(list: string[], value: string) {
    return list.includes(value) ? list.filter((item) => item !== value) : [...list, value];
  }

  // Auto-fill starts from whichever the user typed: hanzi, then English, then pinyin.
  const autofillInput = [form.hanzi, form.english, form.pinyin].map((value) => value.trim()).find(Boolean) ?? "";

  async function autofill() {
    setBusy(true);
    setError("");
    try {
      const filled = await api<Autofill>("/api/flashcards/autofill", {
        method: "POST",
        body: JSON.stringify({ input: autofillInput }),
      });
      setForm((current) => ({
        ...current,
        hanzi: filled.hanzi || current.hanzi,
        pinyin: filled.pinyin,
        tonal_pinyin: filled.tonal_pinyin,
        english: filled.english,
        example_sentence_hanzi: filled.example_sentence_hanzi,
        example_sentence_pinyin: filled.example_sentence_pinyin,
      }));
      if (filled.image_search) {
        setSearch(filled.image_search);
        const images = await api<ImageCandidate[]>(`/api/images/search?q=${encodeURIComponent(filled.image_search)}`);
        setPhotos(images);
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not fill the card");
    } finally {
      setBusy(false);
    }
  }

  async function searchImages() {
    setError("");
    try {
      setPhotos(await api<ImageCandidate[]>(`/api/images/search?q=${encodeURIComponent(search)}`));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not search images");
    }
  }

  function choosePhoto(photo: ImageCandidate) {
    setForm((current) => ({ ...current, image_url: photo.full_url }));
    setAttribution({
      photographer_name: photo.photographer_name,
      photographer_url: photo.photographer_url,
      photo_url: photo.photo_url,
      download_location: photo.download_location,
    });
    setPreview(null);
  }

  async function save() {
    setBusy(true);
    setError("");
    const body = { ...form, unsplash_attribution: attribution ?? null };
    try {
      const saved = await api<Flashcard>(id ? `/api/flashcards/${id}` : "/api/flashcards", {
        method: id ? "PUT" : "POST",
        body: JSON.stringify(body),
      });
      navigate(`/cards/${saved.id}`);
      setPreview(saved);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not save");
    } finally {
      setBusy(false);
    }
  }

  async function upload(file: File) {
    if (!id) {
      setError("Save the card before uploading a photo");
      return;
    }
    const data = new FormData();
    data.set("file", file);
    try {
      const saved = await api<Flashcard>(`/api/flashcards/${id}/image`, { method: "POST", body: data });
      setPreview(saved);
      setForm((current) => ({ ...current, image_url: "" }));
      setAttribution(undefined);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not upload");
    }
  }

  async function remove() {
    if (!id || !confirm("Delete this card?")) return;
    await api(`/api/flashcards/${id}`, { method: "DELETE" });
    navigate("/");
  }

  async function addCategory() {
    const name = newCategory.trim();
    if (!name) return;
    const created = await api<Category>("/api/categories", {
      method: "POST",
      body: JSON.stringify({ name, color: "#9f2d20" }),
    });
    setCategories((current) => [...current, created].sort((a, b) => a.name.localeCompare(b.name)));
    setForm((current) => ({ ...current, category_ids: [...current.category_ids, created.id] }));
    setNewCategory("");
  }

  const shown = preview && !form.image_url ? preview : ({ ...preview, image_url: form.image_url, has_local_image: preview?.has_local_image && !form.image_url, unsplash_attribution: attribution, id: id ?? "draft" } as Flashcard);

  return (
    <div className="grid gap-4">
      <Link className="text-sm text-stone-500" to="/">← Library</Link>
      <h2 className="text-xl font-semibold">{id ? "Edit card" : "New card"}</h2>
      <label className="grid gap-1 text-sm">
        Hanzi
        <input className="rounded-2xl border border-stone-300 bg-white px-3 py-3 font-hanzi text-2xl" value={form.hanzi} onChange={(event) => setForm({ ...form, hanzi: event.target.value })} />
      </label>
      <label className="grid gap-1 text-sm">
        English
        <input className="rounded-2xl border border-stone-300 bg-white px-3 py-3 text-base" value={form.english} onChange={(event) => setForm({ ...form, english: event.target.value })} />
      </label>
      <label className="grid gap-1 text-sm">
        Pinyin
        <input className="rounded-2xl border border-stone-300 bg-white px-3 py-3 text-base" value={form.pinyin} onChange={(event) => setForm({ ...form, pinyin: event.target.value })} />
      </label>
      <button type="button" className="rounded-2xl border border-ink py-3 font-medium" disabled={busy || !autofillInput} onClick={autofill}>
        {busy ? "Filling…" : "Auto-fill & fetch images"}
      </button>
      <p className="-mt-2 text-xs text-stone-500">Type hanzi, English or pinyin, then auto-fill the rest.</p>
      <label className="grid gap-1 text-sm">
        Tonal pinyin
        <input className="rounded-2xl border border-stone-300 bg-white px-3 py-3 text-base" placeholder="ma1 ma5" value={form.tonal_pinyin} onChange={(event) => setForm({ ...form, tonal_pinyin: event.target.value })} />
        <span className="text-xs text-stone-500">One numbered syllable per character, 5 for the neutral tone. Used to grade pronunciation.</span>
      </label>
      <label className="grid gap-1 text-sm">
        Example
        <input className="rounded-2xl border border-stone-300 bg-white px-3 py-3 text-base" value={form.example_sentence_hanzi} onChange={(event) => setForm({ ...form, example_sentence_hanzi: event.target.value })} />
      </label>
      <label className="grid gap-1 text-sm">
        Example pinyin
        <input className="rounded-2xl border border-stone-300 bg-white px-3 py-3 text-base" value={form.example_sentence_pinyin} onChange={(event) => setForm({ ...form, example_sentence_pinyin: event.target.value })} />
      </label>
      {(form.image_url || preview?.has_local_image) && shown.id && (
        <div>
          <CardImage card={shown} className="h-48 w-full rounded-3xl" />
          <Attribution card={{ ...shown, unsplash_attribution: attribution }} />
        </div>
      )}
      <div className="flex gap-2">
        <input className="min-w-0 flex-1 rounded-2xl border border-stone-300 bg-white px-3 py-3 text-base" placeholder="Search photos" value={search} onChange={(event) => setSearch(event.target.value)} />
        <button type="button" className="rounded-2xl bg-stone-200 px-4" onClick={searchImages}>Search</button>
      </div>
      {photos.length > 0 && (
        <div className="grid grid-cols-3 gap-2">
          {photos.map((photo) => (
            <button key={photo.id} type="button" onClick={() => choosePhoto(photo)} className="overflow-hidden rounded-2xl">
              <img src={photo.thumb_url} alt="" className="h-24 w-full object-cover" />
            </button>
          ))}
        </div>
      )}
      <label className="text-sm">
        Or upload a photo
        <input className="mt-1 block w-full text-base" type="file" accept="image/*" onChange={(event) => event.target.files?.[0] && upload(event.target.files[0])} />
      </label>
      <div>
        <p className="text-sm font-medium">Categories</p>
        <div className="mt-2 flex flex-wrap gap-2">
          {categories.map((category) => (
            <button key={category.id} type="button" onClick={() => setForm({ ...form, category_ids: toggle(form.category_ids, category.id) })} className={`rounded-full px-3 py-1 text-sm ${form.category_ids.includes(category.id) ? "bg-ink text-paper" : "bg-white"}`}>
              {category.name}
            </button>
          ))}
        </div>
        <div className="mt-2 flex gap-2">
          <input className="min-w-0 flex-1 rounded-2xl border border-stone-300 bg-white px-3 py-2 text-base" placeholder="New category" value={newCategory} onChange={(event) => setNewCategory(event.target.value)} />
          <button type="button" className="rounded-2xl bg-stone-200 px-3" onClick={addCategory}>Add</button>
        </div>
      </div>
      <div>
        <p className="text-sm font-medium">Sets</p>
        <div className="mt-2 flex flex-wrap gap-2">
          {sets.map((set) => (
            <button key={set.id} type="button" onClick={() => setForm({ ...form, set_ids: toggle(form.set_ids, set.id) })} className={`rounded-full px-3 py-1 text-sm ${form.set_ids.includes(set.id) ? "bg-ink text-paper" : "bg-white"}`}>
              {set.title}
            </button>
          ))}
        </div>
      </div>
      {error && <p className="text-sm text-cinnabar">{error}</p>}
      <button type="button" className="rounded-2xl bg-cinnabar py-3 font-semibold text-white" disabled={busy} onClick={save}>Save card</button>
      {id && <button type="button" className="text-sm text-stone-500" onClick={remove}>Delete card</button>}
    </div>
  );
}
