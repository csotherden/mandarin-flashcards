import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api";
import { Icon } from "../components/Icon";
import { ToneContour } from "../components/ToneContour";
import type { TonesOverview } from "../types";

/** Intro to tones: the lessons in order, what's done, and what's next. */
export function TonesPage() {
  const [data, setData] = useState<TonesOverview | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    api<TonesOverview>("/api/tones").then(setData).catch((err: Error) => setError(err.message));
  }, []);

  if (!data) return <p className="text-muted">{error || "Loading…"}</p>;
  const total = data.lessons.length;
  const finished = data.completed === total;
  const next = data.lessons.find((lesson) => lesson.key === data.next);

  return (
    <div className="mx-auto max-w-3xl">
      <section className="relative overflow-hidden rounded-[28px] bg-hero p-6 text-white lg:p-8">
        <p className="text-sm font-semibold uppercase tracking-[0.08em] text-white/85">{finished ? "Tone practice" : "Tutorial"}</p>
        <h1 className="mt-1.5 text-3xl font-extrabold leading-tight lg:text-4xl">Intro to tones</h1>
        <p className="mt-2 max-w-lg text-white/90">
          {finished
            ? "You've finished the tutorial. Come back to any lesson to tune your tones; practice counts toward your streak."
            : "Four tones and a neutral one. Learn how each sounds, then say real words and see which tone the grader hears."}
        </p>
        <div className="mt-4 flex items-center gap-3">
          <div className="h-2.5 flex-1 overflow-hidden rounded-full bg-white/20" role="progressbar" aria-label="Tutorial progress" aria-valuemin={0} aria-valuemax={total} aria-valuenow={data.completed}>
            <div className="h-full rounded-full bg-gold" style={{ width: `${(data.completed / total) * 100}%` }} />
          </div>
          <span className="text-sm font-bold">
            {data.completed} / {total}
          </span>
        </div>
        {next && (
          <Link to={`/tones/${next.key}`} className="mt-5 flex h-14 items-center justify-center gap-2 rounded-2xl bg-white text-lg font-extrabold text-maroon">
            <Icon name="play" /> {data.completed === 0 ? "Start the first lesson" : `Continue: ${next.title}`}
          </Link>
        )}
      </section>

      <ol className="mt-6 grid grid-cols-[minmax(0,1fr)] gap-3">
        {data.lessons.map((lesson) => {
          const isNext = lesson.key === data.next;
          return (
            <li key={lesson.key}>
              <Link to={`/tones/${lesson.key}`} className={`surface flex items-center gap-4 p-3 pr-4 transition hover:shadow-lift ${isNext ? "ring-2 ring-crimson" : ""}`}>
                <div className="grid h-16 w-20 shrink-0 place-items-center rounded-2xl bg-[#F7F5F2]">
                  {lesson.tone > 0 ? (
                    <ToneContour contour={lesson.contour} tone={lesson.tone} className="h-12 w-16" grid={false} />
                  ) : (
                    <span className="font-hanzi text-2xl font-bold text-ink">{lesson.items[0]?.hanzi}</span>
                  )}
                </div>
                <div className="min-w-0 flex-1">
                  <b className="block leading-snug">{lesson.title}</b>
                  <span className="block truncate text-sm text-muted">
                    {lesson.subtitle} · {lesson.items.length} words
                  </span>
                </div>
                {lesson.completed ? (
                  <span className="flex shrink-0 items-center gap-1 rounded-full bg-ink px-2.5 py-1 text-sm font-bold text-white">
                    <Icon name="check" className="text-gold" /> {Math.round(lesson.best_score)}
                  </span>
                ) : isNext ? (
                  <span className="shrink-0 rounded-full bg-crimson px-2.5 py-1 text-sm font-bold text-white">Next</span>
                ) : (
                  <span className="shrink-0 text-sm font-semibold text-muted">+XP</span>
                )}
              </Link>
            </li>
          );
        })}
      </ol>
      <p className="mt-4 text-center text-sm text-muted">Lessons earn XP the first time through. After that they're practice, and still count toward your streak.</p>
    </div>
  );
}
