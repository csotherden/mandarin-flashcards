import { useEffect, useMemo, useRef, useState, type PointerEvent } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api, ApiError } from "../api";
import { Attribution, CardImage } from "../components/CardImage";
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
  const [error, setError] = useState("");

  useEffect(() => {
    api<Session>(`/api/study/sessions/${id}`)
      .then((row) => {
        if (row.attempt.completed_at) {
          navigate(`/summary/${id}`, { replace: true });
          return;
        }
        setSession(row);
        const done = new Set(row.results.map((result) => result.flashcard_id));
        const next = row.cards.findIndex((card) => !done.has(card.id));
        setIndex(next === -1 ? Math.max(row.cards.length - 1, 0) : next);
      })
      .catch((err: Error) => setError(err.message));
  }, [id, navigate]);

  async function finish() {
    await api(`/api/study/sessions/${id}/complete`, { method: "POST" });
    navigate(`/summary/${id}`);
  }

  if (!session) return <p className="p-6 text-stone-500">{error || "Loading drill…"}</p>;
  if (session.cards.length === 0) return <p className="p-6">This drill has no cards.</p>;

  const round = session.attempt.mode === "picture_hanzi_match" ? (
    <MatchRound session={session} onDone={finish} onError={setError} />
  ) : (
    <SpeakRound session={session} index={index} onIndex={setIndex} onDone={finish} onError={setError} />
  );

  return (
    <>
      {error && <p className="px-4 pt-4 text-sm text-cinnabar">{error}</p>}
      {round}
    </>
  );
}

function SpeakRound({
  session,
  index,
  onIndex,
  onDone,
  onError,
}: {
  session: Session;
  index: number;
  onIndex: (index: number) => void;
  onDone: () => Promise<void>;
  onError: (message: string) => void;
}) {
  const card = session.cards[index];
  const [result, setResult] = useState<AssessResponse | null>(null);
  const [hint, setHint] = useState(false);
  const [recording, setRecording] = useState(false);
  const [busy, setBusy] = useState(false);
  const holding = useRef(false);
  const recorder = useRef<MediaRecorder | null>(null);
  const chunks = useRef<Blob[]>([]);

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
      setResult(await api<AssessResponse>("/api/assess", { method: "POST", body }));
    } catch (err) {
      onError(err instanceof ApiError ? err.message : "Could not grade that take");
    } finally {
      setBusy(false);
    }
  }

  async function pointerDown(event: PointerEvent<HTMLButtonElement>) {
    event.preventDefault();
    unlockSpeech();
    holding.current = true;
    setResult(null);
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

  async function next() {
    if (index + 1 >= session.cards.length) {
      await onDone();
      return;
    }
    onIndex(index + 1);
  }

  return (
    <div className="mx-auto flex min-h-screen max-w-3xl flex-col px-4 py-4">
      <p className="text-sm text-stone-500">{index + 1} / {session.cards.length}</p>
      {session.attempt.mode === "image_speak" ? (
        <div className="mt-3">
          <CardImage card={card} className="h-72 w-full rounded-3xl" />
          <Attribution card={card} />
          <button type="button" className="mt-2 text-sm text-stone-500" onClick={() => setHint((value) => !value)}>
            {hint ? card.english || "No English gloss" : "Tap for English hint"}
          </button>
        </div>
      ) : (
        <p className="mt-8 text-center font-hanzi text-7xl font-semibold">{card.hanzi}</p>
      )}
      {!result && (
        <button
          type="button"
          className={`mic mx-auto mt-8 grid h-28 w-28 place-items-center rounded-full text-sm font-semibold text-white ${recording ? "bg-cinnabar" : "bg-ink"}`}
          onPointerDown={pointerDown}
          onPointerUp={pointerUp}
          onPointerCancel={pointerUp}
          onContextMenu={(event) => event.preventDefault()}
        >
          {recording ? "Release" : busy ? "Grading" : "Hold to speak"}
        </button>
      )}
      {result && <ResultCard card={card} result={result} onAgain={() => setResult(null)} onNext={next} />}
    </div>
  );
}

