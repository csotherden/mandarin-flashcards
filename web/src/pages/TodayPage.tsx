import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api, ApiError } from "../api";
import { CardImage } from "../components/CardImage";
import { Icon } from "../components/Icon";
import { Pinyin } from "../components/Pinyin";
import { SessionSheet } from "../components/SessionSheet";
import { BrandMark } from "../components/Shell";
import { StreakBadge } from "../components/StreakBadge";
import { modeTitle, startSession } from "../session";
import type { Home, SessionSource, StudyMode } from "../types";

export function TodayPage() {
  const navigate = useNavigate();
  const [home, setHome] = useState<Home | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState("");
  const [builder, setBuilder] = useState(false);

  function load() {
    api<Home>("/api/home")
      .then(setHome)
      .catch((err: Error) => setError(err.message));
  }
  useEffect(load, []);

  const mode: StudyMode = home?.last_mode ?? "image_speak";

  async function go(key: string, source: SessionSource, count: number, drill: StudyMode = mode) {
    setBusy(key);
    setError("");
    try {
      navigate(`/drill/${await startSession(source, drill, count)}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not start");
      setBusy("");
    }
  }

  async function forget(id: string, title: string) {
    if (!confirm(`Remove the quick start "${title}"? Its sets and cards stay.`)) return;
    await api(`/api/lessons/${id}`, { method: "DELETE" });
    load();
  }

  if (!home) return <p className="text-muted">{error || "Loading…"}</p>;
  const count = home.suggested_count;
  const minutes = Math.max(1, Math.round(count * 0.4));

  return (
    <div className="grid grid-cols-[minmax(0,1fr)] gap-7">
      <header className="flex items-center justify-between">
        <div className="flex items-center gap-3 whitespace-nowrap text-lg font-bold lg:hidden">
          <BrandMark />
          Mandarin Cards
        </div>
        <h1 className="hidden text-3xl font-extrabold lg:block">Today</h1>
        <StreakBadge days={home.streak_days} today={home.streak_today} />
      </header>

      {error && <p className="text-sm font-semibold text-crimson" role="alert">{error}</p>}

      {home.total_cards === 0 && (
        <TonesCard home={home} />
      )}
      {home.total_cards === 0 ? (
        <section className="relative overflow-hidden rounded-[28px] bg-hero p-6 text-white lg:p-8">
          <Icon name="star" className="absolute right-5 top-5 text-5xl text-gold" />
          <p className="text-sm font-semibold uppercase tracking-[0.08em] text-white/85">Welcome</p>
          <h2 className="mt-1.5 pr-14 text-3xl font-extrabold leading-tight lg:text-4xl">Make your first cards</h2>
          <p className="mt-2 text-white/90">Start with a ready-made set and a lesson, or build your own from words you choose.</p>
          <Link to="/library" className="mt-5 flex h-14 items-center justify-center gap-2 rounded-2xl bg-white text-lg font-extrabold text-maroon">
            <Icon name="play" /> Pick a set and start
          </Link>
          <div className="mt-3 grid grid-cols-2 gap-3">
            <Link to="/add" className="flex h-12 items-center justify-center gap-2 rounded-2xl border-2 border-white/80 font-extrabold text-white">
              <Icon name="plus" /> Add a word
            </Link>
            <Link to="/sets/generate" className="flex h-12 items-center justify-center gap-2 rounded-2xl border-2 border-white/80 font-extrabold text-white">
              <Icon name="spark" /> Generate
            </Link>
          </div>
        </section>
      ) : (
        <div className="grid gap-7 lg:grid-cols-[1.15fr_1fr] lg:items-start">
          <div className="grid gap-3.5">
            <section className="relative overflow-hidden rounded-[28px] bg-hero p-6 text-white lg:p-8">
              <Icon name="star" className="absolute right-5 top-5 text-5xl text-gold" />
              <p className="text-sm font-semibold uppercase tracking-[0.08em] text-white/85">Today's practice</p>
              <h2 className="mt-1.5 pr-14 text-3xl font-extrabold leading-tight lg:text-4xl">Ready for {count} cards?</h2>
              {home.streak_days > 0 && (
                <p className="mt-2 flex items-center gap-1.5 font-semibold text-white">
                  {home.streak_today ? (
                    <>
                      <Icon name="check" className="text-gold" /> Today counts. {home.streak_days}-day streak.
                    </>
                  ) : (
                    <>
                      <Icon name="star-outline" className="text-gold" /> Practice today to keep your {home.streak_days}-day streak.
                    </>
                  )}
                </p>
              )}
              <p className="mt-2 text-white/90">
                {home.slipped.length > 0 ? "Words that slipped come first, then the ones you see least." : "The words you have seen least come first."}
              </p>
              <div className="mt-3.5 flex flex-wrap gap-2 text-sm font-semibold">
                <span className="rounded-full bg-white/15 px-3 py-1.5">{modeTitle(mode)}</span>
                <span className="rounded-full bg-white/15 px-3 py-1.5">about {minutes} min</span>
                {home.practiced_today > 0 && <span className="rounded-full bg-white/15 px-3 py-1.5">{home.practiced_today} practiced today</span>}
                {home.xp_today > 0 && <span className="rounded-full bg-gold px-3 py-1.5 text-ink">{home.xp_today} XP today</span>}
              </div>
              <button
                type="button"
                disabled={!!busy}
                onClick={() => go("today", { kind: "all" }, count)}
                className="mt-5 flex h-14 w-full items-center justify-center gap-2 rounded-2xl bg-white text-lg font-extrabold text-maroon disabled:opacity-70"
              >
                <Icon name="play" /> {busy === "today" ? "Starting…" : "Practice now"}
              </button>
            </section>
            <button type="button" className="btn-ghost h-13 py-3.5" onClick={() => setBuilder(true)}>
              Build a custom session
            </button>
            <TonesCard home={home} />
          </div>

          <div className="grid grid-cols-[minmax(0,1fr)] gap-7">
            {(home.saved_sessions.length > 0 || home.sets.some((item) => item.set.card_count > 0)) && (
              <section>
                <div className="mb-3 flex items-baseline justify-between">
                  <h2 className="font-bold">Quick start</h2>
                  <Link to="/sets" className="link text-sm">All sets</Link>
                </div>
                <div className="no-scrollbar -mx-5 flex gap-3 overflow-x-auto px-5 pb-1 lg:mx-0 lg:grid lg:grid-cols-3 lg:px-0">
                  {home.saved_sessions.map((saved) => (
                    <div key={saved.id} className="surface relative min-w-[150px] p-2.5">
                      <button
                        type="button"
                        aria-label={`Remove ${saved.title}`}
                        onClick={() => forget(saved.id, saved.title)}
                        className="absolute right-2 top-2 z-10 grid h-7 w-7 place-items-center rounded-full bg-white/90 text-sm text-muted shadow-card"
                      >
                        <Icon name="close" />
                      </button>
                      <button type="button" className="block w-full text-left" disabled={!!busy || saved.card_count === 0} onClick={() => go(`l-${saved.id}`, { kind: "lesson", lessonId: saved.id }, 10)}>
                        <div className="grid h-[84px] place-items-center rounded-2xl bg-ink text-3xl text-gold"><Icon name="star" /></div>
                        <QuickLabel title={saved.title} detail={`${saved.card_count} cards · saved`} busy={busy === `l-${saved.id}`} />
                      </button>
                    </div>
                  ))}
                  {home.sets.filter((item) => item.set.card_count > 0).map(({ set, cover }) => (
                    <button key={set.id} type="button" disabled={!!busy} onClick={() => go(`s-${set.id}`, { kind: "sets", setIds: [set.id] }, 10)} className="surface min-w-[150px] p-2.5 text-left">
                      {cover ? <CardImage card={cover} className="h-[84px] w-full rounded-2xl" /> : <div className="h-[84px] rounded-2xl bg-[#F1EEEA]" />}
                      <QuickLabel title={set.title} detail={`${set.card_count} cards`} busy={busy === `s-${set.id}`} />
                    </button>
                  ))}
                </div>
              </section>
            )}

            {home.slipped.length > 0 && (
              <section>
                <div className="mb-3 flex items-baseline justify-between">
                  <h2 className="font-bold">Words that slipped</h2>
                  <button type="button" className="link text-sm" disabled={!!busy} onClick={() => go("slipped", { kind: "all" }, home.slipped.length, "hanzi_speak")}>
                    Practice these
                  </button>
                </div>
                <div className="grid grid-cols-3 gap-2.5">
                  {home.slipped.map(({ card }) => (
                    <Link key={card.id} to={`/cards/${card.id}`} className="surface px-2 py-3.5 text-center">
                      <div className="truncate font-hanzi text-2xl font-bold">{card.hanzi}</div>
                      <Pinyin tonal={card.tonal_pinyin} plain={card.pinyin} className="mt-0.5 block truncate text-sm font-semibold" />
                    </Link>
                  ))}
                </div>
              </section>
            )}
          </div>
        </div>
      )}

      <SessionSheet open={builder} onClose={() => setBuilder(false)} sets={home.sets.map((item) => item.set)} initialMode={mode} />
    </div>
  );
}

function QuickLabel({ title, detail, busy }: { title: string; detail: string; busy: boolean }) {
  return (
    <div className="mt-2.5 flex items-center justify-between gap-2 px-1">
      <div className="min-w-0">
        <b className="block truncate">{title}</b>
        <span className="text-sm text-muted">{busy ? "Starting…" : detail}</span>
      </div>
      <span className="grid h-8 w-8 shrink-0 place-items-center rounded-[10px] bg-crimson text-sm text-white">
        <Icon name="play" />
      </span>
    </div>
  );
}

/** Intro to tones: a prominent "continue" until it's done, then a quiet way back to practice. */
function TonesCard({ home }: { home: Home }) {
  if (home.tones_next) {
    const started = home.tones_completed > 0;
    return (
      <Link to={started ? `/tones/${home.tones_next}` : "/tones"} className="surface flex items-center gap-4 p-4 ring-2 ring-crimson/70">
        <span className="grid h-12 w-12 shrink-0 place-items-center rounded-2xl bg-ink font-hanzi text-2xl font-bold text-gold">声</span>
        <span className="min-w-0 flex-1">
          <b className="block">{started ? "Continue intro to tones" : "New to tones? Start here"}</b>
          <span className="block truncate text-sm text-muted">
            {started ? `${home.tones_completed} of ${home.tones_total} · next: ${home.tones_next_title}` : `${home.tones_total} short lessons with graded speaking`}
          </span>
        </span>
        <Icon name="right" className="shrink-0 text-xl text-crimson" />
      </Link>
    );
  }
  return (
    <Link to="/tones" className="flex items-center justify-center gap-2 rounded-2xl py-2 font-semibold text-muted underline-offset-4 hover:underline">
      <span className="font-hanzi">声</span> Tone practice
    </Link>
  );
}
