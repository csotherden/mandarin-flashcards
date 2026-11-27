import { useEffect, useMemo, useRef, useState, type PointerEvent } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api, ApiError } from "../api";
import { Attribution, CardImage } from "../components/CardImage";
import { Icon } from "../components/Icon";
import { markSyllable, Pinyin, toneLabel } from "../components/Pinyin";
import { speakChinese, unlockSpeech } from "../speech";
import type { AssessResponse, CardResult, Flashcard, Session } from "../types";

function mimeType() {
  const types = ["audio/webm;codecs=opus", "audio/webm", "audio/mp4"];
  return types.find((type) => typeof MediaRecorder !== "undefined" && MediaRecorder.isTypeSupported(type));
}

export function DrillPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const [session, setSession] = useState<Session | null>(null);
  const [index, setIndex] = useState(0);
  const [graded, setGraded] = useState(0);
  const [error, setError] = useState("");

  useEffect(() => {
    api<Session>(`/api/study/sessions/${id}`)
      .then((row) => {
        if (row.attempt.completed_at) {
          navigate(`/summary/${id}`, { replace: true });
          return;
        }
        setSession(row);
        setGraded(row.results.length);
        const done = new Set(row.results.map((result) => result.flashcard_id));
        const next = row.cards.findIndex((card) => !done.has(card.id));
        setIndex(next === -1 ? Math.max(row.cards.length - 1, 0) : next);
      })
      .catch((err: Error) => setError(err.message));
  }, [id, navigate]);

  async function finish() {
    try {
      await api(`/api/study/sessions/${id}/complete`, { method: "POST" });
      navigate(`/summary/${id}`, { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not finish the session");
    }
  }

  function exit() {
    if (graded > 0) {
      if (confirm("End the session now and see how you did?")) void finish();
      return;
    }
    navigate("/");
  }

  if (!session) return <p className="p-6 text-muted">{error || "Loading…"}</p>;
  if (session.cards.length === 0) {
    return (
      <div className="grid min-h-screen place-items-center p-6 text-center">
        <div>
          <p className="font-semibold">This session has no cards.</p>
          <button type="button" className="btn-primary mt-4 h-12" onClick={() => navigate("/")}>Back to Today</button>
        </div>
      </div>
    );
  }

  const match = session.attempt.mode === "picture_hanzi_match";
  const total = session.cards.length;
  const progress = match ? graded / total : index / total;

  return (
    <div className="flex min-h-screen flex-col bg-canvas">
      <header className="mx-auto flex w-full max-w-2xl items-center gap-4 px-5 pt-5">
        <button type="button" aria-label="End session" onClick={exit} className="grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-white text-lg shadow-card">
          <Icon name="close" />
        </button>
        <div className="h-2.5 flex-1 overflow-hidden rounded-full bg-line" role="progressbar" aria-valuemin={0} aria-valuemax={total} aria-valuenow={match ? graded : index}>
          <div className="h-full rounded-full bg-crimson transition-all duration-500" style={{ width: `${Math.max(4, progress * 100)}%` }} />
        </div>
        <span className="w-14 shrink-0 text-right text-sm font-bold text-muted">
          {match ? graded : index + 1} / {total}
        </span>
      </header>
      {error && <p className="mx-auto mt-3 w-full max-w-2xl px-5 text-sm font-semibold text-crimson" role="alert">{error}</p>}
      {match ? (
        <MatchRound session={session} onMatched={() => setGraded((n) => n + 1)} onDone={finish} onError={setError} />
      ) : (
        <SpeakRound
          session={session}
          index={index}
          onGraded={() => setGraded((n) => n + 1)}
          onNext={() => (index + 1 >= total ? finish() : setIndex(index + 1))}
          onError={setError}
        />
      )}
    </div>
  );
}

function SpeakRound({
  session,
  index,
  onGraded,
  onNext,
  onError,
}: {
  session: Session;
  index: number;
  onGraded: () => void;
  onNext: () => void;
  onError: (message: string) => void;
}) {
  const card = session.cards[index];
  const remaining = session.cards.length - index - 1;
  const [result, setResult] = useState<AssessResponse | null>(null);
  const [hint, setHint] = useState(false);
  const [recording, setRecording] = useState(false);
  const [busy, setBusy] = useState(false);
  const holding = useRef(false);
  const recorder = useRef<MediaRecorder | null>(null);
  const chunks = useRef<Blob[]>([]);
  const gradedCard = useRef<string | null>(null);

  useEffect(() => {
    setResult(null);
    setHint(false);
  }, [card.id]);

  async function grade(blob: Blob) {
    setBusy(true);
    const body = new FormData();
    body.set("audio", blob, "take");
    body.set("lesson_attempt_id", session.attempt.id);
    body.set("flashcard_id", card.id);
    body.set("mode", session.attempt.mode);
    try {
      const res = await api<AssessResponse>("/api/assess", { method: "POST", body });
      setResult(res);
      if (gradedCard.current !== card.id) {
        gradedCard.current = card.id;
        onGraded();
      }
    } catch (err) {
      onError(err instanceof ApiError ? err.message : "Could not grade that take");
    } finally {
      setBusy(false);
    }
  }

  async function pointerDown(event: PointerEvent<HTMLButtonElement>) {
    event.preventDefault();
    unlockSpeech();
    onError("");
    holding.current = true;
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      if (!holding.current) {
        stream.getTracks().forEach((track) => track.stop());
        return;
      }
      const mime = mimeType();
      const rec = new MediaRecorder(stream, mime ? { mimeType: mime } : undefined);
      chunks.current = [];
      rec.ondataavailable = (item) => {
        if (item.data.size) chunks.current.push(item.data);
      };
      rec.onstop = () => {
        stream.getTracks().forEach((track) => track.stop());
        void grade(new Blob(chunks.current, { type: rec.mimeType || mime || "audio/webm" }));
      };
      recorder.current = rec;
      rec.start();
      setRecording(true);
    } catch {
      onError("Allow the microphone to grade pronunciation.");
      holding.current = false;
    }
  }

  function pointerUp() {
    holding.current = false;
    if (recorder.current && recorder.current.state === "recording") recorder.current.stop();
    setRecording(false);
  }

  const flipped = result !== null;

  return (
    <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col px-5 pb-8 pt-6">
      {/* A small stack behind the card shows how many are left. */}
      <div className="relative">
        {remaining > 1 && <div className="absolute inset-x-6 -bottom-3 h-full rounded-[32px] bg-white/60 shadow-card" aria-hidden />}
        {remaining > 0 && <div className="absolute inset-x-3 -bottom-1.5 h-full rounded-[32px] bg-white/80 shadow-card" aria-hidden />}
        <div className="flip relative">
          <div className={`flip-inner relative grid ${flipped ? "flipped" : ""}`}>
            {/* Front */}
            <section className="flip-face col-start-1 row-start-1 overflow-hidden rounded-[32px] bg-white shadow-float" aria-hidden={flipped}>
              {session.attempt.mode === "image_speak" ? (
                <>
                  <CardImage card={card} className="aspect-[4/3] w-full" />
                  <div className="flex min-h-[76px] items-center justify-center px-5 py-4 text-center">
                    {hint ? (
                      <span className="rise text-xl font-semibold">{card.english || "No English gloss"}</span>
                    ) : (
                      <span className="text-muted">What is this in Mandarin?</span>
                    )}
                  </div>
                </>
              ) : (
                <div className="grid aspect-[4/3] place-items-center px-6">
                  <div className="text-center">
                    <p className="break-words font-hanzi text-[clamp(4rem,18vw,8rem)] font-bold leading-none">{card.hanzi}</p>
                    {hint && <p className="rise mt-4 text-xl text-muted">{card.english}</p>}
                  </div>
                </div>
              )}
            </section>

            {/* Back */}
            <section className="flip-face flip-back col-start-1 row-start-1 overflow-y-auto rounded-[32px] bg-maroon p-6 text-white shadow-float" aria-hidden={!flipped}>
              {result && <Back card={card} result={result} />}
            </section>
          </div>
        </div>
      </div>
      {session.attempt.mode === "image_speak" && !flipped && (
        <div className="mt-4 px-2">
          <Attribution card={card} />
        </div>
      )}

      <div className="mt-auto pt-8">
        {!flipped ? (
          <div className="flex items-center justify-between gap-4">
            <button type="button" className="btn-soft h-12 w-24 bg-white shadow-card" onClick={() => setHint((value) => !value)} aria-pressed={hint}>
              Hint
            </button>
            <button
              type="button"
              aria-label={recording ? "Release to grade" : "Hold to speak"}
              className={`mic grid h-24 w-24 shrink-0 place-items-center rounded-full text-4xl text-white shadow-[0_12px_28px_rgba(200,16,46,.35)] transition ${recording ? "scale-110 bg-maroon ring-8 ring-crimson/25" : "bg-crimson"}`}
              onPointerDown={pointerDown}
              onPointerUp={pointerUp}
              onPointerCancel={pointerUp}
              onContextMenu={(event) => event.preventDefault()}
              disabled={busy}
            >
              <Icon name="mic" />
            </button>
            <button type="button" className="btn-soft h-12 w-24 bg-white shadow-card" onClick={onNext}>
              Skip
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-[auto_1fr] gap-3">
            <button type="button" className="btn-ghost h-14" onClick={() => setResult(null)}>Try again</button>
            <button type="button" className="btn-primary h-14 text-lg" onClick={onNext}>
              {remaining === 0 ? "Finish" : "Next card"} <Icon name="right" />
            </button>
          </div>
        )}
        <p className="mt-3 text-center text-sm font-semibold text-muted">
          {flipped ? " " : recording ? "Listening… release when done" : busy ? "Grading…" : "Hold the mic and say the word"}
        </p>
      </div>
    </main>
  );
}

