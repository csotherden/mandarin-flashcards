import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, ApiError } from "../api";
import type { Lesson } from "../types";

export function LessonsPage() {
  const [lessons, setLessons] = useState<Lesson[]>([]);
  const [title, setTitle] = useState("");
  const [error, setError] = useState("");

  function load() {
    api<Lesson[]>("/api/lessons").then(setLessons).catch((err: Error) => setError(err.message));
  }
  useEffect(load, []);

  async function create() {
    try {
      await api("/api/lessons", { method: "POST", body: JSON.stringify({ title, set_ids: [] }) });
      setTitle("");
      load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not create the lesson");
    }
  }

  return (
    <div>
      <h2 className="text-xl font-semibold">Lessons</h2>
      <div className="mt-3 flex gap-2">
        <input className="min-w-0 flex-1 rounded-2xl border border-stone-300 bg-white px-3 py-3 text-base" placeholder="Morning commute drill" value={title} onChange={(event) => setTitle(event.target.value)} />
        <button type="button" className="rounded-2xl bg-ink px-4 text-paper" onClick={create}>Create</button>
      </div>
      {error && <p className="mt-3 text-sm text-cinnabar">{error}</p>}
      <ul className="mt-4 grid gap-3">
        {lessons.map((lesson) => (
          <li key={lesson.id}>
            <Link className="block rounded-3xl bg-white p-4" to={`/lessons/${lesson.id}`}>
              <p className="font-semibold">{lesson.title}</p>
              <p className="text-sm text-stone-500">{lesson.enabled_count} of {lesson.card_count} cards on</p>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
