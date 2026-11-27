import { useEffect, useRef, useState, type FormEvent, type PointerEvent } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { api, ApiError } from "../api";
import { Icon } from "../components/Icon";
import { Pinyin } from "../components/Pinyin";
import type { Autofill, CardSet, Flashcard, ImageCandidate } from "../types";

const lastSetKey = "mandarin.lastSet";

/** What the full editor gets when the user wants every field ("Edit all details"). */
export type CardDraft = Autofill & { photo?: ImageCandidate; category_names: string[]; set_ids: string[] };

/**
 * Quick add: one word in (hanzi, English or pinyin), a card preview out with a photo carousel and
 * suggested categories, save, and straight back to an empty input for the next word.
 */
export function AddPage() {
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const inputRef = useRef<HTMLInputElement>(null);
  const [word, setWord] = useState("");
  const [draft, setDraft] = useState<Autofill | null>(null);
  const [draftFor, setDraftFor] = useState("");
  const [photos, setPhotos] = useState<ImageCandidate[]>([]);
  const [photo, setPhoto] = useState(0);
  const [categories, setCategories] = useState<{ name: string; on: boolean }[]>([]);
  const [newCategory, setNewCategory] = useState<string | null>(null);
  const [sets, setSets] = useState<CardSet[]>([]);
  const [setId, setSetId] = useState(params.get("set") ?? localStorage.getItem(lastSetKey) ?? "");
  const [status, setStatus] = useState<"idle" | "filling" | "saving">("idle");
  const [error, setError] = useState("");
  const [saved, setSaved] = useState<{ hanzi: string; set?: string; count: number } | null>(null);

  useEffect(() => {
    api<CardSet[]>("/api/sets")
      .then((list) => {
        setSets(list);
        if (setId && !list.some((set) => set.id === setId)) setSetId("");
      })
      .catch(() => undefined);
    inputRef.current?.focus();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const ready = draft !== null && draftFor === word.trim();

  async function generate() {
    const input = word.trim();
    if (!input) return;
    setStatus("filling");
    setError("");
    setDraft(null);
    setPhotos([]);
    try {
      const filled = await api<Autofill>("/api/flashcards/autofill", { method: "POST", body: JSON.stringify({ input }) });
      setDraft(filled);
      setDraftFor(input);
      setCategories(filled.categories.map((name) => ({ name, on: true })));
      setPhoto(0);
      if (filled.image_search) {
        api<ImageCandidate[]>(`/api/images/search?q=${encodeURIComponent(filled.image_search)}`)
          .then(setPhotos)
          .catch(() => setPhotos([]));
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not fill that word");
    } finally {
      setStatus("idle");
    }
  }

  function chosenPhoto() {
    return photos[photo];
  }

  async function save() {
    if (!draft) return;
    setStatus("saving");
    setError("");
    const picked = chosenPhoto();
    try {
      await api<Flashcard>("/api/flashcards", {
        method: "POST",
        body: JSON.stringify({
          hanzi: draft.hanzi,
          pinyin: draft.pinyin,
          tonal_pinyin: draft.tonal_pinyin,
          english: draft.english,
          example_sentence_hanzi: draft.example_sentence_hanzi,
          example_sentence_pinyin: draft.example_sentence_pinyin,
          image_url: picked?.full_url ?? "",
          unsplash_attribution: picked
            ? {
                photographer_name: picked.photographer_name,
                photographer_url: picked.photographer_url,
                photo_url: picked.photo_url,
                download_location: picked.download_location,
              }
            : null,
          category_names: categories.filter((item) => item.on).map((item) => item.name),
          category_ids: [],
          set_ids: setId ? [setId] : [],
        }),
      });
      if (setId) localStorage.setItem(lastSetKey, setId);
      setSaved((current) => ({ hanzi: draft.hanzi, set: sets.find((set) => set.id === setId)?.title, count: (current?.count ?? 0) + 1 }));
      // Keep anything typed while the save was in flight.
      setWord((current) => (current.trim() === draftFor ? "" : current));
      setDraft(null);
      setDraftFor("");
      setPhotos([]);
      setCategories([]);
      inputRef.current?.focus();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not save the card");
    } finally {
      setStatus("idle");
    }
  }

  function submit(event: FormEvent) {
    event.preventDefault();
    if (status !== "idle") return;
    if (ready) void save();
    else void generate();
  }

  function editAll() {
    if (!draft) return;
    const full: CardDraft = {
      ...draft,
      photo: chosenPhoto(),
      category_names: categories.filter((item) => item.on).map((item) => item.name),
      set_ids: setId ? [setId] : [],
    };
    navigate("/cards/new", { state: { draft: full } });
  }

  function addCategory() {
    const name = (newCategory ?? "").trim();
    if (name && !categories.some((item) => item.name.toLowerCase() === name.toLowerCase())) {
      setCategories((current) => [...current, { name, on: true }]);
    }
    setNewCategory(null);
  }

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-2xl font-extrabold lg:text-3xl">Add cards</h1>
        <button type="button" className="link" onClick={() => (window.history.length > 1 ? navigate(-1) : navigate("/"))}>
          Done
        </button>
      </div>

      {saved && (
        <div className="rise mb-4 flex items-center gap-2 rounded-2xl bg-ink px-4 py-3 text-sm text-white" role="status">
          <Icon name="check" className="text-lg text-gold" />
          <span className="min-w-0 truncate">
            Saved <b className="font-hanzi text-gold">{saved.hanzi}</b>
            {saved.set ? ` to ${saved.set}` : ""} · {saved.count} added
          </span>
        </div>
      )}

      <div className="grid grid-cols-[minmax(0,1fr)] gap-x-7 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.05fr)] lg:[grid-template-areas:'input_preview'_'opts_preview']">
        <form onSubmit={submit} className="lg:[grid-area:input]">
          <div className="flex gap-2 rounded-[20px] bg-white p-1.5 shadow-card">
            <input
              ref={inputRef}
              className="min-w-0 flex-1 bg-transparent px-3 text-lg outline-none"
              placeholder="Type a word"
              aria-label="Hanzi, English or pinyin"
              value={word}
              onChange={(event) => setWord(event.target.value)}
              autoComplete="off"
              autoCapitalize="off"
              enterKeyHint={ready ? "done" : "go"}
            />
            <button type="submit" className="btn-primary h-12 shrink-0 px-4" disabled={status !== "idle" || !word.trim()}>
              {ready ? (
                <>
                  <Icon name="check" /> {status === "saving" ? "Saving…" : "Save"}
                </>
              ) : (
                <>
                  <Icon name="spark" /> {status === "filling" ? "Working…" : "Generate"}
                </>
              )}
            </button>
          </div>
          <p className="mx-1 mt-2 text-sm text-muted">Hanzi, English or pinyin. Enter to generate, Enter again to save.</p>
          {error && <p className="mt-2 text-sm font-semibold text-crimson" role="alert">{error}</p>}
        </form>

        <div className="mt-4 lg:mt-0 lg:[grid-area:preview]">
          {status === "filling" && <PreviewSkeleton />}
          {draft && (
            <div className="rise overflow-hidden rounded-[28px] bg-white shadow-lift">
              <Carousel photos={photos} index={photo} onIndex={setPhoto} />
              <div className="px-5 pb-5 pt-4">
                <div className="font-hanzi text-5xl font-bold leading-tight">{draft.hanzi}</div>
                <Pinyin tonal={draft.tonal_pinyin} plain={draft.pinyin} className="mt-1 block text-[22px] font-semibold" />
                <div className="text-[17px] text-muted">{draft.english}</div>
                {draft.example_sentence_hanzi && (
                  <p className="mt-3 border-t border-line pt-3 text-sm text-muted">
                    <span className="font-hanzi text-base text-ink">{draft.example_sentence_hanzi}</span>
                    <br />
                    {draft.example_sentence_pinyin}
                  </p>
                )}
              </div>
            </div>
          )}
        </div>

        {draft && (
          <div className="mt-4 lg:mt-5 lg:[grid-area:opts]">
            {draft.existing_card_id && (
              <p className="mb-4 rounded-2xl bg-[#FFF4D6] px-4 py-3 text-sm text-ink">
                You already have <b className="font-hanzi">{draft.hanzi}</b>.{" "}
                <Link className="link" to={`/cards/${draft.existing_card_id}`}>Open it</Link> or save another copy.
              </p>
            )}
            <p className="label">Categories</p>
            <div className="mt-2 flex flex-wrap gap-2">
              {categories.map((item, index) => (
                <button
                  key={item.name}
                  type="button"
                  aria-pressed={item.on}
                  className={`chip ${item.on ? "chip-on" : ""}`}
                  onClick={() => setCategories((current) => current.map((c, i) => (i === index ? { ...c, on: !c.on } : c)))}
                >
                  {item.on && <Icon name="check" />} {item.name}
                </button>
              ))}
              {newCategory === null ? (
                <button type="button" className="chip border-[1.5px] border-dashed border-[#BDB5AC] bg-transparent text-muted" onClick={() => setNewCategory("")}>
                  + New
                </button>
              ) : (
                <input
                  autoFocus
                  className="field w-40 rounded-full py-2"
                  placeholder="Category"
                  value={newCategory}
                  onChange={(event) => setNewCategory(event.target.value)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter") {
                      event.preventDefault();
                      addCategory();
                    }
                    if (event.key === "Escape") setNewCategory(null);
                  }}
                  onBlur={addCategory}
                />
              )}
            </div>
            <label className="label mt-4 block" htmlFor="add-set">Add to set</label>
            <select id="add-set" className="field mt-2 font-semibold" value={setId} onChange={(event) => setSetId(event.target.value)}>
              <option value="">No set</option>
              {sets.map((set) => (
                <option key={set.id} value={set.id}>{set.title}</option>
              ))}
            </select>
            <button type="button" className="btn-primary mt-5 h-14 w-full text-[17px]" disabled={status !== "idle"} onClick={save}>
              {status === "saving" ? "Saving…" : "Save & add next"}
            </button>
            <button type="button" className="mt-3 block w-full text-center text-sm font-semibold text-muted underline underline-offset-4" onClick={editAll}>
              Edit all details
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

function PreviewSkeleton() {
  return (
    <div className="overflow-hidden rounded-[28px] bg-white shadow-lift" aria-label="Generating">
      <div className="aspect-[4/3] animate-pulse bg-[#ECE7E1]" />
      <div className="grid gap-2 p-5">
        <div className="h-12 w-32 animate-pulse rounded-xl bg-[#ECE7E1]" />
        <div className="h-5 w-24 animate-pulse rounded bg-[#ECE7E1]" />
      </div>
    </div>
  );
}

/** Swipe, arrows or arrow keys flip through the photos; the one showing is the one saved. */
function Carousel({ photos, index, onIndex }: { photos: ImageCandidate[]; index: number; onIndex: (i: number) => void }) {
  const start = useRef<number | null>(null);
  const n = photos.length;
  if (n === 0) {
    return <div className="grid aspect-[4/3] place-items-center bg-[#F1EEEA] px-6 text-center text-sm text-muted">Finding photos… You can also add one in Edit all details.</div>;
  }
  const go = (delta: number) => onIndex((index + delta + n) % n);
  const current = photos[Math.min(index, n - 1)];
  return (
    <div
      className="relative aspect-[4/3] touch-pan-y select-none bg-[#ECE7E1]"
      onPointerDown={(event: PointerEvent) => (start.current = event.clientX)}
      onPointerUp={(event: PointerEvent) => {
        if (start.current !== null && Math.abs(event.clientX - start.current) > 40) go(event.clientX < start.current ? 1 : -1);
        start.current = null;
      }}
      onKeyDown={(event) => {
        if (event.key === "ArrowRight") go(1);
        if (event.key === "ArrowLeft") go(-1);
      }}
      tabIndex={0}
      role="group"
      aria-label={`Photo ${index + 1} of ${n}`}
    >
      <img src={current.full_url} alt="" draggable={false} className="h-full w-full object-cover" />
      <span className="absolute right-3 top-3 rounded-full bg-ink/70 px-2.5 py-1 text-xs font-semibold text-white">{index + 1} / {n}</span>
      <button type="button" aria-label="Previous photo" onClick={() => go(-1)} className="absolute left-3 top-1/2 grid h-10 w-10 -translate-y-1/2 place-items-center rounded-full bg-white/90 text-ink shadow-card">
        <Icon name="left" />
      </button>
      <button type="button" aria-label="Next photo" onClick={() => go(1)} className="absolute right-3 top-1/2 grid h-10 w-10 -translate-y-1/2 place-items-center rounded-full bg-white/90 text-ink shadow-card">
        <Icon name="right" />
      </button>
      <div className="absolute inset-x-0 bottom-3 flex justify-center gap-1.5">
        {photos.map((item, i) => (
          <span key={item.id} className={`h-[7px] rounded-full transition-all ${i === index ? "w-[18px] bg-white" : "w-[7px] bg-white/60"}`} />
        ))}
      </div>
      <p className="absolute bottom-2 left-3 hidden text-[10px] text-white/80 sm:block">
        Photo: {current.photographer_name} / Unsplash
      </p>
    </div>
  );
}