/** The red back of the card: the answer, the score, and each syllable's tone. */
function Back({ card, result }: { card: Flashcard; result: AssessResponse }) {
  const score = Math.round(result.result.overall_score);
  const syllables = result.result.syllables ?? [];
  return (
    <div>
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <p className="break-words font-hanzi text-5xl font-bold leading-tight">{card.hanzi}</p>
          <Pinyin tonal={card.tonal_pinyin} plain={card.pinyin} colored={false} className="mt-1 block text-xl font-semibold text-white/90" />
          <p className="text-white/80">{card.english}</p>
        </div>
        <div className="shrink-0 text-right">
          <p className="text-5xl font-extrabold leading-none text-gold">{result.assessable ? score : "—"}</p>
          <p className="mt-1 text-sm font-semibold text-white/80">{!result.assessable ? "not graded" : score >= 85 ? "Nice!" : score >= 60 ? "Close" : "Keep at it"}</p>
        </div>
      </div>
      {!result.assessable && <p className="mt-3 text-sm text-white/90">{result.reason || "This take could not be graded. Try again a little louder."}</p>}
      {syllables.length > 0 && (
        <ul className="mt-5 flex flex-wrap gap-2">
          {syllables.map((syllable, i) => {
            const ok = syllable.status === "ok";
            return (
              <li key={`${syllable.character}-${i}`} className={`rounded-2xl px-3 py-2 text-center ${ok ? "bg-white/15" : "bg-white text-ink"}`}>
                <span className="block font-hanzi text-2xl font-bold">{syllable.character}</span>
                <span className="block text-sm font-semibold">{markSyllable(syllable.pinyin).text}</span>
                <span className={`block text-xs font-bold ${ok ? "text-gold" : "text-crimson"}`}>
                  {ok ? "✓" : syllable.heard_tone && syllable.heard_tone !== syllable.expected_tone ? `heard ${toneLabel(syllable.heard_tone)}` : Math.round(syllable.score)}
                </span>
              </li>
            );
          })}
        </ul>
      )}
      {result.result.coaching_tip && <p className="mt-4 rounded-2xl bg-black/15 p-3 text-[15px] leading-6">{result.result.coaching_tip}</p>}
      <div className="mt-4 flex gap-2">
        <button type="button" className="flex h-11 flex-1 items-center justify-center gap-2 rounded-xl bg-white/15 font-semibold" onClick={() => speakChinese(card.hanzi, 1)}>
          <Icon name="volume" /> Listen
        </button>
        <button type="button" className="flex h-11 flex-1 items-center justify-center gap-2 rounded-xl bg-white/15 font-semibold" onClick={() => speakChinese(card.hanzi, 0.7)}>
          <Icon name="volume" /> Slow
        </button>
      </div>
    </div>
  );
}