function ResultCard({
  card,
  result,
  onAgain,
  onNext,
}: {
  card: Flashcard;
  result: AssessResponse;
  onAgain: () => void;
  onNext: () => void;
}) {
  const score = Math.round(result.result.overall_score);
  return (
    <section className="mt-6 rounded-3xl bg-white p-4">
      <p className="font-hanzi text-3xl font-semibold">{card.hanzi}</p>
      <p className="text-stone-500">{card.pinyin}</p>
      <p className={`mt-2 text-4xl font-semibold ${score >= 85 ? "text-moss" : "text-cinnabar"}`}>{result.assessable ? score : "—"}</p>
      {!result.assessable && <p className="text-sm text-stone-500">{result.reason || "This take could not be graded"}</p>}
      <div className="mt-3 flex flex-wrap gap-2">
        {(result.result.syllables ?? []).map((syllable) => (
          <span
            key={`${syllable.character}-${syllable.pinyin}`}
            className={`rounded-2xl px-3 py-2 font-hanzi text-2xl ${syllable.status === "ok" ? "bg-emerald-100 text-moss" : syllable.status === "mismatch" ? "bg-red-100 text-cinnabar" : "bg-amber-100 text-amberglow"}`}
            title={syllable.detail}
          >
            {syllable.character}
          </span>
        ))}
      </div>
      {(result.result.syllables ?? []).some((syllable) => syllable.detail) && (
        <ul className="mt-2 text-sm text-stone-600">
          {(result.result.syllables ?? []).filter((syllable) => syllable.detail).map((syllable) => (
            <li key={`${syllable.character}-detail`}>{syllable.character}: {syllable.detail}</li>
          ))}
        </ul>
      )}
      {result.result.coaching_tip && <p className="mt-3 text-sm leading-6">{result.result.coaching_tip}</p>}
      <div className="mt-4 flex gap-2">
        <button type="button" className="flex-1 rounded-2xl bg-stone-100 py-3" onClick={() => speakChinese(card.hanzi, 1)}>Listen</button>
        <button type="button" className="flex-1 rounded-2xl bg-stone-100 py-3" onClick={() => speakChinese(card.hanzi, 0.7)}>Listen slow</button>
      </div>
      <div className="mt-2 flex gap-2">
        <button type="button" className="flex-1 rounded-2xl border border-ink py-3" onClick={onAgain}>Try again</button>
        <button type="button" className="flex-1 rounded-2xl bg-ink py-3 text-paper" onClick={onNext}>Next card</button>
      </div>
    </section>
  );
}

function MatchRound({ session, onDone, onError }: { session: Session; onDone: () => Promise<void>; onError: (message: string) => void }) {
  const rounds = useMemo(() => {
    const size = 6;
    const groups: Flashcard[][] = [];
    for (let i = 0; i < session.cards.length; i += size) groups.push(session.cards.slice(i, i + size));
    return groups;
  }, [session.cards]);
  const [round, setRound] = useState(0);
  const cards = rounds[round] ?? [];
  const hanzi = useMemo(() => shuffle(cards), [round, cards]);
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
    if (locked[id]) return;
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
        else setRound((value) => value + 1);
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
    }, 450);
  }

  return (
    <div className="mx-auto min-h-screen max-w-3xl px-4 py-4">
      <p className="text-sm text-stone-500">Round {round + 1} / {rounds.length}</p>
      <div className={`mt-3 grid grid-cols-2 gap-2 ${wrong ? "animate-pulse" : ""}`}>
        {cards.map((card) => (
          <button key={card.id} type="button" disabled={locked[card.id]} onClick={() => choose("image", card.id)} className={`overflow-hidden rounded-2xl border-2 ${locked[card.id] ? "border-moss opacity-60" : imageID === card.id ? "border-ink" : "border-transparent"} ${wrong && imageID === card.id ? "border-cinnabar" : ""}`}>
            <CardImage card={card} className="h-28 w-full" />
          </button>
        ))}
      </div>
      <div className="mt-4 grid grid-cols-2 gap-2">
        {hanzi.map((card) => (
          <button key={card.id} type="button" disabled={locked[card.id]} onClick={() => choose("hanzi", card.id)} className={`rounded-2xl border-2 bg-white py-4 font-hanzi text-3xl ${locked[card.id] ? "border-moss text-moss" : hanziID === card.id ? "border-ink" : "border-transparent"} ${wrong && hanziID === card.id ? "border-cinnabar text-cinnabar" : ""}`}>
            {card.hanzi}
          </button>
        ))}
      </div>
    </div>
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
