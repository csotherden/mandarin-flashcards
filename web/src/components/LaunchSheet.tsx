import { useState } from "react";
import type { StudyMode } from "../types";
import { unlockSpeech } from "../speech";

const modes: { id: StudyMode; title: string; detail: string }[] = [
  { id: "image_speak", title: "Picture → speak", detail: "See the photo, say the word" },
  { id: "hanzi_speak", title: "Hanzi → speak", detail: "See the characters, say the word" },
  { id: "picture_hanzi_match", title: "Picture ↔ hanzi", detail: "Match photos to characters" },
];

export function LaunchSheet({
  busy,
  onStart,
}: {
  busy: boolean;
  onStart: (mode: StudyMode, count: number) => void;
}) {
  const [mode, setMode] = useState<StudyMode>("image_speak");
  const [count, setCount] = useState(10);
  const [custom, setCustom] = useState(false);

  return (
    <section className="mt-6 rounded-3xl border border-stone-300 bg-white p-4">
      <h2 className="text-lg font-semibold">Start a drill</h2>
      <div className="mt-3 grid gap-2">
        {modes.map((item) => (
          <button
            key={item.id}
            type="button"
            onClick={() => setMode(item.id)}
            className={`rounded-2xl border px-3 py-3 text-left ${mode === item.id ? "border-ink bg-ink text-paper" : "border-stone-200"}`}
          >
            <span className="block font-medium">{item.title}</span>
            <span className={`block text-sm ${mode === item.id ? "text-stone-300" : "text-stone-500"}`}>{item.detail}</span>
          </button>
        ))}
      </div>
      <div className="mt-4 flex flex-wrap gap-2">
        {[5, 10, 15].map((n) => (
          <button
            key={n}
            type="button"
            onClick={() => {
              setCustom(false);
              setCount(n);
            }}
            className={`rounded-full px-3 py-2 text-sm ${!custom && count === n ? "bg-cinnabar text-white" : "bg-stone-100"}`}
          >
            {n}
          </button>
        ))}
        <button
          type="button"
          onClick={() => {
            setCustom(false);
            setCount(0);
          }}
          className={`rounded-full px-3 py-2 text-sm ${!custom && count === 0 ? "bg-cinnabar text-white" : "bg-stone-100"}`}
        >
          All
        </button>
        <button type="button" onClick={() => setCustom(true)} className={`rounded-full px-3 py-2 text-sm ${custom ? "bg-cinnabar text-white" : "bg-stone-100"}`}>
          Custom
        </button>
        {custom && (
          <input
            className="w-20 rounded-full border border-stone-300 px-3 py-2 text-base"
            inputMode="numeric"
            value={count || ""}
            onChange={(event) => setCount(Number(event.target.value) || 0)}
          />
        )}
      </div>
      <button
        type="button"
        disabled={busy}
        className="mt-4 w-full rounded-2xl bg-cinnabar py-3 font-semibold text-white disabled:opacity-60"
        onClick={() => {
          unlockSpeech();
          onStart(mode, count);
        }}
      >
        {busy ? "Starting…" : "Start"}
      </button>
    </section>
  );
}
