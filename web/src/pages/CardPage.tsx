import { useEffect, useState } from "react";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import { api, ApiError } from "../api";
import { Attribution, CardImage } from "../components/CardImage";
import { Icon } from "../components/Icon";
import { Pinyin } from "../components/Pinyin";
import { speakChinese } from "../speech";
import type { Attribution as Credit, Autofill, CardSet, Category, Flashcard, ImageCandidate } from "../types";
import type { CardDraft } from "./AddPage";

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

type Form = typeof empty;

/** The full editor: every field, photo search or upload, categories and sets. Also where cards are edited. */
export function CardPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const [form, setForm] = useState<Form>(empty);
  const [card, setCard] = useState<Flashcard | null>(null);
  const [attribution, setAttribution] = useState<Credit | undefined>();
  const [categories, setCategories] = useState<Category[]>([]);
  const [newNames, setNewNames] = useState<string[]>([]);
  const [newCategory, setNewCategory] = useState("");
  const [sets, setSets] = useState<CardSet[]>([]);
  const [photos, setPhotos] = useState<ImageCandidate[]>([]);
  const [search, setSearch] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState("");
  const [savedNote, setSavedNote] = useState(false);

  useEffect(() => {
    const draft = (location.state as { draft?: CardDraft } | null)?.draft;
    Promise.all([api<Category[]>("/api/categories").catch(() => [] as Category[]), api<CardSet[]>("/api/sets").catch(() => [] as CardSet[])]).then(
      ([cats, list]) => {
        setCategories(cats);
        setSets(list);
        if (!id && draft) {
          // Coming from quick add: map suggested names onto existing categories, keep the rest as new.
          const byName = new Map(cats.map((c) => [c.name.toLowerCase(), c.id]));
          setForm((current) => ({
            ...current,
            category_ids: draft.category_names.map((name) => byName.get(name.toLowerCase())).filter((v): v is string => !!v),
          }));
          setNewNames(draft.category_names.filter((name) => !byName.has(name.toLowerCase())));
        }
      },
    );
    if (!id) {
      setCard(null);
      setAttribution(undefined);
      if (draft) {
        setForm({ ...fromAutofill(empty, draft), image_url: draft.photo?.full_url ?? "", set_ids: draft.set_ids, category_ids: [] });
        if (draft.photo) setAttribution(credit(draft.photo));
        setSearch(draft.image_search);
      } else setForm(empty);
      return;
    }
    api<Flashcard>(`/api/flashcards/${id}`)
      .then((row) => {
        setCard(row);
        setForm({
          hanzi: row.hanzi,
          pinyin: row.pinyin,
          tonal_pinyin: row.tonal_pinyin,
          english: row.english,
          example_sentence_hanzi: row.example_sentence_hanzi,
          example_sentence_pinyin: row.example_sentence_pinyin,
          image_url: row.image_url,
          category_ids: row.category_ids,
          set_ids: row.set_ids,
        });
        setAttribution(row.unsplash_attribution);
        setSearch(row.english);
      })
      .catch((err: Error) => setError(err.message));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  const set = (patch: Partial<Form>) => setForm((current) => ({ ...current, ...patch }));
  const toggle = (list: string[], value: string) => (list.includes(value) ? list.filter((item) => item !== value) : [...list, value]);

  // Auto-fill starts from whichever the user typed: hanzi, then English, then pinyin.
  const autofillInput = [form.hanzi, form.english, form.pinyin].map((value) => value.trim()).find(Boolean) ?? "";

  async function autofill() {
    setBusy("fill");
    setError("");
    try {
      const filled = await api<Autofill>("/api/flashcards/autofill", { method: "POST", body: JSON.stringify({ input: autofillInput }) });
      setForm((current) => fromAutofill(current, filled));
      const known = new Set(categories.map((c) => c.name.toLowerCase()));
      setNewNames((current) => [...new Set([...current, ...filled.categories.filter((name) => !known.has(name.toLowerCase()))])]);
      const ids = categories.filter((c) => filled.categories.some((name) => name.toLowerCase() === c.name.toLowerCase())).map((c) => c.id);
      setForm((current) => ({ ...current, category_ids: [...new Set([...current.category_ids, ...ids])] }));
      if (filled.image_search) {
        setSearch(filled.image_search);
        setPhotos(await api<ImageCandidate[]>(`/api/images/search?q=${encodeURIComponent(filled.image_search)}`));
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not fill the card");
    } finally {
      setBusy("");
    }
  }

  async function searchImages() {
    if (!search.trim()) return;
    setBusy("search");
    setError("");
    try {
      setPhotos(await api<ImageCandidate[]>(`/api/images/search?q=${encodeURIComponent(search.trim())}`));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not search photos");
    } finally {
      setBusy("");
    }
  }

  function choosePhoto(photo: ImageCandidate) {
    set({ image_url: photo.full_url });
    setAttribution(credit(photo));
  }

  async function save() {
    setBusy("save");
    setError("");
    try {
      const saved = await api<Flashcard>(id ? `/api/flashcards/${id}` : "/api/flashcards", {
        method: id ? "PUT" : "POST",
        body: JSON.stringify({ ...form, category_names: newNames, unsplash_attribution: attribution ?? null }),
      });
      if (newNames.length) {
        setCategories(await api<Category[]>("/api/categories").catch(() => categories));
        setNewNames([]);
      }
      setForm((current) => ({ ...current, category_ids: saved.category_ids }));
      setCard(saved);
      setSavedNote(true);
      window.setTimeout(() => setSavedNote(false), 2200);
      if (!id) navigate(`/cards/${saved.id}`, { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not save");
    } finally {
      setBusy("");
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
      setCard(saved);
      set({ image_url: "" });
      setAttribution(undefined);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not upload");
    }
  }

  async function remove() {
    if (!id || !confirm("Delete this card?")) return;
    await api(`/api/flashcards/${id}`, { method: "DELETE" });
    navigate("/cards");
  }

  function addCategory() {
    const name = newCategory.trim();
    setNewCategory("");
    if (!name) return;
    const existing = categories.find((c) => c.name.toLowerCase() === name.toLowerCase());
    if (existing) set({ category_ids: [...new Set([...form.category_ids, existing.id])] });
    else if (!newNames.some((n) => n.toLowerCase() === name.toLowerCase())) setNewNames((current) => [...current, name]);
  }

  // What the preview shows: a chosen Unsplash photo wins over an uploaded one.
  const shown = {
    ...(card ?? ({} as Flashcard)),
    id: id ?? "draft",
    hanzi: form.hanzi,
    image_url: form.image_url,
    has_local_image: !form.image_url && !!card?.has_local_image,
    unsplash_attribution: attribution,
  } as Flashcard;

  return (
    <div className="min-w-0">
      <div className="mb-4 flex items-center justify-between gap-3">
        <button type="button" className="flex min-h-11 items-center gap-1 font-semibold text-muted" onClick={() => (window.history.length > 1 ? navigate(-1) : navigate("/cards"))}>
          <Icon name="back" /> Back
        </button>
        {savedNote && (
          <span className="rise flex items-center gap-1.5 rounded-full bg-ink px-3 py-1.5 text-sm font-semibold text-white" role="status">
            <Icon name="check" className="text-gold" /> Saved
          </span>
        )}
      </div>
      <h1 className="mb-5 text-2xl font-extrabold lg:text-3xl">{id ? "Edit card" : "New card"}</h1>

      <div className="grid min-w-0 gap-6 lg:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)] lg:items-start">
        {/* Preview and photo */}
        <div className="grid min-w-0 gap-3 lg:sticky lg:top-8">
          <div className="overflow-hidden rounded-[28px] bg-white shadow-lift">
            <CardImage card={shown} className="aspect-[4/3] w-full" />
            <div className="flex items-start justify-between gap-3 px-5 pb-5 pt-4">
              <div className="min-w-0">
                <div className="break-words font-hanzi text-5xl font-bold leading-tight">{form.hanzi || "—"}</div>
                <Pinyin tonal={form.tonal_pinyin} plain={form.pinyin} className="mt-1 block text-xl font-semibold" />
                <div className="text-muted">{form.english}</div>
              </div>
              {form.hanzi && (
                <button type="button" aria-label="Listen" className="grid h-11 w-11 shrink-0 place-items-center rounded-full bg-[#F1EEEA] text-xl" onClick={() => speakChinese(form.hanzi, 0.85)}>
                  <Icon name="volume" />
                </button>
              )}
            </div>
          </div>
          <Attribution card={shown} />

          <div className="surface grid min-w-0 gap-3 p-4">
            <p className="label">Photo</p>
            <div className="flex min-w-0 gap-2">
              <input
                className="field min-w-0 flex-1"
                placeholder="Search photos"
                aria-label="Search photos"
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                onKeyDown={(event) => event.key === "Enter" && (event.preventDefault(), void searchImages())}
              />
              <button type="button" className="btn-soft min-h-11 shrink-0" onClick={searchImages} disabled={busy === "search"} aria-label="Search photos">
                <Icon name="search" /> <span className="hidden sm:inline">Search</span>
              </button>
            </div>
            {photos.length > 0 && (
              <div className="grid grid-cols-3 gap-2">
                {photos.map((photo) => (
                  <button
                    key={photo.id}
                    type="button"
                    aria-label={`Use photo by ${photo.photographer_name}`}
                    aria-pressed={form.image_url === photo.full_url}
                    onClick={() => choosePhoto(photo)}
                    className={`overflow-hidden rounded-2xl ring-offset-2 ${form.image_url === photo.full_url ? "ring-[3px] ring-crimson" : ""}`}
                  >
                    <img src={photo.thumb_url} alt="" className="aspect-square w-full object-cover" />
                  </button>
                ))}
              </div>
            )}
            <label className="btn-soft h-11 cursor-pointer">
              <Icon name="plus" /> Upload your own
              <input className="sr-only" type="file" accept="image/*" onChange={(event) => event.target.files?.[0] && upload(event.target.files[0])} />
            </label>
            {!id && <p className="-mt-1 text-xs text-muted">Uploads work after the card is saved.</p>}
          </div>
        </div>

        {/* Fields */}
        <div className="grid min-w-0 gap-4">
          <div className="surface grid min-w-0 gap-4 p-4">
            <Field label="Hanzi">
              <input className="field font-hanzi text-2xl" value={form.hanzi} onChange={(event) => set({ hanzi: event.target.value })} />
            </Field>
            <Field label="English">
              <input className="field" value={form.english} onChange={(event) => set({ english: event.target.value })} />
            </Field>
            <Field label="Pinyin">
              <input className="field" value={form.pinyin} onChange={(event) => set({ pinyin: event.target.value })} />
            </Field>
            <button type="button" className="btn-ghost h-12" disabled={busy === "fill" || !autofillInput} onClick={autofill}>
              <Icon name="spark" /> {busy === "fill" ? "Filling…" : "Auto-fill the rest"}
            </button>
            <Field label="Tonal pinyin" hint="One numbered syllable per character, 5 for neutral. Used to grade pronunciation.">
              <input className="field font-mono text-[15px]" placeholder="ma1 ma5" value={form.tonal_pinyin} onChange={(event) => set({ tonal_pinyin: event.target.value })} />
            </Field>
            <Field label="Example">
              <input className="field font-hanzi" value={form.example_sentence_hanzi} onChange={(event) => set({ example_sentence_hanzi: event.target.value })} />
            </Field>
            <Field label="Example pinyin">
              <input className="field" value={form.example_sentence_pinyin} onChange={(event) => set({ example_sentence_pinyin: event.target.value })} />
            </Field>
          </div>

          <div className="surface grid min-w-0 gap-3 p-4">
            <p className="label">Categories</p>
            <div className="flex flex-wrap gap-2">
              {categories.map((category) => {
                const on = form.category_ids.includes(category.id);
                return (
                  <button key={category.id} type="button" aria-pressed={on} onClick={() => set({ category_ids: toggle(form.category_ids, category.id) })} className={`chip ${on ? "chip-on" : ""}`}>
                    {on && <Icon name="check" />} {category.name}
                  </button>
                );
              })}
              {newNames.map((name) => (
                <button key={`new-${name}`} type="button" className="chip chip-on" aria-pressed="true" title="New, created on save" onClick={() => setNewNames((current) => current.filter((n) => n !== name))}>
                  <Icon name="check" /> {name} <span className="text-white/80">new</span>
                </button>
              ))}
            </div>
            <div className="flex min-w-0 gap-2">
              <input
                className="field min-w-0 flex-1"
                placeholder="New category"
                aria-label="New category"
                value={newCategory}
                onChange={(event) => setNewCategory(event.target.value)}
                onKeyDown={(event) => event.key === "Enter" && (event.preventDefault(), addCategory())}
              />
              <button type="button" className="btn-soft shrink-0" onClick={addCategory}>Add</button>
            </div>
          </div>

          {sets.length > 0 && (
            <div className="surface grid min-w-0 gap-3 p-4">
              <p className="label">Sets</p>
              <div className="flex flex-wrap gap-2">
                {sets.map((item) => {
                  const on = form.set_ids.includes(item.id);
                  return (
                    <button key={item.id} type="button" aria-pressed={on} onClick={() => set({ set_ids: toggle(form.set_ids, item.id) })} className={`chip ${on ? "chip-on" : ""}`}>
                      {on && <Icon name="check" />} {item.title}
                    </button>
                  );
                })}
              </div>
            </div>
          )}

          {error && <p className="text-sm font-semibold text-crimson" role="alert">{error}</p>}
          <button type="button" className="btn-primary h-14 text-[17px]" disabled={busy === "save" || !form.hanzi.trim()} onClick={save}>
            {busy === "save" ? "Saving…" : id ? "Save changes" : "Save card"}
          </button>
          {id && (
            <button type="button" className="mx-auto flex min-h-11 items-center gap-1.5 px-3 text-sm font-semibold text-muted" onClick={remove}>
              <Icon name="trash" /> Delete card
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

function Field({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
  return (
    <label className="grid min-w-0 gap-1.5">
      <span className="label">{label}</span>
      {children}
      {hint && <span className="text-xs text-muted">{hint}</span>}
    </label>
  );
}

function fromAutofill(current: Form, filled: Autofill): Form {
  return {
    ...current,
    hanzi: filled.hanzi || current.hanzi,
    pinyin: filled.pinyin,
    tonal_pinyin: filled.tonal_pinyin,
    english: filled.english,
    example_sentence_hanzi: filled.example_sentence_hanzi,
    example_sentence_pinyin: filled.example_sentence_pinyin,
  };
}

function credit(photo: ImageCandidate): Credit {
  return {
    photographer_name: photo.photographer_name,
    photographer_url: photo.photographer_url,
    photo_url: photo.photo_url,
    download_location: photo.download_location,
  };
}
