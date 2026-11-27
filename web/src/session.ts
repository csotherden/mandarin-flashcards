import { api } from "./api";
import { unlockSpeech } from "./speech";
import type { Session, SessionSource, StudyMode } from "./types";

export const modes: { id: StudyMode; title: string; detail: string }[] = [
  { id: "image_speak", title: "Picture → speak", detail: "See the photo, say the word" },
  { id: "hanzi_speak", title: "Hanzi → speak", detail: "See the characters, say the word" },
  { id: "picture_hanzi_match", title: "Picture ↔ hanzi", detail: "Match photos to characters" },
];

export function modeTitle(mode?: string) {
  return modes.find((item) => item.id === mode)?.title ?? modes[0].title;
}

/** Starts a drill and returns its id. count 0 means every card. */
export async function startSession(source: SessionSource, mode: StudyMode, count: number): Promise<string> {
  unlockSpeech();
  const body: Record<string, unknown> = { mode, card_count: count };
  if (source.kind === "all") body.all = true;
  if (source.kind === "sets") body.set_ids = source.setIds;
  if (source.kind === "lesson") body.lesson_id = source.lessonId;
  if (source.kind === "missed") body.source_attempt_id = source.attemptId;
  const session = await api<Session>("/api/study/sessions", { method: "POST", body: JSON.stringify(body) });
  return session.attempt.id;
}

/** Every way to practice a set: the graded drills plus Review (flip through, no grading). */
export type PracticeMode = StudyMode | "review";

export const practiceModes: { id: PracticeMode; title: string; detail: string }[] = [
  ...modes,
  { id: "review", title: "Review", detail: "Flip through the cards at your own pace, no grading" },
];

/** The review screen for some sets (or all cards): /review?sets=a,b or ?all=1, &count, &from. */
export function reviewPath(source: { setIds?: string[]; all?: boolean }, count = 0, from = "") {
  const params = new URLSearchParams();
  if (source.all || !source.setIds?.length) params.set("all", "1");
  else params.set("sets", source.setIds.join(","));
  if (count > 0) params.set("count", String(count));
  if (from) params.set("from", from);
  return `/review?${params.toString()}`;
}
