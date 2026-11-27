import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { api, ApiError } from "../api";
import { CardImage } from "../components/CardImage";
import { Icon } from "../components/Icon";
import { Pinyin } from "../components/Pinyin";
import { modeTitle, startSession } from "../session";
import type { Session } from "../types";

export function SummaryPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const [session, setSession] = useState<Session | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    api<Session>(`/api/study/sessions/${id}`).then(setSession).catch((err: Error) => setError(err.message));
  }, [id]);

  async function again(missedOnly: boolean) {
    if (!session) return;
    setBusy(true);
    setError("");
    try {
      const source = missedOnly ? { kind: "missed" as const, attemptId: id } : session.attempt.lesson_id ? { kind: "lesson" as const, lessonId: session.attempt.lesson_id } : { kind: "all" as const };
      navigate(`/drill/${await startSession(source, session.attempt.mode, missedOnly ? 0 : session.cards.length)}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not start");
      setBusy(false);
    }
  }

  if (!session) return <p className="p-6 text-muted">{error || "Loading…"}</p>;
  const byID = new Map(session.results.map((result) => [result.flashcard_id, result]));
  const practiced = session.cards.filter((card) => byID.has(card.id));
  const score = Math.round(session.overall_score ?? session.attempt.overall_score ?? 0);
  const missed = practiced.filter((card) => (byID.get(card.id)?.overall_score ?? 0) < 85);
  const headline = practiced.length === 0 ? "No cards graded" : score >= 90 ? "Excellent!" : score >= 75 ? "Nice work" : "Good practice";

  return (
    <div className="mx-auto min-h-screen max-w-2xl px-5 py-6">
      <section className="rise relative overflow-hidden rounded-[28px] bg-hero p-6 text-center text-white">
        <Icon name="star" className="absolute left-5 top-5 text-3xl text-gold" />
        <Icon name="star" className="absolute right-6 top-10 text-xl text-gold" />
        <p className="text-sm font-semibold uppercase tracking-[0.08em] text-white/85">{modeTitle(session.attempt.mode)}</p>
        <p className="mt-2 text-7xl font-extrabold leading-none text-gold">{practiced.length ? score : "—"}</p>
        <h1 className="mt-2 text-2xl font-extrabold">{headline}</h1>
        <p className="mt-1 text-white/90">
          {practiced.length} of {session.cards.length} cards · {practiced.length - missed.length} at 85 or above
        </p>
      </section>

      {error && <p className="mt-3 text-sm font-semibold text-crimson" role="alert">{error}</p>}
      <div className="mt-4 grid gap-3 sm:grid-cols-2">
        {missed.length > 0 ? (
          <button type="button" className="btn-primary h-14 text-lg" disabled={busy} onClick={() => again(true)}>
            <Icon name="play" /> Practice the {missed.length} that slipped
          </button>
        ) : (
          <button type="button" className="btn-primary h-14 text-lg" disabled={busy} onClick={() => again(false)}>
            <Icon name="play" /> Go again
          </button>
        )}
        <Link to="/" className="btn-ghost h-14">Done</Link>
      </div>

      <ul className="mt-6 grid grid-cols-[minmax(0,1fr)] gap-2.5">
        {session.cards.map((card) => {
          const result = byID.get(card.id);
          const cardScore = result ? Math.round(result.overall_score) : null;
          return (
            <li key={card.id} className="surface flex items-center gap-3 p-2.5">
              <CardImage card={card} className="h-14 w-14 shrink-0 rounded-2xl" />
              <div className="min-w-0 flex-1">
                <div className="flex items-baseline gap-2">
                  <b className="font-hanzi text-xl">{card.hanzi}</b>
                  <Pinyin tonal={card.tonal_pinyin} plain={card.pinyin} className="truncate text-sm font-semibold" />
                </div>
                <p className="truncate text-sm text-muted">{result?.coaching_tip || card.english}</p>
              </div>
              <span
                className={`grid h-11 w-12 shrink-0 place-items-center rounded-xl text-lg font-extrabold ${cardScore === null ? "bg-[#F1EEEA] text-muted" : cardScore >= 85 ? "bg-tone3/10 text-tone3" : "bg-crimson/10 text-crimson"}`}
              >
                {cardScore ?? "–"}
              </span>
            </li>
          );
        })}
      </ul>
    </div>
  );
}
