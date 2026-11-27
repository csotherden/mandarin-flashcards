import { useEffect, useRef, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { api, ApiError } from "../api";
import { Icon } from "../components/Icon";
import { markSyllable, Pinyin, toneLabel } from "../components/Pinyin";
import { ToneContour } from "../components/ToneContour";
import { useRecorder } from "../recorder";
import { speakChinese } from "../speech";
import type { ToneAssess, ToneComplete, ToneLesson, TonesOverview } from "../types";
import { tierLabel, tierPoints } from "../xp";

const contours: Record<number, number[] | undefined> = { 1: [5, 5], 2: [3, 5], 3: [2, 1, 4], 4: [5, 1] };

/** The tone number of each syllable in numbered pinyin ("ma1 ma5" -> [1, 5]). */
function tonesOf(tonal: string) {
  return tonal.trim().split(/\s+/).map((s) => markSyllable(s).tone);
}

/** One lesson: learn (explanation, contour, examples to hear), then say each word, then a wrap-up. */
export function ToneLessonPage() {
  const { key = "" } = useParams();
  const navigate = useNavigate();
  const [lesson, setLesson] = useState<ToneLesson | null>(null);
  const [step, setStep] = useState<"learn" | number | "done">("learn");
  const [error, setError] = useState("");
  const [lessonXP, setLessonXP] = useState(0);
  const [finish, setFinish] = useState<ToneComplete | null>(null);

  useEffect(() => {
    setStep("learn");
    setLessonXP(0);
    setFinish(null);
    api<TonesOverview>("/api/tones")
      .then((data) => {
        const found = data.lessons.find((item) => item.key === key);
        if (found) setLesson(found);
        else setError("Lesson not found");
      })
      .catch((err: Error) => setError(err.message));
  }, [key]);

  async function complete() {
    try {
      setFinish(await api<ToneComplete>(`/api/tones/lessons/${key}/complete`, { method: "POST" }));
      setStep("done");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not finish the lesson");
    }
  }

  if (!lesson) return <p className="p-6 text-muted">{error || "Loading…"}</p>;
  const practice = lesson.completed;
  const total = lesson.items.length;

  return (
    <div className="flex min-h-screen flex-col bg-canvas">
      <header className="mx-auto flex w-full max-w-2xl items-center gap-4 px-5 pt-5">
        <Link to="/tones" aria-label="Back to tones" className="grid h-11 w-11 shrink-0 place-items-center rounded-xl bg-white text-lg shadow-card">
          <Icon name="close" />
        </Link>
        <div className="h-2.5 flex-1 overflow-hidden rounded-full bg-line" role="progressbar" aria-label="Lesson progress" aria-valuemin={0} aria-valuemax={total} aria-valuenow={typeof step === "number" ? step : step === "done" ? total : 0}>
          <div
            className="h-full rounded-full bg-crimson transition-all duration-500"
            style={{ width: `${Math.max(4, ((typeof step === "number" ? step : step === "done" ? total : 0) / total) * 100)}%` }}
          />
        </div>
        <span className="flex shrink-0 items-center gap-1 whitespace-nowrap rounded-full bg-ink px-2.5 py-1 text-sm font-bold text-white" aria-label={practice ? "Practice, no XP" : `${lessonXP} XP this lesson`}>
          <Icon name="star" className="text-gold" />
          {practice ? "Practice" : `${lessonXP} XP`}
        </span>
      </header>
      {error && <p className="mx-auto mt-3 w-full max-w-2xl px-5 text-sm font-semibold text-crimson" role="alert">{error}</p>}

      <main className="mx-auto w-full max-w-2xl flex-1 px-5 pb-10 pt-6">
        {step === "learn" && <Learn lesson={lesson} practice={practice} onStart={() => setStep(0)} />}
        {typeof step === "number" && (
          <Say
            key={lesson.items[step].key}
            lesson={lesson}
            index={step}
            onXP={(n) => setLessonXP((x) => x + n)}
            onError={setError}
            onNext={() => (step + 1 >= total ? void complete() : setStep(step + 1))}
          />
        )}
        {step === "done" && finish && <Done lesson={lesson} finish={finish} xp={lessonXP} onNext={(next) => navigate(next ? `/tones/${next}` : "/tones")} />}
      </main>
    </div>
  );
}

function Learn({ lesson, practice, onStart }: { lesson: ToneLesson; practice: boolean; onStart: () => void }) {
  return (
    <div className="rise">
      <p className="label">{lesson.tone > 0 && lesson.tone < 5 ? `Tone ${lesson.tone}` : lesson.tone === 5 ? "Neutral tone" : "Practice"}</p>
      <h1 className="mt-1 text-3xl font-extrabold leading-tight">{lesson.title}</h1>
      {practice && (
        <p className="mt-3 rounded-2xl bg-white px-4 py-3 text-sm shadow-card">
          <b>Practice round.</b> You've finished this lesson, so it won't earn XP again, but it counts toward your streak.
        </p>
      )}

      {lesson.tone > 0 && (
        <div className="surface mt-5 flex items-center gap-4 p-4">
          <div className="flex h-24 shrink-0 flex-col justify-between py-1.5 text-right text-xs font-semibold text-muted" aria-hidden>
            <span>high</span>
            <span>low</span>
          </div>
          <ToneContour contour={lesson.contour} tone={lesson.tone} className="h-24 w-36 shrink-0" />
          <p className="text-sm text-muted">{lesson.contour ? `Pitch path ${lesson.contour.join("-")} on a 1 (low) to 5 (high) scale.` : "No contour of its own: short, light, mid height."}</p>
        </div>
      )}

      <div className="mt-5 grid gap-3 text-[17px] leading-7">
        {lesson.explain.map((paragraph) => (
          <p key={paragraph}>{paragraph}</p>
        ))}
      </div>
      <p className="mt-4 rounded-2xl bg-ink p-4 text-white">
        <b className="text-gold">Tip:</b> {lesson.tip}
      </p>

      <h2 className="mt-6 font-bold">Listen first</h2>
      <div className="mt-2 grid gap-3 sm:grid-cols-2">
        {lesson.examples.map((example) => (
          <button key={example.hanzi} type="button" onClick={() => speakChinese(example.hanzi, 0.75)} className="surface flex items-center gap-3 p-3 text-left">
            <span className="grid h-11 w-11 shrink-0 place-items-center rounded-full bg-[#F1EEEA] text-xl"><Icon name="volume" /></span>
            <span className="min-w-0">
              <span className="flex items-baseline gap-2">
                <b className="font-hanzi text-2xl">{example.hanzi}</b>
                <Pinyin tonal={example.tonal_pinyin} plain={example.pinyin} spaced className="font-semibold" />
              </span>
              <span className="block text-sm text-muted">{example.note}</span>
            </span>
          </button>
        ))}
      </div>

      <button type="button" onClick={onStart} className="btn-primary mt-7 h-14 w-full text-lg">
        <Icon name="mic" /> Say {lesson.items.length} words
      </button>
    </div>
  );
}

function Say({ lesson, index, onXP, onError, onNext }: { lesson: ToneLesson; index: number; onXP: (n: number) => void; onError: (m: string) => void; onNext: () => void }) {
  const item = lesson.items[index];
  const [result, setResult] = useState<ToneAssess | null>(null);
  const [busy, setBusy] = useState(false);
  const { recording, bind } = useRecorder((take) => void grade(take), onError);
  const last = index + 1 >= lesson.items.length;

  async function grade(take: Blob) {
    setBusy(true);
    const body = new FormData();
    body.set("audio", take, "take");
    body.set("lesson", lesson.key);
    body.set("item", item.key);
    try {
      const res = await api<ToneAssess>("/api/tones/assess", { method: "POST", body });
      setResult(res);
      onXP(res.xp_earned);
    } catch (err) {
      onError(err instanceof ApiError ? err.message : "Could not grade that take");
    } finally {
      setBusy(false);
    }
  }

  const expected = tonesOf(item.tonal_pinyin);
  const score = result ? Math.round(result.overall_score) : 0;
  const full = result ? tierPoints(result.xp_tier, "hanzi_speak") : 0;
  const micRef = useRef<HTMLButtonElement>(null);
  const nextRef = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    if (result) nextRef.current?.focus({ preventScroll: true });
    else micRef.current?.focus({ preventScroll: true });
  }, [result]);

  return (
    <div className="flex min-h-[70vh] flex-col">
      <h1 className="sr-only">
        {lesson.title}: say {item.hanzi}, word {index + 1} of {lesson.items.length}
      </h1>
      <section className="rounded-[32px] bg-white p-6 text-center shadow-float">
        <p className="break-words font-hanzi text-[clamp(4rem,18vw,7rem)] font-bold leading-none">{item.hanzi}</p>
        <Pinyin tonal={item.tonal_pinyin} plain={item.pinyin} spaced className="mt-3 block text-2xl font-bold" />
        <p className="text-muted">{item.english}</p>
        <div className="mt-3 flex justify-center gap-2">
          {expected.map((tone, i) => (
            <ToneContour key={i} contour={contours[tone]} tone={tone} className="h-10 w-14" grid={false} />
          ))}
        </div>
        <div className="mt-3 flex justify-center gap-2">
          <button type="button" className="btn-soft h-11" onClick={() => speakChinese(item.hanzi, 0.9)}>
            <Icon name="volume" /> Listen
          </button>
          <button type="button" className="btn-soft h-11" onClick={() => speakChinese(item.hanzi, 0.6)}>
            <Icon name="volume" /> Slow
          </button>
        </div>
      </section>

      {result && (
        <section className="rise mt-4 rounded-[28px] bg-maroon p-5 text-white" aria-live="polite">
          <div className="flex items-start justify-between gap-3">
            <div>
              <p className="text-sm font-semibold text-white/80">{tierLabel(result.xp_tier)}</p>
              <p className="text-5xl font-extrabold leading-none text-gold">{result.assessable ? score : "—"}</p>
            </div>
            {result.first_pass && result.xp_tier !== "" ? (
              <span className={`rounded-full px-2.5 py-1 text-sm font-bold ${result.xp_earned > 0 ? "bg-gold text-ink" : "bg-black/20 text-white/85"}`}>
                {result.xp_earned > 0 ? (result.xp_earned < full ? `+${result.xp_earned} XP · improved` : `+${result.xp_earned} XP`) : "Already earned"}
              </span>
            ) : (
              !result.first_pass && <span className="rounded-full bg-black/20 px-2.5 py-1 text-sm font-bold text-white/85">Practice</span>
            )}
          </div>
          {!result.assessable && <p className="mt-2 text-sm text-white/90">{result.reason || "That take couldn't be graded. Try again a little louder."}</p>}
          {result.syllables.length > 0 && (
            <ul className="mt-4 flex flex-wrap gap-2">
              {result.syllables.map((s, i) => {
                const ok = s.status === "ok";
                const want = String(expected[i] ?? "");
                const heard = s.heard_tone && s.heard_tone !== (s.expected_tone || want) ? s.heard_tone : "";
                return (
                  <li key={`${s.character}-${i}`} className={`rounded-2xl px-3 py-2 text-center ${ok ? "bg-white/15" : "bg-white text-ink"}`}>
                    <span className="block font-hanzi text-2xl font-bold">{s.character}</span>
                    <span className={`block text-xs font-bold ${ok ? "text-gold" : "text-crimson"}`}>
                      {ok ? `✓ ${toneLabel(s.expected_tone || want)}` : heard ? `heard ${toneLabel(heard)}, want ${toneLabel(s.expected_tone || want)}` : Math.round(s.score)}
                    </span>
                  </li>
                );
              })}
            </ul>
          )}
        </section>
      )}

      <div className="mt-auto pt-8">
        {!result ? (
          <div className="flex items-center justify-between gap-4">
            <span className="w-24" />
            <button
              type="button"
              ref={micRef}
              aria-label={recording ? "Release to grade" : "Hold to speak (or hold Space)"}
              disabled={busy}
              {...bind}
              className={`mic grid h-24 w-24 shrink-0 place-items-center rounded-full text-4xl text-white shadow-[0_12px_28px_rgba(200,16,46,.35)] transition ${recording ? "scale-110 bg-maroon ring-8 ring-crimson/25" : "bg-crimson"}`}
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
            <button ref={nextRef} type="button" className="btn-primary h-14 text-lg" onClick={onNext}>
              {last ? "Finish" : "Next"} <Icon name="right" />
            </button>
          </div>
        )}
        <p className="mt-3 text-center text-sm font-semibold text-muted">
          {result ? " " : recording ? "Listening… release when done" : busy ? "Grading…" : "Hold the mic and say the word"}
        </p>
      </div>
    </div>
  );
}

