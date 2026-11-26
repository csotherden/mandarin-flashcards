import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api, ApiError } from "../api";
import { LaunchSheet } from "../components/LaunchSheet";
import type { CardSet, Session, StudyMode } from "../types";

export function SetsPage() {
  const [sets, setSets] = useState<CardSet[]>([]);
  const [title, setTitle] = useState("");
  const [error, setError] = useState("");
  const [launchID, setLaunchID] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const navigate = useNavigate();

  function load() {
    api<CardSet[]>("/api/sets").then(setSets).catch((err: Error) => setError(err.message));
  }
  useEffect(load, []);

  async function create() {
    try {
      await api("/api/sets", { method: "POST", body: JSON.stringify({ title, description: "" }) });
      setTitle("");
      load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not create the set");
    }
  }

  async function start(mode: StudyMode, count: number) {
    if (!launchID) return;
    setBusy(true);
    try {
      const session = await api<Session>("/api/study/sessions", {
        method: "POST",
        body: JSON.stringify({ set_id: launchID, mode, card_count: count }),
      });
      navigate(`/drill/${session.attempt.id}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not start");
      setBusy(false);
    }
  }

  return (
    <div>
      <h2 className="text-xl font-semibold">Sets</h2>
      <div className="mt-3 flex gap-2">
        <input className="min-w-0 flex-1 rounded-2xl border border-stone-300 bg-white px-3 py-3 text-base" placeholder="Section 3 sticky words" value={title} onChange={(event) => setTitle(event.target.value)} />
        <button type="button" className="rounded-2xl bg-ink px-4 text-paper" onClick={create}>Create</button>
      </div>
      {error && <p className="mt-3 text-sm text-cinnabar">{error}</p>}
      <ul className="mt-4 grid gap-3">
        {sets.map((set) => (
          <li key={set.id} className="rounded-3xl bg-white p-4">
            <Link className="block" to={`/sets/${set.id}`}>
              <p className="font-semibold">{set.title}</p>
              <p className="text-sm text-stone-500">{set.card_count} cards</p>
            </Link>
            <button type="button" className="mt-3 text-sm font-medium text-cinnabar" onClick={() => setLaunchID(launchID === set.id ? null : set.id)}>
              Quick start
            </button>
            {launchID === set.id && <LaunchSheet busy={busy} onStart={start} />}
          </li>
        ))}
      </ul>
    </div>
  );
}
