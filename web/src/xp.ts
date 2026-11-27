import type { StudyMode, XPTier } from "./types";

// Mirrors internal/xp: per word per day, the best tier's points; match drills earn half.
const speak: Record<XPTier, number> = { "": 0, keep: 2, close: 5, nice: 8, perfect: 12 };

export function tierPoints(tier: XPTier, mode: StudyMode) {
  const p = speak[tier] ?? 0;
  return mode === "picture_hanzi_match" ? Math.ceil(p / 2) : p;
}

export function tierLabel(tier: XPTier) {
  return { "": "Not graded", keep: "Keep at it", close: "Close", nice: "Nice!", perfect: "Perfect!" }[tier] ?? "";
}