function Done({ lesson, finish, xp, onNext }: { lesson: ToneLesson; finish: ToneComplete; xp: number; onNext: (next: string) => void }) {
  const nextIsDifferent = finish.next && finish.next !== lesson.key;
  return (
    <div className="rise">
      <section className="relative overflow-hidden rounded-[28px] bg-hero p-6 text-center text-white">
        <Icon name="star" className="absolute left-5 top-5 text-3xl text-gold" />
        <p className="text-sm font-semibold uppercase tracking-[0.08em] text-white/85">{finish.first_time ? "Lesson complete" : "Practice done"}</p>
        <p className="mt-2 text-7xl font-extrabold leading-none text-gold">{Math.round(finish.best_score)}</p>
        <h1 className="mt-2 text-2xl font-extrabold">{lesson.title}</h1>
        {xp > 0 && (
          <p className="mt-3 inline-flex items-center gap-1.5 rounded-full bg-gold px-3.5 py-1.5 font-extrabold text-ink">
            <Icon name="star" /> +{xp} XP
          </p>
        )}
        {!finish.first_time && <p className="mt-3 text-white/90">Counted toward your streak.</p>}
      </section>
      <div className="mt-4 grid gap-3 sm:grid-cols-2">
        {nextIsDifferent ? (
          <button type="button" className="btn-primary h-14 text-lg" onClick={() => onNext(finish.next)}>
            <Icon name="play" /> Next lesson
          </button>
        ) : (
          <Link to="/" className="btn-primary h-14 text-lg">
            {finish.next ? "Back to Today" : "Tutorial done! Back to Today"}
          </Link>
        )}
        <Link to="/tones" className="btn-ghost h-14">All tone lessons</Link>
      </div>
    </div>
  );
}