function MatchRound({
  session,
  onMatched,
  onDone,
  onError,
}: {
  session: Session;
  onMatched: () => void;
  onDone: () => Promise<void>;
  onError: (message: string) => void;
}) {
  const rounds = useMemo(() => {
    const size = 6;
    const groups: Flashcard[][] = [];
    for (let i = 0; i < session.cards.length; i += size) groups.push(session.cards.slice(i, i + size));
    return groups;
  }, [session.cards]);
  const [round, setRound] = useState(0);
  const cards = rounds[round] ?? [];
  // eslint-disable-next-line react-hooks/exhaustive-deps
  const hanzi = useMemo(() => shuffle(cards), [round]);
  const [imageID, setImageID] = useState<string | null>(null);
  const [hanziID, setHanziID] = useState<string | null>(null);
  const [locked, setLocked] = useState<Record<string, boolean>>({});
  const [misses, setMisses] = useState<Record<string, number>>({});
  const [wrong, setWrong] = useState(false);

  useEffect(() => {
    setImageID(null);
    setHanziID(null);
    setLocked({});
    setWrong(false);
  }, [round]);

  async function choose(kind: "image" | "hanzi", id: string) {
    if (locked[id] || wrong) return;
    unlockSpeech();
    const nextImage = kind === "image" ? id : imageID;
    const nextHanzi = kind === "hanzi" ? id : hanziID;
    if (kind === "image") setImageID(id);
    else setHanziID(id);
    if (!nextImage || !nextHanzi) return;
    if (nextImage === nextHanzi) {
      const mismatchCount = misses[nextImage] ?? 0;
      setLocked((current) => ({ ...current, [nextImage]: true }));
      setImageID(null);
      setHanziID(null);
      onMatched();
      speakChinese(cards.find((card) => card.id === nextImage)?.hanzi ?? "", 1);
      try {
        await api<CardResult>(`/api/study/sessions/${session.attempt.id}/matches`, {
          method: "POST",
          body: JSON.stringify({ flashcard_id: nextImage, correct: true, mismatch_count: mismatchCount }),
        });
      } catch (err) {
        onError(err instanceof ApiError ? err.message : "Could not save the match");
      }
      const lockedCount = Object.keys(locked).length + 1;
      if (lockedCount >= cards.length) {
        if (round + 1 >= rounds.length) await onDone();
        else window.setTimeout(() => setRound((value) => value + 1), 500);
      }
      return;
    }
    setWrong(true);
    setMisses((current) => ({ ...current, [nextImage]: (current[nextImage] ?? 0) + 1 }));
    try {
      await api(`/api/study/sessions/${session.attempt.id}/matches`, {
        method: "POST",
        body: JSON.stringify({ flashcard_id: nextImage, correct: false, mismatch_count: (misses[nextImage] ?? 0) + 1 }),
      });
    } catch (err) {
      onError(err instanceof ApiError ? err.message : "Could not save the match");
    }
    window.setTimeout(() => {
      setWrong(false);
      setImageID(null);
      setHanziID(null);
    }, 500);
  }

  const tile = (id: string, selected: boolean) =>
    locked[id]
      ? "border-tone3 opacity-50"
      : wrong && selected
        ? "border-crimson animate-[shake_.4s]"
        : selected
          ? "border-ink -translate-y-0.5 shadow-lift"
          : "border-transparent";

  return (
    <main className="mx-auto w-full max-w-2xl flex-1 px-5 pb-8 pt-6">
      <p className="text-center font-semibold text-muted">
        Match each photo to its word{rounds.length > 1 && ` · round ${round + 1} of ${rounds.length}`}
      </p>
      <div className="mt-4 grid grid-cols-3 gap-2.5">
        {cards.map((card) => (
          <button
            key={card.id}
            type="button"
            aria-label={`Photo ${card.english}`}
            aria-pressed={imageID === card.id}
            disabled={locked[card.id]}
            onClick={() => choose("image", card.id)}
            className={`overflow-hidden rounded-2xl border-[3px] bg-white shadow-card transition ${tile(card.id, imageID === card.id)}`}
          >
            <CardImage card={{ ...card, hanzi: "" }} className="aspect-square w-full" />
          </button>
        ))}
      </div>
      <div className="mt-5 grid grid-cols-3 gap-2.5">
        {hanzi.map((card) => (
          <button
            key={card.id}
            type="button"
            aria-pressed={hanziID === card.id}
            disabled={locked[card.id]}
            onClick={() => choose("hanzi", card.id)}
            className={`rounded-2xl border-[3px] bg-white py-5 font-hanzi text-3xl font-bold shadow-card transition ${tile(card.id, hanziID === card.id)} ${locked[card.id] ? "text-tone3" : ""}`}
          >
            {card.hanzi}
          </button>
        ))}
      </div>
    </main>
  );
}

function shuffle<T>(items: T[]) {
  const copy = [...items];
  for (let i = copy.length - 1; i > 0; i -= 1) {
    const j = Math.floor(Math.random() * (i + 1));
    [copy[i], copy[j]] = [copy[j], copy[i]];
  }
  return copy;
}
