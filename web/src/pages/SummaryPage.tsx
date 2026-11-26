import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { api, ApiError } from "../api";
import { LaunchSheet } from "../components/LaunchSheet";
import type { Session, StudyMode } from "../types";

export function SummaryPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const [session, setSession] = useState<Session | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    api<Session>(`/api/study/sessions/${id}`).then(setSession).catch((err: Error) => setError(err.message));
  }, [id]);

  async function drillMissed(mode: StudyMode, count: number) {
    setBusy(true);
    try {
      const next = await api<Session>("/api/study/sessions", {
        method: "POST",
        body: JSON.stringify({ source_attempt_id: id, mode, card_count: count }),
      });
      navigate(`/drill/${next.attempt.id}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not start");
      setBusy(false);
    }
  }

  if (!session) return <p className="p-6 text-stone-500">{error || "Loading summary…"}</p>;
  const score = Math.round(session.overall_score ?? session.attempt.overall_score ?? 0);
  const byID = new Map(session.results.map((result) => [result.flashcard_id, result]));
  const missed = session.cards.filter((card) => (byID.get(card.id)?.overall_score ?? 0) < 85);

  return (
    <div className="mx-auto min-h-screen max-w-3xl px-4 py-6">
      <Link className="text-sm text-stone-500" to="/">← Library</Link>
      <p className="mt-4 text-sm uppercase tracking-[0.16em] text-cinnabar">Lesson summary</p>
      <p className={`mt-2 text-6xl font-semibold ${score >= 85 ? "text-moss" : "text-ink"}`}>{score}%</p>
      <ul className="mt-6 grid gap-2">
        {session.cards.map((card) => {
          const result = byID.get(card.id);
          const cardScore = Math.round(result?.overall_score ?? 0);
          return (
            <li key={card.id} className="rounded-2xl bg-white px-3 py-3">
              <div className="flex items-baseline justify-between">
                <span className="font-hanzi text-2xl">{card.hanzi}</span>
                <span className={cardScore >= 85 ? "text-moss" : "text-cinnabar"}>{cardScore}</span>
              </div>
              <p className="text-sm text-stone-500">{card.english}</p>
              {result?.coaching_tip && <p className="mt-1 text-sm">{result.coaching_tip}</p>}
            </li>
          );
        })}
      </ul>
      {error && <p className="mt-3 text-sm text-cinnabar">{error}</p>}
      {missed.length > 0 && (
        <div>
          <p className="mt-6 font-medium">Drill the {missed.length} words under 85</p>
          <LaunchSheet busy={busy} onStart={drillMissed} />
        </div>
      )}
    </div>
  );
}
