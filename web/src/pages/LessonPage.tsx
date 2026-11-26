import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { api, ApiError } from "../api";
import { LaunchSheet } from "../components/LaunchSheet";
import type { CardSet, LessonDetail, Session, StudyMode } from "../types";

export function LessonPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const [lesson, setLesson] = useState<LessonDetail | null>(null);
  const [sets, setSets] = useState<CardSet[]>([]);
  const [selected, setSelected] = useState<string[]>([]);
  const [enabled, setEnabled] = useState<Record<string, boolean>>({});
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    api<CardSet[]>("/api/sets").then(setSets).catch(() => undefined);
    api<LessonDetail>(`/api/lessons/${id}`)
      .then((row) => {
        setLesson(row);
        setSelected(row.set_ids);
        const flags: Record<string, boolean> = {};
        row.cards.forEach((item) => {
          flags[item.card.id] = item.enabled;
        });
        setEnabled(flags);
      })
      .catch((err: Error) => setError(err.message));
  }, [id]);

  async function saveSets() {
    setError("");
    try {
      const row = await api<LessonDetail>(`/api/lessons/${id}/sets`, { method: "PUT", body: JSON.stringify({ set_ids: selected }) });
      setLesson(row);
      const flags: Record<string, boolean> = {};
      row.cards.forEach((item) => {
        flags[item.card.id] = item.enabled;
      });
      setEnabled(flags);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not save sets");
    }
  }

  async function saveToggles() {
    const disabled = Object.entries(enabled).filter(([, on]) => !on).map(([cardID]) => cardID);
    try {
      const row = await api<LessonDetail>(`/api/lessons/${id}/disabled`, { method: "PUT", body: JSON.stringify({ flashcard_ids: disabled }) });
      setLesson(row);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not save cards");
    }
  }

  async function start(mode: StudyMode, count: number) {
    setBusy(true);
    try {
      const session = await api<Session>("/api/study/sessions", {
        method: "POST",
        body: JSON.stringify({ lesson_id: id, mode, card_count: count }),
      });
      navigate(`/drill/${session.attempt.id}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not start");
      setBusy(false);
    }
  }

  if (!lesson) return <p className="text-stone-500">{error || "Loading…"}</p>;

  return (
    <div>
      <Link className="text-sm text-stone-500" to="/lessons">← Lessons</Link>
      <h2 className="mt-2 text-xl font-semibold">{lesson.title}</h2>
      <p className="mt-4 text-sm font-medium">Sets in this lesson</p>
      <div className="mt-2 flex flex-wrap gap-2">
        {sets.map((set) => (
          <button
            key={set.id}
            type="button"
            onClick={() => setSelected((current) => current.includes(set.id) ? current.filter((item) => item !== set.id) : [...current, set.id])}
            className={`rounded-full px-3 py-1 text-sm ${selected.includes(set.id) ? "bg-ink text-paper" : "bg-white"}`}
          >
            {set.title}
          </button>
        ))}
      </div>
      <button type="button" className="mt-3 text-sm font-medium text-cinnabar" onClick={saveSets}>Save sets</button>
      <p className="mt-6 text-sm font-medium">Cards</p>
      <ul className="mt-2 grid gap-2">
        {lesson.cards.map((item) => (
          <li key={item.card.id} className="flex items-center justify-between rounded-2xl bg-white px-3 py-2">
            <span>
              <span className="font-hanzi text-lg">{item.card.hanzi}</span>
              <span className="ml-2 text-sm text-stone-500">{item.card.english}</span>
            </span>
            <button type="button" className={`rounded-full px-3 py-1 text-sm ${enabled[item.card.id] ? "bg-moss text-white" : "bg-stone-200"}`} onClick={() => setEnabled((current) => ({ ...current, [item.card.id]: !current[item.card.id] }))}>
              {enabled[item.card.id] ? "On" : "Off"}
            </button>
          </li>
        ))}
      </ul>
      <button type="button" className="mt-3 text-sm font-medium text-cinnabar" onClick={saveToggles}>Save card toggles</button>
      {error && <p className="mt-3 text-sm text-cinnabar">{error}</p>}
      <LaunchSheet busy={busy} onStart={start} />
    </div>
  );
}
