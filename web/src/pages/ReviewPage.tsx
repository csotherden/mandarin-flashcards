import { useCallback, useEffect, useRef, useState, type PointerEvent } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { api } from "../api";
import { Attribution, CardImage } from "../components/CardImage";
import { Icon } from "../components/Icon";
import { Pinyin } from "../components/Pinyin";
import { speakChinese, unlockSpeech } from "../speech";
import type { Flashcard, SetDetail } from "../types";

const frontKey = "mandarin.reviewFront";
type Front = "chinese" | "english";


/**
 * Review: flip through cards at your own pace, no grading, no XP, nothing saved. For learning new
 * words before (or instead of) graded practice. Leaving early leaves nothing behind to resume.
 */
export function ReviewPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const [cards, setCards] = useState<Flashcard[] | null>(null);
  const [title, setTitle] = useState("");
  const [index, setIndex] = useState(0);
  const [flipped, setFlipped] = useState(false);
  const [done, setDone] = useState(false);
  const [front, setFront] = useState<Front>(() => (localStorage.getItem(frontKey) === "english" ? "english" : "chinese"));
  const [error, setError] = useState("");
  const exitTo = params.get("from") || "/";
  const setIds = (params.get("sets") ?? "").split(",").filter(Boolean);

  useEffect(() => {
    const count = Number(params.get("count") ?? 0);
    const load = async () => {
      let list: Flashcard[] = [];
      if (setIds.length) {
        const details = await Promise.all(setIds.map((id) => api<SetDetail>(`/api/sets/${id}`)));
        const seen = new Set<string>();
        for (const d of details) for (const c of d.cards) if (!seen.has(c.id)) (seen.add(c.id), list.push(c));
        setTitle(details.map((d) => d.title).join(" + "));
      } else {
        list = await api<Flashcard[]>("/api/flashcards");
        setTitle("All cards");
      }
      setCards(count > 0 ? list.slice(0, count) : list);
    };
    load().catch((err: Error) => setError(err.message));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params]);

  const total = cards?.length ?? 0;
  const go = useCallback(
    (delta: number) => {
      if (!total) return;
      const next = index + delta;
      if (next >= total) {
        setDone(true);
        return;
      }
      if (next < 0) return;
      setIndex(next);
      setFlipped(false);
    },
    [index, total],
  );

  // Keyboard: Space/Enter flip (when not on a button), arrows move.
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (done || !cards) return;
      const onControl = (event.target as HTMLElement)?.closest("button, a, input, select");
      if (event.key === "ArrowRight") go(1);
      else if (event.key === "ArrowLeft") go(-1);
      else if ((event.key === " " || event.key === "Enter") && !onControl) {
        event.preventDefault();
        setFlipped((value) => !value);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [go, done, cards]);

  function chooseFront(next: Front) {
    setFront(next);
    setFlipped(false);
    localStorage.setItem(frontKey, next);
  }

  if (error) return <p className="p-6 font-semibold text-crimson">{error}</p>;
  if (!cards) return <p className="p-6 text-muted">Loading…</p>;
  if (total === 0) {
    return (
      <main className="grid min-h-dvh place-items-center p-6 text-center">
        <div>
          <p className="font-semibold">There are no cards to review here yet.</p>
          <Link to={exitTo} className="btn-primary mt-4 h-12">Back</Link>
        </div>
      </main>
    );
  }

  if (done) {
    return (
      <main className="mx-auto flex min-h-dvh max-w-xl flex-col justify-center px-5 py-6">
        <section className="rise relative overflow-hidden rounded-[28px] bg-hero p-6 text-center text-white">
          <Icon name="star" className="absolute left-5 top-5 text-3xl text-gold" />
          <p className="text-sm font-semibold uppercase tracking-[0.08em] text-white/85">Review done</p>
          <h1 className="mt-2 text-3xl font-extrabold">{total} {total === 1 ? "card" : "cards"} reviewed</h1>
          <p className="mt-1 text-white/90">{title}</p>
          <p className="mt-3 text-white/90">When they feel familiar, try saying them in a graded drill.</p>
        </section>
        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          <button
            type="button"
            className="btn-primary h-14 text-lg"
            onClick={() => {
              setDone(false);
              setIndex(0);
              setFlipped(false);
            }}
          >
            Review again
          </button>
          <Link to={exitTo} className="btn-ghost h-14">Done</Link>
        </div>
      </main>
    );
  }

  const card = cards[index];
  const last = index + 1 >= total;

  return (
    <div className="flex min-h-dvh flex-col bg-canvas">
      <header className="mx-auto flex w-full max-w-xl items-center gap-4 px-5 pt-5">
        <h1 className="sr-only">Review: {title}</h1>
        <button type="button" aria-label="End review" onClick={() => navigate(exitTo)} className="grid h-11 w-11 shrink-0 place-items-center rounded-xl bg-white text-lg shadow-card">
          <Icon name="close" />
        </button>
        <div className="h-2.5 flex-1 overflow-hidden rounded-full bg-line" role="progressbar" aria-label="Review progress" aria-valuemin={1} aria-valuemax={total} aria-valuenow={index + 1}>
          <div className="h-full rounded-full bg-ink transition-all duration-300" style={{ width: `${((index + 1) / total) * 100}%` }} />
        </div>
        <span className="shrink-0 text-sm font-bold text-muted">
          {index + 1} / {total}
        </span>
        <span className="shrink-0 text-sm font-bold uppercase tracking-[0.06em] text-muted">Review</span>
      </header>

      <main className="mx-auto flex w-full max-w-xl flex-1 flex-col px-5 pb-6 pt-4">
        <div className="mb-3 flex items-center justify-between gap-3">
          <p className="min-w-0 truncate text-sm font-semibold text-muted">{title}</p>
          <div className="flex shrink-0 rounded-full bg-white p-1 shadow-card" role="group" aria-label="Card front">
            {(["chinese", "english"] as Front[]).map((side) => (
              <button
                key={side}
                type="button"
                aria-pressed={front === side}
                onClick={() => chooseFront(side)}
                className={`min-h-9 rounded-full px-3 text-xs font-bold ${front === side ? "bg-ink text-white" : "text-muted"}`}
              >
                {side === "chinese" ? "中 first" : "EN first"}
              </button>
            ))}
          </div>
        </div>

        <ReviewCard key={card.id} card={card} front={front} flipped={flipped} onFlip={() => setFlipped((value) => !value)} onSwipe={go} />

        <div className="mt-4 grid grid-cols-2 gap-3">
          <button type="button" className="btn-soft h-12 bg-white shadow-card" onClick={() => speakChinese(card.hanzi, 1)}>
            <Icon name="volume" /> Listen
          </button>
          <button type="button" className="btn-soft h-12 bg-white shadow-card" onClick={() => speakChinese(card.hanzi, 0.7)}>
            <Icon name="volume" /> Slow
          </button>
        </div>

        <div className="mt-auto grid grid-cols-[auto_1fr_auto] gap-3 pt-6">
          <button type="button" className="btn-ghost h-14 w-14 px-0" aria-label="Previous card" disabled={index === 0} onClick={() => go(-1)}>
            <Icon name="left" className="text-xl" />
          </button>
          <button type="button" className="btn-ghost h-14" aria-pressed={flipped} onClick={() => setFlipped((value) => !value)}>
            {flipped ? "Show front" : "Flip"}
          </button>
          <button type="button" className="btn-primary h-14 px-5 text-lg" onClick={() => go(1)}>
            {last ? "Finish" : <><span className="sr-only">Next card</span><Icon name="right" className="text-xl" /></>}
          </button>
        </div>
        <p className="mt-3 text-center text-sm text-muted">Tap the card to flip · swipe or use the arrows to move</p>
      </main>
    </div>
  );
}

