import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api, ApiError } from "../api";
import { modes, startSession } from "../session";
import type { CardSet, Lesson, StudyMode } from "../types";
import { Icon } from "./Icon";
import { Sheet } from "./Sheet";

/**
 * Builds a session in one sheet: what to practice (all cards or some sets), how many, which drill.
 * A set mix can be kept as a quick start (a saved lesson) for one-tap launches from Today.
 */
export function SessionSheet({
  open,
  onClose,
  sets,
  initialSetIds = [],
  initialMode = "image_speak",
  totalCards,
}: {
  open: boolean;
  onClose: () => void;
  sets: CardSet[];
  initialSetIds?: string[];
  initialMode?: StudyMode;
  /** How many cards "All cards" covers, to label the All option. */
  totalCards?: number;
}) {
  const navigate = useNavigate();
  const [setIds, setSetIds] = useState<string[]>(initialSetIds);
  const [count, setCount] = useState(10);
  const [mode, setMode] = useState<StudyMode>(initialMode);
  const [save, setSave] = useState(false);
  const [name, setName] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (open) {
      setSetIds(initialSetIds);
      setMode(initialMode);
      setSave(false);
      setName("");
      setError("");
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  const usable = sets.filter((set) => set.card_count > 0);
  const chosen = usable.filter((set) => setIds.includes(set.id));
  const available = setIds.length ? chosen.reduce((sum, set) => sum + set.card_count, 0) : totalCards;

  function toggle(id: string) {
    setSetIds((current) => (current.includes(id) ? current.filter((item) => item !== id) : [...current, id]));
  }

  async function start() {
    setBusy(true);
    setError("");
    try {
      let id: string;
      if (save && setIds.length) {
        const title = name.trim() || chosen.map((set) => set.title).join(" + ");
        const lesson = await api<Lesson>("/api/lessons", { method: "POST", body: JSON.stringify({ title, set_ids: setIds }) });
        id = await startSession({ kind: "lesson", lessonId: lesson.id }, mode, count);
      } else {
        id = await startSession(setIds.length ? { kind: "sets", setIds } : { kind: "all" }, mode, count);
      }
      navigate(`/drill/${id}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not start");
      setBusy(false);
    }
  }

  return (
    <Sheet open={open} onClose={onClose} title="Build a session">
      <p className="label">Practice from</p>
      <div className="mt-2 flex flex-wrap gap-2">
        <button type="button" className={`chip ${setIds.length === 0 ? "chip-on" : ""}`} onClick={() => setSetIds([])}>
          {setIds.length === 0 && <Icon name="check" />} All cards
        </button>
        {usable.map((set) => (
          <button key={set.id} type="button" className={`chip ${setIds.includes(set.id) ? "chip-on" : ""}`} onClick={() => toggle(set.id)}>
            {setIds.includes(set.id) && <Icon name="check" />} {set.title}
            <span className={setIds.includes(set.id) ? "text-white/80" : "text-muted"}>{set.card_count}</span>
          </button>
        ))}
      </div>

      <p className="label mt-5">How many</p>
      <div className="mt-2 flex gap-2">
        {[5, 10, 15, 0].map((n) => (
          <button key={n} type="button" className={`chip min-w-14 justify-center ${count === n ? "chip-on" : ""}`} onClick={() => setCount(n)}>
            {n === 0 ? `All${available ? ` ${available}` : ""}` : n}
          </button>
        ))}
      </div>

      <p className="label mt-5">Drill</p>
      <div className="mt-2 grid gap-2">
        {modes.map((item) => (
          <button
            key={item.id}
            type="button"
            onClick={() => setMode(item.id)}
            className={`rounded-2xl border-2 px-4 py-3 text-left transition ${mode === item.id ? "border-crimson bg-crimson/5" : "border-line bg-white"}`}
          >
            <span className="block font-bold">{item.title}</span>
            <span className="block text-sm text-muted">{item.detail}</span>
          </button>
        ))}
      </div>

      {setIds.length > 0 && (
        <div className="mt-5 rounded-2xl bg-[#F1EEEA] p-3">
          <label className="flex items-center gap-3 font-semibold">
            <input type="checkbox" className="h-5 w-5 accent-crimson" checked={save} onChange={(event) => setSave(event.target.checked)} />
            Keep as a quick start on Today
          </label>
          {save && (
            <input
              className="field mt-3"
              placeholder={chosen.map((set) => set.title).join(" + ") || "Name"}
              value={name}
              onChange={(event) => setName(event.target.value)}
            />
          )}
        </div>
      )}

      {error && <p className="mt-3 text-sm font-semibold text-crimson" role="alert">{error}</p>}
      <button type="button" className="btn-primary mt-5 h-14 w-full text-lg" disabled={busy} onClick={start}>
        <Icon name="play" /> {busy ? "Starting…" : "Start"}
      </button>
    </Sheet>
  );
}