/** One card, front and back, flipping in place. Tapping flips; a horizontal swipe moves cards. */
function ReviewCard({ card, front, flipped, onFlip, onSwipe }: { card: Flashcard; front: Front; flipped: boolean; onFlip: () => void; onSwipe: (delta: number) => void }) {
  const start = useRef<{ x: number; y: number } | null>(null);
  const swiped = useRef(false);
  const hasPhoto = !!card.image_url || card.has_local_image;

  const chinese = (
    <div className="flex h-full flex-col">
      {hasPhoto && <CardImage card={card} className="aspect-[16/10] w-full" />}
      <div className="grid flex-1 place-items-center px-6 py-6 text-center">
        <p className="break-words font-hanzi text-[clamp(3.5rem,16vw,6.5rem)] font-bold leading-none">{card.hanzi}</p>
      </div>
    </div>
  );
  const english = (
    <div className="flex h-full flex-col">
      {hasPhoto && <CardImage card={card} className="aspect-[16/10] w-full" />}
      <div className="grid flex-1 place-items-center px-6 py-6 text-center">
        <p className="text-3xl font-extrabold leading-tight">{card.english || "—"}</p>
      </div>
    </div>
  );
  const answer = (
    <div className="flex h-full flex-col justify-center gap-1 p-6 text-white">
      <p className="break-words font-hanzi text-5xl font-bold leading-tight">{card.hanzi}</p>
      <Pinyin tonal={card.tonal_pinyin} plain={card.pinyin} spaced colored={false} className="text-2xl font-semibold text-white/95" />
      <p className="text-xl text-white/90">{card.english}</p>
      {card.example_sentence_hanzi && (
        <div className="mt-4 rounded-2xl bg-black/15 p-3">
          <p className="font-hanzi text-lg">{card.example_sentence_hanzi}</p>
          <p className="text-sm text-white/85">{card.example_sentence_pinyin}</p>
          <button
            type="button"
            tabIndex={flipped ? 0 : -1}
            className="mt-2 inline-flex min-h-11 items-center gap-1.5 rounded-xl bg-white/15 px-3 text-sm font-semibold"
            onClick={(event) => {
              event.stopPropagation();
              speakChinese(card.example_sentence_hanzi, 0.85);
            }}
          >
            <Icon name="volume" /> Example
          </button>
        </div>
      )}
    </div>
  );

  return (
    <div className="flip">
      {/* Tapping the card flips it; keyboard and screen-reader users have the Flip button and Space. */}
      <div
        onClick={() => {
          // The click that ends a swipe shouldn't also flip the card.
          if (swiped.current) swiped.current = false;
          else onFlip();
        }}
        // Photos would otherwise start a native image drag, which cancels the swipe.
        onDragStart={(event) => event.preventDefault()}
        onPointerDown={(event: PointerEvent) => {
          unlockSpeech();
          swiped.current = false;
          start.current = { x: event.clientX, y: event.clientY };
        }}
        onPointerCancel={() => (start.current = null)}
        onPointerUp={(event: PointerEvent) => {
          const s = start.current;
          start.current = null;
          if (!s) return;
          const dx = event.clientX - s.x;
          if (Math.abs(dx) > 60 && Math.abs(dx) > Math.abs(event.clientY - s.y) * 1.5) {
            swiped.current = true;
            onSwipe(dx < 0 ? 1 : -1);
          }
        }}
        aria-live="polite"
        className="block w-full cursor-pointer touch-pan-y select-none text-left"
      >
        <span className={`flip-inner relative grid ${flipped ? "flipped" : ""}`}>
          <span className="flip-face col-start-1 row-start-1 block min-h-[300px] overflow-hidden rounded-[32px] bg-white shadow-float" aria-hidden={flipped}>
            {front === "chinese" ? chinese : english}
          </span>
          <span className="flip-face flip-back col-start-1 row-start-1 block min-h-[300px] overflow-hidden rounded-[32px] bg-maroon shadow-float" aria-hidden={!flipped}>
            {answer}
          </span>
        </span>
      </div>
      {!flipped && hasPhoto && (
        <div className="mt-2 px-2">
          <Attribution card={card} />
        </div>
      )}
    </div>
  );
}
